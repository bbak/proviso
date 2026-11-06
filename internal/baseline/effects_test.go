package baseline

import (
	"fmt"
	"testing"

	"github.com/bbak/proviso/internal/baseline/baselinetest"
)

// through compares the edited tree with the baseline of impactTree and
// returns how each element is listed, with the chains through which an
// element changed through a dependency.
func through(t *testing.T, edit func(map[string]string)) map[string]string {
	t.Helper()
	before := baselinetest.Files(baselinetest.ImpactTree)
	base, _ := Parse(Write(before))
	out := map[string]string{}
	now := baselinetest.Files(baselinetest.EditTree(edit))
	for _, e := range Through(Compare(now, base), now, before) {
		out[e.ID] = e.Change
		if e.Change == ChangedThrough {
			out[e.ID] = fmt.Sprint(e.Through)
		}
	}
	return out
}

func TestDependencies(t *testing.T) {
	got := through(t, func(m map[string]string) { baselinetest.Replace(m, "glossary.md", "A thing.", "Another thing.") })
	want := map[string]string{
		"system.thing": Changed,
		"c.f.r":        "[[system.thing]]",
		"c.g.a":        "[[system.thing]]",
		"system.api":   "[[system.thing]]",
		"c.g.c":        "[[system.thing system.api]]",
		"c.g.d":        "[[system.thing c.f.r]]",
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("a new Definition:\n%v\nwant\n%v", got, want)
	}
	got = through(t, func(m map[string]string) { baselinetest.Replace(m, "references.md", "- Version: 1", "- Version: 2") })
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.ref.std": Changed, "system.std": "[[system.ref.std]]"}) {
		t.Errorf("a new Version: %v", got)
	}
	got = through(t, func(m map[string]string) { baselinetest.Replace(m, "contracts/api.md", "A code.", "A new code.") })
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.api.code": Changed, "c.g.c": "[[system.api.code]]"}) {
		t.Errorf("a member: %v", got)
	}
	got = through(t, func(m map[string]string) { baselinetest.Replace(m, "parameters.md", "- Value: 1", "- Value: 2") })
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.p": Changed, "c.g.b": "[[system.p]]"}) {
		t.Errorf("a parameter: %v", got)
	}
	got = through(t, func(m map[string]string) {
		baselinetest.Replace(m, "capabilities/c/features/f.md", "Then      it is read", "Then      it is read twice")
	})
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"c.f.r": Changed, "c.g.d": "[[c.f.r]]"}) {
		t.Errorf("a hand-over: %v", got)
	}
	// An Examples row decides as much as a step.
	got = through(t, func(m map[string]string) {
		baselinetest.Replace(m, "capabilities/c/features/f.md", "| b | no | not read | |", "| b | no | not read at all | |")
	})
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"c.f.r": Changed, "c.g.d": "[[c.f.r]]"}) {
		t.Errorf("a row: %v", got)
	}
	// A listed element passes on the meaning that changed through what it
	// depends on.
	got = through(t, func(m map[string]string) {
		baselinetest.Replace(m, "glossary.md", "A thing.", "Another thing.")
		baselinetest.Replace(m, "capabilities/c/features/f.md", "why", "why not")
	})
	if got["c.f.r"] != Changed || got["c.g.d"] != "[[system.thing c.f.r]]" {
		t.Errorf("through a listed rule: %v", got)
	}
	// A chain is named from each changed element, ordered by the first.
	got = through(t, func(m map[string]string) {
		baselinetest.Replace(m, "glossary.md", "A thing.", "Another thing.")
		baselinetest.Replace(m, "contracts/api.md", "The api.", "The new api.")
	})
	if got["c.g.c"] != "[[system.api] [system.thing system.api]]" {
		t.Errorf("two chains: %v", got)
	}
	// A removed member passes its change to the rules referring to the
	// contract element that held it.
	got = through(t, func(m map[string]string) {
		baselinetest.Replace(m, "contracts/api.md", "\n### <a id=\"system.api.code\"></a> code\n\nA code.\n", "")
	})
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.api.code": Removed, "c.g.c": "[[system.api.code]]"}) {
		t.Errorf("a removed member: %v", got)
	}
	for _, edit := range []func(map[string]string){
		func(m map[string]string) {
			baselinetest.Replace(m, "glossary.md", "- Aliases: things", "- Aliases: things, stuff")
		},
		func(m map[string]string) {
			baselinetest.Replace(m, "references.md", "- Contact: Ann", "- Contact: Bob")
		},
		func(m map[string]string) {
			baselinetest.Replace(m, "parameters.md", "- Rationale: r", "- Rationale: s")
		},
	} {
		if got := through(t, edit); len(got) != 1 {
			t.Errorf("no dependents: %v", got)
		}
	}
}

// entry is a glossary entry of the system whose Definition reads def.
func entry(id, name, def string) string {
	return "\n## <a id=\"system." + id + "\"></a> " + name + "\n\n- Definition: " + def + "\n- Contact: Ann\n- Status: defined\n"
}

// throughGlossary compares the impact tree with the entries added to its
// glossary against the baseline of the same tree with the entries as
// former holds them, each tree edited further by edits.
func throughGlossary(t *testing.T, former, now string, edits ...func(map[string]string)) map[string]string {
	t.Helper()
	tree := func(entries string) func(map[string]string) {
		return baselinetest.EditTree(func(m map[string]string) {
			m["glossary.md"] += entries
			for _, edit := range edits {
				edit(m)
			}
		})
	}
	before := baselinetest.Files(tree(former))
	base, _ := Parse(Write(before))
	nowEls := baselinetest.Files(tree(now))
	out := map[string]string{}
	for _, e := range Through(Compare(nowEls, base), nowEls, before) {
		out[e.ID] = e.Change
		if e.Change == ChangedThrough {
			out[e.ID] = fmt.Sprint(e.Through)
		}
	}
	return out
}

func TestRepresentsMarks(t *testing.T) {
	// A contract element depends on every term it represents, not only the
	// first its field marks.
	represents := func(m map[string]string) {
		baselinetest.Replace(m, "contracts/api.md", "- Represents: [thing]", "- Represents: [thing] in a [tree]")
	}
	got := throughGlossary(t, entry("tree", "tree", "A tree."), entry("tree", "tree", "A new tree."), represents)
	want := map[string]string{
		"c.g.c":       "[[system.tree system.api]]",
		"system.api":  "[[system.tree]]",
		"system.tree": Changed,
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("a second represented term:\n%v\nwant\n%v", got, want)
	}
}

func TestDefinitionMarks(t *testing.T) {
	// A definition that marks a term relies on its meaning; a rule
	// marking that entry is reached through both.
	got := throughGlossary(t,
		entry("step", "step", "A line.")+entry("guard", "guard", "The [step] that guards; [guard] itself."),
		entry("step", "step", "A new line.")+entry("guard", "guard", "The [step] that guards; [guard] itself."))
	if got["system.step"] != Changed || got["system.guard"] != "[[system.step]]" || len(got) != 2 {
		t.Errorf("a definition marking a term: %v", got)
	}
	// Two entries marking each other: a circle adds no chain of its own.
	got = throughGlossary(t,
		entry("step", "step", "A [guard] line.")+entry("guard", "guard", "The [step] that guards; [keyword] too.")+entry("keyword", "keyword", "Of a [guard]."),
		entry("step", "step", "A new [guard] line.")+entry("guard", "guard", "The [step] that guards; [keyword] too.")+entry("keyword", "keyword", "Of a [guard]."))
	want := map[string]string{
		"system.step":    Changed,
		"system.guard":   "[[system.step]]",
		"system.keyword": "[[system.step system.guard]]",
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("a circle:\n%v\nwant\n%v", got, want)
	}
	// Only the shortest chain from each changed element is named.
	got = throughGlossary(t,
		entry("step", "step", "A line.")+entry("guard", "guard", "The [step] that guards.")+entry("user", "user", "A [guard] and a [step]."),
		entry("step", "step", "A new line.")+entry("guard", "guard", "The [step] that guards.")+entry("user", "user", "A [guard] and a [step]."))
	if got["system.user"] != "[[system.step]]" {
		t.Errorf("the shortest chain: %v", got)
	}
	// Of equally short chains the first is named.
	got = throughGlossary(t,
		entry("a", "a", "A line.")+entry("c", "c", "Of [a].")+entry("b", "b", "Of [a].")+entry("user", "user", "A [c] and a [b]."),
		entry("a", "a", "A new line.")+entry("c", "c", "Of [a].")+entry("b", "b", "Of [a].")+entry("user", "user", "A [c] and a [b]."))
	if got["system.user"] != "[[system.a system.b]]" {
		t.Errorf("the first of equally short chains: %v", got)
	}
	// Deviations pass the changed meaning on as a Definition does.
	dev := func(d string) string {
		return "\n## <a id=\"system.cm\"></a> cm\n\n- Defined by: [Std](references.md#system.ref.std)\n- Deviations: " + d + "\n- Contact: Ann\n- Status: defined\n" +
			entry("user", "user", "Reads [cm].")
	}
	got = throughGlossary(t, dev("none"), dev("some"))
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.cm": Changed, "system.user": "[[system.cm]]"}) {
		t.Errorf("deviations: %v", got)
	}
	// A deviation relies on the terms it marks.
	got = throughGlossary(t, entry("step", "step", "A line.")+dev("a [step]"), entry("step", "step", "A new line.")+dev("a [step]"))
	want = map[string]string{
		"system.cm":   "[[system.step]]",
		"system.step": Changed,
		"system.user": "[[system.step system.cm]]",
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("a deviation marking a term:\n%v\nwant\n%v", got, want)
	}
}
