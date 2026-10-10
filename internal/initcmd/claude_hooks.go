package initcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// hookScriptSubdir is where archon-owned hook scripts live, relative to
// .claude/. Every script archon generates sits under here so merges and
// future updates can recognize "this hook entry is ours" by command path
// alone, without a JSON comment marker (JSON has none) — the same role
// mapgen's "<!-- MAP:START/END -->" markers play for openspec/map.md, adapted
// to a format that can't carry comments.
const hookScriptSubdir = "hooks/archon"

// hookSpec describes one archon-owned hook: which event/tool it matches, the
// script that implements it, and the script's content. Adding a new HARD
// GATE guard (e.g. a state-transition validator) means adding one entry here
// — writeClaudeHooks and mergeClaudeHookSettings already handle N specs.
type hookSpec struct {
	Event         string // e.g. "PreToolUse"
	Matcher       string // e.g. "Bash"
	ScriptName    string // file name under .claude/hooks/archon/
	ScriptContent string
}

// commitAttributionScript blocks `git commit` / `gh pr create` / `gh pr edit`
// calls whose message carries Co-Authored-By or "Generated with" trailers,
// enforcing CLAUDE.md's Commit Attribution HARD RULE deterministically
// instead of relying on the orchestrator model to remember it every time —
// including when a competing system-reminder tries to inject that same
// attribution by default. Fails open (exit 0) when jq is unavailable or the
// input can't be parsed, or on any command that isn't a commit/PR call: a
// missing dependency must never block unrelated bash commands.
const commitAttributionScript = `#!/bin/sh
# Archon commit-attribution guard (PreToolUse / Bash).
# Blocks git commit / gh pr create|edit calls whose message contains
# Co-Authored-By or "Generated with" trailers, per this project's Commit
# Attribution HARD RULE in CLAUDE.md: commits and PR bodies carry only the
# user's own authorship, never tool/agent attribution.
set -eu

command -v jq >/dev/null 2>&1 || exit 0

input=$(cat)
cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' 2>/dev/null) || exit 0

[ -z "$cmd" ] && exit 0

case "$cmd" in
  *"git commit"*|*"gh pr create"*|*"gh pr edit"*)
    lower=$(printf '%s' "$cmd" | tr '[:upper:]' '[:lower:]')
    case "$lower" in
      *co-authored-by*|*"generated with"*)
        echo "Blocked by archon's commit-attribution guard: this project's CLAUDE.md forbids Co-Authored-By / \"Generated with\" attribution in commit messages and PR bodies. Remove it and retry." >&2
        exit 2
        ;;
    esac
    ;;
esac

exit 0
`

// sessionRecoveryScript prints SESSION_STATUS.md's content as context
// whenever it exists at the project root, implementing CLAUDE.md's Session
// Status contract ("On the next session, READ it FIRST to restore context")
// deterministically instead of depending on the orchestrator model to take
// the initiative to read the file before doing anything else. Empty matcher
// (below) means this fires on every SessionStart sub-event — startup,
// resume, clear, and compact — since re-printing a small status file is
// harmless even when "too often", and CLAUDE.md's concern spans all of
// those (a closed-and-reopened session, a /clear mid-change, a compaction
// that might otherwise lose track of the active change). Fails open (exit 0,
// no output) when jq is unavailable, cwd is missing from the hook's input,
// or there is no SESSION_STATUS.md to show.
const sessionRecoveryScript = `#!/bin/sh
# Archon session-recovery guard (SessionStart).
# Prints SESSION_STATUS.md's content as context when it exists at the
# project root, per CLAUDE.md's Session Status contract: "On the next
# session, READ it FIRST to restore context."
set -eu

command -v jq >/dev/null 2>&1 || exit 0

input=$(cat)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null) || exit 0

[ -z "$cwd" ] && exit 0

status_file="$cwd/SESSION_STATUS.md"
[ -f "$status_file" ] || exit 0

printf '## Resuming an in-progress Archon SDD session\n\n'
printf 'SESSION_STATUS.md exists at the project root. Read it first, per the Session Status contract in CLAUDE.md, before taking any other action. Its contents:\n\n'
cat "$status_file"

exit 0
`

// claudeHookSpecs is the fixed, ordered set of hooks archon generates for
// claude projects. Order is part of the determinism contract: re-running
// writeClaudeHooks with the same specs yields byte-identical output.
var claudeHookSpecs = []hookSpec{
	{Event: "PreToolUse", Matcher: "Bash", ScriptName: "commit-attribution.sh", ScriptContent: commitAttributionScript},
	{Event: "SessionStart", Matcher: "", ScriptName: "session-recovery.sh", ScriptContent: sessionRecoveryScript},
}

// WriteClaudeHooks is the exported integration seam, mirroring
// WriteClaudeAgents: it writes each spec's script under
// .claude/hooks/archon/ and additively merges the corresponding hook
// registration into .claude/settings.json, then returns the paths it wrote.
func WriteClaudeHooks(projectDir string, w io.Writer) ([]string, error) {
	return writeClaudeHooks(projectDir, w)
}

// writeClaudeHooks writes claudeHookSpecs' scripts and merges their
// registration into .claude/settings.json. Unlike writeClaudeAgents (whose
// archon-<phase>.md files are exclusively archon's), settings.json is a file
// the user may already have their own hooks/permissions/env in — so this
// never overwrites the file wholesale. It only ever touches the hook entries
// it recognizes as its own (by ScriptName under hookScriptSubdir) and leaves
// every other key and every other hook entry exactly as it found them.
func writeClaudeHooks(projectDir string, w io.Writer) ([]string, error) {
	if len(claudeHookSpecs) == 0 {
		return nil, nil
	}

	scriptDir := filepath.Join(projectDir, ".claude", hookScriptSubdir)
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		return nil, fmt.Errorf("create hooks dir: %w", err)
	}

	var written []string
	for _, spec := range claudeHookSpecs {
		path := filepath.Join(scriptDir, spec.ScriptName)
		if err := os.WriteFile(path, []byte(spec.ScriptContent), 0o755); err != nil {
			return nil, fmt.Errorf("write hook script %s: %w", spec.ScriptName, err)
		}
		written = append(written, path)
	}

	settingsPath, err := mergeClaudeHookSettings(projectDir, claudeHookSpecs)
	if err != nil {
		return nil, err
	}
	if settingsPath != "" {
		written = append(written, settingsPath)
		fmt.Fprintf(w, "merged %d archon hook(s) into %s\n", len(claudeHookSpecs), settingsPath)
	}

	return written, nil
}

// ownCommandMarker is the substring that identifies a hook entry's command as
// archon's own, regardless of how the surrounding path is spelled (absolute,
// or using ${CLAUDE_PROJECT_DIR}).
const ownCommandMarker = "/.claude/" + hookScriptSubdir + "/"

// mergeClaudeHookSettings additively merges claudeHookSpecs into
// <projectDir>/.claude/settings.json's "hooks" key. For each (Event,
// Matcher) pair it removes any existing matcher-group entries that are
// archon's own (identified by ownCommandMarker in their command) and
// replaces them with freshly rendered ones, preserving every other top-level
// key and every other hook entry — including ones added by the user or by
// other tools — untouched and in their original order. Returns "" (no error)
// when there is nothing to merge.
func mergeClaudeHookSettings(projectDir string, specs []hookSpec) (string, error) {
	path := filepath.Join(projectDir, ".claude", "settings.json")

	doc := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &doc); err != nil {
			return "", fmt.Errorf("parse settings.json: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("read settings.json: %w", err)
	}

	hooksDoc, _ := doc["hooks"].(map[string]any)
	if hooksDoc == nil {
		hooksDoc = map[string]any{}
	}

	// Group specs by event, preserving first-seen event order, so a future
	// second PreToolUse spec lands in the same matcher-group handling as the
	// first rather than requiring new code.
	var eventOrder []string
	byEvent := map[string][]hookSpec{}
	for _, s := range specs {
		if _, ok := byEvent[s.Event]; !ok {
			eventOrder = append(eventOrder, s.Event)
		}
		byEvent[s.Event] = append(byEvent[s.Event], s)
	}

	for _, event := range eventOrder {
		existing, _ := hooksDoc[event].([]any)

		// Keep every existing matcher-group that is not archon's own.
		kept := make([]any, 0, len(existing))
		for _, groupAny := range existing {
			if !isArchonOwnedGroup(groupAny) {
				kept = append(kept, groupAny)
			}
		}

		// Re-add archon's own groups, grouped by matcher, in spec order.
		var matcherOrder []string
		byMatcher := map[string][]hookSpec{}
		for _, s := range byEvent[event] {
			if _, ok := byMatcher[s.Matcher]; !ok {
				matcherOrder = append(matcherOrder, s.Matcher)
			}
			byMatcher[s.Matcher] = append(byMatcher[s.Matcher], s)
		}
		for _, matcher := range matcherOrder {
			var handlers []any
			for _, s := range byMatcher[matcher] {
				handlers = append(handlers, map[string]any{
					"type":    "command",
					"command": "${CLAUDE_PROJECT_DIR}/.claude/" + hookScriptSubdir + "/" + s.ScriptName,
				})
			}
			kept = append(kept, map[string]any{
				"matcher": matcher,
				"hooks":   handlers,
			})
		}

		hooksDoc[event] = kept
	}

	doc["hooks"] = hooksDoc

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode settings.json: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create .claude dir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", fmt.Errorf("write temp settings.json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", fmt.Errorf("rename settings.json: %w", err)
	}

	return path, nil
}

// isArchonOwnedGroup reports whether a decoded hooks.<Event> array element
// (expected shape: {"matcher": ..., "hooks": [{"command": ...}, ...]}) is a
// matcher-group archon generated, i.e. every one of its handler commands
// carries ownCommandMarker. A group is only ever replaced wholesale when
// it's entirely archon's; a malformed or foreign entry is left untouched
// rather than guessed at.
func isArchonOwnedGroup(groupAny any) bool {
	group, ok := groupAny.(map[string]any)
	if !ok {
		return false
	}
	handlers, ok := group["hooks"].([]any)
	if !ok || len(handlers) == 0 {
		return false
	}
	for _, hAny := range handlers {
		h, ok := hAny.(map[string]any)
		if !ok {
			return false
		}
		cmd, _ := h["command"].(string)
		if !strings.Contains(cmd, ownCommandMarker) {
			return false
		}
	}
	return true
}
