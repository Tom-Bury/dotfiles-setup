package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStartupLinksConfigAndKeepsWritableState(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	paths := startupPaths{
		Dotfiles: filepath.Join(root, "dotfiles' $(literal)"),
		PiAuth:   filepath.Join(root, "auth"),
	}
	piSource := filepath.Join(paths.Dotfiles, "shared", ".pi", "agent")
	settings := filepath.Join(piSource, "settings.json")
	auth := filepath.Join(paths.PiAuth, "auth.json")
	piHome := filepath.Join(home, ".pi", "agent")
	yazi := filepath.Join(paths.Dotfiles, "shared", "yazi")
	for _, dir := range []string{piSource, paths.PiAuth, yazi,
		filepath.Join(piSource, "prompts"), filepath.Join(piHome, "prompts"), filepath.Join(piHome, "install")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	writeTestFile(t, settings, "{}\n")
	writeTestFile(t, auth, "{}\n")
	writeTestFile(t, filepath.Join(piHome, "install", "pi"), "installed")
	writeTestStartup(t, root, paths)
	for range 2 {
		got := runStartupShell(t, home, root, `
source ./startup.sh
source ./startup.sh
printf '%s\n' "${GOMODCACHE-unset}" "${GOPROXY-unset}" "${GOTOOLCHAIN-unset}" "${GOCACHE-unset}" "$YAZI_CONFIG_HOME" "$PI_CLIPBOARD_OSC52" "$PATH"
`)
		want := strings.Join([]string{"unset", "unset", "unset", "unset", yazi, "1", home + "/.local/bin:/usr/bin:/bin", ""}, "\n")
		if got != want {
			t.Fatalf("startup environment = %q, want %q", got, want)
		}
	}
	assertLink(t, filepath.Join(piHome, "settings.json"), settings)
	assertLink(t, filepath.Join(piHome, "auth.json"), auth)
	if info, err := os.Lstat(filepath.Join(piHome, "prompts")); err != nil || !info.IsDir() {
		t.Fatalf("local prompts directory was not preserved: %v", err)
	}
	if got := readTestFile(t, filepath.Join(piHome, "install", "pi")); got != "installed" {
		t.Fatal("Pi install was not preserved")
	}
}

func TestStartupWithoutDotfilesPreservesLocalAuth(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	paths := startupPaths{PiAuth: filepath.Join(root, "auth")}
	piHome := filepath.Join(home, ".pi", "agent")
	for _, dir := range []string{piHome, paths.PiAuth} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	auth := filepath.Join(paths.PiAuth, "auth.json")
	localAuth := filepath.Join(piHome, "auth.json")
	writeTestFile(t, auth, "shared")
	writeTestFile(t, localAuth, "local")
	writeTestStartup(t, root, paths)
	if got := runStartupShell(t, home, root, `source ./startup.sh; printf '%s' "${GOMODCACHE-unset}"`); got != "unset" {
		t.Fatalf("set GOMODCACHE without a mount: %q", got)
	}
	if got := readTestFile(t, localAuth); got != "local" {
		t.Fatalf("local auth was not preserved: %q", got)
	}
	if err := os.Remove(localAuth); err != nil {
		t.Fatal(err)
	}
	runStartupShell(t, home, root, `source ./startup.sh`)
	assertLink(t, localAuth, auth)
}

func TestInstallerAddsStartupHookOnce(t *testing.T) {
	root := t.TempDir()
	persistent := filepath.Join(root, "persistent.sh")
	writeTestFile(t, persistent, "export KEEP=original\n")
	installer := readTestFile(t, "../../../docker/main.sh")
	_, tail, _ := strings.Cut(installer, "\ninstall_dsb_startup() {")
	body, _, _ := strings.Cut(tail, "\n}\n")
	body = strings.ReplaceAll(body, "/etc/sandbox-persistent.sh", persistent)
	body = strings.ReplaceAll(body, "/mnt/dsb/startup.sh", filepath.Join(root, "startup.sh"))
	runStartupShell(t, root, root, "run_sudo() { \"$@\"; }\ninstall_dsb_startup() {"+body+"\n}\ninstall_dsb_startup\ninstall_dsb_startup")
	got := readTestFile(t, persistent)
	if strings.Count(got, ">>> dotfiles-setup dsb startup >>>") != 1 {
		t.Fatalf("unexpected startup hook: %q", got)
	}
	writeTestFile(t, filepath.Join(root, "startup.sh"), "export STARTED=yes\n")
	if got := runStartupShell(t, root, root, `source ./persistent.sh; printf '%s:%s' "$KEEP" "$STARTED"`); got != "original:yes" {
		t.Fatalf("hook did not preserve existing config and load dsb: %q", got)
	}
}

func runStartupShell(t *testing.T, home, dir, script string) string {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("the sandbox hook needs Linux shell tools")
	}
	cmd := exec.CommandContext(t.Context(), "bash", "--noprofile", "--norc", "-euc", script)
	cmd.Dir = dir
	cmd.Env = []string{"HOME=" + home, "PATH=/usr/bin:/bin"}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("startup shell: %v: %s", err, out)
	}
	return string(out)
}

func writeTestStartup(t *testing.T, dir string, paths startupPaths) {
	t.Helper()
	var script bytes.Buffer
	if err := startupPlan(dir, paths).Render(&script); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "startup.sh"), script.String())
}

func assertLink(t *testing.T, path, want string) {
	t.Helper()
	if got, err := os.Readlink(path); err != nil || got != want {
		t.Fatalf("readlink %s = %q, %v; want %q", path, got, err, want)
	}
}
