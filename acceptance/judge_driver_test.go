package acceptance

// The driver of `proviso judge` and `proviso verify`: runs with what holds
// for every run (the contracts system.cli.judge, system.cli.verify and
// system.cli.verification, the order of the problems, the exit statuses, and
// the invariant baseline.run.writes), and check_result inputs built in the
// spec's terms. What coherence_file holds is not part of the contract, so
// what judge recorded is observed through verify.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// coherenceName is the name of coherence_file in the spec root.
const coherenceName = ".coherence.json"

// coherencePath is coherence_file of the test trees, relative to the
// repository directory.
const coherencePath = "proviso/" + coherenceName

// ---------------------------------------------------------------------------
// check_result

type judgementIn struct {
	Reason   string   `json:"reason"`
	Elements []string `json:"elements"`
}

// findingIn is a finding of a check_result: new, with Kind and Text, or
// changing one recorded before, with ID and Status.
type findingIn struct {
	ID       string   `json:"id,omitempty"`
	Status   string   `json:"status,omitempty"`
	Kind     string   `json:"kind,omitempty"`
	Elements []string `json:"elements,omitempty"`
	Text     string   `json:"text,omitempty"`
	Reason   string   `json:"reason,omitempty"`
}

type checkResult struct {
	Judgements []judgementIn `json:"judgements,omitempty"`
	Findings   []findingIn   `json:"findings,omitempty"`
}

func (c checkResult) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// judging is a check_result with one judgement for ids.
func judging(reason string, ids ...string) checkResult {
	return checkResult{Judgements: []judgementIn{{reason, ids}}}
}

// gap is a new finding of the kind gap with text, naming ids.
func gap(text string, ids ...string) findingIn {
	return findingIn{Kind: "gap", Text: text, Elements: ids}
}

func resolved(id string) findingIn { return findingIn{ID: id, Status: "resolved"} }

func dismissed(id, reason string) findingIn {
	return findingIn{ID: id, Status: "dismissed", Reason: reason}
}

// ---------------------------------------------------------------------------
// Running judge

// judgeIn runs `proviso judge` in cwd with input on standard input. base is
// the directory the test tree was written to; root, if not empty, is the spec
// root the command is expected to work on, whose .coherence.json alone may
// change, and only if the command exits with 0.
func judgeIn(t *testing.T, base, cwd, root, input string) run {
	t.Helper()
	before := snapshotAll(t, base)
	r := execProvisoWith(t, cwd, strings.NewReader(input), "judge")
	after := snapshotAll(t, base)
	allowed := ""
	if r.exit == 0 && root != "" {
		rel, _ := filepath.Rel(base, filepath.Join(root, coherenceName))
		allowed = filepath.ToSlash(rel)
	}
	checkWrites(t, before, after, allowed)
	if r.exit < 0 || r.exit > 2 {
		t.Fatalf("system.cli.judge.exit-status: %d is not 0, 1 or 2\nstderr: %s", r.exit, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("system.cli.judge: judge writes nothing to standard output, got: %q", r.stdout)
	}
	if r.exit != 0 && strings.TrimSpace(r.stderr) == "" {
		t.Errorf("exit status %d with nothing on stderr", r.exit)
	}
	if r.exit == 0 && root != "" {
		b, err := os.ReadFile(filepath.Join(root, coherenceName))
		if err != nil {
			t.Fatalf("baseline.judge.write: exit status 0, but no %s in the spec root: %v", coherenceName, err)
		}
		if !utf8.Valid(b) || !json.Valid(b) {
			t.Errorf("system.coherence-file: %s is not JSON in UTF-8:\n%s", coherenceName, b)
		}
	}
	return r
}

// judge runs `proviso judge` in the repository with the check result c.
func (r *repo) judge(c checkResult) run {
	r.t.Helper()
	return r.judgeRaw(c.String())
}

// judgeRaw runs `proviso judge` in the repository with input as it is.
func (r *repo) judgeRaw(input string) run {
	r.t.Helper()
	return judgeIn(r.t, r.dir, r.dir, r.root(), input)
}

// judgeOK runs `proviso judge` with c, which must succeed.
func (r *repo) judgeOK(c checkResult) {
	r.t.Helper()
	if res := r.judge(c); res.exit != 0 {
		r.t.Fatalf("want judge to record %s, got exit status %d\nstderr: %s", c, res.exit, res.stderr)
	}
}

// refused asserts that judge exited with want, naming each of names on
// stderr.
func refused(t *testing.T, r run, want int, names ...string) {
	t.Helper()
	if r.exit != want {
		t.Fatalf("want exit status %d, got %d\nstderr: %s", want, r.exit, r.stderr)
	}
	for _, n := range names {
		if !strings.Contains(r.stderr, n) {
			t.Errorf("want stderr to name %s, got: %s", n, r.stderr)
		}
	}
}

// candidates are the candidates `impact --json` finds in the repository now.
func (r *repo) candidates() []string {
	r.t.Helper()
	res := execProviso(r.t, r.dir, "impact", "--json")
	if res.exit != 0 {
		r.t.Fatalf("want impact to run, got exit status %d\nstderr: %s", res.exit, res.stderr)
	}
	var is impactSet
	if err := json.Unmarshal([]byte(res.stdout), &is); err != nil {
		r.t.Fatalf("impact --json: %v", err)
	}
	var ids []string
	for _, c := range is.Candidates {
		ids = append(ids, c.ID)
	}
	return ids
}

// judgeAll records one judgement for every current candidate, except those
// in skip, with the findings fs, and returns the candidates judged.
func (r *repo) judgeAll(skip []string, fs ...findingIn) []string {
	r.t.Helper()
	ids := without(r.candidates(), skip...)
	c := checkResult{Findings: fs}
	if len(ids) > 0 {
		c.Judgements = []judgementIn{{"untouched: the amendment does not reach it", ids}}
	}
	r.judgeOK(c)
	return ids
}

// coherence reads coherence_file of the repository.
func (r *repo) coherence() string {
	r.t.Helper()
	b, err := os.ReadFile(filepath.Join(r.root(), coherenceName))
	if err != nil {
		r.t.Fatalf("want %s in the spec root: %v", coherenceName, err)
	}
	return string(b)
}

func (r *repo) hasCoherence() bool {
	_, err := os.Lstat(filepath.Join(r.root(), coherenceName))
	return err == nil
}

// ---------------------------------------------------------------------------
// Running verify

type problem struct {
	Kind    string `json:"kind"`
	Element string `json:"element"`
	Finding string `json:"finding"`
	Text    string `json:"text"`
	Message string `json:"message"`
}

type verification struct {
	Root     string    `json:"root"`
	Problems []problem `json:"problems"`
}

type verifyOutcome struct {
	json run
	text *run // nil unless the test asked for the text form
	v    verification
}

// runs are the runs made: with `--json` and, if asked for, without.
func (o verifyOutcome) runs() []run {
	if o.text == nil {
		return []run{o.json}
	}
	return []run{o.json, *o.text}
}

// kindRank is the order in which `verification` lists the kinds.
var kindRank = map[string]int{"missing": 0, "leftover": 1, "unjudged": 2, "stale": 3, "open": 4}

// verifyIn runs `proviso verify --json` in cwd and checks what holds for
// every run. base is the directory the test tree was
// written to; nothing below it may change. root, if not empty, is the spec
// root expected.
func verifyIn(t *testing.T, base, cwd, root string) verifyOutcome {
	t.Helper()
	return runVerify(t, false, base, cwd, root)
}

// runVerify is verifyIn, with `proviso verify` run as well if text is true.
func runVerify(t *testing.T, text bool, base, cwd, root string) verifyOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	o := verifyOutcome{json: execProviso(t, cwd, "verify", "--json")}
	if text {
		r := execProviso(t, cwd, "verify")
		o.text = &r
	}
	after := snapshotAll(t, base)
	checkWrites(t, before, after, "")
	if o.text != nil && o.json.exit != o.text.exit {
		t.Errorf("exit status %d with --json, %d without", o.json.exit, o.text.exit)
	}
	for _, r := range o.runs() {
		if r.exit < 0 || r.exit > 2 {
			t.Fatalf("system.cli.verify.exit-status: %d is not 0, 1 or 2\nstderr: %s", r.exit, r.stderr)
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
	o.v = parseVerification(t, o.json.stdout)
	if root != "" {
		checkRoot(t, outcome{report: report{Root: o.v.Root}}, root)
	}
	checkVerifyOrder(t, o.v.Problems)
	// baseline.verify.exit
	want := 0
	if len(o.v.Problems) > 0 {
		want = 1
	}
	if o.json.exit != want {
		t.Errorf("baseline.verify.exit: exit status %d, want %d for the problems:\n%s", o.json.exit, want, o.dump())
	}
	// baseline.verify.output: nothing written as text when there is
	// nothing to say, the problems otherwise.
	if o.text != nil && len(o.v.Problems) == 0 && o.text.stdout != "" {
		t.Errorf("baseline.verify.output: no problems, but text output %q", o.text.stdout)
	}
	if o.text != nil && len(o.v.Problems) > 0 && strings.TrimSpace(o.text.stdout) == "" {
		t.Errorf("baseline.verify.output: %d problems, but no text output", len(o.v.Problems))
	}
	// baseline.verify.missing, baseline.verify.leftover: after either,
	// verify looks no further.
	for _, p := range o.v.Problems {
		if (p.Kind == "missing" || p.Kind == "leftover") && len(o.v.Problems) != 1 {
			t.Errorf("baseline.verify.%s: verify looks no further, but reported:\n%s", p.Kind, o.dump())
		}
	}
	return o
}

// parseVerification checks stdout against the schema of `verification` and
// that nothing else is on stdout.
func parseVerification(t *testing.T, stdout string) verification {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Fatalf("system.cli.verification: stdout is not one JSON object: %v\n%s", err, stdout)
	}
	var s string
	if raw["root"] == nil || json.Unmarshal(raw["root"], &s) != nil {
		t.Errorf("system.cli.verification: root missing or not a string: %s", raw["root"])
	}
	if r := bytes.TrimSpace(raw["problems"]); len(r) == 0 || r[0] != '[' {
		t.Fatalf("system.cli.verification: problems missing or not an array: %s", raw["problems"])
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw["problems"], &items); err != nil {
		t.Fatalf("system.cli.verification: problems are not objects: %v", err)
	}
	for i, it := range items {
		for _, k := range []string{"kind", "message"} {
			if _, ok := it[k]; !ok {
				t.Errorf("system.cli.verification: problem %d has no %s", i, k)
			}
		}
		for _, k := range []string{"kind", "element", "finding", "text", "message"} {
			if v, ok := it[k]; ok && json.Unmarshal(v, &s) != nil {
				t.Errorf("system.cli.verification: problem %d: %s is not a string: %s", i, k, v)
			}
		}
		var kind string
		json.Unmarshal(it["kind"], &kind)
		if _, ok := kindRank[kind]; !ok {
			t.Errorf("system.cli.verification: problem %d: kind %q is none of missing, leftover, unjudged, stale, open", i, kind)
		}
		need := map[string][]string{"unjudged": {"element"}, "stale": {"element"}, "open": {"finding", "text"}}[kind]
		for _, k := range need {
			if v, ok := it[k]; !ok || json.Unmarshal(v, &s) != nil || s == "" {
				t.Errorf("system.cli.verification: the %s problem %d has no %s", kind, i, k)
			}
		}
	}
	var v verification
	if err := json.Unmarshal([]byte(stdout), &v); err != nil {
		t.Fatalf("system.cli.verification: %v", err)
	}
	return v
}

// findingNumber is the number of the finding id, as in f10, or -1.
func findingNumber(id string) int {
	if !strings.HasPrefix(id, "f") {
		return -1
	}
	n, err := strconv.Atoi(id[1:])
	if err != nil {
		return -1
	}
	return n
}

// checkVerifyOrder checks baseline.verify.order: by kind as `verification`
// lists them, then by identifier, byte by byte, or by the number of the
// finding.
func checkVerifyOrder(t *testing.T, list []problem) {
	t.Helper()
	for i := 1; i < len(list); i++ {
		a, b := list[i-1], list[i]
		ra, rb := kindRank[a.Kind], kindRank[b.Kind]
		switch {
		case ra < rb:
			continue
		case ra > rb:
		case a.Kind == "open" && findingNumber(a.Finding) < findingNumber(b.Finding):
			continue
		case a.Kind != "open" && a.Element < b.Element:
			continue
		}
		t.Errorf("baseline.verify.order: %s %s%s before %s %s%s", a.Kind, a.Element, a.Finding, b.Kind, b.Element, b.Finding)
	}
}

func (o verifyOutcome) dump() string {
	if len(o.v.Problems) == 0 {
		return "  (no problems)\n"
	}
	var b strings.Builder
	for _, p := range o.v.Problems {
		fmt.Fprintf(&b, "  %s %s%s %q\n", p.Kind, p.Element, p.Finding, p.Text)
	}
	return b.String()
}

func (o verifyOutcome) ran(t *testing.T) {
	t.Helper()
	if o.json.exit == 2 {
		t.Fatalf("want verify to run, but it exited with 2:\n%s", o.json.stderr)
	}
}

// key names a problem as kind, a space and what it is about, as in
// "unjudged cart.add.full", "open f2" or "missing".
func (p problem) key() string {
	about := p.Element
	if p.Kind == "open" {
		about = p.Finding
	}
	return strings.TrimSpace(p.Kind + " " + about)
}

func (o verifyOutcome) keys() []string {
	var ks []string
	for _, p := range o.v.Problems {
		ks = append(ks, p.key())
	}
	return ks
}

// only asserts that exactly the problems keys are reported, in any order.
func (o verifyOutcome) only(t *testing.T, keys ...string) {
	t.Helper()
	o.ran(t)
	got := o.keys()
	sort.Strings(got)
	want := append([]string(nil), keys...)
	sort.Strings(want)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("want the problems %q, got:\n%s", want, o.dump())
	}
}

// has asserts that the problem key is reported, and returns it.
func (o verifyOutcome) has(t *testing.T, key string) problem {
	t.Helper()
	o.ran(t)
	for _, p := range o.v.Problems {
		if p.key() == key {
			return p
		}
	}
	t.Errorf("want the problem %q, got:\n%s", key, o.dump())
	return problem{}
}

// lacks asserts that the problem key is not reported.
func (o verifyOutcome) lacks(t *testing.T, key string) {
	t.Helper()
	o.ran(t)
	for _, p := range o.v.Problems {
		if p.key() == key {
			t.Errorf("want no problem %q, got:\n%s", key, o.dump())
		}
	}
}

// exits asserts the exit status of every run.
func (o verifyOutcome) exits(t *testing.T, want int) {
	t.Helper()
	for _, r := range o.runs() {
		if r.exit != want {
			t.Errorf("want exit status %d, got %d\nstderr: %s", want, r.exit, r.stderr)
		}
	}
}

// stoppedNaming asserts exit status 2 for every run, each naming name on
// stderr.
func (o verifyOutcome) stoppedNaming(t *testing.T, name string) {
	t.Helper()
	o.exits(t, 2)
	for _, r := range o.runs() {
		if !strings.Contains(r.stderr, name) {
			t.Errorf("want stderr to name %s, got: %s", name, r.stderr)
		}
	}
}

func (r *repo) verify() verifyOutcome {
	r.t.Helper()
	return verifyIn(r.t, r.dir, r.dir, r.root())
}

// verifyText is verify, with the text form as well.
func (r *repo) verifyText() verifyOutcome {
	r.t.Helper()
	return runVerify(r.t, true, r.dir, r.dir, r.root())
}

// ---------------------------------------------------------------------------
// Amendments of the impact tree

// amendedRepo accepts the impact tree and amends it with cs.
func amendedRepo(t *testing.T, cs ...change) *repo {
	t.Helper()
	r, _ := accepted(t, impactTree())
	r.set(amended(impactTree(), cs...))
	return r
}

// unjudgedKeys gives the "unjudged" problem of each id.
func unjudgedKeys(ids ...string) []string {
	var ks []string
	for _, id := range ids {
		ks = append(ks, "unjudged "+id)
	}
	return ks
}
