# Design: Graphify Integration — Slice A (advisory code-graph gate)

<!-- Implements [[graphify-integration]] · [proposal](proposal.md) ·
     [spec](specs/graphify-integration/spec.md) · [exploration](exploration.md).
     Link rule: skills/_shared/spec-vault.md. -->

> Preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 · Playwright=off · Impeccable=off. Scope: **Slice A, Approach 2**
> (single PR, TUI tab deferred). Pin **`v0.9.45`** (PyPI `graphifyy==0.9.45`).

Every element below traces to a requirement ID (R-01..R-18). The coverage table
at the end is the map the tasks phase consumes.

## Technical Approach

Mirror the **Impeccable / Playwright** precedent exactly. Graphify is a scalar,
default-off config gate plus one thin orchestration skill and two gated,
read-only phase-consumer hooks. It differs from Impeccable in one structural way:
**advisory / never-blocking**, so there is **no `severity`, no `Validate*`
helper, and no `Load()` validation** — it follows the simpler `Playwright` /
`Security` struct shape (`config.go:31-44`), not the `Impeccable` severity shape
(`config.go:52-73`). All Graphify runtime behaviour lives in skill prose executed
by the agent at `sdd-explore` / `sdd-tasks` time; the Go code only carries config,
CLI, status, init-flag, and preflight text.

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Struct shape | Scalar (`Playwright`/`Security`), not `Impeccable` | Advisory gate has no verdict → no severity/validation (R-01) |
| Defaulting of `version`/`output_dir` | **Pre-seed in `Load()` before `yaml.Unmarshal`** + exported constants, reused by `buildConfig` | Absent block must yield `v0.9.45` / `.archon/graphify` (feature L17-21). This is *defaulting*, mirroring `c.Judge.Enabled=true` at `config.go:109` — NOT validation (R-01) |
| `output_dir` / `version` set | Assign any string, no path/format validation | Spec §"`output_dir` accepts any string" (R-01/R-02) |
| Root `CLAUDE.md`/`AGENTS.md` | **Hand-edit** group G + count + rule; do NOT full-regenerate | `templates.go` has drifted behind the root files (they carry archive-before-PR prose `templates.go` lacks). A full regen would clobber merged work — see Open Questions (R-05) |
| Extraction site | `sdd-explore` only; `sdd-tasks` strictly file-read | Single extraction site keeps `sdd-tasks` structurally shell-free (R-14) |
| Runtime cost | 2nd runtime (Python/uv) contained by default-off + no-install-at-init | R-08, R-10 |

## 1. Go Config Surface (R-01..R-04)

### `internal/config/config.go`
Add exported defaults (near `ValidImpeccableSeverities`, line 60-61 idiom):
```go
const (
	DefaultGraphifyVersion   = "v0.9.45"
	DefaultGraphifyOutputDir = ".archon/graphify"
)
```
Add the struct (after `Impeccable`, line 58; scalar shape, no severity/validate):
```go
// Graphify controls the opt-in, advisory code-graph gate backed by the external
// Python tool `graphify` (tree-sitter AST). Advisory only — never blocks a phase,
// never returns a verdict, so there is no severity and no Load() validation.
// Defaults to disabled (Enabled:false) when the block is absent.
type Graphify struct {
	Enabled     bool   `yaml:"enabled"`
	AutoInstall bool   `yaml:"auto_install"`
	Version     string `yaml:"version"`
	OutputDir   string `yaml:"output_dir"`
	Semantic    bool   `yaml:"semantic"`
}
```
`Config` field after `Impeccable` (line 90): `Graphify Graphify \`yaml:"graphify"\``.
**Defaulting** in `Load()` immediately before `yaml.Unmarshal` (beside the Judge
pre-seed at line 109), so an absent value is overridden only when set in YAML:
```go
c.Graphify.Version = DefaultGraphifyVersion
c.Graphify.OutputDir = DefaultGraphifyOutputDir
```
No Graphify block is added to the severity-normalize/validate region (lines
118-123). `Clone()` gains one line after line 147:
`Graphify: c.Graphify, // value copy — no maps/slices inside`.

### `internal/initcmd/init.go` (R-04)
`Options` gains (line 31 region): `// Graphify enables the advisory code-graph gate.` / `Graphify bool`.
`buildConfig` gains a trailing `graphify bool` param (line 222); the returned
`Config` gains a block using the shared constants:
```go
Graphify: config.Graphify{Enabled: graphify, Version: config.DefaultGraphifyVersion, OutputDir: config.DefaultGraphifyOutputDir},
```
Call site (line 89) passes `opts.Graphify`.

### `cmd/archon/main.go` (R-04)
Add `graphifyFlag bool` (var block line 82-84); pass `Graphify: graphifyFlag`
into `Options` (line 172); register (line 205 region):
`cmd.Flags().BoolVar(&graphifyFlag, "graphify", false, "Enable the Graphify advisory code-graph gate")`.

### `cmd/archon/config.go` (R-02)
`setConfigValue` — five cases before `default` (line 269), mirroring
`impeccable.*` (lines 243-267): `enabled`/`auto_install`/`semantic` via
`parseBool`; `version`/`output_dir` assign `value` directly (no validation).
`getConfigValue` — five parallel cases before `default` (line 333):
bools via `strconv.FormatBool`, strings returned verbatim.
**Both** "supported keys" error strings (lines **297** and **348**) append:
`, graphify.enabled, graphify.auto_install, graphify.version, graphify.output_dir, graphify.semantic`.

### `internal/status/display.go` (R-03)
Insert a block after the Impeccable block's closing `Fprintln` (line 65), before
`Models` (line 67), matching the label/underline idiom:
```go
fmt.Fprintln(w, "  Graphify (Code Graph)")
fmt.Fprintln(w, "  ---------------------")
fmt.Fprintf(w, "    Enabled:   %t\n", cfg.Graphify.Enabled)
if cfg.Graphify.Enabled {
	fmt.Fprintf(w, "    Version:    %s\n", cfg.Graphify.Version)
	fmt.Fprintf(w, "    Output Dir: %s\n", cfg.Graphify.OutputDir)
	fmt.Fprintf(w, "    Semantic:   %t\n", cfg.Graphify.Semantic)
}
fmt.Fprintln(w)
```
Disabled → only `Enabled: false`. Enabled → also version/output_dir/semantic.

## 2. Preflight Group G (R-05) — `internal/initcmd/templates.go`

All edits are inside `orchestratorSections` (shared; backtick = `§`). Verbatim
inserts:

**Preamble** — line 58 `A–F` → `A–G`; line 61 `Ask all\nsix every SDD session.`
→ `Ask all\nseven every SDD session.`

**Group G question** — insert after group F (after line 85):
```
- **G. Graphify (Grafo de código)** — "¿Activar Graphify para análisis de grafo de código?"
  - No (recomendado): no extraer ni consultar el grafo de código.
  - Sí: extraer el grafo de código en sdd-explore y usar comunidades Leiden para sugerir límites de slices en sdd-tasks.
```
Recommended option is **No**, listed first (R-05).

**Mapping paragraph** — insert after the group-F mapping block (after line 94):
```
**Project type & code-graph gate (group G):**
- Group G maps to §graphify.enabled§ in §.archon/config.yaml§. The §--graphify§
  flag at init time sets the same value. When enabled, sdd-explore consults the
  Graphify code graph for repo comprehension and sdd-tasks reads Leiden
  communities to inform slice boundaries — advisory only, never blocking.
```
(Contains the literal substrings `Group G maps to`, `graphify.enabled`,
`--graphify` the feature asserts at L75.)

**Both rules blocks** — `orchestratorRulesClaude` (line 178) and
`orchestratorRulesOpencode` (line 192): insert after rule 7, then renumber old
8→9, 9→10:
```
8. When graphify.enabled, sdd-explore consults the code graph and sdd-tasks reads Leiden communities to inform slice boundaries — advisory only, never blocking (no verdict)
```

## 3. `skills/graphify/SKILL.md` (R-06..R-17) — new, thin orchestration skill

Frontmatter mirrors `skills/impeccable/SKILL.md` (`metadata.scope: reference`,
`version: "1.0"`); `description` carries triggers "graphify.enabled, code graph,
Leiden communities". Adding the file auto-increments the dynamic skill count
(`embeddedSkillCount()`, `main.go:60-75`) — no Go constant (R-06). Section
outline:

1. **Activation Contract** — load only when `graphify.enabled: true`; when
   absent/false, no phase reads this file, no command shells, `output_dir` is not
   created, no phase output changes. Fully inert (R-08).
2. **Two Invocation Surfaces** table (R-09) — model on Impeccable's:

   | Surface | Commands | Shelled by harness? |
   |---|---|---|
   | Shell CLI | `graphify extract\|query\|path\|explain` | **YES** — the only shellable surface |
   | Agent slash-command | `/graphify` | **NEVER** — agent-run only (the Impeccable failure mode) |
   | MCP server | `python -m graphify.serve` | **Deferred** — not a Slice A dep; headless/cron auth caveat |

3. **Per-Phase Invocation Map** — only two rows:

   | Phase | Action | Mechanism |
   |---|---|---|
   | `sdd-explore` | Fresh→read `graph.json`/`GRAPH_REPORT.md`, `graphify query\|explain` for targeted Qs; absent→`graphify extract` then read; stale→re-extract (R-12) then read | file read + shell CLI |
   | `sdd-tasks` | Read Leiden communities from `graph.json`/`GRAPH_REPORT.md` only | **file read — never shell** |

4. **Staleness algorithm** (R-12) — reference = HEAD committer timestamp
   (`git show -s --format=%ct HEAD`); subject = mtime of `<output_dir>/graph.json`.
   **Stale iff `mtime(graph.json) < HEAD_time`; fresh iff `≥`.** Fresh → reuse, do
   NOT re-extract. Stale + binary present → `sdd-explore` **auto** re-runs
   `graphify extract`, refreshes the excerpt (§4), emits exactly
   `graph may be stale — re-extracting`. Stale + binary absent → failure mode (f).
   No config knob.
5. **`sdd-tasks` read-only guarantee** (R-14) — a dedicated subsection stating
   `sdd-tasks` MUST NOT shell `graphify extract` or any graphify command **even
   when the binary is present and `graph.json` is absent**; the map row and this
   subsection both encode it so it is structurally obvious, not merely asserted.
   Missing/unreadable community data → heuristic slice boundaries (R-07).
6. **Advisory-degradation table** (R-07) — eight modes, one note each, always
   fall back and continue; never `blocked`, never fail:

   | # | Failure mode | Advisory note (single line) | Fallback |
   |---|---|---|---|
   | a | binary not on PATH | `graphify unavailable: binary not on PATH; proceeding with baseline grep/read` (if `auto_install:true`, install first per R-10) | baseline grep/read |
   | b | Python & uv & pipx absent | `graphify unavailable: no Python/uv/pipx runtime; proceeding with baseline grep/read` | baseline |
   | c | `graphify extract` non-zero exit | `graphify extract failed (exit N); proceeding with baseline grep/read` | prior graph if any, else baseline |
   | d | `graph.json` absent & binary unavailable | `code graph unavailable and cannot extract; proceeding with baseline grep/read` | baseline |
   | e | `graph.json` unparseable / schema-drift | `code graph unreadable (parse/schema); proceeding with baseline grep/read` | baseline |
   | f | stale & binary unavailable | `graph may be stale and graphify unavailable; using existing graph / baseline grep/read` | existing graph or baseline |
   | g | `output_dir` unwritable | `cannot write to <output_dir>; skipping extraction, proceeding with baseline grep/read` | baseline |
   | h | empty graph (0 nodes/0 edges) | `code graph is empty; proceeding with baseline grep/read` | baseline |

7. **`auto_install` semantics** (R-10) — `false` (default) + missing binary → note
   naming `uv tool install graphifyy` / `pipx install graphifyy`; never install
   silently. `true` + missing binary → run install **once**, then proceed.
8. **`semantic: false` = zero LLM calls** (R-11) — off (default) → pure local
   deterministic AST; MUST NOT invoke any LLM API via Graphify; graph stays
   structurally queryable.
9. **Version-mismatch advisory** (R-16) — installed binary version ≠
   `config.graphify.version` → one advisory note, continue without blocking.
10. **Naming discipline** (R-17) — the file uses "code graph" exclusively for
    Graphify output; never "spec graph"/"vault graph" (those name
    `internal/mapgen`'s `openspec/map.md`).

## 4. Artifact Layout & Tracked Excerpt (R-15, R-18)

`graph.json` / `graph.html` / `GRAPH_REPORT.md` live in `.archon/graphify/`,
already ignored by `.gitignore` **line 10** (`.archon/`) — **no `.gitignore`
edit** (R-18). `sdd-explore` writes a tracked excerpt to
`openspec/changes/<change-name>/graph-report.excerpt.md` (≤40 lines / ≤2 KB),
refreshed on every (re-)extraction incl. auto re-extraction (R-15). Excerpt
format:
```
<!-- graphify code-graph excerpt · source: .archon/graphify/GRAPH_REPORT.md · graphify <version> · <UTC ts> -->
# Code Graph Report (excerpt)
<node/edge counts + EXTRACTED/INFERRED tallies pulled from GRAPH_REPORT.md head>
## Top Leiden communities
<first few community rows>
<!-- truncated at 40 lines / 2 KB — full report: .archon/graphify/GRAPH_REPORT.md -->
```
Truncation: hard cap at whichever of 40 lines / 2 KB is hit first; when content
is cut, the final retained line is the `truncated …` marker. **This excerpt is
produced at explore-time by the skill, not by this change's `apply` — it
contributes 0 lines to this PR** (there is no graph yet at apply).

## 5. Consumer-Phase Edits (R-13, R-14) — conditional-load references only

- `skills/sdd-explore/SKILL.md`: new **Step 3d** after Step 3c (line 117), before
  Step 4. Gated on `graphify.enabled: true`: load `skills/graphify/SKILL.md`,
  run the fresh/absent/stale consumption + excerpt write; all failure modes fall
  back per R-07; MUST NOT shell `/graphify` or depend on MCP (R-13). ~15 lines.
- `skills/sdd-tasks/SKILL.md`: a gated note in "Review Workload Forecast Rules"
  (line 146 region) + one Rules bullet (line 305 idiom) — when
  `graphify.enabled`, **read-only** Leiden communities from
  `graph.json`/`GRAPH_REPORT.md` inform Suggested Work Units / slice boundaries;
  **never shell any graphify command**; missing data → heuristic fallback (R-14).
  ~12 lines.
- `skills/chained-pr/SKILL.md` (optional): one line in Execution Step 1 noting
  community boundaries MAY inform work-unit identification when present. ~6 lines.

## Testing Strategy

| File | Tests (table-driven, existing idiom) | Requirement |
|---|---|---|
| `internal/config/config_test.go` | extend `TestConfig_Load` (absent block → `v0.9.45`/`.archon/graphify`; all-fields fixture) + `TestConfig_CloneRoundtrip`; add `TestGraphify_DefaultsAbsentBlock` mirroring `TestSecurity_DefaultOff` (line 399) | R-01 |
| `cmd/archon/config_test.go` | `TestConfigCmd_GraphifySetGet` (5-key round-trip, mirror `TestConfigCmd_ImpeccableSetGet` line 329) + assert unknown key `graphify.severity` error lists the five keys (mirror line 298) | R-02 |
| `internal/status/display_test.go` | `TestDisplay_Graphify` mirroring `TestDisplay_Impeccable` (line 139): disabled shows only `Enabled:`, enabled shows version/output_dir | R-03 |
| `internal/initcmd/init_test.go` (+ existing buildConfig coverage) | assert `--graphify`/`Options.Graphify` → `graphify.enabled: true` in built config | R-04 |
| `internal/initcmd/templates_test.go` | extend `TestTemplates_ContainSDDSessionPreflight` (line 116) for group-G question, mapping, `seven`, `A–G`, recommended `No`; **update `TestTemplates_FiveRules` (line 188)** for the new graphify rule + renumbering | R-05 |
| `skills/embed_test.go` (optional) | add `graphify` to the asserted subset in `TestFS_ContainsSkills` | R-06 |

**Not unit-testable without process mocking** (agent-executed skill behaviour;
this repo has **no Go exec wrapper to mock** — `npx impeccable` is likewise only
shelled from skill prose in `harness-judge`, never Go): R-07 all eight failure
modes, R-08 byte-identical inertness, R-09 surface separation, R-10 auto_install,
R-11 no-LLM, R-12 staleness/auto-re-extract, R-13/R-14 consumption, R-15 excerpt
write, R-16 version mismatch. The **"Python/uv/pipx absent"** mode (R-07 b) is the
known hard one. **Verify should accept instead:** (1) a SKILL.md content review
against the §3 contract checklist (Activation Contract, both tables, staleness
algo, degradation table present; "code graph" used exclusively per R-17), and
(2) a documented manual/dry-run of at least the binary-missing path showing
note-and-continue. R-18 is verified by a repo check (`.archon/` at `.gitignore:10`
+ `git status` clean), not a Go unit test.

## File Changes

| File | Action | Requirement |
|---|---|---|
| `internal/config/config.go` (+`config_test.go`) | Modify | R-01 |
| `cmd/archon/config.go` (+`config_test.go`) | Modify | R-02 |
| `internal/status/display.go` (+`display_test.go`) | Modify | R-03 |
| `cmd/archon/main.go`, `internal/initcmd/init.go` (+`init_test.go`) | Modify | R-04 |
| `internal/initcmd/templates.go` (+`templates_test.go`) | Modify | R-05 |
| `skills/graphify/SKILL.md` | **Create** | R-06..R-17 |
| `skills/sdd-explore/SKILL.md` | Modify | R-13, R-15 |
| `skills/sdd-tasks/SKILL.md` | Modify | R-14 |
| `skills/chained-pr/SKILL.md` | Modify (optional) | R-14 |
| `CLAUDE.md`, `AGENTS.md` (root) | Modify (hand-edit) | R-05, R-06 |
| `.gitignore` | **No edit** | R-18 |

## Migration / Rollout

No migration. Additive, default-off, single PR. Revert = revert the merge commit;
outputs live only in gitignored `.archon/graphify/`. Known mechanical merge
collision with `local-model-router` slice B on `config.go` + `templates.go`
(adjacent struct fields / adjacent preflight groups; skill count is dynamic —
no constant to bump); second merger resolves.

## Requirement → Design Coverage

| Req | Covered by |
|---|---|
| R-01 | §1 struct + defaults + `Load()` pre-seed + `Clone()` |
| R-02 | §1 `cmd/archon/config.go` get/set + both key strings |
| R-03 | §1 status block |
| R-04 | §1 `--graphify` → `Options` → `buildConfig` |
| R-05 | §2 group G + mapping + count + rule |
| R-06 | §3 SKILL.md (dynamic count) |
| R-07 | §3.6 degradation table |
| R-08 | §3.1 Activation Contract |
| R-09 | §3.2 Two Invocation Surfaces |
| R-10 | §3.7 auto_install |
| R-11 | §3.8 semantic |
| R-12 | §3.4 staleness algorithm |
| R-13 | §5 sdd-explore Step 3d |
| R-14 | §3.5 + §5 sdd-tasks read-only |
| R-15 | §4 tracked excerpt |
| R-16 | §3.9 version advisory |
| R-17 | §3.10 naming discipline |
| R-18 | §4 `.gitignore:10`, no edit |

## Open Questions

- [ ] **Root `CLAUDE.md`/`AGENTS.md`: hand-edit vs regenerate.** `templates.go`
  has drifted behind the root dogfood files (they carry archive-before-PR prose
  `templates.go` lacks). Design assumes a surgical hand-edit (group G + count
  25→26 + graphify rule) to avoid clobbering merged work. Confirm, or accept that
  a full regen would need `templates.go` brought current first.
- [ ] **Graphify numbered rule in both rules blocks.** R-05's feature scenario
  requires only the group-G question/mapping/count. The rule (parallel to
  impeccable's rule 7) is included per the task's stated edit sites and ripples
  into `TestTemplates_FiveRules`. Drop it if a leaner surface is preferred.

## Risks

- **Young upstream (`0.9.x`) schema drift** → parse failure. Mitigated by
  `version` pin + advisory parse-fail fallback (R-07 e, R-16).
- **Second runtime (Python/uv)** friction. Contained by default-off +
  no-install-at-init + advisory fallback (R-08, R-10).
- **Surface confusion** — shelling `/graphify` is the Impeccable failure mode;
  the skill forbids it explicitly (R-09).
- **`templates.go` drift** (above) — a naïve regen reverts archive-before-PR
  content in the root files.
- **Merge collision** with `local-model-router` slice B; mechanical.
- **Code-graph vs spec-graph** naming conflation; kept disjoint to preserve
  Slice C (R-17).
