package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type Client struct {
	baseURL *url.URL
	token   string
	http    *http.Client
}

type MergeRequest struct {
	ID          int      `json:"id"`
	IID         int      `json:"iid"`
	ProjectID   int      `json:"project_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DiffRefs    DiffRefs `json:"diff_refs"`
}

type DiffRefs struct {
	BaseSHA  string `json:"base_sha"`
	StartSHA string `json:"start_sha"`
	HeadSHA  string `json:"head_sha"`
}

type MergeRequestChanges struct {
	Changes []Change `json:"changes"`
}

type Change struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Diff    string `json:"diff"`
	NewFile bool   `json:"new_file"`
	Renamed bool   `json:"renamed_file"`
	Deleted bool   `json:"deleted_file"`
}

type CreateNoteRequest struct {
	Body string `json:"body"`
}

type CreateDiscussionRequest struct {
	Body     string             `json:"body"`
	Position DiscussionPosition `json:"position"`
}

type DiscussionPosition struct {
	BaseSHA  string `json:"base_sha"`
	StartSHA string `json:"start_sha"`
	HeadSHA  string `json:"head_sha"`
	NewPath  string `json:"new_path,omitempty"`
	OldPath  string `json:"old_path,omitempty"`
	NewLine  int    `json:"new_line,omitempty"`
	OldLine  int    `json:"old_line,omitempty"`
}

func NewClient(baseURL, token string, timeout time.Duration) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	return &Client{
		baseURL: parsed,
		token:   token,
		http:    &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) GetMergeRequest(ctx context.Context, projectID, mrIID int) (MergeRequest, error) {
	var mr MergeRequest
	endpoint := c.buildURL(fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d", projectID, mrIID))
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &mr); err != nil {
		return MergeRequest{}, err
	}
	return mr, nil
}

func (c *Client) GetMergeRequestChanges(ctx context.Context, projectID, mrIID int) (MergeRequestChanges, error) {
	var changes MergeRequestChanges
	endpoint := c.buildURL(fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/changes", projectID, mrIID))
	if err := c.doJSON(ctx, http.MethodGet, endpoint, nil, &changes); err != nil {
		return MergeRequestChanges{}, err
	}
	return changes, nil
}

func (c *Client) PostMergeRequestNote(ctx context.Context, projectID, mrIID int, body string) error {
	endpoint := c.buildURL(fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/notes", projectID, mrIID))
	payload := CreateNoteRequest{Body: body}
	return c.doJSON(ctx, http.MethodPost, endpoint, payload, nil)
}

func (c *Client) PostMergeRequestDiscussion(ctx context.Context, projectID, mrIID int, payload CreateDiscussionRequest) error {
	endpoint := c.buildURL(fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/discussions", projectID, mrIID))
	return c.doJSON(ctx, http.MethodPost, endpoint, payload, nil)
}

func (c *Client) buildURL(p string) string {
	cpy := *c.baseURL
	cpy.Path = path.Join(c.baseURL.Path, p)
	return cpy.String()
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab api error: %s", string(payload))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
