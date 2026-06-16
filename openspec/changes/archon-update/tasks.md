# Tasks: First-class `archon update` command

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 650–800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 foundation → PR 2 command+status |
| Delivery strategy | ask-on-risk |
| Chain strategy | feature-branch-chain |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Scaffold version/gap detection + init `refreshSkills` refactor + truthful inventory + tests | PR 1 | Standalone; `Run` unchanged (golden test). base = master |
| 2 | `archon update` command + status hint + tests + README | PR 2 | Depends on PR 1. base = PR 1 branch (feature-branch-chain) or master (stacked) |

Base boundaries: stacked-to-main → both base master in order; feature-branch-chain → PR 2 base = PR 1 branch.

## Phase 1: Foundation — scaffold (PR 1)

- [x] 1.1 `internal/scaffold/version.go`: promote `extractVersion`→exported `ExtractVersion`.
- [x] 1.2 Add `SkillVersion(embeddedFS, name) string` reading `name/SKILL.md` via `ExtractVersion`; `""` if absent.
- [x] 1.3 Add `SkillChange{Name,EmbeddedVer,InstalledVer}` + `GapReport{Added,Changed,Orphaned}`.
- [x] 1.4 Add `ClassifyGaps(embeddedFS, installedDir) (GapReport, error)`; `"1.0"`/empty installed = unknown.
- [x] 1.5 Keep `DetectVersionGaps` as delegate over `ClassifyGaps` (Added+Changed).

## Phase 2: Foundation — init refactor (PR 1)

- [x] 2.1 `internal/initcmd/refresh.go`: `RefreshResult` + `refreshSkills(homeDir,projectDir,agentName,embeddedFS)` (extract+link+inventory, no template/config writes).
- [x] 2.2 Rewire `Run` (`init.go:76-85`) to call `refreshSkills`; keep guard→ensureAgentDir→Save→rollback→writeTemplate→openspec order identical.
- [x] 2.3 `buildConfig` (`init.go:195-203`) takes `res.Inventory`; drop hardcoded `"1.0"`.

## Phase 3: Command + status (PR 2)

- [ ] 3.1 `internal/initcmd/update.go`: `UpdateOptions`/`UpdateResult`, `Update(opts)`; `config.Load` fail → "run init first", no writes.
- [ ] 3.2 `Update`: `ClassifyGaps` (none→"up to date", no writes); copy-mode `Lstat` non-symlink → WARN no re-link; print machine-wide scope.
- [ ] 3.3 `Update`: `--check` prints report, no writes; else `refreshSkills`; `--prune` `RemoveAll` orphans (global + project link).
- [ ] 3.4 `Update`: `cfg.Clone()`, patch only `Version`/`SkillCount`/`SkillInventory`, set `HomeDir`, `Save`; never `writeTemplate`.
- [ ] 3.5 Wire `newUpdateCmd` (`--check`/`--prune`/`--agent`) in `cmd/archon/main.go`.
- [ ] 3.6 `status.DisplayWithUpdate(w,cfg,n)`; `Display` delegates `0`; `newStatusCmd` computes `ClassifyGaps`, appends hint (silent on error).

## Phase 4: Testing

- [x] 4.1 `scaffold/version_test.go`: `SkillVersion` present/missing; table `ClassifyGaps` (added/changed/orphaned/`"1.0"`-unknown). Backs "Missing frontmatter version is handled".
- [x] 4.2 `initcmd` golden test: `Run` still writes template+config+rollback (regression); `refreshSkills` versions ≠ "1.0". Backs "Init records real frontmatter versions".
- [ ] 4.3 `update_test.go`: missing config → error+no writes ("Update before init reports an actionable error"); no-gap → no writes ("No gaps reports already up to date"); Clone preserves models/playwright/judge/created_at ("Update refreshes skills without touching template or user config").
- [ ] 4.4 `update_test.go`: `--check` no writes ("Check reports the diff without writing"); `--prune` removes orphan ("Prune removes orphaned skills"); no-prune keeps it ("Orphans are kept without prune").
- [ ] 4.5 `cmd/archon` test: `update --check` no fs mutation; copy-mode warning ("Copy-mode install warns without re-linking").

## Phase 5: Docs/cleanup

- [ ] 5.1 README: document `archon update` (`--check`/`--prune`/`--agent`, machine-wide scope).
- [ ] 5.2 Fix stale "21 embedded skills" dry-run text in `cmd/archon/main.go` if present.
