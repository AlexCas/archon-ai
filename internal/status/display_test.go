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
		{
			name: "models configured",
			cfg: &config.Config{
				Version:         "1.0.0",
				Agent:           "opencode",
				SkillCount:      10,
				CreatedAt:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				MutationTesting: config.MutationTesting{Enabled: false},
				Models: config.ModelConfig{
					Default: config.ModelRef{Model: "claude-sonnet-4"},
					Phases:  map[string]config.ModelRef{"apply": {Model: "gpt-4o"}},
				},
			},
			contains: []string{
				"Models",
				"claude-sonnet-4",
				"apply:",
				"gpt-4o",
			},
		},
		{
			name: "models not configured",
			cfg: &config.Config{
				Version:         "1.0.0",
				Agent:           "opencode",
				SkillCount:      10,
				CreatedAt:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				MutationTesting: config.MutationTesting{Enabled: false},
			},
			contains: []string{
				"Models",
				"(none configured)",
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

// TestDisplay_Impeccable covers spec scenarios "Status shows Impeccable as
// disabled" and "Status shows Impeccable as enabled with config details".
func TestDisplay_Impeccable(t *testing.T) {
	t.Run("disabled shows only Enabled: false", func(t *testing.T) {
		cfg := &config.Config{
			Version:         "1.0.0",
			Agent:           "claude",
			CreatedAt:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			MutationTesting: config.MutationTesting{Enabled: false},
		}
		var buf bytes.Buffer
		Display(&buf, cfg)
		got := buf.String()

		if !strings.Contains(got, "Impeccable (Design Language)") {
			t.Errorf("output missing Impeccable block:\n%s", got)
		}
		if !strings.Contains(got, "Enabled:   false") {
			t.Errorf("output missing disabled state:\n%s", got)
		}
		if strings.Contains(got, "Severity:") {
			t.Errorf("disabled block should not show Severity:\n%s", got)
		}
	})

	t.Run("enabled shows all fields", func(t *testing.T) {
		cfg := &config.Config{
			Version:         "1.0.0",
			Agent:           "claude",
			CreatedAt:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			MutationTesting: config.MutationTesting{Enabled: false},
			Impeccable: config.Impeccable{
				Enabled:     true,
				AutoInstall: true,
				Severity:    "block-all",
				ProductPath: "PRODUCT.md",
				DesignPath:  "DESIGN.md",
			},
		}
		var buf bytes.Buffer
		Display(&buf, cfg)
		got := buf.String()

		for _, want := range []string{"Enabled:   true", "block-all", "PRODUCT.md", "DESIGN.md"} {
			if !strings.Contains(got, want) {
				t.Errorf("output missing %q:\n%s", want, got)
			}
		}
	})
}

func TestDisplayWithUpdate(t *testing.T) {
	cfg := &config.Config{
		Version:         "1.0.0",
		Agent:           "claude",
		SkillCount:      10,
		CreatedAt:       time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		MutationTesting: config.MutationTesting{Enabled: false},
	}

	t.Run("hint shown when n > 0", func(t *testing.T) {
		var buf bytes.Buffer
		DisplayWithUpdate(&buf, cfg, 3)
		got := buf.String()
		if !strings.Contains(got, "Update available") {
			t.Errorf("output missing update hint:\n%s", got)
		}
		if !strings.Contains(got, "archon update") {
			t.Errorf("output missing 'archon update':\n%s", got)
		}
		if !strings.Contains(got, "3 skill(s)") {
			t.Errorf("output missing skill count:\n%s", got)
		}
	})

	t.Run("hint hidden when n == 0", func(t *testing.T) {
		var buf bytes.Buffer
		DisplayWithUpdate(&buf, cfg, 0)
		if strings.Contains(buf.String(), "Update available") {
			t.Errorf("hint should be hidden when n == 0:\n%s", buf.String())
		}
	})

	t.Run("Display delegates with 0", func(t *testing.T) {
		var buf bytes.Buffer
		Display(&buf, cfg)
		if strings.Contains(buf.String(), "Update available") {
			t.Errorf("Display must not render the hint:\n%s", buf.String())
		}
	})
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
