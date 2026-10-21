//go:build windows

package acceptance

import (
	bin "encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf16"
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

// makeJunction makes link a junction to the folder target. It sets the
// reparse point through the Windows API, so that no shell parses the paths.
func makeJunction(link, target string) error {
	target, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	if err := os.Mkdir(link, 0o755); err != nil {
		return err
	}
	p, err := syscall.UTF16PtrFromString(link)
	if err != nil {
		return err
	}
	const (
		fileFlagOpenReparsePoint = 0x00200000
		fsctlSetReparsePoint     = 0x000900A4
		ioReparseTagMountPoint   = 0xA0000003
	)
	h, err := syscall.CreateFile(p, syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|fileFlagOpenReparsePoint, 0)
	if err != nil {
		os.Remove(link)
		return err
	}

	// The mount point reparse buffer: the substitute name in NT form and
	// the print name, each followed by a NUL.
	sub := utf16.Encode([]rune(`\??\` + target))
	print := utf16.Encode([]rune(target))
	names := make([]uint16, 0, len(sub)+len(print)+2)
	names = append(append(append(append(names, sub...), 0), print...), 0)
	buf := make([]byte, 16+2*len(names))
	bin.LittleEndian.PutUint32(buf[0:], ioReparseTagMountPoint)
	bin.LittleEndian.PutUint16(buf[4:], uint16(8+2*len(names)))
	bin.LittleEndian.PutUint16(buf[8:], 0)                     // substitute name offset
	bin.LittleEndian.PutUint16(buf[10:], uint16(2*len(sub)))   // substitute name length
	bin.LittleEndian.PutUint16(buf[12:], uint16(2*len(sub)+2)) // print name offset
	bin.LittleEndian.PutUint16(buf[14:], uint16(2*len(print))) // print name length
	for i, c := range names {
		bin.LittleEndian.PutUint16(buf[16+2*i:], c)
	}
	var n uint32
	err = syscall.DeviceIoControl(h, fsctlSetReparsePoint, &buf[0], uint32(len(buf)), nil, 0, &n, nil)
	syscall.CloseHandle(h)
	if err != nil {
		os.Remove(link)
	}
	return err
}

// makeSpecial would make an entry that is neither a regular file, a folder,
// a symbolic link nor a junction. Windows has none that a test can place in
// a folder.
func makeSpecial(dir string) (string, error) {
	return "", errors.New("Windows has no such entry besides links and junctions, which their own subtests cover")
}

// makeUnwritable keeps the manifest.sum of the spec root root open without
// sharing it, so that no other process can write, replace or remove it until
// the returned function is called. Windows has no write permission bit on a
// folder that a test could take away.
func makeUnwritable(root string) (func(), error) {
	return makeUnreadable(filepath.Join(root, "manifest.sum"))
}
