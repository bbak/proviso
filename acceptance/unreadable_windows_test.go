//go:build windows

package acceptance

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

// makeUnreadable keeps path open without sharing it, so that no other
// process can read it until the returned function is called. Windows has no
// read permission bit that a test could take away.
func makeUnreadable(path string) (func(), error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	return func() { syscall.CloseHandle(h) }, nil
}

// makeSpecial makes, in the tree written to dir, an entry that is neither a
// regular file nor a folder, and returns its name: here proviso/shared, a
// junction to the folder outside.
func makeSpecial(dir string) (string, error) {
	link := filepath.Join(dir, "proviso", "shared")
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, filepath.Join(dir, "outside")).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}
	return "shared", nil
}
