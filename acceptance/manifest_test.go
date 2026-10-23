package acceptance

// Feature baseline.manifest: proviso/capabilities/baseline/features/manifest.md,
// and the contracts system.manifest-file (proviso/contracts/manifest.md) and
// system.cli.manifest (proviso/contracts/cli.md).

import (
	"os"
	"path/filepath"
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
	t.Run("row4 without write permission", func(t *testing.T) {
		t.Parallel()
		former := manifestFor(plainParts())
		dir := plainTree().edit(featurePath, "Buyers put goods", "Buyers put many goods").with(manifestPath, former).write(t)
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
