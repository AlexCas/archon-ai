# Tasks: Preflight as Per-Group Arrow-Key Questions

## Overview

Docs/template-only change. Three files are touched:
1. `internal/initcmd/templates.go` — replace the `§§§text` fenced preflight block with per-group prose
2. `CLAUDE.md` — surgical hand-edit of the same preflight block (no `archon init` regen)
3. `internal/initcmd/templates_test.go` — update assertions that pin the removed block

Stacked on PR 1 of 3. This is PR 2 of 3 (force-chained strategy).

---

## Phase 1: Template Rewrite (`internal/initcmd/templates.go`)

### T1.1 — Identify the exact deletion range
- File: `internal/initcmd/templates.go`
- Confirm current line 42 reads `**User-facing prompt (Spanish):**` and line 71 is the closing `§§§` fence of the text block.
- The replacement block starts immediately after line 41 (the `Required choices:` list) and ends before line 73 (`**Project type & web testing (group E):**`).
- Acceptance: lines 42-71 (inclusive) are the only lines being removed/replaced.

### T1.2 — Delete the old user-facing prompt block (lines 42-71)
- Remove the entire block from `**User-facing prompt (Spanish):**` through the closing `§§§` triple-backtick fence.
- This is 30 lines of text, including the blank line before `**User-facing prompt:**`, the `§§§text` open, the Spanish code block body (groups A–E with A1/B1/etc. codes and `usar recomendado` instruction), and the `§§§` close.
- Acceptance: no `§§§text`, no `§§§` fence, no `Antes de continuar con SDD`, no `usar recomendado`, no `A1`, `B1`, `C1`, `D1` codes remain in the template string.

### T1.3 — Insert the new per-group prose block
- In place of the deleted lines, insert the following block verbatim (preserving `§` backtick encoding throughout):

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

- Acceptance: block inserted with `§` encoding intact; `§AskUserQuestion§` appears; all five group headers (`A. Ritmo`, `B. Artefactos`, `C. PRs`, `D. Revisión`, `E. Pruebas web (Playwright)`) appear; `Interactivo (recomendado)`, `OpenSpec (recomendado)`, `Preguntarme (recomendado)`, `400 líneas (recomendado)`, `No (recomendado)` appear; `pregunta de texto libre` appears in the D3 entry.

### T1.4 — Remove the two stale bullets from the `**Project type & web testing (group E):**` block
- Current lines 74-75 (inside `**Project type & web testing (group E):**`):
  - Line 74: `- The orchestrator determines whether the project is web during §sdd-explore§ ...`
  - Line 75: `- For a NEW or blank project where explore cannot determine the type, ASK group E together with the rhythm (group A) during preflight.`
- Delete BOTH bullets. Keep only the single remaining bullet starting with `- Group E maps to §playwright.enabled§ in §.archon/config.yaml§...`
- Acceptance: after the edit, `**Project type & web testing (group E):**` is immediately followed by the single `Group E maps to` bullet. The `sdd-explore` detection bullet and the `NEW or blank project` conditional bullet are absent from the template.

### T1.5 — Update the hard gate tail to match new language
- Current line 80: `- If the session has no preflight block, ask the prompt above and **STOP**. Do not run init, delegate phases, or apply tasks in the same turn.`
- Current line 82: `- If the user explicitly provided all four choices in the current conversation, summarize them as the session preflight block and continue.`
- Replace with (verbatim from design):
  - `- If the session has no preflight decision, ask the five per-group questions above and **STOP**. Do not run init, delegate phases, or apply tasks in the same turn.`
  - `- Cache the choices for this session and echo them into later phase prompts.`
  - `- If the user explicitly provided all five choices in the current conversation, summarize them as the session preflight block and continue.`
- Note: "four choices" becomes "five choices"; "include them in later phase prompts" becomes "echo them into later phase prompts"; "ask the prompt above" becomes "ask the five per-group questions above".
- Acceptance: no "four choices", no "ask the prompt above", no "include them in later phase prompts" remain in the hard gate tail.

### T1.6 — Verify templates.go compiles
- Run `go build ./internal/initcmd/` from the repo root.
- Acceptance: zero build errors.

---

## Phase 2: CLAUDE.md Surgical Hand-Edit

### T2.1 — Identify the CLAUDE.md preflight block boundaries
- File: `CLAUDE.md`
- Current block starts at line 26 (`**User-facing prompt (Spanish):**`) and ends at line 66 (`- If the user explicitly provided all four choices in the current conversation, summarize them as the session preflight block and continue.`).
- Lines 68+ (`## Vague Request Guard`) must NOT be touched.
- Diverged sections that must NOT be touched: `## Rules` (Rule 2 text), `## Phase Models` block (all 9 phases with `anthropic/…` ids, binding hard gate prose).
- Acceptance: only the preflight block is modified; a `git diff CLAUDE.md` shows changes only between `**User-facing prompt (Spanish):**` and the last hard gate bullet.

### T2.2 — Replace the fenced Spanish code block in CLAUDE.md
- Delete lines 26-55 (the `**User-facing prompt (Spanish):**` header, the blank line, the ` ```text ` open fence, the code block body, and the ` ``` ` close fence).
- Insert in their place the per-group prose from T1.3, but with LITERAL backticks (not `§` encoding), because CLAUDE.md is post-render output:

```
**Preflight questions (Spanish, arrow-key):**

Ask each group A–E as its OWN separate arrow-key `AskUserQuestion` — never a single
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

- Acceptance: ` ```text ` no longer appears in CLAUDE.md; `Antes de continuar con SDD` is absent; `AskUserQuestion` (with literal backticks) is present.

### T2.3 — Remove the two stale E-conditional bullets in CLAUDE.md
- In the `**Project type & web testing (group E):**` block (currently lines 57-60), delete:
  - The bullet `- The orchestrator determines whether the project is web during \`sdd-explore\`...`
  - The bullet `- For a NEW or blank project where explore cannot determine the type, ASK group E together with the rhythm (group A) during preflight.`
- Keep only: `- Group E maps to \`playwright.enabled\` in \`.archon/config.yaml\`. The \`--playwright\` flag at init time or the Playwright tab in \`archon tui\` set the same value. When enabled, the harness generates Playwright specs from Gherkin scenarios and runs them after the verify and judge phases.`
- Acceptance: `sdd-explore` does NOT appear in the preflight block (it still appears elsewhere in CLAUDE.md, e.g. Vague Request Guard — do not touch those).

### T2.4 — Update the hard gate tail in CLAUDE.md
- Apply the same language changes as T1.5 but with literal backticks:
  - `- If the session has no preflight decision, ask the five per-group questions above and **STOP**. Do not run init, delegate phases, or apply tasks in the same turn.`
  - `- Cache the choices for this session and echo them into later phase prompts.`
  - `- If the user explicitly provided all five choices in the current conversation, summarize them as the session preflight block and continue.`
- Acceptance: "four choices", "ask the prompt above", "include them in later phase prompts" are absent from the hard gate tail; "five choices", "ask the five per-group questions above", "echo them into later phase prompts" are present.

### T2.5 — Verify diverged sections are untouched
- `git diff CLAUDE.md` must NOT show any change to:
  - `## Rules` Rule 2 line (`You MUST delegate each phase by invoking its \`archon-<phase>\` subagent…`)
  - `## Phase Models` section (all 9 phases with `anthropic/…` ids and the binding-hard-gate prose)
- Acceptance: `git diff CLAUDE.md | grep "archon-<phase>\|binding hard gate\|anthropic/"` returns lines prefixed with a space (context), not `+` or `-`.

---

## Phase 3: Test Updates (`internal/initcmd/templates_test.go`)

### T3.1 — Update `TestTemplates_ContainSDDSessionPreflight` (~line 116)
- In the `required` slice (currently lines 138-144):
  - REMOVE the entry `"Antes de continuar con SDD"` (line 142).
  - KEEP: `"## SDD Session Preflight (HARD GATE)"`, `"## Vague Request Guard (MANDATORY)"`, `"## Human Review Gate (MANDATORY)"`, `"¿Quieres ajustar algo en esta fase antes de continuar?"`.
  - ADD the following new entries to `required`:
    - `"AskUserQuestion"`
    - `"A. Ritmo"`
    - `"B. Artefactos"`
    - `"C. PRs"`
    - `"D. Revisión"`
    - `"E. Pruebas web (Playwright)"`
    - `"Interactivo (recomendado)"`
    - `"OpenSpec (recomendado)"`
    - `"Preguntarme (recomendado)"`
    - `"400 líneas (recomendado)"`
    - `"No (recomendado)"`
    - `"pregunta de texto libre"`
- Acceptance: all new entries are in `required`; `"Antes de continuar con SDD"` is not.

### T3.2 — Add negative (legacy-removal) assertions to `TestTemplates_ContainSDDSessionPreflight` or a new `TestTemplates_NoLegacyPreflight`
- After the positive-assertion loop, add explicit negative checks on `content` for BOTH the AGENTS.md and CLAUDE.md renders:
  - `!strings.Contains(content, "Antes de continuar con SDD")`
  - `!strings.Contains(content, "usar recomendado")`
  - `!strings.Contains(content, `\`\`\`text`)` (the triple-backtick text fence)
  - `!strings.Contains(content, "A1")` within a word-boundary check OR `!strings.Contains(content, " A1 ")` — add a comment noting this is checking for the old answer-code pattern, not incidental occurrences
  - `!strings.Contains(content, "B1")`
- If adding to the existing test becomes unwieldy, extract to a new function `TestTemplates_NoLegacyPreflight` with its own `tests` table covering both AGENTS.md and CLAUDE.md renders.
- Acceptance: running `go test ./internal/initcmd/ -run TestTemplates_NoLegacyPreflight` (or the modified existing test) fails before the template change and passes after it.

### T3.3 — Repurpose `TestTemplates_CodeBlockRendering` → `TestTemplates_NoPreflightCodeBlock` (~line 238)
- Rename the function from `TestTemplates_CodeBlockRendering` to `TestTemplates_NoPreflightCodeBlock`.
- Replace the current assertion body:
  ```go
  // OLD (will fail after template change):
  if !strings.Contains(content, "```text") {
      t.Error("RenderAgentsMD() missing ```text code block opening")
  }
  ```
  with:
  ```go
  // NEW (asserts the legacy fence is gone):
  if strings.Contains(content, "```text") {
      t.Error("RenderAgentsMD() contains legacy ```text preflight code block — should be removed")
  }
  ```
- Keep the same `TemplateData` setup and `RenderAgentsMD` call.
- Acceptance: renamed function exists; the assertion is negated; `go test ./internal/initcmd/ -run TestTemplates_NoPreflightCodeBlock` passes after the template change.

### T3.4 — Confirm `TestTemplates_BacktickRendering` needs no changes
- Read the test (~lines 203-236). Verify all checked backtick-wrapped strings (`interactive`, `auto`, `openspec`, `engram`, `sdd-explore`, `sdd-propose`, `internal/billing`) are present in the template outside the deleted preflight block.
- Verify the "no `§` remains" invariant is still enforced (new prose uses `§AskUserQuestion§` and other `§`-encoded terms that will be rendered, so the invariant holds).
- No code change needed. Annotate this task as verified-unchanged.
- Acceptance: `TestTemplates_BacktickRendering` passes without modification after the template and CLAUDE.md edits.

### T3.5 — Run the full test suite
- Run `go test ./internal/initcmd/` from the repo root.
- Acceptance: all tests pass; zero failures; no `§` placeholder leak reported.

---

## Phase 4: Verification and Build Check

### T4.1 — Build check
- Run `go build ./...` from the repo root.
- Acceptance: zero build errors across all packages.

### T4.2 — Grep verification: no legacy strings in templates.go
- Run: `grep -n "Antes de continuar\|usar recomendado\|§§§\|A1 Interactivo\|B1 OpenSpec" internal/initcmd/templates.go`
- Acceptance: zero matches.

### T4.3 — Grep verification: no legacy strings in CLAUDE.md
- Run: `grep -n "Antes de continuar\|usar recomendado\|\`\`\`text\|A1 Interactivo\|B1 OpenSpec" CLAUDE.md`
- Acceptance: zero matches.

### T4.4 — Grep verification: new markers present in CLAUDE.md
- Run: `grep -n "AskUserQuestion\|pregunta de texto libre\|cinco\|five per-group\|No (recomendado)" CLAUDE.md`
- Acceptance: all five group markers and key phrases appear in CLAUDE.md.

### T4.5 — Verify CLAUDE.md diverged sections are untouched (pre-commit sanity)
- Run: `git diff CLAUDE.md`
- Inspect that the diff contains NO changes to:
  - Lines containing `You MUST delegate each phase by invoking its` (Rule 2)
  - Lines containing `anthropic/claude-` (Phase Models list)
  - Lines containing `binding hard gate` (Phase Models prose)
- Acceptance: those lines appear only as context in the diff, not as deletions or additions.

### T4.6 — Confirm test coverage mapping to Gherkin scenarios
- Verify the following mapping from `design.md § Gherkin → Verification Mapping` is satisfied by the updated test suite:
  - Scenario 6 (no legacy code block): enforced by `TestTemplates_NoPreflightCodeBlock` + negative assertions in T3.2
  - Scenarios 1-2 (five groups, recommended defaults): render proxies in `TestTemplates_ContainSDDSessionPreflight`
  - Scenario 3 (D3 free-text): `"pregunta de texto libre"` in `required`
  - Scenarios 4-5 (E always asked): E marker in `required`; no `sdd-explore` project-type conditional string
  - Scenario 7 (hard gate STOP): `"**STOP**"` or similar present
  - Scenario 8 (choices cached/echoed): `"Cache the choices"` and `"echo them into later phase prompts"` present
- Document any scenario that remains doc-review-only as a comment in the test file or a note in this task list.
- Acceptance: all test-enforced scenarios have a corresponding test; doc-review-only scenarios are acknowledged.

---

## Review Workload Forecast

| File | Lines removed | Lines added | Net change estimate |
|------|--------------|-------------|---------------------|
| `internal/initcmd/templates.go` | ~35 (fenced block + 2 E-conditional bullets + 3 hard-gate tail lines) | ~38 (per-group prose + updated tail) | ~+3 net, ~73 changed |
| `CLAUDE.md` | ~35 (same block, literal backticks) | ~38 (same replacement) | ~+3 net, ~73 changed |
| `internal/initcmd/templates_test.go` | ~10 (old assertions in 2 tests) | ~25 (new assertions, renamed test, negative checks) | ~+15 net, ~35 changed |
| **Total** | **~80** | **~101** | **~181 changed lines** |

**400-line budget risk**: Low. Estimated ~181 changed lines, well within the 400-line budget. No chained-PR split needed within PR 2.

**Chained-PR note**: This is PR 2 of 3 (force-chained strategy). PR 1 is the diverged CLAUDE.md edits (Rule 2 + Phase Models). PR 3 (if planned) handles any follow-on work. This PR must be stacked on top of the PR 1 branch and should not be merged until PR 1 is merged first. Apply should target the PR 1 branch as its base, not `master`.
