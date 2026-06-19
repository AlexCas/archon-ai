# Verify + Judge Report — Slice 3a (`opencode-provider-catalog`)

Branch: `feat/structured-models-tui-picker`
File under change: `internal/opencode/models.go` (+ `models_test.go`)
Method: independent read + adversarial reasoning; tests NOT trusted as proof.

## VERDICT: SHIP

Additive, pure, deterministic-per-spec slice. All gates pass, every spec scenario
maps to code and a test, no import cycle, scope is exactly the 3 funcs + const +
`sort` import. One LOW note (tie-break under duplicate Names) is spec-compliant.

## Gate output (exact)

```
go build ./...        -> BUILD_EXIT=0   (no output)
go vet ./...          -> VET_EXIT=0     (no output)
gofmt -l <2 files>    -> GOFMT_EXIT=0   (no output; both files formatted)
go clean -testcache && go test ./... -count=1 -> TEST_EXIT=0
  ok  cmd/archon, internal/agent, internal/config, internal/initcmd,
      internal/models, internal/opencode, internal/scaffold, internal/status,
      internal/tui, internal/version, skills   (all ok)
```

## Traceability: spec scenario -> code -> test

| Scenario | Code | Test | Status |
|---|---|---|---|
| provider WITH tool_call -> true | `hasToolCallModel` loop returns true on first `m.ToolCall` (models.go:89-96) | `TestHasToolCallModel` withTC | OK |
| provider with NO tool_call -> false | same loop falls through to `return false` | `TestHasToolCallModel` noTC + empty Provider | OK |
| filter keeps tool_call, sorted by Name, excludes non-tc | `FilterModelsForSDD` appends only `m.ToolCall`, `sort.Slice` by Name (models.go:100-109) | `TestFilterModelsForSDD` asserts [Alpha,Zeta], Mid excluded | OK |
| providers with tool_call available, sorted | `DetectAvailableProviders` OR + `sort.Strings` (models.go:115-124) | `TestDetectAvailableProviders/tool_call providers sorted` -> [requesty,zeta] | OK |
| opencode always included even w/o tool_call | `id == builtinProviderID` short-circuit (models.go:118) | `.../opencode always included, no-tool_call excluded` | OK |
| absent cache -> (empty map, nil) | `LoadModelsOrEmpty` `errors.Is(err, os.ErrNotExist)` (models.go:79-81) | `TestLoadModelsOrEmpty_Absent` (pre-existing) | OK — seam intact |
| corrupt cache -> error | `LoadModelsOrEmpty` propagates non-NotExist err (models.go:82) | `TestLoadModelsOrEmpty_ParseError` (testdata/invalid.json) | OK — seam intact |

No gaps. Every MUST in the spec has both an implementation site and a test.

## Adversarial judge findings

- **Determinism (FilterModelsForSDD):** sort key is `Name` only via `sort.Slice`
  (unstable). For the real cache where model Names are distinct this is fully
  deterministic and spec-compliant ("sorted ascending by Name"). If two models
  shared an identical `Name`, their relative order would be nondeterministic
  (map iteration + unstable sort). Spec does not define a secondary key, so this
  is NOT a defect — LOW note only. (See LOW-1.)
- **Determinism (DetectAvailableProviders):** `sort.Strings` over string IDs;
  IDs are unique map keys -> fully deterministic. OK.
- **nil/empty map:** both new funcs use `var x []T` and never append -> return
  `nil` slice for empty/nil input. Acceptable for callers (`range`/`len` safe).
- **opencode with zero tool_call models:** included — `id == builtinProviderID`
  short-circuits the OR before `hasToolCallModel`. Correct per spec.
- **Key vs p.ID:** `DetectAvailableProviders` keys off the MAP KEY `id`, not
  `p.ID`. `LoadModels` forces `p.ID = id` (models.go:66), so they are identical
  for loaded data; using the authoritative key is consistent and slightly safer.
  No off-by-one / wrong-key bug.
- **Import cycle:** confirmed `internal/opencode` does NOT import
  `internal/models` (only match is the explanatory comment, models.go:13). Local
  `const builtinProviderID = "opencode"` is used at models.go:118. No cycle.
- **Scope:** only additions are the const + `sort` import + 3 funcs
  (`hasToolCallModel`, `FilterModelsForSDD`, `DetectAvailableProviders`) plus a
  doc-comment expansion on `LoadModelsOrEmpty`. NO change to `Provider`/`Model`
  structs, NO behavior change to `LoadModels`/`LoadModelsOrEmpty`. Nothing else
  touched.
- **Go hygiene:** `sort` import is used (3 call sites); no unused imports, no
  shadowing, no vet complaints.

## LOW notes (non-blocking)

- **LOW-1** `FilterModelsForSDD` (models.go:107): unstable sort with a Name-only
  key. Harmless given unique Names; if a future cache can carry duplicate Names
  and the TUI needs a stable order, add `ID` as a tie-break:
  `return out[i].Name < out[j].Name || (out[i].Name == out[j].Name && out[i].ID < out[j].ID)`.
  Not required by this slice's spec.
- **LOW-2** Both helpers return `nil` (not `[]T{}`) for empty input. Fine for the
  described callers; just noting for any caller that distinguishes nil from empty.
