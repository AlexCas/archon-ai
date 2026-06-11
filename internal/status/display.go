package status

import (
	"fmt"
	"io"
	"strings"

	"github.com/archon-ai/archon/internal/config"
)

func Display(w io.Writer, cfg *config.Config) {
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
}

func Format(cfg *config.Config) string {
	var b strings.Builder
	Display(&b, cfg)
	return b.String()
}
