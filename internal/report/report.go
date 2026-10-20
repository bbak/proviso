// Package report writes lint's findings: as text for people, or as the
// JSON report of the contract lint.cli.report.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/lint"
)

// Text writes one line per finding: `file:line: severity kind: message`.
func Text(w io.Writer, findings []lint.Finding) error {
	for _, f := range findings {
		msg := strings.ReplaceAll(f.Message, "\n", " ")
		if _, err := fmt.Fprintf(w, "%s:%d: %s %s: %s\n", f.File, f.Line, f.Severity, f.Kind, msg); err != nil {
			return err
		}
	}
	return nil
}

type jsonReport struct {
	Root     string         `json:"root"`
	Findings []lint.Finding `json:"findings"`
}

// JSON writes the report: the spec root that was linted and every finding.
func JSON(w io.Writer, root string, findings []lint.Finding) error {
	if findings == nil {
		findings = []lint.Finding{}
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonReport{Root: root, Findings: findings})
}
