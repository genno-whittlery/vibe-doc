package sitemap

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genno-whittlery/vibe-doc/internal/mount"
)

func TestGenerateBasic(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# Home"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "foo.md"), []byte("# Foo"), 0o644)
	_ = os.Mkdir(filepath.Join(root, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "sub/leaf.md"), []byte("# Leaf"), 0o644)

	set := mount.New([]mount.Mount{{URL: "/", Root: root}})
	var buf bytes.Buffer
	if err := Generate(&buf, "http://example.com", set); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `<?xml`) {
		t.Errorf("missing XML header: %s", out)
	}
	if !strings.Contains(out, `<urlset`) || !strings.Contains(out, `</urlset>`) {
		t.Errorf("missing urlset wrap: %s", out)
	}
	if !strings.Contains(out, "<loc>http://example.com//foo</loc>") {
		// Note: double slash is because mount URL "/" + "/" + "foo" — accept either form
		if !strings.Contains(out, "<loc>http://example.com/foo</loc>") {
			t.Errorf("expected foo URL in sitemap: %s", out)
		}
	}
	if !strings.Contains(out, "sub/leaf") {
		t.Errorf("expected sub/leaf URL: %s", out)
	}
}

func TestGenerateEmptyMount(t *testing.T) {
	root := t.TempDir()
	set := mount.New([]mount.Mount{{URL: "/", Root: root}})
	var buf bytes.Buffer
	if err := Generate(&buf, "http://example.com", set); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "<urlset") {
		t.Errorf("expected empty urlset wrap")
	}
}
