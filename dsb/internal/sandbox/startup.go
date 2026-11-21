package sandbox

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// StartupDir returns the host directory for a named sandbox's startup script.
func StartupDir(runtimeDir, name string) string {
	// Hash the name so it cannot introduce host path components.
	return filepath.Join(runtimeDir, fmt.Sprintf("%x", sha256.Sum256([]byte(name))))
}

// Startup describes Bash to source in the sandbox, followed by links to ensure.
// Dir is the host directory mounted for the startup hook.
type Startup struct {
	Dir    string
	Script func(io.Writer) error
	Links  []Symlink
}

// Mount returns a read-only mount that writes startup.sh during creation.
func (s Startup) Mount(target string) Mount {
	return Mount{
		Source:    s.Dir,
		Target:    target,
		ReadOnly:  true,
		CreateDir: true,
		Prepare: func(dir string) error {
			return WriteStartup(dir, s.Render)
		},
	}
}

// Render writes the complete startup script.
func (s Startup) Render(w io.Writer) error {
	if s.Script != nil {
		if err := s.Script(w); err != nil {
			return err
		}
	}
	if len(s.Links) == 0 {
		return nil
	}
	if _, err := io.WriteString(w, "\nif ! (\n"); err != nil {
		return err
	}
	for _, link := range s.Links {
		if err := link.WriteShell(w); err != nil {
			return err
		}
		if _, err := io.WriteString(w, " || exit\n"); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "); then\n  printf '%s\\n' 'dsb: could not create startup links' >&2\nfi\n")
	return err
}

// WriteStartup replaces startup.sh in an existing host directory.
// If rendering fails, it leaves the previous script in place.
func WriteStartup(dir string, render func(io.Writer) error) error {
	file, err := os.CreateTemp(dir, ".startup-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()

	writeErr := render(file)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	// Publish the script only after it is fully written and closed.
	return os.Rename(file.Name(), filepath.Join(dir, "startup.sh"))
}
