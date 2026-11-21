package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSymlinkWriteShell(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the startup script uses Linux shell tools")
	}
	for _, tt := range []struct {
		name             string
		absoluteTarget   bool
		preserveExisting bool
	}{
		{name: "replace broken link"},
		{name: "preserve broken link", preserveExisting: true},
		{name: "absolute target", absoluteTarget: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			source := filepath.Join(home, "source '$(literal)'")
			if err := os.WriteFile(source, nil, 0600); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(home, "target with spaces")
			broken := filepath.Join(home, "missing")
			if err := os.Symlink(broken, target); err != nil {
				t.Fatal(err)
			}
			link := Symlink{Source: source, Target: filepath.Base(target), PreserveExisting: tt.preserveExisting}
			if tt.absoluteTarget {
				link.Target = target
			}
			var script strings.Builder
			if err := link.WriteShell(&script); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(t.Context(), "bash", "--noprofile", "--norc", "-euc", script.String())
			cmd.Env = []string{"HOME=" + home, "PATH=/usr/bin:/bin"}
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("run link command: %v: %s", err, out)
			}
			want := source
			if tt.preserveExisting {
				want = broken
			}
			if got, err := os.Readlink(target); err != nil || got != want {
				t.Errorf("os.Readlink(%q) = %q, %v; want %q", target, got, err, want)
			}
		})
	}
}
