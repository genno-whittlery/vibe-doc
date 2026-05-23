package mount

import (
	"reflect"
	"sort"
	"testing"
)

func TestLongestPrefixMatch(t *testing.T) {
	set := New([]Mount{
		{URL: "/", Root: "/a"},
		{URL: "/engine", Root: "/b"},
		{URL: "/engine/api", Root: "/c"},
	})
	tests := []struct {
		urlPath  string
		wantRoot string
		wantSub  string
	}{
		{"/", "/a", ""},
		{"/foo", "/a", "foo"},
		{"/engineroom", "/a", "engineroom"},
		{"/engine", "/b", ""},
		{"/engine/", "/b", ""},
		{"/engine/foo", "/b", "foo"},
		{"/engine/api", "/c", ""},
		{"/engine/api/v2", "/c", "v2"},
	}
	for _, tt := range tests {
		m, sub, ok := set.Match(tt.urlPath)
		if !ok {
			t.Errorf("Match(%q) returned !ok", tt.urlPath)
			continue
		}
		if m.Root != tt.wantRoot {
			t.Errorf("Match(%q).Root = %q, want %q", tt.urlPath, m.Root, tt.wantRoot)
		}
		if sub != tt.wantSub {
			t.Errorf("Match(%q) sub = %q, want %q", tt.urlPath, sub, tt.wantSub)
		}
	}
}

func TestNoMatch(t *testing.T) {
	set := New([]Mount{{URL: "/engine", Root: "/b"}})
	if _, _, ok := set.Match("/sdk/foo"); ok {
		t.Error("expected !ok for unmatched path")
	}
}

func TestSortedByLengthDesc(t *testing.T) {
	set := New([]Mount{
		{URL: "/a", Root: "/x"},
		{URL: "/a/b/c", Root: "/y"},
		{URL: "/a/b", Root: "/z"},
	})
	got := []string{}
	for _, m := range set.Mounts() {
		got = append(got, m.URL)
	}
	want := []string{"/a/b/c", "/a/b", "/a"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
	_ = sort.IntsAreSorted // shut linter up about unused import in test
}
