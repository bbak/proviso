// Package acceptance holds the black-box acceptance tests of `proviso lint`,
// `proviso manifest`, `proviso diff`, `proviso impact`, `proviso pack`,
// `proviso judge`, `proviso verify` and `proviso init`, written from the spec
// tree in proviso/ alone. The tests build the binary from cmd/proviso
// themselves and observe only its exit status, standard output and standard
// error, and the files and git repositories it works on.
package acceptance
