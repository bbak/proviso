package tree

import (
	"io/fs"
	"syscall"
)

const (
	nameSurrogate = 0x20000000 // the bit IsReparseTagNameSurrogate tests
	tagAFUnix     = 0x80000023 // IO_REPARSE_TAG_AF_UNIX
)

// held returns the type of the entry at path that Go reports as
// irregular, read from the folder's listing as os.ReadDir reads it.
func held(path string, m fs.FileMode) fs.FileMode {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return m
	}
	var fd syscall.Win32finddata
	h, err := syscall.FindFirstFile(p, &fd)
	if err != nil {
		return m
	}
	syscall.FindClose(h)
	return heldType(fd.FileAttributes, fd.Reserved0, m)
}

// heldType returns the type of an entry with the attributes attrs and the
// reparse tag tag that Go reports as m. A reparse point that is neither a
// name surrogate, which is a link, nor a socket counts as what it holds:
// a folder, or a regular file. A cloud file is such a reparse point.
func heldType(attrs, tag uint32, m fs.FileMode) fs.FileMode {
	switch {
	case attrs&syscall.FILE_ATTRIBUTE_REPARSE_POINT == 0, tag&nameSurrogate != 0, tag == tagAFUnix:
		return m
	case attrs&syscall.FILE_ATTRIBUTE_DIRECTORY != 0:
		return fs.ModeDir
	}
	return 0
}
