package acceptance

// The driver of the baseline capability: git repositories for test trees,
// runs of `proviso manifest` and `proviso diff` with what holds for every run
// (the contracts' shapes, the order, the exit status and the invariant
// baseline.run.writes), and an oracle for manifest_file computed from the
// contract alone.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Running proviso

// execProviso runs the binary with args in cwd, with nothing on standard
// input.
func execProviso(t *testing.T, cwd string, args ...string) run {
	t.Helper()
	return execProvisoWith(t, cwd, nil, args...)
}

// execProvisoWith runs the binary with args in cwd, reading stdin, if not
// nil, as its standard input.
func execProvisoWith(t *testing.T, cwd string, stdin io.Reader, args ...string) run {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = cwd
	cmd.Stdin = stdin
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || ctx.Err() != nil {
			t.Fatalf("running proviso %v in %s: %v", args, cwd, err)
		}
		code = exitErr.ExitCode()
	}
	return run{code, out.String(), errOut.String()}
}

// ---------------------------------------------------------------------------
// The invariant baseline.run.writes

// snapshotAll records every entry below base, the git repository's .git
// folder included: folders, file contents and link targets, without
// following links.
func snapshotAll(t *testing.T, base string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(base, p)
		rel = filepath.ToSlash(rel)
		switch {
		case err != nil:
			snap[rel] = "error: " + err.Error()
		case d.Type()&fs.ModeSymlink != 0:
			target, _ := os.Readlink(p)
			snap[rel] = "link: " + target
		case d.IsDir():
			snap[rel] = "folder"
		case !d.Type().IsRegular():
			snap[rel] = "special: " + d.Type().String()
		default:
			b, err := os.ReadFile(p)
			if err != nil {
				snap[rel] = "unreadable: " + err.Error()
			} else {
				snap[rel] = "file: " + string(b)
			}
		}
		return nil
	})
	return snap
}

// checkWrites checks baseline.run.writes: everything below base is as it was,
// except the entries allowed, if not empty (manifest.sum and .coherence.json
// after a run of `manifest` that exited with 0, .coherence.json after a run
// of `judge` that exited with 0).
func checkWrites(t *testing.T, before, after map[string]string, allowed ...string) {
	t.Helper()
	isAllowed := func(p string) bool {
		for _, a := range allowed {
			if a != "" && p == a {
				return true
			}
		}
		return false
	}
	for p, v := range before {
		if isAllowed(p) {
			continue
		}
		if w, ok := after[p]; !ok {
			t.Errorf("baseline.run.writes: %s was removed", p)
		} else if v != w {
			t.Errorf("baseline.run.writes: %s changed", p)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok && !isAllowed(p) {
			t.Errorf("baseline.run.writes: %s was created", p)
		}
	}
}

// manifestIn runs `proviso manifest` in cwd. base is the directory the test
// tree was written to; root, if not empty, is the spec root the command is
// expected to work on, whose manifest.sum alone may change, and whose
// .coherence.json must be removed, and only if the command exits with 0. prepare, if given, runs after the first snapshot and
// before the command; what it returns runs before the second snapshot.
func manifestIn(t *testing.T, base, cwd, root string, prepare func() func()) run {
	t.Helper()
	before := snapshotAll(t, base)
	var undo func()
	if prepare != nil {
		undo = prepare()
	}
	r := execProviso(t, cwd, "manifest")
	if undo != nil {
		undo()
	}
	after := snapshotAll(t, base)
	var allowed []string
	if r.exit == 0 && root != "" {
		for _, name := range []string{"manifest.sum", coherenceName} {
			rel, _ := filepath.Rel(base, filepath.Join(root, name))
			allowed = append(allowed, filepath.ToSlash(rel))
		}
		// baseline.manifest.write: a coherence file is removed, not
		// written.
		rel, _ := filepath.Rel(base, filepath.Join(root, coherenceName))
		if _, ok := after[filepath.ToSlash(rel)]; ok {
			t.Errorf("baseline.manifest.write: manifest exited with 0, but %s is still in the spec root", coherenceName)
		}
	}
	checkWrites(t, before, after, allowed...)
	if r.exit != 0 && r.exit != 2 {
		t.Fatalf("system.cli.manifest.exit-status: %d is not 0 or 2\nstderr: %s", r.exit, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("system.cli.manifest: manifest writes nothing to standard output, got: %q", r.stdout)
	}
	if r.exit == 2 && strings.TrimSpace(r.stderr) == "" {
		t.Errorf("exit status 2 with nothing on stderr")
	}
	if r.exit == 0 && root != "" {
		b, err := os.ReadFile(filepath.Join(root, "manifest.sum"))
		if err != nil {
			t.Fatalf("baseline.manifest.write: exit status 0, but no manifest.sum in the spec root: %v", err)
		}
		checkManifestFormat(t, string(b))
	}
	return r
}

// manifestOK runs `manifest` in the directory holding proviso/ and returns
// the manifest it wrote.
func manifestOK(t *testing.T, dir string) string {
	t.Helper()
	root := filepath.Join(dir, "proviso")
	r := manifestIn(t, dir, dir, root, nil)
	if r.exit != 0 {
		t.Fatalf("want manifest to exit with 0, got %d\nstderr: %s", r.exit, r.stderr)
	}
	b, _ := os.ReadFile(filepath.Join(root, "manifest.sum"))
	return string(b)
}

// manifestOf writes tr into a fresh directory, runs `manifest` there and
// returns the manifest.
func manifestOf(t *testing.T, tr tree) string {
	t.Helper()
	return manifestOK(t, tr.write(t))
}

// ---------------------------------------------------------------------------
// The element diff

type partChange struct {
	Part   string  `json:"part"`
	Change string  `json:"change"`
	Former *string `json:"former"`
}

type diffElement struct {
	ID        string       `json:"id"`
	Change    string       `json:"change"`
	File      *string      `json:"file"`
	Line      *int         `json:"line"`
	Text      *string      `json:"text"`
	Former    *string      `json:"former"`
	Parts     []partChange `json:"parts"`
	Through   [][]string   `json:"through"`
	Undecided []place      `json:"undecided"`
}

type elementDiff struct {
	Root     string        `json:"root"`
	Revision *string       `json:"revision"`
	Elements []diffElement `json:"elements"`
}

// diffOutcome is one diff of a tree: `diff` run once with `--json` and once
// without, from the same directory.
type diffOutcome struct {
	json, text run
	ed         elementDiff
}

// diffIn runs `proviso diff --json` and `proviso diff` in cwd and checks what
// holds for every run. base is the directory the test tree was written to;
// nothing below it may change. root, if not empty, is the spec root the diff
// is expected to compare. args, such as --since and a revision, are given to
// both runs.
func diffIn(t *testing.T, base, cwd, root string, args ...string) diffOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	o := diffOutcome{
		json: execProviso(t, cwd, append([]string{"diff", "--json"}, args...)...),
		text: execProviso(t, cwd, append([]string{"diff"}, args...)...),
	}
	after := snapshotAll(t, base)
	checkWrites(t, before, after, "")
	if o.json.exit != o.text.exit {
		t.Errorf("exit status %d with --json, %d without", o.json.exit, o.text.exit)
	}
	for _, r := range []run{o.json, o.text} {
		if r.exit < 0 || r.exit > 2 {
			t.Fatalf("system.cli.diff.exit-status: %d is not 0, 1 or 2\nstderr: %s", r.exit, r.stderr)
		}
		if r.exit == 2 && strings.TrimSpace(r.stderr) == "" {
			t.Errorf("exit status 2 with nothing on stderr")
		}
	}
	if o.json.exit == 2 {
		for _, r := range []run{o.json, o.text} {
			if r.stdout != "" {
				t.Errorf("exit status 2, but stdout holds: %q", r.stdout)
			}
		}
		return o
	}
	o.ed = parseElementDiff(t, o.json.stdout)
	if root != "" {
		checkRoot(t, outcome{report: report{Root: o.ed.Root}}, root)
	}
	checkDiffOrder(t, o.ed.Elements)
	// baseline.diff.exit
	want := 0
	if len(o.ed.Elements) > 0 {
		want = 1
	}
	if o.json.exit != want {
		t.Errorf("baseline.diff.exit: exit status %d, want %d for %d listed elements", o.json.exit, want, len(o.ed.Elements))
	}
	// baseline.diff.output: nothing written as text when nothing is listed.
	if len(o.ed.Elements) == 0 && o.text.stdout != "" {
		t.Errorf("baseline.diff.output: no elements, but text output %q", o.text.stdout)
	}
	if len(o.ed.Elements) > 0 && strings.TrimSpace(o.text.stdout) == "" {
		t.Errorf("baseline.diff.output: %d elements, but no text output", len(o.ed.Elements))
	}
	return o
}

var diffChanges = map[string]bool{"added": true, "changed": true, "removed": true, "changed-through": true}
var partChanges = map[string]bool{"added": true, "changed": true, "removed": true}

// parseElementDiff checks stdout against the schema of `element_diff` and
// that nothing else is on stdout.
func parseElementDiff(t *testing.T, stdout string) elementDiff {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Fatalf("system.cli.element-diff: stdout is not one JSON object: %v\n%s", err, stdout)
	}
	var s string
	if raw["root"] == nil || json.Unmarshal(raw["root"], &s) != nil {
		t.Errorf("system.cli.element-diff: root missing or not a string: %s", raw["root"])
	}
	if rev, ok := raw["revision"]; !ok {
		t.Errorf("system.cli.element-diff: revision missing")
	} else if string(bytes.TrimSpace(rev)) != "null" && json.Unmarshal(rev, &s) != nil {
		t.Errorf("system.cli.element-diff: revision is neither a string nor null: %s", rev)
	}
	if r := bytes.TrimSpace(raw["elements"]); len(r) == 0 || r[0] != '[' {
		t.Fatalf("system.cli.element-diff: elements missing or not an array: %s", raw["elements"])
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw["elements"], &items); err != nil {
		t.Fatalf("system.cli.element-diff: elements are not objects: %v", err)
	}
	isLine := func(m json.RawMessage) bool {
		var n int
		return json.Unmarshal(m, &n) == nil && n >= 1
	}
	for i, it := range items {
		for _, k := range []string{"id", "change", "undecided"} {
			if _, ok := it[k]; !ok {
				t.Errorf("system.cli.element-diff: element %d has no %s", i, k)
			}
		}
		for _, k := range []string{"id", "change", "file", "text", "former"} {
			if v, ok := it[k]; ok && json.Unmarshal(v, &s) != nil {
				t.Errorf("system.cli.element-diff: element %d: %s is not a string: %s", i, k, v)
			}
		}
		if json.Unmarshal(it["change"], &s) == nil && !diffChanges[s] {
			t.Errorf("system.cli.element-diff: element %d: change %q is not added, changed, removed or changed-through", i, s)
		}
		if v, ok := it["line"]; ok && !isLine(v) {
			t.Errorf("system.cli.element-diff: element %d: line %s is not an integer of at least 1", i, v)
		}
		if v, ok := it["parts"]; ok {
			var parts []map[string]json.RawMessage
			if json.Unmarshal(v, &parts) != nil {
				t.Errorf("system.cli.element-diff: element %d: parts is not an array of objects", i)
			}
			for _, p := range parts {
				if json.Unmarshal(p["part"], &s) != nil {
					t.Errorf("system.cli.element-diff: element %d: a part has no part string", i)
				}
				if json.Unmarshal(p["change"], &s) != nil || !partChanges[s] {
					t.Errorf("system.cli.element-diff: element %d: part change %s is not added, changed or removed", i, p["change"])
				}
				if f, ok := p["former"]; ok && json.Unmarshal(f, &s) != nil {
					t.Errorf("system.cli.element-diff: element %d: part former is not a string", i)
				}
			}
		}
		if v, ok := it["through"]; ok {
			var through [][]string
			if json.Unmarshal(v, &through) != nil {
				t.Errorf("system.cli.element-diff: element %d: through is not an array of arrays of strings", i)
			}
		}
		if v := bytes.TrimSpace(it["undecided"]); len(v) == 0 || v[0] != '[' {
			t.Errorf("system.cli.element-diff: element %d: undecided is not an array: %s", i, it["undecided"])
		} else {
			var places []map[string]json.RawMessage
			if json.Unmarshal(v, &places) != nil {
				t.Errorf("system.cli.element-diff: element %d: undecided is not an array of objects", i)
			}
			for _, p := range places {
				if json.Unmarshal(p["file"], &s) != nil || !isLine(p["line"]) {
					t.Errorf("system.cli.element-diff: element %d: undecided place %v needs a file and a line of at least 1", i, p)
				}
			}
		}
	}
	var ed elementDiff
	if err := json.Unmarshal([]byte(stdout), &ed); err != nil {
		t.Fatalf("system.cli.element-diff: %v", err)
	}
	for _, e := range ed.Elements {
		paths := []string{}
		if e.File != nil {
			paths = append(paths, *e.File)
		}
		for _, u := range e.Undecided {
			paths = append(paths, u.File)
		}
		for _, p := range paths {
			if strings.Contains(p, `\`) {
				t.Errorf("system.cli.element-diff: path %q does not use / as its separator", p)
			}
		}
		// A removed element has neither file nor line, nor text.
		if e.Change == "removed" && (e.File != nil || e.Line != nil || e.Text != nil) {
			t.Errorf("system.cli.element-diff: removed element %s has a file, line or text", e.ID)
		}
		if e.Change != "removed" && (e.File == nil || e.Line == nil) {
			t.Errorf("system.cli.element-diff: %s element %s has no file or line", e.Change, e.ID)
		}
		if e.Change != "removed" && e.Former != nil {
			t.Errorf("system.cli.element-diff: former is for a removed element only, but %s element %s has one", e.Change, e.ID)
		}
	}
	return ed
}

// checkDiffOrder checks baseline.diff.order: elements by identifier, byte by
// byte; the parts of each in the order of their lines in manifest_file, that
// is by the part as written, byte by byte, since every line of one element
// starts with the same identifier and a tab.
func checkDiffOrder(t *testing.T, list []diffElement) {
	t.Helper()
	for i := 1; i < len(list); i++ {
		if list[i-1].ID >= list[i].ID {
			t.Errorf("baseline.diff.order: %s before %s", list[i-1].ID, list[i].ID)
		}
	}
	for _, e := range list {
		for i := 1; i < len(e.Parts); i++ {
			if e.Parts[i-1].Part+"\t" >= e.Parts[i].Part+"\t" {
				t.Errorf("baseline.diff.order: in %s, part %q before %q", e.ID, e.Parts[i-1].Part, e.Parts[i].Part)
			}
		}
	}
}

func (o diffOutcome) dump() string {
	var b strings.Builder
	if len(o.ed.Elements) == 0 {
		b.WriteString("  (no elements)\n")
	}
	for _, e := range o.ed.Elements {
		var parts []string
		for _, p := range e.Parts {
			parts = append(parts, p.Part+":"+p.Change)
		}
		fmt.Fprintf(&b, "  %s %s parts=%v through=%v undecided=%v\n", e.Change, e.ID, parts, e.Through, e.Undecided)
	}
	return b.String()
}

// ran asserts that diff ran, exiting with 0 or 1.
func (o diffOutcome) ran(t *testing.T) {
	t.Helper()
	if o.json.exit == 2 {
		t.Fatalf("want diff to run, but it exited with 2:\n%s", o.json.stderr)
	}
}

// stopped asserts exit status 2 for both runs; that nothing went to stdout
// is checked for every run with exit status 2.
func (o diffOutcome) stopped(t *testing.T) {
	t.Helper()
	for _, r := range []run{o.json, o.text} {
		if r.exit != 2 {
			t.Errorf("want exit status 2, got %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
		}
	}
}

func (o diffOutcome) elem(t *testing.T, id string) diffElement {
	t.Helper()
	o.ran(t)
	for _, e := range o.ed.Elements {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("want %s listed, got:\n%s", id, o.dump())
	return diffElement{}
}

func (o diffOutcome) lacksElem(t *testing.T, id string) {
	t.Helper()
	o.ran(t)
	for _, e := range o.ed.Elements {
		if e.ID == id {
			t.Errorf("want %s not listed, got:\n%s", id, o.dump())
		}
	}
}

// only asserts that exactly the elements ids are listed, each with the
// change given after a colon, as in "cart.add.limit:changed".
func (o diffOutcome) only(t *testing.T, want ...string) {
	t.Helper()
	o.ran(t)
	var got []string
	for _, e := range o.ed.Elements {
		got = append(got, e.ID+":"+e.Change)
	}
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("want listed %v, got:\n%s", want, o.dump())
	}
}

// part returns the part named name of e, failing if it is not named.
func part(t *testing.T, e diffElement, name string) partChange {
	t.Helper()
	for _, p := range e.Parts {
		if p.Part == name {
			return p
		}
	}
	t.Errorf("want %s to name the part %q, got %v", e.ID, name, e.Parts)
	return partChange{}
}

// onlyParts asserts that e names exactly the parts given, each with its
// change after a colon, as in "step Then:changed".
func onlyParts(t *testing.T, e diffElement, want ...string) {
	t.Helper()
	var got []string
	for _, p := range e.Parts {
		got = append(got, p.Part+":"+p.Change)
	}
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("want %s to name the parts %q, got %q", e.ID, want, got)
	}
}

// ---------------------------------------------------------------------------
// Git repositories

// repo is a git repository in a temporary directory holding a test tree.
// Its git commands run with a fixed user and dates, and without the
// machine's own git configuration.
type repo struct {
	t    *testing.T
	dir  string
	cfg  string
	tick int
	tr   tree
}

// newRepo writes tr into a new git repository on the branch main, without
// committing anything.
func newRepo(t *testing.T, tr tree) *repo {
	t.Helper()
	return newRepoIn(t, t.TempDir(), tr)
}

// newRepoIn is newRepo in the empty directory dir.
func newRepoIn(t *testing.T, dir string, tr tree) *repo {
	t.Helper()
	skipWithoutGit(t)
	r := &repo{t: t, dir: dir, tr: tree{}}
	r.cfg = filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(r.cfg, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.name", "Acceptance Test")
	r.git("config", "user.email", "acceptance@example.com")
	r.git("config", "core.autocrlf", "false")
	r.set(tr)
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	r.tick++
	date := fmt.Sprintf("%d +0000", 1700000000+60*r.tick)
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+r.cfg,
		"GIT_AUTHOR_NAME=Acceptance Test", "GIT_AUTHOR_EMAIL=acceptance@example.com",
		"GIT_COMMITTER_NAME=Acceptance Test", "GIT_COMMITTER_EMAIL=acceptance@example.com",
		"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit commits everything in the working tree and returns the commit's
// full name.
func (r *repo) commit(msg string) string {
	r.t.Helper()
	r.git("add", "-A")
	r.git("commit", "-q", "--allow-empty", "-m", msg)
	return r.git("rev-parse", "HEAD")
}

// set makes the working tree hold tr: files of the former tree that tr does
// not hold are removed, with folders left empty, and manifest.sum is kept.
func (r *repo) set(tr tree) {
	r.t.Helper()
	for p := range r.tr {
		if _, ok := tr[p]; !ok {
			full := filepath.Join(r.dir, filepath.FromSlash(p))
			os.Remove(full)
			for d := filepath.Dir(full); d != r.dir; d = filepath.Dir(d) {
				if os.Remove(d) != nil {
					break
				}
			}
		}
	}
	tr.writeTo(r.t, r.dir, false)
	r.tr = tr.clone()
}

func (r *repo) root() string { return filepath.Join(r.dir, "proviso") }

// manifest runs `proviso manifest` in the repository, which must succeed.
func (r *repo) manifest() string {
	r.t.Helper()
	return manifestOK(r.t, r.dir)
}

// diff runs `proviso diff` in the repository.
func (r *repo) diff() diffOutcome {
	r.t.Helper()
	return diffIn(r.t, r.dir, r.dir, r.root())
}

func skipWithoutGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

// accepted is tr accepted: written into a new repository, its manifest
// written by `proviso manifest`, and both committed. It returns the
// repository and the commit. Each tree is accepted once, by the first test
// that asks for it, in a folder of its own below templates; every test gets
// a copy of that folder in its own temporary directory.
func accepted(t *testing.T, tr tree) (*repo, string) {
	t.Helper()
	skipWithoutGit(t)
	v, _ := acceptedRepos.LoadOrStore(tr.key(), &acceptedRepo{})
	a := v.(*acceptedRepo)
	a.once.Do(func() {
		dir, err := os.MkdirTemp(templates, "accepted-")
		if err != nil {
			t.Fatal(err)
		}
		r := newRepoIn(t, dir, tr)
		r.manifest()
		a.commit = r.commit("accept")
		a.dir, a.tick, a.built = dir, r.tick, true
	})
	if !a.built {
		t.Fatal("accepting the tree failed in another test")
	}
	r := &repo{t: t, dir: t.TempDir(), tick: a.tick, tr: tr.clone()}
	copyDir(t, a.dir, r.dir)
	r.cfg = filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(r.cfg, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return r, a.commit
}

// acceptedRepo is a tree accepted once: its repository, the git commands
// run in it so far and the commit.
type acceptedRepo struct {
	once   sync.Once
	built  bool
	dir    string
	tick   int
	commit string
}

// acceptedRepos maps the key of each tree accepted to its *acceptedRepo.
var acceptedRepos sync.Map

// key identifies the tree by its paths and contents.
func (tr tree) key() string {
	paths := make([]string, 0, len(tr))
	for p := range tr {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, p := range paths {
		fmt.Fprintf(&b, "%d:%s%d:%s", len(p), p, len(tr[p]), tr[p])
	}
	return b.String()
}

// copyDir copies the folders and regular files below src into dst.
func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, info.Mode().Perm())
	})
	if err != nil {
		t.Fatalf("copying the accepted repository: %v", err)
	}
}

// skipIfGitAbove skips a test that needs dir to be in no git repository when
// a folder above it is in one.
func skipIfGitAbove(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		t.Skipf("%s is inside the git repository %s", dir, strings.TrimSpace(string(out)))
	}
}

// ---------------------------------------------------------------------------
// The manifest oracle, from manifest_file and part_kinds

// normWS makes each run of spaces, tabs and line feeds one space and leaves
// none at the start or end. Nothing else counts as whitespace.
func normWS(s string) string {
	var b strings.Builder
	space := false
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' {
			space = true
			continue
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(c)
	}
	return b.String()
}

// fp is the fingerprint of text: h1: and the SHA-256 of its normalised
// text in lowercase hexadecimal.
func fp(text string) string {
	sum := sha256.Sum256([]byte(normWS(md(text))))
	return "h1:" + hex.EncodeToString(sum[:])
}

// rowKey is the key of a row with the placeholder cells given as name,
// value pairs: for each column, in the byte order of the names, the name, =
// and the cell, joined by tabs, and fingerprinted.
func rowKey(pairs ...string) string {
	var cells []string
	for i := 0; i+1 < len(pairs); i += 2 {
		cells = append(cells, pairs[i]+"="+pairs[i+1])
	}
	sort.Slice(cells, func(i, j int) bool {
		return strings.SplitN(cells[i], "=", 2)[0] < strings.SplitN(cells[j], "=", 2)[0]
	})
	return fp(strings.Join(cells, "\t"))
}

// mpart is a part of an element with its text, for the oracle.
type mpart struct{ id, part, text string }

// manifestFor writes manifest_file for parts, as the contract describes it.
func manifestFor(parts []mpart) string {
	var lines []string
	for _, p := range parts {
		lines = append(lines, p.id+"\t"+p.part+"\t"+fp(p.text))
	}
	sort.Strings(lines)
	return "model 0\n" + strings.Join(lines, "\n") + "\n"
}

var (
	idPattern   = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)*$`)
	partPattern = regexp.MustCompile(`^(heading|body|field [^\t]+|section [^\t]+|step (Given|When|Provided|Then|Otherwise)|row h1:[0-9a-f]{64}( #([2-9]|[1-9][0-9]+))?)$`)
	fpPattern   = regexp.MustCompile(`^h1:[0-9a-f]{64}$`)
)

// checkManifestFormat checks manifest_file: UTF-8 text, every line ending in
// a line feed, the first naming model_version, then one line per part with
// three fields separated by tabs, sorted byte by byte.
func checkManifestFormat(t *testing.T, content string) {
	t.Helper()
	if !strings.HasSuffix(content, "\n") {
		t.Errorf("system.manifest-file: the last line does not end with a line feed")
	}
	if strings.Contains(content, "\r") {
		t.Errorf("system.manifest-file: holds a carriage return")
	}
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	if lines[0] != "model 0" {
		t.Errorf("system.manifest-file: first line %q, want %q", lines[0], "model 0")
	}
	for i, l := range lines[1:] {
		f := strings.Split(l, "\t")
		if len(f) != 3 || !idPattern.MatchString(f[0]) || !partPattern.MatchString(f[1]) || !fpPattern.MatchString(f[2]) {
			t.Errorf("system.manifest-file: line %d is not an identifier, a part and a fingerprint separated by tabs: %q", i+2, l)
		}
		if i > 0 && lines[i] >= l {
			t.Errorf("system.manifest-file: line %d %q does not sort after %q", i+2, l, lines[i])
		}
	}
}

// manifestParts reads the parts of a manifest: identifier and part to
// fingerprint.
func manifestParts(content string) map[[2]string]string {
	m := map[[2]string]string{}
	for i, l := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
		if i == 0 {
			continue
		}
		if f := strings.Split(l, "\t"); len(f) == 3 {
			m[[2]string{f[0], f[1]}] = f[2]
		}
	}
	return m
}

// partsOf lists the parts the manifest records for id.
func partsOf(content, id string) []string {
	var parts []string
	for k := range manifestParts(content) {
		if k[0] == id {
			parts = append(parts, k[1])
		}
	}
	sort.Strings(parts)
	return parts
}

// stepFPs gives the fingerprints a step's text may have: part_kinds says a
// part's text is what follows its key, and a step's lines with its And
// lines, so the keyword may or may not be part of it (see the report).
func stepFPs(lines string) []string {
	n := normWS(md(lines))
	_, rest, _ := strings.Cut(n, " ")
	return []string{fp(n), fp(rest)}
}

// ---------------------------------------------------------------------------
// Element text

// elementText gives the text of the element id in content, from its heading
// to the next element's heading, and the line of its heading.
func elementText(content, id string) (string, int) {
	content = md(content)
	lines := strings.SplitAfter(content, "\n")
	start := -1
	var b strings.Builder
	for i, l := range lines {
		isElement := strings.HasPrefix(l, "#") && strings.Contains(l, `<a id="`)
		if start < 0 {
			if isElement && strings.Contains(l, `<a id="`+id+`"></a>`) {
				start = i
				b.WriteString(l)
			}
			continue
		}
		if isElement {
			break
		}
		b.WriteString(l)
	}
	if start < 0 {
		panic("elementText: no element " + id)
	}
	return b.String(), start + 1
}

// sameText compares two texts up to whitespace at their ends.
func sameText(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}
