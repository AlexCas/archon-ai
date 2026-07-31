# Delta Spec: chained-pr-archive → [[harness-workflow]]

<!-- Link convention: [[capability]] wikilinks for capability identity.
     Source of truth after archive: openspec/specs/harness-workflow/spec.md -->

Change: `chained-pr-archive`
Slice: 2 (Feature Branch Chain flow, docs+skills only)
Affects: [proposal.md](../../proposal.md)

---

## ADDED Requirements

<!-- ADDED, not MODIFIED: the single-PR Terminal Phase Ordering requirement in
     harness-workflow/spec.md stays VERBATIM. This requirement is strictly
     additive; it defines archive ownership for the Feature Branch Chain
     strategy only. -->

### Requirement: Terminal Phase Ordering (Feature Branch Chain)

In the **Feature Branch Chain** strategy, the archive commit MUST be staged on
the **tracker/integration branch**, AFTER the integrated judge passes on that
branch and BEFORE the tracker PR merges to `main`. The archive commit travels
inside the tracker PR as legitimate integration history; no separate
archive-only PR is needed or permitted.

**Additive clause:** This requirement EXTENDS (does not replace) the
"Terminal Phase Ordering (Single-PR Flow)" requirement. The single-PR rule is
unaffected; this rule governs the Feature Branch Chain strategy exclusively.

**Judge-timing clause:** The integrated judge MUST run on the tracker branch
after all child PRs are merged into it, and MUST pass, BEFORE the tracker PR
merges to `main`. The archive-on-tracker step is only valid once this
integrated judge has passed. An integrated judge that has not yet run, or that
has failed, MUST block both the archive step and the tracker merge.

`SESSION_STATUS.md` MUST remain root-resident from session start through
integrated judge completion, even while child PRs are open and their own
local judges have passed. It MUST be moved into the archived change folder only
when the tracker archive commit is being staged, satisfying the
`[[session-status-contract]]` invariant ("root-resident during work, moved at
archive").

The archive-internal operation order MUST be preserved on the tracker branch
(identical to the single-PR flow; only the owning branch changes):

1. Spec merge (delta spec merged into `openspec/specs/{domain}/spec.md`)
2. Folder move (`openspec/changes/{name}/` → `openspec/changes/archive/YYYY-MM-DD-{name}/`)
3. `archon map --backfill` + `archon map --check`
4. `SESSION_STATUS.md` move into the archived folder (within the same commit)

All four steps MUST be staged into a single archive commit on the tracker
branch. Tracker PR merge to `main` is EXTERNAL to this sequence — the tracker
PR MUST NOT be merged until the archive commit is present on the tracker branch.

#### Scenario: Integrated judge passes on tracker, archive staged, tracker merges to main

- GIVEN a change using the Feature Branch Chain strategy
- AND all child PRs have been merged into the tracker branch
- AND the integrated judge has run on the tracker branch and passed
- AND `SESSION_STATUS.md` exists at the repository root
- WHEN the orchestrator begins the terminal sequence on the tracker branch
- THEN archive operations run (spec merge, folder move, archon map, SESSION_STATUS.md move)
- AND all archive operations are staged into a single archive commit on the tracker branch
- AND the tracker PR is merged to `main` only AFTER the archive commit is staged on the tracker branch
- AND the archive commit is part of the tracker PR (no separate archive-only PR, no polluted child PR diff)

#### Scenario: Archive is blocked on tracker before integrated judge passes

- GIVEN a change using the Feature Branch Chain strategy
- AND all child PRs are merged into the tracker branch
- AND the integrated judge has NOT yet run on the tracker branch (or has failed)
- WHEN the orchestrator attempts to start the archive step on the tracker branch
- THEN `harness-workflow` returns `blocked`
- AND the response indicates the integrated judge must pass on the tracker branch before archive can run
- AND no archive operations are performed

#### Scenario: Integrated judge runs on tracker branch (all changes integrated)

- GIVEN a change using the Feature Branch Chain strategy
- AND at least one child PR carries work that was judged locally before integration
- WHEN the final child PR is merged into the tracker branch
- THEN an integrated judge run MUST be triggered on the tracker branch
- AND the integrated judge evaluates all change files as a unified whole
- AND the integrated judge result is the gate for both the archive step and the tracker merge to `main`

#### Scenario: SESSION_STATUS.md stays root-resident through integrated judge on tracker

- GIVEN a change using the Feature Branch Chain strategy
- AND child PRs are open (or already merged into the tracker branch)
- AND `SESSION_STATUS.md` exists at the repository root
- WHEN the integrated judge runs on the tracker branch and completes (pass or fail)
- THEN `SESSION_STATUS.md` remains at the repository root
- AND it is NOT moved until the tracker archive commit is being staged
- AND no child PR or intermediate merge touches `SESSION_STATUS.md`

#### Scenario: SESSION_STATUS.md moves in the tracker archive commit

- GIVEN the integrated judge has passed on the tracker branch and archive is running
- WHEN the tracker archive commit is being staged
- THEN `SESSION_STATUS.md` is moved from the repository root into
  `openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`
- AND this move is part of the same tracker archive commit as the spec merge and folder move
- AND the tracker PR is merged to `main` only AFTER this commit is staged
- AND `SESSION_STATUS.md` is no longer present at the repository root after the commit

#### Scenario: Archive-internal ordering is preserved on the tracker branch

- GIVEN the integrated judge has passed and archive is executing on the tracker branch
- WHEN the four archive sub-operations run
- THEN spec merge completes before folder move begins
- AND folder move completes before `archon map --backfill` is invoked
- AND `archon map --check` runs after `--backfill` completes
- AND `archon map --check` passes before `SESSION_STATUS.md` is moved
- AND tracker PR merge to `main` is not invoked until all four sub-operations and the archive commit are done

#### Scenario: Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)

- GIVEN a change using the Stacked-to-Main strategy (individual child PRs each target `main`)
- WHEN the terminal phase sequence is evaluated
- THEN the Feature Branch Chain archive rule from this requirement DOES NOT apply
- AND the harness-workflow treats Stacked-to-Main archive ownership as undefined / slice-2b pending
- AND no blocking or enforcement of archive position is applied to Stacked-to-Main flows

---

## Non-Goals (Slice 2)

- **Stacked-to-Main archive ownership** — that strategy has no tracker branch and
  therefore no natural archive home. Archive ownership for Stacked-to-Main is
  deferred to slice 2b. The rule above MUST NOT be applied to Stacked-to-Main flows.
- **`archon archive` CLI command** — archive remains skill/LLM-driven. An optional
  `archon archive` command is a future concern, not part of this slice.
- **Single-PR behavior** — the "Terminal Phase Ordering (Single-PR Flow)" requirement
  and its 6 scenarios are UNCHANGED. This delta is strictly additive.

---

## Consistency Note

The full Feature Branch Chain archive rule lives in two places (source of truth +
primary skill):

1. `openspec/specs/harness-workflow/spec.md` — this spec (source of truth; updated
   at archive via spec merge).
2. `skills/chained-pr/SKILL.md` — replaces the `:24` blanket deferral with the
   Feature Branch Chain rule and tracker ownership narrative.

Two additional locations carry a brief pointer line only (the full rule is NOT
duplicated there):

3. `CLAUDE.md` — Phase Order section gains a one-line reference to the
   chained-archive rule.
4. `skills/harness-workflow/SKILL.md` — terminal-phase narrative gains a one-line
   pointer to the Feature Branch Chain requirement.

The spec is the source of truth. `CLAUDE.md` and the SKILL files are downstream.
The apply phase MUST update all four consistently: full rule in (1) and (2),
pointer lines in (3) and (4).
