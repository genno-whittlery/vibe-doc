// Package shadow detects ambiguity-producing configurations: README+index
// in the same dir, mount overlaps, same-root mounts, etc. Spec ref: §7.
package shadow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/mount"
)

// Kind enumerates shadow types.
type Kind int

const (
	KindReadmeIndex  Kind = 1
	KindMountOverlap Kind = 2
	KindSameRoot     Kind = 3
	KindFileVsDir    Kind = 4
	KindSymlinkOuter Kind = 5
)

func (k Kind) String() string {
	switch k {
	case KindReadmeIndex:
		return "README.md shadows index.md"
	case KindMountOverlap:
		return "mount overlap"
	case KindSameRoot:
		return "two mounts share the same root"
	case KindFileVsDir:
		return "foo.md vs foo/ directory"
	case KindSymlinkOuter:
		return "symlink target outside any mount root"
	default:
		return "unknown"
	}
}

// Conflict describes one shadow.
type Conflict struct {
	Kind    Kind
	Message string
}

// Scan walks the mount set and reports every conflict found. Order is
// deterministic (mount order, then path order).
func Scan(set *mount.Set) []Conflict {
	var out []Conflict
	mounts := set.Mounts()
	// Same-root and overlap checks (O(n²) but n is small).
	for i := range mounts {
		for j := i + 1; j < len(mounts); j++ {
			a, b := mounts[i], mounts[j]
			if absEqual(a.Root, b.Root) {
				out = append(out, Conflict{
					Kind:    KindSameRoot,
					Message: fmt.Sprintf("mounts %s and %s both rooted at %s", a.URL, b.URL, a.Root),
				})
			}
		}
	}
	// Mount overlap: longer-prefix mount shadows a real dir inside a shorter-prefix mount.
	for i := range mounts {
		longer := mounts[i]
		for j := range mounts {
			if i == j {
				continue
			}
			shorter := mounts[j]
			if shorter.URL == "/" || strings.HasPrefix(longer.URL, shorter.URL+"/") {
				rel := strings.TrimPrefix(longer.URL, shorter.URL)
				rel = strings.TrimPrefix(rel, "/")
				if st, err := os.Stat(filepath.Join(shorter.Root, rel)); err == nil && st.IsDir() {
					out = append(out, Conflict{
						Kind:    KindMountOverlap,
						Message: fmt.Sprintf("mount %s shadows %s/%s", longer.URL, shorter.Root, rel),
					})
					break
				}
			}
		}
	}
	// Per-directory README+index check.
	for _, m := range mounts {
		_ = filepath.WalkDir(m.Root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			readme := filepath.Join(p, "README.md")
			index := filepath.Join(p, "index.md")
			_, e1 := os.Stat(readme)
			_, e2 := os.Stat(index)
			if e1 == nil && e2 == nil {
				out = append(out, Conflict{
					Kind:    KindReadmeIndex,
					Message: fmt.Sprintf("%s/index.md shadowed by README.md", p),
				})
			}
			return nil
		})
	}
	return out
}

func absEqual(a, b string) bool {
	aa, _ := filepath.Abs(a)
	bb, _ := filepath.Abs(b)
	return aa == bb
}
