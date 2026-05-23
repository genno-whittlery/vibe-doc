package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestWriteAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.log")
	lg, err := New(path, 1024, LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	lg.Info("hello %s", "world")
	lg.Warn("be careful")

	b, _ := os.ReadFile(path)
	s := string(b)
	if !strings.Contains(s, "INFO") || !strings.Contains(s, "hello world") {
		t.Errorf("missing INFO line: %q", s)
	}
	if !strings.Contains(s, "WARN") {
		t.Errorf("missing WARN line: %q", s)
	}
}

func TestRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.log")
	lg, err := New(path, 64, LevelInfo) // tiny cap to force rotation
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	for i := 0; i < 20; i++ {
		lg.Info("line %d with some text padding", i)
	}
	if _, err := os.Stat(filepath.Join(dir, "out.log.old")); err != nil {
		t.Errorf("expected rotated .old file: %v", err)
	}
	st, _ := os.Stat(path)
	if st.Size() > 64*2 { // some slack for the last write
		t.Errorf("live log file too big: %d", st.Size())
	}
}

func TestConcurrentWritesNoRace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.log")
	lg, err := New(path, 1<<20, LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				lg.Info("g%d i%d", id, i)
			}
		}(g)
	}
	wg.Wait()
}

func TestLevelGating(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.log")
	lg, err := New(path, 1<<20, LevelWarn)
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Close()
	lg.Info("should be dropped")
	lg.Warn("should appear")
	b, _ := os.ReadFile(path)
	s := string(b)
	if strings.Contains(s, "should be dropped") {
		t.Errorf("INFO line written at level=warn: %q", s)
	}
	if !strings.Contains(s, "should appear") {
		t.Errorf("WARN line missing: %q", s)
	}
}
