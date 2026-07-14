# Proposal: Preflight as Per-Group Arrow-Key Questions

## Intent

The SDD preflight is emitted today as one Spanish text block ("Antes de continuar
con SDD…") that asks the user to reply with codes like `A1, B1, C1, D1` or the
`usar recomendado` fast-path. Codes are error-prone and unfriendly. The orchestrator
already has an arrow-key `AskUserQuestion` capability; the preflight should use it —
one selectable question per group (A Ritmo, B Artefactos, C PRs, D Revisión,
E Playwright), each with its recommended option marked. Choices and semantics are
frozen; only the asking mechanism changes.

## Scope

### In Scope
- Rewrite the `## SDD Session Preflight (HARD GATE)` section in
  `internal/initcmd/templates.go` (lines 32-82) to instruct the orchestrator to ask
  each group A–E as its own `AskUserQuestion` with the recommended option marked.
- Regenerate the repo's committed `CLAUDE.md` preflight section (lines 27-82) to match.
- Update `internal/initcmd/templates_test.go` assertions that pin the old block text.

### Out of Scope
- Any change to the set of groups, their options, or the recommended defaults.
- `.archon/config.yaml` schema or the Playwright wiring in `init`/TUI/config.
- Go runtime behavior of `archon init` / `archon tui`; no new Go prompt component.
- README / harness-workflow prose (no consistency touch-up needed this slice).

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- None. The preflight is LLM-interpreted instruction text generated into
  CLAUDE.md/AGENTS.md; no `openspec/specs/` requirement changes. Pure template/docs
  rewording. (`harness-init` describes generation, not preflight wording.)

## Approach

Approach 1 from exploration (docs/template-only). Replace the fenced Spanish code
block plus the "usar recomendado / codes" instruction with directions telling the
orchestrator to issue five per-group arrow-key questions. Per-group defaults replace
the global `usar recomendado` fast-path. Group E is ALWAYS asked (drop the old
"only when project type unknown" conditional). D3 triggers a free-text follow-up.

### Per-Group Question Mapping (AskUserQuestion)

| Group | Question (Spanish) | Options (recommended marked) | Follow-up |
|-------|--------------------|------------------------------|-----------|
| A Ritmo | ¿Qué ritmo quieres para las fases? | Interactivo (recomendado) · Automático | — |
| B Artefactos | ¿Dónde guardamos los artefactos? | OpenSpec (recomendado) · Engram · Ambos | — |
| C PRs | ¿Qué estrategia de PRs? | Preguntarme (recomendado) · Un solo PR · Encadenados · Auto | — |
| D Revisión | ¿Presupuesto de líneas por revisión? | 400 (recomendado) · 800 · Otro | If **Otro**: free-text ask for the number |
| E Playwright | ¿Generar y correr pruebas Playwright? | No (recomendado) · Sí | — |

Each option keeps its current one-line explanation as the option description. The
recommended option is the pre-selected default. Retain the sentence that the TUI
Playwright tab and `--playwright` flag set the same `playwright.enabled` value, and
the Hard-gate rules (cache choices, STOP if none, echo into later phases).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/initcmd/templates.go` (32-82) | Modified | Rewrite preflight section to per-group AskUserQuestion; drop code block + `usar recomendado`. |
| `CLAUDE.md` (27-82) | Modified | Regenerated from the template. |
| `internal/initcmd/templates_test.go` | Modified | Replace the `"Antes de continuar con SDD"` assertion (line 142) with new per-group markers; keep backtick/`interactive`/`auto` checks. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Test pins old literal block (`"Antes de continuar con SDD"`) | High | Update assertion to new wording in the same slice. |
| Orchestrator ignores instruction (no Go enforcement) | Low | Already true of all preflight today; instruction is explicit + STOP gate. |
| CLAUDE.md drifts from template | Med | Regenerate in the same commit; render test covers presence. |

## Rollback Plan

Revert the three edited files (`templates.go`, `CLAUDE.md`, `templates_test.go`) to
their pre-change state via `git revert`/`git checkout`. No data, config, or migration
side effects — the change is text-only.

## Dependencies

- Stacked on PR 1 of the 3-PR chain (force-chained). This is PR 2 of 3.

## Success Criteria

- [ ] Rendered CLAUDE.md/AGENTS.md instruct five per-group arrow-key questions, each
      with the recommended option marked; no `usar recomendado` block or A1/B1 codes.
- [ ] Group D "Otro" documents a free-text follow-up for the custom line count.
- [ ] Group E is always asked (no project-type conditional).
- [ ] Same groups, options, and recommended defaults as before.
- [ ] `go test ./internal/initcmd/` passes.
