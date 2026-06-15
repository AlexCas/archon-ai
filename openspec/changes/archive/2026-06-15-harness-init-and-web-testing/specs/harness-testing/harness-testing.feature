Feature: Gherkin specs and Playwright web testing
  Use cases are authored in formal Gherkin; web projects generate and run
  Playwright tests derived from those use cases.

  Scenario: Spec phase emits a feature file
    Given a change with one affected domain
    When sdd-spec runs in openspec mode
    Then a "{domain}.feature" file exists beside "spec.md"
    And every requirement maps to at least one Gherkin scenario

  Scenario: Blank project triggers a preflight question
    Given a new project where the type cannot be determined from code
    When sdd-explore reports project type "unknown"
    Then the orchestrator asks preflight group E before proceeding

  @web
  Scenario: Apply generates Playwright specs
    Given playwright.enabled is true and a web project
    And a feature file with a "@web" scenario
    When sdd-apply implements the related tasks
    Then a Playwright spec is generated in the configured test_dir
    And the suite is not executed during apply

  @web
  Scenario: Playwright gate runs after judgment-day passes
    Given playwright.enabled is true
    And judgment-day has passed for the change
    When the judge phase evaluates gates
    Then the Playwright suite is executed
    And a failing scenario produces a fail verdict that enters the re-apply loop

  Scenario: Playwright gate skipped when disabled
    Given playwright.enabled is false
    When the judge phase evaluates gates
    Then the Playwright gate is skipped and counts as pass
