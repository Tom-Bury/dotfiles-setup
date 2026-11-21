// Package sandbox wraps the Docker Sandboxes command line tool and prepares dsb mounts.
package sandbox

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Client runs sbx with the supplied streams.
type Client struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// New returns a client that connects sbx to the supplied streams.
func New(stdin io.Reader, stdout, stderr io.Writer) Client {
	return Client{stdin: stdin, stdout: stdout, stderr: stderr}
}

// Exists reports whether sbx lists the named sandbox.
func (client Client) Exists(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "sbx", "ls")
	cmd.Stderr = client.stderr
	listing, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("sbx ls: %w", err)
	}

	for line := range strings.SplitSeq(string(listing), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true, nil
		}
	}

	return false, nil
}

// RunSpec describes how to open an existing sandbox.
type RunSpec struct {
	Name         string
	NetworkAllow []string
	Startup      *Startup // Nil leaves the startup script unchanged.
}

// Run writes the optional startup script, adds network rules, and opens the sandbox.
// It does not remove network rules. An empty list leaves the policy unchanged.
func (client Client) Run(ctx context.Context, spec RunSpec) error {
	if spec.Startup != nil {
		if err := WriteStartup(spec.Startup.Dir, spec.Startup.Render); err != nil {
			return fmt.Errorf("refresh startup for %q: %w", spec.Name, err)
		}
	}
	if len(spec.NetworkAllow) > 0 {
		if err := client.run(ctx, "policy", "allow", "network", "--sandbox", spec.Name, strings.Join(spec.NetworkAllow, ",")); err != nil {
			return err
		}
	}

	return client.run(ctx, "run", "--name", spec.Name)
}

// RemoveAll removes all local sandboxes, including those with active connections.
func (client Client) RemoveAll(ctx context.Context) error {
	return client.run(ctx, "rm", "--all", "--force")
}

// RemoveNetworkRule removes the sandbox-specific network allow rule.
func (client Client) RemoveNetworkRule(ctx context.Context, name, domain string) error {
	return client.run(ctx, "policy", "rm", "network", "--sandbox", name, domain)
}

// SaveTemplate saves a sandbox as a named template.
func (client Client) SaveTemplate(ctx context.Context, name, template string) error {
	return client.run(ctx, "template", "save", name, template)
}

func (client Client) run(ctx context.Context, subcommand string, args ...string) error {
	cmd := exec.CommandContext(ctx, "sbx", append([]string{subcommand}, args...)...)
	cmd.Stdin = client.stdin
	cmd.Stdout = client.stdout
	cmd.Stderr = client.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sbx %s: %w", subcommand, err)
	}
	return nil
}
