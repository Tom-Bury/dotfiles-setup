package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"dsb/internal/sandbox"
	"dsb/internal/task"
)

// RunTasks runs ready tasks until the queue ends or the process receives a stop signal.
func (app launcher) RunTasks(ctx context.Context, name string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	store := task.Store{Root: filepath.Join(app.sharedDir, "tasks")}
	return store.RunReady(ctx, name, app.executeTask)
}

// taskCommand never reads terminal input. Diagnostics go to the private run log.
func taskCommand(ctx context.Context, stderr, stdout io.Writer, program string, args ...string) error {
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil {
		// Arguments can contain the task prompt. Keep them out of the error.
		return fmt.Errorf("%s: %w", program, err)
	}
	return nil
}

func (app launcher) executeTask(ctx context.Context, dir string, cfg task.Settings) (runErr error) {
	logFile, err := os.Create(filepath.Join(dir, "run/stderr"))
	if err != nil {
		return err
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}()
	workspace, err := sandbox.ResolveDirectory(cfg.Workspace)
	if err != nil {
		return err
	}
	git := func(args ...string) (string, error) {
		var out strings.Builder
		err := taskCommand(ctx, logFile, &out, "git", append([]string{"-C", workspace}, args...)...)
		return strings.TrimSpace(out.String()), err
	}
	branch := "refs/heads/" + cfg.Branch
	if _, err := git("check-ref-format", branch); err != nil {
		return err
	}
	existing, err := git("for-each-ref", "--format=%(refname)", branch)
	if err != nil {
		return err
	}
	if existing != "" {
		return fmt.Errorf("host branch already exists: %s", cfg.Branch)
	}
	base, err := git("rev-parse", "--verify", "--end-of-options", cfg.Base+"^{commit}")
	if err != nil {
		return err
	}
	prompt, err := os.ReadFile(filepath.Join(dir, "TASK.md"))
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(prompt)) == "" {
		return errors.New("empty TASK.md")
	}
	domains, err := validateAllowedDomains(app.allowedDomains)
	if err != nil {
		return err
	}
	name := "dsb-task-" + strings.ToLower(rand.Text())
	if err := os.WriteFile(filepath.Join(dir, "run/sandbox"), []byte(name+"\n"), 0600); err != nil {
		return err
	}
	output := filepath.Join(dir, "output")
	if err := os.Mkdir(output, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	info, err := os.Lstat(output)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory, not a link", output)
	}
	spec := sandbox.Creation{
		Name:      name,
		Clone:     true,
		Template:  app.projectTemplateName,
		Workspace: workspace,
	}
	spec.Mounts = app.taskMounts(spec, dir, output, cfg.Mounts)
	defer func() {
		runErr = app.finishTask(name, logFile, runErr)
	}()

	if err := sandbox.New(nil, logFile, logFile).Create(ctx, spec); err != nil {
		return fmt.Errorf("create task sandbox: %w", err)
	}
	if len(domains) > 0 {
		err := taskCommand(ctx, logFile, logFile, "sbx", "--no-input", "policy", "allow", "network",
			"--sandbox", name, strings.Join(domains, ","))
		if err != nil {
			return fmt.Errorf("allow task domains: %w", err)
		}
	}

	inside := func(out io.Writer, args ...string) error {
		prefix := []string{"--no-input", "exec", "-w", workspace, name}
		return taskCommand(ctx, logFile, out, "sbx", append(prefix, args...)...)
	}
	if err := inside(logFile, "git", "fetch", "--no-tags", "/run/sandbox/source", base); err != nil {
		return fmt.Errorf("fetch task base: %w", err)
	}
	if err := inside(logFile, "git", "switch", "-c", cfg.Branch, base); err != nil {
		return fmt.Errorf("create sandbox branch: %w", err)
	}

	if err := runTaskPrompt(inside, dir, cfg, string(prompt)); err != nil {
		return err
	}
	return publishTaskBranch(taskBranch{ref: branch, base: base, remote: "sandbox-" + name}, inside, git)
}

func (app launcher) taskMounts(spec sandbox.Creation, dir, output string, inputs []string) []sandbox.Mount {
	mounts := []sandbox.Mount{
		piAuthMount(app.piAuthDir),
		{Source: dir, Target: "/mnt/task", ReadOnly: true},
		{Source: output, Target: "/mnt/output"},
	}
	for i, path := range inputs {
		mounts = append(mounts, sandbox.Mount{
			Source:   path,
			Target:   fmt.Sprintf("/mnt/input/%d", i+1),
			ReadOnly: true,
		})
	}
	return app.withStartup(mounts, spec, startupPaths{PiAuth: piAuthMountPath})
}

func (app launcher) finishTask(name string, logFile io.Writer, runErr error) error {
	// Cleanup must still work if the task deadline expired during creation.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if runErr == nil {
		runErr = taskCommand(ctx, logFile, logFile, "sbx", "--no-input", "rm", "--force", name)
		if runErr == nil {
			return os.RemoveAll(sandbox.StartupDir(app.runtimeDir, name))
		}
	}
	stopErr := taskCommand(ctx, logFile, logFile, "sbx", "--no-input", "stop", name)
	return errors.Join(runErr, stopErr)
}

type sandboxExec func(io.Writer, ...string) error

func runTaskPrompt(inside sandboxExec, dir string, cfg task.Settings, prompt string) error {
	response, err := os.Create(filepath.Join(dir, "run/response.md"))
	if err != nil {
		return err
	}
	const script = `source /mnt/dsb/startup.sh && exec pi --print --provider "$1" --model "$2" --thinking "$3" "$4"`
	const instructions = "Run unattended. Commit task changes on the current branch. Do not push. " +
		"Read inputs from /mnt/task and /mnt/input if present. " +
		"Write requested artifacts to /mnt/output. Report checks and limitations."
	promptText := "Task specification:\n\n" + prompt + "\n\n" + instructions
	promptErr := inside(response, "bash", "-lc", script, "dsb", cfg.Provider, cfg.Model, cfg.Thinking, promptText)
	return errors.Join(promptErr, response.Close())
}

type taskBranch struct {
	ref    string
	base   string
	remote string
}

func publishTaskBranch(branch taskBranch, inside sandboxExec, git func(...string) (string, error)) error {
	var status, ref, tip strings.Builder
	if err := inside(&status, "git", "status", "--porcelain", "--untracked-files=all"); err != nil {
		return err
	}
	if status.Len() != 0 {
		return errors.New("uncommitted work in sandbox")
	}
	if err := inside(&ref, "git", "symbolic-ref", "HEAD"); err != nil {
		return err
	}
	if strings.TrimSpace(ref.String()) != branch.ref {
		return errors.New("agent switched branches")
	}
	if err := inside(&tip, "git", "rev-parse", "HEAD"); err != nil {
		return err
	}
	oid := strings.TrimSpace(tip.String())
	if _, err := git("fetch", "--no-tags", branch.remote, branch.ref); err != nil {
		return err
	}
	if _, err := git("merge-base", "--is-ancestor", branch.base, oid); err != nil {
		return err
	}
	// An empty old value makes branch creation atomic and refuses overwrites.
	_, err := git("update-ref", branch.ref, oid, "")
	return err
}
