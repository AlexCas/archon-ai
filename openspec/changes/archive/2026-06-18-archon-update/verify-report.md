# Verification Report

- **Change**: archon-update (PR1 `feat/archon-update-foundation` + PR2 working tree on `feat/archon-update-command`)
- **Mode**: openspec, STANDARD testing (`strict_tdd: false`), runner `go test`
- **Artifacts present**: proposal, design, specs (harness-update, harness-init) + `.feature`, tasks
- **Scope verified**: cumulative working tree (PR1 + PR2)

## 1. Task Completeness

| Phase | Task | Status | Evidence |
|-------|------|--------|----------|
| 1.1 | `ExtractVersion` exported | [x] DONE | `version.go:179` |
| 1.2 | `SkillVersion(fs,name)` | [x] DONE | `version.go:169`; tested `TestSkillVersion` |
| 1.3 | `SkillChange`/`GapReport` | [x] DONE | `version.go:22-38` |
| 1.4 | `ClassifyGaps` + `"1.0"`/empty unknown | [x] DONE | `version.go:46-135`; tested |
| 1.5 | `DetectVersionGaps` delegate | [x] DONE | `version.go:141-165` |
| 2.1 | `refresh.go` `refreshSkills` (no template/config writes) | [x] DONE | `refresh.go:30-57` |
| 2.2 | `Run` rewired, order preserved | [x] DONE | `init.go:76-102` |
| 2.3 | `buildConfig` takes inventory, no `"1.0"` | [x] DONE | `init.go:84,192-219` |
| 3.1 | `update.go` `Update`, missing config → error+no writes | [x] DONE | `update.go:53-72` |
| 3.2 | ClassifyGaps none→up-to-date; copy-mode WARN; machine-wide scope | [x] DONE | `update.go:88-110`, `main.go:268-272` |
| 3.3 | `--check` no writes; else refresh; `--prune` RemoveAll (global+link) | [x] DONE | `update.go:114-159` |
| 3.4 | `Clone` patch only Version/SkillCount/SkillInventory, set HomeDir, Save; never writeTemplate | [x] DONE | `update.go:133-141` |
| 3.5 | `newUpdateCmd` `--check`/`--prune`/`--agent` | [x] DONE | `main.go:206-283` |
| 3.6 | `DisplayWithUpdate`; `Display` delegates 0; status computes gaps, silent on error | [x] DONE | `display.go:12-89`, `main.go:374-383` |
| 4.1 | scaffold tests SkillVersion + table ClassifyGaps | [x] DONE | `version_test.go` |
| 4.2 | init golden + truthful versions | [x] DONE | `init_test.go:483,531` |
| 4.3 | update missing-config/no-gap/Clone preservation | [x] DONE | `update_test.go` |
| 4.4 | update --check / --prune / no-prune | [x] DONE | `update_test.go` |
| 4.5 | cmd update --check no mutation; copy-mode warning | [x] DONE | `main_test.go:201,245` |
| 5.1 | README documents update | [x] DONE | `README.md:76,113-138` |
| 5.2 | dynamic dry-run skill count | [x] DONE | `main.go:117` uses `embeddedSkillCount()` |

All 21 tasks checked and substantiated. No unchecked implementation tasks.

## 2. Build / Vet / Test Evidence (real runtime output)

- `go build ./...` → exit 0 (no output)
- `go vet ./...` → exit 0 (no output)
- `go test ./... -count=1` → all packages `ok` (cache defeated):
  ```
  ok  cmd/archon            0.019s
  ok  internal/agent        0.002s
  ok  internal/config       0.004s
  ok  internal/initcmd      0.013s
  ok  internal/scaffold     0.005s
  ok  internal/status       0.002s
  ok  internal/tui          0.540s
  ok  internal/version      0.002s
  ok  skills                0.002s
  TEST_EXIT=0
  ```
- Verbose targeted run confirmed every named scenario test executed and PASSED (see matrix).

## 3. Spec → Test Compliance Matrix

### harness-update.feature

| Scenario | Covering test | Result | Notes |
|----------|---------------|--------|-------|
| Update refreshes skills without touching template or user config | `initcmd.TestUpdate_PreservesUserConfigAndTemplate` | PASS | Asserts CLAUDE.md byte-unchanged; created_at, agent, mutation_testing, playwright, models preserved after a real-gap apply. |
| Update classifies the version gap | `scaffold.TestClassifyGaps/mixed_added_changed_and_orphaned` + `initcmd.TestUpdate_CheckReportsDiffWithoutWriting` | PASS | Classification proven at the classifier (all three buckets) and surfaced via `UpdateResult.GapReport`. Reporting-to-user is rendered in `main.go:254-259`; see WARNING W1. |
| No gaps reports already up to date | `initcmd.TestUpdate_NoGapsReportsAlreadyUpToDate` | PASS | `UpToDate==true`, `Wrote==false`, config byte-identical. |
| Check reports the diff without writing | `initcmd.TestUpdate_CheckReportsDiffWithoutWriting` + `cmd.TestUpdateCommand_CheckNoMutation` | PASS | Changed==1, no config/global write; command asserts removed skill not re-extracted. |
| Prune removes orphaned skills | `initcmd.TestUpdate_PruneRemovesOrphanedSkills` | PASS | Orphan dir removed; `Pruned==[old-skill]`. |
| Orphans are kept without prune | `initcmd.TestUpdate_OrphansKeptWithoutPrune` | PASS | Orphaned reported, `Pruned` empty, dir still present. |
| Copy-mode install warns without re-linking | `cmd.TestUpdateCommand_CopyModeWarning` | PASS | stderr contains "copy-mode"; real dir stays non-symlink (`Lstat`). |
| Output states machine-wide scope | `cmd.TestUpdateCommand_CheckNoMutation` (asserts "machine-wide") | PARTIAL | See WARNING W2 — only asserted on the `--check` path. |
| Update before init reports an actionable error | `initcmd.TestUpdate_BeforeInitReportsActionableError` | PASS | Error contains "archon init"; `.archon` not created. |

### harness-init.feature

| Scenario | Covering test | Result | Notes |
|----------|---------------|--------|-------|
| Init records real frontmatter versions | `initcmd.TestRun_RecordsRealFrontmatterVersions` | PASS | sdd-init→2.0, sdd-propose→3.0, no entry =="1.0", source=="embedded". |
| Missing frontmatter version is handled | `initcmd.TestRun_RecordsRealFrontmatterVersions` (sdd-noversion) + `scaffold.TestSkillVersion/version_missing_in_frontmatter` | PASS | Versionless skill recorded with empty version, init does not abort. |

All 11 scenarios have a covering test that passed at runtime. No CRITICAL UNTESTED/FAILING.

## 4. Design Coherence

| Decision | Verdict | Evidence |
|----------|---------|----------|
| D1 refreshSkills excludes template/config/rollback/openspec | HONORED | `refresh.go` does only Extract+symlink+inventory; `Update` never calls `writeTemplate` (grep: writeTemplate only in `init.go`). |
| D2 command lives in `internal/initcmd` reusing helpers | HONORED | `update.go` reuses `refreshSkills`, `resolveProjectSkillsDir`, `knownAgent`. |
| D3 prune opt-in + machine-wide scope printed | HONORED | Prune gated on `opts.Prune`; scope line `main.go:268`. |
| D4 `"1.0"`/empty installed = unknown | HONORED | `isUnknownVersion` (`version.go:46`); proven by `TestClassifyGaps` unknown/empty cases. |
| D5 ClassifyGaps core; DetectVersionGaps delegates | HONORED | `DetectVersionGaps` calls `ClassifyGaps` and returns Added+Changed only. |
| Config preservation patches only Version/SkillCount/SkillInventory | HONORED | `update.go:133-137`; `Clone` (`config.go:64`) deep-copies all other fields. |

## 5. Correctness Notes

- `Update` copy-mode path refreshes the global dir via `scaffold.Extract` but does NOT relink and does NOT save config (`result.Wrote` stays false unless `--prune`). Consistent with the spec ("project needs its own update").
- Orphan detection keys off `SKILL.md` presence in both embedded and installed dirs, correctly skipping `_shared`.
- `DisplayWithUpdate` hint counts only Added+Changed (not Orphaned), matching the "update available" intent.

## 6. Issues

### CRITICAL
None.

### WARNING
- **W1 — "classification reported to the user" is asserted only at the data layer.** The `Update classifies the version gap` scenario's "And the classification is reported to the user" step is satisfied by `main.go:254-259` (prints Added/Changed/Orphaned counts and names), but no command-layer test asserts that this rendered output actually contains the three buckets on the apply path. The behavior exists in code and the data-layer classification is well tested; the rendered report string is untested.
- **W2 — "Output states machine-wide scope" tested only on `--check`.** The scope line is printed unconditionally for any non-up-to-date result (`main.go:268`), but the only assertion of the "machine-wide" string is in `TestUpdateCommand_CheckNoMutation` (the dry-run path). The apply path that the scenario describes ("the user runs archon update", no `--check`) has no direct assertion of the scope string. Code path is shared, so the risk is low, but the scenario is technically verified by inference rather than by a dedicated apply-path assertion.

### SUGGESTION
- **S1 — Spec references a non-existent `judge` config field.** Both specs/tasks require preserving `judge` in `.archon/config.yaml`, but `config.Config` has no `judge` field (only a comment at `config.go:22` mentions the judge phase). Preservation is therefore vacuously true. Either add the field if judge config is intended, or drop `judge` from the spec's preserved-fields list to avoid implying coverage that does not exist. `TestUpdate_PreservesUserConfigAndTemplate` accordingly does not assert judge (cannot).
- **S2 — `Update` copy-mode + `--prune` removes the project link** (`update.go:152-155`) even though the project was intentionally not re-linked. For a copy-mode project this deletes a real orphan directory under the project; acceptable but worth a doc note.

## 7. Verdict

**PASS WITH WARNINGS**

All 21 tasks are complete and substantiated. Build, vet, and the full test suite pass with the cache defeated. All 11 Gherkin scenarios map to covering tests that passed at runtime; the implementation honors design decisions D1–D5 and the config-preservation contract (template never rewritten, only Version/SkillCount/SkillInventory patched). The warnings concern command-layer output assertions (rendered classification report and the machine-wide scope line on the apply path) that are verified by code inspection and shared-path inference rather than dedicated apply-path tests, plus a spec reference to a `judge` field that does not exist in the config. None block archive; W1/W2 are reasonable candidates for a follow-up test, and S1 should be reconciled in the spec.
