package lint

import (
	"slices"
	"strings"
	"testing"
)

// TestMarkdown checks how lint reads the Markdown of a tree: CommonMark
// with GitHub's tables.
func TestMarkdown(t *testing.T) {
	const block = "When      lint reads <id>\nThen      x\n  And     y\n"
	const table = "| id | outcome |\n| --- | --- |\n| a | x |\n"
	feature := func(f map[string]string, from, to string) map[string]string {
		p := "capabilities/c/features/f.md"
		f[p] = strings.Replace(f[p], from, to, 1)
		return f
	}
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"a later table is prose", withRule(block, table+"\n| notes |\n| --- |\n"), nil},
		{"a table only before the block", feature(withRule(block, ""), "```guarded", table+"\n```guarded"),
			[]string{"examples-missing@capabilities/c/features/f.md:13"}},
		{"an escaped pipe within a cell", withRule(block, "| id | outcome |\n| --- | --- |\n| a \\| b | x |\n"), nil},
		{"attribute names ignore case", feature(base(), `<a id="c.f.one">`, `<a ID="c.f.one">`), nil},
		{"brackets before ( mark nothing", withRule("When      lint reads <id> of the [spec tree](notes.md) of\nThen      lint reads the [rest of it](notes.md)\n  And     lint reads the [spec tree][tree]\n", table),
			[]string{"term-unmarked@capabilities/c/features/f.md:20"}},
		{"a trailing dot stays in the target", withRule("When      lint reads <id>\nProvided  <id> is 1\nThen      x\n  And     y\nOtherwise → c.f.one.\n", "| id | provided | outcome |\n| --- | --- | --- |\n| a | yes | x |\n| b | no | y |\n"),
			[]string{"hand-over-unresolved@capabilities/c/features/f.md:24"}},
		{"the label is the first word of the info string", feature(withRule(block, table), "```guarded\n", "```guarded strict\n"), nil},
		{"placeholders of either case", withRule("When      lint reads <file>\nThen      <File>\n  And     y\n", "| file | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"examples-column-missing@capabilities/c/features/f.md:25", "placeholder-unbound@capabilities/c/features/f.md:21"}},
		{"a byte order mark", func() map[string]string { f := base(); f["system.md"] = "\xef\xbb\xbf" + f["system.md"]; return f }(), nil},
		{"a setext heading", func() map[string]string {
			f := base()
			f["system.md"] = strings.Replace(f["system.md"], "# <a id=\"system\"></a> Sys\n", "<a id=\"system\"></a> Sys\n===\n", 1)
			f["system.md"] = strings.Replace(f["system.md"], "## Goal\n", "Goal\n----\n", 1)
			return f
		}(), nil},
		{"an unplaced element has no sections or fields", feature(base(), "## <a id=\"c.f.one\"></a> Rule: One\n", "## <a id=\"c.f.y\"></a> Something\n\n- Owner: Search team\n- Contact: ?\n\n### Goal\n\n?\n\n## <a id=\"c.f.one\"></a> Rule: One\n"),
			[]string{"element-unplaced@capabilities/c/features/f.md:13"}},
		{"a list marked with * is prose", feature(base(), "- Contact: Ann", "* Contact: Ann"),
			[]string{"field-missing@capabilities/c/features/f.md:1"}},
		{"a numbered list is prose", feature(base(), "- Contact: Ann", "1. Contact: Ann"),
			[]string{"field-missing@capabilities/c/features/f.md:1"}},
		{"a heading in a list item is text", feature(base(), "- Contact: Ann", "- Contact: Ann\n\n- x\n\n  ## Goal\n"), nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := kinds(lintFiles(c.files))
			slices.Sort(got)
			want := slices.Clone(c.want)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("got  %v\nwant %v", got, want)
			}
		})
	}
}

// TestOrderInLine checks that findings of one kind on one line appear in
// the order of their places in the line.
func TestOrderInLine(t *testing.T) {
	open := "\n## <a id=\"c.open\"></a> open thing\n\n- Definition: ?\n- Contact: x\n- Status: open\n- Question: q?\n"
	cases := []struct {
		name, line string
		want       []string
	}{
		{"unmarked", "the guard of the spec tree, then the guard", []string{`"guard"`, `"spec tree"`, `"guard"`}},
		{"unmarked after pieces", "the \"x\" `limit` [spec tree] <id> guard, then spec tree", []string{`"guard"`, `"spec tree"`}},
		{"undecided", "lint reads <id> of the [open thing] and ? or `limit` ?", []string{"[open thing]", `"?"`, `"?"`}},
		{"undecided first", "lint reads <id> ? the [open thing]", []string{`"?"`, "[open thing]"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := withRule("When      "+c.line+"\nThen      x\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n")
			f["capabilities/c/glossary.md"] += open
			var got []string
			for _, fd := range lintFiles(f) {
				if fd.Line != 20 || fd.Kind == "examples-column-missing" {
					continue
				}
				for _, w := range []string{`"guard"`, `"spec tree"`, `"?"`, "[open thing]"} {
					if strings.Contains(fd.Message, w) {
						got = append(got, w)
					}
				}
			}
			if !slices.Equal(got, c.want) {
				t.Errorf("got  %v\nwant %v", got, c.want)
			}
		})
	}
}

// TestAnchorsInLine checks that misplaced anchors on one line appear in
// the order in which they stand.
func TestAnchorsInLine(t *testing.T) {
	for _, para := range []string{
		"Text <a id=\"z\"></a> and <a id=\"a\"></a>.\n",
		"<p><a id=\"z\"></a> and <a id=\"a\"></a></p>\n",
	} {
		f := base()
		f["capabilities/c/features/f.md"] += "\n" + para
		var got []string
		for _, fd := range lintFiles(f) {
			if fd.Kind == "anchor-misplaced" {
				got = append(got, fd.Message)
			}
		}
		if len(got) != 2 || !strings.Contains(got[0], `"z"`) || !strings.Contains(got[1], `"a"`) {
			t.Errorf("%q: got %v", para, got)
		}
	}
}
