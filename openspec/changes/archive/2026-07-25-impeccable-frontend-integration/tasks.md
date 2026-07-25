# Tasks: Impeccable Frontend-Integration (design-language gate)

<!-- Link convention: [[capability]] wikilinks for capability identity; relative links
     for intra-change navigation. This file lives at
     changes/impeccable-frontend-integration/tasks.md. -->

Implements [[impeccable-gate]]. Builds on [proposal](proposal.md), [spec](specs/impeccable-gate/spec.md), and [design](design.md).

---

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~840 (PR1 ~330, PR2 ~360–510 split across PR2+PR3) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR1 (Go wiring) → PR2 (skills + phase hooks) → PR3 (judge gate + templates, conditional if PR2 > 400) |
| Delivery strategy | ask-always |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| PR1 | Go wiring — config, CLI, flag, TUI, status | PR1 → master | Fully testable via `go test ./...`; no behavior change when disabled |
| PR2 | Orchestration — thin skill + 6 phase hooks | PR2 → master (after PR1) | Prose-heavy; confirm line count at ask-always gate |
| PR3 | Judge gate + templates/CLAUDE.md (conditional) | PR3 → master (after PR2) | Split from PR2 if PR2 > 400; each stays under budget |

---

## PR1 — Go Wiring (~330 lines, no behavior change)

### Phase 1: Config Struct and Clone

- [x] 1.1 `internal/config/config.go`: Add the `Impeccable` struct after the `Security` struct (~line 44). Five fields: `Enabled bool`, `AutoInstall bool`, `Severity string yaml:"severity,omitempty"`, `ProductPath string yaml:"product_path,omitempty"`, `DesignPath string yaml:"design_path,omitempty"`. Mirrors design §2.1.
- [x] 1.2 `internal/config/config.go`: Add `Impeccable Impeccable yaml:"impeccable"` field to the `Config` struct after `Security` (~line 60). Mirrors design §2.3.
- [x] 1.3 `internal/config/config.go`: Add `Impeccable: c.Impeccable,` to `Clone()` after the `Security` copy (~line 106). Value copy is a full deep copy (all scalars). Mirrors design §2.4.
- [x] 1.4 `internal/config/config.go`: In `Load()`, after unmarshal, normalize empty `Severity` to `"block-deterministic"` before validation. Then call `validateImpeccableSeverity`. Add package-level `ValidImpeccableSeverities` slice and `validateImpeccableSeverity` helper (exported). Mirrors design §2.2. Covers spec scenarios: "severity defaults to block-deterministic", "Invalid severity value rejected at config load".
- [x] 1.5 `internal/config/config_test.go`: Extend `TestConfig_CloneRoundtrip` fixture with a non-zero `Impeccable` block (all five fields set). Mirrors design §2.5. Covers spec scenario: "CloneRoundtrip fixture catches missing Clone wiring".
- [x] 1.6 `internal/config/config_test.go`: Add `TestImpeccable_DefaultsAndValidation`: (a) absent block → `Severity == "block-deterministic"` and `Enabled == false`; (b) `severity: foobar` fixture → `Load()` returns error naming value + three valid options. Covers spec scenarios: "Zero-value config is fully inert", "Invalid severity value rejected at config load".

### Phase 2: CLI get/set

- [x] 2.1 `cmd/archon/config.go`: In `setConfigValue`, add 5 cases after `security.profile` (~line 189): `impeccable.enabled` (parseBool), `impeccable.auto_install` (parseBool), `impeccable.severity` (call `config.ValidateImpeccableSeverity`), `impeccable.product_path` (string assign), `impeccable.design_path` (string assign). Mirrors design §3.1.
- [x] 2.2 `cmd/archon/config.go`: In `getConfigValue`, add 5 matching cases after `security.profile` (~line 221): format bools with `strconv.FormatBool`, return strings directly. Mirrors design §3.2.
- [x] 2.3 `cmd/archon/config.go`: Append all five impeccable keys to BOTH key-list error strings (lines ~202 and ~230). The two strings must be byte-identical to each other after the update. Mirrors design §3.3.
- [x] 2.4 `cmd/archon/config_test.go` (or equivalent): Assert unknown-key error contains all five impeccable keys. Assert invalid severity via `config set impeccable.severity invalid` exits non-zero and names the value + three valid options. Covers spec scenarios: "Get unknown impeccable key shows updated supported-key list", "CLI rejects invalid severity via set".

### Phase 3: Init Flag Threading

- [x] 3.1 `cmd/archon/main.go`: Add `var impeccableFlag bool` beside `securityFlag` (~line 83). Pass `Impeccable: impeccableFlag` into the `initcmd.Options{...}` literal (~line 170). Register `cmd.Flags().BoolVar(&impeccableFlag, "impeccable", false, "Enable the Impeccable design-language gate")` (~line 202). Mirrors design §4.
- [x] 3.2 `internal/initcmd/init.go`: Add `Impeccable bool` to `Options` struct beside `Security` (~line 27). Thread `opts.Impeccable` into the `buildConfig(...)` call (~line 87). Add `impeccable bool` parameter to `buildConfig` signature (~line 220). In the returned `config.Config{}`, add `Impeccable: config.Impeccable{Enabled: impeccable}` — do NOT set `Severity` (Load normalizes). Mirrors design §4. Covers spec scenarios: "Init with --impeccable flag enables the gate", "Init without --impeccable leaves gate disabled", "buildConfig receives the flag value".

### Phase 4: TUI Tab

- [x] 4.1 `internal/tui/impeccable_tab.go` (NEW): Create as a clone of `playwright_tab.go` extended for Impeccable. Implement `impeccableTabState` struct (5 fields: `enabled bool`, `autoInstall bool`, `severity textinput.Model`, `productPath textinput.Model`, `designPath textinput.Model`, `focused int`). Constant `impeccableFocusCount = 5`. Mirrors design §5.2.
- [x] 4.2 `internal/tui/impeccable_tab.go`: Implement `newImpeccableTabState(cfg config.Impeccable)`: seed toggles from cfg; severity placeholder `"block-deterministic (default)"`; productPath placeholder `"PRODUCT.md (default: project root)"`; designPath placeholder `"DESIGN.md (default: project root)"`.
- [x] 4.3 `internal/tui/impeccable_tab.go`: Implement `update`, `refocus`, `view` (title: `"Impeccable (Design Language) Configuration"`; two toggle lines + three labeled inputs; info footer about `npx impeccable detect`), `applyToConfig`, and `setWidth`. Mirrors design §5.2.
- [x] 4.4 `internal/tui/model.go`: Apply all 10 lockstep edit sites in order per design §5.3: (1) add `ImpeccableTab` after `SecurityTab` before `tabCount` in the Tab enum; (2) add `impeccableTab impeccableTabState` model field; (3) `newImpeccableTabState` in constructor; (4) `setWidth` on resize; (5) `Update` key routing case; (6) reload after `agentInitDoneMsg`; (7) `setWidth` after reload; (8) append `"Impeccable"` to `renderTabs` label slice; (9) `renderTabContent` case; (10) `saveConfig` apply call.
- [x] 4.5 `internal/tui/model_test.go`: Change `TestModel_Update_ShiftTabWrapsFromAgent` (~line 120) expected wrap target from `SecurityTab` to `ImpeccableTab`. Append `"Impeccable"` to the `labels` slice in `TestModel_renderTabs_Order` (~line 221). Mirrors design §5.4. Covers spec scenario: "Tab count and order tests pass with new tab".
- [x] 4.6 `internal/tui/model_test.go`: Add `TestImpeccableTabState_ApplyToConfig`: drive toggles (`enabled`, `autoInstall`) and text inputs (`severity`, `productPath`, `designPath`); assert all five `cfg.Impeccable.*` fields. Mirrors design §5.4. Covers spec scenarios: "Impeccable tab renders current config", "TUI save persists Impeccable changes".

### Phase 5: Status Block

- [x] 5.1 `internal/status/display.go`: Add an "Impeccable (Design Language)" block parallel to "Playwright (Web E2E)". When `Enabled: false` show disabled state only; when `Enabled: true` show all non-empty fields (Enabled, Severity, Product Path, Design Path). Mirrors design §1. Covers spec scenarios: "Status shows Impeccable as disabled", "Status shows Impeccable as enabled with config details".

### Phase 6: Doc Update (PR1)

- [x] 6.1 `README.md` and `AGENTS.md`: Document `--impeccable` beside `--playwright`/`--security` (one line each). Mirrors design §4.

---

## PR2 — Orchestration: Thin Skill + Phase Hooks (~200–510 lines; confirm at ask-always gate)

> If PR2 exceeds 400 lines, split: PR2 retains phase hooks (§7–§12), PR3 takes judge gate + templates/CLAUDE.md (§13–§15). Confirm the split at the ask-always PR gate.

### Phase 7: Thin Impeccable Skill (NEW)

- [x] 7.1 `skills/impeccable/SKILL.md` (NEW): Create the thin orchestration skill. Document: per-phase invocation map (design = read PRODUCT.md/DESIGN.md; apply = `/impeccable <verb>` slash commands; verify = advisory note; judge = `npx impeccable detect --json .`). Document detect invocation signature, exit-code/JSON interpretation, severity mapping, Node/npx-missing → blocked messages (verbatim from design §6.3), auto_install semantics, PRODUCT.md/DESIGN.md ownership, and two-surface warning (npx CLI vs `/impeccable` slash commands). Mirrors design §8. Covers spec scenarios: "Impeccable skill is auto-embedded and skill_count becomes 25", "Skill delegates detection to npx, not Go code". Note: file is auto-embedded by glob; skill_count bumps 24→25 automatically.

### Phase 8: Design-Phase Hook

- [x] 8.1 `skills/sdd-design/SKILL.md`: Add a flag-gated read-only hook. When `impeccable.enabled: true` and `PRODUCT.md`/`DESIGN.md` exist at target root (or `product_path`/`design_path`), read them as input context before drafting design.md. When docs absent, proceed normally and add a note recommending `/impeccable init` (slash command, not npx). When disabled, no change. Explicitly: MUST NOT run `npx impeccable detect`, MUST NOT run any slash command, MUST NOT overwrite SDD design.md. Mirrors design §7. Covers spec scenarios: "Design references PRODUCT.md and DESIGN.md when both exist", "Design continues normally when Impeccable docs are missing", "Design phase unchanged when impeccable is disabled".

### Phase 9: Apply-Phase Hook

- [x] 9.1 `skills/sdd-apply/SKILL.md`: Add a new step parallel to Step 4b. When `impeccable.enabled: true` and the change includes frontend-affecting files, instruct the agent to run `/impeccable <verb>` (craft/polish/harden/animate) as slash commands — NOT npx shell-outs. Add a Rules line: "When impeccable.enabled, run Impeccable design verbs on frontend-affecting changes during apply." Mirrors design §8.1. Covers spec scenario: "Apply step invokes Impeccable on frontend changes when enabled".

### Phase 10: Verify-Phase Hook

- [x] 10.1 `skills/sdd-verify/SKILL.md`: Add a flag-gated advisory check. When `impeccable.enabled: true`, include a NOTE (not CRITICAL) if Impeccable hooks/artifacts are absent. Never blocks. Does NOT run the detection gate. Mirrors design §8.1. Covers spec scenario: "Verify reports missing Impeccable artifacts as advisory".

### Phase 11: Tasks-Phase Hook

- [x] 11.1 `skills/sdd-tasks/SKILL.md`: Add conditional logic: when `impeccable.enabled: true` and the change touches frontend files, emit an "Impeccable pass" task in the task list instructing apply to run the relevant Impeccable subcommands. Mirrors design §8.1. Covers spec requirement: "Tasks phase — add Impeccable pass task".

### Phase 12: Explore and Spec Phase Hooks

- [x] 12.1 `skills/sdd-explore/SKILL.md`: Add detection logic: when the target project is identified as web/frontend, recommend enabling Impeccable (preflight group F). Recommendation only, not auto-activation. Mirrors design §8.1.
- [x] 12.2 `skills/sdd-spec/SKILL.md`: Add lightweight annotation guidance: when writing specs for frontend design-language requirements, suggest a `@design` prose note (not a new hard tag, to avoid coupling with `@web`). Mirrors design §8.1.

---

## PR3 — Judge Gate + Templates (conditional split from PR2 if PR2 > 400, ~165 lines)

### Phase 13: Judge Detection Gate

- [x] 13.1 `skills/harness-judge/SKILL.md`: Add Step 3c "Impeccable Detection Gate (conditional)" after Step 3b (Playwright). Implement the 8-step gate flow from design §6.1: check enabled → check node/npx → auto_install if true → `npx impeccable detect --json .` → parse JSON (deterministic vs LLM critique; never rely on exit code for pass/fail) → apply severity → emit "### Impeccable Gate" section → fold into result table. Mirrors design §6.
- [x] 13.2 `skills/harness-judge/SKILL.md`: Add the "### Impeccable Gate" output section format per design §6.2 (Status, Severity mode, Deterministic violations count, Advisory findings count, Details, blocked Reason). Covers spec scenario: "Gate output contract".
- [x] 13.3 `skills/harness-judge/SKILL.md`: Add blocked message strings verbatim per design §6.3: node/npx missing message and package-missing-with-auto_install-false message. Covers spec scenarios: "Judge gate returns blocked when Node/npx is absent", "npx impeccable not found with auto_install false instructs rather than installs".
- [x] 13.4 `skills/harness-judge/SKILL.md`: Update the 5 additional touchpoints per design §6.4: intro/config-read rules, config snippet block, Step 4 result table (add Impeccable column), edge cases, output contract. Covers spec scenarios: "Judge gate passes when no violations found", "Judge gate blocks on deterministic violations", "Judge gate skipped when impeccable is disabled".

### Phase 14: Templates and CLAUDE.md

- [x] 14.1 `internal/initcmd/templates.go`: In `orchestratorSections`, add preflight group F after group E (~lines 80–85) with the exact text from design §9.1 and the group F mapping paragraph from design §9.2. Covers spec scenario: "Generated CLAUDE.md includes preflight group F".
- [x] 14.2 `internal/initcmd/templates.go`: In BOTH `orchestratorRulesClaude` (~line 175) and `orchestratorRulesOpencode` (~line 188), add the new rule line: "When impeccable.enabled, run Impeccable subcommands during apply and the detection gate after judge passes." The two consts must be byte-consistent with each other. Mirrors design §9.3. Covers spec scenario: "Generated CLAUDE.md includes the Impeccable rule".
- [x] 14.3 `internal/initcmd/templates_test.go`: Add assertions: generated CLAUDE.md contains group F question line ("F. Impeccable"); contains the group F mapping paragraph ("Group F maps to `impeccable.enabled`"); contains the Impeccable rule line in both Claude and Opencode variants. Mirrors design §9.5. Covers spec scenario: "Template tests assert group F and the Impeccable rule".
- [x] 14.4 `CLAUDE.md` (repo root): Manually mirror the template changes — add group F preflight question + mapping paragraph to the Preflight section, and add the Impeccable rule line to the `## Rules` section. Must be byte-consistent with what `archon init` would generate. NOTE: `templates_test.go` does NOT catch drift in this file — this edit is a manual step. Mirrors design §9.4. Covers spec scenario: "Repo CLAUDE.md is consistent with the template".

---

## Final Verification (all PRs)

- [ ] 15.1 Run `go test ./...` after PR1 lands. Confirm: `TestConfig_CloneRoundtrip`, `TestImpeccable_DefaultsAndValidation`, `TestImpeccableTabState_ApplyToConfig`, `TestModel_Update_ShiftTabWrapsFromAgent`, `TestModel_renderTabs_Order`, and CLI config key-list tests all pass.
- [ ] 15.2 Run `go test ./internal/initcmd/...` after PR3 lands. Confirm: all group F and Impeccable rule template assertions pass.
- [ ] 15.3 Manual smoke: `archon init --impeccable` → `.archon/config.yaml` contains `impeccable.enabled: true`; `archon status` shows the Impeccable block; `archon tui` Impeccable tab is navigable and saves correctly; `archon config set impeccable.severity invalid` exits non-zero with correct error.
