Feature: Archive-before-merge ordering in the Feature Branch Chain SDD flow
  As an SDD orchestrator running the Feature Branch Chain flow,
  I want archive to run (and its commit staged on the tracker branch) after the
  integrated judge passes and before the tracker PR merges to main,
  So that the archive commit travels inside the tracker PR with no separate
  archive-only PR and no polluted child PR diff.

  Background:
    Given the project uses the Feature Branch Chain SDD strategy
    And the change is tracked in openspec/changes/{change-name}/state.yaml
    And a tracker/integration branch and draft tracker PR exist

  @happy
  Scenario: Integrated judge passes on tracker, archive staged, tracker merges to main
    Given all child PRs have been merged into the tracker branch
    And the integrated judge has run on the tracker branch and passed
    And SESSION_STATUS.md exists at the repository root
    When the orchestrator begins the terminal sequence on the tracker branch
    Then archive operations run in order: spec merge, folder move, archon map --backfill, archon map --check, SESSION_STATUS.md move
    And all archive operations are staged into a single archive commit on the tracker branch
    And the tracker PR is merged to main only after the archive commit is staged
    And the archive commit is included inside the tracker PR
    And no separate archive-only PR is created
    And no child PR diff is polluted with archive changes

  @gate
  Scenario: Archive is blocked on tracker before integrated judge passes
    Given all child PRs are merged into the tracker branch
    And the integrated judge has not yet run on the tracker branch
    When the orchestrator attempts to start archive on the tracker branch
    Then harness-workflow returns blocked
    And the response indicates the integrated judge must pass on the tracker branch before archive can run
    And no archive operations are performed

  @gate
  Scenario: Archive is blocked when integrated judge failed on tracker
    Given all child PRs are merged into the tracker branch
    And the integrated judge has run on the tracker branch and failed
    When the orchestrator attempts to start archive on the tracker branch
    Then harness-workflow returns blocked
    And the response indicates the integrated judge must pass before archive can run
    And no archive operations are performed
    And the tracker PR is not merged to main

  @gate
  Scenario: Integrated judge runs on tracker with all changes integrated
    Given at least one child PR has been locally judged before integration
    When the final child PR is merged into the tracker branch
    Then an integrated judge run must be triggered on the tracker branch
    And the integrated judge evaluates all change files as a unified whole
    And the integrated judge result is the gate for both the archive step and the tracker merge to main

  @invariant
  Scenario: SESSION_STATUS.md stays root-resident through integrated judge on tracker
    Given child PRs are open or have been merged into the tracker branch
    And SESSION_STATUS.md exists at the repository root
    When the integrated judge runs on the tracker branch and completes
    Then SESSION_STATUS.md remains at the repository root
    And it is not moved before the tracker archive commit is staged
    And no child PR or intermediate merge touches SESSION_STATUS.md

  @invariant
  Scenario: SESSION_STATUS.md moves in the tracker archive commit
    Given the integrated judge has passed on the tracker branch and archive is running
    When the tracker archive commit is being staged
    Then SESSION_STATUS.md is moved from the repository root into openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md
    And this move is part of the same tracker archive commit as the spec merge and folder move
    And the tracker PR is merged to main only after this commit is staged
    And SESSION_STATUS.md is no longer present at the repository root after the commit

  @ordering
  Scenario: Archive-internal sub-operation order is preserved on the tracker branch
    Given the integrated judge has passed and archive is executing on the tracker branch
    When the four archive sub-operations run
    Then spec merge completes before folder move begins
    And folder move completes before archon map --backfill is invoked
    And archon map --check runs after --backfill completes
    And archon map --check passes before SESSION_STATUS.md is moved
    And tracker PR merge to main is not invoked until all four sub-operations and the archive commit are done

  @out-of-scope
  Scenario: Stacked-to-Main archive ownership is out of scope (deferred to slice 2b)
    Given a change using the Stacked-to-Main strategy where child PRs each target main
    When the terminal phase sequence is evaluated
    Then the Feature Branch Chain archive rule does not apply to Stacked-to-Main flows
    And harness-workflow treats Stacked-to-Main archive ownership as deferred to slice 2b
    And no blocking or enforcement of archive position is applied to Stacked-to-Main flows
