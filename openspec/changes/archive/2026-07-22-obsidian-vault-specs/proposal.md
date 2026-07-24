# Proposal: Obsidian-Style Vault for SDD Specs

## Intent

The `openspec/` store is a flat folder-per-thing layout with **zero cross-links** today
(`grep` for `[[` and `](../` returns 0 matches; no `map.md` exists). An AI agent — and a
human — navigating the change/spec graph must read files blindly or grep bare prose
tokens. This change turns `openspec/` into an **Obsidian-style vault** with a single
entry node (`map.md`) that is both an index and a materialized backlink map, plus a
disciplined link convention so the graph is fast and reliable to traverse.

The target root shape:

```
openspec/
  map.md          <- Map of Content (MOC): index + materialized backlink map (entry node)
  changes/        <- as today (active + archive/)
  specs/          <- as today (capabilities: spec.md + {cap}.feature)
```

Gherkin `.feature` files do NOT move or change: they stay beside `spec.md` inside
change/spec folders and are only *referenced* from the `.md`. Go never parses markdown
bodies; the vault work is additive and cannot break the existing build or harness.

## Scope

### In Scope
- A new **vault + link convention** documented as the single source of truth all phase
  skills inherit (new shared module + updates to `openspec-convention.md`).
- `map.md` as a **harness-generated MOC**: capabilities index, active-changes index
  (with phase/status read from `state.yaml`), archive index, and a **materialized
  backlink map** (per capability → the changes that reference it).
- A new, self-contained **Go module** (`internal/`) + CLI subcommand `archon map` that
  walks `openspec/`, regenerates `map.md` between managed markers, rewrites
  boundary-crossing links, and offers a `--check` guard that fails on stale managed
  regions or dangling relative links.
- **Automatic regeneration** of `map.md` after every phase transition (hooked into the
  same transition path used by `harness-workflow` / state updates).
- **Link emission** by phase skills (`sdd-init`, `sdd-propose`, `sdd-spec`, `sdd-design`,
  `sdd-tasks`, `sdd-verify`) in their artifact templates.
- **Archive link rewriting** as a first-class concern in `sdd-archive` + the Go step:
  when a change folder moves, boundary-crossing edges are deterministically rewritten.
- **One-shot backfill** of the 20 existing archived changes in THIS project.

### Out of Scope
- Any change to `.feature` file content or location.
- Any Go parsing of authored artifact prose. Go edits ONLY inside
  `<!-- MAP:START -->` … `<!-- MAP:END -->` managed regions and known link patterns.
- Forcing GitHub-clickability of every link (wikilinks render literally on GitHub by
  design; this is an accepted trade-off for capability-identity stability).
- Backfilling archived changes in OTHER projects that adopt archon — those are
  **forward-only** by default.
- Replacing grep-based navigation entirely; the materialized backlink map complements,
  not removes, the ability to grep stable capability tokens.

## Capabilities

### New Capabilities
- `spec-vault`: the vault convention itself — the root shape (`map.md` + `changes/` +
  `specs/`), the hybrid link convention, the managed-marker policy, and the rule that
  `.feature` files stay put and are only referenced. This is the doc/convention
  capability all phase skills inherit.
- `archon-map`: the deterministic Go module + `archon map` CLI subcommand that
  generates `map.md` (index + backlinks), rewrites boundary-crossing links on archive,
  offers `--check`, supports a one-shot `--backfill` mode, and defaults to forward-only.

### Modified Capabilities
- `openspec-convention` (`skills/_shared/openspec-convention.md`): document the vault
  root shape, `map.md` as entry node, and point to the new `spec-vault` convention.
- `sdd-archive`: rewrite boundary-crossing links and trigger `map.md` regen as part of
  the archive move — the highest-consequence, immutable-audit-trail operation.
- `sdd-init`: seed an initial `map.md` when scaffolding `openspec/`.
- `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-verify`: emit links in their
  artifact templates (wikilinks to capabilities, relative links within the change).
- `harness-workflow` (transition path): fire `archon map` regen after every phase
  transition.

## Approach

### Link convention (LOCKED)
Hybrid, per the explore Human Review Gate:
- **`[[capability]]` wikilinks** for the **stable capability-identity graph** —
  change → the capabilities it touches, and `map.md` → capabilities. These resolve by
  name and survive the archive move with **no rewrite**.
- **Relative links** `[text](../specs/x/spec.md)` for **intra-change navigation**
  (proposal ↔ spec ↔ design ↔ tasks), because those files always move together as a
  unit and stay relatively positioned.

This minimizes the rewrite surface: on archive, only the map→change and change→spec
boundary edges cross the move boundary; the capability-identity graph and intra-change
relative links are untouched.

### Maintenance owner (LOCKED)
A **deterministic Go step** — a small, isolated markdown-writing module — owns `map.md`
regeneration and boundary-link rewriting. It is NOT agent-maintained. Go has never
touched artifact bodies before; this module is surgical: it writes ONLY inside managed
markers and rewrites ONLY known link patterns, never free-form artifact prose.

### `map.md` scope (LOCKED)
Index **AND** materialized backlink map: capabilities index, active-changes index
(phase/status from `state.yaml`), archive index grouped by date, and a per-capability
backlink section listing the changes that reference it.

### Regen trigger (LOCKED)
Both: a standalone **`archon map`** command AND **automatic regeneration after every
phase transition**, hooked into the existing transition path.

### Managed markers (LOCKED)
Go writes ONLY between `<!-- MAP:START -->` … `<!-- MAP:END -->`. A `--check` mode
(for CI/dev) fails if any managed region is stale or if any relative link resolves to a
missing file.

### Backfill (LOCKED)
The Go tool has a one-shot **`--backfill`** mode to fix the 20 existing archived changes
in this project. Default behavior for adopting projects is **forward-only**.

## The Archive Move (Top Risk) — first-class handling

The archive move `changes/{c}/` → `changes/archive/YYYY-MM-DD-{c}/` relocates 6–8 `.md`
files one directory deeper under a renamed parent. This silently breaks:
- Any link pointing INTO the moved folder (e.g. `map.md` → the change, another change).
- Any relative link INSIDE the moved files pointing OUT to `specs/` (shifts one level).

Handling:
1. Wikilink capability-identity edges resolve by name and need **no** rewrite.
2. `sdd-archive` invokes the Go step, which deterministically rewrites the
   boundary-crossing relative edges (map→change, change→spec) as part of the same
   operation, then regenerates `map.md`.
3. `--check` guards against a missed rewrite by failing on any dangling relative link,
   so link rot on the immutable audit trail is caught, not silently accepted.

The archive slice lands **last** and ships with tests over a fixture folder move.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `skills/_shared/spec-vault.md` (new) | New | Vault + link convention, single source of truth |
| `skills/_shared/openspec-convention.md` | Modified | Root shape, `map.md` entry node, pointer to `spec-vault` |
| `internal/mapgen/` (new) | New | Deterministic `map.md` generation + link rewrite/check + backfill |
| `cmd/archon/` | Modified | New `archon map` subcommand |
| `skills/sdd-init/` | Modified | Seed initial `map.md` |
| `skills/sdd-propose,-spec,-design,-tasks,-verify/` | Modified | Emit links in artifact templates |
| `skills/sdd-archive/` | Modified | Trigger link rewrite + map regen on archive move |
| `harness-workflow` transition path | Modified | Auto-run `archon map` after each transition |
| `openspec/map.md` (new) | New | The entry-node MOC for this repo |

## Slice / Chained-PR Plan

Per session preflight (chained-PR strategy `ask-always`, review budget **400 changed
lines**): the explore flagged this exceeds 400 lines as one PR. Four independently
reviewable slices, each within budget:

- **Slice 1 — Vault convention + docs (skills/docs only).** New `spec-vault.md` shared
  module; update `openspec-convention.md`; document the root shape, hybrid link
  convention, managed-marker policy, and `.feature`-stays-put rule. Seed this repo's
  initial `map.md` skeleton (managed markers only). No Go. Smallest, unblocks the rest.

- **Slice 2 — `archon map` Go module (Go, isolated).** New `internal/mapgen/` +
  `archon map` subcommand: index + materialized backlinks generation between managed
  markers, plus `--check`. Ships with unit tests over a fixture `openspec/` tree. No
  skill wiring yet.

- **Slice 3 — Auto-regen + link emission.** Hook `archon map` into the phase-transition
  path (`harness-workflow`); update `sdd-init` (seed `map.md`) and the emitting phase
  skills (`sdd-propose/spec/design/tasks/verify`) to write wikilinks + relative links in
  their templates.

- **Slice 4 — Archive rewrite + backfill (riskiest, lands last).** Add
  boundary-crossing link rewrite to `sdd-archive` + the Go step; add `--backfill`; run
  the one-shot backfill of the 20 existing archived changes. Ships with tests over a
  fixture archive move.

Because the strategy is `ask-always`, the orchestrator confirms the chained split with
the user before opening PRs; each slice is estimated to stay under 400 changed lines and
is reviewable on its own.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Archive move silently breaks boundary-crossing links | High | Wikilink identity needs no rewrite; Go step rewrites the few relative boundary edges deterministically on archive; `--check` fails on any dangling link |
| Go now writes markdown bodies (new capability) | Medium | Confine writes to managed markers + known link patterns only; never touch authored prose; cover with fixture tests |
| Backfilling the 20 archived changes mutates an immutable audit trail | Medium | One-shot, deterministic, reviewed as its own slice (4); adopting projects stay forward-only by default |
| Wikilinks not clickable on GitHub | Low | Accepted trade-off for identity stability; relative intra-change links remain clickable; agent greps stable tokens |
| Two link conventions to teach the skills | Low | Single source of truth in `spec-vault.md`; each phase skill inherits it |
| Managed regions drift from source | Medium | `--check` in CI/dev fails on stale regions; auto-regen after every transition keeps them fresh |

## Rollback Plan

The change is additive. To roll back: delete `openspec/map.md`, remove `internal/mapgen/`
and the `archon map` subcommand, revert the transition-path hook, and revert the skill
template edits. Because Go only wrote inside managed markers, removing those regions
leaves authored artifact prose intact. Backfilled archived changes can be reverted via
git since each slice is a discrete PR.

## Success Criteria

- [ ] `openspec/map.md` exists as the entry node with a capabilities index, active-changes
      index (phase/status from `state.yaml`), archive index, and a per-capability backlink map.
- [ ] `archon map` regenerates `map.md` deterministically and re-runs cleanly (idempotent).
- [ ] `archon map --check` fails on a stale managed region or a dangling relative link.
- [ ] `map.md` auto-regenerates after every phase transition.
- [ ] Phase skills emit `[[capability]]` wikilinks and relative intra-change links per the convention.
- [ ] After an archive move, no relative link is dangling (`--check` passes) and capability
      wikilinks still resolve without rewrite.
- [ ] The 20 existing archived changes are backfilled with correct links; adopting projects
      default to forward-only.
- [ ] No change to any `.feature` file content or location; Go writes only inside managed markers.

## Open Questions (for the review gate)

1. **Module naming**: `internal/mapgen/` and capability `archon-map` are proposals —
   confirm before spec locks them in.
2. **Backlink granularity**: backlink map at capability → change level (proposed). Do we
   also want change → artifact (proposal/spec/design) depth, or is capability-level enough
   for agent navigation?
3. **`--check` in CI**: is there an existing CI workflow to wire `archon map --check` into,
   or is it dev-loop-only for now (explore treated it as CI/dev)?
