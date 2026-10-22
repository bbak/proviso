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

// systemInvariant is an invariant in invariants.md, where its heading has
// no prefix.
func systemInvariant(id string) string {
	return "## <a id=\"" + id + "\"></a> I\n\n" + invariantBody
}

// impactTree adds to the base tree: a second feature with a rule sharing
// the term and the event head of c.f.r and one referring to the parameter
// `p`, an invariant in each feature and one at system level, and a contract
// element representing the term.
func impactTree(m map[string]string) {
	m["capabilities/c/features/f.md"] = feature + "\n" + invariant("c.f.inv")
	m["capabilities/c/features/g.md"] = "# <a id=\"c.g\"></a> G\n\n- Contact: Ann\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n" +
		rule("c.g.a", "someone reads <x> of the [thing]", "it is read") +
		rule("c.g.b", "`p` is set to <x>", "it is set") + invariant("c.g.inv")
	m["invariants.md"] = "# Invariants\n\n" + systemInvariant("system.inv")
	m["parameters.md"] = "# Parameters\n\n## <a id=\"system.p\"></a> p\n\n- Value: 1\n- Configurable: no\n- Rationale: r\n"
	m["contracts/api.md"] = "# API\n\n## <a id=\"system.api\"></a> api\n\n- Represents: [thing]\n\nThe api.\n\n### <a id=\"system.api.code\"></a> code\n\nA code.\n"
}

// impact compares the tree, edited after impactTree, with the baseline of
// the tree impactTree gives, and returns each candidate's reasons as
// "source via from".
func impact(t *testing.T, edit func(map[string]string)) (map[string][]string, []Frame) {
	t.Helper()
	before := files(impactTree)
	base, bad := Parse(Write(before))
	if len(bad) > 0 {
		t.Fatal(bad)
	}
	now := files(func(m map[string]string) {
		impactTree(m)
		edit(m)
	})
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
		m["capabilities/c/features/f.md"] = strings.Replace(m["capabilities/c/features/f.md"], "Then      it is read", "Then      it is read twice", 1)
	})
	want := map[string][]string{
		"c.g.a":      {"terms system.thing c.f.r", "events someone reads c.f.r"},
		"c.f.inv":    {"ranges feature c.f.r"},
		"c.g.inv":    {"ranges capability c.f.r"},
		"system.inv": {"ranges system c.f.r"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) || len(frames) != 0 {
		t.Errorf("%v %v", got, frames)
	}
}

func TestImpactReferences(t *testing.T) {
	got, _ := impact(t, func(m map[string]string) {
		m["parameters.md"] = strings.Replace(m["parameters.md"], "- Value: 1", "- Value: 2", 1)
	})
	if fmt.Sprint(got) != fmt.Sprint(map[string][]string{"c.g.b": {"references p system.p"}}) {
		t.Errorf("a parameter: %v", got)
	}
	got, _ = impact(t, func(m map[string]string) {
		m["parameters.md"] = strings.Replace(m["parameters.md"], "- Rationale: r", "- Rationale: s", 1)
	})
	if len(got) != 0 {
		t.Errorf("a parameter's Rationale: %v", got)
	}
	// A member's top-level element represents the term.
	got, _ = impact(t, func(m map[string]string) {
		m["contracts/api.md"] = strings.Replace(m["contracts/api.md"], "A code.", "A new code.", 1)
	})
	want := map[string][]string{"c.f.r": {"references system.thing system.api.code"}, "c.g.a": {"references system.thing system.api.code"}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("a contract element: %v", got)
	}
}

func TestImpactListed(t *testing.T) {
	got, _ := impact(t, func(m map[string]string) {
		m["glossary.md"] = strings.Replace(m["glossary.md"], "A thing.", "Another thing.", 1)
	})
	// The rules marking the term are changed by it, so listed, not
	// candidates; every part of theirs counts as touched, so the
	// invariants above them are found.
	want := map[string][]string{
		"c.f.inv":    {"ranges capability c.g.a", "ranges feature c.f.r"},
		"c.g.inv":    {"ranges capability c.f.r", "ranges feature c.g.a"},
		"system.inv": {"ranges system c.f.r,c.g.a"},
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("rules changed by a term: %v", got)
	}
}

func TestImpactScope(t *testing.T) {
	got, frames := impact(t, func(m map[string]string) {
		m["capabilities/c/capability.md"] = strings.Replace(m["capabilities/c/capability.md"], "## Goal\n\ng", "## Goal\n\nh", 1)
	})
	ids := []string{"c.f.inv", "c.f.r", "c.g.a", "c.g.b", "c.g.inv"}
	if len(frames) != 1 || frames[0].ID != "c" || !slices.Equal(frames[0].Candidates, ids) || len(got) != 5 || got["c.f.r"][0] != "scope c c" {
		t.Errorf("%v %v", got, frames)
	}
	got, frames = impact(t, func(m map[string]string) {
		m["capabilities/c/features/f.md"] = strings.Replace(m["capabilities/c/features/f.md"], "- Contact: Ann", "- Contact: Bob", 1)
	})
	if len(got) != 0 || len(frames) != 0 {
		t.Errorf("a new Contact: %v %v", got, frames)
	}
	got, frames = impact(t, func(m map[string]string) {
		m["capabilities/c/features/g.md"] = strings.Replace(m["capabilities/c/features/g.md"], "## Goal\n\ng", "## Goal\n\nh", 1)
	})
	if len(frames) != 1 || !slices.Equal(frames[0].Candidates, []string{"c.g.a", "c.g.b", "c.g.inv"}) {
		t.Errorf("a feature's Goal: %v %v", got, frames)
	}
	_, frames = impact(t, func(m map[string]string) {
		m["invariants.md"] = strings.Replace(m["invariants.md"], "### Property\n\np", "### Property\n\nq", 1)
	})
	if len(frames) != 1 || len(frames[0].Candidates) != 5 { // every rule and invariant but itself
		t.Errorf("a system invariant: %v", frames)
	}
}

func TestImpactRemoved(t *testing.T) {
	got, _ := impact(t, func(m map[string]string) {
		m["capabilities/c/features/f.md"] = feature[:strings.Index(feature, "## <a id=\"c.f.r\">")] + invariant("c.f.inv")
	})
	if !slices.Contains(got["c.g.a"], "terms system.thing c.f.r") || !slices.Contains(got["c.f.inv"], "ranges feature c.f.r") {
		t.Errorf("a removed rule: %v", got)
	}
}

func TestEventHead(t *testing.T) {
	cases := map[string]string{
		"lint reads the row <row> of an [Examples table], below the rows <earlier>":      "lint reads row",
		"lint reads the row <row> of an [Examples table] with the columns <columns>":     "lint reads row",
		"`diff` compares the [element] <element> of the [spec tree] with the [baseline]": "diff compares element",
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
