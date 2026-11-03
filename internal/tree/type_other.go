//go:build !windows

package tree

import "io/fs"

// held returns m: only Windows reports what an entry holds as irregular.
func held(path string, m fs.FileMode) fs.FileMode { return m }
