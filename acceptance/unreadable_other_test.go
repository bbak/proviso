//go:build !windows

package acceptance

import (
	"errors"
	"os"
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
