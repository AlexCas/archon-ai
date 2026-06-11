package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/initcmd"
	"github.com/archon-ai/archon/skills"
	"golang.org/x/term"
)

type Tab int

const (
	ModelsTab Tab = iota
	MutationTab
	AgentTab
)

type Model struct {
	config     *config.Config
	projectDir string
	activeTab  Tab
	height     int
	width      int
	quitting   bool
	statusMsg  string
	statusErr  bool
	// Tab states
	modelsTab   modelsTabState
	mutationTab mutationTabState
	agentTab    agentTabState
}

type keyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Save     key.Binding
	Quit     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Save, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.ShiftTab},
		{k.Save, k.Quit},
	}
}

var defaultKeys = keyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q", "q"),
		key.WithHelp("ctrl+q", "quit"),
	),
}

func NewModel(cfg *config.Config, projectDir string) Model {
	return Model{
		config:      cfg,
		projectDir:  projectDir,
		activeTab:   ModelsTab,
		modelsTab:   newModelsTabState(cfg),
		mutationTab: newMutationTabState(cfg.MutationTesting),
		agentTab:    newAgentTabState(cfg.Agent),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.modelsTab.setWidth(m.width)
		m.mutationTab.setWidth(m.width)
		m.agentTab.setWidth(m.width)

	case tea.KeyMsg:
		// Global keys
		switch {
		case key.Matches(msg, defaultKeys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, defaultKeys.Tab):
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil

		case key.Matches(msg, defaultKeys.ShiftTab):
			m.activeTab = (m.activeTab + 2) % 3
			return m, nil

		case key.Matches(msg, defaultKeys.Save):
			cmds = append(cmds, m.saveConfig())
			return m, tea.Batch(cmds...)
		}

		// Tab-specific keys
		switch m.activeTab {
		case ModelsTab:
			cmd, _ := m.modelsTab.update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case MutationTab:
			cmd, _ := m.mutationTab.update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case AgentTab:
			cmd, _ := m.agentTab.update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case agentInitMsg:
		cmd := m.runAgentInit(msg.agent)
		cmds = append(cmds, cmd)

	case string:
		// Status messages
		m.statusMsg = msg
		m.statusErr = false
		return m, tea.Batch(cmds...)

	case error:
		m.statusMsg = msg.Error()
		m.statusErr = true
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content string

	// Tab headers
	content = m.renderTabs()
	content += "\n"

	// Tab content
	content += m.renderTabContent()
	content += "\n"

	// Status message
	if m.statusMsg != "" {
		content += m.renderStatus() + "\n"
	}

	// Help footer
	content += m.renderHelp()

	return content
}

func (m Model) renderTabs() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	activeStyle := style.Copy().
		BorderForeground(lipgloss.Color("63")).
		Bold(true)

	tabs := []string{"Models", "Mutation Testing", "Agent"}
	var rendered []string

	for i, name := range tabs {
		if Tab(i) == m.activeTab {
			rendered = append(rendered, activeStyle.Render(name))
		} else {
			rendered = append(rendered, style.Render(name))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderTabContent() string {
	style := lipgloss.NewStyle().
		Width(m.width - 2).
		Height(m.height - 8).
		Padding(1, 2)

	switch m.activeTab {
	case ModelsTab:
		return style.Render(m.modelsTab.view(m.width, m.height))
	case MutationTab:
		return style.Render(m.mutationTab.view(m.width, m.height))
	case AgentTab:
		return style.Render(m.agentTab.view(m.width, m.height))
	default:
		return style.Render("Unknown tab")
	}
}

func (m Model) renderHelp() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	return style.Render("tab: next | shift+tab: prev | ctrl+s: save | ctrl+q: quit")
}

func (m Model) renderStatus() string {
	style := lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)

	if m.statusErr {
		style = style.Foreground(lipgloss.Color("196"))
	} else {
		style = style.Foreground(lipgloss.Color("82"))
	}

	return style.Render(m.statusMsg)
}

func (m Model) saveConfig() tea.Cmd {
	return func() tea.Msg {
		// Apply tab states to config
		m.modelsTab.applyToConfig(m.config)
		m.mutationTab.applyToConfig(m.config)
		m.agentTab.applyToConfig(m.config)

		if err := m.config.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		return "✓ Configuration saved"
	}
}

func (m Model) runAgentInit(agent string) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			m.agentTab.setInitResult(fmt.Sprintf("Error: %v", err), true)
			return fmt.Errorf("get home directory: %w", err)
		}

		opts := initcmd.Options{
			HomeDir:    homeDir,
			ProjectDir: m.projectDir,
			Agent:      agent,
			Force:      true,
			EmbeddedFS: skills.FS,
		}

		result, err := initcmd.Run(opts)
		if err != nil {
			m.agentTab.setInitResult(fmt.Sprintf("Init failed: %v", err), true)
			return fmt.Errorf("agent init: %w", err)
		}

		m.agentTab.setInitResult(
			fmt.Sprintf("✓ Agent initialized: %s (%d skills)", result.Agent, result.ExtractedCount),
			false)
		return fmt.Sprintf("Agent %s initialized successfully", result.Agent)
	}
}

// CheckTerminal verifies the current process is running in a terminal.
func CheckTerminal() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("not a terminal: archon tui requires an interactive terminal")
	}
	return nil
}
