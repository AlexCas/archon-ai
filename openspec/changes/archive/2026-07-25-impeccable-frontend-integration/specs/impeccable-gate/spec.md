# impeccable-gate Specification

## Purpose

Add an opt-in design-language quality gate to the Archon harness by integrating the
external npm tool [Impeccable](https://github.com/pbakaus/impeccable) via `npx
impeccable`. The gate mirrors the Playwright precedent in shape (config struct, init
flag, TUI tab, status block, CLI get/set, preflight group, phase hooks, judge gate) and
is fully inert when `impeccable.enabled: false` (zero-value default). No Impeccable
source is vendored into the Go binary; Archon only orchestrates when and how to call
it.

---

## Requirement: Config Surface

The `internal/config` package MUST expose an `Impeccable` struct and a corresponding
`Config.Impeccable` field. The zero-value of the struct MUST yield opt-out behavior
(gate is inert). `Clone()` MUST copy the field so `archon update` never silently drops
it. The `TestConfig_CloneRoundtrip` fixture MUST exercise the new struct.

### Config struct shape

The `Impeccable` struct MUST carry these fields:

| Field | Go type | YAML key | Default | Allowed values |
|-------|---------|----------|---------|---------------|
| `Enabled` | `bool` | `enabled` | `false` | `true`, `false` |
| `AutoInstall` | `bool` | `auto_install` | `false` | `true`, `false` |
| `Severity` | `string` | `severity` | `"block-deterministic"` | `"block-deterministic"`, `"block-all"`, `"advisory"` |
| `ProductPath` | `string` | `product_path,omitempty` | `""` (target-project root) | any valid path |
| `DesignPath` | `string` | `design_path,omitempty` | `""` (target-project root) | any valid path |

The `Severity` field controls which Impeccable check categories block the judge gate:
- `block-deterministic` (default): the 58-rule deterministic detector blocks; LLM
  critique results are advisory (report only).
- `block-all`: both deterministic detector and LLM critique block.
- `advisory`: neither blocks; all results are report-only.

### Scenario: Zero-value config is fully inert

```gherkin
@happy
Scenario: Zero-value config is fully inert
  Given ".archon/config.yaml" has no "impeccable" section
  When any Archon phase runs
  Then no Impeccable checks or gate logic executes
  And all phase outputs are identical to a build without this change
```

### Scenario: Enabled config written to disk

```gherkin
@happy
Scenario: Enabled config written to disk
  Given "impeccable.enabled: true" is set in ".archon/config.yaml"
  When "archon status" runs
  Then the Impeccable status block is displayed
  And "severity" defaults to "block-deterministic" when not explicitly set
```

### Scenario: Clone preserves Impeccable fields

```gherkin
@happy
Scenario: Clone preserves Impeccable fields
  Given a Config with Impeccable{Enabled: true, AutoInstall: false, Severity: "block-all",
        ProductPath: "PRODUCT.md", DesignPath: "DESIGN.md"}
  When "Clone()" is called
  Then the returned Config.Impeccable equals the original field-for-field
  And modifying the clone's Impeccable fields does not affect the original
```

### Scenario: CloneRoundtrip fixture covers Impeccable

```gherkin
@happy
Scenario: CloneRoundtrip fixture covers Impeccable
  Given "TestConfig_CloneRoundtrip" sets a non-zero Impeccable struct in the fixture
  When "Clone()" is called and the result is compared field-by-field
  Then no Impeccable field is zero in the clone
  And the test fails if Clone() omits the Impeccable field
```

---

## Requirement: Severity Knob Semantics

The `severity` config field MUST govern which detection output categories cause the
judge gate to return `fail` versus emit an advisory note. The valid values and their
blocking semantics are fixed.

### Scenario: block-deterministic blocks on detector, not LLM critique

```gherkin
@happy
Scenario: block-deterministic severity — detector fails, LLM critique advises
  Given "impeccable.severity: block-deterministic"
  And the Impeccable deterministic detector reports 2 violations
  And the Impeccable LLM critique reports 1 advisory note
  When the judge gate runs
  Then the gate returns "fail" due to the deterministic violations
  And the advisory note is included in the report without blocking
```

### Scenario: block-all blocks on both categories

```gherkin
@edge
Scenario: block-all severity — LLM critique also blocks
  Given "impeccable.severity: block-all"
  And the deterministic detector passes
  And the Impeccable LLM critique reports 1 issue
  When the judge gate runs
  Then the gate returns "fail" due to the LLM critique issue
```

### Scenario: advisory severity never blocks

```gherkin
@edge
Scenario: advisory severity — detector failures do not block
  Given "impeccable.severity: advisory"
  And the deterministic detector reports 3 violations
  When the judge gate runs
  Then the gate returns "pass"
  And all violations are included in the report as advisory notes
```

### Scenario: Invalid severity value is rejected

```gherkin
@error
Scenario: Invalid severity value rejected at config load
  Given ".archon/config.yaml" contains "severity: foobar"
  When "archon status" or any phase command loads the config
  Then Archon exits with a descriptive error naming the invalid value
  And it lists the three valid values: block-deterministic, block-all, advisory
```

---

## Requirement: Init Flag — --impeccable

`archon init` MUST accept a `--impeccable` boolean flag. When the flag is passed,
`buildConfig` MUST set `Impeccable.Enabled: true` in the generated config. The flag
MUST follow the same threading path as `--playwright` and `--security`.

### Scenario: Init with --impeccable sets enabled

```gherkin
@happy
Scenario: Init with --impeccable flag enables the gate
  Given the user runs "archon init --impeccable"
  When init writes ".archon/config.yaml"
  Then "impeccable.enabled: true" is present in the file
  And all other impeccable fields retain their defaults
```

### Scenario: Init without --impeccable leaves gate disabled

```gherkin
@happy
Scenario: Init without --impeccable leaves gate off
  Given the user runs "archon init" with no impeccable flag
  When init writes ".archon/config.yaml"
  Then "impeccable.enabled: false" is the effective value (absent or explicit false)
  And no impeccable phase hook or gate logic activates
```

### Scenario: buildConfig wires the flag value

```gherkin
@happy
Scenario: buildConfig receives the flag value
  Given "impeccableFlag" is set to "true" in "cmd/archon/main.go"
  When "buildConfig" is called
  Then the returned Config contains "Impeccable{Enabled: true}"
```

---

## Requirement: CLI Config Get/Set

`archon config get` and `archon config set` MUST support all Impeccable config keys.
Both the set-value dispatch and the get-value dispatch MUST handle every key. Both
key-list error strings (the "supported keys" error messages in `cmd/archon/config.go`)
MUST be updated to include every new key.

### Supported keys (all five MUST be present in both error strings)

- `impeccable.enabled` — parsed as bool
- `impeccable.auto_install` — parsed as bool
- `impeccable.severity` — string; MUST validate against the three allowed values
- `impeccable.product_path` — string
- `impeccable.design_path` — string

### Scenario: Set impeccable.enabled to true

```gherkin
@happy
Scenario: CLI set impeccable.enabled
  Given ".archon/config.yaml" exists
  When the user runs "archon config set impeccable.enabled true"
  Then ".archon/config.yaml" is updated with "impeccable.enabled: true"
  And "archon config get impeccable.enabled" returns "true"
```

### Scenario: Set impeccable.severity to block-all

```gherkin
@happy
Scenario: CLI set impeccable.severity
  When the user runs "archon config set impeccable.severity block-all"
  Then ".archon/config.yaml" is updated with "impeccable.severity: block-all"
```

### Scenario: Set impeccable.severity to an invalid value

```gherkin
@error
Scenario: CLI rejects invalid severity value
  When the user runs "archon config set impeccable.severity invalid"
  Then the command exits with a non-zero status
  And the error message names the invalid value and lists the three valid options
```

### Scenario: Get unknown key reports supported keys including impeccable keys

```gherkin
@error
Scenario: Get unknown key shows updated supported-key list
  When the user runs "archon config get impeccable.typo"
  Then the error message lists "impeccable.enabled", "impeccable.auto_install",
       "impeccable.severity", "impeccable.product_path", and "impeccable.design_path"
```

---

## Requirement: TUI Impeccable Tab

The TUI MUST include an "Impeccable" tab for managing the impeccable config block
interactively. The tab MUST follow the shape of `playwright_tab.go` and be wired into
`model.go` at all required hook points. The tab enum entry MUST be inserted without
breaking existing tab indices; the `tabs []string` label slice order MUST match the
enum order.

### Tab contents

The Impeccable tab MUST expose these interactive fields:

| Field | TUI element | Bound config key |
|-------|-------------|-----------------|
| Enabled | toggle | `impeccable.enabled` |
| Auto-Install | toggle | `impeccable.auto_install` |
| Severity | text input | `impeccable.severity` |
| Product Path | text input | `impeccable.product_path` |
| Design Path | text input | `impeccable.design_path` |

### model.go wiring requirements

The following hook points in `model.go` MUST be updated for the new tab:

1. Tab enum: one new constant between existing tabs (exact position decided in design);
   `tabCount` sentinel MUST remain the last iota value.
2. `tabs []string` label slice: `"Impeccable"` entry at the position matching the enum.
3. Model struct field: `impeccableTab impeccableTabState`.
4. Constructor: `impeccableTab: newImpeccableTabState(cfg.Impeccable)`.
5. `setWidth` call (resize): `m.impeccableTab.setWidth(w)`.
6. Key routing in `Update`: `case ImpeccableTab: m.impeccableTab = m.impeccableTab.update(msg)`.
7. Reload after save: `m.impeccableTab = newImpeccableTabState(msg.cfg.Impeccable)`.
8. `setWidth` call (after reload): `m.impeccableTab.setWidth(m.width)`.
9. `renderTabContent`: `case ImpeccableTab: return m.impeccableTab.view(...)`.
10. `saveConfig`: `m.impeccableTab.applyToConfig(cfg)`.

### Scenario: TUI tab displays impeccable config

```gherkin
@happy
Scenario: Impeccable tab renders current config
  Given ".archon/config.yaml" contains a non-default impeccable block
  When the user opens "archon tui" and navigates to the Impeccable tab
  Then the tab displays the current values for enabled, auto_install, severity,
       product_path, and design_path
```

### Scenario: Saving TUI writes impeccable config to disk

```gherkin
@happy
Scenario: TUI save persists Impeccable changes
  Given the user opens "archon tui", navigates to Impeccable tab,
        and sets enabled to true and severity to "advisory"
  When the user saves
  Then ".archon/config.yaml" reflects "impeccable.enabled: true" and
       "impeccable.severity: advisory"
```

### Scenario: Tab count and order tests pass after new tab

```gherkin
@happy
Scenario: model_test.go tab count test passes with new tab
  Given "ImpeccableTab" is added to the enum and "tabs" slice
  When "go test ./internal/tui/..." runs
  Then all tab-count and tab-order assertions pass
```

---

## Requirement: archon status Block

`archon status` MUST display an "Impeccable (Design Language)" block parallel to the
existing "Playwright (Web E2E)" block. When `impeccable.enabled: false`, the block
MUST still appear but show the disabled state (no runtime invocation).

### Scenario: Status block when disabled

```gherkin
@happy
Scenario: Status shows impeccable as disabled
  Given "impeccable.enabled: false"
  When "archon status" runs
  Then the output contains an "Impeccable (Design Language)" section
  And it shows "Enabled: false"
  And no paths or severity values are displayed
```

### Scenario: Status block when enabled with non-default values

```gherkin
@happy
Scenario: Status shows impeccable as enabled with config details
  Given "impeccable.enabled: true", "severity: block-all",
        "product_path: PRODUCT.md", "design_path: DESIGN.md"
  When "archon status" runs
  Then the output shows "Enabled: true", "Severity: block-all",
       "Product Path: PRODUCT.md", "Design Path: DESIGN.md"
```

---

## Requirement: Judge Detection Gate

When `impeccable.enabled: true`, the `harness-judge` skill MUST execute a Step 3c
Impeccable Detection Gate after judgment-day passes (parallel to Step 3b Playwright).
The gate MUST invoke `npx impeccable` on the target project and interpret the output
according to the configured `severity`. The gate MUST NOT run if `impeccable.enabled:
false`.

### Gate execution flow

1. Confirm `impeccable.enabled: true` in loaded config.
2. Check that `node` and `npx` are available in the target project's environment.
   - If missing: return `blocked` with an actionable message (see Node-missing
     scenario below). Never silently pass.
3. Run `npx impeccable detect` (or the equivalent detection subcommand) in the target
   project root.
4. Parse stdout/stderr to separate deterministic-detector violations from LLM-critique
   findings.
5. Apply `severity` semantics:
   - `block-deterministic`: deterministic violations → `fail`; LLM critique → advisory.
   - `block-all`: any violation from either category → `fail`.
   - `advisory`: all findings → advisory; gate returns `pass`.
6. Emit the "### Impeccable Gate" output section with status, violation count, and
   advisory-findings summary.
7. Fold the gate result into the overall judge result table.

### Scenario: Gate passes — no violations

```gherkin
@happy
Scenario: Impeccable gate passes when no violations found
  Given "impeccable.enabled: true" and "severity: block-deterministic"
  And "npx impeccable detect" returns exit 0 with no violations
  When the judge gate runs
  Then the "### Impeccable Gate" section shows "Status: pass"
  And the overall judge verdict is not degraded by this gate
```

### Scenario: Gate blocks on deterministic violations (default severity)

```gherkin
@happy
Scenario: Impeccable gate blocks on deterministic violations
  Given "impeccable.enabled: true" and "severity: block-deterministic"
  And "npx impeccable detect" reports 3 deterministic violations
  When the judge gate runs
  Then the "### Impeccable Gate" section shows "Status: fail"
  And the violation count and descriptions are included in the report
  And the overall judge verdict is "fail"
```

### Scenario: Gate is advisory on LLM critique (default severity)

```gherkin
@happy
Scenario: LLM critique is advisory under block-deterministic severity
  Given "impeccable.enabled: true" and "severity: block-deterministic"
  And the deterministic detector reports 0 violations
  And the LLM critique reports 2 advisory notes
  When the judge gate runs
  Then the "### Impeccable Gate" section shows "Status: pass"
  And the advisory notes appear in the report section without blocking
```

### Scenario: Gate skipped when disabled

```gherkin
@happy
Scenario: Judge gate is skipped when impeccable is disabled
  Given "impeccable.enabled: false"
  When "harness-judge" runs
  Then no "npx impeccable" invocation occurs
  And no "### Impeccable Gate" section appears in the output
  And the judge result table has no Impeccable column
```

### Scenario: Gate blocked when Node/npx is missing

```gherkin
@error
Scenario: Judge gate returns blocked when Node/npx absent
  Given "impeccable.enabled: true"
  And "node" or "npx" is not found in the target project environment
  When the judge gate runs
  Then the gate returns "blocked"
  And the output contains an actionable message such as:
      "Impeccable requires Node.js and npx. Install Node.js or set
       impeccable.enabled: false to skip this gate."
  And the overall judge verdict is "blocked"
  And the gate does NOT silently pass
```

### Scenario: Gate output contract

```gherkin
@happy
Scenario: Impeccable gate output section is present when enabled
  Given "impeccable.enabled: true"
  And the gate runs to completion (pass or fail)
  When "harness-judge" completes
  Then the output contains a "### Impeccable Gate" section
  And the section includes at minimum: Status (pass/fail/blocked),
      deterministic violation count, and advisory-findings count
```

---

## Requirement: Design-Phase Minimal Reference

When `impeccable.enabled: true`, the `sdd-design` skill MUST reference the target
project's `PRODUCT.md` and `DESIGN.md` (generated by `impeccable init`) as design
input constraints when they exist. The skill MUST NOT run the Impeccable detector,
MUST NOT generate or overwrite SDD `design.md`, and MUST NOT generate any Impeccable
output files during the design phase. This hook is purely advisory input.

### File ownership rules (non-negotiable)

- `PRODUCT.md` and `DESIGN.md` live at the **target-project root** (Impeccable
  default). Archon does not own or rewrite them.
- The SDD design artifact lives at `openspec/changes/<change>/design.md`. These two
  paths are distinct; neither overwrites the other.

### Scenario: Design phase references Impeccable docs when enabled and present

```gherkin
@happy
Scenario: Design references PRODUCT.md and DESIGN.md when impeccable is enabled
  Given "impeccable.enabled: true"
  And "PRODUCT.md" and "DESIGN.md" exist at the target-project root
  When "sdd-design" runs
  Then the design phase reads PRODUCT.md and DESIGN.md as input context
  And design constraints derived from those files are reflected in the design artifact
  And the design skill does not invoke "npx impeccable" or any detection commands
  And "openspec/changes/<change>/design.md" is the only SDD design file written
```

### Scenario: Design phase when Impeccable docs absent

```gherkin
@edge
Scenario: Design continues normally when PRODUCT.md/DESIGN.md are missing
  Given "impeccable.enabled: true"
  And neither "PRODUCT.md" nor "DESIGN.md" exists at the target-project root
  When "sdd-design" runs
  Then the design phase proceeds without error
  And a note in the design artifact recommends running "impeccable init" to generate
      the design-language foundation documents
```

### Scenario: Design phase when impeccable is disabled

```gherkin
@happy
Scenario: Design phase is unchanged when impeccable is disabled
  Given "impeccable.enabled: false"
  When "sdd-design" runs
  Then the phase behaves identically to today (no impeccable logic executes)
```

---

## Requirement: Other Phase Hooks (Apply, Verify, Tasks, Explore, Spec)

Impeccable MUST be woven into the remaining phase skills in a flag-gated, additive
manner. Each hook is inert when `impeccable.enabled: false`.

### Apply phase

The `sdd-apply` skill MUST include an Impeccable step (parallel to Step 4b for
Playwright) when `impeccable.enabled: true`. This step invokes the relevant Impeccable
subcommands (craft/polish/harden/animate) on frontend-affecting changes. A
corresponding Rules line MUST be added to the skill.

### Verify phase

The `sdd-verify` skill MUST confirm, when `impeccable.enabled: true`, that any
required Impeccable artifacts or hooks are present (advisory check). It MUST NOT run
the detection gate (that is the judge's responsibility). Missing artifacts MUST be
reported as a note, not a CRITICAL blocker.

### Tasks phase

The `sdd-tasks` skill MUST add an "Impeccable pass" task when `impeccable.enabled:
true` and the change affects frontend files. This task instructs the apply phase to
run the relevant Impeccable subcommands.

### Explore phase

The `sdd-explore` skill MUST suggest enabling Impeccable (preflight group F) when the
target project is detected as a web/frontend project. This is a recommendation, not an
automatic activation.

### Spec phase

The `sdd-spec` skill MAY flag requirements that apply to frontend design-language
quality so they can be selected by the Impeccable gate. This is a lightweight
annotation (analogous to `@web` for Playwright). The exact annotation form (tag vs
prose) is left to the design phase.

### Scenario: Apply runs Impeccable subcommands when enabled

```gherkin
@happy
Scenario: Apply step invokes Impeccable on frontend changes when enabled
  Given "impeccable.enabled: true"
  And the change includes modifications to frontend files
  When "sdd-apply" runs
  Then the Impeccable step runs the appropriate subcommands on the changed frontend files
  And a Rules line in the skill enforces this behavior
```

### Scenario: Verify notes missing Impeccable artifacts

```gherkin
@edge
Scenario: Verify reports missing Impeccable artifacts as advisory
  Given "impeccable.enabled: true"
  And no Impeccable hooks or configuration exist in the target project
  When "sdd-verify" runs
  Then the verification report includes a note recommending Impeccable setup
  And the phase does NOT block or fail due to missing Impeccable artifacts
```

---

## Requirement: Preflight Group F and Template Sync

The orchestrator template MUST include a preflight group F ("¿Aplicar Impeccable…?")
for the Impeccable design-language gate, following the same always-emitted structure
as group E (Playwright). The `orchestratorRulesClaude` and `orchestratorRulesOpencode`
string consts in `templates.go` MUST both gain a new rule line for Impeccable. The
repo's `CLAUDE.md` MUST be updated to mirror the template output. Template assertions
in `templates_test.go` MUST cover both group F and the new rule.

### Group F text (required content — must appear verbatim or equivalent)

```
- **F. Impeccable (Diseño de interfaz)** — "¿Activar Impeccable para calidad visual?"
  - No (recomendado): no correr verificaciones de diseño.
  - Sí: activar el gate de Impeccable tras verify/judge cuando esté habilitado.
```

### Group F mapping paragraph (required)

```
**Project type & design-language gate (group F):**
- Group F maps to `impeccable.enabled` in `.archon/config.yaml`. The `--impeccable`
  flag at init time or the Impeccable tab in `archon tui` set the same value. When
  enabled, the harness invokes Impeccable subcommands during apply and runs the
  detection gate after the judge phase.
```

### New orchestrator rule (required in both Claude and Opencode rule consts)

```
When impeccable.enabled, run Impeccable subcommands during apply and the detection
gate after judge passes.
```

### Scenario: Template emits group F

```gherkin
@happy
Scenario: Generated CLAUDE.md includes preflight group F
  Given a project initialized with "archon init"
  When the generated "CLAUDE.md" is read
  Then it contains the group F preflight question for Impeccable
  And it contains the group F mapping paragraph
```

### Scenario: Template emits the Impeccable rule

```gherkin
@happy
Scenario: Generated CLAUDE.md includes the Impeccable rule
  Given a project initialized with "archon init"
  When the generated "CLAUDE.md" is read
  Then it contains a rule line for "impeccable.enabled"
  And this rule appears in both the Claude and Opencode variants
```

### Scenario: templates_test.go asserts group F and new rule

```gherkin
@happy
Scenario: Template tests cover group F and the Impeccable rule
  Given "templates_test.go" has assertions for group F text and the Impeccable rule
  When "go test ./internal/initcmd/..." runs
  Then all template assertions pass
  And a test failure catches any omission of group F or the rule line
```

### Scenario: Repo CLAUDE.md matches template output

```gherkin
@happy
Scenario: Repo CLAUDE.md is consistent with the template
  Given the repo root "CLAUDE.md" is updated to mirror templates.go
  Then group F, the group F mapping paragraph, and the Impeccable rule are present
  And the repo CLAUDE.md is consistent with what "archon init" would generate
```

---

## Requirement: Thin Impeccable Skill

A new file `skills/impeccable/SKILL.md` MUST be created as the single orchestration
skill for the Impeccable gate. It MUST describe WHEN and HOW Archon calls Impeccable
without reimplementing the 58-rule detector or the LLM critique. The file is
auto-embedded by `skills/embed.go` and auto-inventoried by `scaffold/version.go`,
bumping `skill_count` from 24 to 25. No manual edits to embed.go or version.go are
required.

### Skill responsibilities (what SKILL.md must document)

- Which phase each Impeccable subcommand is called from and under what condition.
- The detect invocation signature used by the judge gate.
- How to interpret exit codes and output to determine pass/fail/advisory.
- How to handle Node/npx missing (fail with actionable message).
- Reference to `PRODUCT.md` / `DESIGN.md` ownership and location.

### Scenario: Skill is auto-embedded and inventoried

```gherkin
@happy
Scenario: New impeccable skill is embedded and counted automatically
  Given "skills/impeccable/SKILL.md" exists
  When "archon init" or "archon update" runs
  Then the skill is included in the embedded skill set
  And "skill_count" in the generated config reflects 25 skills
  And "skill_inventory" includes an "impeccable" entry
```

### Scenario: Skill does not reimplement detector

```gherkin
@happy
Scenario: Skill delegates detection to npx impeccable, not Go code
  Given "skills/impeccable/SKILL.md"
  When the skill instructions are followed
  Then the only invocation mechanism is "npx impeccable <subcommand>"
  And no Go code or skill logic reimplements Impeccable's rules
```

---

## Requirement: auto_install Behavior

When `impeccable.auto_install: false` (default), Archon MUST assume Impeccable is
already installed and MUST NOT run any install commands. When `npx impeccable` fails
because the package is not found, Archon MUST instruct the user to install it rather
than installing silently.

When `impeccable.auto_install: true`, Archon MAY run `npx impeccable install` (or
equivalent) to set up the Impeccable hooks in the target project before the first
gate invocation. This is still a one-time setup action, not a repeated install.

### Scenario: auto_install false — missing package instructs rather than installs

```gherkin
@error
Scenario: npx impeccable not found with auto_install false
  Given "impeccable.enabled: true" and "impeccable.auto_install: false"
  And "npx impeccable" fails because the package is not installed
  When the judge gate attempts to run Impeccable
  Then the gate returns "blocked"
  And the output contains an instruction to install Impeccable
      (e.g., "Run 'npm install -g impeccable' or add it to your project devDependencies")
  And no silent install occurs
```

### Scenario: auto_install true — install runs before first gate invocation

```gherkin
@edge
Scenario: auto_install true triggers install before gate
  Given "impeccable.enabled: true" and "impeccable.auto_install: true"
  And Impeccable is not installed in the target project
  When the judge gate runs for the first time
  Then "npx impeccable install" (or equivalent) runs first
  And the detection gate runs after install completes
```

---

## Non-Goals (Explicit)

These are out of scope for this change and MUST NOT be implemented:

1. **No vendoring**: Impeccable source code or its npm package MUST NOT be compiled
   into or bundled with the Archon Go binary.
2. **No detector reimplementation**: The 58-rule deterministic detector and the LLM
   critique logic MUST remain entirely within the `npx impeccable` tool. Archon only
   interprets its output.
3. **No generic gate abstraction**: A shared "web quality gates" struct or interface
   that subsumes Playwright and Impeccable MUST NOT be introduced. Each gate keeps its
   own config struct (one-struct-per-gate convention).
4. **No `archon update` config migration**: `impeccable` config fields ride through
   `archon update` via `Clone()` automatically. No migration code is needed.
5. **No Playwright behavior change**: All Playwright gate logic, config, and TUI remain
   untouched when `impeccable.enabled: false` — and untouched in general. The
   Impeccable gate is strictly additive.

---

## Risks and Open Questions

| # | Risk / Question | Likelihood | Mitigation / Note |
|---|-----------------|------------|-------------------|
| 1 | Impeccable's detection output format changes between versions | Med | Parse by documented exit codes; treat unknown output as advisory, not blocker |
| 2 | LLM critique output is non-deterministic (flaky with block-all severity) | High | Default severity is `block-deterministic`; user must opt into `block-all` consciously |
| 3 | Tab enum insertion position conflicts with future tabs | Low | Exact position decided in design phase; spec requires enum/slice synchrony, not specific index |
| 4 | `impeccable init` PRODUCT.md/DESIGN.md names clash with other tools | Low | Both live at target root; the names are Impeccable defaults; out of Archon's control |
| 5 | Design-phase hook over-scope (no Playwright precedent) | Med | Spec limits it to read-only reference; no detection or file generation in design |
| 6 | PR2 prose (skill + hooks + template) exceeds 400-line budget | High | Pre-planned PR3 split (judge gate + template); confirm at ask-always gate |
| 7 | `npx impeccable detect` subcommand name may differ in Impeccable API | Med | Design phase must verify against Impeccable's actual CLI; this spec uses the canonical name |
