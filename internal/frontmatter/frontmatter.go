// Package frontmatter parses TOML front-matter delimited by +++ … +++ at
// the top of a markdown file. YAML-style --- delimiters are a hard error
// (vibe-doc is TOML-only by design — spec §2.7, §13.1).
package frontmatter

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

// Front holds the parsed front-matter fields.
type Front struct {
	Title        string   `toml:"title"`
	Order        int      `toml:"order"`
	Hidden       bool     `toml:"hidden"`
	Tags         []string `toml:"tags"`
	HideSiblings bool     `toml:"hide_siblings"`
}

var (
	tomlOpen  = []byte("+++")
	tomlClose = []byte("\n+++")
	yamlOpen  = []byte("---")
	errYAML   = errors.New("vibe-doc uses TOML front-matter delimited by `+++ … +++`; YAML-style `---` blocks are not supported")
)

// Parse returns the parsed Front, the remaining body bytes, and any error.
// If no front-matter is present, Front is zero and body equals src.
func Parse(src []byte) (Front, []byte, error) {
	leading := skipBlankPrefix(src)
	rest := src[leading:]

	if bytes.HasPrefix(rest, yamlOpen) && (len(rest) == 3 || rest[3] == '\n' || rest[3] == '\r') {
		return Front{}, nil, errYAML
	}
	if !bytes.HasPrefix(rest, tomlOpen) || (len(rest) > 3 && rest[3] != '\n' && rest[3] != '\r') {
		return Front{}, src, nil
	}
	rest = rest[3:]
	if len(rest) > 0 && rest[0] == '\r' {
		rest = rest[1:]
	}
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}
	closeIdx := bytes.Index(rest, tomlClose)
	if closeIdx < 0 {
		return Front{}, nil, fmt.Errorf("front-matter: unclosed +++ block")
	}
	tomlBytes := rest[:closeIdx]
	bodyStart := closeIdx + 1 + len(tomlOpen)
	if bodyStart < len(rest) && rest[bodyStart] == '\r' {
		bodyStart++
	}
	if bodyStart < len(rest) && rest[bodyStart] == '\n' {
		bodyStart++
	}
	body := rest[bodyStart:]

	var fm Front
	if _, err := toml.Decode(string(tomlBytes), &fm); err != nil {
		return Front{}, nil, fmt.Errorf("front-matter: %w", err)
	}
	return fm, body, nil
}

func skipBlankPrefix(b []byte) int {
	i := 0
	for i < len(b) {
		switch b[i] {
		case ' ', '\t', '\r', '\n':
			i++
		default:
			return i
		}
	}
	return i
}
