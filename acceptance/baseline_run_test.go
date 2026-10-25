package acceptance

// Feature baseline.run: proviso/capabilities/baseline/features/run.md

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// baseRun runs the baseline command cmd in cwd: `manifest`, `judge` with a
// check result of no judgements and no findings, or `diff`, `impact`, `pack`
// or `verify` with --json (diffIn, impactIn, packIn and verifyIn run them
// without as well). root
// is the spec root the command is expected to work on, if any.
func baseRun(t *testing.T, cmd, base, cwd, root string) run {
	t.Helper()
	switch cmd {
	case "manifest":
		return manifestIn(t, base, cwd, root, nil)
	case "judge":
		return judgeIn(t, base, cwd, root, "{}")
	case "verify":
		return verifyIn(t, base, cwd, root).json
	case "impact":
		return impactIn(t, base, cwd, root).json
	case "pack":
		return packIn(t, base, cwd, root).json
	}
	return diffIn(t, base, cwd, root).json
}

// lintLines is what lint writes as text in cwd, line by line.
func lintLines(t *testing.T, cwd string) []string {
	t.Helper()
	var lines []string
	for _, l := range strings.Split(strings.ReplaceAll(execLint(t, cwd).stdout, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func TestBaselineRunLint(t *testing.T) {
	t.Parallel()
	invalid := "# <a id=\"cart.bad\"></a> Bad \xff\xfe\n\n- Contact: Dana Reyes\n"
	for _, cmd := range []string{"manifest", "diff", "impact", "pack", "judge", "verify"} {
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			t.Run("row1 a tree without findings", func(t *testing.T) {
				t.Parallel()
				dir := plainTree().write(t)
				r := baseRun(t, cmd, dir, dir, filepath.Join(dir, "proviso"))
				if r.exit == 2 {
					t.Errorf("want the command to run, got exit status 2\nstderr: %s", r.stderr)
				}
			})
			t.Run("row2 a tree with an ask only", func(t *testing.T) {
				t.Parallel()
				dir := plainTree().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?").write(t)
				lint := lintLines(t, dir)
				if len(lint) == 0 {
					t.Fatal("want lint to report the ask")
				}
				r := baseRun(t, cmd, dir, dir, filepath.Join(dir, "proviso"))
				if r.exit == 2 {
					t.Fatalf("want the command to run, got exit status 2\nstderr: %s", r.stderr)
				}
				// Nothing of what lint would write is written.
				for _, l := range lint {
					if strings.Contains(r.stdout, l) || strings.Contains(r.stderr, l) {
						t.Errorf("want nothing of lint's output written, got %q in\nstdout: %s\nstderr: %s", l, r.stdout, r.stderr)
					}
				}
			})
			t.Run("row3 a tree with an error", func(t *testing.T) {
				t.Parallel()
				dir := plainTree().with("proviso/notes.md", "# Notes\n").write(t)
				lint := lintLines(t, dir)
				r := baseRun(t, cmd, dir, dir, filepath.Join(dir, "proviso"))
				if r.exit != 2 {
					t.Fatalf("want exit status 2, got %d", r.exit)
				}
				// What lint would write as text goes to stderr; that
				// nothing goes to stdout is checked for every run.
				for _, l := range lint {
					if !strings.Contains(r.stderr, l) {
						t.Errorf("want stderr to hold lint's line %q, got: %s", l, r.stderr)
					}
				}
			})
			t.Run("row4 a tree with an unreadable file", func(t *testing.T) {
				t.Parallel()
				dir := plainTree().with("proviso/capabilities/cart/features/a.md", invalid).write(t)
				if r := baseRun(t, cmd, dir, dir, filepath.Join(dir, "proviso")); r.exit != 2 {
					t.Errorf("want exit status 2, got %d", r.exit)
				}
			})
			t.Run("row5 no spec root up to the file system's root", func(t *testing.T) {
				t.Parallel()
				dir := tree{"other/readme.txt": "nothing here\n"}.write(t)
				skipIfSpecRootAbove(t, dir)
				if r := baseRun(t, cmd, dir, filepath.Join(dir, "other"), ""); r.exit != 2 {
					t.Errorf("want exit status 2, got %d", r.exit)
				}
			})
			// lint.run.model stops lint, and with it the command.
			t.Run("derived: a manifest of another model version", func(t *testing.T) {
				t.Parallel()
				dir := plainTree().with(manifestPath, "model 1\n").write(t)
				if r := baseRun(t, cmd, dir, dir, filepath.Join(dir, "proviso")); r.exit != 2 {
					t.Errorf("want exit status 2, got %d", r.exit)
				}
			})
			t.Run("derived: run inside the tree", func(t *testing.T) {
				t.Parallel()
				dir := plainTree().write(t)
				root := filepath.Join(dir, "proviso")
				if r := baseRun(t, cmd, dir, filepath.Join(root, "capabilities", "cart", "features"), root); r.exit == 2 {
					t.Errorf("want the command to run, got exit status 2\nstderr: %s", r.stderr)
				}
			})
			t.Run("derived: the nearest spec root", func(t *testing.T) {
				t.Parallel()
				// The outer tree has an error, the inner none.
				tr := plainTree().with("proviso/notes.md", "# Notes\n")
				for k, v := range plainTree().prefixed("a") {
					tr[k] = v
				}
				dir := tr.write(t)
				inner := filepath.Join(dir, "a", "proviso")
				if r := baseRun(t, cmd, dir, filepath.Join(dir, "a"), inner); r.exit == 2 {
					t.Errorf("want the command to run on the inner tree, got exit status 2\nstderr: %s", r.stderr)
				}
			})
		})
	}
}

// TestBaselineRunWrites checks the invariant baseline.run.writes over runs
// of every kind, however they end: every file and folder below the test
// directory, the git repository's .git included, is as it was, except
// manifest.sum, .coherence.json, and .amendment.md with everything in it,
// after a run of `manifest` that wrote manifest.sum, and .coherence.json
// after a run of `judge` that exits with 0. Every
// other test checks it too, for each run it makes (see manifestIn, diffIn,
// impactIn, packIn, judgeIn and verifyIn).
func TestBaselineRunWrites(t *testing.T) {
	t.Parallel()
	// staged leaves an edit in the index and an untracked file, which a
	// run must leave as they are.
	staged := func(r *repo) {
		r.set(r.tr.apply(change{featurePath, "Buyers put goods into the cart.", "Buyers put goods in the cart."}))
		r.git("add", filepath.ToSlash(featurePath))
		os.WriteFile(filepath.Join(r.dir, "untracked.txt"), []byte("untracked\n"), 0o644)
	}
	cases := []struct {
		name  string
		cmd   string
		setup func(t *testing.T) *repo
	}{
		{"manifest, written", "manifest", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			staged(r)
			return r
		}},
		{"manifest, a tree with an error", "manifest", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
			return r
		}},
		{"manifest, a manifest of another model version", "manifest", func(t *testing.T) *repo {
			return newRepo(t, diffTree().with(manifestPath, "model 1\n"))
		}},
		{"manifest, written, with a coherence file", "manifest", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{})
			staged(r)
			return r
		}},
		{"manifest, a tree with an error, with a coherence file", "manifest", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{})
			r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
			return r
		}},
		{"manifest, written, with a coherence file and an amendment note", "manifest", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{})
			r.writeNote("- refunds\n")
			staged(r)
			return r
		}},
		{"manifest, a tree with an error, with an amendment note", "manifest", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.writeNote("- refunds\n")
			r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
			return r
		}},
		{"diff, nothing changed, with an amendment note", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.writeNote("- refunds\n")
			return r
		}},
		{"diff, changes with a coherence file and an amendment note", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{})
			r.writeNote("- refunds\n")
			staged(r)
			return r
		}},
		{"diff, nothing changed", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			return r
		}},
		{"diff, changes with former text, staged and untracked", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			staged(r)
			return r
		}},
		{"diff, a merge in progress", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.git("checkout", "-q", "-b", "feature")
			r.set(diffTree().apply(xThen))
			r.commit("feature")
			r.git("checkout", "-q", "main")
			r.tr = diffTree()
			r.set(diffTree().apply(limitThen))
			r.commit("mainline")
			r.git("checkout", "-q", "feature")
			r.tr = diffTree().apply(xThen)
			r.git("merge", "-q", "--no-ff", "--no-commit", "main")
			r.tr = diffTree().apply(xThen, limitThen)
			return r
		}},
		{"diff, a malformed manifest", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			b, _ := os.ReadFile(filepath.Join(r.dir, filepath.FromSlash(manifestPath)))
			os.WriteFile(filepath.Join(r.dir, filepath.FromSlash(manifestPath)), append(b, "<<<<<<< HEAD\n"...), 0o644)
			return r
		}},
		{"diff, a manifest written again and not committed", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.manifest()
			r.set(r.tr.apply(limitProvided))
			return r
		}},
		{"diff, without a manifest", "diff", func(t *testing.T) *repo {
			return newRepo(t, diffTree())
		}},
		{"diff, changes with a coherence file", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
			staged(r)
			return r
		}},
		{"diff, a damaged coherence file", "diff", func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			os.WriteFile(filepath.Join(r.root(), coherenceName), []byte("<<<<<<< HEAD\n"), 0o644)
			return r
		}},
	}
	for _, c := range cases {
		// Every case of diff is a case of impact, pack, judge and verify
		// as well.
		cmds := []string{c.cmd}
		if c.cmd == "diff" {
			cmds = append(cmds, "impact", "pack", "judge", "verify")
		}
		for _, cmd := range cmds {
			name := c.name
			if cmd != c.cmd {
				name = cmd + strings.TrimPrefix(c.name, "diff")
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				r := c.setup(t)
				if _, ok := snapshotAll(t, r.dir)[".git/HEAD"]; !ok {
					t.Fatal("the snapshot does not cover the git repository")
				}
				baseRun(t, cmd, r.dir, r.dir, r.root())
			})
		}
	}
}
