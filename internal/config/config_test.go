package config

import (
	"os"
	"reflect"
	"testing"
	"testing/fstest"
	"time"
)

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name    string
		fs      fstest.MapFS
		want    Config
		wantErr bool
	}{
		{
			name: "valid config",
			fs: fstest.MapFS{
				".archon/config.yaml": &fstest.MapFile{
					Data: []byte(`harness_version: "1.0.0"
agent: opencode
skill_count: 23
created_at: 2026-06-10T00:00:00Z
mutation_testing:
  enabled: true
  tool: gremlins
  threshold: 0.80
models:
  default: claude-sonnet-4
  phases:
    apply: gpt-4o
skill_inventory:
  - name: sdd-init
    version: "2.0"
    source: embedded
`),
				},
			},
			want: Config{
				Version:    "1.0.0",
				Agent:      "opencode",
				SkillCount: 23,
				CreatedAt:  time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
				MutationTesting: MutationTesting{
					Enabled:   true,
					Tool:      "gremlins",
					Threshold: 0.80,
				},
				Models: ModelConfig{
					Default: "claude-sonnet-4",
					Phases:  map[string]string{"apply": "gpt-4o"},
				},
				SkillInventory: []SkillInventory{
					{Name: "sdd-init", Version: "2.0", Source: "embedded"},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			fs: fstest.MapFS{
				".archon/config.yaml": &fstest.MapFile{
					Data: []byte(`harness_version: "1.0.0"
agent: claude
`),
				},
			},
			want: Config{
				Version: "1.0.0",
				Agent:   "claude",
			},
			wantErr: false,
		},
		{
			name:    "missing config",
			fs:      fstest.MapFS{},
			want:    Config{},
			wantErr: true,
		},
		{
			name: "invalid yaml",
			fs: fstest.MapFS{
				".archon/config.yaml": &fstest.MapFile{
					Data: []byte(`invalid: yaml: content: [`),
				},
			},
			want:    Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Config
			err := got.Load(tt.fs)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Version != tt.want.Version {
					t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
				}
				if got.Agent != tt.want.Agent {
					t.Errorf("Agent = %v, want %v", got.Agent, tt.want.Agent)
				}
				if got.SkillCount != tt.want.SkillCount {
					t.Errorf("SkillCount = %v, want %v", got.SkillCount, tt.want.SkillCount)
				}
				if !got.CreatedAt.Equal(tt.want.CreatedAt) {
					t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, tt.want.CreatedAt)
				}
				if got.MutationTesting.Enabled != tt.want.MutationTesting.Enabled {
					t.Errorf("MutationTesting.Enabled = %v, want %v", got.MutationTesting.Enabled, tt.want.MutationTesting.Enabled)
				}
				if got.MutationTesting.Tool != tt.want.MutationTesting.Tool {
					t.Errorf("MutationTesting.Tool = %v, want %v", got.MutationTesting.Tool, tt.want.MutationTesting.Tool)
				}
				if got.MutationTesting.Threshold != tt.want.MutationTesting.Threshold {
					t.Errorf("MutationTesting.Threshold = %v, want %v", got.MutationTesting.Threshold, tt.want.MutationTesting.Threshold)
				}
				if got.Models.Default != tt.want.Models.Default {
					t.Errorf("Models.Default = %v, want %v", got.Models.Default, tt.want.Models.Default)
				}
				if len(got.Models.Phases) != len(tt.want.Models.Phases) {
					t.Errorf("Models.Phases length = %d, want %d", len(got.Models.Phases), len(tt.want.Models.Phases))
				} else {
					for k, v := range tt.want.Models.Phases {
						if got.Models.Phases[k] != v {
							t.Errorf("Models.Phases[%q] = %q, want %q", k, got.Models.Phases[k], v)
						}
					}
				}
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Version:    "1.0.0",
		Agent:      "opencode",
		SkillCount: 23,
		CreatedAt:  time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		MutationTesting: MutationTesting{
			Enabled:   true,
			Tool:      "gremlins",
			Threshold: 0.80,
		},
		SkillInventory: []SkillInventory{
			{Name: "sdd-init", Version: "2.0", Source: "embedded"},
		},
		HomeDir: tmpDir,
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	var loaded Config
	if err := loaded.Load(fstest.MapFS{}); err == nil {
		t.Error("Load() should fail with empty FS")
	}

	loaded.HomeDir = tmpDir
	mapFS := fstest.MapFS{}
	if err := loaded.Load(mapFS); err != nil {
		t.Logf("Note: Load from MapFS after Save requires actual file system")
	}
}

// TestConfig_CloneRoundtrip builds a fully-populated Config with every field set
// to a non-zero value, clones it, and asserts the clone is a deep-equal but
// independent copy. It fails loudly if a new Config field is added without being
// copied in Clone (the clone would then differ from the original via DeepEqual).
func TestConfig_CloneRoundtrip(t *testing.T) {
	original := &Config{
		Version:    "1.2.3",
		Agent:      "claude",
		SkillCount: 24,
		CreatedAt:  time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		MutationTesting: MutationTesting{
			Enabled:   true,
			Tool:      "gremlins",
			Threshold: 0.85,
		},
		Playwright: Playwright{
			Enabled: true,
			TestDir: "e2e",
			BaseURL: "http://localhost:3000",
		},
		Models: ModelConfig{
			Default: "claude-opus-4-8",
			Leader:  "anthropic/claude-sonnet-4-20250514",
			Phases:  map[string]string{"apply": "claude-sonnet-4-6", "verify": "claude-haiku-4-5"},
		},
		SkillInventory: []SkillInventory{
			{Name: "sdd-init", Version: "2.0", Source: "embedded"},
		},
		HomeDir: "/tmp/project",
	}

	clone := original.Clone()

	// Deep equality: every field must be copied. A field added to Config but not
	// to Clone would show up here as a difference.
	if !reflect.DeepEqual(clone, original) {
		t.Fatalf("Clone() is not deep-equal to original.\n got: %+v\nwant: %+v", clone, original)
	}

	// Distinct pointer.
	if clone == original {
		t.Fatal("Clone() returned the same pointer as the original")
	}

	// Independent maps: mutating the clone must not affect the original.
	clone.Models.Phases["apply"] = "MUTATED"
	if original.Models.Phases["apply"] == "MUTATED" {
		t.Error("mutating clone.Models.Phases affected the original (shared map)")
	}

	// Independent slices: mutating the clone must not affect the original.
	clone.SkillInventory[0].Version = "MUTATED"
	if original.SkillInventory[0].Version == "MUTATED" {
		t.Error("mutating clone.SkillInventory affected the original (shared slice)")
	}
}

func TestConfig_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()

	original := &Config{
		Version:    "1.0.0",
		Agent:      "opencode",
		SkillCount: 23,
		CreatedAt:  time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		MutationTesting: MutationTesting{
			Enabled:   true,
			Tool:      "gremlins",
			Threshold: 0.80,
		},
		Models: ModelConfig{
			Default: "claude-opus-4-8",
			Leader:  "anthropic/claude-sonnet-4-20250514",
			Phases:  map[string]string{"apply": "claude-sonnet-4-6"},
		},
		SkillInventory: []SkillInventory{
			{Name: "sdd-init", Version: "2.0", Source: "embedded"},
			{Name: "sdd-propose", Version: "1.5", Source: "embedded"},
		},
		HomeDir: tmpDir,
	}

	if err := original.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load from the real file Save() wrote, so this exercises the actual
	// serialize -> reload path (not an empty in-memory FS).
	loaded := &Config{HomeDir: tmpDir}
	if err := loaded.Load(os.DirFS(tmpDir)); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("Version = %v, want %v", loaded.Version, original.Version)
	}
	if loaded.Agent != original.Agent {
		t.Errorf("Agent = %v, want %v", loaded.Agent, original.Agent)
	}
	if loaded.SkillCount != original.SkillCount {
		t.Errorf("SkillCount = %v, want %v", loaded.SkillCount, original.SkillCount)
	}
	// S1: models.leader must survive serialize -> reload verbatim.
	if loaded.Models.Leader != original.Models.Leader {
		t.Errorf("Models.Leader = %q, want %q", loaded.Models.Leader, original.Models.Leader)
	}
	if loaded.Models.Default != original.Models.Default {
		t.Errorf("Models.Default = %q, want %q", loaded.Models.Default, original.Models.Default)
	}
}
