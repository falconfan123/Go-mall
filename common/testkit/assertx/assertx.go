package assertx

import "testing"

// RequireWithinRange checks that got is between min and max inclusively.
func RequireWithinRange(t *testing.T, got, min, max int64) {
	t.Helper()
	if got < min || got > max {
		t.Fatalf("expected %d to be within [%d, %d]", got, min, max)
	}
}
