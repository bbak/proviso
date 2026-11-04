package lint

import (
	"cmp"
	"slices"
)

// Severity is whether a finding is an error or an ask.
type Severity string

// The two severities.
const (
	Error Severity = "error"
	Ask   Severity = "ask"
)

// Pos is a place in the spec tree. File is relative to the directory that
// holds the spec root, with forward slashes.
type Pos struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// Finding is something lint reports.
type Finding struct {
	Kind     string   `json:"kind"`
	Severity Severity `json:"severity"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Message  string   `json:"message"`
	Related  []Pos    `json:"related,omitempty"`
	// col is where the finding stands in its line: a byte offset, or, in a
	// table row or header, the index of the cell. It orders findings of
	// one kind on one line.
	col int
}

// Sort orders findings by file path, compared byte by byte, then line, then
// kind, then position in the line, and, beyond what the tree requires, by
// message.
func Sort(fs []Finding) {
	slices.SortStableFunc(fs, func(a, b Finding) int {
		return cmp.Or(
			cmp.Compare(a.File, b.File),
			cmp.Compare(a.Line, b.Line),
			cmp.Compare(a.Kind, b.Kind),
			cmp.Compare(a.col, b.col),
			cmp.Compare(a.Message, b.Message),
		)
	})
}

// HasError reports whether any finding is an error.
func HasError(fs []Finding) bool {
	return slices.ContainsFunc(fs, func(f Finding) bool { return f.Severity == Error })
}
