# Verification Report: opencode-model-assignment (Slice 1)

**Change**: opencode-model-assignment
**Slice**: Slice 1 (PR 1) — Standalone Bug Fix (S1-1 .. S1-19)
**Branch**: fix/opencode-model-assignment (working tree, uncommitted)
**Mode**: Full artifact verification (proposal context + 2 specs + design + tasks)
**Date**: 2026-06-19 (re-verification after defect fixes)
**Verdict**: **PASS** (both prior defects resolved; no regression; real `$HOME` untouched)

---

## Summary

This is a RE-VERIFICATION of Slice 1 after the two defects from the first run were fixed. Both are now resolved, the full suite passes with `-count=1`, and the run is provably hermetic: the developer's real `~/.config/opencode/opencode.json` is byte-for-byte identical before and after, with zero new backups.

- **HIGH (Defect 2 in history below) — hermeticity — RESOLVED.** `internal/opencode/paths.go` now exposes `SettingsPathFor(homeDir)` / `CachePathFor(homeDir)`, and `internal/initcmd/init.go:80-81` derives the overlay settings + cache paths from `opts.HomeDir` instead of the implicit-home `SettingsPath()`/`CachePath()`. The Apply path (`apply.go`) consumes only `opts.SettingsPath`/`opts.CachePath` and `resolveWithCachePath(name, opts.CachePath)` — no `os.UserHomeDir()` is reached on the Apply route (the only mention in `apply.go` is a doc comment). New guard test `TestRun_OpenCodeOverlay_Hermetic` (`init_test.go:377`) records the real-home file + backup count before the run and asserts both are unchanged afterward.
- **CRITICAL (Defect 1 in history below) — deterministic resolution — RESOLVED.** `resolve.go` now treats the static map as canonical for known names: when a known bare name maps to `provider/name`, that provider wins; the cache is consulted only *within* that provider for a more specific (date-suffixed) ID, and the fallthrough loop iterates providers in sorted key order. With a synthetic multi-provider cache where `claude-sonnet-4` appears under `neon` and `azure` decoys while `anthropic` lists only `claude-sonnet-4-20250514`, `Resolve` deterministically returns `anthropic/claude-sonnet-4-20250514`. Covered by `TestResolve` subcases plus `TestResolve_StaticMapTiebreak_Deterministic` (100 iterations) and `TestResolve_NonStaticName_SortedProviderOrder` (100 iterations).

All other Slice 1 requirements remain satisfied with code + passing tests (no regression).

---

## Defect history (found in first verify run, now fixed)

| # | Severity | Defect | Resolution | Covering test |
|---|----------|--------|------------|---------------|
| D1 | CRITICAL | `Resolve` fell through to a random-map-order provider scan and returned `neon/claude-sonnet-4` for the bare name under a real multi-provider cache; default-fallback could be wrong/empty. | `resolve.go` makes the static map canonical for known names (sorted, deterministic); cache only refines the ID within the canonical provider. | `TestResolve` (decoy subcases), `TestResolve_StaticMapTiebreak_Deterministic`, `TestResolve_NonStaticName_SortedProviderOrder` |
| D2 | HIGH | `Apply`/init used `SettingsPath()`/`CachePath()` (implicit real `$HOME`), so opencode init tests mutated and backed up the developer's real `~/.config/opencode/opencode.json`. | `paths.go` adds `*For(homeDir)` helpers; `init.go` passes `opts.HomeDir`-derived paths; Apply consumes explicit paths only. | `TestRun_OpenCodeOverlay_Hermetic` (asserts real-home invariant) |

---

## Hermeticity Proof (this run)

Captured around the full `go test ./... -count=1` run (which exercises `TestRun`, `TestRun_WithModelFlags`, and `TestRun_OpenCodeOverlay_Hermetic`):

| Metric | Before tests | After tests | Result |
|--------|--------------|-------------|--------|
| `md5sum ~/.config/opencode/opencode.json` | `9f92fd6ea1ac2a4f234dc35ef6d9da43` | `9f92fd6ea1ac2a4f234dc35ef6d9da43` | IDENTICAL |
| `~/.config/opencode/opencode.json.backup.*` count | 0 | 0 | NO NEW BACKUPS |

The real `$HOME` was not touched, backed up, restored, or deleted by this verification. PASS.

---

## Build Evidence (this run)

| Command | Result |
|---------|--------|
| `gofmt -l internal/opencode internal/config internal/initcmd` | Only `internal/config/model_test.go` flagged — pre-existing, unmodified by this change (`git diff --stat` empty). Not a Slice 1 defect. All Slice 1 files are gofmt-clean. |
| `go vet ./...` | PASS — exit 0, no warnings |
| `go build ./...` | PASS — exit 0, clean compilation |
| `go test ./... -count=1` | PASS — all 10 packages green (no stale cache) |

`go test ./... -count=1` output: `ok` for cmd/archon, internal/agent, internal/config, internal/initcmd, internal/opencode, internal/scaffold, internal/status, internal/tui, internal/version, skills.

Defect-fix tests (verbose, this run, all PASS):
- `internal/opencode`: `TestResolve` (incl. `static-map_provider_wins_over_decoys_(multi-provider_cache)` and `static-map_provider_wins_when_canonical_provider_absent_from_cache`), `TestResolve_StaticMapTiebreak_Deterministic`, `TestResolve_NonStaticName_SortedProviderOrder`.
- `internal/initcmd`: `TestRun`, `TestRun_WithModelFlags`, `TestRun_OpenCodeOverlay_Hermetic`.

---

## No-Regression Re-Confirmation (coverage table delta)

The first run's requirement→evidence coverage table is unchanged except that the two previously-failing rows now pass. Spot-checks performed this run:

| Area | Check | Result |
|------|-------|--------|
| Decision tree (3 cases) | `TestInject_ExplicitPhaseWins`, `TestInject_ExistingUserAgentPreserved`, `TestInject_DefaultFallback` | PASS |
| Shared-file backup/restore-not-delete | `internal/config` rollback + FileBackup tests | PASS |
| opencode.json NOT in CreatedPaths | code inspection (`apply.go` never appends; `init.go` records only `FileBackups`) | PASS |
| Overlay integrity | `archon-orchestrator` + exactly 10 hidden `sdd-*` subagents; overlay keys == the 10 `skills/sdd-*` folders (10 == 10) via `TestOverlaySubagentKeysMatchSkillFolders` | PASS |
| Both templates contain `## Phase Models` | `grep -c "## Phase Models" templates.go` == 2; `TestTemplates_PhaseModelsSection` | PASS |
| Init gating on agent==opencode | `init.go:77` gate; `TestRun_ClaudeAgent` confirms claude path unaffected | PASS |
| Cache-based deterministic resolution (model-resolution) | now deterministic — see Defect history D1 | PASS (was FAIL) |
| Default-fallback no longer empty/wrong under multi-provider cache | resolution canonicalizes to static-map provider | PASS (was FAIL) |

Coverage (changed packages, unchanged from prior run within rounding): `internal/opencode` ~75%, `internal/config` ~70%, `internal/initcmd` ~82%.

---

## Issues

### CRITICAL / HIGH
None. Both prior defects are resolved and covered by tests.

### SUGGESTION (carried over, non-blocking)
1. **S1-14 golden fixtures not committed.** `internal/opencode/testdata/` is empty; Apply tests assert behaviour inline rather than against golden files. Behaviourally adequate, but does not match the task's "golden file" wording. Consider adding committed golden fixtures or updating the task text. Not a blocker.

---

## Verdict

**PASS.** Build, vet, gofmt (changed files), and the full test suite pass with `-count=1`. Both prior defects are fixed and guarded by tests: deterministic static-map resolution (CRITICAL D1) and hermetic, `Options.HomeDir`-derived Apply paths (HIGH D2). The verification run left the developer's real `~/.config/opencode/` byte-for-byte unchanged with zero new backups (proof above). No regression in the rest of Slice 1. No production code was modified and nothing was committed during verification. Slice 1 is archive-ready.

---

## Re-judge after fixes (2026-06-19)

Confirmation of the four fixes raised by the prior blind dual review. Verifier ran build, `go vet ./...`, `gofmt -l`, and `go test ./... -count=1` (all green), inspected production code and the locking tests, and proved hermeticity against the real `~/.config/opencode/`.

**Verdict: SHIP.** All four fixes are CONFIRMED in code and locked by tests. No regressions introduced. No new HIGH/MEDIUM issues. One pre-existing non-blocking nit unrelated to these fixes.

### Per-fix confirmation

1. **HIGH — no empty `model`: CONFIRMED.** `internal/opencode/inject.go` `delete(agentMap, "model")` in the default-fallback branch when `defaultModel==""` (lines 68-72); the explicit-phase branch only sets a non-empty value and case-2 removes the agent entirely, so no branch can emit `"model":""`. `internal/opencode/assets/sdd-overlay.json` contains zero `"model"` fields (grep exit 1). Tests `TestInject_EmptyDefaultModel_NoModelKey` and `TestInject_PerPhaseExplicit_OnlyThatPhaseHasModel` assert the key is absent (not `""`) and that the marshaled output contains no `"model": ""`. NOTE (doc nit, non-blocking): the inject.go function docstring lines 19-20 still says the field "is left as `` (not removed)", which now contradicts the implemented delete. Documentation-only; behavior and tests are correct.

2. **MEDIUM — claudeTemplate Phase Models: CONFIRMED.** `claudeTemplate` Phase Models describes the advisory per-delegation `model: <id>` mechanism and the per-phase model list, with no `opencode.json` reference; `agentsTemplate` still describes the `opencode.json` wiring. `TestTemplates_PhaseModelsSection` asserts CLAUDE.md contains "## Phase Models" and "model:" but NOT "opencode.json", and AGENTS.md contains both "## Phase Models" and "opencode.json".

3. **MEDIUM — rollback removes fresh opencode.json: CONFIRMED.** `internal/initcmd/init.go` always appends `config.FileBackup{Target, Backup}` for opencode.json regardless of `Backup==""` (lines 100-103); opencode.json is NOT in `CreatedPaths` (`buildRollbackManifest` adds only config.yaml, rollback.json, and skills). `config.Cleanup` restores when `Backup!=""` and removes when `Backup==""`. Tests `TestRun_OpenCode_RollbackManifest_FreshFile` (fresh → Backup empty → Cleanup removes file) and `TestRun_OpenCode_RollbackManifest_PriorFile` (prior file → Backup non-empty → Cleanup restores prior bytes `{"prior": true}`).

4. **MEDIUM — resolve.go prefix match: CONFIRMED.** `internal/opencode/resolve.go` prefers exact `m.ID==name`, else accepts only `^<QuoteMeta(name)>-\d{8}$` date snapshots and picks the newest via `sort.Strings` last element, else falls back to the static map. Regex is anchored at both ends and QuoteMeta'd. Tests: `TestResolve_Fix4_Gpt4DoesNotMatchGpt4Turbo` (gpt-4 does not resolve to gpt-4-turbo), `TestResolve_Fix4_DateSnapshotNewest` (rejects `claude-sonnet-4-5` decoy, picks newest `...20251022`), `TestResolve_Fix4_Gpt4WithoutExactMatch` (static-map fallback), plus deterministic 100-iteration tests.

### New-problem checks (all clear)

- **Empty-model resurrection via merge:** No path. When the user already has the agent (`existingAgentKeys`), Inject case 2 removes it from the overlay so the merge preserves the user's own entry (correct). When the user lacks the agent, the overlay carries no `model` key and the base has no agent, so the deep-merge produces no `model` key. The merge never reintroduces `"model":""`.
- **FileBackup vs AGENTS.md BackupPath:** Decoupled in `Cleanup` — `BackupPath` restores AGENTS.md only; `FileBackups` handles opencode.json. opencode.json is absent from `CreatedPaths`, so no double-record / double-remove. Prior-file case restores (does not remove): verified by `TestRun_OpenCode_RollbackManifest_PriorFile`.
- **Regex determinism:** Anchored + QuoteMeta'd; newest-snapshot selection is `sort.Strings` then last element, deterministic across the 100-iteration determinism tests.
- **gofmt/vet:** `go vet ./...` clean; the four fixed files (`inject.go`, `resolve.go`, `init.go`, `templates.go`) are gofmt-clean. `gofmt -l` flags only `internal/config/model_test.go` — a struct-alignment nit pre-dating this branch (commit `12ff9cc`, not in this branch's diff), unrelated to these fixes.

### Hermeticity proof

Before: `md5sum ~/.config/opencode/opencode.json` = `9f92fd6ea1ac2a4f234dc35ef6d9da43`, backups = 0.
After full `go test ./... -count=1`: `9f92fd6ea1ac2a4f234dc35ef6d9da43`, backups = 0. Unchanged.

### Test result

`go build ./...` ok; `go vet ./...` clean; `go test ./... -count=1` all packages `ok` (cmd/archon, internal/{agent,config,initcmd,opencode,scaffold,status,tui,version}, skills). No production code modified; nothing committed.
