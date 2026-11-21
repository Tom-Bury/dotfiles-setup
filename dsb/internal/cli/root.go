// Package cli routes host commands and reports their results.
package cli

import (
	"context"
	"fmt"
	"io"
)

const usage = `Usage:
  dsb                   Open the sandbox for this directory.
  dsb open [MOUNT ...]  Open it with extra mounts on first creation.
  dsb task new NAME [MOUNT ...]  Create a draft and edit its prompt.
  dsb auto                       Run ready tasks in folder order.
  dsb task run [NAME]            Run one ready task. Omit NAME to choose.
  dsb task list                  List tasks in folder order.
  dsb task prune                 Remove completed tasks and their branches.
  dsb task open [NAME]           Open a task folder. Omit NAME to choose.
  dsb base              Open the base sandbox.
  dsb base save         Save the base as the project template.
  dsb nuke              Remove all local sandboxes, including the base.`

// Launcher supplies the actions behind the CLI commands.
type Launcher interface {
	RunTasks(context.Context, string) error
	Task(context.Context, []string) error
	OpenProject(context.Context, []string) error
	OpenBase(context.Context) error
	SaveProjectTemplate(context.Context) error
	Nuke(context.Context) error
}

// Execute routes one command. With no arguments, it opens the current project.
func Execute(ctx context.Context, launcher Launcher, args []string, out io.Writer) error {
	if len(args) == 0 {
		return launcher.OpenProject(ctx, nil)
	}

	switch args[0] {
	case "auto":
		if len(args) == 1 {
			return launcher.RunTasks(ctx, "")
		}
		return invalidCommand(args)
	case "task":
		if len(args) == 3 && args[1] == "run" && args[2] != "" {
			return launcher.RunTasks(ctx, args[2])
		}
		if len(args) >= 3 && args[1] == "new" || len(args) == 2 && (args[1] == "list" || args[1] == "run" || args[1] == "prune") || (len(args) == 2 || len(args) == 3) && args[1] == "open" {
			return launcher.Task(ctx, args[1:])
		}
		return invalidCommand(args)
	case "open":
		return launcher.OpenProject(ctx, args[1:])

	case "base":
		switch {
		case len(args) == 1:
			return launcher.OpenBase(ctx)
		case len(args) == 2 && args[1] == "save":
			return launcher.SaveProjectTemplate(ctx)
		}
		return invalidCommand(args)

	case "nuke":
		if len(args) != 1 {
			return invalidCommand(args)
		}
		return launcher.Nuke(ctx)

	case "--help", "-h", "help":
		if len(args) != 1 {
			return invalidCommand(args)
		}
		_, err := fmt.Fprintln(out, usage)
		return err

	default:
		return invalidCommand(args)
	}
}

func invalidCommand(args []string) error {
	return fmt.Errorf("invalid command %q\n%s", args, usage)
}
