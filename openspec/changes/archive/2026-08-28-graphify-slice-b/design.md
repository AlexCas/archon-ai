# Design: Graphify Slice B — graph diff in verify + edge evidence in judge

<!-- [[graphify-integration]] · [proposal](proposal.md) · [spec](specs/graphify-integration/spec.md) -->

## Technical Approach

Skill-prose only (proposal Approach 1). No Go, no config knob, no gate. Three skill
files change: `skills/sdd-verify/SKILL.md` (advisory diff step + Hard Rule + report
section), `skills/harness-judge/SKILL.md` (edge-evidence Hard Rule + delegation
context, read-only), and `skills/graphify/SKILL.md` (R-22 §3/§5/§6 edits, plus two
supporting edits flagged below). All behavior stays inert under `graphify.enabled:
false` and degrades to one advisory line on every failure (Slice A R-07/R-08).

The set-diff has no native tool support (`graphify` has no `diff` verb) — it is a
harness-side set-difference over two `graph.json` snapshots, computed inline by the
verify executor. IDs are stable (proposal: byte-identical re-extracts), so a naive
set-diff is clean.

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Baseline retention | Copy explore-time `<output_dir>/graph.json` → `graph.baseline.json` before verify re-extracts | `graphify update` overwrites the single `graph.json`; a copy-aside is the only way to keep a "before". Both live in gitignored `.archon/graphify/` (line 10). No Go, trivial. |
| Who extracts "after" | verify shells `graphify update <path>` itself (2nd extraction site) | Diff needs a post-apply graph; verify runs after apply. Approved at the gate; relaxes §5. |
| Diff engine | Inline set-diff in the verify executor (no Go helper) | Mirrors Slice A (100% skill-driven runtime). Sets of node `id` and `(source,target,relation)` tuples. |
| Judge edits scoped to `harness-judge` only | Inject edge-evidence permission via Step 2 delegation context + a Hard Rule; leave `judgment-day` untouched | archon-judge invokes judgment-day internally, so criteria injected in harness-judge Step 2 reach the judges. Keeps `judgment-day` (an independently user-invocable skill) free of a graphify coupling. Keeps scope minimal. |
| §1 activation contract (supporting edit, beyond R-22 letter) | Amend "from `sdd-explore` or `sdd-tasks` only" to include `sdd-verify` and `harness-judge` | §1 currently forbids verify/judge from loading the skill; without this edit R-20/R-21 cannot load it. Necessary for the feature to function. |
| Stale `extract` verb (supporting edit) | Correct `graphify extract` → `graphify update` in the §3 explore row, §4, and §6 row c | Proposal empirically confirmed the verb is `update`. Leaving §3/§4 on `extract` while the new verify row says `update` ships a self-contradictory skill. Small, in-file, prevents drift. Flagged for gate sign-off (Open Questions). |

## Data Flow

    sdd-explore ──extract──▶ .archon/graphify/graph.json  (baseline source)
                                     │
    sdd-apply (± retries)            │ copy-aside
                                     ▼
    sdd-verify ── cp graph.json graph.baseline.json
              └─ graphify update <path> ──▶ graph.json (post-apply, overwritten)
              └─ set-diff(baseline, after) ──▶ "### Code Graph Diff (advisory)" NOTE
                                                  (never alters PASS/FAIL)
    harness-judge (read-only) ── read graph.json / graphify query|explain
                             └─ cite EXTRACTED edges in finding descriptions only

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `skills/sdd-verify/SKILL.md` | Modify | Hard Rule bullet (advisory diff), Execution Step 8b, report-section reference |
| `skills/sdd-verify/references/report-format.md` | Modify | Add `### Code Graph Diff (advisory)` to the report template |
| `skills/harness-judge/SKILL.md` | Modify | Edge-evidence Hard Rule; Step 2 delegation context; Step 4 "stays 4 columns" note |
| `skills/graphify/SKILL.md` | Modify | R-22 §3/§5/§6; supporting §1 + verb-correction edits |

---

## Exact edits

### `skills/graphify/SKILL.md`

**§3 — insert two rows after the `sdd-tasks` row (current L44).** New rows:

    | `sdd-verify` | Snapshot baseline (copy `graph.json` → `graph.baseline.json`); shell `graphify update <path>` for the post-apply graph; compute set-diff; emit advisory NOTE per R-20. Failure modes fall back per §6 rows i–j. | file copy + shell CLI |
    | `harness-judge` | Read `graph.json` or call `graphify query`/`graphify explain` to look up edge evidence for findings per R-21. | file read + shell CLI (query/explain only — **never extract**) |

**§3 — replace the sentence at current L46–47:**

- BEFORE: `` `sdd-explore` is the **sole extraction site**. No other phase shells any `graphify` command. ``
- AFTER: `` `sdd-explore` and `sdd-verify` are the two extraction sites (see §5). No other phase shells any `graphify` extraction command. ``

**§5 — replace the opening statement at current L63–65:**

- BEFORE: `` `sdd-explore` is the sole extraction site in Slice A; `sdd-tasks` only ever opens `graph.json`/`GRAPH_REPORT.md` for a read. ``
- AFTER: `` **Slice B amendment**: `sdd-explore` and `sdd-verify` are the two extraction sites. No other phase shells `graphify update` or any extraction command. `sdd-tasks` and `harness-judge` are read-only surfaces — file read or `graphify query`/`graphify explain` only; they MUST NOT shell any extraction command. `` Remainder of §5 (the sdd-tasks read-only guarantee + rationale) unchanged.

**§6 — append rows i and j after row h (current L89):**

    | i | `graph.baseline.json` absent at verify diff time (baseline copy not captured or source `graph.json` was absent) | `code graph baseline absent; skipping graph diff` | skip diff section; verify continues |
    | j | Diff compute error (parse failure of `graph.baseline.json` or post-apply `graph.json`, or schema mismatch) | `code graph diff failed (parse/schema); skipping graph diff` | skip diff section; verify continues |

**§1 — supporting edit (L13):** change `Load this skill from `sdd-explore` or `sdd-tasks` only when` → `Load this skill from `sdd-explore`, `sdd-tasks`, `sdd-verify`, or `harness-judge` only when`.

**Verb correction — supporting edit:** replace `graphify extract` → `graphify update` in the §3 `sdd-explore` row (L43), §4 (L55, L56), and §6 row c (L84). Add row-i/j alignment. (Flagged — Open Questions.)

### `skills/sdd-verify/SKILL.md`

**Hard Rules — new bullet after the Impeccable presence check (after current L50), modeled on it:**

    - **Code graph diff (conditional, advisory only)**: If `graphify.enabled` is true,
      after the compliance matrix and before the verdict, load `skills/graphify/SKILL.md`,
      capture the baseline (R-19), shell `graphify update <path>`, set-diff the two
      snapshots, and emit a `### Code Graph Diff (advisory)` section. This is a NOTE,
      never CRITICAL, and NEVER alters the PASS / PASS WITH WARNINGS / FAIL verdict.
      Re-snapshot and re-extract fresh on each re-apply retry. Every failure mode
      degrades to exactly one advisory line per `skills/graphify/SKILL.md` §6 rows i–j.
      When `graphify.enabled` is false, skip this check entirely.

**Execution Steps — new Step 8b between current Step 8 (matrix, L89) and Step 9 (persist, L90):**

    8b. If `graphify.enabled`, run the code graph diff (R-19/R-20): copy
        `<output_dir>/graph.json` to `<output_dir>/graph.baseline.json`
        (`output_dir` from `.archon/graphify.output_dir`, default `.archon/graphify`);
        if the source is absent, emit the §6 row-i advisory and skip. Else shell
        `graphify update <path>`, parse both snapshots, compute the set-diff, and add
        the `### Code Graph Diff (advisory)` section. Never changes the verdict.

**Set-diff procedure (for sdd-tasks to turn into checklist items):**

    output_dir = config.graphify.output_dir or ".archon/graphify"
    if not exists(output_dir/graph.json):      emit §6 row-i note; skip diff section
    cp output_dir/graph.json output_dir/graph.baseline.json   # R-19
    run: graphify update <project-root>        # overwrites output_dir/graph.json
      on non-zero exit:                        emit §6 row-j note; skip diff section
    baseline = parse(output_dir/graph.baseline.json)
    after    = parse(output_dir/graph.json)
      on parse/schema error of either:         emit §6 row-j note; skip diff section
    baseNodes  = { n.id for n in baseline.nodes }
    afterNodes = { n.id for n in after.nodes }
    baseEdges  = { (e.source, e.target, e.relation) for e in baseline.links }
    afterEdges = { (e.source, e.target, e.relation) for e in after.links }
    addedNodes   = afterNodes - baseNodes
    removedNodes = baseNodes  - afterNodes
    addedEdges   = afterEdges - baseEdges
    removedEdges = baseEdges  - afterEdges
    render diff section (below)

Fields per proposal's confirmed schema: nodes carry `id`; edges live in the `links`
array with `source`, `target`, `relation`, `confidence` (`EXTRACTED`/`INFERRED`).
The edge identity key is the `(source, target, relation)` tuple; `source_location`
is ignored so line drift never churns the diff.

### `skills/sdd-verify/references/report-format.md`

**Insert into the report template between `### Coherence (Design)` and `### Issues
Found` (current L56):**

    ### Code Graph Diff (advisory)
    _Present only when `graphify.enabled: true`; NOTE severity; never changes the verdict._

    | Category | Count |
    |----------|-------|
    | Added nodes | {n} |
    | Removed nodes | {n} |
    | Added edges | {n} |
    | Removed edges | {n} |

    Samples (up to 5 per non-empty category):
    - Added node — `<node_id>`
    - Removed edge — `<source> →[<relation>]→ <target> (EXTRACTED)`

    When all four counts are zero: `No structural changes detected in the code graph.`

Node sample format `` `<node_id>` ``; edge sample format `` <source> →[<relation>]→
<target> (EXTRACTED) `` (R-20 step 5).

### `skills/harness-judge/SKILL.md`

**Hard Rules — new bullet after the Security gate bullet (current L30):**

    - Graphify edge evidence is OPT-IN and ADVISORY. When `graphify.enabled: true`,
      judges MAY cite `EXTRACTED`-confidence edges from the post-apply `graph.json`
      (via file read or `graphify query`/`graphify explain` — NEVER `graphify update`)
      to enrich the description of findings they reached independently. Edge evidence
      is NEVER a new Step 4 result column, NEVER a Decision Gate condition, and NEVER a
      re-apply trigger on its own. When `graphify.enabled: false`, judges access no
      graph data and emit no edge citations.

**Step 2 — append a conditional context line to the archon-judge delegation:**

    - When `graphify.enabled: true`, tell the judges they MAY cite EXTRACTED edges
      from `<output_dir>/graph.json` (`(source, target, relation)`, confidence
      `EXTRACTED`) as supporting evidence for findings they independently confirm,
      e.g. `func X no longer called by Y — edge Y→[calls]→X removed (EXTRACTED, code
      graph)`. INFERRED edges may be cited only when `semantic: true`, labeled
      `(INFERRED, semantic)`. This is enrichment, never the sole basis for a finding.

**Step 4 — one-line note under the result table (current L179):** `` Edge evidence
never adds a column here — the table stays exactly: judgment-day, mutation gate,
playwright gate, impeccable gate. ``

## Interfaces / Contracts

No code interfaces. `graph.json` consumed schema: `nodes[].id`, `links[].{source,
target, relation, confidence}`. Advisory note is the only new output artifact,
threaded into the existing verify report template.

## Testing Strategy

Skill-prose change — no Go tests. Verification is by inspecting the amended skill
files against the 22 Gherkin scenarios (R-19–R-22).

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Content | §3/§5/§6 rows, verify Hard Rule + Step 8b, judge Hard Rule + Step 4 note, report template section | grep the anchor strings the scenarios assert (e.g. "two extraction sites", "never extract", "No structural changes detected", degradation rows i/j) |
| Inertness | `graphify.enabled: false` leaves verify/judge output byte-identical | confirm every new block is guarded by the enabled check |

## Migration / Rollout

No migration. Inert under `graphify.enabled: false`. Rollback = revert the four
file edits (proposal Rollback Plan).

## Estimated size

Skill-prose only, four files: roughly 60–90 changed lines total (two §3 rows, two §6
rows, §1/§5 sentence swaps, verb corrections; one verify bullet + one step; one
report section; one judge bullet + delegation line + table note). Well under the
800-line review budget — a single small PR. PR strategy: ask-always.

## Open Questions

- [ ] **Verb correction scope**: correct the stale `graphify extract` → `update` in
      §3/§4/§6-c now (recommended — avoids a self-contradictory skill), or defer to a
      cleanup change to keep this PR strictly to R-22's enumerated sections?
- [ ] **§1 activation edit**: confirmed necessary (verify/judge cannot load the skill
      otherwise) — flagged only because it is beyond R-22's three named sections.
