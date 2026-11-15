package acceptance

// Checks that a command wrote what a rule says to stdout and nothing else,
// as baseline.pack.output, coherence.impact.output, query.show.match and
// query.show.output say.

import (
	"encoding/json"
	"strings"
	"testing"
)

// onlyJSON checks that stdout is one JSON value and nothing else: no text
// before it, and nothing after it but white space, which JSON allows around
// a value.
func onlyJSON(t *testing.T, rule, stdout string) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var v json.RawMessage
	if err := dec.Decode(&v); err != nil {
		t.Errorf("%s: stdout does not start with a JSON value: %v\n%s", rule, err, stdout)
		return
	}
	if rest := stdout[dec.InputOffset():]; strings.TrimSpace(rest) != "" {
		t.Errorf("%s: want nothing else on stdout, got after the JSON value: %q", rule, rest)
	}
}

// onlyPackText checks that the stdout of the run without --json is pack_text
// and nothing else: its first line is the first heading of the pack, and it
// ends with the text of the last element of the pack, as context_pack holds
// it, followed by nothing but white space.
func onlyPackText(t *testing.T, o packOutcome) {
	t.Helper()
	text := strings.ReplaceAll(o.text.stdout, "\r\n", "\n")
	if len(o.headings) == 0 {
		t.Errorf("baseline.pack.output: want pack_text on stdout, got %q", o.text.stdout)
		return
	}
	if first, _, _ := strings.Cut(text, "\n"); first != "# "+o.headings[0] {
		t.Errorf("baseline.pack.output: want nothing on stdout before the heading %q, got the first line %q", "# "+o.headings[0], first)
	}
	if !o.cpOK {
		return
	}
	last := o.headings[len(o.headings)-1]
	var end string
	switch {
	case packHeadingJSON[last] != "":
		for _, e := range o.cp.Elements {
			if e.Heading == packHeadingJSON[last] {
				end = e.Text
			}
		}
	case last == notInAmendment && o.cp.Note != nil:
		end = *o.cp.Note
	default:
		// Under Amendment, the end of the last element is its last part's
		// fenced block, which the layout check reads (see parsePackText).
		return
	}
	trim := func(s string) string { return strings.TrimRight(strings.ReplaceAll(s, "\r\n", "\n"), " \t\n") }
	if !strings.HasSuffix(trim(text), "\n"+trim(end)) {
		t.Errorf("baseline.pack.output: want nothing on stdout after the last element of the pack\n%s\ngot the end\n%s", trim(end), tail(trim(text), len(trim(end))+200))
	}
}

// tail gives the last n bytes of s, or s.
func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
