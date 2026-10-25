package acceptance

// The driver of the install capability: fresh git repositories, runs of
// `proviso init` with what holds for every run (the contract of its standard
// output, the exit status, the rule install.init.existing as far as the
// output shows it, and the invariant install.init.keeps).

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

// installedFiles are the paths installed_files names, in their byte order.
var installedFiles = []string{
	".claude/agents/proviso-checker.md",
	".claude/agents/proviso-coder.md",
	".claude/agents/proviso-tester.md",
	".claude/skills/proviso-accept/SKILL.md",
	".claude/skills/proviso-build/SKILL.md",
	".claude/skills/proviso-check/SKILL.md",
	".claude/skills/proviso-integrate/SKILL.md",
	".proviso/build.md",
	".proviso/coder.md",
	".proviso/tester.md",
	"proviso/.gitattributes",
	"proviso/.markdownlint.jsonc",
	"proviso/system.md",
}

func isInstalled(p string) bool {
	for _, q := range installedFiles {
		if p == q {
			return true
		}
	}
	return false
}

// initLine is one line of the standard output of `init`.
type initLine struct{ status, path string }

// initOutcome is one run of `init`, with every entry below the base
// directory before and after it, as snapshotAll records them.
type initOutcome struct {
	run
	lines         []initLine
	before, after map[string]string
	prefix        string // the run's directory relative to base, with a trailing /, or ""
}

var initLinePattern = regexp.MustCompile(`^(written|skipped) (.+)$`)

// initIn runs `proviso init` in cwd and checks what holds for every run.
// base is the directory the test set up, cwd at or below it. prepare, if
// given, runs after the first snapshot and before init; what it returns runs
// after init and before the second snapshot.
func initIn(t *testing.T, base, cwd string, prepare func() func()) initOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	var undo func()
	if prepare != nil {
		undo = prepare()
	}
	r := execProviso(t, cwd, "init")
	if undo != nil {
		undo()
	}
	o := initOutcome{run: r, before: before, after: snapshotAll(t, base)}
	if rel, _ := filepath.Rel(base, cwd); rel != "." {
		o.prefix = filepath.ToSlash(rel) + "/"
	}
	if r.exit != 0 && r.exit != 2 {
		t.Fatalf("system.cli.init.exit-status: %d is not 0 or 2\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	if r.exit == 2 && strings.TrimSpace(r.stderr) == "" {
		t.Errorf("exit status 2 with nothing on stderr")
	}
	o.parse(t)
	o.checkKeeps(t)
	o.checkLines(t)
	return o
}

// parse checks system.cli.init: one line per installed file, `written` or
// `skipped`, a space and the path as installed_files gives it, in the byte
// order of the paths.
func (o *initOutcome) parse(t *testing.T) {
	t.Helper()
	text := strings.TrimSuffix(strings.ReplaceAll(o.stdout, "\r\n", "\n"), "\n")
	if text == "" {
		return
	}
	for _, l := range strings.Split(text, "\n") {
		m := initLinePattern.FindStringSubmatch(l)
		if m == nil || !isInstalled(m[2]) {
			t.Errorf("system.cli.init: %q is not `written` or `skipped`, a space and a path installed_files names\nstdout:\n%s", l, o.stdout)
			continue
		}
		if n := len(o.lines); n > 0 && o.lines[n-1].path >= m[2] {
			t.Errorf("system.cli.init: %s named after %s, not once each in the byte order of the paths", m[2], o.lines[n-1].path)
		}
		o.lines = append(o.lines, initLine{m[1], m[2]})
	}
}

// checkKeeps checks install.init.keeps: every entry that existed is as it
// was, and the only entries added are installed files and the folders above
// them, below the run's directory.
func (o initOutcome) checkKeeps(t *testing.T) {
	t.Helper()
	allowed := map[string]bool{}
	for _, q := range installedFiles {
		allowed[o.prefix+q] = true
		for d := path.Dir(q); d != "."; d = path.Dir(d) {
			allowed[o.prefix+d] = true
		}
	}
	var changed, added []string
	for p, v := range o.before {
		if w, ok := o.after[p]; !ok || v != w {
			changed = append(changed, p)
		}
	}
	for p := range o.after {
		if _, ok := o.before[p]; !ok && !allowed[p] {
			added = append(added, p)
		}
	}
	sort.Strings(changed)
	sort.Strings(added)
	for _, p := range changed {
		t.Errorf("install.init.keeps: %s, which existed before, was changed, removed or renamed", p)
	}
	for _, p := range added {
		t.Errorf("install.init.keeps: %s was added, but it is neither an installed file nor a folder above one", p)
	}
}

// checkLines checks the output against the entries: a file named as written
// did not exist and is now a file; a file named as skipped existed. With exit
// status 0, every installed file is named and in place; a file not named was
// not written.
func (o initOutcome) checkLines(t *testing.T) {
	t.Helper()
	named := map[string]bool{}
	for _, l := range o.lines {
		named[l.path] = true
		p := o.prefix + l.path
		_, existed := o.before[p]
		switch l.status {
		case "written":
			if existed {
				t.Errorf("install.init.existing: %s named as written, but an entry existed there", l.path)
			}
			if !strings.HasPrefix(o.after[p], "file: ") {
				t.Errorf("install.init.existing: %s named as written, but no file is there", l.path)
			}
		case "skipped":
			if !existed {
				t.Errorf("install.init.existing: %s named as skipped, but no entry existed there", l.path)
			}
		}
	}
	for _, q := range installedFiles {
		if named[q] {
			continue
		}
		if o.exit == 0 {
			t.Errorf("system.cli.init.exit-status: exit status 0, but %s is not named on stdout\nstdout:\n%s", q, o.stdout)
		}
		p := o.prefix + q
		if o.before[p] != o.after[p] {
			t.Errorf("install.init.exit: %s is not named on stdout, but was written", q)
		}
	}
}

// exits asserts the exit status.
func (o initOutcome) exits(t *testing.T, want int) {
	t.Helper()
	if o.exit != want {
		t.Errorf("want exit status %d, got %d\nstdout:\n%s\nstderr:\n%s", want, o.exit, o.stdout, o.stderr)
	}
}

// says asserts that stdout names path with status.
func (o initOutcome) says(t *testing.T, status, path string) {
	t.Helper()
	for _, l := range o.lines {
		if l.path == path {
			if l.status != status {
				t.Errorf("want %s %s, got %s %s", status, path, l.status, path)
			}
			return
		}
	}
	t.Errorf("want %s %s on stdout, got:\n%s", status, path, o.stdout)
}

// all asserts that stdout names every installed file with status.
func (o initOutcome) all(t *testing.T, status string) {
	t.Helper()
	if len(o.lines) != len(installedFiles) {
		t.Errorf("want every installed file named as %s, got:\n%s", status, o.stdout)
	}
	for _, q := range installedFiles {
		o.says(t, status, q)
	}
}

// nothingWritten asserts that the run left every entry below base as it was
// and wrote nothing to stdout.
func (o initOutcome) nothingWritten(t *testing.T) {
	t.Helper()
	for p := range o.after {
		if _, ok := o.before[p]; !ok {
			t.Errorf("install.init.repository: %s was written", p)
		}
	}
	if o.stdout != "" {
		t.Errorf("install.init.repository: stdout holds %q", o.stdout)
	}
}

// ---------------------------------------------------------------------------
// Fresh repositories

// repoTemplates maps whether the repository has a commit to the
// *acceptedRepo of its template, made once by `git init`.
var repoTemplates sync.Map

// freshRepo is a copy of a new git repository, in a folder named name of its
// own temporary directory; with committed, it holds a README.md committed on
// main, as a repository with code does; without, it has no commit yet.
func freshRepo(t *testing.T, name string, committed bool) *repo {
	t.Helper()
	skipWithoutGit(t)
	v, _ := repoTemplates.LoadOrStore(committed, &acceptedRepo{})
	a := v.(*acceptedRepo)
	a.once.Do(func() {
		dir, err := os.MkdirTemp(templates, "repo-")
		if err != nil {
			t.Fatal(err)
		}
		tr := tree{}
		if committed {
			tr = tree{"README.md": "# A system\n"}
		}
		r := newRepoIn(t, dir, tr)
		if committed {
			r.commit("start")
		}
		a.dir, a.tick, a.built = dir, r.tick, true
	})
	if !a.built {
		t.Fatal("making the repository failed in another test")
	}
	r := &repo{t: t, dir: filepath.Join(t.TempDir(), name), tick: a.tick, tr: tree{}}
	copyDir(t, a.dir, r.dir)
	r.cfg = filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(r.cfg, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return r
}

// init runs `proviso init` in the repository.
func (r *repo) init() initOutcome {
	r.t.Helper()
	return initIn(r.t, r.dir, r.dir, nil)
}

// read returns the content of the file at the slash-separated path p below
// the repository.
func (r *repo) read(p string) string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.dir, filepath.FromSlash(p)))
	if err != nil {
		r.t.Fatalf("reading %s: %v", p, err)
	}
	return string(b)
}
