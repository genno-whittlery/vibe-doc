package shadow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genno-whittlery/vibe-doc/internal/mount"
)

func TestReadmeIndexShadow(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "mocks"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "mocks/README.md"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "mocks/index.md"), []byte("b"), 0o644)
	conflicts := Scan(mount.New([]mount.Mount{{URL: "/", Root: root}}), nil)
	if len(conflicts) == 0 {
		t.Fatal("expected README/index shadow conflict")
	}
	found := false
	for _, c := range conflicts {
		if c.Kind == KindReadmeIndex {
			found = true
		}
	}
	if !found {
		t.Errorf("expected KindReadmeIndex; got %+v", conflicts)
	}
}

func TestMountOverlap(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "engine"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "engine/x.md"), []byte("x"), 0o644)
	conflicts := Scan(mount.New([]mount.Mount{
		{URL: "/", Root: root},
		{URL: "/engine", Root: t.TempDir()},
	}), nil)
	found := false
	for _, c := range conflicts {
		if c.Kind == KindMountOverlap {
			found = true
		}
	}
	if !found {
		t.Errorf("expected mount-overlap shadow; got %v", conflicts)
	}
}

func TestScanRespectsExclude(t *testing.T) {
	root := t.TempDir()
	// A real README/index shadow that SHOULD surface.
	_ = os.MkdirAll(filepath.Join(root, "real"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "real/README.md"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "real/index.md"), []byte("b"), 0o644)
	// A README/index shadow inside node_modules that should be SKIPPED.
	_ = os.MkdirAll(filepath.Join(root, "node_modules/pkg"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "node_modules/pkg/README.md"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "node_modules/pkg/index.md"), []byte("b"), 0o644)

	conflicts := Scan(mount.New([]mount.Mount{{URL: "/", Root: root}}), []string{"node_modules"})
	var realFound bool
	for _, c := range conflicts {
		if strings.Contains(c.Message, "node_modules") {
			t.Errorf("excluded node_modules surfaced shadow conflict: %s", c.Message)
		}
		if c.Kind == KindReadmeIndex {
			realFound = true
		}
	}
	if !realFound {
		t.Errorf("expected real README/index shadow to still surface")
	}
}

func TestSameRoot(t *testing.T) {
	root := t.TempDir()
	conflicts := Scan(mount.New([]mount.Mount{
		{URL: "/a", Root: root},
		{URL: "/b", Root: root},
	}), nil)
	found := false
	for _, c := range conflicts {
		if c.Kind == KindSameRoot {
			found = true
		}
	}
	if !found {
		t.Errorf("expected same-root shadow; got %v", conflicts)
	}
}
