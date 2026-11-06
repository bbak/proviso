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

// Effects is the parameter change_effects: whether a changed part of an
// element of the type typ passes its changed meaning on to its dependents,
// and which candidate sources it opens.
func Effects(typ, part string) (dependents bool, sources []string) {
	kind, name, _ := strings.Cut(part, " ")
	switch {
	case typ == tRule && (kind == "step" || kind == "row"):
		return true, []string{Events, Ranges}
	case typ == tGlossary && (part == "field Definition" || part == "field Defined by" || part == "field Deviations"):
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

// ChangedParts returns the parts of a listed element that count as
// changed: every part of an added or removed element, or of one changed
// through a dependency, and the parts named for a changed one.
func ChangedParts(e *Entry, version *lint.Element) []string {
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
// each element depends on, read from the tree as it stands. removed holds,
// by the contract element that held them, the members an amendment
// removed, on which a rule referring to that contract element still
// depends.
func dependsOn(elements []lint.Element, removed map[string][]string) map[string][]string {
	members := map[string][]string{}
	for holder, ids := range removed {
		members[holder] = slices.Clone(ids)
	}
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
			// An entry that marks its own term depends on nothing through
			// that mark.
			for _, m := range e.LeansOn {
				if m != e.ID {
					d = append(d, m)
				}
			}
		}
		slices.Sort(d)
		deps[e.ID] = slices.Compact(d)
	}
	return deps
}

// throughDependencies lists every element that depends on a changed
// meaning, and is not listed otherwise, as changed through a dependency,
// naming, from each listed element whose changed part passes its meaning
// on and from which a chain of identifiers reaches it, the shortest such
// chain, and of equally short ones the first. former is the tree the
// removed elements stood in, which tells the members among them.
func throughDependencies(elements, former []lint.Element, byID map[string]*Entry) {
	formerElem := map[string]*lint.Element{}
	for i := range former {
		formerElem[former[i].ID] = &former[i]
	}
	removed := map[string][]string{}
	for id, e := range byID {
		if fe := formerElem[id]; e.Change == Removed && fe != nil && fe.Parent != "" {
			removed[fe.Parent] = append(removed[fe.Parent], id)
		}
	}
	deps := dependsOn(elements, removed)
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
		version := e.Element
		if version == nil {
			// Only a removed member is still depended on, by the rules
			// referring to the contract element that held it.
			version = formerElem[id]
		}
		if version == nil {
			continue
		}
		for _, p := range ChangedParts(e, version) {
			if passes, _ := Effects(version.Type, p); passes {
				seeds[id] = true
				queue = append(queue, id)
				break
			}
		}
	}
	slices.Sort(queue)
	seedList := slices.Clone(queue)
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
	// From each seed, a breadth-first walk along the dependents finds the
	// shortest chain to every element it reaches, and of equally short
	// ones the first, compared identifier by identifier: every chain to an
	// element one step further is a chain to the one before it with that
	// one appended, so the first to it extends the first to the one before.
	for _, seed := range seedList {
		best := map[string][]string{seed: {seed}}
		layer := []string{seed}
		for len(layer) > 0 {
			next := map[string][]string{}
			for _, id := range layer {
				for _, x := range dependents[id] {
					if _, ok := best[x]; ok || elem[x] == nil {
						continue
					}
					if c, ok := next[x]; !ok || slices.Compare(best[id], c[:len(c)-1]) < 0 {
						next[x] = append(slices.Clone(best[id]), x)
					}
				}
			}
			layer = slices.Sorted(maps.Keys(next))
			maps.Copy(best, next)
		}
		for x, c := range best {
			if through[x] {
				byID[x].Through = append(byID[x].Through, c[:len(c)-1])
			}
		}
	}
	for id := range through {
		slices.SortFunc(byID[id].Through, slices.Compare[[]string])
	}
}
