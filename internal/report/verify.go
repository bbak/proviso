package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
)

type jsonProblem struct {
	Kind    string `json:"kind"`
	Element string `json:"element,omitempty"`
	Finding string `json:"finding,omitempty"`
	Text    string `json:"text,omitempty"`
	Message string `json:"message"`
}

type jsonVerification struct {
	Root     string        `json:"root"`
	Problems []jsonProblem `json:"problems"`
}

// VerifyJSON writes the verification: the spec root that was verified and
// every problem found.
func VerifyJSON(w io.Writer, root string, problems []baseline.Problem) error {
	v := jsonVerification{Root: root, Problems: []jsonProblem{}}
	for _, p := range problems {
		v.Problems = append(v.Problems, jsonProblem{Kind: p.Kind, Element: p.Element, Finding: p.Finding, Text: p.Text, Message: message(p)})
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// VerifyText writes one line per problem, its kind first. Nothing is
// written when there is no problem.
func VerifyText(w io.Writer, problems []baseline.Problem) error {
	var sb strings.Builder
	for _, p := range problems {
		subject := p.Element + p.Finding
		if subject != "" {
			subject = " " + subject
		}
		fmt.Fprintf(&sb, "%s%s: %s\n", p.Kind, subject, strings.ReplaceAll(message(p), "\n", " "))
	}
	_, err := io.WriteString(w, sb.String())
	return err
}

// message says what a problem means, for people.
func message(p baseline.Problem) string {
	switch p.Kind {
	case baseline.Missing:
		return "the element diff lists elements, but no coherence check is recorded; run the coherence check and record it with judge"
	case baseline.Leftover:
		return "a coherence check is recorded, but the spec tree is as the manifest records it; remove .coherence.json"
	case baseline.Unjudged:
		return p.Element + " is a candidate without a judgement"
	case baseline.Stale:
		return p.Element + " changed since the coherence check covered it"
	}
	return "the finding " + p.Finding + " is neither resolved nor dismissed: " + p.Text
}
