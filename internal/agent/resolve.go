package agent

import (
	"fmt"
)

type Prompter interface {
	Prompt(message string, options []string) (string, error)
}

type Resolver struct {
	Prompter Prompter
}

func (r *Resolver) Resolve(result *Result, flagAgent string) (string, error) {
	if flagAgent != "" {
		for _, d := range result.Dirs {
			if d == flagAgent {
				return flagAgent, nil
			}
		}
		return "", fmt.Errorf("specified agent %q not found in project", flagAgent)
	}

	if len(result.Dirs) == 1 {
		return result.Agent, nil
	}

	if r.Prompter == nil {
		return "", fmt.Errorf("multiple agents found (%v) but no prompter available", result.Dirs)
	}

	selection, err := r.Prompter.Prompt(
		"Multiple AI agents detected. Which one should archon configure?",
		result.Dirs,
	)
	if err != nil {
		return "", fmt.Errorf("agent selection: %w", err)
	}

	return selection, nil
}
