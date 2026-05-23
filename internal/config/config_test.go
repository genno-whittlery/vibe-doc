package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTOMLFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "vibe-doc.toml")
	body := `port = 4001
log = "/tmp/vd.log"
log_max_bytes = 2048

[[mounts]]
url = "/"
root = "/srv/docs"
display = "Docs"

[[mounts]]
url = "/engine"
root = "/srv/engine"
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if cfg.Port != 4001 {
		t.Errorf("Port = %d, want 4001", cfg.Port)
	}
	if cfg.LogMaxBytes != 2048 {
		t.Errorf("LogMaxBytes = %d, want 2048", cfg.LogMaxBytes)
	}
	if len(cfg.Mounts) != 2 {
		t.Fatalf("len(Mounts) = %d, want 2", len(cfg.Mounts))
	}
	if cfg.Mounts[0].URL != "/" || cfg.Mounts[0].Root != "/srv/docs" || cfg.Mounts[0].Display != "Docs" {
		t.Errorf("Mounts[0] = %+v", cfg.Mounts[0])
	}
	if cfg.Mounts[1].URL != "/engine" || cfg.Mounts[1].Display != "" {
		t.Errorf("Mounts[1] = %+v", cfg.Mounts[1])
	}
}

func TestDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Port != 4000 {
		t.Errorf("default Port = %d, want 4000", cfg.Port)
	}
	if cfg.LogMaxBytes != 1<<20 {
		t.Errorf("default LogMaxBytes = %d, want 1MiB", cfg.LogMaxBytes)
	}
	if cfg.Log != "/tmp/vibe-doc.log" {
		t.Errorf("default Log = %q", cfg.Log)
	}
}

func TestParseMountFlag(t *testing.T) {
	m, err := ParseMountFlag("/engine=/Users/x/proj/engine/docs")
	if err != nil {
		t.Fatal(err)
	}
	if m.URL != "/engine" || m.Root != "/Users/x/proj/engine/docs" {
		t.Errorf("got %+v", m)
	}
	if _, err := ParseMountFlag("malformed"); err == nil {
		t.Error("expected error on missing =")
	}
}

func TestExpandHomeInMounts(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir on this OS")
	}
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "vibe-doc.toml")
	body := `port = 4000

[[mounts]]
url = "/"
root = "~/docs"
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(cfg.Mounts[0].Root, home) {
		t.Errorf("Mounts[0].Root = %q, expected expansion starting with %q", cfg.Mounts[0].Root, home)
	}
}

func TestExpandHomeInMountFlag(t *testing.T) {
	home, _ := os.UserHomeDir()
	m, err := ParseMountFlag("/=~/docs")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(m.Root, home) {
		t.Errorf("Root = %q, expected expansion", m.Root)
	}
}

func TestExpandHomeIgnoresNamedUserForm(t *testing.T) {
	// ~/x and ~ expand; ~user/x is intentionally NOT supported.
	for _, input := range []string{"~alice/x", "~bob", "/absolute/no-tilde", "relative/path"} {
		t.Run(input, func(t *testing.T) {
			if got := expandHome(input); got != input {
				t.Errorf("expandHome(%q) = %q, want unchanged", input, got)
			}
		})
	}
}
