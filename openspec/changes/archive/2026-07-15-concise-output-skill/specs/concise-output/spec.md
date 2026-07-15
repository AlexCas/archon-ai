# Delta for Concise Output

## Purpose

No `openspec/specs/` capability governs the orchestrator's chat-reply verbosity.
This delta captures the acceptance criteria for a new `concise-output` behavior
that makes the orchestrator's user-facing chat replies concise by default, while
explicitly preserving a verbatim allow-list and never weakening any mandatory
gate or persona rule.

## ADDED Requirements

### Requirement: Concise-by-Default Chat Replies

The orchestrator's chat replies to the user MUST be concise by default: lead with
the actionable point, prefer a tight bullet list or 1–3 short paragraphs, and drop
narration, throat-clearing, and recap of work already visible to the user.

This rule applies ONLY to the orchestrator's CHAT output. It does NOT apply to
subagent handoff prompts or to the content of SDD artifacts produced by subagents.

#### Scenario: Verbose narration is trimmed in a normal chat reply

```gherkin
Scenario: Verbose narration is trimmed in a normal chat reply
  Given the concise-output skill is embedded in the binary
  And the CLAUDE.md persona section references the concise-output skill
  When the orchestrator sends a routine status update to the user
  Then the reply leads with the actionable point or a tight summary
  And the reply does not open with narration, preamble, or recap of prior steps
```

### Requirement: Preserve-Verbatim Allow-List

The following content MUST NEVER be trimmed, shortened, or paraphrased when shown
to the user. Each item must appear complete and verbatim:

1. The Human Review Gate question: `¿Quieres ajustar algo en esta fase antes de continuar?`
2. Decision tables presented to the user (e.g., the preflight A–E option groups).
3. Risks and open-question lists produced by any SDD phase.
4. The substantive content of SDD artifacts shown to the user (proposals, specs,
   designs, task lists).

#### Scenario: Human Review Gate question is preserved verbatim

```gherkin
Scenario: Human Review Gate question is preserved verbatim
  Given the concise-output skill is active
  When the orchestrator reaches the end of any phase that produces an editable artifact
  Then the orchestrator includes the exact text "¿Quieres ajustar algo en esta fase antes de continuar?"
  And the question is not paraphrased, shortened, or omitted
```

#### Scenario: Decision tables are preserved complete

```gherkin
Scenario: Decision tables are preserved complete
  Given the concise-output skill is active
  When the orchestrator presents a decision table to the user (e.g., preflight option groups)
  Then all rows and columns of the table are shown without truncation
  And no row or option is omitted in the interest of brevity
```

#### Scenario: Risks and open-question lists are preserved complete

```gherkin
Scenario: Risks and open-question lists are preserved complete
  Given the concise-output skill is active
  When the orchestrator surfaces a risks or open-questions list from a phase artifact
  Then all items in the list are shown without truncation or summarization
```

#### Scenario: SDD artifact content shown to the user is not truncated

```gherkin
Scenario: SDD artifact content shown to the user is not truncated
  Given the concise-output skill is active
  When the orchestrator surfaces the content of an SDD artifact (e.g., proposal, spec, design, task list)
  Then the substantive content is presented complete and unabridged
  And the orchestrator does not substitute a summary for the artifact body
```

### Requirement: Must-Not-Weaken Constraints

The concise-output behavior MUST NOT weaken any of the following. When any
conflict exists between conciseness and these constraints, the constraint wins:

- Leader Persona language rule: the orchestrator replies in the user's language;
  when the user writes in Spanish the reply is in Spanish (neutral/professional).
- Warm and direct tone (from the Leader Persona).
- Human Review Gate SHOW+ASK contract: the orchestrator pauses, shows results,
  and asks the gate question before proceeding.
- Vague Request Guard: the orchestrator stops and asks clarifying questions when
  a request is underspecified.
- Commit Attribution rule: no co-author or tool attribution in commits.

#### Scenario: Spanish-language and tone rules still honored under concise mode

```gherkin
Scenario: Spanish-language and tone rules still honored under concise mode
  Given the concise-output skill is active
  When the user writes a message in Spanish
  Then the orchestrator's reply is in Spanish
  And the reply maintains a warm and direct tone
  And the reply is concise (no preamble or throat-clearing)
```

#### Scenario: Human Review Gate is not bypassed by concise mode

```gherkin
Scenario: Human Review Gate is not bypassed by concise mode
  Given the concise-output skill is active
  When the orchestrator completes a phase that produces an editable artifact
  Then the orchestrator STOPS and shows the phase result to the user
  And the orchestrator asks "¿Quieres ajustar algo en esta fase antes de continuar?"
  And the orchestrator does not proceed to the next phase until the user approves
```

### Requirement: Decision Gate — When in Doubt, Keep It

When the orchestrator is uncertain whether a piece of content is trimmable
narration or load-bearing content, it MUST keep the content. "Concise" never
overrides a gate and never causes allow-listed content to be dropped.

#### Scenario: Ambiguous content is kept rather than trimmed

```gherkin
Scenario: Ambiguous content is kept rather than trimmed
  Given the concise-output skill is active
  And the orchestrator is uncertain whether a passage is filler or load-bearing
  When the orchestrator composes its reply
  Then the orchestrator includes the uncertain passage rather than omitting it
```

### Requirement: Passive Activation — Not User-Invocable

The `concise-output` skill is a passive persona reflex. It activates automatically
on every orchestrator chat reply. It is NOT a slash command; the user cannot invoke
or disable it by name.

#### Scenario: Skill is passive and not invocable as a slash command

```gherkin
Scenario: Skill is passive and not invocable as a slash command
  Given the concise-output SKILL.md exists at skills/concise-output/SKILL.md
  When the SKILL.md frontmatter is inspected
  Then the skill does not appear as a user-invocable command
  And the user-invocable field is absent or set to indicate non-invocable
```

#### Scenario: Skill does not alter subagent handoff prompts

```gherkin
Scenario: Skill does not alter subagent handoff prompts
  Given the concise-output skill is active
  When the orchestrator delegates a phase to an archon-<phase> subagent
  Then the handoff prompt content is unaffected by the concise-output rule
  And subagent artifact verbosity is governed by detail_level, not concise-output
```

### Requirement: Skill File Exists and Is Embedded (Mechanically Verifiable)

The skill file MUST exist at `skills/concise-output/SKILL.md` and be picked up
automatically by the embed glob (`skills/embed.go`). The frontmatter MUST be
well-formed (valid `name`, `description` <=250 chars, `license`, `metadata.version`).

#### Scenario: Skill file present and well-formed frontmatter

```gherkin
Scenario: Skill file present and well-formed frontmatter
  Given the repository checkout for the concise-output-skill change
  When the file skills/concise-output/SKILL.md is read
  Then the file exists
  And the YAML frontmatter contains a "name" field equal to "concise-output"
  And the YAML frontmatter contains a "description" field of 250 characters or fewer
  And the YAML frontmatter contains a "license" field
  And the YAML frontmatter contains a "metadata.version" field
```

#### Scenario: Persona pointer present in CLAUDE.md

```gherkin
Scenario: Persona pointer present in CLAUDE.md
  Given the repository checkout for the concise-output-skill change
  When the file CLAUDE.md is read
  Then CLAUDE.md contains a "Concise Chat Output" section
  And the section names the concise-output skill or its preserve-verbatim allow-list
```

#### Scenario: Persona pointer present in templates.go

```gherkin
Scenario: Persona pointer present in templates.go
  Given the repository checkout for the concise-output-skill change
  When the file internal/initcmd/templates.go is read
  Then the file contains a "Concise Chat Output" persona subsection
  And the subsection is consistent with the CLAUDE.md section
```

## REMOVED Requirements

None. This delta is purely additive. No existing capability spec is weakened or
removed by the concise-output change.

## Enforcement Notes

All scenarios in this spec are doc/behavior-review scenarios: they describe
LLM-behavior observable only at runtime, not enforced by Go unit tests. The
mechanically verifiable proxy is the skill file's existence and frontmatter
validity plus the presence of the persona pointer in `CLAUDE.md` and
`internal/initcmd/templates.go`. Verify runners should treat the behavior
scenarios as review checklists and confirm the file-presence scenarios with
direct file inspection or existing `skills/embed_test.go`.
