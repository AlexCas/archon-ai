Feature: config CLI base_url follow-ups (#90, #91)
  Archon users can see and manage base_url endpoints through both
  the "archon config" CLI surface and "archon status", including
  the leader ref, even when no model id is configured.

  # ─── REQ-8: Broaden "none configured" emptiness guard ───────────────────────

  @happy @bug-90
  Scenario: base_url-only default ref is treated as configured
    Given a config where models.default has no provider/model but BaseURL "http://localhost:11434/v1"
    When the user runs: archon config list
    Then the output does NOT contain "(none configured)"
    And the output contains "models.default.base_url = http://localhost:11434/v1"

  @happy @bug-90
  Scenario: base_url-only phase ref is treated as configured
    Given a config where models.phases.apply has no provider/model but BaseURL "http://localhost:11434/v1"
    When the user runs: archon config list
    Then the output does NOT contain "(none configured)"
    And the output contains "models.phases.apply.base_url = http://localhost:11434/v1"

  @regression @bug-90
  Scenario: genuinely empty config still shows (none configured) — regression guard
    Given a config with no models keys set at all
    When the user runs: archon config list
    Then the output is exactly "(none configured)"

  @regression @bug-90
  Scenario: status shows (none configured) for genuinely empty config
    Given a config with no models keys set at all
    When the user runs: archon status
    Then the Models block contains "(none configured)"

  @happy @bug-90
  Scenario: status does NOT show (none configured) when only base_url is set
    Given a config where models.default has no provider/model but BaseURL "http://localhost:11434/v1"
    When the user runs: archon status
    Then the Models block does NOT contain "(none configured)"

  # ─── REQ-9: Suppress empty primary line when model id is absent ───────────

  @happy @bug-90
  Scenario: empty-id ref emits only the base_url line
    Given a config where models.default has BaseURL "http://localhost:11434/v1" and no provider/model
    When the user runs: archon config list
    Then the output contains "models.default.base_url = http://localhost:11434/v1"
    And the output does NOT contain "models.default = "

  @happy @bug-90
  Scenario: ref with both id and base_url emits both lines
    Given a config where models.phases.apply has Provider "ollama", Model "llama3", and BaseURL "http://localhost:11434/v1"
    When the user runs: archon config list
    Then the output contains "models.phases.apply = ollama/llama3"
    And the output contains "models.phases.apply.base_url = http://localhost:11434/v1"

  # ─── REQ-10: status renders base_url lines and suppresses empty primary ────

  @happy @bug-90
  Scenario: status shows base_url for default ref
    Given a config where models.default has BaseURL "http://localhost:11434/v1" and Provider "ollama", Model "llama3"
    When the user runs: archon status
    Then the Models block contains "Default:"
    And the Models block contains "http://localhost:11434/v1"

  @happy @bug-90
  Scenario: status suppresses empty primary for default ref with only base_url
    Given a config where models.default has BaseURL "http://localhost:11434/v1" and no provider/model
    When the user runs: archon status
    Then the Models block contains "http://localhost:11434/v1"
    And the Models block does NOT contain "Default:  " followed immediately by a newline

  @happy @bug-90
  Scenario: status shows base_url for a phase ref
    Given a config where models.phases.apply has Provider "ollama", Model "llama3", and BaseURL "http://localhost:11434/v1"
    When the user runs: archon status
    Then the Models block contains "apply:"
    And the Models block contains "http://localhost:11434/v1"

  # ─── REQ-11: Wire models.leader through config set/get/list ─────────────────

  @happy @feat-91
  Scenario: set and get models.leader round-trip
    Given an archon project with no leader configured
    When the user runs: archon config set models.leader ollama/llama3
    Then archon config get models.leader prints "ollama/llama3"
    And the provider and BaseURL fields of the leader ref are unchanged

  @happy @feat-91
  Scenario: set and get models.leader.base_url round-trip
    Given an archon project with no leader base_url configured
    When the user runs: archon config set models.leader.base_url http://localhost:11434/v1
    Then archon config get models.leader.base_url prints "http://localhost:11434/v1"
    And the provider and model fields of the leader ref are unchanged

  @happy @feat-91
  Scenario: config list shows leader block
    Given a config where models.leader is Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
    When the user runs: archon config list
    Then the output contains "models.leader = ollama/llama3"
    And the output contains "models.leader.base_url = http://localhost:11434/v1"

  @happy @feat-91
  Scenario: get models.leader.base_url when unset returns empty
    Given an archon project with no leader base_url configured
    When the user runs: archon config get models.leader.base_url
    Then the command exits 0 and prints nothing

  @advisory @feat-91
  Scenario: leader base_url set triggers advisory validation — non-blocking on bad URL
    Given an archon project with no leader configured
    When the user runs: archon config set models.leader.base_url ftp://bad-url
    Then the command exits 0
    And stderr contains a warning about an invalid http/https URL
    And archon config get models.leader.base_url prints "ftp://bad-url"

  @advisory @feat-91
  Scenario: leader base_url set with valid URL triggers no warning
    Given an archon project with no leader configured
    When the user runs: archon config set models.leader.base_url http://localhost:11434/v1
    Then the command exits 0
    And stderr contains no base_url warning

  @error @feat-91
  Scenario: unknown models.leader key in set is rejected with updated error listing
    Given an archon project
    When the user runs: archon config set models.leader.typo somevalue
    Then the command exits non-zero
    And stderr contains "models.leader" in the supported-keys message
    And stderr contains "models.leader.base_url" in the supported-keys message

  @error @feat-91
  Scenario: unknown models.leader key in get is rejected with updated error listing
    Given an archon project
    When the user runs: archon config get models.leader.typo
    Then the command exits non-zero
    And stderr contains "models.leader" in the supported-keys message
    And stderr contains "models.leader.base_url" in the supported-keys message

  # ─── REQ-12: status renders Leader block symmetric to Default ───────────────

  @happy @feat-91
  Scenario: status shows Leader block with id and base_url
    Given a config where models.leader is Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
    When the user runs: archon status
    Then the Models block contains "Leader:"
    And the Models block contains "ollama/llama3"
    And the Models block contains "http://localhost:11434/v1"

  @happy @feat-91
  Scenario: status shows Leader block when only base_url is set
    Given a config where models.leader has no provider/model but BaseURL "http://localhost:11434/v1"
    When the user runs: archon status
    Then the Models block contains "Leader:"
    And the Models block contains "http://localhost:11434/v1"
    And the Models block does NOT contain a blank leader model-id line

  @edge @feat-91
  Scenario: status omits Leader block when leader is genuinely empty
    Given a config where models.leader is completely unset
    When the user runs: archon status
    Then the Models block does NOT contain "Leader:"

  @happy @feat-91
  Scenario: config list and status render leader symmetrically
    Given a config where models.leader is Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
    When the user runs both: archon config list and archon status
    Then config list contains "models.leader = ollama/llama3"
    And config list contains "models.leader.base_url = http://localhost:11434/v1"
    And status contains "Leader:" with the same model id and base_url values

  @regression @feat-91
  Scenario: leader guard included in broadened emptiness check
    Given a config where models.leader has no provider/model but BaseURL "http://localhost:1234/v1"
    And models.default and models.phases are all empty
    When the user runs: archon config list
    Then the output does NOT contain "(none configured)"
    And the output contains "models.leader.base_url = http://localhost:1234/v1"
