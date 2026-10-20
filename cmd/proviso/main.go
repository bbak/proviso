// Command proviso is the tooling for a spec tree written in the Proviso
// Model. Its one command so far is `proviso lint`.
package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
)

// The exit statuses of the contract lint.cli.lint.exit-status.
const (
	exitClean   = 0
	exitErrors  = 1
	exitCantRun = 2
)

const usage = "usage: proviso lint [--json]"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "lint" {
		fmt.Fprintln(stderr, usage)
		return exitCantRun
	}
	asJSON := false
	for _, a := range args[1:] {
		if a != "--json" {
			fmt.Fprintf(stderr, "proviso lint: unknown argument %q\n%s\n", a, usage)
			return exitCantRun
		}
		asJSON = true
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "proviso lint: %v\n", err)
		return exitCantRun
	}
	root, err := tree.FindRoot(wd)
	if err != nil {
		fmt.Fprintf(stderr, "proviso lint: no spec root found: no %s/system.md in %s or any directory above it\n", tree.RootName, wd)
		return exitCantRun
	}
	t := tree.Load(root)
	if len(t.Unreadable) > 0 {
		fmt.Fprintf(stderr, "proviso lint: cannot read the whole spec tree at %s:\n", root)
		for _, rel := range t.Unreadable {
			fmt.Fprintf(stderr, "  %s\n", path.Join(tree.RootName, rel))
		}
		return exitCantRun
	}
	findings := lint.Run(t)
	if asJSON {
		err = report.JSON(stdout, reportRoot(root), findings)
	} else {
		err = report.Text(stdout, findings)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso lint: %v\n", err)
		return exitCantRun
	}
	if lint.HasError(findings) {
		return exitErrors
	}
	return exitClean
}

// reportRoot writes the absolute path of the spec root with "/" as its
// separator and, on Windows, its drive letter in uppercase, whichever case
// the working directory was given in.
func reportRoot(root string) string {
	vol := filepath.VolumeName(root)
	if len(vol) == 2 && vol[1] == ':' {
		root = strings.ToUpper(vol) + root[2:]
	}
	return filepath.ToSlash(root)
}
