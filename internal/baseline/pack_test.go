package baseline

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/baseline/baselinetest"
	"github.com/bbak/proviso/internal/lint"
)

// packOf compares the edited tree with the baseline of baselinetest.ImpactTree, and
// returns the identifiers of the pack by heading, "amendment" for the
// element diff.
func packOf(t *testing.T, edit func(map[string]string)) map[string][]string {
	t.Helper()
	before := baselinetest.Files(baselinetest.ImpactTree)
	base, _ := Parse(Write(before))
	now := baselinetest.Files(baselinetest.EditTree(edit))
	entries := Through(Compare(now, base), now, before)
	out := map[string][]string{}
	for _, e := range PackOrder(entries) {
		out["amendment"] = append(out["amendment"], e.ID)
	}
	for _, p := range Pack(entries, now) {
		out[p.Heading] = append(out[p.Heading], p.Element.ID)
	}
	return out
}

func TestPack(t *testing.T) {
	cases := []struct {
		name string
		edit func(map[string]string)
		want map[string][]string
	}{
		{"none", func(map[string]string) {}, map[string][]string{
			HeadingSystem: {"system"},
		}},
		{"a rule referring to a contract element", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/g.md", "it is answered", "it is answered at once")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.c"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api", "system.api.code", "system.thing"},
		}},
		{"a member's holder and the term it represents", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/g.md", "`api` answers", "`api.code` answers")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.c"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api", "system.api.code", "system.thing"},
		}},
		{"every term a contract element represents", func(m map[string]string) {
			baselinetest.Replace(m, "contracts/api.md", "- Represents: [thing]", "- Represents: [thing] by a [std]")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.c", "system.api"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api.code", "system.thing", "system.std", "system.ref.std"},
		}},
		{"a hand-over followed to its end, not a rule sharing a term", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/g.md", "Then      it is written", "Then      it is written down")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.d"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"c.f.r", "system.thing"},
		}},
		{"an unchanged standard", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/g.md", "of the [thing]", "of the [std]")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.a"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.std", "system.ref.std"},
		}},
		{"a removed rule brings nothing", func(m map[string]string) {
			m["capabilities/c/features/g.md"] = strings.Replace(m["capabilities/c/features/g.md"], baselinetest.Rule("c.g.c", "`api` answers <x>", "it is answered"), "", 1)
		}, map[string][]string{
			HeadingSystem: {"system"},
			"amendment":   {"c.g.c"},
		}},
		{"a glossary entry of the system", func(m map[string]string) {
			baselinetest.Replace(m, "glossary.md", "- Contact: Ann\n- Status: defined\n- Aliases: things", "- Contact: Bob\n- Status: defined\n- Aliases: things")
		}, map[string][]string{
			HeadingSystem: {"system"},
			"amendment":   {"system.thing"},
			HeadingFrames: {"system.inv"},
		}},
		{"listed elements appear under Amendment only, others once", func(m map[string]string) {
			baselinetest.Replace(m, "capabilities/c/features/g.md", "## Goal\n\ng\n", "## Goal\n\nh\n")
			baselinetest.Replace(m, "capabilities/c/features/g.md", "it is answered", "it is answered at once")
			baselinetest.Replace(m, "capabilities/c/features/g.md", "it is read", "it is read at once")
			baselinetest.Replace(m, "system.md", "## Scope\n\ns\n", "## Scope\n\nt\n")
		}, map[string][]string{
			"amendment":    {"c.g", "c.g.a", "c.g.c", "system"},
			HeadingFrames:  {"c", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api", "system.api.code", "system.thing"},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := packOf(t, c.edit)
			for _, k := range slices.Sorted(maps.Keys(got)) {
				if !slices.Equal(got[k], c.want[k]) {
					t.Errorf("%s: got %v, want %v", k, got[k], c.want[k])
				}
			}
			for k := range c.want {
				if _, ok := got[k]; !ok {
					t.Errorf("%s: got nothing, want %v", k, c.want[k])
				}
			}
		})
	}
}

func TestPackElement(t *testing.T) {
	elements := baselinetest.Files(baselinetest.ImpactTree)
	cases := []struct {
		id   string
		want map[string][]string
	}{
		{"c.g.c", map[string][]string{
			HeadingSystem:  {"system"},
			HeadingElement: {"c.g.c"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api", "system.api.code", "system.thing"},
		}},
		{"system", map[string][]string{
			HeadingElement: {"system"},
			HeadingFrames:  {"system.inv"},
		}},
	}
	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			packed, ok := PackElement(c.id, elements)
			if !ok {
				t.Fatal("not found")
			}
			got := map[string][]string{}
			var order []string
			for _, p := range packed {
				got[p.Heading] = append(got[p.Heading], p.Element.ID)
				if len(order) == 0 || order[len(order)-1] != p.Heading {
					order = append(order, p.Heading)
				}
			}
			if !maps.EqualFunc(got, c.want, slices.Equal) {
				t.Errorf("got %v, want %v", got, c.want)
			}
			if !slices.IsSortedFunc(order, func(a, b string) int { return rank(a) - rank(b) }) {
				t.Errorf("headings in the order %v", order)
			}
		})
	}
	if _, ok := PackElement("c.g.gone", elements); ok {
		t.Error("an identifier the tree does not hold was found")
	}
}

func TestPackOrder(t *testing.T) {
	el := func(file string, line int) *lint.Element { return &lint.Element{File: file, Line: line} }
	entries := []*Entry{
		{ID: "a.z", Change: Removed},
		{ID: "lint.run.exit", Change: Changed, Element: el("proviso/capabilities/lint/features/run.md", 30)},
		{ID: "a.b", Change: Removed},
		{ID: "lint.run.root", Change: Changed, Element: el("proviso/capabilities/lint/features/run.md", 10)},
		{ID: "system.git", Change: Added, Element: el("proviso/constraints.md", 1)},
		{ID: "lint", Change: Added, Element: el("proviso/capabilities/lint/capability.md", 1)},
	}
	var got []string
	for _, e := range PackOrder(entries) {
		got = append(got, e.ID)
	}
	want := []string{"lint", "lint.run.root", "lint.run.exit", "system.git", "a.b", "a.z"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
