# Tasks: Graphify Slice B — graph diff in verify + edge evidence in judge

<!-- [[graphify-integration]] · [design](design.md) · [spec](specs/graphify-integration/spec.md) -->

**Scope**: Skill-prose only. Four files, roughly 60–90 changed lines total. No Go
code, no config knob, no gate. All behavior is inert under `graphify.enabled: false`.

---

## Group A — `skills/graphify/SKILL.md`

- [x] **A1 · §1 activation scope (L13)**: Replace the phrase
  `Load this skill from \`sdd-explore\` or \`sdd-tasks\` only when`
  with
  `Load this skill from \`sdd-explore\`, \`sdd-tasks\`, \`sdd-verify\`, or \`harness-judge\` only when`.
  (Supporting edit — necessary so verify and judge can legally load this skill per
  R-20/R-21; approved by user 2026-08-28.)

- [x] **A2 · §3 verb correction — sdd-explore row (L43)**: In the `sdd-explore` table
  row, in the Action column, replace
  `Absent → shell \`graphify extract\` (if binary present) then read.`
  with
  `Absent → shell \`graphify update\` (if binary present) then read.`
  (Supporting verb-correction edit; the stale verb in the explore row must match the
  new verify row that correctly uses `update`. Approved by user 2026-08-28.)

- [x] **A3 · §3 insert two rows after the sdd-tasks row (after L44)**: Append the
  following two rows to the Per-Phase Invocation Map table, directly after the
  `sdd-tasks` row:

  ```
  | `sdd-verify` | Snapshot baseline (copy `graph.json` → `graph.baseline.json`); shell `graphify update <path>` for the post-apply graph; compute set-diff; emit advisory NOTE per R-20. Failure modes fall back per §6 rows i–j. | file copy + shell CLI |
  | `harness-judge` | Read `graph.json` or call `graphify query`/`graphify explain` to look up edge evidence for findings per R-21. | file read + shell CLI (query/explain only — **never extract**) |
  ```

- [x] **A4 · §3 replace the "sole extraction site" sentence (L46–47)**: Replace
  `` `sdd-explore` is the **sole extraction site**. No other phase shells any `graphify` command. ``
  with
  `` `sdd-explore` and `sdd-verify` are the two extraction sites (see §5). No other phase shells any `graphify` extraction command. ``

- [x] **A5 · §4 verb correction (L55–56)**: In the Staleness Algorithm section, replace
  `` `sdd-explore` automatically re-runs `graphify extract`, ``
  with
  `` `sdd-explore` automatically re-runs `graphify update`, ``.
  (Only the verb inside the Staleness Algorithm `graphify` invocation; the surrounding
  sentence structure is unchanged.)

- [x] **A6 · §5 replace the opening extraction-site sentence**: Replace the sentence
  starting at the end of the `sdd-tasks` MUST NOT paragraph:
  `` `sdd-explore` is the sole extraction site in Slice A; `sdd-tasks` only ever opens `graph.json`/`GRAPH_REPORT.md` for a read. ``
  with the Slice B amendment block:
  > **Slice B amendment**: `sdd-explore` and `sdd-verify` are the two extraction sites. No other phase shells `graphify update` or any extraction command. `sdd-tasks` and `harness-judge` are read-only surfaces — file read or `graphify query`/`graphify explain` only; they MUST NOT shell any extraction command.

  The remainder of §5 (the sdd-tasks read-only guarantee rationale, "This is
  structural…" through the end of the section) is unchanged.

- [x] **A7 · §6 row c verb correction (L84)**: In the Advisory-Degradation Table, row c,
  replace both occurrences of `graphify extract`:
  - Failure mode cell: `` `graphify extract` exits non-zero `` → `` `graphify update` exits non-zero ``
  - Advisory note cell: `` `graphify extract failed (exit N); proceeding with baseline grep/read` `` → `` `graphify update failed (exit N); proceeding with baseline grep/read` ``

- [x] **A8 · §6 append rows i and j after row h (after L89)**: Add two new rows to the
  Advisory-Degradation Table immediately after row h:

  ```
  | i | `graph.baseline.json` absent at verify diff time (baseline copy not captured or source `graph.json` was absent) | `code graph baseline absent; skipping graph diff` | skip diff section; verify continues |
  | j | Diff compute error (parse failure of `graph.baseline.json` or post-apply `graph.json`, or schema mismatch) | `code graph diff failed (parse/schema); skipping graph diff` | skip diff section; verify continues |
  ```

---

## Group B — `skills/sdd-verify/SKILL.md`

- [x] **B1 · Hard Rules — new bullet after the Impeccable presence check (after L50)**:
  After the closing line of the "Impeccable presence check" bullet
  (`When \`impeccable.enabled\` is false, skip this check entirely.`),
  insert a new Hard Rule bullet:

  ```
  - **Code graph diff (conditional, advisory only)**: If `graphify.enabled` is true,
    after the compliance matrix and before the verdict, load `skills/graphify/SKILL.md`,
    capture the baseline (R-19), shell `graphify update <path>`, set-diff the two
    snapshots, and emit a `### Code Graph Diff (advisory)` section. This is a NOTE,
    never CRITICAL, and NEVER alters the PASS / PASS WITH WARNINGS / FAIL verdict.
    Re-snapshot and re-extract fresh on each re-apply retry. Every failure mode
    degrades to exactly one advisory line per `skills/graphify/SKILL.md` §6 rows i–j.
    When `graphify.enabled` is false, skip this check entirely.
  ```

- [x] **B2 · Execution Steps — new Step 8b between Step 8 and Step 9 (after L89)**:
  After Step 8 (`Build the behavioral compliance matrix from actual test results…`),
  insert a new numbered step before Step 9:

  ```
  8b. If `graphify.enabled`, run the code graph diff (R-19/R-20): copy
      `<output_dir>/graph.json` to `<output_dir>/graph.baseline.json`
      (`output_dir` from `.archon/graphify.output_dir`, default `.archon/graphify`);
      if the source is absent, emit the §6 row-i advisory and skip. Else shell
      `graphify update <path>`, parse both snapshots, compute the set-diff, and add
      the `### Code Graph Diff (advisory)` section. Never changes the verdict.
  ```

  The set-diff procedure (for reference during apply):
  ```
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
  render diff section (report-format template)
  ```

  Edge identity key is `(source, target, relation)`; `confidence` and
  `source_location` are not part of the key (line drift never churns the diff).

---

## Group C — `skills/sdd-verify/references/report-format.md`

- [x] **C1 · Insert Code Graph Diff section into the report template**: In the
  report template, between the `### Coherence (Design)` section (ending at L55)
  and the `### Issues Found` section (beginning at L57), insert:

  ````
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
  ````

  Node sample format: `` `<node_id>` ``.
  Edge sample format: `` <source> →[<relation>]→ <target> (EXTRACTED) ``.

---

## Group D — `skills/harness-judge/SKILL.md`

- [x] **D1 · Hard Rules — new bullet after the Security gate bullet (after L30)**:
  After the sentence ending `…count it as \`pass\`.`, insert:

  ```
  - Graphify edge evidence is OPT-IN and ADVISORY. When `graphify.enabled: true`,
    judges MAY cite `EXTRACTED`-confidence edges from the post-apply `graph.json`
    (via file read or `graphify query`/`graphify explain` — NEVER `graphify update`)
    to enrich the description of findings they reached independently. Edge evidence
    is NEVER a new Step 4 result column, NEVER a Decision Gate condition, and NEVER a
    re-apply trigger on its own. When `graphify.enabled: false`, judges access no
    graph data and emit no edge citations.
  ```

- [x] **D2 · Step 2 — append conditional context line to the archon-judge delegation**:
  After the line `- Criteria: spec compliance, design coherence, code quality`
  (inside the Step 2 delegation block), append:

  ```
  - When `graphify.enabled: true`, tell the judges they MAY cite EXTRACTED edges
    from `<output_dir>/graph.json` (`(source, target, relation)`, confidence
    `EXTRACTED`) as supporting evidence for findings they independently confirm,
    e.g. `func X no longer called by Y — edge Y→[calls]→X removed (EXTRACTED, code
    graph)`. INFERRED edges may be cited only when `semantic: true`, labeled
    `(INFERRED, semantic)`. This is enrichment, never the sole basis for a finding.
  ```

- [x] **D3 · Step 4 — add one-line note under the result table**: After the result
  table in Step 4 (after the final table row ending `→ enter re-apply loop`, around L178),
  insert:

  ```
  Edge evidence never adds a column here — the table stays exactly: judgment-day, mutation gate, playwright gate, impeccable gate.
  ```

---

## Group E — Post-Edit Verification (manual read-through)

- [x] **E1 · Grep for stale anchor strings**: After all edits are applied, confirm
  that the following patterns no longer appear in `skills/graphify/SKILL.md`:
  - `sole extraction site` — should be gone from both §3 and §5 of graphify/SKILL.md
  - bare `graphify extract` as an invocation verb in §3 sdd-explore row, §4, and §6 row c
    (the §2 Shell CLI surface table entry `` `graphify extract\|query\|path\|explain` ``
    and the §5 prohibition `` MUST NOT shell `graphify extract` `` are intentionally
    preserved and are NOT changed by this slice)

  Suggested check:
  ```
  grep -n "sole extraction site\|graphify extract" skills/graphify/SKILL.md
  ```

  Expected remaining hits: the §5 prohibition line (`MUST NOT shell \`graphify extract\``)
  and the §2 Shell CLI table entry — both are correct and unchanged.

- [x] **E2 · Cross-file staleness note**: `skills/sdd-tasks/SKILL.md` contains the
  line `` `sdd-explore` is the sole extraction site (see `skills/graphify/SKILL.md`). ``
  This cross-reference will become slightly stale after the §5 edit (graphify/SKILL.md
  will now say "two extraction sites"). This file is out of scope for Slice B per the
  design's file-change table. Confirm whether to address it now or defer:
  - **Defer** (recommended for this slice): the cross-reference correctly points readers
    to graphify/SKILL.md for the authoritative statement, which will be up to date.
  - **Fix now**: replace `is the sole extraction site` with `and \`sdd-verify\` are the
    two extraction sites` in `skills/sdd-tasks/SKILL.md` — flag as a minor in-scope
    addition and note it in the PR description.

  Document the decision before opening the PR.

  **Decision (2026-08-28, user-approved)**: Deferred. `skills/sdd-tasks/SKILL.md`
  is left unchanged in this slice; the cross-reference still correctly points to
  `skills/graphify/SKILL.md` for the authoritative (now up-to-date) statement.
  Follow-up tracked for a later change.
