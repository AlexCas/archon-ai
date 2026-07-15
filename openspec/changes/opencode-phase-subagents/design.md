# Design — opencode-phase-subagents

**Change**: Extend the opencode writer to emit one `archon-<phase>` subagent per SDD
phase (each with its resolved per-phase model), alongside the existing `archon-leader`.
**Branch**: `feat/opencode-phase-subagents` (base `master`).
**Phase**: design. Single file modified plus call sites and tests.

This is a low-level, implementation-ready design. The apply phase follows it mechanically.
All requirements come from `specs/opencode-phase-subagents/spec.md` and are SETTLED; this
document does not re-open them.

---

## 1. New struct: `archonPhaseAgent`

Added to `internal/initcmd/opencode_mode.go`, next to `archonLeaderAgent`. Struct field
declaration order IS the deterministic JSON output order (Go's `encoding/json` emits struct
fields in declaration order, not sorted — only `map` keys are sorted). The fixed order
required by the spec is `mode, hidden, model, description, prompt`.

```go
// archonPhaseAgent is the fixed shape written under agent.archon-<phase> for each
// SDD phase in config.PhaseOrder. Struct field declaration order is the
// deterministic JSON output order.
type archonPhaseAgent struct {
	Mode        string `json:"mode"`        // always "subagent"
	Hidden      bool   `json:"hidden"`      // always true; see note below
	Model       string `json:"model"`       // verbatim resolved per-phase model
	Description string `json:"description"` // "Archon SDD <phase> phase"
	Prompt      string `json:"prompt"`      // "{file:./AGENTS.md}"
}
```

**`Hidden bool` — no `omitempty`, always serialized as `true`.** Confirmed: every per-phase
subagent is hidden, so `Hidden` is set to `true` for all 8 entries and is never `false`.
`omitempty` on a `bool` would drop the field when the value is `false`; since the value is
always `true` it would never actually be dropped, but adding `omitempty` would be misleading
(it signals the field is optional, which it is not). We deliberately omit `omitempty` so the
field is always present and always `true`, matching the spec requirement "every
`agent.archon-<phase>` has `hidden` equal to true".

A new key constant alongside `leaderAgentName`:

```go
// phaseAgentName returns the opencode.json agent key for an SDD phase, e.g.
// "archon-spec". Mirrors the leaderAgentName scheme.
func phaseAgentName(phase string) string { return "archon-" + phase }
```

---

## 2. Signatures: `mergeOpencodeAgent` and `MergeOpencodeAgent`

We pass the whole `config.ModelConfig`. This is the cleanest option: it carries `Leader`,
`Default`, and `Phases` in one value, lets the writer own the full resolution chain, and
keeps the two call sites trivial (`cfg.Models`). A narrower param set
(`leader string, phases map[string]string, default string`) would force every caller to
destructure the config and would not be more testable.

`config` is already imported transitively by the package (the test file imports it); the
source file `opencode_mode.go` must add `"github.com/archon-ai/archon/internal/config"` to
its import block.

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

The exported `MergeOpencodeAgent` continues to delegate:

```go
func MergeOpencodeAgent(projectDir string, models config.ModelConfig) (written string, err error) {
	return mergeOpencodeAgent(projectDir, models)
}
```

---

## 3. Model-resolution helper: `resolvePhaseModel`

A small pure function in `opencode_mode.go` implementing the VERBATIM fallback chain. It does
NOT call `config.NormalizeModel` (that would corrupt provider-qualified ids into bare aliases,
which opencode cannot use — see exploration "central design problem"). It only picks the first
non-empty value in the chain.

```go
// resolvePhaseModel returns the verbatim model for an SDD phase, applying the
// fallback chain: Models.Phases[phase] -> Models.Default -> Models.Leader.
// No normalization or prefix handling — the value is written to opencode.json
// exactly as configured, mirroring the leader's verbatim contract.
func resolvePhaseModel(mc config.ModelConfig, phase string) string {
	if v := mc.Phases[phase]; v != "" {
		return v
	}
	if mc.Default != "" {
		return mc.Default
	}
	return mc.Leader
}
```

Note: when all three are empty `resolvePhaseModel` returns `""`. That case never reaches the
loop because the whole-merge no-op check (Section 4) returns early when leader, default, and
all phases are empty. When the merge does run, at least one of the three is non-empty, so every
emitted phase gets a non-empty model. (A phase whose own value is empty still emits, using the
non-empty default or leader.)

---

## 4. Merge algorithm (step by step)

`mergeOpencodeAgent(projectDir, models)`:

1. **Whole-merge no-op check.** Return `("", nil)` when `models` is empty by the new
   definition: `Leader == "" AND Default == "" AND every Phases value is empty`. Implement
   with a small method on `config.ModelConfig`? No — `config` is a separate package and we
   should not grow its surface for this. Implement as an unexported helper in
   `opencode_mode.go`:

   ```go
   // modelsEmpty reports whether nothing would be written: no leader, no default,
   // and no non-empty per-phase model.
   func modelsEmpty(mc config.ModelConfig) bool {
   	if mc.Leader != "" || mc.Default != "" {
   		return false
   	}
   	for _, v := range mc.Phases {
   		if v != "" {
   			return false
   		}
   	}
   	return true
   }
   ```

   ```go
   if modelsEmpty(models) {
   	return "", nil
   }
   ```

   (The non-opencode no-op is enforced by the caller's `agentName == "opencode"` /
   `cfg.Agent == "opencode"` gate, unchanged — this function is only reached for opencode.)

2. **Path + read existing doc.** Unchanged:
   `path := filepath.Join(projectDir, "opencode.json")`; read into `doc map[string]any`;
   on `os.ReadFile` success `json.Unmarshal`; on a non-`IsNotExist` read error, wrap and
   return. Missing file -> empty `doc`.

3. **Ensure agent map.** Unchanged:
   `agents, ok := doc["agent"].(map[string]any); if !ok { agents = map[string]any{} }`.

4. **Set leader — only when `models.Leader != ""`.** This PRESERVES the existing leader
   behavior exactly. (Previously the whole function no-op'd on empty leader; now the no-op
   moved up to step 1, so we must guard the leader entry individually so an empty leader does
   not emit an `archon-leader` with an empty model.) Per spec scenario "Default set with empty
   leader still writes subagents", no `archon-leader` is written when leader is empty.

   ```go
   if models.Leader != "" {
   	agents[leaderAgentName] = archonLeaderAgent{
   		Mode:        "primary",
   		Description: "Archon SDD orchestration leader",
   		Model:       models.Leader,
   		Prompt:      "{file:./AGENTS.md}",
   	}
   }
   ```

5. **Loop `config.PhaseOrder` setting `archon-<phase>`.** Iterating the slice (not a map)
   gives deterministic order; `judge` is absent from `PhaseOrder`, so it is never emitted.
   All 8 phases are written.

   ```go
   for _, phase := range config.PhaseOrder {
   	agents[phaseAgentName(phase)] = archonPhaseAgent{
   		Mode:        "subagent",
   		Hidden:      true,
   		Model:       resolvePhaseModel(models, phase),
   		Description: "Archon SDD " + phase + " phase",
   		Prompt:      "{file:./AGENTS.md}",
   	}
   }
   ```

6. **Reattach + marshal + atomic write.** Unchanged from current code:
   `doc["agent"] = agents`; `json.MarshalIndent(doc, "", "  ")` (sorts map keys, so
   `archon-archive` < `archon-apply` ... ordering is alphabetical and stable); append `'\n'`;
   write `path + ".tmp"` with `0o644`; `os.Rename(tmp, path)`.

7. **Return** `path, nil`.

The struct field order is fixed by declaration; map keys (the agent names and the top-level
keys) are sorted by `MarshalIndent`. Together these guarantee byte-identical re-runs.

---

## 5. Call-site edits

### `internal/initcmd/init.go` (around line 102)

**Before:**

```go
if agentName == "opencode" {
	mergedPath, err := mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader)
```

**After:**

```go
if agentName == "opencode" {
	mergedPath, err := mergeOpencodeAgent(opts.ProjectDir, cfg.Models)
```

Everything else in that block (rollback registration of `mergedPath`, error wrap) is
unchanged.

### `internal/tui/model.go` (line 334)

**Before:**

```go
if cfg.Agent == "opencode" {
	if _, err := initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models.Leader); err != nil {
```

**After:**

```go
if cfg.Agent == "opencode" {
	if _, err := initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models); err != nil {
```

The surrounding comment can be updated to "merge the archon-leader plus per-phase subagents",
but that is cosmetic and optional.

---

## 6. Rollback

**No change needed.** `init.go` already appends the returned `mergedPath` to
`rollback.CreatedPaths` before `WriteManifest`. The writer still writes exactly one file
(`opencode.json`) and returns its path. Adding more agents inside that single file does not
create any new path to register. The existing rollback test
(`TestRun_RegistersOpencodeJSONForRollback`) continues to assert the same path is registered;
it keeps passing unchanged.

---

## 7. Test plan

All tests live in `internal/initcmd/opencode_mode_test.go` unless noted. Reuse existing
helpers `readOpencodeDoc`, `leaderAgentFrom`, and the constant `testLeaderModel`. Add a small
helper `phaseAgentFrom(t, doc, phase)` mirroring `leaderAgentFrom` for the per-phase maps, and
a fixture builder for a fully-populated `config.ModelConfig`.

**Signature migration (compile-gating, required first):** every existing call to
`mergeOpencodeAgent(dir, testLeaderModel)` becomes
`mergeOpencodeAgent(dir, config.ModelConfig{Leader: testLeaderModel})`, and the
`tui/model_test.go:702` call `initcmd.MergeOpencodeAgent(initDir, leader)` becomes
`initcmd.MergeOpencodeAgent(initDir, config.ModelConfig{Leader: leader})`. The TUI test's
`cfg.Models` is already a `ModelConfig{Leader: leader}`, so the TUI save path and the
reference path stay equivalent and the byte-identical assertion still holds.

Existing tests to migrate (signature only, behavior preserved):
- `TestMergeOpencodeAgent_CreatesAgent` — wrap leader in `ModelConfig`. Still asserts the
  leader shape. (Optionally extend to also assert the 8 phase agents exist; see new cases.)
- `TestMergeOpencodeAgent_PreservesExisting` — wrap leader; the pre-existing `agent.build`
  and top-level `$schema`/`theme` assertions stay. Maps to spec "Preserve existing keys and
  user agents".
- `TestMergeOpencodeAgent_Idempotent` — wrap leader; still asserts byte-identical re-run.
  Maps to "Re-run is byte-identical".
- `TestMergeOpencodeAgent_EmptyLeaderWritesNothing` — change to pass an
  EMPTY `config.ModelConfig{}` (leader, default, phases all empty) and rename to reflect the
  new no-op condition (`..._EmptyConfigWritesNothing`). Maps to "Everything empty writes
  nothing". Keep asserting no file is created and `written == ""`.
- `TestRun_RegistersOpencodeJSONForRollback` — unchanged (drives `Run`, not the writer
  directly; `ModelLeader` already set).
- `TestRun_NonOpencodeWritesNoOpencodeJSON` — unchanged. Maps to "Non-opencode agent writes
  nothing".

New test cases (one per remaining spec scenario):

| Test | Spec scenario | Asserts |
|------|---------------|---------|
| `TestMergeOpencodeAgent_WritesEightPhaseSubagents` | Init writes 8 per-phase subagents | for each `p` in `config.PhaseOrder`, `agent.archon-<p>` exists; exactly 8; `agent.archon-judge` absent |
| `TestMergeOpencodeAgent_PhaseModelVerbatim` | Phase with its own model uses it verbatim | `Phases["spec"]="anthropic/claude-opus-4-20250514"` -> `agent.archon-spec.model` equals it byte-for-byte |
| `TestMergeOpencodeAgent_PhaseFallsBackToDefault` | Phase without a model falls back to default | `Phases["tasks"]` empty, `Default` set -> `agent.archon-tasks` present, model == default |
| `TestMergeOpencodeAgent_PhaseFallsBackToLeader` | Phase and default empty fall back to leader | `Phases["verify"]`+`Default` empty, `Leader` set -> `agent.archon-verify` present, model == leader |
| `TestMergeOpencodeAgent_DefaultOnlyNoLeader` | Default set with empty leader still writes subagents | `Leader=""`, `Default` set, phases empty -> file written, all 8 subagents present (each model == default), `agent.archon-leader` ABSENT |
| `TestMergeOpencodeAgent_SubagentFixedFields` | Subagent carries the fixed fields | every `archon-<phase>` has `mode=="subagent"`, `hidden==true` (JSON bool), and non-empty `model`, `description`, `prompt`; description == "Archon SDD <phase> phase"; prompt == "{file:./AGENTS.md}" |

The "Re-run is byte-identical" and "8 subagents" cases together exercise idempotency with the
full phase set; the migrated `TestMergeOpencodeAgent_Idempotent` should be upgraded to use a
fully-populated `ModelConfig` (leader + default + per-phase models) so idempotency is proven
over the whole written document, not just the leader.

`tui/model_test.go:702` — update the single `MergeOpencodeAgent` call to the new signature as
described above. No new TUI test is required; the existing equivalence test now covers the
full document (leader + 8 subagents) since both paths receive the same `ModelConfig`.

---

## 8. Edge cases & determinism notes

- **`hidden:true` serialization.** `Hidden bool` with no `omitempty` always emits
  `"hidden": true`. JSON unmarshal into `map[string]any` yields a Go `bool` `true`, so tests
  assert `agent["hidden"] == true` (not the string `"true"`).
- **Field-order determinism.** Struct fields serialize in declaration order
  (`mode, hidden, model, description, prompt`); this is a property of `encoding/json` and is
  not affected by `MarshalIndent`'s key sorting (which only applies to maps). The leader's
  existing 4-field order is unchanged.
- **Map-key determinism.** `MarshalIndent` sorts map keys, so the `agent` object lists
  `archon-apply, archon-archive, archon-design, archon-explore, archon-leader, archon-propose,
  archon-spec, archon-tasks, archon-verify` plus any user agents in alphabetical order on
  every run. Trailing newline appended explicitly. Atomic temp+rename. => byte-identical
  re-runs.
- **`judge` never emitted.** `judge` is intentionally absent from `config.PhaseOrder`, and we
  iterate that slice exactly, so no `archon-judge` key is ever produced. A test explicitly
  asserts its absence.
- **Interaction with the existing leader entry.** The leader is still written with its own
  unchanged struct/shape, gated on `models.Leader != ""`. Phase subagents are written
  independently. When leader is empty but default/phases are not, the document has the 8
  subagents and no leader — exactly the spec's "Default set with empty leader" scenario.
- **Stale `archon-<phase>` across config changes (known limitation, out of scope).** The
  merge OVERWRITES every `archon-<phase>` key on each run (all 8 are always set), so a model
  change is reflected. However, because `judge` is never in `PhaseOrder` and the phase set is
  fixed at 8, there is no scenario where a previously-written `archon-<phase>` key becomes
  orphaned by a config change — the same 8 keys are always rewritten. If a future change
  removed a phase from `PhaseOrder`, the now-unused `archon-<oldphase>` key would remain in
  `opencode.json` (overwritten-not-removed semantics). That pruning is OUT OF SCOPE here and
  noted as a known limitation; it is not triggered by any in-scope configuration change.
- **Existing user agents and top-level keys** are read into `doc`/`agents` and only the
  `archon-*` keys are set, so unrelated agents (e.g. `build`) and top-level keys
  (`$schema`, `theme`) survive untouched — covered by the migrated
  `TestMergeOpencodeAgent_PreservesExisting`.

---

## 9. Size estimate (LOC)

| Area | Approx. changed lines |
|------|----------------------|
| `opencode_mode.go`: struct, `phaseAgentName`, `resolvePhaseModel`, `modelsEmpty`, loop, leader guard, signature, import | ~55 |
| `init.go` call site | 1 |
| `tui/model.go` call site (+comment) | 1–2 |
| `opencode_mode_test.go`: migrate 4 tests + 6 new cases + 1 helper | ~140 |
| `tui/model_test.go`: 1 call-site line | 1 |
| **Total** | **~200** |

Comfortably under the **D1 400-line** review budget. Single PR, no chaining (C1 respected).

---

## Confirmed design decisions

1. **Pass whole `config.ModelConfig`** to both `mergeOpencodeAgent` and `MergeOpencodeAgent`
   (vs a narrower tuple). Cleanest, fewest caller changes, lets the writer own resolution.
2. **`description = "Archon SDD <phase> phase"`**, **`prompt = "{file:./AGENTS.md}"`** for all
   phases (orchestrator-confirmed). Static and identical-shape across phases for determinism.
3. **`Hidden bool` without `omitempty`, always `true`.** Always serialized.
4. **No-op condition** moved up to a `modelsEmpty` guard; leader entry guarded individually so
   an empty leader with a non-empty default/phase still writes the 8 subagents and no leader.
5. **Verbatim model** via `resolvePhaseModel`; `NormalizeModel` deliberately NOT used.

## Open risk to confirm with the user

- **Verbatim, unqualified models.** If a user configured a per-phase model as a bare alias
  (e.g. `opus`) rather than a provider-qualified id (e.g. `anthropic/claude-opus-4-...`),
  opencode may reject the resulting subagent's `model`. This mirrors the leader's existing
  verbatim contract (same risk already shipped for the leader), so it is consistent, but worth
  flagging: this change does not add any validation or normalization for per-phase models.
