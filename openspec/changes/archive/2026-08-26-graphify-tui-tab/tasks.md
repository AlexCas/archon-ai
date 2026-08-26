# Tasks: Graphify tab in `archon tui`

<!-- [[graphify-integration]] · change: graphify-tui-tab · phase: tasks -->

Implements R-19 and the R-05 doc modification. One PR, ~320 lines, budget 800.
All decisions locked by design.md — do not re-derive.

---

## Group A — New tab file

### A-1 Create `internal/tui/graphify_tab.go`

Clone `internal/tui/impeccable_tab.go` (182 lines) into a new file and adapt it
for `config.Graphify`'s five fields (three bools, two textinputs; bools-first).

- [x] Create `/home/skollhowl/Projects/archon-ai/internal/tui/graphify_tab.go`
  with package `tui` and the following structure (adapt every name from
  `impeccable`/`Impeccable` → `graphify`/`Graphify`):

  **Type and constant**
  - `graphifyTabState` struct: fields `enabled bool`, `autoInstall bool`,
    `semantic bool`, `version textinput.Model`, `outputDir textinput.Model`,
    `focused int` (in that exact order).
  - `const graphifyFocusCount = 5`

  **`newGraphifyTabState(cfg config.Graphify) graphifyTabState`**
  - `version` input: `Placeholder = "v0.9.45 (default)"`,
    `SetValue(cfg.Version)`, `Width = 30`.
  - `outputDir` input: `Placeholder = ".archon/graphify (default)"`,
    `SetValue(cfg.OutputDir)`, `Width = 30`.
  - Return struct with `enabled: cfg.Enabled`, `autoInstall: cfg.AutoInstall`,
    `semantic: cfg.Semantic`, `version: version`, `outputDir: outputDir`,
    `focused: 0`.

  **`update(msg tea.Msg) (tea.Cmd, bool)`**
  - Up/Down: cycle `focused` mod `graphifyFocusCount`, call `refocus()`.
  - Enter/Space at focus 0: toggle `enabled`; focus 1: toggle `autoInstall`;
    focus 2: toggle `semantic`.
  - Else forward to the focused input at focus 3 (`version`) or 4 (`outputDir`).

  **`refocus()`**
  - `p.version.Blur()`, `p.outputDir.Blur()`.
  - Focus `p.version` at focus 3, `p.outputDir` at focus 4 (nothing at 0/1/2).

  **`view(width, height int) string`**
  - Title: `"Graphify (Code Graph) Configuration"` in the same bold/color-63
    lipgloss style as Impeccable's title.
  - Three toggle lines: `enabled` (focus 0), `autoInstall` (focus 1),
    `semantic` (focus 2) — each with `[ON/OFF] Label (press Enter to toggle)`
    and highlighted when focused.
  - Two labelled textinput lines: `"Version:"` → `p.version.View()`;
    `"Output dir:"` → `p.outputDir.View()`, using the same `labelStyle`
    (`Width(14).Align(Right).MarginRight(1)`) as Impeccable.
  - Info footer (color 240): `"Graphify is advisory-only (non-blocking). When"` /
    `"enabled, sdd-explore consults the code graph."`.

  **`applyToConfig(cfg *config.Config)`**
  ```go
  cfg.Graphify.Enabled = p.enabled
  cfg.Graphify.AutoInstall = p.autoInstall
  cfg.Graphify.Semantic = p.semantic
  v := strings.TrimSpace(p.version.Value())
  if v == "" { v = config.DefaultGraphifyVersion }
  cfg.Graphify.Version = v
  o := strings.TrimSpace(p.outputDir.Value())
  if o == "" { o = config.DefaultGraphifyOutputDir }
  cfg.Graphify.OutputDir = o
  ```
  No `ValidateGraphify…` call — Graphify is advisory-only (no severity enum,
  no Load()-time validation at config.go L60-63).

  **`setWidth(width int)`**
  - `w := width - 20; if w < 10 { w = 10 }`.
  - Apply `w` to both `p.version.Width` and `p.outputDir.Width`.

  **Imports**: same as `impeccable_tab.go` — `fmt`, `strings`,
  `github.com/charmbracelet/bubbles/textinput`, `tea`, `lipgloss`,
  `github.com/archon-ai/archon/internal/config`.

---

## Group B — `model.go` wiring (nine sites)

All edits are in `internal/tui/model.go`. Apply in source order (top to bottom)
so each subsequent line number remains accurate.

### B-1 Iota block (L22-31): add `GraphifyTab`

- [x] After `ImpeccableTab` (L29) and before `tabCount` (L30), insert:
  ```go
  GraphifyTab
  ```
  Result: `GraphifyTab` becomes the new last named tab; `tabCount` increments
  automatically.

### B-2 Model struct field (L51): add `graphifyTab`

- [x] After the `impeccableTab impeccableTabState` line (L51), insert:
  ```go
  graphifyTab   graphifyTabState
  ```

### B-3 `NewModel` constructor (L113): seed `graphifyTab`

- [x] After `impeccableTab: newImpeccableTabState(cfg.Impeccable),` (L113),
  insert:
  ```go
  graphifyTab:   newGraphifyTabState(cfg.Graphify),
  ```

### B-4 `WindowSizeMsg` setWidth fan-out (L134): add setWidth call

- [x] After `m.impeccableTab.setWidth(m.width)` (L134), insert:
  ```go
  m.graphifyTab.setWidth(m.width)
  ```

### B-5 Key-dispatch switch (L184): add `case GraphifyTab`

- [x] After the `case ImpeccableTab:` block (L184-187):
  ```go
  case ImpeccableTab:
      cmd, _ := m.impeccableTab.update(msg)
      if cmd != nil {
          cmds = append(cmds, cmd)
      }
  ```
  Insert immediately after:
  ```go
  case GraphifyTab:
      cmd, _ := m.graphifyTab.update(msg)
      if cmd != nil {
          cmds = append(cmds, cmd)
      }
  ```

### B-6 `agentInitDoneMsg` rebuild (L216): re-seed `graphifyTab`

- [x] After `m.impeccableTab = newImpeccableTabState(msg.cfg.Impeccable)` (L216),
  insert:
  ```go
  m.graphifyTab = newGraphifyTabState(msg.cfg.Graphify)
  ```

### B-7 `agentInitDoneMsg` setWidth fan-out (L224): add setWidth call

- [x] After `m.impeccableTab.setWidth(m.width)` (L224, inside `if m.width > 0`),
  insert:
  ```go
  m.graphifyTab.setWidth(m.width)
  ```

### B-8 `saveConfig` applyToConfig fan-out (L357): add applyToConfig call

- [x] After `m.impeccableTab.applyToConfig(cfg)` (L357), insert:
  ```go
  m.graphifyTab.applyToConfig(cfg)
  ```

### B-9 Render helpers (L286 + L317): tab label and content case

- [x] In `renderTabs` (L286), append `"Graphify"` to the `tabs` slice so it
  reads:
  ```go
  tabs := []string{"Agent", "Models", "Judge", "Mutation Testing", "Playwright", "Security", "Impeccable", "Graphify"}
  ```

- [x] In `renderTabContent` (L317-318), after the `case ImpeccableTab:` return,
  insert:
  ```go
  case GraphifyTab:
      return style.Render(m.graphifyTab.view(m.width, m.height))
  ```

---

## Group C — Tests in `model_test.go`

### C-1 Fix `TestModel_Update_ShiftTabWrapsFromAgent` (L107-123)

- [x] Update the doc comment at L107-109 from:
  ```
  // TestModel_Update_ShiftTabWrapsFromAgent verifies that a single Shift+Tab
  // from a freshly-constructed model (default AgentTab) wraps around to the
  // last tab, ImpeccableTab.
  ```
  to:
  ```
  // TestModel_Update_ShiftTabWrapsFromAgent verifies that a single Shift+Tab
  // from a freshly-constructed model (default AgentTab) wraps around to the
  // last tab, GraphifyTab.
  ```

- [x] At L121, change:
  ```go
  if model.activeTab != ImpeccableTab {
      t.Errorf("activeTab after Shift+Tab from AgentTab = %d, want %d", model.activeTab, ImpeccableTab)
  }
  ```
  to:
  ```go
  if model.activeTab != GraphifyTab {
      t.Errorf("activeTab after Shift+Tab from AgentTab = %d, want %d", model.activeTab, GraphifyTab)
  }
  ```

### C-2 Fix `TestModel_renderTabs_Order` (L222)

- [x] At L222, append `"Graphify"` to the `labels` slice:
  ```go
  labels := []string{"Agent", "Models", "Judge", "Mutation Testing", "Playwright", "Security", "Impeccable", "Graphify"}
  ```

### C-3 Add `TestGraphifyTabState_ApplyToConfig` (new, after L434)

- [x] Insert a new test function immediately after
  `TestImpeccableTabState_ApplyToConfig_BlankSeverityFallback` (which ends at
  ~L434):

  ```go
  func TestGraphifyTabState_ApplyToConfig(t *testing.T) {
      cfg := &config.Config{}
      state := newGraphifyTabState(config.Graphify{})

      // Drive bool toggles at focus 0, 1, 2.
      state.focused = 0
      state.update(tea.KeyMsg{Type: tea.KeyEnter})
      if !state.enabled {
          t.Error("enabled should be true after toggle")
      }
      state.focused = 1
      state.update(tea.KeyMsg{Type: tea.KeyEnter})
      if !state.autoInstall {
          t.Error("autoInstall should be true after toggle")
      }
      state.focused = 2
      state.update(tea.KeyMsg{Type: tea.KeyEnter})
      if !state.semantic {
          t.Error("semantic should be true after toggle")
      }

      // Set text inputs directly.
      state.version.SetValue("v1.2.3")
      state.outputDir.SetValue(".archon/out")

      state.applyToConfig(cfg)

      if !cfg.Graphify.Enabled {
          t.Error("cfg.Graphify.Enabled should be true")
      }
      if !cfg.Graphify.AutoInstall {
          t.Error("cfg.Graphify.AutoInstall should be true")
      }
      if !cfg.Graphify.Semantic {
          t.Error("cfg.Graphify.Semantic should be true")
      }
      if cfg.Graphify.Version != "v1.2.3" {
          t.Errorf("cfg.Graphify.Version = %q, want %q", cfg.Graphify.Version, "v1.2.3")
      }
      if cfg.Graphify.OutputDir != ".archon/out" {
          t.Errorf("cfg.Graphify.OutputDir = %q, want %q", cfg.Graphify.OutputDir, ".archon/out")
      }
  }

  // TestGraphifyTabState_ApplyToConfig_BlankCoercion asserts that blank version
  // and outputDir inputs fall back to the package defaults on save — never
  // persist "" (config.Load only re-seeds when the block is absent, not for a
  // present-but-empty key).
  func TestGraphifyTabState_ApplyToConfig_BlankCoercion(t *testing.T) {
      cfg := &config.Config{}
      state := newGraphifyTabState(config.Graphify{Version: "", OutputDir: ""})
      state.version.SetValue("")
      state.outputDir.SetValue("")

      state.applyToConfig(cfg)

      if cfg.Graphify.Version != config.DefaultGraphifyVersion {
          t.Errorf("cfg.Graphify.Version = %q, want %q", cfg.Graphify.Version, config.DefaultGraphifyVersion)
      }
      if cfg.Graphify.OutputDir != config.DefaultGraphifyOutputDir {
          t.Errorf("cfg.Graphify.OutputDir = %q, want %q", cfg.Graphify.OutputDir, config.DefaultGraphifyOutputDir)
      }
  }
  ```

---

## Group D — Docs hand-edits (never `archon init --force`)

### D-1 `internal/initcmd/templates.go` — group G clause (L100-101)

- [x] In `internal/initcmd/templates.go` at L100-101, the group G block
  currently reads (§ represents backtick-like escaping inside the template):
  ```
  - Group G maps to §graphify.enabled§ in §.archon/config.yaml§. The §--graphify§
    flag at init time sets the same value. When enabled, ...
  ```
  Change the second sentence to:
  ```
  The §--graphify§ flag at init time or the Graphify tab in §archon tui§ set the same value.
  ```
  so the full two-line group G paragraph in the template matches:
  ```
  - Group G maps to §graphify.enabled§ in §.archon/config.yaml§. The §--graphify§
    flag at init time or the Graphify tab in §archon tui§ set the same value. When enabled, sdd-explore consults the
    Graphify code graph for repo comprehension and sdd-tasks reads Leiden
    communities to inform slice boundaries — advisory only, never blocking.
  ```
  Keep all `§` escapes intact. Do NOT call `archon init --force`.

### D-2 `CLAUDE.md` — group G clause (L95-96)

- [x] In `CLAUDE.md` at L95-96, the group G block currently reads:
  ```
  - Group G maps to `graphify.enabled` in `.archon/config.yaml`. The `--graphify`
    flag at init time sets the same value. When enabled, sdd-explore consults the
  ```
  Change the second sentence to:
  ```
  The `--graphify` flag at init time or the Graphify tab in `archon tui` set the same value.
  ```
  so the full paragraph becomes:
  ```
  - Group G maps to `graphify.enabled` in `.archon/config.yaml`. The `--graphify`
    flag at init time or the Graphify tab in `archon tui` set the same value. When enabled, sdd-explore consults the
    Graphify code graph for repo comprehension and sdd-tasks reads Leiden
    communities to inform slice boundaries — advisory only, never blocking.
  ```
  Do NOT call `archon init --force`.

---

## Group E — Verification

### E-1 Compile check

- [x] Run `go build ./internal/tui/...` — must produce no errors.

### E-2 Run TUI tests

- [x] Run `go test ./internal/tui/...` — all tests must pass, including:
  - `TestModel_Update_ShiftTabWrapsFromAgent` (now wraps to `GraphifyTab`)
  - `TestModel_renderTabs_Order` (now includes `"Graphify"`)
  - `TestGraphifyTabState_ApplyToConfig`
  - `TestGraphifyTabState_ApplyToConfig_BlankCoercion`
  - `TestIntegration_TabStateConsistency` (guards fan-out wiring automatically)

---

## Notes

- **Blank coercion**: `applyToConfig` coerces blank `version` →
  `config.DefaultGraphifyVersion` (`"v0.9.45"`) and blank `outputDir` →
  `config.DefaultGraphifyOutputDir` (`".archon/graphify"`). These constants are
  defined at `internal/config/config.go` L77-78.
- **No validation guard**: unlike Impeccable, `applyToConfig` has no
  `ValidateGraphify…` call — Graphify has no severity enum and no Load()-time
  validation.
- **Docs constraint**: `templates-go-drift` memory note applies — D-1 and D-2
  are hand-edits only. Running `archon init --force` would revert the
  archive-before-PR content merged in earlier PRs.
- **Estimated size**: ~130 lines (new file) + ~15 wiring + ~45 tests + ~4 docs
  ≈ 194-320 lines; well under the 800-line budget. No chained-PR split needed.
