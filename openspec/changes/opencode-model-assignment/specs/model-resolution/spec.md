# model-resolution Specification

## Purpose

Resolve archon's bare model names (e.g. `gpt-4o`, `claude-sonnet-4`) read from
`.archon/config.yaml` (`models.default` / `models.phases`) into the
`provider/model-id` form opencode requires. Resolution prefers opencode's own model
cache, falls back to a static map, and passes already-qualified IDs through verbatim.
All requirements below are Slice 1 (MVP).

## Requirements

### Requirement: Qualified ID Pass-Through

The system MUST pass any value that is already in `provider/model` form through
verbatim, without modification.

#### Scenario: Already-qualified ID is unchanged

- GIVEN a config model value `anthropic/claude-sonnet-4`
- WHEN the value is resolved
- THEN the resolved value is exactly `anthropic/claude-sonnet-4`
- AND no cache or static-map lookup alters it

### Requirement: Cache-Based Resolution

When `~/.cache/opencode/models.json` is present, the system MUST resolve a bare model
name by matching it against the providers and model IDs in that cache, returning the
matched `provider/model-id`.

#### Scenario: Bare name matched from cache

- GIVEN `~/.cache/opencode/models.json` lists a provider whose models include the
  bare name
- WHEN that bare name is resolved
- THEN the result is the `provider/model-id` from the cache

### Requirement: Static-Map Fallback

When the cache is absent or does not contain the bare name, the system MUST fall back
to a static `name → provider/model` map covering the names in `config.KnownModels`.

#### Scenario: Fallback when cache missing

- GIVEN `~/.cache/opencode/models.json` does not exist
- AND the bare name is present in `config.KnownModels`
- WHEN the name is resolved
- THEN the result is the static-map `provider/model` value for that name

#### Scenario: Fallback when name absent from cache

- GIVEN the cache exists but does not contain the bare name
- AND the static map covers the name
- WHEN the name is resolved
- THEN the static-map value is returned

### Requirement: Init Never Fails on Missing Cache

`archon init` MUST NEVER fail because `~/.cache/opencode/models.json` is missing. A
missing cache MUST degrade gracefully to the static-map fallback path.

#### Scenario: Missing cache does not abort init

- GIVEN `~/.cache/opencode/models.json` does not exist
- WHEN `archon init` runs the resolution step for the opencode agent
- THEN init does not fail
- AND bare names resolve via the static-map fallback (or pass through if qualified)

#### Scenario: Unresolvable name does not abort init

- GIVEN a bare name is in neither the cache nor the static map
- WHEN init resolves models
- THEN init does not fail
- AND the resolver surfaces a warning rather than an error
