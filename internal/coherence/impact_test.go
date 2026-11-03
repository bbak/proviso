package coherence

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/baseline/baselinetest"
	"github.com/bbak/proviso/internal/lint"
)

// impact compares the edited tree with the baseline of baselinetest.ImpactTree, and
// returns each candidate's reasons as "source via from".
func impact(t *testing.T, edit func(map[string]string)) (map[string][]string, []Frame) {
	t.Helper()
	before := baselinetest.Files(baselinetest.ImpactTree)
	base, bad := baseline.Parse(baseline.Write(before))
	if len(bad) > 0 {
		t.Fatal(bad)
	}
	now := baselinetest.Files(baselinetest.EditTree(edit))
	entries := baseline.Compare(now, base)
	if err := baseline.Former(entries, base, before, "rev"); err != nil {
		t.Fatal(err)
	}
	entries = baseline.Through(entries, now, before)
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
		baselinetest.Replace(m, "capabilities/c/features/f.md", "Then      it is read", "Then      it is read twice")
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
	got, _ := impact(t, func(m map[string]string) { baselinetest.Replace(m, "glossary.md", "A thing.", "Another thing.") })
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
		{"a capability's Goal", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/capability.md", "## Goal\n\ng", "## Goal\n\nh")
		}, "c", all},
		{"a baselinetest.Feature's Goal", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/g.md", "## Goal\n\ng", "## Goal\n\nh")
		}, "c.g", []string{"c.g.a", "c.g.b", "c.g.c", "c.g.d", "c.g.inv"}},
		{"an invariant of the system", func(m map[string]string) {
			baselinetest.Replace(m, "invariants.md", "### Property\n\np", "### Property\n\nq")
		}, "system.inv", all},
		{"an invariant of the capability", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/invariants.md", "### Property\n\np", "### Property\n\nq")
		}, "c.inv", all[:7]},
		{"an invariant of a feature", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/f.md", "### Property\n\np", "### Property\n\nq")
		}, "c.f.inv", []string{"c.f.r"}},
	}
	for _, c := range cases {
		frames := scope(c.edit)
		if len(frames) != 1 || frames[0].ID != c.frame || !slices.Equal(frames[0].Candidates, c.want) {
			t.Errorf("%s: %v", c.name, frames)
		}
	}
	got, frames := impact(t, func(m map[string]string) {
		baselinetest.Replace(m, "capabilities/c/features/f.md", "- Contact: Ann", "- Contact: Bob")
	})
	if len(got) != 0 || len(frames) != 0 {
		t.Errorf("a new Contact: %v %v", got, frames)
	}
}

func TestImpactRemoved(t *testing.T) {
	got, _ := impact(t, func(m map[string]string) {
		baselinetest.Replace(m, "capabilities/c/features/g.md", baselinetest.Rule("c.g.a", "someone reads <x> of the [thing]", "it is read"), "")
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
