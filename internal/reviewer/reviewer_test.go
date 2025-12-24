package reviewer

import (
	"strings"
	"testing"
)

func TestDiffHasNewLine(t *testing.T) {
	diff := "@@ -1,2 +1,3 @@\n line1\n+line2\n line3\n"
	if !diffHasNewLine(diff, 2) {
		t.Fatal("expected line 2 to be in diff")
	}
	if diffHasNewLine(diff, 5) {
		t.Fatal("did not expect line 5 to be in diff")
	}
}

func TestParseHunkHeader(t *testing.T) {
	newLine, oldLine := parseHunkHeader("@@ -10,7 +20,9 @@")
	if newLine != 20 || oldLine != 10 {
		t.Fatalf("unexpected parse result: new %d old %d", newLine, oldLine)
	}
}

func TestRenderSummaryFallback(t *testing.T) {
	fallback := []ReviewComment{{File: "main.go", Line: 12, Comment: "test"}}
	result := renderSummary("summary", fallback)
	if !strings.Contains(result, "main.go:12") {
		t.Fatal("expected fallback entry in summary")
	}
}
