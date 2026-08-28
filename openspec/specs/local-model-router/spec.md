# local-model-router Specification

## Purpose

Specify the behavior of the hybrid phase-dispatch router: a deterministic code
pre-router (`archon route`) plus a model classifier (`skills/sdd-router/SKILL.md`)
that together resolve which SDD phase to target from a natural-language user message.
This spec is the contract the verify phase tests against.

The router NEVER executes a phase, NEVER writes state, and NEVER bypasses preflight,
vague-request, or human-review gates. It resolves one thing: `target_phase | ASK`.

---

## Decisions on Open Questions

### D1 — Handoff channel: `--json` + human echo line

`archon route` MUST emit a machine-readable JSON object on stdout as its primary
output, followed by a human echo line on stderr (or a second stdout line prefixed
`#`). The JSON schema is:

```json
{
  "phase": "<phase-name | ASK>",
  "rule": "<rule-id>",
  "path": "<code | model>",
  "active_change": "<change-name | none>"
}
```

Rationale: weak leaders parsing prose drop the `archon-` prefix (observed in
FINDINGS.md). A structured field is unambiguous. The human line
`→ Router: archon-<phase>  (rule: <rule-id>, active-change: <name|none>)` echoes the
same data for readability in interactive sessions. When `phase` is `ASK`, the `rule`
field contains the `ask` or `ambiguous` id and the leader invokes the model
classifier.

### D2 — Active-change discovery: router discovers, flags override

The router MUST discover the active change autonomously using this precedence:

1. `--change <name>` CLI flag (explicit override, highest priority).
2. `SESSION_STATUS.md` at repo root: parse `Active change:` field.
3. Single non-archive folder under `openspec/changes/` (fallback).
4. `none` — no active change (bootstrap state).

The `--phase <phase>` and `--status <status>` CLI flags MAY override the values read
from `state.yaml` for testability (e.g., `archon route --phase propose --status
completed "Continuemos"`).

Rationale: router-discovers keeps the SESSION_STATUS precedence in one place (code),
and lets the leader pass the message without needing to resolve active-change itself.
Optional flag overrides keep unit tests deterministic without requiring filesystem
fixtures.

### D3 — Residual #15 dual-action: add one narrow code rule

When `MSG` matches BOTH a judge-verb (`revisa|review|juzga|dictamen`) AND a
verify-verb (`prueba|pruebas|test|valida|verify`) with a coordinating conjunction
(`y|and|e`) present in the same clause, the code pre-router MUST output `ASK` with
rule id `ambiguous` BEFORE falling through to the model classifier.

Rationale: this is one narrow, stable rule (two verb sets + conjunction), not a
growing list. The model alone returned `judge` instead of `ASK` for fixture #15
across all runs (FINDINGS.md); the gate backstop is not sufficient because the human
sees a wrong echo before the gate fires. The rule is testable, non-whack-a-mole, and
the scope is limited to judge+verify overlap.

---

## Requirements

### Requirement: Deterministic Code Pre-router

The system MUST implement a deterministic code resolver in `internal/route/route.go`
that applies four rules in strict top-to-bottom, first-match-wins order:

1. `explicit-agent`: message names an agent token or navigation target directly.
2. `control`: message contains a control-flow word (`siguiente|continuemos|…`).
3. `implicit`: message contains a start/continue verb without explicit agent.
4. `keyword`: message matches exactly one phase's keyword set.

The resolver MUST NOT invoke any LLM or external service.

#### Scenario: Explicit agent token wins over implicit verb

```gherkin
Scenario: Explicit agent wins over implicit verb
  Given STATE is none
  And MSG is "Hagamos esta especificacion. Lanza el agente de exploracion"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "explore"
  And "rule": "explicit-agent"
  And "path": "code"
```

#### Scenario: Control word next resolves successor phase

```gherkin
Scenario: Control word resolves next phase from completed state
  Given STATE has phase "propose" and status "completed"
  And MSG is "Continuemos"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "spec"
  And "rule": "next"
  And "path": "code"
```

#### Scenario: Control word resumes in-progress phase

```gherkin
Scenario: Control word resumes in-progress phase
  Given STATE has phase "design" and status "in_progress"
  And MSG is "Continuemos"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "design"
  And "rule": "resume"
  And "path": "code"
```

#### Scenario: Implicit start verb with no active change starts explore

```gherkin
Scenario: Implicit start verb with no active change starts explore
  Given STATE is none
  And MSG is "Trabajemos en esta especificacion"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "explore"
  And "rule": "implicit-start"
  And "path": "code"
```

#### Scenario: Implicit start verb resumes current phase when change is active

```gherkin
Scenario: Implicit start verb resumes current phase when change is active
  Given STATE has phase "spec" and status "in_progress"
  And MSG is "Trabajemos en esta especificacion"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "spec"
  And "rule": "implicit-resume"
  And "path": "code"
```

#### Scenario: Single-keyword match resolves phase

```gherkin
Scenario: Single-keyword match resolves phase
  Given STATE has phase "spec" and status "completed"
  And MSG is "Disenemos la arquitectura del API"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "design"
  And "rule": "keyword"
  And "path": "code"
```

#### Scenario: Literal phase token in navigation position is explicit-agent

```gherkin
Scenario: Literal phase token in navigation position is explicit-agent
  Given STATE has phase "tasks" and status "completed"
  And MSG is "corre el apply"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "apply"
  And "rule": "explicit-agent"
  And "path": "code"
```

---

### Requirement: Implicit-above-keyword Precedence

The system MUST place the `implicit` rule ABOVE the `keyword` rule so that a
start/continue verb with a loose phase noun routes to resume/start rather than jumping
to the named phase.

#### Scenario: Implicit verb with phase noun does not trigger keyword

```gherkin
Scenario: Implicit verb with loose phase noun does not trigger keyword
  Given STATE is none
  And MSG is "Trabajemos en esta especificacion"
  When the code pre-router processes MSG
  Then the output JSON has "rule": "implicit-start"
  And NOT "rule": "keyword"
```

---

### Requirement: Narrow Dual-Action Ambiguity Rule (D3)

When `MSG` contains both a judge-verb and a verify-verb connected by a coordinating
conjunction, the code pre-router MUST output `ASK` with rule id `ambiguous` without
consulting the model classifier.

Judge-verbs: `revisa`, `review`, `juzga`, `dictamen`.
Verify-verbs: `prueba`, `pruebas`, `test`, `valida`, `verify`.
Coordinating conjunctions: `y`, `and`, `e`.

#### Scenario: Dual-action judge+verify produces ASK

```gherkin
Scenario: Dual-action judge-verb and verify-verb produces ASK
  Given STATE has phase "verify" and status "in_progress"
  And MSG is "Revisa y prueba esto"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "ASK"
  And "rule": "ambiguous"
  And "path": "code"
```

---

### Requirement: Model Classifier Fallback

When no deterministic rule fires, the system MUST emit `{"phase":"CLASSIFY","rule":"classify","path":"model","active_change":"<name|none>"}` to signal the leader to invoke the in-context model classifier (`skills/sdd-router/SKILL.md`). The model classifier MUST output `phase | ASK` per its keyword table.

#### Scenario: No deterministic rule falls through to model classifier signal

```gherkin
Scenario: No rule match emits CLASSIFY signal
  Given STATE has phase "spec" and status "in_progress"
  And MSG is "Que opinas del clima?"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "CLASSIFY"
  And "rule": "classify"
  And "path": "model"
```

#### Scenario: Model classifier returns ASK for unclassifiable message

```gherkin
Scenario: Model classifier returns ASK for non-SDD message
  Given the code pre-router emitted "CLASSIFY" for MSG "Que opinas del clima?"
  When the leader invokes the model classifier
  Then the model classifier output is "ASK"
```

---

### Requirement: Backward Navigation Resolves but Does Not Bypass Workflow Gate

The code pre-router MUST resolve backward navigation (`explicit-agent` to a prior
phase) and emit the JSON with the resolved phase. Legality enforcement is delegated
to `harness-workflow`; the router MUST NOT pre-judge legality.

#### Scenario: Backward navigation resolves then harness-workflow blocks

```gherkin
Scenario: Backward navigation resolved by router, blocked by harness-workflow
  Given STATE has phase "design" and status "completed"
  And MSG is "Volvamos al spec"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "spec"
  And "rule": "explicit-agent"
  And "path": "code"
  When the leader passes "spec" to harness-workflow
  Then harness-workflow returns "blocked" with reason "backward transitions not allowed"
```

---

### Requirement: Bootstrap State — No Active Change

When no active change exists (`STATE = none`), the only legal start is explore.
Control words and implicit verbs MUST route to explore. The model classifier keyword
table applies normally (e.g., "explora el codigo de billing" → explore).

#### Scenario: Control word with no state targets explore

```gherkin
Scenario: Control word with no active change targets explore
  Given STATE is none
  And MSG is "Adelante"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "explore"
  And "rule": "next-nochange"
  And "path": "code"
```

#### Scenario: Implicit start verb with no state targets explore (alt)

```gherkin
Scenario: Implicit start verb empecemos with no active change
  Given STATE is none
  And MSG is "Empecemos con esto"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "explore"
  And "rule": "implicit-start"
  And "path": "code"
```

#### Scenario: Implicit verb hagamos with no state targets explore

```gherkin
Scenario: Implicit verb hagamos with no active change
  Given STATE is none
  And MSG is "Hagamos esta feature"
  When the code pre-router processes MSG
  Then the output JSON has "phase": "explore"
  And "rule": "implicit-start"
  And "path": "code"
```

---

### Requirement: Machine-Readable JSON Output Contract

`archon route [--json] [--change <name>] [--phase <phase>] [--status <status>] "<message>"` MUST write valid JSON to stdout. Schema (all fields required):

| Field | Type | Values |
|-------|------|--------|
| `phase` | string | phase name, `ASK`, or `CLASSIFY` |
| `rule` | string | `explicit-agent`, `next`, `resume`, `next-nochange`, `implicit-start`, `implicit-resume`, `keyword`, `ambiguous`, `ask`, `classify` |
| `path` | string | `code` or `model` |
| `active_change` | string | change folder name or `none` |

Exit codes: `0` = resolved (phase or ASK/CLASSIFY); `1` = internal error (e.g., corrupt state.yaml, filesystem failure). The router MUST NOT exit non-zero for ASK or CLASSIFY — those are valid outputs.

#### Scenario: JSON output on successful route resolution

```gherkin
Scenario: JSON output schema on resolved route
  Given STATE has phase "apply" and status "completed"
  And MSG is "Siguiente"
  When archon route is invoked
  Then stdout is valid JSON
  And the JSON contains "phase", "rule", "path", "active_change"
  And "phase" is "verify"
  And exit code is 0
```

#### Scenario: JSON output for ASK exits zero

```gherkin
Scenario: ASK output exits with code 0
  Given STATE has phase "spec" and status "in_progress"
  And MSG is "Que opinas del clima?"
  When archon route is invoked
  Then the JSON "phase" is "CLASSIFY"
  And exit code is 0
```

---

### Requirement: Text Normalization

The system MUST normalize `MSG` before matching: lowercase, strip diacritics (e.g.,
`especificación` → `especificacion`, `diseño` → `diseno`). Normalization MUST occur
before any rule evaluation.

#### Scenario: Accented keyword matches normalized form

```gherkin
Scenario: Accented keyword matches its normalized equivalent
  Given STATE is none
  And MSG is "Trabajemos en esta especificación"
  When the code pre-router processes MSG
  Then the output JSON has "rule": "implicit-start"
```

---

### Requirement: State is Read-Only for the Router

The router MUST NOT write to `state.yaml`, `SESSION_STATUS.md`, or any SDD state
file. State writes remain exclusively owned by `harness-workflow`.

#### Scenario: Router does not modify state.yaml

```gherkin
Scenario: Router leaves state.yaml unchanged
  Given STATE has phase "spec" and status "in_progress"
  When archon route processes any message
  Then state.yaml is identical before and after the invocation
```

---

### Requirement: PhaseOrder Canonical Source

The router MUST import `internal/config.PhaseOrder` as the sole definition of phase
sequence. No second phase-order list is permitted in `internal/route`.

#### Scenario: Successor resolution uses config.PhaseOrder

```gherkin
Scenario: Next-phase resolution uses config.PhaseOrder not a local list
  Given config.PhaseOrder is ["explore","propose","spec","design","tasks","apply","verify","judge","archive"]
  And STATE has phase "propose" and status "completed"
  And MSG is "Continuemos"
  When the code pre-router resolves the successor
  Then the resolved phase is "spec"
  And no second PhaseOrder definition exists in internal/route
```

---

### Requirement: Active-Change Discovery Precedence

Active-change resolution MUST follow the order: `--change` flag > `SESSION_STATUS.md`
field `Active change:` > sole non-archive folder under `openspec/changes/` > `none`.

#### Scenario: CLI flag overrides SESSION_STATUS.md

```gherkin
Scenario: Explicit --change flag takes priority over SESSION_STATUS.md
  Given SESSION_STATUS.md names active change "foo"
  And archon route is invoked with "--change bar"
  When the router resolves active change
  Then active_change in JSON is "bar"
```

#### Scenario: SESSION_STATUS.md used when no flag provided

```gherkin
Scenario: SESSION_STATUS.md resolution when no flag provided
  Given SESSION_STATUS.md names active change "local-model-router"
  And no --change flag is passed
  When the router resolves active change
  Then active_change in JSON is "local-model-router"
```

---

### Requirement: Keyword-to-Phase Table Coverage

The router MUST cover all nine SDD phases in the keyword table with both Spanish and
English keywords as specified in `ROUTER.md`. No phase is left without at least two
Spanish and two English keywords.

#### Scenario: Keyword coverage for each phase (representative)

```gherkin
Scenario Outline: Keyword resolves each phase
  Given STATE has no conflicting active rule (no implicit verb, no control word)
  And MSG is <message>
  When the code pre-router evaluates MSG
  Then the output JSON has "phase": <phase> and "rule": "keyword"

  Examples:
    | message                      | phase   |
    | Explora el codigo de billing  | explore |
    | Implementa las tareas         | apply   |
    | Corre las pruebas             | verify  |
    | Archiva el cambio             | archive |
```

---

### Requirement: sdd-router Skill — Model Classifier Contract

`skills/sdd-router/SKILL.md` MUST document:
1. When the leader invokes it (only when `archon route` outputs `CLASSIFY`).
2. The keyword table (Spanish + English, all nine phases).
3. The output contract: one echo line `→ Router: archon-<phase>` or `ASK`.
4. That it MUST NOT start executing a phase in the same turn.

The skill MUST be provider-neutral (no Claude-specific delegation primitives).

#### Scenario: Model classifier does not execute phase in same turn

```gherkin
Scenario: Model classifier outputs target then stops
  Given the leader receives CLASSIFY signal from archon route
  When the leader invokes the model classifier for MSG "Disenemos el API"
  Then the model classifier emits exactly one echo line
  And does not begin executing a design phase task
```

---

### Requirement: Leader Wiring (orchestratorRules)

`internal/initcmd/templates.go` MUST add a routing step to both
`orchestratorRulesClaude` and `orchestratorRulesOpencode`: "Before delegating a
phase, run `archon route '<message>'` and use its resolved phase; invoke the model
classifier when output is `CLASSIFY`; surface ASK to the user." The routing step MUST
be inserted so that existing rules (harness-workflow gate, delegate) remain intact and
downstream from it.

#### Scenario: Generated CLAUDE.md includes routing instruction

```gherkin
Scenario: archon init generates CLAUDE.md with routing rule
  Given archon init is run with default config
  When the generated CLAUDE.md is read
  Then it contains the text "archon route"
  And the routing rule precedes the harness-workflow delegation rule
```
