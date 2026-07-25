package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/archon-ai/archon/internal/config"
)

// impeccableTabState holds the state for the Impeccable (design language) tab.
// focus: 0 = enabled toggle, 1 = auto-install toggle, 2 = severity input,
// 3 = product path input, 4 = design path input.
type impeccableTabState struct {
	enabled     bool
	autoInstall bool
	severity    textinput.Model
	productPath textinput.Model
	designPath  textinput.Model
	focused     int
}

const impeccableFocusCount = 5

func newImpeccableTabState(cfg config.Impeccable) impeccableTabState {
	severity := textinput.New()
	severity.Placeholder = "block-deterministic (default)"
	severity.SetValue(cfg.Severity)
	severity.Width = 30

	productPath := textinput.New()
	productPath.Placeholder = "PRODUCT.md (default: project root)"
	productPath.SetValue(cfg.ProductPath)
	productPath.Width = 30

	designPath := textinput.New()
	designPath.Placeholder = "DESIGN.md (default: project root)"
	designPath.SetValue(cfg.DesignPath)
	designPath.Width = 30

	return impeccableTabState{
		enabled:     cfg.Enabled,
		autoInstall: cfg.AutoInstall,
		severity:    severity,
		productPath: productPath,
		designPath:  designPath,
		focused:     0,
	}
}

func (p *impeccableTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			p.focused = (p.focused - 1 + impeccableFocusCount) % impeccableFocusCount
			p.refocus()
			return nil, true
		case tea.KeyDown:
			p.focused = (p.focused + 1) % impeccableFocusCount
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
			}
		}

		// Forward text editing to the focused input.
		var cmd tea.Cmd
		switch p.focused {
		case 2:
			p.severity, cmd = p.severity.Update(msg)
		case 3:
			p.productPath, cmd = p.productPath.Update(msg)
		case 4:
			p.designPath, cmd = p.designPath.Update(msg)
		}
		return cmd, true
	}

	return nil, true
}

func (p *impeccableTabState) refocus() {
	p.severity.Blur()
	p.productPath.Blur()
	p.designPath.Blur()
	switch p.focused {
	case 2:
		p.severity.Focus()
	case 3:
		p.productPath.Focus()
	case 4:
		p.designPath.Focus()
	}
}

func (p *impeccableTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Impeccable (Design Language) Configuration"))
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

	labelStyle := lipgloss.NewStyle().Width(14).Align(lipgloss.Right).MarginRight(1)
	b.WriteString(labelStyle.Render("Severity:"))
	b.WriteString(p.severity.View())
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Product path:"))
	b.WriteString(p.productPath.View())
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Design path:"))
	b.WriteString(p.designPath.View())
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(infoStyle.Render("When enabled, runs 'npx impeccable detect' after judge; severity"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("governs blocking."))

	return b.String()
}

func (p *impeccableTabState) applyToConfig(cfg *config.Config) {
	cfg.Impeccable.Enabled = p.enabled
	cfg.Impeccable.AutoInstall = p.autoInstall
	severity := strings.TrimSpace(p.severity.Value())
	// Normalize blank input and reject invalid values: fall back to the safe
	// default so the written config always passes config.Load() validation.
	if severity == "" || config.ValidateImpeccableSeverity(severity) != nil {
		severity = "block-deterministic"
	}
	cfg.Impeccable.Severity = severity
	cfg.Impeccable.ProductPath = strings.TrimSpace(p.productPath.Value())
	cfg.Impeccable.DesignPath = strings.TrimSpace(p.designPath.Value())
}

func (p *impeccableTabState) setWidth(width int) {
	w := width - 20
	if w < 10 {
		w = 10
	}
	p.severity.Width = w
	p.productPath.Width = w
	p.designPath.Width = w
}
