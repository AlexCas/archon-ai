# model-ref Specification

## Purpose

Define a structured `provider + model` representation (`ModelRef`) that replaces flat model
strings in `config.ModelConfig`, with a `FullID()` accessor and back-compatible YAML
(un)marshalling so existing flat-string `config.yaml` files load and re-save byte-identical
until the user re-picks a model.

## Requirements

### Requirement: ModelRef FullID resolution

`ModelRef{Provider, Model, Effort}` MUST expose `FullID() string`. With a non-empty `Provider`
and a bare `Model`, `FullID()` MUST return `"<provider>/<model>"`. When `Model` already
contains a `/` (already-slashed ids for non-opencode providers), `FullID()` MUST return it
as-is and MUST NOT double-prefix. When `Provider` is empty, `FullID()` MUST return the bare
`Model` with no leading slash (advisory-only state).

#### Scenario: FullID joins provider and bare model

```gherkin
Scenario: FullID joins provider and bare model
  Given a ModelRef with provider "anthropic" and model "claude-sonnet-4-6"
  When FullID() is called
  Then it returns "anthropic/claude-sonnet-4-6"
```

#### Scenario: FullID for opencode bare key

```gherkin
Scenario: FullID for opencode bare key
  Given a ModelRef with provider "opencode" and model "deepseek-v4-pro"
  When FullID() is called
  Then it returns "opencode/deepseek-v4-pro"
```

#### Scenario: FullID never double-prefixes an already-slashed id

```gherkin
@edge
Scenario: FullID never double-prefixes an already-slashed id
  Given a ModelRef with provider "openrouter" and model "xai/grok-4"
  When FullID() is called
  Then it returns "xai/grok-4"
```

#### Scenario: FullID with empty provider returns the bare model

```gherkin
@edge
Scenario: FullID with empty provider returns the bare model
  Given a ModelRef with empty provider and model "opus"
  When FullID() is called
  Then it returns "opus" with no leading slash
```

### Requirement: Structured ModelConfig fields

`ModelConfig` MUST hold `Default ModelRef`, `Leader ModelRef`, and `Phases map[string]ModelRef`,
replacing the previous flat-string fields.

#### Scenario: ModelConfig carries structured refs

```gherkin
Scenario: ModelConfig carries structured refs
  Given a ModelConfig built in memory
  When a provider-qualified ref is assigned to Default, Leader and a Phases entry
  Then each field holds a ModelRef preserving its provider and model
```

### Requirement: Back-compat YAML unmarshalling

`ModelRef.UnmarshalYAML` MUST accept a legacy scalar node: `"a/b"` decodes to
`{Provider:"a", Model:"b"}`, and a bare `"x"` decodes to `{Provider:"", Model:"x"}` —
it MUST NOT guess a provider for a bare alias. A mapping node MUST decode the structured
`{provider, model, effort}` form.

#### Scenario: Legacy slashed scalar splits into provider and model

```gherkin
Scenario: Legacy slashed scalar splits into provider and model
  Given a YAML scalar "anthropic/claude-sonnet-4-20250514"
  When it is unmarshalled into a ModelRef
  Then provider is "anthropic" and model is "claude-sonnet-4-20250514"
```

#### Scenario: Legacy bare alias keeps an empty provider

```gherkin
@edge
Scenario: Legacy bare alias keeps an empty provider
  Given a YAML scalar "opus"
  When it is unmarshalled into a ModelRef
  Then provider is empty and model is "opus"
```

#### Scenario: Mapping node decodes the structured form

```gherkin
Scenario: Mapping node decodes the structured form
  Given a YAML mapping with provider "opencode" and model "deepseek-v4-pro"
  When it is unmarshalled into a ModelRef
  Then provider is "opencode", model is "deepseek-v4-pro"
```

### Requirement: Scalar-on-empty YAML marshalling

`ModelRef.MarshalYAML` MUST emit a SCALAR equal to the bare `Model` when `Provider == ""` and
`Effort == ""`, and a mapping otherwise. As a consequence, an existing flat-string
`config.yaml` MUST load and re-save byte-identical until the user re-picks a model.

#### Scenario: Flat-string config round-trips byte-identical

```gherkin
@happy
Scenario: Flat-string config round-trips byte-identical
  Given a config.yaml whose models use legacy flat strings
  When it is loaded and saved again without edits
  Then the saved bytes are identical to the original
```

#### Scenario: Bare alias re-saves as the same scalar

```gherkin
Scenario: Bare alias re-saves as the same scalar
  Given a ModelRef with empty provider, model "opus" and empty effort
  When it is marshalled to YAML
  Then it emits the scalar "opus" not a mapping
```

### Requirement: Clone deep-copies structured models

`config.Clone` MUST deep-copy the new `ModelConfig`, including the `Phases map[string]ModelRef`,
so a clone round-trips equal to the original.

#### Scenario: Clone round-trip equality

```gherkin
Scenario: Clone round-trip equality
  Given a config with structured Default, Leader and Phases entries
  When the config is cloned
  Then the clone equals the original and shares no Phases map backing
```

### Requirement: Phase resolution emits FullID

`ResolvePhaseModels` MUST yield `PhaseModel.Model == ref.FullID()` when the phase's resolved
ref has a provider, and the bare alias when it does not. The phase-order iteration and the
default-fallback logic MUST remain unchanged. `NormalizeModel` MUST be retained for `Validate`
(offline advisory) and MUST NOT be used in the resolution path.

#### Scenario: Resolution emits FullID when provider is present

```gherkin
Scenario: Resolution emits FullID when provider is present
  Given a phase ref with provider "opencode" and model "deepseek-v4-pro"
  When ResolvePhaseModels runs
  Then that phase's Model equals "opencode/deepseek-v4-pro"
```

#### Scenario: Resolution emits the bare alias when no provider

```gherkin
@edge
Scenario: Resolution emits the bare alias when no provider
  Given a phase ref with empty provider and model "opus"
  When ResolvePhaseModels runs
  Then that phase's Model equals "opus"
```
