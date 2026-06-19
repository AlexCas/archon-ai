Feature: Structured ModelRef with FullID and back-compat YAML
  ModelRef carries a provider alongside the model, exposes FullID() handling the
  opencode key asymmetry and empty-provider advisory state, and (un)marshals so that
  legacy flat-string config.yaml files load and re-save byte-identical until re-picked.

  @happy
  Scenario: FullID joins provider and bare model
    Given a ModelRef with provider "anthropic" and model "claude-sonnet-4-6"
    When FullID() is called
    Then it returns "anthropic/claude-sonnet-4-6"

  @happy
  Scenario: FullID for opencode bare key
    Given a ModelRef with provider "opencode" and model "deepseek-v4-pro"
    When FullID() is called
    Then it returns "opencode/deepseek-v4-pro"

  @edge
  Scenario: FullID never double-prefixes an already-slashed id
    Given a ModelRef with provider "openrouter" and model "xai/grok-4"
    When FullID() is called
    Then it returns "xai/grok-4"

  @edge
  Scenario: FullID with empty provider returns the bare model
    Given a ModelRef with empty provider and model "opus"
    When FullID() is called
    Then it returns "opus" with no leading slash

  @happy
  Scenario: ModelConfig carries structured refs
    Given a ModelConfig built in memory
    When a provider-qualified ref is assigned to Default, Leader and a Phases entry
    Then each field holds a ModelRef preserving its provider and model

  @happy
  Scenario: Legacy slashed scalar splits into provider and model
    Given a YAML scalar "anthropic/claude-sonnet-4-20250514"
    When it is unmarshalled into a ModelRef
    Then provider is "anthropic" and model is "claude-sonnet-4-20250514"

  @edge
  Scenario: Legacy bare alias keeps an empty provider
    Given a YAML scalar "opus"
    When it is unmarshalled into a ModelRef
    Then provider is empty and model is "opus"

  @happy
  Scenario: Mapping node decodes the structured form
    Given a YAML mapping with provider "opencode" and model "deepseek-v4-pro"
    When it is unmarshalled into a ModelRef
    Then provider is "opencode", model is "deepseek-v4-pro"

  @happy
  Scenario: Flat-string config round-trips byte-identical
    Given a config.yaml whose models use legacy flat strings
    When it is loaded and saved again without edits
    Then the saved bytes are identical to the original

  @happy
  Scenario: Bare alias re-saves as the same scalar
    Given a ModelRef with empty provider, model "opus" and empty effort
    When it is marshalled to YAML
    Then it emits the scalar "opus" not a mapping

  @happy
  Scenario: Clone round-trip equality
    Given a config with structured Default, Leader and Phases entries
    When the config is cloned
    Then the clone equals the original and shares no Phases map backing

  @happy
  Scenario: Resolution emits FullID when provider is present
    Given a phase ref with provider "opencode" and model "deepseek-v4-pro"
    When ResolvePhaseModels runs
    Then that phase's Model equals "opencode/deepseek-v4-pro"

  @edge
  Scenario: Resolution emits the bare alias when no provider
    Given a phase ref with empty provider and model "opus"
    When ResolvePhaseModels runs
    Then that phase's Model equals "opus"
