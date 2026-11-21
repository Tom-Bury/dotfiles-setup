package task

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestQueueContinuesAfterTimeoutWithoutRetry(t *testing.T) {
	store := Store{Root: t.TempDir()}
	for _, name := range []string{"03-pass", "02-result-error", "01-timeout"} {
		dir := filepath.Join(store.Root, name)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		for file, data := range map[string]string{"TASK.md": "do work", "settings.json": `{"ready":true,"workspace":"/repo","branch":"task/work","base":"main","provider":"p","model":"m","thinking":"medium","timeout":"10ms"}`} {
			if err := os.WriteFile(filepath.Join(dir, file), []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	var seen []string
	run := func(ctx context.Context, dir string, _ Settings) error {
		name := filepath.Base(dir)
		seen = append(seen, name)
		if name == "02-result-error" {
			// Force the result write to fail without relying on file permissions.
			return os.Mkdir(filepath.Join(dir, "run", "result"), 0700)
		}
		if name == "01-timeout" {
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	}
	err := store.RunReady(t.Context(), "", run)
	var writeErr *os.PathError
	if !errors.Is(err, context.DeadlineExceeded) || !errors.As(err, &writeErr) {
		t.Fatalf("RunReady() error = %v, want both deadline exceeded and a result write error", err)
	}
	if err := store.RunReady(t.Context(), "", run); err != nil {
		t.Fatal(err)
	}
	if want := []string{"01-timeout", "02-result-error", "03-pass"}; !slices.Equal(seen, want) {
		t.Fatalf("task runs = %v, want %v", seen, want)
	}
}
