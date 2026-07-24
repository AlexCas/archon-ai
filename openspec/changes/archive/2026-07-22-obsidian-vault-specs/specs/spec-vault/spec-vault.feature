Feature: Spec vault layout and link convention
  The openspec/ directory follows a fixed vault shape with a hybrid link convention
  so that the capability graph is fast and reliable to traverse for agents and humans.

  @happy
  Scenario: Vault root contains map.md and known subdirectories
    Given an initialized archon project
    When the openspec/ directory is inspected
    Then openspec/map.md exists as the entry node
    And openspec/specs/ exists as the capability source-of-truth tree
    And openspec/changes/ exists with an archive/ subdirectory

  @happy
  Scenario: Capability reference uses a wikilink
    Given a sdd-propose artifact referencing the harness-workflow capability
    When the artifact is written
    Then the reference appears as [[harness-workflow]] in the markdown body
    And not as a relative path link

  @happy
  Scenario: Intra-change navigation uses relative links
    Given a sdd-spec artifact inside changes/my-feature/specs/foo/
    When the spec links to the change's proposal
    Then the link is a relative markdown link like [proposal](../../proposal.md)
    And not a wikilink

  @happy
  Scenario: Wikilink survives the archive move without rewrite
    Given a change artifact containing [[harness-workflow]]
    When sdd-archive moves the change to changes/archive/YYYY-MM-DD-change/
    Then the wikilink still reads [[harness-workflow]] with no modification

  @happy
  Scenario: Tooling writes only inside managed markers
    Given openspec/map.md with authored prose outside managed markers
    When archon map regenerates map.md
    Then the content between <!-- MAP:START --> and <!-- MAP:END --> is replaced
    And all prose outside the markers is byte-identical to before

  @edge
  Scenario: Authored prose outside markers is never touched
    Given openspec/map.md with custom authored sections outside the managed region
    When archon map is run multiple times
    Then no character outside the managed region is modified

  @happy
  Scenario: Archive move does not touch feature files
    Given a change with specs/foo/foo.feature
    When sdd-archive moves the change folder
    Then foo.feature is moved alongside spec.md to the archive location
    And foo.feature content is byte-identical to before the move
