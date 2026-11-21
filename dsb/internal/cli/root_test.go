package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	if err := Execute(t.Context(), nil, []string{"--help"}, &out); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	for _, command := range []string{"dsb task run [NAME]", "dsb auto", "dsb open", "dsb base", "dsb base save", "dsb nuke"} {
		if !strings.Contains(out.String(), command) {
			t.Errorf("help does not list %q", command)
		}
	}
	if strings.Contains(out.String(), "dsb submit") {
		t.Error("help lists removed command dsb submit")
	}
}

func TestInvalidCommands(t *testing.T) {
	for _, args := range [][]string{
		{"typo"}, {"base", "typo"}, {"base", "save", "extra"}, {"--help", "extra"}, {"nuke", "extra"},
		{"task", "new"}, {"submit", "demo"}, {"task", "run", ""}, {"auto", "--once"},
	} {
		err := Execute(t.Context(), nil, args, io.Discard)
		if err == nil || !strings.HasPrefix(err.Error(), "invalid command") {
			t.Errorf("Execute(%q) error = %v, want invalid command error", args, err)
		}
	}
}
