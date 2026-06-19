package tui

import (
	"fmt"
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
	effortSelect                  // choosing an effort/reasoning level for a reasoning model
)

// effortOptions are the fixed effort levels offered for reasoning-capable models.
// "default" maps to an empty Effort (provider default); the rest are passed verbatim.
var effortOptions = []string{"default", "low", "medium", "high"}

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
	effortCursor   int              // index into effortOptions (valid only while mode == effortSelect)

	// Picker scroll state prevents the list overflowing the terminal.
	// pickerOffset is the first visible item index; pickerMaxVisible caps the
	// rendered window. Both are valid only while mode is providerSelect or
	// modelSelect, and are updated in view() plus the updateXxxSelect methods.
	pickerOffset     int
	pickerMaxVisible int

	// Picker live filter. filter holds the typed search string; filteredIndices
	// maps the visible (matching) position back to the original index in
	// available or curModels depending on mode. nil filteredIndices means no
	// filter (show all).
	filter          string
	filteredIndices []int

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
	case effortSelect:
		return m.updateEffortSelect(key)
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
		m.pickerOffset = 0
		m.filter = ""
		m.filteredIndices = nil
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

// rebuildFilter recomputes filteredIndices from m.filter and the current mode.
// It is called whenever the filter text changes or the mode transitions to
// providerSelect/modelSelect with a non-empty filter.
func (m *modelsTabState) rebuildFilter() {
	if m.filter == "" {
		m.filteredIndices = nil
		return
	}
	lower := strings.ToLower(m.filter)
	switch m.mode {
	case providerSelect:
		idx := make([]int, 0, len(m.available))
		for i, p := range m.available {
			if strings.Contains(strings.ToLower(p), lower) {
				idx = append(idx, i)
			}
		}
		m.filteredIndices = idx
	case modelSelect:
		idx := make([]int, 0, len(m.curModels))
		for i, mod := range m.curModels {
			name := mod.Name
			if name == "" {
				name = mod.ID
			}
			if strings.Contains(strings.ToLower(name), lower) {
				idx = append(idx, i)
			}
		}
		m.filteredIndices = idx
	default:
		m.filteredIndices = nil
	}
}

// filteredLen returns the number of items visible after applying m.filter.
func (m *modelsTabState) filteredLen() int {
	if m.filteredIndices == nil {
		switch m.mode {
		case providerSelect:
			return len(m.available)
		case modelSelect:
			return len(m.curModels)
		}
	}
	return len(m.filteredIndices)
}

func (m *modelsTabState) updateProviderSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	total := m.filteredLen()

	switch key.Type {
	case tea.KeyUp:
		if m.providerCursor > 0 {
			m.providerCursor--
		}
		if m.providerCursor < m.pickerOffset {
			m.pickerOffset = m.providerCursor
		}
	case tea.KeyDown:
		if m.providerCursor < total-1 {
			m.providerCursor++
		}
		if m.providerCursor >= m.pickerOffset+m.pickerMaxVisible {
			m.pickerOffset = m.providerCursor - m.pickerMaxVisible + 1
		}
	case tea.KeyEnter:
		if total == 0 {
			return nil, true // no selection possible
		}
		idx := m.providerCursor
		if m.filteredIndices != nil {
			idx = m.filteredIndices[idx]
		}
		m.pickedProvider = m.available[idx]
		m.curModels = opencode.FilterModelsForSDD(m.providers[m.pickedProvider])
		if len(m.curModels) == 0 {
			m.openFreeForm()
			return nil, true
		}
		m.mode = modelSelect
		m.modelCursor = 0
		m.pickerOffset = 0
		m.filter = ""
		m.filteredIndices = nil
	case tea.KeyEsc:
		if m.filter != "" {
			m.filter = ""
			m.filteredIndices = nil
			m.providerCursor = 0
			m.pickerOffset = 0
		} else {
			m.mode = rowNav
		}
	case tea.KeyBackspace:
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildFilter()
			m.providerCursor = 0
			m.pickerOffset = 0
		}
	case tea.KeyRunes:
		for _, r := range key.Runes {
			if r >= 32 && r <= 126 { // printable ASCII
				m.filter += string(r)
			}
		}
		m.rebuildFilter()
		m.providerCursor = 0
		m.pickerOffset = 0
	}
	return nil, true
}

func (m *modelsTabState) updateModelSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	total := m.filteredLen()

	switch key.Type {
	case tea.KeyUp:
		if m.modelCursor > 0 {
			m.modelCursor--
		}
		if m.modelCursor < m.pickerOffset {
			m.pickerOffset = m.modelCursor
		}
	case tea.KeyDown:
		if m.modelCursor < total-1 {
			m.modelCursor++
		}
		if m.modelCursor >= m.pickerOffset+m.pickerMaxVisible {
			m.pickerOffset = m.modelCursor - m.pickerMaxVisible + 1
		}
	case tea.KeyEnter:
		if total == 0 {
			return nil, true
		}
		idx := m.modelCursor
		if m.filteredIndices != nil {
			idx = m.filteredIndices[idx]
		}
		picked := m.curModels[idx]
		m.rows[m.focusedRow].ref = refFromCacheKey(m.pickedProvider, picked.ID)
		m.rows[m.focusedRow].changed = true
		m.filter = ""
		m.filteredIndices = nil
		if picked.Reasoning {
			m.mode = effortSelect
			m.effortCursor = 0
		} else {
			m.mode = rowNav
		}
	case tea.KeyEsc:
		if m.filter != "" {
			m.filter = ""
			m.filteredIndices = nil
			m.modelCursor = 0
			m.pickerOffset = 0
		} else {
			m.mode = providerSelect
		}
	case tea.KeyBackspace:
		if m.filter != "" {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildFilter()
			m.modelCursor = 0
			m.pickerOffset = 0
		}
	case tea.KeyRunes:
		for _, r := range key.Runes {
			if r >= 32 && r <= 126 {
				m.filter += string(r)
			}
		}
		m.rebuildFilter()
		m.modelCursor = 0
		m.pickerOffset = 0
	}
	return nil, true
}

func (m *modelsTabState) updateEffortSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.effortCursor > 0 {
			m.effortCursor--
		}
	case tea.KeyDown:
		if m.effortCursor < len(effortOptions)-1 {
			m.effortCursor++
		}
	case tea.KeyEnter:
		opt := effortOptions[m.effortCursor]
		if opt == "default" {
			opt = ""
		}
		m.rows[m.focusedRow].ref.Effort = opt
		m.mode = rowNav
	case tea.KeyEsc:
		m.mode = modelSelect // step back; model already set
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
		return "↑/↓: choose provider · Enter: next · type to filter · Esc: cancel"
	case modelSelect:
		return "↑/↓: choose model · Enter: set · type to filter · Esc: back"
	case freeForm:
		return "type provider/model · Enter: set · Esc: cancel"
	case effortSelect:
		return "↑/↓: choose effort · Enter: set · Esc: back"
	default:
		if len(m.available) == 0 {
			return "No detected models — press e to type a provider/model"
		}
		return "↑/↓: move · Enter: pick provider/model · e: type a model"
	}
}

func (m *modelsTabState) view(width, height int) string {
	// Compute how many picker items fit in the visible area.
	// Overhead: tab chrome (~2 lines padding) + title (1) + hint (1) + blank (1)
	// + row overhead (2: label + header). For providerSelect and modelSelect
	// the effective item lines = height - ~13. Clamp to a sane range.
	m.pickerMaxVisible = height - 13
	if m.pickerMaxVisible < 3 {
		m.pickerMaxVisible = 3
	}
	if m.pickerMaxVisible > 20 {
		m.pickerMaxVisible = 20
	}

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
			b.WriteString(" [pick provider]")
			if m.filter != "" {
				b.WriteString(dim.Render("  filter: " + m.filter))
			}
			b.WriteString("\n")
			total := m.filteredLen()
			if total == 0 {
				b.WriteString(dim.Render("  (no matches)"))
				return strings.TrimRight(b.String(), "\n")
			}
			// Clamp cursor to valid range.
			if m.providerCursor >= total {
				m.providerCursor = total - 1
			}
			end := m.pickerOffset + m.pickerMaxVisible
			if end > total {
				end = total
			}
			if m.pickerOffset > 0 {
				b.WriteString(dim.Render(fmt.Sprintf("  ↑ %d more", m.pickerOffset)))
				b.WriteString("\n")
			}
			for visIdx := m.pickerOffset; visIdx < end; visIdx++ {
				origIdx := visIdx
				if m.filteredIndices != nil {
					origIdx = m.filteredIndices[visIdx]
				}
				id := m.available[origIdx]
				if visIdx == m.providerCursor {
					b.WriteString(focus.Render("▸ " + id))
				} else {
					b.WriteString("    " + id)
				}
				b.WriteString("\n")
			}
			if end < total {
				b.WriteString(dim.Render(fmt.Sprintf("  ↓ %d more", total-end)))
			}
			return strings.TrimRight(b.String(), "\n")

		case modelSelect:
			b.WriteString(label.Render(row.label + ":"))
			b.WriteString(" Provider: " + m.pickedProvider)
			if m.filter != "" {
				b.WriteString(dim.Render("  filter: " + m.filter))
			}
			b.WriteString("\n")
			total := m.filteredLen()
			if total == 0 {
				b.WriteString(dim.Render("  (no matches)"))
				return strings.TrimRight(b.String(), "\n")
			}
			// Clamp cursor to valid range.
			if m.modelCursor >= total {
				m.modelCursor = total - 1
			}
			end := m.pickerOffset + m.pickerMaxVisible
			if end > total {
				end = total
			}
			if m.pickerOffset > 0 {
				b.WriteString(dim.Render(fmt.Sprintf("  ↑ %d more", m.pickerOffset)))
				b.WriteString("\n")
			}
			for visIdx := m.pickerOffset; visIdx < end; visIdx++ {
				origIdx := visIdx
				if m.filteredIndices != nil {
					origIdx = m.filteredIndices[visIdx]
				}
				mod := m.curModels[origIdx]
				name := mod.Name
				if name == "" {
					name = mod.ID
				}
				if visIdx == m.modelCursor {
					b.WriteString(focus.Render("▸ " + name))
				} else {
					b.WriteString("    " + name)
				}
				b.WriteString("\n")
			}
			if end < total {
				b.WriteString(dim.Render(fmt.Sprintf("  ↓ %d more", total-end)))
			}
			return strings.TrimRight(b.String(), "\n")

		case freeForm:
			b.WriteString(label.Render(row.label + ":"))
			b.WriteString(" ")
			b.WriteString(m.input.View())
			b.WriteString("  ")
			b.WriteString(dim.Render("Enter to set · Esc to cancel"))
			return b.String()

		case effortSelect:
			b.WriteString(label.Render(row.label + ":"))
			b.WriteString(" Effort:\n")
			for j, opt := range effortOptions {
				if j == m.effortCursor {
					b.WriteString(focus.Render("▸ " + opt))
				} else {
					b.WriteString("    " + opt)
				}
				b.WriteString("\n")
			}
			b.WriteString(dim.Render("↑/↓: choose effort · Enter: set · Esc: back"))
			return strings.TrimRight(b.String(), "\n")
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
