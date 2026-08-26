Feature: local-model-router — hybrid deterministic + model phase dispatch
  The hybrid router resolves which SDD phase to target from a natural-language user
  message. A deterministic code pre-router handles state-dependent cases; the model
  classifier handles fuzzy categorization. The router never executes a phase, never
  writes state, and never bypasses SDD gates.

  # ────────────────────────────────────────────────────────────────────
  # Background: text normalization applied before all rule evaluation
  # ────────────────────────────────────────────────────────────────────

  Background:
    Given the router normalizes MSG to lowercase with diacritics stripped

  # ──────────────────────────────────────────────
  # Fixture #1 — Implicit start, no active change
  # ──────────────────────────────────────────────

  @happy @fixture-1
  Scenario: Implicit start verb with no active change starts explore (trabajemos)
    Given STATE is none
    And MSG is "Trabajemos en esta especificacion"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "explore"
    And "rule": "implicit-start"
    And "path": "code"
    And "active_change": "none"

  # ──────────────────────────────────────────────
  # Fixture #2 — Implicit resume, spec in progress
  # ──────────────────────────────────────────────

  @happy @fixture-2
  Scenario: Implicit start verb resumes current phase when spec is in_progress
    Given STATE has phase "spec" and status "in_progress"
    And MSG is "Trabajemos en esta especificacion"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "spec"
    And "rule": "implicit-resume"
    And "path": "code"

  # ──────────────────────────────────────────────────────
  # Fixture #3 — Implicit start (empecemos), no active change
  # ──────────────────────────────────────────────────────

  @happy @fixture-3
  Scenario: Implicit start verb empecemos with no active change targets explore
    Given STATE is none
    And MSG is "Empecemos con esto"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "explore"
    And "rule": "implicit-start"
    And "path": "code"

  # ─────────────────────────────────────────────────────────
  # Fixture #4 — Implicit start (hagamos), no active change
  # ─────────────────────────────────────────────────────────

  @happy @fixture-4
  Scenario: Implicit verb hagamos with no active change targets explore
    Given STATE is none
    And MSG is "Hagamos esta feature"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "explore"
    And "rule": "implicit-start"
    And "path": "code"

  # ────────────────────────────────────────────────────────────────────────────
  # Fixture #5 — Explicit agent wins over implicit verb (regression guard)
  # ────────────────────────────────────────────────────────────────────────────

  @happy @fixture-5 @regression
  Scenario: Explicit agent token wins over implicit verb co-present in message
    Given STATE is none
    And MSG is "Hagamos esta especificacion. Lanza el agente de exploracion"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "explore"
    And "rule": "explicit-agent"
    And "path": "code"

  # ──────────────────────────────────────────────────────────────────────────
  # Fixture #6 — Control word: next from propose/completed → spec
  # ──────────────────────────────────────────────────────────────────────────

  @happy @fixture-6
  Scenario: Control word continuemos resolves next phase from propose completed
    Given STATE has phase "propose" and status "completed"
    And MSG is "Continuemos"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "spec"
    And "rule": "next"
    And "path": "code"

  # ────────────────────────────────────────────────────────────────────
  # Fixture #7 — Control word: next from apply/completed → verify
  # ────────────────────────────────────────────────────────────────────

  @happy @fixture-7
  Scenario: Control word siguiente resolves next phase from apply completed
    Given STATE has phase "apply" and status "completed"
    And MSG is "Siguiente"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "verify"
    And "rule": "next"
    And "path": "code"

  # ────────────────────────────────────────────────────────────────────
  # Fixture #8 — Control word: resume design/in_progress
  # ────────────────────────────────────────────────────────────────────

  @happy @fixture-8
  Scenario: Control word continuemos resumes design when in_progress
    Given STATE has phase "design" and status "in_progress"
    And MSG is "Continuemos"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "design"
    And "rule": "resume"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────────
  # Fixture #9 — Control word: next-nochange when STATE is none
  # ─────────────────────────────────────────────────────────────────────

  @happy @fixture-9
  Scenario: Control word adelante with no active change targets explore
    Given STATE is none
    And MSG is "Adelante"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "explore"
    And "rule": "next-nochange"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────
  # Fixture #10 — Keyword: explora → explore
  # ─────────────────────────────────────────────────────────────────

  @happy @fixture-10
  Scenario: Keyword explora resolves explore phase
    Given STATE is none
    And MSG is "Explora el codigo de billing"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "explore"
    And "rule": "keyword"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────
  # Fixture #11 — Keyword: disenemos → design (not implicit)
  # ─────────────────────────────────────────────────────────────────

  @happy @fixture-11
  Scenario: Keyword disenemos resolves design (not in implicit verb list)
    Given STATE has phase "spec" and status "completed"
    And MSG is "Disenemos la arquitectura del API"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "design"
    And "rule": "keyword"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────
  # Fixture #12 — Keyword: implementa → apply (action-verb beats noun)
  # ─────────────────────────────────────────────────────────────────

  @happy @fixture-12
  Scenario: Keyword implementa resolves apply not tasks
    Given STATE has phase "tasks" and status "completed"
    And MSG is "Implementa las tareas"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "apply"
    And "rule": "keyword"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────
  # Fixture #13 — Keyword: corre las pruebas → verify
  # ─────────────────────────────────────────────────────────────────

  @happy @fixture-13
  Scenario: Keyword corre las pruebas resolves verify
    Given STATE has phase "apply" and status "completed"
    And MSG is "Corre las pruebas"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "verify"
    And "rule": "keyword"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────
  # Fixture #14 — Keyword: archiva → archive
  # ─────────────────────────────────────────────────────────────────

  @happy @fixture-14
  Scenario: Keyword archiva resolves archive
    Given STATE has phase "judge" and status "completed"
    And MSG is "Archiva el cambio"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "archive"
    And "rule": "keyword"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────────────
  # Fixture #15 — Dual-action: revisa y prueba → ASK (D3 narrow code rule)
  # ─────────────────────────────────────────────────────────────────────────

  @edge @fixture-15
  Scenario: Dual-action judge-verb and verify-verb connected by conjunction produces ASK
    Given STATE has phase "verify" and status "in_progress"
    And MSG is "Revisa y prueba esto"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "ASK"
    And "rule": "ambiguous"
    And "path": "code"

  # ─────────────────────────────────────────────────────────────────────────
  # Fixture #16 — Unclassifiable: falls through to model → CLASSIFY
  # ─────────────────────────────────────────────────────────────────────────

  @edge @fixture-16
  Scenario: Non-SDD message falls through to CLASSIFY signal
    Given STATE has phase "spec" and status "in_progress"
    And MSG is "Que opinas del clima?"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "CLASSIFY"
    And "rule": "classify"
    And "path": "model"

  @edge @fixture-16b
  Scenario: Model classifier returns ASK for non-SDD message
    Given the code pre-router emitted CLASSIFY for MSG "Que opinas del clima?"
    When the leader invokes the model classifier
    Then the model classifier output is "ASK"
    And the leader does not begin executing any SDD phase

  # ─────────────────────────────────────────────────────────────────────────
  # Fixture #17 — Backward nav: volvamos al spec → explicit-agent then BLOCKED
  # ─────────────────────────────────────────────────────────────────────────

  @edge @fixture-17 @regression
  Scenario: Backward navigation resolves then harness-workflow blocks
    Given STATE has phase "design" and status "completed"
    And MSG is "Volvamos al spec"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "spec"
    And "rule": "explicit-agent"
    And "path": "code"
    When the leader passes "spec" to harness-workflow
    Then harness-workflow returns "blocked"
    And the block reason includes "backward transitions not allowed"

  # ────────────────────────────────────────────────────────────────────────
  # Fixture #18 — Literal phase token "corre el apply" → explicit-agent
  # ────────────────────────────────────────────────────────────────────────

  @happy @fixture-18 @regression
  Scenario: Literal phase token apply in imperative position is explicit-agent
    Given STATE has phase "tasks" and status "completed"
    And MSG is "corre el apply"
    When the code pre-router processes MSG
    Then the output JSON has "phase": "apply"
    And "rule": "explicit-agent"
    And "path": "code"

  # ─────────────────────────────────────────────
  # Output contract and exit codes
  # ─────────────────────────────────────────────

  @happy
  Scenario: JSON output schema on resolved route
    Given STATE has phase "apply" and status "completed"
    And MSG is "Siguiente"
    When archon route is invoked
    Then stdout is valid JSON
    And the JSON contains fields "phase", "rule", "path", "active_change"
    And exit code is 0

  @edge
  Scenario: CLASSIFY output exits with code 0 not an error
    Given STATE has phase "spec" and status "in_progress"
    And MSG is "Que opinas del clima?"
    When archon route is invoked
    Then the JSON "phase" is "CLASSIFY"
    And exit code is 0

  # ─────────────────────────────────────────────
  # State immutability
  # ─────────────────────────────────────────────

  @edge
  Scenario: Router does not modify state.yaml
    Given STATE has phase "spec" and status "in_progress"
    When archon route processes any message
    Then state.yaml content is identical before and after the call

  # ─────────────────────────────────────────────
  # Active-change discovery precedence
  # ─────────────────────────────────────────────

  @happy
  Scenario: CLI --change flag overrides SESSION_STATUS.md
    Given SESSION_STATUS.md names active change "foo"
    And archon route is invoked with "--change bar"
    When the router resolves active change
    Then active_change in JSON is "bar"

  @happy
  Scenario: SESSION_STATUS.md used when no --change flag provided
    Given SESSION_STATUS.md names active change "local-model-router"
    And no --change flag is passed
    When the router resolves active change
    Then active_change in JSON is "local-model-router"

  # ──────────────────────────────────────────
  # Text normalization
  # ──────────────────────────────────────────

  @edge
  Scenario: Accented keyword matches normalized form
    Given STATE is none
    And MSG is "Trabajemos en esta especificación"
    When the code pre-router processes MSG
    Then the output JSON has "rule": "implicit-start"

  # ──────────────────────────────────────────────────
  # Precedence: implicit above keyword
  # ──────────────────────────────────────────────────

  @edge
  Scenario: Implicit verb with loose phase noun does not trigger keyword rule
    Given STATE is none
    And MSG is "Trabajemos en esta especificacion"
    When the code pre-router processes MSG
    Then "rule" is "implicit-start"
    And "rule" is NOT "keyword"

  # ──────────────────────────────────────────────────────────────────────
  # PhaseOrder: single canonical source
  # ──────────────────────────────────────────────────────────────────────

  @happy
  Scenario: Next-phase resolution uses config.PhaseOrder only
    Given config.PhaseOrder is the canonical 9-phase list
    And STATE has phase "propose" and status "completed"
    And MSG is "Continuemos"
    When the code pre-router resolves the successor
    Then the resolved phase is "spec"

  # ──────────────────────────────────────────────────────────────────────────
  # Leader wiring: generated CLAUDE.md includes routing instruction
  # ──────────────────────────────────────────────────────────────────────────

  @happy
  Scenario: archon init generates CLAUDE.md with archon route instruction
    Given archon init is run with default config
    When the generated CLAUDE.md is read
    Then it contains the text "archon route"
    And the routing rule appears before the harness-workflow delegation rule

  # ──────────────────────────────────────────────────────────────────────────────────
  # Model classifier: does not execute phase in same turn
  # ──────────────────────────────────────────────────────────────────────────────────

  @edge
  Scenario: Model classifier emits one echo line and does not begin phase execution
    Given the leader receives CLASSIFY signal from archon route
    When the leader invokes the model classifier for MSG "Disenemos el API"
    Then the model classifier emits exactly one echo line
    And does not begin executing a design phase task

  # ────────────────────────────────────────────────────────────────────
  # Keyword table: all nine phases covered (data-driven)
  # ────────────────────────────────────────────────────────────────────

  @happy
  Scenario Outline: Keyword table covers all nine SDD phases
    Given STATE has no active implicit verb or control word in MSG
    And MSG is <message>
    When the code pre-router evaluates MSG
    Then the output JSON has "phase": <phase> and "rule": "keyword"

    Examples:
      | message                              | phase   |
      | Explora el codigo de billing         | explore |
      | Propongamos el enfoque               | propose |
      | Escribe los requisitos gherkin       | spec    |
      | Disenemos la arquitectura            | design  |
      | Haz el desglose de tareas            | tasks   |
      | Implementa las tareas                | apply   |
      | Corre las pruebas                    | verify  |
      | Juzga el codigo                      | judge   |
      | Archiva el cambio                    | archive |
