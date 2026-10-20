package markdown

import (
	"slices"
	"testing"
)

func TestAnchors(t *testing.T) {
	cases := []struct {
		src    string
		id     string // the heading's identifier, "" for none
		strays []int  // lines of misplaced anchors
	}{
		{"## <a id=\"a.b\"></a> Rule: B\n", "a.b", nil},
		{"## <a id=\"a.b\"></a>Rule: B\n", "a.b", nil},
		{"<a id=\"a.b\"></a>\n## Rule: B\n", "", []int{1}},
		{"## Rule: B <a id=\"a.b\"></a>\n", "", []int{1}},
		{"# T\n\ntext <a id=\"a.b\"></a> more\n", "", []int{3}},
		{"# T\n\n`<a id=\"a.b\"></a>`\n", "", nil},
		{"# T\n\n```\n<a id=\"a.b\"></a>\n```\n", "", nil},
		{"# T\n\n<div>\n<a id=\"a.b\"></a>\n</div>\n", "", []int{4}},
		{"## <a id='x' class=\"c\"></a> X\n", "x", nil},
		{"## <a href=\"x\"></a> X\n", "", nil},
	}
	for _, c := range cases {
		d := Parse([]byte(c.src))
		id := ""
		for _, b := range d.Blocks {
			if b.Heading != nil && b.Heading.HasID {
				id = b.Heading.ID
			}
		}
		var strays []int
		for _, a := range d.StrayAnchors {
			strays = append(strays, a.Line)
		}
		if id != c.id || !slices.Equal(strays, c.strays) {
			t.Errorf("%q: id %q strays %v, want %q %v", c.src, id, strays, c.id, c.strays)
		}
	}
}

func TestBlocks(t *testing.T) {
	src := "# <a id=\"f\"></a> F `x`\n\n- Contact: Ann\n- Defined by: [x](a.md#b)\n\n## Goal\n\n```guarded\nWhen x\n\nThen y\n```\n\n| a | b |\n| --- | --- |\n| 1 |\n| 2 | 3 | 4 |\n"
	d := Parse([]byte(src))
	if len(d.Blocks) != 5 {
		t.Fatalf("%d blocks", len(d.Blocks))
	}
	if h := d.Blocks[0].Heading; h.Text != "F x" || h.Level != 1 {
		t.Errorf("heading %+v", h)
	}
	l := d.Blocks[1].List
	if !l.Bullet || len(l.Items) != 2 || l.Items[1].Line != 4 || l.Items[1].Text != "Defined by: [x](a.md#b)" || !slices.Equal(l.Items[1].Links, []string{"a.md#b"}) {
		t.Errorf("list %+v", l)
	}
	f := d.Blocks[3].Fence
	if f.Label != "guarded" || len(f.Lines) != 3 || f.Lines[2].N != 11 || d.Blocks[3].Line != 8 {
		t.Errorf("fence %+v", f)
	}
	tb := d.Blocks[4].Table
	if tb.Header.N != 14 || len(tb.Rows) != 2 || tb.Rows[1].Text != "| 2 | 3 | 4 |" {
		t.Errorf("table %+v", tb)
	}
}
