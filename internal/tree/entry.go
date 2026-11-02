package tree

import (
	"io/fs"
	"os"
)

// Entry is what an entry on disk is, as the system's glossary defines it:
// a regular file, a folder, a symbolic link, or none of these.
type Entry int

// The entries the glossary tells apart.
const (
	Other        Entry = iota // neither a regular file, a folder nor a symbolic link, such as a socket
	RegularFile               // holds a file's data itself and is no symbolic link
	Folder                    // holds other entries and is no symbolic link
	SymbolicLink              // stands for another entry by its name
)

// KindOf tells what the entry at path is, given the type bits of its mode
// as os.Lstat or os.ReadDir report them. It follows no link.
func KindOf(path string, mode fs.FileMode) Entry {
	if k, ok := platformKind(path, mode); ok {
		return k
	}
	switch {
	case mode&fs.ModeSymlink != 0:
		return SymbolicLink
	case mode.IsDir():
		return Folder
	case mode.IsRegular():
		return RegularFile
	}
	return Other
}

// Lkind tells what the entry at path is, without following a link.
func Lkind(path string) (Entry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Other, err
	}
	return KindOf(path, info.Mode()), nil
}
