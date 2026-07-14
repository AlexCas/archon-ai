# Verification Report: preflight-arrow-key-questions

- Change: `preflight-arrow-key-questions` (PR 2 of 3, force-chained)
- Phase: verify
- Persistence mode: openspec
- Branch: `feat/preflight-arrow-key-questions`
- Verdict: **PASS**

## Build / Vet / Test Evidence

| Command | Result |
|---|---|
| `go build ./...` | exit 0 (no output) |
| `go vet ./internal/...` | exit 0 (no output) |
| `go test ./...` | exit 0 — all 10 packages `ok` (cmd/archon, internal/agent, internal/config, internal/initcmd, internal/models, internal/scaffold, internal/status, internal/tui, internal/version, skills) |

Targeted run (`go test ./internal/initcmd/ -v`):
- `TestTemplates_ContainSDDSessionPreflight/AGENTS.md` PASS
- `TestTemplates_ContainSDDSessionPreflight/CLAUDE.md` PASS
- `TestTemplates_BacktickRendering` PASS
- `TestTemplates_NoPreflightCodeBlock` PASS (renamed from `TestTemplates_CodeBlockRendering`; assertion negated to require ` ```text ` ABSENT)

## Test-Enforced Scenario 6 (no legacy code block)

Confirmed at `internal/initcmd/templates_test.go`:
- `TestTemplates_NoPreflightCodeBlock` (lines 269-286): asserts `strings.Contains(content, "```text")` is false. Renamed from `TestTemplates_CodeBlockRendering`; the old positive assertion is gone.
- `TestTemplates_ContainSDDSessionPreflight` (lines 116-184): `required` slice (138-155) adds the per-group + default + D3 markers; a `legacy` negative-assertion loop (169-181) asserts absence of `"Antes de continuar con SDD"`, `Responda con "usar recomendado"`, ` ```text `, `"A1 Interactivo"`, `"B1 OpenSpec"` across BOTH the AGENTS.md and CLAUDE.md renders.
- Note (positive deviation from design, not a defect): design suggested raw `!Contains("A1")`/`!Contains("B1")`; apply used the tighter `"A1 Interactivo"`/`"B1 OpenSpec"` patterns. This is correct — the new prose intentionally quotes `"A1"/"B1"/"usar recomendado"` as negative examples, so a raw substring check would false-positive. Documented in a code comment (lines 163-168).
- Both tests run and pass at runtime.

## Render-Marker Proxy Verification (against actual RenderClaudeMD output)

Verified by rendering the live template in-package (temporary test, removed after use) and by grep on committed `CLAUDE.md`. All 15 required markers present, all 10 legacy markers absent.

| # | Scenario | Enforcement | Observable markers | Result |
|---|----------|-------------|--------------------|--------|
| 1 | Five groups asked as separate questions | doc-review + render proxy | `AskUserQuestion` + all 5 headers (`A. Ritmo`, `B. Artefactos`, `C. PRs`, `D. Revisión`, `E. Pruebas web (Playwright)`) present; each framed as its own arrow-key question | PASS |
| 2 | Recommended option pre-selected default | doc-review + render proxy | `Interactivo (recomendado)`, `OpenSpec (recomendado)`, `Preguntarme (recomendado)`, `400 líneas (recomendado)`, `No (recomendado)` all present | PASS |
| 3 | D "Otro" free-text follow-up | doc-review + render proxy | `pregunta de texto libre` present in D3 entry | PASS |
| 4 | Group E asked (non-web project) | doc-review + render proxy | E marker present; project-type conditional strings (`determines whether the project is web during`, `NEW or blank project where explore cannot determine`) ABSENT | PASS |
| 5 | Group E asked (web project) | doc-review + render proxy | same E marker; no conditional gating | PASS |
| 6 | No legacy code block | TEST-ENFORCED | ` ```text ` absent; `Antes de continuar con SDD` absent; `A1 Interactivo`/`B1 OpenSpec` absent; `Responda con "usar recomendado"` absent | PASS |
| 7 | Hard gate STOP when no preflight decision | doc-review + render proxy | `ask the five per-group questions above and **STOP**` present; hard-gate rules block present | PASS |
| 8 | Choices cached & echoed into later phases | doc-review + render proxy | `Cache the choices for this session and echo them into later phase prompts` present; `provided all five choices` present. SESSION_STATUS.md recording is governed by the pre-existing unchanged `## Session Status` section. | PASS |

Legacy-absence confirmation (rendered output + committed CLAUDE.md): `Antes de continuar con SDD` (false), `Responda con "usar recomendado"` (false), ` ```text ` (false), `A1 Interactivo` (false), `B1 OpenSpec` (false), `all four choices` (false), `include them in later phase prompts` (false), residual `§` placeholder (false).

## Scope Integrity

- Files modified in the live working tree (git diff HEAD): exactly the four expected — `internal/initcmd/templates.go`, `CLAUDE.md`, `internal/initcmd/templates_test.go`, `internal/tui/model_test.go`. No files outside this set changed. (`SESSION_STATUS.md` and the change folder are untracked artifacts, not production edits.)
- `internal/tui/model_test.go` change = the flagged fix only: one line in `TestSaveConfig_RegeneratesClaudeMD` swapping the stale `"Antes de continuar con SDD"` required-section marker for `"AskUserQuestion"`.
- CLAUDE.md diverged sections UNCHANGED: `git diff HEAD -- CLAUDE.md | grep -E "^[+-]"` filtered for `You MUST delegate each phase by invoking`, `binding hard gate`, and `anthropic/` returned zero added/removed lines (exit 1). Rule 2 and the 9-id Phase Models block are untouched.
- CLAUDE.md diff is confined to the preflight block (header through the hard-gate tail); `## Vague Request Guard` and everything after are context-only.
- `templates.go` and `CLAUDE.md` edits are byte-parallel (`§`-encoded vs literal backticks), matching the design's surgical-hand-edit mandate. `TestTemplates_AgentsAndClaudeIdentical` still passes, preserving template↔template parity.

## Issues

- CRITICAL: none.
- WARNING: none.
- SUGGESTION: none.

## Verdict

**PASS** — Build, vet, and full test suite green. Test-enforced scenario 6 is covered by a renamed negated test plus negative assertions that run and pass. All 8 Gherkin scenarios satisfied (1 test-enforced, 7 doc-review with render-marker proxies confirmed against actual rendered output). Scope integrity holds: only the four expected files changed, CLAUDE.md's diverged Rule 2 and Phase Models regions are untouched.
