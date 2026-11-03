package report

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/lint"
)

func TestPackText(t *testing.T) {
	sys := &lint.Element{ID: "system", File: "proviso/system.md", Line: 1, Text: "# <a id=\"system\"></a> S\n"}
	rule := &lint.Element{ID: "c.f.r", File: "proviso/capabilities/c/features/f.md", Line: 9, Text: "## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n",
		Undecided: []lint.Pos{{File: "proviso/capabilities/c/features/f.md", Line: 12}, {File: "proviso/capabilities/c/features/f.md", Line: 14}}}
	term := &lint.Element{ID: "system.thing", File: "proviso/glossary.md", Line: 3, Text: "## <a id=\"system.thing\"></a> thing\n"}
	entries := []*baseline.Entry{
		{ID: "c.f.r", Change: baseline.Changed, Element: rule, Parts: []baseline.PartChange{
			{Part: "row h1:ab", Change: baseline.Added, Text: "| a | yes | read | |\n"},
			{Part: "section Rationale", Change: baseline.Removed, Former: "why\n"},
			{Part: "step Then", Change: baseline.Changed, Text: "it is read twice", Former: "it is ```read```"},
		}},
		{ID: "c.f.s", Change: baseline.ChangedThrough, Element: rule, Through: [][]string{{"system.a", "system.b"}, {"system.c"}}},
		{ID: "c.f.t", Change: baseline.Removed, Former: "## <a id=\"c.f.t\"></a> Rule: T\n"},
	}
	packed := []baseline.Packed{{Heading: baseline.HeadingSystem, Element: sys}, {Heading: baseline.HeadingContext, Element: term}}
	var buf bytes.Buffer
	if err := PackText(&buf, "1edea75", entries, "- refunds\n", packed); err != nil {
		t.Fatal(err)
	}
	want := "# System\n\n" +
		"# <a id=\"system\"></a> S\n\n" +
		"# Amendment\n\n" +
		"Former text from `1edea75`.\n\n" +
		"Changed `c.f.r`: `row h1:ab` added, `section Rationale` removed, `step Then` changed.\n" +
		"Not to be built: proviso/capabilities/c/features/f.md:12, proviso/capabilities/c/features/f.md:14.\n\n" +
		"## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n\n" +
		"Now `row h1:ab`:\n\n" +
		"```text\n| a | yes | read | |\n```\n\n" +
		"Former `section Rationale`:\n\n" +
		"```text\nwhy\n```\n\n" +
		"Now `step Then`:\n\n" +
		"```text\nit is read twice\n```\n\n" +
		"Former `step Then`:\n\n" +
		"````text\nit is ```read```\n````\n\n" +
		"Changed through `c.f.s`: `system.a` → `system.b`; `system.c`.\n" +
		"Not to be built: proviso/capabilities/c/features/f.md:12, proviso/capabilities/c/features/f.md:14.\n\n" +
		"## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n\n" +
		"Removed `c.f.t`.\n\n" +
		"```text\n## <a id=\"c.f.t\"></a> Rule: T\n```\n\n" +
		"# Not in this amendment\n\n" +
		"- refunds\n\n" +
		"# Context\n\n" +
		"## <a id=\"system.thing\"></a> thing\n"
	if buf.String() != want {
		t.Errorf("got\n%s\nwant\n%s", buf.String(), want)
	}

	buf.Reset()
	if err := PackJSON(&buf, "C:/repo/proviso", "", entries, "", packed); err != nil {
		t.Fatal(err)
	}
	var p struct {
		Revision *string
		Note     *string
		Diff     []map[string]any
		Elements []struct{ ID, Heading, File string }
	}
	if err := json.Unmarshal(buf.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Revision != nil || p.Note != nil || len(p.Diff) != 3 || p.Diff[2]["change"] != "removed" || len(p.Elements) != 2 || p.Elements[1].Heading != "context" || p.Elements[0].File != sys.File {
		t.Errorf("%s", buf.String())
	}
	parts := p.Diff[0]["parts"].([]any)
	added, removed, changed := parts[0].(map[string]any), parts[1].(map[string]any), parts[2].(map[string]any)
	if _, ok := added["former"]; ok || added["text"] != "| a | yes | read | |\n" {
		t.Errorf("an added part: %v", added)
	}
	if _, ok := removed["text"]; ok || removed["former"] != "why\n" {
		t.Errorf("a removed part: %v", removed)
	}
	if changed["text"] != "it is read twice" || changed["former"] != "it is ```read```" {
		t.Errorf("a changed part: %v", changed)
	}
}

func TestPackTextElement(t *testing.T) {
	sys := &lint.Element{ID: "system", File: "proviso/system.md", Line: 1, Text: "# <a id=\"system\"></a> S\n"}
	rule := &lint.Element{ID: "c.f.r", File: "proviso/capabilities/c/features/f.md", Line: 9, Text: "## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n"}
	packed := []baseline.Packed{{Heading: baseline.HeadingSystem, Element: sys}, {Heading: baseline.HeadingElement, Element: rule}}
	var buf bytes.Buffer
	if err := PackText(&buf, "", nil, "", packed); err != nil {
		t.Fatal(err)
	}
	want := "# System\n\n" +
		"# <a id=\"system\"></a> S\n\n" +
		"# Element\n\n" +
		"## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n"
	if buf.String() != want {
		t.Errorf("got\n%s\nwant\n%s", buf.String(), want)
	}

	buf.Reset()
	if err := PackJSON(&buf, "C:/repo/proviso", "", nil, "", packed); err != nil {
		t.Fatal(err)
	}
	var p map[string]any
	if err := json.Unmarshal(buf.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if d, ok := p["diff"].([]any); !ok || len(d) != 0 || p["note"] != nil || p["revision"] != nil {
		t.Errorf("%s", buf.String())
	}
	if els := p["elements"].([]any); len(els) != 2 || els[1].(map[string]any)["heading"] != "element" {
		t.Errorf("%s", buf.String())
	}
}
