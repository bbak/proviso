package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// baseline.manifest.remove: manifest removes a junction without what it
// points to, standing alone or inside a folder that goes.
func TestManifestRemovesJunction(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system, "outside/x.md": "x"})
	target, _ := filepath.Abs("outside")
	junction := func(link string) {
		t.Helper()
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
			t.Skipf("cannot create a junction: %v %s", err, out)
		}
	}
	note, _ := filepath.Abs(filepath.Join("proviso", ".amendment.md"))
	coherence, _ := filepath.Abs(filepath.Join("proviso", ".coherence.json"))
	junction(note)
	if err := os.MkdirAll(coherence, 0o755); err != nil {
		t.Fatal(err)
	}
	junction(filepath.Join(coherence, "inner"))
	var out, errOut bytes.Buffer
	if status := run([]string{"manifest"}, &out, &errOut); status != 0 {
		t.Fatalf("manifest: %d %s", status, errOut.String())
	}
	for _, p := range []string{note, coherence} {
		if _, err := os.Lstat(p); err == nil {
			t.Errorf("%s stays", p)
		}
	}
	if data, err := os.ReadFile(filepath.Join(target, "x.md")); err != nil || string(data) != "x" {
		t.Errorf("the target changed: %q %v", data, err)
	}
}
