package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"dsb/internal/sandbox"
)

func piAuthMount(hostPath string) sandbox.Mount {
	return sandbox.Mount{
		Source:    hostPath,
		Target:    piAuthMountPath,
		CreateDir: true,
		Prepare:   securePiAuthDir,
	}
}

func securePiAuthDir(authDir string) error {
	if err := os.Chmod(authDir, 0700); err != nil {
		return err
	}

	authPath := filepath.Join(authDir, "auth.json")
	if err := createAuthFileIfMissing(authPath); err != nil {
		return err
	}
	return restrictAuthFile(authPath)
}

func createAuthFileIfMissing(authPath string) error {
	authFile, err := os.OpenFile(authPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}

	_, writeErr := authFile.WriteString("{}\n")
	return errors.Join(writeErr, authFile.Close())
}

func restrictAuthFile(authPath string) (err error) {
	// A running sandbox can replace this entry. Do not follow a link or wait on a pipe.
	authFile, err := os.OpenFile(authPath, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, authFile.Close()) }()

	authInfo, err := authFile.Stat()
	if err != nil {
		return err
	}
	if !authInfo.Mode().IsRegular() {
		return fmt.Errorf("not a regular auth file: %s", authPath)
	}
	return authFile.Chmod(0600)
}
