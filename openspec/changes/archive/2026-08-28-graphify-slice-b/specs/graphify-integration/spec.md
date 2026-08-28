# Graphify Integration Specification — Slice B Delta

<!-- [[graphify-integration]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->
<!-- Slice A base spec: openspec/changes/archive/2026-08-17-graphify-integration/specs/graphify-integration/spec.md -->

> Session preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 · Playwright=off · Impeccable=off · Graphify=on.

## Purpose

Capability spec delta to `graphify-integration` (Slice B). Extends the archived Slice A
requirements (R-01–R-18) with two advisory behaviors:

1. **`sdd-verify` code graph diff** — after apply, snapshot the pre-apply baseline,
   re-extract the post-apply graph via `graphify update <path>`, set-diff the two
   snapshots, and surface the result as an advisory `### Code Graph Diff (advisory)`
   NOTE. Never alters the PASS / PASS WITH WARNINGS / FAIL verdict.

2. **`harness-judge` edge evidence** — when `judgment-day` raises findings, judges
   MAY reference `EXTRACTED`-confidence edges from the post-apply graph as supporting
   evidence in finding descriptions. Never a new gate, never a new column in the
   Step 4 result table, never a re-apply trigger.

Skill-only (no new Go code, no new `graphify.diff` config knob, no new blocking gate).
All new behavior is fully inert when `graphify.enabled: false`.

**Governing constraint**: R-07 (Advisory Absolute) and R-08 (Inertness When Disabled)
from Slice A bind every requirement in this delta. Any failure mode emits exactly one
advisory note and lets the phase continue — never `blocked`, never phase-fail, never
SDD-flow halt.

**CLI verb correction carried from proposal**: the extraction verb is `graphify update
<path>` (not `graphify extract` as assumed in the Slice A exploration). All references
in this delta use `update`.

All scenarios for each requirement are in `graphify-integration.feature` alongside
this file.

## Requirements

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

---

## Deferred (Out of Scope)

- New Go code, `graphify.diff` config knob, `semantic: true`/`INFERRED` edge
  extraction (beyond optional judge citation per R-21 when already configured).
- Any blocking behavior: no 4th verdict column, no re-apply trigger.
- Slice C (code graph ↔ spec graph bridge via `internal/mapgen`).
- MCP surface, TUI tab.

## Cross-References

- Slice A base spec:
  `openspec/changes/archive/2026-08-17-graphify-integration/specs/graphify-integration/spec.md`
- R-07 (Advisory Absolute) and R-08 (Inertness When Disabled) from Slice A govern
  every requirement in this delta without re-statement.
- R-14 and the §5 sdd-tasks guarantee are preserved unchanged; only the framing of
  the extraction-site count is updated.
