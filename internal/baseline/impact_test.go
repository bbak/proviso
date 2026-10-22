package baseline

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/lint"
)

func rule(id, when, then string) string {
	return "## <a id=\"" + id + "\"></a> Rule: R\n\n```guarded\nWhen      " + when + "\nThen      " + then + "\n  And     everything else stays as it was\n```\n\n| x | outcome |\n| --- | --- |\n| 1 | done |\n\n"
}

const invariantBody = "### Rationale\n\nr\n\n### Ranges over\n\nall\n\n### Property\n\np\n\n"

func invariant(id string) string {
	return "## <a id=\"" + id + "\"></a> Invariant: I\n\n" + invariantBody
}

// levelInvariant is an invariant in an invariants.md, where its heading
// has no prefix.
func levelInvariant(id string) string {
	return "## <a id=\"" + id + "\"></a> I\n\n" + invariantBody
}

const handOver = "## <a id=\"c.g.d\"></a> Rule: D\n\n```guarded\nWhen      someone writes <x>\nProvided  <x> is 1\nThen      it is written\n  And     everything else stays as it was\nOtherwise → c.f.r\n```\n\n| x | provided | outcome |\n| --- | --- | --- |\n| 1 | yes | written |\n| 2 | no | handed over |\n\n"

// impactTree adds to the base tree a second feature with a rule on the
// event of c.f.r marking its term, one referring to the parameter `p`, one
// referring to the contract element `api` and one handing over to c.f.r;
// an invariant in each feature, one of the capability and one of the
// system; and a contract element representing the term, with a member.
func impactTree(m map[string]string) {
	m["capabilities/c/features/f.md"] = feature + "\n" + invariant("c.f.inv")
	m["capabilities/c/features/g.md"] = "# <a id=\"c.g\"></a> G\n\n- Contact: Ann\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n" +
		rule("c.g.a", "someone reads <x> of the [thing]", "it is read") +
		rule("c.g.b", "`p` is set to <x>", "it is set") +
		rule("c.g.c", "`api` answers <x>", "it is answered") +
		handOver + invariant("c.g.inv")
	m["capabilities/c/invariants.md"] = "# Invariants\n\n" + levelInvariant("c.inv")
	m["invariants.md"] = "# Invariants\n\n" + levelInvariant("system.inv")
	m["parameters.md"] = "# Parameters\n\n## <a id=\"system.p\"></a> p\n\n- Value: 1\n- Configurable: no\n- Rationale: r\n"
	m["contracts/api.md"] = "# API\n\n## <a id=\"system.api\"></a> api\n\n- Represents: [thing]\n\nThe api.\n\n### <a id=\"system.api.code\"></a> code\n\nA code.\n"
}

func editTree(edit func(map[string]string)) func(map[string]string) {
	return func(m map[string]string) {
		impactTree(m)
		edit(m)
	}
}

func replace(m map[string]string, file, old, new string) {
	if !strings.Contains(m[file], old) {
		panic(old)
	}
	m[file] = strings.Replace(m[file], old, new, 1)
}

// through compares the edited tree with the baseline of impactTree and
// returns how each element is listed, with the chains through which an
// element changed through a dependency.
func through(t *testing.T, edit func(map[string]string)) map[string]string {
	t.Helper()
	base, _ := Parse(Write(files(impactTree)))
	out := map[string]string{}
	for _, e := range Compare(files(editTree(edit)), base) {
		out[e.ID] = e.Change
		if e.Change == ChangedThrough {
			out[e.ID] = fmt.Sprint(e.Through)
		}
	}
	return out
}

func TestDependencies(t *testing.T) {
	got := through(t, func(m map[string]string) { replace(m, "glossary.md", "A thing.", "Another thing.") })
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
	got = through(t, func(m map[string]string) { replace(m, "references.md", "- Version: 1", "- Version: 2") })
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.ref.std": Changed, "system.std": "[[system.ref.std]]"}) {
		t.Errorf("a new Version: %v", got)
	}
	got = through(t, func(m map[string]string) { replace(m, "contracts/api.md", "A code.", "A new code.") })
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.api.code": Changed, "c.g.c": "[[system.api.code]]"}) {
		t.Errorf("a member: %v", got)
	}
	got = through(t, func(m map[string]string) { replace(m, "parameters.md", "- Value: 1", "- Value: 2") })
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"system.p": Changed, "c.g.b": "[[system.p]]"}) {
		t.Errorf("a parameter: %v", got)
	}
	got = through(t, func(m map[string]string) {
		replace(m, "capabilities/c/features/f.md", "Then      it is read", "Then      it is read twice")
	})
	if fmt.Sprint(got) != fmt.Sprint(map[string]string{"c.f.r": Changed, "c.g.d": "[[c.f.r]]"}) {
		t.Errorf("a hand-over: %v", got)
	}
	// A listed element passes on the meaning that changed through what it
	// depends on.
	got = through(t, func(m map[string]string) {
		replace(m, "glossary.md", "A thing.", "Another thing.")
		replace(m, "capabilities/c/features/f.md", "why", "why not")
	})
	if got["c.f.r"] != Changed || got["c.g.d"] != "[[system.thing c.f.r]]" {
		t.Errorf("through a listed rule: %v", got)
	}
	for _, edit := range []func(map[string]string){
		func(m map[string]string) { replace(m, "glossary.md", "- Aliases: things", "- Aliases: things, stuff") },
		func(m map[string]string) { replace(m, "references.md", "- Contact: Ann", "- Contact: Bob") },
		func(m map[string]string) { replace(m, "parameters.md", "- Rationale: r", "- Rationale: s") },
	} {
		if got := through(t, edit); len(got) != 1 {
			t.Errorf("no dependents: %v", got)
		}
	}
}

// impact compares the edited tree with the baseline of impactTree, and
// returns each candidate's reasons as "source via from".
func impact(t *testing.T, edit func(map[string]string)) (map[string][]string, []Frame) {
	t.Helper()
	before := files(impactTree)
	base, bad := Parse(Write(before))
	if len(bad) > 0 {
		t.Fatal(bad)
	}
	now := files(editTree(edit))
	entries := Compare(now, base)
	if err := Former(entries, base, before, "rev"); err != nil {
		t.Fatal(err)
	}
	cands, frames := Impact(entries, now, before)
	out := map[string][]string{}
	for i, c := range cands {
		if i > 0 && cands[i-1].Element.ID >= c.Element.ID {
			t.Error("not ordered by identifier")
		}
		for _, r := range c.Reasons {
			out[c.Element.ID] = append(out[c.Element.ID], fmt.Sprintf("%s %s %s", r.Source, r.Via, strings.Join(r.From, ",")))
		}
	}
	return out, frames
}

func TestImpactRuleChanged(t *testing.T) {
	got, frames := impact(t, func(m map[string]string) {
		replace(m, "capabilities/c/features/f.md", "Then      it is read", "Then      it is read twice")
	})
	want := map[string][]string{
		"c.g.a":      {"events someone reads c.f.r"},
		"c.f.inv":    {"ranges feature c.f.r"},
		"c.g.inv":    {"ranges feature c.g.d"},
		"c.inv":      {"ranges capability c.f.r,c.g.d"},
		"system.inv": {"ranges system c.f.r,c.g.d"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) || len(frames) != 0 {
		t.Errorf("%v\nwant\n%v", got, want)
	}
}

func TestImpactThroughDefinition(t *testing.T) {
	got, _ := impact(t, func(m map[string]string) { replace(m, "glossary.md", "A thing.", "Another thing.") })
	// The rules marking the term are listed, not candidates; every part of
	// theirs counts as changed, so the invariants above them are found.
	want := map[string][]string{
		"c.f.inv":    {"ranges feature c.f.r"},
		"c.g.inv":    {"ranges feature c.g.a,c.g.c,c.g.d"},
		"c.inv":      {"ranges capability c.f.r,c.g.a,c.g.c,c.g.d"},
		"system.inv": {"ranges system c.f.r,c.g.a,c.g.c,c.g.d"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("%v\nwant\n%v", got, want)
	}
}

func TestImpactScope(t *testing.T) {
	scope := func(edit func(map[string]string)) []Frame {
		_, frames := impact(t, edit)
		return frames
	}
	all := []string{"c.f.inv", "c.f.r", "c.g.a", "c.g.b", "c.g.c", "c.g.d", "c.g.inv", "c.inv"}
	cases := []struct {
		name  string
		edit  func(map[string]string)
		frame string
		want  []string
	}{
		{"a capability's Goal", func(m map[string]string) { replace(m, "capabilities/c/capability.md", "## Goal\n\ng", "## Goal\n\nh") }, "c", all},
		{"a feature's Goal", func(m map[string]string) { replace(m, "capabilities/c/features/g.md", "## Goal\n\ng", "## Goal\n\nh") }, "c.g", []string{"c.g.a", "c.g.b", "c.g.c", "c.g.d", "c.g.inv"}},
		{"an invariant of the system", func(m map[string]string) { replace(m, "invariants.md", "### Property\n\np", "### Property\n\nq") }, "system.inv", all},
		{"an invariant of the capability", func(m map[string]string) {
			replace(m, "capabilities/c/invariants.md", "### Property\n\np", "### Property\n\nq")
		}, "c.inv", all[:7]},
		{"an invariant of a feature", func(m map[string]string) {
			replace(m, "capabilities/c/features/f.md", "### Property\n\np", "### Property\n\nq")
		}, "c.f.inv", []string{"c.f.r"}},
	}
	for _, c := range cases {
		frames := scope(c.edit)
		if len(frames) != 1 || frames[0].ID != c.frame || !slices.Equal(frames[0].Candidates, c.want) {
			t.Errorf("%s: %v", c.name, frames)
		}
	}
	got, frames := impact(t, func(m map[string]string) {
		replace(m, "capabilities/c/features/f.md", "- Contact: Ann", "- Contact: Bob")
	})
	if len(got) != 0 || len(frames) != 0 {
		t.Errorf("a new Contact: %v %v", got, frames)
	}
}

func TestImpactRemoved(t *testing.T) {
	got, _ := impact(t, func(m map[string]string) {
		replace(m, "capabilities/c/features/g.md", rule("c.g.a", "someone reads <x> of the [thing]", "it is read"), "")
	})
	if !slices.Contains(got["c.f.r"], "events someone reads c.g.a") || !slices.Contains(got["c.g.inv"], "ranges feature c.g.a") {
		t.Errorf("a removed rule: %v", got)
	}
}

func TestEventHead(t *testing.T) {
	cases := map[string]string{
		"lint reads the row <row> of an [Examples table], below the rows <earlier>":      "lint reads row",
		"lint reads the row <row> of an [Examples table] with the columns <columns>":     "lint reads row",
		"`diff` compares the [element] <element> of the [spec tree] with the [baseline]": "diff compares element",
		"lint reads the [hand-over] to <target>":                                         "lint reads hand over to",
		"lint exits, having found \"x\"":                                                 "lint exits having found x",
	}
	for when, want := range cases {
		e := lint.Element{Steps: []lint.Step{{Keyword: "When", Text: when}}}
		if got, ok := EventHead(&e); !ok || got != want {
			t.Errorf("%q: %q, want %q", when, got, want)
		}
	}
	if _, ok := EventHead(&lint.Element{}); ok {
		t.Error("an element without a When has an event head")
	}
}
