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
// anchors ('#...') are ignored.
func FindRelLinks(md string) []Link {
	var links []Link
	for _, m := range relLinkRe.FindAllStringSubmatch(md, -1) {
		target := m[2]
		if isAbsoluteLink(target) {
			continue
		}
		links = append(links, Link{Raw: m[0], Text: m[1], Target: target})
	}
	return links
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
