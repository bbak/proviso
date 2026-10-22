package lint

import (
	"regexp"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/markdown"
	"github.com/bbak/proviso/internal/tree"
)

// The element types, named as the parameters name them.
const (
	tSystem     = "system"
	tCapability = "capability"
	tFeature    = "feature"
	tRule       = "rule"
	tInvariant  = "invariant"
	tConstraint = "constraint"
	tGlossary   = "glossary entry"
	tParameter  = "parameter"
	tReference  = "reference"
	tContract   = "contract element"
)

// sectionNames is the parameter section_names: the sections of each element
// type, and which of them are required.
var sectionNames = map[string]struct{ all, required []string }{
	tSystem:     {[]string{"Goal", "Rationale", "Scope", "Out of scope", "Audience", "Design boundary"}, []string{"Goal", "Rationale", "Scope", "Out of scope", "Audience", "Design boundary"}},
	tCapability: {[]string{"Goal", "Rationale"}, []string{"Goal", "Rationale"}},
	tFeature:    {[]string{"Goal", "Rationale"}, []string{"Goal", "Rationale"}},
	tRule:       {[]string{"Rationale"}, nil},
	tInvariant:  {[]string{"Rationale", "Ranges over", "Property"}, []string{"Rationale", "Ranges over", "Property"}},
	tConstraint: {[]string{"Rationale", "Property"}, []string{"Rationale", "Property"}},
}

// fieldNames is the parameter field_names: the fields of each element type,
// and which of them are required.
var fieldNames = map[string]struct{ all, required []string }{
	tFeature:   {[]string{"Contact"}, []string{"Contact"}},
	tGlossary:  {[]string{"Definition", "Defined by", "Deviations", "Contact", "Status", "Aliases", "Question"}, []string{"Contact", "Status"}},
	tParameter: {[]string{"Value", "Unit", "Configurable", "Rationale"}, []string{"Configurable", "Rationale"}},
	tReference: {[]string{"Version", "Link", "Contact"}, []string{"Version", "Link", "Contact"}},
	tContract:  {[]string{"Represents"}, nil},
}

type file struct {
	*tree.File
	path string // the report's path
	doc  *markdown.Doc
}

type field struct {
	line       int
	key, value string
	links      []string
}

type section struct {
	name      string
	line      int
	firstWord string // of its body; "" for an empty body
}

type element struct {
	id    string
	typ   string // "" when the layout gives no element type
	file  *file
	line  int
	level int
	name  string // the heading's text after its anchor
	cap   string // "" at system level

	parent   *element // the element a contract member belongs to
	members  map[string][]*element
	fields   []field
	sections []section
	region   []markdown.Block // the blocks after the heading, up to the next element

	// What the commands that build on lint read of the element.
	raw        string   // the heading's source text after its anchor
	fieldsLine int      // the line of its field list; 0 if it has none
	rule       *rule    // its rule block and Examples table, for a rule
	marks      []string // the glossary entries its rule block marks
	definedBy  string   // the reference its Defined by links to
	refs       []Ref    // the backtick references of its rule block
	represents []string // the glossary entries its Represents marks
}

func (e *element) pos() Pos { return Pos{e.file.path, e.line} }

// field returns the first field with the key, if any.
func (e *element) field(key string) (field, bool) {
	for _, f := range e.fields {
		if f.key == key {
			return f, true
		}
	}
	return field{}, false
}

var fieldItem = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9 -]*?)\s*:(.*)$`)

// build reads the elements of one file and reports what the file's layout
// and structure get wrong.
func (l *linter) build(f *file) {
	doc := f.doc
	for _, a := range doc.StrayAnchors {
		l.report(Error, "anchor-misplaced", Pos{f.path, a.Line}, "the anchor %q does not start the text of a heading", a.ID)
	}
	sibling := f.Kind.IsSibling()
	first := -1
	for i, b := range doc.Blocks {
		if b.Heading != nil {
			first = i
			break
		}
	}
	if sibling {
		l.checkTitle(f, first)
	}
	if f.Kind == tree.System || f.Kind == tree.Capability || f.Kind == tree.Feature {
		if first < 0 || doc.Blocks[first].Heading.Level != 1 || !doc.Blocks[first].Heading.HasID {
			l.report(Error, "element-missing", Pos{f.path, 1}, "the file does not open with its element: a level 1 heading with an anchor")
		}
	}
	var cur *element
	var stack []*element // open contract elements, by heading level
	for i, b := range doc.Blocks {
		h := b.Heading
		if h == nil {
			if cur != nil {
				cur.region = append(cur.region, b)
			}
			continue
		}
		if i == first && sibling && isTitle(h) {
			cur = nil
			continue
		}
		if h.HasID {
			e := &element{id: h.ID, typ: placeType(f.Kind, h), file: f, line: b.Line, level: h.Level, name: h.Text, cap: f.Cap, members: map[string][]*element{}, raw: h.Raw}
			if e.typ == "" {
				l.report(Error, "element-unplaced", e.pos(), "the heading of %q gives no element type in this file", e.id)
			}
			if e.typ == tContract {
				for len(stack) > 0 && stack[len(stack)-1].level >= e.level {
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 0 {
					e.parent = stack[len(stack)-1]
					e.parent.members[e.name] = append(e.parent.members[e.name], e)
				}
				stack = append(stack, e)
			}
			if i+1 < len(doc.Blocks) {
				if lst := doc.Blocks[i+1].List; lst != nil && lst.Bullet {
					e.fields = readFields(lst)
					e.fieldsLine = doc.Blocks[i+1].Line
				}
			}
			l.elements = append(l.elements, e)
			cur = e
			continue
		}
		if cur == nil || cur.typ == "" {
			continue
		}
		cur.region = append(cur.region, b)
		if !slices.Contains(sectionNames[cur.typ].all, h.Text) {
			l.report(Error, "heading-unknown", Pos{f.path, b.Line}, "%q is no section of a %s", h.Text, cur.typ)
			continue
		}
		if h.Level != cur.level+1 {
			l.report(Error, "section-level", Pos{f.path, b.Line}, "the section %q is at heading level %d, not one below its element's %d", h.Text, h.Level, cur.level)
		}
		s := section{name: h.Text, line: b.Line}
		if i+1 < len(doc.Blocks) && doc.Blocks[i+1].Heading == nil {
			s.firstWord = doc.Blocks[i+1].FirstWord
		}
		cur.sections = append(cur.sections, s)
	}
}

func (l *linter) checkTitle(f *file, first int) {
	if first >= 0 {
		b := f.doc.Blocks[first]
		if !isTitle(b.Heading) {
			l.report(Error, "title-missing", Pos{f.path, b.Line}, "the first heading is not a title: a level 1 heading without an anchor")
		}
	}
	for i := range f.doc.Blocks {
		if i != first || !isTitle(f.doc.Blocks[i].Heading) {
			return
		}
	}
	l.report(Error, "sibling-empty", Pos{f.path, 1}, "the file holds nothing but its title")
}

func isTitle(h *markdown.Heading) bool { return h.Level == 1 && !h.HasID }

// placeType is the parameter tree_layout: the element type a heading with an
// anchor has in a file of the kind k, or "" if it has none.
func placeType(k tree.Kind, h *markdown.Heading) string {
	if k != tree.Feature && (strings.HasPrefix(h.Text, "Rule:") || strings.HasPrefix(h.Text, "Invariant:")) {
		return ""
	}
	switch k {
	case tree.System:
		if h.Level == 1 {
			return tSystem
		}
	case tree.Capability:
		if h.Level == 1 {
			return tCapability
		}
	case tree.Feature:
		switch {
		case h.Level == 1:
			return tFeature
		case h.Level == 2 && strings.HasPrefix(h.Text, "Rule:"):
			return tRule
		case h.Level == 2 && strings.HasPrefix(h.Text, "Invariant:"):
			return tInvariant
		}
	case tree.Glossary, tree.Parameters, tree.References, tree.Invariants, tree.Constraints:
		if h.Level == 2 {
			return map[tree.Kind]string{tree.Glossary: tGlossary, tree.Parameters: tParameter, tree.References: tReference, tree.Invariants: tInvariant, tree.Constraints: tConstraint}[k]
		}
	case tree.Contract:
		if h.Level >= 2 {
			return tContract
		}
	}
	return ""
}

func readFields(lst *markdown.List) []field {
	var out []field
	for _, it := range lst.Items {
		m := fieldItem.FindStringSubmatch(it.Text)
		if m == nil {
			continue
		}
		out = append(out, field{line: it.Line, key: m[1], value: strings.TrimSpace(m[2]), links: it.Links})
	}
	return out
}
