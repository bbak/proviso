package acceptance

// The driver of `proviso pack`: runs with what holds for every run (the
// layout of pack_text, the shape of context_pack, the two agreeing with each
// other and with the element diff, baseline.pack.contents, baseline.pack.once,
// baseline.pack.order, the exit status and the invariant
// baseline.run.writes).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// The headings of pack_text, in their order, and how context_pack names
// each but Amendment.
var packHeadings = []string{"System", "Amendment", "Frames", "Context"}
var packHeadingJSON = map[string]string{"System": "system", "Frames": "frames", "Context": "context"}

type packElement struct {
	ID      string `json:"id"`
	Heading string `json:"heading"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Text    string `json:"text"`
}

type contextPack struct {
	Root     string        `json:"root"`
	Revision *string       `json:"revision"`
	Diff     []diffElement `json:"diff"`
	Elements []packElement `json:"elements"`
}

// formerBlock is a line Former `<part>`: of pack_text with the fenced block
// after it.
type formerBlock struct {
	part, fence, content string
}

// packItem is one element as pack_text writes it.
type packItem struct {
	heading  string
	id       string        // from its status line under Amendment, from its heading elsewhere
	status   string        // under Amendment: the line saying what happened to it
	notBuilt string        // under Amendment: the line Not to be built, if any
	headIDs  []string      // the identifiers of the anchored headings in its text
	body     []string      // its text, or a removed element's former text
	formers  []formerBlock // under Amendment: the former text of its parts
}

func (it packItem) text() string { return strings.Join(it.body, "\n") }

// packOutcome is one pack of a tree: `pack` run once with `--json` and once
// without, from the same directory, and `diff` run there too.
type packOutcome struct {
	json, text run
	cp         contextPack
	cpOK       bool     // context_pack could be read
	headings   []string // the headings of pack_text, in their order
	revision   string   // the line Former text from under Amendment, if any
	items      []packItem
	diff       diffOutcome
}

// packIn runs `proviso pack --json` and `proviso pack` in cwd and checks what
// holds for every run. base is the directory the test tree was written to;
// nothing below it may change. root, if not empty, is the spec root expected.
// Where pack runs, `diff` is run as well, so that the pack can be checked
// against the element diff.
func packIn(t *testing.T, base, cwd, root string) packOutcome {
	t.Helper()
	before := snapshotAll(t, base)
	o := packOutcome{json: execProviso(t, cwd, "pack", "--json"), text: execProviso(t, cwd, "pack")}
	after := snapshotAll(t, base)
	checkWrites(t, before, after, "")
	if o.json.exit != o.text.exit {
		t.Errorf("exit status %d with --json, %d without", o.json.exit, o.text.exit)
	}
	for _, r := range []run{o.json, o.text} {
		if r.exit != 0 && r.exit != 2 {
			t.Errorf("system.cli.pack.exit-status: %d is not 0 or 2\nstderr: %s", r.exit, r.stderr)
			return o
		}
		if r.exit == 2 && strings.TrimSpace(r.stderr) == "" {
			t.Errorf("exit status 2 with nothing on stderr")
		}
	}
	if o.json.exit == 2 || o.text.exit == 2 {
		for _, r := range []run{o.json, o.text} {
			if r.exit == 2 && r.stdout != "" {
				t.Errorf("exit status 2, but stdout holds: %q", r.stdout)
			}
		}
		return o
	}
	// Each check reports what it finds and the others go on, so that one
	// cause does not hide another.
	o.cp, o.cpOK = parseContextPack(t, o.json.stdout)
	o.headings, o.revision, o.items = parsePackText(t, o.text.stdout)
	if root != "" && o.cpOK {
		checkRoot(t, outcome{report: report{Root: o.cp.Root}}, root)
	}
	checkPackHeadings(t, o)
	if o.cpOK {
		checkPackAgrees(t, o)
	}
	o.diff = diffIn(t, base, cwd, root)
	if o.diff.json.exit == 2 {
		t.Errorf("pack ran, but diff could not:\n%s", o.diff.json.stderr)
		return o
	}
	checkPackDiff(t, o)
	return o
}

func (r *repo) pack() packOutcome {
	r.t.Helper()
	return packIn(r.t, r.dir, r.dir, r.root())
}

// ---------------------------------------------------------------------------
// context_pack

// parseContextPack checks stdout against the schema of `context_pack`: diff
// holds items of element_diff, elements those of its own.
func parseContextPack(t *testing.T, stdout string) (contextPack, bool) {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Errorf("system.cli.context-pack: stdout is not one JSON object: %v\n%s", err, stdout)
		return contextPack{}, false
	}
	var s string
	if raw["root"] == nil || json.Unmarshal(raw["root"], &s) != nil {
		t.Errorf("system.cli.context-pack: root missing or not a string: %s", raw["root"])
	}
	if rev, ok := raw["revision"]; !ok {
		t.Errorf("system.cli.context-pack: revision missing")
	} else if string(bytes.TrimSpace(rev)) != "null" && json.Unmarshal(rev, &s) != nil {
		t.Errorf("system.cli.context-pack: revision is neither a string nor null: %s", rev)
	}
	for _, k := range []string{"diff", "elements"} {
		if r := bytes.TrimSpace(raw[k]); len(r) == 0 || r[0] != '[' {
			t.Errorf("system.cli.context-pack: %s missing or not an array: %s", k, raw[k])
			return contextPack{}, false
		}
	}
	// Each item of diff exactly as an item of elements in element_diff.
	asDiff, _ := json.Marshal(map[string]json.RawMessage{"root": raw["root"], "revision": raw["revision"], "elements": raw["diff"]})
	parseElementDiff(t, string(asDiff))
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw["elements"], &items); err != nil {
		t.Errorf("system.cli.context-pack: elements are not objects: %v", err)
		return contextPack{}, false
	}
	for i, it := range items {
		for _, k := range []string{"id", "heading", "file", "line", "text"} {
			if _, ok := it[k]; !ok {
				t.Errorf("system.cli.context-pack: element %d has no %s", i, k)
			}
		}
		for _, k := range []string{"id", "heading", "file", "text"} {
			if v, ok := it[k]; ok && json.Unmarshal(v, &s) != nil {
				t.Errorf("system.cli.context-pack: element %d: %s is not a string: %s", i, k, v)
			}
		}
		var line int
		if json.Unmarshal(it["line"], &line) != nil || line < 1 {
			t.Errorf("system.cli.context-pack: element %d: line %s is not an integer of at least 1", i, it["line"])
		}
	}
	var cp contextPack
	if err := json.Unmarshal([]byte(stdout), &cp); err != nil {
		t.Errorf("system.cli.context-pack: %v", err)
		return contextPack{}, false
	}
	enum := map[string]bool{"system": true, "frames": true, "context": true}
	for _, e := range cp.Elements {
		if !enum[e.Heading] {
			t.Errorf("system.cli.context-pack: %s: heading %q is none of system, frames, context", e.ID, e.Heading)
		}
		if strings.Contains(e.File, `\`) {
			t.Errorf("system.cli.context-pack: path %q does not use / as its separator", e.File)
		}
	}
	return cp, true
}

// ---------------------------------------------------------------------------
// pack_text

var (
	fenceOpen    = regexp.MustCompile("^(`{3,}|~{3,})(.*)$")
	anchorHead   = regexp.MustCompile(`^#{1,6} <a id="([^"]+)"></a>`)
	statusLine   = regexp.MustCompile("^(Added|Removed) `([^`]+)`\\.$|^(Changed through|Changed) `([^`]+)`: (.+)\\.$")
	notBuiltLine = regexp.MustCompile(`^Not to be built: (.+)\.$`)
	formerLine   = regexp.MustCompile("^Former `([^`]+)`:$")
	revisionLine = regexp.MustCompile("^Former text from `([^`]+)`\\.$")
	backtickRun  = regexp.MustCompile("`+")
)

// parsePackText reads stdout as pack_text and checks its layout: UTF-8, the
// headings of level 1 among the four and in their order, each holding an
// element; the line Former text from, if any, first under Amendment; a
// blank line before every heading, status line, element and line Former;
// the line Not to be built right after a status line; the former
// text of a part in a block fenced with backticks, labelled text, its fence
// longer than any run of backticks in it. Line endings are read as line
// feeds, or carriage returns and line feeds (see the report).
func parsePackText(t *testing.T, stdout string) ([]string, string, []packItem) {
	t.Helper()
	if !utf8.ValidString(stdout) {
		t.Errorf("system.cli.pack-text: not UTF-8")
	}
	text := strings.ReplaceAll(stdout, "\r\n", "\n")
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	isHeading := map[string]bool{}
	for _, h := range packHeadings {
		isHeading["# "+h] = true
	}
	var headings []string
	revision := ""
	var items []packItem
	var cur *packItem
	heading := ""
	blankBefore := func(i int, what string) {
		if i > 0 && lines[i-1] != "" {
			t.Errorf("system.cli.pack-text: no blank line before the %s on line %d: %q", what, i+1, lines[i])
		}
	}
	fence := ""       // the fence open, if any
	inFormer := false // the open fence holds the former text of a part
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if fence != "" {
			if closesFence(l, fence) {
				fence = ""
				if inFormer {
					inFormer = false
					f := cur.formers[len(cur.formers)-1]
					for _, run := range backtickRun.FindAllString(f.content, -1) {
						if len(run) >= len(f.fence) {
							t.Errorf("system.cli.pack-text: the fence %s of the former text of %s in %s is not longer than the run %s in it", f.fence, f.part, cur.id, run)
						}
					}
					continue
				}
			} else if inFormer {
				f := &cur.formers[len(cur.formers)-1]
				f.content += l + "\n"
				continue
			}
			if cur != nil {
				cur.body = append(cur.body, l)
			}
			continue
		}
		switch {
		case isHeading[l]:
			blankBefore(i, "heading")
			heading = strings.TrimPrefix(l, "# ")
			headings = append(headings, heading)
			cur = nil
			continue
		case heading == "Amendment" && revisionLine.MatchString(l):
			blankBefore(i, "line Former text from")
			if revision != "" || headings[len(headings)-1] != "Amendment" || lineAt(lines, i-2) != "# Amendment" {
				t.Errorf("system.cli.pack-text: line %d %q is not the first line under Amendment", i+1, l)
			}
			revision = revisionLine.FindStringSubmatch(l)[1]
			continue
		case heading == "Amendment" && statusLine.MatchString(l):
			blankBefore(i, "status line")
			m := statusLine.FindStringSubmatch(l)
			id := m[2] + m[4]
			items = append(items, packItem{heading: heading, id: id, status: l})
			cur = &items[len(items)-1]
			continue
		case heading == "Amendment" && notBuiltLine.MatchString(l):
			if cur == nil || cur.notBuilt != "" || len(cur.body) > 0 || lines[i-1] != cur.status {
				t.Errorf("system.cli.pack-text: line %d %q does not follow a status line", i+1, l)
			}
			if cur != nil {
				cur.notBuilt = l
			}
			continue
		case heading == "Amendment" && formerLine.MatchString(l):
			blankBefore(i, "line Former")
			if cur == nil {
				t.Errorf("system.cli.pack-text: line %d %q follows no element", i+1, l)
				continue
			}
			cur.formers = append(cur.formers, formerBlock{part: formerLine.FindStringSubmatch(l)[1]})
			// A blank line, then the fenced block.
			j := i + 1
			for j < len(lines) && lines[j] == "" {
				j++
			}
			m := fenceOpen.FindStringSubmatch(lineAt(lines, j))
			if m == nil || m[1][0] != '`' || strings.TrimSpace(m[2]) != "text" {
				t.Errorf("system.cli.pack-text: line %d %q is not followed by a block fenced with backticks and labelled text", i+1, l)
				continue
			}
			fence, inFormer, i = m[1], true, j
			cur.formers[len(cur.formers)-1].fence = fence
			continue
		case anchorHead.MatchString(l):
			id := anchorHead.FindStringSubmatch(l)[1]
			if heading != "Amendment" {
				blankBefore(i, "element")
				items = append(items, packItem{heading: heading, id: id})
				cur = &items[len(items)-1]
			} else if cur != nil && len(cur.formers) == 0 {
				if len(cur.headIDs) == 0 {
					blankBefore(i, "element")
				}
			}
			if cur != nil {
				cur.headIDs = append(cur.headIDs, id)
			}
		}
		if m := fenceOpen.FindStringSubmatch(l); m != nil {
			fence = m[1]
		}
		switch {
		case cur == nil && heading == "":
			t.Errorf("system.cli.pack-text: line %d %q stands before the first heading", i+1, l)
		case cur == nil:
			if l != "" {
				t.Errorf("system.cli.pack-text: line %d %q under %q belongs to no element", i+1, l, heading)
			}
		case len(cur.formers) > 0:
			if l != "" {
				t.Errorf("system.cli.pack-text: line %d %q follows the former text of %s", i+1, l, cur.id)
			}
		default:
			cur.body = append(cur.body, l)
		}
	}
	if fence != "" {
		t.Errorf("system.cli.pack-text: a fenced block %s is not closed", fence)
	}
	return headings, revision, items
}

// closesFence tells whether l closes the fenced block opened with fence.
func closesFence(l, fence string) bool {
	s := strings.TrimRight(l, " ")
	return len(s) >= len(fence) && strings.Trim(s, fence[:1]) == ""
}

func lineAt(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}

// checkPackHeadings checks system.cli.pack-text's headings and
// baseline.pack.contents, baseline.pack.once and baseline.pack.order for
// every run: the headings in their order, none twice and none empty; every
// element once; the elements under each heading in byte order. That the
// system stands under System is checked against the element diff (see
// checkPackDiff).
func checkPackHeadings(t *testing.T, o packOutcome) {
	t.Helper()
	rank := map[string]int{}
	for i, h := range packHeadings {
		rank[h] = i
	}
	count := map[string]int{}
	for _, it := range o.items {
		count[it.heading]++
	}
	for i, h := range o.headings {
		if i > 0 && rank[o.headings[i-1]] >= rank[h] {
			t.Errorf("system.cli.pack-text: the heading %q after %q", h, o.headings[i-1])
		}
		if count[h] == 0 {
			t.Errorf("system.cli.pack-text: no element appears under %q, which is not left out", h)
		}
	}
	seen := map[string]string{}
	for i, it := range o.items {
		if h, ok := seen[it.id]; ok {
			t.Errorf("baseline.pack.once: %s appears under %q and under %q", it.id, h, it.heading)
		}
		seen[it.id] = it.heading
		if i > 0 && o.items[i-1].heading == it.heading && o.items[i-1].id >= it.id {
			t.Errorf("baseline.pack.order: under %q, %s before %s", it.heading, o.items[i-1].id, it.id)
		}
		if it.heading != "Amendment" && (len(it.headIDs) != 1 || it.headIDs[0] != it.id) {
			t.Errorf("system.cli.pack-text: the text of %s under %q holds the headings %v", it.id, it.heading, it.headIDs)
		}
	}
}

// checkPackAgrees checks that context_pack and pack_text hold the same pack:
// the elements outside Amendment with their headings, in the order of
// pack_text, with the same text; the elements under Amendment as diff.
func checkPackAgrees(t *testing.T, o packOutcome) {
	t.Helper()
	var fromText, fromJSON, amendText, amendJSON []string
	texts := map[string]string{}
	for _, it := range o.items {
		if it.heading == "Amendment" {
			amendText = append(amendText, it.id)
			continue
		}
		fromText = append(fromText, packHeadingJSON[it.heading]+":"+it.id)
		texts[it.id] = it.text()
	}
	for _, e := range o.cp.Elements {
		fromJSON = append(fromJSON, e.Heading+":"+e.ID)
		if tx, ok := texts[e.ID]; ok && !sameText(tx, e.Text) {
			t.Errorf("system.cli.context-pack: the text of %s differs from pack_text's:\n%s\n---\n%s", e.ID, e.Text, tx)
		}
	}
	for _, e := range o.cp.Diff {
		amendJSON = append(amendJSON, e.ID)
	}
	if strings.Join(fromText, " ") != strings.Join(fromJSON, " ") {
		t.Errorf("system.cli.context-pack: elements %v, but pack_text holds %v", fromJSON, fromText)
	}
	if strings.Join(amendText, " ") != strings.Join(amendJSON, " ") {
		t.Errorf("system.cli.context-pack: diff %v, but pack_text holds %v under Amendment", amendJSON, amendText)
	}
}

// checkPackDiff checks the pack against the element diff: diff in
// context_pack exactly as the elements of element_diff, the same revision,
// the same revision named in pack_text, the system alone under System unless
// the element diff lists it, and under Amendment of pack_text every element
// diff lists, in its order, each with its status line, its text or former
// text, and the former text of each changed or removed part, in the order
// diff gives them.
func checkPackDiff(t *testing.T, o packOutcome) {
	t.Helper()
	if o.diff.ed.Revision == nil && o.revision != "" {
		t.Errorf("system.cli.pack-text: no former text was needed, but the pack names the revision %s", o.revision)
	}
	if o.diff.ed.Revision != nil && o.revision != *o.diff.ed.Revision {
		t.Errorf("system.cli.pack-text: want the line Former text from `%s`. first under Amendment, got the revision %q", *o.diff.ed.Revision, o.revision)
	}
	wantSystem := "system"
	for _, e := range o.diff.ed.Elements {
		if e.ID == "system" {
			wantSystem = ""
		}
	}
	if got := strings.Join(o.under("System"), " "); got != wantSystem {
		t.Errorf("baseline.pack.contents: want under System %q, got %q", wantSystem, got)
	}
	if !o.cpOK {
		checkAmendments(t, o)
		return
	}
	var packRaw, diffRaw map[string]any
	json.Unmarshal([]byte(o.json.stdout), &packRaw)
	json.Unmarshal([]byte(o.diff.json.stdout), &diffRaw)
	pd, dd := packRaw["diff"], diffRaw["elements"]
	if !reflect.DeepEqual(pd, dd) {
		a, _ := json.MarshalIndent(pd, "", "  ")
		b, _ := json.MarshalIndent(dd, "", "  ")
		t.Errorf("system.cli.context-pack: diff is not the elements of element_diff:\n%s\n---\n%s", a, b)
	}
	if deref(o.cp.Revision) != deref(o.diff.ed.Revision) {
		t.Errorf("system.cli.context-pack: revision %s, but the element diff's is %s", deref(o.cp.Revision), deref(o.diff.ed.Revision))
	}
	checkAmendments(t, o)
}

// checkAmendments checks the elements under Amendment against those the
// element diff lists: the same, in the same order, each as checkAmendment
// says.
func checkAmendments(t *testing.T, o packOutcome) {
	t.Helper()
	var want []string
	for _, e := range o.diff.ed.Elements {
		want = append(want, e.ID)
	}
	if got := o.under("Amendment"); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("baseline.pack.contents: under Amendment %v, but the element diff lists %v", got, want)
	}
	for _, e := range o.diff.ed.Elements {
		for _, it := range o.items {
			if it.heading == "Amendment" && it.id == e.ID {
				checkAmendment(t, it, e)
			}
		}
	}
}

// checkAmendment checks the element e of the element diff as pack_text
// writes it under Amendment.
func checkAmendment(t *testing.T, it packItem, e diffElement) {
	t.Helper()
	if want := statusFor(e); want != "" && it.status != want {
		t.Errorf("system.cli.pack-text: want the line %q, got %q", want, it.status)
	}
	if e.Change == "changed-through" {
		prefix := "Changed through `" + e.ID + "`: "
		var want []string
		for _, c := range e.Through {
			want = append(want, chainText(c))
		}
		got := strings.Split(strings.TrimSuffix(strings.TrimPrefix(it.status, prefix), "."), "; ")
		if !strings.HasPrefix(it.status, prefix) || strings.Join(sorted(got), "|") != strings.Join(sorted(want), "|") {
			t.Errorf("system.cli.pack-text: want %s changed through %q, separated by \"; \", got %q", e.ID, want, it.status)
		}
	}
	// Not to be built.
	if len(e.Undecided) == 0 && it.notBuilt != "" {
		t.Errorf("system.cli.pack-text: %s is to be built, but has the line %q", e.ID, it.notBuilt)
	}
	if len(e.Undecided) > 0 {
		var want []string
		for _, u := range e.Undecided {
			want = append(want, fmt.Sprintf("%s:%d", u.File, u.Line))
		}
		got := strings.Split(strings.TrimSuffix(strings.TrimPrefix(it.notBuilt, "Not to be built: "), "."), ", ")
		if it.notBuilt == "" || strings.Join(sorted(got), " ") != strings.Join(sorted(want), " ") {
			t.Errorf("system.cli.pack-text: want %s with the line Not to be built naming %v, got %q", e.ID, want, it.notBuilt)
		}
	}
	// Its text, or a removed element's former text.
	if e.Change == "removed" {
		former := unfenced(it.body)
		if former == strings.TrimSpace(it.text()) {
			t.Errorf("system.cli.pack-text: the former text of the removed %s does not stand in a block fenced with backticks and labelled text:\n%s", e.ID, it.text())
		}
		if m := fenceOpen.FindStringSubmatch(strings.TrimSpace(it.text())); m != nil && former != strings.TrimSpace(it.text()) {
			for _, run := range backtickRun.FindAllString(former, -1) {
				if m[1][0] != '`' || len(run) >= len(m[1]) {
					t.Errorf("system.cli.pack-text: the fence %s of the former text of %s is not backticks longer than the run %s in it", m[1], e.ID, run)
				}
			}
		}
		if e.Former == nil || !sameText(former, *e.Former) {
			t.Errorf("system.cli.pack-text: want the former text of the removed %s in place of its text, got:\n%s", e.ID, it.text())
		}
	} else {
		if e.Text == nil || !sameText(it.text(), *e.Text) {
			t.Errorf("system.cli.pack-text: want the text of %s as element_diff holds it\n%s\ngot\n%s", e.ID, deref(e.Text), it.text())
		}
		if len(it.headIDs) == 0 || it.headIDs[0] != e.ID {
			t.Errorf("system.cli.pack-text: the text under the status line of %s starts with no heading of it", e.ID)
		}
	}
	// The former text of each changed or removed part, in the order diff
	// gives the parts.
	var want []string
	formers := map[string]string{}
	for _, p := range e.Parts {
		if p.Change != "added" {
			want = append(want, p.Part)
			formers[p.Part] = deref(p.Former)
		}
	}
	var got []string
	for _, f := range it.formers {
		got = append(got, f.part)
		if w, ok := formers[f.part]; ok && !sameText(f.content, w) {
			t.Errorf("system.cli.pack-text: want the former text of %s in %s\n%s\ngot\n%s", f.part, e.ID, w, f.content)
		}
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("system.cli.pack-text: want the lines Former of %s for %q, got %q", e.ID, want, got)
	}
}

// statusFor writes the status line of e, except for an element changed
// through a dependency, whose chains have no order (see the report).
func statusFor(e diffElement) string {
	switch e.Change {
	case "added":
		return "Added `" + e.ID + "`."
	case "removed":
		return "Removed `" + e.ID + "`."
	case "changed":
		var parts []string
		for _, p := range e.Parts {
			parts = append(parts, "`"+p.Part+"` "+p.Change)
		}
		return "Changed `" + e.ID + "`: " + strings.Join(parts, ", ") + "."
	}
	return ""
}

// chainText writes a chain of identifiers as pack_text does.
func chainText(ids []string) string {
	var q []string
	for _, id := range ids {
		q = append(q, "`"+id+"`")
	}
	return strings.Join(q, " → ")
}

// unfenced gives the text of body, or what it holds if it is one block
// fenced with backticks and labelled text.
func unfenced(body []string) string {
	text := strings.TrimSpace(strings.Join(body, "\n"))
	lines := strings.Split(text, "\n")
	if m := fenceOpen.FindStringSubmatch(lines[0]); m != nil && len(lines) >= 2 &&
		strings.TrimSpace(m[2]) == "text" && closesFence(lines[len(lines)-1], m[1]) {
		return strings.Join(lines[1:len(lines)-1], "\n")
	}
	return text
}

// ---------------------------------------------------------------------------
// Assertions

func (o packOutcome) ran(t *testing.T) {
	t.Helper()
	if o.json.exit != 0 || o.text.exit != 0 {
		t.Fatalf("want pack to exit with 0, got %d with --json, %d without:\n%s%s", o.json.exit, o.text.exit, o.json.stderr, o.text.stderr)
	}
}

// stopped asserts exit status 2 for both runs; that nothing went to stdout is
// checked for every run with exit status 2.
func (o packOutcome) stopped(t *testing.T) {
	t.Helper()
	for _, r := range []run{o.json, o.text} {
		if r.exit != 2 {
			t.Errorf("want exit status 2, got %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
		}
	}
}

// under lists the identifiers of the elements under heading in pack_text, in
// their order.
func (o packOutcome) under(heading string) []string {
	var ids []string
	for _, it := range o.items {
		if it.heading == heading {
			ids = append(ids, it.id)
		}
	}
	return ids
}

func (o packOutcome) dump() string {
	var b strings.Builder
	for _, h := range o.headings {
		fmt.Fprintf(&b, "  %s: %v\n", h, o.under(h))
	}
	if len(o.headings) == 0 {
		b.WriteString("  (no headings)\n")
	}
	return b.String()
}

// exactly asserts that exactly the elements ids appear under heading, in any
// order (their order is checked for every run), and that the heading is left
// out if ids is empty.
func (o packOutcome) exactly(t *testing.T, heading string, ids ...string) {
	t.Helper()
	o.ran(t)
	got := o.under(heading)
	sort.Strings(got)
	if strings.Join(got, " ") != strings.Join(sorted(ids), " ") {
		t.Errorf("want under %q exactly %v, got:\n%s", heading, sorted(ids), o.dump())
	}
	present := false
	for _, h := range o.headings {
		present = present || h == heading
	}
	if present != (len(ids) > 0) {
		t.Errorf("want the heading %q present: %v, got:\n%s", heading, len(ids) > 0, o.dump())
	}
}

// headingsAre asserts the headings of pack_text.
func (o packOutcome) headingsAre(t *testing.T, want ...string) {
	t.Helper()
	o.ran(t)
	if strings.Join(o.headings, "|") != strings.Join(want, "|") {
		t.Errorf("want the headings %q, got %q", want, o.headings)
	}
}

// item returns the element id as pack_text writes it.
func (o packOutcome) item(t *testing.T, id string) packItem {
	t.Helper()
	o.ran(t)
	for _, it := range o.items {
		if it.id == id {
			return it
		}
	}
	t.Errorf("want %s in the pack, got:\n%s", id, o.dump())
	return packItem{}
}

// standsAt asserts that each element of the pack outside Amendment has its
// text as tr has it, and in context_pack stands where
// tr has it, at the line of its heading.
func (o packOutcome) standsAt(t *testing.T, tr tree) {
	t.Helper()
	for _, e := range o.cp.Elements {
		path, ok := packStandsIn[e.ID]
		if !ok {
			continue
		}
		text, line := elementText(tr[path], e.ID)
		if e.File != path || e.Line != line {
			t.Errorf("want %s at %s:%d, got %s:%d", e.ID, path, line, e.File, e.Line)
		}
		if !sameText(e.Text, text) {
			t.Errorf("want the text of %s as it stands now\n%s\ngot\n%s", e.ID, text, e.Text)
		}
	}
}
