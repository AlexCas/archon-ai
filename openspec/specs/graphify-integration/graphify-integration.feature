Feature: Graphify Integration (Slice A & B — advisory code-graph gate with graph diff and edge evidence)
  Opt-in, default-off, advisory code-graph gate. When enabled, sdd-explore
  gains graph-informed repo comprehension and sdd-tasks gains Leiden-community
  slice boundaries (Slice A). Additionally, sdd-verify emits an advisory code
  graph diff section and harness-judge may cite EXTRACTED edges to enrich
  findings (Slice B). Never blocks any phase, never returns a verdict.

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

  # ── R-19: Baseline Snapshot Capture ──────────────────────────────────────────

  @happy
  Scenario: sdd-verify copies graph.json to graph.baseline.json before re-extracting
    Given graphify.enabled is true
    And output_dir/graph.json exists from sdd-explore
    When sdd-verify begins the code graph diff step
    Then output_dir/graph.baseline.json is created as a byte-for-byte copy of graph.json

  @edge
  Scenario: sdd-verify skips diff and emits advisory when baseline graph.json is absent
    Given graphify.enabled is true
    And output_dir/graph.json is absent
    When sdd-verify runs
    Then the phase emits exactly one advisory note referencing the missing baseline
    And the "### Code Graph Diff (advisory)" section is omitted from the verify report
    And the PASS/FAIL verdict is not altered

  # ── R-20: sdd-verify Advisory Code Graph Diff ────────────────────────────────

  @happy
  Scenario: sdd-verify emits advisory diff section when apply introduced structural changes
    Given graphify.enabled is true
    And output_dir/graph.baseline.json captures the pre-apply graph
    And "graphify update <path>" produces a post-apply graph.json with at least one structural change
    When sdd-verify runs the code graph diff step
    Then the verify report contains a "### Code Graph Diff (advisory)" section
    And the section lists counts for added nodes, removed nodes, added edges, and removed edges
    And the section severity is NOTE and not CRITICAL
    And the PASS/FAIL verdict is not altered by the diff counts

  @happy
  Scenario: sdd-verify emits "no structural changes" when diff is empty
    Given graphify.enabled is true
    And the post-apply graph.json is structurally identical to graph.baseline.json
    When sdd-verify runs the code graph diff step
    Then the "### Code Graph Diff (advisory)" section contains "No structural changes detected in the code graph."
    And the PASS/FAIL verdict is not altered

  @happy
  Scenario: sdd-verify includes up to 5 sample items per non-empty diff category
    Given graphify.enabled is true
    And the post-apply graph.json has 10 removed edges and 2 added nodes
    When sdd-verify runs the code graph diff step
    Then the removed-edges listing shows at most 5 sample items
    And the added-nodes listing shows all 2 items
    And node samples use format "<node_id>"
    And edge samples use format "<source> →[<relation>]→ <target> (EXTRACTED)"

  @edge
  Scenario: sdd-verify emits advisory and skips diff section when graphify update exits non-zero
    Given graphify.enabled is true
    And output_dir/graph.baseline.json is present
    And "graphify update <path>" exits with a non-zero code
    When sdd-verify runs
    Then the phase emits exactly one advisory note about the extraction failure
    And the "### Code Graph Diff (advisory)" section is omitted from the verify report
    And the PASS/FAIL verdict is not altered
    And the phase completes successfully

  @edge
  Scenario: sdd-verify emits advisory and skips diff when graph.json is unparseable after update
    Given graphify.enabled is true
    And output_dir/graph.baseline.json is present
    And "graphify update <path>" succeeds but produces invalid JSON
    When sdd-verify runs the code graph diff step
    Then the phase emits exactly one advisory note about the parse failure
    And the "### Code Graph Diff (advisory)" section is omitted from the verify report
    And the PASS/FAIL verdict is not altered

  @edge
  Scenario: sdd-verify re-snapshots and re-extracts fresh on each re-apply retry
    Given graphify.enabled is true
    And sdd-apply has run a second time (retry 2 of the re-apply loop)
    When sdd-verify runs again
    Then a fresh graph.baseline.json is captured before re-extraction
    And the diff is recomputed against the new baseline
    And the prior graph.baseline.json from retry 1 is overwritten

  @happy
  Scenario: graphify.enabled false — sdd-verify produces no graph diff section
    Given graphify.enabled is false
    When sdd-verify runs
    Then no graphify commands are executed by sdd-verify
    And the verify report contains no "### Code Graph Diff (advisory)" section
    And verify output is byte-identical to the non-graphify baseline

  # ── R-21: harness-judge Edge Evidence (Advisory Only) ─────────────────────────

  @happy
  Scenario: judge cites an EXTRACTED edge to enrich a finding it independently confirmed
    Given graphify.enabled is true
    And output_dir/graph.json contains a "calls" edge from function Y to function X (EXTRACTED)
    And judgment-day independently identifies a dead-code finding for X through normal review
    When the judge describes the finding
    Then the description may include "edge Y→[calls]→X removed (EXTRACTED, code graph)"
    And no new column appears in the Step 4 pass/fail result table
    And the cited edge does not trigger the re-apply loop by itself

  @happy
  Scenario: INFERRED edge citation is labeled "(INFERRED, semantic)" when semantic is true
    Given graphify.enabled is true and semantic is true
    And output_dir/graph.json contains an INFERRED edge
    And judgment-day cites that edge in a finding description
    When the finding is emitted
    Then the citation carries the label "(INFERRED, semantic)"

  @happy
  Scenario: harness-judge does not gain a new gate column from graph evidence
    Given graphify.enabled is true
    And output_dir/graph.json shows several removed edges
    When harness-judge runs to completion
    Then the Step 4 result table contains only: judgment-day, mutation gate, playwright gate, impeccable gate
    And graph edge evidence is not listed as a blocking condition in the Decision Gates table

  @happy
  Scenario: harness-judge does not trigger re-apply loop from graph evidence alone
    Given graphify.enabled is true
    And output_dir/graph.json shows removed edges for multiple functions
    And judgment-day passes on its own merits
    When harness-judge processes edge evidence
    Then the re-apply loop is not triggered
    And the overall judge result is pass (assuming other gates pass)

  @happy
  Scenario: harness-judge never shells graphify update or any extraction command
    Given graphify.enabled is true
    When harness-judge queries the code graph for edge evidence
    Then only "graphify query" or "graphify explain" are permitted shell invocations
    And "graphify update" is never invoked by harness-judge

  @happy
  Scenario: graphify.enabled false — judges emit no edge citations
    Given graphify.enabled is false
    When harness-judge runs
    Then no graphify commands are executed by harness-judge
    And no edge citations appear in any finding description
    And judge output is byte-identical to the non-graphify baseline

  # ── R-22: graphify SKILL.md Amendments (Slice B) ─────────────────────────────

  @happy
  Scenario: graphify SKILL.md section 3 contains sdd-verify and harness-judge rows
    Given skills/graphify/SKILL.md has been amended for Slice B
    When the Per-Phase Invocation Map (section 3) is read
    Then it contains a "sdd-verify" row describing baseline snapshot and graphify update
    And it contains a "harness-judge" row describing read-only evidence query
    And the harness-judge row states "never extract"

  @happy
  Scenario: graphify SKILL.md section 5 reflects the two-extraction-site invariant
    Given skills/graphify/SKILL.md has been amended for Slice B
    When section 5 is read
    Then it states "sdd-explore and sdd-verify are the two extraction sites"
    And it preserves the sdd-tasks read-only guarantee unchanged
    And it states harness-judge must not shell any extraction command

  @happy
  Scenario: graphify SKILL.md section 6 contains degradation rows i and j
    Given skills/graphify/SKILL.md has been amended for Slice B
    When the Advisory-Degradation Table (section 6) is read
    Then it contains a row for "graph.baseline.json absent at verify diff time"
    And it contains a row for "diff compute error (parse failure or schema mismatch)"
    And both new rows show "skip diff section" as the fallback behavior
    And both new rows show the advisory note text without CRITICAL severity
