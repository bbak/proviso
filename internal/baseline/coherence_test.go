package baseline

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/lint"
)

func TestParseCheckResult(t *testing.T) {
	cases := map[string]string{ // the input, and "" or a part of why it is refused
		`{}`: "",
		`{"judgements": [{"reason": "r", "elements": ["a", "b", "a"]}], "findings": [{"kind": "gap", "text": "t"}]}`: "",
		`{"findings": [{"id": "f3", "status": "dismissed", "reason": "r"}, {"id": "f1", "status": "resolved"}]}`:     "",
		`{"judgements": [{"reason": "r", "elements": ["a"]}], "findings": [{"id": "f1", "status": "resolved"}]}`:     "recorded apart",
		`{"findings": [{"id": "f1", "status": "resolved"}, {"kind": "gap", "text": "t"}]}`:                           "recorded apart",
		`{"findings": [{"id": "f1", "status": "resolved"}], "other": 1}`:                                             "",
		`not json`:             "not JSON",
		`[]`:                   "not a JSON object",
		`{"judgements": null}`: "not an array",
		`{"judgements": [{"reason": "", "elements": ["a"]}]}`:                                      "empty",
		`{"judgements": [{"reason": "r", "elements": []}]}`:                                        "names no element",
		`{"judgements": [{"reason": "r", "elements": ["a"]}, {"reason": "s", "elements": ["a"]}]}`: "would be a guess",
		`{"findings": [{"kind": "gap", "text": "t", "id": "f1"}]}`:                                 "not both",
		`{"findings": [{"kind": "typo", "text": "t"}]}`:                                            "none of",
		`{"findings": [{"kind": "gap", "text": "t", "status": "open"}]}`:                           "member status",
		`{"findings": [{"id": "f1", "status": "dismissed"}]}`:                                      "without a reason",
		`{"findings": [{"id": "f1", "status": "open"}]}`:                                           "neither",
		"{\"findings\": [{\"kind\": \"gap\", \"text\": \"\xff\"}]}":                                "not UTF-8",
	}
	for in, want := range cases {
		r, err := ParseCheckResult([]byte(in))
		switch {
		case want == "" && err != nil:
			t.Errorf("%s: refused: %v", in, err)
		case want != "" && (err == nil || !strings.Contains(err.Error(), want)):
			t.Errorf("%s: %v, want %q", in, err, want)
		case want == "" && strings.Contains(in, `"a", "b", "a"`) && fmt.Sprint(r.Judgements[0].Elements) != "[a b]":
			t.Errorf("an element named twice in one judgement: %v", r.Judgements)
		}
	}
}

// amendment changes the Then step of c.f.r in impactTree, so that c.g.a,
// on the same event, and the invariants above it are candidates.
func amendment(t *testing.T, edit func(map[string]string)) ([]*Entry, []Candidate, []lint.Element, Baseline) {
	t.Helper()
	before := files(impactTree)
	base, _ := Parse(Write(before))
	now := files(editTree(func(m map[string]string) {
		replace(m, "capabilities/c/features/f.md", "Then      it is read", "Then      it is read twice")
		if edit != nil {
			edit(m)
		}
	}))
	entries := Compare(now, base)
	cands, _ := Impact(entries, now, before)
	return entries, cands, now, base
}

func ids(cands []Candidate) []string {
	var out []string
	for _, c := range cands {
		out = append(out, c.Element.ID)
	}
	return out
}

func TestJudgeAndVerify(t *testing.T) {
	entries, cands, now, base := amendment(t, nil)
	if fmt.Sprint(ids(cands)) != "[c.f.inv c.g.a c.g.inv c.inv system.inv]" {
		t.Fatalf("candidates %v", ids(cands))
	}
	if got := Verify(nil, false, entries, cands, base); fmt.Sprint(got) != fmt.Sprint([]Problem{{Kind: Missing}}) {
		t.Errorf("without a file: %v", got)
	}
	if got := Verify(nil, false, nil, nil, base); got != nil {
		t.Errorf("an unchanged tree without a file: %v", got)
	}
	in, _ := ParseCheckResult([]byte(`{"judgements": [{"reason": "untouched", "elements": ["c.g.a", "c.inv"]}], "findings": [{"kind": "gap", "elements": ["c.f.r"], "text": "a gap"}]}`))
	c, err := Judge(nil, in, entries, cands, now, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Touched) != len(entries) || len(c.Judgements) != 2 || c.Findings[0].ID != "f1" || c.Findings[0].Status != Open {
		t.Fatalf("%+v", c)
	}
	parsed, err := ParseCoherence(c.Encode())
	if err != nil {
		t.Fatal(err)
	}
	got := Verify(parsed, false, entries, cands, base)
	want := []Problem{{Kind: Unjudged, Element: "c.f.inv"}, {Kind: Unjudged, Element: "c.g.inv"}, {Kind: Unjudged, Element: "system.inv"}, {Kind: OpenKind, Finding: "f1", Text: "a gap"}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
	if got := Verify(parsed, false, nil, nil, base); fmt.Sprint(got) != fmt.Sprint([]Problem{{Kind: Leftover}}) {
		t.Errorf("leftover: %v", got)
	}
	if got := Verify(nil, true, nil, nil, base); fmt.Sprint(got) != fmt.Sprint([]Problem{{Kind: Leftover}}) {
		t.Errorf("leftover note: %v", got)
	}
	if got := Verify(parsed, true, nil, nil, base); fmt.Sprint(got) != fmt.Sprint([]Problem{{Kind: Leftover}}) {
		t.Errorf("leftover, both: %v", got)
	}
	if got := Verify(nil, true, entries, cands, base); fmt.Sprint(got) != fmt.Sprint([]Problem{{Kind: Missing}}) {
		t.Errorf("a note without a check: %v", got)
	}

	in, _ = ParseCheckResult([]byte(`{"judgements": [{"reason": "also untouched", "elements": ["c.f.inv", "c.g.inv", "system.inv", "c.g.a"]}], "findings": [{"kind": "missed", "text": "x"}]}`))
	c, err = Judge(parsed, in, entries, cands, now, base)
	if err != nil {
		t.Fatal(err)
	}
	in, _ = ParseCheckResult([]byte(`{"findings": [{"id": "f1", "status": "dismissed", "reason": "r"}]}`))
	if c, err = Judge(c, in, entries, cands, now, base); err != nil {
		t.Fatal(err)
	}
	if got := Verify(c, false, entries, cands, base); len(got) != 1 || got[0].Finding != "f2" {
		t.Errorf("after the second run: %v", got)
	}
	if c.Judgements[1].ID != "c.g.a" || c.Judgements[1].Reason != "also untouched" {
		t.Errorf("a former judgement is replaced: %+v", c.Judgements)
	}

	// c.g.a is edited after the check: it is touched now, and its
	// judgement is stale; c.inv's text changes: its judgement is stale.
	entries2, cands2, _, _ := amendment(t, func(m map[string]string) {
		replace(m, "capabilities/c/features/g.md", "it is read", "it is read again")
		replace(m, "capabilities/c/invariants.md", "\np\n", "\nq\n")
	})
	got = Verify(c, false, entries2, cands2, base)
	var stale []string
	for _, p := range got {
		if p.Kind == Stale {
			stale = append(stale, p.Element)
		}
	}
	if fmt.Sprint(stale) != "[c.g.a c.inv]" {
		t.Errorf("stale %v in %v", stale, got)
	}
}

func TestJudgeRefuses(t *testing.T) {
	entries, cands, now, base := amendment(t, nil)
	for in, want := range map[string]string{
		`{"judgements": [{"reason": "r", "elements": ["c.f.r"]}]}`:               "c.f.r is no candidate",
		`{"judgements": [{"reason": "r", "elements": ["c.nosuch"]}]}`:            "c.nosuch is no candidate",
		`{"findings": [{"id": "f9", "status": "resolved"}]}`:                     "no finding has the id f9",
		`{"findings": [{"kind": "gap", "elements": ["c.nosuch"], "text": "t"}]}`: "c.nosuch, which neither",
	} {
		r, err := ParseCheckResult([]byte(in))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Judge(nil, r, entries, cands, now, base); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: %v, want %q", in, err, want)
		}
	}
}

func TestParseCoherence(t *testing.T) {
	good := (&Coherence{Touched: []Covered{}, Judgements: []Judged{}, Findings: []Finding{}}).Encode()
	if _, err := ParseCoherence(good); err != nil {
		t.Errorf("as judge writes it: %v", err)
	}
	if _, err := ParseCoherence([]byte(strings.ReplaceAll(string(good), "\n", "\r\n"))); err != nil {
		t.Errorf("with other line endings: %v", err)
	}
	for _, bad := range []string{"", "<<<<<<< HEAD\n" + string(good), `{"touched": [], "judgements": [], "findings": [], "x": 1}`, `{"Touched": [], "judgements": [], "findings": []}`, `{"touched": null, "judgements": [], "findings": []}`} {
		if _, err := ParseCoherence([]byte(bad)); err == nil {
			t.Errorf("%q: read", bad)
		}
	}
}

func TestProblemOrder(t *testing.T) {
	c := &Coherence{Touched: []Covered{}, Judgements: []Judged{}, Findings: []Finding{
		{ID: "f10", Kind: "gap", Elements: []string{}, Text: "a", Status: Open},
		{ID: "f2", Kind: "gap", Elements: []string{}, Text: "b", Status: Open},
	}}
	entries, cands, _, base := amendment(t, nil)
	got := Verify(c, false, entries, cands[:1], base)
	if got[0].Kind != Unjudged || got[1].Kind != Stale || got[len(got)-2].Finding != "f2" || got[len(got)-1].Finding != "f10" {
		t.Errorf("%v", got)
	}
}

func TestJudgeClosingKeepsTouched(t *testing.T) {
	entries, cands, now, base := amendment(t, nil)
	in, _ := ParseCheckResult([]byte(`{"findings": [{"kind": "gap", "elements": ["c.f.r"], "text": "a gap"}]}`))
	first, err := Judge(nil, in, entries, cands, now, base)
	if err != nil {
		t.Fatal(err)
	}

	// c.f.r is edited after the check, and f1 is resolved by that edit:
	// the run closes a finding, so c.f.r keeps what was recorded, and c.g.a,
	// touched now and recorded for none, stays so.
	entries2, cands2, now2, _ := amendment(t, func(m map[string]string) {
		replace(m, "capabilities/c/features/f.md", "Then      it is read twice", "Then      it is read three times")
		replace(m, "capabilities/c/features/g.md", "it is read", "it is read again")
	})
	in, _ = ParseCheckResult([]byte(`{"findings": [{"id": "f1", "status": "resolved", "reason": "edited"}]}`))
	c, err := Judge(first, in, entries2, cands2, now2, base)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(c.Touched) != fmt.Sprint(first.Touched) {
		t.Errorf("closing a finding recorded the touched elements anew:\n%v\nwant %v", c.Touched, first.Touched)
	}
	var stale []string
	for _, p := range Verify(c, false, entries2, cands2, base) {
		if p.Kind == Stale {
			stale = append(stale, p.Element)
		}
	}
	if fmt.Sprint(stale) != "[c.f.r c.g.a]" {
		t.Errorf("stale %v", stale)
	}

	// A run that closes no finding records every touched element as it is now.
	in, _ = ParseCheckResult([]byte(`{}`))
	c, err = Judge(c, in, entries2, cands2, now2, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Touched) != len(entries2) {
		t.Errorf("touched %v", c.Touched)
	}
	for _, p := range Verify(c, false, entries2, cands2, base) {
		if p.Kind == Stale {
			t.Errorf("stale after a check: %v", p)
		}
	}
}
