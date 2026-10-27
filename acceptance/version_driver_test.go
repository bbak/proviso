package acceptance

// The driver of `proviso version`: binaries built with a given Proviso
// version, runs of `version` with what holds for every run (its standard
// output, the exit status and the invariant install.version.reads-only), the
// lines it writes to standard error as system.cli.version orders them, and an
// oracle for the version stamp, read as the glossary defines it.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// modelVersion is model_version.
const modelVersion = "0"

// stampedFiles are the agent definitions and skills installed_files names, in
// the byte order of their paths.
var stampedFiles = []string{
	".claude/agents/proviso-checker.md",
	".claude/agents/proviso-coder.md",
	".claude/agents/proviso-tester.md",
	".claude/skills/proviso-accept/SKILL.md",
	".claude/skills/proviso-build/SKILL.md",
	".claude/skills/proviso-check/SKILL.md",
	".claude/skills/proviso-integrate/SKILL.md",
}

// ---------------------------------------------------------------------------
// Binaries of a given Proviso version

// versionedBinaries maps a Proviso version to the *versionedBinary built with
// it.
var versionedBinaries sync.Map

type versionedBinary struct {
	once sync.Once
	path string
	err  error
}

// built gives the binary of the Proviso version v: the one TestMain built,
// given no version, for "dev"; otherwise one built once, on first use, with
// the version given as the Taskfile gives it, through the linker flag
// `-X main.version=<v>`. The pack names no other way to give a build its
// version.
func built(t *testing.T, v string) string {
	t.Helper()
	if v == "dev" {
		return binary
	}
	e, _ := versionedBinaries.LoadOrStore(v, &versionedBinary{})
	b := e.(*versionedBinary)
	b.once.Do(func() {
		name := "proviso-" + v
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		b.path = filepath.Join(filepath.Dir(binary), name)
		cmd := exec.Command("go", "build", "-ldflags", "-X main.version="+v, "-o", b.path, "github.com/bbak/proviso/cmd/proviso")
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			b.err = errors.New(err.Error() + "\n" + string(out))
		}
	})
	if b.err != nil {
		t.Fatalf("building proviso with the version %s: %v", v, b.err)
	}
	return b.path
}

// execBinary runs bin with args in cwd, with nothing on standard input.
func execBinary(t *testing.T, bin, cwd string, args ...string) run {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = cwd
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || ctx.Err() != nil {
			t.Fatalf("running %s %v in %s: %v", filepath.Base(bin), args, cwd, err)
		}
		code = exitErr.ExitCode()
	}
	return run{code, out.String(), errOut.String()}
}

// ---------------------------------------------------------------------------
// Running version

// versionOutcome is one run of `version`: the run and the lines on standard
// error, without their line endings, blank lines left out.
type versionOutcome struct {
	run
	lines []string
	base  string
}

// versionIn runs `proviso version` of the Proviso version v, with args, in
// cwd, and checks what holds for every run. base is the directory the test
// set up, cwd at or below it. prepare, if given, runs after the first
// snapshot and before version; what it returns runs after version and before
// the second snapshot.
func versionIn(t *testing.T, v, base, cwd string, prepare func() func(), args ...string) versionOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	var undo func()
	if prepare != nil {
		undo = prepare()
	}
	r := execBinary(t, built(t, v), cwd, append([]string{"version"}, args...)...)
	if undo != nil {
		undo()
	}
	after := snapshotAll(t, base)
	checkVersionReadsOnly(t, before, after)

	o := versionOutcome{run: r, base: base}
	for _, l := range strings.Split(strings.ReplaceAll(r.stderr, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			o.lines = append(o.lines, l)
		}
	}
	if r.exit != 0 && r.exit != 2 {
		t.Fatalf("system.cli.version.exit-status: %d is not 0 or 2\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	// install.version.output
	if len(args) == 0 {
		want := "proviso " + v + "\nmodel " + modelVersion + "\n"
		if got := strings.ReplaceAll(r.stdout, "\r\n", "\n"); got != want {
			t.Errorf("install.version.output: want stdout %q, got %q", want, got)
		}
		if r.exit != 0 {
			t.Errorf("install.version.output: want exit status 0, whatever goes to stderr, got %d\nstderr: %s", r.exit, r.stderr)
		}
	} else {
		if r.stdout != "" {
			t.Errorf("install.version.output: an argument was given, but stdout holds %q", r.stdout)
		}
		if r.exit != 2 {
			t.Errorf("install.version.output: an argument was given, want exit status 2, got %d", r.exit)
		}
		if len(o.lines) == 0 {
			t.Errorf("install.version.output: an argument was given, but stderr says nothing")
		}
	}
	return o
}

// version runs `version` of the Proviso version v in the root of the
// repository, with nothing given.
func (r *repo) version(v string) versionOutcome {
	r.t.Helper()
	return versionIn(r.t, v, r.dir, r.dir, nil)
}

// checkVersionReadsOnly checks install.version.reads-only: every entry below
// the directory is byte for byte as it was, and none was added, removed or
// renamed.
func checkVersionReadsOnly(t *testing.T, before, after map[string]string) {
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
			bad = append(bad, p+" was added")
		}
	}
	sort.Strings(bad)
	for _, b := range bad {
		t.Errorf("install.version.reads-only: %s", b)
	}
}

// names gives the agent definitions and skills the line names, by their
// paths as installed_files gives them, with / or the platform's separator.
func names(line string) []string {
	var found []string
	for _, p := range stampedFiles {
		if strings.Contains(line, p) || strings.Contains(line, filepath.FromSlash(p)) {
			found = append(found, p)
		}
	}
	return found
}

// expect is what a run of `version` in the root of a repository writes to
// standard error, as system.cli.version orders it: first either the line
// saying that no skill or agent is installed (none), or one line for each
// file named, in the byte order of their paths; then, with manifest, one line
// about the manifest.
type expect struct {
	none     bool
	files    []string
	manifest bool
}

// reports checks the lines on standard error against e, and returns the
// line naming each file.
func (o versionOutcome) reports(t *testing.T, e expect) map[string]string {
	t.Helper()
	files := append([]string(nil), e.files...)
	sort.Strings(files)
	want := len(files)
	if e.none {
		want++
	}
	if e.manifest {
		want++
	}
	describe := func() string {
		var parts []string
		if e.none {
			parts = append(parts, "the line saying no skill or agent is installed")
		}
		for _, f := range files {
			parts = append(parts, "a line naming "+f)
		}
		if e.manifest {
			parts = append(parts, "a line about the manifest")
		}
		if len(parts) == 0 {
			return "nothing"
		}
		return strings.Join(parts, ", then ")
	}
	got := map[string]string{}
	if len(o.lines) != want {
		t.Errorf("system.cli.version: want on stderr %s; got %d lines:\n%s", describe(), len(o.lines), o.stderr)
		return got
	}
	i := 0
	if e.none {
		if n := names(o.lines[i]); len(n) != 0 {
			t.Errorf("install.version.kit: the line saying no skill or agent is installed names %v on its own: %q", n, o.lines[i])
		}
		i++
	}
	for _, f := range files {
		n := names(o.lines[i])
		if len(n) != 1 || n[0] != f {
			t.Errorf("system.cli.version: want line %d to name %s alone, in the byte order of the paths; it names %v:\n%s", i+1, f, n, o.stderr)
		}
		got[f] = o.lines[i]
		i++
	}
	if e.manifest {
		if n := names(o.lines[i]); len(n) != 0 {
			t.Errorf("system.cli.version: want the last line about the manifest, but it names %v:\n%s", n, o.stderr)
		}
	}
	return got
}

// manifestLine gives the last line on standard error, the line about the
// manifest where one is expected.
func (o versionOutcome) manifestLine() string {
	if len(o.lines) == 0 {
		return ""
	}
	return o.lines[len(o.lines)-1]
}

// hint checks the one line of hint `version` writes outside the root of a git
// repository, and that it checked nothing: no other line, and none naming a
// skill or agent.
func (o versionOutcome) hint(t *testing.T) {
	t.Helper()
	if len(o.lines) != 1 {
		t.Errorf("install.version.repository: want one line of hint on stderr, got %d:\n%s", len(o.lines), o.stderr)
		return
	}
	if n := names(o.lines[0]); len(n) != 0 {
		t.Errorf("install.version.repository: the hint names %v, but nothing is checked", n)
	}
}

// numberToken is a whole number or a dotted version that stands on its own:
// not part of a word, such as "row1".
var numberToken = regexp.MustCompile(`(?:^|[^\w.])(\d+(?:\.\d+)*)`)

// numbers gives the numbers line names, once each text in drop, the longest
// first, has been taken out of it.
func numbers(line string, drop ...string) []string {
	drop = append([]string(nil), drop...)
	sort.Slice(drop, func(i, j int) bool { return len(drop[i]) > len(drop[j]) })
	for _, d := range drop {
		if d != "" {
			line = strings.ReplaceAll(line, d, " ")
		}
	}
	var found []string
	for _, m := range numberToken.FindAllStringSubmatch(line, -1) {
		found = append(found, m[1])
	}
	return found
}

func count(list []string, s string) int {
	n := 0
	for _, x := range list {
		if x == s {
			n++
		}
	}
	return n
}

// staleLine checks the line about file installed by another version: it
// names stamp and, where model is not empty, that model version; where model
// is empty, no model version. The model version is looked for among the
// numbers the line names besides the file, the stamp and the binary's own
// Proviso version v.
func staleLine(t *testing.T, line, file, stamp, v, model string) {
	t.Helper()
	if !strings.Contains(line, stamp) {
		t.Errorf("install.version.stale: the line about %s does not name %s as the version that installed it: %q", file, stamp, line)
	}
	nums := numbers(line, file, filepath.FromSlash(file), stamp, v)
	if model != "" && count(nums, model) == 0 {
		t.Errorf("install.version.stale: the line about %s does not name model %s, the model version of %s: %q", file, model, stamp, line)
	}
	if model == "" && count(nums, modelVersion) != 0 {
		t.Errorf("install.version.stale: the line about %s names a model version, but %s has none this binary can know: %q", file, stamp, line)
	}
}

// ---------------------------------------------------------------------------
// Installed files

// stampedAs is a skill or agent definition stamped with v, as `init` would
// write it: frontmatter with its name and the stamp below metadata:.
func stampedAs(file, v string) string {
	return "---\nname: " + nameOf(file) + "\ndescription: Installed for the test.\nmetadata:\n  proviso-version: " + v + "\n---\n\nInstructions.\n"
}

// nameOf gives the name of the agent or skill at file.
func nameOf(file string) string {
	if strings.HasSuffix(file, "/SKILL.md") {
		return path.Base(path.Dir(file))
	}
	return strings.TrimSuffix(path.Base(file), ".md")
}

// installedAs is every agent definition and skill stamped with v, and a
// manifest of model_version, so that nothing but what a test changes is
// reported.
func installedAs(v string) tree {
	tr := tree{"proviso/manifest.sum": "model " + modelVersion + "\n"}
	for _, f := range stampedFiles {
		tr[f] = stampedAs(f, v)
	}
	return tr
}

// versionRepo is a fresh git repository holding tr.
func versionRepo(t *testing.T, tr tree) *repo {
	t.Helper()
	r := freshRepo(t, "repo", false)
	tr.writeTo(t, r.dir, false)
	return r
}

// stampsOf reads content as the glossary defines the version stamp, and
// gives the version of every stamp, in their order: the frontmatter is the
// lines between a first line --- and the next line ---; the stamps are the
// lines proviso-version: among the lines directly below a line metadata:
// that start with a space or a tab. Line endings, and spaces and tabs at the
// end of a line, around the colons and before proviso-version, change
// nothing.
func stampsOf(content string) []string {
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(strings.TrimSuffix(l, "\r"), " \t")
	}
	if len(lines) == 0 || lines[0] != "---" {
		return nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil
	}
	metadata := regexp.MustCompile(`^metadata[ \t]*:$`)
	stamp := regexp.MustCompile(`^[ \t]+proviso-version[ \t]*:[ \t]*(.*)$`)
	var stamps []string
	for i := 1; i < end; i++ {
		if !metadata.MatchString(lines[i]) {
			continue
		}
		for j := i + 1; j < end && (strings.HasPrefix(lines[j], " ") || strings.HasPrefix(lines[j], "\t")); j++ {
			if m := stamp.FindStringSubmatch(lines[j]); m != nil {
				stamps = append(stamps, m[1])
			}
		}
	}
	return stamps
}
