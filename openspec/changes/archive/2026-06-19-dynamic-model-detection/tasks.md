# Tasks: Dynamic Model Detection (Slice 3)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~320–380 (PR1 ~190, PR2 ~150) |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Suggested split | PR1 (engine) → PR2 (TUI wiring) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | `internal/models/` detector + lister + parser + resolver + tests (no UI) | PR1 | Base = master OR Slice-1 branch (decided at apply) |
| 2 | TUI consumes resolver: cache at open, hint rename, cycle on cached catalog + tests | PR2 | Base = PR1 branch (stacked) |

Catalog-agnostic: resolver composes with Slice 1's Gemini/OpenAI lists once merged, no edits. `internal/config/model.go` is UNCHANGED.

## Phase 1: Engine — Detector (PR1)

- [x] 1.1 Create `internal/models/detect.go`: `type CLIDetector func() map[string]bool` and `LookPathDetector()` iterating `opencode, claude, codex, gemini` via `exec.LookPath`.

## Phase 2: Engine — Lister & Parser (PR1)

- [x] 2.1 Create `internal/models/opencode.go`: define `OpencodeLister` interface (`List(ctx) ([]string, error)`) and `execLister` running `exec.CommandContext(ctx, "opencode", "models", "opencode-go")`.
- [x] 2.2 In `opencode.go`, add `parseModels([]byte) []string`: split lines, trim, skip blanks, strip `provider/` prefix.

## Phase 3: Engine — Resolver (PR1)

- [x] 3.1 Create `internal/models/resolve.go`: `ResolveModels(detect CLIDetector, lister OpencodeLister) []string` — claude→`config.ClaudeModels`; opencode→live `List` with 2s `context.WithTimeout`, fallback to `config.OpencodeModels` on err/empty.
- [x] 3.2 In `resolve.go`, add `Resolve()` convenience = `ResolveModels(LookPathDetector, execLister{})`.

## Phase 4: Engine Tests (PR1)

- [x] 4.1 `internal/models/resolve_test.go`, "Installed opencode shows the live catalog": detector{opencode:true} + fake lister live list → output has live, not curated.
- [x] 4.2 Same file, "Only installed agents' models are offered": detector{opencode:false} → no opencode models; present-agent models remain.
- [x] 4.3 Same file, "Live enumeration error falls back silently": detector{opencode:true} + lister err/empty → output == `config.OpencodeModels`.
- [x] 4.4 `internal/models/opencode_test.go`: `parseModels` table test (prefixed, blank, no-slash lines).
- [x] 4.5 Verify PR1: `go build ./...`, `go test`/`go vet ./internal/models/...` (no subprocess).

## Phase 5: TUI Wiring (PR2, stacked on PR1)

- [x] 5.1 `internal/tui/models_tab.go`: add `catalog []string` field to `modelsTabState`; change `newModelsTabState(cfg, catalog []string)`; store on state.
- [x] 5.2 Same file, `cycleStaticModel`: build cycle list from `m.catalog` (empty-lead) instead of `config.StaticModels()`.
- [x] 5.3 Same file, `view`: hint reads `m.catalog`; rename label "Static:" → "Available:".
- [x] 5.4 `internal/tui/model.go`: compute `models.Resolve()` once at open; pass into BOTH `newModelsTabState` call sites (~line 88, ~line 175).

## Phase 6: TUI Tests (PR2)

- [x] 6.1 `internal/tui/models_tab_test.go`, "Detection is cached once per Models view": resolver runs once at open; cycle/type reads `m.catalog`; call-counter fake not re-invoked; never at init.
- [x] 6.2 Same file: cycling and "Available:" hint render from injected catalog slice.
- [x] 6.3 Same file, "Free-form entry and advisory behavior unchanged": arbitrary value accepted; `NormalizeModel`/`Validate` intact.
- [x] 6.4 Verify PR2: `go build ./...`, `go test`/`go vet ./internal/tui/...`.
