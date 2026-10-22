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

// The sources of impact candidates, in the order impact_triggers names
// them.
const (
	Terms      = "terms"
	References = "references"
	Events     = "events"
	Ranges     = "ranges"
	Scope      = "scope"
)

var sourceOrder = []string{Terms, References, Events, Ranges, Scope}

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

// The element types impact_triggers names.
const (
	tSystem     = lint.TypeSystem
	tCapability = lint.TypeCapability
	tFeature    = lint.TypeFeature
	tRule       = lint.TypeRule
	tInvariant  = lint.TypeInvariant
	tConstraint = lint.TypeConstraint
	tGlossary   = lint.TypeGlossary
	tParameter  = lint.TypeParameter
	tContract   = lint.TypeContract
)

// opens returns the sources a touched part of an element of the type typ
// opens: the parameter impact_triggers.
func opens(typ, part string) []string {
	kind, name, _ := strings.Cut(part, " ")
	switch {
	case typ == tRule && (kind == "step" || kind == "row"):
		return []string{Terms, References, Events, Ranges}
	case typ == tGlossary && (part == "field Definition" || part == "field Defined by"):
		return []string{Terms}
	case typ == tContract, typ == tParameter && part != "field Rationale":
		return []string{References}
	case (typ == tInvariant || typ == tConstraint) && (part == "section Property" || part == "section Ranges over"):
		return []string{Scope}
	case typ == tSystem && kind == "section":
		return []string{Scope}
	case (typ == tCapability || typ == tFeature) && kind == "section" && name == "Goal":
		return []string{Scope}
	}
	return nil
}

// EventHead returns the event head of a rule: the text of its When step up
// to its first placeholder, with case, the articles, punctuation, term
// marks and backticks ignored. It returns false for an element without a
// When step.
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
		typ := versions[0].Type
		var parts []string
		if e.Change == Changed {
			for _, p := range e.Parts {
				parts = append(parts, p.Part)
			}
		} else {
			for _, v := range versions {
				for _, p := range Parts(*v) {
					parts = append(parts, p.Key)
				}
			}
		}
		open := map[string]bool{}
		for _, p := range parts {
			for _, s := range opens(typ, p) {
				open[s] = true
			}
		}
		if open[Terms] {
			terms := map[string]bool{}
			for _, v := range versions {
				for _, m := range v.Marks {
					terms[m] = true
				}
				if v.Type == tGlossary {
					terms[v.ID] = true
				}
			}
			for i := range now {
				r := &now[i]
				for _, m := range r.Marks {
					if r.Type == tRule && terms[m] {
						add(r, Terms, m, e.ID)
					}
				}
			}
		}
		if open[References] {
			referring(e.ID, versions, now, nowByID, formerByID, add)
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
					inv := &now[i]
					if inv.Type != tInvariant {
						continue
					}
					switch {
					case inv.File == v.File:
						add(inv, Ranges, "feature", e.ID)
					case inv.Cap != "" && inv.Cap == v.Cap:
						add(inv, Ranges, "capability", e.ID)
					case inv.Cap == "":
						add(inv, Ranges, "system", e.ID)
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

// framed reports whether the element c stands at or below the level of the
// frame f: the whole tree for the system and a system constraint or
// invariant, the capability for a capability and a capability's constraint
// or invariant, wherever it stands in it, and the feature's file for a
// feature.
func framed(f, c *lint.Element) bool {
	switch {
	case f.Type == tFeature:
		return c.File == f.File
	case f.Type == tSystem || f.Cap == "":
		return true
	default:
		return c.Cap == f.Cap
	}
}

// referring finds the rules that refer to the touched element, or to a
// contract element or parameter it refers to, and the rules that mark a
// term one of those contract elements represents, or that refer to a
// contract element representing a term the touched element defines.
func referring(from string, versions []*lint.Element, now []lint.Element, nowByID, formerByID map[string]*lint.Element, add func(*lint.Element, string, string, string) bool) {
	targets := map[string]bool{from: true}
	defines := map[string]bool{}
	for _, v := range versions {
		for _, r := range v.Refs {
			targets[r.ID] = true
		}
		if v.Type == tGlossary {
			defines[v.ID] = true
		}
	}
	lookup := func(id string) *lint.Element {
		if e := nowByID[id]; e != nil {
			return e
		}
		return formerByID[id]
	}
	// represented returns the terms a contract element represents, itself
	// or through its top-level element.
	represented := func(id string) []string {
		e := lookup(id)
		if e == nil {
			return nil
		}
		terms := slices.Clone(e.Represents)
		if top := lookup(e.Top); top != nil && top != e {
			terms = append(terms, top.Represents...)
		}
		return terms
	}
	terms := map[string]bool{}
	for id := range targets {
		for _, t := range represented(id) {
			terms[t] = true
		}
	}
	for i := range now {
		r := &now[i]
		if r.Type != tRule {
			continue
		}
		for _, ref := range r.Refs {
			if targets[ref.ID] {
				add(r, References, ref.Name, from)
			}
			if slices.ContainsFunc(represented(ref.ID), func(t string) bool { return defines[t] }) {
				add(r, References, ref.Name, from)
			}
		}
		for _, m := range r.Marks {
			if terms[m] {
				add(r, References, m, from)
			}
		}
	}
}
