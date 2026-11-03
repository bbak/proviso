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
	got := Rewrite(tr, els, renames)
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
	if len(Rewrite(tr, els, nil)) != 0 {
		t.Error("rewritten without renames")
	}
}
