package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
)

func TestLoadModelSupportsNone(t *testing.T) {
	client, backend, err := loadModel("none")
	if err != nil {
		t.Fatalf("loadModel(none) error = %v", err)
	}
	if client != nil {
		t.Fatal("loadModel(none) client != nil")
	}
	if backend != "none" {
		t.Fatalf("backend = %q, want none", backend)
	}
}

func TestLoadModelRejectsUnknownBackend(t *testing.T) {
	_, _, err := loadModel("bogus")
	if err == nil {
		t.Fatal("loadModel(bogus) error = nil, want error")
	}
}

func TestDoctorIncludesIndexedSources(t *testing.T) {
	repoRoot := t.TempDir()
	server, err := NewServer(repoRoot)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	store, err := server.cli.OpenStore()
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	defer store.Close()
	err = store.ReplaceChunks(context.Background(), []indexer.Chunk{
		{Path: "dtm.pub/docs/arch.md", LineStart: 1, LineEnd: 3, Content: "DTM", Priority: 1, Hash: "1", SourceKind: "external_doc", SourceName: "dtm.pub"},
	})
	if err != nil {
		t.Fatalf("ReplaceChunks() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/doctor", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp doctorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	found := false
	for _, item := range resp.IndexedSources {
		if item.Name == "dtm.pub" && item.Count == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected indexed source dtm.pub in response, got %#v", resp.IndexedSources)
	}
}

func TestSearchResultsForTestsIncludesSourcePaths(t *testing.T) {
	results := SearchResultsForTests([]indexer.SearchResult{{
		Chunk: indexer.Chunk{
			Path:      filepath.ToSlash("dtm.pub/docs/practice/arch.md"),
			LineStart: 1,
			LineEnd:   10,
			Content:   "DTM",
		},
		Citation: "[dtm.pub/docs/practice/arch.md:1-10]",
		Score:    42,
	}})
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].Path != "dtm.pub/docs/practice/arch.md" {
		t.Fatalf("path = %q", results[0].Path)
	}
}
