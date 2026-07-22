package mapgen

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// relLinkRe matches Markdown inline links: [text](target). Wikilinks
// ([[capability]]) never match — they lack the required parenthesized target.
var relLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)\)`)

// Link is a single relative Markdown link span found by FindRelLinks.
type Link struct {
	Raw    string // full matched span, e.g. "[design](design.md)"
	Text   string
	Target string // the parenthesized target, e.g. "design.md" or "../proposal.md"
}

// FindRelLinks extracts every relative Markdown link ([text](target)) in md.
// Wikilinks, absolute URLs, mailto:, site-absolute paths, and same-page
// anchors ('#...') are ignored. Link syntax that appears inside fenced code
// blocks or inline code spans (illustrative examples in prose) is ignored
// too — it is not a real link.
func FindRelLinks(md string) []Link {
	masked := maskCodeRegions(md)

	var links []Link
	for _, m := range relLinkRe.FindAllStringSubmatch(masked, -1) {
		target := m[2]
		if isAbsoluteLink(target) {
			continue
		}
		links = append(links, Link{Raw: m[0], Text: m[1], Target: target})
	}
	return links
}

// maskCodeRegions blanks out (replaces with spaces, preserving line
// structure) the contents of fenced code blocks (``` or ~~~ fences) and
// inline code spans (backtick-delimited text on a single line), so link
// syntax used as a documentation example is never mistaken for a real link.
func maskCodeRegions(md string) string {
	lines := strings.Split(md, "\n")

	inFence := false
	var fenceChar byte
	fenceLen := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inFence {
			lines[i] = blank(line)
			if fenceRunLen(trimmed, fenceChar) >= fenceLen {
				inFence = false
			}
			continue
		}

		if c, n := fenceMarker(trimmed); n >= 3 {
			fenceChar, fenceLen, inFence = c, n, true
			lines[i] = blank(line)
			continue
		}

		lines[i] = maskInlineCode(line)
	}

	return strings.Join(lines, "\n")
}

// maskInlineCode blanks backtick-delimited code spans on a single line,
// e.g. `` `[text](target)` `` or ``` ``code`` ```. Per CommonMark, a code
// span opens with a run of N backticks and closes at the next run of
// exactly N backticks (runs of a different length are content, not a
// closer) — a plain regex can't express that "run lengths must match"
// rule without backreferences, so this scans by hand.
func maskInlineCode(line string) string {
	b := []byte(line)
	for i := 0; i < len(b); {
		if b[i] != '`' {
			i++
			continue
		}
		openEnd := i
		for openEnd < len(b) && b[openEnd] == '`' {
			openEnd++
		}
		delimLen := openEnd - i

		closeStart := -1
		for k := openEnd; k < len(b); {
			if b[k] != '`' {
				k++
				continue
			}
			runEnd := k
			for runEnd < len(b) && b[runEnd] == '`' {
				runEnd++
			}
			if runEnd-k == delimLen {
				closeStart = k
				break
			}
			k = runEnd
		}

		if closeStart < 0 {
			// No matching closer on this line: not a code span.
			i = openEnd
			continue
		}
		closeEnd := closeStart + delimLen
		for x := i; x < closeEnd; x++ {
			b[x] = ' '
		}
		i = closeEnd
	}
	return string(b)
}

// fenceMarker reports the fence character (` or ~) and run length if
// trimmed opens or closes a fenced code block, e.g. "```" or "~~~~".
func fenceMarker(trimmed string) (byte, int) {
	if trimmed == "" {
		return 0, 0
	}
	c := trimmed[0]
	if c != '`' && c != '~' {
		return 0, 0
	}
	return c, fenceRunLen(trimmed, c)
}

// fenceRunLen returns how many leading bytes of s equal c.
func fenceRunLen(s string, c byte) int {
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	return n
}

// blank replaces every byte of s with a space, keeping its length so line
// structure (and thus later line-based scanning) is preserved.
func blank(s string) string {
	return strings.Repeat(" ", len(s))
}

func isAbsoluteLink(target string) bool {
	switch {
	case target == "":
		return true
	case strings.HasPrefix(target, "#"):
		return true
	case strings.HasPrefix(target, "/"):
		return true
	case strings.HasPrefix(target, "http://"), strings.HasPrefix(target, "https://"):
		return true
	case strings.HasPrefix(target, "mailto:"):
		return true
	default:
		return false
	}
}

// Resolve resolves link (found in the file at srcPath) against srcPath's
// directory on the real filesystem, matching how Generate/Check touch disk.
// A pure same-file anchor (e.g. "#section") resolves to ok=true.
func Resolve(srcPath, link string) (target string, ok bool) {
	clean := link
	if i := strings.IndexByte(clean, '#'); i >= 0 {
		clean = clean[:i]
	}
	if clean == "" {
		return "", true
	}

	dir := filepath.Dir(srcPath)
	target = filepath.Clean(filepath.Join(dir, filepath.FromSlash(clean)))

	if _, err := os.Stat(target); err != nil {
		return target, false
	}
	return target, true
}

// Rewrite recomputes every relative link in md so it still resolves to the
// same absolute target after the containing file moves from oldDir to
// newDir (slash-separated paths relative to a common root). Only matched
// link spans are edited; wikilinks and prose are left byte-identical.
func Rewrite(md, oldDir, newDir string) string {
	return relLinkRe.ReplaceAllStringFunc(md, func(match string) string {
		sub := relLinkRe.FindStringSubmatch(match)
		text, target := sub[1], sub[2]
		if isAbsoluteLink(target) {
			return match
		}

		frag := ""
		clean := target
		if i := strings.IndexByte(clean, '#'); i >= 0 {
			frag = clean[i:]
			clean = clean[:i]
		}
		if clean == "" {
			return match
		}

		abs := path.Clean(path.Join(oldDir, clean))
		newTarget := relPath(newDir, abs) + frag
		return "[" + text + "](" + newTarget + ")"
	})
}

// relPath returns the slash-separated relative path from dir to target,
// both given as slash-separated paths relative to a common root.
func relPath(dir, target string) string {
	dirParts := splitClean(dir)
	targetParts := splitClean(target)

	i := 0
	for i < len(dirParts) && i < len(targetParts) && dirParts[i] == targetParts[i] {
		i++
	}

	var parts []string
	for range dirParts[i:] {
		parts = append(parts, "..")
	}
	parts = append(parts, targetParts[i:]...)
	if len(parts) == 0 {
		return "."
	}
	return strings.Join(parts, "/")
}

func splitClean(p string) []string {
	p = path.Clean(p)
	if p == "." || p == "" {
		return nil
	}
	return strings.Split(p, "/")
}
