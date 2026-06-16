## Exploration: First-class `archon update` command

### Project Type
**Web testing**: not-web

This is a Go CLI ("archon") built with Cobra (`cmd/archon/main.go`). No web framework, no
`package.json`, no `templates/`/`views/`, no browser-facing routes, and no E2E tooling
(playwright/cypress/chromedp). `playwright.enabled` must stay `false`.

### Current State

**The CLI surface** (`cmd/archon/main.go:39-46`) exposes exactly six commands:
`init`, `rollback`, `version`, `status`, `config`, `tui`. There is NO `update`/`upgrade`/`sync`
command (confirmed by grep — only stray matches in template prose).

**Skills are embedded** in the binary via `//go:embed */SKILL.md all:_shared`
(`skills/embed.go:5`). There are **24** skill directories that contain a `SKILL.md`
(including `_shared`, which has its own `SKILL.md`), all of which `Extract` processes.
Note: `cmd/archon/main.go:93` (init dry-run) prints a stale "21 embedded skills"; `CLAUDE.md`
says 24. The real count is 24.

**Extraction is unconditional overwrite.** `scaffold.Extract` (`internal/scaffold/extract.go:10-56`)
iterates embedded dirs and writes every file with `os.WriteFile(..., 0o644)`
(`extract.go:47`) — no version check, no diffing, always clobbers. It returns the list of
extracted skill names.

**Project skills are symlinks into a shared global dir.** `init` extracts to the global dir
`~/.config/opencode/skills` (hardcoded at `internal/initcmd/init.go:76` regardless of agent),
then `scaffold.SymlinkOrCopy` (`internal/scaffold/symlink.go:12-38`) links
`<project>/.<agent>/skills/<skill>` → `<global>/<skill>`. If symlinking fails (EPERM/EINVAL/
ENOSYS/EACCES — e.g. Windows, restricted FS), it falls back to a recursive **copy**
(`symlink.go:37`, `copyDir`). Symlinked projects therefore "auto-update" the instant the global
dir is re-extracted; copy-fallback projects do NOT — their `.<agent>/skills/<skill>` is a
private copy frozen at install time.

**Version detection exists but is dead code.** `scaffold.DetectVersionGaps`
(`internal/scaffold/version.go:19-71`) parses the `metadata.version` frontmatter of embedded vs
installed `SKILL.md` and returns gaps (added skills via `os.IsNotExist`, changed skills via
version mismatch). It is referenced ONLY by `internal/scaffold/version_test.go` — never in
production. It does NOT detect orphaned/removed skills (it only iterates the embedded set, so a
skill that exists installed but was dropped from the binary is never reported).

**Per-skill versioning is non-functional.** `buildConfig` (`internal/initcmd/init.go:195-203`)
hardcodes every inventory entry to `Version: "1.0", Source: "embedded"`. Meanwhile the actual
`SKILL.md` frontmatter versions are already heterogeneous and ahead of "1.0":
e.g. `sdd-apply: 3.0`, `sdd-init: 3.0`, `sdd-verify: 3.0`, `branch-pr: 2.0`,
`judgment-day: 1.4`, several at `2.0`, and the rest at `1.0`. So today's
`skill_inventory[].version` is meaningless and bears no relation to the real installed version.

**Re-running `init` is the only refresh path today, and it is lossy.** `initcmd.Run`
(`internal/initcmd/init.go:45-113`) does NOT reload or merge the existing `.archon/config.yaml`
before rebuilding it. It calls `buildConfig` (line 87) from scratch every time, then `cfg.Save()`
(line 89) overwrites the file. Contrary to the gathered context, there is **no reload that
preserves user model/playwright/mutation/judge settings** — those come solely from the flags
passed on that invocation (defaults when omitted). So `init` (even without `--force`) silently
resets `created_at`, resets `skill_inventory` versions to "1.0", and rewrites
mutation/judge/playwright/models. The ONLY thing guarding `init` is the orchestrator-file check:
`templateFilePath` + `ErrTemplateExists` (`init.go:63-68`), which aborts before any work unless
`OverwriteTemplate`/`--force`. The CLI wraps this with an interactive y/N prompt
(`main.go:130-139`, `confirmOverwrite`). This is exactly the clobber risk the user wants to
avoid: refreshing skills currently requires going through `init`, which is entangled with the
config rebuild and the orchestrator template.

**Rollback is project-local only.** `RollbackManifest` (`internal/config/rollback.go:11-16`)
records `.archon/config.yaml`, `.archon/rollback.json`, and the per-project skill paths
(`init.go:237-251`). It does NOT track the shared global skills dir, so rollback never touches
`~/.config/opencode/skills`.

**Status display** (`internal/status`) prints `cfg.Version` ("Harness Version") and the skill
inventory, so it is the natural surface to reflect post-update state.

### Affected Areas

- `cmd/archon/main.go` — add `newUpdateCmd(...)` and register it in `newRootCmd` (`main.go:39-46`).
  Wire flags (`--check`/dry-run, `--agent`, maybe `--prune`).
- `internal/initcmd/init.go` — refactor the reusable middle of `Run` (extract-to-global +
  symlink-refresh + inventory build) so `update` can reuse it WITHOUT touching
  `writeTemplate` (`init.go:99`/`253-281`) or the `ErrTemplateExists` guard. Today these are
  inlined in one function; `update` needs the skill/config half only.
- `internal/initcmd/init.go:195-203` (`buildConfig`) — stop hardcoding "1.0"; read the real
  `metadata.version` from each embedded `SKILL.md` so the inventory is meaningful (shared logic
  with `extractVersion`). This is a prerequisite for any version-aware reporting.
- `internal/scaffold/version.go` — promote `DetectVersionGaps` to production use; extend it to
  also detect **orphaned** skills (installed-but-no-longer-embedded) and to surface
  added/changed/removed categories for `--check` output. Export `extractVersion` or move it to a
  shared spot so `buildConfig` can reuse it.
- `internal/scaffold/extract.go` — likely fine as-is for the global re-extract; optionally add a
  selective/targeted extract (only changed skills) if approach B wants minimal writes.
- `internal/config/config.go` — `update` must `Load` the existing config, **preserve**
  user-tunable sections (`Models`, `Playwright`, `MutationTesting`, `Judge`, `CreatedAt`,
  `Agent`), refresh only `Version` (harness_version), `SkillCount`, and `SkillInventory`, then
  `Save`. `Clone` (`config.go:78-96`) is a helper for the preserve-then-patch flow.
- `internal/initcmd` symlink helpers (`createSymlinks`, `resolveProjectSkillsDir`,
  `agentBaseDir`) — reused to re-point/refresh links and to detect copy-mode (a project skill
  path that is a real dir, not a symlink).
- `internal/status` — optionally show "update available" (compare embedded harness version /
  per-skill versions against config) and reflect the refreshed inventory.
- `internal/config/rollback.go` — `update` mutates the shared global dir; the project-local
  rollback manifest does not cover it. Decide whether `update` records anything (probably not,
  to keep rollback semantics about init).

### Approaches

1. **(A) Thin update — re-extract + refresh symlinks + patch config, never touch the template**
   — `update` loads the existing config, calls `Extract` into the global dir, re-runs
   `SymlinkOrCopy` for each embedded skill, refreshes `harness_version` + `skill_count` +
   `skill_inventory` (still possibly with real frontmatter versions), preserves all user
   settings, and explicitly never calls `writeTemplate`. Detects and warns when a project skill
   path is a copy (not a symlink) so the user knows that project must be updated per-project.
   - Pros: smallest change; directly solves the stated pain (no template clobber); reuses
     `Extract`/`SymlinkOrCopy`/`Load`/`Save` almost verbatim; low risk of regressing init.
   - Cons: still overwrites everything in the global dir unconditionally (no per-skill diff
     report); no dry-run insight into what changed; leaves orphaned skills behind; `--check`
     would need at least partial version logic anyway.
   - Effort: Low.

2. **(B) Version-aware update — wire up `DetectVersionGaps`, report added/changed/removed,
   support `--check`/dry-run, write only what changed, optional `--prune`** — Build on (A) but
   first fix `buildConfig`/inventory to carry real frontmatter versions, then use (an extended)
   `DetectVersionGaps` to compute a diff: added skills, changed skills (old→new version),
   orphaned skills. `--check` prints the diff and exits 0/nonzero without writing. A real update
   extracts/relinks only changed+added skills, refreshes inventory with true versions, and with
   `--prune` removes orphaned skills from the global dir and project links. Preserves the
   template and all user config as in (A).
   - Pros: makes per-skill versioning actually meaningful; gives users a trustworthy "what will
     change" preview; handles orphans; turns existing dead code (`DetectVersionGaps`) into value;
     idempotent and minimal writes.
   - Cons: more surface to design/test (orphan detection, prune semantics, exit codes); shared
     global dir means a `--check` in one project reflects machine-wide state, which can confuse;
     touching `buildConfig` versioning also subtly changes `init` output (acceptable, arguably a
     fix). Pruning the shared global dir can break OTHER symlinked projects still expecting an
     old skill — prune must be conservative/opt-in.
   - Effort: Medium.

3. **(C) Alternatives** —
   - *Update-as-init-flag* (`archon init --skills-only` / `--update`): reuses the existing
     command but adds a mode that skips the template and config-reset. Pros: one command. Cons:
     overloads `init`'s already-entangled `Run`; the very entanglement (template + config rebuild)
     is what we're trying to escape, so this fights the design rather than separating concerns.
     Effort: Low-Med.
   - *TUI action*: add an "Update skills" action to the existing `tui` (`internal/tui`). Useful
     as a secondary surface but not a substitute for a scriptable/CI-friendly subcommand; the
     core logic is identical to A/B, so this is additive, not an alternative. Effort: Med (UI
     plumbing).

### Recommendation

Adopt **(B) version-aware update**, delivered as a thin command that internally builds on the
(A) mechanics, but implement it in two clean slices:

1. **Foundation**: factor a reusable `refreshSkills` path out of `initcmd.Run` (extract-to-global
   + symlink refresh + inventory build) and fix `buildConfig` to read real `metadata.version`
   from each `SKILL.md` (export/share `extractVersion`). This makes `skill_inventory` truthful
   for both `init` and `update`.
2. **The command**: `archon update` loads config, computes a diff via an extended
   `DetectVersionGaps` (added / changed / orphaned), supports `--check` (report-only, no writes),
   re-extracts + relinks changed/added skills, refreshes `harness_version` + `skill_count` +
   `skill_inventory` while PRESERVING `models`/`playwright`/`mutation_testing`/`judge`/
   `created_at`/`agent`, and NEVER calls `writeTemplate`. Detect copy-mode project skills (real
   dir instead of symlink) and warn that those are per-project copies. Make orphan pruning
   opt-in (`--prune`) and conservative because the global dir is shared.

This directly removes the orchestrator-clobber risk (no template write, no config reset),
resurrects the existing-but-dead version machinery, and keeps `init` semantics intact. Approach
(A) alone is tempting for speed but leaves per-skill versions meaningless and gives the user no
preview — most of B's cost is the `--check`/diff reporting the user explicitly asked for.

### Risks

- **Shared global dir blast radius**: `~/.config/opencode/skills` is machine-wide
  (`init.go:76`). One `archon update` refreshes skills for EVERY symlinked project on the
  machine simultaneously; `--prune` could break other projects pinned to an old skill. Update
  output must make this explicit; pruning must be opt-in.
- **Copy-fallback projects don't auto-update**: any project where symlinking failed has private
  copies under `.<agent>/skills/`. `update` run from the global dir won't refresh them unless the
  command also re-runs `SymlinkOrCopy` per-project and detects the copy case. A user could
  `update` and still be on stale skills without a clear warning.
- **Per-skill version currently hardcoded "1.0"**: existing installed configs already say "1.0"
  for everything, while frontmatter is at 2.0/3.0/etc. The first real diff will look like "all
  skills changed". Need a sane migration/first-run story (treat config "1.0" as unknown and
  reconcile against frontmatter).
- **Orphan handling**: `DetectVersionGaps` only walks the embedded set, so removed skills are
  invisible today. Reporting/pruning orphans requires walking the installed dir too.
- **`init` regression surface**: changing `buildConfig` versioning and factoring out a shared
  refresh path touches the most-tested command; needs golden/behavior tests to avoid regressions.
- **Rollback gap**: the rollback manifest doesn't cover the global dir, so `update` is not
  cleanly reversible by `archon rollback`. Set expectations (update is forward-only) or extend
  the manifest deliberately.

### Ready for Proposal

Yes. The problem is well-shaped and the codebase confirms the core mechanics. Recommend telling
the user: proceed to `propose` with **Approach B** (version-aware update built on thin-update
mechanics, template untouched, user config preserved), delivered in two slices (foundation:
truthful inventory + reusable refresh; then the `update` command with `--check`/`--prune`).
Open decisions to confirm during propose: (1) whether `--prune` is in the first slice or
deferred; (2) copy-mode handling — warn-only vs. auto re-link per project; (3) whether `status`
should also gain an "update available" hint. Project is `not-web`; keep `playwright.enabled`
false.
