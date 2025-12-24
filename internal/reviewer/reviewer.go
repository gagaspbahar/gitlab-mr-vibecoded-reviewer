package reviewer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gitlab-mr-vibecoded-reviewer/internal/gitlab"
	"gitlab-mr-vibecoded-reviewer/internal/llm"
)

type Reviewer struct {
	gitlab *gitlab.Client
	llm    *llm.Client
}

type ReviewResponse struct {
	Summary  string          `json:"summary"`
	Comments []ReviewComment `json:"comments"`
}

type ReviewComment struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Comment string `json:"comment"`
}

type InlineResult struct {
	Posted   int
	Fallback []ReviewComment
}

func New(gitlabClient *gitlab.Client, llmClient *llm.Client) *Reviewer {
	return &Reviewer{
		gitlab: gitlabClient,
		llm:    llmClient,
	}
}

func (r *Reviewer) Run(ctx context.Context, projectID, mrIID int, note string) error {
	mr, err := r.gitlab.GetMergeRequest(ctx, projectID, mrIID)
	if err != nil {
		return fmt.Errorf("fetch merge request: %w", err)
	}
	changes, err := r.gitlab.GetMergeRequestChanges(ctx, projectID, mrIID)
	if err != nil {
		return fmt.Errorf("fetch merge request changes: %w", err)
	}

	prompt := buildPrompt(mr, changes, note)
	content, err := r.llm.ChatCompletion(ctx, prompt)
	if err != nil {
		return fmt.Errorf("llm completion failed: %w", err)
	}

	var review ReviewResponse
	if err := json.Unmarshal([]byte(content), &review); err != nil {
		return fmt.Errorf("parse llm response: %w", err)
	}

	inlineResult, err := r.postInlineComments(ctx, mr, changes, review.Comments)
	if err != nil {
		return err
	}

	summaryBody := renderSummary(review.Summary, inlineResult.Fallback)
	if err := r.gitlab.PostMergeRequestNote(ctx, projectID, mrIID, summaryBody); err != nil {
		return fmt.Errorf("post summary note: %w", err)
	}
	return nil
}

func buildPrompt(mr gitlab.MergeRequest, changes gitlab.MergeRequestChanges, note string) []llm.ChatMessage {
	var diffBuilder strings.Builder
	for _, change := range changes.Changes {
		diffBuilder.WriteString("File: ")
		diffBuilder.WriteString(change.NewPath)
		diffBuilder.WriteString("\n")
		diffBuilder.WriteString(change.Diff)
		diffBuilder.WriteString("\n\n")
	}

	instruction := `You are a senior code reviewer. Summarize the merge request and suggest improvements.
Return ONLY valid JSON with this shape:
{
  "summary": "...",
  "comments": [
    {"file": "path/to/file.go", "line": 42, "comment": "..."}
  ]
}
If no inline comments are needed, return an empty comments array.`

	userContent := fmt.Sprintf("Merge Request Title: %s\nDescription: %s\nRequester Note: %s\n\nDiffs:\n%s", mr.Title, mr.Description, note, diffBuilder.String())

	return []llm.ChatMessage{
		{Role: "system", Content: instruction},
		{Role: "user", Content: userContent},
	}
}

func (r *Reviewer) postInlineComments(ctx context.Context, mr gitlab.MergeRequest, changes gitlab.MergeRequestChanges, comments []ReviewComment) (InlineResult, error) {
	result := InlineResult{}
	for _, comment := range comments {
		change, ok := findChange(changes, comment.File)
		if !ok || comment.Line <= 0 {
			result.Fallback = append(result.Fallback, comment)
			continue
		}
		if !diffHasNewLine(change.Diff, comment.Line) {
			result.Fallback = append(result.Fallback, comment)
			continue
		}
		payload := gitlab.CreateDiscussionRequest{
			Body: comment.Comment,
			Position: gitlab.DiscussionPosition{
				BaseSHA:  mr.DiffRefs.BaseSHA,
				StartSHA: mr.DiffRefs.StartSHA,
				HeadSHA:  mr.DiffRefs.HeadSHA,
				NewPath:  change.NewPath,
				NewLine:  comment.Line,
			},
		}
		if err := r.gitlab.PostMergeRequestDiscussion(ctx, mr.ProjectID, mr.IID, payload); err != nil {
			return result, fmt.Errorf("post inline discussion: %w", err)
		}
		result.Posted++
	}
	return result, nil
}

func renderSummary(summary string, fallback []ReviewComment) string {
	var builder strings.Builder
	builder.WriteString("### 🤖 Review Summary\n")
	builder.WriteString(summary)
	builder.WriteString("\n\n")
	if len(fallback) == 0 {
		return builder.String()
	}
	builder.WriteString("### 💬 Additional Suggestions\n")
	for _, comment := range fallback {
		builder.WriteString("- ")
		builder.WriteString(comment.File)
		if comment.Line > 0 {
			builder.WriteString(fmt.Sprintf(":%d", comment.Line))
		}
		builder.WriteString(" — ")
		builder.WriteString(comment.Comment)
		builder.WriteString("\n")
	}
	return builder.String()
}

func findChange(changes gitlab.MergeRequestChanges, file string) (gitlab.Change, bool) {
	for _, change := range changes.Changes {
		if change.NewPath == file || change.OldPath == file {
			return change, true
		}
	}
	return gitlab.Change{}, false
}

func diffHasNewLine(diff string, target int) bool {
	lines := strings.Split(diff, "\n")
	newLine := 0
	oldLine := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			parsedNew, parsedOld := parseHunkHeader(line)
			if parsedNew != 0 {
				newLine = parsedNew
				oldLine = parsedOld
			}
			continue
		}
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			if newLine == target {
				return true
			}
			newLine++
			continue
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			oldLine++
			continue
		}
		if newLine == target {
			return true
		}
		newLine++
		oldLine++
	}
	return false
}

func parseHunkHeader(line string) (int, int) {
	// Example: @@ -10,7 +10,9 @@
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return 0, 0
	}
	oldRange := strings.TrimPrefix(parts[1], "-")
	newRange := strings.TrimPrefix(parts[2], "+")
	oldLine := parseRangeStart(oldRange)
	newLine := parseRangeStart(newRange)
	return newLine, oldLine
}

func parseRangeStart(input string) int {
	chunk := strings.SplitN(input, ",", 2)
	if len(chunk) == 0 {
		return 0
	}
	var value int
	_, _ = fmt.Sscanf(chunk[0], "%d", &value)
	return value
}
