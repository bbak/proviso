package lint

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/bbak/proviso/internal/guarded"
	"github.com/bbak/proviso/internal/markdown"
)

// rule is a rule element with its one rule block, read.
type rule struct {
	*element
	fenceLine int
	block     *guarded.Block
	table     *markdown.Table
	next      *element // the rule its hand-over reaches, if any
}

func (l *linter) checkRules() {
	var rules []*rule
	for _, e := range l.elements {
		if e.typ != tRule {
			continue
		}
		var fences []int
		for i, b := range e.region {
			if b.Fence != nil && b.Fence.Label == "guarded" {
				fences = append(fences, i)
			}
		}
		if len(fences) != 1 {
			l.report(Error, "rule-block-count", e.pos(), "the rule %q holds %d rule blocks, not 1", e.id, len(fences))
			continue
		}
		fb := e.region[fences[0]]
		var src []guarded.SourceLine
		for _, ln := range fb.Fence.Lines {
			src = append(src, guarded.SourceLine{N: ln.N, Text: ln.Text})
		}
		r := &rule{element: e, fenceLine: fb.Line, block: guarded.Parse(src)}
		for _, b := range e.region[fences[0]+1:] {
			if b.Table != nil {
				r.table = b.Table
				break
			}
		}
		rules = append(rules, r)
		l.checkBlock(r)
		l.checkExamples(r)
	}
	l.checkCycles(rules)
}

func (l *linter) at(r *rule, line int) Pos { return Pos{r.file.path, line} }

func (l *linter) checkBlock(r *rule) {
	b := r.block
	// Keywords.
	for _, ln := range b.Lines {
		if ln.Keyword == "" {
			l.report(Error, "keyword-unknown", l.at(r, ln.N), "the line does not start with a keyword")
		}
	}
	l.checkOrder(r)
	counts := map[string]int{}
	for _, s := range b.Steps {
		counts[s.Head.Keyword]++
	}
	var wrong []string
	for _, kw := range []string{guarded.When, guarded.Provided, guarded.Then, guarded.Otherwise} {
		n := counts[kw]
		if (kw == guarded.When || kw == guarded.Then) && n != 1 || (kw == guarded.Provided || kw == guarded.Otherwise) && n > 1 {
			wrong = append(wrong, fmt.Sprintf("%s (%d)", kw, n))
		}
	}
	if len(wrong) > 0 {
		l.report(Error, "keyword-count", l.at(r, r.fenceLine), "keywords appear too often or not at all: %s", strings.Join(wrong, ", "))
	}
	if (counts[guarded.Provided] > 0) != (counts[guarded.Otherwise] > 0) {
		l.report(Error, "guard-incomplete", l.at(r, r.fenceLine), "a Provided needs its Otherwise, and an Otherwise a Provided")
	}

	// Outcomes.
	for _, s := range b.Steps {
		switch {
		case s.Head.Keyword == guarded.Otherwise && s.Head.IsHandOver():
			for _, a := range s.Ands {
				l.report(Error, "hand-over-and", l.at(r, a.N), "an And line after a hand-over")
			}
			if t := l.ids[s.Head.Target]; len(t) > 0 && t[0].typ == tRule {
				r.next = t[0]
			} else {
				l.report(Error, "hand-over-unresolved", l.at(r, s.Head.N), "the hand-over target %q is not the identifier of a rule", s.Head.Target)
			}
		case s.Head.Keyword == guarded.Then || s.Head.Keyword == guarded.Otherwise:
			if len(s.Ands) == 0 {
				l.report(Error, "and-missing", l.at(r, s.Head.N), "the %s step has no And line", s.Head.Keyword)
			}
		}
	}

	// Placeholders. The checks that take the block's one When as their
	// reference are left out when there is not exactly one.
	oneWhen := counts[guarded.When] == 1
	event := map[string]bool{}
	for _, s := range b.StepsOf(guarded.When) {
		for p := range s.Placeholders() {
			event[p] = true
		}
	}
	introduced := maps.Clone(event)
	for _, s := range b.StepsOf(guarded.Given) {
		for _, p := range slices.Sorted(maps.Keys(s.Placeholders())) {
			if oneWhen && event[p] {
				l.report(Error, "given-refers-to-event", l.at(r, s.Head.N), "the Given uses <%s>, a placeholder of the When", p)
				break
			}
		}
		for p := range s.Placeholders() {
			introduced[p] = true
		}
	}
	for _, s := range b.StepsOf(guarded.Provided) {
		uses := false
		for p := range s.Placeholders() {
			uses = uses || event[p]
		}
		if oneWhen && !uses {
			l.report(Ask, "provided-without-event", l.at(r, s.Head.N), "the guard uses no placeholder of the When: is it state for a Given?")
		}
	}
	for _, ln := range b.Lines {
		for _, sp := range ln.Spans {
			pos := l.at(r, ln.N)
			switch sp.Kind {
			case guarded.Placeholder:
				if slices.Contains(reserved, sp.Value) {
					l.report(Error, "placeholder-reserved", pos, "the placeholder <%s> is named like a reserved column", sp.Value).col = sp.Start
				}
				if oneWhen && !introduced[sp.Value] {
					l.report(Error, "placeholder-unbound", pos, "the placeholder <%s> is introduced by neither the When nor a Given", sp.Value).col = sp.Start
				}
			case guarded.Mark:
				e := l.resolveTerm(r.cap, sp.Value)
				if e == nil {
					l.report(Error, "term-unresolved", pos, "the term [%s] names no visible glossary entry", sp.Value).col = sp.Start
				} else if st, _ := e.field("Status"); st.value != "defined" {
					f := l.report(Ask, "undecided", pos, "the term [%s] resolves to %q, which is not defined", sp.Value, e.id)
					f.col, f.Related = sp.Start, []Pos{e.pos()}
				}
			case guarded.Code:
				if !l.resolveReference(r.cap, sp.Value) {
					l.report(Error, "reference-unresolved", pos, "`%s` names no visible contract element or parameter", sp.Value).col = sp.Start
				}
			}
		}
		l.checkUnmarked(r, ln)
		for i, w := range strings.Fields(ln.Without(false, guarded.Literal, guarded.Code)) {
			if w == "?" {
				l.report(Ask, "undecided", l.at(r, ln.N), "the line holds an undecided \"?\"").col = i
			}
		}
	}
}

var reserved = []string{"provided", "outcome", "why"}

func (l *linter) checkOrder(r *rule) {
	b := r.block
	maxRank := 0
	for _, ln := range b.Lines {
		if ln.Keyword == "" {
			continue
		}
		if ln.Keyword == guarded.And {
			if maxRank == 0 {
				l.report(Error, "keyword-order", l.at(r, ln.N), "an And comes before any other keyword")
				return
			}
			continue
		}
		rank := guarded.Rank[ln.Keyword]
		if rank < maxRank {
			l.report(Error, "keyword-order", l.at(r, ln.N), "%s is out of the order Given, When, Provided, Then, Otherwise", ln.Keyword)
			return
		}
		maxRank = rank
	}
}

// resolveReference reports whether a backtick reference names a visible
// contract element or parameter: its first segment a top-level element or
// parameter, each further segment a member of the one before.
func (l *linter) resolveReference(cap, ref string) bool {
	segs := strings.Split(ref, ".")
	for _, lv := range visible(cap) {
		cands := l.names[lv][segs[0]]
		if len(cands) == 0 {
			continue
		}
		e := cands[0]
		for _, s := range segs[1:] {
			m := e.members[s]
			if len(m) == 0 {
				return false
			}
			e = m[0]
		}
		return true
	}
	return false
}

// checkUnmarked reports each name or alias of a visible glossary entry that
// the unmarked text of a line uses as a whole word, ignoring case. Where
// names overlap, the longest match at a place wins.
func (l *linter) checkUnmarked(r *rule, ln *guarded.Line) {
	text := ln.Without(true, guarded.Literal, guarded.Code, guarded.Mark, guarded.Placeholder)
	if ln.Keyword != "" {
		i := strings.Index(text, ln.Keyword) // the first word, after leading white space
		text = text[:i] + string(rune(0)) + text[i+len(ln.Keyword):]
	}
	words := splitWords(text)
	var phrases [][]string
	for _, lv := range visible(r.cap) {
		for _, e := range l.glossary[lv] {
			for _, n := range termNames(e) {
				if p := splitWords(n); len(p) > 0 && wordsAndSpaces(n) {
					phrases = append(phrases, lowerAll(p))
				}
			}
		}
	}
	for i := 0; i < len(words); {
		best := 0
		for _, p := range phrases {
			if len(p) > best && matchAt(words, i, p, text) {
				best = len(p)
			}
		}
		if best == 0 {
			i++
			continue
		}
		used := text[words[i].start:words[i+best-1].end]
		l.report(Error, "term-unmarked", l.at(r, ln.N), "%q is a glossary term used without its mark", used).col = words[i].start
		i += best
	}
}

type word struct {
	start, end int
	lower      string
}

func isWordRune(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '-' || c == '_'
}

func splitWords(s string) []word {
	var out []word
	start := -1
	for i, c := range s {
		if isWordRune(c) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			out = append(out, word{start, i, strings.ToLower(s[start:i])})
			start = -1
		}
	}
	if start >= 0 {
		out = append(out, word{start, len(s), strings.ToLower(s[start:])})
	}
	return out
}

// wordsAndSpaces reports whether a name consists only of words and white
// space, so that it can be matched word by word.
func wordsAndSpaces(s string) bool {
	return strings.IndexFunc(s, func(c rune) bool { return !isWordRune(c) && !unicode.IsSpace(c) }) < 0
}

func lowerAll(ws []word) []string {
	var out []string
	for _, w := range ws {
		out = append(out, w.lower)
	}
	return out
}

var spaceOnly = regexp.MustCompile(`^\s+$`)

func matchAt(words []word, i int, phrase []string, text string) bool {
	if i+len(phrase) > len(words) {
		return false
	}
	for k, p := range phrase {
		if words[i+k].lower != p {
			return false
		}
		if k > 0 && !spaceOnly.MatchString(text[words[i+k-1].end:words[i+k].start]) {
			return false
		}
	}
	return true
}

// checkCycles reports each chain of hand-overs that returns to a rule it has
// passed, once per cycle, at the first rule that appears twice on the chain
// from the first rule of the cycle in the tree's order.
func (l *linter) checkCycles(rules []*rule) {
	rules = slices.Clone(rules)
	slices.SortStableFunc(rules, func(a, b *rule) int {
		return cmp.Or(cmp.Compare(a.file.path, b.file.path), cmp.Compare(a.line, b.line))
	})
	byElement := map[*element]*rule{}
	for _, r := range rules {
		byElement[r.element] = r
	}
	reported := map[*element]bool{}
	for _, start := range rules {
		var chain []*element
		cur := start
		for cur != nil {
			if i := slices.Index(chain, cur.element); i >= 0 {
				cycle := chain[i:]
				if !reported[cur.element] {
					var rel []Pos
					var ids []string
					for _, e := range cycle {
						reported[e] = true
						ids = append(ids, e.id)
						if e != cur.element {
							rel = append(rel, e.pos())
						}
					}
					ids = append(ids, cur.id)
					l.report(Error, "hand-over-cycle", cur.pos(), "the hand-overs return to %q: %s", cur.id, strings.Join(ids, " → ")).Related = rel
				}
				break
			}
			chain = append(chain, cur.element)
			if cur.next == nil {
				break
			}
			cur = byElement[cur.next]
		}
	}
}

func (l *linter) checkRepresents(e *element) {
	for _, f := range e.fields {
		if f.key != "Represents" {
			continue
		}
		ln := guarded.Parse([]guarded.SourceLine{{N: f.line, Text: f.value}})
		for _, gl := range ln.Lines {
			for _, sp := range gl.Spans {
				if sp.Kind == guarded.Mark && l.resolveTerm(e.cap, sp.Value) == nil {
					l.report(Error, "term-unresolved", Pos{e.file.path, f.line}, "the term [%s] names no visible glossary entry", sp.Value).col = sp.Start
				}
			}
		}
	}
}
