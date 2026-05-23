package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genno-whittlery/vibe-doc/internal/config"
	"github.com/genno-whittlery/vibe-doc/internal/logger"
)

func newTestServer(t *testing.T, docRootContents map[string]string) *Server {
	t.Helper()
	docRoot := t.TempDir()
	for path, body := range docRootContents {
		full := filepath.Join(docRoot, path)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	logPath := filepath.Join(t.TempDir(), "t.log")
	lg, _ := logger.New(logPath, 1<<20, logger.LevelInfo)
	t.Cleanup(func() { lg.Close() })

	cfg := config.Default()
	cfg.Mounts = []config.Mount{{URL: "/", Root: docRoot}}
	srv, err := New(cfg, lg)
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func TestHealth(t *testing.T) {
	srv := newTestServer(t, map[string]string{"README.md": "# Home"})
	req := httptest.NewRequest("GET", "/__health", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if got := rec.Body.String(); got != "ok\n" {
		t.Errorf("health = %q", got)
	}
}

func TestPageRenderRoot(t *testing.T) {
	srv := newTestServer(t, map[string]string{"README.md": "# Welcome\n\nBody."})
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "<h1") || !strings.Contains(body, "Welcome") {
		t.Errorf("page did not render H1 + body:\n%s", body)
	}
	if !strings.Contains(body, "/__static/style.css") {
		t.Errorf("page did not include stylesheet link")
	}
}

func TestDocStaticAsset(t *testing.T) {
	srv := newTestServer(t, map[string]string{
		"README.md":        "# Home",
		"static/image.png": "PNGDATA",
	})
	req := httptest.NewRequest("GET", "/static/image.png", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("static asset status = %d", rec.Code)
	}
	if rec.Body.String() != "PNGDATA" {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestRedirectBareFolderToTrailingSlash(t *testing.T) {
	srv := newTestServer(t, map[string]string{
		"README.md":     "# Home",
		"sub/README.md": "# Sub",
	})
	req := httptest.NewRequest("GET", "/sub", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code != 301 {
		t.Errorf("expected 301 redirect for /sub → /sub/, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/sub/" {
		t.Errorf("Location = %q, want /sub/", loc)
	}
}

func TestPathTraversalRejected(t *testing.T) {
	srv := newTestServer(t, map[string]string{"README.md": "# Home"})
	req := httptest.NewRequest("GET", "/../etc/passwd", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code == 200 && strings.Contains(rec.Body.String(), "root:") {
		t.Errorf("path traversal leaked system file")
	}
}

func TestSitemapEndpoint(t *testing.T) {
	srv := newTestServer(t, map[string]string{
		"README.md": "# Home",
		"foo.md":    "# Foo",
	})
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<urlset") {
		t.Errorf("missing urlset: %s", body)
	}
	if !strings.Contains(body, "foo") {
		t.Errorf("missing foo URL: %s", body)
	}
}
