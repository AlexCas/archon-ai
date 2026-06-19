package opencode

import (
	"encoding/json"
	"testing"
)

func TestMergeJSONObjects(t *testing.T) {
	tests := []struct {
		name    string
		base    []byte
		overlay []byte
		check   func(t *testing.T, result map[string]any)
		wantErr bool
	}{
		{
			name:    "additive merge preserves user keys",
			base:    []byte(`{"user_key": "preserved", "nested": {"user_nested": 1}}`),
			overlay: []byte(`{"new_key": "added", "nested": {"overlay_nested": 2}}`),
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				if result["user_key"] != "preserved" {
					t.Errorf("user_key = %v, want preserved", result["user_key"])
				}
				if result["new_key"] != "added" {
					t.Errorf("new_key = %v, want added", result["new_key"])
				}
				nested, ok := result["nested"].(map[string]any)
				if !ok {
					t.Fatal("nested is not an object")
				}
				if nested["user_nested"] != float64(1) {
					t.Errorf("nested.user_nested = %v, want 1", nested["user_nested"])
				}
				if nested["overlay_nested"] != float64(2) {
					t.Errorf("nested.overlay_nested = %v, want 2", nested["overlay_nested"])
				}
			},
		},
		{
			name: "__replace__ sentinel replaces whole key",
			base: []byte(`{"task": {"keep": "me", "old_entry": "drop"}}`),
			overlay: []byte(`{
				"task": {
					"__replace__": {"new_entry": "only", "*": "deny"}
				}
			}`),
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				task, ok := result["task"].(map[string]any)
				if !ok {
					t.Fatal("task is not an object after replace")
				}
				if _, exists := task["old_entry"]; exists {
					t.Error("old_entry should have been replaced")
				}
				if _, exists := task["keep"]; exists {
					t.Error("keep should have been replaced (whole key replaced)")
				}
				if task["new_entry"] != "only" {
					t.Errorf("new_entry = %v, want only", task["new_entry"])
				}
				if task["*"] != "deny" {
					t.Errorf("* = %v, want deny", task["*"])
				}
			},
		},
		{
			name:    "empty base treated as empty object",
			base:    []byte{},
			overlay: []byte(`{"key": "value"}`),
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				if result["key"] != "value" {
					t.Errorf("key = %v, want value", result["key"])
				}
			},
		},
		{
			name:    "malformed base treated as empty object",
			base:    []byte(`not valid json`),
			overlay: []byte(`{"key": "value"}`),
			check: func(t *testing.T, result map[string]any) {
				t.Helper()
				if result["key"] != "value" {
					t.Errorf("key = %v, want value", result["key"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := MergeJSONObjects(tt.base, tt.overlay)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MergeJSONObjects error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			var result map[string]any
			if err := json.Unmarshal(merged, &result); err != nil {
				t.Fatalf("unmarshal result: %v", err)
			}
			tt.check(t, result)
		})
	}
}
