# Delta for harness-init

## MODIFIED Requirements

### Requirement: Normalization to real model IDs

Configured model values MUST be normalized to identifiers the delegation tool accepts
before rendering. Normalization MUST recognize curated identifiers from FOUR providers
— Claude, Gemini, OpenAI, Opencode — with canonical output per provider:

| Provider | Match source | Canonical output |
|----------|--------------|------------------|
| Claude | `opus` / `sonnet` / `haiku` family token | short family alias (`opus`/`sonnet`/`haiku`) |
| Gemini | curated `GeminiModels` catalog id | matched catalog id, as-is |
| OpenAI | curated `OpenAIModels` catalog id | matched catalog id, as-is |
| Opencode | curated `OpencodeModels` catalog id | matched catalog id, as-is |

Matching MUST be whole-token: a value merely CONTAINING a family substring (e.g.
`octopus`) MUST NOT match `opus`. Raw display strings (e.g. "Opus 4.8") MUST NOT appear
in the block. When a value normalizes for ANY provider, the phase line renders;
otherwise it returns not-ok, its phase is OMITTED, and the value is accepted with an
advisory warning, never rejected. Catalog contents are a design detail.
(Previously: only Claude `opus|sonnet|haiku` tokens normalized; every Gemini/OpenAI/Opencode value returned not-ok and was dropped from the block.)

#### Scenario: Display name is normalized to an accepted identifier

```gherkin
Scenario: Display name is normalized to an accepted identifier
  Given "models.phases.design" is set to a display string like "Opus 4.8"
  When the orchestrator template is rendered
  Then the "design" line shows a normalized identifier the delegation tool accepts
  And no raw display string appears in the block
```

#### Scenario: Gemini model normalizes to its catalog id

```gherkin
Scenario: Gemini model normalizes to its catalog id
  Given "models.phases.spec" is set to a curated Gemini catalog id
  When the orchestrator template is rendered
  Then the "spec" line shows that Gemini catalog id as-is
```

#### Scenario: OpenAI model normalizes to its catalog id

```gherkin
Scenario: OpenAI model normalizes to its catalog id
  Given "models.phases.tasks" is set to a curated OpenAI catalog id
  When the orchestrator template is rendered
  Then the "tasks" line shows that OpenAI catalog id as-is
```

#### Scenario: Opencode model normalizes to its catalog id

```gherkin
Scenario: Opencode model normalizes to its catalog id
  Given "models.phases.apply" is set to a curated Opencode catalog id
  When the orchestrator template is rendered
  Then the "apply" line shows that Opencode catalog id as-is
```

#### Scenario: Whole-token guard rejects a containing substring

```gherkin
@edge
Scenario: Whole-token guard rejects a containing substring
  Given "models.phases.verify" is set to "octopus"
  When the value is normalized
  Then it does not match the Claude "opus" family
```

#### Scenario: Unresolvable typo is omitted but not rejected

```gherkin
@error
Scenario: Unresolvable typo is omitted but not rejected
  Given "models.phases.propose" is set to an unresolvable value like "Opues 4.8"
  When the configured models are processed for rendering
  Then "propose" is omitted from the block
  And the value is accepted with an advisory warning, not rejected
```

## ADDED Requirements

### Requirement: Cross-provider normalization precedence

When a value could match more than one provider, normalization MUST resolve it by a
fixed precedence: Claude → Gemini → OpenAI → Opencode. The first provider whose
whole-token match succeeds wins. This precedence MUST be stable across runs so the
`## Phase Models` block is byte-identical between the `archon init` and TUI regenerate
paths.

#### Scenario: Colliding value resolves by fixed precedence

```gherkin
@edge
Scenario: Colliding value resolves by fixed precedence
  Given a value that matches both Claude and a later provider
  When the value is normalized
  Then it resolves to the Claude canonical form
```

#### Scenario: Non-Claude default renders an identical block across paths

```gherkin
@happy
Scenario: Non-Claude default renders an identical block across paths
  Given "models.default" is set to a curated non-Claude catalog id
  When the file is rendered via "archon init" and via the TUI regenerate path
  Then both produce a non-empty "## Phase Models" block
  And the two blocks are byte-identical
```
