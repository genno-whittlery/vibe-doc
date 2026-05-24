package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/genno-whittlery/vibe-doc/internal/check"
	"github.com/genno-whittlery/vibe-doc/internal/config"
	"github.com/genno-whittlery/vibe-doc/internal/logger"
	"github.com/genno-whittlery/vibe-doc/internal/server"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		os.Exit(serveCmd(os.Args[2:]))
	case "check":
		os.Exit(checkCmd(os.Args[2:]))
	case "version", "--version", "-v":
		fmt.Println(version)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "vibe-doc — personal web doc system that just works")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  vibe-doc serve [--port N] [--config PATH] [--mount URL=PATH]...")
	fmt.Fprintln(os.Stderr, "  vibe-doc check [--config PATH]")
	fmt.Fprintln(os.Stderr, "  vibe-doc version")
}

type stringList []string

func (s *stringList) String() string         { return fmt.Sprint(*s) }
func (s *stringList) Set(value string) error { *s = append(*s, value); return nil }

func checkCmd(args []string) int {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	cfgPath := fs.String("config", "vibe-doc.toml", "TOML config path")
	jsonOut := fs.Bool("format-json", false, "emit JSON instead of plain text")
	_ = fs.Parse(args)
	cfg, err := config.LoadFile(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	n, err := check.Run(cfg, os.Stdout, *jsonOut)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if n > 0 {
		return 1
	}
	return 0
}

func serveCmd(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	cfgPath := fs.String("config", "vibe-doc.toml", "TOML config path")
	port := fs.Int("port", 0, "HTTP port (overrides config)")
	var mountFlags stringList
	fs.Var(&mountFlags, "mount", "URL=PATH; may be repeated")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg := config.Default()
	if _, err := os.Stat(*cfgPath); err == nil {
		c, err := config.LoadFile(*cfgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		cfg = c
	}
	if *port != 0 {
		cfg.Port = *port
	}
	for _, mf := range mountFlags {
		m, err := config.ParseMountFlag(mf)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		cfg.Mounts = append(cfg.Mounts, m)
	}
	if len(cfg.Mounts) == 0 {
		fmt.Fprintln(os.Stderr, "no mounts configured. pass --mount or create vibe-doc.toml.")
		return 2
	}

	log, err := logger.New(cfg.Log, cfg.LogMaxBytes, logger.LevelFromString(cfg.LogLevel))
	if err != nil {
		fmt.Fprintln(os.Stderr, "logger:", err)
		return 2
	}
	defer log.Close()

	srv, err := server.New(cfg, log)
	if err != nil {
		fmt.Fprintln(os.Stderr, "server:", err)
		return 2
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "run:", err)
		return 1
	}
	return 0
}
