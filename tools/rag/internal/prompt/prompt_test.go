package prompt

import (
	"strings"
	"testing"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
)

func TestBuildAskPromptIncludesCitations(t *testing.T) {
	results := []indexer.SearchResult{
		{
			Chunk:    indexer.Chunk{Path: "docs/flow.md", LineStart: 10, LineEnd: 20, Content: "checkout rollback"},
			Citation: "[docs/flow.md:10-20]",
		},
	}
	prompt := BuildAskPrompt("where is rollback", results)
	if !strings.Contains(prompt, "where is rollback") {
		t.Fatalf("query missing from prompt")
	}
	if !strings.Contains(prompt, "[docs/flow.md:10-20]") {
		t.Fatalf("citation missing from prompt")
	}
}

func TestBuildLoopPromptIncludesAllowedCommands(t *testing.T) {
	prompt := BuildLoopPrompt("session-1", "fix tests", nil, []string{"make test-unit"}, []string{"tool: failed run"}, 1, 3)
	if !strings.Contains(prompt, "make test-unit") {
		t.Fatalf("allowed command missing from prompt")
	}
	if !strings.Contains(prompt, `"actions"`) {
		t.Fatalf("json contract missing from prompt")
	}
}
