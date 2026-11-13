// Command proviso is the tooling for a spec tree written in the Proviso
// Model: `proviso init`, `proviso update`, `proviso lint`,
// `proviso manifest`, `proviso diff`, `proviso impact`, `proviso pack`,
// `proviso judge`, `proviso verify`, `proviso rehome`, `proviso show`,
// `proviso version` and `proviso help`.
package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
	"golang.org/x/text/unicode/norm"
)

// The exit statuses of the contracts system.cli.lint.exit-status,
// system.cli.init.exit-status, system.cli.update.exit-status,
// system.cli.manifest.exit-status, system.cli.diff.exit-status,
// system.cli.impact.exit-status, system.cli.pack.exit-status,
// system.cli.judge.exit-status, system.cli.verify.exit-status,
// system.cli.rehome.exit-status, system.cli.show.exit-status,
// system.cli.version.exit-status and system.cli.help.exit-status.
const (
	exitClean   = 0
	exitErrors  = 1 // lint found an error; diff listed an element; judge refused the check result; verify found a problem; a rename would take an identifier; show's query names no element
	exitCantRun = 2
)

// version is the Proviso version of the binary, set at build time with
// -ldflags "-X main.version=..."; "dev" for a build given none.
var version = "dev"

// stdin is what judge reads its check_result from, and rehome its
// mention_list.
var stdin io.Reader = os.Stdin

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	// cli.help.start: without a command, the page goes to stderr.
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, helpPage)
		return exitCantRun
	}
	// cli.help.command: commands are matched exactly.
	name, opts := args[0], args[1:]
	cmd := name
	if cmd == "--help" || cmd == "-h" {
		cmd = "help"
	}
	if _, ok := commandOptions[cmd]; !ok {
		fmt.Fprintf(stderr, "proviso: %q is no command of Proviso; %s\n", name, helpHint)
		return exitCantRun
	}
	// cli.help.argument: the arguments are read before anything else.
	o, ok := readArgs(cmd, name, opts, stderr)
	if !ok {
		return exitCantRun
	}
	switch cmd {
	case "help":
		return helpRun(stdout)
	case "init":
		return initRun(stdout, stderr)
	case "update":
		return updateRun(stdout, stderr)
	case "version":
		return versionRun(stdout, stderr)
	case "show":
		// query.show.query: without a query, show has nothing to look for.
		if o.query == "" {
			fmt.Fprintf(stderr, "proviso show: needs a query, the identifier or name of the elements to show; %s\n", helpHint)
			return exitCantRun
		}
	}
	// lint.run.model checks the first line of manifest.sum for every command
	// lint precedes, with --since and --element too.
	t, res, ok := check(cmd, stderr)
	if !ok {
		return exitCantRun
	}
	switch cmd {
	case "lint":
		return lintRun(t, res, o.json, stdout, stderr)
	case "show":
		// query.show.tree: whatever lint finds, show runs.
		return showRun(t, res, o, stdout, stderr)
	case "manifest", "diff", "impact", "pack", "judge", "verify", "rehome":
		// baseline.run.lint for manifest, diff and pack, coherence.run.lint
		// for impact, judge and verify, upkeep.rehome.lint for rehome: only
		// where lint finds no error, and without writing what lint would
		// write.
		if lint.HasError(res.Findings) {
			_ = report.Text(stderr, res.Findings)
			return exitCantRun
		}
		switch cmd {
		case "manifest":
			return manifestRun(t, res, stderr)
		case "diff":
			return diffRun(t, res, o.json, o.since, stdout, stderr)
		case "pack":
			if o.element != "" {
				return packElementRun(t, res, o.json, o.element, stdout, stderr)
			}
			return packRun(t, res, o.json, o.since, stdout, stderr)
		case "judge":
			return judgeRun(t, res, stderr)
		case "verify":
			return verifyRun(t, res, o.json, stdout, stderr)
		case "rehome":
			return rehomeRun(t, res, o.json, o.write, o.mentions, o.renames, stdout, stderr)
		}
		return impactRun(t, res, o.json, stdout, stderr)
	}
	return exitCantRun
}

// check finds the spec root as lint does, reads its tree completely,
// checks that it is of this binary's model version, and lints it. It
// writes to stderr why it could not.
func check(cmd string, stderr io.Writer) (*tree.Tree, *lint.Result, bool) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "proviso %s: %v\n", cmd, err)
		return nil, nil, false
	}
	root, err := tree.FindRoot(wd)
	if err != nil {
		fmt.Fprintf(stderr, "proviso %s: no spec root found: no %s/system.md in %s or any directory above it\n", cmd, tree.RootName, wd)
		return nil, nil, false
	}
	t := tree.Load(root)
	if len(t.Unreadable) > 0 {
		fmt.Fprintf(stderr, "proviso %s: cannot read the whole spec tree at %s:\n", cmd, root)
		for _, rel := range t.Unreadable {
			fmt.Fprintf(stderr, "  %s\n", path.Join(tree.RootName, rel))
		}
		return nil, nil, false
	}
	if data, ok := t.Manifest(); ok {
		if err := baseline.CheckModel(data); err != nil {
			fmt.Fprintf(stderr, "proviso %s: %v\n", cmd, err)
			return nil, nil, false
		}
	}
	return t, lint.Check(t), true
}

func lintRun(t *tree.Tree, res *lint.Result, asJSON bool, stdout, stderr io.Writer) int {
	var err error
	if asJSON {
		err = report.JSON(stdout, reportRoot(t.Root), res.Findings)
	} else {
		err = report.Text(stdout, res.Findings)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso lint: %v\n", err)
		return exitCantRun
	}
	if lint.HasError(res.Findings) {
		return exitErrors
	}
	return exitClean
}

// reportRoot writes the absolute path of the spec root with "/" as its
// separator, every name in NFC and, on Windows, its drive letter in
// uppercase, whichever case the working directory was given in.
func reportRoot(root string) string {
	vol := filepath.VolumeName(root)
	if len(vol) == 2 && vol[1] == ':' {
		root = strings.ToUpper(vol) + root[2:]
	}
	return norm.NFC.String(filepath.ToSlash(root))
}
