package tree

import (
	"io/fs"
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

// A reparse point that is no link counts as what it holds; a link and a
// socket stay what Go reports.
func TestHeldType(t *testing.T) {
	const (
		reparse = syscall.FILE_ATTRIBUTE_REPARSE_POINT
		dir     = syscall.FILE_ATTRIBUTE_DIRECTORY
	)
	cases := []struct {
		name       string
		attrs, tag uint32
		want       fs.FileMode
	}{
		{"no reparse point", syscall.FILE_ATTRIBUTE_ARCHIVE, 0, fs.ModeIrregular},
		{"a cloud file being moved", reparse, 0x9000201a, 0},
		{"a cloud file", reparse, 0x9000601a, 0},
		{"a cloud folder", reparse | dir, 0x9000601a, fs.ModeDir},
		{"a deduplicated file", reparse, 0x80000013, 0},
		{"a symbolic link", reparse, 0xa000000c, fs.ModeIrregular},
		{"a junction", reparse | dir, 0xa0000003, fs.ModeIrregular},
		{"a socket", reparse, 0x80000023, fs.ModeIrregular},
	}
	for _, c := range cases {
		if got := heldType(c.attrs, c.tag, fs.ModeIrregular); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}
