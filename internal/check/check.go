// Package check implements the `vibe-doc check` subcommand: walks the
// mount tree and reports broken internal links. Exits non-zero on any
// issue. External (http/https/mailto/tel) links are skipped.
package check

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/config"
	"github.com/genno-whittlery/vibe-doc/internal/mount"
	"github.com/genno-whittlery/vibe-doc/internal/route"
	"github.com/genno-whittlery/vibe-doc/internal/walk"
)

// Issue describes one broken-link finding.
type Issue struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Target string `json:"target"`
	Reason string `json:"reason"`
}

var linkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

// Run walks the mounts and writes issues to w. Returns the issue count.
func Run(cfg config.Config, w io.Writer, jsonOut bool) (int, error) {
	ms := make([]mount.Mount, len(cfg.Mounts))
	for i, m := range cfg.Mounts {
		ms[i] = mount.Mount{URL: m.URL, Root: m.Root, Display: m.Display}
	}
	set := mount.New(ms)
	var issues []Issue
	for _, m := range set.Mounts() {
		files, _ := walk.MD(m.Root, cfg.Exclude)
		for _, f := range files {
			absPath := filepath.Join(m.Root, f)
			content, err := os.ReadFile(absPath)
			if err != nil {
				continue
			}
			lines := strings.Split(string(content), "\n")
			for lineIdx, line := range lines {
				for _, match := range linkRe.FindAllStringSubmatch(line, -1) {
					target := match[2]
					if isExternal(target) {
						continue
					}
					if reason := validate(absPath, target, set); reason != "" {
						issues = append(issues, Issue{
							File:   absPath,
							Line:   lineIdx + 1,
							Target: target,
							Reason: reason,
						})
					}
				}
			}
		}
	}
	if jsonOut {
		return len(issues), json.NewEncoder(w).Encode(issues)
	}
	for _, is := range issues {
		fmt.Fprintf(w, "%s:%d → %s: %s\n", is.File, is.Line, is.Target, is.Reason)
	}
	return len(issues), nil
}

func isExternal(target string) bool {
	return strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "https://") ||
		strings.HasPrefix(target, "mailto:") ||
		strings.HasPrefix(target, "tel:")
}

func validate(srcFile, target string, set *mount.Set) string {
	// Strip anchor. Pure anchor (#foo) accepted; whole-doc-anchor
	// verification deferred to v0.2.
	if i := strings.Index(target, "#"); i >= 0 {
		target = target[:i]
		if target == "" {
			return ""
		}
	}
	if strings.HasPrefix(target, "/") {
		m, sub, ok := set.Match(target)
		if !ok {
			return "no mount covers target"
		}
		res, _ := route.Resolve(m.Root, sub)
		if res.Kind == route.KindNotFound {
			return "target not found in mount " + m.URL
		}
		return ""
	}
	// Relative path — resolve against the source file's directory.
	srcDir := filepath.Dir(srcFile)
	abs := filepath.Join(srcDir, target)
	if _, err := os.Stat(abs); err != nil {
		return "relative target missing on disk"
	}
	return ""
}
