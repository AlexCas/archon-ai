# Exploration: Impeccable frontend-integration (design-language gate mirroring Playwright)

## Project Type
**Web testing**: not-web

This repository (`archon-ai`) is itself a Go CLI/TUI harness, not a web frontend. The
Impeccable integration being explored is a *harness feature* that will apply to the
TARGET projects Archon initializes, exactly like the existing Playwright feature —
which is opt-in and disabled here. This is not the SDD flow for a web app; it is a
harness change. Web-detection / Playwright preflight for THIS repo stays `not-web`.

## Confirmed Integration Model (given, not re-litigated)
1. External npm tool invoked via `npx impeccable ...` into the TARGET project. Never
   vendored into the Go binary. Archon only orchestrates WHEN/HOW to call it.
2. Config-flag gated (`impeccable.enabled`) mirroring `playwright.enabled`. Phases
   touched: design → apply → verify/judge, with a detection gate after verify/judge
   analogous to the Playwright E2E gate.
3. A thin Archon `impeccable` skill orchestrates the calls; it does not reimplement
   Impeccable's detector/rules.

---

## Current State — The Playwright Precedent (file-by-file map)

Two harness "opt-in gate" precedents exist. The best structural template is
**Playwright** (dedicated config struct + TUI tab + init flag + per-phase skill hooks +
a post-verify/judge execution gate). A second, lighter precedent is **Security**
(`Security` config struct, `--security` flag, TUI tab, and phase hooks woven directly
into existing phase skills with NO dedicated skill file). Impeccable needs the FULL
Playwright shape PLUS one thing Playwright lacks: a dedicated thin skill.

### 1. Config — `internal/config/config.go`
- **Struct** `Playwright` (lines 27-35): `Enabled bool yaml:"enabled"`,
  `TestDir string yaml:"test_dir,omitempty"`, `BaseURL string yaml:"base_url,omitempty"`,
  with a doc comment describing generate-during-apply / run-after-verify-and-judge.
- **Config field** (line 59): `Playwright Playwright yaml:"playwright"`.
- **Clone()** (line 105): `Playwright: c.Playwright` — value copy. NOTE the loud
  contract comment (lines 89-95): *every new Config field MUST be added to Clone or it
  is silently dropped on every `archon update`*. `TestConfig_CloneRoundtrip` fails if
  missed.
- **Default**: `Playwright` has NO pre-seed in `Load()` (only `Judge.Enabled=true` is
  pre-seeded at line 79). So the zero value `Enabled:false` is the default when the
  YAML section is absent — the "opt-in, default off" behavior.
- On-disk shape confirmed in `.archon/config.yaml`:
  ```yaml
  playwright:
      enabled: false
  ```

### 2. CLI config get/set — `cmd/archon/config.go`
- **setConfigValue** (lines 157-169): cases `playwright.enabled` (parseBool),
  `playwright.test_dir`, `playwright.base_url`.
- **getConfigValue** (lines 210-215): symmetric getters.
- **Two error strings** (lines 202 and 230) enumerate every supported key —
  `playwright.enabled, playwright.test_dir, playwright.base_url` are listed in BOTH and
  must be updated together for any new keys.

### 3. Init flag + scaffold — `internal/initcmd/`
- **init.go**: `Options.Playwright bool` (lines 26-27). Threaded into
  `buildConfig(..., opts.Playwright, ...)` (line 87). `buildConfig` signature takes
  `playwright bool` (line 220) and sets `Playwright: config.Playwright{Enabled: playwright}`
  (lines 242-244).
- **cmd/archon/main.go**: `playwrightFlag bool` var (line 82); passed into
  `initcmd.Options{... Playwright: playwrightFlag ...}` (line 169); registered as
  `cmd.Flags().BoolVar(&playwrightFlag, "playwright", false, "...")` (line 201).
  `securityFlag` (lines 83, 170, 202) is the same pattern.
- **update.go**: `Update()` PRESERVES `playwright` untouched — it clones the loaded
  config and patches ONLY `Version`, `SkillCount`, `SkillInventory` (lines 130-137,
  doc comment lines 49-52). So a new `impeccable` config block automatically rides
  through `archon update` **as long as it is in `Clone()`**. No update.go edit needed
  for the config field itself, but the skill_inventory / embedded-skill path (below) is
  what actually installs a new skill.

### 4. Skill embedding — `skills/embed.go` + `internal/scaffold/version.go`
- **skills/embed.go**: `//go:embed */SKILL.md all:_shared` — a GLOB. Any new
  `skills/<name>/SKILL.md` directory is embedded automatically; no code edit to add a
  skill to the embedded set.
- **scaffold/version.go** `ClassifyGaps` walks the embedded FS dir-by-dir (any dir with
  a `SKILL.md` counts as a skill) and diffs against the installed set. A new
  `skills/impeccable/SKILL.md` is auto-classified as `Added` and installed by
  `refreshSkills`/`Extract`.
- **skill_inventory** in config is auto-generated from `res.Inventory` (init) and
  `res.Inventory` on update — NOT hand-maintained. Adding the skill dir bumps
  `skill_count` from 24→25 and appends the inventory entry automatically. The
  hardcoded `Skills: 24` line in CLAUDE.md is rendered from `{{.SkillCount}}` at init.

### 5. TUI tab — `internal/tui/playwright_tab.go` + `internal/tui/model.go`
- **playwright_tab.go** is a self-contained ~140-line tab: `playwrightTabState`
  struct (enabled bool + two `textinput.Model`s + focused int), `newPlaywrightTabState`,
  `update`, `refocus`, `view`, `applyToConfig`, `setWidth`, and a `playwrightFocusCount`
  const. Directly clonable into `impeccable_tab.go`.
- **model.go** wiring points (ALL must be mirrored for a new tab):
  - Tab enum (lines 22-28): `PlaywrightTab` between `MutationTab` and `SecurityTab`;
    `tabCount` sentinel closes the iota block. Inserting a tab shifts indices — the
    `tabs []string` slice order (line 274) must match the enum order.
  - Field on Model (line 47): `playwrightTab playwrightTabState`.
  - Constructor (line 108): `playwrightTab: newPlaywrightTabState(cfg.Playwright)`.
  - `setWidth` on resize (line 128) and after reload (line 211).
  - Key routing in `Update` (lines 169-170): `case PlaywrightTab: ...update(msg)`.
  - Reload after save (line 204): `m.playwrightTab = newPlaywrightTabState(msg.cfg.Playwright)`.
  - `renderTabs` labels slice (line 274): `"Playwright"` — order-coupled to enum.
  - `renderTabContent` (lines 301-302): `case PlaywrightTab: ...view(...)`.
  - `saveConfig` (line 341): `m.playwrightTab.applyToConfig(cfg)`.
  - `model_test.go` references Playwright — tab-count / tab-order tests will need the
    new tab added.

### 6. Status display — `internal/status/display.go`
- Lines 40-48 print a "Playwright (Web E2E)" section (Enabled + conditional TestDir /
  BaseURL). `archon status` output. A parallel "Impeccable (Design Language)" block
  belongs here.

### 7. Phase skill hooks (the conditional invocations)
Where Playwright is woven into the SDD phase skills — these are the exact hook points
an Impeccable equivalent replicates (adjusted for design→apply→verify/judge):
- **skills/sdd-spec/SKILL.md** (lines 100, 110, 298): `@web` scenario tagging so
  Playwright generation can select. (Impeccable's analog: tag/flag frontend-affecting
  requirements so the design-language checks can select — OPEN QUESTION whether needed.)
- **skills/sdd-apply/SKILL.md** — **Step 4b** (lines 162-179): "Generate Playwright
  Specs from Gherkin (web projects only) — Only if `playwright.enabled: true`". Also a
  Rules line (line 271). This is the "do work during apply when enabled" hook.
- **skills/sdd-verify/SKILL.md** (line 43): verify CONFIRMS specs were generated for
  `@web` scenarios but does NOT execute them (execution is judge's job). Reports
  missing specs as CRITICAL.
- **skills/harness-judge/SKILL.md** — the execution gate:
  - Line 13, 23, 28: intro + config-read rules.
  - Lines 67-70: config snippet block.
  - **Step 3b (lines 100-115)**: "Playwright E2E Gate (conditional) — Only if
    `playwright.enabled: true` AND judgment-day passed". Ensure server reachable → run
    suite → parse → pass/fail.
  - Step 4 result table (lines 122-127): playwright column folds into overall pass.
  - Edge cases (lines 243-246): config missing default false; "Playwright not
    installed → blocked"; "server unreachable → blocked".
  - Output contract (lines 215-217): "### Playwright Gate — Status / Scenarios".
- **skills/sdd-explore/SKILL.md** (Step 3b, lines 87-107, 136): web detection → drives
  `playwright.enabled` and, for unknown/blank projects, preflight group E.
- **skills/sdd-tasks/SKILL.md** (lines 198-201, 257): add a task to generate Playwright
  specs when web + enabled.
- **skills/sdd-design/SKILL.md**: **ZERO** Playwright references. Playwright currently
  does NOT touch design. Impeccable explicitly wants a design-phase hook, so this is
  NET-NEW surface with no precedent line to copy — must be authored fresh.

### 8. Orchestrator template + CLAUDE.md — `internal/initcmd/templates.go`
- The orchestrator markdown (`CLAUDE.md` / `AGENTS.md`) is generated from hardcoded Go
  string consts in templates.go, NOT from the checked-in CLAUDE.md. Both must stay in
  sync (the repo's own CLAUDE.md is effectively a rendered copy).
- **Preflight group E** lives at lines 80-85 (`orchestratorSections` const): the
  arrow-key question + "Group E maps to `playwright.enabled`" mapping paragraph.
- **Rules line 6** (lines 175 and 188, in both `orchestratorRulesClaude` and
  `orchestratorRulesOpencode`): "When playwright.enabled, run the generated Playwright
  tests after verify and judge pass."
- The template body is **static** (only `{{.SkillCount}}`, `{{.Agent}}`,
  `{{.HarnessVersion}}`, `{{.PhaseModels}}` are dynamic). There is NO `{{if .Playwright}}`
  conditional — the Playwright preflight text is ALWAYS emitted; the runtime
  `playwright.enabled` flag decides behavior. An Impeccable group F would follow the
  same "always-emitted text, flag-gated behavior" model.
- `templates_test.go` asserts on template content (group E, rule 6) → new assertions
  needed.

---

## Affected Areas (touchpoint checklist for the Impeccable integration)
- `internal/config/config.go` — add `Impeccable` struct + `Config.Impeccable` field +
  `Clone()` copy (CloneRoundtrip gate). Decide fields: at minimum `Enabled bool`; likely
  `AutoInstall bool` and a docs path (`ProductPath` / `DesignPath`), plus maybe a
  detector-severity / blocking flag.
- `internal/config/config_test.go` — extend `TestConfig_CloneRoundtrip` fixture (line
  222 area) with the new struct so the round-trip passes.
- `cmd/archon/config.go` — add `impeccable.enabled` (+ any sub-keys) to setConfigValue,
  getConfigValue, AND both key-list error strings (lines 202, 230).
- `internal/initcmd/init.go` — `Options.Impeccable bool`; thread through
  `buildConfig`; set `Impeccable: config.Impeccable{Enabled: impeccable}`.
- `cmd/archon/main.go` — `impeccableFlag` var + `Options` field + `--impeccable`
  BoolVar registration.
- `internal/tui/impeccable_tab.go` — NEW, cloned from `playwright_tab.go`.
- `internal/tui/model.go` — enum entry (+ shift `tabCount`), Model field, constructor,
  setWidth x2, key routing, reload, `tabs` label slice, renderTabContent, saveConfig.
- `internal/tui/model_test.go` — tab count/order test updates.
- `internal/status/display.go` — Impeccable status block.
- `skills/impeccable/SKILL.md` — NEW thin orchestration skill (auto-embedded, auto-
  inventoried; bumps skill_count 24→25).
- `skills/sdd-design/SKILL.md` — NEW hook: when `impeccable.enabled`, run/consult
  Impeccable (e.g. `impeccable init` producing PRODUCT.md/DESIGN.md, or `shape`) and
  fold design-language constraints into design.md.
- `skills/sdd-apply/SKILL.md` — new Step (parallel to 4b): when enabled, invoke the
  relevant Impeccable subcommands (craft/polish/harden/animate) during apply + a Rules
  line.
- `skills/sdd-verify/SKILL.md` — verify confirms Impeccable artifacts/hooks are present
  (advisory), deferring the blocking detection run to judge (mirrors Playwright).
- `skills/harness-judge/SKILL.md` — new "Step 3c: Impeccable Detection Gate" +
  result-table column + edge cases (npx/Node missing → blocked; detector fail →
  fail-or-advisory) + output-contract section.
- `skills/sdd-tasks/SKILL.md` — add an "Impeccable pass" task when enabled.
- `skills/sdd-explore/SKILL.md` — web detection already drives this; add that a `web`
  project should also suggest enabling Impeccable (group F).
- `internal/initcmd/templates.go` — preflight **group F** ("¿Aplicar Impeccable...?")
  + mapping paragraph + a new Rules line; keep `orchestratorRulesClaude` and
  `orchestratorRulesOpencode` in sync.
- `internal/initcmd/templates_test.go` — assertions for group F + new rule.
- `CLAUDE.md` (repo root) — mirror the template changes (group F, rule) so the repo's
  own orchestrator file matches the generator.
- `README.md` / `AGENTS.md` — mention `--impeccable` (docs, low-risk).

---

## Approaches
1. **Full Playwright-mirror (dedicated struct + tab + flag + gate) + thin skill** —
   Replicate every Playwright touchpoint, add the one missing piece (dedicated
   `impeccable` skill), and extend into the design phase (net-new).
   - Pros: consistent with an established, tested pattern; reviewers already understand
     it; each touchpoint is small and mechanical.
   - Cons: touches ~15 files across Go + skills + template + tests; the design-phase
     hook and the judge detection gate are net-new (no line to copy).
   - Effort: Medium.

2. **Config + skill only, no TUI/flag (minimal)** — Add `impeccable.enabled` to config
   and the phase skills, skip the `--impeccable` init flag and TUI tab initially.
   - Pros: much smaller first slice; skills carry most of the behavior.
   - Cons: diverges from Playwright's UX (no tui/flag parity); `archon tui` and
     `archon init --impeccable` would be inconsistent; likely a fast-follow anyway.
   - Effort: Low.

3. **Reuse the Playwright config shape as a generic "web quality gates" block** —
   Fold Impeccable under a shared gate abstraction.
   - Pros: DRY if more gates arrive.
   - Cons: premature abstraction; the repo currently keeps each gate as its own struct
     (Playwright, Security, MutationTesting all separate); would break the established
     one-struct-per-gate convention and the CloneRoundtrip test shape.
   - Effort: High. Not recommended.

## Recommendation
Approach **1 (full Playwright-mirror + thin skill)**, sliced into chained PRs. The Go
plumbing (config/flag/tui/status/tests) is one cohesive, low-risk slice; the skill +
phase-hook + template/CLAUDE.md changes are a second slice. This keeps each PR under
the 400-line review budget and separates "wiring" from "behavior/prose". The design-
phase hook and judge detection gate need fresh authoring — flag those as the highest-
uncertainty items for the propose/design phases, not explore.

## Risks
- **Node/npx availability**: Impeccable runs via `npx impeccable`. Target projects may
  lack Node. The judge gate must fail as `blocked` with an actionable message
  ("impeccable not found — install Node/npx or disable impeccable"), mirroring the
  Playwright "not installed → blocked" edge case (harness-judge lines 245-246).
- **Blocking vs advisory**: Impeccable's detector = 58 deterministic rules + LLM
  critique. Deciding whether a detector failure *blocks* the judge gate (like
  Playwright) or is *advisory-only* is a core design decision. LLM-critique checks are
  non-deterministic and could make the gate flaky if blocking. Recommend: deterministic
  rules block, LLM critique advisory (or a config `severity`/`blocking` knob).
- **PRODUCT.md / DESIGN.md ownership**: `impeccable init` generates these in the TARGET
  project. Where they live, whether Archon commits them, and whether they collide with
  SDD's own `design.md` must be settled (naming/location clash risk with
  `openspec/changes/.../design.md`).
- **Auto-install vs assume-installed**: `npx impeccable install` sets up hooks. Decide
  whether Archon runs install automatically (side effects in the target repo) or assumes
  a pre-installed tool. Recommend a config `auto_install` flag defaulting off, matching
  the "opt-in, no surprise side effects" ethos.
- **Design-phase hook has no precedent**: Playwright never touches design; this surface
  is net-new and the highest-uncertainty part of the change.
- **Template/CLAUDE.md drift**: templates.go consts and the repo CLAUDE.md must stay
  byte-consistent; templates_test.go will catch template drift but not the repo file —
  update both.
- **Tab index shift**: inserting an enum value between existing tabs reorders indices;
  the `tabs []string` slice and any index-based tests must move in lockstep.

## Slicing Hint
Likely **2 chained PRs** (each < 400 lines):
- **PR 1 — Go wiring**: `Impeccable` config struct + Clone + CloneRoundtrip fixture,
  `cmd/archon/config.go` get/set + key strings, `--impeccable` flag (main.go +
  init.go/buildConfig), `impeccable_tab.go` + model.go wiring + model_test.go,
  status/display.go. Self-contained, fully testable via `go test ./...`, no behavior
  change until skills consume the flag.
- **PR 2 — Orchestration behavior**: new `skills/impeccable/SKILL.md`, phase-skill
  hooks (design/apply/verify/judge/tasks/spec/explore), templates.go group F + rule,
  repo CLAUDE.md mirror, README. Prose-heavy; the design-phase hook and judge detection
  gate get the most design attention.
If PR 2's skill prose runs large, split the judge-gate + template changes into a PR 3.

## Ready for Proposal
Yes. The precedent is fully mapped and the touchpoint list is concrete. The propose
phase should resolve the open design decisions (blocking-vs-advisory detector,
auto-install default, PRODUCT.md/DESIGN.md location, whether the design-phase hook is
in-scope for the first change) and confirm the 2-PR chained slicing against the
session's PR strategy.
