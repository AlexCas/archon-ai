# Proposal: Chained-PR Archive Ownership (Feature Branch Chain)

<!-- Link convention: [[capability]] wikilinks for capability identity; relative
     links for intra-change navigation. Full rule: skills/_shared/spec-vault.md. -->

## Intent

Slice 1 (#96) fixed archive-before-PR ordering for the **single-PR flow** only.
For chained flows the harness still says archive ownership is *undefined /
slice-2 pending* (`harness-workflow` spec scenario "Chained-PR flow is out of
scope"; `chained-pr/SKILL.md:24`). This reproduces the #93 pain for chains: with
no owning PR, the archive commit either lands after the integrated judge with no
home, or pollutes a trailing PR. This slice closes that gap for the **Feature
Branch Chain** strategy, whose draft/no-merge tracker PR gives archive a
judge-legal home. Stacked-to-Main has no tracker and is deferred (slice 2b).

## Scope

### In Scope
- Define a **Feature Branch Chain archive rule**: the archive commit is staged on
  the **tracker/integration branch**, AFTER the integrated judge passes and
  BEFORE the tracker PR merges to `main`. It travels inside the tracker PR.
- `SESSION_STATUS.md` moves in that same tracker archive commit; stays
  root-resident through the integrated judge even while child PRs are open.
- Convert the blanket "all chains deferred" language into "Feature Branch Chain
  defined; Stacked-to-Main pending slice 2b" (documented, not silent).
- Preserve slice-1 archive-internal order and one-commit staging unchanged; only
  the OWNING branch changes (change's single PR → tracker PR).

### Out of Scope (Non-goals)
- **Stacked-to-Main archive ownership** — deferred to slice 2b (documented).
- `archon archive` CLI / any Go enforcement (explore: zero Go enforces ordering).
- Any change to single-PR "Terminal Phase Ordering (Single-PR Flow)" behavior.

## Capabilities

> Contract with sdd-spec. Researched `openspec/specs/`.

### New Capabilities
- None.

### Modified Capabilities
- `harness-workflow`: ADD a chained "Terminal Phase Ordering (Feature Branch
  Chain)" requirement — archive staged on the tracker branch after the integrated
  judge, before the tracker merges. Single-PR requirement stays verbatim.

## Approach

Docs-and-skills only. Add the tracker-owned archive rule as a NEW requirement in
`openspec/specs/harness-workflow/spec.md` (single-PR requirement untouched), then
propagate the chained narrative to the skills that describe terminal ordering and
chaining. Slice-1 invariants that MUST be preserved and named:

1. Single-PR "Terminal Phase Ordering (Single-PR Flow)" requirement + its 6
   scenarios stay VERBATIM — slice 2 ADDS a chained rule.
2. Judge gating: archive never runs before judge passes (both flows).
3. Archive-internal order (spec merge → folder move → `archon map --backfill`/
   `--check` STOP gate → `SESSION_STATUS.md` move) + one-commit staging.
4. `SESSION_STATUS.md` root-residency until the archive commit; session-status
   invariants (one file/session, moved-not-deleted, read-first-on-crash).

**Judge-timing requirement (explore Q2):** the archive-on-tracker rule assumes the
**integrated judge runs on the tracker branch BEFORE the tracker merges to
`main`**. The current spec does not guarantee this; slice 2 MUST make it explicit.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `openspec/specs/harness-workflow/spec.md` | Modified | ADD chained terminal-ordering requirement; single-PR requirement verbatim |
| `skills/chained-pr/SKILL.md` | Modified | Replace `:24` blanket deferral with Feature Branch Chain rule + tracker ownership |
| `skills/chained-pr/references/chaining-details.md` | Modified | Add archive step to Feature Branch Chain sequence |
| `skills/harness-workflow/SKILL.md` | Modified | Chained terminal-ordering narrative |
| `skills/sdd-archive/SKILL.md` | Modified | Tracker-branch owning in chained flow |
| `skills/branch-pr/SKILL.md` | Modified | Tracker PR carries the archive commit |
| `skills/sdd-apply/SKILL.md` | Modified | Apply hands off to archive-on-tracker |
| `skills/_shared/session-status-contract.md` | Modified | Move happens in tracker archive commit |
| `CLAUDE.md` | Modified | Orchestrator rule for chained archive ownership |
| `skills/work-unit-commits/SKILL.md` | Optional | Archive as a tracker work unit |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Regressing #96 single-PR behavior | Med | Single-PR requirement + 6 scenarios stay verbatim; chained rule is additive |
| Reintroducing "polluted last PR" | Med | Archive on tracker branch (integration history), not a child PR diff |
| Integrated-judge timing unspecified | Med | Slice 2 makes tracker-before-merge judge explicit (explore Q2) |
| Phase-order copies drift out of sync | Low | Decide copy scope at spec (open question below) |

## Rollback Plan

Docs-only. Revert the slice-2 commits; the single-PR requirement and all skill
prose return to their #96 state. No data, config, or Go behavior to unwind.

## Dependencies

- Slice 1 (#96, archived `2026-07-30-archive-before-pr`) — the invariants extended here.

## Open Questions (for the Human Review Gate)

- **Q (explore #4):** Do all three phase-order copies (`CLAUDE.md`,
  `harness-workflow/SKILL.md`, spec) each need the chained narrative, or only the
  spec + `chained-pr` skill, with the others pointing to them?
- **Q (explore #2):** If the spec does not already guarantee the integrated judge
  runs on the tracker before merge, does slice 2 add that as a new scenario or a
  same-requirement clause?

## Success Criteria

- [ ] `harness-workflow` spec defines Feature Branch Chain archive ownership
      (tracker branch, after integrated judge, before tracker merge).
- [ ] Single-PR requirement + its 6 scenarios remain verbatim (no diff).
- [ ] `chained-pr/SKILL.md:24` blanket deferral becomes "Feature Branch Chain
      defined; Stacked-to-Main pending slice 2b".
- [ ] `SESSION_STATUS.md` chain behavior documented (moves in tracker archive
      commit; root-resident through integrated judge).
- [ ] Stacked-to-Main deferral to slice 2b is explicit, not silent.
