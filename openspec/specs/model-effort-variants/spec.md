# model-effort-variants Specification

## Purpose

Let users pick an effort/reasoning level for reasoning-capable models in the TUI Models picker,
persist it on `config.ModelRef.Effort`, and write it to `opencode.json` as the per-agent `variant`
field — without any opencode plugin, variants cache, or embedded-asset subsystem (availability is
derived from the existing `opencode.Model.Reasoning` flag).

## Requirements

### Requirement: Effort selection for reasoning models only

After a model is picked in the TUI, if the picked `opencode.Model.Reasoning` is true, the picker
MUST enter an effort-selection step offering `default`, `low`, `medium`, `high`; choosing one MUST
set the row's `ModelRef.Effort` (`default` → empty) and return to row navigation. If the picked
model is NOT reasoning-capable, the picker MUST skip the step and leave `Effort` empty.

#### Scenario: Reasoning model offers effort selection
```gherkin
Scenario: Reasoning model offers effort selection
  Given the user picks a provider and a model whose Reasoning is true
  When the model is selected
  Then an effort step is shown offering default/low/medium/high
  And choosing "high" sets the row's ModelRef.Effort to "high" and returns to row navigation
```

#### Scenario: Non-reasoning model skips the effort step
```gherkin
@edge
Scenario: Non-reasoning model skips the effort step
  Given the user picks a model whose Reasoning is false
  When the model is selected
  Then no effort step is shown and the row's ModelRef.Effort is empty
```

#### Scenario: The default effort option maps to empty
```gherkin
@edge
Scenario: The default effort option maps to empty
  Given the effort step is shown
  When the user chooses "default"
  Then the row's ModelRef.Effort is empty
```

### Requirement: Effort propagates through phase resolution

`config.ResolvePhaseModels` MUST carry each resolved ref's `Effort` into the returned
`PhaseModel.Effort` alongside the FullID model string. The phase-order and fallback logic are
unchanged.

#### Scenario: ResolvePhaseModels carries effort
```gherkin
Scenario: ResolvePhaseModels carries effort
  Given a phase ref with model and Effort "medium"
  When ResolvePhaseModels runs
  Then that phase's PhaseModel has the FullID model and Effort "medium"
```

### Requirement: variant written to opencode.json

`mergeOpencodeAgent` MUST write a `variant` field (json `variant,omitempty`) on each `archon-leader`
and `archon-<phase>` agent, set from the resolved `Effort`. When the effort is empty the `variant`
key MUST be omitted. Output MUST stay deterministic and idempotent (a re-run with the same config is
byte-identical).

#### Scenario: variant present when effort set
```gherkin
Scenario: variant present when effort set
  Given a phase resolves to a model with Effort "high"
  When the opencode agents are written
  Then "agent.archon-<phase>.variant" equals "high"
```

#### Scenario: variant omitted when effort empty
```gherkin
@edge
Scenario: variant omitted when effort empty
  Given a phase resolves to a model with empty Effort
  When the opencode agents are written
  Then "agent.archon-<phase>" has no "variant" key
```

#### Scenario: Re-run is byte-identical
```gherkin
@edge
Scenario: Re-run is byte-identical
  Given opencode agents were written with some efforts set and some empty
  When the merge runs again with the same configuration
  Then the resulting opencode.json is byte-identical to the previous run
```

### Requirement: Config round-trip preserves effort

A `ModelRef` with a non-empty `Effort` MUST round-trip through config save/load as a mapping
(`{provider, model, effort}`); an effortless ref MUST stay a scalar (no churn for existing configs).

#### Scenario: Effort persists in config
```gherkin
Scenario: Effort persists in config
  Given a ModelRef with provider, model and Effort "low"
  When the config is saved and reloaded
  Then the reloaded ModelRef has Effort "low"
```
