// Package rehome plans the renames that give every element of a spec tree
// an identifier naming its identifier home, finds the renames that would
// take an identifier, and rewrites the places rehome_references lists.
package rehome

import (
	"cmp"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/guarded"
	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/markdown"
	"github.com/bbak/proviso/internal/tree"
)

// System is the identifier home of an element in a file of the system
// level.
const System = "system"

// Rename is a planned rename of an element's identifier.
type Rename struct {
	From, To string
}

// Collision is an identifier that renames would take: one an element of
// the tree has, or one planned for two elements. From holds the old
// identifiers of the renames planned to take it, compared byte by byte.
type Collision struct {
	Identifier string
	From       []string
}

// Plan returns the renames of upkeep.rehome.home, ordered by their old
// identifier, and the collisions of upkeep.rehome.collision, ordered by the
// identifier each would take, both compared byte by byte.
func Plan(elements []lint.Element) ([]Rename, []Collision) {
	// The identifier home of a capability's folder is the capability's
	// identifier, not the folder's name.
	homes := map[string]string{}
	ids := map[string]bool{}
	for _, e := range elements {
		ids[e.ID] = true
		if e.Type == lint.TypeCapability {
			homes[e.Cap] = e.ID
		}
	}
	var renames []Rename
	for _, e := range elements {
		home := System
		if e.Cap != "" {
			home = homes[e.Cap]
		}
		first, rest, hasRest := strings.Cut(e.ID, ".")
		if home == "" || first == home {
			continue
		}
		to := home
		if hasRest {
			to += "." + rest
		}
		renames = append(renames, Rename{e.ID, to})
	}
	slices.SortFunc(renames, func(a, b Rename) int { return cmp.Compare(a.From, b.From) })
	// An identifier is taken until the run ends, so no two elements swap.
	byTo := map[string][]string{}
	for _, r := range renames {
		byTo[r.To] = append(byTo[r.To], r.From)
	}
	var collisions []Collision
	for _, to := range slices.Sorted(maps.Keys(byTo)) {
		if ids[to] || len(byTo[to]) > 1 {
			collisions = append(collisions, Collision{to, byTo[to]})
		}
	}
	return renames, collisions
}

// edit is a place of a file that is to hold a new identifier.
type edit struct {
	start, end int
	to         string
}

// Rewrite returns, by their paths relative to the spec root, the files of
// t that hold a place rehome_references lists for an old identifier of
// renames, each with every such place holding the new identifier and
// nothing else of it changed. elements are the elements lint read in t.
func Rewrite(t *tree.Tree, elements []lint.Element, renames []Rename) map[string][]byte {
	to := map[string]string{}
	for _, r := range renames {
		to[r.From] = r.To
	}
	out := map[string][]byte{}
	if len(to) == 0 {
		return out
	}
	byFile := map[string][]lint.Element{}
	for _, e := range elements {
		byFile[e.File] = append(byFile[e.File], e)
	}
	for _, f := range t.Files {
		if f.Kind == tree.Settings || f.Kind == tree.Manifest {
			continue
		}
		edits := places(f, byFile[path.Join(tree.RootName, f.Rel)], to)
		if len(edits) == 0 {
			continue
		}
		slices.SortFunc(edits, func(a, b edit) int { return cmp.Compare(b.start, a.start) })
		data := slices.Clone(f.Data)
		for _, e := range edits {
			data = slices.Concat(data[:e.start], []byte(e.to), data[e.end:])
		}
		out[f.Rel] = data
	}
	return out
}

// places returns the places of the file f, whose elements are elements,
// that rehome_references lists and that hold an old identifier of to: the
// anchors, the hand-overs, the fragments of inline links and autolinks
// outside the Examples tables and of link reference definitions, and in a
// questions.md the tags that start its list items.
func places(f *tree.File, elements []lint.Element, to map[string]string) []edit {
	examples := map[int]bool{}
	for _, e := range elements {
		for _, r := range e.Rows {
			examples[r.Line] = true
		}
	}
	var out []edit
	for _, p := range markdown.Places(f.Data) {
		id, ok := to[p.Value]
		switch {
		case !ok:
		case p.Kind == markdown.PlaceLink && examples[p.Line]:
			// An Examples cell may hold an identifier as data.
		case p.Kind == markdown.PlaceTag && f.Kind != tree.Questions:
		default:
			out = append(out, edit{p.Start, p.End, id})
		}
	}
	return append(out, handOvers(f, elements, to)...)
}

// handOvers returns the targets of the hand-overs of the rules among
// elements that are old identifiers of to.
func handOvers(f *tree.File, elements []lint.Element, to map[string]string) []edit {
	var doc *markdown.Doc
	var starts []int // the byte offset of each line of f
	var out []edit
	for _, e := range elements {
		id, ok := to[e.HandOver]
		if e.Type != lint.TypeRule || !ok {
			continue
		}
		if doc == nil {
			doc = markdown.Parse(f.Data)
			starts = lineStarts(f.Data)
		}
		for _, b := range doc.Blocks {
			if b.Fence == nil || b.Fence.Label != "guarded" || b.Line < e.Line || b.Line > e.End {
				continue
			}
			for _, ln := range b.Fence.Lines {
				start := starts[ln.N-1]
				end := len(f.Data)
				if ln.N < len(starts) {
					end = starts[ln.N]
				}
				text := strings.TrimSuffix(string(f.Data[start:end]), "\n")
				parsed := guarded.Parse([]guarded.SourceLine{{N: ln.N, Text: text}})
				if len(parsed.Lines) == 0 {
					continue
				}
				l := parsed.Lines[0]
				if l.Keyword == guarded.Otherwise && l.IsHandOver() && l.Target == e.HandOver {
					out = append(out, edit{start + l.TargetStart, start + l.TargetStart + len(l.Target), id})
				}
			}
		}
	}
	return out
}

func lineStarts(data []byte) []int {
	starts := []int{0}
	for i, c := range data {
		if c == '\n' && i+1 < len(data) {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// Manifest returns what `manifest` would write for the tree t with the
// files of rewritten in place of its own.
func Manifest(t *tree.Tree, rewritten map[string][]byte) []byte {
	after := *t
	after.Files = nil
	for _, f := range t.Files {
		g := *f
		if data, ok := rewritten[f.Rel]; ok {
			g.Data = data
		}
		after.Files = append(after.Files, &g)
	}
	return baseline.Write(lint.Check(&after).Elements)
}
