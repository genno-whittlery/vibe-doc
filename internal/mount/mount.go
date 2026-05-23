// Package mount implements multi-mount routing: a request URL is matched
// against the longest URL-prefix mount first. Spec ref: §5.
package mount

import (
	"sort"
	"strings"
)

// Mount declares a single (URL prefix → filesystem root) mapping.
type Mount struct {
	URL     string
	Root    string
	Display string
}

// Set is an immutable collection of Mounts sorted by URL length descending,
// so longest-prefix wins on the first matching iteration.
type Set struct {
	mounts []Mount
}

// New builds an immutable Set. URL prefixes are normalized: stripped of
// trailing slashes (except the root mount "/").
func New(mounts []Mount) *Set {
	cp := make([]Mount, len(mounts))
	copy(cp, mounts)
	for i := range cp {
		cp[i].URL = normalizeURL(cp[i].URL)
	}
	sort.SliceStable(cp, func(i, j int) bool {
		return len(cp[i].URL) > len(cp[j].URL)
	})
	return &Set{mounts: cp}
}

func normalizeURL(u string) string {
	if u == "" {
		return "/"
	}
	if u == "/" {
		return u
	}
	return strings.TrimRight(u, "/")
}

// Mounts returns the sorted slice (read-only — caller MUST NOT mutate).
func (s *Set) Mounts() []Mount { return s.mounts }

// Match finds the longest-prefix mount for urlPath. Returns the Mount, the
// remaining sub-path (relative to the mount's root, with no leading slash),
// and ok=true. ok=false means no mount covers urlPath.
func (s *Set) Match(urlPath string) (Mount, string, bool) {
	if urlPath == "" {
		urlPath = "/"
	}
	for _, m := range s.mounts {
		if m.URL == "/" {
			sub := strings.TrimPrefix(urlPath, "/")
			return m, sub, true
		}
		if urlPath == m.URL {
			return m, "", true
		}
		if strings.HasPrefix(urlPath, m.URL+"/") {
			return m, strings.TrimPrefix(urlPath, m.URL+"/"), true
		}
	}
	return Mount{}, "", false
}
