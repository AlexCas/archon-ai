# Design: Streamline the Harness

<!-- Link convention: [[capability]] for capability identity; relative links for
     intra-change navigation. Full rule: skills/_shared/spec-vault.md. -->

## Technical Approach

Six independent simplifications, executed as removals-before-behavior. Mechanical Go
deletions (Impeccable, Graphify) land first and unlock the skill cleanup; then the
skill layer loses its engram/none/Impeccable/Graphify branches; then the preflight
collapses to prose; then judge reduces to a single review and the bugfix track is
added. `internal/initcmd/templates.go` remains the single source of truth for
CLAUDE.md/AGENTS.md — editing it regenerates both via `RenderClaudeMD`/`RenderAgentsMD`
(exercised on `archon init`, `archon update`, and TUI save at `model.go:410`).

No new Go struct is introduced (preflight is prose-only, per the fixed decision). The
YAML loader stays lenient (`yaml.Unmarshal` without `KnownFields(true)` at
`config.go:139`), so leftover `impeccable:`/`graphify:` blocks in existing configs are
silently ignored — the removal is safe by construction, and the CLI `set/get` paths
become the only surface that rejects those keys.

## Architecture Decisions

| Decision | Choice | Rationale / alternative rejected |
|---|---|---|
| Config removal safety | Rely on lenient `yaml.Unmarshal` (no `KnownFields`) | Existing configs with leftover blocks load clean with zero migration tooling. `KnownFields(true)` would break every pre-existing config — rejected. |
| Impeccable severity default+validate | Delete the pre-seed at `config.go:146-148` AND the `ValidateImpeccableSeverity` call at `:149-151` | These lines reference removed symbols; leaving either breaks the build. |
| Graphify version/output_dir pre-seed | Delete the two `DefaultGraphify*` consts + the two pre-seed lines at `config.go:136-137` | Same — dangling symbol references. |
| Preflight defaults storage | Prose in `orchestratorSections` only; NO `Preflight` Go struct | Fixed decision. Keeps this change struct-free; no `Clone()`/`CloneRoundtrip`/TUI churn. |
| Tab-iota terminal tab | `SecurityTab` becomes last; drop `ImpeccableTab`+`GraphifyTab` from the iota block | Wrap arithmetic uses `tabCount`; keeping the two constants would leave phantom tabs in the modulo range. |
| Single judge site | `archon-judge` subagent stays the model-gated runner; its body swaps `judgment-day` for a direct focused review | Preserves the frontmatter `model:` hard gate (Opus). Collapsing judge into the orchestrator would lose the model gate — rejected. |
| `judgment-day` | Byte-for-byte unchanged | Fixed decision; only `harness-judge`/`archon-judge` stop calling it. |
| Bugfix track selection | `track` field on `state.yaml`; default-absent = `full`; unknown = error | Back-compatible additive field; no schema migration for in-flight changes. |

## Concern 1 — Preflight → Prose (CLI + init)

### templates.go rewrite (`orchestratorSections`)

The 7-group arrow-key ceremony (`templates.go:46-109`) is replaced by a **standard
default expressed as prose**. Concretely:

- **Lines 56-88** (the A–G arrow-key question list): delete. Replace with a short
  "SDD Session Preflight" prose block stating the fixed default:
  *Ritmo=interactivo · Artefactos=OpenSpec · PRs=ask-but-default-chained · Revisión=800*,
  applied SILENTLY, with the deviation rule below.
- **Lines 90-91** (group E mapping paragraph, `playwright.enabled`): KEEP — Playwright
  survives. Reword to drop the "group E" framing.
- **Lines 93-97** (group F mapping, `impeccable.enabled`): delete.
- **Lines 99-103** (group G mapping, `graphify.enabled`): delete.
- **Lines 105-109** (hard-gate rules referencing "seven per-group questions" and
  "§openspec§, §engram§, or §both§"): replace with the prose deviation contract.

**Deviation rule (prose, verbatim intent):** the default is applied without asking;
the orchestrator surfaces a single scoped one-line question ONLY when (a) the estimate
exceeds the 800-line budget (suggest slicing / budget bump), (b) the change is
trivially small (suggest a single PR), or (c) the user explicitly asks to adjust. This
satisfies `[[harness-workflow]]` "Preflight is a Standard Default Expressed as Prose".

Because Artefactos is fixed to OpenSpec, group B is gone entirely (no engram/both/none
choice) — this simultaneously discharges `[[harness-init]]` "Artifact store is always
OpenSpec".

### Rules-block edits

`orchestratorRulesClaude` (`templates.go:187-199`) and `orchestratorRulesOpencode`
(`:203-215`):
- Rule 8 (`impeccable.enabled`) — delete.
- Rule 9 (`graphify.enabled`) — delete.
- Rule 7 (`playwright.enabled`) — keep.
- Rule 6 ("After verify, invoke harness-judge") — keep (harness-judge stays the SDD
  gate; its internals change, not its invocation). Renumber the trailing rules after
  the two deletions (10→8, 11→9).

### Rendered orchestrator files + README + tests

- `CLAUDE.md` / `AGENTS.md` (repo root): regenerated from the template — but this
  change also edits them directly so the committed copies match (the repo's own
  CLAUDE.md carries the 7-group block today). Mirror the template edits: prose
  preflight, drop groups B-engram/F/G, drop rules 8–9.
- `internal/initcmd/templates_test.go`: remove Group F/G assertions (~:149-151), the
  Group B engram/"Ambos" assertions, and rule 8/9 assertions (~:205); add an assertion
  that the rendered output contains the prose default and does NOT contain "Impeccable",
  "Graphify", "Engram", or "Ambos".
- `README.md`: strip `--impeccable`, the Impeccable/Graphify TUI-tab descriptions, and
  the config snippets (lines ~103, 162, 190-192, 260-262).

## Concern 2 + 3 — Remove Impeccable & Graphify (Config, CLI, init, TUI)

### `internal/config/config.go`

| Anchor | Action |
|---|---|
| `:46-58` `Impeccable` struct | Delete |
| `:60-70` `Graphify` struct | Delete |
| `:72-79` `DefaultGraphifyVersion`/`DefaultGraphifyOutputDir` consts | Delete |
| `:81-94` `ValidImpeccableSeverities` var + `ValidateImpeccableSeverity` func | Delete |
| `:111` `Impeccable Impeccable` field, `:112` `Graphify Graphify` field | Delete both |
| `:136-137` Graphify pre-seed in `Load()` | Delete |
| `:143-151` Impeccable severity normalize + validate in `Load()` | Delete (9 lines) |
| `:175` `Impeccable: c.Impeccable`, `:176` `Graphify: c.Graphify` in `Clone()` | Delete both |

After these deletions `Load()` has no Impeccable/Graphify defaulting and no validation
call; the only remaining pre-seed is `c.Judge.Enabled = true` (`:131`).

### `internal/config/config_test.go`

- Remove `Impeccable`/`Graphify` fields from the `TestConfig_CloneRoundtrip` fixture
  (~:265-269) so the roundtrip compiles.
- Delete `TestImpeccable_DefaultsAndValidation` (~:458-503).
- Delete any Graphify default/validation test analogue.

### `cmd/archon/config.go`

- `setConfigValue`: delete the six `impeccable.*` cases (`:243-267`) and the five
  `graphify.*` cases (`:269-294`).
- `getConfigValue`: delete the five `impeccable.*` cases (`:350-359`) and the five
  `graphify.*` cases (`:360-369`).
- Both unknown-key error strings (`:324`, `:385`): remove every `impeccable.*` and
  `graphify.*` key from the `(supported: …)` list. The trimmed supported list keeps
  `models.*`, `playwright.*`, `mutation_testing.enabled`, `security.enabled`,
  `security.profile`. After removal, `archon config set impeccable.enabled true` falls
  through to the default branch → unknown-key error, satisfying
  `[[harness-init]]`/`[[graphify-integration]]` "unknown-key error" scenarios.

### `cmd/archon/main.go`

- Delete `impeccableFlag` and `graphifyFlag` var decls (`:87-88`).
- Delete `Impeccable: impeccableFlag` / `Graphify: graphifyFlag` from the `Options{}`
  literal (`:176-177`).
- Delete the two `cmd.Flags().BoolVar(...)` registrations (`:210-211`). After removal,
  `archon init --impeccable` / `--graphify` is an unknown flag (Cobra rejects it),
  satisfying the "no init flag" scenarios.

### `internal/initcmd/init.go`

- Delete `Impeccable bool` / `Graphify bool` from `Options` (`:30-33`).
- Delete `opts.Impeccable, opts.Graphify` args from the `buildConfig(...)` call (`:91`).
- Delete the `impeccable bool, graphify bool` params from `buildConfig` (`:224`) and
  the `Impeccable: config.Impeccable{...}` (`:252-254`) + `Graphify: config.Graphify{...}`
  (`:255-258`) fields it sets. Default-config generation no longer emits either block.
- `internal/initcmd/init_test.go`: delete `TestBuildConfig_ImpeccableFlag` (~:621-639)
  and any Graphify analogue.

### TUI — tab-iota model (the load-bearing detail)

**Current iota** (`model.go:22-32`), 8 tabs + sentinel:

```
AgentTab=0, ModelsTab=1, JudgeTab=2, MutationTab=3, PlaywrightTab=4,
SecurityTab=5, ImpeccableTab=6, GraphifyTab=7, tabCount=8
```

**After removal** — delete the `ImpeccableTab` and `GraphifyTab` lines from the const
block (`:29-30`). `SecurityTab` stays at index 5 and becomes the last real tab;
`tabCount` collapses to 6:

```
AgentTab=0, ModelsTab=1, JudgeTab=2, MutationTab=3, PlaywrightTab=4,
SecurityTab=5, tabCount=6
```

Because the iota is contiguous and the two removed tabs were the last two before the
sentinel, **no surviving tab's index changes** — only `tabCount` drops from 8 to 6.
The wrap arithmetic is index-agnostic:
- next-tab `(activeTab + 1) % tabCount` (`:149`) now wraps 5→0.
- prev-tab `(activeTab + tabCount - 1) % tabCount` (`:153`) now wraps 0→5 (Agent→Security),
  satisfying `[[harness-init]]` "Shift-Tab from the first tab wraps to Security".

Struct/field/wiring deletions in `model.go`:
- `:52-53` `impeccableTab`/`graphifyTab` fields — delete.
- `:115-116` `newImpeccableTabState`/`newGraphifyTabState` in `NewModel` — delete.
- `:137-138` `setWidth` calls in `WindowSizeMsg` — delete.
- `:188-197` `case ImpeccableTab:` / `case GraphifyTab:` in key routing — delete.
- `:225-226`, `:234-235` rebuild + resize calls in `agentInitDoneMsg` — delete.
- `:297` tab-label slice: drop `"Impeccable"`, `"Graphify"` → six labels ending
  `"Security"`, satisfying "Tab set has no Impeccable or Graphify tab".
- `:328-331` `case ImpeccableTab:` / `case GraphifyTab:` in `renderTabContent` — delete.
- `:370-371` `applyToConfig` calls in `saveConfig` — delete.

Delete `internal/tui/impeccable_tab.go` entirely. (`graphify_tab.go` exists per the
same pattern — delete it too; the proposal's Affected-Areas row lists both.)

**`internal/tui/model_test.go`:** the prev-tab-from-Agent test (~:109-122) currently
expects `GraphifyTab` as the terminal tab; update its expectation to `SecurityTab`.
Delete `TestImpeccableTabState_ApplyToConfig*` (~:381-440) and any Graphify tab-state
test. Update any test enumerating `tabCount` (was 8 → now 6) or the label list.

### `internal/status/display.go` + test

- Delete the "Impeccable (Design Language)" block (~:53-65) and the Graphify status
  block. Delete `TestDisplay_Impeccable` (~:137-186) and the Graphify display test.

### Embedded skills + retired spec

- `skills/embed_test.go:29`: remove `"graphify"` from the expected embedded-skill list.
  (`"impeccable"` is NOT in that list — no edit for it there.)
- Delete `skills/graphify/` and `skills/impeccable/` directories.
- Delete `openspec/specs/graphify-integration/` (spec + `.feature`) — the physical
  deletion the retirement spec defers to apply. `skill_count` auto-corrects on
  `archon init`/`update` (generated from the embedded FS), so no manual count edit.

## Concern 3 (skill layer) — Engram/none removal + persistence-contract rewrite

### `skills/_shared/persistence-contract.md` rewrite (OpenSpec-only)

The file today (`:5-73`) defines four modes and an `artifact_store.mode` handshake.
Rewrite to a single OpenSpec path:
- Remove the `artifact_store.mode` selection line (`:5`), the "ASK the user which mode"
  handshake (`:7`), and the default-resolution line (`:9`).
- Collapse the mode table (`:20-40`) to a single "OpenSpec is the sole store" statement:
  artifacts are files under `openspec/changes/{change-name}/`, git-tracked, team-shareable.
- Delete the `engram`/`hybrid`/`none` rows, the "engram mode limitation" note (`:29-31`),
  the hybrid section (`:44-52`), and the per-mode read/write table (`:60-73`).
- Replace the state-persistence row with the single OpenSpec form:
  write/read `openspec/changes/{change-name}/state.yaml`.
- Remove the `none`-mode "return only" path (`:69,:74`) and the "default to none"
  fallback — satisfying `[[sdd-phase-skills]]` "No none-mode return-only path remains".
- The trailing prompt templates (`:103,:123`) that echo `{engram|openspec|hybrid|none}`
  become `OpenSpec`.

### `skills/_shared/engram-convention.md`

Delete entirely — satisfies `[[sdd-phase-skills]]` "engram convention module is gone".
Grep for inbound references (`persistence-contract.md`, phase skills, `chained-pr`) and
remove each link.

### Per-phase skill edits (strip engram/hybrid/none + Impeccable/Graphify)

For every phase skill, the pattern is: (a) delete the "Artifact store mode
(engram|openspec|hybrid|none)" input line, (b) delete the `IF mode is engram/hybrid/none`
branch blocks, keeping only the OpenSpec body inline (drop the now-redundant
`IF mode is openspec` guard), (c) delete Impeccable/Graphify conditional blocks.

| Skill | Engram/none removals (anchors from exploration) | Impeccable/Graphify removals |
|---|---|---|
| `sdd-explore` | `:46-48`, `:53-56` | Step 3c Impeccable rec (`:110-118`); Graphify consumption block |
| `sdd-propose` | `:41,:47,:49,:83,:89,:184` | — |
| `sdd-spec` | `:40,:46,:48,:76,:93,:275` | Impeccable annotation notes (`:112-115,:312-315`) |
| `sdd-design` | `:40,:46,:48,:102,:191` | Step 2b Impeccable block (`:64-88`) |
| `sdd-apply` | `:41,:49,:51,:98,:132,:230` | Step 4c (`:193-208`) + Rules line (`:315`) |
| `sdd-verify` | (mode branches) | Impeccable presence check (`:44-50`) |
| `sdd-tasks` | (mode branches) | Impeccable pass task (`:253-254,:315-321`); Graphify Leiden consumption |
| `sdd-archive` | `:40,:48,:50,:59,:107,:144,:172,:191,:217,:237,:258` | — |
| `sdd-init` | `:38,:40,:42,:51,:53,:75-76` (drop `track`-less engram handling) | — |
| `sdd-onboard` | `:37` (`engram|openspec|hybrid|none` → OpenSpec) | — |
| `chained-pr` + `references/chaining-details.md` | any engram refs | — |
| `_shared/session-status-contract.md` | any engram mode branches | — |

**`sdd-spec` (`:40,:46,:48,:76,:93,:275`)** — in addition to engram removal, this skill
is the site of the bugfix scoped-spec branch (Concern 6). The Gherkin-`.feature`
machinery (`:95-122,:273-304`) stays for the FULL track; the bugfix branch bypasses it.

**Security `@security` behavior is preserved** — the `security.enabled`-gated abuse-case
hooks (added by the archived security-baseline change) are NOT touched; the
`[[sdd-phase-skills]]` "Security abuse-case behavior is unchanged" scenario guards this.

## Concern 5 — Judge reduction (single focused review)

### Mechanism

`harness-judge` keeps its role as the SDD-path gate (Rule 6 unchanged), but the review
it delegates changes from a **dual adversarial** review to a **single focused** review,
DELEGATED to the same `archon-judge` subagent so the frontmatter `model:` hard gate
(Opus) still binds. `judgment-day` is no longer invoked by the SDD path.

### `.claude/agents/archon-judge.md` rewrite

Current body (verbatim above) says: *"your job is the dual adversarial review. Run the
`judgment-day` skill…"*. Replace with a single-review brief: read all files modified by
the change plus the change's spec/design, evaluate **spec compliance, design coherence,
and code quality**, and return one verdict (`pass`/`fail`) with any issues. Do NOT run
`judgment-day`; do NOT launch a second/blind judge; do NOT apply fixes or re-verify
(harness-judge still owns the re-apply loop). Frontmatter `model: claude-opus-4-8`
stays. This satisfies `[[harness-judge]]` "Single Focused Judge in the SDD Path" and the
"pinned judge model, not a dual review" edge scenario.

### `skills/harness-judge/SKILL.md` edits

- **Purpose (`:13`)** and **frontmatter description (`:3`)**: "dual adversarial review"
  → "single focused review". Drop "judgment-day" from the trigger prose.
- **Hard Rules (`:24`)**: "delegate the dual review to archon-judge" → "delegate a
  single focused review to archon-judge"; keep "do NOT run inline on the orchestrator's
  model".
- **Step 2 (`:84-91`) "Delegate Dual Review"**: rename to "Delegate Single Review";
  remove "dual adversarial", "two blind judges", "synthesis", and the "archon-judge
  invokes judgment-day internally" sentence. It now says: delegate one focused review;
  capture the returned `pass`/`fail` + issues.
- **Impeccable gate removal (Concern 2)**: delete Step 3c (`:125-165`), the
  `impeccable.enabled` Hard Rule (`:23,:30ff`), the Step-1 impeccable config read
  (`:71-77,:82`), the Step-4 result-table impeccable column (`:172-173`), the
  Output-Contract "### Impeccable Gate" section (`:270-278`), and the Impeccable
  Error-Handling rows (`:307-310`) and Rules bullet (`:321-323`).
- **Preserved (do NOT change)**: Step 0 judge-flag gate (`judge.enabled`), Step 3
  mutation gate, Step 3b Playwright gate, Step 4 evaluate, Step 6 3-retry re-apply
  loop, the Structured Feedback format. Replace every remaining "judgment-day" token
  with "the single judge / archon-judge review" so the pass/fail wording is consistent.
  This satisfies `[[harness-judge]]` "Preserved gates and loop under the single judge".

### `skills/judgment-day/SKILL.md`

**Untouched — byte-for-byte.** Its activation contract already restricts it to explicit
user requests ("juzgar"/"dual review"). Removing `harness-judge`'s automatic call is the
only change needed to make it opt-in-only — satisfies `[[harness-judge]]` "judgment-day
Remains Standalone and Untouched" and "still runs on explicit request".

### `openspec/specs/harness-judge/spec.md`

Live spec update lands in apply (spec-level deltas already authored in the change's
`specs/harness-judge/spec.md`): the MODIFIED "Single Focused Judge" requirement replaces
the dual-review requirement; the ADDED "judgment-day Remains Standalone" requirement is
appended.

### `[[harness-workflow]]` terminal ordering

Full-track terminal ordering keeps exactly one judge between verify and archive
(the "Single Judge in the Full-Track Terminal Ordering" ADDED requirement); no
harness-workflow SKILL change beyond track-awareness (Concern 6) is required for judge.

## Concern 6 — Bugfix track mechanics

### `state.yaml` schema

Add an optional top-level `track` field: `full | bugfix`. Semantics:
- **absent** → resolve as `full` (back-compat; in-flight changes untouched).
- **`bugfix`** → select the 5-phase sequence.
- **any other value** (e.g. `hotfix`) → the resolver reports an error naming the two
  supported values; it MUST NOT silently coerce to `full`. Satisfies
  `[[harness-bugfix-track]]` "Unknown track value is rejected".

```yaml
# state.yaml (bugfix example)
track: bugfix          # optional; absent == full
phase: spec
status: completed
history:
  - {phase: explore, status: completed, ts: "..."}
  - {phase: spec,    status: completed, ts: "..."}
```

### `skills/harness-workflow/SKILL.md` — track-aware PHASE_ORDER

- **Step 1 (`:86`)**: after parsing `phase`/`status`, also parse `track`. Missing →
  `full`. Unknown value → return `blocked` with reason `unknown track: {v} (supported:
  full, bugfix)`.
- **PHASE_ORDER (`:93`)**: replace the single hard-coded array with a track-selected one:

  ```
  PHASE_ORDER = track == "bugfix"
      ? [explore, spec, apply, verify, archive]
      : [explore, propose, spec, design, tasks, apply, verify, judge, archive]
  ```

  All downstream `.index()` transition logic (`:94-98`) is unchanged — it operates on
  whichever array `track` selected. Satisfies `[[harness-workflow]]` "Track selects the
  phase sequence" and `[[harness-bugfix-track]]` "Bugfix advances explore to spec".
- **Phase-skipping-prevention (`:62`)**: the "N → N+1 only" rule now applies within the
  selected array, so requesting `propose` on a bugfix change is an unknown/unreachable
  phase → `blocked` with "propose is not part of the bugfix track"; requesting `apply`
  before `spec` completes → `blocked` naming `spec`. Satisfies "Propose is not reachable"
  and "Spec cannot be skipped".
- **Judge omission**: because `judge` is simply absent from the bugfix array, no
  `judge.enabled` check applies — after `verify` the only next phase is `archive`,
  regardless of config. Satisfies "Verify transitions straight to archive" and
  "judge.enabled does not add a judge to the bugfix track". Add one Hard-Rule sentence:
  *"`judge.enabled` governs the full track only; the bugfix track never runs judge."*
- **Human Review Gate**: unchanged mechanics — it fires per artifact-producing phase.
  In the bugfix array the only such phase before apply is `spec`, so it fires once.

### `sdd-spec` — bugfix detection + 3-section scoped spec

`sdd-spec` reads `track` from `state.yaml` at entry. Branch:
- **`track: bugfix`** → emit a scoped spec and SKIP capability/requirement decomposition
  and Gherkin. Write a single `openspec/changes/{change-name}/spec.md` with EXACTLY three
  sections and NO `.feature` file:
  - `## Bug` — observed vs expected, 1–3 sentences.
  - `## Fix Criteria` — flat checklist of 2–5 verifiable acceptance bullets
    (`- [ ] X returns Y when Z`).
  - `## Non-Regression` — 1–2 bullets naming existing behavior that must not break.
  Explicitly bypass the "Gherkin Feature Files (MANDATORY)" block (`sdd-spec:95-122`) and
  the Rules that mandate a `{domain}.feature` per domain (`:303-304`). Satisfies
  `[[harness-bugfix-track]]` "Scoped spec format contract", "Fix Criteria is a flat
  verifiable checklist", and "Scoped spec emits no Gherkin feature file".
- **`track: full`** (or absent) → the existing full spec flow, unchanged.

### Track selection (where `track: bugfix` is set)

The track is chosen when the change's `state.yaml` is first created (explore/sdd-init).
Routing hint in `CLAUDE.md`/`AGENTS.md` (`templates.go` Vague Request Guard region):
add a brief note that when the user's request is a **bug report** (not a feature), the
orchestrator asks *"¿Es un bug o una nueva funcionalidad?"* and, on "bug", initializes
the change with `track: bugfix`. `sdd-init` accepts a `track` parameter (default `full`)
and writes it into `state.yaml`. No CLI flag is required for this change (natural-language
routing + the state field are sufficient); a flag is a possible future add.

### New live spec

`openspec/specs/harness-bugfix-track/spec.md` (+ `.feature`) is created in apply from the
change's authored `specs/harness-bugfix-track/` delta.

## Migration & Back-Compat

| Surface | Existing state | Post-change behavior |
|---|---|---|
| `.archon/config.yaml` with `impeccable:`/`graphify:` blocks | present | Loads clean — lenient `yaml.Unmarshal` ignores unknown keys. No migration. |
| `archon config set/get impeccable.*\|graphify.*` | worked | Unknown-key error (by design; documented in README). |
| `archon init --impeccable\|--graphify` | worked | Unknown-flag error (Cobra). |
| `state.yaml` without `track` | present (all in-flight changes) | Resolves as `full`; zero change to existing flows. |
| CLAUDE.md/AGENTS.md customized past the generated template | possible in consumer repos | `archon update` regenerates; hand-customized files need a manual merge (noted in README). |

**Rollback:** each concern is a self-contained PR; revert the offending PR(s). Removed Go
config keys are additive-free deletions — reverting restores struct/flag/tab/const
verbatim. No user-data migration is ever performed, so nothing to roll back on disk.

## PR-Slice Mapping (chained; 800-line budget)

The proposal's 5-PR plan holds. Below, each slice is reconciled with the design's file
touch-points and the two spec-expanded capabilities are placed.

| PR | Scope (design sections) | Capability specs it lands | Est. |
|----|---|---|---|
| **1** | Remove Impeccable Go plumbing: `config.go` (struct/field/Clone/Load), `config_test.go`, `cmd/archon/config.go` (impeccable cases + error strings), `main.go` (flag), `init.go` (Options/buildConfig), `impeccable_tab.go` delete, `model.go` iota+wiring for Impeccable, `model_test.go`, `status/display.go` + test | partial `[[harness-init]]` (Impeccable config/CLI/init/TUI invariants) | ~350 removed |
| **2** | Remove Graphify Go plumbing (mirror PR1): config struct+consts, CLI cases, flag, buildConfig, `graphify_tab.go` delete, `model.go` iota (now SecurityTab last, `tabCount`=6), `model_test.go` prev-tab→Security, `status` block; `embed_test.go` graphify entry; delete `skills/graphify/`; delete `openspec/specs/graphify-integration/` | `[[graphify-integration]]` (RETIRED) + remaining `[[harness-init]]` TUI six-tabs + Graphify-init invariants | ~350 removed |
| **3** | Skill layer: rewrite `persistence-contract.md` OpenSpec-only, delete `engram-convention.md`, strip engram/none + Impeccable/Graphify from all 8 phase skills + `sdd-init`/`sdd-onboard`/`chained-pr`/`session-status-contract`; strip Impeccable from `harness-judge` | `[[sdd-phase-skills]]` (both requirements) | ~400–600 removed |
| **4** | Preflight prose: `templates.go` `orchestratorSections` + rules 8–9 removal; `templates_test.go`; `CLAUDE.md`/`AGENTS.md`/`README.md` | `[[harness-workflow]]` "Preflight is a Standard Default Expressed as Prose"; `[[harness-init]]` "Artifact store is always OpenSpec" | ~200 changed |
| **5** | Single judge (`archon-judge.md` body, `harness-judge/SKILL.md` Step 2 + purpose, `harness-judge` live spec) **and** bugfix track (`state.yaml` `track`, `harness-workflow/SKILL.md` PHASE_ORDER + parse, `sdd-spec` scoped-spec branch, `harness-bugfix-track` live spec, CLAUDE.md routing hint) | `[[harness-judge]]`, `[[harness-workflow]]` (Phase State Machine + Single-Judge terminal), `[[harness-bugfix-track]]` | ~300 changed |

### Slice-size risk

- **PR 3 is the at-risk slice** — the exploration bounds it at ~400–600 removed lines,
  but it touches 8 phase skills + 4 shared/aux skills + the persistence-contract rewrite
  + the harness-judge Impeccable strip. If the actual diff approaches or crosses 800,
  **split PR 3 into 3a and 3b**:
  - **3a — Engram/none removal**: `persistence-contract.md` rewrite, delete
    `engram-convention.md`, strip engram/hybrid/none from all 8 phase skills +
    `sdd-init`/`sdd-onboard`/`chained-pr`/`session-status-contract`.
  - **3b — Impeccable/Graphify hook removal from skills**: the per-skill Impeccable/
    Graphify conditional blocks + the `harness-judge` Impeccable gate strip.
  Both sub-slices are independently reviewable; 3b depends on 3a only for touching the
  same files (rebase, not logic). This split is the recommended contingency; keep PR 3
  whole only if the measured diff stays comfortably under 800.
- **PR 5** carries two logically distinct behaviors (single-judge, bugfix-track). If it
  exceeds 800, split at the natural seam: **5a single judge**, **5b bugfix track** — they
  share no files except the (already-authored) spec deltas, so the split is clean.
- PRs 1, 2, 4 are each comfortably under budget.

**Dependencies:** PR3 depends on PRs 1–2 (Go removal unlocks skill cleanup); PRs 4 and 5
are independent of each other. Recommended order: 1 → 2 → 3 → 4 → 5.

## Open Questions

None blocking. The one prior open question (bugfix scoped-spec format) is fixed by the
proposal's 3-section contract, realized above in the `sdd-spec` branch.
