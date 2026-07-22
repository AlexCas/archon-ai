Feature: sdd-init seeds map.md
  archon init scaffolds a vault-ready openspec/ with a managed map.md entry node.

  @happy
  Scenario: Init creates map.md with managed markers
    Given a project with no existing openspec/ directory
    When the user runs archon init
    Then openspec/map.md is created
    And it contains <!-- MAP:START --> and <!-- MAP:END --> markers

  @edge
  Scenario: Init does not overwrite an existing map.md
    Given a project where openspec/map.md already exists with authored content
    When the user runs archon init again
    Then openspec/map.md is left unchanged
