// Package model holds the system parameters about the Proviso Model that
// more than one capability reads: the model versions this binary reads
// and the one it writes, and the releases that write each one.
package model

import (
	"fmt"
	"strconv"
)

// Version is the parameter model_version: the version of the Proviso
// Model this binary writes, and the newest it reads.
const Version = 1

// Oldest is the parameter oldest_model_version: the oldest version of the
// Proviso Model this binary reads. It reads every version from Oldest to
// Version.
const Oldest = 0

// Reads reports whether v, a whole number in decimal digits, is a model
// version from Oldest to Version. A number too large for an int is newer
// than Version.
func Reads(v string) bool {
	n, err := strconv.Atoi(v)
	return err == nil && n >= Oldest && n <= Version
}

// Range names the model versions this binary reads, from Oldest to
// Version, as one number where both are the same.
func Range() string {
	if Oldest == Version {
		return fmt.Sprintf("model version %d", Version)
	}
	return fmt.Sprintf("model versions %d to %d", Oldest, Version)
}

// Release is a row of the parameter model_releases: a model version
// and the first release version that writes it.
type Release struct {
	Model int
	First string
}

// Releases is the parameter model_releases. Each row names the model
// version its releases write, and holds from its first release up to the
// next row's; the last row holds for every later release.
var Releases = []Release{
	{0, "0.1.0"},
	{1, "0.7.0"},
}
