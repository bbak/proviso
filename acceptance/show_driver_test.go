package acceptance

// The driver of the capability query: runs of `proviso show` with what holds
// for every run (the invariant query.read-only, the shapes of element_view
// and show_text, the exit status), the show tree, and an oracle for an
// element as show writes it, computed from the tree alone.

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// The show tree

const (
	showSysGlossaryPath = "proviso/glossary.md"
	showContractPath    = "proviso/contracts/cli.md"
	billingGlossaryPath = "proviso/capabilities/billing/glossary.md"
)

// The system's contract: the command diff with its member since, as Proviso's
// own system.cli.diff.since.
const showContractMD = `# Command line

## <a id="system.cli.diff"></a> diff

The command that compares.

### <a id="system.cli.diff.since"></a> since

The revision to compare with.
`

// The cart's glossary, with the term finding beside item.
const showCartGlossaryMD = glossaryMD + `
## <a id="cart.finding"></a> finding

- Definition: Something the cart reports.
- Contact: Dana Reyes
- Status: defined
`

const showBillingGlossaryMD = `# Glossary

## <a id="billing.finding"></a> finding

- Definition: Something billing reports.
- Contact: Dana Reyes
- Status: defined
`

// showTree is the baseline with what the Examples rows of the capability
// query name, each standing in for an element of Proviso's own tree:
//
//	lint.run.exit       cart.add.limit
//	system.spec-tree    system.spec-tree ("spec tree", alias "tree")
//	coherence.finding   billing.finding
//	lint.finding        cart.finding
//	system.cli.diff.since, as diff.since   the same
//	system.tree-layout, as tree_layout     system.shelf-layout, as shelf_layout
//	lint                cart
//	coherence           billing
//
// lint reports nothing about it (TestFixtureShow).
func showTree() tree {
	return withBilling(baseline(), idBlock, idTable).with(
		showSysGlossaryPath, systemGlossaryMD,
		showContractPath, showContractMD,
		glossaryPath, showCartGlossaryMD,
		billingGlossaryPath, showBillingGlossaryMD,
		sysParamsPath, sysParamsMD,
	)
}

// ---------------------------------------------------------------------------
// Running show

type viewElement struct {
	ID      string  `json:"id"`
	Type    *string `json:"type"`
	Heading string  `json:"heading"`
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Parent  *string `json:"parent"`
	Text    string  `json:"text"`
}

type elementView struct {
	Root     string        `json:"root"`
	Elements []viewElement `json:"elements"`
}

// showOutcome is one run of show; view is the element_view it wrote, where
// it was run with --json and wrote one.
type showOutcome struct {
	run
	view *elementView
}

// showIn runs `proviso show` with args in cwd, and checks what holds for
// every run: query.read-only over every spec root below base, an exit status
// show.exit-status allows, nothing on stdout where it is 2, and, with
// --json and an exit status of 0 or 1, an element_view on stdout and nothing
// else.
func showIn(t *testing.T, base, cwd string, args ...string) showOutcome {
	t.Helper()
	before := snapshot(t, base)
	r := execProviso(t, cwd, append([]string{"show"}, args...)...)
	after := snapshot(t, base)
	checkWrites(t, "query.read-only", before, after)
	o := showOutcome{run: r}
	switch r.exit {
	case 0, 1, 2:
	default:
		t.Fatalf("system.cli.show.exit-status: %d is not 0, 1 or 2\nstderr: %s", r.exit, r.stderr)
	}
	if r.exit == 2 {
		if r.stdout != "" {
			t.Errorf("show exited with 2, but stdout holds: %q", r.stdout)
		}
		if strings.TrimSpace(r.stderr) == "" {
			t.Errorf("show exited with 2 with nothing on stderr")
		}
		return o
	}
	if hasJSON(args) {
		v := parseElementView(t, r.stdout)
		o.view = &v
		if r.exit == 0 && len(v.Elements) == 0 {
			t.Errorf("system.cli.show.exit-status: 0 says at least one element was written, got none")
		}
		if r.exit == 1 && len(v.Elements) != 0 {
			t.Errorf("system.cli.show.exit-status: 1 says the query names no element, got %d", len(v.Elements))
		}
	}
	return o
}

// hasJSON tells whether args hold --json as an option, not as the value of
// --in.
func hasJSON(args []string) bool {
	for i := 0; i < len(args); i++ {
		if args[i] == "--in" {
			i++
			continue
		}
		if args[i] == "--json" {
			return true
		}
	}
	return false
}

// show writes tr into a fresh directory and runs show with args there.
func show(t *testing.T, tr tree, args ...string) (showOutcome, string) {
	t.Helper()
	dir := tr.write(t)
	return showIn(t, dir, dir, args...), dir
}

// elementTypes are the values element_view allows for type, null aside.
var elementTypes = map[string]bool{
	"system": true, "capability": true, "feature": true, "rule": true, "invariant": true,
	"glossary entry": true, "parameter": true, "reference": true, "constraint": true, "contract element": true,
}

// parseElementView checks stdout against the schema of element_view and that
// nothing else is on stdout.
func parseElementView(t *testing.T, stdout string) elementView {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var raw map[string]json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		t.Fatalf("system.cli.element-view: stdout is not one JSON object: %v\n%s", err, stdout)
	}
	if rest, _ := readRest(dec); strings.TrimSpace(rest) != "" {
		t.Errorf("system.cli.element-view: stdout holds more than one JSON object: %q", rest)
	}
	var root string
	if raw["root"] == nil || json.Unmarshal(raw["root"], &root) != nil {
		t.Errorf("system.cli.element-view: root missing or not a string: %s", raw["root"])
	}
	if r := bytes.TrimSpace(raw["elements"]); len(r) == 0 || r[0] != '[' {
		t.Fatalf("system.cli.element-view: elements missing or not an array: %s", raw["elements"])
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw["elements"], &items); err != nil {
		t.Fatalf("system.cli.element-view: elements are not objects: %v", err)
	}
	for i, it := range items {
		for _, k := range []string{"id", "type", "heading", "file", "line", "parent", "text"} {
			if _, ok := it[k]; !ok {
				t.Errorf("system.cli.element-view: element %d has no %s: %v", i, k, it)
			}
		}
		var s string
		for _, k := range []string{"id", "heading", "file", "text"} {
			if json.Unmarshal(it[k], &s) != nil {
				t.Errorf("system.cli.element-view: element %d: %s is not a string: %s", i, k, it[k])
			}
		}
		var typ *string
		if json.Unmarshal(it["type"], &typ) != nil || (typ != nil && !elementTypes[*typ]) {
			t.Errorf("system.cli.element-view: element %d: type %s is no element type nor null", i, it["type"])
		}
		var parent *string
		if json.Unmarshal(it["parent"], &parent) != nil {
			t.Errorf("system.cli.element-view: element %d: parent %s is neither a string nor null", i, it["parent"])
		}
		var line int
		if json.Unmarshal(it["line"], &line) != nil || line < 1 {
			t.Errorf("system.cli.element-view: element %d: line %s is not an integer of at least 1", i, it["line"])
		}
		var f string
		if json.Unmarshal(it["file"], &f) == nil {
			if strings.Contains(f, `\`) {
				t.Errorf("system.cli.element-view: file %q does not use / as its separator", f)
			}
		}
	}
	var v elementView
	if err := json.Unmarshal([]byte(stdout), &v); err != nil {
		t.Fatalf("system.cli.element-view: %v", err)
	}
	return v
}

func readRest(dec *json.Decoder) (string, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(dec.Buffered())
	return b.String(), err
}

// ---------------------------------------------------------------------------
// The oracle

// shown is an element as show is expected to take it: its identifier, the
// file it stands in, its element type ("" for none) and its element parent
// ("" for none). Heading, line and text follow from the tree.
type shown struct {
	id, file, typ, parent string
}

// headingText gives the text of the heading of id in content after its
// anchor.
func headingText(content, id string) string {
	text, _ := elementText(content, id)
	first, _, _ := strings.Cut(text, "\n")
	_, after, _ := strings.Cut(first, `<a id="`+id+`"></a>`)
	return strings.TrimSpace(after)
}

func strOrNull(s *string) string {
	if s == nil {
		return "null"
	}
	return strconv.Quote(*s)
}

// elementsAre checks that the view holds exactly the elements ws, in that
// order, each as the tree tr holds it, and that its root is the spec root
// below dir.
func (o showOutcome) elementsAre(t *testing.T, tr tree, dir string, ws ...shown) {
	t.Helper()
	if o.view == nil {
		t.Fatalf("want an element_view, got exit status %d\nstdout: %s\nstderr: %s", o.exit, o.stdout, o.stderr)
	}
	checkRoot(t, outcome{report: report{Root: o.view.Root}}, filepath.Join(dir, "proviso"))
	var got []string
	for _, e := range o.view.Elements {
		got = append(got, e.ID+"@"+e.File)
	}
	var exp []string
	for _, w := range ws {
		exp = append(exp, w.id+"@"+w.file)
	}
	if strings.Join(got, " ") != strings.Join(exp, " ") {
		t.Fatalf("query.show.match: want the elements %v, in this order, got %v", exp, got)
	}
	for i, w := range ws {
		e := o.view.Elements[i]
		text, line := elementText(tr[w.file], w.id)
		if e.Line != line {
			t.Errorf("query.show.match: %s stands at %s:%d, got line %d", w.id, w.file, line, e.Line)
		}
		if h := headingText(tr[w.file], w.id); e.Heading != h {
			t.Errorf("query.show.match: the heading of %s is %q, got %q", w.id, h, e.Heading)
		}
		if !sameText(e.Text, text) {
			t.Errorf("system.cli.element-view: the text of %s is\n%s\ngot\n%s", w.id, text, e.Text)
		}
		var wt, wp *string
		if w.typ != "" {
			wt = &w.typ
		}
		if w.parent != "" {
			wp = &w.parent
		}
		if strOrNull(e.Type) != strOrNull(wt) {
			t.Errorf("query.show.match: the element type of %s is %s, got %s", w.id, strOrNull(wt), strOrNull(e.Type))
		}
		if strOrNull(e.Parent) != strOrNull(wp) {
			t.Errorf("query.element-parent: the element parent of %s is %s, got %s", w.id, strOrNull(wp), strOrNull(e.Parent))
		}
	}
}

// none checks that the view holds no element and show exited with 1.
func (o showOutcome) none(t *testing.T) {
	t.Helper()
	if o.exit != 1 {
		t.Fatalf("query.show.match: want exit status 1, got %d\nstdout: %s\nstderr: %s", o.exit, o.stdout, o.stderr)
	}
	if o.view == nil || len(o.view.Elements) != 0 {
		t.Errorf("query.show.match: want an element_view without elements, got %q", o.stdout)
	}
}

// refused checks exit status 2, nothing on stdout, and that stderr names
// each of names.
func (o showOutcome) refused(t *testing.T, rule string, names ...string) {
	t.Helper()
	if o.exit != 2 {
		t.Fatalf("%s: want exit status 2, got %d\nstdout: %s\nstderr: %s", rule, o.exit, o.stdout, o.stderr)
	}
	for _, n := range names {
		if !strings.Contains(o.stderr, n) {
			t.Errorf("%s: want stderr to name %q, got: %s", rule, n, o.stderr)
		}
	}
}

// showText gives what show_text holds for the elements of v: per element a
// frontmatter of id, type, heading, file, line and parent, a blank line and
// the element's text without the blank lines at its end; a blank line
// between one element and the next.
func showText(v elementView) string {
	var blocks []string
	for _, e := range v.Elements {
		var b strings.Builder
		b.WriteString("---\n")
		b.WriteString("id: " + jsonString(e.ID) + "\n")
		b.WriteString("type: " + jsonOrNull(e.Type) + "\n")
		b.WriteString("heading: " + jsonString(e.Heading) + "\n")
		b.WriteString("file: " + jsonString(e.File) + "\n")
		b.WriteString("line: " + strconv.Itoa(e.Line) + "\n")
		b.WriteString("parent: " + jsonOrNull(e.Parent) + "\n")
		b.WriteString("---\n\n")
		b.WriteString(strings.TrimRight(e.Text, "\n") + "\n")
		blocks = append(blocks, b.String())
	}
	return strings.Join(blocks, "\n")
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func jsonOrNull(s *string) string {
	if s == nil {
		return "null"
	}
	return jsonString(*s)
}

// isShowText checks that stdout is show_text of the elements of v, up to
// the line endings at its very end.
func isShowText(t *testing.T, stdout string, v elementView) {
	t.Helper()
	got := strings.TrimRight(strings.ReplaceAll(stdout, "\r\n", "\n"), "\n")
	exp := strings.TrimRight(showText(v), "\n")
	if got != exp {
		t.Errorf("system.cli.show-text: want\n%s\ngot\n%s", exp, got)
	}
}

// showTextIDs gives the values of id in the frontmatters of a show_text.
func showTextIDs(stdout string) []string {
	var ids []string
	lines := strings.Split(strings.ReplaceAll(stdout, "\r\n", "\n"), "\n")
	for i := 0; i+1 < len(lines); i++ {
		if lines[i] == "---" && strings.HasPrefix(lines[i+1], "id: ") {
			var id string
			if json.Unmarshal([]byte(strings.TrimPrefix(lines[i+1], "id: ")), &id) == nil {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
