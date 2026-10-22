package baseline

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/bbak/proviso/internal/lint"
)

// The changes an element and a part are listed with.
const (
	Added          = "added"
	Changed        = "changed"
	Removed        = "removed"
	ChangedThrough = "changed-through"
)

// Entry is an element listed in the element diff.
type Entry struct {
	ID      string
	Change  string
	Element *lint.Element // as it stands now; nil for a removed element
	Former  string        // the former text of a removed element
	Parts   []PartChange  // for a changed element
	Through [][]string    // for an element changed through a dependency, every chain that reaches it
}

// PartChange is a part of a changed element that was added, changed or
// removed, with the former text of a changed or removed part.
type PartChange struct {
	Part, Change, Former string
}

// Undecided returns the undecided places of an element to be built, and the
// marks of open terms in it; a removed element has none.
func (e *Entry) Undecided() []lint.Pos {
	if e.Element == nil {
		return nil
	}
	return e.Element.Undecided
}

// Compare lists what the tree's elements did against the baseline, ordered
// by identifier, without former texts: Former fills them in.
func Compare(elements []lint.Element, base Baseline) []*Entry {
	var out []*Entry
	byID := map[string]*Entry{}
	present := map[string]bool{}
	for i := range elements {
		e := &elements[i]
		present[e.ID] = true
		recorded, ok := base[e.ID]
		if !ok {
			byID[e.ID] = &Entry{ID: e.ID, Change: Added, Element: e}
			continue
		}
		var parts []PartChange
		now := map[string]bool{}
		for _, p := range Parts(*e) {
			now[p.Key] = true
			switch fp, ok := recorded[p.Key]; {
			case !ok:
				parts = append(parts, PartChange{Part: p.Key, Change: Added})
			case fp != p.FP:
				parts = append(parts, PartChange{Part: p.Key, Change: Changed})
			}
		}
		for key := range recorded {
			if !now[key] {
				parts = append(parts, PartChange{Part: key, Change: Removed})
			}
		}
		if len(parts) > 0 {
			byID[e.ID] = &Entry{ID: e.ID, Change: Changed, Element: e, Parts: parts}
		}
	}
	for id := range base {
		if !present[id] {
			byID[id] = &Entry{ID: id, Change: Removed}
		}
	}
	throughDependencies(elements, byID)
	for _, id := range slices.Sorted(maps.Keys(byID)) {
		e := byID[id]
		slices.SortFunc(e.Parts, func(a, b PartChange) int { return cmp.Compare(a.Part, b.Part) })
		out = append(out, e)
	}
	return out
}

// NeedsFormer reports whether any entry needs former text: a removed
// element, or a changed or removed part.
func NeedsFormer(entries []*Entry) bool {
	for _, e := range entries {
		if e.Change == Removed || slices.ContainsFunc(e.Parts, func(p PartChange) bool { return p.Change != Added }) {
			return true
		}
	}
	return false
}

// Former fills in the former texts from the elements of the tree at the
// revision that wrote the manifest. Each must be there, with the
// fingerprint the manifest records, or the text found would not be the
// baseline's.
func Former(entries []*Entry, base Baseline, former []lint.Element, rev string) error {
	byID := map[string]*lint.Element{}
	for i := range former {
		byID[former[i].ID] = &former[i]
	}
	for _, e := range entries {
		if e.Change != Removed && !slices.ContainsFunc(e.Parts, func(p PartChange) bool { return p.Change != Added }) {
			continue
		}
		fe := byID[e.ID]
		if fe == nil {
			return fmt.Errorf("the element %s is not in the spec tree at %s, the revision that wrote manifest.sum", e.ID, rev)
		}
		parts := map[string]Part{}
		for _, p := range Parts(*fe) {
			parts[p.Key] = p
		}
		for key, fp := range base[e.ID] {
			if parts[key].FP != fp && (e.Change == Removed || slices.ContainsFunc(e.Parts, func(p PartChange) bool { return p.Part == key && p.Change != Added })) {
				return fmt.Errorf("the part %q of %s at %s, the revision that wrote manifest.sum, is not what manifest.sum records", key, e.ID, rev)
			}
		}
		if e.Change == Removed {
			e.Former = fe.Text
			continue
		}
		for i, p := range e.Parts {
			if p.Change != Added {
				e.Parts[i].Former = parts[p.Part].Text
			}
		}
	}
	return nil
}
