package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/falconfan123/Go-mall/test/rpc/internal/harness"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
)

func main() {
	timeout := flag.Duration("timeout", testenv.Timeout(), "maximum time to wait for the local RPC stack")
	mode := flag.String("mode", "wait", "readycheck mode: wait or monitor")
	interval := flag.Duration("interval", 2*time.Second, "monitor interval")
	flag.Parse()

	if err := testenv.RequireLocalMode(); err != nil {
		fmt.Fprintf(os.Stderr, "local rpc test misconfigured: %v\n", err)
		os.Exit(2)
	}

	switch *mode {
	case "wait":
		runWait(*timeout)
	case "monitor":
		runMonitor(*timeout, *interval)
	default:
		fmt.Fprintf(os.Stderr, "unsupported readycheck mode: %s\n", *mode)
		os.Exit(2)
	}
}

func runWait(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := harness.WaitForStackReady(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "rpc stack not ready: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("rpc stack ready within %s\n", timeout.Round(time.Second))
}

func runMonitor(timeout time.Duration, interval time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := harness.MonitorStackHealth(ctx, interval); err != nil {
		fmt.Fprintf(os.Stderr, "rpc stack became unhealthy: %v\n", err)
		os.Exit(1)
	}
}
