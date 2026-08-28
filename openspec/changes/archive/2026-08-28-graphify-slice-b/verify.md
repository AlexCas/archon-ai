# Verification Report

**Change**: graphify-slice-b
**Version**: Slice B delta to `[[graphify-integration]]` (R-19–R-22)
**Mode**: Standard (skill-prose-only change)

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 15 (A1–A8, B1–B2, C1, D1–D3, E1–E2) |
| Tasks complete | 15 |
| Tasks incomplete | 0 |

All 15 tasks in [tasks.md](tasks.md) are checked and each is backed by landed text
in the four modified files (see traceability below).

## Build & Tests Execution

**Build**: ➖ Not applicable — skill-prose-only change (no Go, no config knob, no gate).
The [design.md](design.md) Testing Strategy explicitly defines verification as
"inspecting the amended skill files against the Gherkin scenarios." There is no
executable test layer for Markdown skill prose; correctness is proven by tracing each
scenario's asserted anchor text to the literal text that landed. Confirmed via
`git diff` on the four files plus targeted `grep` of the preserved-exception lines.

**Tests**: ➖ No runtime test suite exists for this change type (by design).

**Coverage**: ➖ Not available (prose change).

## Spec Compliance Matrix

The `.feature` file contains **18** scenarios (not 22 — see Note 1 in Issues). Each
maps to a landed prose anchor rather than a runtime test.

| Requirement | Scenario | Evidence (landed text) | Result |
|-------------|----------|------------------------|--------|
| R-19 | copies graph.json → baseline before re-extracting | sdd-verify Step 8b: `copy <output_dir>/graph.json to <output_dir>/graph.baseline.json` … `Else shell graphify update` (copy precedes update) | ✅ COMPLIANT |
| R-19 | skips diff + advisory when baseline absent | Step 8b `if the source is absent, emit the §6 row-i advisory and skip`; graphify §6 row i | ✅ COMPLIANT |
| R-20 | emits advisory diff section on structural changes | report-format.md `### Code Graph Diff (advisory)` w/ 4 count rows; Step 8b `add the ### Code Graph Diff (advisory) section` | ✅ COMPLIANT |
| R-20 | "no structural changes" when empty | report-format.md `When all four counts are zero: No structural changes detected in the code graph.` | ✅ COMPLIANT |
| R-20 | up to 5 samples per category; node/edge format | report-format.md `Samples (up to 5 per non-empty category)`, `Added node — <node_id>`, `Removed edge — <source> →[<relation>]→ <target> (EXTRACTED)` | ✅ COMPLIANT |
| R-20 | advisory + skip when `graphify update` exits non-zero | Step 8b set-diff `on non-zero exit: emit §6 row-c note; skip diff section`; graphify §6 row c (`graphify update exits non-zero`) | ✅ COMPLIANT |
| R-20 | advisory + skip when graph.json unparseable | Step 8b `on parse/schema error of either: emit §6 row-j note; skip diff section`; graphify §6 row j | ✅ COMPLIANT |
| R-20 | re-snapshot + re-extract fresh on retry | sdd-verify Hard Rule `Re-snapshot and re-extract fresh on each re-apply retry` | ✅ COMPLIANT |
| R-20 | graphify.enabled false → no diff section | Hard Rule `When graphify.enabled is false, skip this check entirely`; report-format `Present only when graphify.enabled: true` | ✅ COMPLIANT |
| R-21 | judge cites EXTRACTED edge to enrich confirmed finding | harness-judge Hard Rule + Step 2 delegation, example `edge Y→[calls]→X removed (EXTRACTED, code graph)` | ✅ COMPLIANT |
| R-21 | INFERRED labeled `(INFERRED, semantic)` when semantic true | Step 2 `INFERRED edges may be cited only when semantic: true, labeled (INFERRED, semantic)` | ✅ COMPLIANT |
| R-21 | no new gate column from graph evidence | Step 4 note `Edge evidence never adds a column here — … judgment-day, mutation gate, playwright gate, impeccable gate`; Hard Rule `NEVER a new Step 4 result column` | ✅ COMPLIANT |
| R-21 | no re-apply loop from graph evidence alone | Hard Rule `NEVER a re-apply trigger on its own` | ✅ COMPLIANT |
| R-21 | never shells `graphify update` / extraction | Hard Rule `NEVER graphify update`; graphify §3 harness-judge row `query/explain only — never extract` | ✅ COMPLIANT |
| R-21 | graphify.enabled false → no citations | Hard Rule `When graphify.enabled: false, judges access no graph data and emit no edge citations` | ✅ COMPLIANT |
| R-22 | §3 has sdd-verify + harness-judge rows, "never extract" | graphify §3 both rows present; harness-judge row ends `(query/explain only — **never extract**)` | ✅ COMPLIANT |
| R-22 | §5 two-extraction-site invariant + sdd-tasks guarantee preserved | graphify §5 `sdd-explore and sdd-verify are the two extraction sites`; sdd-tasks read-only body intact; `harness-judge … MUST NOT shell any extraction command` | ✅ COMPLIANT |
| R-22 | §6 rows i and j, skip-diff fallback, no CRITICAL | graphify §6 rows i, j present with `skip diff section; verify continues`, advisory-note text, no CRITICAL severity | ✅ COMPLIANT |

**Compliance summary**: 18/18 scenarios compliant (content-traced).

## Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| R-19 Baseline snapshot | ✅ Implemented | Copy-before-update ordering correct in Step 8b; baseline lives in gitignored `.archon/graphify/`; absent-source degrades to §6 row i. |
| R-20 Advisory diff | ✅ Implemented | Set-diff over node `id` and `(source, target, relation)` tuples; `source_location`/`confidence` explicitly excluded from the key (`Edge identity key is (source, target, relation); confidence and source_location are not part of the key`). NOTE-only, never alters verdict — stated in Hard Rule, Step 8b, and report-format subtitle. |
| R-21 Edge evidence advisory | ✅ Implemented | Strictly advisory; read-only (`NEVER graphify update`); no new column / no Decision Gate / no re-apply trigger; inert when disabled. |
| R-22 SKILL.md amendments | ✅ Implemented | §1/§3/§5/§6 all landed; verb corrections applied where required; two exceptions correctly preserved (see below). |

### Verb-correction exceptions (task focus item 2) — confirmed real and justified

`grep -n "sole extraction site\|graphify extract" skills/graphify/SKILL.md` returns
exactly two remaining hits, both matching the E1-documented intentional preservations:

- **§2 Shell CLI surface table (L35)**: `` `graphify extract\|query\|path\|explain` `` —
  preserved. This documents Graphify's CLI surface listing and was explicitly scoped
  out by task E1. (Minor pre-existing staleness, deferred like E2 — not introduced by
  this slice; see Note 2.)
- **§5 prohibition (L65)**: `` `sdd-tasks` MUST NOT shell `graphify extract` or any
  other `graphify` command `` — preserved; the "or any other graphify command"
  clause makes the prohibition verb-agnostic, so the literal `extract` is harmless.

All corrective sites landed: §3 sdd-explore row (L43 `graphify update`), §4 Staleness
Algorithm (L57 `graphify update`), and §6 row c (L92 `graphify update` in both the
failure-mode and advisory-note cells).

### "Sole extraction site" replacement (task focus item 3) — confirmed

Replaced in **both** locations with a two-extraction-site statement:
- §3 (L48): `` `sdd-explore` and `sdd-verify` are the two extraction sites (see §5). No other phase shells any `graphify` extraction command. ``
- §5 (L68): the `> **Slice B amendment**:` blockquote naming both extraction sites and
  marking `sdd-tasks`/`harness-judge` read-only.

The sdd-tasks read-only guarantee text is intact in §5 (L74–80: "only ever opens
`graph.json`/`GRAPH_REPORT.md` for a read. This is structural, not a policy choice…").

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Skill-prose only, four files | ✅ Yes | `git diff --stat`: only the four design-named files changed (+87/-8). |
| Baseline retention via copy-aside | ✅ Yes | Step 8b `cp graph.json graph.baseline.json` before `graphify update`. |
| verify is 2nd extraction site | ✅ Yes | §3/§5 amended; sdd-verify Hard Rule + Step 8b shell `graphify update`. |
| Inline set-diff, no Go helper | ✅ Yes | Procedure embedded in Step 8b; no code files touched. |
| Judge edits scoped to harness-judge; judgment-day untouched | ✅ Yes | Only `skills/harness-judge/SKILL.md` changed; `judgment-day` not in diff. |
| §1 activation scope beyond R-22 letter (item 1) | ✅ Yes | L13 now `Load this skill from sdd-explore, sdd-tasks, sdd-verify, or harness-judge` — user-approved supporting edit. |
| E2 defer — sdd-tasks/SKILL.md untouched (item 6) | ✅ Yes | `git diff --name-only` shows no `sdd-tasks`; L172 cross-reference deliberately left, user-approved defer. |

## Issues Found

**CRITICAL**: None.

**WARNING**: None.

**SUGGESTION**:
- **Note 1 — scenario count**: The verify task brief referenced "22 scenarios", but
  the authoritative `graphify-integration.feature` contains **18** `Scenario:` blocks
  (2 for R-19, 7 for R-20, 6 for R-21, 3 for R-22). All 18 are compliant. The "22"
  appears to be a brief-level miscount / conflation with the label "R-22"; it does not
  affect the result. No missing scenario coverage was found.
- **Note 2 — deferred cleanup carry-over**: The §2 Shell CLI surface table still lists
  `graphify extract` as a verb (pre-existing, intentionally out of scope per E1), and
  `skills/sdd-tasks/SKILL.md:172` still reads "sole extraction site" (E2 deferred,
  user-approved 2026-08-28). Both are correctly deferred, not defects of this slice.
  Recommend folding them into the tracked follow-up cleanup change.

## Verdict

**PASS**
All 15 tasks complete and all 18 Gherkin scenarios (R-19–R-22) are satisfied by the
literal text that landed in the four files; the two verb-correction exceptions and the
E2 defer are real, documented, and user-approved. No CRITICAL or WARNING issues.

---

## Retry 1 Re-verification (2026-08-28)

Re-verify after re-apply retry 1/3, confirming the six fixes for judgment-day Round 1
feedback (C-1, S-1, S-2, S-3, S-4, S-5) landed correctly and introduced no
contradictions. Files touched this retry: `skills/sdd-verify/SKILL.md`,
`skills/harness-judge/SKILL.md`. (Diff stat now +98/-8 across the same four
design-named files — the delta over the +87/-8 pre-judge baseline is the six fixes.)

### Fix confirmation matrix

| Fix | Requirement | Landed evidence | Result |
|-----|-------------|-----------------|--------|
| **C-1** row-c/row-j mapping | R-20 / R-22 §6 | sdd-verify Step 8b: non-zero exit → `emit §6 row-c note` (L116); parse/schema error → `emit §6 row-j note` (L119). graphify §6 row c = "`graphify update` exits non-zero" (L92); row j = "Diff compute error (parse failure … or schema mismatch)" (L99). Hard Rules updated to "§6 rows c, i–j" (L57). Row-c is the exact semantic match for a non-zero `graphify update` exit; no residual row-j misreference. | ✅ FIXED |
| **S-1** fresh-per-retry in procedure | R-20 (step 7) | sdd-verify Step 8b operative body (L106–108): "On every re-apply retry (this run included), the baseline snapshot and the diff MUST be redone fresh — never reuse a snapshot or diff section carried over from a prior attempt." Now stated in the procedure, not only in Hard Rules. | ✅ FIXED |
| **S-2** EXTRACTED/INFERRED parity | R-21(c) | harness-judge Hard Rules (L31–41): EXTRACTED always permitted; INFERRED "ONLY when `graphify.semantic: true`" carrying the "`(INFERRED, semantic)`" label. Matches Step 2 (L102–106) verbatim in intent — contradiction removed. | ✅ FIXED |
| **S-3** R-21(d) sole-reason bar | R-21(d) | harness-judge Hard Rules (L39–41): "Edge evidence MUST NOT be the sole reason a finding is raised (R-21(d)) — it only enriches a finding the judge already confirmed independently." | ✅ FIXED |
| **S-4** both skipped when baseline absent | R-19 / R-20 | sdd-verify Hard Rules (L58–59): "When the baseline source is absent, BOTH re-extraction and diff rendering are skipped (R-19) — do not run `graphify update` first and decide afterward." | ✅ FIXED |
| **S-5** 5-samples cap in narrative | R-20 (step 5) | sdd-verify Step 8b (L128–129): "render diff section (report-format template; up to 5 samples per non-empty category — see references/report-format.md)". Cap stated inline, not only via the reference. | ✅ FIXED |

### No-regression checks

- **row-c/row-j did not break the parse/schema scenario**: feature scenario "graphify
  update exits non-zero" (L61–69, asserts "advisory note about the extraction failure")
  maps to row-c ("`graphify update` failed (exit N)…"); feature scenario "graph.json
  unparseable after update" (L72–79, asserts "advisory note about the parse failure")
  maps to row-j ("code graph diff failed (parse/schema)…"). Both mappings correct; no
  scenario regressed. 18/18 still compliant.
- **`skills/sdd-tasks/SKILL.md` untouched**: `git diff --name-only` shows no
  `sdd-tasks` entry — E2 remains deferred (user-approved).
- **No new contradiction introduced**: harness-judge Hard Rules (L31–41) and Step 2
  (L102–106) now agree on EXTRACTED/INFERRED handling; sdd-verify Hard Rules (L51–60)
  and Step 8b (L99–130) agree on the §6 rows c, i–j degradation and the baseline-absent
  behavior.

### Retry 1 Verdict

**PASS** — All six fixes landed correctly with no contradictions and no scenario
regression. `state.yaml` remains at `verify/completed`. No CRITICAL or WARNING issues.
Ready for re-judge (harness-judge Round 2).
