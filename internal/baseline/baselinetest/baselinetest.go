package baselinetest

import (
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

const Feature = "# <a id=\"c.f\"></a> F\n\n- Contact: Ann\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n" +
	"## <a id=\"c.f.r\"></a> Rule: R\n\n### Rationale\n\nwhy\n\n```guarded\n" +
	"When      someone reads <id> of a [thing]\n" +
	"Provided  <id> is \"a\"\n" +
	"Then      it is read\n  And     everything else stays as it was\n" +
	"Otherwise it is not read\n  And     everything else stays as it was\n```\n\n" +
	"| id | provided | outcome | why |\n| --- | --- | --- | --- |\n| a | yes | read | |\n| b | no | not read | |\n"

func Files(edit func(map[string]string)) []lint.Element {
	f := map[string]string{
		"system.md":                    system,
		"glossary.md":                  glossary,
		"references.md":                references,
		"capabilities/c/capability.md": capability,
		"capabilities/c/features/f.md": Feature,
	}
	if edit != nil {
		edit(f)
	}
	m := map[string][]byte{}
	for k, v := range f {
		m[k] = []byte(v)
	}
	return lint.Check(tree.FromFiles("root", m, nil)).Elements
}

func Element(t testing.TB, els []lint.Element, id string) lint.Element {
	t.Helper()
	for _, e := range els {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("no element %s", id)
	return lint.Element{}
}

func Rule(id, when, then string) string {
	return "## <a id=\"" + id + "\"></a> Rule: R\n\n```guarded\nWhen      " + when + "\nThen      " + then + "\n  And     everything else stays as it was\n```\n\n| x | outcome |\n| --- | --- |\n| 1 | done |\n\n"
}

const invariantBody = "### Rationale\n\nr\n\n### Ranges over\n\nall\n\n### Property\n\np\n\n"

func Invariant(id string) string {
	return "## <a id=\"" + id + "\"></a> Invariant: I\n\n" + invariantBody
}

// LevelInvariant is an invariant in an invariants.md, where its heading
// has no prefix.
func LevelInvariant(id string) string {
	return "## <a id=\"" + id + "\"></a> I\n\n" + invariantBody
}

const handOver = "## <a id=\"c.g.d\"></a> Rule: D\n\n```guarded\nWhen      someone writes <x>\nProvided  <x> is 1\nThen      it is written\n  And     everything else stays as it was\nOtherwise → c.f.r\n```\n\n| x | provided | outcome |\n| --- | --- | --- |\n| 1 | yes | written |\n| 2 | no | handed over |\n\n"

// ImpactTree adds to the base tree a second feature with a rule on the
// event of c.f.r marking its term, one referring to the parameter `p`, one
// referring to the contract element `api` and one handing over to c.f.r;
// an invariant in each Feature, one of the capability and one of the
// system; and a contract element representing the term, with a member.
func ImpactTree(m map[string]string) {
	m["capabilities/c/features/f.md"] = Feature + "\n" + Invariant("c.f.inv")
	m["capabilities/c/features/g.md"] = "# <a id=\"c.g\"></a> G\n\n- Contact: Ann\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n" +
		Rule("c.g.a", "someone reads <x> of the [thing]", "it is read") +
		Rule("c.g.b", "`p` is set to <x>", "it is set") +
		Rule("c.g.c", "`api` answers <x>", "it is answered") +
		handOver + Invariant("c.g.inv")
	m["capabilities/c/invariants.md"] = "# Invariants\n\n" + LevelInvariant("c.inv")
	m["invariants.md"] = "# Invariants\n\n" + LevelInvariant("system.inv")
	m["parameters.md"] = "# Parameters\n\n## <a id=\"system.p\"></a> p\n\n- Value: 1\n- Configurable: no\n- Rationale: r\n"
	m["contracts/api.md"] = "# API\n\n## <a id=\"system.api\"></a> api\n\n- Represents: [thing]\n\nThe api.\n\n### <a id=\"system.api.code\"></a> code\n\nA code.\n"
}

func EditTree(edit func(map[string]string)) func(map[string]string) {
	return func(m map[string]string) {
		ImpactTree(m)
		edit(m)
	}
}

func Replace(m map[string]string, file, old, new string) {
	if !strings.Contains(m[file], old) {
		panic(old)
	}
	m[file] = strings.Replace(m[file], old, new, 1)
}
