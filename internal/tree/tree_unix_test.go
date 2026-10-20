//go:build unix

package tree

import (
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

// An entry that is neither a regular file nor a folder is unreadable
// wherever it stands, even where the layout does not list it.
func TestLoadPipe(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"system.md": "# x"})
	for _, name := range []string{"notes", "glossary.md", ".pipe"} {
		if err := syscall.Mkfifo(filepath.Join(root, name), 0o644); err != nil {
			t.Skipf("cannot create a named pipe: %v", err)
		}
	}
	tr := Load(root)
	if want := []string{"glossary.md", "notes"}; !slices.Equal(tr.Unreadable, want) {
		t.Errorf("unreadable %v, want %v", tr.Unreadable, want)
	}
	if len(tr.Unknown) != 0 {
		t.Errorf("unknown %v", tr.Unknown)
	}
}
