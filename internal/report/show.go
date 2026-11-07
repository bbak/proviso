package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/query"
)

type jsonShown struct {
	ID      string  `json:"id"`
	Type    *string `json:"type"`
	Heading string  `json:"heading"`
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Parent  *string `json:"parent"`
	Text    string  `json:"text"`
}

type jsonView struct {
	Root     string      `json:"root"`
	Elements []jsonShown `json:"elements"`
}

// shown returns an element as element_view holds it: null for no element
// type and for no element parent.
func shown(f query.Found) jsonShown {
	s := jsonShown{ID: f.ID, Heading: f.Heading, File: f.File, Line: f.Line, Text: f.Text}
	if f.Type != "" {
		s.Type = &f.Type
	}
	if f.ElementParent != "" {
		s.Parent = &f.ElementParent
	}
	return s
}

// ShowJSON writes the element_view: the spec root that was read and every
// element found, in the order given; an empty list where none was.
func ShowJSON(w io.Writer, root string, found []query.Found) error {
	v := jsonView{Root: root, Elements: []jsonShown{}}
	for _, f := range found {
		v.Elements = append(v.Elements, shown(f))
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ShowText writes show_text: each element as a frontmatter of its id,
// type, heading, file, line and parent, each value as JSON writes it, then
// a blank line and its text without the blank lines at its end, with a
// blank line between one element and the next.
func ShowText(w io.Writer, found []query.Found) error {
	var sb strings.Builder
	for i, f := range found {
		if i > 0 {
			sb.WriteString("\n")
		}
		s := shown(f)
		sb.WriteString("---\n")
		for _, kv := range []struct {
			key string
			val any
		}{{"id", s.ID}, {"type", s.Type}, {"heading", s.Heading}, {"file", s.File}, {"line", s.Line}, {"parent", s.Parent}} {
			v, err := jsonValue(kv.val)
			if err != nil {
				return err
			}
			fmt.Fprintf(&sb, "%s: %s\n", kv.key, v)
		}
		sb.WriteString("---\n\n")
		sb.WriteString(withoutTrailingBlank(s.Text))
	}
	_, err := io.WriteString(w, sb.String())
	return err
}

// jsonValue writes v as JSON on one line, without escaping HTML.
func jsonValue(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// withoutTrailingBlank returns text without the blank lines at its end,
// its last line ended with a line feed.
func withoutTrailingBlank(text string) string {
	lines := strings.Split(text, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
