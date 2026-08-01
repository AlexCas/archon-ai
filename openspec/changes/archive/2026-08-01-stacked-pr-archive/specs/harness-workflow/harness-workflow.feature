Feature: Stacked-to-Main archive convergence to Feature Branch Chain
  As an SDD orchestrator with archive-before-PR in effect,
  I want the harness-workflow to prevent pure Stacked-to-Main selection when
  archive-before-PR applies and to converge to Feature Branch Chain up-front,
  So that every change always has a single un-merged owning ref when the archive
  commit is staged, eliminating the Stacked-to-Main archive ownership gap.

  Background:
    Given the project uses SDD with artifact store openspec or hybrid
    And archive-before-PR is therefore in effect
    And the change is tracked in openspec/changes/{change-name}/state.yaml

  @happy
  Scenario: Orchestrator converges to FBC at sdd-tasks when archive-before-PR is active
    Given the orchestrator is selecting a chain strategy during sdd-tasks
    And the user preference or default would select Stacked-to-Main
    When harness-workflow evaluates the strategy selection
    Then the orchestrator selects Feature Branch Chain instead of Stacked-to-Main
    And the orchestrator notifies the user that Stacked-to-Main + archive-before-PR is unsupported and FBC is selected
    And a tracker integration branch is created or planned before any child PR is opened
    And the archive step will be governed by the Terminal Phase Ordering Feature Branch Chain requirement

  @gate
  Scenario: Pure Stacked-to-Main + archive-before-PR is blocked before any child PR merges to main
    Given the orchestrator attempted to proceed with pure Stacked-to-Main without a tracker branch
    And archive-before-PR is in effect
    When any child PR would merge to main without a tracker branch in place
    Then harness-workflow returns blocked
    And the response states that Stacked-to-Main + archive-before-PR is an unsupported combination
    And no child PR is merged to main until the strategy is corrected to Feature Branch Chain
    And harness-workflow directs the orchestrator to converge to FBC before opening child PRs

  @happy
  Scenario: After convergence the FBC archive rule governs with no new mechanics
    Given the orchestrator has converged from Stacked-to-Main to Feature Branch Chain at sdd-tasks
    And the change is now proceeding as a Feature Branch Chain flow
    And the integrated judge has passed on the tracker branch
    When the terminal phase sequence runs
    Then the archive step follows the Terminal Phase Ordering Feature Branch Chain requirement verbatim
    And no additional archive mechanics are introduced by the Stacked-to-Main convergence
    And the archive commit is staged on the tracker branch before the tracker PR merges to main

  @out-of-scope
  Scenario: Late Stacked-to-FBC conversion is not the sanctioned path
    Given archive-before-PR is in effect
    And one or more slices have already merged to main under Stacked-to-Main
    When the orchestrator attempts a late conversion to Feature Branch Chain
    Then this conversion path is not the sanctioned approach
    And harness-workflow treats this state as an unresolved hazard due to stranded slices on main without an owning tracker ref
    And the orchestrator must not attempt archive under these conditions without explicit human resolution

  @invariant
  Scenario: Stacked-to-Main is unaffected when archive-before-PR is not in effect
    Given archive-before-PR is not in effect because the artifact store is engram or archive is not applicable
    And the orchestrator selects Stacked-to-Main as the chain strategy during sdd-tasks
    When harness-workflow evaluates the strategy selection
    Then the orchestrator proceeds with Stacked-to-Main without forced convergence
    And no FBC tracker branch is required
    And the Stacked-to-Main Archive Convergence requirement does not apply
