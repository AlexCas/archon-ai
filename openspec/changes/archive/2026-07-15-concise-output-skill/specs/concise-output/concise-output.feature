Feature: Concise-by-Default Orchestrator Chat Replies
  The ARCHON orchestrator MUST apply a concise-by-default reflex to its chat
  replies to the user, trimming narration and filler while never truncating
  the verbatim allow-list (Human Review Gate question, decision tables, risks
  lists, SDD artifact content) and never weakening any mandatory gate or
  Leader Persona rule.

  Background:
    Given the concise-output skill is embedded in the binary
    And the CLAUDE.md persona section references the concise-output skill

  # ── Concise-by-default behavior ──────────────────────────────────────────

  @happy
  Scenario: Verbose narration is trimmed in a normal chat reply
    Given the orchestrator sends a routine status update to the user
    When the reply is composed
    Then the reply leads with the actionable point or a tight summary
    And the reply does not open with narration, preamble, or recap of prior steps

  # ── Preserve-verbatim allow-list ─────────────────────────────────────────

  @happy
  Scenario: Human Review Gate question is preserved verbatim
    Given the orchestrator completes a phase that produces an editable artifact
    When the orchestrator composes its phase-end message
    Then the reply includes the exact text "¿Quieres ajustar algo en esta fase antes de continuar?"
    And the question is not paraphrased, shortened, or omitted

  @happy
  Scenario: Decision tables are preserved complete
    Given the orchestrator presents a decision table to the user
    When the reply is composed
    Then all rows and columns of the table appear without truncation
    And no row or option is omitted in the interest of brevity

  @happy
  Scenario: Risks and open-question lists are preserved complete
    Given the orchestrator surfaces a risks or open-questions list from a phase artifact
    When the reply is composed
    Then all items in the list are shown without truncation or summarization

  @happy
  Scenario: SDD artifact content shown to the user is not truncated
    Given the orchestrator surfaces the content of an SDD artifact (e.g., proposal, spec, design, task list)
    When the reply is composed
    Then the substantive content is presented complete and unabridged
    And the orchestrator does not substitute a summary for the artifact body

  # ── Must-not-weaken constraints ───────────────────────────────────────────

  @happy
  Scenario: Spanish-language and tone rules still honored under concise mode
    Given the user writes a message in Spanish
    When the orchestrator composes its reply under concise mode
    Then the reply is in Spanish
    And the reply maintains a warm and direct tone
    And the reply is concise (no preamble or throat-clearing)

  @error
  Scenario: Human Review Gate is not bypassed by concise mode
    Given the orchestrator completes a phase that produces an editable artifact
    When the orchestrator is under concise mode
    Then the orchestrator STOPS and shows the phase result to the user
    And the orchestrator asks "¿Quieres ajustar algo en esta fase antes de continuar?"
    And the orchestrator does not proceed to the next phase until the user approves

  # ── Decision gate: when in doubt, keep it ────────────────────────────────

  @edge
  Scenario: Ambiguous content is kept rather than trimmed
    Given the orchestrator is uncertain whether a passage is filler or load-bearing
    When the orchestrator composes its reply
    Then the orchestrator includes the uncertain passage rather than omitting it

  # ── Passive activation: not user-invocable ───────────────────────────────

  @happy
  Scenario: Skill is passive and not invocable as a slash command
    Given the file skills/concise-output/SKILL.md exists
    When the SKILL.md frontmatter is inspected
    Then the skill does not appear as a user-invocable slash command
    And the frontmatter has no user-invocable field or marks the skill as non-invocable

  @happy
  Scenario: Skill does not alter subagent handoff prompts
    Given the concise-output skill is active
    When the orchestrator delegates a phase to an archon-<phase> subagent
    Then the handoff prompt content is unaffected by the concise-output rule
    And subagent artifact verbosity is governed by detail_level, not concise-output

  # ── Mechanically verifiable: file existence and frontmatter ──────────────

  @happy
  Scenario: Skill file exists and has well-formed frontmatter
    Given the repository checkout for the concise-output-skill change
    When the file skills/concise-output/SKILL.md is read
    Then the file exists
    And the YAML frontmatter contains a "name" field equal to "concise-output"
    And the YAML frontmatter contains a "description" field of 250 characters or fewer
    And the YAML frontmatter contains a "license" field
    And the YAML frontmatter contains a "metadata.version" field

  @happy
  Scenario: Persona pointer present in CLAUDE.md
    Given the repository checkout for the concise-output-skill change
    When the file CLAUDE.md is read
    Then CLAUDE.md contains a "Concise Chat Output" section
    And the section names the concise-output skill or its preserve-verbatim allow-list

  @happy
  Scenario: Persona pointer present in templates.go
    Given the repository checkout for the concise-output-skill change
    When the file internal/initcmd/templates.go is read
    Then the file contains a "Concise Chat Output" persona subsection
    And the subsection is consistent with the CLAUDE.md section
