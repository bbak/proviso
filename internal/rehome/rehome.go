// Package rehome plans the renames that give every element of a spec tree
// an identifier naming its identifier home, finds the renames that would
// take an identifier, and rewrites the places rehome_references lists.
package rehome

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

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

// Mention is an identifier mention of an old identifier of a rename: the
// line Line of the file File, the report's path, on which the code spans
// whose whole content is Identifier stand Count times, and the identifier
// Element has before the run of the element whose text holds that line,
// "" where none does.
type Mention struct {
	Identifier string
	File       string
	Line       int
	Count      int
	Element    string
}

// Picked names an identifier mention as a mention_list does: by its
// identifier, its file, the report's path, and its line.
type Picked struct {
	Identifier string
	File       string
	Line       int
}

// Mentions returns the identifier mentions of the old identifiers of
// renames in the files of t, one for each file, line and identifier,
// ordered by the file's path, then the line, then the identifier, paths
// and identifiers compared byte by byte, following
// upkeep.rehome.mentions and upkeep.rehome.order. elements are the
// elements lint read in t.
func Mentions(t *tree.Tree, elements []lint.Element, renames []Rename) []Mention {
	to := targets(renames)
	byFile := elementsByFile(elements)
	var out []Mention
	for _, f := range t.Files {
		if f.Kind == tree.Settings || f.Kind == tree.Manifest {
			continue
		}
		file := path.Join(tree.RootName, f.Rel)
		at := map[Picked]int{}
		for _, p := range markdown.CodeSpans(f.Data) {
			if _, ok := to[p.Value]; ok {
				at[Picked{p.Value, file, p.Line}]++
			}
		}
		for k, n := range at {
			out = append(out, Mention{k.Identifier, k.File, k.Line, n, holder(byFile[file], k.Line)})
		}
	}
	slices.SortFunc(out, func(a, b Mention) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Identifier, b.Identifier))
	})
	return out
}

// holder returns the identifier of the element among elements whose text
// holds the line, "" if none does.
func holder(elements []lint.Element, line int) string {
	for _, e := range elements {
		if e.Line <= line && line <= e.End {
			return e.ID
		}
	}
	return ""
}

func targets(renames []Rename) map[string]string {
	to := map[string]string{}
	for _, r := range renames {
		to[r.From] = r.To
	}
	return to
}

func elementsByFile(elements []lint.Element) map[string][]lint.Element {
	byFile := map[string][]lint.Element{}
	for _, e := range elements {
		byFile[e.File] = append(byFile[e.File], e)
	}
	return byFile
}

// edit is a place of a file that is to hold a new identifier.
type edit struct {
	start, end int
	to         string
}

// Rewrite returns, by their paths relative to the spec root, the files of
// t that hold a place rehome_references lists for an old identifier of
// renames, or an identifier mention picked names, each with every such
// place, and every identifier mention of an old identifier on a line
// picked names for it, holding the new identifier, and nothing else of it
// changed. elements are the elements lint read in t.
func Rewrite(t *tree.Tree, elements []lint.Element, renames []Rename, picked []Picked) map[string][]byte {
	to := targets(renames)
	out := map[string][]byte{}
	if len(to) == 0 {
		return out
	}
	byFile := elementsByFile(elements)
	pick := map[Picked]bool{}
	for _, p := range picked {
		pick[p] = true
	}
	for _, f := range t.Files {
		if f.Kind == tree.Settings || f.Kind == tree.Manifest {
			continue
		}
		file := path.Join(tree.RootName, f.Rel)
		edits := places(f, byFile[file], to)
		if len(pick) > 0 {
			for _, p := range markdown.CodeSpans(f.Data) {
				if id, ok := to[p.Value]; ok && pick[Picked{p.Value, file, p.Line}] {
					edits = append(edits, edit{p.Start, p.End, id})
				}
			}
		}
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

// ParseMentionList reads a mention_list: a JSON object whose member
// mentions is an array of objects, each with the string members identifier
// and file and the member line, a whole number from 1. Other members are
// not read. It returns the mentions it names, each once, in the order
// they are first named.
func ParseMentionList(data []byte) ([]Picked, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("it is not UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("it is not JSON: %v", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("it is not JSON: something follows its value")
	}
	top, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("it is not a JSON object")
	}
	raw, ok := top["mentions"]
	if !ok {
		return nil, errors.New("it has no member mentions")
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, errors.New("the member mentions is not an array")
	}
	var out []Picked
	seen := map[Picked]bool{}
	for i, it := range items {
		at := fmt.Sprintf("mention %d", i+1)
		m, ok := it.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s is not an object", at)
		}
		var p Picked
		for key, dst := range map[string]*string{"identifier": &p.Identifier, "file": &p.File} {
			s, ok := m[key].(string)
			if !ok {
				return nil, fmt.Errorf("%s has no string %s", at, key)
			}
			*dst = s
		}
		n, ok := m["line"].(json.Number)
		if !ok {
			return nil, fmt.Errorf("%s has no number line", at)
		}
		f, err := strconv.ParseFloat(n.String(), 64)
		if err != nil || f < 1 || f != math.Trunc(f) || f > math.MaxInt32 {
			return nil, fmt.Errorf("the line of %s, %s, is no whole number from 1", at, n)
		}
		p.Line = int(f)
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out, nil
}

// Unlisted returns the mentions of picked that are none of mentions, in
// the order of picked.
func Unlisted(picked []Picked, mentions []Mention) []Picked {
	listed := map[Picked]bool{}
	for _, m := range mentions {
		listed[Picked{m.Identifier, m.File, m.Line}] = true
	}
	var out []Picked
	for _, p := range picked {
		if !listed[p] {
			out = append(out, p)
		}
	}
	return out
}
