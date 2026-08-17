# Graphify Integration Specification

<!-- [[graphify-integration]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

> Session preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 · Playwright=off · Impeccable=off.

## Purpose

Opt-in, default-off, advisory *code graph* gate backed by
[Graphify](https://github.com/Graphify-Labs/graphify) (Python, tree-sitter AST;
edges tagged `EXTRACTED`/`INFERRED`). Adds graph-informed repo comprehension to
`sdd-explore` and Leiden-community slice boundaries to `sdd-tasks`. **Never
blocks any phase, never returns a verdict.**

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
The `--graphify` flag at init time sets the same value."); and update the preamble
from "A–F"/"six" to "A–G"/"seven".

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

---

## Deferred (Out of Scope)

- **TUI tab** (`internal/tui/graphify_tab.go` + `model.go` wiring + tests, ~320 lines).
- **Slice B**: structural code-graph diff in `sdd-verify`; edge evidence in `harness-judge`.
- **Slice C**: bridge code graph ↔ spec graph (`internal/mapgen`/`openspec/map.md`).
- **MCP surface** (`python -m graphify.serve`): future opt-in; headless/cron auth caveat.
- **`skill-registry` source edit**: auto-scans `*/SKILL.md`; indexing is automatic.
