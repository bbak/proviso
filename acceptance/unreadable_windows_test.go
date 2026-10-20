//go:build windows

package acceptance

import "syscall"

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
