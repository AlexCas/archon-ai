# Tasks: Multi-provider per-phase models (Slice 1)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~120–170 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Multi-provider normalization + catalogs + all tests | PR 1 | Base = master; one function + 2 vars + test edits; tests included |

## Phase 1: Foundation (catalogs)

- [x] 1.1 In `internal/config/model.go`, add `GeminiModels = ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"]` with a doc comment.
- [x] 1.2 In `internal/config/model.go`, add `OpenAIModels = ["gpt-4o", "gpt-4o-mini", "gpt-4.1", "o3", "o4-mini"]` with a doc comment.
- [x] 1.3 Fold both catalogs into `StaticModels()` (order: Claude → Gemini → OpenAI → Opencode); `KnownModels` inherits via `StaticModels()`.

## Phase 2: Core Implementation

- [x] 2.1 In `internal/config/model.go`, add `providerFamily{ families []string; catalog []string }` and ordered `providerFamilies` table (Claude→Gemini→OpenAI→Opencode) per design.
- [x] 2.2 Rewrite `NormalizeModel` body to walk `providerFamilies`: Claude rows = whole-token family match emitting the alias (keep `FieldsFunc` tokenizer, octopus-safe, opus→sonnet→haiku priority); non-Claude rows = exact case-insensitive id match emitting the id as-is; first match wins.
- [x] 2.3 Update the `NormalizeModel` doc comment to describe the four-provider precedence and per-provider canonical output.

## Phase 3: Testing (unit — `internal/config/model_test.go`)

- [x] 3.1 Invert the `glm-5` and `kimi-k2.5` rows to `ok=true` returning the id as-is; change the `gpt-4` row to `gpt-4o` → `gpt-4o`, `ok=true` (scenario: Opencode/OpenAI catalog id).
- [x] 3.2 Add a Gemini table row (e.g. `gemini-2.5-pro` → itself) (scenario: "Gemini model normalizes to its catalog id").
- [x] 3.3 Keep `octopus` and `supushaiku` as `ok=false` (scenario: "Whole-token guard rejects a containing substring"); keep `Opues 4.8` → `ok=false` (scenario: "Unresolvable typo is omitted but not rejected").
- [x] 3.4 Add a precedence/collision row: a value matching Claude + a later catalog resolves to the Claude alias (scenario: "Colliding value resolves by fixed precedence").
- [x] 3.5 In `TestValidate`, add a curated non-Claude id (e.g. `gpt-4o`) with `wantWarn:false`; keep `Opues 4.8` `wantWarn:true` (scenario: typo accepted with advisory warning, not rejected).

## Phase 4: Render assertion (`internal/initcmd/templates_test.go`)

- [x] 4.1 Add a test rendering `RenderClaudeMD` with `PhaseModels: ResolvePhaseModels(ModelConfig{Default: "gemini-2.5-pro"})`; assert a non-empty `## Phase Models` block and the catalog id appears (scenarios: "Opus 4.8" display-name + "Non-Claude default renders an identical block across paths").

## Phase 5: Verify

- [x] 5.1 Run `go test ./internal/config/... ./internal/initcmd/...` and confirm all pass.
- [x] 5.2 Run `go build ./...` to confirm no compile regressions.
