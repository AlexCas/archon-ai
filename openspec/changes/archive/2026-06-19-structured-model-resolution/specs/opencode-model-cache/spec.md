# opencode-model-cache Specification

## Purpose

Define a structured reader for opencode's provider-keyed cache
(`~/.cache/opencode/models.json`) that preserves the provider, degrades gracefully when the
cache is absent, and serves as the primary catalog source in `ResolveModels` with the existing
`opencode models` shell-out retained as a fallback (PR #45 preserved, not reverted).

## Requirements

### Requirement: Default cache path

`DefaultCachePath()` MUST return the opencode cache location
`~/.cache/opencode/models.json` for the current user's home directory.

#### Scenario: Default cache path points at the opencode cache

```gherkin
Scenario: Default cache path points at the opencode cache
  Given a known user home directory
  When DefaultCachePath() is called
  Then it returns "<home>/.cache/opencode/models.json"
```

### Requirement: Load provider-keyed cache

`LoadModels(path)` MUST parse the provider-keyed JSON into `map[string]Provider`, where each
`Provider{ID, Name, ...}` carries `Models map[string]Model` keyed by the cache's model key.
Malformed or partial provider entries MUST be skipped rather than failing the whole load.

#### Scenario: Well-formed cache yields providers and models

```gherkin
@happy
Scenario: Well-formed cache yields providers and models
  Given a well-formed models.json with an "opencode" provider and models
  When LoadModels is called on its path
  Then the result maps "opencode" to a Provider with its models keyed by model key
```

#### Scenario: Malformed provider entry is skipped

```gherkin
@edge
Scenario: Malformed provider entry is skipped
  Given a models.json with one valid provider and one malformed provider entry
  When LoadModels is called
  Then the valid provider is returned and the malformed entry is skipped without error
```

### Requirement: Graceful degrade on missing cache

`LoadModelsOrEmpty(path)` MUST return an empty map and NO error when the cache file is absent.
Other read or parse errors MUST propagate from `LoadModels`.

#### Scenario: Absent cache returns empty map and no error

```gherkin
@edge
Scenario: Absent cache returns empty map and no error
  Given a path where no models.json exists
  When LoadModelsOrEmpty is called
  Then it returns an empty map and a nil error
```

#### Scenario: Parse error propagates

```gherkin
@error
Scenario: Parse error propagates
  Given a models.json file containing invalid JSON
  When LoadModelsOrEmpty is called
  Then a non-nil parse error is returned
```

### Requirement: Cache-first resolution with shell-out fallback

`ResolveModels` MUST use the cache as the PRIMARY catalog source and MUST fall back to the
`opencode models` shell-out when the cache is absent or empty. `parseModels` and `execLister`
MUST be retained (PR #45 preserved). The `Resolve() []string` signature MUST be unchanged, and
names sourced from the cache MUST be provider-qualified (`FullID` form).

#### Scenario: Resolution prefers the cache

```gherkin
@happy
Scenario: Resolution prefers the cache
  Given a non-empty opencode cache and opencode detected
  When ResolveModels runs
  Then the offered names come from the cache in provider-qualified FullID form
  And the shell-out lister is not invoked
```

#### Scenario: Resolution falls back to the shell-out when cache is empty

```gherkin
@edge
Scenario: Resolution falls back to the shell-out when cache is empty
  Given an absent or empty opencode cache and opencode detected
  When ResolveModels runs
  Then the shell-out "opencode models" path supplies the offered names
```

#### Scenario: PR #45 detection path stays reachable

```gherkin
@edge
Scenario: PR #45 detection path stays reachable
  Given opencode detected but no cache and the shell-out returning models
  When ResolveModels runs
  Then parseModels and execLister produce the offered names as before PR #45
```
