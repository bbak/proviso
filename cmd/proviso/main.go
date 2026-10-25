// Command proviso is the tooling for a spec tree written in the Proviso
// Model: `proviso init`, `proviso lint`, `proviso manifest`,
// `proviso diff`, `proviso impact`, `proviso pack`, `proviso judge` and
// `proviso verify`.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/git"
	"github.com/bbak/proviso/internal/kit"
	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
	"golang.org/x/text/unicode/norm"
)

// The exit statuses of the contracts lint.cli.lint.exit-status,
// system.cli.init.exit-status, system.cli.manifest.exit-status,
// system.cli.diff.exit-status, system.cli.impact.exit-status,
// system.cli.pack.exit-status, system.cli.judge.exit-status and
// system.cli.verify.exit-status.
const (
	exitClean   = 0
	exitErrors  = 1 // lint found an error; diff listed an element; judge refused the check result; verify found a problem
	exitCantRun = 2
)

const usage = "usage: proviso init | proviso lint [--json] | proviso manifest | proviso diff [--json] [--since <revision>] | proviso impact [--json] | proviso pack [--json] [--since <revision>] | proviso judge | proviso verify [--json]"

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

// initRun writes every installed file the working directory is missing,
// provided it is the root of a git repository. It never opens an entry
// that exists, and it creates each file exclusively, so that it cannot
// replace one that appears meanwhile.
func initRun(stdout, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "proviso init: %v\n", err)
		return exitCantRun
	}
	// Any entry named .git counts: a folder in a clone, a file in a
	// worktree, whatever its attributes.
	if _, err := os.Lstat(filepath.Join(wd, ".git")); err != nil {
		fmt.Fprintf(stderr, "proviso init: %s is not the root of a git repository\n", wd)
		return exitCantRun
	}
	status := exitClean
	for _, f := range kit.Files(norm.NFC.String(filepath.Base(wd))) {
		dst := filepath.Join(wd, filepath.FromSlash(f.Path))
		if _, err := os.Lstat(dst); err == nil {
			fmt.Fprintf(stdout, "skipped %s\n", f.Path)
			continue
		}
		if err := install(dst, f.Data); err != nil {
			fmt.Fprintf(stderr, "proviso init: cannot write %s: %v\n", f.Path, err)
			status = exitCantRun
			continue
		}
		fmt.Fprintf(stdout, "written %s\n", f.Path)
	}
	return status
}

// install creates the file dst, with the folders above it that are
// missing, and removes what it created of the file if it cannot write it
// completely, so that a later run writes it again.
func install(dst string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(dst)
	}
	return err
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

// manifestRun writes manifest.sum into the spec root, so that a run that
// fails leaves the former manifest as it was, and then removes the
// coherence file and the amendment note, since the tree they served is the
// baseline now.
func manifestRun(t *tree.Tree, res *lint.Result, stderr io.Writer) int {
	if err := replace(t.Root, tree.ManifestName, baseline.Write(res.Elements)); err != nil {
		fmt.Fprintf(stderr, "proviso manifest: cannot write %s: %v\n", path.Join(tree.RootName, tree.ManifestName), err)
		return exitCantRun
	}
	// Anything named as the note is the note, a folder too.
	for _, rm := range []struct {
		name   string
		remove func(string) error
	}{{tree.CoherenceName, os.Remove}, {tree.NoteName, os.RemoveAll}} {
		if err := rm.remove(filepath.Join(t.Root, rm.name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(stderr, "proviso manifest: %s is written, but %s cannot be removed: %v\n", path.Join(tree.RootName, tree.ManifestName), path.Join(tree.RootName, rm.name), err)
			return exitCantRun
		}
	}
	return exitClean
}

// replace writes the file name into the folder root. It writes a new file
// beside the former one and renames it over it, so that a write that
// fails leaves the former file as it was.
func replace(root, name string, data []byte) error {
	dst := filepath.Join(root, name)
	if _, err := os.Lstat(dst); err == nil {
		// Renaming over a read-only file succeeds on Unix and fails on
		// Windows; asking for write access first refuses it everywhere.
		f, err := os.OpenFile(dst, os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		f.Close()
	}
	tmp, err := os.CreateTemp(root, "."+name+".*")
	if err != nil {
		return err
	}
	_, err = tmp.Write(data)
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
	}
	return err
}

// elementDiff derives the element diff with its former texts, against the
// baseline or, if since names a revision, against the tree there. It
// returns the revision the former texts came from and the elements of the
// tree there, nil where none was needed. It writes to stderr why it could
// not.
func elementDiff(cmd string, t *tree.Tree, res *lint.Result, since string, stderr io.Writer) ([]*baseline.Entry, string, []lint.Element, bool) {
	if since != "" {
		return sinceDiff(cmd, t, res, since, stderr)
	}
	base := baseline.Baseline{}
	data, hasManifest := t.Manifest()
	if hasManifest {
		var bad []string
		base, bad = baseline.Parse(data)
		if len(bad) > 0 {
			fmt.Fprintf(stderr, "proviso %s: %s cannot be read:\n", cmd, path.Join(tree.RootName, tree.ManifestName))
			for _, b := range bad {
				fmt.Fprintf(stderr, "  %s\n", b)
			}
			return nil, "", nil, false
		}
	}
	entries := baseline.Compare(res.Elements, base)
	if !baseline.NeedsFormer(entries) {
		return entries, "", nil, true
	}
	revision, formerElements, err := former(t, entries, base, data)
	if err != nil {
		fmt.Fprintf(stderr, "proviso %s: cannot read the former text: %v\n", cmd, err)
		return nil, "", nil, false
	}
	return entries, revision, formerElements, true
}

// sinceDiff derives the element diff against the spec tree at the revision
// since names, fingerprinted as a manifest would record it, provided lint
// accepts that tree or there is none.
func sinceDiff(cmd string, t *tree.Tree, res *lint.Result, since string, stderr io.Writer) ([]*baseline.Entry, string, []lint.Element, bool) {
	fail := func(format string, args ...any) ([]*baseline.Entry, string, []lint.Element, bool) {
		fmt.Fprintf(stderr, "proviso %s: cannot compare with %s: %s\n", cmd, since, fmt.Sprintf(format, args...))
		return nil, "", nil, false
	}
	rev, err := git.Commit(t.Root, since)
	if err != nil {
		return fail("%v", err)
	}
	files, err := git.Files(t.Root, rev)
	if err != nil {
		return fail("%v", err)
	}
	formerTree := tree.FromFiles(t.Root, files)
	var formerElements []lint.Element
	// Without system.md, there is no spec tree at the revision, and the
	// tree compared with is empty.
	if slices.ContainsFunc(formerTree.Files, func(f *tree.File) bool { return f.Kind == tree.System }) {
		if len(formerTree.Unreadable) > 0 {
			return fail("the spec tree there cannot be read completely: %s", strings.Join(formerTree.Unreadable, ", "))
		}
		formerRes := lint.Check(formerTree)
		if lint.HasError(formerRes.Findings) {
			fail("lint finds an error in the spec tree there:")
			report.Text(stderr, formerRes.Findings)
			return nil, "", nil, false
		}
		formerElements = formerRes.Elements
	}
	base := baseline.Record(formerElements)
	entries := baseline.Compare(res.Elements, base)
	if !baseline.NeedsFormer(entries) {
		return entries, "", nil, true
	}
	if err := baseline.Former(entries, base, formerElements, rev); err != nil {
		return fail("%v", err)
	}
	return entries, rev, formerElements, true
}

func diffRun(t *tree.Tree, res *lint.Result, asJSON bool, since string, stdout, stderr io.Writer) int {
	entries, revision, _, ok := elementDiff("diff", t, res, since, stderr)
	if !ok {
		return exitCantRun
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

func impactRun(t *tree.Tree, res *lint.Result, asJSON bool, stdout, stderr io.Writer) int {
	entries, revision, formerElements, ok := elementDiff("impact", t, res, "", stderr)
	if !ok {
		return exitCantRun
	}
	candidates, frames := baseline.Impact(entries, res.Elements, formerElements)
	var buf bytes.Buffer
	var err error
	if asJSON {
		err = report.ImpactJSON(&buf, reportRoot(t.Root), revision, candidates, frames)
	} else {
		err = report.ImpactText(&buf, revision, candidates, frames)
	}
	if err == nil {
		_, err = buf.WriteTo(stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso impact: %v\n", err)
		return exitCantRun
	}
	return exitClean
}

func packRun(t *tree.Tree, res *lint.Result, asJSON bool, since string, stdout, stderr io.Writer) int {
	entries, revision, _, ok := elementDiff("pack", t, res, since, stderr)
	if !ok {
		return exitCantRun
	}
	note, ok := readNote(t, stderr)
	if !ok {
		return exitCantRun
	}
	packed := baseline.Pack(entries, res.Elements)
	var buf bytes.Buffer
	var err error
	if asJSON {
		err = report.PackJSON(&buf, reportRoot(t.Root), revision, entries, note, packed)
	} else {
		err = report.PackText(&buf, revision, entries, note, packed)
	}
	if err == nil {
		_, err = buf.WriteTo(stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso pack: %v\n", err)
		return exitCantRun
	}
	return exitClean
}

// readNote reads amendment_note, following baseline.pack.note and
// baseline.pack.note-unreadable: its text as the file holds it, or ""
// where the spec root holds none or one that holds only whitespace. It
// writes to stderr why it could not read it, as a regular file in UTF-8.
func readNote(t *tree.Tree, stderr io.Writer) (string, bool) {
	p := filepath.Join(t.Root, tree.NoteName)
	info, err := os.Lstat(p)
	if errors.Is(err, fs.ErrNotExist) {
		return "", true
	}
	var data []byte
	if err == nil && !info.Mode().IsRegular() {
		err = errors.New("it is not a regular file")
	}
	if err == nil {
		data, err = os.ReadFile(p)
	}
	if err == nil && !utf8.Valid(data) {
		err = errors.New("it is not UTF-8")
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso pack: %s cannot be read: %v\n", path.Join(tree.RootName, tree.NoteName), err)
		return "", false
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", true
	}
	return string(data), true
}

// hasNote reports whether the spec root holds an amendment_note: an entry
// of its name, whatever it is and holds, even one that cannot be looked
// at.
func hasNote(t *tree.Tree) bool {
	_, err := os.Lstat(filepath.Join(t.Root, tree.NoteName))
	return !errors.Is(err, fs.ErrNotExist)
}

// coherence derives what judge and verify start from: the element diff
// against the baseline, the baseline and the candidates. It writes to
// stderr why it could not.
func coherence(cmd string, t *tree.Tree, res *lint.Result, stderr io.Writer) ([]*baseline.Entry, baseline.Baseline, []baseline.Candidate, bool) {
	entries, _, formerElements, ok := elementDiff(cmd, t, res, "", stderr)
	if !ok {
		return nil, nil, nil, false
	}
	base := baseline.Baseline{}
	if data, ok := t.Manifest(); ok {
		base, _ = baseline.Parse(data) // elementDiff refused it if it was damaged
	}
	candidates, _ := baseline.Impact(entries, res.Elements, formerElements)
	return entries, base, candidates, true
}

// readCoherence reads coherence_file, if the spec root holds one,
// following baseline.judge.unreadable. It writes to stderr why it could
// not.
func readCoherence(cmd string, t *tree.Tree, stderr io.Writer) (*baseline.Coherence, bool) {
	name := path.Join(tree.RootName, tree.CoherenceName)
	p := filepath.Join(t.Root, tree.CoherenceName)
	info, err := os.Lstat(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, true
	}
	var data []byte
	if err == nil && !info.Mode().IsRegular() {
		err = errors.New("it is not a regular file")
	}
	if err == nil {
		data, err = os.ReadFile(p)
	}
	var c *baseline.Coherence
	if err == nil {
		c, err = baseline.ParseCoherence(data)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso %s: %s cannot be read: %v\n", cmd, name, err)
		return nil, false
	}
	return c, true
}

// judgeRun records the check_result on stdin in coherence_file.
func judgeRun(t *tree.Tree, res *lint.Result, stderr io.Writer) int {
	data, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "proviso judge: cannot read standard input: %v\n", err)
		return exitCantRun
	}
	in, err := baseline.ParseCheckResult(data)
	if err != nil {
		fmt.Fprintf(stderr, "proviso judge: the input is no check_result: %v\n", err)
		return exitErrors
	}
	entries, base, candidates, ok := coherence("judge", t, res, stderr)
	if !ok {
		return exitCantRun
	}
	if len(entries) == 0 {
		fmt.Fprintln(stderr, "proviso judge: there is no amendment to judge: the spec tree is as the manifest records it")
		return exitErrors
	}
	before, ok := readCoherence("judge", t, stderr)
	if !ok {
		return exitCantRun
	}
	after, err := baseline.Judge(before, in, entries, candidates, res.Elements, base)
	if err != nil {
		fmt.Fprintf(stderr, "proviso judge: %v; nothing is recorded\n", err)
		return exitErrors
	}
	if err := replace(t.Root, tree.CoherenceName, after.Encode()); err != nil {
		fmt.Fprintf(stderr, "proviso judge: cannot write %s: %v\n", path.Join(tree.RootName, tree.CoherenceName), err)
		return exitCantRun
	}
	return exitClean
}

// verifyRun writes the problems it finds comparing coherence_file with the
// spec tree.
func verifyRun(t *tree.Tree, res *lint.Result, asJSON bool, stdout, stderr io.Writer) int {
	entries, base, candidates, ok := coherence("verify", t, res, stderr)
	if !ok {
		return exitCantRun
	}
	c, ok := readCoherence("verify", t, stderr)
	if !ok {
		return exitCantRun
	}
	problems := baseline.Verify(c, hasNote(t), entries, candidates, base)
	var buf bytes.Buffer
	var err error
	if asJSON {
		err = report.VerifyJSON(&buf, reportRoot(t.Root), problems)
	} else {
		err = report.VerifyText(&buf, problems)
	}
	if err == nil {
		_, err = buf.WriteTo(stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso verify: %v\n", err)
		return exitCantRun
	}
	if len(problems) > 0 {
		return exitErrors
	}
	return exitClean
}

// former fills in the former texts from the revision that wrote the
// manifest, and returns that revision and the elements of the tree there.
func former(t *tree.Tree, entries []*baseline.Entry, base baseline.Baseline, manifest []byte) (string, []lint.Element, error) {
	rev, err := git.ManifestRevision(t.Root, tree.ManifestName, manifest)
	if err != nil {
		return "", nil, err
	}
	files, err := git.Files(t.Root, rev)
	if err != nil {
		return "", nil, err
	}
	formerTree := lint.Check(tree.FromFiles(t.Root, files))
	return rev, formerTree.Elements, baseline.Former(entries, base, formerTree.Elements, rev)
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
