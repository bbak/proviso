package acceptance

// The driver of `proviso impact`: runs with what holds for every run (the
// impact_set's shape, its orders, the exit status, baseline.impact.listed
// against the element diff, and the invariant baseline.run.writes).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
)

type reason struct {
	Source string   `json:"source"`
	Via    string   `json:"via"`
	From   []string `json:"from"`
}

type candidate struct {
	ID      string   `json:"id"`
	File    string   `json:"file"`
	Line    int      `json:"line"`
	Reasons []reason `json:"reasons"`
}

type scopeGroup struct {
	ID         string   `json:"id"`
	Candidates []string `json:"candidates"`
}

type impactSet struct {
	Root       string       `json:"root"`
	Revision   *string      `json:"revision"`
	Candidates []candidate  `json:"candidates"`
	Scopes     []scopeGroup `json:"scopes"`
}

type impactOutcome struct {
	json run
	text *run // nil unless the test asked for the text form
	is   impactSet
	diff diffOutcome
}

// runs are the runs made: with `--json` and, if asked for, without.
func (o impactOutcome) runs() []run {
	if o.text == nil {
		return []run{o.json}
	}
	return []run{o.json, *o.text}
}

// sourceRank is the order of the sources baseline.impact.order gives.
var sourceRank = map[string]int{"events": 0, "ranges": 1, "scope": 2}

// impactIn runs `proviso impact --json` in cwd and checks what holds for
// every run. base is the directory the test tree was
// written to; nothing below it may change. root, if not empty, is the spec
// root expected. Where impact runs, `diff` is run as well, so that no
// candidate can be an element the diff lists (baseline.impact.listed).
func impactIn(t *testing.T, base, cwd, root string) impactOutcome {
	t.Helper()
	return runImpact(t, false, base, cwd, root)
}

// runImpact is impactIn, with `proviso impact` run as well if text is true.
func runImpact(t *testing.T, text bool, base, cwd, root string) impactOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	o := impactOutcome{json: execProviso(t, cwd, "impact", "--json")}
	if text {
		r := execProviso(t, cwd, "impact")
		o.text = &r
	}
	after := snapshotAll(t, base)
	checkWrites(t, before, after, "")
	if o.text != nil && o.json.exit != o.text.exit {
		t.Errorf("exit status %d with --json, %d without", o.json.exit, o.text.exit)
	}
	for _, r := range o.runs() {
		if r.exit != 0 && r.exit != 2 {
			t.Fatalf("system.cli.impact.exit-status: %d is not 0 or 2\nstderr: %s", r.exit, r.stderr)
		}
		if r.exit == 2 && strings.TrimSpace(r.stderr) == "" {
			t.Errorf("exit status 2 with nothing on stderr")
		}
	}
	if o.json.exit == 2 {
		for _, r := range o.runs() {
			if r.stdout != "" {
				t.Errorf("exit status 2, but stdout holds: %q", r.stdout)
			}
		}
		return o
	}
	o.is = parseImpactSet(t, o.json.stdout)
	if root != "" {
		checkRoot(t, outcome{report: report{Root: o.is.Root}}, root)
	}
	// baseline.impact.output: nothing written as text for no candidates.
	if o.text != nil && len(o.is.Candidates) == 0 && o.text.stdout != "" {
		t.Errorf("baseline.impact.output: no candidates, but text output %q", o.text.stdout)
	}
	if o.text != nil && len(o.is.Candidates) > 0 && strings.TrimSpace(o.text.stdout) == "" {
		t.Errorf("baseline.impact.output: %d candidates, but no text output", len(o.is.Candidates))
	}
	// baseline.impact.listed: no element the diff lists is a candidate.
	o.diff = diffIn(t, base, cwd, root)
	if o.diff.json.exit == 2 {
		t.Errorf("impact ran, but diff could not:\n%s", o.diff.json.stderr)
		return o
	}
	listed := map[string]bool{}
	for _, e := range o.diff.ed.Elements {
		listed[e.ID] = true
	}
	for _, c := range o.is.Candidates {
		if listed[c.ID] {
			t.Errorf("baseline.impact.listed: %s is a candidate, but the element diff lists it", c.ID)
		}
		for _, r := range c.Reasons {
			for _, f := range r.From {
				if !listed[f] {
					t.Errorf("system.cli.impact-set: %s names %s among its reasons' touched elements, which the element diff does not list", c.ID, f)
				}
			}
		}
	}
	if o.diff.ed.Revision != nil && (o.is.Revision == nil || *o.is.Revision != *o.diff.ed.Revision) {
		t.Errorf("system.cli.impact-set: revision %s, but the former text of the element diff came from %s", deref(o.is.Revision), *o.diff.ed.Revision)
	}
	return o
}

// parseImpactSet checks stdout against the schema of `impact_set`, its
// orders, and that the scopes agree with the candidates' scope reasons.
func parseImpactSet(t *testing.T, stdout string) impactSet {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Fatalf("system.cli.impact-set: stdout is not one JSON object: %v\n%s", err, stdout)
	}
	var s string
	if raw["root"] == nil || json.Unmarshal(raw["root"], &s) != nil {
		t.Errorf("system.cli.impact-set: root missing or not a string: %s", raw["root"])
	}
	if rev, ok := raw["revision"]; !ok {
		t.Errorf("system.cli.impact-set: revision missing")
	} else if string(bytes.TrimSpace(rev)) != "null" && json.Unmarshal(rev, &s) != nil {
		t.Errorf("system.cli.impact-set: revision is neither a string nor null: %s", rev)
	}
	for _, k := range []string{"candidates", "scopes"} {
		if r := bytes.TrimSpace(raw[k]); len(r) == 0 || r[0] != '[' {
			t.Fatalf("system.cli.impact-set: %s missing or not an array: %s", k, raw[k])
		}
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw["candidates"], &items); err != nil {
		t.Fatalf("system.cli.impact-set: candidates are not objects: %v", err)
	}
	for i, it := range items {
		for _, k := range []string{"id", "file", "line", "reasons"} {
			if _, ok := it[k]; !ok {
				t.Errorf("system.cli.impact-set: candidate %d has no %s", i, k)
			}
		}
		var line int
		if json.Unmarshal(it["line"], &line) != nil || line < 1 {
			t.Errorf("system.cli.impact-set: candidate %d: line %s is not an integer of at least 1", i, it["line"])
		}
		var reasons []map[string]json.RawMessage
		if json.Unmarshal(it["reasons"], &reasons) != nil {
			t.Errorf("system.cli.impact-set: candidate %d: reasons is not an array of objects", i)
		}
		for _, r := range reasons {
			for _, k := range []string{"source", "via", "from"} {
				if _, ok := r[k]; !ok {
					t.Errorf("system.cli.impact-set: candidate %d: a reason has no %s", i, k)
				}
			}
			if json.Unmarshal(r["source"], &s) != nil {
				t.Errorf("system.cli.impact-set: candidate %d: source is not a string", i)
			} else if _, ok := sourceRank[s]; !ok {
				t.Errorf("system.cli.impact-set: candidate %d: source %q is none of events, ranges, scope", i, s)
			}
			var from []string
			if json.Unmarshal(r["from"], &from) != nil {
				t.Errorf("system.cli.impact-set: candidate %d: from is not an array of strings", i)
			}
		}
	}
	var is impactSet
	if err := json.Unmarshal([]byte(stdout), &is); err != nil {
		t.Fatalf("system.cli.impact-set: %v", err)
	}
	scoped := map[string]map[string]bool{}
	for i, c := range is.Candidates {
		if strings.Contains(c.File, `\`) {
			t.Errorf("system.cli.impact-set: path %q does not use / as its separator", c.File)
		}
		if i > 0 && is.Candidates[i-1].ID >= c.ID {
			t.Errorf("baseline.impact.order: %s before %s", is.Candidates[i-1].ID, c.ID)
		}
		if len(c.Reasons) == 0 {
			t.Errorf("baseline.candidate: %s is found through no source", c.ID)
		}
		for j, r := range c.Reasons {
			if j > 0 {
				p := c.Reasons[j-1]
				if sourceRank[p.Source] > sourceRank[r.Source] || (p.Source == r.Source && p.Via >= r.Via) {
					t.Errorf("baseline.impact.order: in %s, reason %s %q before %s %q", c.ID, p.Source, p.Via, r.Source, r.Via)
				}
			}
			if len(r.From) == 0 || !sort.StringsAreSorted(r.From) {
				t.Errorf("system.cli.impact-set: in %s, from %v is empty or not in byte order", c.ID, r.From)
			}
			if r.Source == "scope" {
				if scoped[r.Via] == nil {
					scoped[r.Via] = map[string]bool{}
				}
				scoped[r.Via][c.ID] = true
			}
		}
	}
	// baseline.impact.scope: the candidates found through scope are listed
	// together under each framing element as well.
	grouped := map[string]bool{}
	for _, g := range is.Scopes {
		grouped[g.ID] = true
		if !sort.StringsAreSorted(g.Candidates) {
			t.Errorf("system.cli.impact-set: the candidates of the scope %s are not in byte order: %v", g.ID, g.Candidates)
		}
		if len(g.Candidates) != len(scoped[g.ID]) {
			t.Errorf("system.cli.impact-set: the scope %s lists %v, but the candidates found through it are %v", g.ID, g.Candidates, keys(scoped[g.ID]))
		}
		for _, id := range g.Candidates {
			if !scoped[g.ID][id] {
				t.Errorf("system.cli.impact-set: the scope %s lists %s, which has no scope reason naming it", g.ID, id)
			}
		}
	}
	for via := range scoped {
		if !grouped[via] {
			t.Errorf("system.cli.impact-set: candidates found through the scope of %s, but no scope lists them", via)
		}
	}
	return is
}

func keys(m map[string]bool) []string {
	var k []string
	for s := range m {
		k = append(k, s)
	}
	sort.Strings(k)
	return k
}

func (o impactOutcome) dump() string {
	var b strings.Builder
	if len(o.is.Candidates) == 0 {
		b.WriteString("  (no candidates)\n")
	}
	for _, c := range o.is.Candidates {
		fmt.Fprintf(&b, "  %s %s:%d\n", c.ID, c.File, c.Line)
		for _, r := range c.Reasons {
			fmt.Fprintf(&b, "    %s %q from %v\n", r.Source, r.Via, r.From)
		}
	}
	return b.String()
}

func (o impactOutcome) ran(t *testing.T) {
	t.Helper()
	if o.json.exit != 0 {
		t.Fatalf("want impact to exit with 0, got %d:\n%s", o.json.exit, o.json.stderr)
	}
}

// only asserts that exactly the candidates ids are found.
func (o impactOutcome) only(t *testing.T, ids ...string) {
	t.Helper()
	o.ran(t)
	var got []string
	for _, c := range o.is.Candidates {
		got = append(got, c.ID)
	}
	want := append([]string(nil), ids...)
	sort.Strings(want)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("want the candidates %v, got:\n%s", want, o.dump())
	}
}

func (o impactOutcome) cand(t *testing.T, id string) candidate {
	t.Helper()
	o.ran(t)
	for _, c := range o.is.Candidates {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("want %s a candidate, got:\n%s", id, o.dump())
	return candidate{}
}

func (o impactOutcome) notCand(t *testing.T, id string) {
	t.Helper()
	o.ran(t)
	for _, c := range o.is.Candidates {
		if c.ID == id {
			t.Errorf("want %s no candidate, got:\n%s", id, o.dump())
		}
	}
}

// hasReason asserts that the candidate id is found through source; via, if
// not empty, is what the reason names, and from, if given, the touched
// elements it comes from.
func (o impactOutcome) hasReason(t *testing.T, id, source, via string, from ...string) {
	t.Helper()
	c := o.cand(t, id)
	for _, r := range c.Reasons {
		if r.Source == source && (via == "" || r.Via == via) {
			if from != nil && strings.Join(r.From, " ") != strings.Join(from, " ") {
				t.Errorf("want the %s reason %q of %s to come from %v, got %v", source, r.Via, id, from, r.From)
			}
			return
		}
	}
	t.Errorf("want %s found through %s %q, got:\n%s", id, source, via, o.dump())
}

// lacksSource asserts that no candidate is found through source.
func (o impactOutcome) lacksSource(t *testing.T, source string) {
	t.Helper()
	o.ran(t)
	for _, c := range o.is.Candidates {
		for _, r := range c.Reasons {
			if r.Source == source {
				t.Errorf("want nothing found through %s, got:\n%s", source, o.dump())
				return
			}
		}
	}
}

// at asserts that each candidate stands where the impact tree has it, at the
// line of its heading.
func (o impactOutcome) at(t *testing.T, tr tree) {
	t.Helper()
	for _, c := range o.is.Candidates {
		path, ok := standsIn[c.ID]
		if !ok {
			continue
		}
		_, line := elementText(tr[path], c.ID)
		if c.File != path || c.Line != line {
			t.Errorf("want %s at %s:%d, got %s:%d", c.ID, path, line, c.File, c.Line)
		}
	}
}

func (r *repo) impact() impactOutcome {
	r.t.Helper()
	return impactIn(r.t, r.dir, r.dir, r.root())
}

// impactText is impact, with the text form as well.
func (r *repo) impactText() impactOutcome {
	r.t.Helper()
	return runImpact(r.t, true, r.dir, r.dir, r.root())
}
