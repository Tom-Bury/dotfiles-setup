package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"dsb/internal/sandbox"
)

type launcher struct {
	settings
	client sandbox.Client
}

func (app launcher) OpenProject(ctx context.Context, extraMounts []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("find current directory: %w", err)
	}

	workspace, err := sandbox.ResolveDirectory(cwd)
	if err != nil {
		return fmt.Errorf("resolve project workspace: %w", err)
	}

	return app.open(ctx, sandbox.Creation{
		Name:        filepath.Base(workspace),
		Template:    app.projectTemplateName,
		Workspace:   workspace,
		ExtraMounts: extraMounts,
	}, app.projectMounts)
}

func (app launcher) OpenBase(ctx context.Context) error {
	return app.openBase(ctx)
}

func (app launcher) openBase(ctx context.Context) error {
	spec := sandbox.Creation{Name: app.baseSandboxName, Workspace: app.dotfilesDir}
	exists, err := app.client.Exists(ctx, spec.Name)
	if err != nil {
		return err
	}
	if exists {
		return app.client.Run(ctx, sandbox.RunSpec{Name: spec.Name})
	}

	mounts, err := app.baseMounts(ctx, spec)
	if err != nil {
		return err
	}
	spec.Mounts = mounts
	if err := app.client.Create(ctx, spec); err != nil {
		return err
	}
	if err := app.client.Run(ctx, sandbox.RunSpec{Name: spec.Name, NetworkAllow: []string{"*"}}); err != nil {
		return err
	}
	if err := app.client.RemoveNetworkRule(ctx, spec.Name, "*"); err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, "Save this base as the project template? [y/N] ")
	var answer string
	if _, err := fmt.Fscan(os.Stdin, &answer); err != nil {
		return err
	}
	if answer == "y" || answer == "Y" {
		return app.SaveProjectTemplate(ctx)
	}
	return nil
}

func (app launcher) Nuke(ctx context.Context) error {
	return app.client.RemoveAll(ctx)
}

func (app launcher) SaveProjectTemplate(ctx context.Context) error {
	return app.client.SaveTemplate(ctx, app.baseSandboxName, app.projectTemplateName)
}

func (app launcher) open(ctx context.Context, spec sandbox.Creation, mountsFor func(context.Context, sandbox.Creation) ([]sandbox.Mount, error)) error {
	domains, err := validateAllowedDomains(app.allowedDomains)
	if err != nil {
		return err
	}

	exists, err := app.client.Exists(ctx, spec.Name)
	if err != nil {
		return err
	}
	run := sandbox.RunSpec{Name: spec.Name, NetworkAllow: domains}
	if exists {
		if len(spec.ExtraMounts) > 0 {
			return errors.New("extra mounts are create-time only")
		}
		return app.client.Run(ctx, run)
	}

	mounts, err := mountsFor(ctx, spec)
	if err != nil {
		return err
	}
	spec.Mounts = mounts
	if err := app.client.Create(ctx, spec); err != nil {
		return err
	}
	// Creation already wrote the script through the startup mount.
	return app.client.Run(ctx, run)
}
