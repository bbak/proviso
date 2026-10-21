package baseline

import (
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/tree"
)

const system = "# <a id=\"system\"></a> S\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n## Scope\n\ns\n\n## Out of scope\n\no\n\n## Audience\n\na\n\n## Design boundary\n\nd\n"

const glossary = `# Glossary

## <a id="system.thing"></a> thing

- Definition: A thing.
- Contact: Ann
- Status: defined
- Aliases: things

## <a id="system.std"></a> std

- Defined by: [Std](references.md#system.ref.std)
- Contact: Ann
- Status: defined
`

const references = `# References

## <a id="system.ref.std"></a> Std

- Version: 1
- Link: https://example.org/std
- Contact: Ann
`

const capability = "# <a id=\"c\"></a> C\n\n## Goal\n\ng\n\n## Rationale\n\nr\n"

const feature = "# <a id=\"c.f\"></a> F\n\n- Contact: Ann\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n" +
	"## <a id=\"c.f.r\"></a> Rule: R\n\n### Rationale\n\nwhy\n\n```guarded\n" +
	"When      someone reads <id> of a [thing]\n" +
	"Provided  <id> is \"a\"\n" +
	"Then      it is read\n  And     everything else stays as it was\n" +
	"Otherwise it is not read\n  And     everything else stays as it was\n```\n\n" +
	"| id | provided | outcome | why |\n| --- | --- | --- | --- |\n| a | yes | read | |\n| b | no | not read | |\n"

func files(edit func(map[string]string)) []lint.Element {
	f := map[string]string{
		"system.md":                    system,
		"glossary.md":                  glossary,
		"references.md":                references,
		"capabilities/c/capability.md": capability,
		"capabilities/c/features/f.md": feature,
	}
	if edit != nil {
		edit(f)
	}
	m := map[string][]byte{}
	for k, v := range f {
		m[k] = []byte(v)
	}
	return lint.Check(tree.FromFiles("root", m)).Elements
}

func element(t *testing.T, els []lint.Element, id string) lint.Element {
	t.Helper()
	for _, e := range els {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("no element %s", id)
	return lint.Element{}
}

func keys(ps []Part) []string {
	var out []string
	for _, p := range ps {
		k := p.Key
		if strings.HasPrefix(k, "row ") {
			k = "row"
		}
		out = append(out, k)
	}
	return out
}

func TestFingerprint(t *testing.T) {
	same := [][2]string{
		{"a body\nwrapped here", "a body wrapped\nhere"},
		{"Then      lint exits", "Then lint exits"},
		{"  x\t\ty\n", "x y"},
	}
	for _, c := range same {
		if Fingerprint(c[0]) != Fingerprint(c[1]) {
			t.Errorf("%q and %q differ", c[0], c[1])
		}
	}
	differ := [][2]string{{"*a*", "_a_"}, {"- Contact: Bruno", "- Contact: bruno"}, {"a b", "a b"}}
	for _, c := range differ {
		if Fingerprint(c[0]) == Fingerprint(c[1]) {
			t.Errorf("%q and %q are the same", c[0], c[1])
		}
	}
	if fp := Fingerprint(""); !strings.HasPrefix(fp, "h1:") || len(fp) != 67 {
		t.Errorf("%q", fp)
	}
}

func TestParts(t *testing.T) {
	els := files(nil)
	cases := map[string][]string{
		"c.f.r":          {"heading", "section Rationale", "step When", "step Provided", "step Then", "step Otherwise", "row", "row"},
		"c.f":            {"heading", "field Contact", "section Goal", "section Rationale"},
		"system.thing":   {"heading", "field Definition", "field Contact", "field Status", "field Aliases"},
		"system.ref.std": {"heading", "field Version", "field Link", "field Contact"},
	}
	for id, want := range cases {
		if got := keys(Parts(element(t, els, id))); !slices.Equal(got, want) {
			t.Errorf("%s: %q, want %q", id, got, want)
		}
	}
	r := element(t, els, "c.f.r")
	ps := Parts(r)
	if ps[0].Text != "Rule: R" || ps[1].Text != "why\n" || ps[4].Text != "it is read\n  And     everything else stays as it was" {
		t.Errorf("%q", ps[:5])
	}

	// In a rule, a section ends at its rule block: prose after it is body.
	after := strings.Replace(feature, "```\n\n|", "```\n\nmore prose\n\n|", 1)
	r = element(t, files(func(m map[string]string) { m["capabilities/c/features/f.md"] = after }), "c.f.r")
	if r.Sections[0].Body != "why\n" || r.Body != "more prose\n" || Parts(r)[len(Parts(r))-1].Key != "body" {
		t.Errorf("%+v %q", r.Sections, r.Body)
	}
}

func TestRowKey(t *testing.T) {
	key := func(f string) []string {
		var out []string
		for _, p := range Parts(element(t, files(func(m map[string]string) { m["capabilities/c/features/f.md"] = f }), "c.f.r")) {
			if strings.HasPrefix(p.Key, "row ") {
				out = append(out, p.Key)
			}
		}
		return out
	}
	base := key(feature)
	if base[0] == base[1] {
		t.Fatal("two cases share a key")
	}
	if got := key(strings.Replace(feature, "| a | yes | read | |", "| a | yes | still read | |", 1)); !slices.Equal(got, base) {
		t.Error("a new outcome changed the key")
	}
	swapped := strings.NewReplacer("| id | provided | outcome | why |", "| outcome | provided | id | why |",
		"| a | yes | read | |", "| read | yes | a | |", "| b | no | not read | |", "| not read | no | b | |").Replace(feature)
	if got := key(swapped); !slices.Equal(got, base) {
		t.Errorf("swapped columns: %q, want %q", got, base)
	}
	repeated := strings.Replace(feature, "| b | no | not read | |", "| ? | no | not read | |\n| ? | no | x | |", 1)
	got := key(repeated)
	if len(got) != 3 || got[2] != got[1]+" #2" {
		t.Errorf("repeated: %q", got)
	}
}

func TestManifest(t *testing.T) {
	els := files(nil)
	m := Write(els)
	if !strings.HasPrefix(string(m), "model 0\n") || !strings.HasSuffix(string(m), "\n") {
		t.Fatalf("%q", m)
	}
	lines := strings.Split(strings.TrimSuffix(string(m), "\n"), "\n")[1:]
	if !slices.IsSorted(lines) {
		t.Error("not sorted")
	}
	if string(Write(files(nil))) != string(m) {
		t.Error("not the same bytes for the same tree")
	}
	b, bad := Parse(m)
	if len(bad) > 0 || len(b["c.f.r"]) != 8 {
		t.Errorf("%v %v", bad, b["c.f.r"])
	}
	fp := "h1:" + strings.Repeat("a", 64)
	for _, line := range []string{"<<<<<<< HEAD", "lint.run.exit heading " + fp, "lint.run.exit\theading\th1:5b0d", "Lint\theading\t" + fp, "x\tfoo\t" + fp, ""} {
		if _, bad := Parse([]byte("model 0\n" + line + "\n")); len(bad) != 1 {
			t.Errorf("%q read", line)
		}
	}
	if _, bad := Parse([]byte("model 0\nx\theading\t" + fp + "\nx\theading\t" + fp + "\n")); len(bad) != 1 {
		t.Error("a part recorded twice is read")
	}
	if _, bad := Parse([]byte("model 0\nx\theading\t" + fp)); len(bad) != 1 {
		t.Error("a last line without a line feed is read")
	}
	if _, bad := Parse([]byte("model 0\nx\trow " + fp + " #2\t" + fp + "\n")); len(bad) != 0 {
		t.Error(bad)
	}
}

func TestCheckModel(t *testing.T) {
	for in, ok := range map[string]bool{"model 0\nx": true, "model 0": true, "model 0\r\nx": true, "model 1\n": false, "": false, "model 0 \n": false, "model\n": false} {
		if err := CheckModel([]byte(in)); (err == nil) != ok {
			t.Errorf("%q: %v", in, err)
		}
	}
}

func diff(t *testing.T, edit func(map[string]string)) map[string]*Entry {
	t.Helper()
	base, bad := Parse(Write(files(nil)))
	if len(bad) > 0 {
		t.Fatal(bad)
	}
	out := map[string]*Entry{}
	entries := Compare(files(edit), base)
	for i, e := range entries {
		if i > 0 && entries[i-1].ID >= e.ID {
			t.Error("not ordered by identifier")
		}
		out[e.ID] = e
	}
	return out
}

func TestCompare(t *testing.T) {
	if d := diff(t, nil); len(d) != 0 {
		t.Errorf("an unchanged tree: %v", d)
	}
	reflowed := diff(t, func(m map[string]string) {
		m["capabilities/c/features/f.md"] = strings.NewReplacer("Then      it is read", "Then it is read", "### Rationale\n\nwhy\n", "### Rationale\n\n  why\n\n").Replace(feature)
	})
	if len(reflowed) != 0 {
		t.Errorf("reflowed: %v", reflowed)
	}
	d := diff(t, func(m map[string]string) {
		f := strings.Replace(feature, "Then      it is read", "Then      it is read twice", 1)
		f = strings.Replace(f, "### Rationale\n\nwhy\n\n", "", 1)
		f += "| c | no | not read | |\n"
		m["capabilities/c/features/f.md"] = f
		m["capabilities/c/glossary.md"] = "# Glossary\n\n## <a id=\"c.new\"></a> new\n\n- Definition: ?\n- Contact: Ann\n- Status: open\n- Question: what?\n"
	})
	e := d["c.f.r"]
	if e == nil || e.Change != Changed || len(e.Parts) != 3 {
		t.Fatalf("%+v", e)
	}
	var got []string
	for _, p := range e.Parts {
		got = append(got, p.Change+" "+strings.Fields(p.Part)[0])
	}
	if !slices.Contains(got, "changed step") || !slices.Contains(got, "removed section") || !slices.Contains(got, "added row") {
		t.Errorf("%q", got)
	}
	if d["c.new"] == nil || d["c.new"].Change != Added || len(d["c.new"].Undecided()) != 1 {
		t.Errorf("%+v", d["c.new"])
	}
	if d["c.f"] != nil {
		t.Error("a rule's change changed its feature")
	}
	if !NeedsFormer([]*Entry{e}) || NeedsFormer([]*Entry{d["c.new"]}) {
		t.Error("NeedsFormer")
	}

	removed := diff(t, func(m map[string]string) {
		m["capabilities/c/features/f.md"] = feature[:strings.Index(feature, "## <a id=\"c.f.r\">")]
	})
	if removed["c.f.r"] == nil || removed["c.f.r"].Change != Removed || removed["c.f.r"].Element != nil {
		t.Errorf("%+v", removed)
	}
}

func TestCompareTerms(t *testing.T) {
	d := diff(t, func(m map[string]string) {
		m["glossary.md"] = strings.Replace(glossary, "A thing.", "Another thing.", 1)
	})
	if r := d["c.f.r"]; r == nil || r.Change != ChangedByTerm || !slices.Equal(r.Terms, []string{"system.thing"}) {
		t.Errorf("%+v", r)
	}
	d = diff(t, func(m map[string]string) {
		m["glossary.md"] = strings.Replace(glossary, "- Aliases: things", "- Aliases: things, stuff", 1)
	})
	if d["c.f.r"] != nil || d["system.thing"] == nil {
		t.Errorf("a new alias: %v", d)
	}
	d = diff(t, func(m map[string]string) {
		m["glossary.md"] = strings.Replace(glossary, "A thing.", "Another thing.", 1)
		m["capabilities/c/features/f.md"] = strings.Replace(feature, "it is read", "it is read now", 1)
	})
	if r := d["c.f.r"]; r == nil || r.Change != Changed || !slices.Equal(r.Terms, []string{"system.thing"}) {
		t.Errorf("%+v", r)
	}

	d = diff(t, func(m map[string]string) {
		m["references.md"] = strings.Replace(references, "- Version: 1", "- Version: 2", 1)
	})
	if g := d["system.std"]; g == nil || g.Change != Changed || len(g.Parts) != 1 || g.Parts[0].Part != "field Defined by" {
		t.Errorf("a new Version: %+v", g)
	}
	d = diff(t, func(m map[string]string) {
		m["references.md"] = strings.Replace(references, "- Contact: Ann", "- Contact: Bob", 1)
	})
	if d["system.std"] != nil {
		t.Error("a new Contact changed the terms")
	}
}

func TestFormer(t *testing.T) {
	m := Write(files(nil))
	base, _ := Parse(m)
	entries := Compare(files(func(f map[string]string) {
		f["capabilities/c/features/f.md"] = strings.Replace(feature, "Then      it is read", "Then      it is read twice", 1)
	}), base)
	if err := Former(entries, base, files(nil), "rev"); err != nil {
		t.Fatal(err)
	}
	if p := entries[0].Parts[0]; !strings.HasPrefix(p.Former, "it is read\n") {
		t.Errorf("%+v", p)
	}
	other := files(func(f map[string]string) {
		f["capabilities/c/features/f.md"] = strings.Replace(feature, "Then      it is read", "Then      it is read thrice", 1)
	})
	if err := Former(entries, base, other, "rev"); err == nil {
		t.Error("former text that is not the baseline's was taken")
	}
}
