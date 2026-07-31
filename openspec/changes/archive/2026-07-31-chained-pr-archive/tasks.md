# Tasks: chained-pr-archive (slice 2 of #93)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~310–390 |
| 400-line budget risk | Medium (may touch the ceiling) |
| Chained PRs recommended | No (approved scope: single PR) |
| Suggested split | Single PR, 5 internal work units |
| Delivery strategy | single-pr-default |
| Chain strategy | N/A (docs-only; no code) |

**Decision needed before apply:** No
**400-line budget risk note:** This change defines archive rules for chained PRs but is
itself delivered as a single PR — a noted irony. With 10 target files and the full
ADDED requirement block (~100 lines in the spec alone), the diff is estimated at
310–390 lines. If the actual diff approaches 400, apply MUST stop and flag before
continuing; do not raise the budget or split into chained PRs without an explicit
approval from the review gate.

> **STOP checkpoint (docs-only hard rule):** If ANY task in this list would touch a
> `.go` file, STOP immediately and surface it to the orchestrator. This change is
> docs-and-skills only (confirmed by explore). Do not attempt any code edit. See
> task 5.5 for the final gate.

---

## Suggested Work Units

| Unit | Goal | Files | Edit IDs | Est. lines |
|------|------|-------|----------|------------|
| 1 | Source of truth — live spec | `openspec/specs/harness-workflow/spec.md` | A | ~110 |
| 2 | Full-rule skill — chained-pr | `skills/chained-pr/SKILL.md`, `skills/chained-pr/references/chaining-details.md` | B, C, D, E, F | ~80 |
| 3 | Mechanics skills | `skills/sdd-archive/SKILL.md`, `skills/branch-pr/SKILL.md`, `skills/sdd-apply/SKILL.md`, `skills/_shared/session-status-contract.md` | J, K, L, M, N, O, P, Q, R | ~110 |
| 4 | Pointer lines | `CLAUDE.md`, `skills/harness-workflow/SKILL.md` | G, H, I | ~20 |
| 5 | Verification | (read-only; no file writes) | V1–V6 | 0 |

All five units land in a single PR. Unit 5 is a doc-consistency check, not a file edit.

---

## Work Unit 1: Source of truth — live spec

> Edit A. Do this first. Canonical wording is established here; all downstream
> edits in Units 2–4 copy the three testable facts from what lands here.
> Do NOT proceed to Unit 2 until this edit is applied and verified against the
> delta spec byte-for-byte.

- [x] **1.1 (Edit A)** Edit `openspec/specs/harness-workflow/spec.md` lines 183–189.

  Locate and REMOVE the existing blanket-deferral scenario (lines 183–189):
  ```
  #### Scenario: Chained-PR flow is out of scope (deferred to slice 2)

  - GIVEN a change using the chained-PR flow (multiple stacked PRs)
  - WHEN the terminal phase sequence is evaluated
  - THEN the archive-before-PR ordering rule from this requirement DOES NOT apply
  - AND the harness-workflow treats chained-PR archive ownership as undefined / slice-2 pending
  - AND no blocking or enforcement of archive position is applied to chained-PR flows
  ```

  In place of those 7 lines, append the entire ADDED block from
  `openspec/changes/chained-pr-archive/specs/harness-workflow/spec.md` lines 19–123
  verbatim. The block starts with:

  ```
  ### Requirement: Terminal Phase Ordering (Feature Branch Chain)
  ```

  and ends after the 7th scenario:

  ```
  #### Scenario: Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)
  ```

  The resulting spec MUST contain immediately after the single-PR requirement
  (lines 112–181 of the original file — those lines are NOT touched) the new
  requirement block with all 7 scenarios. No line within 112–181 may be altered.

  **Verify:** `git diff openspec/specs/harness-workflow/spec.md` shows ZERO deletions
  or changes within the original lines 112–181. Confirm the canonical header is
  present verbatim:
  `### Requirement: Terminal Phase Ordering (Feature Branch Chain)`

---

## Work Unit 2: Full-rule skill — chained-pr

> Edits B, C, D, E, F. Mirror the three testable facts from the merged spec:
> (i) owning branch = tracker, (ii) integrated judge gates archive AND tracker merge,
> (iii) Stacked-to-Main deferred to slice 2b.
> Edit B and C are REQUIRED. Edit D is INCLUDED per approved decision.
> Do NOT proceed to Unit 3 until all three testable facts appear in chained-pr/SKILL.md.

- [x] **2.1 (Edit B)** Edit `skills/chained-pr/SKILL.md` line 24.

  REPLACE the single bullet at line 24:
  ```
  - The archive-before-PR single-PR rule does NOT apply to chained/stacked flows. Which PR owns the archive move in a stacked sequence is a slice-2 non-goal (deferred); do not enforce archive position on chained PRs.
  ```

  With the two-bullet Hard Rule pair:
  ```
  - **Feature Branch Chain archive ownership:** the archive commit (spec merge,
    folder move, `archon map`, `SESSION_STATUS.md` move) is staged on the
    **tracker branch**, AFTER the integrated judge passes on the tracker and
    BEFORE the tracker PR merges to `main`. It travels inside the tracker PR — never
    a child PR diff, never a separate archive-only PR. Archive-internal order and
    one-commit staging are unchanged from the single-PR flow (`[[harness-workflow]]`
    "Terminal Phase Ordering (Feature Branch Chain)").
  - **Stacked-to-Main archive ownership** is NOT yet defined — deferred to slice 2b.
    Stacked-to-Main has no tracker branch; do not apply the Feature Branch Chain
    archive rule to it and do not enforce archive position on Stacked-to-Main flows.
  ```

- [x] **2.2 (Edit C)** Edit `skills/chained-pr/SKILL.md` — Execution Steps.

  REPLACE the current Step 6 (line 43):
  ```
  6. Keep tracker PR draft/no-merge until all child PRs are reviewed and integrated.
  ```

  With Step 6 rewritten and Steps 7–8 appended:
  ```
  6. Keep the tracker PR draft/no-merge until all child PRs are reviewed and merged
     into the tracker branch.
  7. After the LAST child PR merges into the tracker branch, run the **integrated
     judge on the tracker branch** (evaluate all change files as a unified whole via
     `[[harness-judge]]`). The integrated judge is the gate for BOTH the archive step
     and the tracker merge. If it fails, re-apply on the tracker and re-judge before
     proceeding — do NOT archive and do NOT merge the tracker.
  8. Once the integrated judge passes, stage the archive commit on the tracker branch
     (spec merge → folder move → `archon map --backfill`/`--check` STOP gate →
     `SESSION_STATUS.md` move, one commit), THEN merge the tracker PR to `main`. See
     `[[sdd-archive]]` for the archive mechanics.
  ```

- [x] **2.3 (Edit D)** Edit `skills/chained-pr/SKILL.md` line 21 — add a trailing
  clause to the Feature Branch Chain Hard Rule.

  Append to the end of the existing line 21 bullet (after
  `"later children target the immediate parent branch."`):
  ```
   The tracker PR is also the archive commit's home (see Execution Steps 7–8).
  ```

  Result: the line reads `"... later children target the immediate parent branch.
  The tracker PR is also the archive commit's home (see Execution Steps 7–8)."`

- [x] **2.4 (Edit E)** Edit `skills/chained-pr/references/chaining-details.md`
  — Feature Branch Chain steps (lines 27–34) and ASCII diagram note.

  REPLACE current step 5 at line 33:
  ```
  5. Merge/integrate children in order; merge the tracker only after the chain is complete.
  ```

  With step 5 rewritten and steps 6–7 appended:
  ```
  5. Merge/integrate children in order into the tracker branch.
  6. After the last child merges, run the integrated judge on the tracker branch; it
     must pass before archive or tracker merge (gate for both).
  7. Stage the archive commit on the tracker branch (spec merge, folder move,
     `archon map`, `SESSION_STATUS.md` move — one commit), THEN merge the tracker PR
     to `main`. Archive lives inside the tracker PR, not a child PR.
  ```

  Then ADD one line immediately after the closing ` ``` ` of the ASCII diagram block
  (after line 25):
  ```
  The tracker branch (`feat/my-feature`) owns the archive commit: it is staged there
  after the integrated judge passes, before the tracker merges to `main`.
  ```

- [x] **2.5 (Edit F)** Edit `skills/chained-pr/references/chaining-details.md`
  — Stacked PRs to Main section.

  After the paragraph "After a parent PR merges, rebase/retarget the next PR so GitHub
  shows only the current slice." (line 45), ADD one line:
  ```
  Archive ownership for Stacked PRs to main is NOT yet defined (deferred to slice
  2b); there is no tracker branch to host the archive commit.
  ```

---

## Work Unit 3: Mechanics skills

> Edits J, K, L, M, N, O, P, Q, R (in this order). Every edit is an "IF Feature
> Branch Chain" scoped clause alongside preserved single-PR text. No single-PR
> sentence is deleted. Edit order: sdd-archive → branch-pr → sdd-apply →
> session-status-contract.
> Do NOT proceed to Unit 4 until all mechanics files consistently name:
> "owning branch = tracker for Feature Branch Chain; integrated judge gates
> archive-then-tracker-merge."

- [x] **3.1 (Edit J)** Edit `skills/sdd-archive/SKILL.md` — Timing paragraph
  (lines 89–97).

  The current Timing paragraph opens with: "Archive runs AFTER judge passes and
  BEFORE the PR is opened."

  REPLACE the Timing paragraph (lines 89–97) with a flow-aware version that:
  - Keeps the first sentence verbatim: "Archive runs AFTER judge passes and BEFORE
    the PR is opened."
  - Adds the Feature Branch Chain variant sentence after the first sentence:
    `In the Feature Branch Chain flow, "judge" means the integrated judge on the
    tracker branch, the owning branch is the tracker branch (not the change branch),
    and "before the PR is opened" becomes "before the tracker PR merges to \`main\`".`
  - Generalizes "change branch" to "owning branch" in the ONE archive commit sentence.
    Current: "Stage the results of … into ONE archive commit on the change branch"
    New: "Stage the results of Step 2 (spec merge), Step 3 (folder move), Step 3b
    (`archon map --backfill` + `--check`), and Step 3c (`SESSION_STATUS.md` move)
    into ONE archive commit on the owning branch (change branch for single-PR,
    tracker branch for Feature Branch Chain)."
  - Keeps all remaining sentences in the paragraph unchanged (PR-open is EXTERNAL,
    archive-internal order unchanged).

- [x] **3.2 (Edit K)** Edit `skills/sdd-archive/SKILL.md` — Step 3c body
  (approximately line 177, after the MOVE block).

  After the line:
  ```
  SESSION_STATUS.md  → openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md
  ```
  (approximately line 180), ADD one sentence:
  ```
  In the Feature Branch Chain flow this move happens on the tracker branch, staged
  into the tracker archive commit (before the tracker merges to `main`), not on an
  individual child branch.
  ```

- [x] **3.3 (Edit L)** Edit `skills/sdd-archive/SKILL.md` — Step 3d (lines 190–208).

  Two sub-edits:

  a. REPLACE the Step 3d heading at line 192:
     `**IF mode is \`openspec\` or \`hybrid\`:** Stage and commit all archive changes onto the change branch as ONE commit:`
     With:
     `**IF mode is \`openspec\` or \`hybrid\`:** Stage and commit all archive changes onto the **owning branch** (change branch for single-PR; tracker branch for Feature Branch Chain) as ONE commit:`

  b. REPLACE precondition bullet 4 (line 204–205):
     `4. This commit satisfies the branch-PR precondition: the archive commit MUST
        be staged on the change branch BEFORE the PR is opened (single-PR flow).`
     With:
     `4. This commit satisfies the branch-PR precondition: the archive commit MUST
        be staged on the owning branch BEFORE the change is opened/merged — on the change
        branch before the PR is opened (single-PR flow), or on the tracker branch
        before the tracker PR merges to \`main\` (Feature Branch Chain flow).`

- [x] **3.4 (Edit M)** Edit `skills/sdd-archive/SKILL.md` — Step 4 checklist
  (lines 210–221).

  PRESERVE lines 220–221 VERBATIM:
  ```
  - [ ] Archive commit created on the change branch (Step 3d) with subject `chore(archive): archive {change-name}`, authored solely by the user's git account
  - [ ] PR has NOT been opened before this archive commit is staged (single-PR flow)
  ```

  After line 221, ADD:
  ```
  - [ ] Feature Branch Chain flow ONLY: the archive commit is created on the
    **tracker branch**, and the **tracker PR has NOT been merged to `main` before
    the archive commit is staged on the tracker branch**; the integrated judge
    passed on the tracker before archive ran.
  ```

- [x] **3.5 (Edit N)** Edit `skills/branch-pr/SKILL.md` — Workflow steps (lines 31–41).

  Current step 5 (line 35–37):
  ```
  5. If this is an SDD single-PR change, the archive commit (spec merge, folder
     move, SESSION_STATUS.md move, `archon map`) MUST already be staged on the
     branch before opening the PR.
  ```

  Keep step 5 VERBATIM and ADD step 5b immediately after it:
  ```
  5b. If this is an SDD Feature Branch Chain change, the integrated judge MUST have
     passed on the tracker branch and the archive commit MUST already be staged on the
     **tracker branch** before the tracker PR is merged to `main`. Child PRs carry no
     archive commit.
  ```

  Do not renumber steps 6–8 (a `5b` sub-step is acceptable and lower-risk).

- [x] **3.6 (Edit O)** Edit `skills/sdd-apply/SKILL.md` — chain-strategy
  branch-targeting note (approximately lines 96–98).

  Locate the `feature-branch-chain` bullet (line 98). After its closing sentence
  ("child PR diffs must stay focused on only the current work unit and must never
  target `main` directly."), ADD one sentence on a new line:
  ```
  For `feature-branch-chain`, the archive commit is NOT part of any child slice; it is
  staged on the tracker branch after the integrated judge passes and before the
  tracker merges to `main` (see `[[sdd-archive]]` and `[[chained-pr]]`).
  ```

- [x] **3.7 (Edit P)** Edit `skills/sdd-apply/SKILL.md` — Workload / PR Boundary
  paragraph (approximately lines 272–276).

  Locate the paragraph that begins "For a single-PR change, apply's work ends with
  the implementation ready for archive-then-PR." (line 273). After that paragraph
  (after line 276), ADD a sibling sentence:
  ```
  For a Feature Branch Chain change, each apply batch delivers one child work unit;
  the archive commit belongs to the tracker integration (staged after the integrated
  judge, before the tracker merges), never to a child apply batch.
  ```

- [x] **3.8 (Edit Q)** Edit `skills/sdd-apply/SKILL.md` — Rules block
  (approximately line 292).

  Locate the rule at line 292:
  `- In the single-PR flow, the PR is opened only after the archive commit is staged
  on the branch; apply's scope ends before archive`

  Keep that rule VERBATIM and ADD a chain-aware sibling rule after it:
  ```
  - In the Feature Branch Chain flow, the archive commit is staged on the tracker
    branch after the integrated judge passes and before the tracker PR merges to
    `main`; it is never part of a child apply batch and never appears in a child PR
    diff.
  ```

- [x] **3.9 (Edit R)** Edit `skills/_shared/session-status-contract.md` —
  Archive bullet (lines 22–24).

  REPLACE the Archive bullet (lines 22–26, from `- **Archive**:` through
  `remove the root file.`) with:
  ```
  - **Archive**: during `sdd-archive`, MOVE the file into the archived change folder
    (`openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`) as part of
    the staged archive commit — before the PR is opened in the single-PR flow, or on
    the **tracker branch** before the tracker PR merges to `main` in the Feature
    Branch Chain flow — then delete it from the root. It stays root-resident through
    the integrated judge in the chain flow. In Engram-only mode, store its final
    contents as an observation and remove the root file.
  ```

  The "one file per session" invariant at lines 27–29 is NOT modified.

---

## Work Unit 4: Pointer lines

> Edits G, H, I. These are done LAST so the pointer lines can quote the
> now-final spec header string byte-for-byte. Confirm the header is present in
> the merged spec before writing these lines.

- [x] **4.1 (Edit G)** Edit `CLAUDE.md` — Phase Order section (after line 30).

  Line 30 currently reads:
  `In the single-PR flow the archive commit is staged BEFORE the PR is opened, so
  archive history travels inside the change's PR.`

  Keep line 30 VERBATIM. ADD a new paragraph immediately after line 30:
  ```
  In the Feature Branch Chain flow the archive commit is staged on the tracker branch
  after the integrated judge passes and before the tracker merges to `main` (full
  rule: `harness-workflow` spec "Terminal Phase Ordering (Feature Branch Chain)" and
  the `chained-pr` skill). Stacked-to-Main archive ownership is deferred to slice 2b.
  ```

  **Literal string check:** the pointer text MUST contain exactly:
  `Terminal Phase Ordering (Feature Branch Chain)` — confirm it matches the merged
  spec requirement header character-for-character.

- [x] **4.2 (Edit H)** Edit `skills/harness-workflow/SKILL.md` — after the single-PR
  terminal note (lines 23–25).

  Lines 23–25 currently read:
  ```
  **Terminal ordering (single-PR)**: in the single-PR flow, the sequence after `judge`
  passes is: archive commit staged (spec merge, folder move, `archon map`,
  `SESSION_STATUS.md` move) → PR opened. Judge gating is unchanged.
  ```

  Keep lines 23–25 VERBATIM. ADD a sibling paragraph immediately after line 25:
  ```
  **Terminal ordering (Feature Branch Chain)**: in the Feature Branch Chain flow the
  archive commit is staged on the **tracker branch** after the integrated judge
  passes on the tracker and before the tracker PR merges to `main`. Full rule:
  `harness-workflow` spec "Terminal Phase Ordering (Feature Branch Chain)" and the
  `chained-pr` skill. Stacked-to-Main archive ownership is deferred to slice 2b.
  ```

- [x] **4.3 (Edit I)** Edit `skills/harness-workflow/SKILL.md` — Rules block,
  SESSION_STATUS.md bullet (line 158).

  REPLACE line 158:
  ```
  - `SESSION_STATUS.md` is updated on every transition and is MOVED into the archived
    change folder during `sdd-archive`, within the staged archive commit, before the
    PR is opened.
  ```

  With a flow-aware version:
  ```
  - `SESSION_STATUS.md` is updated on every transition and is MOVED into the archived
    change folder during `sdd-archive`, within the staged archive commit — before the
    PR is opened in the single-PR flow, or on the **tracker branch** before the
    tracker PR merges to `main` in the Feature Branch Chain flow. See
    `session-status-contract`.
  ```

---

## Work Unit 5: Verification

> Read-only. No file writes. Apply performs these checks before marking tasks
> complete and signaling ready for verify. Reference design §5 (V1–V6).

- [x] **5.1 (V1) Single-PR regression guard.**

  Run: `git diff openspec/specs/harness-workflow/spec.md`

  Confirm ZERO changes within the original lines 112–181 (the single-PR requirement
  "Terminal Phase Ordering (Single-PR Flow)" and its 6 scenarios). Only lines 183–189
  should be removed and new content appended after line 181.

  Also confirm:
  - `CLAUDE.md` line 30 is byte-identical to its pre-edit value.
  - `skills/harness-workflow/SKILL.md` lines 23–25 are byte-identical.
  - `skills/sdd-archive/SKILL.md` lines 220–221 remain present and unchanged
    (the single-PR checklist rows).
  - No single-PR clause in `branch-pr`, `sdd-apply`, or `session-status-contract`
    was deleted or reworded.

- [x] **5.2 (V2) Spec ↔ chained-pr full-rule alignment.**

  Verify that all three testable facts appear in BOTH `openspec/specs/harness-workflow/spec.md`
  AND `skills/chained-pr/SKILL.md`:
  - (i) owning branch = tracker branch
  - (ii) integrated judge on tracker gates archive AND tracker merge
  - (iii) Stacked-to-Main deferred to slice 2b

  If any fact is missing from either file, fix it before proceeding.

- [x] **5.3 (V3) Pointer lines resolve.**

  Grep for the literal string `Terminal Phase Ordering (Feature Branch Chain)` in
  `CLAUDE.md` and `skills/harness-workflow/SKILL.md`. Confirm:
  - The string appears in each file and matches the merged spec requirement header
    exactly (character-for-character).
  - `chained-pr` is named as the skill home in each pointer line.
  - Neither pointer line restates the requirement body — length check: each pointer
    paragraph must be 1–2 sentences (no multi-bullet requirement narrative).

- [x] **5.4 (V4) Delta scenarios — 1:1 coverage map.**

  Verify that each of the 7 delta scenarios is reflected in at least one skill
  sentence per the design §5 V4 table:

  | Delta scenario | Must appear in |
  |---|---|
  | Integrated judge passes, archive staged, tracker merges to main | `chained-pr` Step 8 (Edit C); `chaining-details` step 7 (Edit E); `sdd-archive` Timing (Edit J) |
  | Archive blocked before integrated judge passes | `chained-pr` Step 7 (Edit C); `branch-pr` step 5b (Edit N); `sdd-archive` Step 4 checklist row (Edit M) |
  | Integrated judge runs on tracker (all changes integrated) | `chained-pr` Step 7 (Edit C); `chaining-details` step 6 (Edit E) |
  | SESSION_STATUS stays root-resident through integrated judge | `session-status-contract` Archive bullet (Edit R); `harness-workflow` Rules bullet (Edit I) |
  | SESSION_STATUS moves in tracker archive commit | `sdd-archive` Step 3c (Edit K); `session-status-contract` (Edit R) |
  | Archive-internal ordering preserved on tracker | `sdd-archive` Timing (Edit J) — order sentence untouched; `chained-pr` Step 8 (Edit C) |
  | Stacked-to-Main out of scope (slice 2b) | `chained-pr` Hard Rule (Edit B); `chaining-details` note (Edit F); pointer lines (Edits G, H) |

  If any cell is missing, locate the gap and fix before signaling complete.

- [x] **5.5 (V5) Stacked-to-Main deferral present and explicit.**

  Confirm a Stacked-to-Main deferral statement (not just silent omission) appears in:
  - `openspec/specs/harness-workflow/spec.md` — scenario 7 (Edit A)
  - `skills/chained-pr/SKILL.md` — Hard Rule second bullet (Edit B)
  - `skills/chained-pr/references/chaining-details.md` — Stacked section note (Edit F)
  - `CLAUDE.md` — pointer paragraph (Edit G)
  - `skills/harness-workflow/SKILL.md` — pointer paragraph (Edit H)

  Also confirm: nowhere does the Stacked-to-Main deferral text say "all chains
  deferred" or "slice-2 pending" in unscoped language (the old blanket wording is
  replaced, not just augmented).

  **STOP checkpoint — Go code:** Run:
  ```
  git diff --name-only | grep '\.go$'
  ```
  This MUST return empty. If any `.go` file was modified, STOP and surface it to
  the orchestrator before marking this task complete.

- [x] **5.6 (V6) Link integrity check (deferred to archive).**

  The wikilinks `[[harness-workflow]]`, `[[sdd-archive]]`, `[[chained-pr]]`,
  `[[harness-judge]]`, `[[session-status-contract]]` will be verified by
  `archon map --check` during the archive step. No action needed here, but note:
  if any of these wikilinks are renamed or removed during apply, surface it
  immediately — do not defer the breakage.

---

## state.yaml update

After all tasks are marked complete and verification passes, update
`openspec/changes/chained-pr-archive/state.yaml`:

```yaml
change: chained-pr-archive
phase: tasks
status: completed
updated_at: "2026-07-31T00:00:00Z"
phases_completed:
  - explore
  - propose
  - spec
  - design
  - tasks
```
