package indexer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	chunkLines    = 80
	chunkOverlap  = 20
	maxFileSize   = 500 << 10
	maxLineLength = 4000
)

var ignoredDirs = map[string]struct{}{
	".artifacts":   {},
	".git":         {},
	".idea":        {},
	".vscode":      {},
	"__pycache__":  {},
	"build":        {},
	"dist":         {},
	"node_modules": {},
	"vendor":       {},
}

var ignoredFiles = map[string]struct{}{
	"go.sum":            {},
	"go.work.sum":       {},
	"package-lock.json": {},
	"pnpm-lock.yaml":    {},
	"yarn.lock":         {},
}

var ignoredExtensions = map[string]struct{}{
	".bin": {},
	".exe": {},
}

type Chunk struct {
	ID         int64
	Path       string
	LineStart  int
	LineEnd    int
	Content    string
	Priority   float64
	Hash       string
	SourceKind string
	SourceName string
}

type BuildResult struct {
	Chunks       []Chunk
	FilesIndexed int
	FilesSkipped int
}

type SourceRoot struct {
	Path       string
	Prefix     string
	SourceName string
	DocSource  *DocSource
}

type SearchResult struct {
	Chunk    Chunk
	Score    float64
	Citation string
}

type TopicRule struct {
	Name             string
	Keywords         []string
	PathPrefixes     []string
	Boost            float64
	SoftPenaltyPaths []string
	PathHints        []PathHint
	PathBiases       []PathBias
}

type PathHint struct {
	Keywords       []string
	PathSubstrings []string
	Boost          float64
}

type PathBias struct {
	PathSubstring string
	Delta         float64
}

type DocSource struct {
	Name           string
	Topic          *TopicRule
	EnableImageOCR bool
}

var topicRules = []TopicRule{
	{
		Name: "dtm",
		Keywords: []string{
			"dtm", "saga", "tcc", "xa", "barrier", "2pc", "msg", "workflow",
			"分布式事务", "事务消息", "补偿", "事务屏障", "tm", "rm", "ap",
		},
		PathPrefixes: []string{
			"dtm.pub/",
		},
		Boost: 28,
		SoftPenaltyPaths: []string{
			"CLAUDE.md",
		},
		PathHints: []PathHint{
			{
				Keywords:       []string{"架构", "架构图", "tm", "rm", "ap"},
				PathSubstrings: []string{"dtm.pub/docs/practice/arch.md", "dtm.pub/docs/imgs/arch"},
				Boost:          18,
			},
			{
				Keywords:       []string{"saga"},
				PathSubstrings: []string{"dtm.pub/docs/practice/saga.md", "dtm.pub/docs/guide/e-saga.md"},
				Boost:          15,
			},
			{
				Keywords:       []string{"tcc"},
				PathSubstrings: []string{"dtm.pub/docs/practice/tcc.md", "dtm.pub/docs/guide/e-tcc.md"},
				Boost:          15,
			},
			{
				Keywords:       []string{"workflow", "工作流"},
				PathSubstrings: []string{"dtm.pub/docs/practice/workflow.md"},
				Boost:          15,
			},
			{
				Keywords:       []string{"barrier", "事务屏障"},
				PathSubstrings: []string{"dtm.pub/docs/practice/barrier.md", "dtm.pub/docs/imgs/barrier"},
				Boost:          15,
			},
			{
				Keywords:       []string{"msg", "事务消息", "二阶段消息"},
				PathSubstrings: []string{"dtm.pub/docs/practice/msg.md", "dtm.pub/docs/guide/e-msg.md"},
				Boost:          15,
			},
			{
				Keywords:       []string{"xa", "2pc"},
				PathSubstrings: []string{"dtm.pub/docs/practice/xa.md", "dtm.pub/docs/imgs/xa-"},
				Boost:          15,
			},
		},
		PathBiases: []PathBias{
			{PathSubstring: "dtm.pub/docs/practice/", Delta: 28},
			{PathSubstring: "dtm.pub/docs/guide/", Delta: 16},
			{PathSubstring: "dtm.pub/docs/imgs/", Delta: 8},
			{PathSubstring: "dtm.pub/docs/app/", Delta: -10},
			{PathSubstring: "dtm.pub/docs/ref/", Delta: -18},
			{PathSubstring: "dtm.pub/docs/deploy/", Delta: -12},
		},
	},
}

func TopicByName(name string) *TopicRule {
	for _, rule := range topicRules {
		if rule.Name == name {
			copyRule := rule
			return &copyRule
		}
	}
	return nil
}

var (
	markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	htmlImagePattern     = regexp.MustCompile(`(?i)<img[^>]*src=["']([^"']+)["'][^>]*>`)
	htmlAltPattern       = regexp.MustCompile(`(?i)alt=["']([^"']*)["']`)
)

func Build(ctx context.Context, repoRoot string) (BuildResult, error) {
	return BuildWithRoots(ctx, []SourceRoot{{Path: repoRoot}})
}

func BuildWithRoots(ctx context.Context, roots []SourceRoot) (BuildResult, error) {
	var result BuildResult
	for _, root := range roots {
		if strings.TrimSpace(root.Path) == "" {
			continue
		}
		rootResult, err := buildRoot(ctx, root)
		if err != nil {
			return BuildResult{}, err
		}
		result.FilesIndexed += rootResult.FilesIndexed
		result.FilesSkipped += rootResult.FilesSkipped
		result.Chunks = append(result.Chunks, rootResult.Chunks...)
	}
	sort.Slice(result.Chunks, func(i, j int) bool {
		if result.Chunks[i].Path == result.Chunks[j].Path {
			return result.Chunks[i].LineStart < result.Chunks[j].LineStart
		}
		return result.Chunks[i].Path < result.Chunks[j].Path
	})
	return result, nil
}

func buildRoot(ctx context.Context, root SourceRoot) (BuildResult, error) {
	var result BuildResult
	err := filepath.WalkDir(root.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name(), path, root.Path) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root.Path, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if root.Prefix != "" {
			rel = pathWithPrefix(root.Prefix, rel)
		}
		if !shouldIndexFile(rel) {
			result.FilesSkipped++
			return nil
		}
		chunks, err := chunkFile(path, rel, root)
		if err != nil {
			result.FilesSkipped++
			return nil
		}
		if len(chunks) == 0 {
			result.FilesSkipped++
			return nil
		}
		result.FilesIndexed++
		result.Chunks = append(result.Chunks, chunks...)
		return nil
	})
	if err != nil {
		return BuildResult{}, err
	}
	return result, nil
}

func Search(chunks []Chunk, query string, limit int) []SearchResult {
	terms := normalizeTerms(query)
	pathTerms := normalizePathTerms(query)
	topics := matchedTopics(query)
	if len(terms) == 0 || limit <= 0 {
		return nil
	}
	results := make([]SearchResult, 0, len(chunks))
	for _, chunk := range chunks {
		score := scoreChunk(chunk, query, terms, pathTerms, topics)
		if score <= 0 {
			continue
		}
		results = append(results, SearchResult{
			Chunk:    chunk,
			Score:    score,
			Citation: Citation(chunk),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			if results[i].Chunk.Path == results[j].Chunk.Path {
				return results[i].Chunk.LineStart < results[j].Chunk.LineStart
			}
			return results[i].Chunk.Path < results[j].Chunk.Path
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func Citation(chunk Chunk) string {
	return fmt.Sprintf("[%s:%d-%d]", chunk.Path, chunk.LineStart, chunk.LineEnd)
}

func scoreChunk(chunk Chunk, rawQuery string, terms []string, pathTerms []string, topics []TopicRule) float64 {
	content := strings.ToLower(chunk.Content)
	path := strings.ToLower(chunk.Path)
	score := chunk.Priority*2 + fileTypeScore(chunk.Path)
	uniqueHits := 0
	if phrase := strings.ToLower(strings.TrimSpace(rawQuery)); len(phrase) >= 3 && strings.Contains(content, phrase) {
		score += 8
	}
	for _, term := range pathTerms {
		if strings.Contains(path, term) {
			score += 200
			uniqueHits++
			continue
		}
		trimmed := strings.TrimSuffix(term, filepath.Ext(term))
		if trimmed != term && trimmed != "" && strings.Contains(path, trimmed) {
			score += 40
			uniqueHits++
		}
	}
	for _, term := range terms {
		freq := strings.Count(content, term)
		pathHits := strings.Count(path, term)
		if freq == 0 && pathHits == 0 {
			continue
		}
		uniqueHits++
		score += float64(freq)*1.5 + float64(pathHits)*3
		if strings.HasPrefix(path, term) {
			score += 1
		}
	}
	for _, topic := range topics {
		for _, prefix := range topic.PathPrefixes {
			if strings.HasPrefix(path, strings.ToLower(prefix)) {
				score += topic.Boost
				break
			}
		}
		for _, penaltyPath := range topic.SoftPenaltyPaths {
			if strings.EqualFold(chunk.Path, penaltyPath) {
				score -= topic.Boost * 0.8
			}
		}
		for _, bias := range topic.PathBiases {
			if strings.Contains(path, strings.ToLower(bias.PathSubstring)) {
				score += bias.Delta
			}
		}
		for _, hint := range topic.PathHints {
			if !matchesAnyKeyword(rawQuery, hint.Keywords) {
				continue
			}
			for _, pathHint := range hint.PathSubstrings {
				if strings.Contains(path, strings.ToLower(pathHint)) {
					score += hint.Boost
					break
				}
			}
		}
	}
	score += float64(uniqueHits) * 4
	return score
}

func chunkFile(path, rel string, root SourceRoot) ([]Chunk, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() <= 0 || info.Size() > maxFileSize {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(raw, 0) >= 0 || !utf8.Valid(raw) {
		return nil, nil
	}
	if isGeneratedFile(raw) {
		return nil, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 128<<10), maxFileSize+1)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if utf8.RuneCountInString(line) > maxLineLength {
			return nil, nil
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, nil
		}
		return nil, err
	}
	if len(lines) == 0 {
		return nil, nil
	}

	priority := pathPriority(rel)
	chunks := chunkTextContent(rel, strings.Join(lines, "\n"), priority, sourceKindForRoot(root), root.SourceName)
	if root.DocSource != nil {
		imageChunks, err := imageDerivedChunks(path, rel, raw, root)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, imageChunks...)
	}
	return chunks, nil
}

func chunkTextContent(rel, text string, priority float64, sourceKind, sourceName string) []Chunk {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil
	}
	var chunks []Chunk
	for start := 0; start < len(lines); {
		end := start + chunkLines
		if end > len(lines) {
			end = len(lines)
		}
		content := strings.Join(lines[start:end], "\n")
		sum := sha256.Sum256([]byte(rel + ":" + fmt.Sprint(start) + ":" + content))
		chunks = append(chunks, Chunk{
			Path:       rel,
			LineStart:  start + 1,
			LineEnd:    end,
			Content:    content,
			Priority:   priority,
			Hash:       hex.EncodeToString(sum[:]),
			SourceKind: sourceKind,
			SourceName: sourceName,
		})
		if end == len(lines) {
			break
		}
		start = end - chunkOverlap
	}
	return chunks
}

func normalizeTerms(in string) []string {
	seen := map[string]struct{}{}
	fields := strings.FieldsFunc(strings.ToLower(in), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-')
	})
	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.Trim(field, "-_/")
		if len(field) < 2 {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		terms = append(terms, field)
	}
	return terms
}

func normalizePathTerms(in string) []string {
	seen := map[string]struct{}{}
	fields := strings.FieldsFunc(strings.ToLower(in), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '/' || r == '.')
	})
	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.Trim(field, "-_./")
		if len(field) < 3 {
			continue
		}
		if !strings.Contains(field, "/") && !strings.Contains(field, ".") {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		terms = append(terms, field)
	}
	return terms
}

func shouldSkipDir(name, fullPath, repoRoot string) bool {
	rel, err := filepath.Rel(repoRoot, fullPath)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, ".git") || strings.HasPrefix(rel, ".artifacts") {
		return true
	}
	_, ok := ignoredDirs[name]
	return ok
}

func shouldIndexFile(rel string) bool {
	base := filepath.Base(rel)
	if strings.HasPrefix(base, ".") && base != ".gitignore" {
		return false
	}
	if _, ok := ignoredFiles[strings.ToLower(base)]; ok {
		return false
	}
	ext := strings.ToLower(filepath.Ext(rel))
	if _, ok := ignoredExtensions[ext]; ok {
		return false
	}
	switch ext {
	case ".go", ".md", ".txt", ".yaml", ".yml", ".json", ".sh", ".sql", ".proto", ".mod", ".work", ".xml", ".html":
		return true
	case "":
		return base == "Makefile" || base == "Dockerfile" || base == "CLAUDE.md"
	default:
		return false
	}
}

func pathPriority(rel string) float64 {
	switch {
	case rel == "CLAUDE.md":
		return 6
	case strings.Contains(rel, "/docs/practice/"):
		return 6
	case strings.Contains(rel, "/docs/guide/"):
		return 5
	case strings.Contains(rel, "/docs/app/"):
		return 4
	case strings.Contains(rel, "/docs/ref/"):
		return 3
	case strings.Contains(rel, "/docs/imgs/"):
		return 2
	case strings.HasPrefix(rel, "docs/"):
		return 5
	case strings.HasPrefix(rel, "services/"):
		return 4
	case strings.HasPrefix(rel, "test/"):
		return 4
	case strings.HasPrefix(rel, "scripts/"):
		return 3
	default:
		return 1
	}
}

func fileTypeScore(rel string) float64 {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".go", ".proto", ".md", ".sql", ".sh":
		return 2
	case ".mod", ".json", ".yaml", ".yml", ".xml", ".html":
		return 0.5
	default:
		return 0
	}
}

func isGeneratedFile(raw []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for i := 0; i < 5 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "Code generated") || strings.Contains(line, "DO NOT EDIT") {
			return true
		}
	}
	return false
}

func matchedTopics(query string) []TopicRule {
	var matched []TopicRule
	for _, rule := range topicRules {
		if matchesAnyKeyword(query, rule.Keywords) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func matchesAnyKeyword(query string, keywords []string) bool {
	lower := strings.ToLower(query)
	for _, keyword := range keywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func sourceKindForRoot(root SourceRoot) string {
	if root.DocSource != nil {
		return "external_doc"
	}
	if root.Prefix != "" {
		return "external_repo"
	}
	return "repo"
}

func imageDerivedChunks(path, rel string, raw []byte, root SourceRoot) ([]Chunk, error) {
	if root.DocSource == nil || !root.DocSource.EnableImageOCR {
		return nil, nil
	}
	refs := parseImageRefs(path, raw)
	if len(refs) == 0 {
		return nil, nil
	}
	var chunks []Chunk
	for _, ref := range refs {
		imageText, err := buildImageText(ref, root)
		if err != nil || strings.TrimSpace(imageText) == "" {
			continue
		}
		derivedPath := rootRelativePath(root, ref.AbsPath)
		if derivedPath == "" {
			continue
		}
		derivedPath += "#image"
		for _, chunk := range chunkTextContent(derivedPath, imageText, pathPriority(derivedPath)+1, "image_ocr", root.SourceName) {
			chunks = append(chunks, chunk)
		}
	}
	return chunks, nil
}

type imageRef struct {
	AbsPath     string
	RelPath     string
	Alt         string
	ContextPath string
	ContextText string
}

func parseImageRefs(docPath string, raw []byte) []imageRef {
	text := string(raw)
	var refs []imageRef
	baseDir := filepath.Dir(docPath)
	for _, match := range markdownImagePattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 3 {
			continue
		}
		if ref, ok := buildImageRef(baseDir, docPath, match[2], match[1], text); ok {
			refs = append(refs, ref)
		}
	}
	for _, match := range htmlImagePattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		fullTag := match[0]
		alt := ""
		if altMatch := htmlAltPattern.FindStringSubmatch(fullTag); len(altMatch) > 1 {
			alt = altMatch[1]
		}
		if ref, ok := buildImageRef(baseDir, docPath, match[1], alt, text); ok {
			refs = append(refs, ref)
		}
	}
	return refs
}

func buildImageRef(baseDir, docPath, refPath, alt, contextText string) (imageRef, bool) {
	refPath = strings.TrimSpace(refPath)
	if refPath == "" || strings.HasPrefix(refPath, "http://") || strings.HasPrefix(refPath, "https://") || strings.HasPrefix(refPath, "data:") {
		return imageRef{}, false
	}
	refPath = strings.Split(refPath, "#")[0]
	absPath := filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(refPath)))
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		return imageRef{}, false
	}
	relPath, err := filepath.Rel(baseDir, absPath)
	if err != nil {
		relPath = filepath.Base(absPath)
	}
	return imageRef{
		AbsPath:     absPath,
		RelPath:     filepath.ToSlash(relPath),
		Alt:         strings.TrimSpace(alt),
		ContextPath: docPath,
		ContextText: contextText,
	}, true
}

func buildImageText(ref imageRef, root SourceRoot) (string, error) {
	var parts []string
	imagePath := rootRelativePath(root, ref.AbsPath)
	if imagePath == "" {
		imagePath = filepath.ToSlash(ref.RelPath)
	}
	parts = append(parts, "Referenced image: "+imagePath)
	if ref.Alt != "" {
		parts = append(parts, "Alt text: "+ref.Alt)
	}
	contextPath := rootRelativePath(root, ref.ContextPath)
	if contextPath == "" {
		contextPath = filepath.ToSlash(ref.ContextPath)
	}
	parts = append(parts, "Referenced from: "+contextPath)
	contextSummary := summarizeImageContext(ref.ContextText, ref.RelPath, ref.Alt)
	if contextSummary != "" {
		parts = append(parts, "Context:\n"+contextSummary)
	}
	ocrText, err := extractImageText(ref.AbsPath)
	if err != nil {
		return strings.Join(parts, "\n\n"), nil
	}
	if strings.TrimSpace(ocrText) != "" {
		parts = append(parts, "Extracted text:\n"+ocrText)
	}
	return strings.Join(parts, "\n\n"), nil
}

func summarizeImageContext(text, relPath, alt string) string {
	lines := strings.Split(text, "\n")
	var matches []string
	for i, line := range lines {
		if strings.Contains(line, relPath) || (alt != "" && strings.Contains(line, alt)) {
			start := i - 2
			if start < 0 {
				start = 0
			}
			end := i + 3
			if end > len(lines) {
				end = len(lines)
			}
			matches = append(matches, strings.Join(lines[start:end], "\n"))
			break
		}
	}
	return strings.TrimSpace(strings.Join(matches, "\n"))
}

func extractImageText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".svg":
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		text := string(raw)
		text = strings.ReplaceAll(text, "<![CDATA[", " ")
		text = strings.ReplaceAll(text, "]]>", " ")
		return strings.TrimSpace(text), nil
	case ".png", ".jpg", ".jpeg", ".webp":
		if _, err := exec.LookPath("tesseract"); err != nil {
			return "", err
		}
		cmd := execCommandContext(context.Background(), "tesseract", path, "stdout", "-l", "eng")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return normalizeOCRText(string(output)), nil
	default:
		return "", nil
	}
}

var execCommandContext = func(ctx context.Context, name string, args ...string) combinedOutputRunner {
	return commandRunner{ctx: ctx, name: name, args: args}
}

type combinedOutputRunner interface {
	CombinedOutput() ([]byte, error)
}

type commandRunner struct {
	ctx  context.Context
	name string
	args []string
}

func (c commandRunner) CombinedOutput() ([]byte, error) {
	return exec.CommandContext(c.ctx, c.name, c.args...).CombinedOutput()
}

func normalizeOCRText(text string) string {
	lines := strings.Split(text, "\n")
	var keep []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if utf8.RuneCountInString(line) > maxLineLength {
			line = string([]rune(line)[:maxLineLength])
		}
		keep = append(keep, line)
	}
	return strings.Join(keep, "\n")
}

func rootRelativePath(root SourceRoot, absPath string) string {
	rel, err := filepath.Rel(root.Path, absPath)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "../") || rel == ".." {
		return ""
	}
	return pathWithPrefix(root.Prefix, rel)
}

func pathWithPrefix(prefix, rel string) string {
	prefix = strings.Trim(strings.TrimSpace(filepath.ToSlash(prefix)), "/")
	rel = strings.Trim(strings.TrimSpace(filepath.ToSlash(rel)), "/")
	switch {
	case prefix == "":
		return rel
	case rel == "":
		return prefix
	default:
		return prefix + "/" + rel
	}
}
