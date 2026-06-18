# Tasks: Phase Model Propagation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 120–200 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR (3 internal work units) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Config core: `PhaseOrder`/`PhaseModel`/`NormalizeModel`/`ResolvePhaseModels` + extended `Validate` + unit tests | PR 1 | Standalone; no callers yet. base = master |
| 2 | Template rendering: `TemplateData.PhaseModels` field + guarded `## Phase Models` block + render tests | PR 1 | Depends on unit 1 types |
| 3 | Plumbing: `writeTemplate` signature + caller + TUI line; cross-path identity test | PR 1 | Depends on units 1–2 |

Single PR (C1 = ask-always, but Low risk under 400 lines → no chain decision needed).

## Phase 1: Config core (Unit 1)

- [x] 1.1 `internal/config/model.go`: add `var PhaseOrder = []string{"explore","propose","spec","design","tasks","apply","verify","archive"}` and `type PhaseModel struct{ Phase, Model string }`.
- [x] 1.2 `internal/config/model.go`: add `NormalizeModel(s string) (id string, ok bool)` — trim+lowercase; exact alias match; else first family substring `opus`/`sonnet`/`haiku`; else `("",false)` (per Decision 1 table).
- [x] 1.3 `internal/config/model.go`: add `ResolvePhaseModels(mc ModelConfig) []PhaseModel` — iterate `PhaseOrder`, normalize `Phases[p]`→fallback `Default`→omit; pure, no mutation (Decision 3).
- [x] 1.4 `internal/config/model.go`: extend `Validate(model string)` to also return `""` when `NormalizeModel` resolves (Decision 2 snippet).
- [x] 1.5 `internal/config/model_test.go`: table tests for `NormalizeModel` — typo `Opues 4.8`→`ok=false`, `claude-haiku-4-5-20251001`→`haiku`, idempotent `opus`→`opus`, `glm-5`→false, empty→false (backs "Garbage model value is surfaced", "Display name is normalized").
- [x] 1.6 `internal/config/model_test.go`: `ResolvePhaseModels` — (a) phase set→alias; (b) phase unset+default set→default alias; (c) unset+no default→omitted; (d) mixed→canonical order; (e) twice→`reflect.DeepEqual` (backs fallback/omit/ordering scenarios).
- [x] 1.7 `internal/config/model_test.go`: `Validate` — `Opus 4.8`→no warn; `Opues 4.8`→warn; `glm-5`→no warn; empty→no warn.

## Phase 2: Template rendering (Unit 2)

- [x] 2.1 `internal/initcmd/templates.go`: add `PhaseModels []config.PhaseModel` to `TemplateData`; add `internal/config` import if missing.
- [x] 2.2 `internal/initcmd/templates.go`: insert `{{if .PhaseModels}}`-guarded `## Phase Models` block in `orchestratorTrailer` after `## Configuration`, before `## State Management` (literal snippet in design Decision 4; `§` = backtick placeholder).
- [x] 2.3 `internal/initcmd/templates_test.go`: block-present test — `TemplateData{PhaseModels: ...}` asserts `## Phase Models`, `- propose: sonnet`, no raw `Opus 4.8`, no `§` (backs "Init renders a phase model").
- [x] 2.4 `internal/initcmd/templates_test.go`: empty-omit test — `PhaseModels: nil` excludes `## Phase Models` header (backs "Phase omitted when no model resolves").
- [x] 2.5 `internal/initcmd/templates_test.go`: confirm existing golden tests `_EmptyData`, `_FiveRules`, `_AgentsAndClaudeIdentical` stay green.

## Phase 3: Plumbing (Unit 3)

- [x] 3.1 `internal/initcmd/init.go`: change `writeTemplate` to `writeTemplate(projectDir, agentName string, skillCount int, phaseModels []config.PhaseModel) error`; set `data.PhaseModels = phaseModels`.
- [x] 3.2 `internal/initcmd/init.go` (~L96): caller passes `config.ResolvePhaseModels(cfg.Models)`.
- [x] 3.3 `internal/tui/model.go` (~L333-338): add `PhaseModels: config.ResolvePhaseModels(cfg.Models)` to `regenerateTemplate`'s `TemplateData` literal.
- [x] 3.4 `internal/initcmd/templates_test.go`: cross-path identity test — block from init data vs. identically-built `TemplateData` is byte-identical (backs "TUI regeneration produces the same block as init", "canonical order byte-identical").

## Phase 4: Verification

- [x] 4.1 `gofmt -l internal/` returns empty (or `go fmt ./...`).
- [x] 4.2 `go build ./...` succeeds.
- [x] 4.3 `go vet ./...` clean.
- [x] 4.4 `go test ./...` passes (config + initcmd + tui).
