package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPiAuthMountCreatesPrivateFile(t *testing.T) {
	authDir := t.TempDir()
	if err := securePiAuthDir(authDir); err != nil {
		t.Fatal(err)
	}
	authPath := filepath.Join(authDir, "auth.json")
	if got := readTestFile(t, authPath); got != "{}\n" {
		t.Fatalf("initial auth contents = %q, want empty JSON object", got)
	}
	assertPermissions(t, authDir, 0700)
	assertPermissions(t, authPath, 0600)
}

func TestPiAuthMountPreservesExistingFile(t *testing.T) {
	authDir := t.TempDir()
	authPath := filepath.Join(authDir, "auth.json")
	writeTestFile(t, authPath, "existing")

	if err := securePiAuthDir(authDir); err != nil {
		t.Fatal(err)
	}
	if got := readTestFile(t, authPath); got != "existing" {
		t.Fatalf("auth contents changed to %q", got)
	}
	assertPermissions(t, authPath, 0600)
}

func TestPiAuthMountRejectsSymlink(t *testing.T) {
	authDir := t.TempDir()
	targetPath := filepath.Join(t.TempDir(), "target")
	writeTestFile(t, targetPath, "unchanged")
	if err := os.Symlink(targetPath, filepath.Join(authDir, "auth.json")); err != nil {
		t.Fatal(err)
	}

	if err := securePiAuthDir(authDir); err == nil {
		t.Fatal("accepted an auth file symlink")
	}
	assertPermissions(t, targetPath, 0700)
}

func assertPermissions(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != want {
		t.Fatalf("permissions for %s = %04o, want %04o", path, got, want)
	}
}
