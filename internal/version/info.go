package version

import "fmt"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Print() string {
	return fmt.Sprintf("archon version %s (commit: %s, built: %s)", Version, Commit, Date)
}
