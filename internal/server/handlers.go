package server

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/frontmatter"
	"github.com/genno-whittlery/vibe-doc/internal/render"
	"github.com/genno-whittlery/vibe-doc/internal/route"
	"github.com/genno-whittlery/vibe-doc/internal/sidebar"
)

var rendererSingleton = render.New()

type pageData struct {
	SiteTitle  string
	Title      string
	HTML       template.HTML
	TOC        []render.TOCEntry
	Sidebar    []sidebar.Node
	HasMermaid bool
	HasMath    bool
}

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	m, sub, ok := s.mounts.Match(r.URL.Path)
	if !ok {
		s.serve404(w, r)
		return
	}
	if strings.HasPrefix(sub, "static/") {
		s.serveDocStatic(w, r, m.Root, strings.TrimPrefix(sub, "static/"))
		return
	}
	res, err := route.Resolve(m.Root, sub)
	if err != nil {
		http.Error(w, "internal error", 500)
		s.log.Error("resolve %s: %v", r.URL.Path, err)
		return
	}
	switch res.Kind {
	case route.KindNotFound:
		s.serve404(w, r)
	case route.KindRedirect:
		loc := m.URL + res.Location
		if m.URL == "/" {
			loc = res.Location
		}
		http.Redirect(w, r, loc, 301)
	case route.KindFile:
		s.renderMarkdownFile(w, r, res.AbsPath)
	case route.KindDirListing:
		s.renderDirListing(w, r, res.AbsPath)
	}
}

func (s *Server) renderMarkdownFile(w http.ResponseWriter, r *http.Request, absPath string) {
	body, err := os.ReadFile(absPath)
	if err != nil {
		http.Error(w, "read error", 500)
		s.log.Error("read %s: %v", absPath, err)
		return
	}
	fm, mdBody, err := frontmatter.Parse(body)
	if err != nil {
		s.log.Warn("frontmatter %s: %v", absPath, err)
		mdBody = body
	}
	html, toc, err := rendererSingleton.Render(mdBody)
	if err != nil {
		http.Error(w, "render error", 500)
		s.log.Error("render %s: %v", absPath, err)
		return
	}

	htmlStr := string(html)
	data := pageData{
		SiteTitle:  "vibe-doc",
		Title:      pickTitle(fm.Title, toc),
		HTML:       template.HTML(htmlStr),
		TOC:        toc,
		Sidebar:    s.sidebar,
		HasMermaid: strings.Contains(htmlStr, `class="mermaid"`),
		HasMath:    strings.Contains(htmlStr, `class="math`),
	}

	var buf bytes.Buffer
	if err := s.tpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		http.Error(w, "template error", 500)
		s.log.Error("template %s: %v", absPath, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Vibe-Mount", strings.TrimRight(s.mountForPath(r.URL.Path), "/"))
	_, _ = w.Write(buf.Bytes())
}

func (s *Server) mountForPath(p string) string {
	m, _, _ := s.mounts.Match(p)
	return m.URL
}

func pickTitle(fmTitle string, toc []render.TOCEntry) string {
	if fmTitle != "" {
		return fmTitle
	}
	if len(toc) > 0 {
		return toc[0].Text
	}
	return ""
}

func (s *Server) serveDocStatic(w http.ResponseWriter, r *http.Request, mountRoot, rel string) {
	full := filepath.Join(mountRoot, "static", rel)
	if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(mountRoot)+string(os.PathSeparator)) {
		http.NotFound(w, r)
		s.log.Warn("doc-static escape attempt: %s", r.URL.Path)
		return
	}
	http.ServeFile(w, r, full)
}

func (s *Server) renderDirListing(w http.ResponseWriter, r *http.Request, absPath string) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, "read error", 500)
		s.log.Error("listing %s: %v", absPath, err)
		return
	}
	var html strings.Builder
	fmt.Fprintf(&html, `<h1>%s</h1><ul>`, r.URL.Path)
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() {
			fmt.Fprintf(&html, `<li><a href="%s%s/">%s/</a></li>`, r.URL.Path, name, name)
		} else if strings.HasSuffix(name, ".md") {
			stem := strings.TrimSuffix(name, ".md")
			fmt.Fprintf(&html, `<li><a href="%s%s">%s</a></li>`, r.URL.Path, stem, stem)
		}
	}
	html.WriteString("</ul>")

	data := pageData{
		SiteTitle: "vibe-doc",
		Title:     path.Base(r.URL.Path),
		HTML:      template.HTML(html.String()),
		Sidebar:   s.sidebar,
	}
	var buf bytes.Buffer
	if err := s.tpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		http.Error(w, "template error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (s *Server) serve404(w http.ResponseWriter, r *http.Request) {
	suggestions := s.searchIdx.Search(path.Base(r.URL.Path))
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><link rel=stylesheet href="/__static/style.css"></head><body><main style="padding:2rem;max-width:800px;margin:0 auto"><h1>404 — Not Found</h1><p>No doc at <code>%s</code>.</p>`, r.URL.Path)
	if len(suggestions) > 0 {
		fmt.Fprintln(w, `<p><strong>Did you mean…</strong></p><ul>`)
		for _, sg := range suggestions {
			fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, sg.URL, sg.Title)
		}
		fmt.Fprintln(w, "</ul>")
	}
	fmt.Fprintln(w, "</main></body></html>")
	s.log.Info("404 %s (%d suggestions)", r.URL.Path, len(suggestions))
}
