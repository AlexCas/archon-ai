package agent

import (
	"testing"
	"testing/fstest"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name      string
		fs        fstest.MapFS
		wantAgent string
		wantDirs  []string
		wantErr   bool
	}{
		{
			name: "opencode only",
			fs: fstest.MapFS{
				".opencode/config.json": &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "opencode",
			wantDirs:  []string{"opencode"},
			wantErr:   false,
		},
		{
			name: "claude only",
			fs: fstest.MapFS{
				".claude/settings.json": &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "claude",
			wantDirs:  []string{"claude"},
			wantErr:   false,
		},
		{
			name: "agents only",
			fs: fstest.MapFS{
				".agents/config.yaml": &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "agents",
			wantDirs:  []string{"agents"},
			wantErr:   false,
		},
		{
			name: "codex only",
			fs: fstest.MapFS{
				".codex/config.json": &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "codex",
			wantDirs:  []string{"codex"},
			wantErr:   false,
		},
		{
			name: "priority: opencode over claude",
			fs: fstest.MapFS{
				".opencode/config.json": &fstest.MapFile{Data: []byte("{}")},
				".claude/settings.json": &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "opencode",
			wantDirs:  []string{"opencode", "claude"},
			wantErr:   false,
		},
		{
			name: "priority: opencode over all",
			fs: fstest.MapFS{
				".opencode/config.json": &fstest.MapFile{Data: []byte("{}")},
				".claude/settings.json": &fstest.MapFile{Data: []byte("{}")},
				".agents/config.yaml":  &fstest.MapFile{Data: []byte("{}")},
				".codex/config.json":   &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "opencode",
			wantDirs:  []string{"opencode", "claude", "agents", "codex"},
			wantErr:   false,
		},
		{
			name: "priority: claude over agents and codex",
			fs: fstest.MapFS{
				".claude/settings.json": &fstest.MapFile{Data: []byte("{}")},
				".agents/config.yaml":  &fstest.MapFile{Data: []byte("{}")},
				".codex/config.json":   &fstest.MapFile{Data: []byte("{}")},
			},
			wantAgent: "claude",
			wantDirs:  []string{"claude", "agents", "codex"},
			wantErr:   false,
		},
		{
			name:      "no agents found",
			fs:        fstest.MapFS{},
			wantAgent: "",
			wantDirs:  nil,
			wantErr:   true,
		},
		{
			name: "unrelated directories ignored",
			fs: fstest.MapFS{
				".git/config":           &fstest.MapFile{Data: []byte{}},
				".vscode/settings.json": &fstest.MapFile{Data: []byte{}},
			},
			wantAgent: "",
			wantDirs:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Detect(tt.fs)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Agent != tt.wantAgent {
				t.Errorf("Agent = %q, want %q", got.Agent, tt.wantAgent)
			}

			if len(got.Dirs) != len(tt.wantDirs) {
				t.Errorf("Dirs length = %d, want %d", len(got.Dirs), len(tt.wantDirs))
				return
			}

			for i, d := range got.Dirs {
				if d != tt.wantDirs[i] {
					t.Errorf("Dirs[%d] = %q, want %q", i, d, tt.wantDirs[i])
				}
			}
		})
	}
}
