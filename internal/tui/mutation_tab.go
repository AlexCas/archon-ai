package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/archon-ai/archon/internal/config"
)

// mutationTabState holds the state for the Mutation Testing configuration tab.
type mutationTabState struct {
	enabled bool
	threshold Slider
	focused   int // 0 = toggle, 1 = slider
}

func newMutationTabState(cfg config.MutationTesting) mutationTabState {
	return mutationTabState{
		enabled:   cfg.Enabled,
		threshold: NewSlider("Threshold", FromFloat(cfg.Threshold)),
		focused:   0,
	}
}

func (m *mutationTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			m.focused = 1 - m.focused
			m.updateFocus()
			return nil, true

		case tea.KeyLeft:
			if m.focused == 1 {
				m.threshold.Dec()
			}
			return nil, true

		case tea.KeyRight:
			if m.focused == 1 {
				m.threshold.Inc()
			}
			return nil, true

		case tea.KeyEnter, tea.KeySpace:
			if m.focused == 0 {
				m.enabled = !m.enabled
			}
			return nil, true
		}
	}

	return nil, true
}

func (m *mutationTabState) updateFocus() {
	m.threshold.Focused = (m.focused == 1)
}

func (m *mutationTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Mutation Testing Configuration"))
	b.WriteString("\n\n")

	// Toggle
	toggleStyle := lipgloss.NewStyle()
	if m.focused == 0 {
		toggleStyle = toggleStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}

	status := "OFF"
	if m.enabled {
		status = "ON"
	}

	b.WriteString(toggleStyle.Render(fmt.Sprintf("[%s] Enabled (press Enter to toggle)", status)))
	b.WriteString("\n\n")

	// Slider
	m.threshold.Width = width - 30
	b.WriteString(m.threshold.View())
	b.WriteString("\n")

	// Info
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	b.WriteString(infoStyle.Render("Mutation testing will run after unit tests to verify test quality."))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("Threshold: %.0f%% of mutants must be killed.", m.threshold.AsFloat()*100)))

	return b.String()
}

func (m *mutationTabState) applyToConfig(cfg *config.Config) {
	cfg.MutationTesting.Enabled = m.enabled
	cfg.MutationTesting.Threshold = m.threshold.AsFloat()
}

func (m *mutationTabState) setWidth(width int) {
	m.threshold.Width = width - 30
}
