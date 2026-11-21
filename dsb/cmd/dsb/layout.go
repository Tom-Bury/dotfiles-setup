package main

import (
	"context"

	"dsb/internal/sandbox"
)

const (
	dotfilesMountPath = "/mnt/dotfiles"
	piAuthMountPath   = "/mnt/pi-auth"
	startupMountPath  = "/mnt/dsb"
)

func (app launcher) baseMounts(ctx context.Context, spec sandbox.Creation) ([]sandbox.Mount, error) {
	return app.withStartup(nil, spec, startupPaths{PiAuth: piAuthMountPath}), nil
}

func (app launcher) projectMounts(ctx context.Context, spec sandbox.Creation) ([]sandbox.Mount, error) {
	paths := startupPaths{PiAuth: piAuthMountPath, Dotfiles: spec.Workspace}
	mounts := []sandbox.Mount{piAuthMount(app.piAuthDir)}
	dotfilesDir, err := sandbox.ResolveDirectory(app.dotfilesDir)
	if err != nil || dotfilesDir != spec.Workspace {
		paths.Dotfiles = dotfilesMountPath
		mounts = append(mounts, sandbox.Mount{Source: app.dotfilesDir, Target: dotfilesMountPath, ReadOnly: true})
	}
	mounts = append(mounts, sandbox.Mount{Source: app.sharedDir, Target: "/mnt/shared", CreateDir: true})
	return app.withStartup(mounts, spec, paths), nil
}

func (app launcher) withStartup(mounts []sandbox.Mount, spec sandbox.Creation, paths startupPaths) []sandbox.Mount {
	dir := sandbox.StartupDir(app.runtimeDir, spec.Name)
	return append(mounts, startupPlan(dir, paths).Mount(startupMountPath))
}
