package acceptance

// The driver of `proviso update`: runs of update of a given Proviso version,
// with what holds for every run (the contract install.cli.update of its
// standard output, the exit status, the invariant install.update.keeps, and
// what each line on standard output says about the file it names), and the
// content `init` of a Proviso version writes, which update writes too.

import (
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

func isStamped(p string) bool {
	for _, q := range stampedFiles {
		if p == q {
			return true
		}
	}
	return false
}

// initContents maps a Proviso version to the *kitContent init of it writes.
var initContents sync.Map

type kitContent struct {
	once  sync.Once
	files map[string]string
	err   string
}

// initContent gives the content init of the Proviso version v writes at the
// skill or agent definition file, run once per version in a fresh
// repository: install.update.missing and install.update.stamp write a file
// "as `init` of its own Proviso version writes it".
func initContent(t *testing.T, v, file string) string {
	t.Helper()
	e, _ := initContents.LoadOrStore(v, &kitContent{})
	k := e.(*kitContent)
	k.once.Do(func() {
		r := freshRepo(t, "kit", false)
		if o := execBinary(t, built(t, v), r.dir, "init"); o.exit != 0 {
			k.err = "init of " + v + " exited with " + o.stderr
			return
		}
		k.files = map[string]string{}
		for _, f := range stampedFiles {
			k.files[f] = r.read(f)
		}
	})
	if k.err != "" {
		t.Fatal(k.err)
	}
	return k.files[file]
}

// updateLine is one line of the standard output of `update`.
type updateLine struct{ status, path string }

// updateOutcome is one run of `update`, with every entry below the base
// directory before and after it.
type updateOutcome struct {
	run
	v             string
	lines         []updateLine
	stderrLines   []string
	before, after map[string]string
	prefix        string // the run's directory relative to base, with a trailing /, or ""
}

var updateLinePattern = regexp.MustCompile(`^(written|updated|skipped) (.+)$`)

// updateIn runs `proviso update` of the Proviso version v, with args, in
// cwd, and checks what holds for every run. base is the directory the test
// set up, cwd at or below it.
func updateIn(t *testing.T, v, base, cwd string, args ...string) updateOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	r := execBinary(t, built(t, v), cwd, append([]string{"update"}, args...)...)
	o := updateOutcome{run: r, v: v, before: before, after: snapshotAll(t, base)}
	if rel, _ := filepath.Rel(base, cwd); rel != "." {
		o.prefix = filepath.ToSlash(rel) + "/"
	}
	for _, l := range strings.Split(strings.ReplaceAll(r.stderr, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			o.stderrLines = append(o.stderrLines, l)
		}
	}
	if r.exit != 0 && r.exit != 2 {
		t.Fatalf("install.cli.update.exit-status: %d is not 0 or 2\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	if r.exit == 2 && len(o.stderrLines) == 0 {
		t.Errorf("install.cli.update.exit-status: exit status 2 with nothing on stderr")
	}
	o.parse(t)
	o.checkKeeps(t)
	o.checkLines(t)
	return o
}

// update runs `update` of the Proviso version v in the root of the
// repository.
func (r *repo) update(v string) updateOutcome {
	r.t.Helper()
	return updateIn(r.t, v, r.dir, r.dir)
}

// parse checks install.cli.update: one line per skill or agent definition,
// `written`, `updated` or `skipped`, a space and the path as installed_files
// gives it, in the byte order of the paths.
func (o *updateOutcome) parse(t *testing.T) {
	t.Helper()
	text := strings.TrimSuffix(strings.ReplaceAll(o.stdout, "\r\n", "\n"), "\n")
	if text == "" {
		return
	}
	for _, l := range strings.Split(text, "\n") {
		m := updateLinePattern.FindStringSubmatch(l)
		if m == nil || !isStamped(m[2]) {
			t.Errorf("install.cli.update: %q is not `written`, `updated` or `skipped`, a space and the path of a skill or agent definition installed_files names\nstdout:\n%s", l, o.stdout)
			continue
		}
		if n := len(o.lines); n > 0 && o.lines[n-1].path >= m[2] {
			t.Errorf("install.cli.update: %s named after %s, not once each in the byte order of the paths", m[2], o.lines[n-1].path)
		}
		o.lines = append(o.lines, updateLine{m[1], m[2]})
	}
}

// isFile tells whether a snapshot entry is a file.
func isFile(entry string) bool {
	return strings.HasPrefix(entry, "file: ") || strings.HasPrefix(entry, "unreadable: ")
}

// checkKeeps checks install.update.keeps: every file other than the skills
// and agent definitions is byte for byte as it was, no entry is removed or
// renamed, and the only entries added are those paths and the folders above
// them, below the run's directory.
func (o updateOutcome) checkKeeps(t *testing.T) {
	t.Helper()
	owned := map[string]bool{}
	allowed := map[string]bool{}
	for _, q := range stampedFiles {
		owned[o.prefix+q] = true
		allowed[o.prefix+q] = true
		for d := path.Dir(q); d != "."; d = path.Dir(d) {
			allowed[o.prefix+d] = true
		}
	}
	var bad []string
	for p, v := range o.before {
		w, ok := o.after[p]
		switch {
		case !ok:
			bad = append(bad, p+", which existed before, was removed or renamed")
		case v == w:
		case owned[p] && isFile(v) && isFile(w):
			// A skill or agent definition replaced.
		default:
			bad = append(bad, p+", which existed before, was changed")
		}
	}
	for p := range o.after {
		if _, ok := o.before[p]; !ok && !allowed[p] {
			bad = append(bad, p+" was added, but it is neither a skill or agent definition nor a folder above one")
		}
	}
	sort.Strings(bad)
	for _, b := range bad {
		t.Errorf("install.update.keeps: %s", b)
	}
}

// checkLines checks the output against the entries: a file named as written
// had no entry before; one named as updated had one; both are now what init
// of the binary's version writes, stamped with that version; a file named as
// skipped is as it was. A file not named is as it was, and with exit status
// 0 every file is named.
func (o updateOutcome) checkLines(t *testing.T) {
	t.Helper()
	named := map[string]bool{}
	for _, l := range o.lines {
		named[l.path] = true
		p := o.prefix + l.path
		before, existed := o.before[p]
		after := o.after[p]
		switch l.status {
		case "written", "updated":
			if l.status == "written" && existed {
				t.Errorf("install.update.missing: %s named as written, but an entry existed there", l.path)
			}
			if l.status == "updated" && !existed {
				t.Errorf("install.update.stamp: %s named as updated, but no entry existed there", l.path)
			}
			if !strings.HasPrefix(after, "file: ") {
				t.Errorf("install.cli.update: %s named as %s, but no file is there", l.path, l.status)
				continue
			}
			content := strings.TrimPrefix(after, "file: ")
			if content != initContent(t, o.v, l.path) {
				t.Errorf("install.update.%s: %s named as %s, but it is not as init of %s writes it:\n%s", map[string]string{"written": "missing", "updated": "stamp"}[l.status], l.path, l.status, o.v, content)
			}
			if got := stampsOf(content); len(got) != 1 || got[0] != o.v {
				t.Errorf("install.update.release: %s, written by update of %s, holds the stamps %q", l.path, o.v, got)
			}
		case "skipped":
			if !existed {
				t.Errorf("install.update.stamp: %s named as skipped, but no entry existed there", l.path)
			} else if before != after {
				t.Errorf("install.update.stamp: %s named as skipped, but it changed", l.path)
			}
		}
	}
	for _, q := range stampedFiles {
		if named[q] {
			continue
		}
		if o.exit == 0 {
			t.Errorf("install.update.exit: exit status 0, but %s is not named on stdout\nstdout:\n%s", q, o.stdout)
		}
		p := o.prefix + q
		if o.before[p] != o.after[p] {
			t.Errorf("install.update.exit: %s is not named on stdout, but was written", q)
		}
	}
}

// exits asserts the exit status.
func (o updateOutcome) exits(t *testing.T, want int) {
	t.Helper()
	if o.exit != want {
		t.Errorf("want exit status %d, got %d\nstdout:\n%s\nstderr:\n%s", want, o.exit, o.stdout, o.stderr)
	}
}

// says asserts that stdout names file with status.
func (o updateOutcome) says(t *testing.T, status, file string) {
	t.Helper()
	for _, l := range o.lines {
		if l.path == file {
			if l.status != status {
				t.Errorf("want %s %s, got %s %s", status, file, l.status, file)
			}
			return
		}
	}
	t.Errorf("want %s %s on stdout, got:\n%s", status, file, o.stdout)
}

// all asserts that stdout names every skill and agent definition with
// status.
func (o updateOutcome) all(t *testing.T, status string) {
	t.Helper()
	if len(o.lines) != len(stampedFiles) {
		t.Errorf("want every skill and agent definition named as %s, got:\n%s", status, o.stdout)
	}
	for _, q := range stampedFiles {
		o.says(t, status, q)
	}
}

// stderrNaming gives the lines on standard error that name file.
func (o updateOutcome) stderrNaming(file string) []string {
	var found []string
	for _, l := range o.stderrLines {
		for _, n := range names(l) {
			if n == file {
				found = append(found, l)
			}
		}
	}
	return found
}

// failed asserts, for install.update.exit, that each file could not be
// written: no line on stdout names it, standard error says why, and the exit
// status is 2.
func (o updateOutcome) failed(t *testing.T, files ...string) {
	t.Helper()
	o.exits(t, 2)
	for _, f := range files {
		for _, l := range o.lines {
			if l.path == f {
				t.Errorf("install.update.exit: %s could not be written, but stdout names it as %s", f, l.status)
			}
		}
		if len(o.stderrNaming(f)) == 0 {
			t.Errorf("install.update.exit: %s could not be written, but no line on stderr names it:\n%s", f, o.stderr)
		}
	}
}

// newer asserts the line of install.update.stamp naming file and stamp as
// newer than the binary's version.
func (o updateOutcome) newer(t *testing.T, file, stamp string) {
	t.Helper()
	lines := o.stderrNaming(file)
	if len(lines) != 1 {
		t.Errorf("install.update.stamp: want one line on stderr naming %s as newer, got %d:\n%s", file, len(lines), o.stderr)
		return
	}
	if !strings.Contains(lines[0], stamp) || !strings.Contains(strings.ReplaceAll(lines[0], stamp, " "), o.v) {
		t.Errorf("install.update.stamp: want the line to name %s and %s as newer than %s: %q", file, stamp, o.v, lines[0])
	}
}

// quietAbout asserts that no line on standard error names file.
func (o updateOutcome) quietAbout(t *testing.T, file string) {
	t.Helper()
	if l := o.stderrNaming(file); len(l) != 0 {
		t.Errorf("install.update.stamp: want no line on stderr about %s, got %q", file, l)
	}
}

// refused asserts a refusal before any file is taken: exit status 2, nothing
// on stdout, nothing below base written or changed, and, with one, a single
// line on stderr.
func (o updateOutcome) refused(t *testing.T, rule string, one bool) {
	t.Helper()
	o.exits(t, 2)
	if o.stdout != "" {
		t.Errorf("%s: nothing goes to stdout, got %q", rule, o.stdout)
	}
	checkWritesAs(t, rule, o.before, o.after)
	if one && len(o.stderrLines) != 1 {
		t.Errorf("%s: want one line on stderr, got %d:\n%s", rule, len(o.stderrLines), o.stderr)
	}
}

// checkWritesAs checks that every entry is as it was, and none was added.
func checkWritesAs(t *testing.T, rule string, before, after map[string]string) {
	t.Helper()
	var bad []string
	for p, v := range before {
		if w, ok := after[p]; !ok {
			bad = append(bad, p+" was removed")
		} else if v != w {
			bad = append(bad, p+" changed")
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			bad = append(bad, p+" was written")
		}
	}
	sort.Strings(bad)
	for _, b := range bad {
		t.Errorf("%s: nothing is written, but %s", rule, b)
	}
}
