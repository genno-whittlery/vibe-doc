package server

import "net/http"

// handleSitemap serves /sitemap.xml. STUB: returns 501. Task 15 walks the
// mounts and emits XML urlset.
func (s *Server) handleSitemap(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "sitemap: not yet implemented", http.StatusNotImplemented)
}
