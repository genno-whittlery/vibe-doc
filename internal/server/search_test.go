package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchEndpointReturnsJSON(t *testing.T) {
	srv := newTestServer(t, map[string]string{
		"intro.md":    "# Introduction\n\nA welcome guide for new users.",
		"advanced.md": "# Advanced\n\nDeep technical content.",
	})
	// Force a populate — Server.New calls rebuildSearchIndex once, but the
	// stub from Task 9 was a no-op; with Task 13 wired it should index.
	srv.rebuildSearchIndex()

	req := httptest.NewRequest("GET", "/__search?q=guide", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	var body struct {
		Query     string `json:"query"`
		Truncated bool   `json:"truncated"`
		Results   []struct {
			URL   string `json:"url"`
			Title string `json:"title"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Query != "guide" {
		t.Errorf("query echo = %q", body.Query)
	}
	if len(body.Results) == 0 {
		t.Errorf("expected results for 'guide', got 0")
	}
}

func TestSearchEndpointEmptyQuery(t *testing.T) {
	srv := newTestServer(t, map[string]string{"README.md": "# Home"})
	req := httptest.NewRequest("GET", "/__search?q=", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
}
