package acceptance

// Constraints of the system: proviso/constraints.md. A constraint is checked
// by build or review; these are the parts a build can check.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestSystemPlatforms checks system.platforms as far as a build can: Proviso
// builds for Windows, macOS and Linux, as one static binary (system.go).
func TestSystemPlatforms(t *testing.T) {
	for _, goos := range []string{"windows", "darwin", "linux"} {
		t.Run(goos, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "proviso")
			build := exec.Command("go", "build", "-o", out, "github.com/bbak/proviso/cmd/proviso")
			build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos)
			if msg, err := build.CombinedOutput(); err != nil {
				t.Errorf("system.platforms: proviso does not build for %s: %v\n%s", goos, err, msg)
			}
		})
	}
}
