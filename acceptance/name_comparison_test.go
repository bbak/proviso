package acceptance

// Parameter lint.name-comparison: proviso/capabilities/lint/parameters.md

import (
	"strings"
	"testing"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// sameName is the oracle of lint.name-comparison: two names are the same
// when they are equal after each of its steps, in order.
func sameName(a, b string) bool { return comparedName(a) == comparedName(b) }

func comparedName(s string) string {
	// Step 1: White_Space at either end is dropped.
	s = strings.TrimFunc(s, func(r rune) bool { return unicode.Is(unicode.White_Space, r) })
	// Step 2: each run of spaces and tabs within the name becomes one space.
	var b strings.Builder
	inRun := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !inRun {
				b.WriteRune(' ')
			}
			inRun = true
			continue
		}
		inRun = false
		b.WriteRune(r)
	}
	// Step 3: each character becomes its simple lower-case form.
	s = strings.Map(unicode.ToLower, b.String())
	// Step 4: NFC.
	return norm.NFC.String(s)
}

// TestLintNameComparison checks that a glossary name, a term mark and a run
// of words in unmarked text are compared alike, as the steps of
// name_comparison say: a mark resolves wherever plain text would match, and
// the reverse.
func TestLintNameComparison(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		entry   string // the name of the cart's glossary entry
		written string // as written in the mark and in the text
		// markOnly: white space at either end of written belongs to a mark
		// only; in text it would not be part of the run of words.
		markOnly bool
	}{
		{name: "step1 spaces at either end", entry: "shelf label", written: " shelf label ", markOnly: true},
		{name: "step1 other white space at either end", entry: "shelf label", written: "\u00a0shelf label\u2003", markOnly: true},
		{name: "step1 a zero-width space stays", entry: "shelf label", written: "shelf label\u200b", markOnly: true},
		{name: "step1 white space at the end of a glossary name", entry: "shelf label\u00a0", written: "shelf label"},
		{name: "step2 a run of spaces and tabs", entry: "shelf label", written: "shelf \t label"},
		{name: "step2 a run next to punctuation", entry: "U.S. law", written: "U.S.\t law"},
		{name: "step2 a run in a glossary name", entry: "shelf  label", written: "shelf label"},
		{name: "step2 a non-breaking space joins no words", entry: "shelf label", written: "shelf\u00a0label"},
		{name: "step2 an em space joins no words", entry: "shelf label", written: "shelf\u2003label"},
		{name: "step3 case", entry: "shelf label", written: "SHELF Label"},
		{name: "step3 no full case folding of sharp s", entry: "Stra\u00dfe", written: "STRASSE"},
		{name: "step3 no full case folding of final sigma", entry: "\u039f\u0394\u039f\u03a3", written: "\u03bf\u03b4\u03bf\u03c2"},
		{name: "step3 simple lower case of the capital I with dot above", entry: "istanbul", written: "\u0130stanbul"},
		{name: "step3 simple lower case of the Kelvin sign", entry: "kelvin", written: "\u212aelvin"},
		{name: "step4 a decomposed mark or text", entry: "\u00dcbersicht", written: "u\u0308bersicht"},
		{name: "step4 a decomposed glossary name", entry: "U\u0308bersicht", written: "\u00fcbersicht"},
		{name: "step4 after lower case", entry: "\u01f0", written: "J\u030c"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			same := sameName(c.entry, c.written)
			text := "the cart keeps the [" + c.written + "]"
			if !c.markOnly {
				text += " beside the " + c.written
			}
			tr := withThenAnd(withTerms(baseline(), "cart", entryMD("cart.subject", c.entry)), text)
			at := thenAndLine(tr)
			o := checkGoesOn(t, tr)
			if same {
				o.lacks(t, "term-unresolved", featurePath, at)
			} else {
				o.has(t, "term-unresolved", featurePath, at)
			}
			if c.markOnly {
				return
			}
			if same {
				o.has(t, "term-unmarked", featurePath, at)
			} else {
				o.lacks(t, "term-unmarked", featurePath, at)
			}
		})
	}
}
