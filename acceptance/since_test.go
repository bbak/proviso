package acceptance

// The rule baseline.diff.since: `diff --since` and `pack --since`
// (system.cli.diff.since, system.cli.pack.since). Every case runs with both
// commands; packIn runs `diff` with the same option and checks the pack
// against it, so the cases assert on the element diff.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// sinceOutcome is one run of diff or pack with --since; d is the element
// diff, from diff itself or from the diff packIn ran beside the pack.
type sinceOutcome struct {
	d diffOutcome
	p *packOutcome
}

// since runs cmd, "diff" or "pack", with --since rev in the repository.
func since(t *testing.T, r *repo, cmd, rev string) sinceOutcome {
	t.Helper()
	if cmd == "pack" {
		p := packIn(t, r.dir, r.dir, r.root(), "--since", rev)
		return sinceOutcome{d: p.diff, p: &p}
	}
	return sinceOutcome{d: diffIn(t, r.dir, r.dir, r.root(), "--since", rev)}
}

// ran asserts that the command ran and returns the element diff.
func (o sinceOutcome) ran(t *testing.T) diffOutcome {
	t.Helper()
	if o.p != nil {
		o.p.ran(t)
	}
	o.d.ran(t)
	return o.d
}

func (o sinceOutcome) stopped(t *testing.T) {
	t.Helper()
	if o.p != nil {
		o.p.stopped(t)
		return
	}
	o.d.stopped(t)
}

// checkSinceRevision checks that the former text, where any was needed,
// came from the commit want, named in full. Where none was needed, the
// revision may be null or want (see the report of this round).
func checkSinceRevision(t *testing.T, o diffOutcome, want string, needed bool) {
	t.Helper()
	switch {
	case o.ed.Revision == nil && needed:
		t.Errorf("want the revision %s, got null", want)
	case o.ed.Revision != nil && *o.ed.Revision != want:
		t.Errorf("want the revision %s, named in full, got %s", want, *o.ed.Revision)
	}
}

// damageManifest puts a merge conflict into the manifest.sum in r's spec
// root, below its first line, so that lint still reads the model version
// (lint.run.model). It returns the manifest as it was.
func damageManifest(t *testing.T, r *repo) []byte {
	t.Helper()
	path := filepath.Join(r.root(), "manifest.sum")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	first, rest, _ := strings.Cut(string(b), "\n")
	damaged := first + "\n<<<<<<< HEAD\n" + rest + "=======\n>>>>>>> main\n"
	if err := os.WriteFile(path, []byte(damaged), 0o644); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBaselineDiffSince(t *testing.T) {
	t.Parallel()
	formerThen := "Then      the cart holds <quantity> more of <product>\n  And     nothing else changes"
	emptyText, _ := elementText(diffTree()[featurePath], "cart.add.empty")

	for _, cmd := range []string{"diff", "pack"} {
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			t.Run("row1 a1b2, the commit an agent received in its last round", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				r.set(diffTree().apply(xThen))
				a1b2 := r.commit("round 1")
				r.set(diffTree().apply(xThen, limitThen))
				// Any name git accepts: here an abbreviation.
				d := since(t, r, cmd, a1b2[:7]).ran(t)
				d.only(t, "cart.add.limit:changed")
				e := d.elem(t, "cart.add.limit")
				onlyParts(t, e, "step Then:changed")
				checkFormer(t, part(t, e, "step Then"), formerThen)
				checkSinceRevision(t, d, a1b2, true)
			})
			t.Run("row2 HEAD~1", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				r.set(diffTree().apply(limitThen))
				back := r.commit("round 1")
				r.set(diffTree().apply(limitThen, limitProvided))
				r.commit("round 2")
				d := since(t, r, cmd, "HEAD~1").ran(t)
				d.only(t, "cart.add.limit:changed")
				onlyParts(t, d.elem(t, "cart.add.limit"), "step Provided:changed")
				checkSinceRevision(t, d, back, true)
			})
			t.Run("row3 e5f6, from before the tree existed", func(t *testing.T) {
				t.Parallel()
				r := freshRepo(t, "repo", true)
				e5f6 := r.git("rev-parse", "HEAD")
				r.set(diffTree())
				r.manifest()
				r.commit("the tree, accepted")
				d := since(t, r, cmd, e5f6).ran(t)
				d.only(t, withChange(diffIDs, "added")...)
				checkSinceRevision(t, d, e5f6, false)
			})
			t.Run("row4 a1b2, with the manifest.sum in the spec root damaged by a merge conflict after its first line", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				r.set(diffTree().apply(xThen))
				a1b2 := r.commit("round 1")
				r.set(diffTree().apply(xThen, limitThen))
				// The conflict below the first line, so that lint still
				// reads the model version (lint.run.model).
				damageManifest(t, r)
				if plain := diffIn(t, r.dir, r.dir, r.root()); plain.json.exit != 2 {
					t.Fatalf("want the damage to stop diff without --since (baseline.diff.malformed), got exit status %d", plain.json.exit)
				}
				d := since(t, r, cmd, a1b2).ran(t)
				d.only(t, "cart.add.limit:changed")
				checkSinceRevision(t, d, a1b2, true)
			})
			t.Run("row5 a1b2, whose manifest.sum is damaged by a merge conflict after its first line", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				intact := damageManifest(t, r)
				r.set(diffTree().apply(xThen))
				if l := execLint(t, r.dir); l.exit != 0 {
					t.Fatalf("want lint to accept the tree with the damaged manifest, got exit status %d:\n%s%s", l.exit, l.stdout, l.stderr)
				}
				a1b2 := r.commit("round 1, the manifest damaged")
				// The manifest in the spec root intact again, so that only
				// the one at the revision is damaged.
				if err := os.WriteFile(filepath.Join(r.root(), "manifest.sum"), intact, 0o644); err != nil {
					t.Fatal(err)
				}
				r.set(diffTree().apply(xThen, limitThen))
				d := since(t, r, cmd, a1b2).ran(t)
				d.only(t, "cart.add.limit:changed")
				onlyParts(t, d.elem(t, "cart.add.limit"), "step Then:changed")
				checkSinceRevision(t, d, a1b2, true)
			})
			t.Run("row6 nosuch", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				r.set(diffTree().apply(limitThen))
				since(t, r, cmd, "nosuch").stopped(t)
			})
			t.Run("row7 c3d4, whose tree has a lint error", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				r.set(diffTree().with("proviso/notes.md", "# Notes\n"))
				c3d4 := r.commit("a tree with an error")
				r.set(diffTree().apply(limitThen))
				if l := execLint(t, r.dir); l.exit != 0 {
					t.Fatalf("want the tree as it stands now to pass lint, got exit status %d:\n%s", l.exit, l.stdout)
				}
				since(t, r, cmd, c3d4).stopped(t)
			})
			t.Run("row8 a1b2, whose manifest.sum names another model version", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				path := filepath.Join(r.root(), "manifest.sum")
				intact, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				_, rest, _ := strings.Cut(string(intact), "\n")
				if err := os.WriteFile(path, []byte("model 1\n"+rest), 0o644); err != nil {
					t.Fatal(err)
				}
				r.set(diffTree().apply(xThen))
				if l := execLint(t, r.dir); l.exit == 0 {
					t.Fatalf("want lint to refuse the tree whose manifest names model 1, got exit status 0")
				}
				a1b2 := r.commit("round 1, the manifest of another model version")
				// The manifest in the spec root of this model version again,
				// so that only the one at the revision names another.
				if err := os.WriteFile(path, intact, 0o644); err != nil {
					t.Fatal(err)
				}
				r.set(diffTree().apply(xThen, limitThen))
				if l := execLint(t, r.dir); l.exit != 0 {
					t.Fatalf("want the tree as it stands now to pass lint, got exit status %d:\n%s%s", l.exit, l.stdout, l.stderr)
				}
				since(t, r, cmd, a1b2).stopped(t)
			})
			t.Run("row9 a1b2, whose tree holds x.md as a symbolic link", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				// The link points out of the tree to a file holding x.md as
				// it was, committed beside it: followed, it would give a
				// tree lint accepts; skipped, it
				// would show x.md's elements removed. Either way the command
				// would run. The link goes into the commit through git's
				// index, so that no platform has to create one on disk.
				xMD := diffTree()[xPath]
				r.set(diffTree().without(xPath).with("outside/x.md", xMD))
				r.git("add", "-A")
				target := filepath.Join(t.TempDir(), "target")
				if err := os.WriteFile(target, []byte("../../../../outside/x.md"), 0o644); err != nil {
					t.Fatal(err)
				}
				blob := r.git("hash-object", "-w", target)
				r.git("update-index", "--add", "--cacheinfo", "120000,"+blob+","+xPath)
				r.git("commit", "-q", "-m", "round 1, x.md a link")
				a1b2 := r.git("rev-parse", "HEAD")
				if mode := strings.Fields(r.git("ls-tree", a1b2, "--", xPath)); len(mode) == 0 || mode[0] != "120000" {
					t.Fatalf("want x.md a symbolic link at %s, got %v", a1b2, mode)
				}
				r.set(diffTree().apply(limitThen))
				if l := execLint(t, r.dir); l.exit != 0 {
					t.Fatalf("want the tree as it stands now to pass lint, got exit status %d:\n%s%s", l.exit, l.stdout, l.stderr)
				}
				since(t, r, cmd, a1b2).stopped(t)
			})
			t.Run("derived: the manifest written but never committed", func(t *testing.T) {
				t.Parallel()
				// Without --since, this stops (baseline.diff.former).
				r := newRepo(t, diffTree())
				base := r.commit("tree")
				r.manifest()
				r.set(diffTree().apply(limitThen))
				d := since(t, r, cmd, base).ran(t)
				d.only(t, "cart.add.limit:changed")
				checkSinceRevision(t, d, base, true)
			})
			t.Run("derived: no manifest.sum", func(t *testing.T) {
				t.Parallel()
				// Without --since, every element is added
				// (baseline.diff.baseline).
				r := newRepo(t, diffTree())
				base := r.commit("tree")
				r.set(diffTree().apply(limitThen))
				d := since(t, r, cmd, base).ran(t)
				d.only(t, "cart.add.limit:changed")
			})
			t.Run("derived: an element removed since, with its former text", func(t *testing.T) {
				t.Parallel()
				r, _ := accepted(t, diffTree())
				r.set(diffTree().apply(limitThen))
				a1b2 := r.commit("round 1")
				r.set(diffTree().apply(limitThen).edit(featurePath, emptyText, ""))
				d := since(t, r, cmd, a1b2).ran(t)
				d.only(t, "cart.add.empty:removed")
				if e := d.elem(t, "cart.add.empty"); e.Former == nil || normWS(*e.Former) != normWS(emptyText) {
					t.Errorf("want the former text of cart.add.empty\n%s\ngot\n%s", emptyText, deref(e.Former))
				}
				checkSinceRevision(t, d, a1b2, true)
			})
			t.Run("derived: no git repository", func(t *testing.T) {
				t.Parallel()
				dir := diffTree().write(t)
				skipIfGitAbove(t, dir)
				r := &repo{t: t, dir: dir, tr: diffTree()}
				since(t, r, cmd, "HEAD").stopped(t)
			})
			// baseline.run.lint, changed through the option: the tree as
			// it stands now is checked first.
			t.Run("derived: the tree as it stands now with a lint error", func(t *testing.T) {
				t.Parallel()
				r, base := accepted(t, diffTree())
				r.set(diffTree().with("proviso/notes.md", "# Notes\n"))
				lint := lintLines(t, r.dir)
				o := since(t, r, cmd, base)
				o.stopped(t)
				stderr := o.d.json.stderr
				if o.p != nil {
					stderr = o.p.json.stderr
				}
				for _, l := range lint {
					if !strings.Contains(stderr, l) {
						t.Errorf("want stderr to hold lint's line %q, got: %s", l, stderr)
					}
				}
			})
		})
	}

	// The revision the manifest's tree was committed in stands in for the
	// baseline exactly: every rule of the element diff that is changed
	// through the option gives the same list either way.
	amendments := []struct {
		name string
		base tree
		cs   []change
	}{
		{"derived: since the accepted commit, as without --since: steps, rows and a rationale", diffTree(), []change{limitThen, limitNewRow, limitNoRatio}},
		{"derived: since the accepted commit, as without --since: through a dependency", handTree(), []change{itemDefinition, limitThen}},
		{"derived: since the accepted commit, as without --since: added, removed and reworded", diffTree(), []change{{featurePath, emptyText, ""}, xThen}},
	}
	for _, a := range amendments {
		t.Run(a.name, func(t *testing.T) {
			t.Parallel()
			r, base := accepted(t, a.base)
			tr := amended(a.base, a.cs...).appendTo(featurePath, extraRule)
			r.set(tr)
			plain := r.diff()
			plain.ran(t)
			d := diffIn(t, r.dir, r.dir, r.root(), "--since", base)
			d.ran(t)
			var pe, se map[string]any
			if err := json.Unmarshal([]byte(plain.json.stdout), &pe); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(d.json.stdout), &se); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(pe, se) {
				t.Errorf("want the same element_diff with --since %s as without, got\n%s\n---\n%s", base, plain.dump(), d.dump())
			}
			p := packIn(t, r.dir, r.dir, r.root(), "--since", base)
			p.ran(t)
			if pp := r.pack(); pp.json.stdout != p.json.stdout || pp.text.stdout != p.text.stdout {
				t.Errorf("want the same pack with --since %s as without", base)
			}
		})
	}
}
