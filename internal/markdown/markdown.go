// Package markdown reads one Markdown file of a spec tree into the blocks
// lint works with: headings with their anchors, field lists, fenced blocks,
// tables and the anchors that stand anywhere else. It reads CommonMark with
// GitHub's tables, through goldmark, and knows nothing of the model's
// element types.
package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Doc is a parsed file: its document-level blocks in order, and the anchors
// that do not start the text of a document-level heading.
type Doc struct {
	Blocks       []Block
	StrayAnchors []Anchor
	src          []byte
	lines        []int // byte offset of the start of each line
}

// Block is one document-level block. Exactly one of the pointer fields is
// set; a block that lint does not look into has none set.
type Block struct {
	Line      int
	End       int    // the last line before the next block, or the file's last line
	FirstWord string // the first whitespace-separated word of the block's source
	Heading   *Heading
	List      *List
	Fence     *Fence
	Table     *Table
}

// Heading is a document-level ATX or setext heading.
type Heading struct {
	Level int
	// ID is the id of the anchor that starts the heading's text; HasID says
	// whether there is one.
	ID    string
	HasID bool
	// Text is the heading's plain text, without the starting anchor, trimmed.
	Text string
	// Raw is the heading's source text after the starting anchor, as
	// written, trimmed.
	Raw string
}

// List is a document-level list.
type List struct {
	Bullet bool // a bullet list marked with '-'
	Items  []Item
}

// Item is a list item: the source text of its first paragraph, with lines
// joined by single spaces, and the destinations of the links in it.
type Item struct {
	Line  int
	Text  string
	Links []string
}

// Fence is a fenced code block.
type Fence struct {
	Label string // the first word of the info string
	Lines []Line
}

// Line is one line of source with its line number.
type Line struct {
	N    int
	Text string
}

// Table is a GitHub table. Header and each row hold the source line, so
// that cells can be counted as written, not as goldmark pads them.
type Table struct {
	Header Line
	Rows   []Line
}

// Anchor is an anchor, at a line.
type Anchor struct {
	Line int
	ID   string
}

var md = goldmark.New(goldmark.WithExtensions(extension.Table))

// Parse reads src. A leading byte order mark is ignored.
func Parse(src []byte) *Doc {
	src = bytes.TrimPrefix(src, []byte("\xef\xbb\xbf"))
	d := &Doc{src: src, lines: lineStarts(src)}
	root := md.Parser().Parse(text.NewReader(src))
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		b := Block{Line: d.line(n.Pos()), FirstWord: firstWord(src[max(n.Pos(), 0):])}
		switch v := n.(type) {
		case *ast.Heading:
			b.Heading = d.heading(v, src)
		case *ast.List:
			b.List = d.list(v, src)
			d.strayIn(v, src)
		case *ast.FencedCodeBlock:
			b.Fence = d.fence(v, src)
		case *extast.Table:
			b.Table = d.table(v, src)
			b.Line = b.Table.Header.N
			d.strayIn(v, src)
		case *ast.HTMLBlock:
			d.strayInHTML(v, src)
		case *ast.CodeBlock:
			// An indented code block is text, like a fenced one.
		default:
			d.strayIn(v, src)
		}
		d.Blocks = append(d.Blocks, b)
	}
	for i := range d.Blocks {
		if i+1 < len(d.Blocks) {
			d.Blocks[i].End = max(d.Blocks[i+1].Line-1, d.Blocks[i].Line)
		} else {
			d.Blocks[i].End = len(d.lines)
		}
	}
	return d
}

// Source returns the lines from..to (1-based, inclusive) as written, each
// ended by a line feed, without the blank lines at their end.
func (d *Doc) Source(from, to int) string {
	var sb strings.Builder
	for n := from; n <= to && n <= len(d.lines); n++ {
		sb.WriteString(d.sourceLine(d.src, n))
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), " \t\n") + "\n"
}

// LineCount returns the number of lines of the file.
func (d *Doc) LineCount() int { return len(d.lines) }

// sourceLine returns line n (1-based) of src without its line ending.
func (d *Doc) sourceLine(src []byte, n int) string {
	start := d.lines[n-1]
	end := len(src)
	if n < len(d.lines) {
		end = d.lines[n]
	}
	return strings.TrimRight(string(src[start:end]), "\r\n")
}

func lineStarts(src []byte) []int {
	starts := []int{0}
	for i, c := range src {
		if c == '\n' && i+1 < len(src) {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func (d *Doc) line(offset int) int {
	if offset < 0 {
		return 1
	}
	lo, hi := 0, len(d.lines)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if d.lines[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo + 1
}

func firstWord(b []byte) string {
	f := bytes.Fields(b)
	if len(f) == 0 {
		return ""
	}
	return string(f[0])
}

func (d *Doc) heading(h *ast.Heading, src []byte) *Heading {
	out := &Heading{Level: h.Level}
	first := h.FirstChild()
	if id, ok := anchorAt(first, src); ok {
		out.ID, out.HasID = id, true
		closing := first.NextSibling().(*ast.RawHTML)
		first = closing.NextSibling()
		if lines := h.Lines(); lines.Len() > 0 {
			var sb strings.Builder
			from := closing.Segments.At(closing.Segments.Len() - 1).Stop
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				if seg.Stop <= from {
					continue
				}
				sb.Write(src[max(seg.Start, from):seg.Stop])
				sb.WriteByte('\n')
			}
			out.Raw = strings.TrimSpace(sb.String())
		}
	}
	var sb strings.Builder
	for c := first; c != nil; c = c.NextSibling() {
		if id, ok := anchorAt(c, src); ok {
			d.StrayAnchors = append(d.StrayAnchors, Anchor{d.line(c.Pos()), id})
			c = c.NextSibling()
			continue
		}
		plainText(c, src, &sb)
		d.strayIn(c, src)
	}
	out.Text = strings.TrimSpace(sb.String())
	return out
}

func (d *Doc) list(l *ast.List, src []byte) *List {
	out := &List{Bullet: l.Marker == '-'}
	for it := l.FirstChild(); it != nil; it = it.NextSibling() {
		item := Item{Line: d.line(it.Pos())}
		if c := it.FirstChild(); c != nil {
			if _, isText := c.(*ast.TextBlock); isText || c.Kind() == ast.KindParagraph {
				var parts []string
				lines := c.Lines()
				for i := 0; i < lines.Len(); i++ {
					seg := lines.At(i)
					parts = append(parts, strings.TrimSpace(string(seg.Value(src))))
				}
				item.Text = strings.Join(parts, " ")
				_ = ast.Walk(c, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
					if link, ok := n.(*ast.Link); ok && entering {
						item.Links = append(item.Links, string(link.Destination))
					}
					return ast.WalkContinue, nil
				})
			}
		}
		out.Items = append(out.Items, item)
	}
	return out
}

func (d *Doc) fence(f *ast.FencedCodeBlock, src []byte) *Fence {
	out := &Fence{}
	if f.Info != nil {
		out.Label = firstWord(f.Info.Segment.Value(src))
	}
	lines := f.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		t := strings.TrimRight(string(seg.Value(src)), "\r\n")
		if seg.Padding > 0 {
			t = strings.Repeat(" ", seg.Padding) + t
		}
		out.Lines = append(out.Lines, Line{d.line(seg.Start), t})
	}
	return out
}

func (d *Doc) table(t *extast.Table, src []byte) *Table {
	out := &Table{}
	for r := t.FirstChild(); r != nil; r = r.NextSibling() {
		n := d.line(r.Pos())
		l := Line{n, d.sourceLine(src, n)}
		if _, ok := r.(*extast.TableHeader); ok {
			out.Header = l
		} else {
			out.Rows = append(out.Rows, l)
		}
	}
	return out
}

// strayIn records every anchor inside n: none of them starts a
// document-level heading.
func (d *Doc) strayIn(n ast.Node, src []byte) {
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c.(type) {
		case *ast.CodeSpan, *ast.FencedCodeBlock, *ast.CodeBlock:
			return ast.WalkSkipChildren, nil
		case *ast.HTMLBlock:
			d.strayInHTML(c.(*ast.HTMLBlock), src)
			return ast.WalkSkipChildren, nil
		}
		if id, ok := anchorAt(c, src); ok {
			d.StrayAnchors = append(d.StrayAnchors, Anchor{d.line(c.Pos()), id})
		}
		return ast.WalkContinue, nil
	})
}

var htmlAnchor = regexp.MustCompile(`(?is)(<a(?:\s[^<>]*)?>)\s*</a\s*>`)

func (d *Doc) strayInHTML(h *ast.HTMLBlock, src []byte) {
	lines := h.Lines()
	var buf []byte
	var starts []int
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		starts = append(starts, len(buf))
		buf = append(buf, seg.Value(src)...)
	}
	if h.HasClosure() {
		starts = append(starts, len(buf))
		buf = append(buf, h.ClosureLine.Value(src)...)
	}
	for _, m := range htmlAnchor.FindAllSubmatchIndex(buf, -1) {
		id, ok := openTagID(string(buf[m[2]:m[3]]))
		if !ok {
			continue
		}
		i := 0
		for i+1 < len(starts) && starts[i+1] <= m[0] {
			i++
		}
		var seg text.Segment
		if i < lines.Len() {
			seg = lines.At(i)
		} else {
			seg = h.ClosureLine
		}
		d.StrayAnchors = append(d.StrayAnchors, Anchor{d.line(seg.Start), id})
	}
}

// anchorAt reports whether n is the opening tag of an anchor, that is a raw
// `<a>` tag with an id, directly followed by its closing tag.
func anchorAt(n ast.Node, src []byte) (string, bool) {
	open, ok := n.(*ast.RawHTML)
	if !ok {
		return "", false
	}
	close, ok := n.NextSibling().(*ast.RawHTML)
	if !ok || !closeTag.MatchString(rawValue(close, src)) {
		return "", false
	}
	return openTagID(rawValue(open, src))
}

var (
	openTag  = regexp.MustCompile(`(?is)^<a(\s[^<>]*)?>$`)
	closeTag = regexp.MustCompile(`(?is)^</a\s*>$`)
	attr     = regexp.MustCompile(`(?s)([a-zA-Z_:][-a-zA-Z0-9_:.]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>` + "`" + `]+)))?`)
)

func openTagID(tag string) (string, bool) {
	m := openTag.FindStringSubmatch(tag)
	if m == nil {
		return "", false
	}
	for _, a := range attr.FindAllStringSubmatch(m[1], -1) {
		if strings.EqualFold(a[1], "id") {
			return a[2] + a[3] + a[4], true
		}
	}
	return "", false
}

func rawValue(r *ast.RawHTML, src []byte) string {
	var sb strings.Builder
	for i := 0; i < r.Segments.Len(); i++ {
		seg := r.Segments.At(i)
		sb.Write(seg.Value(src))
	}
	return sb.String()
}

// plainText appends the text of an inline node: text, code span content and
// link text, without raw HTML.
func plainText(n ast.Node, src []byte, sb *strings.Builder) {
	switch v := n.(type) {
	case *ast.Text:
		sb.Write(v.Segment.Value(src))
		if v.SoftLineBreak() || v.HardLineBreak() {
			sb.WriteByte(' ')
		}
		return
	case *ast.String:
		sb.Write(v.Value)
		return
	case *ast.RawHTML:
		return
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		plainText(c, src, sb)
	}
}
