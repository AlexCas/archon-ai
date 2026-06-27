package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/archon-ai/archon/internal/config"
)

// securityProfiles enumerates the two valid security profiles in cycle order.
// The cycle-selector can only ever yield one of these values, making it
// impossible to select an invalid profile via the TUI.
var securityProfiles = []string{"cli", "web"}

// securityTabState holds the state for the Security configuration tab.
// focus: 0 = enabled toggle, 1 = profile cycle-selector.
type securityTabState struct {
	enabled      bool
	profileIndex int // index into securityProfiles; always 0 ("cli") or 1 ("web")
	focused      int
}

const securityFocusCount = 2

// newSecurityTabState initialises the tab from the current config.
// When Profile is empty or unset the cycle defaults to "cli" (index 0).
func newSecurityTabState(cfg config.Security) securityTabState {
	idx := 0
	for i, p := range securityProfiles {
		if p == cfg.Profile {
			idx = i
			break
		}
	}
	return securityTabState{
		enabled:      cfg.Enabled,
		profileIndex: idx,
		focused:      0,
	}
}

// profile returns the currently selected profile string.
func (s *securityTabState) profile() string {
	return securityProfiles[s.profileIndex]
}

func (s *securityTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			s.focused = (s.focused - 1 + securityFocusCount) % securityFocusCount
			return nil, true
		case tea.KeyDown:
			s.focused = (s.focused + 1) % securityFocusCount
			return nil, true
		case tea.KeyEnter, tea.KeySpace:
			switch s.focused {
			case 0:
				// Toggle enabled.
				s.enabled = !s.enabled
				return nil, true
			case 1:
				// Advance the cycle: cli → web → cli.
				s.profileIndex = (s.profileIndex + 1) % len(securityProfiles)
				return nil, true
			}
		}
	}
	return nil, true
}

func (s *securityTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Security Baseline Configuration"))
	b.WriteString("\n\n")

	// Enabled toggle row.
	toggleStyle := lipgloss.NewStyle()
	if s.focused == 0 {
		toggleStyle = toggleStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}
	status := "OFF"
	if s.enabled {
		status = "ON"
	}
	b.WriteString(toggleStyle.Render(fmt.Sprintf("[%s] Enabled (press Enter to toggle)", status)))
	b.WriteString("\n\n")

	// Profile cycle-selector row.
	profileStyle := lipgloss.NewStyle()
	if s.focused == 1 {
		profileStyle = profileStyle.Foreground(lipgloss.Color("63")).Bold(true)
	}
	labelStyle := lipgloss.NewStyle().Width(10).Align(lipgloss.Right).MarginRight(1)
	b.WriteString(labelStyle.Render("Profile:"))
	b.WriteString(profileStyle.Render(fmt.Sprintf("[ %s ] (press Enter to cycle: cli → web)", s.profile())))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(infoStyle.Render("When enabled, the harness injects security-review hooks across the"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("propose, spec, tasks, verify, and judge phases."))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("Profiles: cli (argument/path injection, secrets, deps) |"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("          web (cli controls + OWASP Top 10)."))

	return b.String()
}

func (s *securityTabState) applyToConfig(cfg *config.Config) {
	cfg.Security.Enabled = s.enabled
	cfg.Security.Profile = s.profile()
}

// setWidth is a no-op: the security tab has no width-sensitive text inputs. It
// exists to satisfy the same tab-state shape used by the other tabs.
func (s *securityTabState) setWidth(width int) {}
