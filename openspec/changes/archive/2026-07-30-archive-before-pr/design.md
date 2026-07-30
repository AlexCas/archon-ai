# Design: Run archive before opening the PR

<!-- Link convention: [[capability]] for capability identity; relative links for
     intra-change navigation ([proposal](proposal.md), same folder).
     Full rule: skills/_shared/spec-vault.md. -->

## Technical Approach

Docs-and-skills only. No Go code. We edit prose in one source-of-truth spec plus nine
downstream narrative files so the single-PR terminal sequence reads **judge pass →
archive (one commit) → open PR**. There is NO `archon archive` CLI; archive is
LLM-driven, so every edit is an instruction/order change telling the orchestrator and
executor *when* to act, not new code. Implements [[harness-workflow]] per
[specs/harness-workflow/spec.md](specs/harness-workflow/spec.md).

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Source of truth | `openspec/specs/harness-workflow/spec.md` is authoritative; CLAUDE.md + harness-workflow SKILL are downstream copies kept identical in meaning | Self-hosting: the delta merges here at archive. Three copies must not drift (proposal risk). |
| PR-open placement | Instruction, not code: branch-pr/chained-pr/sdd-apply gain a precondition "archive commit must already be on the branch" | No state machine node for PR; PR-open lives in skill prose outside `harness-workflow`. |
| One-commit staging | Describe the four archive sub-ops as staged into ONE commit before PR-open | Keeps archive history inside the change's PR (issue #93). |
| SESSION_STATUS.md timing | Move it *within* the archive-commit staging, before PR-open | Preserves crash recovery through judge; still "moved at archive" per contract. |
| Chained-PR | Explicit slice-2 non-goal note in chained-pr; rule scoped to single-PR | Avoids silently redefining stacked-PR ownership. |

## Data Flow (terminal sequence, single-PR)

    judge PASS ──► archive (ONE commit): spec merge → folder move →
                   SESSION_STATUS.md move → archon map --backfill/--check
                        │
                        └──► commit staged on change branch ──► open PR
                                                                (carries archive commit)

## File Changes — per-file edit plan

Source of truth first; downstream copies must match its meaning.

| File | Action | Edit (anchored) |
|------|--------|-----------------|
| `openspec/specs/harness-workflow/spec.md` (`:5` purpose; Requirements block) | Modify | Merge the delta's ADDED **"Terminal Phase Ordering (Single-PR Flow)"** requirement + its 6 scenarios verbatim into the Requirements section. This is the archive-time merge target; the design records it so apply lands the same text here that the other copies paraphrase. |
| `CLAUDE.md` (`:27-28` Phase Order; `:129-138` SESSION_STATUS; `:146-155` Rules) | Modify | After the Phase Order line, add one sentence: *"In the single-PR flow the archive commit is staged BEFORE the PR is opened, so archive history travels inside the change's PR."* In SESSION_STATUS rule `:137`, change "During `archive`, MOVE ... then remove it from the root" to note the move happens **as part of the staged archive commit, before the PR is opened**. Add a Rule (renumber): *"In the single-PR flow, run archive (spec merge, folder move, SESSION_STATUS.md move, `archon map`) as one commit AFTER judge passes and BEFORE opening the PR."* |
| `skills/harness-workflow/SKILL.md` (`:15-19` Phase Sequence; `:154` archive note) | Modify | Under Phase Sequence add a short "Terminal ordering (single-PR)" note mirroring the spec: judge pass → archive commit staged → PR opened; judge gating unchanged. Reword `:154` so `SESSION_STATUS.md` "is MOVED into the archived change folder during `sdd-archive`, within the staged archive commit, before the PR is opened." |
| `skills/sdd-archive/SKILL.md` (Purpose `:34`; Step 3b `:143-157`; Step 3c `:160-177`; Rules `:236-245`) | Modify | Add a **"Timing" note at the top of the archive steps**: archive runs after judge passes and BEFORE the PR is opened; stage the Step 2 spec merge, Step 3 folder move, Step 3c SESSION_STATUS.md move, and Step 3b `archon map --backfill`/`--check` into ONE commit; only after that commit is staged does the orchestrator open the PR. State the fixed internal order stays merge → move → SESSION_STATUS move → map (unchanged); PR-open is EXTERNAL. Add a Rule: *"Archive is pre-PR in the single-PR flow — never open the PR before the archive commit is staged."* |
| `skills/harness-judge/SKILL.md` (`:19,23,47` "advance to archive"; `:174,184` On Pass) | Modify | Where it says the orchestrator "advances to archive" on pass, append: *"then, after archive stages its commit, the PR is opened (single-PR flow)."* No change to judge gating or verdict logic — only clarify the post-pass ordering downstream. |
| `skills/branch-pr/SKILL.md` (`:30-38` Workflow; `:187-202` Commands) | Modify | Add a precondition to the Workflow list before "Open PR": *"If this is an SDD single-PR change, the archive commit (spec merge, folder move, SESSION_STATUS.md move, `archon map`) MUST already be staged on the branch before opening the PR."* Optionally a one-line note near Commands. |
| `skills/chained-pr/SKILL.md` (`:14-24` Hard Rules) | Modify | Add a Hard Rule / note: *"The archive-before-PR single-PR rule does NOT apply to chained/stacked flows. Which PR owns the archive move is a slice-2 non-goal; do not enforce archive position on chained PRs."* |
| `skills/sdd-apply/SKILL.md` (`:263-271` Workload / PR Boundary; Rules `:285`) | Modify | In the PR Boundary section note that for a single-PR change the archive commit is staged before PR-open, so apply's reported boundary ends at "ready for archive-then-PR." Add a Rule reinforcing PR-open follows the staged archive commit. No change to chained/stacked slice behavior. |
| `skills/_shared/session-status-contract.md` (`:24-25` Archive lifecycle) | Modify | Reword the **Archive** bullet: *"during `sdd-archive`, MOVE the file into the archived change folder ... **as part of the staged archive commit, before the PR is opened**, then delete it from the root."* Keep the "root-resident during work / read-first on crash recovery" invariants intact (`:16-21`). |
| `skills/_shared/openspec-convention.md` (`:26-27` SESSION_STATUS note; `:49,:58` map/archive-move rows) | Modify | Add a short archive-then-PR note to the SESSION_STATUS comment block (`:26-27`) and, if desired, a one-line note on the `sdd-archive Moves` row (`:58`) that the move is staged into the pre-PR archive commit. Row semantics unchanged. |

## Ordering-of-edits / consistency strategy

Apply MUST edit in this order to avoid partial/inconsistent states:

1. **Spec first** — land the merged requirement text in
   `openspec/specs/harness-workflow/spec.md` (the authoritative wording).
   *Note:* this is also what the archive step re-merges; the wording apply writes here
   and the delta's ADDED block must be byte-identical in meaning.
2. **Two downstream phase-order copies** (`CLAUDE.md`, `harness-workflow/SKILL.md`) —
   paraphrase the same rule; verify each states judge-gating-unchanged + archive-before-PR.
3. **Sequencing skills** (`sdd-archive`, `harness-judge`, `branch-pr`, `chained-pr`,
   `sdd-apply`) — insert the "archive staged before PR-open" instruction.
4. **Shared modules** (`session-status-contract`, `openspec-convention`) — align timing wording.

Consistency check for apply/verify: the three phase-order copies (spec, CLAUDE.md, SKILL)
must each contain, in their own words, (a) judge pass precedes archive, (b) archive is one
commit, (c) PR opens only after that commit. No copy may omit any of the three.

## Sequencing insertion point (precise)

- The **"archive runs before PR-open"** instruction is inserted at the *top of the archive
  step block* in `sdd-archive` (a new Timing note above Step 2) and at the *pre-"Open PR"*
  point in `branch-pr` Workflow (`:30-38`). These are the two operational anchors.
- **"One commit" staging** is described operationally as: after Step 2 (spec merge),
  Step 3 (folder move), Step 3c (SESSION_STATUS.md move), and Step 3b (`archon map
  --backfill`/`--check`) complete, stage all resulting changes and create a single
  archive commit on the change branch; do NOT open the PR until that commit exists. This
  is prose telling the LLM orchestrator the order — no CLI archive command is invoked.
- `harness-judge` only *clarifies* that "advance to archive" is followed by
  archive-then-PR; it does not gain gating logic.

## SESSION_STATUS.md handling

- `session-status-contract.md` Archive bullet and `sdd-archive` Step 3c are reworded so the
  move happens **when the archive commit is staged, before PR-open** — not after.
- Invariants preserved verbatim: root-resident from session start through judge; read-first
  on crash recovery; moved (not deleted) into
  `openspec/changes/archive/YYYY-MM-DD-{name}/SESSION_STATUS.md`. Because the move is inside
  the same pre-PR commit, crash recovery between judge and PR-open still finds the root file.

## Verification hooks

There are **no automated Go tests for skill prose** — verification is doc-consistency based:

- **Cross-copy consistency**: verify the three phase-order copies each narrate the three
  invariants above (judge-first, one-commit, PR-after). A grep for "archive" + "PR" in
  CLAUDE.md, harness-workflow SKILL, and the merged spec should surface matching statements.
- **Scenario coverage**: each of the 6 delta-spec scenarios maps to prose — e.g. Scenario
  "SESSION_STATUS.md moves in the archive commit (before PR-open)" ↔ contract + Step 3c
  wording; "Chained-PR flow is out of scope" ↔ chained-pr note.
- **No orphaned old wording**: verify no file still says archive runs *after* the PR or that
  SESSION_STATUS moves *after* PR-open.

## Migration / Rollout

No migration. Docs-and-skills only. Rollback = revert the change's commits.

## Open Questions

- [ ] Chained-PR archive ownership (which PR owns the move) — deferred to slice 2 by
      approved scope; flagged here only so verify does not treat the chained gap as a defect.
- [ ] None at code level: no Go/CLI implication found. If apply discovers any code path that
      assumes archive-after-PR, STOP and flag — the approved slice is docs-only.
