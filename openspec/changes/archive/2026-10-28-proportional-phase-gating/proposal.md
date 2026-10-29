# Proposal: Proportional Phase Gating

## Intent

Ceremony today is flat: every `full`-track change runs all 9 phases (Rule 3)
and judge is an unconditional static gate (`judge.enabled`, Rule 6). Adapt
ODD's proportionality philosophy — review scales with risk/size — via (1) a
sized `fast` track plus a classifier that selects it, and (2) judge becoming
a per-change computed risk tier.

## Scope

### In Scope
- `fast` track in `trackOrder()` (`internal/workflow/workflow.go`), mirroring
  `BugfixTrack` — same `CheckTransition`/`Advance`/`state.yaml` machinery.
- Sizing step in `sdd-explore` (Step 3c: `small`/`substantial`/`unknown`),
  mirroring Step 3b's web detection; feeds `fast` vs `full` via the existing
  preflight deviation-question pattern — no new UX.
- `judge.enabled` becomes a computed risk tier (`harness-judge` Step 0);
  low-risk full-track changes may skip/lighten judge.
- Spec deltas: new `harness-fast-track`; MODIFIED `harness-workflow` ("Single
  Judge in the Full-Track Terminal Ordering") and `harness-judge`.

### Out of Scope
- `sdd-router` size/risk detection (stays phase-family-only).
- `judgment-day` dual review (untouched).
- `state.yaml` bootstrap fix (already shipped).

## Capabilities

### New Capabilities
- `harness-fast-track`: `fast` track phase sequence and gates, sibling to [[harness-bugfix-track]].

### Modified Capabilities
- `harness-workflow`: adds `fast` as a third `track` value; MODIFIES "Single Judge in the Full-Track Terminal Ordering" to allow a computed risk-tier skip/lighten.
- `harness-judge`: Step 0 gate changes from static boolean to computed per-change tier.

## Approach

Reuse track-as-data (new `fast` value, copying `BugfixTrack`'s shape) and a
`sdd-explore` sizing sub-step (Step 3b's precedent) to select it; layer judge
risk-tiering onto the existing `judge.enabled` seam — all proven patterns.

## Affected Areas

| Area | Impact |
|------|--------|
| `internal/workflow/workflow.go`, `_test.go` | New `fast` track + tests |
| `skills/sdd-explore/SKILL.md` | Step 3c sizing classification |
| `skills/harness-judge/SKILL.md` | Step 0 reads computed tier |
| `skills/harness-workflow/SKILL.md` | Document third track |
| `internal/initcmd/templates.go` | CLAUDE.md Rule 3/6 wording |
| `openspec/specs/harness-workflow/spec.md` | MODIFIED Single Judge delta |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| MODIFIED delta weakens an existing guarantee | Med | Scenario-backed delta, explicit old/new behavior |
| Fuzzy classification under-ceremonies a change | Med | Default `full` on `unknown`; one scoped question |
| Rule 3/6 prose-only; docs/code drift | Low | Spec deltas + `templates.go` updated together |

## Rollback Plan

Independently revertable: drop `fast` from `trackOrder()` (`full`/`bugfix`
unaffected), revert Step 3c, revert `harness-judge` Step 0 to a plain
boolean. No data migration — new fields are additive.

## Dependencies

- `state.yaml` bootstrap fix (shipped).

## Success Criteria

- [ ] `small`-classified change runs `fast` end-to-end, no judge phase.
- [ ] `substantial`-classified change runs full 9 phases with judge.
- [ ] `harness-workflow` delta reviewed without weakening the guarantee for high-risk changes.
- [ ] Go tests cover `fast` transitions like `bugfix`.

## Proposal question round — CONFIRMED by user

- **Auto-select vs suggest `fast`?** Confirmed: suggest only, via the existing preflight deviation-question pattern — never silently auto-pick.
- **Does `judge.enabled=false` become a tier special case, or get replaced?** Confirmed: special-case override (`enabled=false` always skips) layered under the new tier, so existing configs keep working.
- **Does `fast` skip Human Review Gate phases?** Confirmed: no — `fast` shortens the phase list, like `bugfix`; included phases still gate on human review.
