# harness-testing Specification

## Purpose

Author use cases as formal Gherkin, detect web projects, and generate/execute
Playwright E2E tests from those use cases when enabled.

## Requirements

### Requirement: Formal Gherkin feature files

`sdd-spec` MUST author scenarios in formal Gherkin and produce a `{domain}.feature`
file beside each `spec.md`, using `Feature:`, `Scenario:`/`Scenario Outline:`,
`Background:`, and the keywords `Given/When/Then/And/But`. Every requirement MUST map
to at least one scenario with a matching name. Scenarios SHOULD be tagged
(`@happy`, `@edge`, `@error`, `@web`).

#### Scenario: Spec phase emits a feature file

```gherkin
Scenario: Spec phase emits a feature file
  Given a change with one affected domain
  When sdd-spec runs in openspec mode
  Then a "{domain}.feature" file exists beside "spec.md"
  And every requirement maps to at least one Gherkin scenario
```

### Requirement: Web project detection

`sdd-explore` MUST determine whether the project is `web`, `not-web`, or `unknown`.
For a NEW or blank project where the type is `unknown`, the orchestrator MUST ASK the
user (preflight group E) alongside the rhythm. Playwright MUST never be enabled for a
`not-web` project.

#### Scenario: Blank project triggers a preflight question

```gherkin
Scenario: Blank project triggers a preflight question
  Given a new project where the type cannot be determined from code
  When sdd-explore reports project type "unknown"
  Then the orchestrator asks preflight group E before proceeding
```

### Requirement: Playwright generation from Gherkin

When `playwright.enabled` is true and the project is web, `sdd-apply` MUST generate
Playwright specs from the `@web` Gherkin scenarios into `playwright.test_dir`, keeping
scenario names so failures trace back to the `.feature`. Generation MUST NOT execute
the suite.

#### Scenario: Apply generates Playwright specs

```gherkin
Scenario: Apply generates Playwright specs
  Given playwright.enabled is true and a web project
  And a feature file with a "@web" scenario
  When sdd-apply implements the related tasks
  Then a Playwright spec is generated in the configured test_dir
  And the suite is not executed during apply
```

### Requirement: Playwright execution after verify and judge

When `playwright.enabled` is true, the generated Playwright suite MUST run as a
judge-phase gate, AFTER `sdd-verify` and AFTER `judgment-day` passes. A failing
scenario MUST fail the gate and feed the auto re-apply loop. When disabled, the gate
is skipped.

#### Scenario: Playwright gate runs after judgment-day passes

```gherkin
Scenario: Playwright gate runs after judgment-day passes
  Given playwright.enabled is true
  And judgment-day has passed for the change
  When the judge phase evaluates gates
  Then the Playwright suite is executed
  And a failing scenario produces a fail verdict that enters the re-apply loop
```
