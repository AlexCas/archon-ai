package config

import "fmt"

type ModelConfig struct {
	Default string            `yaml:"default,omitempty"`
	Phases  map[string]string `yaml:"phases,omitempty"`
}

var KnownModels = map[string]bool{
	"gpt-4":            true,
	"gpt-4o":           true,
	"gpt-4o-mini":      true,
	"claude-sonnet-4":  true,
	"claude-haiku-4":   true,
	"gemini-2.5-pro":   true,
	"gemini-2.5-flash": true,
	"o3":               true,
	"o3-mini":          true,
	"o4-mini":          true,
}

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
