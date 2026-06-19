# Exploration — structured-models-tui-picker (Slice 3)

**Goal**: replace the free-form text model inputs in the TUI models tab with a TWO-STEP
provider→model picker driven by the structured opencode cache, populating `config.ModelRef`
directly; plus a TUI-safe in-UI warning when the opencode cache is present but unreadable.

**Branch**: `feat/structured-models-tui-picker` (stacked on Slice 2 @ 7050780; PR #48 open)
**Date**: 2026-06-19

## Current tab mechanics (internal/tui/models_tab.go, 310 LOC)
- `modelsTabState{inputs []textinput.Model, focusedInput, phaseNames, autoFillLocks, leaderEnabled, catalog []string}`.
- Rows: default(0) + 8 phases(1..8) + opencode-only leader (appended; `leaderInputIndex()`).
- Reads config via `ModelRef.FullID()` strings (newModelsTabState:50,58,69); writes via
  `config.ParseModelRef(value)` (applyToConfig:283,289-297,302).
- `cycleStaticModel` cycles a flat `catalog []string` from `models.Resolve()`; view shows
  `"Available: "+join(catalog)` (212) + inline `config.Validate` advisory.
- Save path UNCHANGED & structured: applyToConfig→cfg.Models→Save→regenerateTemplate
  (ResolvePhaseModels) + MergeOpencodeAgent(cfg.Models). Slice 3 changes ONLY how the tab
  POPULATES cfg.Models.

## Cache API: present vs needed
- Present (`internal/opencode/models.go`): `DefaultCachePath`, `LoadModels`, `LoadModelsOrEmpty`
  (absent→empty+nil; corrupt top-level→propagates err; malformed provider entry→skipped silently),
  `Provider{ID,Name,Models}`, `Model{ID,Name,ToolCall,Reasoning}`.
- MUST PORT from gentle-ai (`../gentle-ai/internal/opencode/models.go:186-267`):
  `FilterModelsForSDD` (ToolCall filter + sort, trivial), `hasToolCallModel` (trivial),
  `DetectAvailableProviders` (decision: full auth.json/env port vs simplified "providers with
  ≥1 tool_call model + always-opencode"; archon's Provider lacks an `Env` field → full port needs
  a struct change). NOT needed: variants/effort, custom providers, cost/limit, OpenRouter fix.

## Corrupt-vs-absent warning seam (TUI-safe)
- `LoadModelsOrEmpty`: absent (ErrNotExist)→(empty,nil)=NO warning; corrupt top-level JSON→(nil,err)=WARN.
- The picker must call `LoadModelsOrEmpty(DefaultCachePath())` DIRECTLY (not the flat `ResolveModels`,
  which collapses to []string and swallows the error). Store the warning string in the tab state;
  render INLINE in view() (color 214), like the existing `⚠` advisory. NEVER stderr (Resolve()/cache
  read is TUI-only → stderr corrupts Bubbletea).

## Widgets / patterns
- `bubbles/list` NOT used anywhere. Convention = HAND-ROLLED cursor selection
  (`agent_tab.go:83-99` focusedIndex + Up/Down/Enter + lipgloss focus) — reuse this, not bubbles/list.
- Warnings: inline tab text (mechanism b) — correct for construction-time cache reads.

## Back-compat (correctness requirements, not optional)
- Legacy bare-alias `ModelRef{Provider:""}` (e.g. `opus`): picker has no provider to highlight.
  MUST display the loaded FullID as a keepable "current value" and preserve it verbatim if untouched
  — a pure overwrite-on-pick would corrupt legacy aliases.
- Empty/absent cache (no providers): the free-form `textinput` path MUST survive as a fallback
  (also the escape hatch for models not in the cache).
- opencode-provider key asymmetry: bare `opencode` keys → ModelRef{opencode, <key>}; already-slashed
  keys (requesty/xai/grok-4) map so FullID doesn't double-prefix (FullID already guards on "/").

## SLICING — recommend 3a / 3b (one PR would exceed ~400 LOC + a large test rewrite)
- **Slice 3a — cache-backed provider/model data layer (additive, no UI change).** Port
  `DetectAvailableProviders` (simplified or full — decision), `FilterModelsForSDD`, `hasToolCallModel`
  into `internal/opencode` + the corrupt-vs-absent classification helper + tests against existing
  testdata. No `models_tab.go` behavior change. Cut line = the `internal/opencode` ↔ `internal/tui`
  boundary. Easily under 400 LOC, low risk.
- **Slice 3b — the TUI two-step picker + inline warning + legacy preservation + free-form fallback +
  test rewrite.** The larger, riskier half; own focused review. Reuses agent_tab's cursor pattern.

## Key decisions for design/spec
1. Slicing: 3a now, 3b next (recommended) vs one PR.
2. `DetectAvailableProviders` fidelity: simplified (tool_call + always-opencode) vs full auth/env
   (needs adding `Env` to opencode.Provider). Recommend SIMPLIFIED for 3a.
3. Picker hosting (3b): in-tab per-row sub-mode (archon-style) vs full-screen sub-state.
4. Legacy bare-alias preservation contract (3b) — keep verbatim if untouched.
5. Free-form fallback trigger (3b) — always available escape hatch.
6. Leader row (3b): picker vs stay free-form (accepts any provider/model, Validate already guards `/`).
7. Effort: OUT of scope (archon Model has no Variants).

## Key files
- `internal/tui/models_tab.go`; `internal/tui/model.go:88,93,182,311-355`
- `internal/opencode/models.go`; `internal/models/resolve.go` (flat catalog, bypassed by picker)
- `internal/config/model.go:14-78,250-276`; `internal/initcmd/opencode_mode.go:44`
- `internal/tui/agent_tab.go:83-99` (cursor pattern); tests: `models_tab_test.go`, `model_test.go:250-326,675-735`
- Port ref: `../gentle-ai/internal/opencode/models.go:186-267`, `../gentle-ai/internal/tui/screens/model_picker.go`
