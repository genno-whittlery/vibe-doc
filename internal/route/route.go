// Package route resolves a URL sub-path (mount-relative) to a disk
// location. The Resolve function is the implementation of the §4 routing
// table from the spec.
package route

import (
	"os"
	"path/filepath"
	"strings"
)

// Kind describes what Resolve found.
type Kind int

const (
	KindNotFound   Kind = 0
	KindFile       Kind = 1 // serve AbsPath (a .md file)
	KindRedirect   Kind = 2 // 301 to Location
	KindDirListing Kind = 3 // auto-generate listing of AbsPath dir
)

func (k Kind) String() string {
	switch k {
	case KindFile:
		return "file"
	case KindRedirect:
		return "redirect"
	case KindDirListing:
		return "dir-listing"
	default:
		return "not-found"
	}
}

// Result is what Resolve returns.
type Result struct {
	Kind     Kind
	AbsPath  string // file path on disk (for File or DirListing)
	Location string // URL to redirect to (for Redirect; mount-prefix is added by caller)
}

// Resolve maps a URL sub-path (no leading slash; possible trailing slash)
// against a filesystem root and returns the §4 resolution.
//
// Examples (root = /docs):
//
//	""              → file /docs/README.md
//	"foo"           → file /docs/foo.md (if exists)
//	"foo"           → redirect "/foo/" (if /docs/foo is a dir)
//	"foo/"          → file /docs/foo/README.md → index.md → dir listing
//	"foo/bar"       → file /docs/foo/bar.md
//	"missing"       → not found
//
// sub MUST be URL-decoded by the caller; Resolve does not percent-decode.
// os.Stat follows symlinks, so a symlink inside root that points outside
// root will be served — caller is responsible for choosing safe roots.
func Resolve(root, sub string) (Result, error) {
	// filepath.Clean collapses ".." segments before the path ever reaches
	// disk — this is the traversal defense, not any later check.
	cleaned := filepath.ToSlash(filepath.Clean("/" + sub))
	if cleaned == "/" {
		cleaned = ""
	} else {
		cleaned = strings.TrimPrefix(cleaned, "/")
	}

	trailingSlash := strings.HasSuffix(sub, "/") || sub == ""

	if trailingSlash {
		// Folder URL: try README.md, then index.md, then dir listing.
		dir := filepath.Join(root, cleaned)
		st, err := os.Stat(dir)
		if err != nil || !st.IsDir() {
			return Result{Kind: KindNotFound}, nil
		}
		for _, name := range []string{"README.md", "index.md"} {
			candidate := filepath.Join(dir, name)
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
				return Result{Kind: KindFile, AbsPath: candidate}, nil
			}
		}
		return Result{Kind: KindDirListing, AbsPath: dir}, nil
	}

	// Bare URL: prefer file (sub.md), else redirect-to-dir if a dir exists.
	candidate := filepath.Join(root, cleaned+".md")
	if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
		return Result{Kind: KindFile, AbsPath: candidate}, nil
	}
	dir := filepath.Join(root, cleaned)
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return Result{Kind: KindRedirect, Location: "/" + cleaned + "/"}, nil
	}
	return Result{Kind: KindNotFound}, nil
}
