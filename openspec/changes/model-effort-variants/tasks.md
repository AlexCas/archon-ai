# Tasks — model-effort-variants (Slice 4, option b)

## E1 — config (internal/config/model.go)
- [x] E1-1 Add `Effort string` to `PhaseModel`. [effort propagation]
- [x] E1-2 `ResolvePhaseModels` emit `Effort: ref.Effort`. [effort propagation]
- [x] E1-3 `model_test.go`: assert PhaseModel.Effort carried (phase-effort + default-fallback effort).

## E2 — writer (internal/initcmd/opencode_mode.go)
- [x] E2-1 Add `Variant string json:"variant,omitempty"` after Model in archonLeaderAgent + archonPhaseAgent.
- [x] E2-2 Populate leader `Variant: models.Leader.Effort`; phase `Variant: pm.Effort`.
- [x] E2-3 `opencode_mode_test.go`: variant present when effort set; key absent when empty; re-run byte-identical (idempotency).

## E3 — picker (internal/tui/models_tab.go)
- [x] E3-1 Add `effortSelect` to subMode enum + `effortCursor int` + `effortOptions` (default/low/medium/high; default→"").
- [x] E3-2 `updateModelSelect` Enter: build ref + changed; if picked model `.Reasoning` → effortSelect (cursor 0), else rowNav (effort empty).
- [x] E3-3 Add `updateEffortSelect` (Up/Down, Enter→set ref.Effort + rowNav, Esc→modelSelect). Wire into update() dispatch.
- [x] E3-4 Render effortSelect in view()/renderRow (cursor list, "Effort:" header, hint). 
- [x] E3-5 `models_tab_test.go`: reasoning model → effortSelect; "high"→Effort "high"; "default"→""; non-reasoning skips; Esc→modelSelect. Add `reasoningModel` helper.

## E4 — Gates
- [x] E4-1 `gofmt -l` changed files empty.
- [x] E4-2 `go build ./...` clean.
- [x] E4-3 `go vet ./...` clean.
- [x] E4-4 `go test ./...` all green.

## Definition of Done
- Reasoning models offer default/low/medium/high; non-reasoning skip; Effort set on the ModelRef.
- variant written to opencode.json (omitempty) for leader + phases; effortless output byte-identical to today; re-run idempotent.
- Effort round-trips in config (mapping form, already supported).
- build/vet/test green; no plugin/cache/embed; free-form leaves Effort empty.
