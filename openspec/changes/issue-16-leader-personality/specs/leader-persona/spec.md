# leader-persona Specification

## Purpose

Embeds persona, language, tone, and behavioral rules into generated orchestrator templates (AGENTS.md/CLAUDE.md) so the Leader agent consistently matches the user's language, uses neutral Spanish, communicates warmly, and provides constructive guidance in chat.

## Requirements

### Requirement: Persona Section Placement

The rendered template MUST include a `## Leader Persona` section placed before `## Phase Order`.

#### Scenario: Persona appears before workflow rules in both templates

- GIVEN `RenderAgentsMD()` or `RenderClaudeMD()` is called with valid TemplateData
- WHEN the output is rendered
- THEN `## Leader Persona` appears before `## Phase Order`

### Requirement: Persona Content Coverage

The persona section MUST cover scope, language, tone, and behavior with explicit directives.

#### Scenario: All four domains are present

- GIVEN the rendered template
- THEN scope is stated: persona governs ONLY chat replies; technical artifacts default to English
- AND Language includes: MUST reply in user's language; when Spanish, use neutral Spanish, do NOT use voseo
- AND Tone includes: warm and direct; avoid ALL-CAPS
- AND Behavior includes: seek clarification before acting; never evade with "I didn't do this because you didn't ask me to"

### Requirement: Template Integrity After Persona Insertion

Adding the persona section MUST NOT corrupt existing template structure or rendering.

#### Scenario: Both templates remain byte-identical

- GIVEN identical TemplateData for `RenderAgentsMD` and `RenderClaudeMD`
- WHEN both are rendered
- THEN the outputs are identical (shared `orchestratorSections`)

#### Scenario: No § placeholder artifacts remain

- GIVEN the rendered template
- THEN the `§` character does NOT appear anywhere in the output

#### Scenario: Exactly five rules preserved

- GIVEN the rendered template
- THEN exactly five numbered rules follow `## Rules`
- AND no sixth rule exists

#### Scenario: Existing workflow sections intact

- GIVEN the rendered template
- THEN it contains `## Phase Order`, `## SDD Session Preflight (HARD GATE)`, `## Vague Request Guard (MANDATORY)`, and `## Human Review Gate (MANDATORY)`
