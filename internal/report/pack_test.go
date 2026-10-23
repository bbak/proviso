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
			{Part: "row h1:ab", Change: baseline.Added},
			{Part: "step Then", Change: baseline.Changed, Former: "it is ```read```"},
		}},
		{ID: "c.f.s", Change: baseline.ChangedThrough, Element: rule, Through: [][]string{{"system.a", "system.b"}, {"system.c"}}},
		{ID: "c.f.t", Change: baseline.Removed, Former: "## <a id=\"c.f.t\"></a> Rule: T\n"},
	}
	packed := []baseline.Packed{{Heading: baseline.HeadingSystem, Element: sys}, {Heading: baseline.HeadingContext, Element: term}}
	var buf bytes.Buffer
	if err := PackText(&buf, "1edea75", entries, packed); err != nil {
		t.Fatal(err)
	}
	want := "# System\n\n" +
		"# <a id=\"system\"></a> S\n\n" +
		"# Amendment\n\n" +
		"Former text from `1edea75`.\n\n" +
		"Changed `c.f.r`: `row h1:ab` added, `step Then` changed.\n" +
		"Not to be built: proviso/capabilities/c/features/f.md:12, proviso/capabilities/c/features/f.md:14.\n\n" +
		"## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n\n" +
		"Former `step Then`:\n\n" +
		"````text\nit is ```read```\n````\n\n" +
		"Changed through `c.f.s`: `system.a` → `system.b`; `system.c`.\n" +
		"Not to be built: proviso/capabilities/c/features/f.md:12, proviso/capabilities/c/features/f.md:14.\n\n" +
		"## <a id=\"c.f.r\"></a> Rule: R\n\nnow\n\n" +
		"Removed `c.f.t`.\n\n" +
		"```text\n## <a id=\"c.f.t\"></a> Rule: T\n```\n\n" +
		"# Context\n\n" +
		"## <a id=\"system.thing\"></a> thing\n"
	if buf.String() != want {
		t.Errorf("got\n%s\nwant\n%s", buf.String(), want)
	}

	buf.Reset()
	if err := PackJSON(&buf, "C:/repo/proviso", "", entries, packed); err != nil {
		t.Fatal(err)
	}
	var p struct {
		Revision *string
		Diff     []map[string]any
		Elements []struct{ ID, Heading, File string }
	}
	if err := json.Unmarshal(buf.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Revision != nil || len(p.Diff) != 3 || p.Diff[2]["change"] != "removed" || len(p.Elements) != 2 || p.Elements[1].Heading != "context" || p.Elements[0].File != sys.File {
		t.Errorf("%s", buf.String())
	}
}
