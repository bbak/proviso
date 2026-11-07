package rehome

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

- Defined by: [Std](references.md#c.ref.std)
- Contact: Ann
- Status: defined
- Aliases: things
`

const references = `# References

## <a id="c.ref.std"></a> Std

- Version: 1
- Link: https://example.org/std
- Contact: Ann
`

const capability = "# <a id=\"c\"></a> C\n\n## Goal\n\ng\n\n## Rationale\n\nr\n"

const other = "# <a id=\"k\"></a> K\n\n## Goal\n\ng\n\n## Rationale\n\nSee [s](../c/features/f.md#x.f.s), `[s](f.md#x.f.s)`, [t][d].\n\n[d]: ../c/features/f.md#x.f.s\n"

const feature = "# <a id=\"c.f\"></a> F\n\n- Contact: Ann\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n" +
	"## <a id=\"c.f.r\"></a> Rule: R\n\n```guarded\n" +
	"When      someone reads <id> of a [thing]\n" +
	"Provided  <id> is \"a\"\n" +
	"Then      it is read\n  And     everything else stays as it was\n" +
	"Otherwise → x.f.s\n```\n\n" +
	"| id | provided | outcome | why |\n| --- | --- | --- | --- |\n| a | yes | read | |\n| b | no | handed over | [x.f.s](#x.f.s) |\n\n" +
	"## <a id=\"x.f.s\"></a> Rule: S\n\n```guarded\n" +
	"When      someone reads <id> of a [thing]\n" +
	"Then      it is not read\n  And     everything else stays as it was\n```\n\n" +
	"| id | outcome |\n| --- | --- |\n| b | not read |\n"

const questions = "# Questions\n\n- [x.f.s] Should it be read? (ask: Ann)\n- [c.f.r] Fine?\n"

func load(t *testing.T, edit func(map[string]string)) (*tree.Tree, []lint.Element) {
	t.Helper()
	f := map[string]string{
		"system.md":                    system,
		"glossary.md":                  glossary,
		"references.md":                references,
		"capabilities/c/capability.md": capability,
		"capabilities/c/features/f.md": feature,
		"capabilities/c/questions.md":  questions,
		"capabilities/d/capability.md": other,
	}
	if edit != nil {
		edit(f)
	}
	m := map[string][]byte{}
	for k, v := range f {
		m[k] = []byte(v)
	}
	tr := tree.FromFiles("root", m, nil)
	res := lint.Check(tr)
	if lint.HasError(res.Findings) {
		t.Fatalf("lint: %+v", res.Findings)
	}
	return tr, res.Elements
}

func TestPlan(t *testing.T) {
	_, els := load(t, nil)
	renames, collisions := Plan(els)
	// k stands in the folder d, and is its own identifier home.
	want := []Rename{{"c.ref.std", "system.ref.std"}, {"x.f.s", "c.f.s"}}
	if !slices.Equal(renames, want) || collisions != nil {
		t.Errorf("renames %v collisions %v", renames, collisions)
	}
}

func TestCollisions(t *testing.T) {
	_, els := load(t, func(f map[string]string) {
		// system.ref.std is taken, and x.f.s and y.f.s both go to c.f.s.
		f["references.md"] += "\n## <a id=\"system.ref.std\"></a> Std 2\n\n- Version: 1\n- Link: https://example.org/std2\n- Contact: Ann\n"
		f["capabilities/c/features/f.md"] += "\n## <a id=\"y.f.s\"></a> Rule: T\n\n```guarded\nWhen      someone writes <id>\nThen      it is written\n  And     everything else stays as it was\n```\n\n| id | outcome |\n| --- | --- |\n| b | written |\n"
	})
	renames, collisions := Plan(els)
	if len(renames) != 3 {
		t.Errorf("renames %v", renames)
	}
	want := []Collision{{"c.f.s", []string{"x.f.s", "y.f.s"}}, {"system.ref.std", []string{"c.ref.std"}}}
	if len(collisions) != 2 || collisions[0].Identifier != want[0].Identifier || !slices.Equal(collisions[0].From, want[0].From) ||
		collisions[1].Identifier != want[1].Identifier || !slices.Equal(collisions[1].From, want[1].From) {
		t.Errorf("collisions %v", collisions)
	}
}

func TestRewrite(t *testing.T) {
	tr, els := load(t, nil)
	renames, _ := Plan(els)
	got := Rewrite(tr, els, renames, nil)
	if len(got) != 5 {
		t.Errorf("%d files rewritten", len(got))
	}
	f := string(got["capabilities/c/features/f.md"])
	for _, s := range []string{`<a id="c.f.s"></a>`, "Otherwise → c.f.s\n", "| [x.f.s](#x.f.s) |"} {
		if !strings.Contains(f, s) {
			t.Errorf("feature lacks %q", s)
		}
	}
	if s := string(got["capabilities/c/questions.md"]); s != strings.Replace(questions, "[x.f.s]", "[c.f.s]", 1) {
		t.Errorf("questions %q", s)
	}
	if s := string(got["capabilities/d/capability.md"]); !strings.Contains(s, "(../c/features/f.md#c.f.s), `[s](f.md#x.f.s)`, [t][d].\n\n[d]: ../c/features/f.md#c.f.s\n") {
		t.Errorf("capability %q", s)
	}
	if s := string(got["glossary.md"]); !strings.Contains(s, "(references.md#system.ref.std)") {
		t.Errorf("glossary %q", s)
	}
	// The manifest of the rewritten tree is that of a tree lint reads with
	// the new identifiers.
	m := string(Manifest(tr, got))
	if !strings.Contains(m, "\nc.f.s\theading\t") || strings.Contains(m, "x.f.s\t") || !strings.Contains(m, "\nsystem.ref.std\t") {
		t.Errorf("manifest %q", m)
	}
	if len(Rewrite(tr, els, nil, nil)) != 0 {
		t.Error("rewritten without renames")
	}
}

// mentioned adds identifier mentions of x.f.s: twice in one Examples row,
// once in a Rationale, once in a question, and places that are none.
func mentioned(f map[string]string) {
	f["capabilities/c/features/f.md"] = strings.Replace(f["capabilities/c/features/f.md"], "| a | yes | read | |", "| a | yes | read | `x.f.s` and ` x.f.s ` |", 1)
	f["capabilities/c/features/f.md"] = strings.Replace(f["capabilities/c/features/f.md"], "## Rationale\n\nr\n", "## Rationale\n\nSee `x.f.s`, `x.f.s.y`, x.f.s and `<a id=\"x.f.s\"></a>`.\n\n```text\n`x.f.s`\n```\n", 1)
	f["capabilities/c/questions.md"] += "- [c.f.r] And `x.f.s`?\n"
}

func TestMentions(t *testing.T) {
	tr, els := load(t, mentioned)
	renames, _ := Plan(els)
	got := Mentions(tr, els, renames)
	want := []Mention{
		{"x.f.s", "proviso/capabilities/c/features/f.md", 11, 1, "c.f"},
		{"x.f.s", "proviso/capabilities/c/features/f.md", 29, 2, "c.f.r"},
		{"x.f.s", "proviso/capabilities/c/questions.md", 5, 1, ""},
	}
	if !slices.Equal(got, want) {
		t.Errorf("mentions %v", got)
	}
	if len(Mentions(tr, els, nil)) != 0 {
		t.Error("mentions without renames")
	}
}

func TestRewriteMentions(t *testing.T) {
	tr, els := load(t, mentioned)
	renames, _ := Plan(els)
	f := "proviso/capabilities/c/features/f.md"
	got := Rewrite(tr, els, renames, nil)
	if s := string(got["capabilities/c/features/f.md"]); !strings.Contains(s, "`x.f.s` and ` x.f.s `") || !strings.Contains(s, "See `x.f.s`") {
		t.Errorf("unpicked mentions renamed: %q", s)
	}
	got = Rewrite(tr, els, renames, []Picked{{"x.f.s", f, 29}})
	s := string(got["capabilities/c/features/f.md"])
	if !strings.Contains(s, "| a | yes | read | `c.f.s` and ` c.f.s ` |") || !strings.Contains(s, "See `x.f.s`, `x.f.s.y`, x.f.s and `<a id=\"x.f.s\"></a>`") || !strings.Contains(s, "```text\n`x.f.s`\n```") {
		t.Errorf("picked line: %q", s)
	}
	if s := string(Rewrite(tr, els, renames, []Picked{{"x.f.s", "proviso/capabilities/c/questions.md", 5}})["capabilities/c/questions.md"]); !strings.HasSuffix(s, "- [c.f.r] And `c.f.s`?\n") {
		t.Errorf("question %q", s)
	}
}

func TestParseMentionList(t *testing.T) {
	got, err := ParseMentionList([]byte(`{"mentions": [{"identifier": "a.b", "file": "proviso/f.md", "line": 3, "count": 2, "element": null}, {"identifier": "a.b", "file": "proviso/f.md", "line": 3}], "root": "x"}`))
	if err != nil || !slices.Equal(got, []Picked{{"a.b", "proviso/f.md", 3}}) {
		t.Errorf("%v %v", got, err)
	}
	if got, err := ParseMentionList([]byte(`{"mentions": []}`)); err != nil || len(got) != 0 {
		t.Errorf("empty: %v %v", got, err)
	}
	for _, in := range []string{`{`, `[]`, `{}`, `{"mentions": {}}`, `{"mentions": [1]}`, `{"mentions": [{"file": "f", "line": 1}]}`,
		`{"mentions": [{"identifier": "a", "file": "f", "line": 0}]}`, `{"mentions": [{"identifier": "a", "file": "f", "line": 1.5}]}`,
		`{"mentions": [{"identifier": "a", "file": "f", "line": "1"}]}`, `{"mentions": []} {}`, "\xff"} {
		if _, err := ParseMentionList([]byte(in)); err == nil {
			t.Errorf("%q taken", in)
		}
	}
}

func TestUnlisted(t *testing.T) {
	listed := []Mention{{"a.b", "f", 3, 1, ""}}
	if u := Unlisted([]Picked{{"a.b", "f", 3}, {"a.b", "f", 4}, {"c.d", "f", 3}}, listed); !slices.Equal(u, []Picked{{"a.b", "f", 4}, {"c.d", "f", 3}}) {
		t.Errorf("%v", u)
	}
}
