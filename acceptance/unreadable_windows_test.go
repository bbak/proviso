//go:build windows

package acceptance

import (
	"errors"
	"fmt"
	"os/exec"
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

// makeJunction makes link a junction to the folder target.
func makeJunction(link, target string) error {
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, out)
	}
	return nil
}

// makeSpecial would make an entry that is neither a regular file, a folder,
// a symbolic link nor a junction. Windows has none that a test can place in
// a folder.
func makeSpecial(dir string) (string, error) {
	return "", errors.New("Windows has no such entry besides links and junctions, which their own subtests cover")
}
