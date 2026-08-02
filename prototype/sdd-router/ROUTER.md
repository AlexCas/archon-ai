# SDD Router (prototype) — deterministic phase dispatch for weak local models

> **Purpose.** Replace intent *inference* with a *lookup*. A weak local model must
> not have to reason "trabajemos en la spec" → explore phase → launch agent. It
> runs the numbered algorithm below and follows the table. No judgment required.
>
> **Scope of prototype.** This file is standalone. It does NOT modify CLAUDE.md.
> To test: paste this file's contents into the local model's context (or load it
> as a skill) and run the fixtures in `fixtures.md`.

## Contract

- The router NEVER executes a phase itself. It resolves ONE thing: `target_phase`
  (or `ASK`), then hands off to `harness-workflow` for gate validation and to the
  `archon-<target_phase>` subagent for execution.
- The router ALWAYS echoes its decision in one line before handing off:
  `→ Router: archon-<phase>  (rule: <rule-id>, active-change: <name|none>)`
  This makes routing an explicit, correctable step.
- On no confident match the router outputs `ASK` — it must NOT guess a phase.

## Inputs

1. `MSG` = the user's latest message, lowercased, accents stripped.
2. `STATE` = contents of `openspec/changes/<active>/state.yaml` if a change is
   active, else `none`. From it read `current_phase` and `status`.
   (Active change = the one named in `SESSION_STATUS.md`, else the only live
   folder under `openspec/changes/` that is not `archive/`.)

## Phase order (for "next" resolution)

`explore → propose → spec → design → tasks → apply → verify → judge → archive`

## Algorithm (run top-to-bottom, FIRST match wins)

**Precedence matters.** `explicit-agent` beats `control` beats `implicit` beats
`keyword`. The implicit-verb rule is deliberately ABOVE the keyword table so that
"trabajemos en esta **especificacion**" is read as "start/continue the flow", not
as a jump to the spec phase (the word "especificacion" here is loose, not a phase
request).

**Rule `explicit-agent`.** If `MSG` names an agent or a phase token directly as the
thing to run OR as a navigation target — e.g. "lanza explore", "agente de
exploracion", "corre el apply", "vamos a la fase spec", "volvamos al spec",
"regresa a design", "salta a tasks", "archon-design" — resolve to THAT phase.
Always wins. If the target is behind or ahead of `current_phase`, still resolve it;
`harness-workflow` will block an illegal jump and surface the reason (the router
does not pre-judge legality). A bare phase name used as a destination (preceded by
"a", "al", "a la fase", "volvamos a", "regresa a", "salta a") counts here, NOT as a
`keyword` match.

**Rule `control`.** Else if `MSG` contains any of
`siguiente | continuemos | continua | adelante | sigamos | next | procede | avanza`:
- `STATE = none` → `target = explore`  (id `next-nochange`)
- `status = in_progress` → `target = current_phase` (resume)  (id `resume`)
- `status = completed` → `target =` phase AFTER `current_phase`  (id `next`)

**Rule `implicit`.** Else if `MSG` contains any start/continue verb —
`trabajemos | empecemos | comencemos | hagamos | armemos | pongamonos | arranquemos`
(optionally + a loose noun like "esta especificacion", "esta feature", "esto"):
- `STATE = none` → `target = explore`  (id `implicit-start`)
- else → `target = current_phase` (resume where we are)  (id `implicit-resume`)

  > This is the exact case the user reported: "Trabajemos en esta especificacion"
  > fires explore (or resume) WITHOUT the user naming the agent. Note "volvamos",
  > "regresemos", "cambiemos a" are NOT in this list — when they precede a named
  > phase they are handled by `explicit-agent` (navigation target); the deliberate
  > jump is then gated by `harness-workflow`.

**Rule `keyword`.** Else scan `MSG` against the KEYWORD TABLE below.
- exactly one phase matches → that phase  (id `keyword`)
- two or more phases match → `ASK`  (id `ambiguous`)

**Rule `ask`.** Else output `ASK` with a menu of the 3 most likely phases given
`STATE`  (id `ask`). Never guess.

## Keyword table (rule `keyword`)

| Phase    | Spanish keywords                                             | English keywords                          |
|----------|--------------------------------------------------------------|-------------------------------------------|
| explore  | explora, exploremos, investiga, entender, analiza el codigo  | explore, investigate, understand, research|
| propose  | propon, propuesta, propongamos, idea, enfoque                | propose, proposal, approach, suggest      |
| spec     | especificacion, spec, requisitos, gherkin, escenarios        | spec, specification, requirements         |
| design   | diseno, disenemos, arquitectura, plan tecnico                | design, architecture, technical plan      |
| tasks    | tareas, desglose, checklist, plan de tareas, to-do           | tasks, breakdown, checklist, work items   |
| apply    | implementa, codifica, aplica, escribe el codigo, construye   | apply, implement, code, build             |
| verify   | verifica, prueba, pruebas, corre las pruebas, valida         | verify, test, run tests, validate         |
| judge    | juzga, revisa, revisar, revision, revisa el codigo, dictamen  | judge, review, dual review                |
| archive  | archiva, finaliza, cierra, completa el cambio                | archive, finalize, close, complete        |

## Handoff (after resolution)

1. Echo the decision line.
2. Call `harness-workflow` with `target_phase`. If it returns `blocked`, show the
   block reason to the user (missing phases / backward move) — do NOT override.
3. If `allowed`, delegate to the `archon-<target_phase>` subagent.
4. Preflight / Vague-Request / Human-Review gates from CLAUDE.md still apply and are
   NOT bypassed by the router — the router only decides WHICH phase, not whether
   the gates run.

## Trade-off toggle

If your team wants "trabajemos en la spec" to JUMP to spec rather than resume the
current phase, move rule `keyword` ABOVE rule `implicit`. Safer default is
implicit-first ("don't skip phases"). `harness-workflow` blocks illegal jumps
either way, so keyword-first is safe but noisier with block messages.
