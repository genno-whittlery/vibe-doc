package server

import "net/http"

// SSEHub broadcasts reload events to connected /__events listeners. STUB:
// Task 14 fills in fanout, subscribe/unsubscribe, and message queue.
type SSEHub struct{}

func newSSEHub() *SSEHub { return &SSEHub{} }

// handleSSE serves the /__events stream. STUB: closes the connection
// immediately. Task 14 keeps the stream open and writes "data: reload\n\n"
// on debounced fsnotify events.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)
}
