package main

import (
	"fmt"
	"io"

	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/query"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
)

// showRun writes every element the query names, among those of the
// capability --in names and of the system level where it names one
// (query.show.in, query.show.match, query.show.output). It writes nothing
// under the spec root (query.read-only).
func showRun(t *tree.Tree, res *lint.Result, o options, stdout, stderr io.Writer) int {
	if o.in != "" && !query.HasCapability(t, o.in) {
		fmt.Fprintf(stderr, "proviso show: the spec tree holds no capability %q\n", o.in)
		return exitCantRun
	}
	found := query.Find(res.Elements, o.query, o.in)
	if len(found) == 0 && !o.json {
		within := ""
		if o.in != "" {
			within = fmt.Sprintf(" within the capability %q", o.in)
		}
		fmt.Fprintf(stderr, "proviso show: %q names no element%s\n", o.query, within)
		return exitErrors
	}
	if !emit("show", stdout, stderr, func(w io.Writer) error {
		if o.json {
			return report.ShowJSON(w, reportRoot(t.Root), found)
		}
		return report.ShowText(w, found)
	}) {
		return exitCantRun
	}
	if len(found) == 0 {
		return exitErrors
	}
	return exitClean
}
