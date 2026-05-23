package frontmatter

import (
	"testing"
)

func TestParseTOML(t *testing.T) {
	src := `+++
title = "Welcome"
order = 1
tags = ["intro", "guide"]
hidden = false
+++
# Body header
Body text.
`
	fm, body, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if fm.Title != "Welcome" {
		t.Errorf("Title = %q", fm.Title)
	}
	if fm.Order != 1 {
		t.Errorf("Order = %d", fm.Order)
	}
	if len(fm.Tags) != 2 || fm.Tags[0] != "intro" {
		t.Errorf("Tags = %v", fm.Tags)
	}
	if !startsWith(body, "# Body header") {
		t.Errorf("body trimmed wrong: %q", body[:30])
	}
}

func TestNoFrontMatter(t *testing.T) {
	src := []byte("# Just a header\nBody.\n")
	fm, body, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if fm.Title != "" {
		t.Errorf("expected zero fm, got %+v", fm)
	}
	if string(body) != string(src) {
		t.Errorf("body should equal input when no front-matter")
	}
}

func TestYAMLDelimiterIsHardError(t *testing.T) {
	src := []byte("---\ntitle: x\n---\nBody.\n")
	_, _, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for YAML --- front-matter")
	}
	if !contains(err.Error(), "TOML") {
		t.Errorf("error should mention TOML, got %q", err.Error())
	}
}

func TestMalformedTOMLReturnsError(t *testing.T) {
	src := []byte("+++\nthis is not = valid toml [[\n+++\nbody")
	_, _, err := Parse(src)
	if err == nil {
		t.Error("expected parse error")
	}
}

// Trailing whitespace on the closing +++ line is tolerated (the search
// needle `\n+++` still matches). The body keeps the trailing space as its
// first char, which renderers harmlessly ignore. This guards against a
// regression that would either drop the close or error.
func TestParseTOMLTrailingWhitespaceOnFence(t *testing.T) {
	src := []byte("+++\ntitle = \"X\"\n+++ \n# Body\n")
	fm, body, err := Parse(src)
	if err != nil {
		t.Fatalf("expected to parse despite trailing space, got err: %v", err)
	}
	if fm.Title != "X" {
		t.Errorf("Title = %q, want X", fm.Title)
	}
	if !contains(string(body), "# Body") {
		t.Errorf("body should contain # Body header, got %q", string(body))
	}
}

func startsWith(b []byte, s string) bool { return len(b) >= len(s) && string(b[:len(s)]) == s }
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
