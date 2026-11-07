// Package query finds the elements of a spec tree that an identifier or a
// name stands for, as lint read them, with the element each belongs to:
// what `proviso show` writes.
package query

import (
	"cmp"
	"path"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/tree"
)

// Found is an element the query names, with its element parent: the
// identifier of the element it belongs to, "" for none.
type Found struct {
	lint.Element
	ElementParent string
}

// HasCapability reports whether the tree holds a folder named name in the
// folder capabilities of its spec root, with or without its capability.md
// (query.show.in). Hidden folders are no part of the tree.
func HasCapability(t *tree.Tree, name string) bool {
	return slices.ContainsFunc(t.CapFolders, func(c tree.CapFolder) bool {
		return c.Rel == path.Join("capabilities", name)
	})
}

// Find returns every element of elements that q names, each once, with its
// element parent, by the path of its file below the spec root, compared
// byte by byte, then by where it stands in that file (query.show.match).
// With in not "", it looks only among the elements of the capability in
// and of the system level (query.show.in).
func Find(elements []lint.Element, q, in string) []Found {
	parents := parentsOf(elements)
	var out []Found
	for i := range elements {
		e := &elements[i]
		if in != "" && e.Cap != in && e.Cap != "" {
			continue
		}
		if e.Named(q) {
			out = append(out, Found{*e, parents[i]})
		}
	}
	slices.SortStableFunc(out, func(a, b Found) int {
		return cmp.Or(strings.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line))
	})
	return out
}

// parentsOf returns the element parent of each element, by its index, as
// query.element-parent defines it: an element of a feature file but the
// feature belongs to the feature above it; a feature to its capability; a
// capability to the system; a member to the contract element it is a
// member of; any other element to the capability whose folder holds its
// file, or to the system for a file of the system level. The system
// belongs to none, and so does an element whose parent the tree does not
// hold, such as a feature of a capability whose capability.md is missing.
func parentsOf(elements []lint.Element) []string {
	system := ""
	caps := map[string]string{}
	for _, e := range elements {
		switch e.Type {
		case lint.TypeSystem:
			if system == "" {
				system = e.ID
			}
		case lint.TypeCapability:
			if _, ok := caps[e.Cap]; !ok {
				caps[e.Cap] = e.ID
			}
		}
	}
	level := func(cap string) string {
		if cap == "" {
			return system
		}
		return caps[cap]
	}
	out := make([]string, len(elements))
	feature := map[string]string{} // the last feature seen in each file
	for i, e := range elements {
		kind, _ := tree.Classify(strings.TrimPrefix(e.File, tree.RootName+"/"))
		switch {
		case e.Type == lint.TypeSystem:
		case e.Type == lint.TypeCapability:
			out[i] = system
		case e.Type == lint.TypeFeature:
			feature[e.File] = e.ID
			out[i] = level(e.Cap)
		case kind == tree.Feature && feature[e.File] != "":
			out[i] = feature[e.File]
		case e.Parent != "":
			out[i] = e.Parent
		default:
			out[i] = level(e.Cap)
		}
	}
	return out
}
