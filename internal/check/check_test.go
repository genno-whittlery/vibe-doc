package check

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genno-whittlery/vibe-doc/internal/config"
)

func TestCheckReportsBrokenLink(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# T\n[good](./good.md)\n[bad](./missing.md)\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "good.md"), []byte("# Good"), 0o644)

	cfg := config.Default()
	cfg.Mounts = []config.Mount{{URL: "/", Root: root}}
	var buf bytes.Buffer
	n, err := Run(cfg, &buf, false)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("expected 1 issue, got %d; output: %s", n, buf.String())
	}
	if !strings.Contains(buf.String(), "missing.md") {
		t.Errorf("issue should mention missing.md, got: %s", buf.String())
	}
}

func TestCheckSkipsExternal(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# T\n[ext](https://example.com)\n[mail](mailto:x@y.com)\n"), 0o644)

	cfg := config.Default()
	cfg.Mounts = []config.Mount{{URL: "/", Root: root}}
	var buf bytes.Buffer
	n, _ := Run(cfg, &buf, false)
	if n != 0 {
		t.Errorf("expected 0 issues for external links, got %d: %s", n, buf.String())
	}
}

func TestCheckAcceptsAbsolutePathInMount(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# T\n[ok](/foo)\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "foo.md"), []byte("# Foo"), 0o644)

	cfg := config.Default()
	cfg.Mounts = []config.Mount{{URL: "/", Root: root}}
	var buf bytes.Buffer
	n, _ := Run(cfg, &buf, false)
	if n != 0 {
		t.Errorf("expected 0 issues for /foo (resolves to foo.md), got %d: %s", n, buf.String())
	}
}
