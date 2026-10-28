// Command proviso is the tooling for a spec tree written in the Proviso
// Model: `proviso init`, `proviso lint`, `proviso manifest`,
// `proviso diff`, `proviso impact`, `proviso pack`, `proviso judge`,
// `proviso verify` and `proviso version`.
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

// The exit statuses of the contracts lint.cli.lint.exit-status,
// system.cli.init.exit-status, system.cli.manifest.exit-status,
// system.cli.diff.exit-status, system.cli.impact.exit-status,
// system.cli.pack.exit-status, system.cli.judge.exit-status,
// system.cli.verify.exit-status and system.cli.version.exit-status.
const (
	exitClean   = 0
	exitErrors  = 1 // lint found an error; diff listed an element; judge refused the check result; verify found a problem
	exitCantRun = 2
)

const usage = "usage: proviso init | proviso lint [--json] | proviso manifest | proviso diff [--json] [--since <revision>] | proviso impact [--json] | proviso pack [--json] [--since <revision>] | proviso judge | proviso verify [--json] | proviso version"

// version is the Proviso version of the binary, set at build time with
// -ldflags "-X main.version=..."; "dev" for a build given none.
var version = "dev"

// stdin is what judge reads its check_result from.
var stdin io.Reader = os.Stdin

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return exitCantRun
	}
	cmd, opts := args[0], args[1:]
	switch cmd {
	case "init":
		if len(opts) > 0 {
			fmt.Fprintf(stderr, "proviso init: unknown argument %q\n%s\n", opts[0], usage)
			return exitCantRun
		}
		return initRun(stdout, stderr)
	case "version":
		if len(opts) > 0 {
			fmt.Fprintf(stderr, "proviso version: takes no argument\n%s\n", usage)
			return exitCantRun
		}
		return versionRun(stdout, stderr)
	case "lint", "manifest", "diff", "impact", "pack", "judge", "verify":
	default:
		fmt.Fprintln(stderr, usage)
		return exitCantRun
	}
	asJSON, since := false, ""
	for i := 0; i < len(opts); i++ {
		switch a := opts[i]; {
		case a == "--json" && cmd != "manifest" && cmd != "judge":
			asJSON = true
		case a == "--since" && (cmd == "diff" || cmd == "pack") && since == "":
			if i+1 == len(opts) || opts[i+1] == "" {
				fmt.Fprintf(stderr, "proviso %s: --since needs a revision\n%s\n", cmd, usage)
				return exitCantRun
			}
			i++
			since = opts[i]
		default:
			fmt.Fprintf(stderr, "proviso %s: unknown argument %q\n%s\n", cmd, a, usage)
			return exitCantRun
		}
	}
	// baseline.diff.since: since a revision, manifest.sum is not read.
	t, res, ok := check(cmd, since == "", stderr)
	if !ok {
		return exitCantRun
	}
	switch cmd {
	case "lint":
		return lintRun(t, res, asJSON, stdout, stderr)
	case "manifest", "diff", "impact", "pack", "judge", "verify":
		// baseline.run.lint: only where lint finds no error, and without
		// writing what lint would write.
		if lint.HasError(res.Findings) {
			report.Text(stderr, res.Findings)
			return exitCantRun
		}
		switch cmd {
		case "manifest":
			return manifestRun(t, res, stderr)
		case "diff":
			return diffRun(t, res, asJSON, since, stdout, stderr)
		case "pack":
			return packRun(t, res, asJSON, since, stdout, stderr)
		case "judge":
			return judgeRun(t, res, stderr)
		case "verify":
			return verifyRun(t, res, asJSON, stdout, stderr)
		}
		return impactRun(t, res, asJSON, stdout, stderr)
	}
	return exitCantRun
}

// check finds the spec root as lint does, reads its tree completely,
// checks, if readManifest, that it is of this binary's model version, and
// lints it. It writes to stderr why it could not.
func check(cmd string, readManifest bool, stderr io.Writer) (*tree.Tree, *lint.Result, bool) {
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
	if data, ok := t.Manifest(); ok && readManifest {
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
