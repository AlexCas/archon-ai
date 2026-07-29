# SESSION STATUS

## Active Change
`local-model-provider` — provider local (openai-compatible) para apuntar fases SDD a Ollama y LocalAI.

## Current Phase
PR-A: DONE — apply→verify→judge APPROVED, committed as 73808c9 on branch feat/local-model-provider (not pushed).
PR-B: DONE — apply→verify→judge APPROVED, committed as dde17a8 on branch feat/local-model-provider-tui-guard (stacked on PR-A). Not pushed.
Phase: `archive` in_progress (closing the SDD cycle).

## Pending after archive (needs user)
- Push both branches + open chained PRs (PR-A→master, PR-B→PR-A). Requires switching to the AlexCas gh account first (agent cannot).

## Follow-up issues (non-blocking, deferred)
- `config list` shows "(none configured)" when only `models.default.base_url` is set without a model (pre-existing REQ-2 edge case).
- Leader BaseURL not CLI-settable (REQ-2 usability gap).

## Budget Decision (PR-A)
PR-A = 896 changed lines vs 400 budget. Overage is test coverage (~617 test lines; code only ~277). Split can't bring all slices under 400 (opencode-emission alone = 478). User APPROVED **size:exception** for a single cohesive PR-A on 2026-07-28.

## Preflight Choices
- A. Ritmo: Interactivo
- B. Artefactos: OpenSpec
- C. PRs: Preguntarme (presupuesto de revisión: 400 líneas)
- D. Revisión: 400 líneas
- E. Playwright: No
- F. Impeccable: No

## Completed Phases
- explore — completed 2026-07-29T00:37:42Z (artifact: openspec/changes/local-model-provider/exploration.md)
- propose — completed 2026-07-28 (artifact: openspec/changes/local-model-provider/proposal.md)
- spec    — completed 2026-07-28 (artifacts: openspec/changes/local-model-provider/specs/local-model-provider/spec.md + local-model-provider.feature)
- design  — completed 2026-07-29T01:02:41Z (artifact: openspec/changes/local-model-provider/design.md)
- tasks   — completed 2026-07-28 (artifact: openspec/changes/local-model-provider/tasks.md)

## Key Artifacts / Paths
- openspec/changes/local-model-provider/exploration.md
- openspec/changes/local-model-provider/proposal.md
- openspec/changes/local-model-provider/specs/local-model-provider/spec.md
- openspec/changes/local-model-provider/specs/local-model-provider/local-model-provider.feature
- openspec/changes/local-model-provider/tasks.md
- openspec/changes/local-model-provider/state.yaml — phase=tasks, status=completed

## Resolved Decisions (locked)
- OpenCode V1 schema: `provider.<id>.npm = @ai-sdk/openai-compatible`, `options.baseURL`, `models` map; `apiKey` OMITTED keyless servers.
- Config surface: per-ref `ModelRef.BaseURL` (Approach 1); `models.providers` block deferred.
- Provider naming: user-named ids; ref `Provider` IS the OpenCode provider id.
- Claude path: warn-and-skip (no silent drop, no hard reject).
- API key: omitted for keyless local servers.
- Validation: advisory (non-empty id, http(s) BaseURL parse).
- Chained-PR split APPROVED: PR-A (REQ-1–REQ-5) + PR-B (REQ-6–REQ-7).

## PR Split Summary
- **PR-A** (REQ-1–5): ModelRef.BaseURL YAML, CLI set/get/list, advisory validation, opencode provider-block emission + coalescing + idempotency.
- **PR-B** (REQ-6–7): Claude path warn-and-skip guard, TUI BaseURL sub-mode + render.

## Open Questions
- None blocking. OpenCode V2 schema divergence is a documented risk; no action until archon migrates.

## Next Recommended Step
Apply PR-B (REQ-6–7) on feat/local-model-provider-tui-guard, then verify + judge. On judge pass, archive.

_Last updated: 2026-07-28 — phase apply (PR-B start)_
