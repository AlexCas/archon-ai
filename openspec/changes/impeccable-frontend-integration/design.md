# Design: Impeccable Frontend-Integration (design-language gate)

Implements [[impeccable-gate]]. Builds on [proposal](proposal.md), [exploration](exploration.md),
and [spec](specs/impeccable-gate/spec.md). This document resolves the spec's two open
questions (real Impeccable CLI; TUI enum insertion position) and specifies every edit
site for apply.

---

## 0. Verified Impeccable CLI facts (resolves spec open question #7)

Fetched from `github.com/pbakaus/impeccable` README on 2026-07-24. **The spec's
assumptions were partly wrong** — Impeccable has TWO invocation surfaces, and most of
its commands are agent slash-commands, not `npx` CLI commands.

| Surface | Commands | Invocation | Notes |
|---------|----------|-----------|-------|
| **npx CLI** | `install`, `update`, `detect` | `npx impeccable <cmd>` | These are the only true shell commands. |
| **Agent slash-commands** | `init` + 23 design verbs (`craft`, `shape`, `critique`, `audit`, `polish`, `bolder`, `quieter`, `distill`, `harden`, `onboard`, `animate`, `colorize`, `typeset`, `layout`, `delight`, `overdrive`, `clarify`, `adapt`, `optimize`, `live`, `document`, `extract`, `pin`) | `/impeccable <cmd>` inside the AI coding tool | Run by the agent, NOT via `npx`. |

Confirmed exact invocations:

- **Detection (judge gate):** `npx impeccable detect <target>` — targets can be a dir
  (`src/`), a file (`index.html`), or a URL (`https://...`). Flags: `--json`
  (CI-friendly machine output — the gate MUST use this), `--no-config` (raw scan,
  ignore project context), `--no-inline-ignores` (bypass inline waivers). The judge gate
  uses **`npx impeccable detect --json .`** from the target-project root.
- **Install (auto_install):** `npx impeccable install` (from project root), refresh via
  `npx impeccable update`.
- **Init (design-phase docs):** **`/impeccable init`** — a **slash command run inside
  the AI coding tool**, NOT `npx impeccable init`. It writes `PRODUCT.md` and offers
  `DESIGN.md` at the project root (audience, brand lane, voice, anti-references, colors,
  type, components). Subsequent commands read them.
- **Apply-phase design verbs:** the craft/polish/harden/animate family are
  **`/impeccable <verb>` slash commands run by the agent**, NOT `npx` commands.

### Corrections vs the spec (must be reflected in artifacts)

1. **`init` is not a CLI command.** Spec Decision 3, the design-phase requirement, and
   the skill said "`impeccable init` produces PRODUCT.md/DESIGN.md." The real form is the
   slash command `/impeccable init`. Design-phase and skill copy must say
   `/impeccable init`, and any "run `impeccable init`" recommendation in the design
   artifact (edge scenario) becomes "run `/impeccable init` in your AI coding tool."
2. **Apply-phase verbs are slash commands.** The apply hook and skill must instruct the
   agent to run `/impeccable <verb>` (agent-driven), not shell out to `npx impeccable craft`.
   This does NOT change the flag-gating or the Rules line — only the invocation string.
3. **Exit code on violations is undocumented.** The README does not state that `detect`
   exits non-zero on violations. The gate therefore MUST NOT rely on exit code alone for
   pass/fail: it parses `--json` output. Exit code is used only to distinguish
   *ran-successfully* from *tool-missing/crashed* (see gate flow). This hardens spec
   Risk #1 (output-format drift).
4. **`detect` still matches the spec's judge-gate assumption** (`npx impeccable detect`),
   so the gate's core invocation in the spec is correct; only add `--json .`.

No spec requirement is invalidated — the gate, config surface, flag, TUI, and status
all stand. Only the invocation strings for `init` and the design verbs are corrected
from "npx CLI" to "agent slash-command," and the gate is made exit-code-tolerant.

---

## 1. Architecture and data flow

`impeccable.enabled` (plus `auto_install`, `severity`, `product_path`, `design_path`)
is a config-gated feature, inert at zero value. The flow mirrors Playwright:

```
[config source]                     [runtime consumer]
--impeccable flag ─┐
archon tui tab ────┼──► .archon/config.yaml ──► config.Load() ──┬─► archon status (display block)
config set ────────┘        (impeccable:)                       │
                                                                ├─► sdd-design  (read PRODUCT.md/DESIGN.md; no detect)
                                                                ├─► sdd-tasks   (add "Impeccable pass" task)
                                                                ├─► sdd-apply   (/impeccable <verb> on frontend changes)
                                                                ├─► sdd-verify  (advisory presence note)
                                                                └─► harness-judge Step 3c
                                                                       └─► npx impeccable detect --json .
                                                                            └─► severity → pass | fail | (advisory) | blocked
```

- **Config → init:** `--impeccable` sets `Impeccable.Enabled: true` via `buildConfig`.
- **Config → TUI:** the Impeccable tab reads `cfg.Impeccable` and writes it back on save.
- **Config → status:** `archon status` always prints the block; contents depend on `Enabled`.
- **Config → phase skills:** each SDD phase skill reads `impeccable.enabled` and branches.
  Everything is a no-op when false (zero value).
- **Judge gate:** the only place that *executes* Impeccable (`npx impeccable detect`).
  `severity` decides how detector vs LLM-critique findings map to the verdict.
- **auto_install:** consulted only by the judge gate (and optionally apply) before the
  first invocation.

`archon update` needs no code changes for the config block itself: `Update()` clones the
loaded config and patches only Version/SkillCount/SkillInventory, so `Impeccable` rides
through **provided it is in `Clone()`**. Skill embedding is a glob, so
`skills/impeccable/SKILL.md` is auto-embedded and bumps `skill_count` 24→25 with no
`embed.go`/`version.go` edits.

---

## 2. Go config design — `internal/config/config.go`

### 2.1 `Impeccable` struct (insert after the `Security` struct, ~line 44)

```go
// Impeccable controls the opt-in design-language quality gate backed by the
// external npm tool `npx impeccable`. When enabled, the harness references the
// target project's Impeccable design docs during design, runs Impeccable design
// verbs during apply, and executes the `npx impeccable detect` gate after the
// judge phase. Severity governs which finding categories block the judge gate.
// Defaults to disabled (Enabled:false) when the block is absent.
type Impeccable struct {
	Enabled     bool   `yaml:"enabled"`
	AutoInstall bool   `yaml:"auto_install"`
	Severity    string `yaml:"severity,omitempty"`
	ProductPath string `yaml:"product_path,omitempty"`
	DesignPath  string `yaml:"design_path,omitempty"`
}
```

Design notes:
- `Enabled` and `AutoInstall` have **no `omitempty`** — they mirror `Playwright.Enabled`
  / `Security.Enabled` so `enabled: false` / `auto_install: false` render explicitly
  (matches the on-disk `playwright:\n  enabled: false` precedent and the status/spec
  expectation that a disabled block is still visible on disk).
- `Severity` uses `omitempty`: the spec default is `"block-deterministic"`, but writing
  it always would clutter the file. **The default is resolved at read time**, not stored
  (see 2.2). `omitempty` keeps a defaulted config clean.
- `ProductPath` / `DesignPath` use `omitempty` per the spec table (empty = target-project
  root, Impeccable's own default).

### 2.2 Severity default + validation in `Load()` (~line 79, beside `c.Judge.Enabled = true`)

The spec requires (a) `severity` defaults to `block-deterministic` when unset and (b)
config load REJECTS an invalid severity ("Invalid severity value rejected at config
load"). The current `security.profile` precedent only validates at CLI set-time, so this
is a small net-new behavior at `Load()`:

```go
// Impeccable severity defaults to the safe "block-deterministic" mode; an absent
// or empty value is normalized here so downstream consumers never see "".
if c.Impeccable.Severity == "" {
	c.Impeccable.Severity = "block-deterministic"
}
```

Add, after unmarshal succeeds, a validation guard:

```go
if err := validateImpeccableSeverity(c.Impeccable.Severity); err != nil {
	return fmt.Errorf("config: %w", err)
}
```

with a shared helper (exported for reuse by the CLI set path — DRY, single source of the
three valid values):

```go
// ValidImpeccableSeverities is the fixed set of allowed severity values.
var ValidImpeccableSeverities = []string{"block-deterministic", "block-all", "advisory"}

func validateImpeccableSeverity(s string) error {
	switch s {
	case "block-deterministic", "block-all", "advisory":
		return nil
	default:
		return fmt.Errorf("invalid impeccable.severity %q (valid: block-deterministic, block-all, advisory)", s)
	}
}
```

Order in `Load()`: pre-seed `Judge.Enabled = true` → unmarshal → normalize empty
severity → validate severity. (Normalize BEFORE validate so an absent block is not
rejected as `""`.)

### 2.3 `Config.Impeccable` field (after `Security`, ~line 60)

```go
	Security        Security         `yaml:"security"`
	Impeccable      Impeccable       `yaml:"impeccable"`
```

### 2.4 `Clone()` addition (after `Security: c.Security,`, ~line 106)

```go
		Security:        c.Security,
		Impeccable:      c.Impeccable,   // value copy — no maps/slices inside
```

`Impeccable` is all scalars, so a value copy is a full deep copy. No extra loop needed.

### 2.5 `CloneRoundtrip` fixture — `internal/config/config_test.go` (~line 227, after `Security:`)

```go
		Security: Security{
			Enabled: true,
			Profile: "web",
		},
		Impeccable: Impeccable{
			Enabled:     true,
			AutoInstall: true,
			Severity:    "block-all",
			ProductPath: "PRODUCT.md",
			DesignPath:  "DESIGN.md",
		},
```

All five fields non-zero → the existing `reflect.DeepEqual` check fails loudly if
`Clone()` omits the field. Add a parallel `TestImpeccable_DefaultsAndValidation` test
(new) asserting: absent block → `Severity == "block-deterministic"`, `Enabled == false`;
and a `severity: foobar` fixture → `Load()` returns an error naming the value + three
valid options (spec "Invalid severity value rejected at config load").

---

## 3. CLI get/set design — `cmd/archon/config.go`

### 3.1 `setConfigValue` (add cases after `security.profile`, ~line 189)

```go
	case "impeccable.enabled":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.Impeccable.Enabled = b
		return nil
	case "impeccable.auto_install":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.Impeccable.AutoInstall = b
		return nil
	case "impeccable.severity":
		if err := config.ValidateImpeccableSeverity(value); err != nil {
			return err   // names the invalid value + the three valid options
		}
		cfg.Impeccable.Severity = value
		return nil
	case "impeccable.product_path":
		cfg.Impeccable.ProductPath = value
		return nil
	case "impeccable.design_path":
		cfg.Impeccable.DesignPath = value
		return nil
```

`ValidateImpeccableSeverity` is the exported form of the `Load()` helper (2.2) — one
source of truth for both the load path and the CLI set path (mirrors how the spec's two
"invalid severity" scenarios, one at load, one at CLI set, must give the same message).

### 3.2 `getConfigValue` (add cases after `security.profile`, ~line 221)

```go
	case "impeccable.enabled":
		return strconv.FormatBool(cfg.Impeccable.Enabled), nil
	case "impeccable.auto_install":
		return strconv.FormatBool(cfg.Impeccable.AutoInstall), nil
	case "impeccable.severity":
		return cfg.Impeccable.Severity, nil
	case "impeccable.product_path":
		return cfg.Impeccable.ProductPath, nil
	case "impeccable.design_path":
		return cfg.Impeccable.DesignPath, nil
```

### 3.3 Both key-list error strings (lines 202 and 230)

Both currently end with `..., security.enabled, security.profile)`. Append all five
impeccable keys to **both** (spec: "all five MUST be present in both error strings"):

```
..., security.enabled, security.profile, impeccable.enabled, impeccable.auto_install, impeccable.severity, impeccable.product_path, impeccable.design_path)
```

These two strings must stay byte-identical to each other. A CLI-level test (extend the
existing config get/set test file) asserts the unknown-key error contains all five
impeccable keys.

---

## 4. Init flag `--impeccable` threading

Mirror `--playwright`/`--security` exactly (three files):

**`cmd/archon/main.go`:**
- Var: `var impeccableFlag bool` (beside `securityFlag`, ~line 83).
- Pass into Options: `Impeccable: impeccableFlag` in the `initcmd.Options{...}` literal (~line 170).
- Register: `cmd.Flags().BoolVar(&impeccableFlag, "impeccable", false, "Enable the Impeccable design-language gate")` (~line 202).

**`internal/initcmd/init.go`:**
- `Options.Impeccable bool` (beside `Security`, ~line 27).
- Thread into the `buildConfig(...)` call (~line 87): add `opts.Impeccable` argument.
- `buildConfig` signature (~line 220): add `impeccable bool` parameter (placed after the
  `security` param, keeping positional order == Options field order).
- In the returned `config.Config{...}` (~line 242): add
  `Impeccable: config.Impeccable{Enabled: impeccable}`. Do NOT set `Severity` here —
  `Load()` normalizes the empty value to the default on the next read, matching the spec
  scenario "all other impeccable fields retain their defaults."

**README.md / AGENTS.md:** document `--impeccable` beside `--playwright`/`--security`
(one line each; low-risk doc).

---

## 5. TUI tab — concrete design (resolves spec open question / TUI insertion decision)

### 5.1 Insertion decision

**Insert `ImpeccableTab` immediately after `SecurityTab`, as the last real tab before
`tabCount`.** Rationale:

- Impeccable is the newest gate; appending it after Security keeps every existing tab
  index (`AgentTab..SecurityTab`) unchanged, minimizing test churn and the "index shift"
  risk (proposal Risk row). No existing index-based assertion moves.
- `SecurityTab` stays the value it is today; `ImpeccableTab` takes the slot currently
  held by `tabCount`, and `tabCount` shifts up by one — the standard "append before
  sentinel" pattern already used when Security was added.

Resulting enum (model.go lines 21-29):

```go
const (
	AgentTab Tab = iota
	ModelsTab
	JudgeTab
	MutationTab
	PlaywrightTab
	SecurityTab
	ImpeccableTab   // NEW — appended after Security
	tabCount        // sentinel stays last
)
```

**Test consequence — the ONE existing test that must change:** the `ShiftTab` wrap test
`TestModel_Update_ShiftTabWrapsFromAgent` (model_test.go lines 109-123) asserts that
one Shift+Tab from `AgentTab` lands on **`SecurityTab`** (the current last tab). After
insertion the last tab is `ImpeccableTab`, so line 120-122 must change
`SecurityTab` → `ImpeccableTab`. This is the single behavioral test coupled to
"which tab is last." The order test `TestModel_renderTabs_Order` (lines 216-233) checks
a *prefix* of labels ending at "Playwright", so it still passes unchanged, but we will
**append `"Impeccable"` to its `labels` slice** (line 221) to lock the new order.

### 5.2 `internal/tui/impeccable_tab.go` (NEW — clone of playwright_tab.go)

Five interactive fields (spec tab-contents table): 2 toggles + 3 text inputs. Focus
model: `focused int`, `impeccableFocusCount = 5`.

```go
type impeccableTabState struct {
	enabled     bool             // focus 0 — toggle
	autoInstall bool             // focus 1 — toggle
	severity    textinput.Model  // focus 2 — text input
	productPath textinput.Model  // focus 3 — text input
	designPath  textinput.Model  // focus 4 — text input
	focused     int
}

const impeccableFocusCount = 5
```

- `newImpeccableTabState(cfg config.Impeccable)`: seed `enabled`/`autoInstall` from cfg;
  `severity` textinput placeholder `"block-deterministic (default)"`, value `cfg.Severity`;
  `productPath` placeholder `"PRODUCT.md (default: project root)"`, value `cfg.ProductPath`;
  `designPath` placeholder `"DESIGN.md (default: project root)"`, value `cfg.DesignPath`.
- `update`: up/down cycle focus over `impeccableFocusCount`; `Enter`/`Space` toggles when
  `focused == 0` (enabled) or `focused == 1` (autoInstall); otherwise forward keystrokes
  to the focused textinput (severity=2, productPath=3, designPath=4). Same shape as
  playwright's `update`, extended for the second toggle.
- `refocus`: blur all three inputs; focus the one matching `focused` (2/3/4).
- `view`: title `"Impeccable (Design Language) Configuration"`; two toggle lines
  (`[ON|OFF] Enabled`, `[ON|OFF] Auto-Install`); three labeled inputs
  (`Severity:`, `Product path:`, `Design path:`); info footer:
  `"When enabled, runs 'npx impeccable detect' after judge; severity governs blocking."`
- `applyToConfig(cfg)`:
  ```go
  cfg.Impeccable.Enabled = p.enabled
  cfg.Impeccable.AutoInstall = p.autoInstall
  cfg.Impeccable.Severity = strings.TrimSpace(p.severity.Value())
  cfg.Impeccable.ProductPath = strings.TrimSpace(p.productPath.Value())
  cfg.Impeccable.DesignPath = strings.TrimSpace(p.designPath.Value())
  ```
  Note: an empty severity written back is normalized to the default by `Load()` on the
  next read, and `Save()`→`Load()` round-trips safely. (If we want the TUI-saved file to
  never carry an invalid value, `applyToConfig` may fall back to `"block-deterministic"`
  when the trimmed value is empty; decided at apply, low-risk either way since `Load()`
  normalizes.)
- `setWidth(width)`: same body as playwright but set width on all three inputs.

### 5.3 `internal/tui/model.go` lockstep edit list (10 sites + enum)

Every site below MUST move together, in this order:

| # | Site | Line (current) | Edit |
|---|------|---------------|------|
| 1 | Tab enum | 21-29 | Add `ImpeccableTab` after `SecurityTab`, before `tabCount`. |
| 2 | Model struct field | 47-49 | Add `impeccableTab impeccableTabState` after `securityTab`. |
| 3 | Constructor `NewModel` | 109-110 | Add `impeccableTab: newImpeccableTabState(cfg.Impeccable),`. |
| 4 | `setWidth` on resize | 129-130 | Add `m.impeccableTab.setWidth(m.width)` after `m.securityTab.setWidth`. |
| 5 | Key routing in `Update` | 174-178 | Add `case ImpeccableTab: cmd, _ := m.impeccableTab.update(msg); if cmd != nil { cmds = append(cmds, cmd) }` after the `SecurityTab` case. |
| 6 | Reload after `agentInitDoneMsg` | 206 | Add `m.impeccableTab = newImpeccableTabState(msg.cfg.Impeccable)` after `m.securityTab = ...`. |
| 7 | `setWidth` after reload | 213 | Add `m.impeccableTab.setWidth(m.width)` inside the `if m.width > 0` block. |
| 8 | `renderTabs` label slice | 274 | Append `"Impeccable"` to `tabs []string` (order-coupled to enum: `..., "Security", "Impeccable"`). |
| 9 | `renderTabContent` | 303-304 | Add `case ImpeccableTab: return style.Render(m.impeccableTab.view(m.width, m.height))` after the `SecurityTab` case. |
| 10 | `saveConfig` | 342-343 | Add `m.impeccableTab.applyToConfig(cfg)` after `m.securityTab.applyToConfig(cfg)`. |

### 5.4 `internal/tui/model_test.go` edits

- Line 120-122: `TestModel_Update_ShiftTabWrapsFromAgent` — change expected wrap target
  `SecurityTab` → `ImpeccableTab` (the new last tab). This is REQUIRED; the test fails
  otherwise.
- Line 221: `TestModel_renderTabs_Order` — append `"Impeccable"` to the `labels` slice so
  the new tab's order is asserted.
- Add a `TestImpeccableTabState_ApplyToConfig` (new, parallel to
  `TestMutationTabState_ApplyToConfig`) driving toggles + text inputs and asserting the
  five `cfg.Impeccable.*` fields (spec "TUI save persists Impeccable changes"). Include a
  toggle test for both `enabled` and `autoInstall`.

No other model_test.go assertions are index-coupled (they navigate relatively or check
prefixes), so nothing else moves.

---

## 6. Judge detection gate — `skills/harness-judge/SKILL.md` (Step 3c)

Add a **Step 3c: Impeccable Detection Gate (conditional)**, parallel to Step 3b
(Playwright), running only after judgment-day passes AND `impeccable.enabled: true`.

### 6.1 Gate flow

```
1. Read config. If impeccable.enabled != true → skip entirely
   (no invocation, no "### Impeccable Gate" section, no result-table column).
2. Check `node` and `npx` are on PATH in the target project.
   - Missing → return `blocked` with the actionable message (6.3). Never silent-pass.
3. If auto_install == true AND impeccable is not yet installed → run
   `npx impeccable install` once, then continue. If auto_install == false and the
   package is missing (npx reports not-found) → return `blocked` with the install
   instruction (6.3, package-missing variant). No silent install.
4. Run `npx impeccable detect --json .` from the target-project root.
5. Interpret results (do NOT rely on exit code for pass/fail — see CLI fact #3):
   - Parse the --json payload into deterministic-detector violations vs
     LLM-critique findings.
   - Exit code is used only to detect tool crash / not-found → `blocked`.
   - Unrecognized/unparseable JSON → treat findings as advisory, note the parse
     failure, do NOT hard-fail (spec Risk #1).
6. Apply severity:
   - block-deterministic (default): deterministic violations > 0 → `fail`;
     LLM critique → advisory (reported, non-blocking).
   - block-all: any finding from either category → `fail`.
   - advisory: all findings advisory; gate returns `pass`.
7. Emit the "### Impeccable Gate" section (6.2).
8. Fold the gate status into the overall judge result table as a new column;
   `fail`/`blocked` degrade the overall verdict exactly like the Playwright column.
```

### 6.2 Output section format (spec "Gate output contract")

```
### Impeccable Gate
- Status: pass | fail | blocked
- Severity mode: block-deterministic | block-all | advisory
- Deterministic violations: <n>
- Advisory findings (LLM critique): <n>
- Details:
  - <rule id / description per violation, when n > 0>
- (blocked only) Reason: <node/npx missing | package not installed | detect crashed>
```

Minimum required fields per spec: Status, deterministic violation count, advisory count.

### 6.3 Blocked messages (verbatim intent)

- Node/npx missing:
  `Impeccable requires Node.js and npx. Install Node.js or set impeccable.enabled: false to skip this gate.`
- Package missing, `auto_install: false`:
  `Impeccable is not installed. Run 'npx impeccable install' (or set impeccable.auto_install: true), or set impeccable.enabled: false to skip this gate.`

### 6.4 Other harness-judge edits (mirror the Playwright touchpoints)

- Intro / config-read rules (skill lines ~13, 23, 28): add impeccable alongside playwright.
- Config snippet block (~lines 67-70): add the `impeccable:` example.
- Step 4 result table (~lines 122-127): add an Impeccable column.
- Edge cases (~lines 243-246): add "Impeccable: node/npx missing → blocked",
  "package not installed + auto_install:false → blocked", "config absent → default off".
- Output contract (~lines 215-217): add the "### Impeccable Gate" section.

---

## 7. Design-phase minimal reference — `skills/sdd-design/SKILL.md`

Net-new hook (Playwright has none). Strictly read-only:

- When `impeccable.enabled: true`, before drafting design.md, the skill reads
  `PRODUCT.md` and `DESIGN.md` from the target-project root (or `product_path`/
  `design_path` if set) **if they exist**, and folds their design-language constraints
  (audience, brand lane, voice, colors, type, components) into the design artifact as
  input context.
- The hook MUST NOT run `npx impeccable detect`, MUST NOT run any slash command, MUST NOT
  generate/overwrite SDD `design.md`, and MUST NOT create any Impeccable output file.
  The SDD artifact stays `openspec/changes/<change>/design.md`; the Impeccable docs stay
  at the target root. Distinct paths, neither overwrites the other (spec file-ownership
  rules).
- If the docs are absent: proceed normally and add a note recommending the user run
  **`/impeccable init` in their AI coding tool** (corrected from "impeccable init") to
  generate the design-language foundation docs.
- When `impeccable.enabled: false`: identical to today (no impeccable logic).

---

## 8. Thin skill — `skills/impeccable/SKILL.md` (NEW)

Single orchestration skill; does NOT reimplement the 58-rule detector or LLM critique.
Auto-embedded via the `skills/embed.go` glob and auto-inventoried by
`scaffold/version.go`, bumping `skill_count` 24→25 (no embed.go/version.go edits;
the `{{.SkillCount}}` template value renders 25 automatically).

Documents (spec "skill responsibilities"):
- **Per-phase invocation map:**
  - design → read `PRODUCT.md`/`DESIGN.md` (read-only, no command).
  - apply → run `/impeccable <verb>` (craft/polish/harden/animate) on frontend-affecting
    changes (slash command, agent-driven — corrected from `npx impeccable <verb>`).
  - verify → advisory presence note only.
  - judge → the detection gate: `npx impeccable detect --json .`.
- **Detect invocation signature** used by the judge gate: `npx impeccable detect --json .`
  from target root; `--no-config`/`--no-inline-ignores` documented as available knobs.
- **Exit-code / output interpretation:** parse `--json`; exit code only distinguishes
  ran vs crashed/not-found; severity maps categories to pass/fail/advisory.
- **Node/npx missing → blocked** with the actionable message (6.3).
- **auto_install semantics:** false = assume installed, instruct on missing; true = one-time
  `npx impeccable install` before first gate run.
- **PRODUCT.md/DESIGN.md ownership:** target-project root, Impeccable-owned, generated by
  `/impeccable init`; never overwrites SDD design.md.
- **Two-surface warning:** `install`/`update`/`detect` are `npx` CLI; `init` + the 23
  design verbs are `/impeccable` slash commands. Do not shell out to slash commands.

### 8.1 Hooks into other phase skills (flag-gated, additive)

- **sdd-apply**: new step parallel to Step 4b — when enabled + frontend changes, run
  `/impeccable <verb>`; add a Rules line ("When impeccable.enabled, run Impeccable design
  verbs on frontend-affecting changes during apply").
- **sdd-verify**: when enabled, advisory NOTE if Impeccable hooks/artifacts absent — never
  CRITICAL, never blocks (spec: note, not blocker; contrast Playwright's CRITICAL).
- **sdd-tasks**: add an "Impeccable pass" task when enabled AND the change touches
  frontend files.
- **sdd-explore**: web/frontend detection → recommend enabling Impeccable (preflight
  group F). Recommendation only, not auto-activation.
- **sdd-spec**: MAY annotate frontend design-language requirements (lightweight tag/prose,
  analogous to `@web`). Exact form left flexible per spec; recommend a `@design` prose
  note over a new hard tag to avoid coupling the Playwright `@web` selector.

---

## 9. Templates / CLAUDE.md — `internal/initcmd/templates.go`

Follow the "always-emitted text, flag-gated behavior" model (like group E). The template
body is static aside from `{{.SkillCount}}` etc.; there is no `{{if}}` conditional.

### 9.1 Preflight group F (in `orchestratorSections`, after group E ~lines 80-85)

```
- **F. Impeccable (Diseño de interfaz)** — "¿Activar Impeccable para calidad visual?"
  - No (recomendado): no correr verificaciones de diseño.
  - Sí: activar el gate de Impeccable tras verify/judge cuando esté habilitado.
```

### 9.2 Group F mapping paragraph (after the group E mapping paragraph)

```
**Project type & design-language gate (group F):**
- Group F maps to `impeccable.enabled` in `.archon/config.yaml`. The `--impeccable`
  flag at init time or the Impeccable tab in `archon tui` set the same value. When
  enabled, the harness invokes Impeccable subcommands during apply and runs the
  detection gate after the judge phase.
```

### 9.3 New Rules line — in BOTH `orchestratorRulesClaude` (~line 175) and `orchestratorRulesOpencode` (~line 188)

```
When impeccable.enabled, run Impeccable subcommands during apply and the detection
gate after judge passes.
```

Both consts must gain the identical line (byte-consistent). It should be added as a new
numbered rule following the existing Playwright rule 6, renumbering downstream rules if
they are numbered (verify the exact numbering during apply).

### 9.4 Repo `CLAUDE.md` mirror

The repo-root `CLAUDE.md` is a rendered copy of the template. Add group F + the mapping
paragraph to its Preflight section (group E area) and the new Rules line to its `## Rules`
section, byte-consistent with templates.go. `templates_test.go` catches template-side
drift but NOT the repo file, so this edit is manual and easy to forget — call it out in
the apply task list.

### 9.5 `templates_test.go` assertions to add

- Assert generated CLAUDE.md contains the group F question line ("F. Impeccable").
- Assert it contains the group F mapping paragraph ("Group F maps to `impeccable.enabled`").
- Assert the Impeccable Rules line appears — and assert it appears in **both** the Claude
  and Opencode rendered variants (mirror the existing group-E / rule-6 assertions).

---

## 10. PR slicing (finalized against ask-always, 400-line budget)

Per session preference (memory: split over raise-budget). Three chained PRs:

### PR1 — Go wiring (~330 lines, no behavior change)

| File | Est. lines |
|------|-----------|
| `internal/config/config.go` (struct + field + Clone + Load default/validate + helper) | ~40 |
| `internal/config/config_test.go` (CloneRoundtrip fixture + defaults/validation test) | ~35 |
| `cmd/archon/config.go` (get/set 5 keys + 2 error strings) | ~45 |
| `cmd/archon/config_test.go` or equivalent (key-list assertions) | ~20 |
| `cmd/archon/main.go` (flag var + Options + BoolVar) | ~5 |
| `internal/initcmd/init.go` (Options field + buildConfig param + set) | ~10 |
| `internal/tui/impeccable_tab.go` (NEW, cloned) | ~150 |
| `internal/tui/model.go` (enum + 10 lockstep sites) | ~15 |
| `internal/tui/model_test.go` (wrap-test fix + order label + ApplyToConfig test) | ~35 |
| `internal/status/display.go` (Impeccable block) | ~12 |
| `README.md` / `AGENTS.md` (`--impeccable` doc) | ~4 |

Fully testable via `go test ./...`; no skill consumes the flag yet, so behavior is
unchanged. Comfortably under 400.

### PR2 — Orchestration (skills + templates), estimated ~360 → likely triggers PR3

| File | Est. lines |
|------|-----------|
| `skills/impeccable/SKILL.md` (NEW thin skill) | ~120 |
| `skills/sdd-design/SKILL.md` (read-only hook) | ~30 |
| `skills/sdd-apply/SKILL.md` (step + Rules line) | ~35 |
| `skills/sdd-verify/SKILL.md` (advisory note) | ~20 |
| `skills/sdd-tasks/SKILL.md` (Impeccable pass task) | ~20 |
| `skills/sdd-explore/SKILL.md` (group F recommendation) | ~20 |
| `skills/sdd-spec/SKILL.md` (@design annotation note) | ~15 |
| `internal/initcmd/templates.go` (group F + mapping + 2 rule lines) | ~30 |
| `internal/initcmd/templates_test.go` (group F + rule assertions) | ~25 |
| `CLAUDE.md` (repo mirror) | ~15 |
| `README.md` (feature note) | ~10 |

Prose-heavy; likely > 400. **Pre-planned PR3 split below.** Confirm at the ask-always gate.

### PR3 (conditional, ~150 lines) — judge gate + template, split out of PR2 if PR2 > 400

| File | Est. lines |
|------|-----------|
| `skills/harness-judge/SKILL.md` (Step 3c + result column + edge cases + output contract) | ~110 |
| `internal/initcmd/templates.go` (group F + rule) | ~30 |
| `internal/initcmd/templates_test.go` (assertions) | ~25 |
| `CLAUDE.md` (repo mirror) | ~15 |

If PR3 is taken, PR2 sheds the templates + judge rows and lands ~200 (skills only).
Boundary is clean: PR2 = phase-skill prose + thin skill; PR3 = the blocking gate + the
orchestrator-template rule that advertises it. Each stays < 400.

---

## 11. Risks and residual open questions

| # | Risk / Question | Likelihood | Mitigation |
|---|-----------------|------------|-----------|
| 1 | `npx impeccable detect --json` schema undocumented; parsing may drift | Med | Parse defensively; unparseable → advisory + note, never hard-fail; exit code only for crash/not-found |
| 2 | `detect` exit code on violations undocumented (CLI fact #3) | Med | Never rely on exit code for pass/fail; derive verdict from `--json` payload |
| 3 | LLM critique non-deterministic under `block-all` | High | Default `block-deterministic`; user opts into `block-all` consciously |
| 4 | Node/npx absent in target repo | High | Step 3c returns `blocked` with actionable message; never silent-pass |
| 5 | `severity` validation added at `Load()` is net-new (no security.profile precedent at load) | Low | Small shared helper; normalize empty→default BEFORE validate; covered by new test |
| 6 | Repo CLAUDE.md drift vs templates.go (not caught by templates_test.go) | Med | Manual mirror edit called out as an explicit apply task |
| 7 | TUI wrap test coupling to "last tab" | Low (resolved) | Insert ImpeccableTab last; only `TestModel_Update_ShiftTabWrapsFromAgent` line 120 changes |
| 8 | Design-phase hook over-scope (no precedent) | Med | Strictly read-only; no detect, no file gen; deferrable to a later slice (Decision 5) |
| 9 | PR2 exceeds 400 | High | Pre-planned PR3 split; confirm at ask-always gate |

### Residual open questions (for the Human Review Gate)

1. **Apply-phase verb invocation.** The design verbs are `/impeccable <verb>` slash
   commands run by the AI agent, not shell commands. The apply-phase hook therefore
   instructs the agent to run them rather than shelling out. Confirm this is acceptable —
   it means the apply step is agent-behavioral, not a deterministic harness call. (This
   is the biggest correction vs the spec's implicit "npx everything" model.)
2. **Design-phase hook scope.** Decision 5 flagged it as deferrable. Keep it in PR2 as a
   read-only hook, or defer to a fast-follow to shrink PR2? Recommend keep (it is small
   and read-only).
3. **spec `@design` annotation.** Recommend a lightweight prose note over a new hard tag,
   to avoid coupling to Playwright's `@web` selector. Confirm the annotation form.
4. **TUI severity write-back.** On save, should `applyToConfig` fall back to
   `block-deterministic` when the input is left blank, or write empty and rely on
   `Load()` normalization? Both are safe; recommend the fallback for a clean on-disk file.
