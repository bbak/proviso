package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
)

// packHeadings are the headings of pack_text, in its order, with the
// heading context_pack names; Amendment holds the element diff.
var packHeadings = []struct{ title, heading string }{
	{"Builder notes", baseline.HeadingNotes},
	{"System", baseline.HeadingSystem},
	{"Amendment", ""},
	{"Frames", baseline.HeadingFrames},
	{"Context", baseline.HeadingContext},
}

// PackText writes pack_text: under each heading its elements, each with a
// blank line before and after, and under Amendment, before each element,
// what happened to it, and after it the former text of its parts. A
// heading under which no element appears is left out.
func PackText(w io.Writer, entries []*baseline.Entry, packed []baseline.Packed) error {
	var blocks []string // separated by a blank line
	for _, h := range packHeadings {
		var under []string
		if h.heading == "" {
			for _, e := range entries {
				under = append(under, amendment(e)...)
			}
		} else {
			for _, p := range packed {
				if p.Heading == h.heading {
					under = append(under, p.Element.Text)
				}
			}
		}
		if len(under) > 0 {
			blocks = append(blocks, "# "+h.title+"\n")
			blocks = append(blocks, under...)
		}
	}
	_, err := io.WriteString(w, strings.Join(blocks, "\n"))
	return err
}

// amendment returns the blocks of a listed element: the line saying what
// happened to it, and the second line if it is not to be built; its text,
// or its former text if it was removed; and the former text of each of its
// changed or removed parts.
func amendment(e *baseline.Entry) []string {
	var head string
	switch e.Change {
	case baseline.Added:
		head = fmt.Sprintf("Added `%s`.\n", e.ID)
	case baseline.Removed:
		head = fmt.Sprintf("Removed `%s`.\n", e.ID)
	case baseline.Changed:
		var parts []string
		for _, p := range e.Parts {
			parts = append(parts, fmt.Sprintf("`%s` %s", p.Part, p.Change))
		}
		head = fmt.Sprintf("Changed `%s`: %s.\n", e.ID, strings.Join(parts, ", "))
	case baseline.ChangedThrough:
		var chains []string
		for _, c := range e.Through {
			var ids []string
			for _, id := range c {
				ids = append(ids, "`"+id+"`")
			}
			chains = append(chains, strings.Join(ids, " → "))
		}
		head = fmt.Sprintf("Changed through `%s`: %s.\n", e.ID, strings.Join(chains, "; "))
	}
	if u := e.Undecided(); len(u) > 0 {
		var at []string
		for _, p := range u {
			at = append(at, fmt.Sprintf("%s:%d", p.File, p.Line))
		}
		head += fmt.Sprintf("Not to be built: %s.\n", strings.Join(at, ", "))
	}
	out := []string{head}
	if e.Element != nil {
		out = append(out, e.Element.Text)
	} else {
		out = append(out, fenced(e.Former))
	}
	for _, p := range e.Parts {
		if p.Change != baseline.Added {
			out = append(out, fmt.Sprintf("Former `%s`:\n", p.Part), fenced(p.Former))
		}
	}
	return out
}

// fenced puts former text in a fenced block labelled text, its fence
// longer than any run of backticks in it.
func fenced(text string) string {
	longest, run := 0, 0
	for _, r := range text {
		if r == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	fence := strings.Repeat("`", max(3, longest+1))
	body := strings.TrimRight(text, "\n")
	if body != "" {
		body += "\n"
	}
	return fence + "text\n" + body + fence + "\n"
}

type jsonPacked struct {
	ID      string `json:"id"`
	Heading string `json:"heading"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Text    string `json:"text"`
}

type jsonPack struct {
	Root     string        `json:"root"`
	Revision *string       `json:"revision"`
	Diff     []jsonElement `json:"diff"`
	Elements []jsonPacked  `json:"elements"`
}

// PackJSON writes the context_pack: the spec root the pack was assembled
// from, the revision the former text came from, or null, the element diff,
// and every other element of the pack, in the order of pack_text.
func PackJSON(w io.Writer, root, revision string, entries []*baseline.Entry, packed []baseline.Packed) error {
	p := jsonPack{Root: root, Diff: diffElements(entries), Elements: []jsonPacked{}}
	if revision != "" {
		p.Revision = &revision
	}
	for _, x := range packed {
		e := x.Element
		p.Elements = append(p.Elements, jsonPacked{e.ID, x.Heading, e.File, e.Line, e.Text})
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}
