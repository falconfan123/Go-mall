package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/tools/rag/internal/app"
	"github.com/falconfan123/Go-mall/tools/rag/internal/model"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "rag:", err)
		os.Exit(1)
	}
}

func run() error {
	repoRoot, backend, args, err := parseGlobalArgs(os.Args[1:])
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return usage()
	}

	cli, err := app.NewCLI(repoRoot)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	switch args[0] {
	case "index":
		result, err := cli.Index(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("indexed %d files into %d chunks\n", result.FilesIndexed, len(result.Chunks))
		return nil
	case "ask":
		fs := newFlagSet("ask")
		topK := fs.Int("top-k", 6, "number of context chunks")
		refresh := fs.Bool("refresh", false, "rebuild the index before asking")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		query := strings.TrimSpace(strings.Join(fs.Args(), " "))
		if query == "" {
			return fmt.Errorf("ask requires a question")
		}
		m, err := loadModel(backend)
		if err != nil {
			return err
		}
		answer, _, err := cli.Ask(ctx, app.AskOptions{
			Query:   query,
			TopK:    *topK,
			Model:   m,
			Refresh: *refresh,
		})
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	case "loop":
		fs := newFlagSet("loop")
		topK := fs.Int("top-k", 6, "number of context chunks")
		dryRun := fs.Bool("dry-run", false, "print the first model plan without executing changes")
		maxIterations := fs.Int("max-iterations", 6, "maximum repair iterations")
		autoCommit := fs.Bool("commit", true, "commit when final validation succeeds")
		var extras stringSlice
		fs.Var(&extras, "allow-command", "additional allowed command")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		prompt := strings.TrimSpace(strings.Join(fs.Args(), " "))
		if prompt == "" {
			return fmt.Errorf("loop requires a task prompt")
		}
		m, err := loadModel(backend)
		if err != nil {
			return err
		}
		result, err := cli.Loop(ctx, app.LoopOptions{
			Prompt:          prompt,
			Model:           m,
			AllowedCommands: extras,
			MaxIterations:   *maxIterations,
			DryRun:          *dryRun,
			TopK:            *topK,
			AutoCommit:      *autoCommit,
		})
		if err != nil {
			return err
		}
		fmt.Printf("session=%s status=%s summary=%s\n", result.Session.ID, result.Session.Status, result.Session.Summary)
		if result.Plan.Summary != "" {
			fmt.Println(result.Plan.Summary)
		}
		for _, out := range result.Output {
			if out.Action.Type == "run" {
				fmt.Printf("run %s\n%s\n", out.Action.Command, out.Output)
				continue
			}
			fmt.Printf("%s %s\n%s\n", out.Action.Type, out.Action.Path, out.Output)
		}
		return nil
	case "resume":
		if len(args) < 2 {
			return fmt.Errorf("resume requires a session id")
		}
		m, err := loadModel(backend)
		if err != nil {
			return err
		}
		result, err := cli.Resume(ctx, args[1], m)
		if err != nil {
			return err
		}
		fmt.Printf("session=%s status=%s summary=%s\n", result.Session.ID, result.Session.Status, result.Session.Summary)
		return nil
	case "doctor":
		report, err := cli.Doctor(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("repo: %s\n", report.RepoRoot)
		fmt.Printf("db: %s\n", report.DBPath)
		fmt.Printf("writable: %t\n", report.Writable)
		fmt.Printf("git: %t\n", report.GitReady)
		fmt.Printf("anthropic_auth_token: %t\n", report.AuthTokenReady)
		fmt.Printf("anthropic_api_key: %t\n", report.APIKeyReady)
		fmt.Printf("auth_ready: %t\n", report.AuthReady)
		if len(report.IndexedSources) > 0 {
			fmt.Println("indexed_sources:")
			for name, count := range report.IndexedSources {
				fmt.Printf("  %s: %d\n", name, count)
			}
		}
		for target, err := range report.AllowedChecks {
			if err != nil {
				fmt.Printf("%s: FAIL (%v)\n", target, err)
			} else {
				fmt.Printf("%s: OK\n", target)
			}
		}
		return nil
	default:
		return usage()
	}
}

func loadModel(name string) (model.Client, error) {
	switch strings.TrimSpace(name) {
	case "", "anthropic":
		return app.NewAnthropicClient()
	case "none":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported backend: %s", name)
	}
}

func usage() error {
	fmt.Fprintln(os.Stderr, "usage: rag [--repo PATH] [--backend anthropic|none] <index|ask|loop|resume|doctor> ...")
	return fmt.Errorf("invalid command")
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	return fs
}

func parseGlobalArgs(args []string) (repoRoot string, backend string, rest []string, err error) {
	backend = "anthropic"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--repo":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("--repo requires a value")
			}
			repoRoot = args[i+1]
			i++
		case strings.HasPrefix(arg, "--repo="):
			repoRoot = strings.TrimPrefix(arg, "--repo=")
		case arg == "--backend":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("--backend requires a value")
			}
			backend = args[i+1]
			i++
		case strings.HasPrefix(arg, "--backend="):
			backend = strings.TrimPrefix(arg, "--backend=")
		default:
			rest = append(rest, arg)
		}
	}
	return repoRoot, backend, rest, nil
}

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, strings.TrimSpace(value))
	return nil
}
