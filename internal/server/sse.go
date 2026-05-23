package server

import (
	"fmt"
	"net/http"
	"sync"
)

// SSEHub fans out reload events to connected browser tabs.
type SSEHub struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func newSSEHub() *SSEHub {
	return &SSEHub{subs: map[chan struct{}]struct{}{}}
}

func (h *SSEHub) Subscribe() chan struct{} {
	ch := make(chan struct{}, 4)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *SSEHub) Unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.subs, ch)
	close(ch)
	h.mu.Unlock()
}

func (h *SSEHub) Broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// handleSSE serves the /__events stream. Sends "data: reload" on every
// broadcast and closes when the client disconnects.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := s.sse.Subscribe()
	defer s.sse.Unsubscribe(ch)
	fmt.Fprintln(w, "data: hello")
	flusher.Flush()
	for {
		select {
		case <-ch:
			fmt.Fprint(w, "data: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
