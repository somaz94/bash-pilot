// Package testutil provides shared helpers for unit tests across the
// bash-pilot internal packages. Only test files (_test.go) should import this
// package; production code should not depend on it.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// WriteFile writes content to dir/name with mode 0600 and returns the full
// path. It fails the test via t.Fatal on error.
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

// MakeDir creates parent/name with the given mode and returns the full path.
// Callers pass the mode explicitly because some tests intentionally set
// stricter or laxer modes (e.g. 0700 for ~/.ssh, 0755 for .config). It fails
// the test via t.Fatal on error.
func MakeDir(t *testing.T, parent, name string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, mode); err != nil {
		t.Fatal(err)
	}
	return path
}
