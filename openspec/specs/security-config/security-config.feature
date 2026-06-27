Feature: Security config surface
  An opt-in security config block, CLI get/set, init flag, and TUI tab,
  defaulting to OFF so existing projects are unaffected.

  @happy
  Scenario: Absent block defaults to disabled
    Given a config file with no security block
    When the config is loaded
    Then security.enabled is false
    And security.profile is empty

  @happy
  Scenario: Clone preserves security fields
    Given a config with security.enabled true and profile "web"
    When the config is cloned
    Then the clone reports security.enabled true and profile "web"

  @happy
  Scenario: Setting a valid profile succeeds
    When the user runs "archon config set security.profile web"
    Then the command succeeds
    And "archon config get security.profile" returns "web"

  @error @security
  Scenario: Setting an invalid profile is rejected
    When the user runs "archon config set security.profile llm"
    Then the command fails with a non-zero exit code
    And the error names the supported values cli and web

  @happy
  Scenario: Init flag enables the gate
    When the user runs "archon init --security"
    Then the emitted .archon/config.yaml contains security.enabled true

  @happy
  Scenario: Init without flag leaves security off
    When the user runs "archon init" without --security
    Then the emitted config has security.enabled false
