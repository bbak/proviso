package baseline

import (
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/baseline/baselinetest"
)

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
	els := baselinetest.Files(nil)
	cases := map[string][]string{
		"c.f.r":          {"heading", "section Rationale", "step When", "step Provided", "step Then", "step Otherwise", "row", "row"},
		"c.f":            {"heading", "field Contact", "section Goal", "section Rationale"},
		"system.thing":   {"heading", "field Definition", "field Contact", "field Status", "field Aliases"},
		"system.ref.std": {"heading", "field Version", "field Link", "field Contact"},
	}
	for id, want := range cases {
		if got := keys(Parts(baselinetest.Element(t, els, id))); !slices.Equal(got, want) {
			t.Errorf("%s: %q, want %q", id, got, want)
		}
	}
	r := baselinetest.Element(t, els, "c.f.r")
	ps := Parts(r)
	if ps[0].Text != "Rule: R" || ps[1].Text != "why\n" || ps[4].Text != "it is read\n  And     everything else stays as it was" {
		t.Errorf("%q", ps[:5])
	}

	// In a rule, a section ends at its rule block: prose after it is body.
	after := strings.Replace(baselinetest.Feature, "```\n\n|", "```\n\nmore prose\n\n|", 1)
	r = baselinetest.Element(t, baselinetest.Files(func(m map[string]string) { m["capabilities/c/features/f.md"] = after }), "c.f.r")
	if r.Sections[0].Body != "why\n" || r.Body != "more prose\n" || Parts(r)[len(Parts(r))-1].Key != "body" {
		t.Errorf("%+v %q", r.Sections, r.Body)
	}
}

func TestRowKey(t *testing.T) {
	key := func(f string) []string {
		var out []string
		for _, p := range Parts(baselinetest.Element(t, baselinetest.Files(func(m map[string]string) { m["capabilities/c/features/f.md"] = f }), "c.f.r")) {
			if strings.HasPrefix(p.Key, "row ") {
				out = append(out, p.Key)
			}
		}
		return out
	}
	base := key(baselinetest.Feature)
	if base[0] == base[1] {
		t.Fatal("two cases share a key")
	}
	if got := key(strings.Replace(baselinetest.Feature, "| a | yes | read | |", "| a | yes | still read | |", 1)); !slices.Equal(got, base) {
		t.Error("a new outcome changed the key")
	}
	swapped := strings.NewReplacer("| id | provided | outcome | why |", "| outcome | provided | id | why |",
		"| a | yes | read | |", "| read | yes | a | |", "| b | no | not read | |", "| not read | no | b | |").Replace(baselinetest.Feature)
	if got := key(swapped); !slices.Equal(got, base) {
		t.Errorf("swapped columns: %q, want %q", got, base)
	}
	repeated := strings.Replace(baselinetest.Feature, "| b | no | not read | |", "| ? | no | not read | |\n| ? | no | x | |", 1)
	got := key(repeated)
	if len(got) != 3 || got[2] != got[1]+" #2" {
		t.Errorf("repeated: %q", got)
	}
}

func TestManifest(t *testing.T) {
	els := baselinetest.Files(nil)
	m := Write(els)
	if !strings.HasPrefix(string(m), "model 0\n") || !strings.HasSuffix(string(m), "\n") {
		t.Fatalf("%q", m)
	}
	lines := strings.Split(strings.TrimSuffix(string(m), "\n"), "\n")[1:]
	if !slices.IsSorted(lines) {
		t.Error("not sorted")
	}
	if string(Write(baselinetest.Files(nil))) != string(m) {
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
	for in, ok := range map[string]bool{"model 0\nx": true, "model 0": true, "model 0\r\nx": true, "model 0\r": false, "model 0\rfoo\n": false, "model 1\r\r\n": false, "model 1\n": false, "": false, "model 0 \n": false, "model\n": false, "model 00\n": false} {
		if err := CheckModel([]byte(in)); (err == nil) != ok {
			t.Errorf("%q: %v", in, err)
		}
	}
}

func TestNamedModel(t *testing.T) {
	for in, want := range map[string]string{"model 0": "0", "model 12": "12", "model 00": "", "model": "", "model one": "", "": "", "<<<<<<< HEAD": "", "model 1 ": ""} {
		if got, ok := NamedModel(in); got != want || ok != (want != "") {
			t.Errorf("%q: %q %v", in, got, ok)
		}
	}
}

func diff(t *testing.T, edit func(map[string]string)) map[string]*Entry {
	t.Helper()
	base, bad := Parse(Write(baselinetest.Files(nil)))
	if len(bad) > 0 {
		t.Fatal(bad)
	}
	out := map[string]*Entry{}
	now := baselinetest.Files(edit)
	entries := Through(Compare(now, base), now, baselinetest.Files(nil))
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
		m["capabilities/c/features/f.md"] = strings.NewReplacer("Then      it is read", "Then it is read", "### Rationale\n\nwhy\n", "### Rationale\n\n  why\n\n").Replace(baselinetest.Feature)
	})
	if len(reflowed) != 0 {
		t.Errorf("reflowed: %v", reflowed)
	}
	d := diff(t, func(m map[string]string) {
		f := strings.Replace(baselinetest.Feature, "Then      it is read", "Then      it is read twice", 1)
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
	for _, p := range e.Parts {
		switch {
		case p.Change == Removed && p.Text != "",
			p.Change == Changed && !strings.HasPrefix(p.Text, "it is read twice"),
			p.Change == Added && !strings.Contains(p.Text, "| c | no | not read |"):
			t.Errorf("the text now of %+v", p)
		}
	}
	if d["c.new"] == nil || d["c.new"].Change != Added || len(d["c.new"].Undecided()) != 1 {
		t.Errorf("%+v", d["c.new"])
	}
	twice := diff(t, func(m map[string]string) {
		m["capabilities/c/features/f.md"] = strings.Replace(baselinetest.Feature, "| b | no | not read | |", "| b | no | ? | ? |", 1)
	})
	if u := twice["c.f.r"].Undecided(); len(u) != 1 {
		t.Errorf("a row with two undecided cells: %v", u)
	}
	if d["c.f"] != nil {
		t.Error("a rule's change changed its feature")
	}
	if !NeedsFormer([]*Entry{e}) || NeedsFormer([]*Entry{d["c.new"]}) {
		t.Error("NeedsFormer")
	}

	removed := diff(t, func(m map[string]string) {
		m["capabilities/c/features/f.md"] = baselinetest.Feature[:strings.Index(baselinetest.Feature, "## <a id=\"c.f.r\">")]
	})
	if removed["c.f.r"] == nil || removed["c.f.r"].Change != Removed || removed["c.f.r"].Element != nil {
		t.Errorf("%+v", removed)
	}
}

func TestFormer(t *testing.T) {
	m := Write(baselinetest.Files(nil))
	base, _ := Parse(m)
	entries := Compare(baselinetest.Files(func(f map[string]string) {
		f["capabilities/c/features/f.md"] = strings.Replace(baselinetest.Feature, "Then      it is read", "Then      it is read twice", 1)
	}), base)
	if err := Former(entries, base, baselinetest.Files(nil), "rev"); err != nil {
		t.Fatal(err)
	}
	if p := entries[0].Parts[0]; !strings.HasPrefix(p.Former, "it is read\n") {
		t.Errorf("%+v", p)
	}
	other := baselinetest.Files(func(f map[string]string) {
		f["capabilities/c/features/f.md"] = strings.Replace(baselinetest.Feature, "Then      it is read", "Then      it is read thrice", 1)
	})
	if err := Former(entries, base, other, "rev"); err == nil {
		t.Error("former text that is not the baseline's was taken")
	}
	// baseline.diff.former: a part the element diff does not list as
	// changed or removed may differ at the revision.
	entries = Compare(baselinetest.Files(func(f map[string]string) {
		f["capabilities/c/features/f.md"] = strings.Replace(baselinetest.Feature, "Then      it is read", "Then      it is read twice", 1)
	}), base)
	elsewhere := baselinetest.Files(func(f map[string]string) {
		f["capabilities/c/features/f.md"] = strings.Replace(baselinetest.Feature, "\nwhy\n", "\nanother reason\n", 1)
	})
	if err := Former(entries, base, elsewhere, "rev"); err != nil {
		t.Errorf("a difference in a part not listed: %v", err)
	}
}
