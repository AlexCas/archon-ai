// Package models composes the offered model catalog from injected dependencies:
// a PATH-based CLI detector and an opencode lister. The engine is pure and
// catalog-agnostic; the real implementations shell out via os/exec, while unit
// tests inject fakes (no real subprocess).
package models

import "os/exec"

// CLIDetector reports which agent CLIs are present on PATH. The returned map is
// keyed by CLI name with a true value when the CLI is found. Injecting this as a
// function makes detection trivial to fake in tests.
type CLIDetector func() map[string]bool

// LookPathDetector is the real CLIDetector. It probes the known agent CLIs via
// exec.LookPath and records each one found.
func LookPathDetector() map[string]bool {
	out := map[string]bool{}
	for _, c := range []string{"opencode", "claude", "codex", "gemini"} {
		if _, err := exec.LookPath(c); err == nil {
			out[c] = true
		}
	}
	return out
}
