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
