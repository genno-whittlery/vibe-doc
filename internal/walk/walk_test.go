package walk

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func TestWalkBasic(t *testing.T) {
	root := t.TempDir()
	mk := func(p string) {
		full := filepath.Join(root, p)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		_ = os.WriteFile(full, []byte("x"), 0o644)
	}
	mk("a.md")
	mk("sub/b.md")
	mk("sub/deep/c.md")
	mk("_hidden.md") // underscore-prefixed; still walked, filter is consumer's job

	got, err := MD(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"_hidden.md", "a.md", "sub/b.md", "sub/deep/c.md"}
	if !equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWalkFollowsDirSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require admin on Windows")
	}
	root := t.TempDir()
	target := t.TempDir()
	_ = os.WriteFile(filepath.Join(target, "z.md"), []byte("z"), 0o644)
	if err := os.Symlink(target, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	got, err := MD(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"linked/z.md"}
	if !equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWalkDetectsSymlinkLoop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require admin on Windows")
	}
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "a.md"), []byte("a"), 0o644)
	// loop: root/self → root
	if err := os.Symlink(root, filepath.Join(root, "self")); err != nil {
		t.Fatal(err)
	}
	got, err := MD(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Must terminate. Either skips "self" or visits it exactly once.
	if len(got) > 5 {
		t.Errorf("loop not detected; got %d entries", len(got))
	}
}

func TestWalkRespectsExclude(t *testing.T) {
	root := t.TempDir()
	mk := func(p string) {
		full := filepath.Join(root, p)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		_ = os.WriteFile(full, []byte("x"), 0o644)
	}
	mk("keep.md")
	mk("node_modules/skip.md")
	mk("node_modules/nested/also-skip.md")
	mk("dist/skip-too.md")
	mk("real/a.md")

	got, err := MD(root, []string{"node_modules", "dist"})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"keep.md", "real/a.md"}
	if !equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
