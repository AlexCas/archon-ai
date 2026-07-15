# Design — opencode-phase-writers (Slice 2)

Apply-ready. Extends `internal/initcmd/opencode_mode.go` to emit `archon-<phase>` subagents
(reusing `config.ResolvePhaseModels`) alongside the existing `archon-leader`. Settled
requirements from `specs/opencode-phase-writers/spec.md`.

## 1. New struct `archonPhaseAgent`

Add next to `archonLeaderAgent` in `opencode_mode.go`. Field declaration order = JSON order.

```go
// phaseAgentName returns the opencode.json agent key for an SDD phase, e.g.
// "archon-spec".
func phaseAgentName(phase string) string { return "archon-" + phase }

// archonPhaseAgent is the fixed shape written under agent.archon-<phase> for each
// resolvable SDD phase. Field declaration order is the deterministic JSON order.
type archonPhaseAgent struct {
	Mode        string `json:"mode"`        // always "subagent"
	Hidden      bool   `json:"hidden"`      // always true (no omitempty)
	Model       string `json:"model"`       // resolved per-phase FullID
	Description string `json:"description"` // "Archon SDD <phase> phase"
	Prompt      string `json:"prompt"`      // "{file:./AGENTS.md}"
}
```

`Hidden bool` has NO `omitempty` (always serialized `true`).

## 2. Signature change

`opencode_mode.go` must import `"github.com/archon-ai/archon/internal/config"` (verify exact
module path from existing imports in the package; `init.go` already imports config).

**Before:**
```go
func MergeOpencodeAgent(projectDir, leader string) (written string, err error)
func mergeOpencodeAgent(projectDir, leader string) (written string, err error)
```
**After:**
```go
func MergeOpencodeAgent(projectDir string, models config.ModelConfig) (written string, err error)
func mergeOpencodeAgent(projectDir string, models config.ModelConfig) (written string, err error)
```
Exported delegates: `return mergeOpencodeAgent(projectDir, models)`.

## 3. Merge algorithm

```go
func mergeOpencodeAgent(projectDir string, models config.ModelConfig) (written string, err error) {
	leaderFull := models.Leader.FullID()
	phases := config.ResolvePhaseModels(models) // []PhaseModel, FullID, omit-when-empty
	if leaderFull == "" && len(phases) == 0 {
		return "", nil // nothing to write
	}

	path := filepath.Join(projectDir, "opencode.json")
	// ... unchanged read/parse into doc map[string]any ...

	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		agents = map[string]any{}
	}
	if leaderFull != "" {
		agents[leaderAgentName] = archonLeaderAgent{
			Mode:        "primary",
			Description: "Archon SDD orchestration leader",
			Model:       leaderFull,
			Prompt:      "{file:./AGENTS.md}",
		}
	}
	for _, pm := range phases {
		agents[phaseAgentName(pm.Phase)] = archonPhaseAgent{
			Mode:        "subagent",
			Hidden:      true,
			Model:       pm.Model, // already FullID from ResolvePhaseModels
			Description: "Archon SDD " + pm.Phase + " phase",
			Prompt:      "{file:./AGENTS.md}",
		}
	}
	doc["agent"] = agents
	// ... unchanged MarshalIndent + trailing newline + atomic temp+rename ...
	return path, nil
}
```

Notes:
- Leader gated individually on `leaderFull != ""` (previously the whole func no-op'd on empty
  leader; the no-op moved up to the combined condition).
- `pm.Model` is already the FullID (ResolvePhaseModels emits `ref.FullID()`), so no extra join.
- Determinism: struct field order fixed; `MarshalIndent` sorts the `agent` map keys
  (archon-apply, archon-archive, …, archon-leader, …) + user agents; trailing newline; atomic
  rename — byte-identical re-runs.
- Update the doc comment to describe leader + per-phase subagents.

## 4. Call sites

- `internal/initcmd/init.go:102`:
  before `mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader.FullID())`
  after  `mergeOpencodeAgent(opts.ProjectDir, cfg.Models)`
- `internal/tui/model.go:334`:
  before `initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models.Leader.FullID())`
  after  `initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models)`

Rollback (`init.go:106-108`) unchanged — same single `opencode.json` path.

## 5. Test plan (opencode_mode_test.go + model_test.go)

Signature migration (compile-gating): the 6 `mergeOpencodeAgent(dir, testLeaderModel)` /
`MergeOpencodeAgent(...)` calls become
`mergeOpencodeAgent(dir, config.ModelConfig{Leader: config.ParseModelRef(testLeaderModel)})`.
`model_test.go:675` reference merge likewise. `testLeaderModel` stays
`"anthropic/claude-sonnet-4-20250514"`.

Add helper `phaseAgentFrom(t, doc, phase) map[string]any` (mirrors `leaderAgentFrom`).

Migrate existing tests (signature only, behavior preserved): `_CreatesAgent`,
`_RegistersOpencodeJSONForRollback`, `_PreservesExisting`, `_Idempotent`,
`_EmptyLeaderWritesNothing`→`_NothingConfiguredWritesNothing` (empty ModelConfig{}),
`Run_NonOpencodeWritesNoOpencodeJSON`.

New tests (one per spec scenario):
- `_WritesSubagentPerResolvablePhase` — Default + a phase set → archon-<phase> for each
  ResolvePhaseModels phase; no archon-judge.
- `_PhaseModelMatchesResolvedFullID` — phases.spec = {opencode, deepseek-v4-pro} →
  archon-spec.model == "opencode/deepseek-v4-pro".
- `_PhaseFallsBackToDefault` — phases.tasks empty, default set → archon-tasks.model == default FullID.
- `_SubagentFixedFields` — mode/hidden/description/prompt per phase.
- `_PhasesSetEmptyLeaderWritesSubagentsNoLeader` — leader empty, default set → subagents present,
  archon-leader absent.
- `_PreservesAndIdempotent` (extend existing) — user agent preserved; byte-identical re-run over
  the full leader+subagent set.

## 6. Size
Prod ~45-60 LOC (struct + helper + loop + no-op + signature + comment); 2 call-site lines.
Tests ~120-160 LOC. Total ~180-230 LOC. Under D1 400 — single PR.
