# Proposal: opencode Model Assignment Bridge

## Intent

After `archon init` for the opencode agent, models in `.archon/config.yaml` (`models.default` / `models.phases`) have NO effect — opencode runs every phase on its built-in default. Root cause (confirmed in explore): archon writes models only to its private `.archon/config.yaml` (read solely by status + TUI) and never generates the `agent.<name>.model` overlay in `~/.config/opencode/opencode.json`, the ONLY place opencode reads per-agent models. This makes the "Archon Leader" config silently inert.

## Scope

### In Scope
- **Slice 1 (MVP bug fix, ~280-360 LOC):** embedded default-profile overlay (one `mode: primary` orchestrator + one `mode: subagent` per SDD phase, prompts pointing at `~/.config/opencode/skills/sdd-<phase>/SKILL.md`); bare-name→`provider/model` resolver (static map over `KnownModels` + optional `~/.cache/opencode/models.json` read + qualified pass-through); JSON deep-merge with `__replace__` sentinel into `opencode.json` with pre-merge backup; injection decision tree (explicit phase > existing user agent > default fallback); wired into `archon init` gated on `agent == opencode`; backup recorded in rollback (never delete the shared file); "## Phase Models" section in `agentsTemplate`.
- **Slice 2 (multi-profile, ~350-450 LOC):** named SDD profiles, each with per-phase overlays/agents (mirror gentle-ai `inject.go`/`profiles.go`/`sdd-overlay-*.json`); stale-agent cleanup; profile selection; `archon sync` to re-apply after TUI/CLI model edits.

### Out of Scope
- Other agents (claude, codex) — only opencode reads this overlay.
- Auth-aware provider detection, custom-provider merge, reasoning `variant`/effort (archon's `ModelConfig` has no effort field).
- TUI provider/model picker (still bare names; resolver handles them).

## Capabilities

### New Capabilities
- `opencode-overlay`: generate/merge opencode agent+model overlay during init/sync.
- `model-resolution`: resolve bare model names to provider/model IDs.

### Modified Capabilities
None.

## Approach

Add `internal/opencode/` (path helpers, cache loader, `Resolve()`), an embedded `sdd-overlay.json` asset, a JSON deep-merge helper, and an injection step. On `archon init` for opencode: back up `opencode.json`, inject resolved models into the overlay, deep-merge, write, record backup. Slice 2 generalizes to multi-profile + `archon sync`. Reference: gentle-ai at `/home/alexcasdev/Projects/gentle-ai`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/opencode/` | New | Path helpers, cache loader, bare-name resolver |
| `internal/opencode/assets/sdd-overlay*.json` | New | Embedded overlay(s) |
| `internal/opencode/jsonmerge.go` | New | Deep-merge + `__replace__` |
| `internal/opencode/inject.go` | New | 3-case decision tree |
| `internal/initcmd/init.go` | Modified | Backup + inject + merge step (opencode-gated) |
| `internal/initcmd/templates.go` | Modified | "Phase Models" in `agentsTemplate` |
| `internal/config/rollback.go` | Modified | Restore backup, do NOT delete shared file |
| `cmd/archon/main.go` | New (slice 2) | `archon sync` subcommand |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Clobbering shared `opencode.json` | High | Additive deep-merge + pre-merge backup; rollback restores backup, never deletes |
| Model cache missing on fresh machine | Med | Static map fallback + qualified pass-through; init never fails, warn only |
| Wrong provider for ambiguous bare name | Med | Prefer cache; allow qualified `provider/model` override |
| Overlay agent keys drift from skill folders | Med | Keys are `sdd-<phase>`, matching extracted folders; covered by golden test |

## Rollback Plan

`opencode.json` is a SHARED global file. Before merge, copy it to a backup recorded in the rollback manifest. Rollback restores the backup (or removes only if no prior file existed); the file is NEVER added to `CreatedPaths` for deletion. All other artifacts roll back via existing manifest paths.

## Dependencies

- opencode reads `agent.<name>.model` from `~/.config/opencode/opencode.json`.
- gentle-ai reference port (`internal/opencode/models.go`, `filemerge/json_merge.go`, `sdd/inject.go`).

## Success Criteria

- [ ] After `archon init` (opencode) with `--model`/`--model-<phase>`, `opencode.json` has `agent.sdd-<phase>.model` set to resolved `provider/model` IDs.
- [ ] Missing model cache does not fail init; bare names resolve via static map.
- [ ] Pre-existing `opencode.json` content is preserved through the merge; rollback restores the backup.
- [ ] `agentsTemplate` output includes a "Phase Models" section.
- [ ] Slice 2: `archon sync` re-applies edited models; multi-profile overlays generated with stale agents cleaned up.
