//go:build !windows

package acceptance

import "errors"

// markHidden fails: the hidden attribute is Windows' own.
func markHidden(path string) error {
	return errors.New("the hidden attribute exists on Windows only")
}
