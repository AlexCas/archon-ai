# Graphify Integration Specification

<!-- [[graphify-integration]] -->

> Session preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 · Playwright=off · Impeccable=off · Graphify=on.

## Purpose

Opt-in, default-off, advisory *code graph* gate backed by
[Graphify](https://github.com/Graphify-Labs/graphify) (Python, tree-sitter AST;
edges tagged `EXTRACTED`/`INFERRED`). Adds graph-informed repo comprehension to
`sdd-explore` and Leiden-community slice boundaries to `sdd-tasks` (Slice A),
plus advisory code graph diff in `sdd-verify` and optional edge evidence support
in `harness-judge` (Slice B). **Never blocks any phase, never returns a verdict.**

**Pinned version: `v0.9.45`** (GitHub `Graphify-Labs/graphify`, PyPI
`graphifyy==0.9.45`, published 2026-08-16). The string `"v8"` in the proposal
draft is an upstream *development branch name*, not a release; the release series
is `0.9.x`. Default stays pinned at `v0.9.45`; R-16 advisory note is the drift
mechanism. `output_dir` accepts any string — no path-format validation, consistent
with the no-`Load()`-validation decision.

All scenarios for each requirement are in
`graphify-integration.feature` alongside this file.

## Requirements

### R-01 Config Struct

The harness MUST add `Graphify` to `internal/config/config.go` (alongside
`Playwright`/`Security` scalar shape, lines 31–44) with exactly five YAML fields:

| Field | Type | Default |
|-------|------|---------|
| `enabled` | bool | `false` |
| `auto_install` | bool | `false` |
| `version` | string | `"v0.9.45"` |
| `output_dir` | string | `".archon/graphify"` |
| `semantic` | bool | `false` |

`Load()` MUST NOT add Graphify validation — no blocking verdict exists to
validate. `Clone()` MUST copy the struct as a value (line 147 pattern).

### R-02 `archon config` get/set

The harness MUST support `archon config set/get graphify.<field>` for all five
fields in `cmd/archon/config.go`. Both "supported keys" error strings (lines 297
and 348) MUST list the five new `graphify.*` keys.

### R-03 `archon status` Block

`internal/status/display.go` MUST render "Graphify (Code Graph)" after the
Impeccable block (line 65 region). When `enabled: false`: show `Enabled: false`
only. When `enabled: true`: also show `version`, `output_dir`, and `semantic`.

### R-04 `--graphify` Init Flag

`cmd/archon/main.go` MUST register `--graphify` (bool, default `false`, matching
line 84 `--impeccable` pattern). `internal/initcmd/init.go` MUST add `Graphify
bool` to `Options` (line 31 region) and wire it into `buildConfig`.

### R-05 Preflight Group G

`internal/initcmd/templates.go` MUST add group G after group F (lines 83–94
region) in both `orchestratorRules` blocks and the preflight section. Group G
MUST: be its own arrow-key `AskUserQuestion`; ask in Spanish "¿Activar Graphify
para análisis de grafo de código?"; pre-select "No (recomendado)"; include a
mapping paragraph ("Group G maps to `graphify.enabled` in `.archon/config.yaml`.
The `--graphify` flag at init time or the Graphify tab in `archon tui` set the
same value."); and update the preamble from "A–F"/"six" to "A–G"/"seven".

### R-06 Graphify Skill File

`skills/graphify/SKILL.md` MUST be a thin orchestration skill with: an Activation
Contract (load only when `graphify.enabled: true`); a Two Invocation Surfaces
table; graceful-degradation rules. MUST NOT reimplement AST parsing. Adding the
file auto-increments skill count via `embeddedSkillCount()` (lines 60–75 of
`cmd/archon/main.go`); no Go constant to bump.

### R-07 Advisory Absolute

For every one of the eight failure modes below, the consuming phase MUST emit a
single advisory note and continue with baseline grep/read. MUST NOT fail, MUST
NOT return `blocked`, MUST NOT halt the SDD flow.

**Failure modes:** (a) binary missing; (b) Python/uv/pipx absent; (c) `graphify
extract` non-zero exit; (d) `graph.json` absent or unreadable; (e) `graph.json`
unparseable or schema-drifted; (f) graph stale and binary unavailable for
re-extraction; (g) `output_dir` unwritable; (h) empty graph.

The `.feature` Scenario Outline covers all eight rows with a shared Examples
table.

### R-08 Inertness When Disabled

When `graphify.enabled: false`, the harness MUST NOT read `skills/graphify/
SKILL.md`, MUST NOT shell any `graphify` command, MUST NOT create `output_dir`,
and MUST NOT change any phase output. Fully inert, matching the Impeccable
inertness contract.

### R-09 Surface Separation

The harness MUST shell ONLY `graphify extract|query|path|explain`. `/graphify`
(agent slash-command) MUST NEVER be shelled — this is the documented Impeccable
failure mode (`skills/impeccable/SKILL.md` "Two Invocation Surfaces").
`python -m graphify.serve` (MCP) MUST NOT be a Slice A dependency; MAY be a
future option with the headless/cron auth caveat.

### R-10 `auto_install` Semantics

`auto_install: false` (default) + missing binary → advisory note naming `uv tool
install graphifyy` / `pipx install graphifyy`; MUST NOT install silently.
`auto_install: true` + missing binary → run install once, then proceed.

### R-11 `semantic: false` Means No LLM Calls

When `semantic: false` (default), the harness MUST NOT invoke any LLM API via
Graphify. Extraction uses pure local deterministic AST only. The code graph MUST
remain structurally queryable in this mode.

### R-12 Staleness Advisory + Auto Re-extraction

The skill MUST compare `graph.json` mtime against git `HEAD` commit timestamp.
**If stale:** `sdd-explore` MUST automatically re-run `graphify extract` (when
binary is present), refresh `graph-report.excerpt.md`, and emit "graph may be
stale — re-extracting". **If fresh** (mtime ≥ HEAD): reuse the existing graph
without re-extracting. No config knob for staleness policy.

### R-13 `sdd-explore` Advisory Consumption

When `graphify.enabled: true`:
- If graph is **fresh**: read `graph.json`/`GRAPH_REPORT.md` from `output_dir`;
  shell `graphify query`/`graphify explain` for targeted questions. No
  re-extraction.
- If graph is **absent**: shell `graphify extract` to produce it (if binary
  present), then read.
- If graph is **stale**: re-extract per R-12, then read.
- All failure modes fall back per R-07. MUST NOT shell `/graphify` or depend on
  MCP.

### R-14 `sdd-tasks` Leiden Communities (Read-Only)

When `graphify.enabled: true`, `sdd-tasks` SHOULD read Leiden community data from
`graph.json`/`GRAPH_REPORT.md` (read-only file access — **no shell invocation**).
`sdd-tasks` MUST NOT shell `graphify extract` or any other graphify command, **even
when the binary is present and graph.json is absent** — `sdd-explore` is the sole
extraction site in Slice A. Community data informs PR/slice boundary suggestions
fed to `chained-pr` and the review-budget split. Missing or unreadable community
data falls back to heuristic slice boundaries per R-07.

### R-15 Tracked Excerpt

`sdd-explore` MUST write a ≤40-line / ≤2 KB excerpt of `GRAPH_REPORT.md` to
`openspec/changes/<change-name>/graph-report.excerpt.md` (tracked in git; lives
in `openspec/`). Full `graph.json`, `graph.html`, `GRAPH_REPORT.md` remain in
`output_dir` (untracked via R-18). Excerpt MUST be refreshed on every
re-extraction, including auto re-extraction per R-12.

### R-16 Version Pin Advisory

When the installed `graphify` binary version does not match
`config.graphify.version`, the consuming phase MUST emit a single advisory note
and continue without blocking.

### R-17 Naming Discipline

All harness code, skill prose, and requirements MUST use "code graph"
exclusively for Graphify's AST output and "spec graph"/"vault graph" exclusively
for `internal/mapgen`'s `openspec/map.md` output. Terms MUST remain disjoint to
preserve Slice C feasibility.

### R-18 No `.gitignore` Edit Required

The default `output_dir` (`.archon/graphify/`) is already gitignored by
`.gitignore` **line 10** (`.archon/`; line 9 is the `# Local directories`
comment). The harness MUST NOT add a new `.gitignore` rule.

### R-19 Baseline Snapshot Capture

Before `sdd-verify` calls `graphify update <path>` to produce the post-apply graph,
it MUST copy `<output_dir>/graph.json` (the explore-phase graph) to
`<output_dir>/graph.baseline.json`. This establishes the "before" state for the
set-difference in R-20.

If `<output_dir>/graph.json` is absent at baseline-capture time (no explore-phase
extraction ran, or the file was removed between explore and verify), `sdd-verify`
MUST emit exactly one advisory note and skip both the re-extraction and the diff
section entirely — no CRITICAL severity, no verdict change, no `blocked`.

`graph.baseline.json` lives in the gitignored `.archon/graphify/` directory
(covered by `.gitignore` line 10 per Slice A R-18). No new `.gitignore` rule is
required.

### R-20 `sdd-verify` Advisory Code Graph Diff

When `graphify.enabled: true`, after the compliance matrix and before the final
verdict, `sdd-verify` MUST execute the following sequence:

1. Capture the baseline per R-19. If the baseline copy fails or the source is
   absent, fall back per R-07: emit one advisory note and skip the diff section.
2. Shell `graphify update <path>` (using the project root or `output_dir` as
   `<path>`) to produce a fresh post-apply `graph.json` in `output_dir`. This is
   the second extraction site introduced by Slice B — permitted by the R-22 §5
   amendment. All failure modes for this shell call fall back per R-07.
3. Parse both `graph.baseline.json` and the new `graph.json`. If either is
   unparseable or schema-drifted, emit one advisory note and skip the diff section.
4. Compute the set-difference:
   - **Added nodes**: node IDs present in `graph.json` but absent from
     `graph.baseline.json`.
   - **Removed nodes**: node IDs present in `graph.baseline.json` but absent
     from `graph.json`.
   - **Added edges**: `(source, target, relation)` tuples present in `graph.json`
     but absent from `graph.baseline.json`.
   - **Removed edges**: `(source, target, relation)` tuples present in
     `graph.baseline.json` but absent from `graph.json`.

   Node and edge IDs are stable across extractions (empirically confirmed in the
   proposal: `graphifyy==0.9.45`, two consecutive extracts, byte-identical output;
   a single added function produced exactly one new node, zero spurious churn;
   line info lives in `source_location`, not in IDs).

5. Emit a `### Code Graph Diff (advisory)` section in the verification report:
   - Severity: NOTE (never CRITICAL).
   - Content: counts per category plus up to 5 sample items per non-empty category.
   - Sample item format: node — `<node_id>`; edge —
     `<source> →[<relation>]→ <target> (EXTRACTED)`.
   - When all four counts are zero: emit the line
     `No structural changes detected in the code graph.`

6. The diff NOTE MUST NOT alter the PASS / PASS WITH WARNINGS / FAIL verdict on its
   own. A diff with non-zero counts is advisory context only — not a new failure
   condition.

7. `sdd-verify` MUST re-snapshot and re-extract on each re-apply retry rather than
   reusing a stale baseline or after-file. Each verify run starts the sequence at
   step 1 above.

When `graphify.enabled: false`, skip this requirement entirely: no extraction, no
`### Code Graph Diff` section, no change to verify output. Fully inert, matching
the Impeccable inertness contract (Slice A R-08).

### R-21 `harness-judge` Edge Evidence (Advisory Only)

When `graphify.enabled: true` and `judgment-day` is running, the judge subagent MAY:

1. Read `<output_dir>/graph.json` (the post-apply graph produced by R-20's
   extraction step) or call `graphify query`/`graphify explain` to look up callers,
   callees, or import relationships for any symbol under review.
2. Cite `EXTRACTED`-confidence edges in the description of a finding to ground it in
   deterministic AST facts. Example citation format:
   `func X is no longer called by Y — edge Y→[calls]→X removed (EXTRACTED, code graph)`.
3. When `semantic: true` is configured, `INFERRED` edges MAY also be cited but MUST
   be labeled `(INFERRED, semantic)` to distinguish them from deterministic facts. In
   the default `semantic: false` mode only `EXTRACTED` edges exist; no `INFERRED`
   citation is possible.

The following constraints apply without exception:

- Edge evidence MUST NOT become a new column in the Step 4 pass/fail result table in
  `harness-judge`.
- Edge evidence MUST NOT be added as a blocking condition in the harness-judge
  Decision Gates table.
- Edge evidence MUST NOT trigger the re-apply loop by itself. The re-apply loop is
  driven by Step 4 fail verdicts from `judgment-day` and enabled gates (mutation /
  Playwright / Impeccable) — never by graph findings.
- Judges MUST cite edge evidence only to enrich finding descriptions they reached
  independently through normal dual-review. Edge evidence must never be the sole
  reason a finding is raised.
- `harness-judge` MUST NOT shell `graphify update` or any extraction command. It is
  read-only with respect to the code graph (file read or `graphify query`/`graphify
  explain` only — never extract).

When `graphify.enabled: false`, judges MUST NOT access `graph.json` or call any
`graphify` command; no edge citations may appear in any finding description.

### R-22 `graphify` SKILL.md Amendments (Slice B)

`skills/graphify/SKILL.md` MUST be amended to reflect the Slice B relaxations and
new phase rows. Three sections require changes.

**§3 Per-Phase Invocation Map** — add two rows after the existing `sdd-tasks` row:

| Phase | Action | Mechanism |
|-------|--------|-----------|
| `sdd-verify` | Snapshot baseline (copy `graph.json` → `graph.baseline.json`); shell `graphify update <path>` for post-apply graph; compute set-diff; emit advisory NOTE per R-20. All failure modes fall back per §6 rows i–j. | file copy + shell CLI |
| `harness-judge` | Read `graph.json` or call `graphify query`/`graphify explain` to look up edge evidence for findings per R-21. | file read + shell CLI (query/explain only — **never extract**) |

Also remove or replace the sentence immediately following the existing table:

> `sdd-explore` is the **sole extraction site**. No other phase shells any `graphify` command.

Replace it with the updated statement from §5 below.

**§5 Extraction-Site Invariant (Slice B amendment)** — the section currently titled
"`sdd-tasks` Read-Only Guarantee" MUST have its opening statement amended. Replace:

> `sdd-explore` is the sole extraction site in Slice A; `sdd-tasks` only ever opens
> `graph.json`/`GRAPH_REPORT.md` for a read.

with:

> **Slice B amendment**: `sdd-explore` and `sdd-verify` are the two extraction sites.
> No other phase shells `graphify update` or any extraction command. `sdd-tasks` and
> `harness-judge` are read-only surfaces — file read or `graphify query`/`graphify
> explain` only; they MUST NOT shell any extraction command.

The remainder of the §5 body (the `sdd-tasks` read-only guarantee and its rationale)
remains unchanged.

**§6 Advisory-Degradation Table** — add two new rows:

| # | Failure mode | Advisory note | Fallback |
|---|---|---|---|
| i | `graph.baseline.json` absent at verify diff time (baseline copy not captured or source `graph.json` was absent) | `code graph baseline absent; skipping graph diff` | skip diff section; verify continues |
| j | Diff compute error (parse failure of `graph.baseline.json` or post-apply `graph.json`, or schema mismatch between the two) | `code graph diff failed (parse/schema); skipping graph diff` | skip diff section; verify continues |

### R-23 TUI Tab

The harness MUST provide a "Graphify" tab in `archon tui` (`internal/tui/graphify_tab.go`)
that exposes all five `config.Graphify` fields for interactive editing and saving,
mirroring the Impeccable tab structure.

Focus order (bools-first, `graphifyFocusCount = 5`):

| Focus | Field | Control |
|-------|-------|---------|
| 0 | `enabled` | toggle (Enter/Space) |
| 1 | `auto_install` | toggle (Enter/Space) |
| 2 | `semantic` | toggle (Enter/Space) |
| 3 | `version` | textinput |
| 4 | `output_dir` | textinput |

`applyToConfig` MUST coerce a blank `version` input to `DefaultGraphifyVersion`
(`"v0.9.45"`) and a blank `output_dir` input to `DefaultGraphifyOutputDir`
(`".archon/graphify"`). MUST NOT persist `""` for either field.

The tab MUST be wired into `model.go` at all nine canonical sites: `Tab` iota
constant, `Model` field, `NewModel` ctor, two `setWidth` fan-outs
(`WindowSizeMsg` + `agentInitDoneMsg`), key-dispatch switch, `agentInitDoneMsg`
rebuild, `saveConfig` applyToConfig fan-out, `renderTabs` label slice,
`renderTabContent` switch case. No live install probe; no blocking verdict —
parity with the Impeccable tab.

`TestModel_Update_ShiftTabWrapsFromAgent` MUST be updated (Shift+Tab from AgentTab
now wraps to GraphifyTab, the new last tab). `TestModel_renderTabs_Order` MUST be
updated (append `"Graphify"` to the expected label list). `TestGraphifyTabState_ApplyToConfig`
MUST be added to `model_test.go` alongside `TestImpeccableTabState_ApplyToConfig`.

---

## Deferred (Out of Scope)
- **Slice C**: bridge code graph ↔ spec graph (`internal/mapgen`/`openspec/map.md`).
- **MCP surface** (`python -m graphify.serve`): future opt-in; headless/cron auth caveat.
- **`skill-registry` source edit**: auto-scans `*/SKILL.md`; indexing is automatic.
