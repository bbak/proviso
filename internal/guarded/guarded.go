// Package guarded reads the lines of a rule block written in Guarded
// Gherkin: their keywords, the steps they form, and the parts of each line,
// such as literals, backtick references, term marks and placeholders.
package guarded

import (
	"regexp"
	"strings"
)

// The keywords, matched exactly.
const (
	Given     = "Given"
	When      = "When"
	Provided  = "Provided"
	Then      = "Then"
	Otherwise = "Otherwise"
	And       = "And"
)

// HandOverArrow opens the target of a hand-over.
const HandOverArrow = "→"

// Rank gives the order keywords follow; And has none.
var Rank = map[string]int{Given: 1, When: 2, Provided: 3, Then: 4, Otherwise: 5}

// SpanKind is what a part of a line is.
type SpanKind int

// The kinds of span a line is split into.
const (
	Text        SpanKind = iota
	Literal              // "…"
	Code                 // `…`, a backtick reference
	Mark                 // [term]
	Placeholder          // <name>
)

// Span is a part of a line. Value is its content without delimiters.
type Span struct {
	Kind  SpanKind
	Start int // byte offset in the line
	End   int
	Value string
}

// Line is a non-blank line of a rule block.
type Line struct {
	N       int    // line number in the file
	Text    string // the line as written
	Keyword string // the first word if it is a keyword, else ""
	Spans   []Span
	// Target is the target of a hand-over on an Otherwise line, and
	// TargetStart its byte offset; TargetStart is -1 if there is none.
	Target      string
	TargetStart int
}

// IsHandOver reports whether the line is an Otherwise that hands over.
func (l *Line) IsHandOver() bool { return l.TargetStart >= 0 }

// Step is a line with a keyword other than And, and the And lines after it.
type Step struct {
	Head *Line
	Ands []*Line
}

// Lines returns the step's head and its And lines.
func (s *Step) Lines() []*Line { return append([]*Line{s.Head}, s.Ands...) }

// Placeholders returns the names of the placeholders the step uses.
func (s *Step) Placeholders() map[string]bool {
	m := map[string]bool{}
	for _, l := range s.Lines() {
		for _, sp := range l.Spans {
			if sp.Kind == Placeholder {
				m[sp.Value] = true
			}
		}
	}
	return m
}

// Block is a parsed rule block.
type Block struct {
	Lines []*Line
	Steps []*Step
	// LeadingAnds are And lines with no step before them.
	LeadingAnds []*Line
}

// SourceLine is a line of a rule block's source with its line number.
type SourceLine struct {
	N    int
	Text string
}

// Parse reads the lines of a rule block. Blank lines are left out.
func Parse(src []SourceLine) *Block {
	b := &Block{}
	var cur *Step
	for _, sl := range src {
		if strings.TrimSpace(sl.Text) == "" {
			continue
		}
		l := parseLine(sl)
		b.Lines = append(b.Lines, l)
		switch l.Keyword {
		case "":
			// Not part of any step.
		case And:
			if cur == nil {
				b.LeadingAnds = append(b.LeadingAnds, l)
			} else {
				cur.Ands = append(cur.Ands, l)
			}
		default:
			cur = &Step{Head: l}
			b.Steps = append(b.Steps, cur)
		}
	}
	return b
}

// StepsOf returns the steps whose keyword is kw.
func (b *Block) StepsOf(kw string) []*Step {
	var out []*Step
	for _, s := range b.Steps {
		if s.Head.Keyword == kw {
			out = append(out, s)
		}
	}
	return out
}

func parseLine(sl SourceLine) *Line {
	l := &Line{N: sl.N, Text: sl.Text, TargetStart: -1}
	trimmed := strings.TrimLeft(sl.Text, " \t")
	word := trimmed
	if i := strings.IndexAny(trimmed, " \t"); i >= 0 {
		word = trimmed[:i]
	}
	if _, ok := Rank[word]; ok || word == And {
		l.Keyword = word
	}
	if l.Keyword == Otherwise {
		offset := len(sl.Text) - len(trimmed) + len(word)
		rest := sl.Text[offset:]
		r := strings.TrimLeft(rest, " \t")
		if strings.HasPrefix(r, HandOverArrow) {
			after := r[len(HandOverArrow):]
			target := strings.TrimSpace(after)
			l.Target = target
			l.TargetStart = offset + (len(rest) - len(r)) + len(HandOverArrow) + (len(after) - len(strings.TrimLeft(after, " \t")))
		}
	}
	l.Spans = lex(sl.Text)
	return l
}

var placeholder = regexp.MustCompile(`^<([A-Za-z0-9_-]+)>`)

// lex splits a line into spans, left to right: a literal runs from a double
// quote to the next one, a backtick reference from a run of backticks to the
// next run of the same length, a term mark from `[` to the next `]` (not a
// Markdown link), and a placeholder is `<name>`. An opening delimiter
// without its closing one is text.
func lex(s string) []Span { return lexIn(s, false) }

// FieldMarks returns the term marks in the value of a field, left to right.
// In a field, only code spans, backslash escapes and term marks are read:
// brackets within a code span or escaped by a backslash mark nothing, and
// within a mark an escape stands for the character it escapes, as
// CommonMark reads backslash escapes. Quotes, angle brackets and character
// references are text.
func FieldMarks(s string) []Span {
	var out []Span
	for _, sp := range lexIn(s, true) {
		if sp.Kind == Mark {
			out = append(out, sp)
		}
	}
	return out
}

// lexIn is lex; in a field, it reads backslash escapes, and no literals and
// no placeholders.
func lexIn(s string, field bool) []Span {
	var spans []Span
	textStart := 0
	flush := func(end int) {
		if end > textStart {
			spans = append(spans, Span{Text, textStart, end, s[textStart:end]})
		}
	}
	i := 0
	for i < len(s) {
		var sp *Span
		switch s[i] {
		case '\\':
			if field && isEscapable(s, i) {
				i += 2
				continue
			}
		case '"':
			if field {
				break
			}
			if j := strings.IndexByte(s[i+1:], '"'); j >= 0 {
				end := i + 1 + j + 1
				sp = &Span{Literal, i, end, s[i+1 : end-1]}
			}
		case '`':
			n := 1
			for i+n < len(s) && s[i+n] == '`' {
				n++
			}
			if end, ok := closingTicks(s, i+n, n); ok {
				sp = &Span{Code, i, end + n, strings.TrimSpace(s[i+n : end])}
			} else {
				i += n
				continue
			}
		case '[':
			if j := closingBracket(s, i+1, field); j > i+1 {
				end := j + 1
				if end >= len(s) || s[end] != '(' {
					v := s[i+1 : j]
					if field {
						v = unescape(v)
					}
					sp = &Span{Mark, i, end, v}
				}
			}
		case '<':
			if field {
				break
			}
			if m := placeholder.FindStringSubmatch(s[i:]); m != nil {
				sp = &Span{Placeholder, i, i + len(m[0]), m[1]}
			}
		}
		if sp == nil {
			i++
			continue
		}
		flush(sp.Start)
		spans = append(spans, *sp)
		i = sp.End
		textStart = i
	}
	flush(len(s))
	return spans
}

// closingBracket returns the offset of the `]` that closes a mark opened
// before from, or -1 if a `[` comes first or none follows. In a field, a
// bracket escaped by a backslash neither opens nor closes.
func closingBracket(s string, from int, field bool) int {
	for i := from; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if field && isEscapable(s, i) {
				i++
			}
		case '[':
			return -1
		case ']':
			return i
		}
	}
	return -1
}

// isEscapable reports whether the backslash at i escapes the character
// after it: CommonMark escapes every ASCII punctuation character.
func isEscapable(s string, i int) bool {
	return i+1 < len(s) && strings.IndexByte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", s[i+1]) >= 0
}

// unescape replaces each backslash escape in s by the character it escapes.
func unescape(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && isEscapable(s, i) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func closingTicks(s string, from, n int) (int, bool) {
	for i := from; i < len(s); {
		if s[i] != '`' {
			i++
			continue
		}
		j := i
		for j < len(s) && s[j] == '`' {
			j++
		}
		if j-i == n {
			return i, true
		}
		i = j
	}
	return 0, false
}

// Without returns the line with the spans of the given kinds, and the
// hand-over target if withTarget, replaced by NUL bytes, which are neither
// word characters nor white space, one for each byte they replace, so that
// every byte stands where it stands in the line.
func (l *Line) Without(withTarget bool, kinds ...SpanKind) string {
	drop := map[SpanKind]bool{}
	for _, k := range kinds {
		drop[k] = true
	}
	var sb strings.Builder
	for _, sp := range l.Spans {
		if drop[sp.Kind] {
			sb.WriteString(strings.Repeat("\x00", sp.End-sp.Start))
			continue
		}
		v := l.Text[sp.Start:sp.End]
		if withTarget && l.IsHandOver() && sp.Kind == Text && sp.Start <= l.TargetStart && l.TargetStart < sp.End {
			end := min(l.TargetStart+len(l.Target), sp.End)
			v = l.Text[sp.Start:l.TargetStart] + strings.Repeat("\x00", end-l.TargetStart) + l.Text[end:sp.End]
		}
		sb.WriteString(v)
	}
	return sb.String()
}
