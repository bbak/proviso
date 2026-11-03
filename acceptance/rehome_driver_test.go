package acceptance

// The driver of the feature upkeep.rehome: runs of `proviso rehome` with what
// holds for every run (the contract rehome_plan, the exit status, the order,
// the rules upkeep.rehome.exit and upkeep.rehome.collision as far as the plan
// shows them, and the invariant upkeep.rehome.writes), and oracles of the
// identifier home and the collisions, computed from the test tree alone.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

type planRename struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type planCollision struct {
	Identifier string   `json:"identifier"`
	From       []string `json:"from"`
}

type rehomePlan struct {
	Root       string          `json:"root"`
	Renames    []planRename    `json:"renames"`
	Collisions []planCollision `json:"collisions"`
}

// rehomeOutcome is one run of `rehome`, with every entry below the base
// directory before and after it, as snapshotAll records them.
type rehomeOutcome struct {
	run
	plan          rehomePlan
	before, after map[string]string
	dir           string // the base directory, taken out of stderr by names
}

// rehomeIn runs `proviso rehome` with args in cwd and checks what holds for
// every run. base is the directory the test set up, the git repository; root,
// if not empty, is the spec root the command is expected to work on. prepare,
// if given, runs after the first snapshot and before the command; what it
// returns runs after the command and before the second snapshot.
func rehomeIn(t *testing.T, base, cwd, root string, prepare func() func(), args ...string) rehomeOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	var undo func()
	if prepare != nil {
		undo = prepare()
	}
	r := execProviso(t, cwd, append([]string{"rehome"}, args...)...)
	if undo != nil {
		undo()
	}
	o := rehomeOutcome{run: r, before: before, after: snapshotAll(t, base), dir: base}
	has := func(opt string) bool {
		for _, a := range args {
			if a == opt {
				return true
			}
		}
		return false
	}
	write, asJSON := has("--write"), has("--json")

	if r.exit < 0 || r.exit > 2 {
		t.Fatalf("system.cli.rehome.exit-status: %d is not 0, 1 or 2\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
	}
	if r.exit == 2 {
		if strings.TrimSpace(r.stderr) == "" {
			t.Errorf("exit status 2 with nothing on stderr")
		}
		if r.stdout != "" {
			t.Errorf("system.cli.rehome.exit-status: exit status 2, but stdout holds: %q", r.stdout)
		}
	}
	if asJSON && r.exit != 2 {
		o.plan = parsePlan(t, r.stdout)
		if root != "" {
			checkRoot(t, outcome{report: report{Root: o.plan.Root}}, root)
		}
		checkPlanOrder(t, o.plan)
		// upkeep.rehome.exit
		if want := exitFor(o.plan); r.exit != want {
			t.Errorf("upkeep.rehome.exit: exit status %d, want %d for the collisions %v", r.exit, want, o.plan.Collisions)
		}
	}

	// upkeep.rehome.writes
	rootRel := ""
	if root != "" {
		rel, _ := filepath.Rel(base, root)
		rootRel = filepath.ToSlash(rel)
	}
	if write && r.exit == 0 && asJSON {
		checkRenamedOnly(t, o.before, o.after, rootRel, o.plan.Renames)
	} else if !(write && r.exit == 0) {
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	}
	return o
}

func exitFor(p rehomePlan) int {
	if len(p.Collisions) == 0 {
		return 0
	}
	return 1
}

// rehome runs `proviso rehome` with args in the repository.
func (r *repo) rehome(args ...string) rehomeOutcome {
	r.t.Helper()
	return rehomeIn(r.t, r.dir, r.dir, r.root(), nil, args...)
}

// parsePlan checks stdout against the schema of rehome_plan and that nothing
// else is on stdout.
func parsePlan(t *testing.T, stdout string) rehomePlan {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Fatalf("system.cli.rehome-plan: stdout is not one JSON object: %v\n%s", err, stdout)
	}
	var s string
	if raw["root"] == nil || json.Unmarshal(raw["root"], &s) != nil {
		t.Errorf("system.cli.rehome-plan: root missing or not a string: %s", raw["root"])
	}
	for _, k := range []string{"renames", "collisions"} {
		if v := bytes.TrimSpace(raw[k]); len(v) == 0 || v[0] != '[' {
			t.Fatalf("system.cli.rehome-plan: %s missing or not an array: %s", k, raw[k])
		}
	}
	var renames []map[string]json.RawMessage
	if err := json.Unmarshal(raw["renames"], &renames); err != nil {
		t.Fatalf("system.cli.rehome-plan: renames are not objects: %v", err)
	}
	for i, it := range renames {
		for _, k := range []string{"from", "to"} {
			if v, ok := it[k]; !ok || json.Unmarshal(v, &s) != nil {
				t.Errorf("system.cli.rehome-plan: rename %d: %s missing or not a string", i, k)
			}
		}
	}
	var collisions []map[string]json.RawMessage
	if err := json.Unmarshal(raw["collisions"], &collisions); err != nil {
		t.Fatalf("system.cli.rehome-plan: collisions are not objects: %v", err)
	}
	for i, it := range collisions {
		if v, ok := it["identifier"]; !ok || json.Unmarshal(v, &s) != nil {
			t.Errorf("system.cli.rehome-plan: collision %d: identifier missing or not a string", i)
		}
		var from []string
		if v, ok := it["from"]; !ok || json.Unmarshal(v, &from) != nil {
			t.Errorf("system.cli.rehome-plan: collision %d: from missing or not an array of strings", i)
		} else if len(from) == 0 {
			t.Errorf("system.cli.rehome-plan: collision %d names no old identifier planned for it", i)
		}
	}
	var p rehomePlan
	if err := json.Unmarshal([]byte(stdout), &p); err != nil {
		t.Fatalf("system.cli.rehome-plan: %v", err)
	}
	if p.Root == "" {
		t.Errorf("system.cli.rehome-plan: root is empty")
	}
	return p
}

// checkPlanOrder checks upkeep.rehome.order: the renames by their old
// identifier, the collisions by the identifier each would take, each naming
// its old identifiers in that order, all compared byte by byte.
func checkPlanOrder(t *testing.T, p rehomePlan) {
	t.Helper()
	for i := 1; i < len(p.Renames); i++ {
		if p.Renames[i-1].From >= p.Renames[i].From {
			t.Errorf("upkeep.rehome.order: the rename of %s before that of %s", p.Renames[i-1].From, p.Renames[i].From)
		}
	}
	for i := 1; i < len(p.Collisions); i++ {
		if p.Collisions[i-1].Identifier >= p.Collisions[i].Identifier {
			t.Errorf("upkeep.rehome.order: the collision on %s before that on %s", p.Collisions[i-1].Identifier, p.Collisions[i].Identifier)
		}
	}
	for _, c := range p.Collisions {
		for i := 1; i < len(c.From); i++ {
			if c.From[i-1] >= c.From[i] {
				t.Errorf("upkeep.rehome.order: the collision on %s names %s before %s", c.Identifier, c.From[i-1], c.From[i])
			}
		}
	}
}

// identifierToken is what a test tree writes as an identifier, or any word
// that could be one.
var identifierToken = regexp.MustCompile(`[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)*`)

// renamedOnly tells whether after is before with some identifiers of renames
// replaced by their new identifiers, and nothing else changed.
func renamedOnly(before, after string, renames []planRename) bool {
	to := map[string]string{}
	for _, r := range renames {
		to[r.From] = r.To
	}
	bi := identifierToken.FindAllStringIndex(before, -1)
	ai := identifierToken.FindAllStringIndex(after, -1)
	if len(bi) != len(ai) {
		return false
	}
	bEnd, aEnd := 0, 0
	for i := range bi {
		if before[bEnd:bi[i][0]] != after[aEnd:ai[i][0]] {
			return false
		}
		b, a := before[bi[i][0]:bi[i][1]], after[ai[i][0]:ai[i][1]]
		if a != b && to[b] != a {
			return false
		}
		bEnd, aEnd = bi[i][1], ai[i][1]
	}
	return before[bEnd:] == after[aEnd:]
}

// checkRenamedOnly checks upkeep.rehome.writes after a run with --write that
// exited with 0: every entry is where it was, the git repository as it was,
// and under the spec root rootRel nothing changed but manifest.sum and the
// identifiers renamed.
func checkRenamedOnly(t *testing.T, before, after map[string]string, rootRel string, renames []planRename) {
	t.Helper()
	const rule = "upkeep.rehome.writes"
	for p, v := range before {
		w, ok := after[p]
		switch {
		case !ok:
			t.Errorf("%s: %s was removed", rule, p)
		case v == w:
		case rootRel != "" && p == rootRel+"/manifest.sum":
		case rootRel != "" && strings.HasPrefix(p, rootRel+"/") && strings.HasPrefix(v, "file: ") && strings.HasPrefix(w, "file: ") && renamedOnly(v, w, renames):
		default:
			t.Errorf("%s: %s changed beyond the identifiers renamed", rule, p)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			t.Errorf("%s: %s was created", rule, p)
		}
	}
}

// ---------------------------------------------------------------------------
// Assertions

// exits asserts the exit status.
func (o rehomeOutcome) exits(t *testing.T, want int) {
	t.Helper()
	if o.exit != want {
		t.Fatalf("system.cli.rehome.exit-status: want %d, got %d\nstdout: %s\nstderr: %s", want, o.exit, o.stdout, o.stderr)
	}
}

// names asserts that stderr names each of names, with the test's directory
// taken out of stderr first, since its name holds the test's name.
func (o rehomeOutcome) names(t *testing.T, names ...string) {
	t.Helper()
	stderr := o.stderr
	for _, d := range []string{o.dir, filepath.ToSlash(o.dir)} {
		stderr = strings.ReplaceAll(stderr, d, "")
	}
	for _, n := range names {
		if !strings.Contains(stderr, n) {
			t.Errorf("want stderr to name %s, got: %s", n, o.stderr)
		}
	}
}

// renames asserts the renames of the plan, in their order.
func (o rehomeOutcome) renames(t *testing.T, want []planRename) {
	t.Helper()
	sameRenames(t, o.plan.Renames, want)
}

// collisions asserts the collisions of the plan, in their order.
func (o rehomeOutcome) collisions(t *testing.T, want []planCollision) {
	t.Helper()
	sameCollisions(t, o.plan.Collisions, want)
}

// file gives the content of the slash-separated path p after the run, or ""
// with ok false where it is no file.
func (o rehomeOutcome) file(p string) (string, bool) {
	v, ok := o.after[p]
	if !ok || !strings.HasPrefix(v, "file: ") {
		return "", false
	}
	return strings.TrimPrefix(v, "file: "), true
}

// holds asserts that the file at p holds want after the run.
func (o rehomeOutcome) holds(t *testing.T, rule, p, want string) {
	t.Helper()
	got, ok := o.file(p)
	if !ok {
		t.Errorf("%s: %s is no file after the run", rule, p)
		return
	}
	if got != want {
		t.Errorf("%s: want %s to hold\n%s\ngot\n%s", rule, p, want, got)
	}
}

// unchanged asserts that the entry at p is as it was before the run.
func (o rehomeOutcome) unchanged(t *testing.T, rule, p string) {
	t.Helper()
	if o.before[p] != o.after[p] {
		t.Errorf("%s: %s changed", rule, p)
	}
}

func sameRenames(t *testing.T, got, want []planRename) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) && !(len(got) == 0 && len(want) == 0) {
		t.Errorf("want the renames %v, got %v", want, got)
	}
}

func sameCollisions(t *testing.T, got, want []planCollision) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) && !(len(got) == 0 && len(want) == 0) {
		t.Errorf("want the collisions %v, got %v", want, got)
	}
}

// ---------------------------------------------------------------------------
// The oracles of upkeep.identifier-home and upkeep.rehome.collision

var (
	headingAnchor = regexp.MustCompile(`^#{1,6}\s+<a id="([^"]+)"></a>`)
	capabilityDir = regexp.MustCompile(`^proviso/capabilities/([^/]+)/`)
)

// treeElements lists the identifiers of the elements of tr, the anchors that
// start a heading outside fenced blocks in the Markdown files of the spec
// root, each with the path of its file.
func treeElements(tr tree) [][2]string {
	paths := make([]string, 0, len(tr))
	for p := range tr {
		name := p[strings.LastIndex(p, "/")+1:]
		if strings.HasPrefix(p, "proviso/") && strings.HasSuffix(p, ".md") && !strings.HasPrefix(name, ".") {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	var out [][2]string
	for _, p := range paths {
		fence := ""
		for _, l := range strings.Split(tr[p], "\n") {
			trimmed := strings.TrimLeft(l, " ")
			if fence != "" {
				if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]+" ") == "" {
					fence = ""
				}
				continue
			}
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				c := trimmed[:1]
				fence = trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, c))]
				continue
			}
			if m := headingAnchor.FindStringSubmatch(l); m != nil {
				out = append(out, [2]string{m[1], p})
			}
		}
	}
	return out
}

// identifierHome gives the identifier home of an element in the file p of
// tr: the identifier of the capability whose folder holds p, or system for a
// file of the system level.
func identifierHome(tr tree, p string) string {
	m := capabilityDir.FindStringSubmatch(p)
	if m == nil {
		return "system"
	}
	capFile := "proviso/capabilities/" + m[1] + "/capability.md"
	for _, e := range treeElements(tree{capFile: tr[capFile]}) {
		return e[0]
	}
	return ""
}

// homeRenames gives the renames upkeep.rehome.home plans for tr, ordered by
// their old identifier: every element whose first segment is not its
// identifier home, renamed to its home followed by what follows the first
// segment.
func homeRenames(tr tree) []planRename {
	var out []planRename
	for _, e := range treeElements(tr) {
		id, home := e[0], identifierHome(tr, e[1])
		first, rest, _ := strings.Cut(id, ".")
		if first == home {
			continue
		}
		to := home
		if rest != "" {
			to += "." + rest
		}
		out = append(out, planRename{id, to})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].From < out[j].From })
	return out
}

// collisionsOf gives the collisions upkeep.rehome.collision finds in tr,
// ordered as upkeep.rehome.order says: every new identifier that is the
// identifier of an element of tr, or is planned for two elements, with the
// old identifiers planned for it.
func collisionsOf(tr tree) []planCollision {
	ids := map[string]bool{}
	for _, e := range treeElements(tr) {
		ids[e[0]] = true
	}
	from := map[string][]string{}
	for _, r := range homeRenames(tr) {
		from[r.To] = append(from[r.To], r.From)
	}
	var out []planCollision
	for to, f := range from {
		if ids[to] || len(f) > 1 {
			sort.Strings(f)
			out = append(out, planCollision{to, f})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Identifier < out[j].Identifier })
	return out
}

// planned asserts that the plan holds the renames and collisions the oracles
// give for tr.
func (o rehomeOutcome) planned(t *testing.T, tr tree) {
	t.Helper()
	sameRenames(t, o.plan.Renames, homeRenames(tr))
	sameCollisions(t, o.plan.Collisions, collisionsOf(tr))
}

// ---------------------------------------------------------------------------
// The manifest of a rewritten tree

var rehomedManifests sync.Map

type rehomedManifest struct {
	once    sync.Once
	content string
	built   bool
}

// manifestWritten gives what `manifest` writes for tr, written once per tree
// into a folder below templates.
func manifestWritten(t *testing.T, tr tree) string {
	t.Helper()
	v, _ := rehomedManifests.LoadOrStore(tr.key(), &rehomedManifest{})
	m := v.(*rehomedManifest)
	m.once.Do(func() {
		m.content = manifestOf(t, tr)
		m.built = true
	})
	if !m.built {
		t.Fatal("writing the manifest of the tree failed in another test")
	}
	return m.content
}
