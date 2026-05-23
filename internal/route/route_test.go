package route

import (
	"os"
	"path/filepath"
	"testing"
)

// Build a temp doc tree:
//
//	root/
//	  README.md
//	  foo.md
//	  bar/
//	    README.md
//	    baz.md
//	  onlyindex/
//	    index.md
//	  onlyboth/
//	    README.md
//	    index.md  (will be shadowed; routing prefers README)
//	  emptydir/
//	    (no md files)
func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(p, body string) {
		full := filepath.Join(root, p)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("README.md", "# Root")
	mk("foo.md", "# Foo")
	mk("bar/README.md", "# Bar")
	mk("bar/baz.md", "# Baz")
	mk("onlyindex/index.md", "# OnlyIndex")
	mk("onlyboth/README.md", "# Both A")
	mk("onlyboth/index.md", "# Both B")
	_ = os.MkdirAll(filepath.Join(root, "emptydir"), 0o755)
	return root
}

func TestResolve(t *testing.T) {
	root := buildTree(t)
	tests := []struct {
		sub      string
		wantKind Kind
		wantFile string // disk path relative to root (forward slash)
		wantLoc  string // for KindRedirect
	}{
		{"", KindFile, "README.md", ""},
		{"foo", KindFile, "foo.md", ""},
		{"bar", KindRedirect, "", "/bar/"},
		{"bar/", KindFile, "bar/README.md", ""},
		{"bar/baz", KindFile, "bar/baz.md", ""},
		{"onlyindex/", KindFile, "onlyindex/index.md", ""},
		{"onlyboth/", KindFile, "onlyboth/README.md", ""}, // README wins
		{"emptydir/", KindDirListing, "emptydir", ""},
		{"missing.md", KindNotFound, "", ""},
		{"bar/missing", KindNotFound, "", ""},
	}
	for _, tt := range tests {
		got, err := Resolve(root, tt.sub)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", tt.sub, err)
		}
		if got.Kind != tt.wantKind {
			t.Errorf("Resolve(%q).Kind = %v, want %v", tt.sub, got.Kind, tt.wantKind)
			continue
		}
		switch tt.wantKind {
		case KindFile:
			rel, _ := filepath.Rel(root, got.AbsPath)
			if filepath.ToSlash(rel) != tt.wantFile {
				t.Errorf("Resolve(%q).AbsPath rel = %q, want %q", tt.sub, rel, tt.wantFile)
			}
		case KindRedirect:
			if got.Location != tt.wantLoc {
				t.Errorf("Resolve(%q).Location = %q, want %q", tt.sub, got.Location, tt.wantLoc)
			}
		case KindDirListing:
			rel, _ := filepath.Rel(root, got.AbsPath)
			if filepath.ToSlash(rel) != tt.wantFile {
				t.Errorf("Resolve(%q) dir rel = %q, want %q", tt.sub, rel, tt.wantFile)
			}
		}
	}
}
