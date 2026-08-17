# Exploration: Graphify integration — Slice A (advisory graph gate)

> Session preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 lines · Playwright=disabled · Impeccable=disabled.

## Project Type

**Web testing**: `not-web` — this repo is the Archon Go CLI/TUI harness itself
(`cmd/archon`, `internal/tui` via Bubbletea, no browser surface). Playwright stays
disabled; no Impeccable recommendation. No project-type ambiguity to escalate.

## Scope Under Exploration

Slice A only: a `graphify.*` opt-in gate in `.archon/config.yaml`, a preflight
**group G** question, a new thin-orchestration skill `skills/graphify/SKILL.md`, and
**advisory, never-blocking** read-only consumption in exactly two phases —
`sdd-explore` (repo comprehension) and `sdd-tasks` (Leiden communities inform slice
boundaries). Slices B and C are out of scope (see Deferred / future slices).

## The Precedent — Impeccable is a near-isomorph (verified)

The Impeccable integration is the exact same shape and is the template. Verified
touch points:

- `internal/config/config.go` — `Impeccable` struct (lines 52-58); field on `Config`
  (line 90); default normalization in `Load()` (lines 118-120); value-copy in
  `Clone()` (line 147, comment "value copy — no maps/slices inside"). Simpler
  siblings `Playwright` (lines 31-35) and `Security` (lines 41-44) are pure
  `Enabled`-plus-scalar structs with no validation and no `Load()` defaulting — that
  is the pattern Graphify should follow, because Graphify has no blocking/severity
  concept.
- `cmd/archon/config.go` — set cases (lines 243-267) + get cases (lines 323-332), and
  BOTH long "supported keys" error strings (lines 297 and 348) that must be extended.
- `internal/tui/impeccable_tab.go` — whole-file tab pattern (state struct, focus
  count, `update`/`refocus`/`view`/`applyToConfig`/`setWidth`).
- `internal/tui/model.go` — `ImpeccableTab` const in the `Tab` iota (line 29), struct
  field (line 51), constructor (line 113), and SIX wiring sites: `WindowSizeMsg`
  (line 134), key routing (line 184), `agentInitDoneMsg` rebuild (lines 216, 224),
  `renderTabs` labels (line 286), `renderTabContent` (line 317), `saveConfig`
  (line 357). A new tab touches every one of these.
- `internal/status/display.go` — Impeccable status block (lines 53-64).
- `internal/initcmd/templates.go` — preflight group F text (lines 83-85) and the
  "group F maps to impeccable.enabled" mapping paragraph (lines 90-94), inside the
  raw-string `orchestratorSections` (`§` = backtick placeholder). Rules 6/7 live in
  `orchestratorRulesClaude`/`Opencode` (lines 185, 199).
- `internal/initcmd/init.go` — `Options.Impeccable` (line 31), `buildConfig` param +
  `Impeccable{Enabled: impeccable}` (lines 222, 250-252).
- `cmd/archon/main.go` — `impeccableFlag` var (line 84), `Impeccable:` into `Options`
  (line 172), `--impeccable` flag registration (line 205).
- `skills/impeccable/SKILL.md` — Activation Contract, "Two Invocation Surfaces" table,
  Per-Phase Invocation Map, and the load-me-only-when-`enabled` inertness rule.

## Current State (grounded)

- **Config is dynamic on skill count.** `skill_count` = `len(extracted)`
  (`internal/initcmd/init.go:236`, `buildConfig:236`; `update.go:135`). The rendered
  orchestrator files use `{{.SkillCount}}` (`templates.go:168`). `embeddedSkillCount()`
  (`cmd/archon/main.go:60-75`) counts dirs carrying `SKILL.md`. There is **no numeric
  Go constant to bump** — adding `skills/graphify/SKILL.md` auto-increments the count.
  `skills/embed.go` embeds `*/SKILL.md all:_shared`; `skills/embed_test.go`
  (`TestFS_ContainsSkills`) only asserts non-empty + presence of a named subset, so
  **adding graphify does not break it.**
- The tracked root `CLAUDE.md` / `AGENTS.md` hardcode the rendered `Skills: 25`
  (CLAUDE.md:170, AGENTS.md:151). These are generated dogfood files; regenerating
  after the new skill bumps them **25 → 26**. Also bumps `.archon/config.yaml`
  `skill_count` — but `.archon/` is gitignored (see below), so that copy is untracked.
- **`.gitignore` line 11: `.archon/`** — the entire `.archon/` tree is untracked.
  `openspec/specs/` and `openspec/changes/archive/` are intentionally tracked
  (audit trail). This is decisive for artifact placement (Q3).
- Two existing "graphs" already live in the repo and must NOT be conflated with
  Graphify's code graph: `internal/mapgen` produces `openspec/map.md` (the openspec
  spec/wikilink vault graph), and the router work (`archon route` / `sdd-router`)
  classifies messages. Neither is a code-AST graph (Q7).

## Answers to the Exploration Questions

### Q1 — Touch surface, line estimates, one-PR-vs-chain verdict

| # | File | New? | Est. changed lines | Purpose |
|---|------|------|--------------------|---------|
| 1 | `internal/config/config.go` | edit | ~20 | `Graphify` struct + `Config` field + `Clone()` copy (no `Load()` validation needed) |
| 2 | `internal/config/config_test.go` | edit | ~55 | Load fixture + `Clone` roundtrip (`TestConfig_CloneRoundtrip`) |
| 3 | `cmd/archon/config.go` | edit | ~50 | set/get cases for each field + extend both "supported keys" strings |
| 4 | `cmd/archon/config_test.go` | edit | ~90 | set/get + validation tests |
| 5 | `internal/initcmd/init.go` | edit | ~10 | `Options.Graphify`, `buildConfig` param + call site |
| 6 | `cmd/archon/main.go` | edit | ~5 | `--graphify` flag + pass into `Options` |
| 7 | `internal/initcmd/templates.go` | edit | ~18 | preflight group G + mapping paragraph |
| 8 | `internal/initcmd/templates_test.go` | edit | ~25 | assert group G rendered |
| 9 | `internal/tui/graphify_tab.go` | **new** | ~175 | tab state/update/view/applyToConfig (mirror impeccable_tab.go) |
| 10 | `internal/tui/model.go` | edit | ~28 | `GraphifyTab` const + field + 6 wiring sites + label |
| 11 | `internal/tui/graphify_tab_test.go` (+model_test) | **new/edit** | ~120 | tab behavior tests |
| 12 | `internal/status/display.go` | edit | ~12 | Graphify status block |
| 13 | `internal/status/display_test.go` | edit | ~18 | status assertion |
| 14 | `CLAUDE.md` (root, regenerated) | edit | ~15 | group G + skill count 25→26 |
| 15 | `AGENTS.md` (root, regenerated) | edit | ~15 | group G + skill count 25→26 |
| 16 | `skills/graphify/SKILL.md` | **new** | ~150 | thin orchestration skill |
| 17 | `skills/sdd-explore/SKILL.md` | edit | ~15 | advisory graph-comprehension step (gated) |
| 18 | `skills/sdd-tasks/SKILL.md` | edit | ~12 | Leiden-communities-inform-slices note (gated) |
| 19 | `skills/chained-pr/SKILL.md` | edit (optional) | ~6 | note community boundaries may inform work units |

**Total ≈ 840 changed lines** — marginally OVER the 800 budget. `skill-registry`
needs no edit (`.atl/skill-registry.md` is regenerated and gitignored).

**Verdict: over budget → chain into 2 PRs (Feature Branch Chain is mandatory here).**
Because the artifact store is `openspec`, archive-before-PR is in effect, so the
`sdd-tasks` Convergence Gate forbids Stacked-to-Main and forces **Feature Branch
Chain**. Natural cut (clean, independently reviewable):

- **PR 1 — the gate plumbing (Go + generated files), ~655 lines.** Rows 1-15. Lands
  `graphify.*` end-to-end (config struct, CLI set/get, init flag, TUI tab, status,
  preflight group G) with the flag **default false and fully inert** — harmless on
  its own, nothing consumes it yet.
- **PR 2 — the orchestration + consumers (pure Markdown), ~185 lines.** Rows 16-19.
  Lands `skills/graphify/SKILL.md` and the two phase-consumer edits that read the
  gate. Depends on PR 1.

**Swing lever for the human:** the TUI tab (rows 9+11 ≈ 295 lines) is the single
heaviest, least-essential piece — `archon config set graphify.*` already covers
configuration. Dropping/deferring the TUI tab brings Slice A to **~545 lines, which
fits comfortably in ONE PR under 800.** So the real decision is: *mirror Impeccable
fully (TUI tab included) → 2-PR FBC*, **or** *ship config-only parity now, defer the
TUI tab → single PR*. Recommendation: single PR without the TUI tab is the leaner,
lower-risk path and still delivers the full advisory capability; add the TUI tab as a
tiny follow-up if desired.

### Q2 — Config shape

Recommended MINIMAL set (`graphify:` block), each justified:

| Field | Type | Default | Keep? | Rationale |
|-------|------|---------|-------|-----------|
| `enabled` | bool | `false` | KEEP | The gate. Mirrors every sibling. |
| `auto_install` | bool | `false` | KEEP | Mirrors Impeccable. Whether the harness may run `uv tool install graphifyy` / `graphify install`. Default false contains the Python-runtime cost (Q6); never install silently. |
| `version` | string | `"8"` (pin) | KEEP | v8 is young + fast-moving → real API-instability risk. Pin the installed/expected version. |
| `output_dir` | string | `.archon/graphify` | KEEP | Where `graph.json`/`graph.html`/`GRAPH_REPORT.md` land. `.archon/` is gitignored, so this keeps the large `graph.json` untracked automatically (Q3). |
| `semantic` | bool | `false` | KEEP | ONE switch for all LLM-dependent features (Leiden community labels + doc/PDF semantic extraction). Default false = pure local deterministic AST graph, no LLM API calls → best privacy/cost/headless story. |

**Deliberately dropped / deferred (resist over-config):**

- **No `severity`** (unlike Impeccable). Slice A is advisory / never-blocking, so
  there is no block-vs-advisory verdict to configure. This is the key structural
  *difference* from Impeccable — and it means **no `Load()` validation and no
  `Validate*` helper** are needed; Graphify follows the simpler `Playwright`/`Security`
  struct shape, not the `Impeccable` severity shape.
- **Staleness policy → NOT a config field.** Handle it in skill prose: compare the
  graph file mtime against git `HEAD` and emit an advisory "graph may be stale" note.
  Adding a `max_age`/`staleness_days` field is over-configuration for an advisory
  feature. (Flagged as an open question in case a human wants an explicit knob.)
- No per-file paths beyond `output_dir`.

### Q3 — Artifact location and tracked/untracked split

- **`graph.json`, `graph.html`, `GRAPH_REPORT.md` → `.archon/graphify/` (untracked).**
  `.gitignore:11` already ignores `.archon/`, so the potentially large `graph.json`
  and the HTML viz never enter version control — no new `.gitignore` rule required.
- **Do NOT commit any graph output.** `openspec/` is tracked, so SDD artifacts must
  never embed `graph.json`. They MAY quote short, relevant excerpts from
  `GRAPH_REPORT.md` inline (e.g. a suggested-questions bullet or a community summary)
  as advisory context, referencing the file by path rather than copying it wholesale.
- **Open question surfaced:** whether to persist a tiny, human-readable excerpt of
  `GRAPH_REPORT.md` into the change folder for traceability. Default: no (keep it in
  the untracked cache; reference by path).

### Q4 — Invocation surface per phase; does the harness shell out?

Graphify has three surfaces (mirror Impeccable's "two surfaces" warning table):

| Surface | Commands | Shelled out by harness? |
|---------|----------|-------------------------|
| Shell CLI | `graphify extract`, `query`, `path`, `explain` | YES — real shell commands, safe to invoke |
| Agent slash-command | `/graphify .` | **NEVER** — agent-run only, exactly like `/impeccable <verb>` |
| MCP server | `python -m graphify.serve` | Optional; see caveat |

- **`sdd-explore`**: primary surface = **read the already-generated files**
  (`graph.json`, `GRAPH_REPORT.md`) directly, and/or shell `graphify query` /
  `graphify explain` (real CLI verbs) against the extracted graph, in place of blind
  grep/read. Extraction is a prerequisite: the harness MAY run `graphify extract`
  (real CLI) when the binary is present and allowed; it MUST NEVER shell out
  `/graphify` (that is the agent slash-command — the Impeccable failure mode to avoid).
- **`sdd-tasks`**: read-only — Leiden communities live inside `graph.json` /
  `GRAPH_REPORT.md`. Read the file; **no shell needed.** Feed community boundaries to
  `chained-pr` work-unit suggestions and the review-budget split decision.
- **MCP assessment:** `python -m graphify.serve` is a *reasonable* surface for
  interactive subagent querying, but it is **not** recommended as the Slice A default.
  Known caveat (must be recorded): interactively-authenticated MCP servers may be
  absent in headless/cron runs, so a subagent that depends on MCP would silently lose
  its data source there. File reads + CLI verbs have no such dependency. Reserve MCP
  as an opt-in enhancement, not the default path.

### Q5 — Graceful degradation (advisory = never fail)

Trigger conditions and the single rule: **the phase continues, emits a one-line note,
and falls back — it never fails and never returns `blocked`** (unlike the Impeccable
judge gate, which can block). Fallbacks:

- `graphify` binary missing / Python|uv absent / `version` mismatch → note
  `graphify unavailable: <reason>; proceeding with baseline grep/read` and continue.
- Extraction fails or times out → same note; use whatever prior `graph.json` exists,
  else baseline behavior.
- Graph stale (mtime older than `HEAD`) → advisory "graph may be stale" note; still
  usable.
- `sdd-explore` fallback = current grep/read comprehension. `sdd-tasks` fallback =
  current heuristic slice boundaries (`chained-pr` unchanged).

### Q6 — Second external runtime (Python) cost

Archon already shells out to Node (`npx impeccable`). Graphify adds **Python + uv/pipx**
as a second external runtime. Honest cost. Containment, all already in the design:
`graphify.enabled` defaults **false**; `auto_install` defaults **false**; **no install
at init time** (init only sets the flag, never runs `uv`/`graphify install`); and
advisory-only semantics mean absence of the runtime **never blocks** a phase. Net: zero
runtime cost for users who leave it off; opt-in and self-contained for those who enable it.

### Q7 — Interactions with existing gates / mapgen / router

- **`internal/mapgen` (`openspec/map.md`)**: a *different* graph — the openspec
  spec/wikilink vault, not code AST. No code coupling in Slice A. Bridging the two is
  explicitly **Slice C** (deferred). Note the naming overlap so reviewers don't
  conflate them.
- **`archon route` / `sdd-router`**: keyword phase classifier; no functional
  interaction. Only collision is the shared skill-count/templates surface (Q8).
- **Existing gates (Impeccable / Playwright / Security / Judge)**: independent
  booleans. Graphify is advisory and touches only explore + tasks, so it does **not**
  affect the judge gate or any verdict.

### Q8 — Known future merge conflict (record, don't design)

`internal/config/config.go` and `internal/initcmd/templates.go` (the `skill_count` /
preflight surface) are also touched by the **unmerged `local-model-router` slice B**
on `feat/local-model-router` (confirmed as a flagged collision in the repo
`SESSION_STATUS.md:43`). Both changes add a config gate + a new skill (bumping the
dynamic skill count) and edit the same preflight/template region. **Whichever merges
second resolves the conflict** — mechanical (adjacent struct fields + adjacent
preflight groups + a +1 count), not semantic. Note: a `git log master..feat/local-model-router`
on those two paths returned no committed delta from this branch's vantage, so the
lmr edits may still be uncommitted/in-flight on that branch; treat the collision as
expected regardless.

## Approaches

1. **Full Impeccable isomorph (config + CLI + TUI tab + status + preflight + skill +
   2 consumers).** Pro: maximum surface parity, TUI discoverability. Con: ~840 lines →
   over budget → forces a 2-PR Feature Branch Chain. Effort: Medium.
2. **Config-only parity now, defer the TUI tab.** Same as (1) minus
   `graphify_tab.go` + its tests. Pro: ~545 lines, fits ONE PR under 800; leaner, lower
   risk; full advisory capability still delivered (config via `archon config set` +
   preflight + skill). Con: no TUI toggle until a follow-up. Effort: Low-Medium.

## Recommendation

Take **Approach 2** unless the human specifically wants day-one TUI parity. It keeps
Slice A in a single reviewable PR, delivers the complete advisory capability, and
leaves the TUI tab as a trivial follow-up. If TUI parity is required now, fall back to
**Approach 1** as a 2-PR Feature Branch Chain (PR 1 = Go/gate plumbing incl. tab,
PR 2 = skill + consumers), because archive-before-PR (openspec store) forbids
Stacked-to-Main.

Structural commitments for downstream phases: no `severity`/validation (follow the
`Playwright`/`Security` struct shape); artifacts under gitignored `.archon/graphify/`;
advisory/never-blocking in explore + tasks only; harness may shell `graphify extract|
query|explain` but NEVER `/graphify` or depend on MCP by default.

## Risks

- **Budget:** ~840 lines with the TUI tab clears 800; the tab is the swing item. Under
  the user's known preference (split over raising the budget), plan on either dropping
  the tab (single PR) or a 2-PR FBC.
- **Young, fast-moving upstream (v8):** CLI/output-schema drift could break parsing.
  Mitigated by the `version` pin and by advisory-only degradation (parse failure → note,
  never fail).
- **Second runtime (Python/uv):** environment friction; contained by default-false +
  no-install-at-init + advisory fallback.
- **MCP in headless runs:** interactively-authenticated MCP servers may be absent in
  cron/headless; do not make Slice A depend on MCP.
- **Surface confusion:** shelling `/graphify` (agent slash-command) is the exact
  Impeccable failure mode — the skill MUST forbid it explicitly.
- **Merge collision** with `local-model-router` slice B on `config.go` +
  `templates.go`; second-merger resolves (mechanical).
- **Graph naming overlap** with `internal/mapgen`'s `map.md` — keep the code graph and
  the spec graph clearly distinct (Slice C bridges them later).

## Deferred / future slices (NOT in this change)

- **Slice B:** structural graph diff in `sdd-verify`; edge (`EXTRACTED`/`INFERRED`)
  evidence in `harness-judge`.
- **Slice C:** bridge `internal/mapgen` (openspec spec/wikilink graph) with Graphify's
  code graph in `sdd-archive`.

## Ready for Proposal

**Yes.** Scope is well-shaped and the precedent is mapped. The one decision the
orchestrator should put to the user before/at propose: **single PR without the TUI tab
(Approach 2, ~545 lines) vs. full Impeccable isomorph as a 2-PR Feature Branch Chain
(Approach 1, ~840 lines).**
