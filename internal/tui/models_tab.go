package tui

import (
	"fmt"
	"strings"

	"github.com/archon-ai/archon/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// modelsTabState holds the state for the Models configuration tab.
type modelsTabState struct {
	inputs        []textinput.Model
	focusedInput  int
	phaseNames    []string
	autoFillLocks map[int]bool // tracks which phase inputs have been manually edited
	// leaderEnabled is true for opencode projects; when set, the leader-model
	// input is appended to inputs and rendered/focus-traversed. The live value
	// lives in the appended inputs entry (see leaderInputIndex), never a separate
	// field, so it cannot desync.
	leaderEnabled bool
}

// modelInputIndices maps input indices to their purpose.
const (
	modelInputDefault = 0
	modelInputExplore = 1
	modelInputPropose = 2
	modelInputSpec    = 3
	modelInputDesign  = 4
	modelInputTasks   = 5
	modelInputApply   = 6
	modelInputVerify  = 7
	modelInputArchive = 8
)

func newModelsTabState(cfg *config.Config) modelsTabState {
	phaseNames := []string{
		"explore", "propose", "spec", "design",
		"tasks", "apply", "verify", "archive",
	}

	// Create inputs: 1 for default + 1 per phase
	inputs := make([]textinput.Model, 1+len(phaseNames))

	// Default model input
	inputs[modelInputDefault] = newModelInput("Default model", cfg.Models.Default)

	// Phase inputs
	for i, phase := range phaseNames {
		idx := i + 1
		value := ""
		if cfg.Models.Phases != nil {
			value = cfg.Models.Phases[phase]
		}
		inputs[idx] = newModelInput(phase, value)
	}

	inputs[modelInputDefault].Focus()

	// Leader input is appended to the focus ring only for opencode projects so
	// the TUI can write the project opencode.json agent.archon-leader. For other
	// agents it is constructed but never rendered or focus-traversed.
	leaderEnabled := cfg.Agent == "opencode"
	if leaderEnabled {
		inputs = append(inputs, newModelInput("Leader model", cfg.Models.Leader))
	}

	state := modelsTabState{
		inputs:        inputs,
		focusedInput:  modelInputDefault,
		phaseNames:    phaseNames,
		autoFillLocks: make(map[int]bool),
		leaderEnabled: leaderEnabled,
	}

	// Update auto-fill placeholders
	state.updateAutoFill()

	return state
}

// leaderInputIndex returns the index of the leader input within m.inputs, or -1
// when the leader field is disabled (non-opencode agent). The leader input is
// always the final element of m.inputs when enabled.
func (m *modelsTabState) leaderInputIndex() int {
	if !m.leaderEnabled {
		return -1
	}
	return len(m.inputs) - 1
}

func newModelInput(placeholder, value string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	ti.Width = 30
	return ti
}

func (m *modelsTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n":
			m.cycleStaticModel(1)
			return nil, true
		case "ctrl+p":
			m.cycleStaticModel(-1)
			return nil, true
		}
		switch msg.Type {
		case tea.KeyUp:
			m.focusPrev()
			return nil, true
		case tea.KeyDown, tea.KeyTab:
			m.focusNext()
			return nil, true
		case tea.KeyEnter:
			// Lock auto-fill for current input if user pressed enter
			m.autoFillLocks[m.focusedInput] = true
			return nil, true
		}
	}

	// Update the focused input
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)

	// If user typed something, lock auto-fill for this input
	if m.focusedInput > modelInputDefault {
		currentValue := m.inputs[m.focusedInput].Value()
		if currentValue != "" {
			m.autoFillLocks[m.focusedInput] = true
		}
	}

	// If default model changed and this is the default input, update auto-fill
	if m.focusedInput == modelInputDefault {
		m.updateAutoFill()
	}

	return cmd, true
}

// cycleStaticModel sets the focused input to the next/previous model in the
// static catalog (Claude + Opencode Go). An empty entry leads the catalog so
// users can cycle back to "no value" and still type a free-form model name.
func (m *modelsTabState) cycleStaticModel(dir int) {
	catalog := append([]string{""}, config.StaticModels()...)
	current := m.inputs[m.focusedInput].Value()

	idx := 0
	for i, name := range catalog {
		if name == current {
			idx = i
			break
		}
	}

	idx = (idx + dir + len(catalog)) % len(catalog)
	m.inputs[m.focusedInput].SetValue(catalog[idx])
	m.autoFillLocks[m.focusedInput] = true
	if m.focusedInput == modelInputDefault {
		m.updateAutoFill()
	}
}

func (m *modelsTabState) focusNext() {
	m.inputs[m.focusedInput].Blur()
	m.focusedInput = (m.focusedInput + 1) % len(m.inputs)
	m.inputs[m.focusedInput].Focus()
}

func (m *modelsTabState) focusPrev() {
	m.inputs[m.focusedInput].Blur()
	m.focusedInput = (m.focusedInput - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focusedInput].Focus()
}

func (m *modelsTabState) updateAutoFill() {
	defaultValue := m.inputs[modelInputDefault].Value()

	// Update all phase inputs that are not locked
	for i := range m.phaseNames {
		idx := i + 1
		if !m.autoFillLocks[idx] {
			// Show default as placeholder if empty
			m.inputs[idx].Placeholder = fmt.Sprintf("%s (default: %s)", m.phaseNames[i], defaultValue)
		}
	}
}

func (m *modelsTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Model Configuration"))
	b.WriteString("\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(hintStyle.Render("ctrl+n/ctrl+p: cycle static models  ·  or type any model name"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Static: " + strings.Join(config.StaticModels(), ", ")))
	b.WriteString("\n\n")

	// Default model
	labelStyle := lipgloss.NewStyle().
		Width(15).
		Align(lipgloss.Right).
		MarginRight(1)

	inputStyle := lipgloss.NewStyle()

	b.WriteString(labelStyle.Render("Default:"))
	b.WriteString(inputStyle.Render(m.inputs[modelInputDefault].View()))
	b.WriteString("\n\n")

	// Phase models
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Per-Phase Models:"))
	b.WriteString("\n\n")

	for i, phase := range m.phaseNames {
		idx := i + 1
		b.WriteString(labelStyle.Render(phase + ":"))
		b.WriteString(inputStyle.Render(m.inputs[idx].View()))

		// Validation warning
		value := m.inputs[idx].Value()
		if value == "" {
			value = m.inputs[modelInputDefault].Value()
		}
		if value != "" {
			if warning := config.Validate(value); warning != "" {
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")). // yellow
					MarginLeft(16)
				b.WriteString("\n")
				b.WriteString(warningStyle.Render("⚠ " + warning))
			}
		}

		b.WriteString("\n")
	}

	// Leader model (opencode only). Written verbatim into opencode.json as the
	// archon-leader primary agent model.
	if idx := m.leaderInputIndex(); idx >= 0 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Leader Model (opencode):"))
		b.WriteString("\n\n")
		b.WriteString(labelStyle.Render("Leader:"))
		b.WriteString(inputStyle.Render(m.inputs[idx].View()))

		// The leader model is an opencode provider/model-id (e.g.
		// "openai/gpt-4o") rather than a Claude family alias, so only run the
		// Claude-oriented advisory check when the value is not already in
		// provider/model-id form (contains "/"). Mirrors the `--leader` CLI flag.
		if value := m.inputs[idx].Value(); value != "" && !strings.Contains(value, "/") {
			if warning := config.Validate(value); warning != "" {
				warningStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")). // yellow
					MarginLeft(16)
				b.WriteString("\n")
				b.WriteString(warningStyle.Render("⚠ " + warning))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *modelsTabState) applyToConfig(cfg *config.Config) {
	cfg.Models.Default = m.inputs[modelInputDefault].Value()

	if cfg.Models.Phases == nil {
		cfg.Models.Phases = make(map[string]string)
	}

	for i, phase := range m.phaseNames {
		idx := i + 1
		value := m.inputs[idx].Value()
		if value != "" {
			cfg.Models.Phases[phase] = value
		} else {
			delete(cfg.Models.Phases, phase)
		}
	}

	// Leader model is opencode-only; for other agents the input is not in the
	// focus ring and we leave cfg.Models.Leader as loaded.
	if idx := m.leaderInputIndex(); idx >= 0 {
		cfg.Models.Leader = m.inputs[idx].Value()
	}
}

func (m *modelsTabState) setWidth(width int) {
	for i := range m.inputs {
		m.inputs[i].Width = width - 20
	}
}
