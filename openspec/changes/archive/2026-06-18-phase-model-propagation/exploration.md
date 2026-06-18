# Exploration: phase-model-propagation

## Project Type
**Web testing**: not-web

This is the ARCHON harness's own codebase: a Go CLI (`cmd/archon`) plus internal
packages (`internal/config`, `internal/initcmd`, `internal/tui`, `internal/status`).
No web framework, no `package.json`, no browser surface, no E2E tooling. Playwright
stays disabled. This matches the already-decided preflight (no Playwright).

## Current State

The harness lets users configure a model per SDD phase, but **that configuration is
never consumed to control which model a delegated phase sub-agent runs on.** The
value is read, persisted, displayed, and edited — then dead-ends.

### How a phase model is stored
- Struct: `internal/config/model.go:5-8` — `ModelConfig{ Default string; Phases map[string]string }`.
- Valid phase keys: `internal/config/model.go:55-64` (`ValidPhases`).
- Loaded from `.archon/config.yaml`: `internal/config/config.go:60-77` (`Load`).
- Deep-copied on every `archon update`: `internal/config/config.go:96-101` (`Clone`).

### Every place `Models.Phases` is touched — all read/persist/UI, none delegate
- `internal/config/config.go:96-101` — `Clone` copies the map (persistence plumbing).
- `internal/status/display.go:55-68` — prints the per-phase models in `archon status`.
- `internal/tui/models_tab.go:50-53` (reads into inputs) and `234-244` (writes edits back).
- `cmd/archon/config.go:135` (`config get`) and `setConfigValue` (`config set`) — CRUD.
- CLI flags feed the map at init: `cmd/archon/main.go:121-148` wires
  `--model-explore … --model-archive` into `initcmd.Options.ModelPhases`.

### The generated orchestrator template has no model awareness
- The orchestrator instructions (`CLAUDE.md` / `AGENTS.md`) are rendered from a
  static template: `internal/initcmd/templates.go:15-169`.
- `TemplateData` carries only `ProjectName, Agent, HarnessVersion, SkillCount`
  (`templates.go:171-176`). There is **no model field and no phase→model
  instruction anywhere in the rendered text.**
- Two render call sites populate `TemplateData`, and neither reads `cfg.Models`:
  - `internal/initcmd/init.go:241-256` (`writeTemplate`, called at `archon init`).
  - `internal/tui/model.go:332-358` (`regenerateTemplate`, called after a TUI save).
- The template's `## Rules` block (rule 2: "Delegate each phase to sdd-* sub-agent")
  never tells the orchestrator which model to pass per phase.

### Conclusion (root cause confirmed)
Per-phase model config is **aspirational**: the full read → persist → display → edit
loop works, but nothing closes the loop to delegation. The orchestrator (the agent
reading `CLAUDE.md`) has no knowledge of the configured models, so it cannot honor
them when delegating. This is a missing-consumer bug, not a wiring bug.

## Affected Areas
- `internal/initcmd/templates.go` — add a phase→model block to the template and a
  field to `TemplateData`. (Primary change for "vía 1".)
- `internal/initcmd/init.go:241-256` — `writeTemplate` must accept/forward the
  configured models so the rendered template is populated at init time.
- `internal/tui/model.go:332-358` — `regenerateTemplate` must forward `cfg.Models`
  so a TUI edit re-renders the model block (otherwise the TUI lets you edit models
  but the regenerated `CLAUDE.md` would still omit them).
- `internal/initcmd/templates_test.go` — template tests assert rendered content; new
  field and block need coverage.
- Possibly `internal/config/model.go` — if we decide to normalize display names to
  model IDs (see Risks), validation/normalization would live here.

## Approaches

1. **Vía 1 — Template injection (recommended)** — Add the resolved phase→model map
   to `TemplateData` and render a new section in the orchestrator template (e.g.
   under `## Configuration` or a new `## Phase Models` block) that lists each phase
   and the model the orchestrator MUST request when delegating that phase. Both
   render call sites (`init.go`, `tui/model.go`) pass `cfg.Models` through.
   - Pros: Smallest, most review-friendly diff (config struct, struct field,
     template string, two call sites, tests). No new runtime delegation machinery.
     Keeps the harness's "the orchestrator is an LLM reading CLAUDE.md" design.
     Naturally regenerates on every init and TUI save. Estimable well under the
     400-line review budget.
   - Cons: Enforcement is advisory — it depends on the orchestrator agent obeying
     the instruction text, not on a hard runtime gate. If a phase has no configured
     model, the block must degrade gracefully (omit or fall back to default).
   - Effort: Low.

2. **Vía 2 — Harness-driven delegation** — Have the harness itself pass the model
   to the Agent/Task delegation primitive at runtime (i.e. the Go code that spawns
   each phase sub-agent reads `Models.Phases[phase]` and sets the model on the
   delegation call).
   - Pros: Authoritative — the model is enforced in code, not by prompt obedience.
   - Cons: **There is no such delegation code today.** Delegation is performed by
     the orchestrator LLM via its platform's Task tool, not by the Go harness. This
     approach requires inventing a runtime delegation layer (or a plugin/hook that
     intercepts Task calls) — a large, speculative change well beyond the budget and
     outside the current architecture. Also needs a reliable mapping from config
     model names to the host platform's model-selection mechanism.
   - Effort: High.

## Recommendation

**Adopt Vía 1 (template injection).** It is the cheapest, most review-friendly fix
and fits the existing architecture, where the orchestrator is an LLM driven by the
generated `CLAUDE.md`. Render a clear, normative phase→model block so the
orchestrator passes the configured model when delegating each phase, and forward
`cfg.Models` through both render paths so init and TUI edits stay in sync. Define a
graceful fallback for phases with no configured model (fall back to `models.default`,
or omit the line). Vía 2 should be recorded as a possible future hardening, not this
change's scope.

## Risks
- **Display names vs model IDs (data-quality risk).** The live `.archon/config.yaml`
  stores human display names — `Opus 4.8`, `Sonnet 4.6`, `Haiku 4.5` — not the IDs
  the code knows (`claude-opus-4-8`, etc. in `model.go:13-17`). These display names
  are **not** in `KnownModels`, so `config.Validate` would already warn on them. The
  free-form TUI text inputs (`models_tab.go:53`) are why such values exist. Decide
  in propose/design whether the injected block passes the raw stored string as-is,
  or normalizes display names → real model IDs. Passing raw strings may instruct the
  orchestrator to "use Opus 4.8", which is not a usable model identifier.
- **Typo in the live config: `propose: Opues 4.8`** (`.archon/config.yaml`). A real
  example of the unvalidated-free-form problem. If the block passes values verbatim,
  the orchestrator would be told to use a nonexistent model for `propose`. Validation
  is advisory-only by design (`model.go:66-74`), so this slipped through.
- **How does the orchestrator actually pass a model to the Task/Agent tool?** Vía 1
  emits an instruction, but whether the host agent platform lets the orchestrator
  choose a sub-agent model per delegation is unverified. The instruction is only as
  effective as the platform's support for per-delegation model selection. Needs a
  decision on what the block tells the orchestrator to do when the platform can't
  honor it.
- **Advisory enforcement.** Vía 1 cannot guarantee the orchestrator obeys; it relies
  on prompt adherence. Acceptable for this slice but should be stated as a known
  limitation.
- **Migration of existing configs.** Configs already on disk (like the live one)
  carry display names and the typo. Decide whether this change also normalizes/repairs
  them, warns on regenerate, or leaves them untouched (template just reflects what's
  stored).
- **`TemplateData` is value-typed and widely tested.** Adding a field touches several
  tests in `templates_test.go`; the new field must render deterministically (stable
  ordering of the phase map, since Go map iteration is random) to keep golden-style
  assertions stable.

## Ready for Proposal
**Yes** — root cause is confirmed with file:line evidence and Vía 1 is a well-shaped,
low-effort slice. Before `propose`, the orchestrator should raise these decisions
with the user:
1. **Name normalization**: pass the configured model string verbatim into the block,
   or normalize display names ("Opus 4.8") to real IDs ("claude-opus-4-8")? This
   determines whether `propose: Opues 4.8` is repaired or surfaced as-is.
2. **Fallback for unset phases**: omit the line, or fall back to `models.default`?
3. **Scope of the typo/display-name cleanup**: fix only the template path now, or
   also tighten input/validation (out of scope for the cheapest fix)?
4. **Platform model-selection mechanism**: confirm how the orchestrator is expected
   to apply a per-phase model when delegating, so the block's wording is actionable.
