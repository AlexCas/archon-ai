# Design: Local Model Provider (Ollama / LocalAI via OpenCode)

<!-- proposal: proposal.md | spec: specs/local-model-provider/spec.md -->

Implements [[local-model-provider]]. HOW to satisfy REQ-1..REQ-7. Decisions in
proposal/spec are LOCKED; this doc is the deterministic implementation map.

## Technical Approach

Add a scalar `ModelRef.BaseURL` (YAML `base_url`, `omitempty`). Its only functional
consumers are the two generation backends. On the OpenCode path, refs that carry a
BaseURL are coalesced by provider id into ONE top-level `provider.<id>` block in
`opencode.json`, additively merged like today's `agent` block. On the Claude path a
warn-and-skip guard fires (BaseURL cannot be honored) and the bare-model agent file is
still written. CLI and TUI gain read/write access to the field. All warnings go to an
injected `io.Writer` (stderr), never to `os.Stderr` inline — matching the existing
`stdout/stderr` seam in `cmd/archon/config.go`.

Two seams must widen to carry the new data (both are the crux of the design):

- `ResolvePhaseModels` today returns `PhaseModel{Phase, Model(FullID), Effort}` — it
  drops the raw `Provider` and never had a `BaseURL`. The provider block needs the raw
  provider id AND the BaseURL per resolved ref. **Add `Provider string` and `BaseURL
  string` to `PhaseModel`** and also expose the resolved Leader ref the same way. This
  keeps `mergeOpencodeAgent`/`writeClaudeAgents` operating on resolved refs (Default
  fallback already applied) rather than re-reading `ModelConfig`.
- `mergeOpencodeAgent` / `writeClaudeAgents` (and their exported `MergeOpencodeAgent` /
  `WriteClaudeAgents` seams) gain an `io.Writer` param for warnings. Callers: `init.go`
  passes `os.Stderr`; TUI `model.go:372/381` passes its own writer (or `io.Discard`).

## Architecture Decisions

| Decision | Choice | Alternative rejected | Rationale |
|----------|--------|----------------------|-----------|
| Scalar-vs-mapping switch | Mapping iff `Effort != "" \|\| BaseURL != ""`; else scalar `FullID()` | New field always forces mapping | Preserves byte-identical round-trip for every existing config (REQ-1) |
| Carry provider+baseURL to backends | Extend `PhaseModel` + a resolved leader accessor | Re-derive from `ModelConfig` inside backends | Backends already consume `ResolvePhaseModels`; keeps Default-fallback logic in ONE place |
| Warning sink | Inject `io.Writer` into the two writers | `os.Stderr` inline / return a `[]warning` | Matches CLI stdout/stderr seam; testable; no API to marshal warnings |
| Provider block build | Pure helper returns `map[string]any`, merged into `doc["provider"]` | Typed struct per provider | `json.MarshalIndent` sorts `map[string]any` keys → free deterministic ordering (REQ-5) |
| Coalescing order | Iterate resolved refs in `PhaseOrder`, leader last; first BaseURL per id wins | Sort refs first | Spec fixes "first encountered" = PhaseOrder traversal, default/leader last |
| CLI base_url key | Dotted suffix `.base_url` on existing model keys | New `models.providers.*` tree | Approach 1 is locked; leaner CLI surface |

## Data Flow

    config.yaml ─load─▶ ModelConfig ─ResolvePhaseModels─▶ []PhaseModel{Provider,Model,BaseURL,Effort}
                                                              │
                          ┌───────────────────────────────────┴─────────────┐
              agent==opencode                                        agent==claude
                          │                                                  │
             buildProviderBlock(refs) ──coalesce/sort──▶ doc["provider"]     guardLocalRef → warn(w)
             agents[...]=FullID()      ──▶ doc["agent"]                       renderClaudeAgent(bare model)
                          │                                                  │
                 opencode.json (merged, atomic)                     .claude/agents/archon-<phase>.md

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modify | Add `BaseURL` field; `MarshalYAML` mapping-switch includes BaseURL; add `Provider`+`BaseURL` to `PhaseModel` and populate in `ResolvePhaseModels`; add `ResolveLeader(ModelConfig) ModelRef` (or reuse `Leader`); add `ValidateBaseURL(ref) string` advisory helper |
| `internal/config/config.go` | Modify | Clone already value-copies ModelRef (scalar field auto-covered) — no code change; note only |
| `internal/initcmd/opencode_mode.go` | Modify | Add `io.Writer` param; `buildProviderBlock(refs, w) map[string]any`; merge into `doc["provider"]` preserving unknown keys |
| `internal/initcmd/claude_mode.go` | Modify | Add `io.Writer` param; warn-and-skip guard per local phase before writing bare-model file |
| `internal/initcmd/init.go` | Modify | Pass `os.Stderr` to both writers |
| `internal/tui/model.go` | Modify | Pass a writer (e.g. `io.Discard`) to `MergeOpencodeAgent`/`WriteClaudeAgents` |
| `cmd/archon/config.go` | Modify | `set/get` handle `.base_url` suffix; `list` prints `base_url` lines |
| `internal/tui/models_tab.go` | Modify | New `baseURLEdit` sub-mode: state, key handling, render, `applyToConfig` persistence |

## Interfaces / Contracts

```go
// model.go — mapping switch (REQ-1)
func (r ModelRef) MarshalYAML() (any, error) {
    if r.Effort == "" && r.BaseURL == "" {
        return r.FullID(), nil // scalar, byte-identical
    }
    type alias ModelRef
    return alias(r), nil
}

// PhaseModel gains Provider + BaseURL; ResolvePhaseModels populates from the
// resolved ref (after Default fallback), leaving Model=FullID() unchanged.

// opencode_mode.go — provider block (REQ-4/5). refs = resolved phase refs + leader.
func buildProviderBlock(refs []config.PhaseModel, w io.Writer) map[string]any
// per ref with BaseURL!="" : provider[id].npm="@ai-sdk/openai-compatible";
//   options.baseURL=first-seen (warn on conflict); models[model]={"name":model}.
// returns nil when no ref carries a BaseURL → doc["provider"] untouched.
```

Resulting JSON shape (REQ-4):

```json
"provider": {
  "ollama": {
    "npm": "@ai-sdk/openai-compatible",
    "options": { "baseURL": "http://localhost:11434/v1" },
    "models": { "llama3": { "name": "llama3" } }
  }
}
```

Merge rule: read existing `doc["provider"]` (if a `map[string]any`), set only archon-built
ids, never delete user ids. Model keys and provider ids sort for free via
`json.MarshalIndent`.

## Warning strings (exact, from spec)

- Validation: `warning: base_url is set but provider is empty — provider id required for local model routing`
- Validation: `warning: base_url "<value>" is not a valid http/https URL`
- Coalesce conflict: `warning: provider "<id>" declared with conflicting baseURLs — using first occurrence "<url>"`
- Claude guard: `warning: phase "<phase>" has base_url set but agent is "claude" — local endpoint ignored; claude agents do not support custom baseURLs`

## TUI baseURLEdit sub-mode (REQ-7)

- Add `baseURLEdit subMode` to the const block; reuse the shared `m.input` `textinput`.
- Enter `baseURLEdit` from `rowNav` on a new key (`u`), seeding `m.input` with the row's
  current `ref.BaseURL`. `hintLine`/`renderRow` gain a `baseURLEdit` case mirroring
  `freeForm` render.
- Enter commits `strings.TrimSpace(input)` to `m.rows[focused].ref.BaseURL`, sets
  `changed=true`, back to `rowNav`. Empty commit → `BaseURL=""` (clear). Escape cancels.
- Plain row display appends ` @ <baseURL>` when `ref.BaseURL != ""`.
- `applyToConfig` already assigns `row.ref` verbatim (struct value copy) → BaseURL
  persists with no extra code; the phase-empty `delete` guard stays keyed on `Model`.

## Testing Strategy

| REQ | Layer | Where | Approach |
|-----|-------|-------|----------|
| REQ-1 | Unit | `internal/config/model_test.go` | Table: scalar round-trip bytes; mapping when BaseURL set; scalar decodes BaseURL="" ; extend `TestConfig_CloneRoundtrip` fixture w/ BaseURL |
| REQ-2 | Unit | `cmd/archon/config_test.go` | set/get/list `.base_url`; provider/model untouched; unset→empty |
| REQ-3 | Unit | `internal/config/model_test.go` | `ValidateBaseURL` table: valid=no warn, ftp=warn, empty-provider=warn |
| REQ-4/5 | Golden+table | `internal/initcmd/opencode_mode_test.go` | Build a tmp `opencode.json`, assert JSON via `os.ReadFile`+`json.Unmarshal` (existing pattern, no testdata dir today); coalesce, mixed, conflict-warn (capture `bytes.Buffer`), preserve-user-provider, idempotent (write twice, compare bytes), no-BaseURL→no provider key |
| REQ-6 | Unit | `internal/initcmd/claude_mode_test.go` | Local ref → warn to `bytes.Buffer` + file still bare model; remote ref no warn; two locals two warns |
| REQ-7 | teatest | `internal/tui/models_tab_test.go` | `go-testing` teatest: open sub-mode, type+Enter→BaseURL set; clear+Enter→""; Escape→unchanged; render shows endpoint |

Follow `go-testing`: golden-style byte comparison for generated config, teatest for TUI.
Capture warnings via an injected `*bytes.Buffer`, not `os.Stderr`.

## PR-A vs PR-B boundary

**PR-A (REQ-1..5) — config core + opencode emission.** Independently mergeable; no PR-B
symbol referenced.

| File | PR-A scope |
|------|-----------|
| `internal/config/model.go` | BaseURL field, MarshalYAML switch, PhaseModel+ResolvePhaseModels, ValidateBaseURL |
| `internal/config/config_test.go` | Clone roundtrip fixture |
| `internal/config/model_test.go` | REQ-1, REQ-3 tests |
| `cmd/archon/config.go` + `_test.go` | REQ-2 set/get/list |
| `internal/initcmd/opencode_mode.go` + `_test.go` | REQ-4/5 provider block, io.Writer param |
| `internal/initcmd/init.go` | pass `os.Stderr` to opencode writer |
| `internal/tui/model.go` | pass writer to `MergeOpencodeAgent` (signature change) |

**PR-B (REQ-6,7) — Claude guard + TUI.** Builds on PR-A's `PhaseModel.BaseURL` and the
`io.Writer` signature convention.

| File | PR-B scope |
|------|-----------|
| `internal/initcmd/claude_mode.go` + `_test.go` | REQ-6 warn-and-skip, io.Writer param |
| `internal/initcmd/init.go` | pass `os.Stderr` to claude writer |
| `internal/tui/models_tab.go` + `_test.go` | REQ-7 baseURLEdit sub-mode |
| `internal/tui/model.go` | pass writer to `WriteClaudeAgents` |

**Boundary risk — `internal/tui/model.go` and `init.go` are touched by BOTH PRs**
(each adds a writer arg to a different function). PR-B rebases cleanly on PR-A since the
edits are on adjacent, non-overlapping lines (line 372 opencode in PR-A, line 381 claude
in PR-B). No shared symbol conflict.

### Line re-estimate vs 400 budget

| PR | Code | Tests | Total |
|----|------|-------|-------|
| PR-A | ~120 (field+marshal+PhaseModel+CLI+buildProviderBlock+wiring) | ~180 (marshal, validate, CLI, 8 opencode scenarios) | **~300** |
| PR-B | ~90 (claude guard + TUI sub-mode/render/keys) | ~130 (3 claude + 5 teatest) | **~220** |

Both slices sit under 400. PR-A is the closer call (~300, golden scenarios dominate); if
the opencode scenario tests balloon past 400, split the CLI (REQ-2) into a PR-A0. Flag at
the tasks gate but no pre-emptive third slice.

## Migration / Rollout

No migration. Additive behind a set BaseURL; refs without one emit byte-identical
`opencode.json`/YAML. Rollback = revert PRs or unset BaseURL.

## Open Questions

- [ ] `MergeOpencodeAgent`/`WriteClaudeAgents` signature change ripples to the TUI save
  path (`model.go`) and any other external caller — confirm no third caller exists
  before tasks (grep showed only init + TUI). Low risk; the exported seams are the only
  cross-package entry.
- [ ] OpenCode V2 (`settings.baseURL`, `aisdk:` npm prefix) — target V1 per repo `$schema`;
  documented in code comment, not abstracted. Already accepted in proposal.
