Feature: Structured opencode model cache reader
  archon reads opencode's provider-keyed cache (~/.cache/opencode/models.json),
  preserving the provider, degrading gracefully when the cache is missing, and using
  the cache as the primary catalog source in ResolveModels with the existing
  "opencode models" shell-out retained as a fallback (PR #45 preserved).

  @happy
  Scenario: Default cache path points at the opencode cache
    Given a known user home directory
    When DefaultCachePath() is called
    Then it returns "<home>/.cache/opencode/models.json"

  @happy
  Scenario: Well-formed cache yields providers and models
    Given a well-formed models.json with an "opencode" provider and models
    When LoadModels is called on its path
    Then the result maps "opencode" to a Provider with its models keyed by model key

  @edge
  Scenario: Malformed provider entry is skipped
    Given a models.json with one valid provider and one malformed provider entry
    When LoadModels is called
    Then the valid provider is returned and the malformed entry is skipped without error

  @edge
  Scenario: Absent cache returns empty map and no error
    Given a path where no models.json exists
    When LoadModelsOrEmpty is called
    Then it returns an empty map and a nil error

  @error
  Scenario: Parse error propagates
    Given a models.json file containing invalid JSON
    When LoadModelsOrEmpty is called
    Then a non-nil parse error is returned

  @happy
  Scenario: Resolution prefers the cache
    Given a non-empty opencode cache and opencode detected
    When ResolveModels runs
    Then the offered names come from the cache in provider-qualified FullID form
    And the shell-out lister is not invoked

  @edge
  Scenario: Resolution falls back to the shell-out when cache is empty
    Given an absent or empty opencode cache and opencode detected
    When ResolveModels runs
    Then the shell-out "opencode models" path supplies the offered names

  @edge
  Scenario: PR #45 detection path stays reachable
    Given opencode detected but no cache and the shell-out returning models
    When ResolveModels runs
    Then parseModels and execLister produce the offered names as before PR #45
