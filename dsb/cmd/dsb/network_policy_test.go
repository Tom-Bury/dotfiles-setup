package main

import (
	"context"
	"testing"
)

func TestOpenBaseReuseDoesNotChangeAllowlist(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	t.Chdir(launcher.dotfilesDir)
	t.Setenv("EXISTING_SANDBOX", launcher.baseSandboxName)
	launcher.allowedDomains = "api.openai.com,  updates.example.com "

	if err := launcher.OpenBase(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertCallOrder(t, "ls\nrun\n")
}

func TestBaseCreationIgnoresConfiguredAllowlist(t *testing.T) {
	launcher := launcherWithFakeSbx(t)
	t.Chdir(launcher.dotfilesDir)
	launcher.allowedDomains = "api.openai.com,**"
	if err := launcher.OpenBase(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertSbxCall(t, "policy", "allow", "network", "--sandbox", launcher.baseSandboxName, "*")
}

func TestAllowlistRejectsInvalidDomains(t *testing.T) {
	for _, invalidDomain := range []string{
		"https://example.com", "example.com:443", "example.com/path", "127.0.0.1",
		"*.example.com", "-example.com", "foo..example.com", "example.com,",
		"example.com other.example.com",
	} {
		if _, err := validateAllowedDomains(invalidDomain); err == nil {
			t.Errorf("accepted %q", invalidDomain)
		}
	}
}
