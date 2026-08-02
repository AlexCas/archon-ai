---
name: sdd-router
description: "Trigger: archon route CLASSIFY signal only. Classify a user message into one of the nine SDD phases, or ASK, using topic keywords."
user-invocable: false
license: MIT
metadata:
  
  version: "1.0"
  scope: orchestrator-gate
---

## Purpose

Single-responsibility model classifier: "which phase family is this message
about?" → one of the nine SDD phases, or `ASK`. This skill is the FALLBACK
layer of the hybrid router — the deterministic `archon route` CLI already
resolved `explicit-agent`, `control`, `implicit`, and `ambiguous` (D3) rules in
code. If you are consulting this skill, none of those matched.

## Activation Contract

Invoke this skill ONLY when `archon route '<message>'` returned
`{"phase":"CLASSIFY","rule":"classify","path":"model", ...}` on stdout. Do NOT
run this classifier speculatively, and do NOT re-derive state arithmetic
(successor phase, resume-vs-start) — the code path already owns that. This
skill never reads `state.yaml` and never computes a "next phase."

## Keyword Table (Spanish + English)

| Phase | Spanish | English |
|-------|---------|---------|
| explore | explora, investiga, entender, analiza el codigo | explore, investigate, understand, research |
| propose | propon, propuesta, propongamos, enfoque | propose, proposal, approach, suggest |
| spec | especificacion, requisitos, gherkin, escenarios | spec, specification, requirements, scenarios |
| design | diseno, disenemos, arquitectura, plan tecnico | design, architecture, technical plan |
| tasks | tareas, desglose, checklist, plan de tareas | tasks, breakdown, checklist, work items |
| apply | implementa, codifica, aplica, construye | apply, implement, code, build |
| verify | verifica, prueba, corre las pruebas, valida | verify, test, run tests, validate |
| judge | juzga, revisa, revision, dictamen | judge, review, dual review |
| archive | archiva, finaliza, cierra, completa el cambio | archive, finalize, close, complete |

## Hard Rules

- Action verb beats object noun: classify by the VERB, not by an object noun
  it happens to touch (e.g. "implementa las tareas" → `apply`, not `tasks`,
  because "implementa" is the action).
- Unclassifiable, off-topic, or ambiguous message → `ASK`. Never guess a
  phase you are not confident about.
- Zero or multiple equally plausible phase matches → `ASK`.
- Do NOT perform state arithmetic, successor-phase lookup, or resume-vs-start
  reasoning — those are code-path responsibilities you never re-derive here.
- Provider-neutral: this skill contains no Claude-specific or other
  provider-specific delegation primitives; it only classifies text.

## Output Contract

Emit exactly ONE line and then STOP — do not begin executing the classified
phase in the same turn:

```
→ Router: archon-<phase>  (rule: classify, active-change: <name|none>)
```

or, when unclassifiable:

```
→ Router: ASK
```

The leader then hands the result to `harness-workflow` (for a resolved
phase) or surfaces the question to the user (for `ASK`) — this skill never
performs either step itself.
