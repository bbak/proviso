package acceptance

// The driver: it builds the binary, writes test trees, runs lint and checks
// what holds for every run (the report's shape, the order of findings, the
// exit status and the read-only invariant), so that the tests read in the
// spec's terms.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "proviso-acceptance-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	name := "proviso"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary = filepath.Join(dir, name)
	build := exec.Command("go", "build", "-o", binary, "github.com/bbak/proviso/cmd/proviso")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building proviso: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// Kinds of finding, from contracts/findings.md, with their severity as the
// rules report them: `ask` where a rule reports an [ask], `error` otherwise.
var severities = map[string]string{
	"anchor-misplaced":             "error",
	"identifier-malformed":         "error",
	"identifier-duplicate":         "error",
	"system-identifier":            "error",
	"file-unknown":                 "error",
	"name-malformed":               "error",
	"line-ending":                  "error",
	"capability-file-missing":      "error",
	"title-missing":                "error",
	"element-unplaced":             "error",
	"heading-unknown":              "error",
	"section-missing":              "error",
	"section-repeated":             "error",
	"sibling-empty":                "error",
	"rule-block-count":             "error",
	"keyword-unknown":              "error",
	"keyword-order":                "error",
	"keyword-count":                "error",
	"guard-incomplete":             "error",
	"hand-over-and":                "error",
	"and-missing":                  "error",
	"given-refers-to-event":        "error",
	"provided-without-event":       "ask",
	"placeholder-unbound":          "error",
	"examples-missing":             "error",
	"placeholder-reserved":         "error",
	"examples-column-missing":      "error",
	"provided-column":              "error",
	"examples-column-unknown":      "error",
	"examples-empty":               "error",
	"examples-row-incomplete":      "error",
	"provided-value-unknown":       "error",
	"examples-side-missing":        "error",
	"examples-part-missing":        "error",
	"term-unresolved":              "error",
	"term-unmarked":                "error",
	"term-ambiguous":               "error",
	"hand-over-unresolved":         "error",
	"hand-over-cycle":              "error",
	"reference-unresolved":         "error",
	"reference-ambiguous":          "error",
	"field-unknown":                "error",
	"field-missing":                "error",
	"field-repeated":               "error",
	"status-unknown":               "error",
	"open-entry-incomplete":        "error",
	"defined-entry-open":           "error",
	"definition-count":             "error",
	"defined-by-unresolved":        "error",
	"deviations-without-reference": "error",
	"undecided":                    "ask",
	"element-missing":              "error",
	"section-level":                "error",
}

// ---------------------------------------------------------------------------
// Test trees

// md turns the acute accent into a backtick, since Go raw strings cannot hold
// backticks.
func md(s string) string { return strings.ReplaceAll(s, "´", "`") }

// tree maps slash-separated paths, relative to the repository directory that
// holds proviso/, to file contents.
type tree map[string]string

func (tr tree) clone() tree {
	c := tree{}
	for k, v := range tr {
		c[k] = v
	}
	return c
}

// with adds or replaces files, given as path, content pairs.
func (tr tree) with(pairs ...string) tree {
	c := tr.clone()
	for i := 0; i+1 < len(pairs); i += 2 {
		c[pairs[i]] = md(pairs[i+1])
	}
	return c
}

func (tr tree) without(paths ...string) tree {
	c := tr.clone()
	for _, p := range paths {
		delete(c, p)
	}
	return c
}

// edit replaces old, which must occur exactly once, by new in path.
func (tr tree) edit(path, old, new string) tree {
	old, new = md(old), md(new)
	content, ok := tr[path]
	if !ok {
		panic("edit: no file " + path)
	}
	if n := strings.Count(content, old); n != 1 {
		panic(fmt.Sprintf("edit: %q occurs %d times in %s", old, n, path))
	}
	c := tr.clone()
	c[path] = strings.Replace(content, old, new, 1)
	return c
}

// appendTo adds text to the end of path.
func (tr tree) appendTo(path, text string) tree {
	c := tr.clone()
	c[path] = tr[path] + md(text)
	return c
}

// prefixed moves every file below dir.
func (tr tree) prefixed(dir string) tree {
	c := tree{}
	for k, v := range tr {
		c[dir+"/"+k] = v
	}
	return c
}

// line gives the 1-based line of the n-th (1-based) line of path holding sub.
func (tr tree) lineN(path, sub string, n int) int {
	sub = md(sub)
	for i, l := range strings.Split(tr[path], "\n") {
		if strings.Contains(l, sub) {
			n--
			if n == 0 {
				return i + 1
			}
		}
	}
	panic(fmt.Sprintf("line: %q not found in %s", sub, path))
}

func (tr tree) line(path, sub string) int { return tr.lineN(path, sub, 1) }

// fence gives the lines of the opening and closing fence of the n-th
// guarded block of path.
func (tr tree) fence(path string, n int) (int, int) {
	lines := strings.Split(tr[path], "\n")
	open := 0
	for i, l := range lines {
		if open == 0 && strings.HasPrefix(l, "```guarded") {
			n--
			if n == 0 {
				open = i + 1
			}
			continue
		}
		if open != 0 && strings.TrimSpace(l) == "```" {
			return open, i + 1
		}
	}
	panic("fence: no guarded block in " + path)
}

// table gives the first and last line of the first table at or after line
// from of path.
func (tr tree) table(path string, from int) (int, int) {
	lines := strings.Split(tr[path], "\n")
	first := 0
	for i := from - 1; i < len(lines); i++ {
		isRow := strings.HasPrefix(lines[i], "|")
		if isRow && first == 0 {
			first = i + 1
		}
		if !isRow && first != 0 {
			return first, i
		}
	}
	if first != 0 {
		return first, len(lines)
	}
	panic("table: no table in " + path)
}

func (tr tree) write(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	tr.writeTo(t, dir, false)
	return dir
}

func (tr tree) writeTo(t *testing.T, dir string, reverse bool) {
	t.Helper()
	paths := make([]string, 0, len(tr))
	for p := range tr {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	if reverse {
		for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
			paths[i], paths[j] = paths[j], paths[i]
		}
	}
	for _, p := range paths {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(tr[p]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// Running lint

type run struct {
	exit           int
	stdout, stderr string
}

type place struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type finding struct {
	Kind     string  `json:"kind"`
	Severity string  `json:"severity"`
	File     string  `json:"file"`
	Line     int     `json:"line"`
	Message  string  `json:"message"`
	Related  []place `json:"related"`
}

type report struct {
	Root     string    `json:"root"`
	Findings []finding `json:"findings"`
}

// outcome is one check of a tree: lint run once with `--json` and once
// without, from the same directory.
type outcome struct {
	json, text run
	report     report
}

func execLint(t *testing.T, cwd string, args ...string) run {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, append([]string{"lint"}, args...)...)
	cmd.Dir = cwd
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || ctx.Err() != nil {
			t.Fatalf("running proviso lint in %s: %v", cwd, err)
		}
		code = exitErr.ExitCode()
	}
	return run{code, out.String(), errOut.String()}
}

// check writes tr and runs lint in the directory holding proviso/.
func check(t *testing.T, tr tree) outcome {
	t.Helper()
	dir := tr.write(t)
	o := lintIn(t, dir, dir, nil)
	if o.json.exit != 2 {
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	}
	return o
}

// checkRoot asserts that the report's root is the absolute path of the spec
// root want, written with / as its separator and, on Windows, its drive
// first, as an uppercase letter, as in C:/repo/proviso.
func checkRoot(t *testing.T, o outcome, want string) {
	t.Helper()
	if strings.Contains(o.report.Root, `\`) {
		t.Errorf("lint.cli.report: root %q does not use / as its separator", o.report.Root)
	}
	if runtime.GOOS == "windows" && len(filepath.VolumeName(want)) == 2 && !hasDrive(o.report.Root) {
		t.Errorf("lint.cli.report: root %q does not start with its drive, as an uppercase letter, as in C:/repo/proviso", o.report.Root)
	}
	got := filepath.FromSlash(o.report.Root)
	if !filepath.IsAbs(got) {
		t.Errorf("lint.cli.report: root %q is not an absolute path", o.report.Root)
		return
	}
	canon := func(p string) string {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			p = r
		}
		p = filepath.Clean(p)
		if runtime.GOOS == "windows" {
			p = strings.ToLower(p)
		}
		return p
	}
	if canon(got) != canon(want) {
		t.Errorf("lint.cli.report: root %q, want the spec root %q", o.report.Root, want)
	}
}

// hasDrive tells whether p starts with an uppercase drive letter, a colon and
// a /.
func hasDrive(p string) bool {
	if len(p) < 3 || p[1] != ':' || p[2] != '/' {
		return false
	}
	return p[0] >= 'A' && p[0] <= 'Z'
}

// sentinels are files the layout does not list, one sorting before and one
// after the files a test is about. Reporting both shows that lint went on
// with the rest of the spec tree.
var sentinels = []string{"proviso/0-sentinel.md", "proviso/zz-sentinel.md"}

func checkGoesOn(t *testing.T, tr tree) outcome {
	t.Helper()
	return checkGoesOnAfter(t, tr, nil)
}

// checkGoesOnAfter is checkGoesOn, with setup, if given, run on the directory
// the tree was written to before lint runs.
func checkGoesOnAfter(t *testing.T, tr tree, setup func(dir string)) outcome {
	t.Helper()
	dir := tr.with(sentinels[0], "# Sentinel\n", sentinels[1], "# Sentinel\n").write(t)
	if setup != nil {
		setup(dir)
	}
	o := lintIn(t, dir, dir, nil)
	if o.json.exit != 2 {
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	}
	for _, s := range sentinels {
		o.has(t, "file-unknown", s)
	}
	return o
}

// lintIn runs lint in cwd. base is the temporary directory the test tree was
// written to; every spec root below it must be unchanged by the run.
// prepare, if given, runs after the snapshot and before lint; the function it
// returns runs after lint and before the second snapshot.
func lintIn(t *testing.T, base, cwd string, prepare func() func()) outcome {
	t.Helper()
	before := snapshot(t, base)
	var undo func()
	if prepare != nil {
		undo = prepare()
	}
	o := outcome{json: execLint(t, cwd, "--json"), text: execLint(t, cwd)}
	if undo != nil {
		undo()
	}
	after := snapshot(t, base)
	checkReadOnly(t, before, after)
	checkRun(t, &o)
	return o
}

// checkRun checks what holds for every run.
func checkRun(t *testing.T, o *outcome) {
	t.Helper()
	if o.json.exit != o.text.exit {
		t.Errorf("exit status %d with --json, %d without", o.json.exit, o.text.exit)
	}
	for _, r := range []run{o.json, o.text} {
		if r.exit < 0 || r.exit > 2 {
			t.Fatalf("lint.cli.lint.exit-status: %d is not 0, 1 or 2\nstderr: %s", r.exit, r.stderr)
		}
		if r.exit == 2 && strings.TrimSpace(r.stderr) == "" {
			t.Errorf("exit status 2 with nothing on stderr")
		}
	}
	if o.json.exit == 2 {
		// lint.run.root, lint.run.unreadable: nothing on stdout, not
		// even an empty report.
		for _, r := range []run{o.json, o.text} {
			if r.stdout != "" {
				t.Errorf("exit status 2, but stdout holds: %q", r.stdout)
			}
		}
		return
	}
	o.report = parseReport(t, o.json.stdout)
	for _, f := range o.report.Findings {
		paths := []string{f.File}
		for _, r := range f.Related {
			paths = append(paths, r.File)
		}
		for _, p := range paths {
			if strings.Contains(p, `\`) {
				t.Errorf("lint.cli.report: path %q does not use / as its separator", p)
			}
			if strings.HasSuffix(p, "/") {
				t.Errorf("lint.cli.report: path %q of a folder has a trailing /", p)
			}
		}
	}
	checkOrder(t, o.report.Findings)
	errorsFound := false
	for _, f := range o.report.Findings {
		sev, known := severities[f.Kind]
		if !known {
			t.Errorf("lint.cli.report: kind %q is no member of findings", f.Kind)
		} else if f.Severity != sev {
			t.Errorf("finding %s at %s:%d has severity %q, the rules report it as %q", f.Kind, f.File, f.Line, f.Severity, sev)
		}
		if f.Severity == "error" {
			errorsFound = true
		}
	}
	want := 0
	if errorsFound {
		want = 1
	}
	if o.json.exit != want {
		t.Errorf("lint.run.exit: exit status %d, want %d for findings:\n%s", o.json.exit, want, dump(o.report.Findings))
	}
	// lint.run.output: text, one finding per line, nothing for no findings.
	n := len(o.report.Findings)
	text := strings.TrimRight(strings.ReplaceAll(o.text.stdout, "\r\n", "\n"), "\n")
	if n == 0 {
		if o.text.stdout != "" {
			t.Errorf("lint.run.output: no findings, but text output %q", o.text.stdout)
		}
	} else {
		lines := strings.Split(text, "\n")
		if len(lines) != n {
			t.Errorf("lint.run.output: %d findings, but %d lines of text:\n%s", n, len(lines), o.text.stdout)
		}
	}
}

// parseReport checks stdout against the schema of `report` and that nothing
// else is on stdout.
func parseReport(t *testing.T, stdout string) report {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Fatalf("lint.cli.report: stdout is not one JSON object: %v\n%s", err, stdout)
	}
	var root string
	if err := json.Unmarshal(raw["root"], &root); err != nil || raw["root"] == nil {
		t.Errorf("lint.cli.report: root missing or not a string: %s", raw["root"])
	}
	var items []map[string]json.RawMessage
	if r := bytes.TrimSpace(raw["findings"]); len(r) == 0 || r[0] != '[' {
		t.Fatalf("lint.cli.report: findings missing or not an array: %s", raw["findings"])
	}
	if err := json.Unmarshal(raw["findings"], &items); err != nil {
		t.Fatalf("lint.cli.report: findings are not objects: %v", err)
	}
	for i, it := range items {
		for _, k := range []string{"kind", "severity", "file", "line", "message"} {
			if _, ok := it[k]; !ok {
				t.Errorf("lint.cli.report: finding %d has no %s: %v", i, k, it)
			}
		}
		var s string
		for _, k := range []string{"kind", "file", "message"} {
			if json.Unmarshal(it[k], &s) != nil {
				t.Errorf("lint.cli.report: finding %d: %s is not a string", i, k)
			}
		}
		if json.Unmarshal(it["severity"], &s) != nil || (s != "error" && s != "ask") {
			t.Errorf("lint.cli.report: finding %d: severity %s is not error or ask", i, it["severity"])
		}
		var line int
		if json.Unmarshal(it["line"], &line) != nil || line < 1 {
			t.Errorf("lint.cli.report: finding %d: line %s is not an integer of at least 1", i, it["line"])
		}
		if rel, ok := it["related"]; ok {
			var places []map[string]json.RawMessage
			if json.Unmarshal(rel, &places) != nil {
				t.Errorf("lint.cli.report: finding %d: related is not an array of objects", i)
			}
			for _, p := range places {
				if json.Unmarshal(p["file"], &s) != nil {
					t.Errorf("lint.cli.report: finding %d: related file is not a string", i)
				}
				if json.Unmarshal(p["line"], &line) != nil || line < 1 {
					t.Errorf("lint.cli.report: finding %d: related line is not an integer of at least 1", i)
				}
			}
		}
	}
	var rep report
	if err := json.Unmarshal([]byte(stdout), &rep); err != nil {
		t.Fatalf("lint.cli.report: %v", err)
	}
	if rep.Root == "" {
		t.Errorf("lint.cli.report: root is empty")
	}
	return rep
}

// checkOrder checks lint.run.order: by file path byte by byte, then line,
// then kind.
func checkOrder(t *testing.T, list []finding) {
	t.Helper()
	for i := 1; i < len(list); i++ {
		a, b := list[i-1], list[i]
		if a.File < b.File || (a.File == b.File && (a.Line < b.Line || (a.Line == b.Line && a.Kind <= b.Kind))) {
			continue
		}
		t.Errorf("lint.run.order: %s:%d %s before %s:%d %s", a.File, a.Line, a.Kind, b.File, b.Line, b.Kind)
	}
}

// ---------------------------------------------------------------------------
// The read-only invariant

// snapshot records every entry below every spec root under base: folders,
// file contents and link targets, without following links.
func snapshot(t *testing.T, base string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	var roots []string
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "proviso" {
			if fi, err := os.Lstat(filepath.Join(p, "system.md")); err == nil && fi.Mode().IsRegular() {
				roots = append(roots, p)
			}
		}
		return nil
	})
	for _, root := range roots {
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			rel, _ := filepath.Rel(base, p)
			switch {
			case err != nil:
				snap[rel] = "error: " + err.Error()
			case d.Type()&fs.ModeSymlink != 0:
				target, _ := os.Readlink(p)
				snap[rel] = "link: " + target
			case d.IsDir():
				snap[rel] = "folder"
			case !d.Type().IsRegular():
				// Such as a named pipe, which reading could block.
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
	}
	return snap
}

func checkReadOnly(t *testing.T, before, after map[string]string) {
	t.Helper()
	for p, v := range before {
		if w, ok := after[p]; !ok {
			t.Errorf("lint.run.read-only: %s was removed", p)
		} else if v != w {
			t.Errorf("lint.run.read-only: %s changed", p)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			t.Errorf("lint.run.read-only: %s was created", p)
		}
	}
}

// ---------------------------------------------------------------------------
// Assertions on findings

func isSentinel(p string) bool {
	for _, s := range sentinels {
		if p == s {
			return true
		}
	}
	return false
}

func dump(list []finding) string {
	if len(list) == 0 {
		return "  (no findings)\n"
	}
	var b strings.Builder
	for _, f := range list {
		fmt.Fprintf(&b, "  %s %s %s:%d %v\n", f.Severity, f.Kind, f.File, f.Line, f.Related)
	}
	return b.String()
}

// at tells whether line is within lines: any line if lines is empty, that
// line if it holds one, the range from the first to the second otherwise.
func at(line int, lines []int) bool {
	switch len(lines) {
	case 0:
		return true
	case 1:
		return line == lines[0]
	default:
		return line >= lines[0] && line <= lines[1]
	}
}

func where(file string, lines []int) string {
	switch len(lines) {
	case 0:
		return file
	case 1:
		return fmt.Sprintf("%s:%d", file, lines[0])
	default:
		return fmt.Sprintf("%s:%d-%d", file, lines[0], lines[1])
	}
}

func (o outcome) find(kind, file string, lines []int) []finding {
	var found []finding
	for _, f := range o.report.Findings {
		inFile := f.File == file || (file == "" && !isSentinel(f.File))
		if f.Kind == kind && inFile && at(f.Line, lines) {
			found = append(found, f)
		}
	}
	return found
}

// has asserts a finding of kind in file (any but the sentinels if empty) at lines, and
// returns it.
func (o outcome) has(t *testing.T, kind, file string, lines ...int) finding {
	t.Helper()
	if o.json.exit == 2 {
		t.Fatalf("want %s at %s, but lint could not run:\n%s", kind, where(file, lines), o.json.stderr)
	}
	found := o.find(kind, file, lines)
	if len(found) == 0 {
		t.Errorf("want %s at %s, got:\n%s", kind, where(file, lines), dump(o.report.Findings))
		return finding{}
	}
	return found[0]
}

// lacks asserts no finding of kind in file (anywhere but in the sentinels
// if empty) at lines.
func (o outcome) lacks(t *testing.T, kind, file string, lines ...int) {
	t.Helper()
	if o.json.exit == 2 {
		t.Fatalf("want no %s, but lint could not run:\n%s", kind, o.json.stderr)
	}
	if found := o.find(kind, file, lines); len(found) != 0 {
		t.Errorf("want no %s at %s, got:\n%s", kind, where(file, lines), dump(o.report.Findings))
	}
}

// clean asserts that lint reported nothing and exited with 0.
func (o outcome) clean(t *testing.T) {
	t.Helper()
	if o.json.exit != 0 || len(o.report.Findings) != 0 {
		t.Errorf("want no findings and exit status 0, got exit status %d:\n%s%s", o.json.exit, dump(o.report.Findings), o.json.stderr)
	}
}

// relates asserts that f names file:line among its related places.
func relates(t *testing.T, f finding, file string, line int) {
	t.Helper()
	for _, r := range f.Related {
		if r.File == file && r.Line == line {
			return
		}
	}
	t.Errorf("want %s at %s:%d to relate to %s:%d, got related %v", f.Kind, f.File, f.Line, file, line, f.Related)
}

// ---------------------------------------------------------------------------
// Examples rows

// row is one Examples row: a tree, whether the guard holds, and where the
// finding is expected when it does not.
type row struct {
	name  string
	tree  tree
	holds bool
	file  string
	lines []int
}

// rows runs each row: where the guard holds, no finding of kind; where it
// fails, a finding of kind at the row's place. In both, lint goes on with the
// rest of the spec tree.
func rows(t *testing.T, kind string, rs []row) {
	t.Helper()
	for _, r := range rs {
		t.Run(r.name, func(t *testing.T) {
			o := checkGoesOn(t, r.tree)
			if r.holds {
				o.lacks(t, kind, r.file, r.lines...)
			} else {
				o.has(t, kind, r.file, r.lines...)
			}
		})
	}
}

func lines(l ...int) []int { return l }
