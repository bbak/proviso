package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// PlaceKind is what kind of place holds a value.
type PlaceKind int

// The places Places finds.
const (
	// PlaceAnchor is the id of an anchor that starts a document-level
	// heading.
	PlaceAnchor PlaceKind = iota
	// PlaceLink is the fragment of an inline link's destination, or of an
	// autolink, after its first "#".
	PlaceLink
	// PlaceTag is the text in brackets that starts a list item, such as
	// "a.b" in "- [a.b] Should …".
	PlaceTag
)

// Place is where a value stands in the source: the bytes Start..End of
// the source as given, on the line Line.
type Place struct {
	Kind       PlaceKind
	Line       int
	Start, End int
	Value      string
}

// Places returns the places in src that may hold an identifier, in the
// order of the source: the anchors that start document-level headings,
// the fragments of links outside code spans and fenced blocks, and the
// tags in brackets that start list items. A reference link, whose
// destination stands in its definition, and an image are no such place,
// and neither is a fragment written with an escape or an entity, which
// does not stand in the source as it reads, or on another line than the
// end of its link's text.
func Places(src []byte) []Place {
	trimmed := bytes.TrimPrefix(src, []byte("\xef\xbb\xbf"))
	bom := len(src) - len(trimmed)
	d := &Doc{src: trimmed, lines: lineStarts(trimmed)}
	root := md.Parser().Parse(text.NewReader(trimmed))
	var out []Place
	add := func(kind PlaceKind, start, end int) {
		out = append(out, Place{kind, d.line(start), start + bom, end + bom, string(trimmed[start:end])})
	}
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*ast.Heading); ok {
			if start, end, ok := anchorID(h.FirstChild(), trimmed); ok {
				add(PlaceAnchor, start, end)
			}
		}
	}
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.CodeSpan, *ast.FencedCodeBlock, *ast.CodeBlock, *ast.HTMLBlock, *ast.Image:
			return ast.WalkSkipChildren, nil
		case *ast.Link:
			if v.Reference == nil {
				if start, end, ok := fragment(v, string(v.Destination), trimmed); ok {
					add(PlaceLink, start, end)
				}
			}
		case *ast.AutoLink:
			if start, end, ok := fragment(v, string(v.URL(trimmed)), trimmed); ok {
				add(PlaceLink, start, end)
			}
		case *ast.ListItem:
			if start, end, ok := tag(v, trimmed); ok {
				add(PlaceTag, start, end)
			}
		}
		return ast.WalkContinue, nil
	})
	return out
}

// anchorID returns the bytes of the id of the anchor that n opens, if it
// opens one.
func anchorID(n ast.Node, src []byte) (int, int, bool) {
	if n == nil {
		return 0, 0, false
	}
	if _, ok := anchorAt(n, src); !ok {
		return 0, 0, false
	}
	open := n.(*ast.RawHTML)
	// The value as goldmark gives it, mapped back to the source.
	var value []byte
	var at []int // the source offset of each byte of value
	for i := 0; i < open.Segments.Len(); i++ {
		seg := open.Segments.At(i)
		for p := seg.Start; p < seg.Stop; p++ {
			value = append(value, src[p])
			at = append(at, p)
		}
	}
	m := openTag.FindSubmatchIndex(value)
	if m == nil || m[2] < 0 {
		return 0, 0, false
	}
	for _, a := range attr.FindAllSubmatchIndex(value[m[2]:m[3]], -1) {
		if !strings.EqualFold(string(value[m[2]+a[2]:m[2]+a[3]]), "id") {
			continue
		}
		for g := 4; g < len(a); g += 2 {
			if a[g] >= 0 {
				from, to := m[2]+a[g], m[2]+a[g+1]
				if from == to {
					return 0, 0, false
				}
				// An id that spans a line break is not one place.
				if at[to-1]-at[from] != to-1-from {
					return 0, 0, false
				}
				return at[from], at[to-1] + 1, true
			}
		}
		return 0, 0, false
	}
	return 0, 0, false
}

// fragment returns the bytes of the fragment of dest, the destination of
// the link or autolink n, where it stands in the source after the link's
// text, on the same line, and before whatever follows the link.
func fragment(n ast.Node, dest string, src []byte) (int, int, bool) {
	_, frag, ok := strings.Cut(dest, "#")
	if !ok || frag == "" || n.Pos() < 0 {
		return 0, 0, false
	}
	from := n.Pos()
	if _, isLink := n.(*ast.Link); isLink {
		from++
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			from = max(from, lastStop(c))
		}
	}
	// The destination stands on the line its link's text ends on.
	to := len(src)
	if eol := bytes.IndexByte(src[from:], '\n'); eol >= 0 {
		to = from + eol
	}
	if next := firstStart(n); next >= 0 && next < to {
		to = next
	}
	needle := "#" + frag
	for i := from; i < to; {
		j := bytes.Index(src[i:to], []byte(needle))
		if j < 0 {
			break
		}
		start := i + j + 1
		end := start + len(frag)
		if end == len(src) || strings.IndexByte(")> \t\r\n\"'", src[end]) >= 0 {
			return start, end, true
		}
		i = start
	}
	return 0, 0, false
}

// lastStop returns the end of the last source bytes n or what it holds
// stands on, or -1 if none is known.
func lastStop(n ast.Node) int {
	stop := -1
	switch v := n.(type) {
	case *ast.Text:
		stop = v.Segment.Stop
	case *ast.RawHTML:
		if v.Segments.Len() > 0 {
			stop = v.Segments.At(v.Segments.Len() - 1).Stop
		}
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		stop = max(stop, lastStop(c))
	}
	return stop
}

// firstStart returns where the inline content after n starts, within its
// block, or -1 if nothing known follows it.
func firstStart(n ast.Node) int {
	for p := n; p != nil && p.Type() == ast.TypeInline; p = p.Parent() {
		for s := p.NextSibling(); s != nil; s = s.NextSibling() {
			if s.Pos() >= 0 {
				return s.Pos()
			}
			if t, ok := s.(*ast.Text); ok {
				return t.Segment.Start
			}
		}
	}
	return -1
}

// tag returns the bytes of the text in brackets that starts the list item
// it, unless the brackets open a link.
func tag(it *ast.ListItem, src []byte) (int, int, bool) {
	c := it.FirstChild()
	if c == nil || c.Lines().Len() == 0 {
		return 0, 0, false
	}
	if _, isText := c.(*ast.TextBlock); !isText && c.Kind() != ast.KindParagraph {
		return 0, 0, false
	}
	seg := c.Lines().At(0)
	line := src[seg.Start:seg.Stop]
	if len(line) == 0 || line[0] != '[' {
		return 0, 0, false
	}
	end := bytes.IndexByte(line, ']')
	if end <= 1 {
		return 0, 0, false
	}
	if end+1 < len(line) && (line[end+1] == '(' || line[end+1] == '[') {
		return 0, 0, false
	}
	return seg.Start + 1, seg.Start + end, true
}
