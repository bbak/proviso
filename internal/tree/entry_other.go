//go:build !windows

package tree

import "io/fs"

// platformKind leaves every entry to its mode: outside Windows, the mode
// tells a symbolic link, a folder and a regular file apart.
func platformKind(string, fs.FileMode) (Entry, bool) {
	return Other, false
}
