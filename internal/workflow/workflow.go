// Package workflow is the authoritative, code-level implementation of the
// SDD phase state machine described in skills/harness-workflow/SKILL.md:
// reading and writing openspec/changes/{name}/state.yaml, validating phase
// transitions against the track's fixed sequence, and tracking the judge
// phase's retry count.
//
// Before this package existed, every rule in harness-workflow/SKILL.md's
// Decision Gates table was enforced only by an orchestrating LLM choosing to
// read and follow that prose correctly, every time, in every session — with
// no code-level check and no automated test of the actual enforcement. This
// package is the single source of truth those rules are ported from;
// skills/harness-workflow/SKILL.md and skills/harness-judge/SKILL.md should
// call `archon workflow ...` rather than reimplementing this logic inline.
package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/archon-ai/archon/internal/config"
	"gopkg.in/yaml.v3"
)

// BugfixTrack is the fixed phase sequence for track "bugfix": a scoped
// 3-section spec with no design/tasks/judge phases (see CLAUDE.md "Bug vs
// feature routing").
var BugfixTrack = []string{"explore", "spec", "apply", "verify", "archive"}

// FastTrack is the fixed phase sequence for track "fast": a sized track for
// small, well-understood feature work (see harness-fast-track). Its phase
// names coincide with BugfixTrack's today, but it is kept as its own named
// slice rather than a shared reference: fast's "spec" phase produces a full
// Gherkin spec, while bugfix's produces the scoped 3-section format
// (sdd-spec branches on track, not slice identity) — sharing the slice
// would make a future divergence a surprising shared-reference edit.
var FastTrack = []string{"explore", "spec", "apply", "verify", "archive"}

// trackOrder resolves track to its fixed phase sequence. "" resolves to
// "full" for backward compatibility with state.yaml files written before
// the track field existed (CLAUDE.md / harness-workflow's documented
// back-compat rule) — it is never silently coerced for any other value.
func trackOrder(track string) ([]string, error) {
	switch track {
	case "", "full":
		return config.PhaseOrder, nil
	case "bugfix":
		return BugfixTrack, nil
	case "fast":
		return FastTrack, nil
	default:
		return nil, fmt.Errorf("unknown track: %s (supported: full, bugfix, fast)", track)
	}
}

// HistoryEntry is one row of state.yaml's history list.
type HistoryEntry struct {
	Phase  string `yaml:"phase"`
	Status string `yaml:"status"`
	TS     string `yaml:"ts"`
}

// State mirrors openspec/changes/{name}/state.yaml. RetryCount is new: it
// persists the judge phase's auto-re-apply attempt count so it survives a
// context compaction or a closed-and-reopened session, instead of living
// only in the orchestrating model's own count-as-you-go reasoning.
type State struct {
	Phase      string         `yaml:"phase"`
	Status     string         `yaml:"status"`
	Track      string         `yaml:"track,omitempty"`
	Size       string         `yaml:"size,omitempty"` // small|substantial|unknown
	RetryCount int            `yaml:"retry_count,omitempty"`
	History    []HistoryEntry `yaml:"history,omitempty"`
}

// StatePath returns the canonical state.yaml path for change changeName
// under projectDir.
func StatePath(projectDir, changeName string) string {
	return filepath.Join(projectDir, "openspec", "changes", changeName, "state.yaml")
}

// ErrStateNotFound is returned by ReadState when state.yaml does not exist —
// callers render harness-workflow's exact message: "state.yaml not found —
// run sdd-explore first".
var ErrStateNotFound = fmt.Errorf("state.yaml not found")

// ReadState reads and parses state.yaml at path. A missing file returns
// ErrStateNotFound (wrapped); invalid YAML returns a distinct error so
// callers can render harness-workflow's "corrupt state.yaml" message instead
// of the "not found" one.
func ReadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrStateNotFound, path)
		}
		return nil, fmt.Errorf("read state.yaml: %w", err)
	}

	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("corrupt state.yaml — delete and re-run sdd-explore: %w", err)
	}
	return &s, nil
}

// Write writes s to path atomically (temp file + rename), per
// harness-workflow's Hard Rule "Use atomic write (temp file + rename) when
// updating state.yaml."
func (s *State) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create change dir: %w", err)
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("encode state.yaml: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp state.yaml: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename state.yaml: %w", err)
	}
	return nil
}

// TransitionResult is the outcome of CheckTransition, mirroring
// harness-workflow/SKILL.md's "Workflow State" report (Step 4) and its
// Decision Gates table.
type TransitionResult struct {
	Allowed       bool
	Resuming      bool     // allowed && same phase as current (re-entry)
	Reason        string   // populated when !Allowed
	MissingPhases []string // populated when blocked by skipping >1 phase ahead
}

// CheckTransition implements skills/harness-workflow/SKILL.md's "Decision
// Gates" table verbatim:
//
//	current phase is "" (new change, no state.yaml yet) and requested ==
//	  the track's first phase                   -> allowed (bootstrap)
//	current phase is "" and requested != first   -> blocked (must start at first phase)
//	requested == current, status in_progress   -> allowed (resuming)
//	requested == current, status completed     -> blocked (already completed)
//	requested == next sequential phase          -> allowed (advancing)
//	requested is >1 step ahead                  -> blocked (missing phases listed)
//	requested is before current                 -> blocked (backward not allowed)
//	requested not in the track's sequence        -> blocked (not part of this track)
//
// Any divergence from skills/harness-workflow/SKILL.md "Decision Gates" here
// is a bug in this port, not a reinterpretation.
func CheckTransition(track, currentPhase, currentStatus, requestedPhase string) (TransitionResult, error) {
	order, err := trackOrder(track)
	if err != nil {
		return TransitionResult{}, err
	}

	requestedIdx := indexOf(order, requestedPhase)
	if requestedIdx == -1 {
		resolvedTrack := track
		if resolvedTrack == "" {
			resolvedTrack = "full"
		}
		return TransitionResult{
			Allowed: false,
			Reason:  fmt.Sprintf("%s is not part of the %s track", requestedPhase, resolvedTrack),
		}, nil
	}

	if currentPhase == "" {
		// Bootstrap case: no state.yaml existed yet for this change, so
		// there is no current phase to compare against. The only valid
		// first move is the track's first phase; anything else would skip
		// phases before the change even starts.
		if requestedIdx == 0 {
			return TransitionResult{Allowed: true}, nil
		}
		return TransitionResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("new change must start at %s, not %s", order[0], requestedPhase),
			MissingPhases: order[0:requestedIdx],
		}, nil
	}

	currentIdx := indexOf(order, currentPhase)
	if currentIdx == -1 {
		// current_phase isn't in this track's sequence either — can only
		// happen from hand-edited or cross-track-migrated state.yaml. Treat
		// it the same as any other unreachable-phase case rather than
		// panicking or guessing an index.
		resolvedTrack := track
		if resolvedTrack == "" {
			resolvedTrack = "full"
		}
		return TransitionResult{
			Allowed: false,
			Reason:  fmt.Sprintf("current phase %s is not part of the %s track", currentPhase, resolvedTrack),
		}, nil
	}

	switch {
	case requestedIdx == currentIdx && currentStatus == "in_progress":
		return TransitionResult{Allowed: true, Resuming: true}, nil
	case requestedIdx == currentIdx && currentStatus == "completed":
		return TransitionResult{Allowed: false, Reason: fmt.Sprintf("%s is already completed", requestedPhase)}, nil
	case requestedIdx == currentIdx+1 && currentStatus == "completed":
		return TransitionResult{Allowed: true}, nil
	case requestedIdx > currentIdx+1:
		return TransitionResult{
			Allowed:       false,
			Reason:        "requested phase skips one or more mandatory phases",
			MissingPhases: order[currentIdx+1 : requestedIdx],
		}, nil
	case requestedIdx <= currentIdx:
		return TransitionResult{Allowed: false, Reason: "backward transitions not allowed"}, nil
	default:
		// requestedIdx == currentIdx+1 && currentStatus == "in_progress":
		// the current phase hasn't been marked completed yet, so advancing
		// past it is a skip of exactly the current phase's completion.
		return TransitionResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("%s is still in_progress — complete it before advancing", currentPhase),
			MissingPhases: []string{currentPhase},
		}, nil
	}
}

// Advance validates the transition to requestedPhase against s (via
// CheckTransition) and, if allowed, mutates s in place: sets Phase to
// requestedPhase, Status to "in_progress", and appends a history entry
// timestamped now. It does not write to disk — callers persist via
// s.Write(path) so the check-then-write sequence stays a single caller-owned
// operation (no partial state between validation and persistence).
func (s *State) Advance(requestedPhase string, now time.Time) (TransitionResult, error) {
	result, err := CheckTransition(s.Track, s.Phase, s.Status, requestedPhase)
	if err != nil {
		return result, err
	}
	if !result.Allowed {
		return result, nil
	}

	s.Phase = requestedPhase
	s.Status = "in_progress"
	s.History = append(s.History, HistoryEntry{
		Phase:  requestedPhase,
		Status: "in_progress",
		TS:     now.UTC().Format(time.RFC3339),
	})
	return result, nil
}

// Complete marks the current phase completed and appends a history entry.
// It is the Go-side counterpart of harness-workflow's "On phase completion
// (reported by the sub-agent): update status to completed and record the
// timestamp."
func (s *State) Complete(now time.Time) {
	s.Status = "completed"
	s.History = append(s.History, HistoryEntry{
		Phase:  s.Phase,
		Status: "completed",
		TS:     now.UTC().Format(time.RFC3339),
	})
}

// MaxJudgeRetries is harness-judge's documented cap: "Maximum 3 retry
// cycles. The 4th failure returns blocked with max_retries_exceeded: true."
const MaxJudgeRetries = 3

// ErrMaxRetriesExceeded is returned by IncrementRetry once RetryCount is
// already at MaxJudgeRetries — the caller (harness-judge) must report
// `blocked` with `max_retries_exceeded: true` and stop, per its Hard Rules.
var ErrMaxRetriesExceeded = fmt.Errorf("max judge retries (%d) already reached", MaxJudgeRetries)

// IncrementRetry increments s.RetryCount and returns the new value, or
// ErrMaxRetriesExceeded without incrementing further if the cap is already
// reached. This makes the 3-retry cap durable across a context compaction or
// a closed-and-reopened session — previously it existed only in the
// orchestrating model's own count-as-you-go reasoning within one
// conversation.
func (s *State) IncrementRetry() (int, error) {
	if s.RetryCount >= MaxJudgeRetries {
		return s.RetryCount, ErrMaxRetriesExceeded
	}
	s.RetryCount++
	return s.RetryCount, nil
}

// ResetRetry zeroes s.RetryCount — called by harness-judge once the single
// focused review finally passes, so a later, unrelated judge cycle for this
// change (if any) starts clean.
func (s *State) ResetRetry() {
	s.RetryCount = 0
}

// PhaseHistoryMarkdown renders the "## Phase History" checklist block that
// skills/_shared/session-status-contract.md's Format section specifies,
// mechanically from s rather than having the orchestrator hand-type it —
// this is the one part of SESSION_STATUS.md fully derivable from state.yaml,
// so it's the one part this package generates; the narrative sections
// (preflight, artifacts, open questions, resume hint) still require the
// orchestrator's own judgment and stay hand-written prose.
func (s *State) PhaseHistoryMarkdown() (string, error) {
	order, err := trackOrder(s.Track)
	if err != nil {
		return "", err
	}

	completedTS := map[string]string{}
	inProgressTS := map[string]string{}
	for _, h := range s.History {
		switch h.Status {
		case "completed":
			completedTS[h.Phase] = h.TS
		case "in_progress":
			inProgressTS[h.Phase] = h.TS
		}
	}

	var b strings.Builder
	b.WriteString("## Phase History\n")
	for _, phase := range order {
		if ts, ok := completedTS[phase]; ok {
			fmt.Fprintf(&b, "- [x] %s — completed %s\n", phase, ts)
			continue
		}
		if phase == s.Phase && s.Status == "in_progress" {
			fmt.Fprintf(&b, "- [ ] %s — in_progress %s\n", phase, inProgressTS[phase])
			continue
		}
		fmt.Fprintf(&b, "- [ ] %s\n", phase)
	}
	return b.String(), nil
}

// StatusLine renders the "**Current phase**: X (Y)" line from
// session-status-contract.md's Format section.
func (s *State) StatusLine() string {
	return fmt.Sprintf("**Current phase**: %s (%s)", s.Phase, s.Status)
}

func indexOf(items []string, item string) int {
	for i, v := range items {
		if v == item {
			return i
		}
	}
	return -1
}
