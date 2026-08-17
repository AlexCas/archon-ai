# Tasks: Graphify Integration — Slice A

<!-- Implements [[graphify-integration]] · [design](design.md) · [spec](specs/graphify-integration/spec.md) · [exploration](exploration.md) · [proposal](proposal.md) -->

> Preflight (cached): mode=interactive · store=openspec · PR=ask-always · budget=800 ·
> Playwright=off · Impeccable=off · Strategy=**single PR** (~430 lines, under 800).

## Review Workload Forecast

| Signal | Value |
|---|---|
| Files touched | 13 production + 5 test + 1 new skill + 3 skill edits |
| Estimated net additions | ~421 lines |
| Estimated deletions | ~10 lines (in-place renumbering, call-site updates) |
| Total changed lines | ~430 |
| Chained PRs recommended | **No** — comfortably under 800-line budget |
| Decision needed before apply | No |

`openspec/changes/graphify-integration/graph-report.excerpt.md` contributes **0 lines**
to this PR — the skill writes it at explore-time; no graph exists when apply runs.

## Coverage Declaration

**Unit-tested (Go tests emitted below):** R-01 · R-02 · R-03 · R-04 · R-05 · R-06

**No automated coverage (agent-executed skill behaviour; this repo has no Go exec
wrapper — `npx impeccable` is likewise only shelled from skill prose, never Go):**
R-07 (all 8 failure modes) · R-08 (inertness) · R-09 (surface separation) ·
R-10 (auto_install) · R-11 (semantic/no-LLM) · R-12 (staleness + auto-re-extract) ·
R-13 (sdd-explore consumption) · R-14 (sdd-tasks read-only) · R-15 (excerpt write) ·
R-16 (version advisory)

Verification approach for the above: SKILL.md content review against design §3 contract
checklist (T-19) + documented dry-run of the binary-missing path (T-20).

**R-18**: repo check only (`git status` confirms `.archon/` gitignored at line 10).

## Follow-up (Out of Scope)

Bringing `internal/initcmd/templates.go` fully current with the archive-before-PR prose
(CLAUDE.md rules 9-11, Session Status Archive paragraph, single-PR archive ordering) is a
**separate follow-up**, explicitly out of scope here.

---

## Tasks

### Phase 1 — Go Config Surface (R-01)

- [ ] **T-01** `internal/config/config.go`: add exported constants (`DefaultGraphifyVersion = "v0.9.45"`, `DefaultGraphifyOutputDir = ".archon/graphify"`) near line 61; add `Graphify` struct (5 fields: `enabled bool`, `auto_install bool`, `version string`, `output_dir string`, `semantic bool`) after the `Impeccable` struct (line 58), with doc comment "Advisory only — never blocks a phase, never returns a verdict, so there is no severity and no Load() validation"; add `Graphify Graphify \`yaml:"graphify"\`` field to `Config` after `Impeccable` (line 90); pre-seed `c.Graphify.Version = DefaultGraphifyVersion` and `c.Graphify.OutputDir = DefaultGraphifyOutputDir` in `Load()` immediately before `yaml.Unmarshal` (beside the Judge pre-seed at line 109 — this is defaulting, not validation; no `ValidateGraphify` call); add `Graphify: c.Graphify, // value copy — no maps/slices inside` to `Clone()` after line 147. Scalar shape only — mirror `Playwright`/`Security`, NOT `Impeccable`. **R-01**

- [ ] **T-02** `internal/config/config_test.go`: (a) extend `TestConfig_Load` with a fixture setting all five graphify fields to non-default values and assert they round-trip correctly; (b) extend `TestConfig_CloneRoundtrip` (line 210) to populate `Graphify` fields and assert the clone matches; (c) add `TestGraphify_DefaultsAbsentBlock` mirroring `TestSecurity_DefaultOff` (line 399) — absent `graphify` block must yield `Enabled:false`, `Version:"v0.9.45"`, `OutputDir:".archon/graphify"`. **R-01**

### Phase 2 — CLI Surface (R-02, R-03, R-04)

- [ ] **T-03** `cmd/archon/config.go` `setConfigValue`: add five `case` branches before `default` (line 269), mirroring the `impeccable.*` cases (lines 243-267): `graphify.enabled`/`auto_install`/`semantic` use `parseBool`; `graphify.version`/`output_dir` assign `value` directly (no path/format validation). Append `, graphify.enabled, graphify.auto_install, graphify.version, graphify.output_dir, graphify.semantic` to **both** "supported keys" error strings — `setConfigValue` line 297 and `getConfigValue` line 348. `getConfigValue`: add five parallel cases before `default` (line 333) — bools via `strconv.FormatBool`, strings returned verbatim. **R-02**

- [ ] **T-04** `cmd/archon/config_test.go`: add `TestConfigCmd_GraphifySetGet` (5-key round-trip mirroring `TestConfigCmd_ImpeccableSetGet` at line 329); add a test case asserting that `graphify.severity` produces an error whose message lists all five supported `graphify.*` keys (mirror line 298 pattern). **R-02**

- [ ] **T-05** `internal/status/display.go`: insert a Graphify block after the Impeccable block's closing `fmt.Fprintln(w)` (line 65), before the `Models` block (line 67). When disabled: print `"  Graphify (Code Graph)"`, underline `"  ---------------------"`, `"    Enabled:   false"`. When enabled: also print `Version`, `OutputDir`, `Semantic`. Mirror the label/underline/conditional idiom of the Impeccable block exactly. **R-03**

- [ ] **T-06** `internal/status/display_test.go`: add `TestDisplay_Graphify` mirroring `TestDisplay_Impeccable` (line 139) — two sub-tests: (a) disabled: output contains `"Graphify (Code Graph)"` and `"Enabled:   false"` but not the version/dir strings; (b) enabled with version `"v0.9.45"` and output_dir `".archon/graphify"`: output contains both strings. **R-03**

- [ ] **T-07** `cmd/archon/main.go`: add `graphifyFlag bool` to the var block near `impeccableFlag` (line 84 region); pass `Graphify: graphifyFlag` into the `initcmd.Options` literal (line 172 region); register `cmd.Flags().BoolVar(&graphifyFlag, "graphify", false, "Enable the Graphify advisory code-graph gate")` after the `--impeccable` registration (line 205 region). **R-04**

- [ ] **T-08** `internal/initcmd/init.go`: add `// Graphify enables the advisory code-graph gate.` + `Graphify bool` to `Options` after `Impeccable` (line 31 region); add trailing `graphify bool` parameter to `buildConfig` signature (line 222) — note this shifts positional args at **all existing call sites in `init_test.go`** (lines 610 and 633), which must be updated to pass the new trailing `false`; add `Graphify: config.Graphify{Enabled: graphify, Version: config.DefaultGraphifyVersion, OutputDir: config.DefaultGraphifyOutputDir}` to the returned `Config` struct; update the `buildConfig` call site (line 89) to pass `opts.Graphify`. **R-04**

- [ ] **T-09** `internal/initcmd/init_test.go`: add `TestBuildConfig_GraphifyFlag` mirroring `TestBuildConfig_ImpeccableFlag` (line 624) — two cases: `graphify:true` and `graphify:false`; assert `cfg.Graphify.Enabled` matches; assert `cfg.Graphify.Version == config.DefaultGraphifyVersion`; assert `cfg.Graphify.OutputDir == config.DefaultGraphifyOutputDir`. Also update the two existing `buildConfig` call sites in the file (lines 610, 633) to include a trailing `false` for the new `graphify` parameter. **R-04**

### Phase 3 — Harness Templates (R-05)

- [ ] **T-10** `internal/initcmd/templates.go` — `orchestratorSections`: (a) change `A–F` → `A–G` (line 58) and `six` → `seven` (line 61); (b) insert group G block after group F (after line 85): `- **G. Graphify (Grafo de código)** — "¿Activar Graphify para análisis de grafo de código?"` / `No (recomendado)` first / `Sí: extraer el grafo de código en sdd-explore y usar comunidades Leiden para sugerir límites de slices en sdd-tasks`; (c) insert group G mapping paragraph after the group-F mapping block (after line 94): `**Project type & code-graph gate (group G):**` / `Group G maps to §graphify.enabled§ in §.archon/config.yaml§. The §--graphify§ flag at init time sets the same value. When enabled, sdd-explore consults the Graphify code graph for repo comprehension and sdd-tasks reads Leiden communities to inform slice boundaries — advisory only, never blocking.` — In **both** `orchestratorRulesClaude` (line 178) and `orchestratorRulesOpencode` (line 192): insert `8. When graphify.enabled, sdd-explore consults the code graph and sdd-tasks reads Leiden communities to inform slice boundaries — advisory only, never blocking (no verdict)` after rule 7; renumber old 8 → 9 and old 9 → 10. **R-05**

- [ ] **T-11** `internal/initcmd/templates_test.go` — two tests require updates: (a) `TestTemplates_ContainSDDSessionPreflight` (line 116): add to the `required` slice `"G. Graphify (Grafo de código)"`, `"¿Activar Graphify"`, `"Group G maps to"`, `"graphify.enabled"`, `"--graphify"`, `"seven"`, `"A–G"`; (b) `TestTemplates_FiveRules` (line 188): add `"8. When graphify.enabled, sdd-explore consults the code graph"` to `sharedRules`; update existing entries from `"8. On judge fail..."` → `"9. On judge fail..."` and `"9. Commits carry..."` → `"10. Commits carry..."`; change the rule-boundary guard from `strings.Contains(content, "10. ")` → `strings.Contains(content, "11. ")`. **R-05**

### Phase 4 — Root Orchestrator Docs (R-05, R-06)

**Critical constraint: NEVER regenerate via `archon init --force`.** `internal/initcmd/templates.go` has drifted behind the root dogfood files — CLAUDE.md and AGENTS.md carry archive-before-PR prose from PRs #96-#99 (rules 9-10 and the Session Status Archive paragraph) that `templates.go` does not have. A full regeneration would clobber that merged content. Perform surgical hand-edits only.

- [ ] **T-12** `CLAUDE.md` (root, hand-edit): (a) `A–F` → `A–G` and `six` → `seven` in the preflight preamble; (b) insert group G question after group F (`- **G. Graphify...` with the same Spanish text and `No (recomendado)` first); (c) insert group G mapping paragraph after group F mapping; (d) insert `8. When graphify.enabled, sdd-explore consults the code graph and sdd-tasks reads Leiden communities to inform slice boundaries — advisory only, never blocking (no verdict)` after rule 7 — existing rules 8-10 become 9-11; (e) bump `Skills: 25` → `Skills: 26` in `## Configuration`. **R-05, R-06**

- [ ] **T-13** `AGENTS.md` (root, hand-edit — identical constraint applies): same five edits as T-12 — `A–F` → `A–G`, `six` → `seven`, group G question + mapping, graphify rule 8 inserted with renumbering of existing rules 8-10 → 9-11, skill count 25 → 26. **R-05, R-06**

### Phase 5 — Skill Files (R-06..R-17)

- [ ] **T-14** Create `skills/graphify/SKILL.md` (new file). Frontmatter: `name: graphify`, `description` covering triggers "graphify.enabled, code graph, Leiden communities", `metadata.scope: reference`, `version: "1.0"`. Ten sections: (1) Activation Contract — fully inert when `graphify.enabled: false`; no command shelled, `output_dir` not created; (2) Two Invocation Surfaces table — shell CLI (`graphify extract|query|path|explain`) shelled by harness; `/graphify` NEVER shelled (documented Impeccable failure mode); MCP (`python -m graphify.serve`) deferred; (3) Per-Phase Invocation Map — sdd-explore (fresh→read, absent→extract+read, stale→re-extract+read) + sdd-tasks (file-read only, never shell, even when binary present and graph.json absent); (4) Staleness algorithm — stale iff `mtime(graph.json) < git show -s --format=%ct HEAD`; fresh iff `≥`; stale+binary present→auto re-extract and emit `graph may be stale — re-extracting`; stale+binary absent→failure mode (f); (5) sdd-tasks read-only guarantee subsection (structural, not merely asserted); (6) Advisory-degradation table — all eight failure modes (a-h) with advisory note text and fallback; (7) auto_install semantics — false+missing→note naming `uv tool install graphifyy`/`pipx install graphifyy`, never install silently; true+missing→install once; (8) semantic:false = zero LLM calls — pure local deterministic AST; (9) version-mismatch advisory — one note, continue; (10) Naming discipline — "code graph" exclusively for Graphify output; "spec graph"/"vault graph" reserved for `internal/mapgen`. All references use `v0.9.45` / `.archon/graphify`. **R-06..R-17**

- [ ] **T-15** `skills/sdd-explore/SKILL.md`: add **Step 3d** after Step 3c (line 117, before Step 4). Gated on `graphify.enabled: true`: load `skills/graphify/SKILL.md`; run the fresh/absent/stale consumption per the Per-Phase Invocation Map; write `openspec/changes/<change-name>/graph-report.excerpt.md` (≤40 lines / ≤2 KB; refresh on every re-extraction). All failure modes fall back per R-07; MUST NOT shell `/graphify`; MUST NOT depend on MCP. ~15 lines. **R-13, R-15**

- [ ] **T-16** `skills/sdd-tasks/SKILL.md`: add a gated note in "Review Workload Forecast Rules" (line 146 region) — when `graphify.enabled`, read Leiden community data from `graph.json`/`GRAPH_REPORT.md` (read-only file access, never shell any graphify command); add one Rules bullet (line 305 idiom) — missing/unreadable community data → heuristic slice boundaries per R-07; never shell graphify commands. ~12 lines. **R-14**

- [ ] **T-17** `skills/chained-pr/SKILL.md`: add one advisory line in Execution Step 1 noting that Leiden community boundaries from `graph.json` MAY inform work-unit identification when `graphify.enabled` and data is present. ~6 lines. **R-14**

### Phase 6 — Optional Embed Test (R-06)

- [ ] **T-18** `skills/embed_test.go`: add `"graphify"` to the asserted subset in `TestFS_ContainsSkills`. ~3 lines. **R-06**

### Phase 7 — Verification Preparation (R-07..R-17, R-18)

- [ ] **T-19** After authoring T-14, review `skills/graphify/SKILL.md` against the design §3 contract checklist: Activation Contract present and correct; Two Invocation Surfaces table — CLI-only shellable surface; per-phase map accurate (sdd-explore shells, sdd-tasks file-reads only); staleness algorithm matches mtime-vs-HEAD-timestamp spec exactly; sdd-tasks read-only guarantee appears in both the map row and a dedicated subsection; all eight degradation modes (a-h) present with single-line advisory notes; auto_install semantics accurate; `semantic:false` = zero LLM calls; version-mismatch advisory present; "code graph" used exclusively for Graphify output (scan for "spec graph"/"vault graph" — must not appear in Graphify context). **R-07..R-17**

- [ ] **T-20** Document a step-by-step dry-run of the binary-missing path (failure modes a and b) in the verify artifact: (1) set `graphify.enabled: true`; (2) ensure no `graphify` binary on PATH; (3) `auto_install: false` — confirm skill emits advisory note naming `uv tool install graphifyy`/`pipx install graphifyy` and does NOT install; (4) confirm sdd-explore continues with baseline grep/read. This constitutes the manual verification record for R-07 a/b and R-10 (no-install path). **R-07, R-10**

---

## Proposed Commit Sequence

| # | Conventional subject | Files | ~Lines |
|---|---|---|---|
| C-1 | `feat(config): add Graphify struct, defaults, Load pre-seed, and Clone copy` | `internal/config/config.go`, `internal/config/config_test.go` | ~76 |
| C-2 | `feat(cli): wire graphify config into archon config get/set, status, and init flag` | `cmd/archon/config.go`, `cmd/archon/config_test.go`, `internal/status/display.go`, `internal/status/display_test.go`, `cmd/archon/main.go`, `internal/initcmd/init.go`, `internal/initcmd/init_test.go` | ~148 |
| C-3 | `feat(templates): add preflight group G and graphify rule to harness templates` | `internal/initcmd/templates.go`, `internal/initcmd/templates_test.go` | ~35 |
| C-4 | `feat(docs): hand-edit CLAUDE.md and AGENTS.md for group G, graphify rule, skill count` | `CLAUDE.md`, `AGENTS.md` | ~26 |
| C-5 | `feat(skills): add advisory code-graph skill and wire into sdd-explore, sdd-tasks, chained-pr` | `skills/graphify/SKILL.md` (new), `skills/sdd-explore/SKILL.md`, `skills/sdd-tasks/SKILL.md`, `skills/chained-pr/SKILL.md`, `skills/embed_test.go` | ~126 |

No `Co-Authored-By`, no "Generated with" lines, no tool attribution in any commit.

---

## Requirements with No Automated Coverage

The following requirements have no Go unit tests — agent-executed skill behaviour; this
repo has no Go exec wrapper (same situation as `npx impeccable`, which is only shelled
from skill prose in `harness-judge`):

**R-07** · **R-08** · **R-09** · **R-10** · **R-11** · **R-12** · **R-13** · **R-14** · **R-15** · **R-16**

Verify inherits T-19 (SKILL.md content review) and T-20 (documented dry-run) as the
coverage record for these requirements.

**R-18** is verified by repo check only (`git status` after apply; `.archon/` gitignored
at `.gitignore:10`).

## Open Questions

None requiring a human decision before apply. Both design open questions (hand-edit vs
regenerate; numbered graphify rule) were resolved in the design gate and encoded in the
tasks above.

## Top Risks

1. **templates.go drift / regeneration hazard** — a naive `archon init --force` during
   apply would silently clobber archive-before-PR content from PRs #96-#99. T-12/T-13
   make this constraint prominent; the "NEVER regenerate" constraint is repeated verbatim
   in Phase 4 above.
2. **`TestTemplates_FiveRules` rule-boundary check** — the guard that asserts no rule
   10 must be updated to assert no rule 11; missing this update while adding the new
   rule will cause the test to pass falsely.
3. **`buildConfig` signature change** — adding the trailing `graphify bool` param at
   T-08 shifts all existing positional callers in `init_test.go` (lines 610 and 633);
   both must be updated in T-09 or the package will not compile.
4. **Merge collision with `local-model-router` slice B** — both branches touch
   `internal/config/config.go` (adjacent struct fields) and
   `internal/initcmd/templates.go` (adjacent preflight group blocks). Mechanical but
   the second merger must resolve carefully, particularly that `templates.go` rule
   renumbering is consistent.
5. **Young upstream schema drift** — Graphify `0.9.x` may change `graph.json` schema;
   mitigated by version pin (`v0.9.45`) and parse-fail fallback (R-07 e, R-16).
