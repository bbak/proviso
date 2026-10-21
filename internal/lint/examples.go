package lint

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/guarded"
)

var noPart = regexp.MustCompile(`^no \(.+\)$`)

func isNo(v string) bool { return v == "no" || noPart.MatchString(v) }

func (l *linter) checkExamples(r *rule) {
	t := r.table
	if t == nil {
		l.report(Error, "examples-missing", r.pos(), "the rule %q has no Examples table after its rule block", r.id)
		return
	}
	tablePos := l.at(r, t.Header.N)
	columns := splitRow(t.Header.Text)

	var placeholders []string
	for _, ln := range r.block.Lines {
		for _, sp := range ln.Spans {
			if sp.Kind == guarded.Placeholder && !slices.Contains(placeholders, sp.Value) {
				placeholders = append(placeholders, sp.Value)
			}
		}
	}
	var missing []string
	for _, p := range append(slices.Clone(placeholders), "outcome") {
		if !slices.Contains(columns, p) {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		l.report(Error, "examples-column-missing", tablePos, "the Examples table lacks the columns: %s", strings.Join(missing, ", "))
	}
	guard := r.block.StepsOf(guarded.Provided)
	hasProvided := slices.Contains(columns, "provided")
	if hasProvided != (len(guard) > 0) {
		if hasProvided {
			l.report(Error, "provided-column", tablePos, "a provided column on a rule without a guard")
		} else {
			l.report(Error, "provided-column", tablePos, "no provided column on a rule with a guard")
		}
	}
	for i, c := range columns {
		if !slices.Contains(placeholders, c) && !slices.Contains(reserved, c) {
			l.report(Error, "examples-column-unknown", tablePos, "the column %q is neither a placeholder nor provided, outcome or why", c).col = i
		}
	}
	if len(t.Rows) == 0 {
		l.report(Error, "examples-empty", tablePos, "the Examples table has no rows")
		return
	}
	prov := slices.Index(columns, "provided")
	var values []string
	var keyCols []int
	for i, c := range columns {
		if slices.Contains(placeholders, c) {
			keyCols = append(keyCols, i)
		}
	}
	seen := map[string]int{}
	for _, row := range t.Rows {
		pos := l.at(r, row.N)
		cells := splitRow(row.Text)
		incomplete := len(cells) != len(columns)
		for i, c := range cells {
			if i < len(columns) && columns[i] != "why" && c == "" {
				incomplete = true
			}
			if firstWord(c) == "?" {
				l.report(Ask, "undecided", pos, "the cell %q is undecided", columnName(columns, i)).col = i
			}
		}
		if incomplete {
			l.report(Error, "examples-row-incomplete", pos, "the row needs one cell per column, each with a value except why")
		}
		if key, ok := rowKey(cells, keyCols); ok {
			if first, dup := seen[key]; dup {
				l.report(Error, "examples-row-repeated", pos, "the row has the same placeholder values as the row on line %d", first)
			} else {
				seen[key] = pos.Line
			}
		}
		if prov >= 0 && prov < len(cells) {
			v := cells[prov]
			if v != "yes" && v != "?" && !isNo(v) {
				l.report(Error, "provided-value-unknown", pos, "the provided cell %q is not yes, no, no (<part>) or ?", v)
			}
			values = append(values, v)
		}
	}
	if len(guard) == 0 || prov < 0 {
		return
	}
	yes, no := 0, 0
	for _, v := range values {
		switch {
		case v == "yes":
			yes++
		case isNo(v):
			no++
		}
	}
	var sides []string
	if yes == 0 {
		sides = append(sides, "the holding side (a yes)")
	}
	if no == 0 {
		sides = append(sides, "the failing side (a no)")
	}
	if len(sides) > 0 {
		l.report(Error, "examples-side-missing", tablePos, "no row shows %s", strings.Join(sides, " or "))
	}
	if parts := 1 + len(guard[0].Ands); parts >= 2 && no < parts {
		l.report(Error, "examples-part-missing", tablePos, "the guard has %d parts but only %d rows show it failing", parts, no)
	}
}

// rowKey joins the cells of a row in the placeholder columns, which tell
// one case from another. A row that lacks one of those cells is not
// compared: it is reported incomplete, and its case cannot be told. Nor
// is a row with an undecided placeholder cell, which equals no other cell.
func rowKey(cells []string, keyCols []int) (string, bool) {
	var key []string
	for _, i := range keyCols {
		if i >= len(cells) || firstWord(cells[i]) == "?" {
			return "", false
		}
		key = append(key, cells[i])
	}
	return fmt.Sprintf("%q", key), true
}

func columnName(columns []string, i int) string {
	if i < len(columns) {
		return columns[i]
	}
	return "beyond the last column"
}

// splitRow splits a table line into its cells as GitHub's tables do: at
// every pipe not escaped by a backslash, without a leading and a trailing
// pipe, each cell trimmed.
func splitRow(line string) []string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "|")
	if strings.HasSuffix(s, "|") {
		s = s[:len(s)-1]
	}
	var cells []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '|' && (i == 0 || s[i-1] != '\\') {
			cells = append(cells, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	return append(cells, strings.TrimSpace(s[start:]))
}
