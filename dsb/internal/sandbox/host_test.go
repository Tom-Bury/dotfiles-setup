package sandbox

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunNetworkRules(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if err := os.WriteFile(filepath.Join(dir, "sbx"), []byte(`#!/bin/sh
printf '%s\n' "$*"
if [ "$1" = policy ]; then exit 1; fi
`), 0700); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	client := New(nil, &out, nil)
	if err := client.Run(t.Context(), RunSpec{Name: "example"}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "run --name example\n" {
		t.Fatalf("without rules: %q", got)
	}
	out.Reset()
	err := client.Run(t.Context(), RunSpec{Name: "example", NetworkAllow: []string{"one.example", "two.example"}})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Errorf("Run error = %v, want sbx exit status 1", err)
	}
	if got := out.String(); got != "policy allow network --sandbox example one.example,two.example\n" {
		t.Fatalf("after failed policy: %q", got)
	}
}
