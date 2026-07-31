# Design: Chained-PR Archive Ownership (Feature Branch Chain)

<!-- Link convention: [[capability]] wikilinks for capability identity; relative
     links for intra-change navigation. Full rule: skills/_shared/spec-vault.md. -->

Change: `chained-pr-archive` (slice 2 of #93)
Phase: design. Contract: [specs/harness-workflow/spec.md](specs/harness-workflow/spec.md).
Type: docs-and-skills only. NO Go code (explore confirmed no Go enforces terminal
ordering — `harness-workflow` is skill/LLM prose, and `sdd-archive`/`branch-pr`
gates are prose too). Any code implication surfaced during apply is an OPEN QUESTION,
not a design item.

---

## 0. Design principles (apply MUST honor)

- **Additive, never rewriting.** Every single-PR sentence, requirement, and scenario
  stays VERBATIM. Slice 2 only ADDS the Feature Branch Chain rule and REPLACES the
  two blanket-deferral passages that explicitly say "chained is out of scope / slice-2
  pending" (spec scenario `:183-189`, `chained-pr/SKILL.md:24`).
- **Two full-rule homes, two pointer homes.** Full narrative lives in (1) the spec
  and (2) `chained-pr/SKILL.md`. `CLAUDE.md` and `harness-workflow/SKILL.md` get a
  ONE-LINE pointer each — no duplicated narrative (per approved decision +
  Consistency Note `:139-159`).
- **Owning branch is the only variable.** Slice 1's archive-internal order (spec merge
  → folder move → `archon map --backfill`/`--check` STOP gate → `SESSION_STATUS.md`
  move) and one-commit staging are unchanged. For Feature Branch Chain the owning
  branch becomes the **tracker branch** instead of the change branch.
- **Integrated judge is the gate.** Archive-on-tracker is only legal after the
  integrated judge passes on the tracker branch, and the tracker PR merges to `main`
  only after the archive commit is staged. This is prose/LLM guidance (no CLI, no
  new state machine transition).
- **Stacked-to-Main stays deferred, but LOUDLY.** The blanket "all chains out of
  scope" language is replaced by an explicit "Feature Branch Chain defined;
  Stacked-to-Main pending slice 2b" statement wherever the old blanket text lived.

---

## 1. Per-file edit plan

Legend for role tags:
- **[SOURCE-OF-TRUTH]** — `openspec/specs/harness-workflow/spec.md` (merged from the
  delta at archive). Authoritative wording.
- **[FULL-RULE SKILL]** — `chained-pr/SKILL.md` (carries the full narrative).
- **[POINTER-ONLY]** — `CLAUDE.md`, `harness-workflow/SKILL.md` (one line each,
  reference the rule; NO narrative).
- **[MECHANICS]** — `sdd-archive`, `branch-pr`, `sdd-apply`,
  `session-status-contract`, `chaining-details.md` (operational how-to that must
  name the tracker branch as the owning branch for the chain flow).

### 1.1 [SOURCE-OF-TRUTH] `openspec/specs/harness-workflow/spec.md`

The delta at `specs/harness-workflow/spec.md` is a pure ADDED requirement; the spec
merge at archive appends it. Apply does the following in the LIVE spec file:

**Edit A — REPLACE the blanket-deferral scenario (`:183-189`).**
The current scenario "Chained-PR flow is out of scope (deferred to slice 2)" is the
exact passage the delta supersedes. Apply REMOVES those 7 lines
(`:183-189`, from `#### Scenario: Chained-PR flow is out of scope` through
`...applied to chained-PR flows`) and in their place APPENDS the entire ADDED
requirement block from the delta:

- `### Requirement: Terminal Phase Ordering (Feature Branch Chain)` with its three
  clauses (main rule, Additive clause, Judge-timing clause), the
  `SESSION_STATUS.md` root-residency paragraph, the 4-step archive-internal order,
  and the single-commit staging paragraph (delta `:19-54`).
- All 6 delta scenarios verbatim (delta `:56-123`):
  1. Integrated judge passes on tracker, archive staged, tracker merges to main
  2. Archive is blocked on tracker before integrated judge passes
  3. Integrated judge runs on tracker branch (all changes integrated)
  4. SESSION_STATUS.md stays root-resident through integrated judge on tracker
  5. SESSION_STATUS.md moves in the tracker archive commit
  6. Archive-internal ordering is preserved on the tracker branch
  7. Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)
  (Note: that is 7 scenarios in the delta; the "8 scenarios" framing in the phase
  brief counts the requirement's clauses + scenarios — see §5 verification map for
  the explicit 1:1 mapping. The apply target is: every delta scenario present in
  the merged spec, byte-for-byte from the delta.)

**Placement:** the ADDED requirement block goes AFTER the single-PR requirement's
last scenario. Because the single-PR requirement ends where the old blanket scenario
began (`:183`), removing `:183-189` and appending the new block keeps the single-PR
requirement (`:112-181`) UNTOUCHED and the new requirement immediately after it.

**Do NOT touch** `:112-181` (the entire "Terminal Phase Ordering (Single-PR Flow)"
requirement and its 6 scenarios). Verify byte-identity after the edit (§5).

Target header the merged spec must contain (from delta `:19`):

> ### Requirement: Terminal Phase Ordering (Feature Branch Chain)
> In the **Feature Branch Chain** strategy, the archive commit MUST be staged on the
> **tracker/integration branch**, AFTER the integrated judge passes on that branch
> and BEFORE the tracker PR merges to `main`. …

### 1.2 [FULL-RULE SKILL] `skills/chained-pr/SKILL.md`

**Edit B — REPLACE the blanket deferral Hard Rule (`:24`).**
Current line `:24`:

> - The archive-before-PR single-PR rule does NOT apply to chained/stacked flows.
>   Which PR owns the archive move in a stacked sequence is a slice-2 non-goal
>   (deferred); do not enforce archive position on chained PRs.

Replace with a two-bullet Hard Rule pair (Feature Branch Chain defined; Stacked
deferred):

> - **Feature Branch Chain archive ownership:** the archive commit (spec merge,
>   folder move, `archon map`, `SESSION_STATUS.md` move) is staged on the
>   **tracker branch**, AFTER the integrated judge passes on the tracker and
>   BEFORE the tracker PR merges to `main`. It travels inside the tracker PR — never
>   a child PR diff, never a separate archive-only PR. Archive-internal order and
>   one-commit staging are unchanged from the single-PR flow (`[[harness-workflow]]`
>   "Terminal Phase Ordering (Feature Branch Chain)").
> - **Stacked-to-Main archive ownership** is NOT yet defined — deferred to slice 2b.
>   Stacked-to-Main has no tracker branch; do not apply the Feature Branch Chain
>   archive rule to it and do not enforce archive position on Stacked-to-Main flows.

**Edit C — add the integrated-judge trigger + archive step to Execution Steps
(`:36-44`).** The current Step 6 (`:43`) keeps the tracker draft/no-merge, but says
nothing about the integrated judge or archive. Expand the tail of Execution Steps so
the orchestrator is told, in order, to (a) integrate children, (b) run the integrated
judge on the tracker, (c) stage archive on the tracker, (d) then merge the tracker.
Rewrite/append Step 6 and add Steps 7–8:

> 6. Keep the tracker PR draft/no-merge until all child PRs are reviewed and merged
>    into the tracker branch.
> 7. After the LAST child PR merges into the tracker branch, run the **integrated
>    judge on the tracker branch** (evaluate all change files as a unified whole via
>    `[[harness-judge]]`). The integrated judge is the gate for BOTH the archive step
>    and the tracker merge. If it fails, re-apply on the tracker and re-judge before
>    proceeding — do NOT archive and do NOT merge the tracker.
> 8. Once the integrated judge passes, stage the archive commit on the tracker branch
>    (spec merge → folder move → `archon map --backfill`/`--check` STOP gate →
>    `SESSION_STATUS.md` move, one commit), THEN merge the tracker PR to `main`. See
>    `[[sdd-archive]]` for the archive mechanics.

**Edit D — reference the new tracker-archive detail in References or the tracker
Hard Rule (`:21`).** The existing Hard Rule `:21` about the draft/no-merge tracker
PR is correct and stays. Add nothing there except (optional, low-risk) a trailing
clause so the tracker-PR rule connects to archive:

> - In Feature Branch Chain, create a draft/no-merge tracker PR; child PR #1 targets
>   the tracker branch, later children target the immediate parent branch. The
>   tracker PR is also the archive commit's home (see Execution Steps 7–8).

This edit is OPTIONAL — Edit B + Edit C already carry the rule. Apply MAY skip it if
it risks redundancy; the design flags it as nice-to-have, not required.

### 1.3 [MECHANICS] `skills/chained-pr/references/chaining-details.md`

**Edit E — add archive placement to the Feature Branch Chain steps (`:27-34`) and a
one-line note to the diagram (`:16-25`).** Extend the numbered steps so archive and
the integrated judge appear in the sequence. Rewrite step 5 (`:33`) and add steps 6–7:

> 5. Merge/integrate children in order into the tracker branch.
> 6. After the last child merges, run the integrated judge on the tracker branch; it
>    must pass before archive or tracker merge (gate for both).
> 7. Stage the archive commit on the tracker branch (spec merge, folder move,
>    `archon map`, `SESSION_STATUS.md` move — one commit), THEN merge the tracker PR
>    to `main`. Archive lives inside the tracker PR, not a child PR.

Add a one-line annotation under the ASCII diagram (after `:25`) so the tracker branch
is visibly the archive owner:

> The tracker branch (`feat/my-feature`) owns the archive commit: it is staged there
> after the integrated judge passes, before the tracker merges to `main`.

**Edit F — Stacked deferral note in the Stacked section (`:36-45`).** Add one line
after the Stacked ASCII/paragraph (`:45`) so the deferral is explicit here too:

> Archive ownership for Stacked PRs to main is NOT yet defined (deferred to slice
> 2b); there is no tracker branch to host the archive commit.

### 1.4 [POINTER-ONLY] `CLAUDE.md` (Phase Order, `:27-30`)

**Edit G — add ONE pointer line after `:30`.** Line `:30` already states the
single-PR archive-before-PR behavior. Add a sibling sentence for the chain flow,
pointing to the rule (NO narrative):

> In the Feature Branch Chain flow the archive commit is staged on the tracker branch
> after the integrated judge passes and before the tracker merges to `main` (full
> rule: `harness-workflow` spec "Terminal Phase Ordering (Feature Branch Chain)" and
> the `chained-pr` skill). Stacked-to-Main archive ownership is deferred to slice 2b.

Single-PR line `:30` stays VERBATIM.

### 1.5 [POINTER-ONLY] `skills/harness-workflow/SKILL.md`

**Edit H — add ONE pointer line after the single-PR terminal-ordering note
(`:23-25`).** Line `:23-25` is the single-PR terminal note; it stays VERBATIM. Add a
sibling paragraph:

> **Terminal ordering (Feature Branch Chain)**: in the Feature Branch Chain flow the
> archive commit is staged on the **tracker branch** after the integrated judge
> passes on the tracker and before the tracker PR merges to `main`. Full rule:
> `harness-workflow` spec "Terminal Phase Ordering (Feature Branch Chain)" and the
> `chained-pr` skill. Stacked-to-Main archive ownership is deferred to slice 2b.

**Edit I — extend the Rules bullet at `:158`.** The bullet currently says
`SESSION_STATUS.md` "is MOVED into the archived change folder during `sdd-archive`,
within the staged archive commit, before the PR is opened." Make it flow-aware
without duplicating the rule:

> - `SESSION_STATUS.md` is updated on every transition and is MOVED into the archived
>   change folder during `sdd-archive`, within the staged archive commit — before the
>   PR is opened in the single-PR flow, or on the **tracker branch** before the
>   tracker PR merges to `main` in the Feature Branch Chain flow. See
>   `session-status-contract`.

This is a pointer-level clarification (names the branch, does not restate the full
requirement). The state-machine transition table (`:145-148`) is NOT modified — see
§3 for why no new `blocked` transition is added.

### 1.6 [MECHANICS] `skills/sdd-archive/SKILL.md`

The single-PR path MUST keep working; the chain path is expressed as an "IF Feature
Branch Chain" variant on the branch that owns the commit. Wording template:
"on the change branch (single-PR flow) or the tracker branch (Feature Branch Chain
flow)."

**Edit J — Timing (`:89-97`).** Current text says archive runs "AFTER judge passes
and BEFORE the PR is opened … on the change branch." Add flow-awareness:

> Archive runs AFTER judge passes and BEFORE the PR is opened. **In the Feature
> Branch Chain flow, "judge" means the integrated judge on the tracker branch, the
> owning branch is the tracker branch (not the change branch), and "before the PR is
> opened" becomes "before the tracker PR merges to `main`."** Stage the results of
> Step 2 (spec merge), Step 3 (folder move), Step 3b (`archon map`), and Step 3c
> (`SESSION_STATUS.md` move) into ONE archive commit on the owning branch (change
> branch for single-PR, tracker branch for Feature Branch Chain). …

Rest of the Timing paragraph (archive-internal order unchanged; PR-open EXTERNAL)
stays as-is; it already generalizes.

**Edit K — Step 3c (`:170-188`).** Add one sentence after the MOVE block (`:181`) so
the chain owning branch is named:

> In the Feature Branch Chain flow this move happens on the tracker branch, staged
> into the tracker archive commit (before the tracker merges to `main`), not on an
> individual child branch.

**Edit L — Step 3d (`:190-208`).** The commit-staging step currently hard-codes "onto
the change branch." Generalize the owning branch and the precondition:

- `:192` heading text: "Stage and commit all archive changes onto the **owning
  branch** (change branch for single-PR; tracker branch for Feature Branch Chain) as
  ONE commit."
- `:204-205` precondition bullet 4: keep the single-PR sentence and ADD a
  chain-aware sibling:

  > 4. This commit satisfies the branch-PR precondition: the archive commit MUST be
  >    staged on the owning branch BEFORE the change is opened/merged — on the change
  >    branch before the PR is opened (single-PR flow), or on the tracker branch
  >    before the tracker PR merges to `main` (Feature Branch Chain flow).

**Edit M — Step 4 checklist (`:210-221`).** The last checklist item (`:220-221`) is
the single-PR assertion "PR has NOT been opened before this archive commit is staged
(single-PR flow)" and "Archive commit created on the change branch." Keep both, and
ADD a Feature-Branch-Chain-aware variant so both flows are asserted:

- Keep `:220` as-is (change branch, single-PR).
- Keep `:221` as-is (PR not opened before archive — single-PR).
- ADD after `:221`:
  > - [ ] Feature Branch Chain flow ONLY: the archive commit is created on the
  >   **tracker branch**, and the **tracker PR has NOT been merged to `main` before
  >   the archive commit is staged on the tracker branch**; the integrated judge
  >   passed on the tracker before archive ran.

This is the "tracker PR not merged before the archive commit is staged on the
tracker" assertion the phase brief requires. The single-PR "PR NOT opened" assertion
is preserved verbatim; the two coexist as flow-scoped checklist rows.

### 1.7 [MECHANICS] `skills/branch-pr/SKILL.md` (Workflow `:34-41`)

**Edit N — add the tracker-merge precondition (`:35-37`).** Current step 5 covers the
single-PR archive precondition. Add a step 5b (or expand) for the chain:

> 5. If this is an SDD single-PR change, the archive commit (spec merge, folder move,
>    SESSION_STATUS.md move, `archon map`) MUST already be staged on the branch
>    before opening the PR.
> 5b. If this is an SDD Feature Branch Chain change, the integrated judge MUST have
>    passed on the tracker branch and the archive commit MUST already be staged on the
>    **tracker branch** before the tracker PR is merged to `main`. Child PRs carry no
>    archive commit.

Renumber the trailing steps only if apply prefers integers; a `5b` sub-step is
acceptable and lower-risk (avoids touching steps 6–8 numbering).

### 1.8 [MECHANICS] `skills/sdd-apply/SKILL.md`

Apply hands off to archive-on-tracker; name where archive lands per strategy.

**Edit O — chain-strategy branch-targeting note (`:96-98`).** After the
`feature-branch-chain` bullet (`:98`), add one sentence naming the archive owner:

> For `feature-branch-chain`, the archive commit is NOT part of any child slice; it is
> staged on the tracker branch after the integrated judge passes and before the
> tracker merges to `main` (see `[[sdd-archive]]` and `[[chained-pr]]`).

**Edit P — PR Boundary / apply-scope tail (`:263-276`).** The paragraph at `:273-276`
describes the single-PR handoff ("apply's work ends with … archive-then-PR"). Add a
sibling sentence for the chain:

> For a Feature Branch Chain change, each apply batch delivers one child work unit;
> the archive commit belongs to the tracker integration (staged after the integrated
> judge, before the tracker merges), never to a child apply batch.

**Edit Q — Rules (`:290-292`).** The rule at `:292` is single-PR archive-after-stage.
Add a chain-aware sibling rule after it:

> - In the Feature Branch Chain flow, the archive commit is staged on the tracker
>   branch after the integrated judge passes and before the tracker PR merges to
>   `main`; it is never part of a child apply batch and never appears in a child PR
>   diff.

### 1.9 [MECHANICS] `skills/_shared/session-status-contract.md` (`:22-29`)

**Edit R — chain clarification in the Archive bullet (`:22-24`).** Keep the existing
single-PR wording; append the chain case:

> - **Archive**: during `sdd-archive`, MOVE the file into the archived change folder
>   (`openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`) as part of
>   the staged archive commit — before the PR is opened in the single-PR flow, or on
>   the **tracker branch** before the tracker PR merges to `main` in the Feature
>   Branch Chain flow — then delete it from the root. It stays root-resident through
>   the integrated judge in the chain flow. In Engram-only mode, store its final
>   contents as an observation and remove the root file.

The "one file per session" invariant (`:27-29`) is unchanged; it already covers the
chain (one file, moved once, at the tracker archive commit).

### 1.10 [MECHANICS, OPTIONAL] `skills/work-unit-commits/SKILL.md` (`:54-62`)

**Edit S — note archive is not a work-unit slice (OPTIONAL).** After the PR
Relationship list (`:61`), add one line:

> The archive commit is NOT a work unit and NOT a chained-PR slice; in the Feature
> Branch Chain flow it belongs to the tracker integration commit, staged after the
> integrated judge passes and before the tracker merges to `main`.

OPTIONAL — apply MAY skip if it adds no clarity beyond `chained-pr`/`sdd-apply`. Flag
as low-priority; not required for the 8 scenarios to hold.

---

## 2. Consistency strategy + edit order

**Source-of-truth vs downstream.** The spec (Edit A) is authoritative. `chained-pr`
(Edits B–D) carries the full operational narrative for the orchestrator. `CLAUDE.md`
(G) and `harness-workflow/SKILL.md` (H, I) carry POINTER lines that name the rule and
the two full-rule homes — they must NOT restate the requirement body. All mechanics
files (J–S) express the SAME single fact ("owning branch = tracker for Feature Branch
Chain; integrated judge gates archive-then-tracker-merge") in operational voice, each
scoped to its concern.

**How apply keeps them aligned without duplication:**
- The full rule text is authored ONCE against the delta (Edit A). Edits B/C reuse the
  SAME clauses (main rule + judge-timing + owning branch + Stacked deferral) in skill
  voice; they must not diverge in the three testable facts: (i) owning branch =
  tracker, (ii) integrated judge on tracker gates archive AND tracker merge, (iii)
  Stacked-to-Main deferred to slice 2b.
- Pointer lines (G, H) must contain a literal reference string that resolves:
  `harness-workflow` spec "Terminal Phase Ordering (Feature Branch Chain)" and the
  `chained-pr` skill. §5 verifies these strings resolve.

**Recommended edit order (lowest-risk first, source-of-truth before downstream):**
1. Edit A — spec (source of truth) — establishes canonical wording.
2. Edits B, C, D — `chained-pr/SKILL.md` (full-rule skill) — mirror the spec clauses.
3. Edits E, F — `chaining-details.md` (mechanics for chained-pr).
4. Edits J, K, L, M — `sdd-archive/SKILL.md` (owning-branch mechanics).
5. Edit N — `branch-pr/SKILL.md`.
6. Edits O, P, Q — `sdd-apply/SKILL.md`.
7. Edit R — `session-status-contract.md`.
8. Edits G, H, I — POINTER lines (`CLAUDE.md`, `harness-workflow/SKILL.md`) LAST, so
   they can quote the now-final rule header string exactly.
9. Edit S — `work-unit-commits` (optional), last.

Doing pointer lines last guarantees the referenced header string matches the merged
spec header byte-for-byte.

---

## 3. Integrated-judge trigger — where/how the skills fire it

**Requirement mapped:** delta scenario "Integrated judge runs on tracker branch (all
changes integrated)" (delta `:78-85`) + judge-timing clause (delta `:31-35`) + the
"Archive is blocked before integrated judge passes" scenario (delta `:68-76`).

**Where the skills instruct it (prose, no CLI):**
- `chained-pr/SKILL.md` Execution Steps 7–8 (Edit C) are the PRIMARY trigger: after
  the LAST child merges into the tracker, the orchestrator runs the integrated judge
  on the tracker via `[[harness-judge]]`, and only on pass proceeds to archive then
  tracker merge. This is the operational home of the trigger.
- `chaining-details.md` steps 6–7 (Edit E) mirror it in the concrete sequence.
- `sdd-archive` Timing (Edit J) redefines "judge" as "the integrated judge on the
  tracker branch" for the chain flow, so the existing "archive runs AFTER judge
  passes" gate transfers to the integrated judge without new machinery.
- `branch-pr` step 5b (Edit N) and `sdd-apply` Rules (Edit Q) restate the gate at the
  PR/merge boundary.

**No CLI, no new state-machine transition.** The integrated judge is skill/LLM prose
driven by `[[harness-judge]]`; there is no `archon` subcommand for it and none is
added. The `harness-workflow/SKILL.md` state table (`:145-148`) is NOT given a new
`blocked` row: the state machine already has a single `judge` phase gate before
`archive`, and the "integrated judge" is that same `judge` phase re-run on the tracker
after child integration — it is a branch-context nuance, not a new phase. Edits H
(pointer paragraph) + I (SESSION_STATUS bullet) carry the chain-aware note at prose
level; the transition enforcement table stays flow-agnostic. **This is the deliberate
design choice: no code, no new transition; the gate is the existing `judge → archive`
edge, re-anchored to the tracker branch by prose.** (See §6 risk "judge-timing
well-founded.")

---

## 4. sdd-archive branch-ownership (both flows coexist)

The design expresses owning-branch as a scoped variant, never a replacement:

- **Timing (Edit J):** one added sentence: single-PR = change branch + "before PR
  opened"; Feature Branch Chain = tracker branch + integrated judge + "before tracker
  merges to `main`." The archive-internal order sentence is untouched (flow-agnostic).
- **Step 3c (Edit K):** single-PR MOVE block untouched; one appended sentence names
  the tracker branch for the chain.
- **Step 3d (Edit L):** heading generalized to "owning branch (change branch for
  single-PR; tracker branch for Feature Branch Chain)"; precondition bullet 4 keeps
  the single-PR clause and adds the chain clause.
- **Step 4 checklist (Edit M):** the single-PR rows (`:220`, `:221`) are PRESERVED
  verbatim; a new flow-scoped row asserts the tracker-branch + integrated-judge +
  "tracker PR not merged before archive staged" facts. Each row is explicitly scoped
  ("single-PR flow" / "Feature Branch Chain flow ONLY") so neither flow's checklist
  is ambiguous.

Because every chain sentence is added as an "IF Feature Branch Chain" scoped clause
alongside the preserved single-PR text, the single-PR archive path reads identically
after the edits (§5 verifies no single-PR sentence is deleted or reworded).

---

## 5. Verification hooks (docs-consistency; no automated Go tests)

Prose change — NO Go tests exist or are added for terminal ordering (explore-
confirmed). Apply and verify use these manual doc-consistency checks:

**V1 — single-PR untouched (regression guard for #96).**
- `git diff` on `openspec/specs/harness-workflow/spec.md` shows ZERO changes within
  `:112-181` (single-PR requirement + its 6 scenarios). Only lines `:183-189` are
  removed and new content appended after.
- `CLAUDE.md:30`, `harness-workflow/SKILL.md:23-25`, and every single-PR sentence in
  `sdd-archive`/`branch-pr`/`sdd-apply`/`session-status-contract` are byte-identical
  (only additions, never edits, to single-PR clauses). Grep the single-PR checklist
  rows `sdd-archive:220-221` remain present and unchanged.

**V2 — spec ↔ chained-pr full rule alignment.** The three testable facts appear in
BOTH the merged spec and `chained-pr/SKILL.md`:
- owning branch = tracker;
- integrated judge on tracker gates archive AND tracker merge;
- Stacked-to-Main deferred to slice 2b.

**V3 — pointer lines resolve.** In `CLAUDE.md` (Edit G) and
`harness-workflow/SKILL.md` (Edit H), grep for the literal string
`Terminal Phase Ordering (Feature Branch Chain)` and confirm it matches the merged
spec requirement header exactly, and that `chained-pr` is named as the skill home.
Confirm neither pointer restates the requirement body (length check: 1–2 sentences).

**V4 — each delta scenario reflected in prose.** Map every delta scenario to at least
one skill sentence:

| Delta scenario (spec) | Reflected in |
|---|---|
| Integrated judge passes, archive staged, tracker merges | `chained-pr` Step 8 (C); `chaining-details` step 7 (E); `sdd-archive` Timing (J) |
| Archive blocked before integrated judge passes | `chained-pr` Step 7 (C); `branch-pr` 5b (N); `sdd-archive` Step 4 row (M) |
| Integrated judge runs on tracker (all integrated) | `chained-pr` Step 7 (C); `chaining-details` step 6 (E) |
| SESSION_STATUS root-resident through integrated judge | `session-status-contract` (R); `harness-workflow` Rules bullet (I) |
| SESSION_STATUS moves in tracker archive commit | `sdd-archive` Step 3c (K); `session-status-contract` (R) |
| Archive-internal order preserved on tracker | `sdd-archive` Timing (J) — order sentence untouched; `chained-pr` Step 8 (C) |
| Stacked-to-Main out of scope (slice 2b) | `chained-pr` Hard Rule (B); `chaining-details` (F); pointers (G, H) |

Requirement clauses (main rule, additive clause, judge-timing clause) each map to
Edits A + B/C. This table is the "8 scenarios reflected" evidence for the Human
Review Gate (7 scenarios + the requirement's clause set).

**V5 — Stacked-to-Main deferral present and explicit** (not silent) in: spec scenario
7 (A), `chained-pr` Hard Rule (B), `chaining-details` note (F), and both pointer lines
(G, H).

**V6 — link integrity.** After edits, wikilinks `[[harness-workflow]]`,
`[[sdd-archive]]`, `[[chained-pr]]`, `[[harness-judge]]`,
`[[session-status-contract]]` resolve (run `archon map --check` at archive; the
delta's own relative links are handled by the archive backfill).

---

## 6. Risk-targeted mitigations (per proposal risk)

| Proposal risk | Design choice that mitigates it |
|---|---|
| **Regressing #96 single-PR behavior** (Med) | Every edit is ADDITIVE or an "IF Feature Branch Chain" scoped clause; single-PR requirement (`spec:112-181`), `CLAUDE.md:30`, `harness-workflow:23-25`, and single-PR checklist rows (`sdd-archive:220-221`) are PRESERVED verbatim. V1 is a hard regression guard: `git diff` must show no single-PR line changed. |
| **Reintroducing "polluted last PR"** (Med) | Archive is placed on the tracker branch (integration history), explicitly "never a child PR diff, never a separate archive-only PR" — stated in Edit B, C, O, P, Q. `sdd-apply` PR Boundary (P) says the archive commit is never part of a child apply batch. |
| **Integrated-judge timing unspecified** (Med) | Judge-timing clause is a first-class requirement clause (spec delta `:31-35`) + scenario (delta `:78-85`, `:68-76`), operationalized in `chained-pr` Steps 7–8 (C), `chaining-details` step 6 (E), `branch-pr` 5b (N), `sdd-archive` Timing/Step 4 (J, M). §3 documents WHY no new state-machine transition or CLI is needed — the gate is the existing `judge → archive` edge re-anchored to the tracker by prose. |
| **Phase-order copies drift out of sync** (Low) | Full rule in exactly two homes (spec + `chained-pr`); `CLAUDE.md`/`harness-workflow` get pointer lines that quote the spec header string. Edit order (§2) does pointers LAST so the quoted header is final. V2 + V3 verify alignment and pointer resolution. |

---

## 7. Open questions (for the Human Review Gate)

1. **Scenario count framing.** The delta defines 1 requirement (3 clauses) + 7
   scenarios. The phase brief says "8 scenarios." §5 V4 treats "8" as
   7 scenarios + the requirement clause-set; the design maps all of them. Confirm this
   interpretation, or point to a missing 8th scenario to add to the delta before
   apply.
2. **Optional edits D and S.** Edit D (`chained-pr:21` trailing clause) and Edit S
   (`work-unit-commits`) are flagged OPTIONAL to avoid redundancy. Approve
   include-both, or drop them and rely on B/C/O/P/Q. Recommendation: include D
   (cheap, connects tracker-PR rule to archive), skip S unless reviewers want the
   explicit "archive is not a work unit" statement in that skill too.
3. **Code implication (should be none).** Explore confirmed no Go enforces terminal
   ordering. If apply discovers ANY Go path (e.g., a `harness-workflow` validator that
   inspects PR/branch state), STOP and surface it — this design covers docs/skills
   only and would need a follow-up.
