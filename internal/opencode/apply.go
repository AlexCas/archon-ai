package opencode

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ApplyOptions configures an Apply call.
type ApplyOptions struct {
	// SettingsPath is the full path to opencode.json (e.g. from SettingsPath()).
	// When empty, Apply returns a warning and exits without touching any file.
	SettingsPath string

	// CachePath is the full path to the opencode models cache (models.json).
	// When empty or the file is missing, resolution falls back to the static map.
	CachePath string

	// DefaultModel is the bare or qualified model name to use as the fallback
	// for any agent without an explicit phase assignment.
	DefaultModel string

	// Phases maps SDD phase names (e.g. "apply", "explore") to bare or qualified
	// model names. Phase entries override DefaultModel for the matching agent.
	Phases map[string]string
}

// Apply generates the opencode agent overlay, injects model assignments, and
// deep-merges the result into the opencode.json settings file. It:
//
//  1. Returns a warning (not an error) when SettingsPath is empty.
//  2. Resolves DefaultModel and each Phases entry; unresolvable names are
//     skipped with a warning — init never fails due to model resolution.
//  3. Reads existing opencode.json (treated as {} when absent).
//  4. Calls Inject with the resolved models and the existing agent keys.
//  5. Creates a timestamped backup of opencode.json if one already exists.
//  6. Deep-merges the injected overlay into the existing content.
//  7. Writes the merged result atomically (temp-file + rename).
//
// backupPath is non-empty only when a prior opencode.json existed and was backed up.
// opencode.json is NEVER added to CreatedPaths — it is a shared global file.
func Apply(opts ApplyOptions) (backupPath string, warnings []string, err error) {
	if opts.SettingsPath == "" {
		warnings = append(warnings, "opencode settings path unavailable (no home dir?); skipping overlay")
		return "", warnings, nil
	}

	// Resolve the default model using the provided CachePath for isolation.
	resolvedDefault := ""
	if opts.DefaultModel != "" {
		q, ok := resolveWithCachePath(opts.DefaultModel, opts.CachePath)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("cannot resolve default model %q; agent model fields will be empty", opts.DefaultModel))
		} else {
			resolvedDefault = q
		}
	}

	// Resolve per-phase models using the provided CachePath.
	resolvedPhases := make(map[string]string, len(opts.Phases))
	for phase, name := range opts.Phases {
		if name == "" {
			continue
		}
		q, ok := resolveWithCachePath(name, opts.CachePath)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("cannot resolve model %q for phase %q; skipping", name, phase))
			continue
		}
		resolvedPhases[phase] = q
	}

	// Read existing opencode.json (may not exist yet).
	existingContent, readErr := os.ReadFile(opts.SettingsPath)
	fileExisted := readErr == nil

	// Read existing agent keys for the decision tree.
	existingAgentKeys, _ := readExistingAgentKeys(opts.SettingsPath)

	// Inject models into the embedded overlay.
	injected, err := Inject(overlayJSON, resolvedDefault, resolvedPhases, existingAgentKeys)
	if err != nil {
		return "", warnings, fmt.Errorf("inject models into overlay: %w", err)
	}

	// Backup existing file before writing.
	if fileExisted {
		ts := time.Now().Format("20060102T150405")
		backupPath = opts.SettingsPath + ".backup." + ts

		if err := atomicCopy(opts.SettingsPath, backupPath); err != nil {
			return "", warnings, fmt.Errorf("backup opencode.json: %w", err)
		}
	}

	// Merge injected overlay into the existing content.
	var base []byte
	if fileExisted {
		base = existingContent
	}
	merged, err := MergeJSONObjects(base, injected)
	if err != nil {
		return backupPath, warnings, fmt.Errorf("merge opencode.json: %w", err)
	}

	// Write atomically.
	if err := os.MkdirAll(filepath.Dir(opts.SettingsPath), 0o755); err != nil {
		return backupPath, warnings, fmt.Errorf("create settings dir: %w", err)
	}

	tmp := opts.SettingsPath + ".tmp"
	if err := os.WriteFile(tmp, append(merged, '\n'), 0o644); err != nil {
		return backupPath, warnings, fmt.Errorf("write temp opencode.json: %w", err)
	}
	if err := os.Rename(tmp, opts.SettingsPath); err != nil {
		os.Remove(tmp)
		return backupPath, warnings, fmt.Errorf("rename opencode.json: %w", err)
	}

	return backupPath, warnings, nil
}
