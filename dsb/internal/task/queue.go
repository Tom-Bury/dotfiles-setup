package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// RunFunc runs one claimed task in its task directory.
type RunFunc func(context.Context, string, Settings) error

// RunReady claims each ready task once. A nonempty name limits the run to that task.
// The run directory remains after exit so an interrupted task cannot restart by accident.
func (s Store) RunReady(ctx context.Context, name string, run RunFunc) error {
	entries, err := s.List()
	if err != nil {
		return err
	}

	var failures []error
	found := false
	for _, entry := range entries {
		if name != "" && name != entry.Name {
			continue
		}
		found = true
		if err := ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		if entry.Err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", entry.Name, entry.Err))
			continue
		}
		if !entry.Settings.Ready {
			if name != "" {
				return fmt.Errorf("%s is not ready", name)
			}
			continue
		}
		dir := filepath.Join(s.Root, entry.Name)
		err := os.Mkdir(filepath.Join(dir, "run"), 0700)
		if errors.Is(err, os.ErrExist) && name == "" {
			continue
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("claim %s: %w", entry.Name, err))
			continue
		}

		if err := runClaimed(ctx, dir, entry, run); err != nil {
			failures = append(failures, err)
		}
	}
	if name != "" && !found {
		return fmt.Errorf("unknown task %q", name)
	}
	return errors.Join(failures...)
}

// RunMetadata records the outcome of one claimed task.
type RunMetadata struct {
	Version    int       `json:"version"`
	Task       string    `json:"task"`
	Settings   Settings  `json:"settings"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	DurationMS int64     `json:"duration_ms"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
}

func runClaimed(ctx context.Context, dir string, entry Entry, run RunFunc) error {
	log.Printf("running %s", entry.Name)
	started := time.Now()
	err := runWithDeadline(ctx, dir, entry.Settings, run)
	finished := time.Now()
	metadata := RunMetadata{Version: 1, Task: entry.Name, Settings: entry.Settings, StartedAt: started.UTC(), FinishedAt: finished.UTC(), DurationMS: finished.Sub(started).Milliseconds(), Status: "done"}
	if err != nil {
		metadata.Status = "failed"
		metadata.Error = err.Error()
		if errors.Is(err, context.DeadlineExceeded) {
			metadata.Status = "timeout"
		} else if errors.Is(err, context.Canceled) {
			metadata.Status = "canceled"
		}
	}
	data, marshalErr := json.MarshalIndent(metadata, "", "  ")
	if marshalErr == nil {
		marshalErr = os.WriteFile(filepath.Join(dir, "run", "metadata.json"), append(data, '\n'), 0600)
	}
	var metadataErr error
	if marshalErr != nil {
		metadataErr = fmt.Errorf("record metadata for %s: %w", entry.Name, marshalErr)
	}
	result := "done\n"
	if err != nil {
		result = "failed: " + err.Error() + "\n"
		if errors.Is(err, context.DeadlineExceeded) {
			result = "timeout: " + err.Error() + "\n"
		}
		err = fmt.Errorf("%s: %w", entry.Name, err)
	}

	log.Printf("%s: %s", entry.Name, result)
	path := filepath.Join(dir, "run/result")
	if writeErr := os.WriteFile(path, []byte(result), 0600); writeErr != nil {
		return errors.Join(err, metadataErr, fmt.Errorf("record result for %s: %w", entry.Name, writeErr))
	}
	return errors.Join(err, metadataErr)
}

func runWithDeadline(ctx context.Context, dir string, cfg Settings, run RunFunc) error {
	timeout := 30 * time.Minute
	if cfg.Timeout != "" {
		d, err := time.ParseDuration(cfg.Timeout)
		if err != nil || d <= 0 {
			return fmt.Errorf("invalid timeout %q", cfg.Timeout)
		}
		timeout = d
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := run(ctx, dir, cfg)
	return errors.Join(err, ctx.Err())
}
