// Command proviso is the tooling for a spec tree written in the Proviso
// Model: `proviso lint`, `proviso manifest` and `proviso diff`.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/git"
	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
	"golang.org/x/text/unicode/norm"
)

// The exit statuses of the contracts lint.cli.lint.exit-status,
// system.cli.manifest.exit-status and system.cli.diff.exit-status.
const (
	exitClean   = 0
	exitErrors  = 1 // lint found an error; diff listed an element
	exitCantRun = 2
)

const usage = "usage: proviso lint [--json] | proviso manifest | proviso diff [--json]"

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
	case "lint", "manifest", "diff":
	default:
		fmt.Fprintln(stderr, usage)
		return exitCantRun
	}
	asJSON := false
	for _, a := range opts {
		if a != "--json" || cmd == "manifest" {
			fmt.Fprintf(stderr, "proviso %s: unknown argument %q\n%s\n", cmd, a, usage)
			return exitCantRun
		}
		asJSON = true
	}
	t, res, ok := check(cmd, stderr)
	if !ok {
		return exitCantRun
	}
	switch cmd {
	case "lint":
		return lintRun(t, res, asJSON, stdout, stderr)
	case "manifest", "diff":
		// baseline.run.lint: only where lint finds no error, and without
		// writing what lint would write.
		if lint.HasError(res.Findings) {
			report.Text(stderr, res.Findings)
			return exitCantRun
		}
		if cmd == "manifest" {
			return manifestRun(t, res, stderr)
		}
		return diffRun(t, res, asJSON, stdout, stderr)
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

// manifestRun writes manifest.sum into the spec root. It writes a new file
// beside the former one and renames it over it, so that a run that fails
// leaves the former manifest as it was.
func manifestRun(t *tree.Tree, res *lint.Result, stderr io.Writer) int {
	fail := func(err error) int {
		fmt.Fprintf(stderr, "proviso manifest: cannot write %s: %v\n", path.Join(tree.RootName, tree.ManifestName), err)
		return exitCantRun
	}
	dst := filepath.Join(t.Root, tree.ManifestName)
	if _, ok := t.Manifest(); ok {
		// Renaming over a read-only file succeeds on Unix and fails on
		// Windows; asking for write access first refuses it everywhere.
		f, err := os.OpenFile(dst, os.O_WRONLY, 0)
		if err != nil {
			return fail(err)
		}
		f.Close()
	}
	tmp, err := os.CreateTemp(t.Root, "."+tree.ManifestName+".*")
	if err != nil {
		return fail(err)
	}
	_, err = tmp.Write(baseline.Write(res.Elements))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), 0o644)
	}
	if err == nil {
		err = os.Rename(tmp.Name(), dst)
	}
	if err != nil {
		os.Remove(tmp.Name())
		return fail(err)
	}
	return exitClean
}

func diffRun(t *tree.Tree, res *lint.Result, asJSON bool, stdout, stderr io.Writer) int {
	base := baseline.Baseline{}
	data, hasManifest := t.Manifest()
	if hasManifest {
		var bad []string
		base, bad = baseline.Parse(data)
		if len(bad) > 0 {
			fmt.Fprintf(stderr, "proviso diff: %s cannot be read:\n", path.Join(tree.RootName, tree.ManifestName))
			for _, b := range bad {
				fmt.Fprintf(stderr, "  %s\n", b)
			}
			return exitCantRun
		}
	}
	entries := baseline.Compare(res.Elements, base)
	revision := ""
	if baseline.NeedsFormer(entries) {
		var err error
		if revision, err = former(t, entries, base, data); err != nil {
			fmt.Fprintf(stderr, "proviso diff: cannot read the former text: %v\n", err)
			return exitCantRun
		}
	}
	var buf bytes.Buffer
	var err error
	if asJSON {
		err = report.DiffJSON(&buf, reportRoot(t.Root), revision, entries)
	} else {
		err = report.DiffText(&buf, revision, entries)
	}
	if err == nil {
		_, err = buf.WriteTo(stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso diff: %v\n", err)
		return exitCantRun
	}
	if len(entries) > 0 {
		return exitErrors
	}
	return exitClean
}

// former fills in the former texts from the revision that wrote the
// manifest, and returns that revision.
func former(t *tree.Tree, entries []*baseline.Entry, base baseline.Baseline, manifest []byte) (string, error) {
	rev, err := git.ManifestRevision(t.Root, tree.ManifestName, manifest)
	if err != nil {
		return "", err
	}
	files, err := git.Files(t.Root, rev)
	if err != nil {
		return "", err
	}
	formerTree := lint.Check(tree.FromFiles(t.Root, files))
	return rev, baseline.Former(entries, base, formerTree.Elements, rev)
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
