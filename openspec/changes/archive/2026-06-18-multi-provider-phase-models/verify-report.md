# Verification Report: multi-provider-phase-models

- **Change**: multi-provider-phase-models
- **Artifact store**: openspec
- **Mode**: Standard verify (config `testing.strict_tdd: false`; apply reported `strict_tdd: false`)
- **Project type**: non-web; Playwright disabled — no Playwright checks
- **Artifacts inspected**: proposal.md, design.md, tasks.md, specs/harness-init/spec.md, specs/harness-init/harness-init.feature
- **Verdict**: **PASS WITH WARNINGS**

---

## 1. Task Completeness

All 13 tasks are checked `[x]` in tasks.md and verified against the implementation.

| Task | Description | Checked | Implemented | Evidence |
|------|-------------|---------|-------------|----------|
| 1.1 | Add `GeminiModels` catalog + doc comment | [x] | Yes | `model.go:23-30` |
| 1.2 | Add `OpenAIModels` catalog + doc comment | [x] | Yes | `model.go:32-41` |
| 1.3 | Fold both catalogs into `StaticModels()` (Claude→Gemini→OpenAI→Opencode); `KnownModels` inherits | [x] | Yes | `model.go:61-79` |
| 2.1 | Add `providerFamily` struct + ordered `providerFamilies` table | [x] | Yes | `model.go:107-121` |
| 2.2 | Rewrite `NormalizeModel` to walk table (Claude whole-token alias; non-Claude exact id; first match wins) | [x] | Yes | `model.go:147-172` |
| 2.3 | Update `NormalizeModel` doc comment for four-provider precedence | [x] | Yes | `model.go:129-146` |
| 3.1 | Invert `glm-5`/`kimi-k2.5` to `ok=true`; change `gpt-4` row to `gpt-4o`→`gpt-4o` | [x] | Yes | `model_test.go:56-58` |
| 3.2 | Add Gemini row (`gemini-2.5-pro`→itself) | [x] | Yes | `model_test.go:60` |
| 3.3 | Keep `octopus`, `supushaiku`, `Opues 4.8` as `ok=false` | [x] | Yes | `model_test.go:55,61,62` |
| 3.4 | Add precedence/collision row → Claude alias | [x] | Yes | `model_test.go:68` (`opus gpt-4o`→`opus`) |
| 3.5 | `TestValidate`: curated non-Claude id `wantWarn:false`; `Opues 4.8` `wantWarn:true` | [x] | Yes | `model_test.go:18,21` |
| 4.1 | Render assertion: non-Claude default emits `## Phase Models` + catalog id | [x] | Yes | `templates_test.go:318-337` |
| 5.1 / 5.2 | Run tests + build | [x] | Yes | See section 2 |

No unchecked tasks. **0 CRITICAL** from task completeness.

---

## 2. Build / Test / Vet Evidence (real execution)

### `go test -count=1 ./internal/config/... ./internal/initcmd/...`
Ran with `-v` and `-count=1` (cache bypassed). Result:

```
ok  github.com/archon-ai/archon/internal/config   0.003s   CONFIG_EXIT=0
ok  github.com/archon-ai/archon/internal/initcmd  0.002s   INITCMD_EXIT=0
```

`TestNormalizeModel` — all 24 subtests PASS. `TestValidate` — all 10 subtests PASS.
`TestTemplates_PhaseModelsBlock`, `TestTemplates_PhaseModelsNonClaudeDefault`,
`TestTemplates_PhaseModelsOmittedWhenEmpty`, `TestTemplates_PhaseModelsBlockMatchesAcrossPaths` — all PASS.

### `go build ./...`
```
BUILD_EXIT=0   (no output)
```

### `go vet ./internal/config/... ./internal/initcmd/...`
```
VET_EXIT=0   (no output)
```

All three commands exited 0. No CRITICAL from execution.

---

## 3. Spec Compliance Matrix (8 Gherkin scenarios)

Source of truth: `specs/harness-init/harness-init.feature`. A scenario is compliant only when a covering test PASSED at runtime (all listed tests passed — see section 2).

| # | Scenario | Covering test(s) | Result | Status |
|---|----------|------------------|--------|--------|
| 1 | Display name is normalized to an accepted identifier | `TestNormalizeModel/display_name_with_version` (`Opus 4.8`→`opus`); `TestTemplates_PhaseModelsBlock` (asserts `- explore: opus`, and no raw `Opus 4.8`) | PASS | COMPLIANT |
| 2 | Gemini model normalizes to its catalog id | `TestNormalizeModel/gemini_pro` (`gemini-2.5-pro`→itself); `TestValidate/known_gemini_model` | PASS | COMPLIANT |
| 3 | OpenAI model normalizes to its catalog id | `TestNormalizeModel/openai_gpt-4o` + `openai_gpt-4o_uppercase`; `TestValidate/known_openai_model` | PASS | COMPLIANT |
| 4 | Opencode model normalizes to its catalog id | `TestNormalizeModel/opencode_glm` (`glm-5`), `opencode_kimi` (`kimi-k2.5`); `TestValidate/known_opencode_model` | PASS | COMPLIANT |
| 5 | Whole-token guard rejects a containing substring (octopus) | `TestNormalizeModel/substring_not_whole_token` (`octopus`→`ok=false`); `family_embedded_in_word_rejected` (`supushaiku`) | PASS | COMPLIANT |
| 6 | Colliding value resolves by fixed precedence | `TestNormalizeModel/claude_row_wins_over_later_providers` (`opus gpt-4o`→`opus`) | PASS | COMPLIANT (see WARNING-1) |
| 7 | Non-Claude default renders an identical block across paths | `TestTemplates_PhaseModelsNonClaudeDefault` (non-empty block + `gemini-2.5-pro`); `TestTemplates_PhaseModelsBlockMatchesAcrossPaths` (init vs TUI byte-identical) | PASS | COMPLIANT |
| 8 | Unresolvable typo is omitted but not rejected | `TestNormalizeModel/typo_no_family` (`Opues 4.8`→`ok=false`); `TestValidate/typo_display_name` (`wantWarn:true`); `TestResolvePhaseModels/unresolvable_phase_value_falls_through_to_default` | PASS | COMPLIANT |

All 8 scenarios map to at least one covering test that passed at runtime.

---

## 4. Design Coherence

| Design decision | Expected | Observed | Status |
|-----------------|----------|----------|--------|
| One ordered `providerFamilies` table, fixed order Claude→Gemini→OpenAI→Opencode | Single table, deterministic by order | `model.go:116-121` — exactly Claude→Gemini→OpenAI→Opencode | MATCH |
| `providerFamily` struct shape (`families []string`, `catalog []string`) | As specified in design interfaces block | `model.go:107-110` matches verbatim | MATCH |
| Canonical output: Claude→alias; non-Claude→catalog id as-is | opus/sonnet/haiku for Claude; id verbatim otherwise | `model.go:159` (`return fam`) and `:166` (`return id`) | MATCH |
| Matching semantics: Claude whole-token via `FieldsFunc`; non-Claude exact case-insensitive id | Octopus-safe tokenizer retained; exact id check | `model.go:152-169` retains `FieldsFunc` tokenizer; catalog uses `s == strings.ToLower(id)` | MATCH |
| Signature unchanged | `func NormalizeModel(s string) (id string, ok bool)` | Unchanged (HEAD `:97` → working tree `:147`, same signature) | MATCH |
| `templates.go`, `init.go`, `tui/model.go`, `models_tab.go` untouched | No edits | Working-tree diff touches ONLY `model.go`, `model_test.go`, `templates_test.go` | MATCH |
| Diff size within forecast (~120–170 / 400 budget) | Low risk | `git diff --stat`: 109 insertions, 21 deletions across 3 files (~130 changed lines) | MATCH |
| `ResolvePhaseModels`/`Validate`/`StaticModels`/`KnownModels` call through unchanged, inherit multi-provider | No logic change outside `NormalizeModel`/catalogs | `Validate` and `ResolvePhaseModels` bodies unchanged; `StaticModels` extended only to append new catalogs | MATCH |

No design deviations.

---

## 5. Issues

### CRITICAL
None.

### WARNING

- **WARNING-1 — Collision scenario (#6) coverage is weaker than the literal Gherkin intent.**
  The Gherkin reads "a value that matches **both Claude and a later provider**." The
  covering test uses `"opus gpt-4o"`, which is a multi-token string carrying a Claude
  family token AND a separate OpenAI catalog token; it proves the Claude row is consulted
  first. As design.md / apply noted, **no real catalog id contains a Claude family
  substring**, so a single token that genuinely belongs to two providers' catalogs cannot
  exist with the current data — making a "true" double-membership collision unconstructable.
  Judgment: the test **adequately demonstrates the fixed-precedence guarantee** the scenario
  exists to protect (earliest matching row wins), and the precedence is also structurally
  guaranteed by the ordered table walk in `NormalizeModel`. This is acceptable coverage,
  but it is a *proxy* for the literal scenario rather than a direct double-catalog-membership
  case. Flagging as WARNING for reviewer awareness, not as a blocker.

### SUGGESTION

- **SUGGESTION-1 — Make the collision test intent explicit.** Consider adding a short test
  comment (or a second precedence row) clarifying that genuine single-token cross-catalog
  collisions are impossible by construction, so `"opus gpt-4o"` is the strongest realizable
  case. The existing inline comment at `model_test.go:65-67` is close; extending it to state
  "no single catalog id collides across providers today" would preempt the WARNING above.

- **SUGGESTION-2 — Scenario #7 byte-identity test uses identical inputs for both paths.**
  `TestTemplates_PhaseModelsBlockMatchesAcrossPaths` constructs `initData` and `tuiData` from
  the same `ResolvePhaseModels(mc)` call shape, so it proves render determinism but does not
  exercise genuinely different code paths (`init.writeTemplate` vs the TUI `regenerateTemplate`).
  The test comment acknowledges it "mirrors" each path's data construction. Coverage is
  adequate for the spec (both produce a non-empty, byte-identical block), but a future
  hardening could invoke the actual init and TUI builders.

---

## 6. Final Verdict

**PASS WITH WARNINGS**

All 13 tasks implemented and checked. `go test` (config + initcmd), `go build ./...`, and
`go vet` all exit 0 with real captured output. All 8 Gherkin scenarios map to covering tests
that passed at runtime. Implementation matches design.md on every dimension (ordered table,
canonical-output split, unchanged signature, render/init/TUI untouched, diff size within
forecast). The two warnings concern coverage *strength* of the collision and cross-path
scenarios, not correctness — both are adequately, if indirectly, covered. No CRITICAL issues;
the change is archive-ready pending the orchestrator's review of the warnings.
