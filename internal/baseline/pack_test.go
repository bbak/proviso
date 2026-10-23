package baseline

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

// packOf compares the edited tree with the baseline of impactTree, and
// returns the identifiers of the pack by heading, "amendment" for the
// element diff.
func packOf(t *testing.T, edit func(map[string]string)) map[string][]string {
	t.Helper()
	base, _ := Parse(Write(files(impactTree)))
	now := files(editTree(edit))
	entries := Compare(now, base)
	out := map[string][]string{}
	for _, e := range entries {
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
			replace(m, "capabilities/c/features/g.md", "it is answered", "it is answered at once")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.c"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api", "system.api.code", "system.thing"},
		}},
		{"a member's holder and the term it represents", func(m map[string]string) {
			replace(m, "capabilities/c/features/g.md", "`api` answers", "`api.code` answers")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.c"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.api", "system.api.code", "system.thing"},
		}},
		{"a hand-over followed to its end, not a rule sharing a term", func(m map[string]string) {
			replace(m, "capabilities/c/features/g.md", "Then      it is written", "Then      it is written down")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.d"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"c.f.r", "system.thing"},
		}},
		{"an unchanged standard", func(m map[string]string) {
			replace(m, "capabilities/c/features/g.md", "of the [thing]", "of the [std]")
		}, map[string][]string{
			HeadingSystem:  {"system"},
			"amendment":    {"c.g.a"},
			HeadingFrames:  {"c", "c.g", "c.g.inv", "c.inv", "system.inv"},
			HeadingContext: {"system.ref.std", "system.std"},
		}},
		{"a removed rule brings nothing", func(m map[string]string) {
			m["capabilities/c/features/g.md"] = strings.Replace(m["capabilities/c/features/g.md"], rule("c.g.c", "`api` answers <x>", "it is answered"), "", 1)
		}, map[string][]string{
			HeadingSystem: {"system"},
			"amendment":   {"c.g.c"},
		}},
		{"a glossary entry of the system", func(m map[string]string) {
			replace(m, "glossary.md", "- Contact: Ann\n- Status: defined\n- Aliases: things", "- Contact: Bob\n- Status: defined\n- Aliases: things")
		}, map[string][]string{
			HeadingSystem: {"system"},
			"amendment":   {"system.thing"},
			HeadingFrames: {"system.inv"},
		}},
		{"listed elements appear under Amendment only, others once", func(m map[string]string) {
			replace(m, "capabilities/c/features/g.md", "## Goal\n\ng\n", "## Goal\n\nh\n")
			replace(m, "capabilities/c/features/g.md", "it is answered", "it is answered at once")
			replace(m, "capabilities/c/features/g.md", "it is read", "it is read at once")
			replace(m, "system.md", "## Scope\n\ns\n", "## Scope\n\nt\n")
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
