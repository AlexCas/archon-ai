package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/archon-ai/archon/internal/config"
)

// graphifyTabState holds the state for the Graphify (code graph) tab.
// focus: 0 = enabled toggle, 1 = auto-install toggle, 2 = semantic toggle,
// 3 = version input, 4 = output dir input.
type graphifyTabState struct {
	enabled     bool
	autoInstall bool
	semantic    bool
	version     textinput.Model
	outputDir   textinput.Model
	focused     int
}

const graphifyFocusCount = 5

func newGraphifyTabState(cfg config.Graphify) graphifyTabState {
	version := textinput.New()
	version.Placeholder = "v0.9.45 (default)"
	version.SetValue(cfg.Version)
	version.Width = 30

	outputDir := textinput.New()
	outputDir.Placeholder = ".archon/graphify (default)"
	outputDir.SetValue(cfg.OutputDir)
	outputDir.Width = 30

	return graphifyTabState{
		enabled:     cfg.Enabled,
		autoInstall: cfg.AutoInstall,
		semantic:    cfg.Semantic,
		version:     version,
		outputDir:   outputDir,
		focused:     0,
	}
}

func (p *graphifyTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			p.focused = (p.focused - 1 + graphifyFocusCount) % graphifyFocusCount
			p.refocus()
			return nil, true
		case tea.KeyDown:
			p.focused = (p.focused + 1) % graphifyFocusCount
			p.refocus()
			return nil, true
		case tea.KeyEnter, tea.KeySpace:
			switch p.focused {
			case 0:
				p.enabled = !p.enabled
				return nil, true
			case 1:
				p.autoInstall = !p.autoInstall
				return nil, true
			case 2:
				p.semantic = !p.semantic
				return nil, true
			}
		}

		// Forward text editing to the focused input.
		var cmd tea.Cmd
		switch p.focused {
		case 3:
			p.version, cmd = p.version.Update(msg)
		case 4:
			p.outputDir, cmd = p.outputDir.Update(msg)
		}
		return cmd, true
	}

	return nil, true
}

func (p *graphifyTabState) refocus() {
	p.version.Blur()
	p.outputDir.Blur()
	switch p.focused {
	case 3:
		p.version.Focus()
	case 4:
		p.outputDir.Focus()
	}
}

func (p *graphifyTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Graphify (Code Graph) Configuration"))
	b.WriteString("\n\n")

	enabledStyle := lipgloss.NewStyle()
	if p.focused == 0 {
		enabledStyle = enabledStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}
	enabledStatus := "OFF"
	if p.enabled {
		enabledStatus = "ON"
	}
	b.WriteString(enabledStyle.Render(fmt.Sprintf("[%s] Enabled (press Enter to toggle)", enabledStatus)))
	b.WriteString("\n\n")

	autoInstallStyle := lipgloss.NewStyle()
	if p.focused == 1 {
		autoInstallStyle = autoInstallStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}
	autoInstallStatus := "OFF"
	if p.autoInstall {
		autoInstallStatus = "ON"
	}
	b.WriteString(autoInstallStyle.Render(fmt.Sprintf("[%s] Auto-Install (press Enter to toggle)", autoInstallStatus)))
	b.WriteString("\n\n")

	semanticStyle := lipgloss.NewStyle()
	if p.focused == 2 {
		semanticStyle = semanticStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}
	semanticStatus := "OFF"
	if p.semantic {
		semanticStatus = "ON"
	}
	b.WriteString(semanticStyle.Render(fmt.Sprintf("[%s] Semantic (press Enter to toggle)", semanticStatus)))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Width(14).Align(lipgloss.Right).MarginRight(1)
	b.WriteString(labelStyle.Render("Version:"))
	b.WriteString(p.version.View())
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Output dir:"))
	b.WriteString(p.outputDir.View())
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(infoStyle.Render("Graphify is advisory-only (non-blocking). When"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("enabled, sdd-explore consults the code graph."))

	return b.String()
}

func (p *graphifyTabState) applyToConfig(cfg *config.Config) {
	cfg.Graphify.Enabled = p.enabled
	cfg.Graphify.AutoInstall = p.autoInstall
	cfg.Graphify.Semantic = p.semantic
	v := strings.TrimSpace(p.version.Value())
	if v == "" {
		v = config.DefaultGraphifyVersion
	}
	cfg.Graphify.Version = v
	o := strings.TrimSpace(p.outputDir.Value())
	if o == "" {
		o = config.DefaultGraphifyOutputDir
	}
	cfg.Graphify.OutputDir = o
}

func (p *graphifyTabState) setWidth(width int) {
	w := width - 20
	if w < 10 {
		w = 10
	}
	p.version.Width = w
	p.outputDir.Width = w
}
