# Tasks: stacked-pr-archive (slice 2b of #93)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~200–280 |
| 400-line budget risk | Low–Medium (well under 400; estimate may approach 280 with optional N) |
| Chained PRs recommended | No — single PR (see recursion note below) |
| Suggested split | Single PR, 5 internal work units |
| Delivery strategy | single-pr-default |
| Chain strategy | N/A (docs-and-skills only; no Go) |

**Decision needed before apply:** No
**Budget note:** This change is itself the one that resolves the Stacked-to-Main
archive policy. It is delivered as a single PR — a noted recursion: the change
being delivered is ABOUT chain-strategy, yet must not be chained. The estimated
diff (spec: ~80 lines net, sdd-tasks: ~30, chained-pr: ~40, chaining-details: ~15,
sdd-apply: ~15, branch-pr: ~10, pointer lines: ~15 total) lands at roughly 200–280
lines. If the actual diff during apply approaches 400, apply MUST STOP and flag
before continuing. Do NOT raise the budget or split into chained PRs without an
explicit approval from the review gate.

> **STOP checkpoint (docs-only hard rule):** If ANY task in this list would require
> touching a `.go` file, STOP immediately and surface it to the orchestrator. This
> change is docs-and-skills only (confirmed by explore and design). See task 5.6 for
> the final gate.

---

## Suggested Work Units

| Unit | Goal | Files | Edit IDs | Est. lines |
|------|------|-------|----------|------------|
| 1 | Source of truth — live spec | `openspec/specs/harness-workflow/spec.md` | A1, A2 | ~80 |
| 2 | Crux — strategy selection home | `skills/sdd-tasks/SKILL.md` | T1, T2 | ~30 |
| 3 | Full-rule skill + mechanics note | `skills/chained-pr/SKILL.md`, `skills/chained-pr/references/chaining-details.md` | B, C, E | ~55 |
| 4 | Backstop + optional breadcrumb | `skills/sdd-apply/SKILL.md`, `skills/branch-pr/SKILL.md` | O, N | ~25 |
| 5 | Pointer lines (LAST) | `CLAUDE.md`, `skills/harness-workflow/SKILL.md` | G, H | ~15 |
| 6 | Verification | (read-only; no file writes) | V1–V7 | 0 |

All six units land in a single PR. Unit 6 is a doc-consistency check, not a file edit.

---

## Work Unit 1: Source of truth — live spec

> Edits A1 and A2. Do these first. Canonical wording is established here; all
> downstream edits copy the three testable facts from the merged spec:
> (1) Trigger = archive-before-PR in effect (openspec/hybrid).
> (2) Decision = MUST NOT select pure Stacked-to-Main; select FBC + notify; up-front
>     at sdd-tasks strategy selection.
> (3) Late Stacked-to-FBC conversion unsanctioned, no recovery; engram unaffected.
> Do NOT proceed to Unit 2 until A1 and A2 are applied and the exact header
> "Stacked-to-Main Archive Convergence" is present in the live spec.

- [x] **1.1 (Edit A1) — REMOVE the deferral scenario from `openspec/specs/harness-workflow/spec.md`.**

  Grep for the exact header:
  ```
  #### Scenario: Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)
  ```
  Delete the entire 7-line block that begins with that header. The full block to
  remove (grep for the opening line and confirm the five GIVEN/WHEN/THEN/AND/AND
  lines follow):
  ```
  #### Scenario: Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)

  - GIVEN a change using the Stacked-to-Main strategy (individual child PRs each target `main`)
  - WHEN the terminal phase sequence is evaluated
  - THEN the Feature Branch Chain archive rule from this requirement DOES NOT apply
  - AND the harness-workflow treats Stacked-to-Main archive ownership as undefined / slice-2b pending
  - AND no blocking or enforcement of archive position is applied to Stacked-to-Main flows
  ```

  This is the ONLY deletion inside the FBC requirement. Every other line in the FBC
  requirement — the header, three clauses, SESSION_STATUS paragraph, 4-step
  archive-internal order, and all six retained scenarios — stays byte-identical.

  **Verify:** `git diff openspec/specs/harness-workflow/spec.md` shows exactly 7
  deleted lines corresponding to this scenario and zero other deletions or changes
  within the FBC requirement body.

- [x] **1.2 (Edit A2) — APPEND the ADDED "Stacked-to-Main Archive Convergence" requirement.**

  Locate the last retained scenario of the FBC requirement. Grep for the final line:
  ```
  - AND tracker PR merge to `main` is not invoked until all four sub-operations and the archive commit are done
  ```
  Immediately after that line (and after any trailing blank line), append the entire
  new requirement verbatim from
  `openspec/changes/stacked-pr-archive/specs/harness-workflow/spec.md` lines 151–234.
  The block starts with:
  ```
  ### Requirement: Stacked-to-Main Archive Convergence
  ```
  and ends after the fifth scenario:
  ```
  #### Scenario: Archive-before-PR not in effect — Stacked-to-Main is unaffected
  ```

  The five scenarios to include verbatim:
  1. Archive-before-PR active, Stacked-to-Main requested — orchestrator converges to FBC at sdd-tasks
  2. Pure Stacked-to-Main + archive-before-PR is unsupported — no partial main merges before owning ref
  3. After convergence, FBC archive rule governs — no new mechanics
  4. Late Stacked-to-FBC conversion is out of scope — stranded slices hazard
  5. Archive-before-PR not in effect — Stacked-to-Main is unaffected

  **Verify:** Grep the merged spec for the exact header:
  ```
  ### Requirement: Stacked-to-Main Archive Convergence
  ```
  Confirm it is present and all five scenario headers appear after it. Confirm the
  single-PR requirement (grep `### Requirement: Terminal Phase Ordering (Single-PR Flow)`)
  is byte-identical — ZERO changes within it.

---

## Work Unit 2: Crux — strategy selection home

> Edits T1 and T2. THE CRUX. The converge-to-FBC decision lands here, BEFORE the
> user's chain choice is cached. This guarantees the cached `Chain strategy` value
> is already FBC when archive-before-PR is in effect; all downstream phases inherit
> the corrected value. Apply MUST confirm T1 fires before the "Cache the user's
> choice" step. Literal guard values are NOT changed.

- [x] **2.1 (Edit T1 — CRUX) — INSERT converge precondition as step 4a in `skills/sdd-tasks/SKILL.md`.**

  Grep for the step-4 header:
  ```
  **Ask the user which chain strategy to use**
  ```
  Also locate the immediately-following step-5 header (grep for):
  ```
  Cache the user's choice
  ```
  Insert the following new step 4a as a bolded paragraph or numbered step BETWEEN
  step 4 and step 5 — specifically, it must appear BEFORE any "Cache" sentence:

  ```
  4a. **Archive-before-PR convergence gate (MANDATORY).** If archive-before-PR is in
     effect for this session — i.e. the artifact store is `openspec` or `hybrid`
     (from the SDD session preflight) — the orchestrator MUST NOT select pure
     **Stacked PRs to main**. Stacked-to-Main ships each slice independently to
     `main`, so no single un-merged ref can own the archive commit; the
     archive-before-PR invariant cannot be satisfied. In that case the orchestrator
     MUST select **Feature Branch Chain** (which supplies the tracker branch that
     owns the archive commit) and MUST notify the user that
     `Stacked-to-Main + archive-before-PR is unsupported, so Feature Branch Chain was
     selected`. This decision is up-front and decisive: made here, before any child
     PR is opened. A late Stacked→FBC conversion (after slices already merged to
     `main`) is NOT sanctioned and has no recovery procedure — see the
     `harness-workflow` spec requirement "Stacked-to-Main Archive Convergence." When
     the artifact store is `engram` (archive-before-PR not in effect), Stacked-to-Main
     is unaffected and remains a valid choice.
  ```

  **Critical ordering check:** Confirm the inserted step 4a appears textually BEFORE
  the "Cache the user's choice" step. If any reordering is needed, fix it. Do NOT
  alter the literal guard values.

- [x] **2.2 (Edit T2) — ADD constraint sentence after the `Chain strategy` fenced forecast block in `skills/sdd-tasks/SKILL.md`.**

  Grep for the fenced block containing:
  ```
  Chain strategy: stacked-to-main|feature-branch-chain|size-exception|pending
  ```
  (This string appears at two locations in the file — one in the forecast table, one
  earlier. Target the one that appears inside the forecast/guard-contract block,
  currently around line 174. Apply must handle BOTH occurrences consistently, but the
  constraint sentence belongs immediately after the FORECAST block's closing fence, not
  after the earlier occurrence.)

  After the closing fence of the forecast block (before the next prose paragraph), add:
  ```
  When archive-before-PR is in effect (`openspec`/`hybrid`), `Chain strategy` MUST NOT
  be `stacked-to-main`; the convergence gate (step 4a) sets it to
  `feature-branch-chain` instead. `stacked-to-main` remains valid only when the
  artifact store is `engram`.
  ```

  **Verify:** The literal pipe-delimited guard string
  `stacked-to-main|feature-branch-chain|size-exception|pending` is UNCHANGED.

---

## Work Unit 3: Full-rule skill + mechanics note

> Edits B, C, E. Mirror the three testable facts from the merged spec.
> Edit D (sdd-tasks step-2 clause) is SKIPPED per approved decision.
> Do NOT proceed to Unit 4 until all three facts appear in chained-pr/SKILL.md.

- [x] **3.1 (Edit B) — REPLACE the deferral Hard Rule bullet in `skills/chained-pr/SKILL.md`.**

  Grep for the opening of the deferral bullet:
  ```
  **Stacked-to-Main archive ownership** is NOT yet defined — deferred to slice 2b.
  ```
  Replace that entire bullet (all lines that form this bullet through the line ending
  `do not enforce archive position on Stacked-to-Main flows.`) with the full converge
  rule:

  ```
  - **Stacked-to-Main + archive-before-PR converges to Feature Branch Chain.** When
    archive-before-PR is in effect (artifact store `openspec`/`hybrid`), pure
    Stacked-to-Main is unsupported: it ships slices independently to `main`, leaving
    no un-merged ref to own the archive commit. At `sdd-tasks` strategy selection the
    orchestrator MUST select **Feature Branch Chain** instead and notify the user.
    After convergence the change is FBC and the Feature Branch Chain archive rule
    above governs — zero new archive mechanics. A late Stacked→FBC conversion (after
    slices already merged to `main`) strands those slices and is NOT sanctioned; there
    is no recovery procedure. When archive-before-PR is NOT in effect (`engram`),
    Stacked-to-Main is unaffected and does not converge. Full rule:
    `[[harness-workflow]]` "Stacked-to-Main Archive Convergence".
  ```

  The Feature Branch Chain Hard Rule bullet above this one stays VERBATIM.

- [x] **3.2 (Edit C) — UPDATE the Decision Gates table row for Stacked in `skills/chained-pr/SKILL.md`.**

  Grep for the table row:
  ```
  | PR >400, each slice can land independently | Use Stacked PRs to main. |
  ```
  Replace that single row with two rows:
  ```
  | PR >400, each slice can land independently, archive-before-PR NOT in effect (`engram`) | Use Stacked PRs to main. |
  | PR >400, each slice independent, BUT archive-before-PR in effect (`openspec`/`hybrid`) | Converge to Feature Branch Chain (Stacked-to-Main unsupported here). |
  ```

  The existing FBC row in the Decision Gates table is NOT touched.

- [x] **3.3 (Edit E) — REPLACE the Stacked deferral note in `skills/chained-pr/references/chaining-details.md`.**

  Grep for the opening of the deferral note:
  ```
  Archive ownership for Stacked PRs to main is NOT yet defined (deferred to slice
  ```
  Replace that entire note (through its closing line ending `to host the archive commit.`) with:

  ```
  Archive ownership for Stacked PRs to main: when archive-before-PR is in effect
  (`openspec`/`hybrid`), pure Stacked-to-Main is unsupported — there is no tracker
  branch to host the archive commit, so the orchestrator converges to Feature Branch
  Chain at `sdd-tasks` strategy selection and the FBC archive step (above) applies.
  When archive-before-PR is not in effect (`engram`), Stacked-to-Main is unaffected.
  Full rule: `[[harness-workflow]]` "Stacked-to-Main Archive Convergence" and the
  `chained-pr` skill.
  ```

  The FBC steps 6–7 and the Stacked ASCII/paragraph above this note are NOT touched.

---

## Work Unit 4: Backstop + optional breadcrumb

> Edit O (REQUIRED backstop) + Edit N (OPTIONAL breadcrumb, included per approved
> decision). Edit J (sdd-archive Timing breadcrumb) is SKIPPED per approved decision.

- [x] **4.1 (Edit O — REQUIRED backstop) — ANNOTATE the `stacked-to-main` bullet in `skills/sdd-apply/SKILL.md`.**

  Grep for the exact bullet:
  ```
  - `stacked-to-main`: each PR targets the previous PR's branch (or `main` after the previous merges).
  ```
  Append a scoped sentence on the next line (part of the same bullet, indented or on
  a new line after the existing text):
  ```
  This value is only valid when archive-before-PR is NOT in effect (`engram`). If the
  tasks artifact carries `Chain strategy: stacked-to-main` while archive-before-PR is
  in effect (`openspec`/`hybrid`), the convergence gate in `sdd-tasks` was skipped:
  STOP and return `blocked` with `Stacked-to-Main + archive-before-PR is unsupported;
  converge to feature-branch-chain before opening any child PR` (see
  `[[harness-workflow]]` "Stacked-to-Main Archive Convergence"). Do not merge any
  child PR to `main`.
  ```

  The `feature-branch-chain` bullet and archive-commit note that follow are NOT
  touched. This backstop returns a `blocked` string matching the existing blocked-message
  pattern — it is NOT a Go error.

- [x] **4.2 (Edit N — OPTIONAL breadcrumb, INCLUDED) — ADD cross-ref in `skills/branch-pr/SKILL.md`.**

  Grep for step 5b text (from slice 2):
  ```
  Child PRs carry no
     archive commit.
  ```
  (or the equivalent final line of the step 5b paragraph). Immediately after that
  sentence, append:
  ```
  (A change with archive-before-PR in effect that would otherwise be Stacked-to-Main
  has already converged to Feature Branch Chain at `sdd-tasks`; there is no
  Stacked-to-Main archive path — see `[[harness-workflow]]` "Stacked-to-Main Archive
  Convergence".)
  ```

  If the exact step 5b closing text differs from the grep above (apply MUST re-locate
  it), adjust the anchor accordingly. Step 5a (single-PR precondition) is NOT touched.

---

## Work Unit 5: Pointer lines

> Edits G and H. Do these LAST. Pointer lines quote the exact header
> "Stacked-to-Main Archive Convergence" — do not write them until A2 is confirmed
> and the header string is final in the merged spec. Each pointer is 1–2 sentences;
> no narrative body is restated. The single-PR sentence and the FBC pointer in each
> file stay VERBATIM.

- [x] **5.1 (Edit G) — REPLACE the deferral pointer in `CLAUDE.md`.**

  Grep for the final sentence of the existing deferral pointer:
  ```
  Stacked-to-Main archive ownership is deferred to slice 2b.
  ```
  Keep all text on the preceding line(s) (the FBC sentence ending with
  `"Terminal Phase Ordering (Feature Branch Chain)" and the \`chained-pr\` skill).`)
  byte-identical. Replace only the deferral sentence with:
  ```
  When archive-before-PR is in effect, pure Stacked-to-Main is unsupported: the
  orchestrator converges to Feature Branch Chain at `sdd-tasks` strategy selection
  (full rule: `openspec/specs/harness-workflow/spec.md` "Stacked-to-Main Archive
  Convergence" and the `chained-pr` skill).
  ```

  **Literal string check:** The pointer MUST contain the exact string
  `Stacked-to-Main Archive Convergence`. The single-PR line above and FBC pointer
  before it are NOT touched.

- [x] **5.2 (Edit H) — REPLACE the deferral pointer in `skills/harness-workflow/SKILL.md`.**

  Grep for the final sentence of the existing deferral pointer:
  ```
  Stacked-to-Main archive ownership is deferred to slice 2b.
  ```
  Keep all text on the preceding line(s) (the FBC pointer ending with
  `...Chain)" and the \`chained-pr\` skill.`) byte-identical. Replace only the
  deferral sentence with:
  ```
  When archive-before-PR is in effect, pure Stacked-to-Main is unsupported and
  converges to Feature Branch Chain at `sdd-tasks` strategy selection. Full rule:
  `openspec/specs/harness-workflow/spec.md` "Stacked-to-Main Archive Convergence" and
  the `chained-pr` skill.
  ```

  **Literal string check:** The pointer MUST contain the exact string
  `Stacked-to-Main Archive Convergence`. The single-PR terminal note and FBC terminal
  note before it are NOT touched.

---

## Work Unit 6: Verification

> Read-only. No file writes. Apply performs these checks before marking tasks
> complete and signaling ready for verify. Reference design §5 (V1–V7).

- [x] **6.1 (V1) — Single-PR and FBC requirement bodies unchanged (regression guard).**

  Run: `git diff openspec/specs/harness-workflow/spec.md`

  Confirm:
  - ZERO changes within the single-PR requirement (grep
    `### Requirement: Terminal Phase Ordering (Single-PR Flow)`) — header, clauses,
    and all its scenarios byte-identical.
  - Within the FBC requirement, the ONLY change is the deletion of the 7-line
    Stacked-to-Main deferral scenario (Edit A1). Header, three clauses, SESSION_STATUS
    paragraph, 4-step archive-internal order, and all six retained scenarios are
    byte-identical.
  - The appended "Stacked-to-Main Archive Convergence" requirement (Edit A2) follows
    the last retained FBC scenario.

- [x] **6.2 (V2) — Up-front decision present in `sdd-tasks` (the crux).**

  Grep `skills/sdd-tasks/SKILL.md` for `archive-before-PR` and `Feature Branch Chain`:
  - Step 4a (Edit T1) MUST state all three: MUST-NOT-select-Stacked AND MUST-select-FBC
    AND notify the user.
  - Step 4a MUST appear textually BEFORE the "Cache the user's choice" step.
  - The T2 constraint sentence MUST be present after the fenced forecast block.
  - The literal pipe-delimited guard string
    `stacked-to-main|feature-branch-chain|size-exception|pending` is UNCHANGED.

- [x] **6.3 (V3) — Pointer lines quote the exact header.**

  Grep `CLAUDE.md` and `skills/harness-workflow/SKILL.md` for the literal string:
  ```
  Stacked-to-Main Archive Convergence
  ```
  Confirm:
  - The string appears in each file and matches the appended spec requirement header
    character-for-character.
  - `chained-pr` is named as the skill home in each pointer.
  - Neither pointer restates the requirement body — length check: 1–2 sentences each.

- [x] **6.4 (V4) — The 5 new scenarios reflected in prose.**

  Verify each ADDED spec scenario maps to at least one skill sentence:

  | ADDED spec scenario | Must appear in |
  |---|---|
  | Converges to FBC at sdd-tasks | `sdd-tasks` step 4a (T1); `chained-pr` Hard Rule (B) + Decision Gate row (C) |
  | Pure Stacked + archive-before-PR unsupported — no partial main merges | `sdd-apply` backstop (O); `chained-pr` Decision Gate row (C) |
  | After convergence, FBC rule governs — no new mechanics | `chained-pr` Hard Rule (B); `chaining-details` (E) |
  | Late Stacked-to-FBC out of scope — stranded slices hazard | `sdd-tasks` step 4a (T1, "not sanctioned, no recovery"); `chained-pr` Hard Rule (B) |
  | Archive-before-PR not in effect — Stacked unaffected (engram) | `sdd-tasks` step 4a (T1); `chained-pr` Decision Gate row (C); `chaining-details` (E); pointers (G, H) |

  If any cell is missing, locate the gap and fix before signaling complete.

- [x] **6.5 (V5) — Deferral fully replaced (zero hits).**

  Grep the repo for:
  - `deferred to slice 2b`
  - `NOT yet defined`
  - `slice-2b`

  MUST return zero hits in these five files:
  `openspec/specs/harness-workflow/spec.md`,
  `skills/chained-pr/SKILL.md`,
  `skills/chained-pr/references/chaining-details.md`,
  `CLAUDE.md`,
  `skills/harness-workflow/SKILL.md`.

  Each of those files now states the converge-to-FBC decision (§3 map in design).

- [x] **6.6 (V6) — SESSION_STATUS root-residency holds through convergence (confirm no edit needed).**

  Read `skills/_shared/session-status-contract.md` Archive bullet (around lines 22–28).
  Confirm it still reads "root-resident through the integrated judge, moved on the
  tracker branch at archive" — unchanged from slice 2 and correct for a converged FBC
  change. No edit is expected here. If V6 finds a gap, surface it before proceeding.

- [x] **6.7 (V7 — HARD GATE) — No Go files touched.**

  Run: `git diff --name-only | grep '\.go$'`

  This MUST return empty. If any `.go` file appears in the diff, STOP immediately
  and surface it to the orchestrator. This is a docs-and-skills-only change; no Go
  path is expected or acceptable (design §7 open question 3).

---

## state.yaml update

After all tasks are marked complete and verification passes, update
`openspec/changes/stacked-pr-archive/state.yaml`:

```yaml
change: stacked-pr-archive
phase: tasks
status: completed
updated_at: "2026-08-01T00:00:00Z"
phases_completed:
  - explore
  - propose
  - spec
  - design
  - tasks
```
