package mapgen

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// RewriteMove rewrites every relative Markdown link in the .md files now
// living at {root}/openspec/{newRel} (post-move) so each still resolves to
// the same absolute target as when the folder lived at
// {root}/openspec/{oldRel}. oldRel/newRel are slash-separated paths relative
// to openspec/, e.g. "changes/x" and "changes/archive/2026-01-01-x".
// Wikilinks and .feature files are never touched. Writes are atomic.
func RewriteMove(root, oldRel, newRel string) error {
	newAbs := filepath.Join(root, "openspec", filepath.FromSlash(newRel))

	return filepath.WalkDir(newAbs, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}

		relFromNew, err := filepath.Rel(newAbs, p)
		if err != nil {
			return fmt.Errorf("rel %s: %w", p, err)
		}
		subDir := path.Dir(filepath.ToSlash(relFromNew))

		fileOldDir := path.Join(oldRel, subDir)
		fileNewDir := path.Join(newRel, subDir)

		data, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		content := string(data)

		// Idempotency guard: if every relative link already resolves from
		// this file's current location, the shift was already applied (or
		// never needed) — leave the file untouched so repeated Backfill
		// runs produce zero diff.
		if !hasDanglingRelLink(p, content) {
			return nil
		}

		rewritten := Rewrite(content, fileOldDir, fileNewDir)
		if rewritten == content {
			return nil
		}

		tmp := p + ".tmp"
		if err := os.WriteFile(tmp, []byte(rewritten), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", tmp, err)
		}
		if err := os.Rename(tmp, p); err != nil {
			return fmt.Errorf("rename %s: %w", tmp, err)
		}
		return nil
	})
}

// hasDanglingRelLink reports whether any relative Markdown link in content
// fails to resolve from srcPath's current location.
func hasDanglingRelLink(srcPath, content string) bool {
	for _, l := range FindRelLinks(content) {
		if _, ok := Resolve(srcPath, l.Target); !ok {
			return true
		}
	}
	return false
}
