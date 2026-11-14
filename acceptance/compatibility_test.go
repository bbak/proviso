package acceptance

// The invariant system.model-compatibility: a tree accepted under a model
// version from oldest_model_version to model_version, and the manifest
// written with it, read as the release that accepted it read them.
//
// testdata/model-<v>/spec/ holds, for each model version v, a tree accepted
// under it and its manifest.sum, frozen when v was model_version: the
// tree's files are copied into proviso/ as they are. The manifest of
// model-0 was computed by the manifest oracle (manifestFor) from the plain
// tree, whose parts' texts part_kinds fixes alone; it is never rewritten, so
// that a later change of the reading shows as a failure here.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// acceptedUnder reads the frozen tree accepted under the model version v,
// with paths below proviso/, and its manifest.
func acceptedUnder(t *testing.T, v string) (tree, string) {
	t.Helper()
	base := filepath.Join("testdata", "model-"+v, "spec")
	tr := tree{}
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(base, p)
		tr["proviso/"+filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatalf("system.model-compatibility: no tree accepted under model %s in %s: %v", v, base, err)
	}
	m, ok := tr[manifestPath]
	if !ok {
		t.Fatalf("system.model-compatibility: the tree accepted under model %s holds no manifest.sum", v)
	}
	if first, _, _ := strings.Cut(m, "\n"); first != "model "+v {
		t.Fatalf("system.model-compatibility: the manifest accepted under model %s starts with %q", v, first)
	}
	return tr, m
}

func TestSystemModelCompatibility(t *testing.T) {
	t.Parallel()
	for _, v := range readModelVersions() {
		t.Run("derived: model "+v, func(t *testing.T) {
			t.Parallel()
			t.Run("lint reports no error the release did not report", func(t *testing.T) {
				t.Parallel()
				// The release that accepted the tree reported no error in
				// it, or it would not have been accepted.
				tr, _ := acceptedUnder(t, v)
				o := check(t, tr)
				if o.json.exit != 0 {
					t.Errorf("system.model-compatibility: want lint to exit with 0 on the tree accepted under model %s, got %d\nstdout: %s\nstderr: %s", v, o.json.exit, o.json.stdout, o.json.stderr)
				}
			})
			t.Run("the same elements, parts and fingerprints", func(t *testing.T) {
				t.Parallel()
				tr, _ := acceptedUnder(t, v)
				r := newRepo(t, tr)
				r.commit("accepted under model " + v)
				d := r.diff()
				d.ran(t)
				if d.json.exit != 0 || len(d.ed.Elements) != 0 {
					t.Errorf("system.model-compatibility: want the tree accepted under model %s as its manifest records it, exit status 0 and no element, got %d:\n%s", v, d.json.exit, d.dump())
				}
			})
			t.Run("the next manifest names model_version, with the same parts", func(t *testing.T) {
				t.Parallel()
				tr, m := acceptedUnder(t, v)
				dir := tr.write(t)
				_, parts, _ := strings.Cut(m, "\n")
				if got, want := manifestOK(t, dir), "model "+modelVersion+"\n"+parts; got != want {
					t.Errorf("system.model-compatibility: want the manifest of the tree accepted under model %s rewritten as\n%s\ngot\n%s", v, want, got)
				}
			})
		})
	}
}
