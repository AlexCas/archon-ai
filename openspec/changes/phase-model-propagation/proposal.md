# Proposal: Phase Model Propagation

## Intent

Per-phase model config (`models.phases.<phase>` in `.archon/config.yaml`) is read,
persisted, displayed, and edited — but never **consumed**. Nothing tells the
orchestrator which model to delegate each SDD phase with, so the setting silently
dead-ends. This closes the loop by injecting a phase→model map into the generated
orchestrator template (`CLAUDE.md`/`AGENTS.md`), so the orchestrator can honor the
configured model when delegating each phase.

## Scope

### In Scope
- Add a resolved phase→model field to `TemplateData` and render a normative
  `## Phase Models` block in the orchestrator template.
- Forward `cfg.Models` through BOTH render paths: `writeTemplate` (init.go) and
  `regenerateTemplate` (tui/model.go).
- **Normalize to IDs**: map display strings ("Opus 4.8") to identifiers the
  delegation tool accepts (aliases `opus`/`sonnet`/`haiku` or full IDs like
  `claude-opus-4-8`) before rendering. Approach proposed here; exact table is a
  design detail.
- **Fallback**: phase with no model → use `models.default`; no default either →
  omit that phase's line.
- Deterministic phase ordering in the rendered block (Go map iteration is random)
  for stable golden tests; cover in `templates_test.go`.

### Out of Scope
- **Vía 2 — Go-driven runtime delegation** (harness sets the model on the Task/Agent
  call). No such delegation layer exists today; deferred as future hardening.
- Editing the live `.archon/config.yaml` (developer test data; contains a `Opues 4.8`
  typo). Left untouched on purpose.
- Migrating/repairing existing on-disk configs.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `harness-init`: template generation MUST emit a phase→model block resolved from
  configured models (normalized to IDs, with default-fallback/omit behavior). This
  applies to both `archon init` and TUI-triggered regeneration.

## Approach

Vía 1 — **template injection**. Resolve `cfg.Models` (per-phase value → default →
omit), normalize each value to a real model ID via a small mapping helper (likely
`internal/config/model.go`), and render a `## Phase Models` block instructing the
orchestrator which model to request per delegated phase. Both render call sites pass
`cfg.Models` through. The block is **advisory**: it instructs the orchestrator LLM;
it is not a hard runtime gate, and its effect depends on the host platform supporting
per-delegation model selection.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/initcmd/templates.go` | Modified | Add field to `TemplateData`; add `## Phase Models` block to template. |
| `internal/initcmd/init.go:96,241` | Modified | `writeTemplate` accepts/forwards configured models. |
| `internal/tui/model.go:332-358` | Modified | `regenerateTemplate` forwards `cfg.Models`. |
| `internal/config/model.go` | Modified | Normalization helper (display→ID) + resolution/fallback. |
| `internal/initcmd/templates_test.go` | Modified | Cover new field, block, deterministic ordering. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Bad/typo model names reach the block (e.g. `Opues 4.8`) | Med | Strengthen validation for END USERS (warn vs reject — deferred to spec/design). |
| Advisory-only enforcement (orchestrator may ignore) | Med | State as known limitation; Vía 2 = future work. |
| Non-deterministic map ordering breaks golden tests | Med | Render phases in fixed canonical order. |
| Platform may not honor per-delegation model selection | Low | Block wording confirmed in design; degrade gracefully. |

## Rollback Plan

Revert the template field, the `## Phase Models` block, and the two render-site
signature changes. Generated `CLAUDE.md` returns to its prior content on next
init/regenerate. No config or on-disk data is migrated, so there is nothing to undo
in user configs.

## Dependencies

- Confirmation of how the host agent platform applies a per-phase model on delegation
  (informs block wording) — resolved in design.

## Size / Risk Forecast

- **Forecast**: ~120-200 changed lines (struct field, normalization helper, template
  block, 2 call sites, tests). **Likely UNDER the 400-line review budget.**
- **400-line budget risk**: Low.
- **Chained PRs**: not anticipated. Re-forecast at tasks.
- **Risk level**: Medium (advisory enforcement + normalization/validation decisions).

## Open Questions (deferred to spec/design)

1. **Exact normalization table** — which display strings map to which IDs, and
   whether to emit aliases (`opus`) or full IDs (`claude-opus-4-8`).
2. **Validate vs warn** for bad/unknown end-user model values (advisory warning vs
   rejection on init/save).
3. **Canonical phase ordering** to surface in the block (likely the phase-order
   sequence from CLAUDE.md).
4. **Block wording** for the case where the platform cannot honor per-delegation
   model selection.

## Success Criteria

- [ ] Generated `CLAUDE.md`/`AGENTS.md` contains a `## Phase Models` block listing
      each configured phase and a normalized model ID.
- [ ] Phases without a model fall back to `models.default`; with no default, the line
      is omitted.
- [ ] Block renders deterministically; `templates_test.go` passes.
- [ ] Both `archon init` and a TUI model edit regenerate the block consistently.
