package sandbox

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMountPreparesResolvedDirectory(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Join(dir, "cache")
	var preparedDir string
	mount := Mount{
		Source:    filepath.Join(link, "cache"),
		Target:    "/mnt/example",
		ReadOnly:  true,
		CreateDir: true,
		Prepare: func(dir string) error {
			preparedDir = dir
			return os.WriteFile(filepath.Join(dir, "ready"), nil, 0600)
		},
	}
	arg, err := mount.arg()
	if err != nil {
		t.Fatalf("mount.arg() error = %v", err)
	}
	if preparedDir != wantDir {
		t.Errorf("Prepare directory = %q, want %q", preparedDir, wantDir)
	}
	if want := wantDir + ":/mnt/example:ro"; arg != want {
		t.Errorf("mount.arg() = %q, want %q", arg, want)
	}
}

func TestMountRejectsInvalidSourceBeforeCreatingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "invalid:source")
	mount := Mount{Source: dir, Target: "/mnt/example", CreateDir: true}

	if _, err := mount.arg(); err == nil {
		t.Fatal("mount.arg() accepted a source path containing a colon")
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("os.Stat(%q) error = %v, want os.ErrNotExist", dir, err)
	}
}

func TestCreateStopsOnMountPreparationError(t *testing.T) {
	want := errors.New("not ready")
	dir := t.TempDir()
	client := New(nil, nil, nil)
	err := client.Create(t.Context(), Creation{Workspace: dir, Mounts: []Mount{{
		Source: dir, Target: "/mnt/example",
		Prepare: func(string) error { return want },
	}}})
	if !errors.Is(err, want) {
		t.Fatalf("Create error = %v, want %v", err, want)
	}
}
