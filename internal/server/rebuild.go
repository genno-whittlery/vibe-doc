package server

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/frontmatter"
	"github.com/genno-whittlery/vibe-doc/internal/search"
	"github.com/genno-whittlery/vibe-doc/internal/walk"
)

// rebuildSearchIndex re-populates s.searchIdx by walking each mount,
// parsing front-matter, and calling s.searchIdx.AddDoc for every .md
// file. Holds s.mu.Lock() while swapping the index pointer so concurrent
// readers see the fresh index atomically.
func (s *Server) rebuildSearchIndex() {
	idx := search.New()
	for _, m := range s.mounts.Mounts() {
		files, err := walk.MD(m.Root)
		if err != nil {
			s.log.Warn("walk %s: %v", m.Root, err)
			continue
		}
		for _, f := range files {
			absPath := filepath.Join(m.Root, f)
			body, err := os.ReadFile(absPath)
			if err != nil {
				continue
			}
			fm, mdBody, err := frontmatter.Parse(body)
			if err != nil {
				s.log.Warn("frontmatter %s: %v", absPath, err)
				mdBody = body
			}
			stem := strings.TrimSuffix(f, ".md")
			urlSuffix := stem
			// README.md → folder URL ("foo/README" → URL ".../foo/")
			if strings.HasSuffix(stem, "/README") || stem == "README" {
				urlSuffix = strings.TrimSuffix(stem, "README")
			}
			var urlPath string
			if m.URL == "/" {
				urlPath = "/" + urlSuffix
			} else {
				urlPath = m.URL + "/" + urlSuffix
			}
			title := fm.Title
			if title == "" {
				if h := firstH1(mdBody); h != "" {
					title = h
				} else {
					title = stem
				}
			}
			heads := extractHeadings(mdBody)
			idx.AddDoc(search.IndexedDoc{
				ID:       urlPath,
				URL:      urlPath,
				File:     f,
				Title:    title,
				Body:     string(mdBody),
				Headings: heads,
				Tags:     fm.Tags,
			})
		}
	}
	s.mu.Lock()
	s.searchIdx = idx
	s.mu.Unlock()
}

var (
	h1Re = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)
	hRe  = regexp.MustCompile(`(?m)^(#{1,3})\s+(.+?)\s*$`)
)

func firstH1(b []byte) string {
	if m := h1Re.FindSubmatch(b); len(m) >= 2 {
		return string(m[1])
	}
	return ""
}

func extractHeadings(b []byte) []search.HeadingHit {
	matches := hRe.FindAllSubmatch(b, -1)
	out := make([]search.HeadingHit, 0, len(matches))
	for _, m := range matches {
		out = append(out, search.HeadingHit{
			Level: len(m[1]),
			Text:  string(m[2]),
		})
	}
	return out
}
