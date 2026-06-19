# opencode-provider-catalog Specification

## Purpose

Provide provider/model query helpers over the opencode cache (`internal/opencode`) so the TUI
(Slice 3b) can build a provider→model picker: which providers are usable for SDD and which of
their models support tool_call. Additive, pure, deterministic. No UI behavior change in this slice.

## Requirements

### Requirement: tool_call model detection

`hasToolCallModel(Provider) bool` MUST return true iff the provider has at least one model whose
`tool_call` is true.

#### Scenario: provider with a tool_call model

```gherkin
Scenario: provider with a tool_call model
  Given a provider whose models include one with tool_call true
  When hasToolCallModel is called
  Then it returns true
```

#### Scenario: provider with no tool_call model

```gherkin
@edge
Scenario: provider with no tool_call model
  Given a provider whose models all have tool_call false
  When hasToolCallModel is called
  Then it returns false
```

### Requirement: SDD model filter

`FilterModelsForSDD(Provider) []Model` MUST return only the provider's models with `tool_call`
true, sorted ascending by `Name`. Models without tool_call MUST be excluded.

#### Scenario: filter keeps tool_call models sorted by name

```gherkin
Scenario: filter keeps tool_call models sorted by name
  Given a provider with tool_call models "Zeta" and "Alpha" and a non-tool_call model "Mid"
  When FilterModelsForSDD is called
  Then the result is ["Alpha", "Zeta"] and excludes "Mid"
```

### Requirement: available provider detection (simplified)

`DetectAvailableProviders(map[string]Provider) []string` MUST return the provider IDs that have
≥1 tool_call model OR whose ID is `opencode` (the built-in provider, always included if present),
sorted ascending. It MUST NOT consult auth.json or environment variables in this slice.

#### Scenario: providers with tool_call models are available, sorted

```gherkin
Scenario: providers with tool_call models are available, sorted
  Given providers "requesty" (has a tool_call model) and "zeta" (has a tool_call model)
  When DetectAvailableProviders is called
  Then it returns ["requesty", "zeta"]
```

#### Scenario: opencode is always included if present

```gherkin
@edge
Scenario: opencode is always included if present
  Given an "opencode" provider with no tool_call model and a "foo" provider with no tool_call model
  When DetectAvailableProviders is called
  Then it includes "opencode" and excludes "foo"
```

### Requirement: corrupt-vs-absent cache classification (existing seam)

The cache reader MUST let a caller distinguish an ABSENT cache from a CORRUPT one so Slice 3b can
warn only on corruption: `LoadModelsOrEmpty` MUST return `(empty map, nil)` when the cache file
does not exist, and MUST propagate the error when the file exists but is unreadable/unparseable.

#### Scenario: absent cache is not an error

```gherkin
Scenario: absent cache is not an error
  Given the cache file does not exist
  When LoadModelsOrEmpty is called
  Then it returns an empty map and no error
```

#### Scenario: corrupt cache propagates an error

```gherkin
@error
Scenario: corrupt cache propagates an error
  Given the cache file exists but contains invalid top-level JSON
  When LoadModelsOrEmpty is called
  Then it returns an error
```
