# Design: Preflight as Per-Group Arrow-Key Questions

## Technical Approach

Docs/template-only (Approach 1). Replace the fenced Spanish code block plus the
"usar recomendado / A1,B1" instruction inside `orchestratorSections`
(`internal/initcmd/templates.go`, lines 42-71) with prose that directs the
orchestrator to issue five per-group `AskUserQuestion` calls. Same groups, options,
and recommended defaults; only the asking mechanism changes. Then surgically hand-edit
the same preflight block in the committed `CLAUDE.md` (lines 26-55) and update
`templates_test.go` assertions that pin the old block.

The preflight is LLM-interpreted instruction text with no Go runtime enforcement, so
most Gherkin scenarios are doc-review-only; tests only enforce the RENDERED text
(presence of markers, absence of legacy strings, backtick integrity).

## Architecture Decisions

### Decision: CLAUDE.md update — surgical hand-edit, NOT `archon init` regen

**Choice**: Hand-edit only the preflight block (lines 26-55) in committed `CLAUDE.md`.
**Alternatives considered**: Re-run `archon init` to regenerate CLAUDE.md from template.
**Rationale**: VERIFIED — the committed CLAUDE.md has already diverged from the
template via PR-1 manual edits. `diff` of a fresh render vs. committed CLAUDE.md shows
three divergences a regen would clobber:
- Rule 2 (`## Rules`): committed says "You MUST delegate each phase by invoking its
  `archon-<phase>` subagent…"; template still says "Delegate each phase to sdd-* sub-agent".
- Phase Models prose: committed says "…binding hard gate — Claude Code selects the model
  from the subagent definition…"; template says "Advisory: …passing `model: <id>`…".
- Phase Models list: committed has all 9 phases with `anthropic/…` ids; template renders
  one placeholder (`- explore: opus`).

A regen would overwrite all three. Therefore apply MUST NOT regen. Apply edits the
preflight block in CLAUDE.md by hand to match the new template block byte-for-byte
(after `§`→backtick substitution). The `TestTemplates_AgentsAndClaudeIdentical` render
test still guarantees template↔template parity; committed-file parity is out of scope
(already broken pre-change, tracked separately).

### Decision: Backtick encoding in the new prose

**Choice**: Use `§` placeholder for every backtick in the template block; the new prose
needs backticks for `AskUserQuestion`, `playwright.enabled`, `.archon/config.yaml`,
`--playwright`, `archon tui`, `openspec/config.yaml`, `sdd-init`. The former `§§§text`
triple-backtick fence is DELETED (no code block remains).
**Alternatives considered**: Drop inline code formatting.
**Rationale**: Keep the existing `§` mechanism (line 15 / renderTemplate line 200). The
committed CLAUDE.md uses literal backticks directly (it is post-render output).

### Decision: D3 free-text follow-up stays instruction-only

**Choice**: Document "If Otro → ask a free-text question for the line count, then cache".
**Rationale**: No Go component; AskUserQuestion has no numeric-input mode. Orchestrator
issues a plain follow-up. Doc-review-only.

## Literal Replacement Block (templates.go, replaces lines 42-71)

Replaces from `**User-facing prompt (Spanish):**` through the closing `§§§` fence.
The `**Required choices:**` list (lines 36-40) and `**Project type…**` /
`**Hard gate rules:**` blocks (lines 73-82) are PRESERVED, except the E project-type
conditional bullet is removed (see below).

```
**Preflight questions (Spanish, arrow-key):**

Ask each group A–E as its OWN separate arrow-key §AskUserQuestion§ — never a single
text block, never answer codes like "A1"/"B1", never a global "usar recomendado"
shortcut. Pre-select the recommended option as the default in each question. Ask all
five every SDD session.

- **A. Ritmo** — "¿Qué ritmo quieres para las fases?"
  - Interactivo (recomendado): mostrar cada fase y esperar confirmación antes de continuar.
  - Automático: ejecutar las fases seguidas y frenar solo ante riesgo alto.
- **B. Artefactos** — "¿Dónde guardamos los artefactos?"
  - OpenSpec (recomendado): archivos en el repo, trazables en revisión.
  - Engram: más rápido, sin archivos de especificación en el repo.
  - Ambos: archivos OpenSpec más copia en Engram.
- **C. PRs** — "¿Qué estrategia de PRs?"
  - Preguntarme (recomendado): frenar y preguntar si la estimación supera el presupuesto.
  - Un solo PR: intentar mantener el cambio en un PR.
  - Encadenados: separar en PRs encadenados desde el inicio.
  - Auto: decidir según la estimación de tamaño.
- **D. Revisión** — "¿Presupuesto de líneas por revisión?"
  - 400 líneas (recomendado): frenar si la estimación supera 400 líneas cambiadas.
  - 800 líneas: más permisivo; útil para cambios medianos.
  - Otro: al elegir esta opción, hacer UNA pregunta de texto libre pidiendo el número
    de líneas y usar ese valor como presupuesto.
- **E. Pruebas web (Playwright)** — "¿Generar y correr pruebas Playwright?"
  - No (recomendado): no generar ni ejecutar pruebas Playwright.
  - Sí: generar pruebas Playwright desde los escenarios Gherkin y ejecutarlas tras verify y jueces.
```

Then the PRESERVED tail (E conditional bullet dropped):

```
**Project type & web testing (group E):**
- Group E maps to §playwright.enabled§ in §.archon/config.yaml§. The §--playwright§ flag
  at init time or the Playwright tab in §archon tui§ set the same value. When enabled,
  the harness generates Playwright specs from Gherkin scenarios and runs them after the
  verify and judge phases.

**Hard gate rules:**
- §openspec/config.yaml§, existing SDD artifacts, or previous §sdd-init§ results do NOT
  satisfy this preflight.
- If the session has no preflight decision, ask the five per-group questions above and
  **STOP**. Do not run init, delegate phases, or apply tasks in the same turn.
- Cache the choices for this session and echo them into later phase prompts.
- If the user explicitly provided all five choices in the current conversation, summarize
  them as the session preflight block and continue.
```

Note: the old bullet "For a NEW or blank project where explore cannot determine the type,
ASK group E together with the rhythm…" is DELETED (group E always asked). The old
`sdd-explore` project-type detection bullet is also dropped from this block as it only
served the removed conditional.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/initcmd/templates.go` (42-76) | Modify | Replace code block + E conditional with the per-group prose above; keep `§` encoding. |
| `CLAUDE.md` (26-60) | Modify | Surgical hand-edit of the SAME block, backticks literal (post-render). No regen. |
| `internal/initcmd/templates_test.go` (~142, ~250-253) | Modify | Swap old-literal assertions for new markers (below). |

## Test Update Plan (templates_test.go)

**`TestTemplates_ContainSDDSessionPreflight` (line ~138-144)** — in `required`:
- REMOVE `"Antes de continuar con SDD"`.
- KEEP `"## SDD Session Preflight (HARD GATE)"`, `"## Vague Request Guard (MANDATORY)"`,
  `"## Human Review Gate (MANDATORY)"`, `"¿Quieres ajustar algo…"`.
- ADD per-group + behavior markers (arrow-key + all five groups + defaults + D3 + no-legacy):
  - `"AskUserQuestion"`
  - `"A. Ritmo"`, `"B. Artefactos"`, `"C. PRs"`, `"D. Revisión"`, `"E. Pruebas web (Playwright)"`
  - `"Interactivo (recomendado)"`, `"OpenSpec (recomendado)"`, `"Preguntarme (recomendado)"`,
    `"400 líneas (recomendado)"`, `"No (recomendado)"`
  - `"pregunta de texto libre"` (D3 follow-up)

Add explicit NEGATIVE assertions (satisfies @error scenario 6) in this test or a new
`TestTemplates_NoLegacyPreflight`:
- `!strings.Contains(content, "Antes de continuar con SDD")`
- `!strings.Contains(content, "usar recomendado")`
- `!strings.Contains(content, "A1")` and `!strings.Contains(content, "B1")`

**`TestTemplates_CodeBlockRendering` (line ~238-254)** — MUST change. It asserts
` ```text ` exists; the fenced block is removed. Either delete this test or repurpose it
to assert the code block is GONE (`!strings.Contains(content, "```text")`). Recommend
repurpose+rename to `TestTemplates_NoPreflightCodeBlock`.

**`TestTemplates_BacktickRendering`** — KEEP unchanged (`interactive`, `auto`, `openspec`,
`engram`, `sdd-explore`, `sdd-propose`, `internal/billing` all still present elsewhere;
the `§`→backtick and "no `§` remains" invariants still hold with the new backticks).

## Verification Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | Rendered template has 5 group markers, recommended defaults, D3 follow-up, no legacy strings, no `text` fence, backtick integrity | `go test ./internal/initcmd/` (updated tests above) |
| Doc | Committed CLAUDE.md preflight block matches new prose | Manual inspection / grep in verify |

### Gherkin → Verification Mapping

| Scenario | Enforcement |
|----------|-------------|
| 1. Five groups asked as separate questions | Doc-review (runtime behavior). Test proxy: 5 group markers + `AskUserQuestion` present in render. |
| 2. Recommended option is pre-selected default | Doc-review. Test proxy: `(recomendado)` markers on the 5 defaults. |
| 3. D "Otro" triggers free-text follow-up | Doc-review. Test proxy: `"pregunta de texto libre"` present. |
| 4. Group E asked (non-web) | Doc-review. Test proxy: E marker present + no project-type conditional string. |
| 5. Group E asked (web) | Doc-review (same marker; conditional removed). |
| 6. No legacy code block | TEST-ENFORCED: negative assertions (no `"Antes de continuar con SDD"`, no `A1`/`B1`, no `usar recomendado`, no `text` fence). |
| 7. Hard gate blocks when no preflight | Doc-review. Test proxy: STOP text + hard-gate rules present. |
| 8. Choices cached & echoed | Doc-review. Test proxy: "Cache the choices… echo them into later phase prompts" present. |

Test-enforced: 6 (fully), plus render proxies for 1-5,7-8. Doc-review-only for the
actual orchestrator runtime behavior in 1-5,7-8.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `TestTemplates_CodeBlockRendering` left asserting removed ` ```text ` → red build | High | Explicitly repurpose/delete it (called out above). |
| CLAUDE.md regen clobbers PR-1 edits | High if regen used | VERIFIED via diff; design mandates surgical hand-edit, no regen. |
| CLAUDE.md drifts from new template block | Med | Apply copies the same block (post-render backticks); verify greps both. |

## Open Questions

- None. CLAUDE.md regen-vs-surgical resolved: **surgical hand-edit only**.
