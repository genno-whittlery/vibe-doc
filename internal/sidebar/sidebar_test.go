package sidebar

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildBasic(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "README.md"), []byte("# Root"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "foo.md"), []byte("# Foo"), 0o644)
	_ = os.Mkdir(filepath.Join(root, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "sub/README.md"), []byte("# Sub Group"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "sub/leaf.md"), []byte("# Leaf"), 0o644)

	tree, err := Build("Docs", "/", root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Title != "Docs" {
		t.Errorf("root title = %q, want Docs", tree.Title)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("len(children) = %d, want 2 (foo + sub group)", len(tree.Children))
	}

	var fooNode, subNode *Node
	for i := range tree.Children {
		c := &tree.Children[i]
		switch c.URL {
		case "/foo":
			fooNode = c
		case "/sub/":
			subNode = c
		}
	}
	if fooNode == nil || !fooNode.IsLeaf || fooNode.Title != "Foo" {
		t.Errorf("foo node = %+v", fooNode)
	}
	if subNode == nil || subNode.IsLeaf || subNode.Title != "Sub Group" {
		t.Errorf("sub node = %+v", subNode)
	}
	if len(subNode.Children) != 1 || subNode.Children[0].URL != "/sub/leaf" {
		t.Errorf("sub children = %+v", subNode.Children)
	}
}

func TestHiddenAndUnderscorePrefixed(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "_private.md"), []byte("# Private"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "visible.md"), []byte("# Visible"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "hidden-by-fm.md"), []byte("+++\nhidden = true\n+++\n# H"), 0o644)
	tree, err := Build("Docs", "/", root, nil)
	if err != nil {
		t.Fatal(err)
	}
	titles := []string{}
	for _, c := range tree.Children {
		titles = append(titles, c.Title)
	}
	if !contains(titles, "Visible") {
		t.Errorf("visible.md missing from %v", titles)
	}
	if contains(titles, "Private") || contains(titles, "H") {
		t.Errorf("hidden file present in sidebar: %v", titles)
	}
}

func TestBuildRespectsExclude(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "a.md"), []byte("# A"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "node_modules/pkg"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "node_modules/pkg/README.md"), []byte("# nope"), 0o644)

	tree, err := Build("Docs", "/", root, []string{"node_modules"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range tree.Children {
		if c.URL == "/node_modules/" {
			t.Errorf("node_modules surfaced in sidebar despite exclude")
		}
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
