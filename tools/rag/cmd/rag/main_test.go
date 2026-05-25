package main

import "testing"

func TestParseGlobalArgsAcceptsFlagsAnywhere(t *testing.T) {
	repo, backend, rest, err := parseGlobalArgs([]string{"ask", "--backend", "none", "hello", "--repo=/tmp/repo"})
	if err != nil {
		t.Fatalf("parseGlobalArgs() error = %v", err)
	}
	if repo != "/tmp/repo" {
		t.Fatalf("unexpected repo %q", repo)
	}
	if backend != "none" {
		t.Fatalf("unexpected backend %q", backend)
	}
	if len(rest) != 2 || rest[0] != "ask" || rest[1] != "hello" {
		t.Fatalf("unexpected rest %#v", rest)
	}
}
