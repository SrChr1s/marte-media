package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDirMigratesOldName(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	old := filepath.Join(base, "marte-desc")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(old, "history.json"), []byte("[]"), 0o644)

	dir, err := DefaultDir()
	if err != nil || dir != filepath.Join(base, "marte-media") {
		t.Fatalf("dir %q, err %v", dir, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "history.json")); err != nil {
		t.Fatalf("history was not carried over: %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("the old folder should be gone")
	}
}
