package sandbox

import (
	"context"
	"fmt"
)

// Creation describes an sbx shell sandbox. An empty Template uses the default image.
type Creation struct {
	// Clone creates a private Git checkout without prompting.
	Clone     bool
	Name      string
	Template  string
	Workspace string
	Mounts    []Mount

	// ExtraMounts contains mount arguments in sbx syntax.
	ExtraMounts []string
}

// Create prepares the mount directories and creates the sandbox.
func (client Client) Create(ctx context.Context, spec Creation) error {
	workspace, err := ResolveDirectory(spec.Workspace)
	if err != nil {
		return fmt.Errorf("sandbox workspace: %w", err)
	}

	args := []string{"--name", spec.Name}
	if spec.Clone {
		args = append(args, "--clone", "--no-input")
	}

	if spec.Template != "" {
		args = append(args, "--template", spec.Template)
	}

	args = append(args, "shell", workspace)

	for _, mount := range spec.Mounts {
		arg, err := mount.arg()
		if err != nil {
			return fmt.Errorf("prepare mount %s: %w", mount.Target, err)
		}
		args = append(args, arg)
	}

	args = append(args, spec.ExtraMounts...)
	return client.run(ctx, "create", args...)
}
