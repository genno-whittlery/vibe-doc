package search

import (
	"strings"
	"testing"
	"time"
)

func TestIndexAndSearch(t *testing.T) {
	idx := New()
	idx.Add("doc1", "Introduction", "Welcome to the introduction. This is the first guide.")
	idx.Add("doc2", "Advanced Topics", "Advanced topics like agent-native authoring.")
	idx.Add("doc3", "Glossary", "Definitions and terms used throughout the guide.")

	res := idx.Search("guide")
	if len(res) == 0 {
		t.Fatal("expected results for 'guide'")
	}
	got := map[string]bool{}
	for _, r := range res {
		got[r.DocID] = true
	}
	if !got["doc1"] || !got["doc3"] {
		t.Errorf("got %v, expected doc1 and doc3", got)
	}

	// Phrase search.
	res = idx.Search(`"agent-native"`)
	if len(res) != 1 || res[0].DocID != "doc2" {
		t.Errorf("phrase search: %+v", res)
	}
}

func TestSnippet(t *testing.T) {
	idx := New()
	idx.Add("d1", "Title", "This is a long body with a unique-keyword in the middle and more text after.")
	res := idx.Search("unique-keyword")
	if len(res) != 1 {
		t.Fatalf("len=%d", len(res))
	}
	if !strings.Contains(res[0].Snippet, "<mark>unique-keyword</mark>") {
		t.Errorf("snippet missing mark: %q", res[0].Snippet)
	}
}

func TestStopwords(t *testing.T) {
	idx := New()
	idx.Add("d1", "T", "the quick brown fox")
	res := idx.Search("the")
	if len(res) != 0 {
		t.Errorf("stopword 'the' should produce no hits, got %d", len(res))
	}
}

// Heading-level boost: a doc that has the term ONLY in an H1 should rank
// above a doc that has the term only in body when both have similar lengths.
func TestSectionBoost(t *testing.T) {
	idx := New()
	idx.AddDoc(IndexedDoc{
		ID: "headerdoc", URL: "/headerdoc", File: "headerdoc.md",
		Title: "Other Title",
		Body:  "irrelevant body text padding padding padding padding padding",
		Headings: []HeadingHit{{Level: 1, Text: "spaceship"}},
	})
	idx.AddDoc(IndexedDoc{
		ID: "bodydoc", URL: "/bodydoc", File: "bodydoc.md",
		Title: "Other Title",
		Body:  "irrelevant body text padding spaceship padding padding padding",
	})
	res := idx.Search("spaceship")
	if len(res) < 2 {
		t.Fatalf("expected 2 hits, got %d", len(res))
	}
	if res[0].DocID != "headerdoc" {
		t.Errorf("headerdoc should outrank bodydoc, got %+v", res)
	}
}

func TestSearchWithTimeout(t *testing.T) {
	idx := New()
	idx.Add("d", "T", "alpha beta gamma")
	res, truncated := idx.SearchWithTimeout("alpha", 100*time.Millisecond)
	if truncated {
		t.Error("trivial query should not time out")
	}
	if len(res) != 1 {
		t.Errorf("expected 1 hit, got %d", len(res))
	}
}
