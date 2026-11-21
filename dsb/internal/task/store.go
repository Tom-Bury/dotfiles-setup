// Package task manages host task folders and their specifications.
package task

import (
	"context"
	_ "embed" // Required by go:embed.
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"dsb/internal/sandbox"
)

// Store locates task folders and the Pi settings used for new drafts.
type Store struct {
	Root           string
	PiSettingsPath string
}

// Settings describes the execution settings saved in settings.json.
type Settings struct {
	Timeout   string   `json:"timeout,omitempty"`
	Ready     bool     `json:"ready"`
	Workspace string   `json:"workspace"`
	Base      string   `json:"base"`
	Branch    string   `json:"branch"`
	Mounts    []string `json:"mounts_ro"`
	Provider  string   `json:"provider"`
	Model     string   `json:"model"`
	Thinking  string   `json:"thinking"`
}

// Entry describes one folder. Err reports an invalid task without hiding its neighbors.
type Entry struct {
	Name     string
	Settings Settings
	State    string
	Err      error
}

// List returns task folders in name order. A missing root is an empty list.
func (s Store) List() ([]Entry, error) {
	dirs, err := os.ReadDir(s.Root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		cfg, err := load(filepath.Join(s.Root, dir.Name()))
		state := taskState(filepath.Join(s.Root, dir.Name()), cfg)
		entries = append(entries, Entry{Name: dir.Name(), Settings: cfg, State: state, Err: err})
	}
	return entries, nil
}

// taskState derives execution state from the claim and its result.
func taskState(dir string, cfg Settings) string {
	data, err := os.ReadFile(filepath.Join(dir, "run", "result"))
	if err == nil {
		result := string(data)
		switch {
		case strings.TrimSpace(result) == "done":
			return "done"
		case strings.HasPrefix(result, "timeout:"):
			return "timeout"
		case strings.HasPrefix(result, "failed:"):
			return "failure"
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "run")); err == nil {
		return "running/interrupted"
	}
	if cfg.Ready {
		return "ready"
	}
	return "draft"
}

func load(dir string) (Settings, error) {
	var cfg Settings
	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if cfg.Workspace == "" || cfg.Base == "" || cfg.Branch == "" ||
		cfg.Model == "" || cfg.Provider == "" || cfg.Thinking == "" {
		return cfg, errors.New("missing task settings")
	}

	info, err := os.Stat(filepath.Join(dir, "TASK.md"))
	if err != nil {
		return cfg, err
	}

	if !info.Mode().IsRegular() {
		return cfg, errors.New("TASK.md is not a regular file")
	}
	return cfg, nil
}

// Prune removes valid, successful tasks after confirm approves the batch.
// The confirmation callback owns user interaction. The task list is checked
// again before each deletion so a changed task is never pruned silently.
func (s Store) Prune(ctx context.Context, confirm func([]Entry) (bool, error)) error {
	entries, err := s.List()
	if err != nil {
		return err
	}

	var done []Entry
	for _, entry := range entries {
		if entry.Err == nil && entry.State == "done" {
			done = append(done, entry)
		}
	}
	ok, err := confirm(done)
	if err != nil || !ok {
		return err
	}

	for _, entry := range done {
		current, err := s.List()
		if err != nil {
			return err
		}
		valid := false
		for _, candidate := range current {
			if candidate.Name == entry.Name && candidate.Err == nil &&
				candidate.State == "done" &&
				candidate.Settings.Workspace == entry.Settings.Workspace &&
				candidate.Settings.Branch == entry.Settings.Branch {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("task %s changed; refusing to prune", entry.Name)
		}

		cmd := exec.CommandContext(ctx, "git", "-C", entry.Settings.Workspace, "branch", "-D", "--", entry.Settings.Branch)
		out, err := cmd.CombinedOutput()
		if err != nil && !missingBranch(out) {
			return fmt.Errorf("delete branch for %s: %s: %w", entry.Name, strings.TrimSpace(string(out)), err)
		}
		dir, err := s.Directory(entry.Name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if err := os.RemoveAll(dir); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func missingBranch(output []byte) bool {
	message := strings.ToLower(string(output))
	return strings.Contains(message, "branch") && strings.Contains(message, "not found")
}

// Directory returns the path to an existing task, even if its settings need repair.
func (s Store) Directory(name string) (string, error) {
	if !filepath.IsLocal(name) || filepath.Base(name) != name || name == "." {
		return "", errors.New("expected a task folder name")
	}

	dir := filepath.Join(s.Root, name)
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("not a task directory: %s", dir)
	}
	return dir, nil
}

//go:embed template.md
var promptTemplate []byte

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
var orderPrefix = regexp.MustCompile(`^[0-9]+-`)

// Create writes a draft for the current Git working tree and returns its prompt path.
// It never replaces an existing task.
func (s Store) Create(ctx context.Context, name string, mounts []string) (string, error) {
	if !validName.MatchString(name) {
		return "", errors.New("use letters, digits, hyphens or underscores for the task name; start with a letter or digit")
	}

	slug := orderPrefix.ReplaceAllString(name, "")
	if slug == "" {
		return "", errors.New("missing task name after order prefix")
	}

	cfg, err := defaults(s.PiSettingsPath)
	if err != nil {
		return "", err
	}

	cfg.Branch = "task/" + slug
	cfg.Workspace, err = revParse(ctx, "--show-toplevel")
	if err != nil {
		return "", err
	}

	cfg.Base, err = revParse(ctx, "HEAD")
	if err != nil {
		return "", err
	}

	cfg.Mounts = []string{}
	for _, mount := range mounts {
		path, err := sandbox.ResolveDirectory(mount)
		if err != nil {
			return "", err
		}
		cfg.Mounts = append(cfg.Mounts, path)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.Root, 0700); err != nil {
		return "", err
	}

	dir := filepath.Join(s.Root, name)
	if err := os.Mkdir(dir, 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), append(data, '\n'), 0600); err != nil {
		return "", err
	}

	prompt := filepath.Join(dir, "TASK.md")
	if err := os.WriteFile(prompt, promptTemplate, 0600); err != nil {
		return "", err
	}

	return prompt, nil
}

func revParse(ctx context.Context, arg string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", arg).Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse %s: %w", arg, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func defaults(path string) (Settings, error) {
	var pi struct {
		DefaultProvider      string
		DefaultModel         string
		DefaultThinkingLevel string
		ModelThinkingLevels  map[string]string
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	if err := json.Unmarshal(data, &pi); err != nil {
		return Settings{}, fmt.Errorf("%s: %w", path, err)
	}

	cfg := Settings{
		Provider: pi.DefaultProvider,
		Model:    pi.DefaultModel,
		Thinking: pi.DefaultThinkingLevel,
	}
	if level := pi.ModelThinkingLevels[cfg.Provider+"/"+cfg.Model]; level != "" {
		cfg.Thinking = level
	}
	if cfg.Provider == "" || cfg.Model == "" || cfg.Thinking == "" {
		return Settings{}, fmt.Errorf("missing Pi defaults in %s", path)
	}
	return cfg, nil
}
