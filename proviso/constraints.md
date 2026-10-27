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

## <a id="system.git"></a> Git is reached through its command line

### Rationale

Proviso needs git only for history: which revision holds the baseline,
and what a file held there. The `git` command line is installed wherever
engineers work with the tree, and calling it adds no dependency. Asking
only those questions keeps every git call in one place, so that another
version control system could answer them later.

### Property

Proviso reads from git only by running the `git` command line, never
through a Go library, and only to find a revision and to list and read
the files of the spec tree at that revision. It never writes to git.

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

## <a id="system.test-names"></a> An acceptance test is named after the element it checks

### Rationale

Proviso's engineers' notes tell a build how to run the acceptance tests
of a list of identifiers, so that it runs only those of what an
amendment touches and may affect, and the whole suite once at its end.
With this pattern, that is one anchored pattern for Go's test runner,
such as `^TestBaselinePackNote(_.*)?$`: the suffix after an underscore
keeps the tests of one element together when they are spread over
several functions, and the anchor keeps out an element whose identifier
merely starts the same way. A test of the test trees themselves checks no element,
and every run needs it, since a broken test tree makes every other
result meaningless.

### Property

Every top-level acceptance test that checks an element is named `Test`
followed by the element's identifier, with the dots and hyphens
dropped and each part between them that starts with a letter starting
with a capital letter, while a digit stays as it is, such as
`TestInstallVersionStale` for `install.version.stale` and
`TestLintRunRoot2` for `lint.run.root-2`;
optionally followed by `_` and a suffix of letters and digits, such as
`TestBaselineDiffFormer_Pack`. Every such test belongs to exactly one
element, whose Examples rows and derived checks it holds; a check of
another element within it is an addition and never replaces a test
named after that element. A top-level test that checks only the
test trees is named `TestFixture` followed by letters and digits, the
first a capital letter, such as `TestFixtureLint`. The only other top-level test function is `TestMain`. No two identifiers in
the spec tree give the same name, none gives `TestMain`, and none gives
a name that starts with `TestFixture`.

## <a id="system.model-release"></a> A new model version starts with a minor or major release

### Rationale

Every change to the notation breaks an older binary, so a model version
never changes in a patch release, which promises to break nothing.
Whether a change takes a minor or a major release is decided when it is
released. Rising columns keep the lookup in `model_releases` to one row
for every release.

### Property

Every first release in `model_releases` is a release version whose
patch number is 0. From row to row, both the model version and the
first release rise, and the last row names `model_version`.

## <a id="system.build-version"></a> Only a release gives the binary a release version

### Rationale

A local build may hold changes no release has, so it must never pass
for one; naming the release it builds on still tells which one that is.
A pre-release of the next patch sorts after that release and before
any later one, as semantic versioning orders them. The git tag is the
only source of a release version, so nothing beside it has to be kept
in step.

### Property

A release is built from its git tag `v<release version>` and gives the
binary that release version. A local build through the Taskfile gives
it the next patch version after the latest tag, with the pre-release
`dev`, such as `0.1.1-dev` after `v0.1.0`, or no version where no tag
exists. A build given no version names its Proviso version `dev`.
