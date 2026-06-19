# Design — structured-models-tui-picker, Slice 3a (provider catalog data layer)

Apply-ready. Adds 3 pure helpers to `internal/opencode/models.go`, simplified per user decision
(no auth/env, no `Provider.Env` field). Settled by `specs/opencode-provider-catalog/spec.md`.

## Functions to add (internal/opencode/models.go)

```go
// hasToolCallModel reports whether the provider has at least one tool_call model.
func hasToolCallModel(p Provider) bool {
	for _, m := range p.Models {
		if m.ToolCall {
			return true
		}
	}
	return false
}

// FilterModelsForSDD returns the provider's tool_call-capable models, sorted by Name.
// Non-tool_call models are excluded (SDD phases require tool calling).
func FilterModelsForSDD(p Provider) []Model {
	var out []Model
	for _, m := range p.Models {
		if m.ToolCall {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// DetectAvailableProviders returns the provider IDs usable for SDD, sorted. A provider
// qualifies when it has at least one tool_call model, OR its ID is "opencode" (the built-in
// provider is always offered when present). Simplified: no auth.json / env-var detection.
func DetectAvailableProviders(providers map[string]Provider) []string {
	var available []string
	for id, p := range providers {
		if id == opencodeProviderID || hasToolCallModel(p) {
			available = append(available, id)
		}
	}
	sort.Strings(available)
	return available
}
```

Notes:
- `opencodeProviderID` const ("opencode") currently lives in `internal/models/resolve.go`. To
  avoid an import cycle / duplication, define a local const in `opencode/models.go`
  (e.g. `const builtinProviderID = "opencode"`) — `internal/opencode` must NOT import
  `internal/models`. Do not reuse resolve.go's const across packages.
- `sort` is already imported by the package (LoadModels sorts? verify; if not, add the import).
- `Model.Name` and `Model.ToolCall` already exist (Model{ID,Name,ToolCall,Reasoning}).
- These are unexported-consumable by `internal/tui` only if exported: `FilterModelsForSDD` and
  `DetectAvailableProviders` are EXPORTED (3b/tui consumes them); `hasToolCallModel` stays
  unexported (internal helper).

## Corrupt-vs-absent seam (no new code; confirm + test)

`LoadModelsOrEmpty(path)` already: absent (os.IsNotExist) → `(empty map, nil)`; corrupt →
propagated error (via `LoadModels`). Slice 3b's inline warning keys off the returned error. 3a
adds/keeps explicit tests for both branches (absent and corrupt). No behavior change.

## Test plan (internal/opencode/models_test.go)

Reuse existing `testdata/`:
- `testdata/models.json` (well-formed: opencode + requesty/slashed),
- `testdata/malformed.json` (good provider + a broken entry),
- `testdata/invalid.json` (truncated JSON).

Add:
- `TestHasToolCallModel` — true when a model has tool_call; false when none.
- `TestFilterModelsForSDD` — only tool_call models, sorted by Name; a non-tool_call excluded.
  (If a fixture lacks a non-tool_call model, build the Provider in-memory.)
- `TestDetectAvailableProviders` — sorted IDs; a provider with a tool_call model included; a
  non-opencode provider with no tool_call model excluded; `opencode` included even with no
  tool_call model (in-memory Provider map).
- Confirm/extend `TestLoadModelsOrEmpty_*`: absent path → empty+nil; `invalid.json` → error.
  (If these already exist from Slice 1, just ensure both branches are covered.)

## Determinism
All outputs sorted (provider IDs via sort.Strings; models via sort.Slice on Name). No map-order
leakage.

## Size
Prod ~30-40 LOC (3 funcs + maybe a const/import). Tests ~50-70 LOC. Total ~80-110 LOC. Under D1 400.

## Out of scope (3b)
TUI picker, inline warning display, legacy preservation, free-form fallback, models_tab rewrite.
