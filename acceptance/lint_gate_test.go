package acceptance

// The system invariant system.lint-gate: the tree commands run only where
// lint finds no error (proviso/system.md). It ranges over every run of
// `manifest`, `diff`, `pack`, `impact`, `judge`, `verify` and `rehome`, with
// or without --since; each case below runs with each of them.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

// gateCommands are the runs system.lint-gate ranges over, each with
// arguments it takes. "pack --element" names the system; "judge, not JSON"
// gives judge input that is no check_result, which lint's gate must stop
// before judge reads it (coherence.judge.input); "rehome --write" is the run
// that would rewrite the tree.
var gateCommands = []string{
	"manifest", "diff", "pack", "pack --element", "impact", "judge", "judge, not JSON",
	"verify", "rehome", "rehome --write",
}

// gateRun runs the command cmd of gateCommands in cwd, through the driver
// that checks what holds for every run of it (see baseRun).
func gateRun(t *testing.T, cmd, base, cwd, root string) run {
	t.Helper()
	switch cmd {
	case "judge, not JSON":
		return judgeIn(t, base, cwd, root, "this is no JSON")
	case "rehome":
		return rehomeIn(t, base, cwd, root, nil, "--json").run
	case "rehome --write":
		return rehomeIn(t, base, cwd, root, nil, "--write", "--json").run
	}
	return baseRun(t, cmd, base, cwd, root)
}

// underRoot is the part of the snapshot snap below the spec root rel, the
// whole snapshot where rel is empty.
func underRoot(snap map[string]string, rel string) map[string]string {
	out := map[string]string{}
	for p, v := range snap {
		if rel == "" || rel == "." || p == rel || strings.HasPrefix(p, rel+"/") {
			out[p] = v
		}
	}
	return out
}

// gateCase is one directory a run starts in: base is the folder the test
// set up, cwd the directory the command runs in, root the spec root lint
// would check, if any. accepts says whether lint, run in cwd, exits with 0.
type gateCase struct {
	name    string
	accepts bool
	setup   func(t *testing.T) (base, cwd, root string)
}

// acceptedCase is the tree tr accepted, then changed to now, if not nil.
func acceptedCase(tr tree, now func() tree) func(t *testing.T) (string, string, string) {
	return func(t *testing.T) (string, string, string) {
		r, _ := accepted(t, tr)
		if now != nil {
			r.set(now())
		}
		return r.dir, r.dir, r.root()
	}
}

func TestSystemLintGate(t *testing.T) {
	t.Parallel()
	invalid := "# <a id=\"cart.bad\"></a> Bad \xff\xfe\n\n- Contact: Dana Reyes\n"
	ask := func() tree { return plainTree().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?") }
	place := func() tree { return plainTree().with(placeAPIPath, placeAPIMD) }
	// Every tree is accepted first, so that it is as its manifest records
	// it, and rehome is stopped by lint alone where it is stopped.
	cases := []gateCase{
		{"a tree without findings", true, acceptedCase(plainTree(), nil)},
		{"a tree with an ask only", true, acceptedCase(ask(), nil)},
		// okf-get: the contract of the cart in placeAPIMD, whose
		// identifiers name no place: an ask, which rehome repairs.
		{"a tree with the contract okf-get in the capability access", true, acceptedCase(place(), nil)},
		// notes.md is no file of the layout and no element: lint reports
		// an error, while the tree is still as the manifest records it.
		{"a tree with an error", false, acceptedCase(plainTree(), func() tree {
			return plainTree().with("proviso/notes.md", "# Notes\n")
		})},
		{"a tree with an unreadable file", false, acceptedCase(plainTree(), func() tree {
			return plainTree().with("proviso/capabilities/cart/features/a.md", invalid)
		})},
		{"no spec root up to the file system's root", false, func(t *testing.T) (string, string, string) {
			dir := tree{"other/readme.txt": "nothing here\n"}.write(t)
			skipIfSpecRootAbove(t, dir)
			return dir, filepath.Join(dir, "other"), ""
		}},
		{"a manifest.sum naming a model version lint does not read", false, func(t *testing.T) (string, string, string) {
			r, _ := accepted(t, plainTree())
			if err := os.WriteFile(filepath.Join(r.root(), "manifest.sum"), []byte("model 2\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return r.dir, r.dir, r.root()
		}},
		{"run inside the tree", true, func(t *testing.T) (string, string, string) {
			r, _ := accepted(t, plainTree())
			return r.dir, filepath.Join(r.root(), "capabilities", "cart", "features"), r.root()
		}},
		{"the nearest spec root, the outer tree with an error", true, func(t *testing.T) (string, string, string) {
			tr := plainTree().with("proviso/notes.md", "# Notes\n")
			for k, v := range plainTree().prefixed("a") {
				tr[k] = v
			}
			r := newRepo(t, tr)
			manifestOK(t, filepath.Join(r.dir, "a"))
			r.commit("accept")
			return r.dir, filepath.Join(r.dir, "a"), filepath.Join(r.dir, "a", "proviso")
		}},
		{"the nearest spec root, the inner tree with an error", false, func(t *testing.T) (string, string, string) {
			tr := plainTree()
			for k, v := range plainTree().prefixed("a") {
				tr[k] = v
			}
			r := newRepo(t, tr)
			manifestOK(t, r.dir)
			manifestOK(t, filepath.Join(r.dir, "a"))
			r.commit("accept")
			r.set(tr.with("a/proviso/notes.md", "# Notes\n"))
			return r.dir, filepath.Join(r.dir, "a"), filepath.Join(r.dir, "a", "proviso")
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, cmd := range gateCommands {
				t.Run(cmd, func(t *testing.T) {
					t.Parallel()
					base, cwd, root := c.setup(t)
					lint := execLint(t, cwd)
					if (lint.exit == 0) != c.accepts {
						t.Fatalf("test tree: want lint to accept it: %v, got exit status %d\n%s%s", c.accepts, lint.exit, lint.stdout, lint.stderr)
					}
					lines := lintLines(t, cwd)
					rootRel := ""
					if root != "" {
						rel, _ := filepath.Rel(base, root)
						rootRel = filepath.ToSlash(rel)
					}
					before := underRoot(snapshotAll(t, base), rootRel)
					r := gateRun(t, cmd, base, cwd, root)
					checkGate(t, c.accepts, lines, r)
					if !c.accepts {
						checkWrites(t, "system.lint-gate", before, underRoot(snapshotAll(t, base), rootRel))
					}
				})
			}
		})
	}

	// diff and pack with --since: the tree as it stands now is checked
	// first, also where the revision holds a tree lint accepts.
	sinceCases := []struct {
		name    string
		accepts bool
		now     func(t *testing.T, r *repo)
	}{
		{"with --since HEAD~1, a tree lint accepts", true, func(t *testing.T, r *repo) {}},
		{"with --since HEAD~1, a tree with an error", false, func(t *testing.T, r *repo) {
			r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
		}},
		{"with --since HEAD~1, a manifest.sum naming a model version lint does not read", false, func(t *testing.T, r *repo) {
			if err := os.WriteFile(filepath.Join(r.root(), "manifest.sum"), []byte("model 2\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, c := range sinceCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, cmd := range []string{"diff", "pack"} {
				t.Run(cmd, func(t *testing.T) {
					t.Parallel()
					r, _ := accepted(t, diffTree())
					r.set(diffTree().apply(limitThen))
					r.commit("round 1")
					c.now(t, r)
					lines := lintLines(t, r.dir)
					before := underRoot(snapshotAll(t, r.dir), "proviso")
					o := since(t, r, cmd, "HEAD~1")
					res := o.d.json
					if o.p != nil {
						res = o.p.json
					}
					checkGate(t, c.accepts, lines, res)
					if !c.accepts {
						o.stopped(t)
						checkWrites(t, "system.lint-gate", before, underRoot(snapshotAll(t, r.dir), "proviso"))
					}
				})
			}
		})
	}
}

// checkGate checks the run r against system.lint-gate, given whether lint
// accepts the tree and the lines lint writes as text: where it does, the
// command runs and writes nothing of lint's; otherwise lint's lines go to
// stderr, the exit status is 2 and nothing goes to stdout.
func checkGate(t *testing.T, accepts bool, lint []string, r run) {
	t.Helper()
	if accepts {
		if r.exit == 2 {
			t.Fatalf("system.lint-gate: want the command to run, got exit status 2\nstderr: %s", r.stderr)
		}
		for _, l := range lint {
			if strings.Contains(r.stdout, l) || strings.Contains(r.stderr, l) {
				t.Errorf("system.lint-gate: want nothing of lint's output written, got %q in\nstdout: %s\nstderr: %s", l, r.stdout, r.stderr)
			}
		}
		return
	}
	if r.exit != 2 {
		t.Errorf("system.lint-gate: want exit status 2, got %d\nstderr: %s", r.exit, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("system.lint-gate: want nothing on stdout, got: %q", r.stdout)
	}
	for _, l := range lint {
		if !strings.Contains(r.stderr, l) {
			t.Errorf("system.lint-gate: want stderr to hold lint's line %q, got: %s", l, r.stderr)
		}
	}
}
