package opencode

import "encoding/json"

// replacesentinel is the key used as a sentinel inside a nested object to signal
// that the parent key should be replaced wholesale rather than deep-merged.
// Example: {"permission": {"task": {"__replace__": {"*": "deny", ...}}}}
// causes the "task" value to be replaced with the inner map.
const replacesentinel = "__replace__"

// MergeJSONObjects deep-merges overlay bytes into base bytes. Rules:
//   - Both base and overlay are expected to be JSON objects ({…}).
//   - Merge is additive: keys present only in base are preserved.
//   - Overlay keys win on collision for scalar/non-object values.
//   - Nested object values are merged recursively.
//   - When an overlay value is an object containing only the key "__replace__",
//     the corresponding base value is replaced wholesale with the "__replace__"
//     child, not merged.
//   - A missing, empty, or malformed base is treated as {} (no error, no crash).
func MergeJSONObjects(base, overlay []byte) ([]byte, error) {
	baseMap := make(map[string]any)

	if len(base) > 0 {
		// Tolerate malformed base by ignoring unmarshal errors.
		_ = json.Unmarshal(base, &baseMap)
	}

	var overlayMap map[string]any
	if err := json.Unmarshal(overlay, &overlayMap); err != nil {
		return nil, err
	}

	merged := mergeObjects(baseMap, overlayMap)

	return json.MarshalIndent(merged, "", "  ")
}

// mergeObjects recursively merges src (overlay) into dst (base).
// Returns a new map; dst and src are not mutated.
func mergeObjects(dst, src map[string]any) map[string]any {
	result := make(map[string]any, len(dst)+len(src))

	// Copy all base keys first.
	for k, v := range dst {
		result[k] = v
	}

	// Merge overlay keys.
	for k, srcVal := range src {
		srcObj, srcIsObj := srcVal.(map[string]any)

		// Check for __replace__ sentinel: {"some_key": {"__replace__": {...}}}
		// means replace result[k] with the value of the __replace__ child.
		if srcIsObj {
			if replaceVal, isReplace := srcObj[replacesentinel]; len(srcObj) == 1 && isReplace {
				result[k] = replaceVal
				continue
			}
		}

		dstVal, dstExists := result[k]
		dstObj, dstIsObj := dstVal.(map[string]any)

		if srcIsObj && dstExists && dstIsObj {
			// Both sides are objects — recurse.
			result[k] = mergeObjects(dstObj, srcObj)
		} else {
			// Scalar, array, or base did not have this key — overlay wins.
			result[k] = srcVal
		}
	}

	return result
}
