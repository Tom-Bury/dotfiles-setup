package main

import (
	"os"
	"path/filepath"
)

const allowedDomainsEnv = "DOCKER_SBX_ALLOWED_DOMAINS"

type settings struct {
	baseSandboxName     string
	projectTemplateName string
	dotfilesDir         string
	sharedDir           string
	piAuthDir           string
	allowedDomains      string
	runtimeDir          string
}

func loadSettings(home string) (settings, error) {
	stateDir := filepath.Join(home, ".dsb")

	cfg := settings{
		baseSandboxName:     "tom-base",
		projectTemplateName: "tom-dev:v1",
		dotfilesDir:         filepath.Join(home, "personal/code/dotfiles-setup"),
		sharedDir:           filepath.Join(stateDir, "shared"),
		piAuthDir:           filepath.Join(stateDir, ".pi", "auth"),
		allowedDomains:      "auth.openai.com,proxy.golang.org,sum.golang.org,storage.googleapis.com,registry.npmjs.org",
		runtimeDir:          filepath.Join(stateDir, "runtime"),
	}
	values := map[string]*string{
		"DOCKER_SBX_BASE_NAME":    &cfg.baseSandboxName,
		"DOCKER_SBX_TEMPLATE":     &cfg.projectTemplateName,
		"DOCKER_SBX_DOTFILES_DIR": &cfg.dotfilesDir,
		"DOCKER_SBX_SHARED_DIR":   &cfg.sharedDir,
		"DOCKER_SBX_PI_AUTH_DIR":  &cfg.piAuthDir,
		allowedDomainsEnv:         &cfg.allowedDomains,
	}
	for key, target := range values {
		if value := os.Getenv(key); value != "" {
			*target = value
		}
	}
	return cfg, nil
}
