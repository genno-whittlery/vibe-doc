package server

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// watchMounts starts an fsnotify watcher for every mount root (recursive)
// and broadcasts a debounced reload event on changes. Spec §3: events are
// debounced 100ms to absorb editor save bursts.
func (s *Server) watchMounts() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	add := func(p string) {
		if err := w.Add(p); err != nil {
			s.log.Warn("watch %s: %v", p, err)
		}
	}
	for _, m := range s.mounts.Mounts() {
		_ = filepath.WalkDir(m.Root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				add(p)
			}
			return nil
		})
	}
	// Coalesce burst events into a single rebuild + broadcast (100ms window).
	var (
		mu      sync.Mutex
		pending bool
	)
	fire := func() {
		mu.Lock()
		if pending {
			mu.Unlock()
			return
		}
		pending = true
		mu.Unlock()
		time.AfterFunc(100*time.Millisecond, func() {
			mu.Lock()
			pending = false
			mu.Unlock()
			s.rebuildSidebar()
			s.rebuildSearchIndex()
			s.sse.Broadcast()
			s.log.Info("fs event: rebuilt sidebar + search index")
		})
	}
	go func() {
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if ev.Op&fsnotify.Create != 0 {
					if st, err := os.Stat(ev.Name); err == nil && st.IsDir() {
						add(ev.Name)
					}
				}
				fire()
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				s.log.Warn("watcher: %v", err)
			}
		}
	}()
	return nil
}
