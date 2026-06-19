package models

import (
	"context"
	"os/exec"
	"strings"
)

// OpencodeLister enumerates the models available through the opencode CLI. The
// real implementation runs the subprocess; tests inject a fake returning canned
// data or an error.
type OpencodeLister interface {
	List(ctx context.Context) ([]string, error)
}

// execLister is the real OpencodeLister. It runs `opencode models opencode-go`
// and parses the output into bare model names.
type execLister struct{}

func (execLister) List(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "opencode", "models", "opencode-go")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseModels(out), nil
}

// parseModels turns the raw `opencode models` output into a slice of model
// names. Each line is trimmed (so trailing CR from CRLF output is removed),
// blank lines are skipped, and the leading `provider/` prefix is stripped
// (everything up to and including the FIRST "/"; a namespaced `a/b/c` keeps
// `b/c`). Lines with no slash are kept as-is. A line that is empty AFTER
// stripping the prefix (e.g. a bare `provider/`) is dropped, so malformed input
// never injects an empty "model" into the catalog.
func parseModels(b []byte) []string {
	var ms []string
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if i := strings.IndexByte(ln, '/'); i >= 0 {
			ln = ln[i+1:]
		}
		if ln == "" {
			continue
		}
		ms = append(ms, ln)
	}
	return ms
}
