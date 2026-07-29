Feature: Local Model Provider via OpenAI-compatible endpoint
  Archon users running local models (Ollama, LocalAI) can declare a BaseURL on
  a ModelRef so that archon emits the correct OpenCode V1 provider block, while
  existing configs without a BaseURL remain byte-identical.

  # ─── REQ-1: ModelRef.BaseURL YAML round-trip ────────────────────────────────

  @happy @pr-a
  Scenario: Scalar ref round-trips byte-identically
    Given a config YAML with a phase ref set to the scalar "ollama/llama3"
    When the config is loaded and saved without modification
    Then the emitted YAML is byte-identical to the input

  @happy @pr-a
  Scenario: BaseURL ref marshals as mapping
    Given a ModelRef with Provider "ollama", Model "llama3", and BaseURL "http://localhost:11434/v1"
    When the ref is marshaled to YAML
    Then the output is a mapping with keys provider, model, and base_url
    And the scalar form is NOT emitted

  @happy @pr-a
  Scenario: Scalar input decodes with empty BaseURL
    Given a YAML scalar value "anthropic/claude-sonnet-4-6"
    When the value is unmarshaled into a ModelRef
    Then Provider is "anthropic", Model is "claude-sonnet-4-6", and BaseURL is ""

  # ─── REQ-2: CLI set/get/list ─────────────────────────────────────────────────

  @happy @pr-a
  Scenario: Set and get base_url for a phase
    Given an archon project with no base_url configured
    When the user runs: archon config set models.phases.apply.base_url http://localhost:11434/v1
    Then archon config get models.phases.apply.base_url prints "http://localhost:11434/v1"
    And the provider and model fields of the apply ref are unchanged

  @happy @pr-a
  Scenario: list shows base_url lines
    Given a config where models.phases.apply has BaseURL "http://localhost:11434/v1"
    When the user runs: archon config list
    Then the output includes a line "models.phases.apply.base_url = http://localhost:11434/v1"

  @edge @pr-a
  Scenario: Get base_url when unset returns empty
    Given a config where models.default has no BaseURL
    When the user runs: archon config get models.default.base_url
    Then the command exits 0 and prints nothing

  # ─── REQ-3: Advisory validation ──────────────────────────────────────────────

  @happy @pr-a
  Scenario: Valid BaseURL produces no warning
    Given a ModelRef with Provider "ollama" and BaseURL "http://localhost:11434/v1"
    When advisory validation runs
    Then no warning is emitted

  @error @pr-a
  Scenario: Non-http BaseURL triggers a warning
    Given a ModelRef with Provider "ollama" and BaseURL "ftp://localhost/v1"
    When advisory validation runs
    Then stderr contains: warning: base_url "ftp://localhost/v1" is not a valid http/https URL

  @error @pr-a
  Scenario: BaseURL set but provider empty triggers a warning
    Given a ModelRef with empty Provider and BaseURL "http://localhost:11434/v1"
    When advisory validation runs
    Then stderr contains: warning: base_url is set but provider is empty — provider id required for local model routing

  # ─── REQ-4: OpenCode provider block emission ──────────────────────────────────

  @happy @pr-a
  Scenario: Ollama happy path — single phase
    Given a config with models.phases.apply set to Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
    When archon init runs with agent "opencode"
    Then opencode.json contains a top-level "provider" key
    And provider.ollama.npm equals "@ai-sdk/openai-compatible"
    And provider.ollama.options.baseURL equals "http://localhost:11434/v1"
    And provider.ollama.models contains key "llama3" with value {"name":"llama3"}
    And agent.archon-apply.model equals "ollama/llama3"
    And the "options" object does NOT contain an "apiKey" key

  @happy @pr-a
  Scenario: LocalAI happy path — single phase
    Given a config with models.phases.spec set to Provider "localai", Model "gpt-4-vision", BaseURL "http://localhost:8080/v1"
    When archon init runs with agent "opencode"
    Then provider.localai.npm equals "@ai-sdk/openai-compatible"
    And provider.localai.options.baseURL equals "http://localhost:8080/v1"
    And provider.localai.models contains key "gpt-4-vision"
    And agent.archon-spec.model equals "localai/gpt-4-vision"

  @happy @pr-a
  Scenario: Multiple phases same provider are coalesced
    Given phases apply and verify both use Provider "ollama" and BaseURL "http://localhost:11434/v1" with models "llama3" and "mistral" respectively
    When archon init runs with agent "opencode"
    Then the "provider" object contains exactly ONE "ollama" key
    And provider.ollama.models contains both "llama3" and "mistral"

  @happy @pr-a
  Scenario: Mixed local and remote phases
    Given phase apply uses Provider "ollama" with BaseURL "http://localhost:11434/v1" and model "llama3"
    And phase spec uses Provider "anthropic" with no BaseURL and model "claude-sonnet-4-6"
    When archon init runs with agent "opencode"
    Then opencode.json contains a "provider" block for "ollama"
    And opencode.json does NOT contain a "provider" block for "anthropic"
    And agent.archon-spec.model equals "anthropic/claude-sonnet-4-6"

  @edge @error @pr-a
  Scenario: Conflicting BaseURLs for same provider id
    Given phase apply uses Provider "ollama", BaseURL "http://localhost:11434/v1"
    And phase spec uses Provider "ollama", BaseURL "http://remote-ollama:11434/v1"
    When archon init runs with agent "opencode"
    Then stderr contains: warning: provider "ollama" declared with conflicting baseURLs — using first occurrence "http://localhost:11434/v1"
    And provider.ollama.options.baseURL equals "http://localhost:11434/v1"

  @edge @pr-a
  Scenario: No BaseURL refs — no provider block emitted
    Given a config where no ModelRef has a BaseURL
    When archon init runs with agent "opencode"
    Then opencode.json does NOT contain a top-level "provider" key
    And the output is byte-identical to previous runs with the same config

  @edge @pr-a
  Scenario: Existing user-defined provider entries are preserved
    Given opencode.json already has provider.myprovider with custom user data
    And a config with models.phases.apply using Provider "ollama" with BaseURL
    When archon init runs with agent "opencode"
    Then provider.myprovider is unchanged
    And provider.ollama is added alongside it

  # ─── REQ-5: Deterministic and idempotent output ───────────────────────────────

  @happy @pr-a
  Scenario: Re-run produces identical output
    Given a config with two local providers "aaa" and "zzz" each with one model
    When archon init runs twice with the same config
    Then both runs produce byte-identical opencode.json content
    And provider keys appear in lexicographic order: "aaa" before "zzz"

  # ─── REQ-6: Claude path warn-and-skip guard ───────────────────────────────────

  @happy @pr-b
  Scenario: Local ref on Claude path triggers warn-and-skip
    Given a config with models.phases.apply set to Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
    When archon init runs with agent "claude"
    Then stderr contains: warning: phase "apply" has base_url set but agent is "claude" — local endpoint ignored; claude agents do not support custom baseURLs
    And the file .claude/agents/archon-apply.md is written
    And its model: frontmatter equals "llama3"

  @happy @pr-b
  Scenario: Remote ref on Claude path has no warning
    Given a config with models.phases.apply set to Provider "anthropic", Model "claude-sonnet-4-6", and no BaseURL
    When archon init runs with agent "claude"
    Then stderr contains no base_url warning
    And .claude/agents/archon-apply.md is written normally

  @edge @pr-b
  Scenario: Multiple local phases each emit a warning
    Given phases apply and verify both have BaseURL set and agent is "claude"
    When archon init runs with agent "claude"
    Then stderr contains one warning per local phase
    And both agent files are written with bare model ids

  # ─── REQ-7: TUI Models-tab BaseURL editing ───────────────────────────────────

  @happy @pr-b
  Scenario: Row with BaseURL shows endpoint in display
    Given the TUI Models tab is open
    And the apply row has BaseURL "http://localhost:11434/v1"
    When the tab renders
    Then the apply row shows "http://localhost:11434/v1" in its display text

  @happy @pr-b
  Scenario: User can set BaseURL via TUI sub-mode
    Given the TUI Models tab is open and the apply row is focused
    When the user activates the BaseURL sub-mode and types "http://localhost:11434/v1" and presses Enter
    Then the apply row's BaseURL is "http://localhost:11434/v1"
    And saving the config persists the BaseURL to .archon/config.yaml

  @edge @pr-b
  Scenario: User can clear BaseURL via TUI sub-mode
    Given the apply row has BaseURL "http://localhost:11434/v1"
    When the user activates the BaseURL sub-mode, clears the input, and presses Enter
    Then the apply row's BaseURL is ""
    And the config YAML for that ref reverts to scalar form

  @edge @pr-b
  Scenario: Escape cancels BaseURL edit
    Given the apply row has BaseURL "http://localhost:11434/v1"
    When the user activates the BaseURL sub-mode, types a new value, and presses Escape
    Then the apply row's BaseURL is still "http://localhost:11434/v1"
