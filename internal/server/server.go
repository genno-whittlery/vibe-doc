// Package server wires HTTP routes for vibe-doc.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"sync"

	"github.com/genno-whittlery/vibe-doc/assets"
	"github.com/genno-whittlery/vibe-doc/internal/config"
	"github.com/genno-whittlery/vibe-doc/internal/logger"
	"github.com/genno-whittlery/vibe-doc/internal/mount"
	"github.com/genno-whittlery/vibe-doc/internal/search"
	"github.com/genno-whittlery/vibe-doc/internal/shadow"
	"github.com/genno-whittlery/vibe-doc/internal/sidebar"
)

type Server struct {
	cfg       config.Config
	log       *logger.Logger
	mounts    *mount.Set
	mux       *http.ServeMux
	tpl       *template.Template
	sidebar   []sidebar.Node
	searchIdx *search.Index
	sse       *SSEHub
	mu        sync.RWMutex
}

func New(cfg config.Config, log *logger.Logger) (*Server, error) {
	ms := make([]mount.Mount, len(cfg.Mounts))
	for i, m := range cfg.Mounts {
		ms[i] = mount.Mount{URL: m.URL, Root: m.Root, Display: m.Display}
	}
	s := &Server{
		cfg:    cfg,
		log:    log,
		mounts: mount.New(ms),
		mux:    http.NewServeMux(),
		sse:    newSSEHub(),
	}
	// Template-arithmetic helpers. Go's html/template won't auto-convert
	// between numeric types — `.Level` arrives as int, literals like `1`
	// are int, and `0.7` is float64. Accept `any` and coerce so the
	// `{{mul (sub .Level 1) 0.7}}` expression in base.html works without
	// requiring every TOC entry to be float64.
	toFloat := func(v any) float64 {
		switch x := v.(type) {
		case float64:
			return x
		case float32:
			return float64(x)
		case int:
			return float64(x)
		case int64:
			return float64(x)
		case int32:
			return float64(x)
		default:
			return 0
		}
	}
	funcs := template.FuncMap{
		"mul": func(a, b any) float64 { return toFloat(a) * toFloat(b) },
		"sub": func(a, b any) float64 { return toFloat(a) - toFloat(b) },
	}
	tpl, err := template.New("base").Funcs(funcs).ParseFS(assets.FS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("server: parse templates: %w", err)
	}
	s.tpl = tpl

	s.rebuildSidebar()

	s.searchIdx = search.New()
	s.rebuildSearchIndex()

	// Spec §7: scan for shadow conflicts at startup and log each one. With
	// --strict (cfg.Strict), promote to fatal.
	conflicts := shadow.Scan(s.mounts, s.cfg.Exclude)
	for _, c := range conflicts {
		s.log.Warn("shadow %s: %s", c.Kind, c.Message)
	}
	if cfg.Strict && len(conflicts) > 0 {
		return nil, fmt.Errorf("server: %d shadow conflict(s) under --strict", len(conflicts))
	}

	s.routes()
	return s, nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("/__health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})
	staticSub, err := fs.Sub(assets.FS, "static")
	if err != nil {
		panic("vibe-doc: assets.FS missing 'static' subdir: " + err.Error())
	}
	s.mux.Handle("/__static/", http.StripPrefix("/__static/", http.FileServer(http.FS(staticSub))))
	s.mux.HandleFunc("/__events", s.handleSSE)
	s.mux.HandleFunc("/__shadow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(shadow.Scan(s.mounts, s.cfg.Exclude))
	})
	s.mux.HandleFunc("/__search", s.handleSearch)
	s.mux.HandleFunc("/sitemap.xml", s.handleSitemap)
	s.mux.HandleFunc("/", s.handlePage)
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.cfg.Port)
	srv := &http.Server{Addr: addr, Handler: s.mux}
	if err := s.watchMounts(); err != nil {
		s.log.Warn("watch: %v", err)
	}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	s.log.Info("vibe-doc serving %d mount(s) on http://%s", len(s.mounts.Mounts()), addr)
	fmt.Printf("vibe-doc serving %d mount(s) on http://%s\n", len(s.mounts.Mounts()), addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// rebuildSidebar walks each mount and rebuilds s.sidebar. STUB-friendly:
// uses sidebar.Build which is a stub today; Task 10 fills it in.
func (s *Server) rebuildSidebar() {
	var trees []sidebar.Node
	for _, m := range s.mounts.Mounts() {
		tree, err := sidebar.Build(m.Display, m.URL, m.Root, s.cfg.Exclude)
		if err != nil {
			s.log.Warn("sidebar %s: %v", m.URL, err)
			continue
		}
		trees = append(trees, tree)
	}
	s.mu.Lock()
	s.sidebar = trees
	s.mu.Unlock()
}
