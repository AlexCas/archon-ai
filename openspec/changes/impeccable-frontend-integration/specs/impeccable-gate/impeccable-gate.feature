Feature: Impeccable design-language gate
  An opt-in design-language quality gate that integrates the Impeccable npm tool
  via "npx impeccable" into the Archon harness. Mirrors the Playwright precedent:
  config struct, --impeccable init flag, TUI tab, archon status block, CLI get/set,
  preflight group F, additive phase hooks (design/apply/verify/tasks/explore/spec),
  and a blocking/advisory judge detection gate. Fully inert when disabled.

  Background:
    Given an Archon project with a valid ".archon/config.yaml"

  # --- Config surface ---

  @happy
  Scenario: Zero-value config is fully inert
    Given ".archon/config.yaml" has no "impeccable" section
    When any Archon phase runs
    Then no Impeccable checks or gate logic executes
    And all phase outputs are identical to a build without this change

  @happy
  Scenario: Clone preserves all Impeccable fields
    Given a Config with a fully populated Impeccable struct
    When "Clone()" is called
    Then the returned Config.Impeccable equals the original field-for-field
    And modifying the clone does not affect the original

  @happy
  Scenario: CloneRoundtrip fixture catches missing Clone wiring
    Given "TestConfig_CloneRoundtrip" uses a non-zero Impeccable struct
    When "Clone()" is called and the result is compared field-by-field
    Then the test fails if any Impeccable field is missing from Clone()

  # --- Severity semantics ---

  @happy
  Scenario: block-deterministic severity — detector fails, LLM critique advises
    Given "impeccable.severity: block-deterministic"
    And the deterministic detector reports 2 violations
    And the LLM critique reports 1 advisory note
    When the judge gate runs
    Then the gate returns "fail" due to deterministic violations
    And the advisory note is included in the report without blocking

  @edge
  Scenario: advisory severity — detector violations do not block
    Given "impeccable.severity: advisory"
    And the deterministic detector reports 3 violations
    When the judge gate runs
    Then the gate returns "pass"
    And all violations appear as advisory notes in the report

  @error
  Scenario: Invalid severity value rejected at config load
    Given ".archon/config.yaml" contains "severity: foobar"
    When any Archon command loads the config
    Then Archon exits with a descriptive error naming the invalid value
    And the error lists the three valid values: block-deterministic, block-all, advisory

  # --- Init flag ---

  @happy
  Scenario: Init with --impeccable flag enables the gate
    When the user runs "archon init --impeccable"
    Then ".archon/config.yaml" contains "impeccable.enabled: true"
    And all other impeccable fields retain their defaults

  @happy
  Scenario: Init without --impeccable leaves gate disabled
    When the user runs "archon init" with no impeccable flag
    Then the effective value of impeccable.enabled is false
    And no impeccable phase hook or gate logic activates

  # --- CLI get/set ---

  @happy
  Scenario: CLI set impeccable.enabled to true
    When the user runs "archon config set impeccable.enabled true"
    Then ".archon/config.yaml" is updated with "impeccable.enabled: true"
    And "archon config get impeccable.enabled" returns "true"

  @error
  Scenario: CLI rejects invalid severity via set
    When the user runs "archon config set impeccable.severity invalid"
    Then the command exits with a non-zero status
    And the error names the invalid value and lists the three valid options

  @error
  Scenario: Get unknown impeccable key shows updated supported-key list
    When the user runs "archon config get impeccable.typo"
    Then the error message lists all five impeccable keys

  # --- TUI tab ---

  @happy
  Scenario: Impeccable tab renders current config
    Given ".archon/config.yaml" contains a non-default impeccable block
    When the user opens "archon tui" and navigates to the Impeccable tab
    Then the tab displays current values for all five impeccable fields

  @happy
  Scenario: TUI save persists Impeccable changes
    Given the user sets enabled to true and severity to "advisory" in the Impeccable tab
    When the user saves in the TUI
    Then ".archon/config.yaml" reflects the new values

  @happy
  Scenario: Tab count and order tests pass with new tab
    Given "ImpeccableTab" is added to the enum and tabs slice
    When "go test ./internal/tui/..." runs
    Then all tab-count and tab-order assertions pass

  # --- archon status ---

  @happy
  Scenario: Status shows Impeccable as disabled
    Given "impeccable.enabled: false"
    When "archon status" runs
    Then the output contains an "Impeccable (Design Language)" section showing "Enabled: false"

  @happy
  Scenario: Status shows Impeccable as enabled with config details
    Given "impeccable.enabled: true", "severity: block-all", custom paths set
    When "archon status" runs
    Then the section shows enabled, severity, product_path, and design_path

  # --- Judge detection gate ---

  @happy
  Scenario: Judge gate passes when no violations found
    Given "impeccable.enabled: true" and "severity: block-deterministic"
    And "npx impeccable detect" returns exit 0 with no violations
    When the judge gate runs
    Then the "### Impeccable Gate" section shows "Status: pass"

  @happy
  Scenario: Judge gate blocks on deterministic violations
    Given "impeccable.enabled: true" and "severity: block-deterministic"
    And "npx impeccable detect" reports 3 deterministic violations
    When the judge gate runs
    Then the "### Impeccable Gate" section shows "Status: fail"
    And the overall judge verdict is "fail"

  @happy
  Scenario: Judge gate skipped when impeccable is disabled
    Given "impeccable.enabled: false"
    When "harness-judge" runs
    Then no "npx impeccable" invocation occurs
    And no "### Impeccable Gate" section appears in the output

  @error
  Scenario: Judge gate returns blocked when Node/npx is absent
    Given "impeccable.enabled: true"
    And "node" or "npx" is not available in the target environment
    When the judge gate runs
    Then the gate returns "blocked" with an actionable install message
    And the overall judge verdict is "blocked"
    And the gate does NOT silently pass

  # --- Design-phase minimal reference ---

  @happy
  Scenario: Design references PRODUCT.md and DESIGN.md when both exist
    Given "impeccable.enabled: true"
    And "PRODUCT.md" and "DESIGN.md" exist at the target-project root
    When "sdd-design" runs
    Then the design phase reads those files as input context
    And "openspec/changes/<change>/design.md" is the only SDD design file written
    And no "npx impeccable" invocation occurs during design

  @edge
  Scenario: Design continues normally when Impeccable docs are missing
    Given "impeccable.enabled: true"
    And neither "PRODUCT.md" nor "DESIGN.md" exists
    When "sdd-design" runs
    Then the phase proceeds without error
    And the design artifact recommends running "impeccable init"

  @happy
  Scenario: Design phase unchanged when impeccable is disabled
    Given "impeccable.enabled: false"
    When "sdd-design" runs
    Then the phase behaves identically to today

  # --- auto_install ---

  @error
  Scenario: npx impeccable not found with auto_install false instructs rather than installs
    Given "impeccable.enabled: true" and "impeccable.auto_install: false"
    And "npx impeccable" fails because the package is not installed
    When the judge gate runs
    Then the gate returns "blocked" with an instruction to install Impeccable
    And no silent install occurs

  @edge
  Scenario: auto_install true triggers install before gate
    Given "impeccable.enabled: true" and "impeccable.auto_install: true"
    And Impeccable is not installed in the target project
    When the judge gate runs for the first time
    Then "npx impeccable install" runs before detection
    And the detection gate runs after install completes

  # --- Preflight group F ---

  @happy
  Scenario: Generated CLAUDE.md includes preflight group F
    When "archon init" runs
    Then the generated "CLAUDE.md" contains the group F preflight question for Impeccable
    And it contains the group F mapping paragraph

  @happy
  Scenario: Generated CLAUDE.md includes the Impeccable rule
    When "archon init" runs
    Then the generated "CLAUDE.md" contains a rule line referencing "impeccable.enabled"
    And this rule appears in both the Claude and Opencode rule-const variants

  @happy
  Scenario: Template tests assert group F and the Impeccable rule
    When "go test ./internal/initcmd/..." runs
    Then all template assertions pass including group F and the Impeccable rule

  # --- Thin impeccable skill ---

  @happy
  Scenario: Impeccable skill is auto-embedded and skill_count becomes 25
    Given "skills/impeccable/SKILL.md" exists
    When "archon init" or "archon update" runs
    Then "skill_count" in the generated config is 25
    And "skill_inventory" includes an "impeccable" entry

  @happy
  Scenario: Skill delegates detection to npx, not Go code
    Given "skills/impeccable/SKILL.md" is authored
    When the skill instructions are followed
    Then the only invocation mechanism is "npx impeccable <subcommand>"
