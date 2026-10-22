package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
)

type jsonReason struct {
	Source string   `json:"source"`
	Via    string   `json:"via"`
	From   []string `json:"from"`
}

type jsonCandidate struct {
	ID      string       `json:"id"`
	File    string       `json:"file"`
	Line    int          `json:"line"`
	Reasons []jsonReason `json:"reasons"`
}

type jsonScope struct {
	ID         string   `json:"id"`
	Candidates []string `json:"candidates"`
}

type jsonImpact struct {
	Root       string          `json:"root"`
	Revision   *string         `json:"revision"`
	Candidates []jsonCandidate `json:"candidates"`
	Scopes     []jsonScope     `json:"scopes"`
}

// ImpactJSON writes the impact_set: the spec root that was compared, the
// revision the former text came from, or null, every candidate with its
// reasons, and the candidates found through scope, by what framed them.
func ImpactJSON(w io.Writer, root, revision string, candidates []baseline.Candidate, frames []baseline.Frame) error {
	d := jsonImpact{Root: root, Candidates: []jsonCandidate{}, Scopes: []jsonScope{}}
	if revision != "" {
		d.Revision = &revision
	}
	for _, c := range candidates {
		jc := jsonCandidate{ID: c.Element.ID, File: c.Element.File, Line: c.Element.Line, Reasons: []jsonReason{}}
		for _, r := range c.Reasons {
			jc.Reasons = append(jc.Reasons, jsonReason(r))
		}
		d.Candidates = append(d.Candidates, jc)
	}
	for _, f := range frames {
		d.Scopes = append(d.Scopes, jsonScope(f))
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

// ImpactText writes the candidates for people: the revision the former
// text came from, then each candidate with where it stands and its
// reasons, then the candidates found through scope, by what framed them.
// Nothing is written when there is no candidate.
func ImpactText(w io.Writer, revision string, candidates []baseline.Candidate, frames []baseline.Frame) error {
	if len(candidates) == 0 {
		return nil
	}
	var sb strings.Builder
	if revision != "" {
		fmt.Fprintf(&sb, "former text from %s\n\n", revision)
	}
	for _, c := range candidates {
		fmt.Fprintf(&sb, "%s (%s:%d)\n", c.Element.ID, c.Element.File, c.Element.Line)
		for _, r := range c.Reasons {
			fmt.Fprintf(&sb, "  %s %s, from %s\n", r.Source, r.Via, strings.Join(r.From, ", "))
		}
	}
	for _, f := range frames {
		fmt.Fprintf(&sb, "\nframed by %s: %d candidates\n", f.ID, len(f.Candidates))
		for _, id := range f.Candidates {
			fmt.Fprintf(&sb, "  %s\n", id)
		}
	}
	_, err := io.WriteString(w, sb.String())
	return err
}
