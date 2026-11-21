package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"dsb/internal/sandbox"
)

func launcherWithFakeSbx(t *testing.T) launcher {
	t.Helper()
	hostDir := canonicalTestDir(t)
	t.Setenv("PATH", hostDir)
	t.Setenv("EXISTING_SANDBOX", "")
	t.Setenv(allowedDomainsEnv, "")
	t.Setenv("SBX_CALL_LOG", filepath.Join(hostDir, "calls"))
	writeTestFile(t, filepath.Join(hostDir, "sbx"), `#!/bin/sh
command=$1
shift
printf '%s\n' "$command" >> "$SBX_CALL_LOG"
printf '%s\n' "$@" > "$SBX_CALL_LOG.$command"
if [ "$command" = ls ]; then
    printf 'NAME  STATUS\n'
    if [ -n "$EXISTING_SANDBOX" ]; then
        printf '%s  stopped\n' "$EXISTING_SANDBOX"
    fi
fi
`)
	return launcher{client: sandbox.New(nil, nil, nil), settings: settings{
		baseSandboxName:     "test-base",
		projectTemplateName: "dev:v1",
		dotfilesDir:         hostDir,
		sharedDir:           filepath.Join(hostDir, "shared"),
		piAuthDir:           filepath.Join(hostDir, "private auth"),
		allowedDomains:      "auth.openai.com,proxy.golang.org,sum.golang.org,storage.googleapis.com,registry.npmjs.org",
		runtimeDir:          filepath.Join(hostDir, "runtime"),
	}}
}

func assertInteractiveCalls(t *testing.T, sandboxName, callOrder string) {
	t.Helper()
	assertCallOrder(t, callOrder)
	if callOrder == "ls\ncreate\npolicy\nrun\n" && sandboxName == "test-base" {
		assertSbxCall(t, "policy", "allow", "network", "--sandbox", sandboxName, "*")
	} else if callOrder != "ls\nrun\n" {
		assertSbxCall(t, "policy", "allow", "network", "--sandbox", sandboxName, "auth.openai.com,proxy.golang.org,sum.golang.org,storage.googleapis.com,registry.npmjs.org")
	}
	assertSbxCall(t, "run", "--name", sandboxName)
}

func assertCallOrder(t *testing.T, want string) {
	t.Helper()
	if got := readTestFile(t, os.Getenv("SBX_CALL_LOG")); got != want {
		t.Fatalf("sbx call order = %q, want %q", got, want)
	}
}

func assertSbxCall(t *testing.T, command string, want ...string) {
	t.Helper()
	callLog := readTestFile(t, os.Getenv("SBX_CALL_LOG")+"."+command)
	got := strings.Split(strings.TrimSuffix(callLog, "\n"), "\n")
	if !slices.Equal(got, want) {
		t.Fatalf("sbx %s args = %q, want %q", command, got, want)
	}
}

func startupDir(runtimeDir, name string) string {
	return filepath.Join(runtimeDir, fmt.Sprintf("%x", sha256.Sum256([]byte(name))))
}

func canonicalTestDir(t *testing.T) string {
	t.Helper()
	testDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return testDir
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0700); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}
