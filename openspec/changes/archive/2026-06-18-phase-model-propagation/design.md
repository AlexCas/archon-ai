# Design: Phase Model Propagation

## Context

Per-phase model config (`ModelConfig{Default, Phases}`, `internal/config/model.go:5-8`)
is read/persisted/displayed/edited but never reaches the generated orchestrator
template. `TemplateData` (`templates.go:171-176`) has no model field, and neither
render path reads `cfg.Models`: `writeTemplate` (`init.go:241-256`, called at
`init.go:96` with `(projectDir, agentName, skillCount)` — no `cfg`) nor
`regenerateTemplate` (`tui/model.go:332-358`, which does receive `*config.Config`).
Vía 1: resolve+normalize `cfg.Models` into an ordered slice on `TemplateData` and
render a normative `## Phase Models` block. Block is **advisory** — it instructs the
orchestrator LLM; it is not a runtime gate (Vía 2 deferred).

## Architecture overview

    cfg.Models ──► config.ResolvePhaseModels(cfg.Models) ──► []PhaseModel (ordered)
                        │  per phase: Phases[p]→normalize → Default→normalize → omit
                        ▼
    init.writeTemplate ─┐
    tui.regenerateTemplate ─┴─► TemplateData{...,PhaseModels} ──► template ──► CLAUDE.md/AGENTS.md

The resolver lives in `internal/config` so both render sites and `Validate` share one
normalization source of truth. Renderers stay dumb: they only iterate the pre-ordered
slice the resolver hands them.

## Decision 1 — Normalization (display → ID), canonical output = short alias

**Choice**: emit short aliases `opus` / `sonnet` / `haiku` in the block.
**Rationale**: aliases are accepted by the delegation tool, are stable across dated
point-releases (no `-20251001` churn), and read cleanly in human-facing CLAUDE.md.
Full IDs would couple the generated doc to exact dated IDs and bloat each line.

`NormalizeModel` is **case-insensitive**, tolerant of extra version digits, and an
idempotent passthrough for values already in canonical/accepted form.

| Input (any case)                              | Match rule                          | Output   |
|-----------------------------------------------|-------------------------------------|----------|
| `opus`, `Opus`                                | exact alias                         | `opus`   |
| `sonnet`, `haiku`                             | exact alias                         | alias    |
| `Opus 4.8`, `opus 4`, `opus-4-8`              | family `opus` + version digits      | `opus`   |
| `Sonnet 4.6`, `sonnet 4`                      | family `sonnet`                     | `sonnet` |
| `Haiku 4.5`, `haiku 4`                        | family `haiku`                      | `haiku`  |
| `claude-opus-4-8`                             | full ID → family                    | `opus`   |
| `claude-sonnet-4-6`, `claude-sonnet-4`       | full ID → family                    | `sonnet` |
| `claude-haiku-4-5`, `claude-haiku-4-5-20251001` | full ID → family                 | `haiku`  |
| `Opues 4.8` (typo), `gpt-4`, `glm-5`, ``      | no Claude family match              | `("", false)` |

Algorithm: lowercase + trim; split the value on non-alphanumeric boundaries into
tokens; if any token equals a family alias `opus`/`sonnet`/`haiku` → return that
alias (covers `claude-opus-4-8`, `Opus 4.8`, `opus-4`); else `("", false)`. Families
are checked in the fixed priority order `opus`→`sonnet`→`haiku`, so a value naming
more than one family resolves deterministically. Matching on whole tokens (not raw
substrings) means a word that merely contains a family — e.g. `octopus` — does NOT
resolve. Non-Claude families (Opencode models like `glm-5`) do not resolve here —
see Decision 2. The typo `Opues` yields no family token, so it correctly fails.

**Signature** (new in `internal/config/model.go`):

```go
// NormalizeModel maps a configured/display model value to an alias the delegation
// tool accepts (opus|sonnet|haiku). ok is false when no known model resolves.
func NormalizeModel(s string) (id string, ok bool)
```

## Decision 2 — Unknown values: WARN, never reject; OMIT from block

**Choice**: non-fatal, consistent with the existing advisory `Validate`
(`model.go:66-74`, which only returns a warning string and never errors). At render
time an unresolvable value is **omitted** (treated as "no model for this phase",
which then falls through to default→omit). The block MUST NEVER contain a string the
delegation tool can't use.
**Rationale**: rejecting would be a behavior change and would break Opencode users
whose models (`glm-5`, etc.) aren't Claude aliases. Surfacing happens via `Validate`,
satisfying the spec's "actionable feedback" requirement without a hard gate.

`Validate` is extended to also accept anything `NormalizeModel` resolves, so
`Opus 4.8` no longer warns while `Opues 4.8` still does:

```go
func Validate(model string) string {
    if model == "" { return "" }
    if KnownModels[model] { return "" }
    if _, ok := NormalizeModel(model); ok { return "" }
    return fmt.Sprintf("warning: %q is not a known model (accepted anyway)", model)
}
```

(Note: Opencode models stay "known" via `KnownModels`; they validate clean but
`NormalizeModel` returns `ok=false`, so they are omitted from the Claude phase block.
That is correct — the block only advises Claude per-delegation model selection.)

## Decision 3 — Resolution + deterministic ordering

Canonical order constant (matches the spec and CLAUDE.md phase order, excluding
non-delegated `judge`):

```go
var PhaseOrder = []string{"explore","propose","spec","design","tasks","apply","verify","archive"}

type PhaseModel struct{ Phase, Model string }

// ResolvePhaseModels returns phase→alias pairs in canonical order, omitting any
// phase that resolves to nothing. Pure; never mutates mc.
func ResolvePhaseModels(mc ModelConfig) []PhaseModel {
    var out []PhaseModel
    for _, p := range PhaseOrder {
        raw := mc.Phases[p]
        id, ok := NormalizeModel(raw)
        if !ok { id, ok = NormalizeModel(mc.Default) }
        if !ok { continue } // omit line
        out = append(out, PhaseModel{Phase: p, Model: id})
    }
    return out
}
```

Iterating `PhaseOrder` (not the map) gives byte-identical output across runs.

## Decision 4 — Template block shape (literal)

Inserted via a `{{if .PhaseModels}}` guard so a fully-empty resolution emits nothing
(no dangling header). Goes after `## Configuration` in `orchestratorTrailer`
(`templates.go:155-159`), before `## State Management`:

```
{{if .PhaseModels}}
## Phase Models

Advisory: when delegating an SDD phase, request the model below for that phase by
passing §model: <id>§ to the Agent/Task delegation tool. This is a preference, not a
hard gate; if the platform cannot honor per-delegation model selection, proceed with
the default model.

{{range .PhaseModels}}- {{.Phase}}: {{.Model}}
{{end}}{{end}}
```

(`§` is the existing backtick placeholder, `templates.go:13`.) Rendered example:

```
## Phase Models

Advisory: when delegating an SDD phase, request the model below ...

- explore: opus
- propose: sonnet
- design: opus
```

## Decision 5 — Plumbing (both paths feed the same resolver)

`TemplateData` gains one field:

```go
type TemplateData struct {
    ProjectName    string
    Agent          string
    HarnessVersion string
    SkillCount     int
    PhaseModels    []config.PhaseModel // ordered, resolved; nil ⇒ block omitted
}
```

(`templates.go` already may import `internal/config`; if not, add the import.)

**init.go**: change `writeTemplate` to take the resolved models and have the caller
pass them. `cfg` exists at the call site (`init.go:84-96`).

```go
// was: func writeTemplate(projectDir, agentName string, skillCount int) error
func writeTemplate(projectDir, agentName string, skillCount int, phaseModels []config.PhaseModel) error
// caller (init.go:96):
writeTemplate(opts.ProjectDir, agentName, len(extracted), config.ResolvePhaseModels(cfg.Models))
```

Set `data.PhaseModels = phaseModels` inside `writeTemplate`.

**tui/model.go**: `regenerateTemplate` already has `cfg *config.Config`. Add one line
to its `TemplateData` literal (`model.go:333-338`):

```go
PhaseModels: config.ResolvePhaseModels(cfg.Models),
```

No signature change there. Both paths therefore call the identical resolver on the
identical `cfg.Models`, guaranteeing byte-identical blocks (spec scenario "TUI
regeneration produces the same block as init").

## File-by-file changes

| File | Action | Change |
|------|--------|--------|
| `internal/config/model.go` | Modify | Add `PhaseOrder`, `PhaseModel`, `NormalizeModel`, `ResolvePhaseModels`; extend `Validate` to accept normalizable values. |
| `internal/initcmd/templates.go` | Modify | Add `PhaseModels` field to `TemplateData`; insert guarded `## Phase Models` block in `orchestratorTrailer`. |
| `internal/initcmd/init.go` | Modify | `writeTemplate` signature gains `phaseModels`; set on `TemplateData`; caller at L96 passes `config.ResolvePhaseModels(cfg.Models)`. |
| `internal/tui/model.go` | Modify | `regenerateTemplate` sets `PhaseModels: config.ResolvePhaseModels(cfg.Models)`. |
| `internal/config/model_test.go` | Modify | Tests for `NormalizeModel` table, `ResolvePhaseModels` fallback/omit/order, extended `Validate`. |
| `internal/initcmd/templates_test.go` | Modify | New render tests for the block; keep existing golden checks green. |

## Testing strategy

| Layer | What | How |
|-------|------|-----|
| Unit (config) | `NormalizeModel` table | Each row of the Decision-1 table incl. typo→`ok=false`, idempotent alias, dated ID, Opencode→false. |
| Unit (config) | `ResolvePhaseModels` | (a) phase set→that alias; (b) phase unset + default set→default alias; (c) phase unset + no default→omitted; (d) mixed set→canonical order; (e) called twice→`reflect.DeepEqual`. |
| Unit (config) | `Validate` | `Opus 4.8`→no warn; `Opues 4.8`→warn; `glm-5`→no warn (KnownModels); empty→no warn. |
| Render (initcmd) | block present | `TemplateData{PhaseModels: ResolvePhaseModels(...)}` → assert `## Phase Models`, `- propose: sonnet`, no raw `Opus 4.8`, no `§`. |
| Render (initcmd) | empty omits header | `PhaseModels: nil` → content excludes `## Phase Models`. |
| Render (initcmd) | both paths identical | render via init data vs. a TemplateData built the same way → byte-identical block substring. |

**Existing golden assertions touched**: none of the current `templates_test.go`
checks break — they assert substrings unaffected by the new section and zero-value
`TemplateData{}` (which now yields `PhaseModels: nil` ⇒ block omitted, so
`TestRenderAgentsMD_EmptyData`, `_FiveRules`, `_AgentsAndClaudeIdentical`, etc. stay
green). The `AgentsAndClaudeIdentical` test still holds because both templates share
`orchestratorTrailer`.

## Residual risks

- **Advisory only**: enforcement depends on orchestrator/platform honoring
  `model: <id>`. Stated in the block; Vía 2 (Go-driven delegation) is future work.
- **Alias acceptance**: assumes the delegation tool accepts `opus`/`sonnet`/`haiku`.
  Confirmed against the harness's delegation model identifiers; if a platform needs
  full IDs, swap the canonical output form in `NormalizeModel` (single function).
- **`judge` excluded** from `PhaseOrder` by design (not an `sdd-*` delegated phase);
  matches `ValidPhases`, which also omits `judge`.

## Open questions

None blocking.
