# Session Status

- **Session started**: 2026-08-17T16:53:00Z
- **Last updated**: 2026-08-17T18:02:00Z
- **Active change**: graphify-integration
- **Current phase**: apply (in_progress)
- **Next recommended**: complete apply (T-01..T-18), then verify

Tasks gate PASSED (user-approved 2026-08-17). Approved planned scope: ~430 changed lines,
the 5-commit sequence below, single PR. T-19/T-20 are handed to verify as documented
non-automatable checks.

## RULE RENUMBERING TRAP — orchestrator-verified, apply MUST NOT get this wrong
The two files have **different rule counts**, so the renumbering is NOT the same edit:
- `internal/initcmd/templates.go`: **9 rules** (rule 9 = commit attribution).
- root `CLAUDE.md`: **10 rules** (rule 9 = the archive-before-PR rule that templates.go
  lacks entirely; rule 10 = commit attribution). Root rule 8 also carries a Feature Branch
  Chain clause that templates.go's rule 8 does not.

Inserting graphify as rule 8 therefore shifts trailing rules 8→9, 9→10 in `templates.go`,
but 8→9, 9→10, 10→11 in root `CLAUDE.md`. **Never copy rule text between the two files.**

`internal/initcmd/templates_test.go:188` `TestTemplates_FiveRules` needs three edits:
its `sharedRules` entries `"8. On judge fail..."` → `"9. ..."` and
`"9. Commits carry ONLY..."` → `"10. ..."`, plus the trailing guard
`strings.Contains(content, "10. ")` → `"11. "` (currently asserting "exactly 9 rules").

## Post-Design Decisions (user-confirmed 2026-08-17, gate passed)
- **Apply uses a SURGICAL HAND-EDIT.** Never `archon init --force`, never a full template
  regeneration — it would clobber the merged archive-before-PR content. Group G, the skill
  count 25→26, and the graphify rule are hand-edited into root `CLAUDE.md`/`AGENTS.md`,
  and `templates.go` is edited separately. Bringing `templates.go` current is a SEPARATE
  follow-up, out of this change's scope.
- **Include a numbered graphify rule** in both `orchestratorRules` blocks, parallel to
  Impeccable's rule 7: when `graphify.enabled`, consult the code graph in explore and
  tasks. Accepted cost: the ripple into `TestTemplates_FiveRules`
  (`internal/initcmd/templates_test.go:188`) and rule renumbering.

Spec gate PASSED on revision pass 2 (user-approved 2026-08-17). The 421-line spec is the
binding contract for design.

## templates.go DRIFT — verified by the orchestrator, matters for apply
`internal/initcmd/templates.go` is **behind** the root dogfood files: it contains **zero**
occurrences of the archive-before-PR / Feature Branch Chain / Stacked-to-Main prose that
root `CLAUDE.md` carries in **six** places. A full template regeneration during apply
would therefore CLOBBER merged content (PRs #96–#99). Apply must use a **surgical
hand-edit** of the root files, not `archon init --force`.

Separate pre-existing debt, NOT this change's job: bring `templates.go` current with the
merged archive-before-PR content. Note that `local-model-router`'s W4 records drift in the
other direction on its own branch (root files lack the `archon route` rule), so the two
follow-ups interact.

## Post-Spec Gate Decisions (user-confirmed 2026-08-17)
- **Trim the spec to ~450 lines total.** Same 18 requirements, consolidated Gherkin:
  R-07's eight failure modes become a Scenario Outline; redundant scenarios merged.
- **Re-extraction: automatic when R-12 detects the graph is stale.** `sdd-explore`
  re-extracts on its own and refreshes the excerpt.
- **`sdd-tasks` stays strictly read-only.** If `graph.json` is absent it falls back to
  heuristics and never shells out — `sdd-explore` is the single extraction site.
- Orchestrator-decided (consistent with prior approvals, not re-asked): `output_dir`
  accepts any string with no path validation; `version` keeps the pinned `v0.9.45`
  default with R-16's advisory drift note rather than an empty default.

## Post-Propose Decisions (user-confirmed 2026-08-17, gate passed)
- **`archon status` parity: YES.** Keep the `internal/status/display.go` block (+test,
  ~30 lines). Estimate stays ~516 lines, single PR.
- **Excerpt convention confirmed as proposed**: `graph-report.excerpt.md`, ≤40 lines /
  ~2 KB, written by `sdd-explore`, refreshed on re-extraction.
- **`version`: pin an EXACT published tag**, not the `v8` major line. The spec phase must
  resolve the concrete tag from upstream releases and record it.
- `skill-registry` needs no source edit (orchestrator verified this directly against
  `skills/skill-registry/SKILL.md` — it auto-scans; output is gitignored).

## Post-Explore Decisions (user-confirmed 2026-08-17, gate passed)
- **Approach 2**: single PR, **TUI tab deferred** to a trivial follow-up. Target ~545
  changed lines, under the 800 budget. No Feature Branch Chain needed.
- **GRAPH_REPORT.md**: persist a short **tracked excerpt** in
  `openspec/changes/<name>/` for review traceability (user overrode the untracked
  default). The `graphify` skill must define excerpt location, max size, and
  regeneration trigger. Full `graph.json`/`graph.html` stay untracked in
  `.archon/graphify/`.
- **Staleness**: skill-side mtime-vs-HEAD advisory logic. No config knob.

## Preflight
- Execution mode: interactive
- Artifact store: openspec
- Chained PR strategy: ask-always
- Review budget: 800 lines
- Web project (Playwright): no
- Impeccable: no

## Scope Decision (user-confirmed 2026-08-17)
- **Slice A only**: `graphify.enabled` flag + preflight group G + `graphify`
  orchestration skill + read-only advisory use in `sdd-explore` and `sdd-tasks`.
- Explicitly OUT of scope for this change: verify structural diff, judge edge
  evidence (Slice B), mapgen↔Graphify bridge in archive (Slice C).
- Base branch: `feat/graphify-integration` off `master`.

## Phase History
- [x] explore — completed 2026-08-17T16:59:58Z
- [x] propose — completed 2026-08-17T17:08:21Z
- [x] spec — completed (revision pass 2) 2026-08-17T17:30:00Z
- [x] design — completed 2026-08-17T17:42:00Z
- [x] tasks — completed 2026-08-17T17:58:00Z
- [ ] apply — in_progress 2026-08-17T18:02:00Z
- [ ] verify
- [ ] judge
- [ ] archive

## Artifacts
- exploration: `openspec/changes/graphify-integration/exploration.md` (294 lines)
- proposal: `openspec/changes/graphify-integration/proposal.md` (116 lines)
- spec: `openspec/changes/graphify-integration/specs/graphify-integration/spec.md` (186 lines, 18 numbered requirements R-01..R-18)
- feature: `.../graphify-integration.feature` (235 lines, 26 scenario blocks incl. 3 Scenario Outlines)
- Total spec artifact: **421 lines**, down from 824, with all 18 requirements intact.
  Orchestrator verified: all of R-01..R-18 present, R-18 citation fixed to `.gitignore`
  line 10, R-14's no-shell prohibition asserted in both spec.md:145 and the .feature:194.
- design: `openspec/changes/graphify-integration/design.md` (340 lines)
- tasks: `openspec/changes/graphify-integration/tasks.md` (158 lines, T-01..T-20)

## Commit plan (single PR, ~430 changed lines)
- C-1 `feat(config)` — config.go + config_test.go (~76)
- C-2 `feat(cli)` — cmd/archon/config.go+test, status/display.go+test, main.go, init.go+test (~148)
- C-3 `feat(templates)` — templates.go + templates_test.go (~35)
- C-4 `feat(docs)` — root CLAUDE.md + AGENTS.md, hand-edit only (~26)
- C-5 `feat(skills)` — skills/graphify/SKILL.md (new) + sdd-explore + sdd-tasks + chained-pr + embed_test (~126)

No `Co-Authored-By`, no tool attribution, in any commit.

Orchestrator-verified against code: `buildConfig` currently takes 9 params
(`internal/initcmd/init.go:222`) with call sites at `init.go:89`, `init_test.go:610` and
`init_test.go:633` — T-08/T-09 cover the positional shift correctly.
- state: `openspec/changes/graphify-integration/state.yaml` — design/completed

## Design decisions of note
- Defaulting for `version`/`output_dir` uses exported constants
  (`DefaultGraphifyVersion`, `DefaultGraphifyOutputDir`) pre-seeded in `Load()` **before**
  `yaml.Unmarshal`, following the `c.Judge.Enabled = true` pattern at
  `internal/config/config.go:109`. This is defaulting, not validation — R-01's
  no-validation rule holds.
- Refreshed estimate: **~481–500 changed lines, single PR**. `graph-report.excerpt.md`
  contributes **0 lines** to this PR — the skill writes it at explore-time, and no graph
  exists yet when apply runs.
- Unit-testable in Go: R-01..R-06 only. R-07..R-16 are agent-executed skill behaviour and
  this repo has no exec wrapper to mock (`npx impeccable` is likewise only shelled from
  skill prose). Verify should accept: SKILL.md content review against the design §3
  checklist, plus a documented dry-run of the binary-missing note-and-continue path.
  R-18 is a repo check, not a unit test. "Python/uv absent" (R-07 b) is the hard one.

## Version Pin — RESOLVED
`v0.9.45`. The earlier `"v8"` was a **development branch name, not a release**. The PyPI
package is `graphifyy` and its release series is `0.9.x`. Orchestrator independently
confirmed `0.9.45` as latest via the PyPI JSON API on 2026-08-17.

Minor drift to fix in spec R-18: it cites `.gitignore` line 11 for `.archon/`; the actual
line is 10.

## Proposed `graphify.*` config surface (from propose)
`enabled` (false) · `auto_install` (false) · `version` ("v8") · `output_dir`
(`.archon/graphify`) · `semantic` (false, single switch for all LLM-dependent features).
Deliberately dropped: no `severity`, no `Load()` validation (advisory has no verdict),
no staleness knob.

Verified independently: `.archon/` is already gitignored and untracked, so `output_dir`
needs no `.gitignore` edit. `skills/skill-registry/SKILL.md` auto-scans and writes a
gitignored `.atl/skill-registry.md`, so Graphify is indexed by regeneration with no
source edit to that skill — propose's reading is correct.

Estimate: **~516 changed lines → single PR**, under the 800 budget (below the ~545 target
because dropping the TUI tab also drops its `model.go` wiring).

## Explore Findings (summary)
- Slice A is a near-isomorph of the merged Impeccable gate; every precedent touch point
  is line-referenced in exploration.md.
- Size estimate: **~840 changed lines with the TUI tab** (over the 800 budget),
  **~545 without it**. The TUI tab is the swing item.
- Correction to an earlier assumption: skill count is **dynamic** (`len(extracted)`,
  `internal/initcmd/init.go:98,236`, templated at `templates.go:168`) — there is no Go
  constant to bump and `embed_test.go` will not break.
- If chaining is needed, Feature Branch Chain is mandatory (not Stacked-to-Main) because
  the openspec store puts archive-before-PR in effect.

## Open Questions / Blockers
- All three explore-phase questions are resolved (see Post-Explore Decisions above).
- Follow-up parked (not this change): `internal/tui/graphify_tab.go` + `model.go` wiring
  + tab test (~320 lines).
- `local-model-router` is unmerged: 4 unpushed branches (tracker + 3 slices) plus a
  pending `sdd-archive` wave. Its `SESSION_STATUS.md` is committed on
  `feat/local-model-router` at e472e34 — do NOT look for it at the root of this branch.
- Known future merge conflict: this change edits `internal/config/config.go` and
  `internal/initcmd/templates.go` (skill_count), the same files as lmr slice B.
  Whichever merges second resolves.

## Resume Hint
Explore phase for the Graphify integration (Slice A) is running via the
`archon-explore` subagent. If interrupted, re-read this file, confirm the branch is
`feat/graphify-integration`, and re-run explore scoped to Slice A only — the
Impeccable gate (`skills/impeccable/SKILL.md`, `internal/config/config.go`,
preflight group F) is the structural precedent to mirror.
