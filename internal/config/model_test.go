package config

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		wantWarn   bool
	}{
		{name: "known claude model opus", model: "claude-opus-4-8", wantWarn: false},
		{name: "known claude model sonnet", model: "claude-sonnet-4-6", wantWarn: false},
		{name: "known claude model haiku", model: "claude-haiku-4-5", wantWarn: false},
		{name: "known opencode model", model: "glm-5", wantWarn: false},
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
