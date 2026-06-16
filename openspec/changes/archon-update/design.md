# Design: First-class `archon update` command

## Context

`init` is the only path that refreshes skills today, and it rebuilds config from
scratch (`init.go:87-91`) and hardcodes `version: "1.0"` (`init.go:198-202`). We add a
safe, version-aware `archon update` that reuses init's extract+symlink+inventory work
without ever calling `writeTemplate` and without resetting user config.

## 1. Shared refresh routine

Extract the skill-side body of `Run` (`init.go:76-85`) plus a truthful inventory
builder into one helper, leaving template/openspec/rollback in init only.

New in `internal/initcmd/refresh.go`:

```go
type RefreshResult struct {
    GlobalSkillsDir  string
    ProjectSkillsDir string
    Extracted        []string
    Inventory        []config.SkillInventory
}

// refreshSkills extracts embedded skills to the machine-wide global dir, links
// them into the project dir, and builds a truthful inventory. No template/config writes.
func refreshSkills(homeDir, projectDir, agentName string, embeddedFS fs.FS) (*RefreshResult, error)
```

`Run` becomes: guard template → ensureAgentDir → `refreshSkills` → `buildConfig` (now
takes `res.Inventory`) → Save → rollback → writeTemplate → openspec. `update` calls
`refreshSkills` then patches an existing config. `writeTemplate`/`ErrTemplateExists`
stay out of the shared path (decision table D1).

## 2. Truthful versions

Promote `extractVersion` (`version.go:73`) to an exported `scaffold.ExtractVersion`
(reuse, do not duplicate). Add `scaffold.SkillVersion(embeddedFS, name) string` that
reads `name/SKILL.md` and returns `ExtractVersion(...)`, or `""` if absent. The
inventory builder maps each extracted skill to `{Name, Version: SkillVersion(...),
Source: "embedded"}`. Missing version → empty string, never abort (spec edge). Legacy
configs literally storing `"1.0"` are treated as *unknown* in the diff (see §3, D4).

## 3. Extended gap detection (added / changed / orphaned)

`DetectVersionGaps` only walks embedded skills, so it cannot see orphans. Add a sibling
in `version.go` that also walks the installed dir:

```go
type SkillChange struct{ Name, EmbeddedVer, InstalledVer string }
type GapReport struct{ Added, Changed, Orphaned []SkillChange }
func ClassifyGaps(embeddedFS fs.FS, installedDir string) (GapReport, error)
```

- **Added**: embedded skill with no installed `SKILL.md`.
- **Changed**: both present and versions differ; an installed `"1.0"` (or empty) is
  treated as unknown and only counts as changed when embedded differs from it.
- **Orphaned**: installed dir entry with `SKILL.md` but no embedded counterpart.

`DetectVersionGaps` is kept (delegates to `ClassifyGaps` → Added+Changed) so existing
callers/tests are unaffected. `installedDir` is the **global** skills dir.

## 4. `archon update` command

Lives in `internal/initcmd` (decision D2 — reuses `refreshSkills`, agent resolution,
`resolveProjectSkillsDir`). New `internal/initcmd/update.go` exposing `Update(opts
UpdateOptions) (*UpdateResult, error)`; wired in `cmd/archon/main.go` via
`newUpdateCmd` with `--check`, `--prune`, `--agent`.

Flow:
1. `cfg.Load(os.DirFS(projectDir))`; if not exist → actionable error "run archon init
   first", write nothing (spec @error).
2. Resolve agent from `cfg.Agent` (flag override allowed); compute
   `globalSkillsDir` and `projectSkillsDir`.
3. `ClassifyGaps(embeddedFS, globalSkillsDir)`. If empty → "already up to date", write
   nothing.
4. Detect copy-mode: `Lstat(projectSkills/<firstSkill>)`; non-symlink → WARN "project
   needs its own update", do not re-link (spec @edge). Always print machine-wide scope.
5. `--check`: print Added/Changed/Orphaned, return — **no writes**.
6. Apply: `refreshSkills` (re-extract + relink). If `--prune`, `os.RemoveAll` each
   orphan under `globalSkillsDir` *and* its project link.
7. Config preservation: `next := cfg.Clone()`; patch only `Version`,
   `SkillCount`, `SkillInventory` from the refresh result; `next.HomeDir = projectDir`;
   `next.Save()`. `created_at`, `models`, `playwright`, `mutation_testing`, `judge`,
   `agent` ride through `Clone` untouched (spec requirement). **Never** `writeTemplate`.

```
user → main.newUpdateCmd → initcmd.Update
  Update → config.Load ──fail──> "run init first" (no writes)
  Update → ClassifyGaps(embedded, globalDir)
         → none ─────────────> "up to date" (no writes)
         → copy-mode? ───────> WARN (no re-link)
         → --check ──────────> print report ──END (no writes)
         → refreshSkills (Extract + SymlinkOrCopy)
         → --prune ──────────> RemoveAll orphans (global + link)
         → cfg.Clone → patch {Version,SkillCount,Inventory} → Save
```

## 5. status hint

`status.Display` gains an optional gap signal computed in `newStatusCmd`: call
`ClassifyGaps(skills.FS, globalSkillsDir)`; if any Added/Changed, append one line
`Update available — run 'archon update' (N skill(s))`. Keep it a single line; on
detection error, silently skip (status must never fail because of the hint). Pass the
count into a new `status.DisplayWithUpdate(w, cfg, n)` so the pure renderer stays
testable; `Display` delegates with `0`.

## 6. Testing strategy (`go test ./...`)

- **scaffold**: extend `version_test.go` for `SkillVersion` (present/missing) and
  table-driven `ClassifyGaps` covering added, changed, orphaned, `"1.0"`-as-unknown.
- **initcmd**: golden/behavior test asserting `Run` still writes template + config +
  rollback after the refactor (guard regression); `refreshSkills` produces a truthful
  inventory (versions ≠ "1.0"). `Update` tests: missing config → error + no writes;
  no-gap → no writes; `--check` writes nothing; `--prune` removes orphan; Clone
  round-trip preserves `models`/`playwright`/`judge`/`created_at`.
- **command**: `cmd/archon` test for `update --check` (no fs mutation) and copy-mode
  warning, mirroring existing init command tests.

## 7. Decisions

| # | Question | Options | Decision |
|---|----------|---------|----------|
| D1 | Shared-routine boundary | (a) move template too (b) skills-only | **(b)** — `refreshSkills` covers extract+link+inventory; template/openspec/rollback stay in `Run`. Keeps update from ever touching CLAUDE.md. |
| D2 | Where update lives | (a) new `internal/updatecmd` (b) `internal/initcmd` | **(b)** — reuses `refreshSkills`, `resolveProjectSkillsDir`, `detectAgent`; avoids exporting internals. |
| D3 | Prune safety (shared global dir) | (a) prune by default (b) opt-in + scope warning | **(b)** — `--prune` only; always print machine-wide scope so users know other symlinked projects are affected. |
| D4 | Legacy `"1.0"` reconciliation | (a) trust stored version (b) treat `"1.0"`/empty as unknown | **(b)** — diff against embedded frontmatter, not the fake stored value; avoids "everything changed" noise. |
| D5 | Gap detection extension | (a) overload `DetectVersionGaps` (b) add `ClassifyGaps`, keep old as delegate | **(b)** — orphan detection needs the installed-dir walk; preserves existing callers/tests. |

## Open questions

None blocking. Note: `skills.FS` is embedded as `*/SKILL.md` + `_shared`, so `Extract`
only materializes `SKILL.md` per skill — orphan/copy-mode logic keys off `SKILL.md`
presence, which is consistent with what is actually installed.
