// Package sidebar builds the auto-generated navigation tree from the
// filesystem. Spec ref: §6.
package sidebar

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/frontmatter"
)

// Node is one entry in the sidebar tree. A Node is either a leaf .md file
// (IsLeaf=true), a group / directory (IsLeaf=false, may have Children), or
// a root / mount (IsLeaf=false, Title from mount Display).
type Node struct {
	Title    string
	URL      string
	IsLeaf   bool
	Order    int
	Hidden   bool
	Children []Node
}

// Build constructs a Node tree for one mount.
//
// urlPrefix is the mount's URL (e.g. "/" or "/engine"). root is the
// filesystem path the mount is rooted at. display is the mount's display
// name (sidebar section title); falls back to the prefix basename.
// exclude lists directory basenames to skip entirely.
func Build(display, urlPrefix, root string, exclude []string) (Node, error) {
	if display == "" {
		display = mountTitle(urlPrefix)
	}
	urlPrefix = strings.TrimRight(urlPrefix, "/")
	tree := Node{
		Title: display,
		URL:   urlPrefix + "/",
	}
	children, err := buildChildren(urlPrefix, root, exclude)
	if err != nil {
		return Node{}, err
	}
	tree.Children = children
	return tree, nil
}

func isExcluded(name string, exclude []string) bool {
	for _, e := range exclude {
		if e == name {
			return true
		}
	}
	return false
}

func mountTitle(prefix string) string {
	b := path.Base(strings.TrimSuffix(prefix, "/"))
	if b == "/" || b == "." || b == "" {
		return "Docs"
	}
	return b
}

func buildChildren(urlPrefix, dir string, exclude []string) ([]Node, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Node
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		if isExcluded(name, exclude) {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Stat(full)
		if err != nil {
			continue
		}
		switch {
		case info.IsDir():
			children, err := buildChildren(urlPrefix+"/"+name, full, exclude)
			if err != nil {
				return nil, err
			}
			groupTitle, hidden := dirGroupMeta(full, name)
			if hidden {
				continue
			}
			out = append(out, Node{
				Title:    groupTitle,
				URL:      urlPrefix + "/" + name + "/",
				Children: children,
			})
		case strings.EqualFold(name, "README.md"):
			// README is used for the parent group's title; not a separate leaf.
			continue
		case strings.HasSuffix(name, ".md"):
			leaf, ok := buildLeaf(full, urlPrefix, name)
			if ok {
				out = append(out, leaf)
			}
		}
	}
	sortChildren(out)
	return out, nil
}

func dirGroupMeta(dir, dirName string) (title string, hidden bool) {
	readme := filepath.Join(dir, "README.md")
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
	if h := extractFirstH1(mdBody); h != "" {
		return h, false
	}
	return dirName, false
}

func buildLeaf(absPath, urlPrefix, filename string) (Node, bool) {
	body, err := os.ReadFile(absPath)
	if err != nil {
		return Node{}, false
	}
	fm, mdBody, err := frontmatter.Parse(body)
	if err != nil {
		// Recoverable per §7.1: render without front-matter.
		fm = frontmatter.Front{}
		mdBody = body
	}
	if fm.Hidden {
		return Node{}, false
	}
	title := fm.Title
	if title == "" {
		title = extractFirstH1(mdBody)
	}
	if title == "" {
		title = strings.TrimSuffix(filename, ".md")
	}
	return Node{
		Title:  title,
		URL:    urlPrefix + "/" + strings.TrimSuffix(filename, ".md"),
		IsLeaf: true,
		Order:  fm.Order,
	}, true
}

var h1Re = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)

func extractFirstH1(b []byte) string {
	m := h1Re.FindSubmatch(b)
	if len(m) >= 2 {
		return string(m[1])
	}
	return ""
}

func sortChildren(nodes []Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		oi, oj := nodes[i].Order, nodes[j].Order
		if oi != 0 || oj != 0 {
			if oi == 0 {
				return false
			}
			if oj == 0 {
				return true
			}
			if oi != oj {
				return oi < oj
			}
		}
		return strings.ToLower(nodes[i].Title) < strings.ToLower(nodes[j].Title)
	})
}
