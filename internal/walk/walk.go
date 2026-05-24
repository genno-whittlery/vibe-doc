// Package walk traverses a doc-tree directory, following directory
// symlinks (so the engine-docs symlink target gets walked) but detecting
// cycles via a resolved-inode visited set. Spec ref: §4 (symlinks work).
package walk

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MD returns paths to all .md files under root, relative to root, using
// forward slashes. Symlinked directories are followed; symlink cycles are
// detected and skipped. Directory entries whose basename appears in
// exclude are skipped (and not recursed into).
func MD(root string, exclude []string) ([]string, error) {
	root = filepath.Clean(root)
	visited := map[string]struct{}{}
	var out []string
	err := walkDir(root, root, exclude, visited, &out)
	return out, err
}

func isExcluded(name string, exclude []string) bool {
	for _, e := range exclude {
		if e == name {
			return true
		}
	}
	return false
}

func walkDir(root, dir string, exclude []string, visited map[string]struct{}, out *[]string) error {
	// De-dup by canonical (symlink-resolved) directory path.
	canon, err := filepath.EvalSymlinks(dir)
	if err != nil {
		// Dangling symlink — skip silently; the caller emits a WARN via the
		// shadow detector.
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if _, seen := visited[canon]; seen {
		return nil
	}
	visited[canon] = struct{}{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if isExcluded(e.Name(), exclude) {
			continue
		}
		full := filepath.Join(dir, e.Name())
		info, err := os.Stat(full) // follows symlinks
		if err != nil {
			continue // dangling — skip
		}
		switch {
		case info.IsDir():
			if err := walkDir(root, full, exclude, visited, out); err != nil {
				return err
			}
		case strings.HasSuffix(e.Name(), ".md"):
			rel, _ := filepath.Rel(root, full)
			*out = append(*out, filepath.ToSlash(rel))
		}
	}
	return nil
}
