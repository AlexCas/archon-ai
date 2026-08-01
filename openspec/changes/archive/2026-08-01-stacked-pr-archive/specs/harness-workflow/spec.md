# Delta Spec: stacked-pr-archive → [[harness-workflow]]

<!-- Link convention: [[capability]] wikilinks for capability identity.
     Source of truth after archive: openspec/specs/harness-workflow/spec.md -->

Change: `stacked-pr-archive`
Slice: 2b (Stacked-to-Main archive ownership — converge-to-FBC decision)
Affects: [proposal.md](../../proposal.md)

---

<!-- Modeling choice: NEW REQUIREMENT, not appended scenarios.

     Rationale: The converge-to-FBC rule has a distinct invariant class —
     it fires at sdd-tasks strategy selection time (upstream of archive), not at
     archive/terminal-sequence time. Appending it to the "Terminal Phase Ordering
     (Feature Branch Chain)" requirement would conflate two concerns: (1) how to
     order archive within an FBC flow (FBC requirement), and (2) whether to enter
     FBC at all when archive-before-PR is active (this requirement). A dedicated
     requirement keeps the gate logic readable and makes the MODIFIED marker below
     unambiguous: only the deferral scenario inside the FBC requirement is removed;
     the FBC requirement body itself is untouched.

     The deferral scenario in the FBC requirement (spec.md:281-287) is superseded
     here via a MODIFIED block that drops it. The converge rule is ADDED as a new
     requirement. Both markers are required: the MODIFIED removes the gap; the ADDED
     states the decision. -->

---

## MODIFIED Requirements

<!-- MODIFIED: "Terminal Phase Ordering (Feature Branch Chain)"
     The single ADDITIVE change is the removal of the deferred/out-of-scope
     Stacked-to-Main scenario (formerly spec.md:281-287). All other requirement
     text, clauses, and scenarios are preserved VERBATIM. Slice 2b replaces that
     deferral with the ADDED requirement below. -->

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

<!-- The deferral scenario that previously appeared here (spec.md:281-287) is
     REMOVED by this delta. It is superseded by the new requirement below.
     No silent gap remains. -->

---

## ADDED Requirements

<!-- ADDED: new requirement for Stacked-to-Main archive convergence.
     Does NOT modify the FBC requirement body above (which stays verbatim).
     Does NOT modify the single-PR requirement (which is unaffected entirely).
     This requirement fires at strategy-selection time (sdd-tasks), upstream of
     the terminal sequence governed by the two preceding requirements. -->

### Requirement: Stacked-to-Main Archive Convergence

When **archive-before-PR is in effect** (artifact store is `openspec` or `hybrid`)
AND the orchestrator is selecting a chain strategy during `sdd-tasks`, the
orchestrator MUST NOT select pure Stacked-to-Main. It MUST select (or silently
convert to) **Feature Branch Chain** instead. This decision is **up-front and
decisive**: made at chain-strategy selection in `sdd-tasks`, before any child PR
is opened.

**Rationale:** Stacked-to-Main ships each slice independently to `main`. By the
time an archive commit is needed there is no single un-merged owning ref; the
archive-before-PR invariant ("archive commit inside the change's PR, not separate")
cannot be satisfied. Feature Branch Chain supplies the required tracker branch.

**Up-front convergence is mandatory:** A late Stacked-to-Main → FBC conversion
(attempted after one or more slices have already merged to `main`) is NOT the
sanctioned path. It would strand already-merged slices on `main` without an
owning tracker ref, violating the archive-before-PR invariant retroactively. The
orchestrator MUST converge at selection time.

**Unsupported combination:** Pure independent-shipping Stacked-to-Main +
archive-before-PR is explicitly **unsupported**. If the user has archive-before-PR
in effect and requests Stacked-to-Main, the orchestrator MUST inform them that this
combination is unsupported and MUST select Feature Branch Chain instead. No partial
`main` merges are permitted before an owning tracker ref exists.

**After convergence:** The change is now Feature Branch Chain. The "Terminal Phase
Ordering (Feature Branch Chain)" requirement governs the archive commit — staged on
the tracker branch, after the integrated judge passes, before the tracker PR merges
to `main`. Zero additional archive mechanics are introduced by this requirement.

**Scope boundary:** When archive-before-PR is NOT in effect (artifact store is
`engram` or `none`, or archive is not applicable to the session), this requirement does NOT
apply. Stacked-to-Main is unaffected; no forced convergence occurs.

#### Scenario: Archive-before-PR active, Stacked-to-Main requested — orchestrator converges to FBC at sdd-tasks

- GIVEN archive-before-PR is in effect (artifact store is `openspec` or `hybrid`)
- AND the orchestrator is selecting a chain strategy during `sdd-tasks`
- AND the user's preference or default would be Stacked-to-Main
- WHEN the orchestrator executes `sdd-tasks` strategy selection
- THEN the orchestrator selects Feature Branch Chain instead of Stacked-to-Main
- AND the orchestrator notifies the user that Stacked-to-Main + archive-before-PR is unsupported and FBC is selected
- AND a tracker/integration branch is created (or planned) before any child PR is opened
- AND the archive step will be governed by the "Terminal Phase Ordering (Feature Branch Chain)" requirement

#### Scenario: Pure Stacked-to-Main + archive-before-PR is unsupported — no partial main merges before owning ref

- GIVEN archive-before-PR is in effect (artifact store is `openspec` or `hybrid`)
- AND the orchestrator attempted to proceed with pure Stacked-to-Main (no tracker branch)
- WHEN any child PR would merge to `main` without a tracker branch in place
- THEN the `sdd-apply` backstop returns `blocked`
- AND the response states that Stacked-to-Main + archive-before-PR is an unsupported combination
- AND no child PR is merged to `main` until the strategy is corrected to Feature Branch Chain
- AND `harness-workflow` directs the orchestrator to converge to FBC before opening child PRs

#### Scenario: After convergence, FBC archive rule governs — no new mechanics

- GIVEN the orchestrator has converged from Stacked-to-Main to Feature Branch Chain at sdd-tasks
- AND the change is now proceeding as a Feature Branch Chain flow
- AND the integrated judge has passed on the tracker branch
- WHEN the terminal phase sequence runs
- THEN the archive step follows the "Terminal Phase Ordering (Feature Branch Chain)" requirement verbatim
- AND no additional archive mechanics are introduced by the Stacked-to-Main convergence
- AND the archive commit is staged on the tracker branch before the tracker PR merges to `main`

#### Scenario: Late Stacked-to-FBC conversion is out of scope — stranded slices hazard

- GIVEN archive-before-PR is in effect
- AND one or more slices have already merged to `main` under Stacked-to-Main
- WHEN the orchestrator attempts a late conversion to Feature Branch Chain
- THEN this conversion path is NOT the sanctioned approach
- AND `harness-workflow` treats this state as an unresolved hazard (stranded slices on `main` without an owning tracker ref)
- AND the orchestrator MUST NOT attempt archive under these conditions without explicit human resolution
- AND this scenario exists to document the hazard, not to define a recovery procedure

#### Scenario: Archive-before-PR not in effect — Stacked-to-Main is unaffected

- GIVEN archive-before-PR is NOT in effect (artifact store is `engram` or `none`, or archive is not applicable)
- AND the orchestrator selects Stacked-to-Main as the chain strategy during `sdd-tasks`
- WHEN `harness-workflow` evaluates the strategy selection
- THEN the orchestrator proceeds with Stacked-to-Main without forced convergence
- AND no FBC tracker branch is required
- AND this requirement does NOT apply

---

## Non-Goals (Slice 2b)

- **No new archive mechanics** — the converge-to-FBC rule reuses the "Terminal Phase
  Ordering (Feature Branch Chain)" requirement verbatim. Archive internals are unchanged.
- **Pure Stacked-to-Main + archive-before-PR is unsupported** — this combination is
  explicitly out of scope; the orchestrator converges to FBC rather than implementing a
  separate Stacked-to-Main archive path.
- **No `archon archive` CLI** — archive remains skill/LLM-driven. An optional
  `archon archive` command is a future concern, not part of this slice.
- **Single-PR behavior unchanged** — the "Terminal Phase Ordering (Single-PR Flow)"
  requirement and its scenarios are VERBATIM and UNAFFECTED by this delta.
- **FBC requirement unchanged** — the "Terminal Phase Ordering (Feature Branch Chain)"
  requirement body and all its scenarios (except removal of the deferral scenario) are
  VERBATIM and UNAFFECTED. The deferral scenario is replaced, not the requirement.

---

## Consistency Note

The full Stacked-to-Main archive convergence rule lives in two places (source of truth +
primary skill):

1. `openspec/specs/harness-workflow/spec.md` — this spec (source of truth; updated
   at archive via spec merge). Contains the new "Stacked-to-Main Archive Convergence"
   requirement and the MODIFIED FBC requirement (deferral scenario removed).
2. `skills/chained-pr/SKILL.md` — replaces the remaining deferral passage with the
   converge-to-FBC rule and up-front selection narrative.

Two additional locations carry a brief pointer line only (the full rule is NOT
duplicated there):

3. `CLAUDE.md` — Phase Order section or chain-strategy note gains a one-line reference
   to the converge-to-FBC rule.
4. `skills/harness-workflow/SKILL.md` — terminal-phase narrative gains a one-line
   pointer to the "Stacked-to-Main Archive Convergence" requirement.

The spec is the source of truth. `CLAUDE.md` and the SKILL files are downstream.
The apply phase MUST update all four consistently: full rule in (1) and (2),
pointer lines in (3) and (4). This mirrors the slice-2 copy fan-out decision.
