package acceptance

// Feature baseline.manifest: proviso/capabilities/baseline/features/manifest.md,
// and the contracts system.manifest-file (proviso/contracts/manifest.md) and
// system.cli.manifest (proviso/contracts/cli.md).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestSystemManifestFile checks the bytes of manifest_file against the
// oracle: for the plain tree, every part's text is fixed by part_kinds, so
// the whole file follows from the contract. The format itself is checked
// for every manifest written (see manifestIn).
func TestSystemManifestFile(t *testing.T) {
	t.Parallel()
	t.Run("the plain tree, byte for byte", func(t *testing.T) {
		t.Parallel()
		got, want := manifestOf(t, plainTree()), manifestFor(plainParts())
		if got != want {
			t.Errorf("want the manifest\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("a tree of every element type, in the format", func(t *testing.T) {
		t.Parallel()
		m := manifestOf(t, richTree())
		if !strings.HasPrefix(m, "model 0\n") {
			t.Errorf("want the first line %q, got %q", "model 0", strings.SplitN(m, "\n", 2)[0])
		}
	})
}

func TestBaselineManifestWrite(t *testing.T) {
	t.Parallel()
	written := func(t *testing.T, dir string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(manifestPath)))
		if err != nil {
			t.Fatalf("want manifest.sum in the spec root: %v", err)
		}
		return string(b)
	}

	t.Run("row1 without a manifest", func(t *testing.T) {
		t.Parallel()
		dir := plainTree().write(t)
		r := manifestIn(t, dir, dir, filepath.Join(dir, "proviso"), nil)
		if r.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", r.exit, r.stderr)
		}
		if got, want := written(t, dir), manifestFor(plainParts()); got != want {
			t.Errorf("want the manifest\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("row2 with a manifest written before an edit", func(t *testing.T) {
		t.Parallel()
		dir := plainTree().write(t)
		manifestOK(t, dir)
		goal := "Buyers put goods into the cart."
		edited := plainTree().edit(featurePath, goal, "Buyers put many goods into the cart.")
		edited.writeTo(t, dir, false)
		manifestOK(t, dir)
		want := plainParts()
		for i := range want {
			if want[i].id == "cart.add" && want[i].part == "section Goal" {
				want[i].text = "Buyers put many goods into the cart."
			}
		}
		if got := written(t, dir); got != manifestFor(want) {
			t.Errorf("want the manifest replaced by\n%s\ngot\n%s", manifestFor(want), got)
		}
	})
	t.Run("derived: replaced, not updated", func(t *testing.T) {
		t.Parallel()
		stale := "model 0\ncart.gone\theading\th1:" + strings.Repeat("0", 64) + "\n"
		dir := plainTree().with(manifestPath, stale).write(t)
		manifestOK(t, dir)
		if got, want := written(t, dir), manifestFor(plainParts()); got != want {
			t.Errorf("want the former manifest replaced by\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("row3 with a manifest written from the same tree", func(t *testing.T) {
		t.Parallel()
		dir := richTree().write(t)
		first := manifestOK(t, dir)
		second := manifestOK(t, dir)
		if first != second {
			t.Errorf("want the same bytes as before, got\n%s\nthen\n%s", first, second)
		}
		// The file depends only on the tree: written in the other
		// order, into another folder, it is the same.
		other := t.TempDir()
		richTree().writeTo(t, other, true)
		if third := manifestOK(t, other); third != first {
			t.Errorf("want the same bytes for the same tree written in another order, got\n%s\nand\n%s", first, third)
		}
	})
	t.Run("row4 with .coherence.json", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil, gap("A gap."))
		// That the manifest is written and .coherence.json removed is
		// checked for every run that exits with 0 (see manifestIn).
		r.manifest()
		if r.hasCoherence() {
			t.Errorf("want %s removed", coherenceName)
		}
	})
	t.Run("row5 with .coherence.json and .amendment.md", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil)
		r.writeNote("- refunds\n")
		// That the manifest is written and both files removed is checked
		// for every run that exits with 0 (see manifestIn).
		r.manifest()
		if r.hasCoherence() || r.hasNote() {
			t.Errorf("want %s and %s removed, got %s: %v, %s: %v", coherenceName, noteName, coherenceName, r.hasCoherence(), noteName, r.hasNote())
		}
	})
	t.Run("derived: with .amendment.md only", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.writeNote("- refunds\n")
		r.manifest()
		if r.hasNote() {
			t.Errorf("want %s removed", noteName)
		}
	})
	t.Run("row6 without write permission", func(t *testing.T) {
		t.Parallel()
		former := manifestFor(plainParts())
		// The coherence file and the amendment note stay as they were, too
		// (see manifestIn).
		dir := plainTree().edit(featurePath, "Buyers put goods", "Buyers put many goods").with(manifestPath, former, coherencePath, "{}\n", notePath, "- refunds\n").write(t)
		root := filepath.Join(dir, "proviso")
		undo, err := makeUnwritable(root)
		if err != nil {
			t.Skipf("cannot take the write permission here: %v", err)
		}
		undo()
		r := manifestIn(t, dir, dir, root, func() func() {
			undo, err := makeUnwritable(root)
			if err != nil {
				t.Fatal(err)
			}
			return undo
		})
		// That the former manifest stays as it was is checked for every
		// run that does not exit with 0 (see manifestIn).
		if r.exit != 2 {
			t.Errorf("want exit status 2, got %d", r.exit)
		}
		if got := written(t, dir); got != former {
			t.Errorf("want the former manifest as it was, got\n%s", got)
		}
	})
	t.Run("derived: run inside the tree, written into the spec root", func(t *testing.T) {
		t.Parallel()
		dir := plainTree().write(t)
		root := filepath.Join(dir, "proviso")
		r := manifestIn(t, dir, filepath.Join(root, "capabilities", "cart"), root, nil)
		if r.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", r.exit, r.stderr)
		}
		if got, want := written(t, dir), manifestFor(plainParts()); got != want {
			t.Errorf("want the manifest\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("derived: the nearest spec root", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		for k, v := range plainTree().prefixed("a") {
			tr[k] = v
		}
		dir := tr.write(t)
		inner := filepath.Join(dir, "a", "proviso")
		// Only the inner manifest.sum may be written (see manifestIn).
		r := manifestIn(t, dir, filepath.Join(dir, "a"), inner, nil)
		if r.exit != 0 {
			t.Fatalf("want exit status 0, got %d\nstderr: %s", r.exit, r.stderr)
		}
		if got, want := written(t, filepath.Join(dir, "a")), manifestFor(plainParts()); got != want {
			t.Errorf("want the inner tree's manifest\n%s\ngot\n%s", want, got)
		}
	})
}

func TestBaselineManifestRemove(t *testing.T) {
	t.Parallel()
	// That the removed entries are gone after exit status 0, and that
	// nothing else changed, is checked for every run (see manifestIn).
	t.Run("row1 none", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.manifest()
	})
	t.Run("row2 .coherence.json and .amendment.md, both files", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil)
		r.writeNote("- refunds\n")
		r.manifest()
		if r.hasCoherence() || r.hasNote() {
			t.Errorf("want %s and %s removed", coherenceName, noteName)
		}
	})
	t.Run("row3 .amendment.md, a folder holding a file", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.writeNoteFolder()
		r.manifest()
		if r.hasNote() {
			t.Errorf("want the folder %s removed", noteName)
		}
	})
	t.Run("row4 .amendment.md, a link to a file of the tree", func(t *testing.T) {
		t.Parallel()
		// capabilities/x.md: the feature file cart.add stands in. Where no
		// symbolic link can be made, a junction to the folder of the
		// features stands in for the link.
		r := amendedRepo(t, iLimitThen)
		link := filepath.Join(r.root(), noteName)
		if err := os.Symlink(filepath.Join(r.dir, filepath.FromSlash(featurePath)), link); err != nil {
			if err := makeJunction(link, filepath.Dir(filepath.Join(r.dir, filepath.FromSlash(featurePath)))); err != nil {
				t.Skipf("cannot make a link here: %v", err)
			}
			t.Logf("a junction stands in for the link: %v", err)
		}
		// That what it points to is unchanged is checked by manifestIn.
		r.manifest()
		if r.hasNote() {
			t.Errorf("want the link %s removed", noteName)
		}
		if _, err := os.Stat(filepath.Join(r.dir, filepath.FromSlash(featurePath))); err != nil {
			t.Errorf("want %s unchanged: %v", featurePath, err)
		}
	})
	t.Run("row5 .coherence.json, which cannot be removed", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil)
		want := manifestOf(t, amended(impactTree(), iLimitThen))
		prepare := lockCoherence(t, r)
		res := manifestIn(t, r.dir, r.dir, r.root(), prepare)
		if res.exit != 2 {
			t.Fatalf("want exit status 2, got %d\nstderr: %s", res.exit, res.stderr)
		}
		b, err := os.ReadFile(filepath.Join(r.root(), "manifest.sum"))
		if err != nil || string(b) != want {
			t.Errorf("want manifest.sum written and left as written\n%s\ngot\n%s", want, b)
		}
		if !r.hasCoherence() {
			t.Errorf("want %s, which could not be removed, still there", coherenceName)
		}
	})
	t.Run("row6 .coherence.json, which cannot be removed, and .amendment.md, which can", func(t *testing.T) {
		t.Parallel()
		r := amendedRepo(t, iLimitThen)
		r.judgeAll(nil)
		r.writeNote("- refunds\n")
		prepare := lockCoherence(t, r)
		res := manifestIn(t, r.dir, r.dir, r.root(), prepare)
		if res.exit != 2 {
			t.Fatalf("want exit status 2, got %d\nstderr: %s", res.exit, res.stderr)
		}
		if r.hasNote() {
			t.Errorf("want %s removed, though %s could not be", noteName, coherenceName)
		}
		if !r.hasCoherence() {
			t.Errorf("want %s, which could not be removed, still there", coherenceName)
		}
	})
	t.Run("derived: .coherence.json, a folder holding a file", func(t *testing.T) {
		t.Parallel()
		// system.coherence-file: the entry, whatever it is.
		r := amendedRepo(t, iLimitThen)
		dir := filepath.Join(r.root(), coherenceName)
		if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "sub", "f.json"), []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		r.manifest()
		if r.hasCoherence() {
			t.Errorf("want the folder %s removed", coherenceName)
		}
	})
}

// lockCoherence makes the coherence file of r one that cannot be removed,
// and returns what manifestIn is to run around the command, if anything. On
// Windows the file is held open without sharing its deletion; elsewhere it
// is replaced by a folder that holds a file in a folder without write
// permission, which counts as the coherence file, whatever it is.
func lockCoherence(t *testing.T, r *repo) func() func() {
	t.Helper()
	if runtime.GOOS == "windows" {
		return func() func() {
			undo, err := makeCoherenceUnwritable(r.root())
			if err != nil {
				t.Fatal(err)
			}
			return undo
		}
	}
	c := filepath.Join(r.root(), coherenceName)
	os.Remove(c)
	locked := filepath.Join(c, "locked")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "f"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Chmod(locked, 0o555)
	t.Cleanup(func() { os.Chmod(locked, 0o755) })
	if err := os.Remove(filepath.Join(locked, "f")); err == nil {
		t.Skip("the folder stays writable without permission (running as root?)")
	}
	return nil
}
