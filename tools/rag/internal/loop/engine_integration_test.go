package loop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
	"github.com/falconfan123/Go-mall/tools/rag/internal/storage"
)

type fakeModel struct {
	responses []string
	index     int
}

func (f *fakeModel) Generate(_ context.Context, _ model.GenerateRequest) (model.GenerateResponse, error) {
	if f.index >= len(f.responses) {
		return model.GenerateResponse{}, fmt.Errorf("no more responses")
	}
	resp := f.responses[f.index]
	f.index++
	return model.GenerateResponse{Text: resp, Model: "fake"}, nil
}

func (f *fakeModel) Name() string { return "fake" }

type fakeRunner struct {
	commands []string
	outputs  map[string]string
}

func (f *fakeRunner) Run(_ context.Context, dir, command string) (string, error) {
	f.commands = append(f.commands, command)
	if strings.HasPrefix(command, "git ") {
		return "", nil
	}
	if out, ok := f.outputs[command]; ok {
		return out, nil
	}
	return "", nil
}

func TestEngineRunCompletesWithFakeModelAndRunner(t *testing.T) {
	repo := t.TempDir()
	sourcePath := filepath.Join(repo, "services", "sample", "sample.go")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("package sample\n\nfunc Value() string { return \"old\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := storage.Open(filepath.Join(t.TempDir(), "rag.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	chunks, err := indexer.Build(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceChunks(context.Background(), chunks.Chunks); err != nil {
		t.Fatal(err)
	}

	newContent := "package sample\n\nfunc Value() string { return \"new\" }\n"
	m := &fakeModel{responses: []string{
		fmt.Sprintf(`{"summary":"update sample","done":true,"commit_message":"feat: update sample","actions":[{"type":"write_file","path":"services/sample/sample.go","content":%q},{"type":"run","command":"make test-unit"}]}`, newContent),
	}}
	runner := &fakeRunner{outputs: map[string]string{
		"make test-unit":     "ok",
		"make lint":          "ok",
		"git status --short": " M services/sample/sample.go",
	}}
	engine := Engine{
		Model:           m,
		Runner:          runner,
		Store:           store,
		RepoRoot:        repo,
		Chunks:          chunks.Chunks,
		AutoCommit:      true,
		FinalValidation: []string{"make lint", "make test-unit"},
		Session: storage.Session{
			ID:              "session-1",
			Prompt:          "update sample",
			Status:          "running",
			RepoRoot:        repo,
			WorktreePath:    repo,
			Branch:          "rag/session-1",
			AllowedCommands: []string{"make test-unit", "make lint"},
			MaxIterations:   2,
		},
	}

	result, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Engine.Run() error = %v", err)
	}
	if result.Session.Status != "completed" {
		t.Fatalf("expected completed status, got %s", result.Session.Status)
	}
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != newContent {
		t.Fatalf("unexpected file content: %s", string(raw))
	}
	if len(runner.commands) < 3 {
		t.Fatalf("expected command execution and final validation, got %v", runner.commands)
	}
}
