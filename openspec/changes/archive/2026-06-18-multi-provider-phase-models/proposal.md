# Proposal: Multi-provider per-phase models (Slice 1)

## Intent

`NormalizeModel` (`internal/config/model.go:97-113`) only matches Claude tokens
(`opus|sonnet|haiku`) and returns `ok=false` for every Gemini/OpenAI/Opencode value.
Since `ResolvePhaseModels` drops phases where both value and default fail to normalize,
a non-Claude-only project resolves to an empty slice and the `## Phase Models` block is
omitted. The Opencode delegation runtime DOES accept provider model IDs, so this is a
real gap. This change teaches normalization about the other providers.

## Scope

### In Scope
- Multi-provider `NormalizeModel` covering Claude + Gemini + OpenAI + Opencode.
- New curated static catalogs `GeminiModels`, `OpenAIModels`; fold into
  `StaticModels()` / `KnownModels`.
- Tests: invert the existing `ok=false` non-Claude rows, add per-provider tables,
  add a non-Claude render assertion proving the block emits.

### Out of Scope (deferred, not foreclosed)
- Slice 2: Opencode "archon-leader" mode in the opencode config + TUI leader-model field.
- Slice 3: dynamic detection of installed agents/models (catalogs stay static).

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `harness-init`: model normalization MUST recognize curated Gemini/OpenAI/Opencode
  IDs (advisory, never rejecting), so the `## Phase Models` block renders for any
  supported provider across both `archon init` and TUI regeneration.

## Approach

Provider-family table (exploration Approach 1). Generalize `claudeFamilies` into one
ordered `{provider, tokens, canonical}` table walked in fixed precedence
(**Claude → Gemini → OpenAI → Opencode**); first whole-token match wins, reusing the
existing octopus-safe tokenizer. **Per-provider canonical**: Claude keeps its short
alias (`opus`/`sonnet`/`haiku`); non-Claude providers emit the matched catalog ID as-is
(no comparable alias system). Render paths are already provider-agnostic — no edits there.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modified | Per-provider `NormalizeModel`; add `GeminiModels`/`OpenAIModels`; fold into `StaticModels()`/`KnownModels`. |
| `internal/config/model_test.go` | Modified | Invert non-Claude `ok=false` rows; add per-provider tables; keep typo/whole-token cases. |
| `internal/initcmd/templates_test.go` | Modified | Add render assertion: non-Claude config emits `## Phase Models`. |
| `internal/initcmd/templates.go`, `init.go`, `internal/tui/model.go`, `models_tab.go` | None | Provider-agnostic already; no edits expected. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Inverted expectations break prior tests | High | Intended; rewrite the three `ok=false` rows. |
| Catalog staleness vs real provider lists | Med | Keep lists small/current; free-form entry stays the escape hatch. |
| Canonical-ID not accepted by runtime | Low | User confirmed Opencode accepts provider IDs; form lives in one function, swappable. |
| Cross-provider collision | Low | Fixed documented precedence keeps it deterministic. |
| Test rewrites push change near/over 400-line budget | Med | Scope is one function + catalogs + tests; re-forecast at tasks. |

## Rollback Plan

Revert `model.go` (restore Claude-only `NormalizeModel`, drop the two new catalogs)
and the test edits. Generated `CLAUDE.md`/`AGENTS.md` returns to prior content on next
init/regenerate. No config or on-disk data is migrated, so nothing to undo in user
configs.

## Dependencies

- Confirmation of the curated `GeminiModels`/`OpenAIModels` contents and the
  canonical-ID form per provider — resolved during spec/design.

## Success Criteria

- [ ] `NormalizeModel` returns `ok=true` for curated Gemini/OpenAI/Opencode IDs;
      Claude still emits its short alias.
- [ ] A non-Claude-only config renders a non-empty `## Phase Models` block in both
      `archon init` and TUI regenerate, byte-identical across paths.
- [ ] Cross-provider precedence is fixed (Claude → Gemini → OpenAI → Opencode).
- [ ] `Validate` stays advisory (warn, never reject); free-form entry still works.
- [ ] `go test ./...` passes with inverted/extended tables.
