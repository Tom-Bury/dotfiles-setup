package task_test

import (
	"os"
	"path/filepath"
	"testing"

	"dsb/internal/task"
)

func TestPruneRemovesDoneTasksAfterConfirmation(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	called := filepath.Join(dir, "git-called")
	git := filepath.Join(bin, "git")
	if err := os.WriteFile(git, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \""+called+"\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)

	store := task.Store{Root: filepath.Join(dir, "tasks")}
	taskDir := filepath.Join(store.Root, "01-done")
	if err := os.MkdirAll(filepath.Join(taskDir, "run"), 0700); err != nil {
		t.Fatal(err)
	}
	settings := `{"workspace":"/repo","base":"main","branch":"task/done","provider":"p","model":"m","thinking":"medium"}`
	for name, data := range map[string]string{
		"settings.json": settings,
		"TASK.md":       "done",
		"run/result":    "done",
	} {
		if err := os.WriteFile(filepath.Join(taskDir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}

	var confirmed []string
	if err := store.Prune(t.Context(), func(entries []task.Entry) (bool, error) {
		for _, entry := range entries {
			confirmed = append(confirmed, entry.Name)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(confirmed) != 1 || confirmed[0] != "01-done" {
		t.Fatalf("confirmed tasks = %v", confirmed)
	}
	if _, err := os.Stat(taskDir); !os.IsNotExist(err) {
		t.Fatalf("task directory still exists: %v", err)
	}
	if _, err := os.Stat(called); err != nil {
		t.Fatalf("git was not called: %v", err)
	}
}

func TestPruneIgnoresAlreadyDeletedBranch(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	git := filepath.Join(bin, "git")
	if err := os.WriteFile(git, []byte("#!/bin/sh\nrm -rf \"$TASK_DIR\"\necho \"error: branch 'task/done' not found.\" >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)

	store := task.Store{Root: filepath.Join(dir, "tasks")}
	taskDir := filepath.Join(store.Root, "01-done")
	if err := os.MkdirAll(filepath.Join(taskDir, "run"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TASK_DIR", taskDir)
	for name, data := range map[string]string{
		"settings.json": `{"workspace":"/repo","base":"main","branch":"task/done","provider":"p","model":"m","thinking":"medium"}`,
		"TASK.md":       "done",
		"run/result":    "done",
	} {
		if err := os.WriteFile(filepath.Join(taskDir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Prune(t.Context(), func([]task.Entry) (bool, error) { return true, nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(taskDir); !os.IsNotExist(err) {
		t.Fatalf("task directory still exists: %v", err)
	}
}

func TestListKeepsInvalidTasksInOrder(t *testing.T) {
	store := task.Store{Root: filepath.Join(t.TempDir(), "tasks")}
	entries, err := store.List()
	if err != nil || len(entries) != 0 {
		t.Fatalf("List before creation = %v, %v; want empty list", entries, err)
	}
	for _, name := range []string{"02-broken", "01-ready"} {
		if err := os.MkdirAll(filepath.Join(store.Root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	dir, err := store.Directory("01-ready")
	if err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{"TASK.md": "Check defaults.", "settings.json": `{"ready":true,"workspace":"/repo","base":"main","branch":"task/demo","provider":"p","model":"m","thinking":"medium"}`} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	entries, err = store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("List returned %d entries; want 2", len(entries))
	}
	if e := entries[0]; e.Name != "01-ready" || e.Err != nil || !e.Settings.Ready {
		t.Errorf("first entry = %+v; want valid ready task", e)
	}
	if e := entries[1]; e.Name != "02-broken" || e.Err == nil {
		t.Errorf("second entry = %+v; want invalid task", e)
	}
}
