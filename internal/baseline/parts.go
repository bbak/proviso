// Package baseline splits the elements lint read into parts with
// fingerprints, writes and reads the manifest that records them, and
// compares a tree with it: the element diff.
package baseline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/lint"
)

// Part is a part of an element: its key as part_kinds writes it, its text
// as written, and its fingerprint.
type Part struct {
	Key, Text, FP string
}

// Normalize makes each run of spaces, tabs and line feeds one space and
// leaves none at the start or end. Nothing else counts as the same.
func Normalize(s string) string {
	return strings.Join(strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }), " ")
}

// Fingerprint is "h1:" and the SHA-256, in lowercase hexadecimal, of the
// normalised text.
func Fingerprint(s string) string {
	sum := sha256.Sum256([]byte(Normalize(s)))
	return "h1:" + hex.EncodeToString(sum[:])
}

// Parts splits an element into its parts, one for each entry of
// part_kinds it holds. An element below it is not one of them.
func Parts(e lint.Element) []Part {
	out := []Part{part("heading", e.Heading)}
	for _, f := range e.Fields {
		out = append(out, part("field "+f.Key, f.Value))
	}
	for _, s := range e.Sections {
		out = append(out, part("section "+s.Name, s.Body))
	}
	seen := map[string]int{}
	numbered := func(key string) string {
		seen[key]++
		if n := seen[key]; n > 1 {
			return fmt.Sprintf("%s #%d", key, n)
		}
		return key
	}
	for _, s := range e.Steps {
		out = append(out, part(numbered("step "+s.Keyword), s.Text))
	}
	for _, r := range e.Rows {
		out = append(out, part(numbered("row "+rowKey(e, r)), r.Text))
	}
	if Normalize(e.Body) != "" {
		out = append(out, part("body", e.Body))
	}
	return out
}

func part(key, text string) Part { return Part{key, text, Fingerprint(text)} }

// rowKey is the fingerprint of a row's placeholder values: for each
// placeholder column, in the byte order of the column names, the name, "="
// and the cell, joined by tabs.
func rowKey(e lint.Element, r lint.Row) string {
	keys := slices.Clone(e.Keys)
	slices.Sort(keys)
	var vals []string
	for _, k := range keys {
		cell := ""
		if i := slices.Index(e.Columns, k); i < len(r.Cells) {
			cell = r.Cells[i]
		}
		vals = append(vals, k+"="+cell)
	}
	return Fingerprint(strings.Join(vals, "\t"))
}
