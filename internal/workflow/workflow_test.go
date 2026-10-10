package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCheckTransition_DecisionGates covers skills/harness-workflow/SKILL.md's
// "Decision Gates" table row by row, full track. Any failure here means this
// Go port has diverged from that table.
func TestCheckTransition_DecisionGates(t *testing.T) {
	tests := []struct {
		name         string
		current      string
		status       string
		requested    string
		wantAllowed  bool
		wantResuming bool
		wantMissing  []string
	}{
		{"same phase, in_progress -> allowed resuming", "design", "in_progress", "design", true, true, nil},
		{"same phase, completed -> blocked already completed", "design", "completed", "design", false, false, nil},
		{"next sequential phase, completed -> allowed advancing", "design", "completed", "tasks", true, false, nil},
		{"next sequential phase, in_progress -> blocked, current must complete first", "design", "in_progress", "tasks", false, false, []string{"design"}},
		{">1 step ahead -> blocked, missing phases listed", "design", "completed", "apply", false, false, []string{"tasks"}},
		{"far ahead -> blocked, all missing phases listed", "explore", "completed", "verify", false, false, []string{"propose", "spec", "design", "tasks", "apply"}},
		{"before current -> blocked backward", "tasks", "completed", "design", false, false, nil},
		{"requested == current - 1 (one back) -> blocked backward", "tasks", "in_progress", "design", false, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CheckTransition("full", tt.current, tt.status, tt.requested)
			if err != nil {
				t.Fatalf("CheckTransition() error = %v", err)
			}
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v (reason: %q)", result.Allowed, tt.wantAllowed, result.Reason)
			}
			if result.Resuming != tt.wantResuming {
				t.Errorf("Resuming = %v, want %v", result.Resuming, tt.wantResuming)
			}
			if tt.wantMissing != nil && !equalStrings(result.MissingPhases, tt.wantMissing) {
				t.Errorf("MissingPhases = %v, want %v", result.MissingPhases, tt.wantMissing)
			}
			if !tt.wantAllowed && result.Reason == "" {
				t.Errorf("blocked transition has no Reason")
			}
		})
	}
}

func TestCheckTransition_UnknownTrack(t *testing.T) {
	_, err := CheckTransition("nonsense", "explore", "completed", "propose")
	if err == nil {
		t.Fatal("CheckTransition() with unknown track should error, got nil")
	}
}

func TestCheckTransition_PhaseNotInBugfixTrack(t *testing.T) {
	// "propose" is not part of the bugfix track (explore, spec, apply, verify, archive).
	result, err := CheckTransition("bugfix", "explore", "completed", "propose")
	if err != nil {
		t.Fatalf("CheckTransition() error = %v", err)
	}
	if result.Allowed {
		t.Fatal("propose should not be allowed in the bugfix track")
	}
	if result.Reason != "propose is not part of the bugfix track" {
		t.Errorf("Reason = %q, want the exact harness-workflow message", result.Reason)
	}
}

func TestCheckTransition_BugfixTrackHappyPath(t *testing.T) {
	// explore -> spec -> apply -> verify -> archive, skipping propose/design/tasks/judge.
	steps := []string{"spec", "apply", "verify", "archive"}
	current := "explore"
	for _, next := range steps {
		result, err := CheckTransition("bugfix", current, "completed", next)
		if err != nil {
			t.Fatalf("CheckTransition(%q -> %q) error = %v", current, next, err)
		}
		if !result.Allowed {
			t.Fatalf("CheckTransition(%q -> %q) blocked: %s", current, next, result.Reason)
		}
		current = next
	}
}

func TestCheckTransition_EmptyTrackDefaultsToFull(t *testing.T) {
	// Back-compat: track absent ("") behaves as "full".
	result, err := CheckTransition("", "explore", "completed", "propose")
	if err != nil {
		t.Fatalf("CheckTransition() error = %v", err)
	}
	if !result.Allowed {
		t.Fatalf("empty track should resolve to full and allow explore->propose, got blocked: %s", result.Reason)
	}
}

func TestState_ReadWriteRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	original := &State{
		Phase:  "design",
		Status: "completed",
		Track:  "full",
		History: []HistoryEntry{
			{Phase: "explore", Status: "completed", TS: "2026-06-10T10:00:00Z"},
			{Phase: "design", Status: "completed", TS: "2026-06-10T10:30:00Z"},
		},
	}
	if err := original.Write(path); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := ReadState(path)
	if err != nil {
		t.Fatalf("ReadState() error = %v", err)
	}
	if got.Phase != original.Phase || got.Status != original.Status || got.Track != original.Track {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", got, original)
	}
	if len(got.History) != 2 {
		t.Fatalf("History length = %d, want 2", len(got.History))
	}
}

func TestReadState_MissingFile(t *testing.T) {
	_, err := ReadState(filepath.Join(t.TempDir(), "state.yaml"))
	if !errors.Is(err, ErrStateNotFound) {
		t.Errorf("error = %v, want ErrStateNotFound", err)
	}
}

func TestReadState_CorruptYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(path, []byte("phase: [this is not valid: yaml"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	_, err := ReadState(path)
	if err == nil {
		t.Fatal("ReadState() on corrupt YAML should error")
	}
	if errors.Is(err, ErrStateNotFound) {
		t.Error("corrupt YAML should not be reported as ErrStateNotFound")
	}
}

func TestState_Advance(t *testing.T) {
	s := &State{Phase: "spec", Status: "completed", Track: "full"}
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	result, err := s.Advance("design", now)
	if err != nil {
		t.Fatalf("Advance() error = %v", err)
	}
	if !result.Allowed {
		t.Fatalf("Advance() blocked: %s", result.Reason)
	}
	if s.Phase != "design" || s.Status != "in_progress" {
		t.Errorf("state after Advance = %+v, want phase=design status=in_progress", s)
	}
	last := s.History[len(s.History)-1]
	if last.Phase != "design" || last.Status != "in_progress" || last.TS != "2026-10-10T12:00:00Z" {
		t.Errorf("last history entry = %+v", last)
	}
}

func TestState_Advance_BlockedDoesNotMutate(t *testing.T) {
	s := &State{Phase: "spec", Status: "completed", Track: "full"}
	before := *s

	result, err := s.Advance("apply", time.Now()) // skips design, tasks
	if err != nil {
		t.Fatalf("Advance() error = %v", err)
	}
	if result.Allowed {
		t.Fatal("Advance() to a far-ahead phase should be blocked")
	}
	if s.Phase != before.Phase || s.Status != before.Status || len(s.History) != len(before.History) {
		t.Errorf("state mutated on a blocked Advance(): got %+v, want unchanged %+v", s, before)
	}
}

func TestState_Complete(t *testing.T) {
	s := &State{Phase: "design", Status: "in_progress"}
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	s.Complete(now)

	if s.Status != "completed" {
		t.Errorf("Status = %q, want completed", s.Status)
	}
	last := s.History[len(s.History)-1]
	if last.Phase != "design" || last.Status != "completed" {
		t.Errorf("last history entry = %+v", last)
	}
}

func TestState_RetryLifecycle(t *testing.T) {
	s := &State{Phase: "judge", Status: "in_progress"}

	for i := 1; i <= MaxJudgeRetries; i++ {
		got, err := s.IncrementRetry()
		if err != nil {
			t.Fatalf("IncrementRetry() #%d error = %v", i, err)
		}
		if got != i {
			t.Errorf("IncrementRetry() #%d = %d, want %d", i, got, i)
		}
	}

	// The 4th increment must fail without changing RetryCount further.
	_, err := s.IncrementRetry()
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("4th IncrementRetry() error = %v, want ErrMaxRetriesExceeded", err)
	}
	if s.RetryCount != MaxJudgeRetries {
		t.Errorf("RetryCount = %d after exceeding cap, want it to stay at %d", s.RetryCount, MaxJudgeRetries)
	}

	s.ResetRetry()
	if s.RetryCount != 0 {
		t.Errorf("RetryCount after ResetRetry() = %d, want 0", s.RetryCount)
	}
}

func TestState_PhaseHistoryMarkdown(t *testing.T) {
	s := &State{
		Phase:  "spec",
		Status: "in_progress",
		Track:  "full",
		History: []HistoryEntry{
			{Phase: "explore", Status: "completed", TS: "2026-06-10T10:00:00Z"},
			{Phase: "propose", Status: "completed", TS: "2026-06-10T10:05:00Z"},
			{Phase: "spec", Status: "in_progress", TS: "2026-06-10T10:10:00Z"},
		},
	}

	md, err := s.PhaseHistoryMarkdown()
	if err != nil {
		t.Fatalf("PhaseHistoryMarkdown() error = %v", err)
	}

	want := "## Phase History\n" +
		"- [x] explore — completed 2026-06-10T10:00:00Z\n" +
		"- [x] propose — completed 2026-06-10T10:05:00Z\n" +
		"- [ ] spec — in_progress 2026-06-10T10:10:00Z\n" +
		"- [ ] design\n" +
		"- [ ] tasks\n" +
		"- [ ] apply\n" +
		"- [ ] verify\n" +
		"- [ ] judge\n" +
		"- [ ] archive\n"

	if md != want {
		t.Errorf("PhaseHistoryMarkdown() =\n%s\nwant\n%s", md, want)
	}
}

func TestState_StatusLine(t *testing.T) {
	s := &State{Phase: "apply", Status: "in_progress"}
	if got, want := s.StatusLine(), "**Current phase**: apply (in_progress)"; got != want {
		t.Errorf("StatusLine() = %q, want %q", got, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
