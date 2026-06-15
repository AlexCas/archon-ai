Feature: User-only commit authorship
  Commits made through the harness carry only the user's git authorship.

  Scenario: Commit created without co-author trailers
    Given the harness is asked to commit a completed work unit
    When the commit is created
    Then the author is the user's git account
    And the commit message contains no "Co-Authored-By" trailer
    And the commit message contains no agent or tool attribution
