Feature: Archive-before-PR ordering in the single-PR SDD flow
  As an SDD orchestrator running the single-PR flow,
  I want archive to run (and its commit staged) before the PR is opened,
  So that the archive commit travels inside the change's own PR with no trailing archive-only commit.

  Background:
    Given the project uses the single-PR SDD flow
    And the change is tracked in openspec/changes/{change-name}/state.yaml

  @happy
  Scenario: Judge passes, archive runs, then PR is opened
    Given the change has reached judge phase with status completed
    And SESSION_STATUS.md exists at the repository root
    When the orchestrator begins the terminal sequence
    Then archive operations run in order: spec merge, folder move, SESSION_STATUS.md move, archon map
    And all archive operations are staged into a single archive commit on the change branch
    And the PR is opened only after the archive commit is staged
    And the archive commit is included inside the change's PR

  @gate
  Scenario: Archive is blocked before judge passes
    Given the change is in verify phase with status completed
    When the orchestrator attempts to start archive
    Then harness-workflow returns blocked
    And the response indicates judge must complete before archive can run
    And no archive operations are performed

  @invariant
  Scenario: SESSION_STATUS.md stays root-resident through judge
    Given the change has passed verify and is entering judge
    And SESSION_STATUS.md exists at the repository root
    When judge runs and completes
    Then SESSION_STATUS.md remains at the repository root
    And it is not moved before the archive commit is staged

  @invariant
  Scenario: SESSION_STATUS.md moves in the archive commit before PR is opened
    Given judge has passed and archive is running
    When the archive commit is being staged
    Then SESSION_STATUS.md is moved from the repository root into openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md
    And this move is part of the same archive commit as the spec merge and folder move
    And the PR is opened only after this commit is staged
    And SESSION_STATUS.md is no longer present at the repository root after the commit

  @ordering
  Scenario: Archive-internal sub-operation order is preserved
    Given judge has passed and archive is executing
    When the four archive sub-operations run
    Then spec merge completes before folder move begins
    And folder move completes before archon map --backfill is invoked
    And archon map --check runs after --backfill completes
    And PR-open is not invoked until all four sub-operations and the archive commit are done

  @out-of-scope
  Scenario: Chained-PR flow is not affected by the archive-before-PR rule
    Given a change using the chained-PR flow
    When the terminal phase sequence is evaluated
    Then the archive-before-PR ordering rule does not apply to chained flows
    And harness-workflow treats chained-PR archive ownership as deferred to slice 2
    And no blocking or enforcement of archive position is applied to chained-PR flows
