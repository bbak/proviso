package baseline

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bbak/proviso/internal/lint"
)

// Coherence is what coherence_file records: every touched element with the
// fingerprints of its parts now and in the baseline, the judgements with
// the fingerprints of the candidate they were made on, and the findings.
// Its form is not part of the contract; only judge writes it.
type Coherence struct {
	Touched    []Covered `json:"touched"`
	Judgements []Judged  `json:"judgements"`
	Findings   []Finding `json:"findings"`
}

// Covered is a touched element as the coherence check covered it.
type Covered struct {
	ID       string            `json:"id"`
	Now      map[string]string `json:"now"`
	Baseline map[string]string `json:"baseline"`
}

// Judged is a judgement recorded for a candidate.
type Judged struct {
	ID     string            `json:"id"`
	Reason string            `json:"reason"`
	Parts  map[string]string `json:"parts"`
}

// The statuses of a recorded finding.
const (
	Open      = "open"
	Resolved  = "resolved"
	Dismissed = "dismissed"
)

// Finding is a finding recorded by the coherence check.
type Finding struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Elements []string `json:"elements"`
	Text     string   `json:"text"`
	Status   string   `json:"status"`
	Reason   string   `json:"reason,omitempty"`
}

var (
	findingID    = regexp.MustCompile(`^f[1-9][0-9]*$`)
	findingKinds = []string{"contradiction", "ignored", "gap", "missed"}
)

// findingNumber returns the number of a finding's id, "f" and a number.
func findingNumber(id string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(id, "f"))
	return n
}

// Encode returns coherence_file as judge writes it.
func (c *Coherence) Encode() []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(c) // a Coherence always encodes
	return b.Bytes()
}

// ParseCoherence reads coherence_file. Anything judge would not have
// written, such as a merge conflict left in it, is refused: it is JSON in
// UTF-8 whose members are those judge writes, in its order, with the
// values it writes. Whitespace is not compared, so line endings do not
// matter.
func ParseCoherence(data []byte) (*Coherence, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("it is not UTF-8")
	}
	var c Coherence
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("it is not what judge writes: %v", err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		return nil, fmt.Errorf("it is not what judge writes: %v", err)
	}
	var again bytes.Buffer
	if err := json.Compact(&again, c.Encode()); err != nil || !bytes.Equal(compact.Bytes(), again.Bytes()) {
		return nil, errors.New("it is not what judge writes")
	}
	if err := c.valid(); err != nil {
		return nil, fmt.Errorf("it is not what judge writes: %v", err)
	}
	return &c, nil
}

// valid checks what the JSON form alone does not: no list or map is null,
// every fingerprint is one, and every finding has an id, a kind and a
// status judge gives.
func (c *Coherence) valid() error {
	if c.Touched == nil || c.Judgements == nil || c.Findings == nil {
		return errors.New("a list is missing")
	}
	fps := func(m map[string]string) bool {
		if m == nil {
			return false
		}
		for k, v := range m {
			if !partKey.MatchString(k) || !fpValue.MatchString(v) {
				return false
			}
		}
		return true
	}
	for _, t := range c.Touched {
		if t.ID == "" || !fps(t.Now) || !fps(t.Baseline) {
			return fmt.Errorf("the touched element %q is not recorded as judge records it", t.ID)
		}
	}
	for _, j := range c.Judgements {
		if j.ID == "" || j.Reason == "" || !fps(j.Parts) {
			return fmt.Errorf("the judgement for %q is not recorded as judge records it", j.ID)
		}
	}
	seen := map[string]bool{}
	for _, f := range c.Findings {
		if !findingID.MatchString(f.ID) || seen[f.ID] || !slices.Contains(findingKinds, f.Kind) || f.Elements == nil || f.Text == "" ||
			!slices.Contains([]string{Open, Resolved, Dismissed}, f.Status) || f.Status == Dismissed && f.Reason == "" {
			return fmt.Errorf("the finding %q is not recorded as judge records it", f.ID)
		}
		seen[f.ID] = true
	}
	return nil
}

// CheckResult is a check_result as judge reads it.
type CheckResult struct {
	Judgements []InputJudgement
	Findings   []InputFinding
}

// InputJudgement is a judgement of a check_result.
type InputJudgement struct {
	Reason   string
	Elements []string
}

// InputFinding is a finding of a check_result: a new one, without an id,
// or one that changes the status of a recorded finding.
type InputFinding struct {
	ID, Status, Reason string
	Kind, Text         string
	Elements           []string
}

// ParseCheckResult reads a check_result, or says why the input is none.
func ParseCheckResult(data []byte) (*CheckResult, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("it is not UTF-8")
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("it is not JSON: %v", err)
	}
	top, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("it is not a JSON object")
	}
	var r CheckResult
	judgements, err := list(top, "judgements", "the member judgements")
	if err != nil {
		return nil, err
	}
	judgedIn := map[string]int{}
	for i, j := range judgements {
		at := fmt.Sprintf("judgement %d", i+1)
		m, ok := j.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s is not an object", at)
		}
		reason, err := text(m, "reason", at, true)
		if err != nil {
			return nil, err
		}
		elements, err := names(m, "elements", at, true)
		if err != nil {
			return nil, err
		}
		if len(elements) == 0 {
			return nil, fmt.Errorf("%s names no element", at)
		}
		var unique []string
		for _, id := range elements {
			if n, ok := judgedIn[id]; ok && n != i {
				return nil, fmt.Errorf("%s is named in judgement %d and in judgement %d: which reason holds would be a guess", id, n+1, i+1)
			}
			if _, ok := judgedIn[id]; !ok {
				unique = append(unique, id)
			}
			judgedIn[id] = i
		}
		r.Judgements = append(r.Judgements, InputJudgement{Reason: reason, Elements: unique})
	}
	findings, err := list(top, "findings", "the member findings")
	if err != nil {
		return nil, err
	}
	for i, f := range findings {
		at := fmt.Sprintf("finding %d", i+1)
		m, ok := f.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s is not an object", at)
		}
		fi, err := inputFinding(m, at)
		if err != nil {
			return nil, err
		}
		r.Findings = append(r.Findings, fi)
	}
	return &r, nil
}

// inputFinding reads one finding of a check_result: with an id, it changes
// a recorded finding; without one, it is new.
func inputFinding(m map[string]any, at string) (InputFinding, error) {
	_, hasID := m["id"]
	_, hasKind := m["kind"]
	if hasID && hasKind {
		return InputFinding{}, fmt.Errorf("%s has both a kind and an id: it is new or it changes one, not both", at)
	}
	allowed := []string{"kind", "elements", "text"}
	if hasID {
		allowed = []string{"id", "status", "reason"}
	}
	for _, k := range slices.Sorted(maps.Keys(m)) {
		if !slices.Contains(allowed, k) {
			return InputFinding{}, fmt.Errorf("%s has the member %s, which a finding with%s an id does not have", at, k, map[bool]string{true: "", false: "out"}[hasID])
		}
	}
	var f InputFinding
	var err error
	if hasID {
		if f.ID, err = text(m, "id", at, false); err != nil {
			return f, err
		}
		if f.Status, err = text(m, "status", at, true); err != nil {
			return f, err
		}
		if f.Status != Resolved && f.Status != Dismissed {
			return f, fmt.Errorf("%s has the status %q, which is neither %q nor %q", at, f.Status, Resolved, Dismissed)
		}
		if _, ok := m["reason"]; ok {
			if f.Reason, err = text(m, "reason", at, true); err != nil {
				return f, err
			}
		}
		if f.Status == Dismissed && f.Reason == "" {
			return f, fmt.Errorf("%s dismisses %s without a reason", at, f.ID)
		}
		return f, nil
	}
	if f.Kind, err = text(m, "kind", at, true); err != nil {
		return f, err
	}
	if !slices.Contains(findingKinds, f.Kind) {
		return f, fmt.Errorf("%s has the kind %q, which is none of %s", at, f.Kind, strings.Join(findingKinds, ", "))
	}
	if f.Text, err = text(m, "text", at, true); err != nil {
		return f, err
	}
	if f.Elements, err = names(m, "elements", at, false); err != nil {
		return f, err
	}
	return f, nil
}

// list returns the array member key of m, empty if it is left out.
func list(m map[string]any, key, at string) ([]any, error) {
	v, ok := m[key]
	if !ok {
		return nil, nil
	}
	a, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%s is not an array", at)
	}
	return a, nil
}

// text returns the string member key of m, which must be there, and must
// not be empty if nonEmpty.
func text(m map[string]any, key, at string, nonEmpty bool) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("%s has no %s", at, key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("the %s of %s is not a string", key, at)
	}
	if nonEmpty && s == "" {
		return "", fmt.Errorf("the %s of %s is empty", key, at)
	}
	return s, nil
}

// names returns the member key of m, an array of strings, which must be
// there if required.
func names(m map[string]any, key, at string, required bool) ([]string, error) {
	v, ok := m[key]
	if !ok {
		if required {
			return nil, fmt.Errorf("%s has no %s", at, key)
		}
		return []string{}, nil
	}
	a, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("the %s of %s is not an array", key, at)
	}
	out := []string{}
	for _, e := range a {
		s, ok := e.(string)
		if !ok {
			return nil, fmt.Errorf("the %s of %s holds something other than a string", key, at)
		}
		out = append(out, s)
	}
	return out, nil
}

// fingerprints returns the fingerprint of every part of an element, by
// part; none for an element that is not there.
func fingerprints(e *lint.Element) map[string]string {
	m := map[string]string{}
	if e != nil {
		for _, p := range Parts(*e) {
			m[p.Key] = p.FP
		}
	}
	return m
}

// Judge takes the judgements and findings of a check result one by one
// into what coherence_file recorded before, nil if there is none, and
// records every touched element of the element diff now and in the
// baseline. It returns why it refuses the check result, if it does; the
// caller then records nothing.
func Judge(before *Coherence, in *CheckResult, entries []*Entry, candidates []Candidate, now []lint.Element, base Baseline) (*Coherence, error) {
	if before == nil {
		before = &Coherence{}
	}
	isCandidate := map[string]*lint.Element{}
	for _, c := range candidates {
		isCandidate[c.Element.ID] = c.Element
	}
	judged := map[string]Judged{}
	for _, j := range before.Judgements {
		judged[j.ID] = j
	}
	for _, j := range in.Judgements {
		for _, id := range j.Elements {
			e := isCandidate[id]
			if e == nil {
				return nil, fmt.Errorf("%s is no candidate of the amendment", id)
			}
			judged[id] = Judged{ID: id, Reason: j.Reason, Parts: fingerprints(e)}
		}
	}
	nowByID := byID(now)
	findings := slices.Clone(before.Findings)
	next := 1
	for _, f := range findings {
		next = max(next, findingNumber(f.ID)+1)
	}
	for _, f := range in.Findings {
		if f.ID == "" {
			for _, id := range f.Elements {
				if _, ok := base[id]; nowByID[id] == nil && !ok {
					return nil, fmt.Errorf("a new finding names %s, which neither the spec tree nor the baseline holds", id)
				}
			}
			findings = append(findings, Finding{ID: "f" + strconv.Itoa(next), Kind: f.Kind, Elements: f.Elements, Text: f.Text, Status: Open})
			next++
			continue
		}
		i := slices.IndexFunc(findings, func(r Finding) bool { return r.ID == f.ID })
		if i < 0 {
			return nil, fmt.Errorf("no finding has the id %s", f.ID)
		}
		findings[i].Status, findings[i].Reason = f.Status, f.Reason
	}
	c := &Coherence{Touched: []Covered{}, Judgements: []Judged{}, Findings: findings}
	for _, e := range entries {
		c.Touched = append(c.Touched, Covered{ID: e.ID, Now: fingerprints(e.Element), Baseline: recorded(base, e.ID)})
	}
	for _, id := range slices.Sorted(maps.Keys(judged)) {
		if isCandidate[id] != nil {
			c.Judgements = append(c.Judgements, judged[id])
		}
	}
	if c.Findings == nil {
		c.Findings = []Finding{}
	}
	return c, nil
}

// recorded returns what the baseline records for an element, a copy, and
// none for an element it does not hold.
func recorded(base Baseline, id string) map[string]string {
	m := maps.Clone(base[id])
	if m == nil {
		m = map[string]string{}
	}
	return m
}

// The kinds of problem verify reports, in the order verification lists
// them.
const (
	Missing  = "missing"
	Leftover = "leftover"
	Unjudged = "unjudged"
	Stale    = "stale"
	OpenKind = "open"
)

var problemOrder = []string{Missing, Leftover, Unjudged, Stale, OpenKind}

// Problem is what verify found wrong: its kind, the element an unjudged
// or stale problem is about, and the id and text of an open finding.
type Problem struct {
	Kind, Element, Finding, Text string
}

// Verify compares what coherence_file records, nil if the spec root holds
// none, with the element diff and the candidates, and returns the
// problems, ordered.
func Verify(c *Coherence, entries []*Entry, candidates []Candidate, base Baseline) []Problem {
	switch {
	case c == nil && len(entries) == 0:
		return nil
	case c == nil:
		return []Problem{{Kind: Missing}}
	case len(entries) == 0:
		return []Problem{{Kind: Leftover}}
	}
	var out []Problem
	covered := map[string]Covered{}
	for _, t := range c.Touched {
		covered[t.ID] = t
	}
	for _, e := range entries {
		t, ok := covered[e.ID]
		if !ok || !maps.Equal(t.Now, fingerprints(e.Element)) || !maps.Equal(t.Baseline, recorded(base, e.ID)) {
			out = append(out, Problem{Kind: Stale, Element: e.ID})
		}
	}
	judged := map[string]Judged{}
	for _, j := range c.Judgements {
		judged[j.ID] = j
	}
	for _, cand := range candidates {
		j, ok := judged[cand.Element.ID]
		switch {
		case !ok:
			out = append(out, Problem{Kind: Unjudged, Element: cand.Element.ID})
		case !maps.Equal(j.Parts, fingerprints(cand.Element)):
			out = append(out, Problem{Kind: Stale, Element: cand.Element.ID})
		}
	}
	for _, f := range c.Findings {
		if f.Status == Open {
			out = append(out, Problem{Kind: OpenKind, Finding: f.ID, Text: f.Text})
		}
	}
	slices.SortFunc(out, func(a, b Problem) int {
		return cmp.Or(
			cmp.Compare(slices.Index(problemOrder, a.Kind), slices.Index(problemOrder, b.Kind)),
			cmp.Compare(a.Element, b.Element),
			cmp.Compare(findingNumber(a.Finding), findingNumber(b.Finding)))
	})
	return out
}
