package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
)

func TestDoctorAuthReadiness(t *testing.T) {
	t.Run("token only", func(t *testing.T) {
		report := runDoctorWithEnv(t, "token-123", "")
		if !report.AuthTokenReady {
			t.Fatal("AuthTokenReady = false, want true")
		}
		if report.APIKeyReady {
			t.Fatal("APIKeyReady = true, want false")
		}
		if !report.AuthReady {
			t.Fatal("AuthReady = false, want true")
		}
	})

	t.Run("api key only", func(t *testing.T) {
		report := runDoctorWithEnv(t, "", "key-123")
		if report.AuthTokenReady {
			t.Fatal("AuthTokenReady = true, want false")
		}
		if !report.APIKeyReady {
			t.Fatal("APIKeyReady = false, want true")
		}
		if !report.AuthReady {
			t.Fatal("AuthReady = false, want true")
		}
	})

	t.Run("missing credentials", func(t *testing.T) {
		report := runDoctorWithEnv(t, "", "")
		if report.AuthTokenReady {
			t.Fatal("AuthTokenReady = true, want false")
		}
		if report.APIKeyReady {
			t.Fatal("APIKeyReady = true, want false")
		}
		if report.AuthReady {
			t.Fatal("AuthReady = true, want false")
		}
	})
}

func TestIndexIncludesExistingExternalRoots(t *testing.T) {
	repoRoot := t.TempDir()
	writeAppTestFile(t, filepath.Join(repoRoot, "docs", "guide.md"), "repo content\n")

	externalRoot := t.TempDir()
	writeAppTestFile(t, filepath.Join(externalRoot, "services", "order", "main.go"), "package order\n// external content\n")

	oldRoots := extraIndexRoots
	extraIndexRoots = []string{externalRoot}
	t.Cleanup(func() {
		extraIndexRoots = oldRoots
	})

	cli := &CLI{
		RepoRoot: repoRoot,
		DataDir:  filepath.Join(repoRoot, ".artifacts", "rag"),
		DBPath:   filepath.Join(repoRoot, ".artifacts", "rag", "rag.db"),
	}

	result, err := cli.Index(context.Background())
	if err != nil {
		t.Fatalf("Index() error = %v", err)
	}
	if len(result.Chunks) == 0 {
		t.Fatal("Index() returned no chunks")
	}

	store, err := cli.OpenStore()
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	defer store.Close()

	chunks, err := store.LoadChunks(context.Background())
	if err != nil {
		t.Fatalf("LoadChunks() error = %v", err)
	}
	wantPath := filepath.ToSlash(filepath.Join(filepath.Base(externalRoot), "services", "order", "main.go"))
	if !containsChunkPath(chunks, wantPath) {
		t.Fatalf("expected indexed external path %q, got %#v", wantPath, chunkPaths(chunks))
	}
}

func TestDoctorReportsIndexedSources(t *testing.T) {
	repoRoot := t.TempDir()
	cli := &CLI{
		RepoRoot: repoRoot,
		DataDir:  filepath.Join(repoRoot, ".artifacts", "rag"),
		DBPath:   filepath.Join(repoRoot, ".artifacts", "rag", "rag.db"),
	}

	store, err := cli.OpenStore()
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	defer store.Close()
	err = store.ReplaceChunks(context.Background(), []indexer.Chunk{
		{Path: "dtm.pub/docs/practice/arch.md", LineStart: 1, LineEnd: 10, Content: "DTM", Priority: 1, Hash: "a", SourceKind: "external_doc", SourceName: "dtm.pub"},
		{Path: "services/a.go", LineStart: 1, LineEnd: 2, Content: "repo", Priority: 1, Hash: "b", SourceKind: "repo", SourceName: ""},
	})
	if err != nil {
		t.Fatalf("ReplaceChunks() error = %v", err)
	}

	report, err := cli.Doctor(context.Background())
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	if report.IndexedSources["dtm.pub"] != 1 {
		t.Fatalf("expected dtm.pub source count 1, got %#v", report.IndexedSources)
	}
	if report.IndexedSources["repo"] != 1 {
		t.Fatalf("expected repo source count 1, got %#v", report.IndexedSources)
	}
}

func runDoctorWithEnv(t *testing.T, authToken, apiKey string) DoctorReport {
	t.Helper()

	repoRoot := t.TempDir()
	cli := &CLI{
		RepoRoot: repoRoot,
		DataDir:  filepath.Join(repoRoot, ".artifacts", "rag"),
		DBPath:   filepath.Join(repoRoot, ".artifacts", "rag", "rag.db"),
	}
	t.Setenv("ANTHROPIC_AUTH_TOKEN", authToken)
	t.Setenv("ANTHROPIC_API_KEY", apiKey)

	report, err := cli.Doctor(context.Background())
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	return report
}

func writeAppTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func containsChunkPath(chunks []indexer.Chunk, want string) bool {
	for _, chunk := range chunks {
		if chunk.Path == want {
			return true
		}
	}
	return false
}

func chunkPaths(chunks []indexer.Chunk) []string {
	seen := map[string]struct{}{}
	var paths []string
	for _, chunk := range chunks {
		if _, ok := seen[chunk.Path]; ok {
			continue
		}
		seen[chunk.Path] = struct{}{}
		paths = append(paths, chunk.Path)
	}
	return paths
}
