# Delta for SDD Preflight

## Purpose

No `openspec/specs/` capability governs the preflight interaction wording (it is
LLM-interpreted instruction text generated into CLAUDE.md/AGENTS.md). This delta
captures the acceptance criteria for replacing the single Spanish code-block preflight
with five per-group arrow-key questions.

## ADDED Requirements

### Requirement: Per-Group Arrow-Key Questions

The generated CLAUDE.md/AGENTS.md MUST instruct the orchestrator to ask each preflight
group (A Ritmo, B Artefactos, C PRs, D Revisión, E Playwright) as its own separate
arrow-key AskUserQuestion, with the recommended option pre-selected as the default.

The instruction text SHALL NOT contain a fenced code block starting with "Antes de
continuar con SDD" nor reference answer codes such as "A1", "B1", "C1", "D1".

#### Scenario: Five groups asked as separate questions

```gherkin
Scenario: Five groups asked as separate questions
  Given an archon-initialized CLAUDE.md
  When the orchestrator starts an SDD session with no prior preflight decision
  Then the orchestrator asks group A (Ritmo) as an arrow-key question
  And the orchestrator asks group B (Artefactos) as an arrow-key question
  And the orchestrator asks group C (PRs) as an arrow-key question
  And the orchestrator asks group D (Revisión) as an arrow-key question
  And the orchestrator asks group E (Playwright) as an arrow-key question
```

#### Scenario: Recommended option is pre-selected default

```gherkin
Scenario: Recommended option is pre-selected default
  Given an archon-initialized CLAUDE.md
  When the orchestrator renders each preflight group question
  Then group A defaults to "Interactivo"
  And group B defaults to "OpenSpec"
  And group C defaults to "Preguntarme"
  And group D defaults to "400 líneas"
  And group E defaults to "No"
```

### Requirement: Group D "Otro" Free-Text Follow-Up

When the user selects the "Otro" option for group D (Revisión), the orchestrator MUST
ask a free-text follow-up question requesting the custom line-count number before
caching the review budget.

#### Scenario: D "Otro" triggers free-text follow-up

```gherkin
Scenario: D "Otro" triggers free-text follow-up
  Given the orchestrator is asking group D (Revisión)
  When the user selects "Otro"
  Then the orchestrator asks a free-text question for the custom line count
  And the orchestrator caches the entered number as the review budget
```

### Requirement: Group E Always Asked

The orchestrator MUST ask group E (Playwright) for every SDD session regardless of
whether the project type has been determined as web or not-web during exploration.

#### Scenario: Group E asked for a non-web project

```gherkin
Scenario: Group E asked for a non-web project
  Given an archon-initialized CLAUDE.md for a non-web Go project
  When the orchestrator starts an SDD session
  Then the orchestrator asks group E (Playwright) as an arrow-key question
```

#### Scenario: Group E asked for a web project

```gherkin
Scenario: Group E asked for a web project
  Given an archon-initialized CLAUDE.md for a web project
  When the orchestrator starts an SDD session
  Then the orchestrator asks group E (Playwright) as an arrow-key question
```

## REMOVED Requirements

### Requirement: Single-Block Code Prompt with Answer Codes

(Reason: The fenced "Antes de continuar con SDD" code block that instructs users to
reply with codes like "A1, B1, C1, D1" or "usar recomendado" is replaced by five
per-group arrow-key questions. The block is error-prone and less user-friendly than
selectable options.)
(Migration: No migration for callers. The same five groups and options are preserved;
only the asking mechanism changes. Tests asserting the literal block text must be
updated to assert per-group question markers instead.)

### Requirement: Global "usar recomendado" Fast-Path

(Reason: Per-group questions each default to the recommended option, making a global
fast-path redundant. Selecting the default on each question is equivalent.)
(Migration: None. Users get the same result by accepting the default on each of the
five questions.)

### Requirement: Group E Project-Type Conditional

(Reason: The old instruction asked group E only for new/blank/unknown projects. The
new instruction always asks group E, simplifying the flow and removing the conditional
branch from the orchestrator's instructions.)
(Migration: None. The semantic of group E (maps to playwright.enabled) is unchanged;
only the ask conditionality is removed.)
