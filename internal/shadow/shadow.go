// Package shadow detects routing conflicts: README + index in the same
// directory, mount-prefix overlaps, etc. Spec ref: §7. This file is a
// STUB — Task 11 replaces it with the full implementation.
package shadow

import "github.com/genno-whittlery/vibe-doc/internal/mount"

// ConflictKind enumerates shadow types. Task 11 will add constants as it
// detects more cases.
type ConflictKind int

const (
	KindUnknown ConflictKind = iota
	KindReadmeIndex
	KindMountOverlap
)

func (k ConflictKind) String() string {
	switch k {
	case KindReadmeIndex:
		return "readme-index"
	case KindMountOverlap:
		return "mount-overlap"
	default:
		return "unknown"
	}
}

// Conflict describes one detected shadow. Task 11 will populate URL/Path
// fields as it walks each mount.
type Conflict struct {
	Kind    ConflictKind
	Message string
}

// Scan returns all detected conflicts across mounts. STUB: returns nil.
// Task 11 walks each mount root + cross-mount prefix overlaps.
func Scan(_ *mount.Set) []Conflict {
	return nil
}
