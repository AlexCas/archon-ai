# Session Status

- **Active change**: `archive-before-pr`
- **Source**: issue #93 — feat(workflow): run archive before opening the PR in the SDD cycle
- **Current phase**: archive — PENDING (2026-07-30); judge PASS on retry 1 (e474e30); archive runs BEFORE the PR (dogfooding this change)
- **Judge retry**: 1 of 3. BLOCKING-1 resolution = align to actual order "merge → move → map → SESSION_STATUS move"
- **Branch**: feat/archive-before-pr (from origin/master); single-PR flow

## Preflight decisions
- A. Ritmo: interactive
- B. Artefactos: openspec
- C. PRs: ask-always
- D. Review budget: 400 lines
- E. Playwright: disabled
- F. Impeccable: disabled

## Completed phases
- explore — 2026-07-30 (findings: archive is terminal + judge-gated; PR creation lives outside the state machine; proven post-merge in local-model-provider cycle)
- propose — 2026-07-30 (proposal.md written; rule: judge pass → archive as one commit → open PR; single-PR slice 1; chained-PR ownership deferred to slice 2)

- spec — 2026-07-30 (delta spec + Gherkin feature; 1 requirement "Terminal Phase Ordering (Single-PR Flow)", 6 scenarios; state.yaml at spec/completed)
- design — 2026-07-30 (design.md written; per-file edit plan for 10 artifacts; edit order spec→phase-order copies→sequencing skills→shared modules; verification is doc-consistency based; state.yaml at design/completed)
- tasks — 2026-07-30 (tasks.md written; 4 work units, ~200-280 line delta within 400 budget; state.yaml at tasks/completed)
- apply — 2026-07-30 (3 commits on feat/archive-before-pr: c7a7fe1 phase-order sources, 3f7d843 sequencing skills, 37afa0a shared modules; 10 files, +127/-16=143 lines; no Go touched; STOP checkpoint not triggered)
- verify — 2026-07-30 (PASS 7/7: 10 files/0 Go, 3-copy consistency, 6/6 scenarios, no orphaned wording, judge gate intact, non-goal documented, go build OK)
- judge — 2026-07-30 (FAIL, retry 1: BLOCKING-1 3b/3c order contradiction, BLOCKING-2 missing git commit step in sdd-archive, BLOCKING-3 Scenario 5 missing SESSION_STATUS ordering assertion)
- apply (judge-fix retry 1) — 2026-07-30 (commit e474e30: fixed all 3 blocking + NB-2; order canonicalized to merge→move→map→SESSION_STATUS; Step 3d commit added; ~56 real lines)
- re-judge — 2026-07-30 (APPROVED: 3/3 blocking resolved, 0 confirmed issues, gates intact, go build clean, no self-hosting paradox)

## Key artifacts (added)
- openspec/changes/archive-before-pr/proposal.md
- openspec/changes/archive-before-pr/specs/harness-workflow/spec.md
- openspec/changes/archive-before-pr/specs/harness-workflow/harness-workflow.feature
- openspec/changes/archive-before-pr/design.md
- openspec/changes/archive-before-pr/tasks.md
- openspec/changes/archive-before-pr/state.yaml

## Key artifacts / paths
- Issue: #93
- Change dir (planned): openspec/changes/archive-before-pr/
- Relevant: harness-workflow skill, session-status-contract, CLAUDE.md phase order

## Open questions
- Which PR in a chained-PR flow owns the archive move (deferred to slice 2 by approved scope).
- No code-level implication found; if apply hits a code path assuming archive-after-PR, STOP and flag (slice is docs-only).

## Next recommended step
- Human Review Gate on design.md, then tasks (sdd-tasks).
