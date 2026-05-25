package loop

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type ActionType string

const (
	ActionReadFile  ActionType = "read_file"
	ActionWriteFile ActionType = "write_file"
	ActionDelete    ActionType = "delete_file"
	ActionRun       ActionType = "run"
)

type Action struct {
	Type      ActionType `json:"type"`
	Path      string     `json:"path,omitempty"`
	StartLine int        `json:"start_line,omitempty"`
	EndLine   int        `json:"end_line,omitempty"`
	Content   string     `json:"content,omitempty"`
	Command   string     `json:"command,omitempty"`
}

type TurnPlan struct {
	Summary       string   `json:"summary"`
	Done          bool     `json:"done"`
	CommitMessage string   `json:"commit_message"`
	Actions       []Action `json:"actions"`
}

type ActionResult struct {
	Action  Action
	Success bool
	Output  string
}

func ParsePlan(raw string) (TurnPlan, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return TurnPlan{}, fmt.Errorf("no JSON object found in model response")
	}
	var plan TurnPlan
	if err := json.Unmarshal([]byte(raw[start:end+1]), &plan); err != nil {
		return TurnPlan{}, fmt.Errorf("decode plan JSON: %w", err)
	}
	return plan, nil
}

func ValidatePlan(plan TurnPlan, allowedCommands map[string]struct{}) error {
	if len(plan.Actions) == 0 && !plan.Done {
		return fmt.Errorf("plan must include actions unless done=true")
	}
	if len(plan.Actions) > 8 {
		return fmt.Errorf("plan contains too many actions: %d", len(plan.Actions))
	}
	for _, action := range plan.Actions {
		if err := ValidateAction(action, allowedCommands); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAction(action Action, allowedCommands map[string]struct{}) error {
	switch action.Type {
	case ActionReadFile:
		if err := validatePath(action.Path); err != nil {
			return fmt.Errorf("read_file path: %w", err)
		}
		if action.StartLine < 0 || action.EndLine < 0 || (action.EndLine > 0 && action.EndLine < action.StartLine) {
			return fmt.Errorf("read_file line range is invalid")
		}
	case ActionWriteFile:
		if err := validatePath(action.Path); err != nil {
			return fmt.Errorf("write_file path: %w", err)
		}
		if action.Content == "" {
			return fmt.Errorf("write_file content is empty")
		}
	case ActionDelete:
		if err := validatePath(action.Path); err != nil {
			return fmt.Errorf("delete_file path: %w", err)
		}
	case ActionRun:
		command := strings.TrimSpace(action.Command)
		if command == "" {
			return fmt.Errorf("run command is empty")
		}
		if _, ok := allowedCommands[command]; !ok {
			return fmt.Errorf("run command is not allowed: %s", command)
		}
	default:
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}
	return nil
}

func validatePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("path must be relative")
	}
	clean := filepath.Clean(path)
	if clean == "." || strings.HasPrefix(clean, "..") || strings.Contains(filepath.ToSlash(clean), "../") {
		return fmt.Errorf("path escapes repository")
	}
	return nil
}

func Done(plan TurnPlan, iteration, maxIterations int, hadFailure bool) bool {
	if iteration >= maxIterations {
		return true
	}
	if plan.Done && !hadFailure {
		return true
	}
	return false
}
