package server

import (
	"fmt"
	"net/http"

	"github.com/genno-whittlery/vibe-doc/internal/sitemap"
)

// handleSitemap serves /sitemap.xml — a flat urlset over every .md file
// across every mount.
func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	baseURL := fmt.Sprintf("http://%s", r.Host)
	if err := sitemap.Generate(w, baseURL, s.mounts, s.cfg.Exclude); err != nil {
		s.log.Error("sitemap: %v", err)
	}
}
