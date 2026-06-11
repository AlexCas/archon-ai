package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/archon-ai/archon/internal/config"
)

var availableAgents = []string{"opencode", "claude", "codex", "agents"}

// agentTabState holds the state for the Agent configuration tab.
type agentTabState struct {
	selectedAgent   string
	focusedIndex    int
	confirmingInit  bool
	initResult      string
	initError       bool
}

func newAgentTabState(currentAgent string) agentTabState {
	return agentTabState{
		selectedAgent: currentAgent,
		focusedIndex:  -1, // -1 means no selection focused
	}
}

func (a *agentTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.confirmingInit {
			switch msg.Type {
			case tea.KeyRunes:
				if len(msg.Runes) == 1 {
					switch msg.Runes[0] {
					case 'y', 'Y':
						return a.triggerInit(), true
					case 'n', 'N':
						a.confirmingInit = false
						return nil, true
					}
				}
			case tea.KeyEnter:
				return a.triggerInit(), true
			case tea.KeyEsc:
				a.confirmingInit = false
				return nil, true
			}
			return nil, true
		}

		switch msg.Type {
		case tea.KeyUp:
			if a.focusedIndex > 0 {
				a.focusedIndex--
			}
			return nil, true
		case tea.KeyDown:
			if a.focusedIndex < len(availableAgents)-1 {
				a.focusedIndex++
			}
			if a.focusedIndex < 0 {
				a.focusedIndex = 0
			}
			return nil, true
		case tea.KeyEnter:
			if a.focusedIndex >= 0 && a.focusedIndex < len(availableAgents) {
				selected := availableAgents[a.focusedIndex]
				if selected != a.selectedAgent {
					a.selectedAgent = selected
					a.confirmingInit = true
				}
			}
			return nil, true
		}
	}

	return nil, true
}

func (a *agentTabState) triggerInit() tea.Cmd {
	agent := a.selectedAgent
	a.confirmingInit = false

	return func() tea.Msg {
		// Return a message to trigger init in the main model
		return agentInitMsg{agent: agent}
	}
}

type agentInitMsg struct {
	agent string
}

func (a *agentTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Agent Configuration"))
	b.WriteString("\n\n")

	if a.confirmingInit {
		return a.renderConfirmDialog()
	}

	// Current agent
	infoStyle := lipgloss.NewStyle().
		MarginBottom(1)

	b.WriteString(infoStyle.Render(fmt.Sprintf("Current agent: %s", a.selectedAgent)))
	b.WriteString("\n\n")

	// Agent list
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Available Agents:"))
	b.WriteString("\n\n")

	for i, agent := range availableAgents {
		style := lipgloss.NewStyle().
			Padding(0, 1).
			MarginLeft(2)

		if i == a.focusedIndex {
			style = style.
				Background(lipgloss.Color("63")).
				Foreground(lipgloss.Color("0"))
		}

		marker := "  "
		if agent == a.selectedAgent {
			marker = "▸ "
		}

		b.WriteString(style.Render(fmt.Sprintf("%s%s", marker, agent)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: navigate | Enter: select | Changes require re-initialization"))

	// Show init result if any
	if a.initResult != "" {
		b.WriteString("\n\n")
		resultStyle := lipgloss.NewStyle()
		if a.initError {
			resultStyle = resultStyle.Foreground(lipgloss.Color("196"))
		} else {
			resultStyle = resultStyle.Foreground(lipgloss.Color("82"))
		}
		b.WriteString(resultStyle.Render(a.initResult))
	}

	return b.String()
}

func (a *agentTabState) renderConfirmDialog() string {
	var b strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(50)

	b.WriteString(boxStyle.Render(
		fmt.Sprintf("⚠ Re-initialization Required\n\n"+
			"Switching to %s requires re-running `archon init`.\n\n"+
			"This will recreate agent files and symlinks.\n\n"+
			"Proceed? [y/N]", a.selectedAgent)))

	return b.String()
}

func (a *agentTabState) applyToConfig(cfg *config.Config) {
	// Agent is applied via re-init, not direct config change
	cfg.Agent = a.selectedAgent
}

func (a *agentTabState) setInitResult(result string, isError bool) {
	a.initResult = result
	a.initError = isError
}

func (a *agentTabState) setWidth(width int) {
	// No-op for agent tab
}
