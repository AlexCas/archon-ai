# Design: Stacked-to-Main Archive Convergence (converge to Feature Branch Chain)

<!-- Link convention: [[capability]] wikilinks for capability identity; relative
     links for intra-change navigation. Full rule: skills/_shared/spec-vault.md. -->

Change: `stacked-pr-archive` (slice 2b of #93)
Phase: design. Contract: [specs/harness-workflow/spec.md](specs/harness-workflow/spec.md)
and [proposal.md](proposal.md).
Type: **docs-and-skills only. NO Go.** Explore + slice 2 confirmed no Go enforces
terminal ordering or chain-strategy selection — `harness-workflow`, `sdd-tasks`,
`sdd-apply`, `chained-pr`, `sdd-archive`, and `branch-pr` are skill/LLM prose. The
"blocked"/"notify" behavior in this slice is orchestrator prose, not a Go error path.
Any Go path apply discovers is an OPEN QUESTION (§7), never a design item here.

---

## 0. Framing — what slice 2b actually changes

Slice 2 (#97) already shipped the full Feature Branch Chain (FBC) archive mechanics
into every downstream skill: `sdd-apply` (`:96-101`), `sdd-archive`
(Timing `:91-94`, Step 3c `:187-188`, Step 3d `:200`/`:212-215`, checklist `:232-235`),
`branch-pr` (step 5b `:38-44`), and `session-status-contract` (`:22-28`). **Those
files are already FBC-aware and need NO further archive-mechanics edits.** Confirmed
by grep during design (see anchors in §1).

What slice 2b adds is a NARROW convergence layer on top of the shipped FBC rule:

1. **The up-front decision** — the crux. When archive-before-PR is in effect
   (`openspec`/`hybrid`), the orchestrator MUST NOT select pure Stacked-to-Main at
   `sdd-tasks` strategy selection; it selects FBC and notifies the user. This is the
   single new behavioral rule, and its operational home is `sdd-tasks/SKILL.md`
   (chain-strategy selection, steps 4–5).
2. **Deferral replacement** — five locations currently say "deferred to slice 2b" /
   "NOT yet defined." Each is replaced by the converge-to-FBC decision (full rule in
   spec + `chained-pr`; pointer lines in `CLAUDE.md` + `harness-workflow/SKILL.md`;
   a mechanics note in `chaining-details.md`). No silent gap remains.
3. **Zero new archive mechanics** — after convergence the change IS an FBC change, and
   the shipped slice-2 rule governs the archive commit verbatim. Slice 2b introduces
   no new tracker/archive machinery.

Design principle throughout: **additive, never rewriting.** Single-PR requirement and
the FBC requirement body stay verbatim; the FBC requirement loses ONLY its
Stacked-to-Main deferral scenario. Every other edit is an added sentence/bullet.

---

## 1. Per-file edit plan

Legend for role tags:
- **[SOURCE-OF-TRUTH]** — `openspec/specs/harness-workflow/spec.md` (merged from the
  delta at archive). Authoritative wording.
- **[FULL-RULE SKILL]** — `skills/chained-pr/SKILL.md` (carries the full converge
  narrative for the orchestrator).
- **[STRATEGY-SELECTION HOME]** — `skills/sdd-tasks/SKILL.md` (where the up-front
  converge decision is enforced — the crux).
- **[POINTER-ONLY]** — `CLAUDE.md`, `skills/harness-workflow/SKILL.md` (one line each;
  reference the rule, NO narrative).
- **[MECHANICS]** — `skills/sdd-apply/SKILL.md`, `skills/sdd-archive/SKILL.md`,
  `skills/branch-pr/SKILL.md`, `skills/chained-pr/references/chaining-details.md`,
  `skills/_shared/session-status-contract.md` (operational how-to; most already
  FBC-aware from slice 2 and touched only where a deferral note lives or a
  convergence precondition belongs).

**Apply MUST grep, not trust line numbers.** #97/#98 already shifted numbers; anchors
below are the CURRENT state observed during design (2026-08-01). Every edit gives a
grep target string so apply re-locates it.

### 1.1 [SOURCE-OF-TRUTH] `openspec/specs/harness-workflow/spec.md`

The delta (`specs/harness-workflow/spec.md` in this change folder) does two things:
a MODIFIED FBC requirement (deferral scenario removed, body verbatim) and an ADDED
"Stacked-to-Main Archive Convergence" requirement. Apply mirrors both into the LIVE
spec (the spec merge at archive is what physically applies them; the design tells
apply exactly what the merged result must contain).

**Edit A1 — REMOVE the Stacked-to-Main deferral scenario.**
Grep target (currently `:281-287`):

> #### Scenario: Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)
> - GIVEN a change using the Stacked-to-Main strategy (individual child PRs each target `main`)
> - WHEN the terminal phase sequence is evaluated
> - THEN the Feature Branch Chain archive rule from this requirement DOES NOT apply
> - AND the harness-workflow treats Stacked-to-Main archive ownership as undefined / slice-2b pending
> - AND no blocking or enforcement of archive position is applied to Stacked-to-Main flows

Delete these 7 lines in full. This is the ONLY change to the FBC requirement — its
header, three clauses, SESSION_STATUS paragraph, 4-step archive-internal order, and
its six retained scenarios stay VERBATIM. The delta's MODIFIED block carries the FBC
requirement body byte-for-byte minus this scenario; apply must confirm byte-identity
of the retained body (V1).

**Edit A2 — APPEND the ADDED "Stacked-to-Main Archive Convergence" requirement.**
Immediately after the FBC requirement's last retained scenario ("Archive-internal
ordering is preserved on the tracker branch", ending `...all four sub-operations and
the archive commit are done`, currently `:279`), append the entire ADDED requirement
from the delta (`specs/harness-workflow/spec.md:151-234` in this change folder):

- `### Requirement: Stacked-to-Main Archive Convergence` header
- the five prose paragraphs: main rule, Rationale, Up-front convergence is mandatory,
  Unsupported combination, After convergence, Scope boundary
- all FIVE scenarios verbatim:
  1. Archive-before-PR active, Stacked-to-Main requested — orchestrator converges to FBC at sdd-tasks
  2. Pure Stacked-to-Main + archive-before-PR is unsupported — no partial main merges before owning ref
  3. After convergence, FBC archive rule governs — no new mechanics
  4. Late Stacked-to-FBC conversion is out of scope — stranded slices hazard
  5. Archive-before-PR not in effect — Stacked-to-Main is unaffected

Header the merged spec MUST contain (quoted by the pointer edits G, H):

> ### Requirement: Stacked-to-Main Archive Convergence

**Do NOT touch** the single-PR requirement (grep `### Requirement: Terminal Phase
Ordering (Single-PR Flow)`, currently ~`:112-181`) or the FBC requirement body. V1 is
the regression guard.

### 1.2 [STRATEGY-SELECTION HOME] `skills/sdd-tasks/SKILL.md` — THE CRUX

The up-front decision lands in the "Ask the user which chain strategy to use" block.
Grep target: step 4 "**Ask the user which chain strategy to use**" and step 5 "Cache
the user's choice" (currently `:157-165`).

**Edit T1 — insert the converge precondition as a new step 4a (or a bolded
paragraph) between step 4 and step 5, at `:159`/`:160`.** It must fire BEFORE the
user's choice is cached, so pure Stacked-to-Main is never selectable under
archive-before-PR. Exact prose to insert:

> 4a. **Archive-before-PR convergence gate (MANDATORY).** If archive-before-PR is in
>    effect for this session — i.e. the artifact store is `openspec` or `hybrid`
>    (from the SDD session preflight) — the orchestrator MUST NOT select pure
>    **Stacked PRs to main**. Stacked-to-Main ships each slice independently to
>    `main`, so no single un-merged ref can own the archive commit; the
>    archive-before-PR invariant cannot be satisfied. In that case the orchestrator
>    MUST select **Feature Branch Chain** (which supplies the tracker branch that
>    owns the archive commit) and MUST notify the user that
>    `Stacked-to-Main + archive-before-PR is unsupported, so Feature Branch Chain was
>    selected`. This decision is up-front and decisive: made here, before any child
>    PR is opened. A late Stacked→FBC conversion (after slices already merged to
>    `main`) is NOT sanctioned and has no recovery procedure — see the
>    `harness-workflow` spec requirement "Stacked-to-Main Archive Convergence." When
>    the artifact store is `engram` (archive-before-PR not in effect), Stacked-to-Main
>    is unaffected and remains a valid choice.

**Edit T2 — annotate the `Chain strategy` guard-contract line** so the literal
`stacked-to-main` value carries the constraint inline. Grep target: the fenced block
containing `Chain strategy: stacked-to-main|feature-branch-chain|size-exception|pending`
(currently `:174`, and the mirror at `:100`/`:96`). Do NOT change the literal values
(downstream guards in `sdd-apply` match them). Add ONE sentence directly under the
fenced forecast block (after `:176`, before "You may keep the table..."):

> When archive-before-PR is in effect (`openspec`/`hybrid`), `Chain strategy` MUST NOT
> be `stacked-to-main`; the convergence gate (step 4a) sets it to
> `feature-branch-chain` instead. `stacked-to-main` remains valid only when the
> artifact store is `engram`.

Rationale for landing here: `sdd-tasks` is where the user's chain choice is solicited
and cached (`Cache the user's choice`, `:161`). Enforcing at selection time — not at
apply or archive — is the approved "up-front and decisive" decision and directly
implements delta scenario 1. Encoding it before the cache write guarantees the cached
`Chain strategy` value already reflects the convergence, so `sdd-apply` and
`chained-pr` inherit a corrected value with no additional gate needed.

### 1.3 [FULL-RULE SKILL] `skills/chained-pr/SKILL.md`

**Edit B — REPLACE the deferral Hard Rule bullet with the converge rule.** Grep
target (currently `:31-33`):

> - **Stacked-to-Main archive ownership** is NOT yet defined — deferred to slice 2b.
>   Stacked-to-Main has no tracker branch; do not apply the Feature Branch Chain
>   archive rule to it and do not enforce archive position on Stacked-to-Main flows.

Replace with (full rule, orchestrator voice):

> - **Stacked-to-Main + archive-before-PR converges to Feature Branch Chain.** When
>   archive-before-PR is in effect (artifact store `openspec`/`hybrid`), pure
>   Stacked-to-Main is unsupported: it ships slices independently to `main`, leaving
>   no un-merged ref to own the archive commit. At `sdd-tasks` strategy selection the
>   orchestrator MUST select **Feature Branch Chain** instead and notify the user.
>   After convergence the change is FBC and the Feature Branch Chain archive rule
>   above governs — zero new archive mechanics. A late Stacked→FBC conversion (after
>   slices already merged to `main`) strands those slices and is NOT sanctioned; there
>   is no recovery procedure. When archive-before-PR is NOT in effect (`engram`),
>   Stacked-to-Main is unaffected and does not converge. Full rule:
>   `[[harness-workflow]]` "Stacked-to-Main Archive Convergence".

**Edit C — update the Decision Gates table row for Stacked.** Grep target (currently
`:40`):

> | PR >400, each slice can land independently | Use Stacked PRs to main. |

Replace with:

> | PR >400, each slice can land independently, archive-before-PR NOT in effect (`engram`) | Use Stacked PRs to main. |
> | PR >400, each slice independent, BUT archive-before-PR in effect (`openspec`/`hybrid`) | Converge to Feature Branch Chain (Stacked-to-Main unsupported here). |

This makes the gate table itself enforce the converge decision at the point the
orchestrator consults it. Low-risk: the existing FBC row (`:41`) is untouched.

**Edit D (OPTIONAL) — Execution Steps note.** Step 2 (`:48`) says "Ask for a chain
strategy when none is cached and the budget is exceeded." Apply MAY append a trailing
clause: "— and if archive-before-PR is in effect, do not offer pure Stacked-to-Main;
converge to Feature Branch Chain per the Hard Rule." OPTIONAL — Edits B + C already
carry the rule; include only if it adds clarity without redundancy.

### 1.4 [MECHANICS] `skills/chained-pr/references/chaining-details.md`

**Edit E — REPLACE the Stacked deferral note.** Grep target (currently `:56-57`):

> Archive ownership for Stacked PRs to main is NOT yet defined (deferred to slice
> 2b); there is no tracker branch to host the archive commit.

Replace with:

> Archive ownership for Stacked PRs to main: when archive-before-PR is in effect
> (`openspec`/`hybrid`), pure Stacked-to-Main is unsupported — there is no tracker
> branch to host the archive commit, so the orchestrator converges to Feature Branch
> Chain at `sdd-tasks` strategy selection and the FBC archive step (above) applies.
> When archive-before-PR is not in effect (`engram`), Stacked-to-Main is unaffected.
> Full rule: `[[harness-workflow]]` "Stacked-to-Main Archive Convergence" and the
> `chained-pr` skill.

The FBC steps 6–7 and the Stacked ASCII/paragraph (`:44-54`) are untouched.

### 1.5 [POINTER-ONLY] `CLAUDE.md`

**Edit G — REPLACE the deferral pointer with a converge pointer.** Grep target
(currently `:35-36`):

> "Terminal Phase Ordering (Feature Branch Chain)" and the `chained-pr` skill).
> Stacked-to-Main archive ownership is deferred to slice 2b.

Replace the final sentence ("Stacked-to-Main archive ownership is deferred to slice
2b.") — keep the FBC sentence at `:32-35` verbatim — with:

> When archive-before-PR is in effect, pure Stacked-to-Main is unsupported: the
> orchestrator converges to Feature Branch Chain at `sdd-tasks` strategy selection
> (full rule: `openspec/specs/harness-workflow/spec.md` "Stacked-to-Main Archive
> Convergence" and the `chained-pr` skill).

Pointer only: names the rule + its two full-rule homes, does NOT restate the
narrative. The single-PR line (`:30`) and FBC pointer (`:32-35`) stay verbatim.

### 1.6 [POINTER-ONLY] `skills/harness-workflow/SKILL.md`

**Edit H — REPLACE the deferral pointer with a converge pointer.** Grep target
(currently `:31-32`):

> Chain)" and the `chained-pr` skill. Stacked-to-Main archive ownership is deferred
> to slice 2b.

Replace the final sentence ("Stacked-to-Main archive ownership is deferred to slice
2b.") — keep the FBC pointer at `:27-31` verbatim — with:

> When archive-before-PR is in effect, pure Stacked-to-Main is unsupported and
> converges to Feature Branch Chain at `sdd-tasks` strategy selection. Full rule:
> `openspec/specs/harness-workflow/spec.md` "Stacked-to-Main Archive Convergence" and
> the `chained-pr` skill.

Pointer only; no narrative. The single-PR terminal note (`:23-25`) and FBC terminal
note (`:27-31`) stay verbatim.

### 1.7 [MECHANICS] `skills/sdd-apply/SKILL.md`

The FBC branch-targeting and archive-ownership notes (`:96-101`, `:277-303`) are
already correct from slice 2 and stay VERBATIM. Slice 2b touches ONE spot: the
`stacked-to-main` branch-targeting bullet, so an apply that inherited a Stacked value
under archive-before-PR is caught.

**Edit O — annotate the `stacked-to-main` bullet.** Grep target (currently `:97`):

> - `stacked-to-main`: each PR targets the previous PR's branch (or `main` after the previous merges).

Append a scoped sentence:

> This value is only valid when archive-before-PR is NOT in effect (`engram`). If the
> tasks artifact carries `Chain strategy: stacked-to-main` while archive-before-PR is
> in effect (`openspec`/`hybrid`), the convergence gate in `sdd-tasks` was skipped:
> STOP and return `blocked` with `Stacked-to-Main + archive-before-PR is unsupported;
> converge to feature-branch-chain before opening any child PR` (see
> `[[harness-workflow]]` "Stacked-to-Main Archive Convergence"). Do not merge any
> child PR to `main`.

This is the operational backstop for delta scenario 2 ("no partial main merges before
owning ref"): if the up-front gate was somehow bypassed, apply blocks before the first
child PR merges. It is prose returning a `blocked` string (matching the existing
blocked-message pattern at `:103`), NOT a Go error. The existing `feature-branch-chain`
bullet and archive-commit note (`:98-101`) stay verbatim.

### 1.8 [MECHANICS] `skills/sdd-archive/SKILL.md`

Already fully FBC-aware from slice 2 (Timing `:91-94`, Step 3c `:187-188`, Step 3d
`:200`/`:212-215`, checklist `:232-235`). **No archive-mechanics edit needed.** Slice
2b adds ONE optional convergence-precondition note so the archive step names that a
converged change is an FBC change.

**Edit J (OPTIONAL) — precondition note in Timing.** After the FBC sentence in Timing
(grep target: `...becomes "before the tracker PR merges to `main`."**`, currently
`:94`), apply MAY append:

> A change that began as Stacked-to-Main but had archive-before-PR in effect will have
> already converged to Feature Branch Chain at `sdd-tasks` strategy selection (see
> `[[harness-workflow]]` "Stacked-to-Main Archive Convergence"); by archive time it is
> an FBC change and this Timing rule applies unchanged.

OPTIONAL — the shipped FBC Timing rule already governs post-convergence; this note is
a clarifying breadcrumb, not a new mechanic. Apply MAY skip if it adds no clarity.

### 1.9 [MECHANICS] `skills/branch-pr/SKILL.md`

Already carries the FBC archive precondition (step 5b `:38-44`) from slice 2. **No new
mechanics.** Slice 2b's cross-ref is OPTIONAL.

**Edit N (OPTIONAL) — cross-ref on step 5b.** After step 5b (grep target: `...Child
PRs carry no\n   archive commit.`, currently `:40-41`), apply MAY append:

> (A change with archive-before-PR in effect that would otherwise be Stacked-to-Main
> has already converged to Feature Branch Chain at `sdd-tasks`; there is no
> Stacked-to-Main archive path — see `[[harness-workflow]]` "Stacked-to-Main Archive
> Convergence".)

OPTIONAL — step 5b already covers the FBC precondition; this only names why no Stacked
path exists. Apply MAY skip.

### 1.10 [MECHANICS] `skills/_shared/session-status-contract.md`

Already FBC-aware from slice 2 (Archive bullet `:22-28`: root-resident through the
integrated judge, moved on the tracker branch). **Confirm ONLY — no edit required.**
The convergence produces an FBC change, so the existing "root-resident during work,
moved on the tracker branch at archive" wording already holds through convergence.
V-check V6 asserts this; apply makes NO edit here unless V6 finds a gap.

---

## 2. Where the up-front decision lands (crux, restated for apply)

The converge decision is enforced in **exactly one behavioral home**:
`skills/sdd-tasks/SKILL.md`, Edit T1 (new step 4a, between the strategy-choice
prompt and the cache-write). Everything else is a copy/pointer/backstop:

- **Enforced (behavioral):** `sdd-tasks` step 4a (T1) — MUST NOT select Stacked, MUST
  select FBC + notify, before caching the choice.
- **Guard-contract note:** `sdd-tasks` T2 — the `Chain strategy` literal line carries
  the constraint so the cached value is already converged.
- **Full-rule copies:** spec (A2) + `chained-pr` Hard Rule (B) + Decision Gate row (C).
- **Backstop:** `sdd-apply` O — if the gate was bypassed and a Stacked value reaches
  apply under archive-before-PR, block before any child PR merges (delta scenario 2).
- **Pointers:** `CLAUDE.md` (G), `harness-workflow/SKILL.md` (H).

The reason for a single behavioral home + one backstop (not enforcement scattered
across skills): the approved decision is "up-front and decisive at `sdd-tasks`." The
`sdd-apply` backstop exists ONLY to honor the spec's scenario-2 "no partial main
merges before an owning ref" without adding a new state-machine transition (§3).

---

## 3. Deferral replacement map (no silent gap)

Every current "deferred to slice 2b" / "NOT yet defined" passage and its replacement:

| Location (grep) | Current deferral text | Replacement | Role |
|---|---|---|---|
| `spec.md` (~`:281-287`) | Scenario "Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)" | REMOVED (A1); superseded by ADDED "Stacked-to-Main Archive Convergence" requirement + 5 scenarios (A2) | SOURCE-OF-TRUTH |
| `chained-pr/SKILL.md` (~`:31-33`) | Hard Rule "NOT yet defined — deferred to slice 2b" | Full converge rule (B) + Decision Gate row (C) | FULL-RULE |
| `chaining-details.md` (~`:56-57`) | "NOT yet defined (deferred to slice 2b)" note | Converge note (E) | MECHANICS |
| `CLAUDE.md` (~`:35-36`) | "Stacked-to-Main archive ownership is deferred to slice 2b." | Converge pointer (G) | POINTER |
| `harness-workflow/SKILL.md` (~`:31-32`) | "Stacked-to-Main archive ownership is deferred to slice 2b." | Converge pointer (H) | POINTER |

No other file contains the deferral string (grep confirmed during design). After
apply, V5 re-greps `deferred to slice 2b` / `NOT yet defined` across the repo and
MUST return zero hits in these five files.

---

## 4. Consistency strategy + edit order

**Source-of-truth vs downstream.** The spec (A1/A2) is authoritative. The behavioral
enforcement lives in `sdd-tasks` (T1/T2). `chained-pr` (B/C/D) carries the full
orchestrator narrative. `CLAUDE.md` (G) and `harness-workflow/SKILL.md` (H) carry
POINTER lines that quote the exact new header `Stacked-to-Main Archive Convergence`
and name the two full-rule homes — they must NOT restate the body. Mechanics files
(O, and optional J/N) express the same single fact ("pure Stacked-to-Main +
archive-before-PR is unsupported → converge to FBC at `sdd-tasks`") in operational
voice, scoped to each concern.

**Three testable facts every full-rule/pointer copy must agree on:**
1. Trigger = archive-before-PR in effect (`openspec`/`hybrid`).
2. Decision = MUST NOT select pure Stacked-to-Main; select FBC + notify the user;
   up-front at `sdd-tasks` strategy selection.
3. Late Stacked→FBC conversion is unsanctioned with NO recovery procedure; `engram`
   is unaffected.

**Recommended edit order (source-of-truth first, pointers last):**
1. **A1, A2** — spec (source of truth) — establishes canonical wording + the exact
   header string.
2. **T1, T2** — `sdd-tasks` (behavioral home) — the crux.
3. **B, C, D** — `chained-pr/SKILL.md` (full-rule skill) — mirror the three facts.
4. **E** — `chaining-details.md` (mechanics note).
5. **O** — `sdd-apply/SKILL.md` (backstop).
6. **J, N** — OPTIONAL mechanics notes (`sdd-archive`, `branch-pr`).
7. **G, H** — POINTER lines (`CLAUDE.md`, `harness-workflow/SKILL.md`) LAST, so they
   can quote the now-final header string `Stacked-to-Main Archive Convergence`
   byte-for-byte.

Doing pointers last guarantees the quoted header matches the merged spec exactly (V3).

---

## 5. Verification hooks (docs-consistency; no automated Go tests)

Prose change — NO Go tests exist or are added (explore + slice 2 confirmed). Apply and
verify use these manual doc-consistency checks:

**V1 — single-PR + FBC requirement bodies unchanged (regression guard for #96/#97).**
- `git diff` on `openspec/specs/harness-workflow/spec.md`: ZERO changes within the
  single-PR requirement (grep `### Requirement: Terminal Phase Ordering (Single-PR
  Flow)`). ZERO changes within the FBC requirement body EXCEPT the removal of the
  Stacked-to-Main deferral scenario. The FBC header, three clauses, SESSION_STATUS
  paragraph, 4-step order, and six retained scenarios are byte-identical.
- The only spec additions are the removed 7-line deferral scenario (deletion) and the
  appended "Stacked-to-Main Archive Convergence" requirement.

**V2 — up-front decision present in `sdd-tasks` (the crux).** Grep `sdd-tasks/SKILL.md`
for `archive-before-PR` and `Feature Branch Chain`: step 4a (T1) MUST state MUST-NOT-
select-Stacked + MUST-select-FBC + notify, positioned BEFORE the "Cache the user's
choice" step. T2's `Chain strategy` constraint sentence MUST be present after the
fenced forecast block. The literal guard values
`stacked-to-main|feature-branch-chain|size-exception|pending` are UNCHANGED.

**V3 — pointer lines quote the exact header.** In `CLAUDE.md` (G) and
`harness-workflow/SKILL.md` (H), grep for the literal string `Stacked-to-Main Archive
Convergence` and confirm it matches the appended spec requirement header exactly, and
that `chained-pr` is named as the skill home. Confirm neither pointer restates the
requirement body (length check: 1–2 sentences each).

**V4 — the 5 new scenarios reflected in prose.** Map each ADDED spec scenario to at
least one skill sentence:

| ADDED scenario (spec) | Reflected in |
|---|---|
| Converges to FBC at sdd-tasks | `sdd-tasks` 4a (T1); `chained-pr` Hard Rule (B) + gate row (C) |
| Pure Stacked + archive-before-PR unsupported — no partial main merges | `sdd-apply` backstop (O); `chained-pr` gate row (C) |
| After convergence, FBC rule governs — no new mechanics | `chained-pr` Hard Rule (B); `chaining-details` (E); optional `sdd-archive` note (J) |
| Late Stacked→FBC out of scope — stranded slices hazard | `sdd-tasks` 4a (T1, "not sanctioned, no recovery"); `chained-pr` Hard Rule (B) |
| Archive-before-PR not in effect — Stacked unaffected (`engram`) | `sdd-tasks` 4a (T1); `chained-pr` gate row (C); `chaining-details` (E); pointers (G, H) |

**V5 — deferral fully replaced (not silent).** Grep the repo for `deferred to slice
2b` and `NOT yet defined`: ZERO hits in `spec.md`, `chained-pr/SKILL.md`,
`chaining-details.md`, `CLAUDE.md`, `harness-workflow/SKILL.md`. Each now states the
converge decision (§3 map).

**V6 — SESSION_STATUS root-residency holds through convergence.** Confirm
`session-status-contract.md` Archive bullet (`:22-28`) still reads "root-resident
through the integrated judge, moved on the tracker branch at archive" — unchanged, and
correct for a converged FBC change. No edit expected; V6 asserts the invariant holds.

**V7 — NO Go touched.** `git diff --stat` shows changes ONLY under
`openspec/`, `skills/`, and `CLAUDE.md`. ZERO `.go` files. If any `.go` appears, STOP
(§7 open question 3).

**V8 — link integrity.** Wikilinks `[[harness-workflow]]`, `[[chained-pr]]`,
`[[sdd-archive]]`, `[[harness-judge]]`, `[[session-status-contract]]` resolve; run
`archon map --check` at archive (the delta's own relative links are handled by
archive backfill).

Note: there are NO automated Go tests for this prose; V1–V8 are manual/grep checks
run by apply and verify.

---

## 6. Risk-targeted mitigations (per proposal risk)

| Proposal risk | Design choice that mitigates it |
|---|---|
| **Convergence timing ambiguous (tasks vs. archive time)** (Med) | Resolved decisively UP-FRONT at `sdd-tasks` strategy selection (Edit T1, new step 4a, before the choice is cached). Timing is not left to archive; by archive time the change is already FBC. §2 documents the single behavioral home. |
| **Late conversion strands already-merged slices on `main`** (Med) | Up-front gate (T1) means an archive-before-PR user NEVER selects pure Stacked-to-Main, so no slice merges to `main` without an owning tracker ref. The `sdd-apply` backstop (O) blocks before the FIRST child PR merge if the gate was bypassed (delta scenario 2). Late conversion is documented as an unsanctioned hazard with NO recovery procedure (T1, B, spec scenario 4) — a loud dead-end, not a silent trap. |
| **Three phase-order copies drift** (Low) | Full rule in exactly two homes (spec A2 + `chained-pr` B/C); `CLAUDE.md` (G) and `harness-workflow/SKILL.md` (H) get pointer lines that quote the exact header `Stacked-to-Main Archive Convergence`. Edit order (§4) does pointers LAST so the quoted header is final. V2/V3/V5 verify alignment, pointer resolution, and zero remaining deferral. Mirrors the slice-2 fan-out decision. |

---

## 7. Open questions (for the Human Review Gate)

1. **Backstop scope in `sdd-apply` (Edit O).** The design adds a `blocked`-returning
   backstop so a bypassed gate can't merge a child PR to `main` under
   archive-before-PR (honoring spec scenario 2 without a new state-machine
   transition). Confirm this belongs in `sdd-apply` prose, or prefer it live ONLY in
   the spec/`chained-pr` and rely on the up-front gate alone. Recommendation: keep O —
   it is the only enforcement point between selection and the first child merge, and
   it costs one prose paragraph.

2. **Optional mechanics notes (Edits D, J, N).** D (`chained-pr` step 2 clause), J
   (`sdd-archive` Timing breadcrumb), N (`branch-pr` step 5b cross-ref) are flagged
   OPTIONAL to avoid redundancy — the shipped slice-2 FBC mechanics already govern
   post-convergence. Approve include-all, or drop them and rely on T1/B/C/E/O/G/H.
   Recommendation: include N (cheap, names why no Stacked archive path exists), treat
   D and J as skip-unless-clarifying.

3. **Code implication (should be none).** Explore + slice 2 confirmed no Go enforces
   chain-strategy selection or terminal ordering; the "blocked"/"notify" behavior is
   orchestrator prose. If apply discovers ANY Go path (e.g. a `harness-workflow`
   validator inspecting chain strategy or PR/branch state), STOP and surface it — this
   design covers docs/skills only and would need a follow-up. (V7 is the guard.)
