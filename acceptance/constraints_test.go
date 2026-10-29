package acceptance

// Constraints of the system: proviso/constraints.md. A constraint is checked
// by build or review; these are the parts a build can check.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestSystemPlatforms checks system.platforms as far as a build can: Proviso
// builds for Windows, macOS and Linux, as one static binary (system.go).
func TestSystemPlatforms(t *testing.T) {
	t.Parallel()
	for _, goos := range []string{"windows", "darwin", "linux"} {
		t.Run(goos, func(t *testing.T) {
			t.Parallel()
			out := filepath.Join(t.TempDir(), "proviso")
			build := exec.Command("go", "build", "-o", out, "github.com/bbak/proviso/cmd/proviso")
			build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos)
			if msg, err := build.CombinedOutput(); err != nil {
				t.Errorf("system.platforms: proviso does not build for %s: %v\n%s", goos, err, msg)
			}
		})
	}
}

// TestSystemGit checks system.git as far as a build can: no Go library
// reaches git, since the binary links no module beyond those agreed in
// system.dependencies. That git is run only to find a revision and read the
// tree's files is for review; that it is never written is the invariant
// baseline.run.writes.
func TestSystemGit(t *testing.T) {
	t.Parallel()
	out, err := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", "github.com/bbak/proviso/cmd/proviso").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	agreed := []string{"github.com/bbak/proviso/", "github.com/yuin/goldmark", "golang.org/x/text/"}
	for _, p := range strings.Fields(string(out)) {
		ok := false
		for _, a := range agreed {
			ok = ok || strings.HasPrefix(p, a)
		}
		if !ok {
			t.Errorf("system.git: the binary links %s, a module beyond those agreed", p)
		}
	}
}

// TestSystemBuildVersion checks system.build-version as far as a build can:
// a build given no version names its Proviso version dev, and a local build
// through the Taskfile gives the next patch version after the latest tag,
// with the pre-release dev, or no version where no tag exists. How a release
// is built from its tag the pack does not say; the release version a binary
// is given is checked by install.version.output. The module is copied into a
// repository of its own, whose tags the steps set one after the other, so
// the steps share it and run in order.
func TestSystemBuildVersion(t *testing.T) {
	t.Parallel()
	skipWithoutGit(t)
	if _, err := exec.LookPath("task"); err != nil {
		t.Skip("Task is not installed")
	}
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").Output()
	if err != nil {
		t.Fatalf("go list -m: %v", err)
	}
	module := strings.TrimSpace(string(out))
	dir := filepath.Join(t.TempDir(), "proviso")
	// Only the files git does not ignore are copied, as they stand in the
	// working tree: whatever is ignored, such as builds and local notes,
	// cannot reach the copy, and .git is never listed.
	cmd := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	cmd.Dir = module
	out, err = cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	for _, rel := range strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00") {
		p := filepath.Join(module, filepath.FromSlash(rel))
		// A file deleted from the working tree, or one that is no regular
		// file, is not copied.
		if fi, err := os.Lstat(p); err != nil || !fi.Mode().IsRegular() {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("copying the module: %v", err)
		}
		target := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
		t.Fatal("the copy of the module holds a .git entry; it would reach the module's own repository")
	}
	r := newRepoIn(t, dir, tree{})
	r.commit("start")

	name := "proviso"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	outside := t.TempDir()
	taskBuild := func(t *testing.T, want string) {
		t.Helper()
		cmd := exec.Command("task", "build")
		cmd.Dir = dir
		if msg, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("task build: %v\n%s", err, msg)
		}
		got := execBinary(t, filepath.Join(dir, "bin", name), outside, "version")
		if first := strings.SplitN(strings.ReplaceAll(got.stdout, "\r\n", "\n"), "\n", 2)[0]; first != "proviso "+want {
			t.Errorf("system.build-version: want the local build to name its Proviso version %s, got %q", want, first)
		}
	}

	t.Run("derived: a build given no version", func(t *testing.T) {
		got := execBinary(t, binary, outside, "version")
		if first := strings.SplitN(strings.ReplaceAll(got.stdout, "\r\n", "\n"), "\n", 2)[0]; first != "proviso dev" {
			t.Errorf("system.build-version: want a build given no version to name its Proviso version dev, got %q", first)
		}
	})
	t.Run("derived: through the Taskfile, no tag", func(t *testing.T) {
		taskBuild(t, "dev")
	})
	t.Run("derived: through the Taskfile, a commit after v0.1.0", func(t *testing.T) {
		r.t = t
		r.git("tag", "v0.1.0")
		r.commit("after the release")
		taskBuild(t, "0.1.1-dev")
	})
	t.Run("derived: through the Taskfile, at v0.2.0, the latest tag", func(t *testing.T) {
		r.t = t
		r.git("tag", "v0.2.0")
		taskBuild(t, "0.2.1-dev")
	})
}
