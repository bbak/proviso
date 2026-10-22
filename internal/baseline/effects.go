package baseline

import (
	"maps"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/lint"
)

// The candidate sources, in the order they are written.
const (
	Events = "events"
	Ranges = "ranges"
	Scope  = "scope"
)

var sourceOrder = []string{Events, Ranges, Scope}

// The element types change_effects and dependencies name.
const (
	tSystem     = lint.TypeSystem
	tCapability = lint.TypeCapability
	tFeature    = lint.TypeFeature
	tRule       = lint.TypeRule
	tInvariant  = lint.TypeInvariant
	tConstraint = lint.TypeConstraint
	tGlossary   = lint.TypeGlossary
	tParameter  = lint.TypeParameter
	tReference  = lint.TypeReference
	tContract   = lint.TypeContract
)

// effects is the parameter change_effects: whether a changed part of an
// element of the type typ passes its changed meaning on to its dependents,
// and which candidate sources it opens.
func effects(typ, part string) (dependents bool, sources []string) {
	kind, name, _ := strings.Cut(part, " ")
	switch {
	case typ == tRule && (kind == "step" || kind == "row"):
		return true, []string{Events, Ranges}
	case typ == tGlossary && (part == "field Definition" || part == "field Defined by"):
		return true, nil
	case typ == tReference && (part == "field Version" || part == "field Link"):
		return true, nil
	case typ == tContract, typ == tParameter && part != "field Rationale":
		return true, nil
	case (typ == tInvariant || typ == tConstraint) && (part == "section Property" || part == "section Ranges over"):
		return false, []string{Scope}
	case typ == tSystem && kind == "section":
		return false, []string{Scope}
	case (typ == tCapability || typ == tFeature) && kind == "section" && name == "Goal":
		return false, []string{Scope}
	}
	return false, nil
}

// changedParts returns the parts of a listed element that count as
// changed: every part of an added or removed element, or of one changed
// through a dependency, and the parts named for a changed one.
func changedParts(e *Entry, version *lint.Element) []string {
	var out []string
	if e.Change == Changed {
		for _, p := range e.Parts {
			out = append(out, p.Part)
		}
		return out
	}
	for _, p := range Parts(*version) {
		out = append(out, p.Key)
	}
	return out
}

// dependsOn is the parameter dependencies: the identifiers of the elements
// each element depends on, read from the tree as it stands.
func dependsOn(elements []lint.Element) map[string][]string {
	members := map[string][]string{}
	for _, e := range elements {
		if e.Parent != "" {
			members[e.Parent] = append(members[e.Parent], e.ID)
		}
	}
	var withMembers func(id string, out []string) []string
	withMembers = func(id string, out []string) []string {
		if slices.Contains(out, id) {
			return out
		}
		out = append(out, id)
		for _, m := range members[id] {
			out = withMembers(m, out)
		}
		return out
	}
	deps := map[string][]string{}
	for _, e := range elements {
		var d []string
		switch e.Type {
		case tRule:
			d = append(d, e.Marks...)
			for _, r := range e.Refs {
				d = withMembers(r.ID, d)
			}
			if e.HandOver != "" {
				d = append(d, e.HandOver)
			}
		case tContract:
			d = append(d, e.Represents...)
		case tGlossary:
			if e.DefinedBy != "" {
				d = append(d, e.DefinedBy)
			}
		}
		slices.Sort(d)
		deps[e.ID] = slices.Compact(d)
	}
	return deps
}

// throughDependencies lists every element that depends on a changed
// meaning, and is not listed otherwise, as changed through a dependency,
// with every chain of identifiers that reaches it.
func throughDependencies(elements []lint.Element, byID map[string]*Entry) {
	deps := dependsOn(elements)
	dependents := map[string][]string{}
	for id, ds := range deps {
		for _, d := range ds {
			dependents[d] = append(dependents[d], id)
		}
	}
	elem := map[string]*lint.Element{}
	for i := range elements {
		elem[elements[i].ID] = &elements[i]
	}
	seeds := map[string]bool{}
	var queue []string
	for id, e := range byID {
		if e.Element == nil {
			continue // nothing in the tree depends on a removed element
		}
		for _, p := range changedParts(e, e.Element) {
			if passes, _ := effects(e.Element.Type, p); passes {
				seeds[id] = true
				queue = append(queue, id)
				break
			}
		}
	}
	slices.Sort(queue)
	// changed holds every element whose meaning changed: one whose changed
	// part passes it on, and every element depending on one, transitively,
	// listed or not. Each is taken once.
	changed := maps.Clone(seeds)
	through := map[string]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		ds := slices.Clone(dependents[id])
		slices.Sort(ds)
		for _, x := range ds {
			if changed[x] || elem[x] == nil {
				continue
			}
			changed[x] = true
			queue = append(queue, x)
			if byID[x] == nil {
				through[x] = true
				byID[x] = &Entry{ID: x, Change: ChangedThrough, Element: elem[x]}
			}
		}
	}
	// chains returns every chain of identifiers from a first changed
	// element to one the element id depends on directly.
	memo := map[string][][]string{}
	var chains func(id string, visiting map[string]bool) [][]string
	chains = func(id string, visiting map[string]bool) [][]string {
		if c, ok := memo[id]; ok {
			return c
		}
		visiting[id] = true
		var out [][]string
		for _, d := range deps[id] {
			if !changed[d] || visiting[d] {
				continue
			}
			if seeds[d] {
				out = append(out, []string{d})
			}
			for _, c := range chains(d, visiting) {
				out = append(out, append(slices.Clone(c), d))
			}
		}
		delete(visiting, id)
		slices.SortFunc(out, slices.Compare[[]string])
		memo[id] = out
		return out
	}
	for id := range through {
		byID[id].Through = chains(id, map[string]bool{})
	}
}
