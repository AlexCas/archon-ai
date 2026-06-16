Feature: Truthful skill inventory versions at init
  archon init records each skill's real SKILL.md frontmatter version in
  skill_inventory instead of a hardcoded value.

  Background:
    Given a blank project directory

  @happy
  Scenario: Init records real frontmatter versions
    Given embedded skills whose "SKILL.md" frontmatter declares versions like "2.0" and "3.0"
    When the user runs "archon init --agent claude"
    Then each "skill_inventory" entry records that skill's real frontmatter version
    And no inventory entry uses a hardcoded version

  @edge
  Scenario: Missing frontmatter version is handled
    Given an embedded skill whose "SKILL.md" declares no metadata.version
    When the user runs "archon init --agent claude"
    Then that skill is still recorded in "skill_inventory"
    And init does not abort
