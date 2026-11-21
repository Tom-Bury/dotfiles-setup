package main

import (
	"path/filepath"
	"testing"
)

func TestSettingsDefaultsAndOverrides(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DOCKER_SBX_BASE_NAME", "custom")
	cfg, err := loadSettings(home)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.baseSandboxName != "custom" {
		t.Fatalf("overrides not applied: %+v", cfg)
	}
	if cfg.allowedDomains != "auth.openai.com,proxy.golang.org,sum.golang.org,storage.googleapis.com,registry.npmjs.org" {
		t.Errorf("domains = %q", cfg.allowedDomains)
	}
	if cfg.runtimeDir != filepath.Join(home, ".dsb", "runtime") {
		t.Errorf("runtime dir = %q", cfg.runtimeDir)
	}
}
