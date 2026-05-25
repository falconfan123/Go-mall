package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAndSearchPrioritizesRepositoryHotspots(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "CLAUDE.md"), "# Guide\ncheckout rollback details\n")
	writeTestFile(t, filepath.Join(root, "docs", "flow.md"), "checkout rollback docs\n")
	writeTestFile(t, filepath.Join(root, "services", "checkout", "logic.go"), "package checkout\n// rollback checkout status\n")
	writeTestFile(t, filepath.Join(root, "test", "checkout_test.go"), "rollback test case\n")
	writeTestFile(t, filepath.Join(root, "go.sum"), "checkout rollback noise\n")
	writeTestFile(t, filepath.Join(root, "frontend", "ignored.ts"), "const ignored = true;\n")

	result, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.FilesIndexed < 4 {
		t.Fatalf("expected at least 4 indexed files, got %d", result.FilesIndexed)
	}
	results := Search(result.Chunks, "checkout rollback", 3)
	if len(results) != 3 {
		t.Fatalf("expected 3 search results, got %d", len(results))
	}
	if got := results[0].Chunk.Path; got != "CLAUDE.md" {
		t.Fatalf("expected top result CLAUDE.md, got %s", got)
	}
	if !strings.Contains(results[0].Citation, "CLAUDE.md:1-") {
		t.Fatalf("unexpected citation %s", results[0].Citation)
	}
	for _, result := range results {
		if result.Chunk.Path == "go.sum" {
			t.Fatalf("unexpected dependency noise in results: %+v", result)
		}
	}
}

func TestChunkingProducesLineRanges(t *testing.T) {
	root := t.TempDir()
	var lines []string
	for i := 0; i < 130; i++ {
		lines = append(lines, "line")
	}
	writeTestFile(t, filepath.Join(root, "docs", "big.md"), strings.Join(lines, "\n"))
	result, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(result.Chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(result.Chunks))
	}
	if result.Chunks[1].LineStart >= result.Chunks[1].LineEnd {
		t.Fatalf("invalid line range: %+v", result.Chunks[1])
	}
}

func TestBuildIgnoresNoiseFilesAndDirectories(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "services", "checkout", "logic.go"), "package checkout\nfunc doctor() {}\n")
	writeTestFile(t, filepath.Join(root, "docs", "guide.md"), "doctor checks writable and git\n")
	writeTestFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n")
	writeTestFile(t, filepath.Join(root, "go.sum"), "noise\n")
	writeTestFile(t, filepath.Join(root, "go.work.sum"), "noise\n")
	writeTestFile(t, filepath.Join(root, "package-lock.json"), "{}\n")
	writeTestFile(t, filepath.Join(root, "pnpm-lock.yaml"), "lockfileVersion: 9\n")
	writeTestFile(t, filepath.Join(root, "yarn.lock"), "noise\n")
	writeTestFile(t, filepath.Join(root, "dist", "bundle.md"), "should be ignored\n")
	writeTestFile(t, filepath.Join(root, "build", "artifact.txt"), "should be ignored\n")
	writeTestFile(t, filepath.Join(root, "node_modules", "pkg", "README.md"), "should be ignored\n")
	writeTestFile(t, filepath.Join(root, ".git", "config"), "should be ignored\n")
	writeTestFile(t, filepath.Join(root, ".idea", "workspace.xml"), "<xml/>\n")
	writeTestFile(t, filepath.Join(root, ".vscode", "settings.json"), "{}\n")
	writeTestFile(t, filepath.Join(root, "bin", "tool.exe"), "binary\n")
	writeTestFile(t, filepath.Join(root, "bin", "blob.bin"), "binary\n")

	result, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	got := indexedPaths(result.Chunks)
	wantPresent := []string{
		"docs/guide.md",
		"go.mod",
		"services/checkout/logic.go",
	}
	for _, path := range wantPresent {
		if _, ok := got[path]; !ok {
			t.Fatalf("expected %s to be indexed, got paths=%v", path, mapsKeys(got))
		}
	}
	wantAbsent := []string{
		".idea/workspace.xml",
		".vscode/settings.json",
		"bin/blob.bin",
		"bin/tool.exe",
		"build/artifact.txt",
		"dist/bundle.md",
		"go.sum",
		"go.work.sum",
		"node_modules/pkg/README.md",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
	}
	for _, path := range wantAbsent {
		if _, ok := got[path]; ok {
			t.Fatalf("expected %s to be ignored, got paths=%v", path, mapsKeys(got))
		}
	}
}

func TestBuildSkipsLargeAndLongLineFiles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "ok.md"), "doctor checks git\n")
	writeTestFile(t, filepath.Join(root, "docs", "large.md"), strings.Repeat("a", maxFileSize+1))
	writeTestFile(t, filepath.Join(root, "docs", "long-line.md"), strings.Repeat("b", maxLineLength+1)+"\nshort\n")

	result, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	got := indexedPaths(result.Chunks)
	if _, ok := got["docs/ok.md"]; !ok {
		t.Fatalf("expected docs/ok.md to be indexed, got paths=%v", mapsKeys(got))
	}
	if _, ok := got["docs/large.md"]; ok {
		t.Fatalf("expected docs/large.md to be skipped, got paths=%v", mapsKeys(got))
	}
	if _, ok := got["docs/long-line.md"]; ok {
		t.Fatalf("expected docs/long-line.md to be skipped, got paths=%v", mapsKeys(got))
	}
}

func TestSearchPrefersSourceFilesOverConfigFiles(t *testing.T) {
	root := t.TempDir()
	content := "doctor auth ready git writable checks\n"
	writeTestFile(t, filepath.Join(root, "services", "rag", "logic.go"), "package rag\n// "+content)
	writeTestFile(t, filepath.Join(root, "services", "rag", "config.json"), "{\n\"note\": \""+strings.TrimSpace(content)+"\"\n}\n")

	result, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	results := Search(result.Chunks, "doctor auth ready", 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 search results, got %d", len(results))
	}
	if got := results[0].Chunk.Path; got != "services/rag/logic.go" {
		t.Fatalf("expected source file to rank first, got %s", got)
	}
}

func TestSearchPrefersExplicitPathMatchesOverGoModNoise(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "tools", "rag", "internal", "app", "app.go"), "package app\n// Doctor checks writable git auth_ready\n")
	writeTestFile(t, filepath.Join(root, "services", "activity", "go.mod"), "module example.com/activity\ngo 1.25.0\n// Doctor checks\n")
	writeTestFile(t, filepath.Join(root, "services", "search", "go.mod"), "module example.com/search\ngo 1.25.0\n// Doctor checks\n")

	result, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	results := Search(result.Chunks, "tools/rag/internal/app/app.go 里的 Doctor 会检查哪些项？", 3)
	if len(results) == 0 {
		t.Fatal("expected search results, got none")
	}
	if got := results[0].Chunk.Path; got != "tools/rag/internal/app/app.go" {
		t.Fatalf("expected explicit path match to rank first, got %s", got)
	}
}

func TestBuildWithDocSourceIndexesImageDerivedChunks(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "practice", "arch.md"), "# DTM\n\n![arch](../imgs/arch.svg)\n")
	writeTestFile(t, filepath.Join(root, "docs", "imgs", "arch.svg"), `<svg><text>TM RM AP Saga</text></svg>`)

	result, err := BuildWithRoots(context.Background(), []SourceRoot{{
		Path:       root,
		Prefix:     "dtm.pub",
		SourceName: "dtm.pub",
		DocSource: &DocSource{
			Name:           "dtm",
			Topic:          TopicByName("dtm"),
			EnableImageOCR: true,
		},
	}})
	if err != nil {
		t.Fatalf("BuildWithRoots() error = %v", err)
	}
	got := indexedPaths(result.Chunks)
	if _, ok := got["dtm.pub/docs/practice/arch.md"]; !ok {
		t.Fatalf("expected markdown path to be indexed, got paths=%v", mapsKeys(got))
	}
	if _, ok := got["dtm.pub/docs/imgs/arch.svg#image"]; !ok {
		t.Fatalf("expected derived image path to be indexed, got paths=%v", mapsKeys(got))
	}
}

func TestSearchPrefersDTMDocSourceOverClaude(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "CLAUDE.md"), "DTM 是个分布式事务框架\n")
	docRoot := t.TempDir()
	writeTestFile(t, filepath.Join(docRoot, "docs", "practice", "arch.md"), "DTM 架构里 TM RM AP 用于 Saga 和 TCC\n")

	repoResult, err := Build(context.Background(), root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	docResult, err := BuildWithRoots(context.Background(), []SourceRoot{{
		Path:       docRoot,
		Prefix:     "dtm.pub",
		SourceName: "dtm.pub",
		DocSource: &DocSource{
			Name:           "dtm",
			Topic:          TopicByName("dtm"),
			EnableImageOCR: true,
		},
	}})
	if err != nil {
		t.Fatalf("BuildWithRoots() error = %v", err)
	}
	chunks := append(repoResult.Chunks, docResult.Chunks...)
	results := Search(chunks, "DTM 的 Saga/TCC 工作流怎么用？", 3)
	if len(results) == 0 {
		t.Fatal("expected search results, got none")
	}
	if !strings.HasPrefix(results[0].Chunk.Path, "dtm.pub/") {
		t.Fatalf("expected dtm.pub result first, got %s", results[0].Chunk.Path)
	}
}

func indexedPaths(chunks []Chunk) map[string]struct{} {
	paths := make(map[string]struct{}, len(chunks))
	for _, chunk := range chunks {
		paths[chunk.Path] = struct{}{}
	}
	return paths
}

func mapsKeys(in map[string]struct{}) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	return keys
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
