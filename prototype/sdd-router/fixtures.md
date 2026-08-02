# SDD Router — test fixtures

Each row: user message + assumed state → expected router output. Run these against
the local model with `ROUTER.md` in context. The model should output the echo line
and nothing else (it must not start executing a phase).

The critical cases are #1–#4: implicit phrasings that fail on weak models today.

| #  | User message (MSG)                                           | State (phase / status)    | Expected target | Rule id          |
|----|--------------------------------------------------------------|---------------------------|-----------------|------------------|
| 1  | Trabajemos en esta especificacion                            | none                      | explore         | implicit-start   |
| 2  | Trabajemos en esta especificacion                            | spec / in_progress        | spec (resume)   | implicit-resume  |
| 3  | Empecemos con esto                                           | none                      | explore         | implicit-start   |
| 4  | Hagamos esta feature                                         | none                      | explore         | implicit-start   |
| 5  | Hagamos esta especificacion. Lanza el agente de exploracion  | none                      | explore         | explicit-agent   |
| 6  | Continuemos                                                  | propose / completed       | spec            | next             |
| 7  | Siguiente                                                    | apply / completed         | verify          | next             |
| 8  | Continuemos                                                  | design / in_progress      | design (resume) | resume           |
| 9  | Adelante                                                     | none                      | explore         | next-nochange    |
| 10 | Explora el codigo de billing                                 | none                      | explore         | keyword          |
| 11 | Disenemos la arquitectura del API                            | spec / completed          | design          | keyword          |
| 12 | Implementa las tareas                                        | tasks / completed         | apply           | keyword          |
| 13 | Corre las pruebas                                            | apply / completed         | verify          | keyword          |
| 14 | Archiva el cambio                                            | judge / completed         | archive         | keyword          |
| 15 | Revisa y prueba esto                                         | verify / in_progress      | ASK (ambiguous) | ambiguous        |
| 16 | Que opinas del clima?                                        | spec / in_progress        | ASK             | ask              |
| 17 | Volvamos al spec                                             | design / completed        | spec → BLOCKED  | explicit-agent → workflow blocks backward |
| 18 | corre el apply                                               | tasks / completed         | apply           | explicit-agent   |

## Trace of the tricky rows (with corrected precedence)

Precedence is `explicit-agent > control > implicit > keyword`.

- **#1** "trabajemos en esta especificacion", none → not explicit-agent, not control,
  `implicit` matches ("trabajemos") → state none → **explore**. The word
  "especificacion" never reaches the keyword table because implicit runs first.
  This is the whole point of moving implicit above keyword.

- **#2** same message, but spec is in_progress → `implicit` → resume → **spec**.

- **#5** implicit verb present ("hagamos") BUT "lanza el agente de exploracion" is an
  explicit agent command → `explicit-agent` wins → **explore**. Proves the current
  WORKING phrasing the user described stays working.

- **#11** "disenemos" is a `design` keyword and is NOT in the implicit-verb list, so
  it falls through to `keyword` → **design**. (Only trabajemos/empecemos/comencemos/
  hagamos/armemos/pongamonos/arranquemos are implicit; conjugated phase verbs are
  not.)

- **#15** "revisa" (judge) + "prueba" (verify) both match → **ASK / ambiguous**. A
  weak model must NOT silently pick one.

- **#17** "volvamos al spec" names `spec` as a navigation target, so `explicit-agent`
  resolves it to spec (a bare phase name after "al" is navigation, not a keyword
  scan). Current phase design is ahead, so `harness-workflow` BLOCKS the backward
  move and the router surfaces the reason instead of swallowing it. Correct division
  of labor: router picks the phase, workflow enforces legality.
  (Blind-test note: this row exposed the original gap — "spec" was both a phase
  token and a keyword, and the weak model defaulted to ASK. The `explicit-agent`
  navigation clause resolves the overlap deterministically.)

- **#18** "corre el apply" names the phase token `apply` as the thing to run →
  `explicit-agent`. Contrast #13 "corre las pruebas" where "pruebas" is descriptive →
  `keyword`.

## How to score a run

For each row the local model output is CORRECT iff:
1. It emits exactly one echo line `→ Router: archon-<phase>` (or `ASK`).
2. `<phase>` matches Expected target (or it outputs ASK where expected).
3. It does NOT begin executing the phase in the same turn.

Count pass/fail per row. **Rows #1–#4 passing is the headline metric** — those are
the implicit phrasings that fail today. Rows #5, #17, #18 guard against regressions
(explicit still wins; backward jumps still blocked).
