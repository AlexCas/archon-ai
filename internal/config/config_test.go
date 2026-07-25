package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"gopkg.in/yaml.v3"
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
				// S1c-1: flat-string scalars decode to ModelRef{Model: value}
				Models: ModelConfig{
					Default: ModelRef{Model: "claude-sonnet-4"},
					Phases:  map[string]ModelRef{"apply": {Model: "gpt-4o"}},
				},
				SkillInventory: []SkillInventory{
					{Name: "sdd-init", Version: "2.0", Source: "embedded"},
				},
			},
			wantErr: false,
		},
		{
			// S1c-1: mapping-form fixture decodes to structured ModelRef
			name: "mapping-form model ref",
			fs: fstest.MapFS{
				".archon/config.yaml": &fstest.MapFile{
					Data: []byte(`harness_version: "1.0.0"
agent: opencode
models:
  default:
    provider: opencode
    model: deepseek-v4-pro
`),
				},
			},
			want: Config{
				Version: "1.0.0",
				Agent:   "opencode",
				Models: ModelConfig{
					Default: ModelRef{Provider: "opencode", Model: "deepseek-v4-pro"},
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
					t.Errorf("Models.Default = %+v, want %+v", got.Models.Default, tt.want.Models.Default)
				}
				if len(got.Models.Phases) != len(tt.want.Models.Phases) {
					t.Errorf("Models.Phases length = %d, want %d", len(got.Models.Phases), len(tt.want.Models.Phases))
				} else {
					for k, v := range tt.want.Models.Phases {
						if got.Models.Phases[k] != v {
							t.Errorf("Models.Phases[%q] = %+v, want %+v", k, got.Models.Phases[k], v)
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
	// S1c-3: Default/Leader/Phases use ModelRef values
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
		Security: Security{
			Enabled: true,
			Profile: "web",
		},
		Impeccable: Impeccable{
			Enabled:     true,
			AutoInstall: true,
			Severity:    "block-all",
			ProductPath: "PRODUCT.md",
			DesignPath:  "DESIGN.md",
		},
		Models: ModelConfig{
			Default: ModelRef{Provider: "anthropic", Model: "claude-sonnet-4-20250514"},
			Leader:  ModelRef{Provider: "anthropic", Model: "claude-opus-4-8"},
			Phases: map[string]ModelRef{
				"apply":  {Model: "claude-sonnet-4-6"},
				"verify": {Model: "claude-haiku-4-5"},
				"judge":  {Model: "claude-opus-4-8"},
			},
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

	// S1c-3: Independent maps — mutating the clone must not affect the original.
	clone.Models.Phases["apply"] = ModelRef{Model: "MUTATED"}
	if original.Models.Phases["apply"] == (ModelRef{Model: "MUTATED"}) {
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

	// S1c-4: Default/Leader/Phases use ModelRef values
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
			Default: ModelRef{Provider: "anthropic", Model: "claude-opus-4-8"},
			Leader:  ModelRef{Provider: "anthropic", Model: "claude-sonnet-4-20250514"},
			Phases:  map[string]ModelRef{"apply": {Provider: "openai", Model: "claude-sonnet-4-6"}},
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
	// S1c-4: ModelRef is a comparable value type — direct comparison works.
	if loaded.Models.Leader != original.Models.Leader {
		t.Errorf("Models.Leader = %+v, want %+v", loaded.Models.Leader, original.Models.Leader)
	}
	if loaded.Models.Default != original.Models.Default {
		t.Errorf("Models.Default = %+v, want %+v", loaded.Models.Default, original.Models.Default)
	}
}

// TestConfig_FlatStringRoundtripByteIdentical asserts that a legacy flat-string
// models block loads and re-marshals with its models: block byte-identical. (S1c-2, M4)
func TestConfig_FlatStringRoundtripByteIdentical(t *testing.T) {
	// The legacy flat fixture as it would appear in config.yaml
	modelsBlock := "models:\n  default: claude-sonnet-4\n  phases:\n    apply: gpt-4o\n"
	fullYAML := "harness_version: 1.0.0\nagent: opencode\n" + modelsBlock

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".archon", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(fullYAML), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := &Config{HomeDir: tmpDir}
	if err := cfg.Load(os.DirFS(tmpDir)); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Marshal just the models block to compare
	got, err := yaml.Marshal(cfg.Models)
	if err != nil {
		t.Fatalf("yaml.Marshal(models) error = %v", err)
	}

	// The expected models block (without the "models:" wrapper key, just the value)
	// yaml.Marshal of ModelConfig should produce: default: claude-sonnet-4\nphases:\n    apply: gpt-4o\n
	wantModels := "default: claude-sonnet-4\nphases:\n    apply: gpt-4o\n"
	gotStr := string(got)
	if gotStr != wantModels {
		t.Errorf("models block round-trip mismatch:\n got: %q\nwant: %q", gotStr, wantModels)
	}

	// Assert empty-leader config does NOT produce a leader: key
	if strings.Contains(gotStr, "leader:") {
		t.Error("models block contains unexpected 'leader:' key for empty leader")
	}

	// Assert a config with NO models key saves without inventing a models: block
	minimalYAML := "harness_version: 1.0.0\nagent: opencode\n"
	if err := os.WriteFile(configPath, []byte(minimalYAML), 0o644); err != nil {
		t.Fatalf("WriteFile minimal: %v", err)
	}
	var minCfg Config
	minCfg.HomeDir = tmpDir
	if err := minCfg.Load(os.DirFS(tmpDir)); err != nil {
		t.Fatalf("Load() minimal error = %v", err)
	}
	// Save and re-read; models block must not appear
	if err := minCfg.Save(); err != nil {
		t.Fatalf("Save() minimal error = %v", err)
	}
	savedBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(savedBytes), "models:") {
		t.Errorf("saved minimal config contains unexpected 'models:' block:\n%s", savedBytes)
	}
}

// TestSecurity_DefaultOff asserts that loading a config with no security block
// yields Security{Enabled:false, Profile:""} — the zero value, default-off.
func TestSecurity_DefaultOff(t *testing.T) {
	fs := fstest.MapFS{
		".archon/config.yaml": &fstest.MapFile{
			Data: []byte("harness_version: \"1.0.0\"\nagent: claude\n"),
		},
	}
	var cfg Config
	if err := cfg.Load(fs); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Security.Enabled {
		t.Error("Security.Enabled = true; want false when block is absent")
	}
	if cfg.Security.Profile != "" {
		t.Errorf("Security.Profile = %q; want empty when block is absent", cfg.Security.Profile)
	}
}

// TestImpeccable_DefaultsAndValidation asserts (a) an absent impeccable block
// defaults Severity to "block-deterministic" and leaves Enabled false, and (b)
// an invalid severity value is rejected at Load() naming the value plus the
// three valid options.
func TestImpeccable_DefaultsAndValidation(t *testing.T) {
	t.Run("absent block defaults", func(t *testing.T) {
		fs := fstest.MapFS{
			".archon/config.yaml": &fstest.MapFile{
				Data: []byte("harness_version: \"1.0.0\"\nagent: claude\n"),
			},
		}
		var cfg Config
		if err := cfg.Load(fs); err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Impeccable.Enabled {
			t.Error("Impeccable.Enabled = true; want false when block is absent")
		}
		if cfg.Impeccable.Severity != "block-deterministic" {
			t.Errorf("Impeccable.Severity = %q; want %q", cfg.Impeccable.Severity, "block-deterministic")
		}
	})

	t.Run("invalid severity rejected", func(t *testing.T) {
		fs := fstest.MapFS{
			".archon/config.yaml": &fstest.MapFile{
				Data: []byte("harness_version: \"1.0.0\"\nagent: claude\nimpeccable:\n  severity: foobar\n"),
			},
		}
		var cfg Config
		err := cfg.Load(fs)
		if err == nil {
			t.Fatal("Load() error = nil, want error for invalid severity")
		}
		if !strings.Contains(err.Error(), "foobar") {
			t.Errorf("error = %q, want contains %q", err.Error(), "foobar")
		}
		for _, v := range ValidImpeccableSeverities {
			if !strings.Contains(err.Error(), v) {
				t.Errorf("error = %q, want contains valid option %q", err.Error(), v)
			}
		}
	})
}

// TestConfig_SlashedScalarRoundtripByteIdentical asserts that a legacy
// provider-qualified scalar (the documented `provider/model` form, e.g. the
// --leader flag) re-marshals as the SAME one-line scalar, not a mapping — so an
// existing config that used provider/model strings is not churned on save. (M4,
// judge issue 1)
func TestConfig_SlashedScalarRoundtripByteIdentical(t *testing.T) {
	// leader uses the documented provider/model scalar; a phase too
	modelsBlock := "default: claude-sonnet-4\nleader: anthropic/claude-sonnet-4-20250514\nphases:\n    apply: opencode/deepseek-v4-pro\n"

	var mc ModelConfig
	if err := yaml.Unmarshal([]byte(modelsBlock), &mc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// Sanity: the slashed scalar split into provider + model
	if mc.Leader.Provider != "anthropic" || mc.Leader.Model != "claude-sonnet-4-20250514" {
		t.Fatalf("leader split wrong: %+v", mc.Leader)
	}

	got, err := yaml.Marshal(mc)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(got) != modelsBlock {
		t.Errorf("slashed-scalar round-trip mismatch:\n got: %q\nwant: %q", string(got), modelsBlock)
	}
	// Explicitly assert no value churned into a mapping
	if strings.Contains(string(got), "provider:") {
		t.Errorf("a scalar value was rewritten as a mapping:\n%s", got)
	}
}
