package prompt

import (
	"fmt"
	"strings"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
)

func BuildAskPrompt(query string, results []indexer.SearchResult) string {
	var b strings.Builder
	b.WriteString("You are answering questions about the current repository.\n")
	b.WriteString("Use only the repository context below. If the context is insufficient, say that explicitly.\n")
	b.WriteString("Every factual claim must include at least one citation in the format [path:start-end].\n\n")
	b.WriteString("Question:\n")
	b.WriteString(query)
	b.WriteString("\n\nRepository context:\n")
	for i, result := range results {
		fmt.Fprintf(&b, "Context %d %s\n", i+1, result.Citation)
		b.WriteString(result.Chunk.Content)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func BuildLoopPrompt(sessionID, task string, results []indexer.SearchResult, allowedCommands []string, events []string, iteration, maxIterations int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Session: %s\n", sessionID)
	fmt.Fprintf(&b, "Iteration: %d/%d\n\n", iteration, maxIterations)
	b.WriteString("Task:\n")
	b.WriteString(task)
	b.WriteString("\n\n")
	b.WriteString("Repository context:\n")
	for i, result := range results {
		fmt.Fprintf(&b, "Context %d %s\n", i+1, result.Citation)
		b.WriteString(result.Chunk.Content)
		b.WriteString("\n\n")
	}
	b.WriteString("Allowed commands:\n")
	for _, command := range allowedCommands {
		b.WriteString("- ")
		b.WriteString(command)
		b.WriteString("\n")
	}
	if len(events) > 0 {
		b.WriteString("\nExecution log:\n")
		for _, event := range events {
			b.WriteString(event)
			if !strings.HasSuffix(event, "\n") {
				b.WriteString("\n")
			}
		}
	}
	b.WriteString(`

Respond with a single JSON object only, with this shape:
{
  "summary": "short status update",
  "done": false,
  "commit_message": "feat: concise summary",
  "actions": [
    {"type": "read_file", "path": "relative/path.go", "start_line": 1, "end_line": 120},
    {"type": "write_file", "path": "relative/path.go", "content": "entire file content"},
    {"type": "delete_file", "path": "relative/path.go"},
    {"type": "run", "command": "make test-unit"}
  ]
}

Rules:
- Paths must be repository-relative, never absolute, and must not contain "..".
- Commands must exactly match one of the allowed commands.
- Prefer small, incremental steps.
- After a failing command, inspect files or modify code before retrying.
- Set "done" to true only when the code is ready for final validation and commit.
`)
	return strings.TrimSpace(b.String())
}
