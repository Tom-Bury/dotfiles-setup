package main

import (
	"context"
	"fmt"
	"os"

	"dsb/internal/cli"
	"dsb/internal/sandbox"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dsb:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}

	cfg, err := loadSettings(home)
	if err != nil {
		return err
	}

	app := launcher{settings: cfg, client: sandbox.New(os.Stdin, os.Stdout, os.Stderr)}
	return cli.Execute(ctx, app, args, os.Stdout)
}
