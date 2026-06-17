package status

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/archon-ai/archon/internal/config"
)

func Display(w io.Writer, cfg *config.Config) {
	DisplayWithUpdate(w, cfg, 0)
}

// DisplayWithUpdate renders the harness status and, when n > 0, appends a single
// line hinting that an update is available. n is the number of skills the
// embedded set adds or changes relative to what is installed. The hint is purely
// advisory; the count is computed by the caller and a zero value renders nothing.
func DisplayWithUpdate(w io.Writer, cfg *config.Config, n int) {
	fmt.Fprintln(w, "Archon Harness Status")
	fmt.Fprintln(w, "=====================")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  Agent:            %s\n", cfg.Agent)
	fmt.Fprintf(w, "  Harness Version:  %s\n", cfg.Version)
	fmt.Fprintf(w, "  Skill Count:      %d\n", cfg.SkillCount)
	fmt.Fprintf(w, "  Created At:       %s\n", cfg.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w)

	fmt.Fprintln(w, "  Mutation Testing")
	fmt.Fprintln(w, "  ----------------")
	fmt.Fprintf(w, "    Enabled:   %t\n", cfg.MutationTesting.Enabled)
	if cfg.MutationTesting.Enabled {
		fmt.Fprintf(w, "    Tool:      %s\n", cfg.MutationTesting.Tool)
		fmt.Fprintf(w, "    Threshold: %.2f\n", cfg.MutationTesting.Threshold)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "  Playwright (Web E2E)")
	fmt.Fprintln(w, "  --------------------")
	fmt.Fprintf(w, "    Enabled:   %t\n", cfg.Playwright.Enabled)
	if cfg.Playwright.Enabled {
		if cfg.Playwright.TestDir != "" {
			fmt.Fprintf(w, "    Test Dir:  %s\n", cfg.Playwright.TestDir)
		}
		if cfg.Playwright.BaseURL != "" {
			fmt.Fprintf(w, "    Base URL:  %s\n", cfg.Playwright.BaseURL)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "  Models")
	fmt.Fprintln(w, "  ------")
	if cfg.Models.Default == "" && len(cfg.Models.Phases) == 0 {
		fmt.Fprintln(w, "    (none configured)")
	} else {
		if cfg.Models.Default != "" {
			fmt.Fprintf(w, "    Default:  %s\n", cfg.Models.Default)
		}
		if len(cfg.Models.Phases) > 0 {
			phases := make([]string, 0, len(cfg.Models.Phases))
			for k := range cfg.Models.Phases {
				phases = append(phases, k)
			}
			sort.Strings(phases)
			for _, phase := range phases {
				fmt.Fprintf(w, "    %-8s %s\n", phase+":", cfg.Models.Phases[phase])
			}
		}
	}
	fmt.Fprintln(w)

	if len(cfg.SkillInventory) > 0 {
		fmt.Fprintln(w, "  Installed Skills")
		fmt.Fprintln(w, "  ----------------")
		for _, s := range cfg.SkillInventory {
			fmt.Fprintf(w, "    %-25s v%-6s (%s)\n", s.Name, s.Version, s.Source)
		}
	} else {
		fmt.Fprintln(w, "  Installed Skills: none")
	}
	fmt.Fprintln(w)

	if n > 0 {
		fmt.Fprintf(w, "  Update available — run 'archon update' (%d skill(s))\n", n)
		fmt.Fprintln(w)
	}
}

func Format(cfg *config.Config) string {
	var b strings.Builder
	Display(&b, cfg)
	return b.String()
}
