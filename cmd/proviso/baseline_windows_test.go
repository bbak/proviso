package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// baseline.manifest.remove: a junction goes without what it points to,
// standing alone or inside a folder that goes.
func TestRemoveEntryJunction(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(target, "x.md"), []byte("x"), 0o644)
	junction := func(link string) {
		t.Helper()
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
			t.Skipf("cannot create a junction: %v %s", err, out)
		}
	}
	alone := filepath.Join(dir, ".amendment.md")
	junction(alone)
	if err := removeEntry(alone); err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(dir, ".coherence.json")
	os.MkdirAll(folder, 0o755)
	junction(filepath.Join(folder, "inner"))
	if err := removeEntry(folder); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{alone, folder} {
		if _, err := os.Lstat(p); err == nil {
			t.Errorf("%s stays", p)
		}
	}
	if data, err := os.ReadFile(filepath.Join(target, "x.md")); err != nil || string(data) != "x" {
		t.Errorf("the target changed: %q %v", data, err)
	}
}
