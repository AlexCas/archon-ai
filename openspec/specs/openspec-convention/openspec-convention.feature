Feature: openspec-convention vault documentation
  The openspec-convention shared module is updated to document the vault root shape,
  map.md entry node, and hybrid link convention as the canonical rule for all phase skills.

  @happy
  Scenario: Convention documents map.md as entry node
    Given the openspec-convention shared module
    When a phase skill reads the convention for directory structure
    Then it finds openspec/map.md listed as the vault entry node
    And a pointer to the spec-vault convention for link rules

  @happy
  Scenario: Phase skill reads unambiguous link rule
    Given the openspec-convention shared module
    When a phase skill decides how to link to a capability
    Then the convention specifies [[capability]] wikilink for capability-identity references
    And relative markdown links for intra-change artifact navigation

  @happy
  Scenario: Table lists map.md creation and regen
    Given the artifact file paths table in openspec-convention.md
    When a skill author reads the table
    Then it shows sdd-init creates openspec/map.md
    And archon map regenerates openspec/map.md after every phase transition and archive
