//go:build !windows

package acceptance

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// makeUnreadable takes the read permission from path until the returned
// function is called.
func makeUnreadable(path string) (func(), error) {
	if err := os.Chmod(path, 0); err != nil {
		return nil, err
	}
	restore := func() { os.Chmod(path, 0o644) }
	if f, err := os.Open(path); err == nil {
		f.Close()
		restore()
		return nil, errors.New("the file stays readable without permission (running as root?)")
	}
	return restore, nil
}

// makeSpecial makes, in the tree written to dir, an entry that is neither a
// regular file nor a folder, and returns its name: here the capability's
// glossary.md, a named pipe in place of the file.
func makeSpecial(dir string) (string, error) {
	path := filepath.Join(dir, filepath.FromSlash(glossaryPath))
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return "glossary.md", syscall.Mkfifo(path, 0o644)
}

// makeJunction fails: junctions exist only on Windows.
func makeJunction(link, target string) error {
	return errors.New("junctions exist only on Windows")
}

// makeUnwritable takes the write permission from the spec root root and its
// manifest.sum until the returned function is called.
func makeUnwritable(root string) (func(), error) {
	manifest := filepath.Join(root, "manifest.sum")
	if err := os.Chmod(manifest, 0o444); err != nil {
		return nil, err
	}
	if err := os.Chmod(root, 0o555); err != nil {
		os.Chmod(manifest, 0o644)
		return nil, err
	}
	restore := func() { os.Chmod(root, 0o755); os.Chmod(manifest, 0o644) }
	probe := filepath.Join(root, "probe")
	if f, err := os.Create(probe); err == nil {
		f.Close()
		os.Remove(probe)
		restore()
		return nil, errors.New("the folder stays writable without permission (running as root?)")
	}
	return restore, nil
}

// makeCoherenceUnwritable takes the write permission from the spec root root
// and its .coherence.json until the returned function is called; both stay
// readable.
func makeCoherenceUnwritable(root string) (func(), error) {
	file := filepath.Join(root, coherenceName)
	if err := os.Chmod(file, 0o444); err != nil {
		return nil, err
	}
	if err := os.Chmod(root, 0o555); err != nil {
		os.Chmod(file, 0o644)
		return nil, err
	}
	restore := func() { os.Chmod(root, 0o755); os.Chmod(file, 0o644) }
	probe := filepath.Join(root, "probe")
	if f, err := os.Create(probe); err == nil {
		f.Close()
		os.Remove(probe)
		restore()
		return nil, errors.New("the folder stays writable without permission (running as root?)")
	}
	return restore, nil
}
