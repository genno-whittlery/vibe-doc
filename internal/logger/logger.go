// Package logger writes timestamped INFO/WARN/ERROR lines to a single file,
// rotated when it exceeds a configurable cap. All writes go through a
// sync.Mutex so concurrent goroutines don't race on rotation.
//
// Spec ref: §7 (logging limits + race-free rotation).
package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelInfo  Level = 0
	LevelWarn  Level = 1
	LevelError Level = 2
)

func LevelFromString(s string) Level {
	switch s {
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

type Logger struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	level    Level
	f        *os.File
	size     int64
}

func New(path string, maxBytes int64, level Level) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	st, _ := f.Stat()
	return &Logger{
		path:     path,
		maxBytes: maxBytes,
		level:    level,
		f:        f,
		size:     st.Size(),
	}, nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f != nil {
		err := l.f.Close()
		l.f = nil
		return err
	}
	return nil
}

func (l *Logger) Info(format string, args ...any)  { l.write(LevelInfo, "INFO", format, args...) }
func (l *Logger) Warn(format string, args ...any)  { l.write(LevelWarn, "WARN", format, args...) }
func (l *Logger) Error(format string, args ...any) { l.write(LevelError, "ERROR", format, args...) }

func (l *Logger) write(lv Level, label, format string, args ...any) {
	if lv < l.level {
		return
	}
	line := fmt.Sprintf("%s %s %s\n",
		time.Now().UTC().Format(time.RFC3339Nano),
		label,
		fmt.Sprintf(format, args...),
	)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return
	}
	if l.size+int64(len(line)) > l.maxBytes {
		l.rotateLocked()
	}
	n, _ := l.f.WriteString(line)
	l.size += int64(n)
}

func (l *Logger) rotateLocked() {
	if err := l.f.Close(); err != nil {
		_ = err
	}
	_ = os.Rename(l.path, l.path+".old")
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		l.f = nil
		l.size = 0
		return
	}
	l.f = f
	l.size = 0
}
