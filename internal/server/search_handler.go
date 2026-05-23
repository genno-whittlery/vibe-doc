package server

import (
	"encoding/json"
	"net/http"
	"time"
)

const searchTimeout = 50 * time.Millisecond

// handleSearch serves /__search?q=... as JSON. The 50ms timeout caps a
// runaway query so a slow goroutine can't tie up the handler.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	s.mu.RLock()
	idx := s.searchIdx
	s.mu.RUnlock()
	results, truncated := idx.SearchWithTimeout(q, searchTimeout)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"query":     q,
		"truncated": truncated,
		"results":   results,
	})
}
