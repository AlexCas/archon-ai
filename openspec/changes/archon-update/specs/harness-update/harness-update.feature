Feature: Safe version-aware skill update
  archon update refreshes installed skills from the embedded set without
  rewriting the orchestrator template or resetting user config.

  Background:
    Given an initialized project with installed skills

  @happy
  Scenario: Update refreshes skills without touching template or user config
    Given an initialized project with a customized "CLAUDE.md" and user config values
    When the user runs "archon update"
    Then the installed skills are refreshed from the embedded set
    And "CLAUDE.md" is left unchanged
    And "models", "playwright", "mutation_testing", "judge", "created_at", and "agent" are preserved
    And only "harness_version", "skill_count", and "skill_inventory" may change

  @happy
  Scenario: Update classifies the version gap
    Given installed skills differing from the embedded set
    When the user runs "archon update"
    Then each skill is classified as added, changed, or orphaned
    And the classification is reported to the user

  @edge
  Scenario: No gaps reports already up to date
    Given installed skills matching the embedded set
    When the user runs "archon update"
    Then the command reports "already up to date"
    And nothing is written

  @happy
  Scenario: Check reports the diff without writing
    Given installed skills differing from the embedded set
    When the user runs "archon update --check"
    Then the added, changed, and orphaned skills are reported
    And no files are written

  @happy
  Scenario: Prune removes orphaned skills
    Given an installed skill that is no longer embedded
    When the user runs "archon update --prune"
    Then the orphaned skill is removed

  @edge
  Scenario: Orphans are kept without prune
    Given an installed skill that is no longer embedded
    When the user runs "archon update"
    Then the orphaned skill is reported
    And the orphaned skill is kept

  @edge
  Scenario: Copy-mode install warns without re-linking
    Given a project whose installed skill path is a real directory, not a symlink
    When the user runs "archon update"
    Then a warning is emitted that the project needs its own update
    And the skill path is not re-linked automatically

  @happy
  Scenario: Output states machine-wide scope
    Given a symlinked project
    When the user runs "archon update"
    Then the output makes clear the refresh affects all symlinked projects

  @error
  Scenario: Update before init reports an actionable error
    Given a project with no ".archon/config.yaml"
    When the user runs "archon update"
    Then an actionable error is reported telling the user to run init first
    And nothing is written
