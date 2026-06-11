package status

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/archon-ai/archon/internal/config"
)

func TestDisplay(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		contains []string
	}{
		{
			name: "full config",
			cfg: &config.Config{
				Version:    "1.0.0",
				Agent:      "opencode",
				SkillCount: 23,
				CreatedAt:  time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				MutationTesting: config.MutationTesting{
					Enabled:   true,
					Tool:      "gremlins",
					Threshold: 0.80,
				},
				SkillInventory: []config.SkillInventory{
					{Name: "sdd-init", Version: "2.0", Source: "embedded"},
					{Name: "sdd-propose", Version: "1.5", Source: "embedded"},
				},
			},
			contains: []string{
				"Archon Harness Status",
				"opencode",
				"1.0.0",
				"23",
				"2026-06-10",
				"Mutation Testing",
				"Enabled:   true",
				"gremlins",
				"0.80",
				"sdd-init",
				"sdd-propose",
				"v2.0",
				"v1.5",
				"embedded",
			},
		},
		{
			name: "mutation testing disabled",
			cfg: &config.Config{
				Version:    "1.0.0",
				Agent:      "claude",
				SkillCount: 10,
				CreatedAt:  time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				MutationTesting: config.MutationTesting{
					Enabled: false,
				},
				SkillInventory: []config.SkillInventory{
					{Name: "sdd-init", Version: "2.0", Source: "embedded"},
				},
			},
			contains: []string{
				"claude",
				"Enabled:   false",
			},
		},
		{
			name: "no skills installed",
			cfg: &config.Config{
				Version:         "1.0.0",
				Agent:           "agents",
				SkillCount:      0,
				CreatedAt:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				MutationTesting: config.MutationTesting{Enabled: false},
				SkillInventory:  nil,
			},
			contains: []string{
				"agents",
				"Installed Skills: none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Display(&buf, tt.cfg)
			got := buf.String()

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Display() output missing %q\ngot:\n%s", want, got)
				}
			}
		})
	}
}

func TestFormat(t *testing.T) {
	cfg := &config.Config{
		Version:    "1.0.0",
		Agent:      "opencode",
		SkillCount: 23,
		CreatedAt:  time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		MutationTesting: config.MutationTesting{
			Enabled:   true,
			Tool:      "gremlins",
			Threshold: 0.80,
		},
		SkillInventory: []config.SkillInventory{
			{Name: "sdd-init", Version: "2.0", Source: "embedded"},
		},
	}

	got := Format(cfg)

	if !strings.Contains(got, "opencode") {
		t.Errorf("Format() missing agent name")
	}
	if !strings.Contains(got, "1.0.0") {
		t.Errorf("Format() missing version")
	}
	if !strings.Contains(got, "sdd-init") {
		t.Errorf("Format() missing skill name")
	}
}
