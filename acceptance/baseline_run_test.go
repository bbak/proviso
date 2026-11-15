package acceptance

// The invariant baseline.run.writes, over the runs of the baseline commands.
// What lint's gate decides for them is checked by TestSystemLintGate.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// baseRun runs the baseline or coherence command cmd in cwd: `manifest`,
// `judge` with a check result of no judgements and no findings, or `diff`,
// `impact`, `pack` or `verify` with --json (diffIn, impactIn, packIn and
// verifyIn run them without as well), or "pack --element", which names the
// system (packElementIn runs it without --json as well). root
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
	case "pack --element":
		return packElementIn(t, base, cwd, root, "system").json
	}
	return diffIn(t, base, cwd, root).json
}

// TestBaselineRunWrites checks the invariant baseline.run.writes over runs
// of `manifest`, `diff` and `pack`, however they end: every file and folder
// below the test directory, the git repository's .git included, is as it
// was, except manifest.sum, .coherence.json, and .amendment.md with
// everything in it, after a run of `manifest` that wrote manifest.sum. Every
// other test checks it too, for each run it makes (see manifestIn, diffIn
// and packIn).
func TestBaselineRunWrites(t *testing.T) {
	t.Parallel()
	// Every case of diff is a case of pack as well.
	writesOver(t, "manifest", "diff", "pack")
}

// writesOver runs the commands cmds in the cases that check the invariants
// baseline.run.writes and coherence.run.writes. A case of `manifest` runs
// only where manifest is among cmds; every case of `diff` runs every other
// command of cmds. The runs check the invariant themselves (see baseRun).
func writesOver(t *testing.T, cmds ...string) {
	t.Helper()
	// staged leaves an edit in the index and an untracked file, which a
	// run must leave as they are.
	staged := func(r *repo) {
		r.set(r.tr.apply(change{featurePath, "Buyers put goods into the cart.", "Buyers put goods in the cart."}))
		r.git("add", filepath.ToSlash(featurePath))
		if err := os.WriteFile(filepath.Join(r.dir, "untracked.txt"), []byte("untracked\n"), 0o644); err != nil {
			t.Fatal(err)
		}
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
			return newRepo(t, diffTree().with(manifestPath, "model 2\n"))
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
			if err := os.WriteFile(filepath.Join(r.dir, filepath.FromSlash(manifestPath)), append(b, "<<<<<<< HEAD\n"...), 0o644); err != nil {
				t.Fatal(err)
			}
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
			if err := os.WriteFile(filepath.Join(r.root(), coherenceName), []byte("<<<<<<< HEAD\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return r
		}},
	}
	for _, c := range cases {
		for _, cmd := range cmds {
			if (c.cmd == "manifest") != (cmd == "manifest") {
				continue
			}
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
