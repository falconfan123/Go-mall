package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/tools/rag/internal/anthropic"
	"github.com/falconfan123/Go-mall/tools/rag/internal/deepseek"
	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
	"github.com/falconfan123/Go-mall/tools/rag/internal/loop"
	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
	"github.com/falconfan123/Go-mall/tools/rag/internal/prompt"
	"github.com/falconfan123/Go-mall/tools/rag/internal/storage"
)

var defaultAllowedCommands = []string{"make lint", "make test-unit", "make mock"}
var extraIndexRoots = []string{
	"/Volumes/Fan/gorder",
	"/Volumes/Fan/dtm.hub",
	"/Volumes/Fan/dtm.pub",
}

type CLI struct {
	RepoRoot string
	DataDir  string
	DBPath   string
	Stdout   *os.File
	Stderr   *os.File
}

type AskOptions struct {
	Query      string
	TopK       int
	Model      model.Client
	Refresh    bool
	MaxTokens  int
	PrintCites bool
}

type LoopOptions struct {
	Prompt          string
	Model           model.Client
	AllowedCommands []string
	MaxIterations   int
	DryRun          bool
	TopK            int
	AutoCommit      bool
}

type DoctorReport struct {
	RepoRoot       string
	DBPath         string
	Writable       bool
	GitReady       bool
	AuthTokenReady bool
	APIKeyReady    bool
	AuthReady      bool
	IndexedSources map[string]int
	AllowedChecks  map[string]error
}

func NewCLI(repoRoot string) (*CLI, error) {
	if repoRoot == "" {
		root, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		repoRoot = root
	}
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	dataDir := filepath.Join(repoRoot, ".artifacts", "rag")
	return &CLI{
		RepoRoot: repoRoot,
		DataDir:  dataDir,
		DBPath:   filepath.Join(dataDir, "rag.db"),
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
	}, nil
}

func (c *CLI) OpenStore() (*storage.Store, error) {
	return storage.Open(c.DBPath)
}

func (c *CLI) Index(ctx context.Context) (indexer.BuildResult, error) {
	store, err := c.OpenStore()
	if err != nil {
		return indexer.BuildResult{}, err
	}
	defer store.Close()

	result, err := indexer.BuildWithRoots(ctx, c.indexRoots())
	if err != nil {
		return indexer.BuildResult{}, err
	}
	if err := store.ReplaceChunks(ctx, result.Chunks); err != nil {
		return indexer.BuildResult{}, err
	}
	return result, nil
}

func (c *CLI) Ask(ctx context.Context, options AskOptions) (string, []indexer.SearchResult, error) {
	store, err := c.OpenStore()
	if err != nil {
		return "", nil, err
	}
	defer store.Close()
	if options.TopK <= 0 {
		options.TopK = 6
	}
	if options.MaxTokens <= 0 {
		options.MaxTokens = 2048
	}
	if options.Refresh {
		if _, err := c.Index(ctx); err != nil {
			return "", nil, err
		}
	}
	chunks, err := store.LoadChunks(ctx)
	if err != nil {
		return "", nil, err
	}
	if len(chunks) == 0 {
		if _, err := c.Index(ctx); err != nil {
			return "", nil, err
		}
		chunks, err = store.LoadChunks(ctx)
		if err != nil {
			return "", nil, err
		}
	}
	results := indexer.Search(chunks, options.Query, options.TopK)
	if options.Model == nil {
		var sb strings.Builder
		sb.WriteString("No model configured. Matching repository context:\n")
		for _, result := range results {
			sb.WriteString(result.Citation)
			sb.WriteByte('\n')
			sb.WriteString(result.Chunk.Content)
			sb.WriteString("\n\n")
		}
		return strings.TrimSpace(sb.String()), results, nil
	}
	resp, err := options.Model.Generate(ctx, model.GenerateRequest{
		System:      "You answer questions using repository context and always cite files with line ranges.",
		Messages:    []model.Message{{Role: "user", Content: prompt.BuildAskPrompt(options.Query, results)}},
		MaxTokens:   options.MaxTokens,
		Temperature: 0.1,
		Metadata:    map[string]string{"mode": "ask"},
	})
	if err != nil {
		return "", results, err
	}
	return strings.TrimSpace(resp.Text), results, nil
}

func (c *CLI) Loop(ctx context.Context, options LoopOptions) (loop.Result, error) {
	store, err := c.OpenStore()
	if err != nil {
		return loop.Result{}, err
	}
	defer store.Close()
	if options.MaxIterations <= 0 {
		options.MaxIterations = 6
	}
	if options.TopK <= 0 {
		options.TopK = 6
	}
	allowed := storage.NormalizeCommands(append(defaultAllowedCommands, options.AllowedCommands...))
	if len(allowed) == 0 {
		return loop.Result{}, fmt.Errorf("no allowed commands configured")
	}
	if _, err := ensureIndex(ctx, c, store); err != nil {
		return loop.Result{}, err
	}
	chunks, err := store.LoadChunks(ctx)
	if err != nil {
		return loop.Result{}, err
	}
	id := sessionID()
	worktreePath := filepath.Join(c.DataDir, "worktrees", id)
	branch := "rag/" + filepath.Base(worktreePath)
	if !options.DryRun {
		if err := prepareWorktree(ctx, c.RepoRoot, worktreePath, branch); err != nil {
			return loop.Result{}, err
		}
	} else {
		worktreePath = c.RepoRoot
	}
	session := storage.Session{
		ID:              id,
		Prompt:          options.Prompt,
		Status:          "running",
		RepoRoot:        c.RepoRoot,
		WorktreePath:    worktreePath,
		Branch:          branch,
		AllowedCommands: allowed,
		MaxIterations:   options.MaxIterations,
		DryRun:          options.DryRun,
		CommitMessage:   "chore: apply rag loop changes",
	}
	if options.DryRun && options.Model == nil {
		session.Status = "dry-run"
		session.Summary = "dry-run preview only; no model configured"
		if err := store.SaveSession(ctx, session); err != nil {
			return loop.Result{}, err
		}
		return loop.Result{
			Session: session,
			Plan: loop.TurnPlan{
				Summary: "dry-run preview only; no model configured",
				Done:    false,
			},
		}, nil
	}
	engine := loop.Engine{
		Model:           options.Model,
		Runner:          loop.ShellRunner{},
		Store:           store,
		RepoRoot:        c.RepoRoot,
		Session:         session,
		Chunks:          chunks,
		AutoCommit:      options.AutoCommit,
		ContextResults:  options.TopK,
		FinalValidation: []string{"make lint", "make test-unit"},
	}
	return engine.Run(ctx)
}

func (c *CLI) Resume(ctx context.Context, id string, m model.Client) (loop.Result, error) {
	store, err := c.OpenStore()
	if err != nil {
		return loop.Result{}, err
	}
	defer store.Close()
	session, err := store.GetSession(ctx, id)
	if err != nil {
		return loop.Result{}, err
	}
	if session.Status == "completed" {
		return loop.Result{Session: session}, nil
	}
	if _, err := ensureIndex(ctx, c, store); err != nil {
		return loop.Result{}, err
	}
	chunks, err := store.LoadChunks(ctx)
	if err != nil {
		return loop.Result{}, err
	}
	engine := loop.Engine{
		Model:           m,
		Runner:          loop.ShellRunner{},
		Store:           store,
		RepoRoot:        c.RepoRoot,
		Session:         session,
		Chunks:          chunks,
		AutoCommit:      !session.DryRun,
		ContextResults:  6,
		FinalValidation: []string{"make lint", "make test-unit"},
	}
	return engine.Run(ctx)
}

func (c *CLI) Doctor(ctx context.Context) (DoctorReport, error) {
	report := DoctorReport{
		RepoRoot:       c.RepoRoot,
		DBPath:         c.DBPath,
		IndexedSources: map[string]int{},
		AllowedChecks:  map[string]error{},
	}
	if err := os.MkdirAll(c.DataDir, 0o755); err != nil {
		return report, err
	}
	f, err := os.CreateTemp(c.DataDir, "doctor-*")
	if err == nil {
		report.Writable = true
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
	if _, err := exec.LookPath("git"); err == nil {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
		cmd.Dir = c.RepoRoot
		if output, err := cmd.Output(); err == nil && strings.TrimSpace(string(output)) != "" {
			report.GitReady = true
		}
	}
	report.AuthTokenReady = strings.TrimSpace(os.Getenv("ANTHROPIC_AUTH_TOKEN")) != ""
	report.APIKeyReady = strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) != ""
	report.AuthReady = report.AuthTokenReady || report.APIKeyReady
	if store, err := c.OpenStore(); err == nil {
		if counts, err := store.CountChunksBySource(ctx); err == nil {
			report.IndexedSources = counts
		}
		_ = store.Close()
	}
	for _, target := range defaultAllowedCommands[:2] {
		report.AllowedChecks[target] = checkMakeTarget(ctx, c.RepoRoot, target)
	}
	report.AllowedChecks["make mock"] = checkMakeTarget(ctx, c.RepoRoot, "make mock")
	return report, nil
}

func ensureIndex(ctx context.Context, c *CLI, store *storage.Store) (int, error) {
	count, err := store.ChunkCount(ctx)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return count, nil
	}
	result, err := c.Index(ctx)
	if err != nil {
		return 0, err
	}
	return len(result.Chunks), nil
}

func checkMakeTarget(ctx context.Context, repoRoot, command string) error {
	args := strings.Fields(command)
	if len(args) < 2 {
		return fmt.Errorf("invalid make command")
	}
	cmd := exec.CommandContext(ctx, args[0], "-n", args[1])
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func prepareWorktree(ctx context.Context, repoRoot, worktreePath, branch string) error {
	if _, err := os.Stat(worktreePath); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "--detach", worktreePath, "HEAD")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(output)))
	}
	checkout := exec.CommandContext(ctx, "git", "checkout", "-b", branch)
	checkout.Dir = worktreePath
	if output, err := checkout.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func sessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UTC().UnixNano())
}

func NewAnthropicClient() (model.Client, error) {
	return anthropic.NewFromEnv()
}

func NewDeepSeekClient() (model.Client, error) {
	return deepseek.NewFromEnv()
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (c *CLI) indexRoots() []indexer.SourceRoot {
	roots := []indexer.SourceRoot{{Path: c.RepoRoot}}
	for _, externalRoot := range extraIndexRoots {
		externalRoot = strings.TrimSpace(externalRoot)
		if externalRoot == "" {
			continue
		}
		info, err := os.Stat(externalRoot)
		if err != nil || !info.IsDir() {
			continue
		}
		docSource := (*indexer.DocSource)(nil)
		if filepath.Base(externalRoot) == "dtm.pub" {
			docSource = &indexer.DocSource{
				Name:           "dtm",
				Topic:          indexer.TopicByName("dtm"),
				EnableImageOCR: true,
			}
		}
		roots = append(roots, indexer.SourceRoot{
			Path:       externalRoot,
			Prefix:     filepath.Base(externalRoot),
			SourceName: filepath.Base(externalRoot),
			DocSource:  docSource,
		})
	}
	return roots
}
