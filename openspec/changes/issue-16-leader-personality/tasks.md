# Tasks: Leader Persona Template Section

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~75 (0 del + 75 add) |
| 400-line budget risk | **Low** |
| Chained PRs recommended | **No** |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Template Changes

- [ ] 1.1 Insert `## Leader Persona` block (scope, Language, Tone, Behavior) into `orchestratorSections` at line 16 of `internal/initcmd/templates.go` — between `# ARCHON AI Orchestrator\n` and `\n## Phase Order`
- [ ] 1.2 Verify persona has zero literal backticks (no `§` escaping needed) and static markdown with no template interpolation

## Phase 2: Testing

- [ ] 2.1 Add `TestTemplates_LeaderPersona` to `internal/initcmd/templates_test.go` — render both `RenderAgentsMD` and `RenderClaudeMD`, assert `## Leader Persona` header present
- [ ] 2.2 Assert ordering: `## Leader Persona` index < `## Phase Order` index in both templates (guard against -1 — check header exists first)
- [ ] 2.3 Assert all four domains present: scope (`"governs ONLY your chat replies"`), language (`"ALWAYS reply in the user's current language"`), tone (`"Warm and direct"`), behavior (`"Never say"`)
- [ ] 2.4 Run full test suite: `go test ./internal/initcmd/...` — all existing tests must pass unchanged

## Phase 3: Regeneration & Verification (if applicable)

- [ ] 3.1 Run `go build ./...` to confirm compilation
- [ ] 3.2 (Optional) Run `go run . init` (or equivalent) to regenerate `CLAUDE.md` and verify persona appears before `## Phase Order` in rendered output
- [ ] 3.3 (Optional) Verify `CLAUDE.md` diff shows persona block inserted at correct position without corrupting existing sections
