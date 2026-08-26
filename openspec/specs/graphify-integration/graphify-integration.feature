Feature: Graphify Integration (Slice A — advisory code-graph gate)
  Opt-in, default-off, advisory code-graph gate. When enabled, sdd-explore
  gains graph-informed repo comprehension and sdd-tasks gains Leiden-community
  slice boundaries. Never blocks any phase.

  Background:
    Given an archon project with a valid .archon/config.yaml

  # ── R-01: Config Struct ──────────────────────────────────────────────────────

  @happy
  Scenario: Graphify config fields survive Load and Clone
    Given a config file with all five graphify fields set to non-default values
    When Load() parses the file and Clone() copies the result
    Then all five fields match the originals in both loaded and cloned config

  @happy
  Scenario: Absent graphify block defaults all fields
    Given a config file with no graphify block
    When Load() parses the file
    Then Graphify.Enabled is false, Version is "v0.9.45", OutputDir is ".archon/graphify"

  # ── R-02: archon config get/set ──────────────────────────────────────────────

  @happy
  Scenario Outline: graphify.<field> round-trips through set and get
    When the user runs "archon config set graphify.<field> <value>"
    Then "archon config get graphify.<field>" returns "<value>"

    Examples:
      | field        | value            |
      | enabled      | true             |
      | auto_install | true             |
      | version      | v0.9.45          |
      | output_dir   | .archon/graphify |
      | semantic     | true             |

  @error
  Scenario: Unknown graphify key produces error listing supported keys
    When the user runs "archon config set graphify.severity block"
    Then the command returns an error listing the supported graphify.* keys

  # ── R-03: archon status Block ─────────────────────────────────────────────────

  @happy
  Scenario: Status shows Graphify section when disabled
    Given graphify.enabled is false
    When the user runs "archon status"
    Then the output contains "Graphify (Code Graph)" and "Enabled:   false"

  @happy
  Scenario: Status shows graphify details when enabled
    Given graphify.enabled true, version "v0.9.45", output_dir ".archon/graphify"
    When the user runs "archon status"
    Then the output contains "v0.9.45" and ".archon/graphify"

  # ── R-04: --graphify init flag ───────────────────────────────────────────────

  @happy
  Scenario: archon init --graphify writes enabled true
    When the user runs "archon init --graphify"
    Then .archon/config.yaml contains graphify.enabled: true

  @happy
  Scenario: Default init leaves graphify disabled
    When the user runs "archon init" without --graphify
    Then .archon/config.yaml has graphify.enabled false or the graphify block is absent

  # ── R-05: Preflight group G ──────────────────────────────────────────────────

  @happy
  Scenario: Graphify tab appears in archon tui tab bar
    Given archon tui is running
    When the user navigates through the tab bar
    Then a "Graphify" tab is listed after the "Impeccable" tab

  @happy
  Scenario: All five config.Graphify fields are visible in the Graphify tab
    Given the user has selected the Graphify tab
    When the tab renders
    Then the tab displays controls for enabled, auto_install, semantic, version, and output_dir

  @happy
  Scenario: Toggling enabled with Enter updates the field
    Given the Graphify tab is focused at index 0 (enabled)
    When the user presses Enter
    Then the enabled toggle flips to its opposite value

  @happy
  Scenario: Toggling auto_install with Space updates the field
    Given the Graphify tab is focused at index 1 (auto_install)
    When the user presses Space
    Then the auto_install toggle flips to its opposite value

  @happy
  Scenario: Toggling semantic at focus index 2 updates the field
    Given the Graphify tab is focused at index 2 (semantic)
    When the user presses Enter
    Then the semantic toggle flips to its opposite value

  @happy
  Scenario: Editing version text input updates config on save
    Given the Graphify tab is focused at index 3 (version)
    When the user types "v0.9.50" and saves
    Then config.Graphify.Version equals "v0.9.50"

  @happy
  Scenario: Editing output_dir text input updates config on save
    Given the Graphify tab is focused at index 4 (output_dir)
    When the user types ".archon/custom" and saves
    Then config.Graphify.OutputDir equals ".archon/custom"

  @edge
  Scenario: Blank version input coerces to DefaultGraphifyVersion on save
    Given the Graphify tab has an empty version field
    When the user saves
    Then config.Graphify.Version equals "v0.9.45"
    And the saved YAML does not contain an empty version key

  @edge
  Scenario: Blank output_dir input coerces to DefaultGraphifyOutputDir on save
    Given the Graphify tab has an empty output_dir field
    When the user saves
    Then config.Graphify.OutputDir equals ".archon/graphify"
    And the saved YAML does not contain an empty output_dir key

  @happy
  Scenario: Graphify tab wired in model — Shift+Tab from AgentTab wraps to GraphifyTab
    Given archon tui is running with GraphifyTab as the last tab
    When the user presses Shift+Tab from AgentTab
    Then the focused tab is GraphifyTab

  @happy
  Scenario: renderTabs lists all tab labels in order including Graphify
    Given archon tui renders the tab bar
    When renderTabs is called
    Then the label list ends with "Impeccable" followed by "Graphify"

  @happy
  Scenario: Tab resize propagates to Graphify tab inputs
    Given the Graphify tab is active
    When a WindowSizeMsg is received with a new width
    Then the Graphify tab's text inputs resize to the new width

  # ── R-05 (MODIFIED): Preflight Group G mapping paragraph ─────────────────────

  @happy
  Scenario: Group G mapping paragraph references both the init flag and the TUI tab
    Given internal/initcmd/templates.go and root CLAUDE.md are current
    When the orchestrator renders CLAUDE.md
    Then the group G mapping paragraph contains "The `--graphify` flag at init time or the Graphify tab in `archon tui` set the same value."
    And the file still contains "¿Activar Graphify para análisis de grafo de código?"
    And the group count reads "seven" and the range reads "A-G"

  # ── R-06: Graphify skill file ────────────────────────────────────────────────

  @happy
  Scenario: SKILL.md auto-increments skill count and contains required sections
    Given skills/graphify/SKILL.md is present
    And the file contains an Activation Contract and a Two Invocation Surfaces table
    When "archon status" runs
    Then Skill Count is one greater than before the file was added

  # ── R-07: Advisory absolute — all failure modes (Scenario Outline) ───────────

  @error
  Scenario Outline: Every failure mode emits an advisory note and lets the phase continue
    Given graphify.enabled is true
    And <failure_condition>
    When sdd-explore runs
    Then the phase emits an advisory note
    And the phase completes successfully using baseline grep/read
    And the phase does not return blocked and does not fail

    Examples:
      | failure_condition                                   |
      | the graphify binary is not on PATH                  |
      | Python and uv and pipx are absent                   |
      | "graphify extract" exits with a non-zero code       |
      | graph.json is absent or unreadable (IO/permissions) |
      | graph.json contains invalid JSON                    |
      | graph.json has zero nodes and zero edges            |
      | output_dir is not writable                          |
      | graph.json schema does not match the expected shape |

  # ── R-08: Inertness when disabled ────────────────────────────────────────────

  @happy
  Scenario: graphify.enabled false is fully inert across all phases
    Given graphify.enabled is false
    When sdd-explore and sdd-tasks run
    Then no graphify commands are executed
    And .archon/graphify/ is not created
    And phase outputs are byte-identical to the non-graphify baseline

  # ── R-09: Surface separation ─────────────────────────────────────────────────

  @happy
  Scenario: Extraction uses graphify CLI, not /graphify or MCP
    Given graphify.enabled true and the binary is available
    When sdd-explore triggers extraction
    Then the harness shells "graphify extract"
    And "/graphify" is never invoked as a shell command
    And "python -m graphify.serve" is never invoked

  # ── R-10: auto_install semantics ─────────────────────────────────────────────

  @error
  Scenario: auto_install false — advisory note with install command, no install
    Given graphify.enabled true, auto_install false, binary not on PATH
    When sdd-explore runs
    Then the phase emits a note containing "uv tool install graphifyy"
    And no install command is executed

  @happy
  Scenario: auto_install true — install runs once then extraction proceeds
    Given graphify.enabled true, auto_install true, binary not on PATH
    When sdd-explore runs
    Then the harness executes the install command once
    And then proceeds with graphify extraction

  # ── R-11: semantic: false means no LLM calls ─────────────────────────────────

  @happy
  Scenario: semantic false — pure AST graph, no LLM API calls
    Given graphify.enabled true and semantic false
    When sdd-explore triggers extraction
    Then no LLM API calls are made via graphify
    And graph.json contains structural AST data queryable without LLM

  # ── R-12 + R-13: Staleness, auto re-extraction, and sdd-explore consumption ──

  @edge
  Scenario: Stale graph triggers automatic re-extraction and excerpt refresh
    Given graphify.enabled true and graph.json mtime predates HEAD commit
    When sdd-explore runs
    Then sdd-explore re-runs "graphify extract"
    And refreshes graph-report.excerpt.md
    And emits "graph may be stale — re-extracting"

  @happy
  Scenario: Fresh graph is reused without re-extracting
    Given graphify.enabled true and graph.json mtime is at or after HEAD commit
    When sdd-explore runs
    Then sdd-explore reads graph.json for repo comprehension
    And "graphify extract" is not re-run

  @happy
  Scenario: sdd-explore extracts graph when absent and binary is present
    Given graphify.enabled true, graph.json absent, binary on PATH
    When sdd-explore runs
    Then the harness shells "graphify extract" to produce the graph
    And sdd-explore reads the resulting graph.json for comprehension

  # ── R-14: sdd-tasks Leiden communities (strictly read-only) ──────────────────

  @happy
  Scenario: sdd-tasks reads communities read-only, never shells graphify
    Given graphify.enabled true and graph.json contains Leiden community data
    When sdd-tasks runs
    Then the task breakdown reflects community boundaries as PR/slice suggestions
    And no graphify CLI command is executed by sdd-tasks

  @error
  Scenario: sdd-tasks falls back and never shells even when binary present and graph absent
    Given graphify.enabled true, graph.json absent, binary on PATH
    When sdd-tasks runs
    Then sdd-tasks uses heuristic slice boundaries
    And emits an advisory note about the missing graph
    And does not shell "graphify extract" or any graphify command

  # ── R-15: Tracked excerpt ────────────────────────────────────────────────────

  @happy
  Scenario: sdd-explore writes tracked excerpt after successful extraction
    Given graphify.enabled true and extraction succeeds
    When sdd-explore completes
    Then openspec/changes/<change-name>/graph-report.excerpt.md exists and is at most 40 lines
    And graph.json and graph.html remain in .archon/graphify/ untracked

  @edge
  Scenario: Excerpt is refreshed when auto re-extraction occurs
    Given graphify.enabled true and a prior excerpt exists
    When sdd-explore auto-re-extracts because the graph is stale
    Then graph-report.excerpt.md is overwritten with the new excerpt

  # ── R-16: Version pin advisory ───────────────────────────────────────────────

  @edge
  Scenario: Version mismatch emits advisory note, phase continues without blocking
    Given graphify.enabled true, config.graphify.version "v0.9.45", binary reports "v0.9.44"
    When sdd-explore runs
    Then the phase emits an advisory note about the version mismatch
    And the phase continues without blocking

  # ── R-17: Naming discipline ──────────────────────────────────────────────────

  @happy
  Scenario: skills/graphify/SKILL.md uses "code graph" exclusively for Graphify output
    Given skills/graphify/SKILL.md is written
    Then the file uses "code graph" for Graphify AST output
    And does not use "spec graph" or "vault graph" to refer to Graphify output

  # ── R-18: No .gitignore edit required ────────────────────────────────────────

  @happy
  Scenario: graph.json ignored by existing .gitignore line 10, no modification needed
    Given graphify runs and produces graph.json in .archon/graphify/
    When "git status" is run in the project root
    Then graph.json does not appear as untracked
    And .gitignore has not been modified
