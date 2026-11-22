package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinishTaskAlwaysForcesRemoval(t *testing.T) {
	for _, failRemoval := range []bool{false, true} {
		t.Run(map[bool]string{false: "removed", true: "removal fails"}[failRemoval], func(t *testing.T) {
			dir := t.TempDir()
			argsFile := filepath.Join(dir, "args")
			script := "#!/bin/sh\nprintf '%s\n' \"$@\" > \"$ARGS_FILE\"\n"
			if failRemoval {
				script += "exit 1\n"
			}
			if err := os.WriteFile(filepath.Join(dir, "sbx"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("ARGS_FILE", argsFile)
			original := errors.New("task failed")
			var log bytes.Buffer
			err := (launcher{runtimeDir: dir}).finishTask("example", &log, original)
			if !errors.Is(err, original) {
				t.Fatalf("lost task error: %v", err)
			}
			if failRemoval && !strings.Contains(err.Error(), "remove task sandbox example") {
				t.Fatalf("missing removal error: %v", err)
			}
			args, readErr := os.ReadFile(argsFile)
			if readErr != nil || string(args) != "--no-input\nrm\n--force\nexample\n" {
				t.Fatalf("removal args = %q, error = %v", args, readErr)
			}
		})
	}
}
