// Package model holds the system parameters about the Proviso Model that
// more than one capability reads: the model version this binary reads, and
// the releases that read each one.
package model

// Version is the parameter model_version: the version of the Proviso
// Model this binary reads and writes.
const Version = 0

// Release is a row of the parameter model_releases: a model version
// and the first release version that reads it.
type Release struct {
	Model int
	First string
}

// Releases is the parameter model_releases. Each row holds from its
// first release up to the next row's; the last row holds for every later
// release.
var Releases = []Release{
	{0, "0.1.0"},
}
