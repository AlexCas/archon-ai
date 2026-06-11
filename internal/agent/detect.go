package agent

import (
	"fmt"
	"io/fs"
)

var agentDirs = []struct {
	name string
	dir  string
}{
	{"opencode", ".opencode"},
	{"claude", ".claude"},
	{"agents", ".agents"},
	{"codex", ".codex"},
}

type Result struct {
	Agent string
	Dirs  []string
}

func Detect(fsys fs.FS) (*Result, error) {
	var found []string

	for _, a := range agentDirs {
		if _, err := fs.Stat(fsys, a.dir); err == nil {
			found = append(found, a.name)
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no agent directory found")
	}

	return &Result{
		Agent: found[0],
		Dirs:  found,
	}, nil
}
