package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const version = "0.0.1-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		serveCmd(os.Args[2:])
	case "check":
		fmt.Fprintln(os.Stderr, "check: not yet implemented")
		os.Exit(1)
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
	fmt.Fprintln(os.Stderr, "  vibe-doc serve [--port N] [--mount URL=PATH]...")
	fmt.Fprintln(os.Stderr, "  vibe-doc check")
	fmt.Fprintln(os.Stderr, "  vibe-doc version")
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 4000, "HTTP port")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	http.HandleFunc("/__health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	fmt.Printf("vibe-doc %s — serving on http://%s\n", version, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
