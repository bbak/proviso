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
	// Every chain is named, ordered, a chain that starts another first.
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
