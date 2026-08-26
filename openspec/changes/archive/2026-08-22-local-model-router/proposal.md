# Proposal: Local Model Router — deterministic SDD phase dispatch

## Intent

Today the SDD flow relies entirely on the leader MODEL inferring which phase to
launch. On a weak local leader (`ollama qwen3-orch` ~5B), implicit phrasings fail:
the reported bug is **"Trabajemos en esta especificación" not launching explore**,
while the explicit "…Lanza el agente de exploración" works. State arithmetic
("the phase after propose") and resume-vs-start are unreliable in-model. The
prototype (`prototype/sdd-router/`) settled the fix empirically: a HYBRID router
scoring **17/18** on the local model. This change formalizes it in the harness.

## Scope

### In Scope
- `internal/route/route.go` — deterministic resolver: control words, start verbs,
  literal `archon-<phase>` tokens; reads `state.yaml` + `config.PhaseOrder`; READ-only.
- `archon route` cobra subcommand (`cmd/archon/main.go`) as the leader's shell boundary.
- `skills/sdd-router/SKILL.md` — MODEL classifier ("which phase family?" → phase|ASK).
- Leader wiring in `internal/initcmd/templates.go` (both Claude + Opencode variants).
- Reconcile pre-existing `skill_count` drift (see Risks).

### Out of Scope
- Re-deciding the hybrid architecture (settled by the prototype).
- Auto-triggering routing via provider hooks (Claude-only, future).
- Changing harness-workflow gate logic or state-write ownership.

## Capabilities

### New Capabilities
- `local-model-router`: deterministic pre-router + model classifier for phase dispatch.

### Modified Capabilities
- None (harness-workflow gate and state ownership are untouched).

## Approach

Deterministic slice → code; fuzzy slice → model; legality → harness-workflow.
`archon route` reads current phase/status and resolves control/start/literal cases;
everything else falls through to the in-context classifier; harness-workflow gates
legality; the `archon-<phase>` subagent runs. The router NEVER writes state and
NEVER bypasses preflight, vague-request, or human-review gates. CLI + SKILL.md stay
provider-neutral (a Go binary, no Claude-only primitives).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/route/route.go` (+test) | New | Deterministic resolver + fixtures port |
| `cmd/archon/main.go` | Modified | `+AddCommand(newRouteCmd)` |
| `skills/sdd-router/SKILL.md` | New | Model classifier; auto-embedded |
| `internal/initcmd/templates.go` (+test) | Modified | Routing step + count reconcile |
| `.archon/config.yaml`, `CLAUDE.md` | Modified | skill_count reconciliation |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| >400-line review budget | High | Split into 2-slice chained PR (A: resolver+CLI+tests; B: leader wiring + count reconcile). User prefers splitting over raising the budget. |
| skill_count drift (config 24 / CLAUDE.md 25 / 26 dirs) — pre-existing, not caused here | Med | Reconcile via `archon update` + prose edits + `skill-registry`; slice B owns it. |
| Residual #15 "revisa y prueba" dual-action | Low | Accepted limitation; harness-workflow + human-review gate backstop execution. |
| Leader doesn't call `archon route` (prompt-enforced) | Med | Baked into both orchestrator rule variants; ASK output feeds vague-request guard. |

## Rollback Plan

Revert per slice: slice B (templates/skill/config) is prose+embed, revertable with no
runtime dependency; slice A removes `newRouteCmd` + `internal/route/`. State ownership
never moves, so no data migration to unwind.

## Dependencies

- `internal/config` (`PhaseOrder`, `ValidPhases`) and the `mapgen readState` pattern.

## Success Criteria

- [ ] `archon route` resolves fixtures #1–#4 (implicit) and #5/#17/#18 (regressions) correctly.
- [ ] Router READs state only; harness-workflow remains sole writer.
- [ ] `skill_count` consistent across config, CLAUDE.md, and templates.
- [ ] Instruction + SKILL.md provider-neutral (no Claude-only assumptions).
