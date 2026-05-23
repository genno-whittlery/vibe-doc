package server

import "net/http"

// handleSearch serves /__search?q=... JSON. STUB: returns 501. Task 13
// queries s.searchIdx with a 50ms timeout and returns ranked hits.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "search: not yet implemented", http.StatusNotImplemented)
}
