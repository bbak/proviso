package tree

import (
	"io/fs"
	"syscall"
)

const (
	reparseTagNameSurrogate = 0x20000000 // the bit that marks a name surrogate
	reparseTagAFUnix        = 0x80000023 // IO_REPARSE_TAG_AF_UNIX, a socket
)

// platformKind tells a reparse point apart by its tag, following
// system.symbolic-link and system.regular-file: a name surrogate, such as a
// symbolic link, a junction or a mount point, is a symbolic link; any
// other reparse point counts as what it holds, a folder where it holds
// other entries, such as a OneDrive folder, neither where it holds no data,
// such as a socket, and a regular file otherwise, such as a OneDrive
// placeholder or a deduplicated file. Go marks every reparse point but a
// deduplicated file as a symbolic link or as irregular, and one of these
// only, so an entry without either mark is left to its mode.
func platformKind(path string, mode fs.FileMode) (Entry, bool) {
	if mode&(fs.ModeSymlink|fs.ModeIrregular) == 0 {
		return Other, false
	}
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return Other, true
	}
	// FindFirstFile looks at the entry itself, never at what it stands for,
	// and does not fetch a OneDrive placeholder's data.
	var d syscall.Win32finddata
	h, err := syscall.FindFirstFile(p, &d)
	if err != nil {
		return Other, true
	}
	syscall.FindClose(h)
	if d.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		return Other, false
	}
	switch tag := d.Reserved0; {
	case tag&reparseTagNameSurrogate != 0:
		return SymbolicLink, true
	case tag == reparseTagAFUnix:
		return Other, true
	case d.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0:
		return Folder, true
	}
	return RegularFile, true
}
