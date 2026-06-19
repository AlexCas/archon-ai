# Proposal — structured-models-tui-picker, Slice 3a (provider catalog data layer)

**Slice 3 split into 3a (this PR) + 3b (later).** Part of `structured-model-resolution`.
Foundation (#47) merged; Slice 2 writers (#48) open; this branch stacks on Slice 2.

## Problem
The TUI models tab offers a flat free-form catalog (`models.Resolve()` → `[]string`, scoped to
the opencode provider). To build a provider→model PICKER (Slice 3b), the TUI needs a
PROVIDER-KEYED view of the cache: which providers are usable for SDD and which of their models
support tool_call. archon's `internal/opencode` has the cache reader but none of these query
helpers — they exist only in gentle-ai.

## Goal (3a)
Add the provider/model query helpers to `internal/opencode` (the data layer the 3b picker will
consume), with no UI behavior change. Pure, additive, fully tested.

## Scope
### In (3a)
- `hasToolCallModel(Provider) bool` — provider has ≥1 tool_call model.
- `FilterModelsForSDD(Provider) []Model` — the provider's tool_call models, sorted by Name.
- `DetectAvailableProviders(map[string]Provider) []string` — SIMPLIFIED: provider IDs that have
  ≥1 tool_call model OR are the built-in `opencode` provider; sorted. (No auth.json/env detection,
  no `Env` field added to `Provider`, no custom-provider args — deferred/unneeded.)
- Tests against the existing `testdata/` fixtures (well-formed, malformed-entry, invalid JSON).
- Confirm + test the corrupt-vs-absent seam already in `LoadModelsOrEmpty` (absent→empty,nil;
  corrupt→propagated err) — this is what 3b's inline warning will key off.

### Out (3b and beyond)
- The TUI two-step picker, the inline corrupt-cache warning DISPLAY, legacy bare-alias preservation,
  free-form fallback, and the `models_tab.go`/`models_tab_test.go` rewrite.
- Full auth/env provider detection; effort/variants; custom providers; cost/limit; OpenRouter fix.

## Affected files
- `internal/opencode/models.go` (+ `models_test.go`). No other package touched.

## New capability to spec
- `opencode-provider-catalog` — provider/model query helpers over the opencode cache.

## Size / PR forecast
~60-100 LOC (prod ~35-45 + tests ~40-60). Well under D1 400. Stacks on Slice 2 (#48); PR base =
`feat/structured-models-writers`, retarget to master after #48 merges.

## Risks
- Looks like dead code until 3b consumes it → mitigated: it's small, pure, fully tested, and the
  explore documents the 3b consumer.
- Determinism: map iteration is unordered → all outputs sorted (provider IDs, model lists).
