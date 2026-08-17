---
name: graphify
description: "Trigger: graphify.enabled, code graph, Leiden communities. Orchestrate WHEN and HOW Archon calls the external `graphify` tool across SDD phases."
license: MIT
metadata:
  
  version: "1.0"
  scope: reference
---

## 1. Activation Contract

Load this skill from `sdd-explore` or `sdd-tasks` only when `.archon/config.yaml`
→ `graphify.enabled: true`. When the flag is absent or `false`, no phase reads
this file, no `graphify` command is shelled, `output_dir` (default
`.archon/graphify`) is never created, and no phase output changes. Fully
inert, matching the Impeccable inertness contract (see
`skills/impeccable/SKILL.md`).

This skill is a **thin orchestration layer**. It does NOT reimplement
[Graphify](https://github.com/Graphify-Labs/graphify)'s tree-sitter AST
extraction or its Leiden community detection. Every actual check runs inside
the `graphify` Python tool — Archon only decides when to call it, how to read
its output, and how to degrade when it is unavailable. Pinned version:
`v0.9.45` (PyPI `graphifyy==0.9.45`).

## 2. Two Invocation Surfaces (do not confuse them)

Graphify exposes distinct surfaces. Mixing them up is the documented
Impeccable failure mode (see `skills/impeccable/SKILL.md` "Two Invocation
Surfaces") — do not repeat it here.

| Surface | Commands | Shelled by harness? |
|---------|----------|----------------------|
| **Shell CLI** | `graphify extract\|query\|path\|explain` | **YES** — the only shellable surface |
| **Agent slash-command** | `/graphify` | **NEVER** — this is an agent-run slash command, not a shell command |
| **MCP server** | `python -m graphify.serve` | **Deferred** — not a Slice A dependency; has a headless/cron auth caveat if adopted later |

## 3. Per-Phase Invocation Map

| Phase | Action | Mechanism |
|-------|--------|-----------|
| `sdd-explore` | Fresh → read `graph.json`/`GRAPH_REPORT.md`, shell `graphify query`/`graphify explain` for targeted questions. Absent → shell `graphify extract` (if binary present) then read. Stale → re-extract per §4, then read. | file read + shell CLI |
| `sdd-tasks` | Read Leiden community data from `graph.json`/`GRAPH_REPORT.md` **only** | **file read — never shell** |

`sdd-explore` is the **sole extraction site**. No other phase shells any
`graphify` command.

## 4. Staleness Algorithm

- Reference timestamp: `git show -s --format=%ct HEAD` (HEAD committer time).
- Subject timestamp: mtime of `<output_dir>/graph.json`.
- **Stale iff `mtime(graph.json) < HEAD_time`. Fresh iff `mtime(graph.json) ≥ HEAD_time`.**
- Fresh → reuse the existing graph; do NOT re-extract.
- Stale + binary present → `sdd-explore` automatically re-runs `graphify
  extract`, refreshes `graph-report.excerpt.md` (§ tracked excerpt), and emits
  exactly: `graph may be stale — re-extracting`.
- Stale + binary absent → failure mode (f) in §6.
- There is no config knob for staleness policy — this algorithm is unconditional.

## 5. `sdd-tasks` Read-Only Guarantee

`sdd-tasks` MUST NOT shell `graphify extract` or any other `graphify` command
— **even when the binary is present on PATH and `graph.json` is absent.**
`sdd-explore` is the sole extraction site in Slice A; `sdd-tasks` only ever
opens `graph.json`/`GRAPH_REPORT.md` for a read. This is structural, not a
policy choice re-derived per run: the Per-Phase Invocation Map in §3 already
marks the `sdd-tasks` row "file read — never shell," and this subsection
restates it so the constraint cannot be missed by a partial read of this
file. Missing or unreadable community data falls back to heuristic slice
boundaries (failure mode behavior per §6) — `sdd-tasks` never extracts to fix
that gap itself.

## 6. Advisory-Degradation Table (Absolute — R-07)

Every failure mode below emits exactly one single-line advisory note and lets
the consuming phase continue with baseline grep/read. **Never** return
`blocked`, **never** fail the phase, **never** halt the SDD flow.

| # | Failure mode | Advisory note | Fallback |
|---|---|---|---|
| a | `graphify` binary not on PATH | `graphify unavailable: binary not on PATH; proceeding with baseline grep/read` (if `auto_install: true`, install first per §7) | baseline grep/read |
| b | Python and uv and pipx all absent | `graphify unavailable: no Python/uv/pipx runtime; proceeding with baseline grep/read` | baseline grep/read |
| c | `graphify extract` exits non-zero | `graphify extract failed (exit N); proceeding with baseline grep/read` | prior graph if any, else baseline |
| d | `graph.json` absent and binary unavailable | `code graph unavailable and cannot extract; proceeding with baseline grep/read` | baseline grep/read |
| e | `graph.json` unparseable or schema-drifted | `code graph unreadable (parse/schema); proceeding with baseline grep/read` | baseline grep/read |
| f | graph stale and binary unavailable for re-extraction | `graph may be stale and graphify unavailable; using existing graph / baseline grep/read` | existing graph or baseline |
| g | `output_dir` unwritable | `cannot write to <output_dir>; skipping extraction, proceeding with baseline grep/read` | baseline grep/read |
| h | empty graph (0 nodes, 0 edges) | `code graph is empty; proceeding with baseline grep/read` | baseline grep/read |

## 7. `auto_install` Semantics

- `auto_install: false` (default) + binary missing → emit an advisory note
  naming the install commands (`uv tool install graphifyy` / `pipx install
  graphifyy`). Never install silently.
- `auto_install: true` + binary missing → run the install command **once**,
  then proceed with extraction. This is a one-time setup action, not a
  repeated install on every run.

## 8. `semantic: false` = Zero LLM Calls

When `semantic: false` (the default), the harness MUST NOT invoke any LLM API
via Graphify. Extraction uses pure local, deterministic tree-sitter AST
analysis only — no network calls, no model cost. The code graph remains fully
structurally queryable in this mode; `semantic: true` is an opt-in that adds
LLM-assisted edges but is not required for baseline usefulness.

## 9. Version-Mismatch Advisory

When the installed `graphify` binary's reported version does not match
`config.graphify.version` (default `v0.9.45`), emit a single advisory note
about the mismatch and continue without blocking. Never fail the phase over a
version drift alone.

## 10. Naming Discipline

Use **"code graph"** exclusively for Graphify's tree-sitter AST output
(`graph.json`, `graph.html`, `GRAPH_REPORT.md`). Never use "spec graph" or
"vault graph" for Graphify output — those terms belong exclusively to
`internal/mapgen`'s `openspec/map.md` output, which is an entirely separate
concept (the SDD change/capability vault, not a code-structure graph). Keeping
these vocabularies disjoint preserves the future Slice C bridge between the
two graphs.

## Artifact Layout

`graph.json`, `graph.html`, and `GRAPH_REPORT.md` live in `<output_dir>`
(default `.archon/graphify/`), already covered by `.gitignore` line 10
(`.archon/`) — no `.gitignore` edit is required. `sdd-explore` additionally
writes a tracked excerpt to
`openspec/changes/<change-name>/graph-report.excerpt.md` (at most 40 lines /
2 KB, whichever limit is hit first), refreshed on every (re-)extraction
including the automatic re-extraction described in §4.

## Rules

- NEVER reimplement AST extraction or Leiden community detection in Go code or
  skill prose — the only invocation mechanism is the `graphify` shell CLI.
- NEVER shell out to `/graphify` — that is an agent-run slash command, not a
  shell command (the documented Impeccable failure mode).
- NEVER depend on the MCP surface (`python -m graphify.serve`) — deferred,
  not a Slice A dependency.
- NEVER let `sdd-tasks` shell any `graphify` command, under any condition —
  see §5.
- NEVER fail a phase or return `blocked` because of Graphify — every failure
  mode in §6 degrades to an advisory note plus baseline grep/read.
- NEVER install Graphify silently — respect `auto_install` (§7).
- When `graphify.enabled: false`, this skill is not loaded and no phase
  changes behavior (§1).
