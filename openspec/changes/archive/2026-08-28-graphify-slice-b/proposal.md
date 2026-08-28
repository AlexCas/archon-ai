# Proposal: Graphify Slice B — graph diff in verify + edge evidence in judge

## Intent

Slice A wired Graphify into `sdd-explore` (extract) and `sdd-tasks` (read-only).
Slice B closes the deferred second half: surface *structural* change after apply.
`sdd-verify` gains an advisory code-graph diff (what symbols/edges appeared or
vanished), and `harness-judge` lets judges cite deterministic `EXTRACTED` edges
as backing for findings they already reached. First Graphify coupling into
verify/judge — must stay fully inert when `graphify.enabled: false`.

## Empirical findings (blocker 1 — resolved)

Installed the pinned `graphifyy==0.9.45` in a throwaway venv and ran the extract
(real verb is `update <path>`, no LLM) twice on `internal/config`:

- **Node/edge IDs are stable and semantic**, e.g. `config_config` = `pkg_symbol`,
  never line-number/hash-derived. Line info lives in a separate `source_location`
  field, so line drift does NOT churn IDs.
- **Two consecutive extracts: byte-identical** — 70 nodes / 146 unique links,
  full records equal (incl. `source_location`).
- **A single added function → exactly one new node, zero spurious churn.**
- Edge provenance is the field `confidence: "EXTRACTED"` (vs `INFERRED`); relation
  in `relation`. **No native `diff` verb** (verbs incl. `update|path|explain|
  diagnose|merge-graphs`) — a diff is still a harness-side set-difference.

Conclusion: a naive set-diff is viable and clean, NOT noisy. The safer fallback
(Approach 3) is no longer forced; **Approach 1 (both features) is proposed.**

## Scope

### In Scope
- `sdd-verify`: advisory `### Code Graph Diff (advisory)` NOTE, gated on
  `graphify.enabled`, modeled on the existing Impeccable presence NOTE.
- `harness-judge`/`judgment-day`: gated note that judges MAY cite `EXTRACTED`
  edges to enrich existing findings.
- `graphify` skill: extend §3 per-phase map with a verify row + baseline-snapshot
  rule; amend the §5 "sole extraction site" invariant.

### Out of Scope
- New Go code, new `graphify.diff` config knob, `semantic: true`/`INFERRED` edges.
- Any blocking behavior: no 4th column in judge Step 4, no re-apply trigger.
- Slice C (mapgen spec-graph ↔ code-graph bridge).

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `graphify-integration`: extraction-site invariant relaxes to allow a
  verify-time re-extract; per-phase map gains verify/judge rows.

## Approach

Skill-only (Approach 1), mirroring Slice A. Baseline = explore-time `graph.json`
copied aside before verify re-extracts the "after"; diff = set-difference over
node IDs and (source,target,relation) tuples, rendered as an advisory NOTE.

## Requires Human Review Gate sign-off (blocker 2)

Approach 1 makes **verify a second extraction site**, relaxing Slice A's
`skills/graphify/SKILL.md` §5 "sdd-explore is the sole extraction site" invariant.
This is NOT silently assumed — the user must approve it at the gate. If declined,
fall back to **Approach 3 (judge edge-evidence only)**, which queries the single
post-apply graph, needs no baseline, and leaves the invariant untouched.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Diff parse fails on 0.9.x schema drift | Low | v0.9.45 pin; degrade to baseline + one-line advisory note |
| Edge evidence creeps into a blocking gate | Med | Explicit "never a column, never a re-apply trigger" guardrail |
| Double extraction cost on large repos / retries | Low | `semantic:false` local-only; diff once at first verify |
| Invariant relaxation done silently | Low | Gate sign-off required; amend §5 in same change |

## Rollback Plan

Skill-only: revert the prose edits to `sdd-verify`, `harness-judge`/`judgment-day`,
and `graphify` skills. No Go, migrations, or config to unwind. Feature is inert
under `graphify.enabled: false` regardless.

## Dependencies

- Graphify `graphifyy==0.9.45` present and `graphify.enabled: true` (both opt-in).

## Success Criteria

- [ ] verify emits an advisory graph-diff NOTE only when `graphify.enabled`; never alters PASS/FAIL.
- [ ] judge may cite `EXTRACTED` edges without adding a gate/column/re-apply trigger.
- [ ] All three skills stay byte-inert when `graphify.enabled: false`.
- [ ] Extraction-site invariant change is explicitly approved and reflected in §5.
