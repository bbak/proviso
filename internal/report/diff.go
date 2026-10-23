package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/lint"
)

type jsonPart struct {
	Part   string  `json:"part"`
	Change string  `json:"change"`
	Former *string `json:"former,omitempty"`
}

type jsonElement struct {
	ID        string     `json:"id"`
	Change    string     `json:"change"`
	File      string     `json:"file,omitempty"`
	Line      int        `json:"line,omitempty"`
	Text      string     `json:"text,omitempty"`
	Former    string     `json:"former,omitempty"`
	Parts     []jsonPart `json:"parts,omitempty"`
	Through   [][]string `json:"through,omitempty"`
	Undecided []lint.Pos `json:"undecided"`
}

type jsonDiff struct {
	Root     string        `json:"root"`
	Revision *string       `json:"revision"`
	Elements []jsonElement `json:"elements"`
}

// DiffJSON writes the element_diff: the spec root that was compared, the
// revision the former text came from, or null, and every element listed.
func DiffJSON(w io.Writer, root, revision string, entries []*baseline.Entry) error {
	d := jsonDiff{Root: root, Elements: diffElements(entries)}
	if revision != "" {
		d.Revision = &revision
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

// diffElements returns the elements of an element_diff.
func diffElements(entries []*baseline.Entry) []jsonElement {
	out := []jsonElement{}
	for _, e := range entries {
		je := jsonElement{ID: e.ID, Change: e.Change, Former: e.Former, Through: e.Through, Undecided: e.Undecided()}
		if je.Undecided == nil {
			je.Undecided = []lint.Pos{}
		}
		if el := e.Element; el != nil {
			je.File, je.Line, je.Text = el.File, el.Line, el.Text
		}
		for _, p := range e.Parts {
			jp := jsonPart{Part: p.Part, Change: p.Change}
			if p.Change != baseline.Added {
				jp.Former = &p.Former
			}
			je.Parts = append(je.Parts, jp)
		}
		out = append(out, je)
	}
	return out
}

// DiffText writes the element diff for people: the revision the former
// text came from, then each element with where it stands, the parts that
// changed, the terms it changed through and whether it is to be built,
// followed by its text, or its former text if it was removed. Nothing is
// written when nothing is listed.
func DiffText(w io.Writer, revision string, entries []*baseline.Entry) error {
	var sb strings.Builder
	if revision != "" && len(entries) > 0 {
		fmt.Fprintf(&sb, "former text from %s\n", revision)
	}
	for _, e := range entries {
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "%s %s", e.Change, e.ID)
		if el := e.Element; el != nil {
			fmt.Fprintf(&sb, " (%s:%d)", el.File, el.Line)
		}
		sb.WriteString("\n")
		for _, p := range e.Parts {
			fmt.Fprintf(&sb, "  %s %s\n", p.Change, p.Part)
			if p.Change != baseline.Added {
				indent(&sb, "    - ", p.Former)
			}
		}
		for _, c := range e.Through {
			fmt.Fprintf(&sb, "  through %s\n", strings.Join(c, " → "))
		}
		if u := e.Undecided(); len(u) > 0 {
			var at []string
			for _, p := range u {
				at = append(at, fmt.Sprintf("%s:%d", p.File, p.Line))
			}
			fmt.Fprintf(&sb, "  not to be built, undecided at %s\n", strings.Join(at, ", "))
		}
		if e.Element != nil {
			indent(&sb, "    | ", e.Element.Text)
		} else {
			indent(&sb, "    - ", e.Former)
		}
	}
	_, err := io.WriteString(w, strings.TrimPrefix(sb.String(), "\n"))
	return err
}

func indent(sb *strings.Builder, prefix, text string) {
	for _, l := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		sb.WriteString(strings.TrimRight(prefix+l, " ") + "\n")
	}
}
