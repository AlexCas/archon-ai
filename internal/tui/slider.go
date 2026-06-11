package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Slider represents a 0-100% slider component.
type Slider struct {
	Value       int // 0-100
	Width       int
	Label       string
	Focused     bool
}

// NewSlider creates a new slider with the given label and initial value.
func NewSlider(label string, value int) Slider {
	return Slider{
		Value:   value,
		Label:   label,
		Width:   40,
		Focused: false,
	}
}

// SetValue sets the slider value (clamped to 0-100).
func (s *Slider) SetValue(v int) {
	if v < 0 {
		v = 0
	} else if v > 100 {
		v = 100
	}
	s.Value = v
}

// Inc increments the value by 1.
func (s *Slider) Inc() {
	s.SetValue(s.Value + 1)
}

// Dec decrements the value by 1.
func (s *Slider) Dec() {
	s.SetValue(s.Value - 1)
}

// View renders the slider.
func (s Slider) View() string {
	if s.Width < 10 {
		s.Width = 10
	}

	filled := s.Value * s.Width / 100
	empty := s.Width - filled

	barStyle := lipgloss.NewStyle()
	if s.Focused {
		barStyle = barStyle.Foreground(lipgloss.Color("63"))
	}

	filledStyle := barStyle.Copy().Foreground(lipgloss.Color("63"))
	emptyStyle := barStyle.Copy().Foreground(lipgloss.Color("240"))

	filledBar := filledStyle.Render(strings.Repeat("█", filled))
	emptyBar := emptyStyle.Render(strings.Repeat("░", empty))

	labelStyle := lipgloss.NewStyle().
		Width(20).
		Align(lipgloss.Left)

	valueStyle := lipgloss.NewStyle().
		Width(5).
		Align(lipgloss.Right)

	return fmt.Sprintf("%s [%s%s] %s",
		labelStyle.Render(s.Label),
		filledBar,
		emptyBar,
		valueStyle.Render(fmt.Sprintf("%d%%", s.Value)))
}

// AsFloat returns the value as a 0.0-1.0 float for config storage.
func (s Slider) AsFloat() float64 {
	return float64(s.Value) / 100.0
}

// FromFloat creates a slider value from a 0.0-1.0 float.
func FromFloat(f float64) int {
	v := int(f * 100.0)
	if v < 0 {
		v = 0
	} else if v > 100 {
		v = 100
	}
	return v
}
