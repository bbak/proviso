package lint

import (
	"strings"

	"github.com/bbak/proviso/internal/guarded"
)

// The element types, as the parameters name them.
const (
	TypeSystem     = tSystem
	TypeCapability = tCapability
	TypeFeature    = tFeature
	TypeRule       = tRule
	TypeInvariant  = tInvariant
	TypeConstraint = tConstraint
	TypeGlossary   = tGlossary
	TypeParameter  = tParameter
	TypeReference  = tReference
	TypeContract   = tContract
)

// Result is what lint found in a tree, and the elements it read there, for
// the commands that build on lint's reading of the tree.
type Result struct {
	Findings []Finding
	Elements []Element
}

// Element is an element as lint read it, split into the pieces it holds.
// Texts are source text as written.
type Element struct {
	ID   string
	Type string // "" when the layout gives no element type
	File string // the report's path
	Line int    // the line of its heading
	End  int    // the last line before the next element's heading, or of the file
	Text string // its own text, from its heading to the next element's heading

	Heading  string // the heading's text after its anchor
	Fields   []Field
	Sections []Section
	Steps    []Step // of its rule block, for a rule
	Columns  []string
	Keys     []string // the columns of its Examples table that are placeholders
	Rows     []Row
	Body     string // everything else it holds

	Cap        string   // the capability it stands in; "" at system level
	Top        string   // the top-level contract element it is or belongs to
	Marks      []string // the identifiers of the glossary entries its rule block marks
	Refs       []Ref    // the backtick references of its rule block
	Represents []string // the identifiers of the glossary entries its Represents marks
	DefinedBy  string   // the identifier of the reference its Defined by links to
	Undecided  []Pos    // the places lint asks about as undecided
}

// Ref is a backtick reference as written, and the identifier of the
// contract element or parameter it names.
type Ref struct {
	Name, ID string
}

// Field is a field of an element.
type Field struct {
	Key, Value string
	Line       int
}

// Section is a section of an element, with its body.
type Section struct {
	Name, Body string
	Line       int
}

// Step is a step of a rule block: its keyword, and its text after the
// keyword with its And lines.
type Step struct {
	Keyword, Text string
}

// Row is a row of an Examples table.
type Row struct {
	Line  int
	Text  string
	Cells []string
}

func (l *linter) export() []Element {
	var out []Element
	for i, e := range l.elements {
		end := e.file.doc.LineCount()
		if i+1 < len(l.elements) && l.elements[i+1].file == e.file {
			end = l.elements[i+1].line - 1
		}
		x := Element{
			ID: e.id, Type: e.typ, File: e.file.path, Line: e.line, End: end,
			Text:    e.file.doc.Source(e.line, end),
			Heading: e.raw, Marks: e.marks, DefinedBy: e.definedBy,
			Cap: e.cap, Refs: e.refs, Represents: e.represents,
		}
		if e.typ == tContract {
			top := e
			for top.parent != nil {
				top = top.parent
			}
			x.Top = top.id
		}
		for _, f := range e.fields {
			x.Fields = append(x.Fields, Field{f.key, f.value, f.line})
		}
		l.split(e, &x)
		for _, f := range l.findings {
			if f.Kind == "undecided" && f.File == x.File && x.Line <= f.Line && f.Line <= end {
				x.Undecided = append(x.Undecided, Pos{f.File, f.Line})
			}
		}
		out = append(out, x)
	}
	return out
}

// split places the blocks of an element: its field list, its sections, its
// rule block and Examples table, and everything else, its body. A section
// runs to the next heading or, in a rule, to its rule block. The Examples
// table's header and separator rows belong to no piece.
func (l *linter) split(e *element, x *Element) {
	doc := e.file.doc
	var body []string
	var cur *Section
	add := func(s string) {
		if cur != nil {
			cur.Body += s
		} else {
			body = append(body, s)
		}
	}
	for _, b := range e.region {
		switch {
		case b.Heading != nil:
			x.Sections = append(x.Sections, Section{Name: b.Heading.Text, Line: b.Line})
			cur = &x.Sections[len(x.Sections)-1]
		case e.rule != nil && b.Line == e.rule.fenceLine:
			cur = nil // in a rule, a section ends at its rule block
		case e.rule != nil && e.rule.table != nil && b.Table == e.rule.table:
		case b.Line == e.fieldsLine && b.List != nil:
			for _, it := range b.List.Items {
				if fieldItem.FindStringSubmatch(it.Text) == nil {
					add("- " + it.Text + "\n")
				}
			}
		default:
			add(doc.Source(b.Line, b.End))
		}
	}
	x.Body = strings.Join(body, "")
	if e.rule == nil {
		return
	}
	for _, s := range e.rule.block.Steps {
		text := strings.TrimLeft(s.Head.Text, " \t")[len(s.Head.Keyword):]
		for _, a := range s.Ands {
			text += "\n" + a.Text
		}
		x.Steps = append(x.Steps, Step{s.Head.Keyword, strings.TrimSpace(text)})
	}
	if t := e.rule.table; t != nil {
		x.Columns = splitRow(t.Header.Text)
		ph := placeholdersOf(e.rule.block)
		for _, c := range x.Columns {
			if ph[c] {
				x.Keys = append(x.Keys, c)
			}
		}
		for _, r := range t.Rows {
			x.Rows = append(x.Rows, Row{r.N, r.Text, splitRow(r.Text)})
		}
	}
}

func placeholdersOf(b *guarded.Block) map[string]bool {
	m := map[string]bool{}
	for _, ln := range b.Lines {
		for _, sp := range ln.Spans {
			if sp.Kind == guarded.Placeholder {
				m[sp.Value] = true
			}
		}
	}
	return m
}
