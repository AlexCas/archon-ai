# Tasks: AI Orchestration Harness CLI

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1,600 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | 5 stacked PRs |
| Delivery strategy | ask-on-risk |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | PR | Notes |
|------|------|----|-------|
| 1 | Foundation: go.mod, config, version, embed.FS, rollback | PR 1 | base=main; tests incl |
| 2 | Agent detection + status display | PR 2 | base=main; dep PR 1 |
| 3 | Scaffold + init orchestrator | PR 3 | base=main; dep PR 1,2 |
| 4 | CLI entry point (cobra) + rollback + e2e | PR 4 | base=main; dep PR 3 |
| 5 | Meta-skills: harness-workflow, harness-judge | PR 5 | base=main; independent |

## Phase 1: Foundation

- [ ] 1.1 Init Go module, add deps (cobra v1.8, yaml.v3) — `go.mod`, `go.sum`
- [ ] 1.2 `internal/config/config.go` — Config struct, Load/Save with injectable HomeDir, YAML roundtrip
- [ ] 1.3 `internal/config/rollback.go` — RollbackManifest, WriteManifest, Cleanup, backup/restore AGENTS.md
- [ ] 1.4 `internal/version/info.go` — ldflags: Version, Commit, Date; Print()
- [ ] 1.5 `skills/embed.go` — `//go:embed */SKILL.md` with 21 existing skills
- [ ] 1.6 Tests: config Load/Save (MapFS), rollback manifest lifecycle, version output

## Phase 2: Agent Detection

- [x] 2.1 `internal/agent/detect.go` — scan `.opencode/`, `.claude/`, `.agents/`, `.codex/` with priority sort
- [x] 2.2 Multi-agent detection — interactive prompt when >1 found
- [x] 2.3 `internal/status/display.go` — read `.archon/config.yaml`, format agent/harness state
- [x] 2.4 Tests: agent detect with fstest.MapFS multi-agent scenario, status output

## Phase 3: Scaffold & Init Orchestrator

- [ ] 3.1 `internal/scaffold/extract.go` — fs.ReadDir → MkdirAll → WriteFile for all 21 skills
- [ ] 3.2 `internal/scaffold/symlink.go` — SymlinkOrCopy with copy fallback on EPERM/EINVAL
- [ ] 3.3 `internal/init/templates.go` — embed AGENTS.md/CLAUDE.md orchestrator template
- [ ] 3.4 `internal/init/init.go` — orchestrator: detect→extract→symlink→config→template chain
- [ ] 3.5 Version-gap detection — compare metadata.version frontmatter, prompt on mismatch
- [ ] 3.6 Tests: extract idempotency, symlink fallback, init orchestration flow, version-gap prompt

## Phase 4: CLI & Rollback

- [x] 4.1 `cmd/archon/main.go` — cobra root with `init`, `version`, `status`, `rollback` subcommands
- [x] 4.2 Wire init (`--agent`, `--force`), rollback (`--dry-run`), version, status to internal packages
- [x] 4.3 Integration tests: golden-file output for each CLI command (init in tmpdir)
- [x] 4.4 E2E test: `archon init` → `archon status` → `archon rollback` cycle in `t.TempDir()`

## Phase 5: Meta-Skills

- [x] 5.1 `skills/harness-workflow/SKILL.md` — phase state machine with state.yaml read/write; block invalid transitions; report current state
- [x] 5.2 `skills/harness-judge/SKILL.md` — invoke judgment-day, parse verdict, mutation gate via gremlins, auto-re-apply loop (max 3), structured feedback output
- [x] 5.3 Update `.atl/skill-registry.md` — index harness-workflow and harness-judge
