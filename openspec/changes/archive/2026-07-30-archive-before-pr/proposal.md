# Proposal: Run archive before opening the PR

<!-- Link convention: [[capability]] wikilinks for capability identity, relative
     links for intra-change navigation. Full rule: skills/_shared/spec-vault.md. -->

## Intent

The archive step runs AFTER the PR is opened, so the archive commit either lands as
an extra archive-only PR or pollutes the last PR. Proven in the `local-model-provider`
cycle: archive commit `2a62962` (2026-07-28 21:55) landed ~2 min AFTER PR-B commit
`dde17a8` (21:53) as a separate archive-only commit — the exact pain (issue #93).
Archive redefines the change's own history (spec merge, folder move, `SESSION_STATUS.md`
move); that history should travel INSIDE the change's PR, not chase it.

## Scope

### In Scope (slice 1 — single-PR flow, docs+skills only)
- Redefine archive as a **pre-PR, judge-gated** step: archive operations run AFTER
  judge passes but BEFORE the PR is opened, so the archive commit is included in the
  change's own PR.
- Sequence `SESSION_STATUS.md` move against `session-status-contract`: keep the root
  copy until the archive commit is staged, then move it into the archived folder as
  part of that same commit (crash recovery preserved until the move commits).
- Update the phase-order / gate narrative and the executor/PR skills so the flow is:
  judge pass → archive (spec merge, folder move, `SESSION_STATUS.md` move, `archon map`
  backfill+check) → open PR carrying the archive commit.

### Out of Scope
- **Chained-PR archive ownership** — which PR in a stack owns the archive move (explore
  open-question #2). Deferred to **slice 2**; single-PR flow only here.
- **`archon archive` CLI command** — archive stays skill/LLM-driven. Optional future,
  not this slice.
- Any change to judge gating semantics — judge must still pass before archive.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `harness-workflow`: the phase state machine's terminal narrative changes — archive
  precedes PR-open in the single-PR flow, and the archive commit is included in the
  change's PR. `openspec/specs/harness-workflow/spec.md` is the harness's own source
  of truth (self-hosting), so the delta merges back into it.

## Approach

Docs-and-skills only. Concrete sequencing rule (single-PR flow):

1. **Judge passes** (gate unchanged — archive still requires judge pass).
2. **Archive runs pre-PR**: spec merge, folder move to
   `openspec/changes/archive/YYYY-MM-DD-{name}/`, `SESSION_STATUS.md` move into that
   folder, `archon map --backfill` + `--check` — all staged into ONE archive commit.
3. **PR is opened** on the change branch; the archive commit is already part of the
   branch, so it travels inside the change's PR (no trailing archive-only commit).

`SESSION_STATUS.md` timing: keep the root copy live through judge; move it only when
the archive commit is being staged. The `session-status-contract` invariant
("root-resident during work, moved at archive") holds — the move still happens at
archive, just before PR-open instead of after.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `CLAUDE.md` | Modified | Phase-order / gate narrative: archive precedes PR-open (single-PR) |
| `skills/harness-workflow/SKILL.md` | Modified | Terminal-phase + PR-relative ordering narrative |
| `skills/sdd-archive/SKILL.md` | Modified | Executor runs pre-PR; stage archive into one commit; `SESSION_STATUS.md` move timing |
| `skills/harness-judge/SKILL.md` | Modified | Downstream after judge pass = archive-then-PR |
| `skills/branch-pr/SKILL.md` | Modified | Open PR AFTER the archive commit exists on the branch |
| `skills/chained-pr/SKILL.md` | Modified | Note single-PR archive-before-PR; chained ownership = slice-2 non-goal |
| `skills/sdd-apply/SKILL.md` | Modified | PR-open boundary references archive-then-PR order |
| `skills/_shared/session-status-contract.md` | Modified | Move timing: at archive, before PR-open |
| `skills/_shared/openspec-convention.md` | Modified | Archive-then-PR note if convention narrates PR order |
| `openspec/specs/harness-workflow/spec.md` | Modified | Self-hosting source of truth; delta merges here at archive |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Chained-PR flow left ambiguous by a single-PR rule | Med | Explicit slice-2 non-goal; chained-pr skill flags "single-PR only" |
| `SESSION_STATUS.md` moved too early breaks crash recovery | Med | Keep root copy until the archive commit is staged; move IN that commit |
| `archon map --backfill/--check` assumes folder move is last mutation | Med | Keep archive-internal order (merge → move → map); PR-open is external to it |
| Judge gate weakened by re-sequencing | Low | Gate unchanged — archive still requires judge pass; only PR-relative position moves |
| Docs across three phase-order copies drift | Med | Enumerate all three copies (CLAUDE.md, SKILL, spec) as affected; keep in sync |

## Rollback Plan

Docs-and-skills only — no code. Revert the change's commits to restore archive-after-PR
narrative. No state, data, or CLI migration.

## Dependencies

- Judge phase continues to gate archive (unchanged).
- `archon map --backfill`/`--check` behavior unchanged.

## Success Criteria

- [ ] Single-PR flow: the archive commit (spec merge + folder move + `SESSION_STATUS.md`
      move + map backfill) lands INSIDE the change's PR — no trailing archive-only commit.
- [ ] Judge still gates archive (archive never runs before judge passes).
- [ ] `SESSION_STATUS.md` stays root-resident until the archive commit is staged, then
      moves into the archived folder in that same commit.
- [ ] All three phase-order copies (CLAUDE.md, harness-workflow SKILL, harness-workflow
      spec) narrate archive-before-PR consistently.
- [ ] Chained-PR ownership is explicitly documented as a slice-2 non-goal, not silently changed.
