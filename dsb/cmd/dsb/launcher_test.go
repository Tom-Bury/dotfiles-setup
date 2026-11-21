package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenBase(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	t.Chdir(launcher.dotfilesDir)

	if err := launcher.OpenBase(context.Background()); err != nil {
		t.Fatal(err)
	}

	assertSbxCall(t, "create", "--name", launcher.baseSandboxName, "shell",
		launcher.dotfilesDir,
		startupDir(launcher.runtimeDir, launcher.baseSandboxName)+":/mnt/dsb:ro")
	assertInteractiveCalls(t, launcher.baseSandboxName, "ls\ncreate\npolicy\nrun\n")
}

func TestReuseBase(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	t.Chdir(launcher.dotfilesDir)
	t.Setenv("EXISTING_SANDBOX", launcher.baseSandboxName)
	dir := startupDir(launcher.runtimeDir, launcher.baseSandboxName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "startup.sh")
	const legacy = "# legacy script without a mount record\n"
	writeTestFile(t, script, legacy)

	if err := launcher.OpenBase(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := readTestFile(t, script); got != legacy {
		t.Errorf("legacy startup script = %q, want %q", got, legacy)
	}
	assertInteractiveCalls(t, launcher.baseSandboxName, "ls\nrun\n")
}

func TestOpenProject(t *testing.T) {
	t.Run("dotfiles", func(t *testing.T) {
		launcher := launcherWithFakeSbx(t)
		testProjectCreation(t, launcher, launcher.dotfilesDir)
	})
	t.Run("dotfiles through a symlink", func(t *testing.T) {
		launcher := launcherWithFakeSbx(t)
		projectDir := launcher.dotfilesDir
		launcher.dotfilesDir = filepath.Join(t.TempDir(), "dotfiles-link")
		if err := os.Symlink(projectDir, launcher.dotfilesDir); err != nil {
			t.Fatal(err)
		}
		t.Chdir(projectDir)

		if err := launcher.OpenProject(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
		assertSbxCall(t, "create", "--name", filepath.Base(projectDir),
			"--template", launcher.projectTemplateName, "shell", projectDir,
			launcher.piAuthDir+":/mnt/pi-auth", launcher.sharedDir+":/mnt/shared",
			startupDir(launcher.runtimeDir, filepath.Base(projectDir))+":/mnt/dsb:ro")
	})
	t.Run("separate project", func(t *testing.T) {
		launcher := launcherWithFakeSbx(t)
		testProjectCreation(t, launcher, canonicalTestDir(t))
	})
}

func testProjectCreation(t *testing.T, launcher launcher, projectDir string) {
	t.Helper()
	t.Chdir(projectDir)
	extraMount := "/tmp:/mnt/extra:ro"
	if err := launcher.OpenProject(context.Background(), []string{extraMount}); err != nil {
		t.Fatal(err)
	}

	projectName := filepath.Base(projectDir)
	want := []string{"--name", projectName, "--template", launcher.projectTemplateName, "shell",
		projectDir,
		launcher.piAuthDir + ":/mnt/pi-auth"}
	if projectDir != launcher.dotfilesDir {
		want = append(want, launcher.dotfilesDir+":/mnt/dotfiles:ro")
	}
	startup := startupDir(launcher.runtimeDir, projectName)
	want = append(want, launcher.sharedDir+":/mnt/shared", startup+":/mnt/dsb:ro", extraMount)
	readTestFile(t, filepath.Join(startup, "startup.sh"))
	assertSbxCall(t, "create", want...)
	assertInteractiveCalls(t, projectName, "ls\ncreate\npolicy\nrun\n")
}

func TestReuseProject(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	t.Chdir(launcher.dotfilesDir)
	projectName := filepath.Base(launcher.dotfilesDir)
	t.Setenv("EXISTING_SANDBOX", projectName)
	ctx := context.Background()

	if err := launcher.OpenProject(ctx, []string{"extra"}); err == nil {
		t.Fatal("accepted mounts for an existing sandbox")
	}
	assertCallOrder(t, "ls\n")
	if err := launcher.OpenProject(ctx, nil); err != nil {
		t.Fatal(err)
	}
	assertInteractiveCalls(t, projectName, "ls\nls\npolicy\nrun\n")
}

func TestReopenKeepsStartup(t *testing.T) {
	app := launcherWithFakeSbx(t)
	t.Chdir(app.dotfilesDir)
	name := filepath.Base(app.dotfilesDir)
	if err := app.OpenProject(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(startupDir(app.runtimeDir, name), "startup.sh")
	const original = "# existing startup\n"
	writeTestFile(t, script, original)
	t.Setenv("EXISTING_SANDBOX", name)
	if err := app.OpenProject(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	if got := readTestFile(t, script); got != original {
		t.Errorf("startup changed: %q", got)
	}
}

func TestOpenProjectRejectsInvalidManagedMount(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	t.Chdir(launcher.dotfilesDir)
	launcher.sharedDir = "relative/shared"

	if err := launcher.OpenProject(context.Background(), nil); err == nil {
		t.Fatal("opened a project with an invalid shared mount")
	}
	assertCallOrder(t, "ls\n")
}

func TestNuke(t *testing.T) {
	app := launcherWithFakeSbx(t)
	if err := app.Nuke(t.Context()); err != nil {
		t.Fatal(err)
	}
	assertCallOrder(t, "rm\n")
	assertSbxCall(t, "rm", "--all", "--force")
}

func TestSaveProjectTemplate(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	if err := launcher.SaveProjectTemplate(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertCallOrder(t, "template\n")
	assertSbxCall(t, "template", "save", launcher.baseSandboxName, launcher.projectTemplateName)
}
