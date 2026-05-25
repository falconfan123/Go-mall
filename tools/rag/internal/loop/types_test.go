package loop

import "testing"

func TestValidateAction(t *testing.T) {
	allowed := map[string]struct{}{"make test-unit": {}}
	tests := []struct {
		name   string
		action Action
		ok     bool
	}{
		{name: "read ok", action: Action{Type: ActionReadFile, Path: "docs/a.md", StartLine: 1, EndLine: 2}, ok: true},
		{name: "write ok", action: Action{Type: ActionWriteFile, Path: "services/a.go", Content: "package a"}, ok: true},
		{name: "delete ok", action: Action{Type: ActionDelete, Path: "services/a.go"}, ok: true},
		{name: "run ok", action: Action{Type: ActionRun, Command: "make test-unit"}, ok: true},
		{name: "path escapes", action: Action{Type: ActionReadFile, Path: "../secret"}, ok: false},
		{name: "command denied", action: Action{Type: ActionRun, Command: "go test ./..."}, ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAction(tc.action, allowed)
			if tc.ok && err != nil {
				t.Fatalf("ValidateAction() unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("ValidateAction() expected error")
			}
		})
	}
}

func TestDone(t *testing.T) {
	if !Done(TurnPlan{Done: true}, 1, 3, false) {
		t.Fatalf("expected done when model signals done and no failure")
	}
	if Done(TurnPlan{Done: true}, 1, 3, true) {
		t.Fatalf("should not finish on failure before max iterations")
	}
	if !Done(TurnPlan{}, 3, 3, true) {
		t.Fatalf("expected termination at max iterations")
	}
}
