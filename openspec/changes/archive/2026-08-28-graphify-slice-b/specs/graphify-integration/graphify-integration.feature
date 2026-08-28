Feature: Graphify Integration — Slice B (graph diff in verify, edge evidence in judge)
  Advisory-only capability delta extending R-01–R-18 (Slice A). When
  graphify.enabled is true, sdd-verify surfaces a structural code-graph diff
  after apply (never alters PASS/FAIL), and harness-judge may cite EXTRACTED
  edges to enrich findings (never a new gate or column). All behavior is fully
  inert when graphify.enabled is false.

  Background:
    Given an archon project with a valid .archon/config.yaml

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
