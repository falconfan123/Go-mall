package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/falconfan123/Go-mall/tools/rag/internal/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "rag-web:", err)
		os.Exit(1)
	}
}

func run() error {
	repoRoot := flag.String("repo", "", "repository root")
	addr := flag.String("addr", "127.0.0.1:5080", "listen address")
	flag.Parse()

	server, err := web.NewServer(*repoRoot)
	if err != nil {
		return err
	}

	fmt.Printf("rag-web listening on http://%s\n", *addr)
	return http.ListenAndServe(*addr, server.Handler())
}
