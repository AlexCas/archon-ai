# Delta Spec: archive-before-pr → [[harness-workflow]]

<!-- Link convention: [[capability]] wikilinks for capability identity.
     Source of truth after archive: openspec/specs/harness-workflow/spec.md -->

Change: `archive-before-pr`
Slice: 1 (single-PR flow, docs+skills only)
Affects: [proposal.md](../../proposal.md)

---

## ADDED Requirements

<!-- ADDED, not MODIFIED: harness-workflow/spec.md has no prior explicit
     terminal-ordering requirement; this introduces one. -->

### Requirement: Terminal Phase Ordering (Single-PR Flow)

The harness-workflow MUST enforce archive-before-PR ordering in the single-PR flow:
after `judge` passes, the archive step MUST run (and its commit MUST be staged on the
change branch) BEFORE the PR is opened. This ensures the archive commit travels inside
the change's own PR with no trailing archive-only commit.

Judge gating is UNCHANGED: archive MUST NOT run until judge passes.

The archive-internal operation order MUST be preserved:

1. Spec merge (delta spec merged into `openspec/specs/{domain}/spec.md`)
2. Folder move (`openspec/changes/{name}/` → `openspec/changes/archive/YYYY-MM-DD-{name}/`)
3. `archon map --backfill` + `archon map --check`
4. `SESSION_STATUS.md` move into the archived folder (within the same commit)

All four steps MUST be staged into a single archive commit. PR-open is EXTERNAL to
this sequence — the PR is opened only after the archive commit is on the branch.

`SESSION_STATUS.md` MUST remain root-resident from session start through judge
completion. It MUST be moved into the archived change folder only when the archive
commit is being staged, satisfying the `[[session-status-contract]]` invariant
("root-resident during work, moved at archive"). The move happening before PR-open
(instead of after) does NOT violate this invariant.

#### Scenario: Judge passes, archive runs, then PR is opened

- GIVEN the change has reached `judge` phase with status `completed` (judge passed)
- AND `SESSION_STATUS.md` exists at the repository root
- WHEN the orchestrator begins the terminal sequence
- THEN archive operations run (spec merge, folder move, archon map, SESSION_STATUS.md move)
- AND all archive operations are staged into a single archive commit on the change branch
- AND ONLY AFTER the archive commit is staged, the PR is opened
- AND the archive commit is part of the change's PR (no separate archive-only commit after PR creation)

#### Scenario: Archive is not attempted before judge passes

- GIVEN the change is in `verify` phase with status `completed`
- WHEN the orchestrator attempts to start archive
- THEN `harness-workflow` returns `blocked`
- AND the response indicates `judge` must complete before `archive` can run
- AND no archive operations are performed

#### Scenario: SESSION_STATUS.md stays root-resident through judge

- GIVEN the change has passed `verify` and is entering `judge`
- AND `SESSION_STATUS.md` exists at the repository root
- WHEN judge runs and completes (pass or fail)
- THEN `SESSION_STATUS.md` remains at the repository root
- AND it is NOT moved until the archive commit is being staged

#### Scenario: SESSION_STATUS.md moves in the archive commit (before PR-open)

- GIVEN judge has passed and archive is running
- WHEN the archive commit is being staged
- THEN `SESSION_STATUS.md` is moved from the repository root into
  `openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`
- AND this move is part of the same archive commit as the spec merge and folder move
- AND the PR is opened only AFTER this commit is staged
- AND `SESSION_STATUS.md` is no longer present at the repository root after the commit

#### Scenario: Archive-internal ordering is preserved

- GIVEN judge has passed and archive is executing
- WHEN the four archive sub-operations run
- THEN spec merge completes before folder move begins
- AND folder move completes before `archon map --backfill` is invoked
- AND `archon map --check` runs after `--backfill` completes
- AND `archon map --check` passes before `SESSION_STATUS.md` is moved
- AND PR-open is not invoked until all four sub-operations and the archive commit are done

#### Scenario: Chained-PR flow is out of scope (deferred to slice 2)

- GIVEN a change using the chained-PR flow (multiple stacked PRs)
- WHEN the terminal phase sequence is evaluated
- THEN the archive-before-PR ordering rule from this requirement DOES NOT apply
- AND the harness-workflow treats chained-PR archive ownership as undefined / slice-2 pending
- AND no blocking or enforcement of archive position is applied to chained-PR flows

---

## Non-Goals (Slice 1)

- **Chained-PR archive ownership** — which PR in a stacked/chained sequence owns the
  archive move is deferred to slice 2. The single-PR rule above MUST NOT be applied
  to chained flows. This gap is intentional and tracked; see `proposal.md`.
- **`archon archive` CLI command** — archive remains skill/LLM-driven. An optional
  `archon archive` command is a future concern, not part of this slice.
- **Judge gating semantics** — judge must still pass before archive; this requirement
  does NOT change when or how judge passes, only what happens after.

---

## Consistency Note

Three phase-order copies narrate the terminal sequence and MUST be kept in sync by
the apply phase:

1. `CLAUDE.md` — Phase Order section and gate narrative
2. `skills/harness-workflow/SKILL.md` — terminal-phase and PR-relative ordering narrative
3. `openspec/specs/harness-workflow/spec.md` — this spec (source of truth)

The spec is the source of truth. CLAUDE.md and the SKILL are downstream. At archive,
this delta merges into `openspec/specs/harness-workflow/spec.md` and the apply phase
MUST update all three copies consistently.
