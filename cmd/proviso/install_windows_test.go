package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// install.repository-root: a .git folder marked hidden, as Git for Windows
// marks it, counts as well.
func TestIsRepositoryRootHidden(t *testing.T) {
	dir := t.TempDir()
	git := filepath.Join(dir, ".git")
	if err := os.Mkdir(git, 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := syscall.UTF16PtrFromString(git)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_DIRECTORY); err != nil {
		t.Fatal(err)
	}
	if !isRepositoryRoot(dir) {
		t.Error("a hidden .git folder is not taken as a repository root")
	}
}
