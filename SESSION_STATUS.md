# Session Status

- **Session started**: 2026-08-01
- **Last updated**: 2026-08-01 (apply — start, Slice A)
- **Active change**: local-model-router
- **Current phase**: apply (in_progress) — Slice A: internal/route + archon route CLI + tests
- **Next recommended**: apply Slice B, then verify

## Preflight
- Execution mode: interactive
- Artifact store: openspec
- Chained PR strategy: ask-always (single-pr unless estimate > budget)
- Review budget: 400 lines
- Playwright: no
- Impeccable: no

## Goal
Formalize a HYBRID phase router that makes the SDD flow friendly for weak local
models (e.g. ollama qwen3-orch:latest ~5B) which need explicit execution orders.
Empirically validated in `prototype/sdd-router/` (17/18 on the local model).

Architecture decided: deterministic CODE pre-router (control words + start verbs +
literal `archon-<phase>`, resolved from state.yaml + PhaseOrder) → MODEL classifier
(fuzzy "which phase family") → harness-workflow gate → archon-<phase> subagent.

## Phase History
- [x] explore — completed 2026-08-01
- [x] propose — completed 2026-08-01
- [x] spec — completed 2026-08-01
- [x] design — completed 2026-08-01
- [x] tasks — completed 2026-08-01
- [ ] apply — in_progress 2026-08-01 (Slice A)
- [ ] verify
- [ ] judge
- [ ] archive

## Key Artifacts
- openspec/changes/local-model-router/exploration.md — integration map
- openspec/changes/local-model-router/proposal.md — approved proposal
- openspec/changes/local-model-router/specs/local-model-router/spec.md — capability spec
- openspec/changes/local-model-router/specs/local-model-router/local-model-router.feature — Gherkin feature file (18+ scenarios)
- openspec/changes/local-model-router/design.md — technical design (package API, single-source verb table, 2-slice forecast)
- openspec/changes/local-model-router/tasks.md — ordered task breakdown (8 phases, 28 tasks)
- openspec/changes/local-model-router/state.yaml — phase state
- prototype/sdd-router/ROUTER.md — deterministic rule spec (read-only reference)
- prototype/sdd-router/fixtures.md — 18 test cases
- prototype/sdd-router/FINDINGS.md — empirical results + architecture

## Decisions Made in Spec
- D1 (Handoff): --json primary (machine-readable), stderr human echo line.
- D2 (Active-change): router discovers (flag > SESSION_STATUS.md > sole folder > none).
- D3 (#15 dual-action): narrow code rule — judge-verb AND verify-verb + conjunction → ASK.

## Chained-PR Strategy (Feature Branch Chain)
Converged from Stacked-to-Main to Feature Branch Chain (archive-before-PR in effect).

Branch chain:
  main
   └── feat/local-model-router        (tracker)
        └── feat/lmr-slice-a          (PR A: internal/route/ + CLI + tests, ~330 lines)
             └── feat/lmr-slice-b     (PR B: sdd-router SKILL.md + templates + golden + skill_count, ~180 lines)

Test-split commit pre-planned: if slice A diff exceeds 400 lines at apply time,
isolate fixture rows 9–18 into a follow-up commit before opening PR A.

## Open Questions
- None. Confirm test-split at apply if slice A exceeds 400 lines at diff time.

## Resume Hint
Tasks complete. Human Review Gate required before sdd-apply. Next: show tasks.md to
user, ask "¿Quieres ajustar algo en esta fase antes de continuar?"
