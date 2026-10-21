// Package lint checks a spec tree against the Proviso Model's notation and
// returns its findings.
package lint

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/bbak/proviso/internal/markdown"
	"github.com/bbak/proviso/internal/tree"
)

type linter struct {
	findings []Finding
	files    []*file
	elements []*element

	ids      map[string][]*element
	glossary map[string][]*element            // glossary entries by level
	names    map[string]map[string][]*element // top-level contract elements and parameters, by level and name
	refsIn   map[string]map[string]bool       // reference identifiers by the file that holds them
}

// Run checks a tree read completely and returns its findings, ordered.
func Run(t *tree.Tree) []Finding { return Check(t).Findings }

// Check checks a tree read completely and returns its findings, ordered,
// with the elements it read.
func Check(t *tree.Tree) *Result {
	l := &linter{
		ids:      map[string][]*element{},
		glossary: map[string][]*element{},
		names:    map[string]map[string][]*element{},
		refsIn:   map[string]map[string]bool{},
	}
	for _, rel := range t.Unknown {
		l.report(Error, "file-unknown", Pos{reportPath(rel), 1}, "the layout does not list this file")
	}
	for _, c := range t.CapFolders {
		if !c.HasCapability {
			l.report(Error, "capability-file-missing", Pos{reportPath(c.Rel), 1}, "the capability folder holds no capability.md")
		}
	}
	for _, rel := range t.Malformed {
		l.report(Error, "name-malformed", Pos{reportPath(rel), 1}, "the name holds something other than lowercase letters, marks, digits, hyphens and dots, or ends with a dot, or another entry in its folder has the same name in NFC")
	}
	for _, tf := range t.Files {
		if tf.Kind == tree.Manifest {
			continue // lint reads only its first line, before the run: lint.run.model
		}
		if i := bytes.IndexByte(tf.Data, '\r'); i >= 0 {
			line := bytes.Count(tf.Data[:i], []byte("\n")) + 1
			l.report(Error, "line-ending", Pos{reportPath(tf.Rel), line}, "the line ends with a carriage return; the tree ends lines with a line feed only")
		}
		if tf.Kind == tree.Settings {
			continue
		}
		f := &file{File: tf, path: reportPath(tf.Rel), doc: markdown.Parse(tf.Data)}
		l.files = append(l.files, f)
		l.build(f)
	}
	l.index()
	for _, e := range l.elements {
		l.checkIdentifier(e)
		l.checkSections(e)
		l.checkFields(e)
		if e.typ == tGlossary {
			l.checkEntry(e)
		}
		if e.typ == tContract {
			l.checkRepresents(e)
		}
	}
	l.checkAmbiguity()
	l.checkRules()
	Sort(l.findings)
	if l.findings == nil {
		l.findings = []Finding{}
	}
	return &Result{Findings: l.findings, Elements: l.export()}
}

func reportPath(rel string) string { return path.Join(tree.RootName, rel) }

func (l *linter) report(sev Severity, kind string, p Pos, format string, args ...any) *Finding {
	l.findings = append(l.findings, Finding{Kind: kind, Severity: sev, File: p.File, Line: p.Line, Message: fmt.Sprintf(format, args...)})
	return &l.findings[len(l.findings)-1]
}

// index builds the lookups that references resolve against.
func (l *linter) index() {
	for _, e := range l.elements {
		l.ids[e.id] = append(l.ids[e.id], e)
		switch {
		case e.typ == tGlossary:
			l.glossary[e.cap] = append(l.glossary[e.cap], e)
		case e.typ == tParameter || (e.typ == tContract && e.parent == nil):
			if l.names[e.cap] == nil {
				l.names[e.cap] = map[string][]*element{}
			}
			l.names[e.cap][e.name] = append(l.names[e.cap][e.name], e)
		case e.typ == tReference:
			if l.refsIn[e.file.Rel] == nil {
				l.refsIn[e.file.Rel] = map[string]bool{}
			}
			l.refsIn[e.file.Rel][e.id] = true
		}
	}
}

// visible returns the levels whose glossaries, contracts and parameters an
// element at the level cap sees, its own first.
func visible(cap string) []string {
	if cap == "" {
		return []string{""}
	}
	return []string{cap, ""}
}

var wellFormed = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)*$`)

func (l *linter) checkIdentifier(e *element) {
	if !wellFormed.MatchString(e.id) {
		l.report(Error, "identifier-malformed", e.pos(), "the identifier %q is not well-formed", e.id)
	}
	if others := l.ids[e.id]; len(others) > 1 {
		var rel []Pos
		var where []string
		for _, o := range others {
			if o != e {
				rel = append(rel, o.pos())
				where = append(where, fmt.Sprintf("%s:%d", o.file.path, o.line))
			}
		}
		l.report(Error, "identifier-duplicate", e.pos(), "the identifier %q is also used at %s", e.id, strings.Join(where, ", ")).Related = rel
	}
	if e.typ == tSystem && e.id != "system" {
		l.report(Error, "system-identifier", e.pos(), "the system is identified as %q, not \"system\"", e.id)
	}
}

func (l *linter) checkSections(e *element) {
	names, ok := sectionNames[e.typ]
	if !ok {
		return
	}
	seen := map[string]bool{}
	for _, s := range e.sections {
		if seen[s.name] {
			l.report(Error, "section-repeated", Pos{e.file.path, s.line}, "the section %q appears again in %q", s.name, e.id)
		}
		seen[s.name] = true
		if s.firstWord == "?" {
			l.report(Ask, "undecided", Pos{e.file.path, s.line}, "the section %q of %q is undecided", s.name, e.id)
		}
	}
	var missing []string
	for _, r := range names.required {
		if !seen[r] {
			missing = append(missing, r)
		}
	}
	if len(missing) > 0 {
		l.report(Error, "section-missing", e.pos(), "the %s %q lacks the sections: %s", e.typ, e.id, strings.Join(missing, ", "))
	}
}

func (l *linter) checkFields(e *element) {
	names := fieldNames[e.typ] // empty for a type without fields
	typed := e.typ != ""
	seen := map[string]bool{}
	for _, f := range e.fields {
		p := Pos{e.file.path, f.line}
		if typed && !slices.Contains(names.all, f.key) {
			l.report(Error, "field-unknown", p, "a %s has no field %q", e.typ, f.key)
		}
		if seen[f.key] {
			l.report(Error, "field-repeated", p, "the field %q appears again in %q", f.key, e.id)
		}
		seen[f.key] = true
		if firstWord(f.value) == "?" {
			l.report(Ask, "undecided", p, "the field %q of %q is undecided", f.key, e.id)
		}
	}
	var missing []string
	for _, r := range names.required {
		if f, ok := e.field(r); !ok || f.value == "" {
			missing = append(missing, r)
		}
	}
	if len(missing) > 0 {
		l.report(Error, "field-missing", e.pos(), "the %s %q lacks a value for the fields: %s", e.typ, e.id, strings.Join(missing, ", "))
	}
}

func firstWord(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

func (l *linter) checkEntry(e *element) {
	if st, ok := e.field("Status"); ok {
		def, hasDef := e.field("Definition")
		undecidedDef := hasDef && firstWord(def.value) == "?"
		q, hasQ := e.field("Question")
		switch st.value {
		case "open":
			if !hasQ || q.value == "" || !undecidedDef {
				l.report(Error, "open-entry-incomplete", e.pos(), "the open entry %q needs a Question and the Definition \"?\"", e.id)
			}
		case "defined":
			if hasQ || undecidedDef {
				l.report(Error, "defined-entry-open", e.pos(), "the defined entry %q has a Question or an undecided Definition", e.id)
			}
		default:
			l.report(Error, "status-unknown", Pos{e.file.path, st.line}, "the status %q is neither \"defined\" nor \"open\"", st.value)
		}
	}
	_, hasDef := e.field("Definition")
	by, hasBy := e.field("Defined by")
	if hasDef == hasBy {
		l.report(Error, "definition-count", e.pos(), "the entry %q needs exactly one of Definition and Defined by", e.id)
	}
	if hasBy {
		e.definedBy = l.definedByResolves(e, by)
	}
	if hasBy && e.definedBy == "" {
		l.report(Error, "defined-by-unresolved", Pos{e.file.path, by.line}, "the Defined by of %q has no link to a visible reference", e.id)
	}
	if _, ok := e.field("Deviations"); ok && !hasBy {
		l.report(Error, "deviations-without-reference", e.pos(), "the entry %q has Deviations but no Defined by", e.id)
	}
}

// definedByResolves returns the identifier of the visible reference a
// Defined by links to, or "" if it links to none.
func (l *linter) definedByResolves(e *element, by field) string {
	dir := path.Dir(e.file.Rel)
	allowed := []string{"references.md"}
	if e.cap != "" {
		allowed = append(allowed, path.Join("capabilities", e.cap, "references.md"))
	}
	for _, dest := range by.links {
		p, frag, ok := strings.Cut(dest, "#")
		if !ok || p == "" || strings.Contains(p, ":") || strings.HasPrefix(p, "/") {
			continue
		}
		target := path.Join(dir, p)
		if slices.Contains(allowed, target) && l.refsIn[target][frag] {
			return frag
		}
	}
	return ""
}

// checkAmbiguity reports term names and contract or parameter names that
// name more than one visible element.
func (l *linter) checkAmbiguity() {
	for lv, entries := range l.glossary {
		for _, e := range entries {
			for _, n := range uniqueNorm(termNames(e)) {
				var others []*element
				for _, olv := range neighbours(lv, l.glossaryLevels()) {
					for _, o := range l.glossary[olv] {
						if o != e && slices.Contains(uniqueNorm(termNames(o)), n) {
							others = append(others, o)
						}
					}
				}
				l.reportAmbiguous("term-ambiguous", e, n, others)
			}
		}
	}
	for lv, byName := range l.names {
		for name, els := range byName {
			for _, e := range els {
				var others []*element
				for _, olv := range neighbours(lv, l.nameLevels()) {
					for _, o := range l.names[olv][name] {
						if o != e {
							others = append(others, o)
						}
					}
				}
				l.reportAmbiguous("reference-ambiguous", e, name, others)
			}
		}
	}
}

func (l *linter) reportAmbiguous(kind string, e *element, name string, others []*element) {
	if len(others) == 0 {
		return
	}
	var rel []Pos
	var ids []string
	for _, o := range others {
		rel = append(rel, o.pos())
		ids = append(ids, fmt.Sprintf("%q (%s:%d)", o.id, o.file.path, o.line))
	}
	l.report(Error, kind, e.pos(), "the name %q of %q also names %s", name, e.id, strings.Join(ids, ", ")).Related = rel
}

// neighbours returns a level together with the levels above and below it.
func neighbours(lv string, all []string) []string {
	if lv != "" {
		return []string{lv, ""}
	}
	return all
}

func (l *linter) glossaryLevels() []string {
	var out []string
	for lv := range l.glossary {
		out = append(out, lv)
	}
	slices.Sort(out)
	return out
}

func (l *linter) nameLevels() []string {
	var out []string
	for lv := range l.names {
		out = append(out, lv)
	}
	slices.Sort(out)
	return out
}

// termNames returns an entry's name and aliases.
func termNames(e *element) []string {
	names := []string{e.name}
	if a, ok := e.field("Aliases"); ok {
		for _, s := range strings.Split(a.value, ",") {
			if s = strings.TrimSpace(s); s != "" {
				names = append(names, s)
			}
		}
	}
	return names
}

// norm folds case and white space, as term names are matched.
func norm(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

func uniqueNorm(names []string) []string {
	var out []string
	for _, n := range names {
		if n := norm(n); !slices.Contains(out, n) {
			out = append(out, n)
		}
	}
	return out
}

// resolveTerm returns the visible entry a term name names, if any.
func (l *linter) resolveTerm(cap, name string) *element {
	n := norm(name)
	for _, lv := range visible(cap) {
		for _, e := range l.glossary[lv] {
			if slices.Contains(uniqueNorm(termNames(e)), n) {
				return e
			}
		}
	}
	return nil
}
