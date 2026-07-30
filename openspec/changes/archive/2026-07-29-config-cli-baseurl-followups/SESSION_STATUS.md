# SESSION_STATUS

## Active change
`config-cli-baseurl-followups` — GitHub issues #90 + #91

## Current phase
judge — in progress

## PR strategy decision (apply gate)
Budget overage: apply diff = 603 lines (47 prod + 556 tests) vs 400 ceiling.
Decision: SPLIT into 2 chained PRs (matches review-budget-split-preference).
- PR-A (#90): HasAny() + guard swap + primary suppression + base_url rendering in config list & status + #90 tests.
- PR-B (#91, chained on A): models.leader set/get/list/baseURLRefForKey/error strings + status Leader block + leader tests.
Split happens at PR-open time (after verify + judge on the full working tree).

## Preflight decisions
- A. Ritmo: interactive
- B. Artefactos: openspec
- C. PRs: ask-always (budget 400)
- D. Review budget: 400 changed lines
- E. Playwright: disabled
- F. Impeccable: disabled

## Scope
- #90 (bug): `archon config list` shows "(none configured)" when only `base_url` is set and model id is empty; base_url value is hidden.
- #91 (feat): `models.leader.base_url` cannot be set/get/listed via the `archon config` CLI (only default and phases are wired).

## Completed phases
- explore — done
- propose — done (proposal.md + state.yaml; approved at gate)
- spec — done (spec.md + .feature file; 5 requirements, 24 Gherkin scenarios)
- design — done (design.md; HasAny() helper + inline render, leader routing; approved at gate)
- tasks — done (tasks.md; 9 phases / 26 tasks; approved at gate)
- apply — done (all 26 tasks; go build/vet/test all pass; working tree clean-building)
- verify — done (PASS; verify-report.md; all REQ-8..12 traced to code+tests; hunk→PR map confirmed)

## Key artifacts / paths
- `openspec/changes/config-cli-baseurl-followups/proposal.md`
- `openspec/changes/config-cli-baseurl-followups/specs/local-model-provider/spec.md`
- `openspec/changes/config-cli-baseurl-followups/specs/local-model-provider/config-cli-baseurl-followups.feature`
- `openspec/changes/config-cli-baseurl-followups/state.yaml`
- Reference prior change: `openspec/changes/archive/2026-07-28-local-model-provider/`

## Open questions
None — scope fully locked at propose gate.

## Next step
Human Review Gate for spec artifacts, then design phase.
