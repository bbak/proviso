# Constraints

## <a id="system.go"></a> Proviso is one static Go binary

### Rationale

Engineers and agents install Proviso in every repository that holds a
spec tree. A single binary without a runtime to install, cross-compiled
for each platform, keeps that step to a download. Pure Go, without CGO,
is what makes the cross-compilation work without a C toolchain per
platform.

### Property

Proviso is written in Go and builds with `CGO_ENABLED=0` into one
statically linked binary named `proviso`, which needs nothing else
installed to run.

## <a id="system.platforms"></a> Proviso is written for Windows and Unix-like systems

### Rationale

Engineers keep spec trees on Windows, macOS and Linux, and a tree must
give the same result wherever it is checked. Proviso is developed and
tested on Windows only, so its portability rests on how it is written:
Go's standard library hides most differences between the platforms, and
those that still show, such as path separators, drive letters, line
endings and case in file names, are decided in the tree, where a test
can name them, rather than left to whichever platform the code met
first.

### Property

Proviso builds for Windows, macOS and Linux, and its code assumes no
one platform's path separator, line ending or rules for file names. A
difference between platforms that shows in what Proviso reports or
writes is decided in the spec tree, never by the code alone.

## <a id="system.binary-source"></a> The binary is built from `cmd/proviso/` only

### Rationale

The spec tree lives in a folder named `proviso/` in the repository root.
A binary of the same name built into the root would collide with it, so
the main package sits in `cmd/proviso/`, and no build writes `proviso`
into the root.

### Property

The only main package is `cmd/proviso/`, and no build step writes a file
named `proviso` or `proviso.exe` into the repository root.

## <a id="system.dependencies"></a> Dependencies are agreed one by one

### Rationale

Every dependency is code that the coding agent did not write and the
tests did not specify, and its defaults are decisions nobody made. Each
one is therefore a decision of its own, taken before code relies on it.

### Property

Proviso uses the Go standard library and, beyond it, only
`github.com/yuin/goldmark`, for reading Markdown as CommonMark with
source positions and GitHub's tables, and the package
`golang.org/x/text/unicode/norm`, for reading names in the same Unicode
form on every platform. Any further module is added here before code
uses it.

## <a id="system.acceptance-tests"></a> Acceptance tests belong to the testing agent

### Rationale

The tests that count at acceptance are written from the spec tree by a
testing agent that never sees the code, so that they check what the
tree says rather than what the code does. If the coding agent could
change them, a test could be bent to fit the code. Its own unit tests
are design aids and stay beside the code.

### Property

The acceptance tests and their test trees live in `acceptance/` and are
changed only by the testing agent. They run the binary as a black box,
built by the tests themselves. The coding agent's tests live beside its
code in `cmd/` and `internal/`, and the coding agent never changes
anything in `acceptance/`.
