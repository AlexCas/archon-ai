# Verify + Judge Report — Slice 3b (`tui-model-picker`)

Branch: `feat/structured-models-picker-ui` · Scope: in-tab provider→model picker rewrite (uncommitted).
Files: `internal/tui/models_tab.go`, `model.go`, `models_tab_test.go`, `model_test.go`, `e2e_test.go`.

## VERDICT: SHIP

No correctness defects found. All gates pass. Every spec scenario (S1–S11 + nav clamp) maps to
production code and at least one passing test. Two LOW notes and one pre-existing known-limitation
(out of this slice's scope) are recorded below.

---

## PART 1 — Gates (exact output)

| Gate | Result |
|------|--------|
| `go build ./...` | clean, exit 0 |
| `go vet ./...` | clean, exit 0 |
| `gofmt -l` (5 changed files) | empty (all formatted) |
| `go clean -testcache && go test ./... -count=1` | all packages `ok`, exit 0 |

Test output (all packages): `cmd/archon`, `internal/agent`, `internal/config`, `internal/initcmd`,
`internal/models`, `internal/opencode`, `internal/scaffold`, `internal/status`, `internal/tui`
(1.262s), `internal/version`, `skills` — all `ok`.

## PART 2 — Spec traceability

| Scenario | Code | Test | Status |
|----------|------|------|--------|
| S1 pick sets ref | `updateProviderSelect`/`updateModelSelect` (215-216) | `TestModelsTab_PickProviderModelSetsRef` | PASS |
| S2 opencode bare | `refFromCacheKey` else-branch (233) | `TestModelsTab_OpencodeBareKeyMapsToRef` | PASS |
| S2 slashed no double-prefix | `refFromCacheKey` "/" short-circuit (230-232) + `FullID` (25-27) | `TestModelsTab_SlashedKeyNoDoublePrefix` | PASS |
| S3 Esc steps back/cancels | `updateModelSelect` Esc→providerSelect (218-219); `updateProviderSelect` Esc→rowNav (198-199) | `TestModelsTab_EscCancelsPicker`, `TestModelsTab_EscFromModelSelectGoesBack` | PASS |
| S4 free-form parses + Esc | `updateFreeForm` (236-252) | `TestModelsTab_FreeFormTogglesAndParses`, `TestModelsTab_FreeFormEscCancels` | PASS |
| S5 legacy untouched preserved | `applyToConfig` writes seeded ref verbatim (367-386) | `TestModelsTab_LegacyUntouchedPreserved` | PASS |
| S6 cleared phase deleted | `applyToConfig` empty-Model delete (378-380) | `TestModelsTab_ApplyDeletesClearedPhase` | PASS |
| S7 corrupt warning inline | `cacheWarning` set on loadErr (112-114); rendered in `view` (283-287) | `TestModelsTab_CorruptCacheWarningShown` | PASS |
| S8 absent→free-form, no warning | rowNav Enter empty-available→openFreeForm (156-159) | `TestModelsTab_AbsentCacheFreeFormOnly` | PASS |
| S9 leader opencode-only | `leaderEnabled` gate (80,91-93); apply leaves Leader for non-OC (no rowLeader case fires) | `TestModelsTab_LeaderUsesPicker`, `TestModelsTab_LeaderWarningGuard` | PASS |
| S10 sorted + nav clamp | `DetectAvailableProviders`/`FilterModelsForSDD` sorted; rowNav clamps (146-153) | `TestModelsTab_DeterministicSortedLists`, `TestModelsTab_NavigationClamps` | PASS |
| S11 provider no-SDD-models→free-form | `updateProviderSelect` len(curModels)==0→openFreeForm (192-195) | `TestModelsTab_ProviderNoSDDModelsFallsBackToFreeForm` | PASS |

No spec gaps.

## PART 2 — Adversarial findings (all cleared)

- **State machine**: re-entering the picker on a different row resets `providerCursor=0`;
  `pickedProvider`/`curModels` are recomputed on every providerSelect→modelSelect Enter, so no
  stale cross-row leak. Esc modelSelect→providerSelect preserves `providerCursor` (not reset) —
  the chosen provider stays highlighted. `focusedRow` is never mutated by a picker transition;
  the single-open-row invariant holds. `modelCursor` reset to 0 on each modelSelect entry.
- **Picked item integrity**: `view` and the Enter handler iterate `curModels` (sorted by Name)
  with the same cursor and use `.ID` for the ref — displayed item == picked item.
- **Legacy preservation**: rows seeded from `cfg.Models.{Default,Phases[phase],Leader}`; untouched
  rows (`changed==false`) written verbatim — a `{Provider:"",Model:"opus"}` default round-trips
  (provider stays empty, never guessed). Untouched empty phase row → `delete` (no empty entry
  created; safe on freshly-made map). Non-opencode leader: no rowLeader row, `applyToConfig` never
  touches `cfg.Models.Leader`, and `saveConfig` clones config first so the loaded value survives.
- **refFromCacheKey**: bare key under any provider → `{Provider,Model}`; any key containing "/" →
  `{Model:key}` so `FullID`'s "/" short-circuit returns it verbatim. No mislabel path.
- **Cache wiring**: `NewModel` loads once via `DefaultCachePath`+`LoadModelsOrEmpty`, stores
  `providers`+`cacheErr` on `Model`; `agentInitDoneMsg` reuses them (no re-read). `LoadModelsOrEmpty`
  returns `(nil,err)` on corrupt, `(empty,nil)` on absent — so when `cacheErr!=nil`, `providers` is
  nil (→ converted to empty map, `available` empty). No path sets cacheErr with non-empty providers.
  `DefaultCachePath` error sets `cacheErr` sensibly (providers nil → empty).
- **TUI safety**: no `os.Stderr`/`fmt.Print`/`Fprint`/`println` added in changed code (grep clean).
  Warning is inline in `view()` only.
- **Determinism**: `available` sorted (DetectAvailableProviders), model lists sorted by Name
  (FilterModelsForSDD), rows in fixed `[Default, PhaseOrder…, Leader]` order, `view()` has no
  map-range. Reproducible.
- **Parent routing**: global `tab/shift+tab/ctrl+s/ctrl+q/q` matched in `Model.Update` with early
  `return` BEFORE tab routing (unchanged by this diff) — never reach `update()`. rowNav consumes
  only Up/Down/Enter/`e`; everything else returns `nil,true`. `e` fires only in rowNav; in freeForm
  it is a literal routed to the textinput.
- **Scope**: save path unchanged; `internal/models/resolve.go` byte-identical (git diff empty);
  `models.Resolve()` no longer called by the TUI (only in `internal/models` self + tests). No 3a /
  foundation changes. Removed identifiers (`inputs`, `autoFillLocks`, `updateAutoFill`,
  `cycleStaticModel`, `focusedInput`, `phaseNames`, `modelInput*`, `leaderInputIndex`, `.catalog`)
  fully gone — grep clean across the package.
- **Go safety**: `Phases` map nil-guarded before write (368-369); cursor bounds clamped against
  `len(available)`/`len(curModels)`/`len(rows)`; textinput Focus on open, Blur on Enter/Esc (no
  leak); no unused imports (build/vet clean).

## LOW notes (non-blocking)

1. **Dead code**: `sortedProviderIDs` (`models_tab.go:393-402`) is defined and documented as "Used
   internally to assert determinism in tests" but is referenced nowhere (no test, no prod). Build
   passes (package-level funcs may be unused). Recommend deleting or wiring into a determinism test.
2. **Weakened-but-justified test**: `TestEdgeCases_UnknownModelWarning` (`e2e_test.go:98`) no longer
   asserts a warning — it only checks the view renders without panicking. This is the intended D2
   consequence (per-row advisory dropped). The test NAME still says "Warning"; consider renaming to
   reflect it now only guards against panic.

## Known limitation (pre-existing, out of scope)

- **Free-form cannot contain `q`**: the parent's global key switch (`model.go:131`, UNCHANGED by
  this diff) matches `q`→quit before routing to the tab, so while in freeForm, typing `q` quits the
  TUI instead of inserting the character. Real model names contain `q` (e.g. `qwen`, `grok` is fine
  but `qwen2.5` is not typeable). This is pre-existing parent routing (the old free-form grid had the
  same constraint) and parent-routing changes are explicitly out of slice 3b's scope (design R1).
  Recommend a follow-up: in freeForm, scope/disable the global single-letter `q` quit (keep `ctrl+q`).
