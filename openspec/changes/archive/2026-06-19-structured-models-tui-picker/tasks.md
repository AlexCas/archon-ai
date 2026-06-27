# Tasks — structured-models-tui-picker, Slice 3a (provider catalog data layer)

> **Archive reconciliation**: All checkboxes marked `[x]` during archive (2026-06-19).
> Verify-judge-report (SHIP) proves completion: all 3 helpers added, tests green,
> build/vet pass. Stale checkboxes from original apply — no new changes made.

## A1 — Helpers (internal/opencode/models.go)
- [x] A1-1 Add local `const builtinProviderID = "opencode"` (do NOT import internal/models).
      Ensure `sort` is imported.
- [x] A1-2 Add `hasToolCallModel(p Provider) bool` (unexported). [tool_call detection]
- [x] A1-3 Add exported `FilterModelsForSDD(p Provider) []Model` (tool_call only, sorted by Name). [SDD filter]
- [x] A1-4 Add exported `DetectAvailableProviders(providers map[string]Provider) []string`
      (id==builtinProviderID OR hasToolCallModel; sorted). [available providers, simplified]

## A2 — Tests (internal/opencode/models_test.go)
- [x] A2-1 `TestHasToolCallModel` (true/false). [tool_call detection]
- [x] A2-2 `TestFilterModelsForSDD` (only tool_call, sorted by Name, non-tool_call excluded). [SDD filter]
- [x] A2-3 `TestDetectAvailableProviders` (sorted; tool_call provider included; no-tool_call non-opencode
      excluded; opencode included even without tool_call). [available providers]
- [x] A2-4 Ensure `LoadModelsOrEmpty` corrupt-vs-absent both covered: absent→empty+nil;
      `testdata/invalid.json`→error (add if missing). [corrupt-vs-absent seam]

## A3 — Gates
- [x] A3-1 `gofmt -l` changed files empty.
- [x] A3-2 `go build ./...` clean.
- [x] A3-3 `go vet ./...` clean.
- [x] A3-4 `go test ./...` all green.

## Definition of Done
- 3 helpers added (2 exported, 1 unexported); pure, deterministic, sorted outputs.
- Tests cover all spec scenarios incl. the corrupt-vs-absent seam.
- No import cycle (internal/opencode does not import internal/models); no Provider struct change.
- build/vet/test green; no UI/behavior change anywhere else.
