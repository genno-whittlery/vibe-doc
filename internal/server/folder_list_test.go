package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genno-whittlery/vibe-doc/internal/config"
	"github.com/genno-whittlery/vibe-doc/internal/logger"
	"github.com/genno-whittlery/vibe-doc/internal/mount"
)

// minimal Server fixture for folderListing — sets cfg.Exclude and a
// no-op logger; nothing else from Server is touched by folderListing.
func newFolderTestServer(t *testing.T, exclude []string) *Server {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "log")
	lg, err := logger.New(logPath, 1<<10, logger.LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	return &Server{cfg: config.Config{Exclude: exclude}, log: lg, mounts: mount.New(nil)}
}

func TestFolderListingSiblingsAndSubdirs(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# Section"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "alpha.md"), []byte("# Alpha"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "beta.md"), []byte("+++\ntitle = \"Beta!\"\n+++"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "_hidden.md"), []byte("# Hidden by underscore"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "sub/README.md"), []byte("# Sub Group"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "node_modules"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "node_modules/should-skip.md"), []byte("x"), 0o644)

	s := newFolderTestServer(t, []string{"node_modules"})
	got := s.folderListing(filepath.Join(root, "README.md"), "/foo/")

	titles := []string{}
	for _, e := range got {
		titles = append(titles, e.Title)
	}
	joined := strings.Join(titles, ",")
	for _, want := range []string{"Sub Group", "Alpha", "Beta!"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in folder listing %v", want, titles)
		}
	}
	for _, unwant := range []string{"Hidden by underscore", "should-skip"} {
		if strings.Contains(joined, unwant) {
			t.Errorf("excluded entry %q surfaced: %v", unwant, titles)
		}
	}
	if !got[0].IsDir {
		t.Errorf("expected dirs sorted first; got %+v", got)
	}
	if !strings.HasSuffix(got[0].URL, "/sub/") {
		t.Errorf("dir URL should end with trailing slash; got %q", got[0].URL)
	}
}

func TestFolderListingNilWhenNotIndex(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "leaf.md"), []byte("# Leaf"), 0o644)
	s := newFolderTestServer(t, nil)
	got := s.folderListing(filepath.Join(root, "leaf.md"), "/foo/leaf")
	if got != nil {
		t.Errorf("expected nil for non-index file; got %+v", got)
	}
}
