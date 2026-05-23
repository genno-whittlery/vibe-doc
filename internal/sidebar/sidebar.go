// Package sidebar builds the auto-generated navigation tree from the
// filesystem. Spec ref: §6. This file is a STUB — Task 10 replaces it with
// the full implementation. The stub exists so the server package can
// compile against the sidebar API while Task 10 is in flight.
package sidebar

// Node is one entry in the sidebar tree. A Node is either a leaf .md file
// (IsLeaf=true), a group / directory (IsLeaf=false, may have Children), or
// a root / mount (IsLeaf=false, Title from mount Display).
type Node struct {
	Title    string
	URL      string
	IsLeaf   bool
	Order    int
	Hidden   bool
	Children []Node
}

// Build constructs a Node tree for one mount. STUB: returns a single
// root-only node. Task 10 walks the filesystem and populates Children.
func Build(display, urlPrefix, root string) (Node, error) {
	return Node{Title: display, URL: urlPrefix}, nil
}
