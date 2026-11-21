package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"dsb/internal/task"
)

// Task creates, lists, or opens task drafts.
func (app launcher) Task(ctx context.Context, args []string) error {
	store := task.Store{
		Root:           filepath.Join(app.sharedDir, "tasks"),
		PiSettingsPath: filepath.Join(app.dotfilesDir, "shared/.pi/agent/settings.json"),
	}

	switch args[0] {
	case "prune":
		return pruneTasks(ctx, store)
	case "new":
		prompt, err := store.Create(ctx, args[1], args[2:])
		if err != nil {
			return err
		}
		return editTask(ctx, prompt)
	case "open":
		if len(args) == 2 {
			return openTask(ctx, store, args[1])
		}
	}
	entries, err := store.List()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("No tasks.")
		return nil
	}
	for i, e := range entries {
		fmt.Printf("%d. %s\t%s\n", i+1, e.Name, summary(e))
	}
	if args[0] == "list" {
		return nil
	}
	fmt.Print("Task number: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return err
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(entries) {
		return fmt.Errorf("choose a number from 1 to %d", len(entries))
	}
	name := entries[n-1].Name
	if args[0] == "run" {
		return app.RunTasks(ctx, name)
	}
	return openTask(ctx, store, name)
}

func summary(e task.Entry) string {
	if e.Err != nil {
		return "invalid: " + e.Err.Error()
	}
	return e.State + "\t" + e.Settings.Workspace
}

func openTask(ctx context.Context, store task.Store, name string) error {
	dir, err := store.Directory(name)
	if err != nil {
		return err
	}
	return editTask(ctx, dir)
}

func editTask(ctx context.Context, path string) error {
	editor := strings.Fields(os.Getenv("EDITOR"))
	if len(editor) == 0 {
		editor = []string{"nvim"}
	}
	cmd := exec.CommandContext(ctx, editor[0], append(editor[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// pruneTasks handles terminal interaction; task lifecycle and deletion stay in
// the task package.
func pruneTasks(ctx context.Context, store task.Store) error {
	return store.Prune(ctx, func(done []task.Entry) (bool, error) {
		if len(done) == 0 {
			fmt.Println("No completed tasks to prune.")
			return false, nil
		}
		fmt.Println("The following tasks and branches will be removed:")
		for _, e := range done {
			fmt.Printf("  %s: %s (workspace: %s)\n", e.Name, e.Settings.Branch, e.Settings.Workspace)
		}
		fmt.Print("Continue? [y/N] ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return false, err
		}
		if strings.ToLower(strings.TrimSpace(line)) != "y" {
			fmt.Println("Cancelled.")
			return false, nil
		}
		return true, nil
	})
}
