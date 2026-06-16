package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/archon-ai/archon/internal/config"
)

// judgeTabState holds the state for the Judge phase configuration tab.
// The tab exposes a single toggle that controls whether the judge phase runs.
type judgeTabState struct {
	enabled bool
}

func newJudgeTabState(cfg config.Judge) judgeTabState {
	return judgeTabState{enabled: cfg.Enabled}
}

func (j *judgeTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeySpace:
			j.enabled = !j.enabled
			return nil, true
		}
	}
	return nil, true
}

func (j *judgeTabState) view(width, height int) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("Judge Phase Configuration"))
	b.WriteString("\n\n")

	// Single toggle; it is always the focused control on this tab.
	toggleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	status := "OFF"
	if j.enabled {
		status = "ON"
	}
	b.WriteString(toggleStyle.Render(fmt.Sprintf("[%s] Enabled (press Enter to toggle)", status)))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(infoStyle.Render("When enabled, the harness runs the judge phase after verify:"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("dual adversarial review (judgment-day) plus any enabled gates."))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("When disabled, the workflow goes from verify straight to archive."))

	return b.String()
}

func (j *judgeTabState) applyToConfig(cfg *config.Config) {
	cfg.Judge.Enabled = j.enabled
}

// setWidth is a no-op: the judge tab has no width-sensitive controls. It exists
// to satisfy the same tab-state shape used by the other tabs.
func (j *judgeTabState) setWidth(width int) {}
