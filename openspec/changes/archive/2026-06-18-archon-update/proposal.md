# Proposal: First-class `archon update` command

## Intent

Refreshing skills today only happens via `init`, which rebuilds `.archon/config.yaml`
from scratch (`init.go:45-113,87,89`) — resetting `created_at`, user
`models`/`playwright`/`mutation_testing`/`judge`, and `skill_inventory`, and risking an
orchestrator-template clobber. Per-skill versions are also fake: `buildConfig` hardcodes
`"1.0"` (`init.go:195-203`) while real frontmatter is at 2.0/3.0. Users need a safe,
version-aware refresh that preserves config and never rewrites CLAUDE.md.

## Scope

### In Scope
- Factor a reusable skill-refresh path out of `initcmd.Run` (extract-to-global + symlink
  refresh + inventory build), leaving `writeTemplate`/`ErrTemplateExists` untouched.
- Make `skill_inventory[].version` truthful by reading real `metadata.version` frontmatter.
- New `archon update` command: extends `DetectVersionGaps` (added / changed / orphaned);
  `--check` (dry-run, report-only, no writes); `--prune` (opt-in orphan removal).
- PRESERVE `models`, `playwright`, `mutation_testing`, `judge`, `created_at`, `agent`;
  refresh only `harness_version`, `skill_count`, `skill_inventory`. NEVER call `writeTemplate`.
- WARN-ONLY on copy-mode project installs (skill path is a real dir, not a symlink).
- `archon status`: add an "update available" hint via the same version-gap detection.

### Out of Scope
- Auto re-linking copy-mode projects (warn only).
- Tracking the shared global dir in the rollback manifest (`update` stays forward-only).
- TUI "update" action; rewriting `Extract` to be selective beyond changed/added skills.

## Capabilities

> Contract with sdd-spec.

### New Capabilities
- `harness-update`: safe, version-aware skill refresh — diff (added/changed/orphaned),
  `--check`/`--prune`, config preservation, template never rewritten, copy-mode warning,
  and the `status` "update available" hint.

### Modified Capabilities
- `harness-init`: `skill_inventory` versions MUST reflect real `SKILL.md` frontmatter
  (no hardcoded `"1.0"`); the skill-refresh path becomes a shared routine init reuses.

## Approach

Approach B (exploration). Slice 1 — foundation: extract `refreshSkills` + share
`extractVersion` so both init and update build a truthful inventory. Slice 2 — the
command: load config, compute gaps, report or apply, preserve user sections, warn on
copy-mode, opt-in prune. Reuse `Extract`, `SymlinkOrCopy`, `Load`/`Save`, `Clone`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/archon/main.go:39-46` | New | Register `newUpdateCmd`; wire `--check`/`--prune`/`--agent` |
| `internal/initcmd/init.go:45-113` | Modified | Factor shared `refreshSkills` out of `Run` |
| `internal/initcmd/init.go:195-203` | Modified | `buildConfig` reads real frontmatter version |
| `internal/scaffold/version.go:19-71` | Modified | Productionize + add orphaned detection |
| `internal/config/config.go` | Modified | Preserve-then-patch via `Load`/`Clone`/`Save` |
| `internal/status` | Modified | "update available" hint |
| `internal/initcmd` (new) | New | `update` command logic |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Shared global dir blast radius — prune breaks other symlinked projects | Med | `--prune` opt-in/conservative; output states machine-wide scope |
| Copy-mode projects silently stale | Med | Detect non-symlink skill path; WARN explicitly |
| Legacy configs say `"1.0"` → "all changed" on first diff | High | Treat config `"1.0"` as unknown; reconcile against frontmatter |
| `init` regression from refactor | Med | Golden/behavior tests around shared refresh path |

## Rollback Plan

`update` is forward-only (global dir not in rollback manifest). Revert the code change
to remove the command; re-run `archon init --force` to restore prior skill state. No
config migration is written, so reverting the binary leaves `.archon/config.yaml` valid.

## Dependencies

- None external. Builds on existing `Extract`, `SymlinkOrCopy`, `DetectVersionGaps`, config `Load/Save/Clone`.

## Success Criteria

- [ ] `archon update` refreshes skills WITHOUT modifying CLAUDE.md/AGENTS.md or resetting user config.
- [ ] `--check` reports added/changed/orphaned and writes nothing.
- [ ] `skill_inventory` versions match real `SKILL.md` frontmatter after init and update.
- [ ] Copy-mode installs produce a clear warning; `--prune` removes orphans only when passed.
- [ ] `archon status` surfaces an "update available" hint when embedded > installed.
