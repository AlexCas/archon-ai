package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/archon-ai/archon/internal/config"
)

// playwrightTabState holds the state for the Playwright (web E2E) tab.
// focus: 0 = enabled toggle, 1 = test dir input, 2 = base url input.
type playwrightTabState struct {
	enabled bool
	testDir textinput.Model
	baseURL textinput.Model
	focused int
}

const playwrightFocusCount = 3

func newPlaywrightTabState(cfg config.Playwright) playwrightTabState {
	testDir := textinput.New()
	testDir.Placeholder = "e2e (default)"
	testDir.SetValue(cfg.TestDir)
	testDir.Width = 30

	baseURL := textinput.New()
	baseURL.Placeholder = "http://localhost:3000"
	baseURL.SetValue(cfg.BaseURL)
	baseURL.Width = 30

	return playwrightTabState{
		enabled: cfg.Enabled,
		testDir: testDir,
		baseURL: baseURL,
		focused: 0,
	}
}

func (p *playwrightTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			p.focused = (p.focused - 1 + playwrightFocusCount) % playwrightFocusCount
			p.refocus()
			return nil, true
		case tea.KeyDown:
			p.focused = (p.focused + 1) % playwrightFocusCount
			p.refocus()
			return nil, true
		case tea.KeyEnter, tea.KeySpace:
			if p.focused == 0 {
				p.enabled = !p.enabled
				return nil, true
			}
		}

		// Forward text editing to the focused input.
		var cmd tea.Cmd
		switch p.focused {
		case 1:
			p.testDir, cmd = p.testDir.Update(msg)
		case 2:
			p.baseURL, cmd = p.baseURL.Update(msg)
		}
		return cmd, true
	}

	return nil, true
}

func (p *playwrightTabState) refocus() {
	p.testDir.Blur()
	p.baseURL.Blur()
	switch p.focused {
	case 1:
		p.testDir.Focus()
	case 2:
		p.baseURL.Focus()
	}
}

func (p *playwrightTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Playwright (Web E2E) Configuration"))
	b.WriteString("\n\n")

	toggleStyle := lipgloss.NewStyle()
	if p.focused == 0 {
		toggleStyle = toggleStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}
	status := "OFF"
	if p.enabled {
		status = "ON"
	}
	b.WriteString(toggleStyle.Render(fmt.Sprintf("[%s] Enabled (press Enter to toggle)", status)))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Width(12).Align(lipgloss.Right).MarginRight(1)
	b.WriteString(labelStyle.Render("Test dir:"))
	b.WriteString(p.testDir.View())
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Base URL:"))
	b.WriteString(p.baseURL.View())
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(infoStyle.Render("When enabled, the harness generates Playwright specs from Gherkin"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("scenarios and runs them after the verify and judge phases."))

	return b.String()
}

func (p *playwrightTabState) applyToConfig(cfg *config.Config) {
	cfg.Playwright.Enabled = p.enabled
	cfg.Playwright.TestDir = strings.TrimSpace(p.testDir.Value())
	cfg.Playwright.BaseURL = strings.TrimSpace(p.baseURL.Value())
}

func (p *playwrightTabState) setWidth(width int) {
	w := width - 20
	if w < 10 {
		w = 10
	}
	p.testDir.Width = w
	p.baseURL.Width = w
}
