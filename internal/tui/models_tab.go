package tui

import (
	"strings"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/opencode"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// subMode is the per-tab interaction mode. Only one row can be "open" (in
// providerSelect/modelSelect/freeForm) at a time; that row is m.focusedRow.
type subMode int

const (
	rowNav         subMode = iota // navigating between rows (Up/Down)
	providerSelect                // choosing a provider for the focused row
	modelSelect                   // choosing a model within the picked provider
	freeForm                      // typing a raw provider/model string
)

// modelRow is one editable assignment line: Default, one per SDD phase, and
// (opencode only) Leader. ref holds the live value seeded from cfg; changed is
// set the moment the user picks/clears/free-forms this row, and gates legacy
// preservation in applyToConfig.
type modelRow struct {
	label   string  // "Default" / phase name / "Leader"
	kind    rowKind // rowDefault | rowPhase | rowLeader
	phase   string  // phase name when kind==rowPhase, else ""
	ref     config.ModelRef
	changed bool
}

type rowKind int

const (
	rowDefault rowKind = iota
	rowPhase
	rowLeader
)

// modelsTabState is the rewritten Models tab. The leader row, when present, is
// always the final element of rows (kind==rowLeader); leaderEnabled mirrors the
// old behavior (opencode only).
type modelsTabState struct {
	rows       []modelRow
	focusedRow int
	mode       subMode

	// Provider catalog loaded once at construction from the opencode cache.
	providers map[string]opencode.Provider // provider id -> Provider
	available []string                     // sorted provider ids (DetectAvailableProviders)

	// Picker cursors (valid only while mode != rowNav).
	providerCursor int              // index into m.available
	modelCursor    int              // index into the active provider's FilterModelsForSDD list
	pickedProvider string           // provider chosen in providerSelect, used by modelSelect
	curModels      []opencode.Model // FilterModelsForSDD(providers[pickedProvider]), cached for modelSelect

	// Free-form text entry (shared; reset on each open).
	input textinput.Model

	// cacheWarning is non-empty ONLY when LoadModelsOrEmpty returned an error
	// (cache present but unreadable). Absent cache => "" (silent). Rendered inline.
	cacheWarning string

	leaderEnabled bool
	width         int
}

// newModelsTabState builds the Models tab from cfg and the opencode provider
// catalog. providers is the map from opencode.LoadModelsOrEmpty (empty when the
// cache is absent); loadErr is the error it returned (non-nil only when the
// cache is present but unreadable — the corrupt-cache seam from 3a). The tab
// seeds each row's ModelRef from cfg.Models and never re-reads the cache.
func newModelsTabState(cfg *config.Config, providers map[string]opencode.Provider, loadErr error) modelsTabState {
	leaderEnabled := cfg.Agent == "opencode"

	rows := make([]modelRow, 0, 1+len(config.PhaseOrder)+1)
	rows = append(rows, modelRow{label: "Default", kind: rowDefault, ref: cfg.Models.Default})
	for _, phase := range config.PhaseOrder { // deterministic order, 8 phases
		var ref config.ModelRef
		if cfg.Models.Phases != nil {
			ref = cfg.Models.Phases[phase] // zero ModelRef when unset
		}
		rows = append(rows, modelRow{label: phase, kind: rowPhase, phase: phase, ref: ref})
	}
	if leaderEnabled {
		rows = append(rows, modelRow{label: "Leader", kind: rowLeader, ref: cfg.Models.Leader})
	}

	ti := textinput.New()
	ti.Placeholder = "provider/model"
	ti.Width = 30

	if providers == nil {
		providers = map[string]opencode.Provider{}
	}

	st := modelsTabState{
		rows:          rows,
		focusedRow:    0,
		mode:          rowNav,
		providers:     providers,
		available:     opencode.DetectAvailableProviders(providers), // sorted, may be empty
		input:         ti,
		leaderEnabled: leaderEnabled,
	}
	if loadErr != nil {
		st.cacheWarning = "⚠ opencode model cache is unreadable — pick is unavailable; type a model name (e)"
	}
	return st
}

func (m *modelsTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		// Non-key msgs only matter to the textinput while in freeForm.
		if m.mode == freeForm {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return cmd, true
		}
		return nil, true
	}

	switch m.mode {
	case rowNav:
		return m.updateRowNav(key)
	case providerSelect:
		return m.updateProviderSelect(key)
	case modelSelect:
		return m.updateModelSelect(key)
	case freeForm:
		return m.updateFreeForm(key)
	}
	return nil, true
}

func (m *modelsTabState) updateRowNav(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.focusedRow > 0 {
			m.focusedRow--
		}
		return nil, true
	case tea.KeyDown:
		if m.focusedRow < len(m.rows)-1 {
			m.focusedRow++
		}
		return nil, true
	case tea.KeyEnter:
		if len(m.available) == 0 {
			m.openFreeForm()
			return nil, true
		}
		m.mode = providerSelect
		m.providerCursor = 0
		return nil, true
	case tea.KeyRunes:
		if len(key.Runes) == 1 && key.Runes[0] == 'e' {
			m.openFreeForm()
			return nil, true
		}
	}
	return nil, true
}

func (m *modelsTabState) openFreeForm() {
	m.mode = freeForm
	m.input.SetValue(m.rows[m.focusedRow].ref.FullID())
	m.input.CursorEnd()
	m.input.Focus()
}

func (m *modelsTabState) updateProviderSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.providerCursor > 0 {
			m.providerCursor--
		}
	case tea.KeyDown:
		if m.providerCursor < len(m.available)-1 {
			m.providerCursor++
		}
	case tea.KeyEnter:
		m.pickedProvider = m.available[m.providerCursor]
		m.curModels = opencode.FilterModelsForSDD(m.providers[m.pickedProvider])
		if len(m.curModels) == 0 {
			m.openFreeForm() // provider qualified (built-in) but has no SDD models
			return nil, true
		}
		m.mode = modelSelect
		m.modelCursor = 0
	case tea.KeyEsc:
		m.mode = rowNav
	}
	return nil, true
}

func (m *modelsTabState) updateModelSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.modelCursor > 0 {
			m.modelCursor--
		}
	case tea.KeyDown:
		if m.modelCursor < len(m.curModels)-1 {
			m.modelCursor++
		}
	case tea.KeyEnter:
		m.rows[m.focusedRow].ref = refFromCacheKey(m.pickedProvider, m.curModels[m.modelCursor].ID)
		m.rows[m.focusedRow].changed = true
		m.mode = rowNav
	case tea.KeyEsc:
		m.mode = providerSelect
	}
	return nil, true
}

// refFromCacheKey builds a ModelRef from a provider id and a cache model key.
// Under the "opencode" provider keys are BARE (e.g. "deepseek-v4-pro") so we set
// Provider+Model and FullID yields "opencode/deepseek-v4-pro". Under other
// providers the key is ALREADY-SLASHED (e.g. "xai/grok-4"); putting it in Model
// alone lets FullID's "/" short-circuit return it as-is (no double-prefix).
func refFromCacheKey(provider, key string) config.ModelRef {
	if strings.Contains(key, "/") {
		return config.ModelRef{Model: key} // already provider-qualified; FullID returns as-is
	}
	return config.ModelRef{Provider: provider, Model: key}
}

func (m *modelsTabState) updateFreeForm(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyEnter:
		m.rows[m.focusedRow].ref = config.ParseModelRef(strings.TrimSpace(m.input.Value()))
		m.rows[m.focusedRow].changed = true
		m.input.Blur()
		m.mode = rowNav
		return nil, true
	case tea.KeyEsc:
		m.input.Blur()
		m.mode = rowNav
		return nil, true
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(key)
	return cmd, true
}

// hintLine returns the mode-specific keyboard hint.
func (m *modelsTabState) hintLine() string {
	switch m.mode {
	case providerSelect:
		return "↑/↓: choose provider · Enter: next · Esc: cancel"
	case modelSelect:
		return "↑/↓: choose model · Enter: set · Esc: back"
	case freeForm:
		return "type provider/model · Enter: set · Esc: cancel"
	default:
		if len(m.available) == 0 {
			return "No detected models — press e to type a provider/model"
		}
		return "↑/↓: move · Enter: pick provider/model · e: type a model"
	}
}

func (m *modelsTabState) view(width, height int) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).MarginBottom(1)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	label := lipgloss.NewStyle().Width(15).Align(lipgloss.Right).MarginRight(1)
	focus := lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("63")).Foreground(lipgloss.Color("0"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	b.WriteString(title.Render("Model Configuration"))
	b.WriteString("\n")
	b.WriteString(hint.Render(m.hintLine()))
	b.WriteString("\n")
	if m.cacheWarning != "" {
		warn := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		b.WriteString(warn.Render(m.cacheWarning))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	for i := range m.rows {
		b.WriteString(m.renderRow(i, label, focus, dim))
		b.WriteString("\n")
	}
	return b.String()
}

func (m *modelsTabState) renderRow(i int, label, focus, dim lipgloss.Style) string {
	row := m.rows[i]
	var b strings.Builder

	// When this row is the focused one in a non-rowNav mode, render the picker/free-form.
	if i == m.focusedRow && m.mode != rowNav {
		switch m.mode {
		case providerSelect:
			b.WriteString(label.Render(row.label + ":"))
			b.WriteString(" [pick provider]\n")
			for j, id := range m.available {
				marker := "  "
				if j == m.providerCursor {
					marker = "▸ "
					b.WriteString(focus.Render(marker + id))
				} else {
					b.WriteString("  " + marker + id)
				}
				b.WriteString("\n")
			}
			return strings.TrimRight(b.String(), "\n")

		case modelSelect:
			b.WriteString(label.Render(row.label + ":"))
			b.WriteString(" Provider: " + m.pickedProvider + "\n")
			for j, mod := range m.curModels {
				name := mod.Name
				if name == "" {
					name = mod.ID
				}
				if j == m.modelCursor {
					b.WriteString(focus.Render("▸ " + name))
				} else {
					b.WriteString("    " + name)
				}
				b.WriteString("\n")
			}
			return strings.TrimRight(b.String(), "\n")

		case freeForm:
			b.WriteString(label.Render(row.label + ":"))
			b.WriteString(" ")
			b.WriteString(m.input.View())
			b.WriteString("  ")
			b.WriteString(dim.Render("Enter to set · Esc to cancel"))
			return b.String()
		}
	}

	// Plain row display.
	marker := "  "
	rowLabel := label.Render(row.label + ":")
	value := row.ref.FullID()
	if value == "" && row.kind == rowPhase {
		value = dim.Render("(default)")
	}

	if i == m.focusedRow {
		marker = "▸ "
		b.WriteString(marker)
		b.WriteString(focus.Render(rowLabel + " " + value))
	} else {
		b.WriteString(marker)
		b.WriteString(rowLabel)
		b.WriteString(" ")
		b.WriteString(value)
	}
	return b.String()
}

func (m *modelsTabState) applyToConfig(cfg *config.Config) {
	if cfg.Models.Phases == nil {
		cfg.Models.Phases = make(map[string]config.ModelRef)
	}
	for _, row := range m.rows {
		switch row.kind {
		case rowDefault:
			cfg.Models.Default = row.ref // seeded verbatim when unchanged
		case rowLeader:
			cfg.Models.Leader = row.ref
		case rowPhase:
			if row.ref.Model == "" {
				// Empty: delete the phase entry (whether seeded-empty or cleared).
				delete(cfg.Models.Phases, row.phase)
			} else {
				cfg.Models.Phases[row.phase] = row.ref
			}
		}
	}
}

func (m *modelsTabState) setWidth(width int) {
	m.width = width
	m.input.Width = width - 20
}
