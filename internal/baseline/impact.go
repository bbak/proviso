package baseline

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/bbak/proviso/internal/guarded"
	"github.com/bbak/proviso/internal/lint"
)

// Reason is why an element is a candidate: the source it was found
// through, what it shares with the touched elements, and the touched
// elements the reason comes from.
type Reason struct {
	Source, Via string
	From        []string
}

// Candidate is a rule or invariant the element diff does not list but the
// amendment may affect.
type Candidate struct {
	Element *lint.Element
	Reasons []Reason
}

// Frame is a touched element that opened "scope", with the candidates
// found through it.
type Frame struct {
	ID         string
	Candidates []string
}

// EventHead returns the event head of a rule: the words of its When step
// up to its first placeholder, or of the whole step if it has none, every
// character but a letter, a mark or a digit separating them, with case
// ignored and the articles left out. It returns false for an element
// without a When step.
func EventHead(e *lint.Element) (string, bool) {
	i := slices.IndexFunc(e.Steps, func(s lint.Step) bool { return s.Keyword == guarded.When })
	if i < 0 {
		return "", false
	}
	var sb strings.Builder
	var src []guarded.SourceLine
	for n, line := range strings.Split(e.Steps[i].Text, "\n") {
		src = append(src, guarded.SourceLine{N: n + 1, Text: line})
	}
done:
	for _, ln := range guarded.Parse(src).Lines {
		for _, sp := range ln.Spans {
			if sp.Kind == guarded.Placeholder {
				break done
			}
			sb.WriteString(" " + sp.Value + " ")
		}
	}
	words := strings.Fields(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, sb.String()))
	words = slices.DeleteFunc(words, func(w string) bool { return w == "the" || w == "a" || w == "an" })
	return strings.Join(words, " "), true
}

// inFeature reports whether an element stands in a feature's file.
func inFeature(e *lint.Element) bool { return strings.Contains(e.File, "/features/") }

// Impact finds the candidates of an amendment from the touched elements
// of its element diff: the elements now, and those of the baseline, or nil
// where no former text was read, as then no touched element held a part
// in the baseline that it does not hold now.
func Impact(entries []*Entry, now, former []lint.Element) ([]Candidate, []Frame) {
	listed := map[string]bool{}
	for _, e := range entries {
		listed[e.ID] = true
	}
	nowByID := byID(now)
	formerByID := byID(former)
	found := map[string]map[[2]string]map[string]bool{} // candidate, source and via, from
	add := func(c *lint.Element, source, via, from string) bool {
		if listed[c.ID] {
			return false
		}
		if found[c.ID] == nil {
			found[c.ID] = map[[2]string]map[string]bool{}
		}
		k := [2]string{source, via}
		if found[c.ID][k] == nil {
			found[c.ID][k] = map[string]bool{}
		}
		found[c.ID][k][from] = true
		return true
	}
	var frames []Frame
	for _, e := range entries {
		cur := nowByID[e.ID]
		old := formerByID[e.ID]
		if former == nil && e.Change != Added {
			old = cur
		}
		var versions []*lint.Element
		for _, v := range []*lint.Element{cur, old} {
			if v != nil && !slices.Contains(versions, v) {
				versions = append(versions, v)
			}
		}
		if len(versions) == 0 {
			continue
		}
		open := map[string]bool{}
		for _, v := range versions {
			for _, p := range changedParts(e, v) {
				_, sources := effects(v.Type, p)
				for _, s := range sources {
					open[s] = true
				}
			}
		}
		if open[Events] {
			heads := map[string]bool{}
			for _, v := range versions {
				if h, ok := EventHead(v); ok {
					heads[h] = true
				}
			}
			for i := range now {
				r := &now[i]
				if h, ok := EventHead(r); ok && r.Type == tRule && heads[h] {
					add(r, Events, h, e.ID)
				}
			}
		}
		if open[Ranges] {
			for _, v := range versions {
				for i := range now {
					if inv := &now[i]; inv.Type == tInvariant {
						if level := ranges(v, inv); level != "" {
							add(inv, Ranges, level, e.ID)
						}
					}
				}
			}
		}
		if open[Scope] {
			f := Frame{ID: e.ID, Candidates: []string{}}
			for _, v := range versions {
				for i := range now {
					c := &now[i]
					if (c.Type == tRule || c.Type == tInvariant) && framed(v, c) && add(c, Scope, e.ID, e.ID) && !slices.Contains(f.Candidates, c.ID) {
						f.Candidates = append(f.Candidates, c.ID)
					}
				}
			}
			slices.Sort(f.Candidates)
			frames = append(frames, f)
		}
	}
	var out []Candidate
	for _, id := range slices.Sorted(maps.Keys(found)) {
		c := Candidate{Element: nowByID[id]}
		for k, from := range found[id] {
			c.Reasons = append(c.Reasons, Reason{Source: k[0], Via: k[1], From: slices.Sorted(maps.Keys(from))})
		}
		slices.SortFunc(c.Reasons, func(a, b Reason) int {
			return cmp.Or(cmp.Compare(slices.Index(sourceOrder, a.Source), slices.Index(sourceOrder, b.Source)), cmp.Compare(a.Via, b.Via))
		})
		out = append(out, c)
	}
	slices.SortFunc(frames, func(a, b Frame) int { return cmp.Compare(a.ID, b.ID) })
	return out, frames
}

func byID(els []lint.Element) map[string]*lint.Element {
	m := map[string]*lint.Element{}
	for i := range els {
		m[els[i].ID] = &els[i]
	}
	return m
}

// ranges returns the level of the invariant inv above the rule r: its
// feature, for an invariant in the rule's feature file; its capability, for
// one in the capability's invariants.md; the system, for one in the
// system's; and "" for any other.
func ranges(r, inv *lint.Element) string {
	switch {
	case inv.File == r.File:
		return "feature"
	case inFeature(inv):
		return ""
	case inv.Cap != "" && inv.Cap == r.Cap:
		return "capability"
	case inv.Cap == "":
		return "system"
	}
	return ""
}

// framed reports whether the element c stands at or below the level the
// frame f stands at: its feature's file for a feature and an invariant in
// a feature's file, the whole tree for the system and a system constraint
// or invariant, and the capability for a capability and a capability's
// constraint or invariant.
func framed(f, c *lint.Element) bool {
	switch {
	case f.Type == tFeature || inFeature(f):
		return c.File == f.File
	case f.Type == tSystem || f.Cap == "":
		return true
	default:
		return c.Cap == f.Cap
	}
}
