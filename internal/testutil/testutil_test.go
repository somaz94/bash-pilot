package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := WriteFile(t, dir, "x.txt", "hello")
	if path != filepath.Join(dir, "x.txt") {
		t.Errorf("unexpected path %q", path)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("expected content %q, got %q", "hello", string(got))
	}
}

func TestMakeDir(t *testing.T) {
	parent := t.TempDir()
	// 0777 is narrowed by any common umask, so this fails without the chmod.
	path := MakeDir(t, parent, "sub", 0777)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory at %s", path)
	}
	if got := info.Mode().Perm(); got != 0777 {
		t.Errorf("mode = %04o, want 0777", got)
	}
}
