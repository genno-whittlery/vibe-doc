package server

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/frontmatter"
)

// FolderEntry is one child of a folder-index page (the README.md / index.md
// served at a directory URL). Used by base.html to render a "Folder
// contents" appendix below the README body.
type FolderEntry struct {
	Title string
	URL   string
	IsDir bool
}

// folderListing returns the sibling files + subdirectories of indexFile,
// each resolved with its display title. urlPath is the directory URL (with
// trailing slash) that links should be built relative to. Returns nil if
// indexFile is not a README.md / index.md page.
func (s *Server) folderListing(indexFile, urlPath string) []FolderEntry {
	base := filepath.Base(indexFile)
	if !strings.EqualFold(base, "README.md") && !strings.EqualFold(base, "index.md") {
		return nil
	}
	if !strings.HasSuffix(urlPath, "/") {
		urlPath += "/"
	}
	dir := filepath.Dir(indexFile)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]FolderEntry, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		if isExcluded(name, s.cfg.Exclude) {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Stat(full)
		if err != nil {
			continue
		}
		switch {
		case info.IsDir():
			title, hidden := dirTitleFromReadme(full, name)
			if hidden {
				continue
			}
			out = append(out, FolderEntry{Title: title, URL: urlPath + name + "/", IsDir: true})
		case strings.HasSuffix(name, ".md"):
			if strings.EqualFold(name, "README.md") || strings.EqualFold(name, "index.md") {
				continue
			}
			title, hidden := leafTitle(full, name)
			if hidden {
				continue
			}
			stem := strings.TrimSuffix(name, ".md")
			out = append(out, FolderEntry{Title: title, URL: urlPath + stem, IsDir: false})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir // dirs first
		}
		return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
	})
	return out
}

func isExcluded(name string, exclude []string) bool {
	for _, e := range exclude {
		if e == name {
			return true
		}
	}
	return false
}

var folderH1Re = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)

func dirTitleFromReadme(absDir, dirName string) (title string, hidden bool) {
	readme := filepath.Join(absDir, "README.md")
	body, err := os.ReadFile(readme)
	if err != nil {
		return dirName, false
	}
	fm, mdBody, err := frontmatter.Parse(body)
	if err != nil {
		return dirName, false
	}
	if fm.Hidden {
		return "", true
	}
	if fm.Title != "" {
		return fm.Title, false
	}
	if m := folderH1Re.FindSubmatch(mdBody); len(m) >= 2 {
		return string(m[1]), false
	}
	return dirName, false
}

func leafTitle(absPath, filename string) (title string, hidden bool) {
	body, err := os.ReadFile(absPath)
	if err != nil {
		return strings.TrimSuffix(filename, ".md"), false
	}
	fm, mdBody, err := frontmatter.Parse(body)
	if err != nil {
		return strings.TrimSuffix(filename, ".md"), false
	}
	if fm.Hidden {
		return "", true
	}
	if fm.Title != "" {
		return fm.Title, false
	}
	if m := folderH1Re.FindSubmatch(mdBody); len(m) >= 2 {
		return string(m[1]), false
	}
	return strings.TrimSuffix(filename, ".md"), false
}
