package main

import (
	"os"
	"path/filepath"
	"testing"
)

// install.repository-root: any entry named .git makes a repository root,
// and only the directory itself is looked at.
func TestIsRepositoryRoot(t *testing.T) {
	clone := t.TempDir()
	if err := os.Mkdir(filepath.Join(clone, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(clone, "src")
	if err := os.Mkdir(src, 0o755); err != nil {
		t.Fatal(err)
	}
	worktree := t.TempDir()
	if err := os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		dir  string
		want bool
	}{
		{clone, true},
		{worktree, true},
		{src, false},
		{t.TempDir(), false},
	} {
		if got := isRepositoryRoot(c.dir); got != c.want {
			t.Errorf("isRepositoryRoot(%s) = %v, want %v", c.dir, got, c.want)
		}
	}
}
