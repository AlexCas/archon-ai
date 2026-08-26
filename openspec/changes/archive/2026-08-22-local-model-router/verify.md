# Verification Report — local-model-router

- **Change**: local-model-router
- **Branch (integrated)**: `feat/lmr-slice-b` (linearly contains A1 + A2 + B)
- **Mode**: openspec (Feature Branch Chain, 400-line budget, no Playwright, no Impeccable)
- **Verdict**: **PASS WITH WARNINGS**

The integrated branch builds, vets, and passes the full test suite; all 18 fixtures
conform end-to-end; the read-only, gate-integrity, and skill_count invariants hold; and
every one of the 13 spec requirements maps to a passing test or fixture. The only
warnings are review-budget observations (Slice A code diff exceeds 400 lines) — no
functional defects were found.

## Completeness

| Dimension | Result |
|-----------|--------|
| Tasks checked complete | 8 phases / all task rows `[x]` (Slice A 1.1–4.4, Slice B 5.1–8.5) |
| Unchecked implementation tasks | none |
| Artifacts present | proposal, spec + `.feature`, design, tasks — full spec-driven verify |

## Build / Test / Static Analysis

| Command | Result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l` (8 touched files) | clean (no files listed) |
| `go test ./...` | exit 0 — all packages `ok` (route, initcmd, cmd/archon, config, + rest) |

## Fixture Conformance — 18/18 PASS

Binary built from the branch (`go build ./cmd/archon`); each fixture invoked with
deterministic `--change` / `--phase` / `--status` overrides. Every row matched the
expected phase + rule + path from `prototype/sdd-router/fixtures.md`, all exit 0.

| #  | Message                                                  | State            | Expected (phase / rule)   | Got | Pass |
|----|----------------------------------------------------------|------------------|---------------------------|-----|------|
| 1  | Trabajemos en esta especificacion                        | none             | explore / implicit-start  | ✓   | ✅ |
| 2  | Trabajemos en esta especificacion                        | spec/in_progress | spec / implicit-resume    | ✓   | ✅ |
| 3  | Empecemos con esto                                       | none             | explore / implicit-start  | ✓   | ✅ |
| 4  | Hagamos esta feature                                     | none             | explore / implicit-start  | ✓   | ✅ |
| 5  | Hagamos esta especificacion. Lanza el agente de exploracion | none          | explore / explicit-agent  | ✓   | ✅ |
| 6  | Continuemos                                              | propose/completed| spec / next               | ✓   | ✅ |
| 7  | Siguiente                                                | apply/completed  | verify / next             | ✓   | ✅ |
| 8  | Continuemos                                              | design/in_progress | design / resume         | ✓   | ✅ |
| 9  | Adelante                                                 | none             | explore / next-nochange   | ✓   | ✅ |
| 10 | Explora el codigo de billing                             | none             | explore / keyword         | ✓   | ✅ |
| 11 | Disenemos la arquitectura del API                        | spec/completed   | design / keyword          | ✓   | ✅ |
| 12 | Implementa las tareas                                    | tasks/completed  | apply / keyword           | ✓   | ✅ |
| 13 | Corre las pruebas                                        | apply/completed  | verify / keyword          | ✓   | ✅ |
| 14 | Archiva el cambio                                        | judge/completed  | archive / keyword         | ✓   | ✅ |
| 15 | Revisa y prueba esto                                     | verify/in_progress | ASK / ambiguous (code)  | ✓   | ✅ |
| 16 | Que opinas del clima?                                    | spec/in_progress | CLASSIFY / classify (model) | ✓ | ✅ |
| 17 | Volvamos al spec                                         | design/completed | spec / explicit-agent     | ✓   | ✅ |
| 18 | corre el apply                                           | tasks/completed  | apply / explicit-agent    | ✓   | ✅ |

Special-attention rows confirmed: #1–#4 implicit → explore/resume; #5 explicit-agent
wins over co-present implicit verb; #12 "implementa las tareas" → apply (not tasks,
not ambiguous); #15 dual-action → ASK on the CODE path (not model); #16 → CLASSIFY;
#17 backward resolves to spec (gate blocks separately); #18 literal token → apply.

## Read-Only Invariant — PASS

With a real `state.yaml` (spec/in_progress) and `SESSION_STATUS.md` naming an active
change, five `archon route` invocations (control, implicit, dual-action, CLASSIFY,
next-nochange) left both files byte-identical (SHA-256 unchanged before/after). Genuine
read confirmed: the router resolved `active_change` from `SESSION_STATUS.md` and phase
from `state.yaml` (→ `{"phase":"spec","rule":"resume",...,"active_change":"mychange"}`).

## Gate-Integrity — PASS

`internal/initcmd/templates.go`, both `orchestratorRulesClaude` and
`orchestratorRulesOpencode`:
- Rule 1 = harness-workflow gate (unchanged, first).
- Rule 2 = the new `archon route` step — inserted AFTER the gate, BEFORE the delegate rule.
- Rule 3 = delegate (renumbered, wording intact).
- Preflight (HARD GATE), Vague Request Guard, and Human Review Gate are separate `##`
  sections above the Rules block — untouched, not weakened, not reordered.
- The router rule only decides WHICH phase; it does not bypass any gate.
- Both variants in sync (differ only in the pre-existing rule-3 model-gate wording:
  "frontmatter model" vs "opencode.json" — predates this change).
- Golden test `TestOrchestratorRules…RoutingOrder` asserts route precedes delegate AND
  follows the harness-workflow gate, in both rendered docs — passing.

## skill_count Consistency — PASS (value = 27)

| Source | Value |
|--------|-------|
| `.archon/config.yaml` `skill_count` | 27 |
| Embedded FS (`go:embed */SKILL.md` → `len(res.Extracted)`) | 27 SKILL.md files |
| Rendered `CLAUDE.md` "Skills:" line (fresh `archon init`) | 27 |

`sdd-router/SKILL.md` is present in the embedded set and auto-counted by the glob (no
manual registration). Note: the count includes `_shared/SKILL.md` — pre-existing
convention, not introduced here.

## Spec Coverage — 13/13 requirements verified

| Requirement | Verifying test / fixture |
|-------------|--------------------------|
| Deterministic Code Pre-router | `TestResolve` fixtures 5,6,8,11 + fixture run |
| Implicit-above-keyword Precedence | `TestImplicitAboveKeyword`; fixtures 1,11 |
| Narrow Dual-Action Ambiguity (D3) | `TestResolve` fixture 15; live fixture #15 |
| Model Classifier Fallback | `TestResolve` fixture 16 (CLASSIFY code-path); live #16 |
| Backward Nav Resolves, No Bypass | `TestResolve` fixture 17 (router resolves spec); gate blocks in harness-workflow (see below) |
| Bootstrap State — No Active Change | fixtures 1,3,4,9 (`implicit-start`/`next-nochange` → explore) |
| Machine-Readable JSON Output Contract | live fixture runs — valid JSON, all 4 fields, exit 0 incl. ASK/CLASSIFY |
| Text Normalization | `TestNormalize`; fixture 1 (accented "especificación") |
| State is Read-Only for the Router | read-only invariant check (SHA-256 unchanged) |
| PhaseOrder Canonical Source | `nextPhase` uses `config.PhaseOrder`; no 2nd list in `internal/route` (grep-confirmed); fixtures 6,7 |
| Active-Change Discovery Precedence | `TestActiveChange` (flag > SESSION_STATUS > sole-folder > none) |
| Keyword-to-Phase Table Coverage | `TestResolveKeywordOutline` (all 9 phases); `keywordTable` has ≥2 ES + 2 EN each |
| sdd-router Skill — Model Classifier Contract | `skills/sdd-router/SKILL.md` (invocation-on-CLASSIFY, 9-phase table, one-echo-line + stop, provider-neutral) |
| Leader Wiring (orchestratorRules) | `TestOrchestratorRules…RoutingOrder` (both variants) |

No requirement is left without verification.

**Note on the "Backward navigation … blocked by harness-workflow" scenario**: the
router half (resolves to spec via explicit-agent) is verified by fixture #17 and the
live run. The harness-workflow blocking half is that skill's own behavior (backward
transitions not allowed) and is out of scope for `internal/route` tests — no defect,
correct division of labor per spec.

## Line Budget (code-only, excludes openspec/ + prototype/ planning docs)

| Slice | Code changed lines | vs 400 budget |
|-------|--------------------|---------------|
| A (route pkg + `archon route` CLI + tests) | 771 insertions (0 del) | **OVER** |
| B (SKILL.md + templates wiring + golden + skill_count) | 154 ins / 33 del (net 121) | under |

Slice A code exceeds 400. The tasks.md forecast pre-planned a fixture test-split commit
for exactly this case; on the integrated branch the two Slice-A commits
(`e806f04` resolver+tests 488+, `fa2197c` CLI 303+) are already split, but the resolver
commit alone (488) is over budget. This is a REVIEW-workload observation for the judge /
PR-split decision, not a functional defect.

## Issues

### CRITICAL
- none

### WARNING
- **W1 (review budget)**: Slice A code diff (771 lines; resolver commit 488) exceeds the
  400-line review budget. The forecast anticipated this and pre-planned a test-split; the
  judge / PR flow should confirm the Slice-A PR is reviewable within budget or split
  further. Not a code defect.

### SUGGESTION
- **S1**: `skill_count` includes `_shared/SKILL.md`. Consistent across all three sources
  so not a drift bug, but if `_shared` is conceptually a shared module rather than a
  user-facing skill, consider excluding it from the count in a future cleanup. Out of
  scope here.

## Final Verdict

**PASS WITH WARNINGS** — build/vet/gofmt/tests all green, 18/18 fixtures conform,
read-only + gate-integrity + skill_count invariants hold, 13/13 spec requirements
verified. The sole warning (W1) is a review-budget observation already anticipated by
the plan, with no functional defect.

**Next recommended**: proceed to `harness-judge` (integrated judge on the tracker for
the Feature Branch Chain). Carry W1 forward to the PR-split / review decision.
