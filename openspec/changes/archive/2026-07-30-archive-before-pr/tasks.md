# Tasks: archive-before-pr

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~200–280 |
| 400-line budget risk | Low–Medium |
| Chained PRs recommended | No (approved scope: single PR) |
| Suggested split | Single PR, 4 internal work units |
| Delivery strategy | ask-on-risk |
| Chain strategy | single-pr-default |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: single-pr-default
400-line budget risk: Low–Medium (docs+skills prose; no Go code)

> **Line-delta note**: Each target file receives a small prose insertion (2–10 lines).
> Ten files touched; estimated total delta is ~200–280 lines added/changed across the
> PR diff. This is within the 400-line budget. If the spec merge in task 1.1 is large
> (the full ADDED block from the delta spec is ~60 lines), the cumulative total could
> approach 300 lines — still within budget. Flagged for review gate awareness; no split
> is needed.

### Suggested Work Units

| Unit | Goal | Files | Est. lines |
|------|------|-------|------------|
| 1 | Source of truth + phase-order copies | `openspec/specs/harness-workflow/spec.md`, `CLAUDE.md`, `skills/harness-workflow/SKILL.md` | ~90–120 |
| 2 | Sequencing skills | `skills/sdd-archive/SKILL.md`, `skills/harness-judge/SKILL.md`, `skills/branch-pr/SKILL.md`, `skills/chained-pr/SKILL.md`, `skills/sdd-apply/SKILL.md` | ~80–110 |
| 3 | Shared modules | `skills/_shared/session-status-contract.md`, `skills/_shared/openspec-convention.md` | ~20–30 |
| 4 | Verification | (read-only grep + cross-check; no file writes) | 0 |

All four work units land in a single commit (archive-style) and a single PR. Unit 4
is a doc-consistency verification pass, not a file edit.

---

## Work Unit 1: Source of truth + phase-order copies

> Edit in this order. The spec is the authoritative wording; the two downstream
> copies must mirror it. Do NOT proceed to Work Unit 2 until all three files in
> this unit are consistent.

- [ ] **1.1** Merge the delta spec ADDED block into `openspec/specs/harness-workflow/spec.md`.
  Append the full **"Requirement: Terminal Phase Ordering (Single-PR Flow)"** section
  (including all 6 scenarios verbatim) from
  `openspec/changes/archive-before-pr/specs/harness-workflow/spec.md` into the
  `## Requirements` section of `openspec/specs/harness-workflow/spec.md`, after the
  existing "Automatic map.md Regen" requirement. Preserve all existing requirements
  unchanged. This file is the source of truth; the exact wording applied here is the
  reference for the downstream copies.

- [ ] **1.2** Edit `CLAUDE.md` — three locations (design `:41`):
  - **Phase Order block** (~`:27-28`): After the phase order line
    (`explore → propose → spec → …`), add one sentence:
    *"In the single-PR flow the archive commit is staged BEFORE the PR is opened,
    so archive history travels inside the change's PR."*
  - **SESSION_STATUS rule** (~`:137`): Change "During `archive`, MOVE `SESSION_STATUS.md`
    into the archived change folder alongside the feature artifacts, then remove it
    from the root." to: "During `archive`, MOVE `SESSION_STATUS.md` into the archived
    change folder **as part of the staged archive commit, before the PR is opened**,
    then remove it from the root."
  - **Rules block** (~`:146-155`): Add a new numbered rule:
    *"In the single-PR flow, run archive (spec merge, folder move, SESSION_STATUS.md
    move, `archon map`) as one commit AFTER judge passes and BEFORE opening the PR."*
    Renumber subsequent rules if needed.

- [ ] **1.3** Edit `skills/harness-workflow/SKILL.md` — two locations (design `:42`):
  - **Phase Sequence block** (~`:15-19`): After the sequence diagram, add a short
    "Terminal ordering (single-PR)" note: *"In the single-PR flow, the sequence after
    `judge` passes is: archive commit staged (spec merge, folder move,
    `SESSION_STATUS.md` move, `archon map`) → PR opened. Judge gating is unchanged."*
  - **Last Rules bullet** (~`:154`): Reword "SESSION_STATUS.md` is updated on every
    transition and is MOVED into the archived change folder during `sdd-archive`." to:
    "`SESSION_STATUS.md` is updated on every transition and is MOVED into the archived
    change folder during `sdd-archive`, within the staged archive commit, before the
    PR is opened."

  **Consistency check (apply must confirm before moving on):** All three files now
  each contain, in their own words, (a) judge pass precedes archive, (b) archive is
  one commit, (c) PR opens only after that commit. If any of the three is missing any
  of the three invariants, fix it before continuing.

---

## Work Unit 2: Sequencing skills

> Edit in design order: sdd-archive → harness-judge → branch-pr → chained-pr →
> sdd-apply. Each edit is an instruction/order change only; no code is touched.

- [ ] **2.1** Edit `skills/sdd-archive/SKILL.md` — four locations (design `:43`):
  - **Top of archive step block** (before Step 2, ~`:127`): Insert a new **"Timing"
    note**: *"Archive runs AFTER judge passes and BEFORE the PR is opened. Stage the
    results of Step 2 (spec merge), Step 3 (folder move), Step 3c (SESSION_STATUS.md
    move), and Step 3b (`archon map --backfill` + `--check`) into ONE archive commit on
    the change branch. The PR is opened only after this commit is staged. The
    archive-internal order (merge → move → SESSION_STATUS move → map) is unchanged;
    PR-open is EXTERNAL to this sequence."*
  - **Step 3c** (~`:160-177`): In the openspec/hybrid branch instruction, reword to
    clarify the move happens *within the archive commit staging, before the PR is
    opened*: "MOVE `SESSION_STATUS.md` from the repo root into the archived change
    folder as part of the archive commit staging (before the PR is opened)."
  - **Step 4 verification checklist** (~`:179-190`): Ensure the checklist already
    includes (or add) a line: `- [ ] PR has NOT been opened before this archive commit
    is staged (single-PR flow)`. If any existing checklist item contradicts
    archive-after-PR, correct it.
  - **Rules block** (~`:236-245`): Add a rule: *"Archive is pre-PR in the single-PR
    flow — never open the PR before the archive commit is staged."*

- [ ] **2.2** Edit `skills/harness-judge/SKILL.md` — "advance to archive" language
  (design `:44`). Locate the passages at approximately lines `:19`, `:23`, `:47`,
  `:174`, `:184` where the skill says the orchestrator "advances to archive" on a pass
  verdict. For each, append: *"then, after archive stages its commit on the branch,
  the PR is opened (single-PR flow)."* Do NOT change judge gating logic, pass/fail
  criteria, or retry semantics — only the post-pass ordering note is added.

- [ ] **2.3** Edit `skills/branch-pr/SKILL.md` — Workflow and Commands sections
  (design `:45`):
  - **Workflow list** (~`:30-38`): Before the "Open PR" step, insert a precondition
    bullet: *"If this is an SDD single-PR change, the archive commit (spec merge,
    folder move, SESSION_STATUS.md move, `archon map`) MUST already be staged on the
    branch before opening the PR."*
  - **Commands section** (~`:187-202`, if present): Add a brief one-line note near the
    `gh pr create` command: *"(For SDD single-PR changes, run after the archive commit
    is on the branch.)"*

- [ ] **2.4** Edit `skills/chained-pr/SKILL.md` — Hard Rules block (~`:14-24`)
  (design `:46`). Add a Hard Rule / note: *"The archive-before-PR single-PR rule does
  NOT apply to chained/stacked flows. Which PR owns the archive move in a stacked
  sequence is a slice-2 non-goal (deferred); do not enforce archive position on
  chained PRs. See `proposal.md` for context."* This is the explicit non-goal
  documentation required by the approved scope.

- [ ] **2.5** Edit `skills/sdd-apply/SKILL.md` — PR Boundary section and Rules
  (design `:47`):
  - **Workload / PR Boundary section** (~`:263-271`): Add a note: *"For a single-PR
    change, apply's work ends with the implementation ready for archive-then-PR.
    The archive commit (spec merge, folder move, SESSION_STATUS.md move, `archon map`)
    is staged after judge passes and before the PR is opened — it is NOT part of apply's
    workload."*
  - **Rules block** (~`:285`): Add a rule reinforcing: *"In the single-PR flow, the PR
    is opened only after the archive commit is staged on the branch; apply's scope
    ends before archive."*
  - **No change** to chained/stacked slice behavior.

  > **STOP checkpoint**: If any of the above edits in Work Unit 2 would require
  > touching Go code (e.g., a `.go` file contains archive-after-PR logic), STOP and
  > flag to the orchestrator. The approved slice is docs-and-skills only. Do NOT
  > proceed with any code edit.

---

## Work Unit 3: Shared modules

- [ ] **3.1** Edit `skills/_shared/session-status-contract.md` — Archive lifecycle
  bullet (~`:24-25`) (design `:48`). Reword the **Archive** bullet: *"during
  `sdd-archive`, MOVE the file into the archived change folder
  (`openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`) as part of
  the staged archive commit, before the PR is opened, then delete it from the root.
  In Engram-only mode, store its final contents as an observation and remove the root
  file."* Keep the "root-resident during work / read-first on crash recovery"
  invariants (~`:16-21`) intact and unchanged.

- [ ] **3.2** Edit `skills/_shared/openspec-convention.md` — two locations (design `:49`):
  - **SESSION_STATUS comment block** (~`:26-27`): Add a short archive-then-PR note:
    *"The move into the archived change folder happens as part of the staged archive
    commit, before the PR is opened (single-PR flow)."*
  - **`sdd-archive Moves` row** (~`:58`): Optionally add a parenthetical on the
    `openspec/changes/{change-name}/` row: *"(move is staged into the pre-PR archive
    commit in the single-PR flow)"*. Row semantics unchanged.

---

## Work Unit 4: Verification

> Read-only. No file writes. Apply performs these checks before marking tasks
> complete and signaling ready for verify.

- [ ] **4.1 Cross-copy consistency grep**: Confirm all three phase-order copies narrate
  all three invariants. For each of the three files — `openspec/specs/harness-workflow/spec.md`,
  `CLAUDE.md`, `skills/harness-workflow/SKILL.md` — verify each contains:
  (a) judge pass precedes archive ("judge" AND "archive" in proximity),
  (b) archive is one commit ("one commit" or "single commit" or "staged … commit"),
  (c) PR opens only after that commit ("PR" AND "after" AND "archive" or "commit").
  If any invariant is absent from any copy, fix that copy before proceeding.

- [ ] **4.2 Scenario coverage map** — confirm each of the 6 delta-spec scenarios is
  addressed by at least one edited prose passage:
  | Scenario | Must be covered by |
  |---|---|
  | Judge passes, archive runs, then PR is opened | `sdd-archive` Timing note + `branch-pr` precondition |
  | Archive is not attempted before judge passes | `harness-judge` "advance to archive" text (judge gate unchanged) |
  | SESSION_STATUS.md stays root-resident through judge | `session-status-contract` Archive bullet + `CLAUDE.md` SESSION_STATUS rule |
  | SESSION_STATUS.md moves in the archive commit (before PR-open) | `sdd-archive` Step 3c + `session-status-contract` Archive bullet |
  | Archive-internal ordering is preserved | `sdd-archive` Timing note (merge → move → SESSION_STATUS move → map) |
  | Chained-PR flow is out of scope (deferred to slice 2) | `chained-pr` Hard Rules note (task 2.4) |

- [ ] **4.3 No orphaned old wording**: Search all edited files for any remaining text
  that states archive runs *after* the PR is opened, or that SESSION_STATUS.md moves
  *after* PR-open. None should exist after the edits. If found, fix before signaling
  complete.
  Suggested grep: `grep -rn "after.*PR\|after the PR\|SESSION_STATUS.*after" \`
  `  CLAUDE.md skills/harness-workflow/SKILL.md skills/sdd-archive/SKILL.md \`
  `  skills/harness-judge/SKILL.md skills/branch-pr/SKILL.md skills/chained-pr/SKILL.md \`
  `  skills/sdd-apply/SKILL.md skills/_shared/session-status-contract.md \`
  `  skills/_shared/openspec-convention.md openspec/specs/harness-workflow/spec.md`

- [ ] **4.4 No Go code touched**: Confirm no `.go` files were modified during apply.
  Suggested check: `git diff --name-only | grep '\.go$'` — must return empty.

- [ ] **4.5 Non-goal confirmed documented**: Verify `skills/chained-pr/SKILL.md` contains
  the slice-2 non-goal note (task 2.4) and does NOT enforce archive position on
  chained PRs. Verify no `archon archive` CLI command was added to any file.
