package tree

import (
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

// A junction is a link as well: it is unreadable wherever it stands.
func TestLoadJunction(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"system.md": "# x", "target/a.md": ""})
	link := filepath.Join(root, "shared")
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, filepath.Join(root, "target")).CombinedOutput(); err != nil {
		t.Skipf("cannot create a junction: %v %s", err, out)
	}
	tr := Load(root)
	if want := []string{"shared"}; !slices.Equal(tr.Unreadable, want) {
		t.Errorf("unreadable %v, want %v", tr.Unreadable, want)
	}
}

// Windows' hidden attribute does not hide a file: only a dot does.
func TestLoadHiddenAttribute(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"system.md": "# x", "notes.md": ""})
	p, err := syscall.UTF16PtrFromString(filepath.Join(root, "notes.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_HIDDEN); err != nil {
		t.Fatal(err)
	}
	if tr := Load(root); !slices.Equal(tr.Unknown, []string{"notes.md"}) {
		t.Errorf("unknown %v, want [notes.md]", tr.Unknown)
	}
}
