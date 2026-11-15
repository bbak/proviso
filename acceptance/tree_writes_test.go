package acceptance

// The invariant system.tree-writes and the parameter command_writes
// (system.command-writes) it names: every run of a command command_writes
// lists leaves the spec root as it was, but for what command_writes lists for
// that command, and the git repository holding the spec root as it was.
// Every run goes through the command's driver, which checks the same for the
// directory the test set up (see checkWrites); the runs here check it again
// against the oracle mayChange, which reads command_writes row by row, over
// the spec root and the git repository's .git folder only, as the invariant
// says. What lint's gate decides for the tree commands is checked by
// TestSystemLintGate.

import (
	"os"
	"path/filepath"
	"slices"
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

// runCommand runs the command cmd with args in base through the command's
// driver and gives its exit status. root is the spec root the command is
// expected to work on, if any; input, if not empty, goes to standard input.
// `lint` runs with --json, as lintIn does; `version` is a build of the
// Proviso version dev.
func runCommand(t *testing.T, cmd, base, root, input string, args ...string) int {
	t.Helper()
	switch cmd {
	case "help":
		return provisoIn(t, base, base, input, append([]string{"help"}, args...)...).exit
	case "version":
		return versionIn(t, "dev", base, base, nil, args...).exit
	case "lint":
		return lintIn(t, base, base, nil).json.exit
	case "show":
		return showIn(t, base, base, args...).exit
	case "judge":
		if input == "" {
			input = "{}"
		}
		return judgeIn(t, base, base, root, input).exit
	case "rehome":
		return rehomeWith(t, base, base, root, input, nil, args...).exit
	case "pack":
		if len(args) == 2 && args[0] == "--element" {
			return packElementIn(t, base, base, root, args[1]).json.exit
		}
	}
	return baseRun(t, cmd, base, base, root).exit
}

// mayChange is the oracle of command_writes: whether the run of cmd with
// args, which exited with exit, may leave the entry p under the spec root
// rootRel as after holds it, rather than as before held it.
func mayChange(cmd string, args []string, exit int, rootRel string, before, after map[string]string, p string) bool {
	if rootRel == "" {
		return false
	}
	in := func(name string) bool {
		q := rootRel + "/" + name
		return p == q || strings.HasPrefix(p, q+"/")
	}
	switch cmd {
	case "manifest":
		// manifest.sum, and .coherence.json and .amendment.md, each with
		// everything in it, after a run that wrote manifest.sum. A run that
		// writes the same bytes again and then fails cannot be told from one
		// that wrote nothing.
		m := rootRel + "/manifest.sum"
		_, ok := after[m]
		wrote := exit == 0 || (ok && after[m] != before[m])
		return wrote && (in("manifest.sum") || in(coherenceName) || in(noteName))
	case "judge":
		// .coherence.json, after a run that exits with 0.
		return exit == 0 && p == rootRel+"/"+coherenceName
	case "rehome":
		// After a run with --write that exits with 0: manifest.sum, and the
		// places of the spec tree its rules rename, nothing else of their
		// files. A rename changes a file's text, never which entries there
		// are; which text may change, the driver checks (checkRenamedOnly).
		if exit != 0 || !slices.Contains(args, "--write") {
			return false
		}
		if p == rootRel+"/manifest.sum" {
			return true
		}
		v, okB := before[p]
		w, okA := after[p]
		return okB && okA && strings.HasPrefix(v, "file: ") && strings.HasPrefix(w, "file: ") && strings.HasPrefix(p, rootRel+"/")
	}
	// help, version, lint, show, diff, pack, impact and verify: nothing.
	return false
}

// checkTreeWrites checks system.tree-writes for one run of cmd with args,
// which exited with exit: of the entries before and after hold, those under
// the spec root rootRel, if not empty, and those of the git repository's .git
// folder at the top of the snapshot are as they were, but what mayChange
// allows.
func checkTreeWrites(t *testing.T, cmd string, args []string, exit int, rootRel string, before, after map[string]string) {
	t.Helper()
	covered := func(p string) bool {
		if p == ".git" || strings.HasPrefix(p, ".git/") {
			return true
		}
		return rootRel != "" && (p == rootRel || strings.HasPrefix(p, rootRel+"/"))
	}
	isGit := func(p string) bool { return p == ".git" || strings.HasPrefix(p, ".git/") }
	report := func(p, what string) {
		if isGit(p) {
			t.Errorf("system.tree-writes: %s of the git repository %s after `%s %s`", p, what, cmd, strings.Join(args, " "))
		} else {
			t.Errorf("system.tree-writes: %s %s after `%s %s`, which command_writes does not allow", p, what, cmd, strings.Join(args, " "))
		}
	}
	for p, v := range before {
		if !covered(p) || (!isGit(p) && mayChange(cmd, args, exit, rootRel, before, after, p)) {
			continue
		}
		if w, ok := after[p]; !ok {
			report(p, "was removed")
		} else if v != w {
			report(p, "changed")
		}
	}
	for p := range after {
		if _, ok := before[p]; ok || !covered(p) || (!isGit(p) && mayChange(cmd, args, exit, rootRel, before, after, p)) {
			continue
		}
		report(p, "was created")
	}
}

// writesIn runs cmd with args in base, through runCommand, and checks
// system.tree-writes for the run against command_writes. root is the spec
// root the command is expected to work on, if any. It gives the exit status.
func writesIn(t *testing.T, cmd, base, root, input string, args ...string) int {
	t.Helper()
	before := snapshotAll(t, base)
	exit := runCommand(t, cmd, base, root, input, args...)
	after := snapshotAll(t, base)
	rootRel := ""
	if root != "" {
		rel, _ := filepath.Rel(base, root)
		rootRel = filepath.ToSlash(rel)
	}
	checkTreeWrites(t, cmd, args, exit, rootRel, before, after)
	return exit
}

// staged leaves an edit of the spec tree in the index and an untracked file,
// which a run must leave as they are. The repository holds diffTree.
func staged(t *testing.T, r *repo) {
	t.Helper()
	r.set(r.tr.apply(change{featurePath, "Buyers put goods into the cart.", "Buyers put goods in the cart."}))
	r.git("add", filepath.ToSlash(featurePath))
	untracked(t, r)
}

// untracked leaves an untracked file beside the spec root.
func untracked(t *testing.T, r *repo) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(r.dir, "untracked.txt"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// pending is an accepted diffTree with a change of the spec tree in the
// index and an untracked file beside it.
func pending(t *testing.T) *repo {
	t.Helper()
	r, _ := accepted(t, diffTree())
	r.set(r.tr.apply(limitThen))
	r.git("add", filepath.ToSlash(featurePath))
	untracked(t, r)
	return r
}

// mergeInProgress is an accepted diffTree with a merge in progress, its
// changes in the index and no commit made.
func mergeInProgress(t *testing.T) *repo {
	t.Helper()
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
}

// writesCase is one run of a command: setup gives the directory the test set
// up and the spec root the command is expected to work on, if any.
type writesCase struct {
	name  string
	setup func(t *testing.T) (base, root string)
	input string
	args  []string
}

// inRepo is a setup of the repository setup gives.
func inRepo(setup func(t *testing.T) *repo) func(t *testing.T) (string, string) {
	return func(t *testing.T) (string, string) {
		r := setup(t)
		return r.dir, r.root()
	}
}

// inDir is a setup of tr written into a directory in no git repository.
func inDir(tr tree) func(t *testing.T) (string, string) {
	return func(t *testing.T) (string, string) {
		dir := tr.write(t)
		return dir, filepath.Join(dir, "proviso")
	}
}

// runCases runs every case of cmd, each in a subtest of its own.
func runCases(t *testing.T, cmd string, cases []writesCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			base, root := c.setup(t)
			writesIn(t, cmd, base, root, c.input, c.args...)
		})
	}
}

// treeCases are the cases of diff, which every command that reads the
// element diff runs as well.
func treeCases() []writesCase {
	return []writesCase{
		{name: "nothing changed", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			return r
		})},
		{name: "nothing changed, with an amendment note", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.writeNote("- refunds\n")
			return r
		})},
		{name: "changes with former text, staged and untracked", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			staged(t, r)
			return r
		})},
		{name: "changes with a coherence file and an amendment note", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{})
			r.writeNote("- refunds\n")
			staged(t, r)
			return r
		})},
		{name: "changes with a coherence file", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.judgeOK(checkResult{Findings: []findingIn{gap("A gap.")}})
			staged(t, r)
			return r
		})},
		{name: "a merge in progress", setup: inRepo(mergeInProgress)},
		{name: "a malformed manifest", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			p := filepath.Join(r.dir, filepath.FromSlash(manifestPath))
			b, _ := os.ReadFile(p)
			if err := os.WriteFile(p, append(b, "<<<<<<< HEAD\n"...), 0o644); err != nil {
				t.Fatal(err)
			}
			return r
		})},
		{name: "a manifest written again and not committed", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.manifest()
			r.set(r.tr.apply(limitProvided))
			return r
		})},
		{name: "without a manifest", setup: inRepo(func(t *testing.T) *repo {
			return newRepo(t, diffTree())
		})},
		{name: "a damaged coherence file", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			if err := os.WriteFile(filepath.Join(r.root(), coherenceName), []byte("<<<<<<< HEAD\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return r
		})},
		{name: "a tree with an error", setup: inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen))
			r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
			return r
		})},
	}
}

// withArgs gives the cases cs, each run with args and input.
func withArgs(cs []writesCase, input string, args ...string) []writesCase {
	out := make([]writesCase, len(cs))
	for i, c := range cs {
		c.input, c.args = input, args
		out[i] = c
	}
	return out
}

// TestSystemCommandWrites checks command_writes row by row: the runs of each
// command, ending every way, change under the spec root nothing but what its
// row lists, and nothing of the git repository.
func TestSystemCommandWrites(t *testing.T) {
	t.Parallel()
	t.Run("row1 help", func(t *testing.T) {
		t.Parallel()
		runCases(t, "help", []writesCase{
			{name: "the help page, in an amendment in progress", setup: inRepo(pending)},
			{name: "an argument help does not take", setup: inRepo(pending), args: []string{"diff"}},
		})
	})
	t.Run("row2 version", func(t *testing.T) {
		t.Parallel()
		runCases(t, "version", []writesCase{
			{name: "in an amendment in progress", setup: inRepo(pending)},
			{name: "an argument version does not take", setup: inRepo(pending), args: []string{"--since"}},
		})
	})
	t.Run("row3 lint", func(t *testing.T) {
		t.Parallel()
		runCases(t, "lint", []writesCase{
			{name: "clean", setup: inDir(baseline())},
			{name: "errors", setup: inDir(baseline().with("proviso/notes.md", "# Notes\n", "proviso/capabilities/billing/glossary.md", "# Glossary\n"))},
			{name: "asks", setup: inDir(baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?"))},
			{name: "hidden", setup: inDir(baseline().with("proviso/.DS_Store", "\x00\x01", "proviso/.markdownlint.jsonc", "{}\n"))},
			{name: "unreadable", setup: inDir(baseline().with("proviso/capabilities/cart/features/a.md", "\xff\xfe\n"))},
			{name: "many errors", setup: inDir(withRule("But the shop\nWhen <provided>", "| x |\n| --- |\n"))},
			{name: "in an amendment in progress", setup: inRepo(pending)},
			{name: "a merge in progress", setup: inRepo(mergeInProgress)},
		})
	})
	t.Run("row4 show", func(t *testing.T) {
		t.Parallel()
		tr := showTree()
		amended := inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, tr)
			r.set(tr.edit(glossaryPath, "A product placed in the cart.", "A product put into the cart."))
			untracked(t, r)
			return r
		})
		runCases(t, "show", []writesCase{
			{name: "found, as element_view", setup: inDir(tr), args: []string{"cart.add.limit", "--json"}},
			{name: "found, as show_text", setup: inDir(tr), args: []string{"finding"}},
			{name: "none found", setup: inDir(tr), args: []string{"nosuch"}},
			{name: "no query", setup: inDir(tr)},
			{name: "a capability the tree does not hold", setup: inDir(tr), args: []string{"system", "--in", "nosuch"}},
			{name: "an argument show does not take", setup: inDir(tr), args: []string{"system", "--write"}},
			{name: "in an amendment in progress", setup: amended, args: []string{"cart.item", "--json"}},
			{name: "a tree lint cannot read", setup: inDir(tr.with("proviso/manifest.sum", "model 2\n")), args: []string{"system", "--json"}},
		})
	})
	for _, row := range []struct{ name, cmd string }{
		{"row5 diff", "diff"},
		{"row6 pack", "pack"},
		{"row7 impact", "impact"},
		{"row8 verify", "verify"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			runCases(t, row.cmd, treeCases())
		})
	}
	t.Run("row6 pack --element", func(t *testing.T) {
		t.Parallel()
		cs := withArgs(treeCases(), "", "--element", "system")
		cs = append(cs, writesCase{name: "an element the tree does not hold", setup: inRepo(pending), args: []string{"--element", "nosuch"}})
		runCases(t, "pack", cs)
	})
	t.Run("row9 manifest", func(t *testing.T) {
		t.Parallel()
		runCases(t, "manifest", []writesCase{
			{name: "written", setup: inRepo(func(t *testing.T) *repo {
				r, _ := accepted(t, diffTree())
				staged(t, r)
				return r
			})},
			{name: "written, with a coherence file", setup: inRepo(func(t *testing.T) *repo {
				r, _ := accepted(t, diffTree())
				r.set(r.tr.apply(limitThen))
				r.judgeOK(checkResult{})
				staged(t, r)
				return r
			})},
			{name: "written, with a coherence file and an amendment note", setup: inRepo(func(t *testing.T) *repo {
				r, _ := accepted(t, diffTree())
				r.set(r.tr.apply(limitThen))
				r.judgeOK(checkResult{})
				r.writeNote("- refunds\n")
				staged(t, r)
				return r
			})},
			{name: "written, a merge in progress", setup: inRepo(mergeInProgress)},
			{name: "a tree with an error", setup: inRepo(func(t *testing.T) *repo {
				r, _ := accepted(t, diffTree())
				r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
				return r
			})},
			{name: "a tree with an error, with a coherence file", setup: inRepo(func(t *testing.T) *repo {
				r, _ := accepted(t, diffTree())
				r.set(r.tr.apply(limitThen))
				r.judgeOK(checkResult{})
				r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
				return r
			})},
			{name: "a tree with an error, with an amendment note", setup: inRepo(func(t *testing.T) *repo {
				r, _ := accepted(t, diffTree())
				r.set(r.tr.apply(limitThen))
				r.writeNote("- refunds\n")
				r.set(r.tr.with("proviso/notes.md", "# Notes\n"))
				return r
			})},
			{name: "a manifest of another model version", setup: inRepo(func(t *testing.T) *repo {
				return newRepo(t, diffTree().with(manifestPath, "model 2\n"))
			})},
		})
	})
	t.Run("row10 judge", func(t *testing.T) {
		t.Parallel()
		cs := treeCases()
		cs = append(cs,
			writesCase{name: "a check result that is no JSON", setup: inRepo(pending), input: "this is no JSON"},
			writesCase{name: "a finding, with a coherence file", setup: inRepo(func(t *testing.T) *repo {
				r := pending(t)
				r.judgeOK(checkResult{})
				return r
			}), input: checkResult{Findings: []findingIn{gap("A gap.")}}.String()},
		)
		runCases(t, "judge", cs)
	})
	t.Run("row11 rehome", func(t *testing.T) {
		t.Parallel()
		withCoherence := inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, mainRehomeTree())
			r.set(mainRehomeTree().with(coherencePath, "{}\n"))
			return r
		})
		renames := inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, mainRehomeTree())
			untracked(t, r)
			return r
		})
		taken := inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, takenRehomeTree())
			return r
		})
		mentions := inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, mentionsTree())
			return r
		})
		picked := mentionList(mentionOn(featurePath, mentionLineB), mentionOn(cartQuestions, mentionLineG))
		runCases(t, "rehome", []writesCase{
			{name: "without --write, exit status 0", setup: renames, args: []string{"--json"}},
			{name: "with --write, exit status 0", setup: renames, args: []string{"--write", "--json"}},
			{name: "with --write, in the text form, exit status 0", setup: renames, args: []string{"--write"}},
			{name: "with --write, exit status 1", setup: taken, args: []string{"--write", "--json"}},
			{name: "with --write, exit status 2", setup: withCoherence, args: []string{"--write", "--json"}},
			{name: "with --write --mentions, exit status 0", setup: mentions, input: picked, args: []string{"--write", "--mentions", "--json"}},
			{name: "with --write --mentions, exit status 1", setup: taken, input: mentionList(), args: []string{"--write", "--mentions", "--json"}},
			{name: "with --write, a merge in progress", setup: inRepo(mergeInProgress), args: []string{"--write", "--json"}},
		})
	})
}

// everyCommand is every command command_writes lists, with the arguments a
// run of it is given where the derived checks of system.tree-writes run it
// in every state.
var everyCommand = []struct {
	cmd, input string
	args       []string
}{
	{"help", "", nil},
	{"version", "", nil},
	{"lint", "", nil},
	{"show", "", []string{"system", "--json"}},
	{"diff", "", nil},
	{"pack", "", nil},
	{"pack", "", []string{"--element", "system"}},
	{"impact", "", nil},
	{"verify", "", nil},
	{"manifest", "", nil},
	{"judge", "{}", nil},
	{"rehome", "", []string{"--write", "--json"}},
}

// TestSystemTreeWrites checks the invariant system.tree-writes over every
// command command_writes lists, whatever it finds and however it ends: in a
// repository with an amendment in progress, staged and untracked, with a
// merge in progress, with a tree lint cannot read, and with no spec root.
// TestSystemCommandWrites checks the runs of each command row by row.
func TestSystemTreeWrites(t *testing.T) {
	t.Parallel()
	states := []struct {
		name  string
		setup func(t *testing.T) (string, string)
	}{
		{"an amendment in progress, staged and untracked", inRepo(pending)},
		{"a merge in progress", inRepo(mergeInProgress)},
		{"a tree lint cannot read", inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.apply(limitThen).with("proviso/capabilities/cart/features/a.md", "\xff\xfe\n"))
			return r
		})},
		{"a manifest of another model version", inRepo(func(t *testing.T) *repo {
			r, _ := accepted(t, diffTree())
			r.set(r.tr.with(manifestPath, "model 2\n"))
			return r
		})},
		{"derived: no spec root, in a git repository", func(t *testing.T) (string, string) {
			r := newRepo(t, tree{"notes/todo.md": "# To do\n"})
			r.commit("notes")
			untracked(t, r)
			skipIfSpecRootAbove(t, r.dir)
			return r.dir, ""
		}},
	}
	for _, s := range states {
		for _, c := range everyCommand {
			name := s.name + ", " + strings.TrimSpace(c.cmd+" "+strings.Join(c.args, " "))
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				base, root := s.setup(t)
				if _, ok := snapshotAll(t, base)[".git/HEAD"]; !ok {
					t.Fatal("the snapshot does not cover the git repository")
				}
				writesIn(t, c.cmd, base, root, c.input, c.args...)
			})
		}
	}
}
