# Tasks — tui-model-picker (Slice 3b)

Ordered. The models_tab rewrite + model.go wiring + test rewrites are one cohesive unit; apply
together (build won't be green mid-rewrite). Tag = spec scenario.

## P1 — models_tab.go rewrite (internal/tui/models_tab.go)
- [x] P1-1 Replace state: `subMode` enum (rowNav/providerSelect/modelSelect/freeForm), `rowKind`,
      `modelRow{label,kind,phase,ref,changed}`, new `modelsTabState{rows,focusedRow,mode,providers,
      available,providerCursor,modelCursor,pickedProvider,curModels,input,cacheWarning,leaderEnabled,width}`.
      Remove inputs grid / autoFillLocks / cycleStaticModel / phaseNames / catalog.
- [x] P1-2 `newModelsTabState(cfg, providers map[string]opencode.Provider, loadErr error)`: seed rows
      from cfg.Models (Default + PhaseOrder + Leader if opencode); `available = DetectAvailableProviders`;
      set `cacheWarning` only when loadErr != nil. [S5 seed, S7/S8 warning]
- [x] P1-3 `update()` dispatch + `updateRowNav` (Up/Down clamp; Enter→providerSelect, or free-form if
      available empty; `e`→free-form; else ignored). [S1,S3,S4,S8,nav]
- [x] P1-4 `updateProviderSelect` (Up/Down, Enter→modelSelect or free-form if no SDD models, Esc→rowNav). [S1,S3,S11]
- [x] P1-5 `updateModelSelect` (Up/Down, Enter→set ref via `refFromCacheKey` + changed + rowNav, Esc→providerSelect). [S1,S2,S3]
- [x] P1-6 `refFromCacheKey(provider,key)`: slashed key→{Model:key}; bare→{Provider,Model}. [S2]
- [x] P1-7 `updateFreeForm` (Enter→ParseModelRef+changed+rowNav, Esc→cancel, else textinput.Update) + `openFreeForm`. [S4,S6]
- [x] P1-8 `view()` + `renderRow` + `hintLine`: deterministic; inline provider/model cursor lists on the
      focused row; `(default)` dim for empty phase rows; focus style from agent_tab; inline cacheWarning
      (color 214); mode-specific hints. Drop the per-row Validate advisory + leader `/`-guard. [S7,S10, D2]
- [x] P1-9 `applyToConfig`: Default/Leader = row.ref; phase with empty Model→delete, else set. Untouched
      rows write seeded ref verbatim (legacy preservation). [S5,S6]
- [x] P1-10 `setWidth`/width handling preserved as today.

## P2 — model.go wiring (internal/tui/model.go)
- [x] P2-1 In `NewModel`: load cache once via `opencode.DefaultCachePath` + `LoadModelsOrEmpty`; store
      `providers`+`cacheErr` on `Model`; build tab via `newModelsTabState(cfg, providers, cacheErr)`.
- [x] P2-2 At `agentInitDoneMsg` rebuild (~182): `newModelsTabState(msg.cfg, m.providers, m.cacheErr)`.
- [x] P2-3 Add `opencode` import; remove now-unused `internal/models` import from model.go.
      Save path (saveConfig→applyToConfig→Save→regenerateTemplate→MergeOpencodeAgent) UNCHANGED.

## P3 — Tests
- [x] P3-1 Rewrite `internal/tui/models_tab_test.go`: helpers `prov`/`toolModel`; the 13 tests from design §7
      (S1..S10 + nav clamps + slashed-key + provider-no-models). Drive update() with synthetic tea.KeyMsg.
- [x] P3-2 `internal/tui/model_test.go`: delete `TestModelsTabState_AutoFill`, `TestModelsTabState_LockOnEdit`;
      rewrite `TestNewModel` default-seed assertion, `TestModelsTabState_ApplyToConfig`,
      `TestSaveConfig_FailureDoesNotMutateInMemoryConfig` set, and `TestModelsTab_LeaderWarningGuard`
      (assert leader-row presence for opencode / absence otherwise, not the ⚠). Fix any
      `newModelsTabState` call to the new 3-arg signature.

## P4 — Gates
- [x] P4-1 `gofmt -l` changed files empty.
- [x] P4-2 `go build ./...` clean.
- [x] P4-3 `go vet ./...` clean.
- [x] P4-4 `go test ./...` all green.

## Definition of Done
- Provider→model picker sets ModelRef (incl. opencode bare-key + slashed-key no double-prefix); Esc cancels;
  free-form always available + parses; untouched legacy refs preserved; cleared phase deleted; corrupt-cache
  inline warning (never stderr); absent-cache free-form only; leader picker opencode-only; lists sorted; nav clamps.
- Save/merge path unchanged; `internal/models/resolve.go` left intact (TUI no longer calls Resolve — noted for future cleanup).
- build/vet/test green.
