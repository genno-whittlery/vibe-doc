// Package config loads vibe-doc.toml and merges CLI flags. The TOML file
// is the canonical source; CLI flags override individual fields.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Mount declares a single (URL prefix → filesystem root) mapping. Multiple
// mounts compose the doc site; longest URL-prefix match wins (see §5 of
// the spec).
type Mount struct {
	URL     string `toml:"url"`
	Root    string `toml:"root"`
	Display string `toml:"display,omitempty"`
}

// Config holds the merged configuration after parsing TOML + flags.
type Config struct {
	Port        int     `toml:"port"`
	Log         string  `toml:"log"`
	LogMaxBytes int64   `toml:"log_max_bytes"`
	LogLevel    string  `toml:"log_level"`
	Mounts      []Mount `toml:"mounts"`
	Strict      bool    `toml:"strict"`
}

// Default returns the baseline config. CLI flag parsing layers on top of
// this; TOML file parsing replaces matching fields if a file is present.
func Default() Config {
	return Config{
		Port:        4000,
		Log:         "/tmp/vibe-doc.log",
		LogMaxBytes: 1 << 20, // 1 MiB
		LogLevel:    "info",
	}
}

// LoadFile reads a TOML config file from disk. Returns the merged config
// (defaults + file values). Caller layers CLI flags on top.
func LoadFile(path string) (Config, error) {
	cfg := Default()
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
	}
	// ~ expansion for mount roots and log path (Plan Revision 2).
	for i := range cfg.Mounts {
		cfg.Mounts[i].Root = expandHome(cfg.Mounts[i].Root)
	}
	cfg.Log = expandHome(cfg.Log)
	return cfg, nil
}

// ParseMountFlag parses a single `--mount URL=PATH` CLI value into a Mount.
func ParseMountFlag(s string) (Mount, error) {
	i := strings.Index(s, "=")
	if i <= 0 || i == len(s)-1 {
		return Mount{}, fmt.Errorf("mount: expected URL=PATH, got %q", s)
	}
	return Mount{URL: s[:i], Root: expandHome(s[i+1:])}, nil
}

func expandHome(p string) string {
	if !strings.HasPrefix(p, "~/") && p != "~" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	return home + p[1:]
}
