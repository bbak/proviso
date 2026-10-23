package acceptance

// Constraints of the system: proviso/constraints.md. A constraint is checked
// by build or review; these are the parts a build can check.

import (
	"os"
	"os/exec"
	"path/filepath"
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
