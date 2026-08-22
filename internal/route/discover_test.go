package route

import (
	"testing"
	"testing/fstest"
)

// TestActiveChange covers spec §Active-Change Discovery Precedence:
// --change flag > SESSION_STATUS.md "Active change:" field > sole
// non-archive folder under openspec/changes/ > "none". Uses an in-memory
// fstest.MapFS — no real filesystem fixtures required.
func TestActiveChange(t *testing.T) {
	tests := []struct {
		name string
		fsys fstest.MapFS
		flag string
		want string
	}{
		{
			name: "flag override wins over SESSION_STATUS.md",
			fsys: fstest.MapFS{
				"SESSION_STATUS.md": {Data: []byte("- **Active change**: foo\n")},
			},
			flag: "bar",
			want: "bar",
		},
		{
			name: "SESSION_STATUS.md used when no flag provided",
			fsys: fstest.MapFS{
				"SESSION_STATUS.md": {Data: []byte("- **Active change**: local-model-router\n")},
			},
			flag: "",
			want: "local-model-router",
		},
		{
			name: "SESSION_STATUS.md plain field format",
			fsys: fstest.MapFS{
				"SESSION_STATUS.md": {Data: []byte("Active change: plain-format\n")},
			},
			flag: "",
			want: "plain-format",
		},
		{
			name: "sole non-archive folder fallback when no SESSION_STATUS.md",
			fsys: fstest.MapFS{
				"openspec/changes/only-change/state.yaml": {Data: []byte("phase: explore\nstatus: in_progress\n")},
			},
			flag: "",
			want: "only-change",
		},
		{
			name: "archive folder excluded from sole-folder fallback",
			fsys: fstest.MapFS{
				"openspec/changes/archive/old-change/state.yaml": {Data: []byte("phase: archive\nstatus: completed\n")},
				"openspec/changes/current-change/state.yaml":     {Data: []byte("phase: spec\nstatus: in_progress\n")},
			},
			flag: "",
			want: "current-change",
		},
		{
			name: "multiple non-archive folders -> none (ambiguous)",
			fsys: fstest.MapFS{
				"openspec/changes/change-a/state.yaml": {Data: []byte("phase: spec\nstatus: in_progress\n")},
				"openspec/changes/change-b/state.yaml": {Data: []byte("phase: design\nstatus: in_progress\n")},
			},
			flag: "",
			want: "none",
		},
		{
			name: "no SESSION_STATUS.md, no changes dir -> none",
			fsys: fstest.MapFS{},
			flag: "",
			want: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := activeChangeFS(tt.fsys, tt.flag); got != tt.want {
				t.Errorf("activeChangeFS(...) = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestReadState covers readState's tolerant behavior: valid, missing, and
// corrupt state.yaml all return sane results without erroring.
func TestReadState(t *testing.T) {
	tests := []struct {
		name       string
		fsys       fstest.MapFS
		path       string
		wantPhase  string
		wantStatus string
	}{
		{
			name:       "valid state.yaml",
			fsys:       fstest.MapFS{"state.yaml": {Data: []byte("phase: design\nstatus: in_progress\n")}},
			path:       "state.yaml",
			wantPhase:  "design",
			wantStatus: "in_progress",
		},
		{
			name:       "missing file returns empty, not an error",
			fsys:       fstest.MapFS{},
			path:       "state.yaml",
			wantPhase:  "",
			wantStatus: "",
		},
		{
			name:       "corrupt yaml returns empty, not an error",
			fsys:       fstest.MapFS{"state.yaml": {Data: []byte("not: [valid: yaml")}},
			path:       "state.yaml",
			wantPhase:  "",
			wantStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, status := readState(tt.fsys, tt.path)
			if phase != tt.wantPhase || status != tt.wantStatus {
				t.Errorf("readState(...) = (%q, %q), want (%q, %q)", phase, status, tt.wantPhase, tt.wantStatus)
			}
		})
	}
}
