Feature: sdd-archive vault-aware archive move
  The archive operation must rewrite boundary-crossing links and regenerate map.md
  as part of the atomic move, then guard integrity with --check.

  @happy
  Scenario: Archive operation rewrites boundary links
    Given an active change my-feature with relative links into openspec/specs/
    When sdd-archive archives my-feature
    Then archon map is invoked to rewrite boundary-crossing relative links
    And map.md is regenerated to reflect the new archive path
    And archon map --check passes after the operation

  @happy
  Scenario: Wikilinks survive archive unchanged
    Given a change artifact containing [[harness-workflow]] and [[spec-vault]]
    When sdd-archive archives the change
    Then both wikilinks are byte-identical after the archive move

  @error
  Scenario: Archive aborts if --check fails
    Given an archive operation that produces a dangling relative link
    When archon map --check runs after the move
    Then the command exits non-zero
    And sdd-archive surfaces the failure to the orchestrator
    And the archive is not marked complete
