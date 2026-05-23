package server

// rebuildSearchIndex re-populates s.searchIdx by walking each mount,
// parsing front-matter, and calling s.searchIdx.AddDoc. STUB: no-op.
// Task 13 fills in the walker → frontmatter → render → AddDoc pipeline,
// and MUST hold s.mu.Lock() while swapping the index pointer.
func (s *Server) rebuildSearchIndex() {}
