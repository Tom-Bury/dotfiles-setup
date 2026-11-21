package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dsb/internal/task"
)

func TestTaskDraft(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	put := func(path, text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0700); err != nil {
			t.Fatal(err)
		}
	}
	put(filepath.Join(bin, "git"), "#!/bin/sh\ncase $2 in --show-toplevel) echo /repo;; HEAD) echo abc123;; *) exit 1;; esac\n")
	put(filepath.Join(bin, "editor"), "#!/bin/sh\nprintf '%s' \"$1\" > \"$RECORD\"\n")
	t.Setenv("PATH", bin)
	t.Setenv("EDITOR", "editor")
	record := filepath.Join(dir, "record")
	t.Setenv("RECORD", record)
	pi := filepath.Join(dir, "shared/.pi/agent")
	if err := os.MkdirAll(pi, 0700); err != nil {
		t.Fatal(err)
	}
	put(filepath.Join(pi, "settings.json"), `{"defaultProvider":"p","defaultModel":"m","defaultThinkingLevel":"minimal","modelThinkingLevels":{"p/m":"medium"}}`)
	app := launcher{settings: settings{dotfilesDir: dir, sharedDir: dir}}
	ctx := t.Context()
	args := []string{"new", "01-demo", bin}
	if err := app.Task(ctx, args); err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(dir, "tasks/01-demo")
	read := func(path string) string {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if read(record) != filepath.Join(folder, "TASK.md") {
		t.Fatal("did not edit prompt")
	}
	if prompt := read(filepath.Join(folder, "TASK.md")); !strings.Contains(prompt, "## Done when") {
		t.Errorf("draft prompt lacks the completion checklist: %q", prompt)
	}
	var s task.Settings
	if err := json.Unmarshal([]byte(read(filepath.Join(folder, "settings.json"))), &s); err != nil {
		t.Fatal(err)
	}
	if s.Ready || s.Base != "abc123" || s.Branch != "task/demo" || s.Workspace != "/repo" || s.Thinking != "medium" || len(s.Mounts) != 1 || s.Mounts[0] != bin {
		t.Fatalf("bad defaults: %+v", s)
	}
	if err := app.Task(ctx, args); err == nil {
		t.Fatal("overwrote draft")
	}
	if err := app.Task(ctx, []string{"open", "01-demo"}); err != nil {
		t.Fatal(err)
	}
	if read(record) != folder {
		t.Fatal("did not open folder")
	}
	if err := app.Task(ctx, []string{"open", ".."}); err == nil {
		t.Fatal("accepted parent path")
	}
}
