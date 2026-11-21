package sandbox

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteStartupKeepsPreviousScriptOnRenderFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "startup.sh")
	const original = "# original startup script\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	wantErr := errors.New("render failed")
	err := WriteStartup(dir, func(w io.Writer) error {
		if _, err := io.WriteString(w, "# incomplete script\n"); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("WriteStartup() error = %v, want %v", err, wantErr)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Errorf("startup script = %q, want %q", got, original)
	}
}
