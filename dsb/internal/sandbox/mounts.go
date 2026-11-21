package sandbox

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Mount binds a host directory to an absolute path in the sandbox.
type Mount struct {
	Source    string
	Target    string
	ReadOnly  bool
	CreateDir bool

	// Prepare runs on the resolved host directory before sbx creates the sandbox.
	Prepare func(dir string) error
}

func (mount Mount) arg() (string, error) {
	if !path.IsAbs(mount.Target) || strings.ContainsAny(mount.Target, ":\r\n\x00") {
		return "", fmt.Errorf("invalid sandbox mount path %q", mount.Target)
	}

	if mount.CreateDir {
		// Check the path before creating anything on the host.
		if err := validateHostPath(mount.Source); err != nil {
			return "", err
		}
		if err := os.MkdirAll(mount.Source, 0700); err != nil {
			return "", err
		}
	}

	dir, err := ResolveDirectory(mount.Source)
	if err != nil {
		return "", err
	}

	if mount.Prepare != nil {
		if err := mount.Prepare(dir); err != nil {
			return "", err
		}
	}

	arg := dir + ":" + mount.Target
	if mount.ReadOnly {
		arg += ":ro"
	}

	return arg, nil
}

// ResolveDirectory returns the real path of a host directory accepted by sbx.
func ResolveDirectory(path string) (string, error) {
	if err := validateHostPath(path); err != nil {
		return "", err
	}

	dir, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	if err := validateHostPath(dir); err != nil {
		return "", err
	}

	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("not a host directory: %q", dir)
	}

	return dir, nil
}

func validateHostPath(path string) error {
	if !filepath.IsAbs(path) || strings.ContainsAny(path, ":\r\n\x00") {
		return fmt.Errorf("invalid absolute host path %q", path)
	}
	return nil
}
