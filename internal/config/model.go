package config

import "fmt"

type ModelConfig struct {
	Default string            `yaml:"default,omitempty"`
	Phases  map[string]string `yaml:"phases,omitempty"`
}

// ClaudeModels is the curated, ordered list of Claude models offered as static
// choices in the TUI. The list is intentionally small and current; users who
// need another model can always type a free-form string instead.
var ClaudeModels = []string{
	"claude-opus-4-8",
	"claude-sonnet-4-6",
	"claude-haiku-4-5",
}

// OpencodeModels is the curated, ordered list of models available through the
// Opencode Go subscription, offered as static choices in the TUI. Adjust this
// list as the subscription catalog changes; the free-form string option remains
// available for anything not listed here.
var OpencodeModels = []string{
	"deepseek-v4-flash",
	"deepseek-v4-pro",
	"glm-5",
	"glm-5.1",
	"kimi-k2.5",
	"kimi-k2.6",
	"qwen3.6-plus",
	"qwen3.7-plus",
}

// StaticModels returns the full ordered list of statically-selectable models,
// Claude first then Opencode Go. The TUI renders these as pickable options and
// still allows a free-form entry.
func StaticModels() []string {
	out := make([]string, 0, len(ClaudeModels)+len(OpencodeModels))
	out = append(out, ClaudeModels...)
	out = append(out, OpencodeModels...)
	return out
}

// KnownModels is the validation set: any statically-listed model is "known".
// Validation is advisory only — unknown (free-form) models are accepted with a
// warning, never rejected.
var KnownModels = func() map[string]bool {
	m := make(map[string]bool)
	for _, name := range StaticModels() {
		m[name] = true
	}
	return m
}()

var ValidPhases = map[string]bool{
	"explore": true,
	"propose": true,
	"spec":    true,
	"design":  true,
	"tasks":   true,
	"apply":   true,
	"verify":  true,
	"archive": true,
}

func Validate(model string) string {
	if model == "" {
		return ""
	}
	if KnownModels[model] {
		return ""
	}
	return fmt.Sprintf("warning: %q is not a known model (accepted anyway)", model)
}
