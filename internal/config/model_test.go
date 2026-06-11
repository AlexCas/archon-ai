package config

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		wantWarn   bool
	}{
		{name: "known model gpt-4o", model: "gpt-4o", wantWarn: false},
		{name: "known model claude-sonnet-4", model: "claude-sonnet-4", wantWarn: false},
		{name: "known model gemini-2.5-pro", model: "gemini-2.5-pro", wantWarn: false},
		{name: "known model o4-mini", model: "o4-mini", wantWarn: false},
		{name: "unknown model", model: "future-model-v2", wantWarn: true},
		{name: "empty string", model: "", wantWarn: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.model)
			hasWarn := got != ""
			if hasWarn != tt.wantWarn {
				t.Errorf("Validate(%q) warning = %v, want %v (msg: %q)", tt.model, hasWarn, tt.wantWarn, got)
			}
		})
	}
}
