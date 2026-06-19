# Proposal — tui-model-picker (Slice 3b)

**Final slice of the structured-model-resolution TUI work.** Builds on 3a (#49, provider catalog
helpers) and the merged Foundation. Stacks on 3a; base retargets to master after #48/#49 merge.

## Problem
The TUI models tab still uses free-form text inputs and a flat `models.Resolve()` catalog. Users
must type a `provider/model` string by hand, with no awareness of which providers/models opencode
actually offers. The structured foundation + 3a catalog helpers now make a guided picker possible.

## Goal
Replace the free-form rows with an in-tab, per-row two-step provider→model picker driven by the
opencode cache (via 3a's `DetectAvailableProviders`/`FilterModelsForSDD`), populating
`config.ModelRef{Provider, Model}` directly — while preserving legacy values and a free-form escape
hatch, and warning (in-UI) when the cache is corrupt.

## Scope
### In
- In-tab per-row SUB-MODE picker (reuse `agent_tab` cursor pattern): focus a row, Enter → provider
  list → model list (`FilterModelsForSDD`), pick sets `ModelRef{Provider, Model}`. Esc cancels.
- Applies to all rows: default + 8 phases + the opencode-only leader.
- FREE-FORM always available: a key toggles free-form text entry on any row (escape hatch for models
  not in the cache); free-form value parsed via `ParseModelRef`.
- LEGACY PRESERVATION: a row whose current `ModelRef` is a legacy bare alias (Provider=="") or any
  value the user does NOT re-pick is kept verbatim in `applyToConfig`.
- CORRUPT-CACHE WARNING: inline in the tab view when `LoadModelsOrEmpty` returns an error (cache
  present but unreadable); absent cache → no warning, picker shows empty → free-form only.
- Rewrite `models_tab.go` (state/update/view/applyToConfig) + its tests.

### Out
- Effort/variants selection (archon `Model` has no Variants) — Slice 4.
- Any change to the save/merge path (applyToConfig→cfg.Models→Save→MergeOpencodeAgent+ResolvePhaseModels
  is already structured), to 3a helpers, or to the cache reader.

## Settled decisions (from user)
1. In-tab per-row sub-mode (not full-screen). 2. Leader also uses the picker. 3. Free-form always
available. 4. Legacy bare-alias preserved verbatim if untouched. 5. Inline corrupt-cache warning.

## Affected files
- `internal/tui/models_tab.go` (rewrite) + `internal/tui/models_tab_test.go` (rewrite).
- `internal/tui/model.go` (catalog wiring at 88/93/182 — switch from flat `models.Resolve()` to the
  provider-keyed cache load for the tab; save path unchanged).
- Possibly `internal/tui/model_test.go` (input-driving tests around 250-326, 719-735).

## New capability to spec
- `tui-model-picker` — the in-tab provider→model picker, free-form fallback, legacy preservation,
  and inline corrupt-cache warning.

## Size / PR forecast
~300-400 LOC (prod ~180-240 + test rewrite ~120-180). At/near the D1 400 budget — flag at PR (C1).
Reuses the agent_tab cursor pattern to keep it lean; no bubbles/list.

## Risks
- Largest, UI-heavy slice → mitigated by reusing the existing hand-rolled cursor pattern + a
  thorough apply-ready design + full test rewrite.
- Legacy corruption (the main hazard) → explicit preservation requirement + a dedicated test.
- Determinism → providers/models sorted (3a already sorts); deterministic view rendering.
- TUI-safety → warning rendered inline, never stderr.
