package agent

import (
	"fmt"
	"testing"
)

type mockPrompter struct {
	response string
	err      error
}

func (m *mockPrompter) Prompt(message string, options []string) (string, error) {
	return m.response, m.err
}

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		result    *Result
		flagAgent string
		prompter  Prompter
		want      string
		wantErr   bool
	}{
		{
			name:      "single agent no prompt needed",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode"}},
			flagAgent: "",
			prompter:  nil,
			want:      "opencode",
			wantErr:   false,
		},
		{
			name:      "flag override single agent",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode"}},
			flagAgent: "opencode",
			prompter:  nil,
			want:      "opencode",
			wantErr:   false,
		},
		{
			name:      "flag override multi agent",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode", "claude"}},
			flagAgent: "claude",
			prompter:  nil,
			want:      "claude",
			wantErr:   false,
		},
		{
			name:      "flag override not found",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode"}},
			flagAgent: "codex",
			prompter:  nil,
			want:      "",
			wantErr:   true,
		},
		{
			name:      "multi agent with prompter",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode", "claude"}},
			flagAgent: "",
			prompter:  &mockPrompter{response: "claude"},
			want:      "claude",
			wantErr:   false,
		},
		{
			name:      "multi agent prompter error",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode", "claude"}},
			flagAgent: "",
			prompter:  &mockPrompter{err: fmt.Errorf("user cancelled")},
			want:      "",
			wantErr:   true,
		},
		{
			name:      "multi agent no prompter",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode", "claude"}},
			flagAgent: "",
			prompter:  nil,
			want:      "",
			wantErr:   true,
		},
		{
			name:      "all four agents with prompter",
			result:    &Result{Agent: "opencode", Dirs: []string{"opencode", "claude", "agents", "codex"}},
			flagAgent: "",
			prompter:  &mockPrompter{response: "agents"},
			want:      "agents",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resolver{Prompter: tt.prompter}
			got, err := r.Resolve(tt.result, tt.flagAgent)

			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
