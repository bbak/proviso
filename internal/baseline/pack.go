package baseline

import (
	"cmp"
	"slices"

	"github.com/bbak/proviso/internal/lint"
)

// The headings of a context pack other than Amendment, as context_pack
// names them, in the order pack_text gives them.
const (
	HeadingSystem  = "system"
	HeadingFrames  = "frames"
	HeadingContext = "context"
)

// Packed is an element of a context pack outside Amendment, with the
// heading it appears under.
type Packed struct {
	Heading string
	Element *lint.Element
}

// Pack assembles the context pack of an element diff from the elements of
// the tree as it stands: the elements the diff lists, under Amendment, and
// every other element of the pack, under its heading. Each element appears
// once: a listed one under Amendment only, any other under the heading it
// was taken for. The elements under each heading are ordered by
// identifier.
func Pack(entries []*Entry, elements []lint.Element) []Packed {
	byID := byID(elements)
	listed := map[string]bool{}
	for _, e := range entries {
		listed[e.ID] = true
	}
	taken := map[string]string{} // identifier, the first heading it was taken for
	take := func(heading string, e *lint.Element) {
		if e == nil || listed[e.ID] {
			return
		}
		if _, ok := taken[e.ID]; !ok || rank(heading) < rank(taken[e.ID]) {
			taken[e.ID] = heading
		}
	}

	// baseline.pack.contents
	for i := range elements {
		if elements[i].Type == tSystem {
			take(HeadingSystem, &elements[i])
		}
	}

	// baseline.pack.frames: the frames of every listed element that is not
	// removed.
	for _, e := range entries {
		if e.Element == nil {
			continue
		}
		for i := range elements {
			if frames(e.Element, &elements[i]) {
				take(HeadingFrames, &elements[i])
			}
		}
	}

	// baseline.pack.chain: from every listed element that is not removed,
	// and from every element taken so, each once.
	deps := dependsOn(elements)
	seen := map[string]bool{}
	var queue []string
	for _, e := range entries {
		if e.Element != nil {
			seen[e.ID] = true
			queue = append(queue, e.ID)
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		next := slices.Clone(deps[id])
		if e := byID[id]; e != nil && e.Parent != "" {
			next = append(next, e.Parent)
		}
		for _, d := range next {
			if byID[d] == nil || seen[d] {
				continue
			}
			seen[d] = true
			queue = append(queue, d)
			take(HeadingContext, byID[d])
		}
	}

	var out []Packed
	for id, h := range taken {
		out = append(out, Packed{Heading: h, Element: byID[id]})
	}
	slices.SortFunc(out, func(a, b Packed) int {
		return cmp.Or(cmp.Compare(rank(a.Heading), rank(b.Heading)), cmp.Compare(a.Element.ID, b.Element.ID))
	})
	return out
}

// rank is the place of a heading in pack_text, Amendment standing between
// System and Frames.
func rank(heading string) int {
	return slices.Index([]string{HeadingSystem, HeadingFrames, HeadingContext}, heading)
}

// frames reports whether the element f frames the listed element e: the
// feature and the capability that hold e or are e, every invariant of that
// feature, every invariant and constraint of that capability, and every
// invariant and constraint of the system.
func frames(e, f *lint.Element) bool {
	switch f.Type {
	case tFeature:
		return f.File == e.File
	case tCapability:
		return f.Cap != "" && f.Cap == e.Cap
	case tInvariant, tConstraint:
		switch {
		case inFeature(f):
			return f.File == e.File
		case f.Cap == "":
			return true
		default:
			return f.Cap == e.Cap
		}
	}
	return false
}
