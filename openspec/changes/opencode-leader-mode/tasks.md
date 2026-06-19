# Tasks: Opencode archon-leader mode + configurable leader model (Slice 2)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 350-450 (split keeps each PR under 400) |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Suggested split | PR1 (core) → PR2 (TUI, stacked on PR1) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | config field + `mergeOpencodeAgent` writer + init wiring + rollback fix + tests | PR1 | base `master`; self-contained, mergeable alone |
| 2 | TUI leader field + save-path merge + parity test | PR2 | base = PR1 branch; stacked-to-main |

# PR1 — Core (config + writer + init wiring + tests)

## Phase 1: Config field (PR1)

- [x] 1.1 In `internal/config/model.go`, add `Leader string \`yaml:"leader,omitempty"\`` to `ModelConfig`.
- [x] 1.2 In `internal/config/config.go` `Clone()`, set `Leader: c.Models.Leader` on the `ModelConfig` literal (line ~96).
- [x] 1.3 In `internal/config/config_test.go`, extend `TestConfig_CloneRoundtrip` fixture with a `Models.Leader` value (`anthropic/claude-sonnet-4-20250514`) — covers **S1**.

## Phase 2: Writer (PR1)

- [x] 2.1 Create `internal/initcmd/opencode_mode.go`: `const leaderAgentName = "archon-leader"` + struct `archonLeaderAgent{Mode,Description,Model,Prompt}` (json tags; declaration order = output order).
- [x] 2.2 Implement `mergeOpencodeAgent(projectDir, leader string) (written string, err error)`: `"", nil` if empty; read `opencode.json` into `map[string]any`; set only `agent["archon-leader"]` (`mode:"primary"`, `prompt:"{file:./AGENTS.md}"`, `model:leader`), preserving other keys; never `default_agent`; `MarshalIndent`+`\n`; atomic `.tmp`+`os.Rename`.

## Phase 3: Init wiring + rollback fix (PR1)

- [x] 3.1 In `internal/initcmd/init.go`, add `ModelLeader string` to `Options`; thread into `buildConfig` → `config.ModelConfig{Leader: ...}`.
- [x] 3.2 In `Run()`, after `writeTemplate` call `mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader)` only when `agentName == "opencode"`; wrap errors `merge opencode agent: %w`.
- [x] 3.3 Rollback ordering fix: move `WriteManifest()` to AFTER `writeTemplate`+merge; append merge's non-empty path to `CreatedPaths` BEFORE `WriteManifest()` so `opencode.json` is registered.
- [x] 3.4 CLI flag: add `--leader` to `archon init` in `cmd/archon/main.go` (consistent with the existing `--model*` flags), advisory-validate it, and thread into `initcmd.Options.ModelLeader` so init populates `models.leader` end-to-end.

## Phase 4: Writer tests + verify (PR1)

- [x] 4.1 Create `internal/initcmd/opencode_mode_test.go`: merge creates `opencode.json` (mode `primary`, prompt `{file:./AGENTS.md}`, model=leader, no `default_agent`) + manifest registers the path (**S2**).
- [x] 4.2 Seed `opencode.json` with unrelated keys/agents; assert `agent.archon-leader` added, all pre-existing left untouched (**S3**).
- [x] 4.3 Run merge twice; assert byte-identical via `bytes.Equal` (**S4**).
- [x] 4.4 Assert no file written for `agent != opencode` (**S5**) and empty `models.leader` (**S6**).
- [x] 4.5 `go build ./...` + `go test ./internal/config/... ./internal/initcmd/...` green.

# PR2 — TUI (field + merge + parity), stacked on PR1

## Phase 5: TUI leader field (PR2)

- [ ] 5.1 In `internal/tui/models_tab.go`, add a leader `textinput`, rendered + focus-traversed only when `cfg.Agent == "opencode"`.
- [ ] 5.2 In `applyToConfig`, set `cfg.Models.Leader` from that input (opencode only).

## Phase 6: Save-path merge (PR2)

- [ ] 6.1 In `internal/tui/model.go` `saveConfig`, after `cfg.Save()`, call `mergeOpencodeAgent(m.projectDir, cfg.Models.Leader)` when `cfg.Agent == "opencode"`; surface errors. Leave `archon update` untouched.

## Phase 7: Parity test + verify (PR2)

- [ ] 7.1 TUI==init parity test: drive `saveConfig`, compare `agent.archon-leader` bytes to a direct `mergeOpencodeAgent` call with the same leader (**S7**).
- [ ] 7.2 Assert `archon update` leaves an existing `opencode.json` unwritten (mtime/bytes unchanged) (**S8**).
- [ ] 7.3 `go build ./...` + `go test ./...` green.
