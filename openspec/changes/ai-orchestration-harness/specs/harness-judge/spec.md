# harness-judge Specification

## Purpose

The `harness-judge` meta-skill wraps `judgment-day` for dual adversarial review and optionally integrates mutation testing. On failure, it automatically re-runs `sdd-apply` with the judge's structured feedback.

## Requirements

### Requirement: Judgment-Day Wrapping

The meta-skill MUST invoke `judgment-day` on the current change and capture its verdict.

#### Scenario: Judgment-day passes

- GIVEN `sdd-verify` completed successfully for change `my-feature`
- WHEN `harness-judge` runs
- THEN `judgment-day` dual review is invoked
- AND if both judges approve, the verdict is `pass`
- AND `state.yaml` advances to `judge` phase with status `completed`

#### Scenario: Judgment-day finds issues

- GIVEN `judgment-day` returns one or more issues
- WHEN `harness-judge` processes the verdict
- THEN the verdict is `fail`
- AND all issues are collected into structured feedback

### Requirement: Auto-Re-Run on Failure

When judgment fails, the meta-skill MUST automatically re-run `sdd-apply` with the judge's feedback.

#### Scenario: Failed judgment triggers auto-re-apply

- GIVEN `harness-judge` returns verdict `fail` with structured feedback
- WHEN the meta-skill processes the failure
- THEN `sdd-apply` is automatically invoked with the judge's feedback as the input prompt
- AND the orchestrator does NOT pause for manual intervention
- AND after re-apply, `sdd-verify` runs automatically

#### Scenario: Auto-re-apply succeeds then re-judge

- GIVEN `sdd-apply` completed with judge feedback applied
- AND `sdd-verify` passes
- WHEN `harness-judge` detects the re-apply cycle completed
- THEN `judgment-day` is invoked again automatically
- AND the loop continues until verdict is `pass` or max retries reached

#### Scenario: Max retries exhausted

- GIVEN 3 consecutive judgment failures with re-apply loops
- WHEN the 4th judgment also fails
- THEN `harness-judge` returns `blocked` with all accumulated issues
- AND reports `max_retries_exceeded: true`

### Requirement: Mutation Testing Integration

Mutation testing MUST be opt-in and configurable via `.archon/config.yaml`.

#### Scenario: Mutation testing enabled and passes

- GIVEN `.archon/config.yaml` contains `mutation_testing: {enabled: true, tool: "gremlins"}`
- AND `judgment-day` passes
- WHEN `harness-judge` runs the mutation gate
- THEN the configured mutation tool is invoked on changed Go files
- AND if mutation score meets the threshold, the gate passes

#### Scenario: Mutation testing disabled

- GIVEN `.archon/config.yaml` contains `mutation_testing: {enabled: false}`
- WHEN `harness-judge` runs
- THEN the mutation testing step is skipped entirely
- AND only `judgment-day` verdict determines the result

#### Scenario: Mutation score below threshold

- GIVEN mutation testing is enabled with threshold `0.80`
- AND the mutation tool reports score `0.72`
- WHEN `harness-judge` processes the result
- THEN the gate fails with structured feedback listing surviving mutants
- AND the verdict is `fail` even if `judgment-day` passed

### Requirement: Feedback Format

Judge feedback MUST be structured so `sdd-apply` can consume it directly.

#### Scenario: Structured feedback block

- GIVEN judgment-day finds 2 issues and mutation testing finds 3 surviving mutants
- WHEN `harness-judge` produces feedback
- THEN the output contains a `## Issues` section with bulleted issue descriptions
- AND a `## Action Required` section with ONE actionable directive per issue
- AND each directive maps to a specific file and requirement
