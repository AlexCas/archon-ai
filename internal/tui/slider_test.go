package tui

import (
	"testing"
)

func TestSlider_NewSlider(t *testing.T) {
	s := NewSlider("Threshold", 50)
	if s.Label != "Threshold" {
		t.Errorf("Label = %q, want %q", s.Label, "Threshold")
	}
	if s.Value != 50 {
		t.Errorf("Value = %d, want 50", s.Value)
	}
	if s.Width != 40 {
		t.Errorf("Width = %d, want 40", s.Width)
	}
	if s.Focused {
		t.Error("Focused should be false")
	}
}

func TestSlider_SetValue(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{50, 50},
		{0, 0},
		{100, 100},
		{-1, 0},
		{101, 100},
		{999, 100},
	}

	for _, tt := range tests {
		s := NewSlider("Test", 0)
		s.SetValue(tt.input)
		if s.Value != tt.expected {
			t.Errorf("SetValue(%d) = %d, want %d", tt.input, s.Value, tt.expected)
		}
	}
}

func TestSlider_IncDec(t *testing.T) {
	s := NewSlider("Test", 50)

	s.Inc()
	if s.Value != 51 {
		t.Errorf("Inc() = %d, want 51", s.Value)
	}

	s.Dec()
	if s.Value != 50 {
		t.Errorf("Dec() = %d, want 50", s.Value)
	}
}

func TestSlider_IncAtMax(t *testing.T) {
	s := NewSlider("Test", 100)
	s.Inc()
	if s.Value != 100 {
		t.Errorf("Inc() at max = %d, want 100", s.Value)
	}
}

func TestSlider_DecAtMin(t *testing.T) {
	s := NewSlider("Test", 0)
	s.Dec()
	if s.Value != 0 {
		t.Errorf("Dec() at min = %d, want 0", s.Value)
	}
}

func TestSlider_AsFloat(t *testing.T) {
	tests := []struct {
		value    int
		expected float64
	}{
		{0, 0.0},
		{50, 0.5},
		{100, 1.0},
		{25, 0.25},
		{75, 0.75},
	}

	for _, tt := range tests {
		s := NewSlider("Test", tt.value)
		got := s.AsFloat()
		if got != tt.expected {
			t.Errorf("AsFloat(%d) = %f, want %f", tt.value, got, tt.expected)
		}
	}
}

func TestFromFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{0.0, 0},
		{0.5, 50},
		{1.0, 100},
		{0.25, 25},
		{0.75, 75},
		{-0.1, 0},
		{1.1, 100},
	}

	for _, tt := range tests {
		got := FromFloat(tt.input)
		if got != tt.expected {
			t.Errorf("FromFloat(%f) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestSlider_View(t *testing.T) {
	s := NewSlider("Test", 50)
	s.Width = 10
	view := s.View()
	if view == "" {
		t.Error("View() should not be empty")
	}
	if !contains(view, "Test") {
		t.Error("View() should contain label")
	}
	if !contains(view, "50%") {
		t.Error("View() should contain value percentage")
	}
}


