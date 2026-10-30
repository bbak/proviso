package acceptance

// The capability cli: proviso/capabilities/cli/, the feature cli.help and the
// contract cli.cli.help. The help page's wording is for people, so the tests
// check where it goes, what it names and the exit status; a refusal is one
// line on standard error, and only what it names is checked: the command,
// the argument and "proviso help".

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// commands are the commands of Proviso, as cli.help.command names them.
var commands = []string{"init", "update", "version", "lint", "manifest", "diff", "impact", "pack", "judge", "verify", "help"}

// options are the options the commands take.
var options = []string{"--json", "--since"}

// anywhere gives a fresh directory in no git repository and with no spec
// root, as far as the temporary folder allows.
func anywhere(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "notes")
	tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
	return dir
}

// stderrLines gives the lines on standard error without their line endings,
// the last line ending dropped.
func stderrLines(s string) []string {
	s = strings.TrimSuffix(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// provisoIn runs proviso with args in cwd, with stdin, if not empty, on its
// standard input, and checks that nothing below base changed.
func provisoIn(t *testing.T, base, cwd, stdin string, args ...string) run {
	t.Helper()
	before := snapshotAll(t, base)
	var r run
	if stdin != "" {
		r = execProvisoWith(t, cwd, strings.NewReader(stdin), args...)
	} else {
		r = execProviso(t, cwd, args...)
	}
	after := snapshotAll(t, base)
	checkWrites(t, before, after, "")
	return r
}

// helpPage runs `proviso help` anywhere and returns the page, which must have
// been written.
func helpPage(t *testing.T) string {
	t.Helper()
	dir := anywhere(t)
	r := provisoIn(t, dir, dir, "", "help")
	if r.exit != 0 || strings.TrimSpace(r.stdout) == "" || r.stderr != "" {
		t.Fatalf("cli.help.page: want the page on stdout, nothing on stderr and exit status 0, got exit status %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	return r.stdout
}

// isPage checks that r wrote the help page to stdout, nothing to stderr, and
// exited with 0.
func isPage(t *testing.T, r run, page string) {
	t.Helper()
	if r.exit != 0 {
		t.Errorf("cli.cli.help.exit-status: want 0, got %d\nstderr: %s", r.exit, r.stderr)
	}
	if r.stderr != "" {
		t.Errorf("cli.help.page: nothing goes to stderr, got %q", r.stderr)
	}
	if norm := strings.ReplaceAll(r.stdout, "\r\n", "\n"); norm != strings.ReplaceAll(page, "\r\n", "\n") {
		t.Errorf("cli.help.page: want the help page on stdout, got:\n%s", r.stdout)
	}
}

// misused checks the refusal of cli.help.argument: exit status 2, nothing on
// stdout, and one line on stderr naming the command, the argument and that
// "proviso help" shows the help page. That nothing was written is checked
// by the run (provisoIn, versionIn).
func misused(t *testing.T, r run, command, argument string) {
	t.Helper()
	if r.exit != 2 {
		t.Errorf("cli.help.argument: want exit status 2, got %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("cli.help.argument: nothing goes to stdout, got %q", r.stdout)
	}
	lines := stderrLines(r.stderr)
	if len(lines) != 1 {
		t.Errorf("cli.help.argument: want one line on stderr, got %d:\n%s", len(lines), r.stderr)
		return
	}
	for _, want := range []string{command, argument, "proviso help"} {
		if want != "" && !strings.Contains(lines[0], want) {
			t.Errorf("cli.help.argument: want the line to name %q, got %q", want, lines[0])
		}
	}
}

// unknown checks the refusal of cli.help.command: exit status 2, nothing on
// stdout, and one line on stderr naming the command, if not empty, and that
// "proviso help" shows the help page.
func unknown(t *testing.T, r run, command string) {
	t.Helper()
	if r.exit != 2 {
		t.Errorf("cli.cli.help.exit-status: want 2 for an unknown command, got %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("cli.help.command: nothing goes to stdout, got %q", r.stdout)
	}
	lines := stderrLines(r.stderr)
	if len(lines) != 1 {
		t.Errorf("cli.help.command: want one line on stderr, got %d:\n%s", len(lines), r.stderr)
		return
	}
	for _, want := range []string{command, "proviso help"} {
		if want != "" && !strings.Contains(lines[0], want) {
			t.Errorf("cli.help.command: want the line to name %q, got %q", want, lines[0])
		}
	}
}

// sinceAccepted gives an accepted diffTree amended by limitThen and committed, so
// that HEAD~1 is the accepted tree, and every baseline command has something
// to work on: the element diff lists cart.add.limit.
func sinceAccepted(t *testing.T) *repo {
	t.Helper()
	r, _ := accepted(t, diffTree())
	r.set(diffTree().apply(limitThen))
	r.commit("round 1")
	return r
}

func TestCliCliHelp(t *testing.T) {
	t.Parallel()
	page := helpPage(t)
	for _, arg := range []string{"help", "--help", "-h"} {
		t.Run("derived: proviso "+arg+", run anywhere, writes the page", func(t *testing.T) {
			t.Parallel()
			dir := anywhere(t)
			isPage(t, provisoIn(t, dir, dir, "", arg), page)
		})
	}
	t.Run("derived: run inside a repository, the same page", func(t *testing.T) {
		t.Parallel()
		r := sinceAccepted(t)
		isPage(t, provisoIn(t, r.dir, r.dir, "", "help"), page)
	})
}

func TestCliCliHelpPage(t *testing.T) {
	t.Parallel()
	page := helpPage(t)
	// The page names every command and every option; help is a command
	// too.
	for _, name := range append(append([]string(nil), commands...), options...) {
		re := regexp.MustCompile(`(^|[^A-Za-z0-9-])` + regexp.QuoteMeta(name) + `($|[^A-Za-z0-9-])`)
		if !re.MatchString(page) {
			t.Errorf("cli.cli.help.page: the page does not name %s:\n%s", name, page)
		}
	}
	// The commands in the page's order: first those engineers run, then
	// those the skills run. A command is listed on the first line that
	// starts with it, or with "proviso" and it.
	order := []string{"init", "update", "version", "lint", "diff", "impact", "verify", "help", "pack", "judge", "manifest"}
	lines := strings.Split(strings.ReplaceAll(page, "\r\n", "\n"), "\n")
	last, lastName := -1, ""
	for _, name := range order {
		re := regexp.MustCompile(`^\s*(?:proviso\s+)?` + regexp.QuoteMeta(name) + `($|[^A-Za-z0-9-])`)
		at := -1
		for i, l := range lines {
			if re.MatchString(l) {
				at = i
				break
			}
		}
		if at < 0 {
			t.Errorf("cli.cli.help.page: no line of the page lists %s:\n%s", name, page)
			continue
		}
		if at <= last {
			t.Errorf("cli.cli.help.page: %s is listed on line %d, not after %s on line %d:\n%s", name, at+1, lastName, last+1, page)
		}
		last, lastName = at, name
	}
}

func TestCliCliHelpExitStatus(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		args []string
		want int
	}{
		{"0: the page was written", []string{"help"}, 0},
		{"2: an argument given to help", []string{"help", "--json"}, 2},
		{"2: no command", nil, 2},
		{"2: an unknown command", []string{"build"}, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := anywhere(t)
			if r := provisoIn(t, dir, dir, "", c.args...); r.exit != c.want {
				t.Errorf("cli.cli.help.exit-status: want %d, got %d\nstderr: %s", c.want, r.exit, r.stderr)
			}
		})
	}
}

func TestCliHelpPage(t *testing.T) {
	t.Parallel()
	page := helpPage(t)
	t.Run("row1 none", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		isPage(t, provisoIn(t, dir, dir, "", "help"), page)
	})
	for _, c := range []struct{ name, arg string }{
		{"row2 diff", "diff"},
		{"row3 --json", "--json"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := anywhere(t)
			// Handed over to cli.help.argument.
			misused(t, provisoIn(t, dir, dir, "", "help", c.arg), "help", c.arg)
		})
	}
}

func TestCliHelpStart(t *testing.T) {
	t.Parallel()
	page := helpPage(t)
	t.Run("row1 lint --json", func(t *testing.T) {
		t.Parallel()
		check(t, baseline()).clean(t)
	})
	t.Run("row2 help", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		isPage(t, provisoIn(t, dir, dir, "", "help"), page)
	})
	t.Run(`row3 ""`, func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		// "" is the command, and no command of Proviso.
		unknown(t, provisoIn(t, dir, dir, "", ""), "")
	})
	t.Run("row4 none", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		r := provisoIn(t, dir, dir, "")
		if r.exit != 2 {
			t.Errorf("cli.help.start: want exit status 2, got %d", r.exit)
		}
		if r.stdout != "" {
			t.Errorf("cli.help.start: nothing goes to stdout, got %q", r.stdout)
		}
		if got := strings.ReplaceAll(r.stderr, "\r\n", "\n"); got != strings.ReplaceAll(page, "\r\n", "\n") {
			t.Errorf("cli.help.start: want the help page on stderr, got:\n%s", r.stderr)
		}
	})
}

func TestCliHelpCommand(t *testing.T) {
	t.Parallel()
	page := helpPage(t)
	t.Run("row1 lint", func(t *testing.T) {
		t.Parallel()
		check(t, baseline()).clean(t)
	})
	t.Run("row2 verify", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, diffTree())
		// verifyIn checks the verification on stdout and the exit status.
		if o := r.verify(); o.json.exit == 2 {
			t.Errorf("cli.help.command: want verify to run, got exit status 2\nstderr: %s", o.json.stderr)
		}
	})
	t.Run("row3 update", func(t *testing.T) {
		t.Parallel()
		// A release, after init: update runs and updates every file.
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.all(t, "updated")
	})
	for _, c := range []struct{ name, cmd string }{
		{"row4 help", "help"},
		{"row5 --help", "--help"},
		{"row6 -h", "-h"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := anywhere(t)
			isPage(t, provisoIn(t, dir, dir, "", c.cmd), page)
		})
	}
	for _, c := range []struct{ name, cmd string }{
		{"row7 Lint", "Lint"},
		{"row8 --version", "--version"},
		{"row9 build", "build"},
		{`row10 ""`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// In a tree lint accepts, so that a command run anyway would
			// show.
			dir := baseline().write(t)
			unknown(t, provisoIn(t, dir, dir, "", c.cmd), c.cmd)
		})
	}
	t.Run("derived: an unknown command is refused before its arguments", func(t *testing.T) {
		t.Parallel()
		dir := baseline().write(t)
		unknown(t, provisoIn(t, dir, dir, "", "build", "--json"), "build")
	})
}

func TestCliHelpArgument(t *testing.T) {
	t.Parallel()
	t.Run("row1 lint, none, --json, none", func(t *testing.T) {
		t.Parallel()
		check(t, baseline()).clean(t)
	})
	t.Run("row2 lint, --json, --json, none", func(t *testing.T) {
		t.Parallel()
		dir := baseline().write(t)
		r := provisoIn(t, dir, dir, "", "lint", "--json", "--json")
		if r.exit != 0 {
			t.Fatalf("cli.help.argument: want lint to read --json twice and run, got exit status %d\nstderr: %s", r.exit, r.stderr)
		}
		if rep := parseReport(t, r.stdout); len(rep.Findings) != 0 {
			t.Errorf("want no findings, got:\n%s", dump(rep.Findings))
		}
	})
	t.Run("row3 diff, none, --since, HEAD~1", func(t *testing.T) {
		t.Parallel()
		r := sinceAccepted(t)
		// The element diff lists cart.add.limit since HEAD~1, the
		// accepted tree, so diff exits with 1.
		o := provisoIn(t, r.dir, r.dir, "", "diff", "--since", "HEAD~1")
		if o.exit != 1 {
			t.Errorf("cli.help.argument: want diff to compare with HEAD~1 and exit with 1, got %d\nstderr: %s", o.exit, o.stderr)
		}
	})
	t.Run("row4 pack, --json, --since, a1b2c3d", func(t *testing.T) {
		t.Parallel()
		r := sinceAccepted(t)
		rev := r.git("rev-parse", "HEAD~1")[:7]
		// packIn runs `pack --json --since <rev>`, and diff with the same
		// arguments.
		p := packIn(t, r.dir, r.dir, r.root(), "--since", rev)
		p.ran(t)
		p.diff.only(t, "cart.add.limit:changed")
	})

	// The refused rows. Each runs where the command, had it run, would
	// have checked or written something.
	for _, c := range []struct {
		name     string
		where    string // "anywhere", "lint", "fresh" (a repository before init), "amended"
		stdin    string
		args     []string
		command  string
		argument string
	}{
		{"row6 init, none, --json, none", "fresh", "", []string{"init", "--json"}, "init", "--json"},
		{"row7 judge, none, --json, none", "amended", "{}", []string{"judge", "--json"}, "judge", "--json"},
		{"row8 help, none, diff, none", "anywhere", "", []string{"help", "diff"}, "help", "diff"},
		{"row9 lint, none, --since, HEAD~1", "lint", "", []string{"lint", "--since", "HEAD~1"}, "lint", "--since"},
		{"row10 diff, none, --help, none", "amended", "", []string{"diff", "--help"}, "diff", "--help"},
		{"row11 diff, none, --since, none", "amended", "", []string{"diff", "--since"}, "diff", "--since"},
		{`row12 pack, none, --since, ""`, "amended", "", []string{"pack", "--since", ""}, "pack", "--since"},
		{"row13 diff, none, --since, --json", "amended", "", []string{"diff", "--since", "--json"}, "diff", "--since"},
		{"row14 diff, --since HEAD~1, --since, HEAD~2", "amended", "", []string{"diff", "--since", "HEAD~1", "--since", "HEAD~2"}, "diff", "--since"},
		{"derived: a revision that starts with a hyphen", "amended", "", []string{"diff", "--since", "-HEAD"}, "diff", "--since"},
		{"derived: the arguments are read before the spec root is looked for", "anywhere", "", []string{"lint", "--bogus"}, "lint", "--bogus"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var dir string
			switch c.where {
			case "anywhere":
				dir = anywhere(t)
			case "lint":
				dir = baseline().write(t)
			case "fresh":
				dir = freshRepo(t, "repo", false).dir
			case "amended":
				r := sinceAccepted(t)
				// HEAD~2 exists, so that two revisions both name commits.
				r.set(diffTree().apply(limitThen, limitProvided))
				r.commit("round 2")
				dir = r.dir
			}
			misused(t, provisoIn(t, dir, dir, c.stdin, c.args...), c.command, c.argument)
		})
	}
	t.Run("row5 version, none, 0.1.0, none", func(t *testing.T) {
		t.Parallel()
		// After init, in a root whose skills are stale, so that anything
		// checked would show; versionIn checks that nothing is written.
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		o := versionIn(t, "0.1.0", r.dir, r.dir, nil, "0.1.0")
		misused(t, o.run, "version", "0.1.0")
	})

	t.Run("derived: every command refuses an argument it does not take", func(t *testing.T) {
		t.Parallel()
		for _, cmd := range commands {
			t.Run(cmd, func(t *testing.T) {
				t.Parallel()
				var dir string
				if cmd == "init" || cmd == "update" {
					dir = freshRepo(t, "repo", false).dir
				} else {
					dir = sinceAccepted(t).dir
				}
				stdin := ""
				if cmd == "judge" {
					stdin = "{}"
				}
				misused(t, provisoIn(t, dir, dir, stdin, cmd, "--bogus"), cmd, "--bogus")
			})
		}
	})
}

// TestSystemCliManifest checks the contract system.cli.manifest: it takes no
// argument.
func TestSystemCliManifest(t *testing.T) {
	t.Parallel()
	t.Run("derived: it takes no argument", func(t *testing.T) {
		t.Parallel()
		r := sinceAccepted(t)
		misused(t, provisoIn(t, r.dir, r.dir, "", "manifest", "--json"), "manifest", "--json")
		if _, err := os.Lstat(filepath.Join(r.root(), "manifest.sum")); err != nil {
			t.Errorf("system.cli.manifest: the manifest written before is gone: %v", err)
		}
	})
}

// TestSystemCliJudge checks the contract system.cli.judge: it takes no
// argument.
func TestSystemCliJudge(t *testing.T) {
	t.Parallel()
	t.Run("derived: it takes no argument", func(t *testing.T) {
		t.Parallel()
		r := sinceAccepted(t)
		misused(t, provisoIn(t, r.dir, r.dir, "{}", "judge", "f1"), "judge", "f1")
	})
}
