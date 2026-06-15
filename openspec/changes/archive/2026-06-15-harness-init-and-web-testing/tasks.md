# Tasks: Harness init UX, session status, Gherkin & web testing

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~900–1100 (Go + skills markdown) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 config/models → PR 2 init+TUI UX → PR 3 harness skills (session/Gherkin/Playwright/commits) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Playwright config + curated models | PR 1 | config, model.go, config get/set, status display + tests |
| 2 | init folder creation + existing-file guard + TUI (Playwright tab, model catalog, init-without-config) | PR 2 | initcmd, cmd/archon, tui + tests |
| 3 | Harness skills: SESSION_STATUS, Gherkin, web/Playwright flow, commit rule, templates | PR 3 | skills markdown + templates.go + CLAUDE.md |

> Note: implemented as a single batch during this retrofit. Slicing into the PRs
> above is recommended before merge (preflight C = ask-always).

## Phase 1: Config foundation

- [x] 1.1 Add `Playwright{Enabled,TestDir,BaseURL}` struct + Config field + Clone in `internal/config/config.go`
- [x] 1.2 Curate `ClaudeModels`/`OpencodeModels`/`StaticModels()` and derive `KnownModels` in `internal/config/model.go`
- [x] 1.3 Add `playwright.*` and `mutation_testing.enabled` to `config get/set` in `cmd/archon/config.go`
- [x] 1.4 Show Playwright section in `internal/status/display.go`
- [x] 1.5 Update `model_test.go` for the curated catalog

## Phase 2: init + TUI UX

- [x] 2.1 `detectAgent` no longer requires folder; `ensureAgentDir` creates it (`internal/initcmd/init.go`)
- [x] 2.2 `ErrTemplateExists` guard + `OverwriteTemplate`; early abort before any work
- [x] 2.3 `--playwright` flag, `confirmOverwrite` prompt, default-config TUI launch (`cmd/archon/main.go`)
- [x] 2.4 New `playwright_tab.go`; wire `PlaywrightTab` into `model.go`
- [x] 2.5 Static model catalog cycling in `models_tab.go`
- [x] 2.6 In-TUI replace confirmation + config reload after init (`agent_tab.go`, `model.go`)
- [x] 2.7 init tests: abort-on-existing, create-folder-when-missing; TUI tab-count test

## Phase 3: Harness skills

- [x] 3.1 New `_shared/session-status-contract.md`
- [x] 3.2 `harness-workflow` writes SESSION_STATUS per transition; `sdd-archive` moves it
- [x] 3.3 `sdd-spec` emits formal Gherkin `.feature` files
- [x] 3.4 `sdd-explore` web detection; preflight group E in templates + CLAUDE.md
- [x] 3.5 `sdd-apply` generates Playwright from Gherkin; `harness-judge` Playwright gate
- [x] 3.6 `sdd-verify`/`sdd-tasks`/`openspec-convention` consume Gherkin
- [x] 3.7 Commit-authorship rule in templates, `work-unit-commits`, `sdd-apply`

## Phase 4: Verification

- [x] 4.1 `go build ./...` and `go test ./...` pass
- [x] 4.2 Binary smoke test: blank-project init, existing-file decline, config get/set
