package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/initcmd"
	"github.com/archon-ai/archon/internal/models"
	"github.com/archon-ai/archon/skills"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Tab int

const (
	ModelsTab Tab = iota
	JudgeTab
	MutationTab
	PlaywrightTab
	AgentTab
	tabCount
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
	modelsTab     modelsTabState
	judgeTab      judgeTabState
	mutationTab   mutationTabState
	playwrightTab playwrightTabState
	agentTab      agentTabState
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
	// Detect the offered model catalog once when the TUI opens. Detection must
	// not run per keystroke nor during "archon init"; the cached slice is reused
	// for the lifetime of the Models view.
	catalog := models.Resolve()
	return Model{
		config:        cfg,
		projectDir:    projectDir,
		activeTab:     ModelsTab,
		modelsTab:     newModelsTabState(cfg, catalog),
		judgeTab:      newJudgeTabState(cfg.Judge),
		mutationTab:   newMutationTabState(cfg.MutationTesting),
		playwrightTab: newPlaywrightTabState(cfg.Playwright),
		agentTab:      newAgentTabState(cfg.Agent),
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
		m.judgeTab.setWidth(m.width)
		m.mutationTab.setWidth(m.width)
		m.playwrightTab.setWidth(m.width)
		m.agentTab.setWidth(m.width)

	case tea.KeyMsg:
		// Global keys
		switch {
		case key.Matches(msg, defaultKeys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, defaultKeys.Tab):
			m.activeTab = (m.activeTab + 1) % tabCount
			return m, nil

		case key.Matches(msg, defaultKeys.ShiftTab):
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount
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
		case JudgeTab:
			cmd, _ := m.judgeTab.update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case MutationTab:
			cmd, _ := m.mutationTab.update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case PlaywrightTab:
			cmd, _ := m.playwrightTab.update(msg)
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
		cmd := m.runAgentInit(msg.agent, msg.overwrite)
		cmds = append(cmds, cmd)

	case templateExistsMsg:
		// init found an existing orchestrator file; ask before replacing it.
		m.agentTab.askOverwrite(msg.agent, msg.path)
		return m, tea.Batch(cmds...)

	case agentInitDoneMsg:
		// Adopt the freshly written config and rebuild tab states from it.
		msg.cfg.HomeDir = m.projectDir
		m.config = msg.cfg
		// Reuse the catalog detected when the view first opened; init does not
		// re-run detection.
		m.modelsTab = newModelsTabState(msg.cfg, m.modelsTab.catalog)
		m.judgeTab = newJudgeTabState(msg.cfg.Judge)
		m.mutationTab = newMutationTabState(msg.cfg.MutationTesting)
		m.playwrightTab = newPlaywrightTabState(msg.cfg.Playwright)
		m.agentTab = newAgentTabState(msg.cfg.Agent)
		if m.width > 0 {
			m.modelsTab.setWidth(m.width)
			m.judgeTab.setWidth(m.width)
			m.mutationTab.setWidth(m.width)
			m.playwrightTab.setWidth(m.width)
			m.agentTab.setWidth(m.width)
		}
		m.statusMsg = msg.summary
		m.statusErr = false
		return m, tea.Batch(cmds...)

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

	tabs := []string{"Models", "Judge", "Mutation Testing", "Playwright", "Agent"}
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
		Width(m.width-2).
		Height(m.height-8).
		Padding(1, 2)

	switch m.activeTab {
	case ModelsTab:
		return style.Render(m.modelsTab.view(m.width, m.height))
	case JudgeTab:
		return style.Render(m.judgeTab.view(m.width, m.height))
	case MutationTab:
		return style.Render(m.mutationTab.view(m.width, m.height))
	case PlaywrightTab:
		return style.Render(m.playwrightTab.view(m.width, m.height))
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
		// Clone config before mutation so we can rollback on failure
		cfg := m.config.Clone()
		m.modelsTab.applyToConfig(cfg)
		m.judgeTab.applyToConfig(cfg)
		m.mutationTab.applyToConfig(cfg)
		m.playwrightTab.applyToConfig(cfg)
		m.agentTab.applyToConfig(cfg)

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		// Regenerate orchestrator template to reflect updated config
		if err := regenerateTemplate(m.projectDir, cfg); err != nil {
			return fmt.Errorf("saved config but failed to regenerate orchestrator: %w", err)
		}

		// For opencode projects, merge the archon-leader agent into the project
		// opencode.json using the same writer as init so both paths produce
		// byte-identical output. No-op when models.leader is empty.
		if cfg.Agent == "opencode" {
			if _, err := initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models.Leader.FullID()); err != nil {
				return fmt.Errorf("saved config but failed to merge opencode agent: %w", err)
			}
		}

		// Only update in-memory config after save, regenerate, and merge succeed
		m.config = cfg

		return "✓ Configuration saved"
	}
}

// regenerateTemplate re-renders the orchestrator markdown file (CLAUDE.md or
// AGENTS.md) so it stays in sync with the current config after a save.
func regenerateTemplate(projectDir string, cfg *config.Config) error {
	data := initcmd.TemplateData{
		ProjectName:    filepath.Base(projectDir),
		Agent:          cfg.Agent,
		HarnessVersion: cfg.Version,
		SkillCount:     cfg.SkillCount,
		PhaseModels:    config.ResolvePhaseModels(cfg.Models),
	}

	var content string
	var filename string

	switch cfg.Agent {
	case "claude":
		var err error
		content, err = initcmd.RenderClaudeMD(data)
		if err != nil {
			return fmt.Errorf("render CLAUDE.md: %w", err)
		}
		filename = "CLAUDE.md"
	default:
		var err error
		content, err = initcmd.RenderAgentsMD(data)
		if err != nil {
			return fmt.Errorf("render AGENTS.md: %w", err)
		}
		filename = "AGENTS.md"
	}

	path := filepath.Join(projectDir, filename)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write template: %w", err)
	}

	return os.Rename(tmp, path)
}

func (m Model) runAgentInit(agent string, overwrite bool) tea.Cmd {
	projectDir := m.projectDir
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}

		opts := initcmd.Options{
			HomeDir:           homeDir,
			ProjectDir:        projectDir,
			Agent:             agent,
			Force:             true,
			EmbeddedFS:        skills.FS,
			OverwriteTemplate: overwrite,
		}

		result, err := initcmd.Run(opts)
		if errors.Is(err, initcmd.ErrTemplateExists) {
			// Surface a confirmation request instead of failing.
			return templateExistsMsg{
				agent: agent,
				path:  templateFileName(agent),
			}
		}
		if err != nil {
			return fmt.Errorf("agent init: %w", err)
		}

		// Reload the freshly written config so a later save does not clobber it.
		reloaded := &config.Config{HomeDir: projectDir}
		if err := reloaded.Load(os.DirFS(projectDir)); err != nil {
			return fmt.Errorf("init succeeded but reloading config failed: %w", err)
		}

		return agentInitDoneMsg{
			cfg:     reloaded,
			summary: fmt.Sprintf("✓ Agent initialized: %s (%d skills)", result.Agent, result.ExtractedCount),
		}
	}
}

type agentInitDoneMsg struct {
	cfg     *config.Config
	summary string
}

// templateFileName returns the orchestrator filename for an agent, for display.
func templateFileName(agent string) string {
	if agent == "claude" {
		return "CLAUDE.md"
	}
	return "AGENTS.md"
}

type templateExistsMsg struct {
	agent string
	path  string
}

// CheckTerminal verifies the current process is running in a terminal.
func CheckTerminal() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("not a terminal: archon tui requires an interactive terminal")
	}
	return nil
}
