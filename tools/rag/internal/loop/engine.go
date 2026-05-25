package loop

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/falconfan123/Go-mall/tools/rag/internal/indexer"
	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
	"github.com/falconfan123/Go-mall/tools/rag/internal/prompt"
	"github.com/falconfan123/Go-mall/tools/rag/internal/storage"
)

type CommandRunner interface {
	Run(ctx context.Context, dir, command string) (string, error)
}

type Engine struct {
	Model            model.Client
	Runner           CommandRunner
	Store            *storage.Store
	RepoRoot         string
	Session          storage.Session
	Chunks           []indexer.Chunk
	FinalValidation  []string
	AutoCommit       bool
	ContextResults   int
	ModelMaxTokens   int
	ModelTemperature float64
}

type Result struct {
	Session storage.Session
	Plan    TurnPlan
	Output  []ActionResult
}

type ShellRunner struct{}

func (ShellRunner) Run(ctx context.Context, dir, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-lc", command)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return strings.TrimSpace(limitString(buf.String(), 16000)), err
}

func (e *Engine) Run(ctx context.Context) (Result, error) {
	if e.Runner == nil {
		e.Runner = ShellRunner{}
	}
	if e.ContextResults <= 0 {
		e.ContextResults = 6
	}
	if e.ModelMaxTokens <= 0 {
		e.ModelMaxTokens = 2048
	}
	if len(e.FinalValidation) == 0 {
		e.FinalValidation = []string{"make lint", "make test-unit"}
	}
	if err := e.Store.SaveSession(ctx, e.Session); err != nil {
		return Result{}, err
	}

	allowed := map[string]struct{}{}
	for _, command := range e.Session.AllowedCommands {
		allowed[command] = struct{}{}
	}
	results := indexer.Search(e.Chunks, e.Session.Prompt, e.ContextResults)
	session := e.Session
	var lastOutputs []ActionResult

	for {
		session.Iteration++
		events, err := e.Store.Events(ctx, session.ID)
		if err != nil {
			return Result{}, err
		}
		var lines []string
		for _, event := range events {
			lines = append(lines, fmt.Sprintf("%s: %s", event.Role, event.Content))
		}
		promptText := prompt.BuildLoopPrompt(session.ID, session.Prompt, results, session.AllowedCommands, lines, session.Iteration, session.MaxIterations)
		resp, err := e.Model.Generate(ctx, model.GenerateRequest{
			System:    "You are operating as a repository-local coding agent. Return only valid JSON.",
			Messages:  []model.Message{{Role: "user", Content: promptText}},
			MaxTokens: e.ModelMaxTokens,
			Metadata:  map[string]string{"session_id": session.ID, "mode": "loop"},
		})
		if err != nil {
			session.Status = "failed"
			session.LastError = err.Error()
			_ = e.Store.SaveSession(ctx, session)
			return Result{}, err
		}
		_ = e.Store.AppendEvent(ctx, session.ID, "assistant", resp.Text)
		plan, err := ParsePlan(resp.Text)
		if err != nil {
			session.LastError = err.Error()
			_ = e.Store.AppendEvent(ctx, session.ID, "system", "plan parse error: "+err.Error())
			if session.Iteration >= session.MaxIterations {
				session.Status = "failed"
				_ = e.Store.SaveSession(ctx, session)
				return Result{Session: session}, err
			}
			_ = e.Store.SaveSession(ctx, session)
			continue
		}
		if err := ValidatePlan(plan, allowed); err != nil {
			session.LastError = err.Error()
			_ = e.Store.AppendEvent(ctx, session.ID, "system", "plan validation error: "+err.Error())
			if session.Iteration >= session.MaxIterations {
				session.Status = "failed"
				_ = e.Store.SaveSession(ctx, session)
				return Result{Session: session, Plan: plan}, err
			}
			_ = e.Store.SaveSession(ctx, session)
			continue
		}

		lastOutputs = nil
		hadFailure := false
		for _, action := range plan.Actions {
			output, err := e.executeAction(ctx, session, action)
			actionResult := ActionResult{Action: action, Output: output, Success: err == nil}
			if err != nil {
				hadFailure = true
				actionResult.Output = strings.TrimSpace(output + "\nERROR: " + err.Error())
				session.LastError = err.Error()
			}
			lastOutputs = append(lastOutputs, actionResult)
			_ = e.Store.AppendEvent(ctx, session.ID, "tool", formatActionResult(actionResult))
			if err != nil {
				break
			}
		}

		session.Summary = plan.Summary
		if plan.CommitMessage != "" {
			session.CommitMessage = plan.CommitMessage
		}

		if session.DryRun {
			session.Status = "dry-run"
			_ = e.Store.SaveSession(ctx, session)
			return Result{Session: session, Plan: plan, Output: lastOutputs}, nil
		}

		if Done(plan, session.Iteration, session.MaxIterations, hadFailure) {
			if plan.Done && !hadFailure {
				if err := e.runFinalValidation(ctx, session); err != nil {
					session.LastError = err.Error()
					_ = e.Store.AppendEvent(ctx, session.ID, "tool", "final validation failed: "+err.Error())
					hadFailure = true
				} else if e.AutoCommit {
					if err := e.commit(ctx, session, plan.CommitMessage); err != nil {
						session.LastError = err.Error()
						_ = e.Store.AppendEvent(ctx, session.ID, "tool", "commit failed: "+err.Error())
						hadFailure = true
					}
				}
			}
			if hadFailure {
				session.Status = "failed"
			} else {
				session.Status = "completed"
			}
			_ = e.Store.SaveSession(ctx, session)
			return Result{Session: session, Plan: plan, Output: lastOutputs}, nil
		}
		session.Status = "running"
		_ = e.Store.SaveSession(ctx, session)
	}
}

func (e *Engine) executeAction(ctx context.Context, session storage.Session, action Action) (string, error) {
	switch action.Type {
	case ActionReadFile:
		return readFile(filepath.Join(session.WorktreePath, action.Path), action.StartLine, action.EndLine)
	case ActionWriteFile:
		path := filepath.Join(session.WorktreePath, action.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(action.Content), 0o644); err != nil {
			return "", err
		}
		return "wrote " + action.Path, nil
	case ActionDelete:
		if err := os.Remove(filepath.Join(session.WorktreePath, action.Path)); err != nil {
			return "", err
		}
		return "deleted " + action.Path, nil
	case ActionRun:
		out, err := e.Runner.Run(ctx, session.WorktreePath, action.Command)
		if err != nil {
			return out, err
		}
		return out, nil
	default:
		return "", fmt.Errorf("unsupported action type: %s", action.Type)
	}
}

func (e *Engine) runFinalValidation(ctx context.Context, session storage.Session) error {
	for _, command := range e.FinalValidation {
		out, err := e.Runner.Run(ctx, session.WorktreePath, command)
		_ = e.Store.AppendEvent(ctx, session.ID, "tool", fmt.Sprintf("final validation `%s`:\n%s", command, out))
		if err != nil {
			return fmt.Errorf("%s failed: %w", command, err)
		}
	}
	return nil
}

func (e *Engine) commit(ctx context.Context, session storage.Session, commitMessage string) error {
	if strings.TrimSpace(commitMessage) == "" {
		commitMessage = "chore: apply rag loop changes"
	}
	status, err := e.Runner.Run(ctx, session.WorktreePath, "git status --short")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) == "" {
		return nil
	}
	if _, err := e.Runner.Run(ctx, session.WorktreePath, "git add -A"); err != nil {
		return err
	}
	_, err = e.Runner.Run(ctx, session.WorktreePath, "git commit -m "+shellQuote(commitMessage))
	return err
}

func readFile(path string, startLine, endLine int) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(raw), "\n")
	if startLine <= 0 {
		startLine = 1
	}
	if endLine <= 0 || endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > len(lines) {
		return "", fmt.Errorf("start line %d out of range", startLine)
	}
	return strings.Join(lines[startLine-1:endLine], "\n"), nil
}

func formatActionResult(result ActionResult) string {
	status := "ok"
	if !result.Success {
		status = "failed"
	}
	switch result.Action.Type {
	case ActionRun:
		return fmt.Sprintf("%s run `%s`\n%s", status, result.Action.Command, result.Output)
	default:
		return fmt.Sprintf("%s %s %s\n%s", status, result.Action.Type, result.Action.Path, result.Output)
	}
}

func limitString(in string, max int) string {
	if len(in) <= max {
		return in
	}
	return in[:max] + "\n...[truncated]"
}

func shellQuote(in string) string {
	return "'" + strings.ReplaceAll(in, "'", `'\''`) + "'"
}
