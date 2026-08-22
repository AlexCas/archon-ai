// Package route implements the deterministic code pre-router for SDD phase
// dispatch: it resolves which SDD phase a natural-language message targets
// (or signals that a model classifier / human ASK is needed) without ever
// invoking an LLM. The router is READ-ONLY on state — harness-workflow
// remains the sole state writer and legality gate.
package route

import (
	"strings"
	"unicode"

	"github.com/archon-ai/archon/internal/config"
	"golang.org/x/text/unicode/norm"
)

// pathCode and pathModel are the two possible Result.Path values.
const (
	pathCode  = "code"
	pathModel = "model"
)

// Input is the resolver's sole input. All IO — state.yaml reads, active
// change discovery, --phase/--status flag overrides — is resolved into
// Input by the CLI layer before calling Resolve.
type Input struct {
	// Message is the raw user message; Resolve normalizes it internally
	// (lowercase + diacritic strip) before matching any rule.
	Message string
	// Phase is the current SDD phase for ActiveChange, or "" if none.
	Phase string
	// Status is the current phase's status (in_progress|completed), or ""
	// if none.
	Status string
	// ActiveChange is the resolved active change name, or "none" when no
	// change is active (bootstrap state).
	ActiveChange string
}

// Result is the resolver's sole output.
type Result struct {
	Phase        string `json:"phase"`
	Rule         string `json:"rule"`
	Path         string `json:"path"`
	ActiveChange string `json:"active_change"`
}

// Resolve applies the deterministic phase-dispatch rules in strict
// top-to-bottom, first-match-wins precedence:
//
//	explicit-agent > control > implicit > ambiguous (D3) > keyword > else CLASSIFY
//
// Resolve is pure: it performs no IO and never invokes an LLM or external
// service. When no code rule fires, it emits Phase="CLASSIFY", Path="model"
// to signal the leader to invoke the model classifier (skills/sdd-router).
func Resolve(in Input) Result {
	msg := Normalize(in.Message)
	activeChange := in.ActiveChange
	if activeChange == "" {
		activeChange = "none"
	}

	if phase, ok := matchExplicitAgent(msg); ok {
		return Result{Phase: phase, Rule: "explicit-agent", Path: pathCode, ActiveChange: activeChange}
	}

	if matchAny(msg, controlWords) {
		return resolveControl(in.Phase, in.Status, activeChange)
	}

	if matchAny(msg, implicitVerbs) {
		return resolveImplicit(in.Phase, activeChange)
	}

	if isAmbiguous(msg) {
		return Result{Phase: "ASK", Rule: "ambiguous", Path: pathCode, ActiveChange: activeChange}
	}

	if phase, ok := matchKeyword(msg); ok {
		return Result{Phase: phase, Rule: "keyword", Path: pathCode, ActiveChange: activeChange}
	}

	return Result{Phase: "CLASSIFY", Rule: "classify", Path: pathModel, ActiveChange: activeChange}
}

// resolveControl implements the "control" rule: a control word resolved
// against ActiveChange/Status. ActiveChange=="none" (bootstrap) always wins
// over status, since there is no current phase to resume or advance from.
func resolveControl(phase, status, activeChange string) Result {
	if activeChange == "none" {
		return Result{Phase: "explore", Rule: "next-nochange", Path: pathCode, ActiveChange: activeChange}
	}
	if status == "in_progress" {
		return Result{Phase: phase, Rule: "resume", Path: pathCode, ActiveChange: activeChange}
	}
	next := phase
	if n, ok := nextPhase(phase); ok {
		next = n
	}
	return Result{Phase: next, Rule: "next", Path: pathCode, ActiveChange: activeChange}
}

// resolveImplicit implements the "implicit" rule: a start/continue verb with
// no explicit agent named. ActiveChange=="none" starts explore; otherwise it
// resumes the current phase — status is not consulted (task 2.4).
func resolveImplicit(phase, activeChange string) Result {
	if activeChange == "none" {
		return Result{Phase: "explore", Rule: "implicit-start", Path: pathCode, ActiveChange: activeChange}
	}
	return Result{Phase: phase, Rule: "implicit-resume", Path: pathCode, ActiveChange: activeChange}
}

// nextPhase returns the phase immediately after phase in config.PhaseOrder —
// the sole canonical phase-sequence source (no second list in this package).
func nextPhase(phase string) (string, bool) {
	for i, p := range config.PhaseOrder {
		if p == phase && i+1 < len(config.PhaseOrder) {
			return config.PhaseOrder[i+1], true
		}
	}
	return "", false
}

// matchExplicitAgent detects a literal archon-<phase> token or a navigation
// marker phrase (e.g. "corre el apply", "volvamos al spec") naming a phase
// as the thing to run or the destination. It always wins over every other
// rule regardless of whether the target is ahead of or behind the current
// phase — legality is harness-workflow's job, not the router's.
func matchExplicitAgent(msg string) (string, bool) {
	if m := archonAgentRe.FindStringSubmatch(msg); m != nil {
		return m[1], true
	}
	if m := navigationRe.FindStringSubmatch(msg); m != nil {
		if phase, ok := phaseAliases[m[1]]; ok {
			return phase, true
		}
	}
	return "", false
}

// isAmbiguous implements the D3 narrow dual-action rule: a judge-verb AND a
// verify-verb AND a coordinating conjunction all present in msg.
func isAmbiguous(msg string) bool {
	return matchAny(msg, judgeVerbs) && matchAny(msg, verifyVerbs) && matchAny(msg, conjunctions)
}

// matchKeyword scans keywordTable in config.PhaseOrder order (for
// deterministic iteration — Go map order is randomized). Exactly one
// matching phase resolves; zero or multiple matches fall through to
// CLASSIFY (design decision A5 — CLASSIFY, not ASK, for the fallthrough).
func matchKeyword(msg string) (string, bool) {
	var matched string
	count := 0
	for _, phase := range config.PhaseOrder {
		words, ok := keywordTable[phase]
		if !ok {
			continue
		}
		if matchAny(msg, words) {
			matched = phase
			count++
		}
	}
	if count == 1 {
		return matched, true
	}
	return "", false
}

// Normalize lowercases s and strips diacritics (e.g. "especificación" ->
// "especificacion", "diseño" -> "diseno"), so rule matching is accent- and
// case-insensitive. Normalization runs before any rule evaluation.
func Normalize(s string) string {
	lower := strings.ToLower(s)
	decomposed := norm.NFD.String(lower)
	var b strings.Builder
	b.Grow(len(decomposed))
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
