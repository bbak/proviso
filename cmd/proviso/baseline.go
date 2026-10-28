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
	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
)

// manifestRun writes manifest.sum into the spec root, so that a run that
// fails leaves the former manifest as it was, and then removes the
// coherence file and the amendment note, since the tree they served is the
// baseline now.
func manifestRun(t *tree.Tree, res *lint.Result, stderr io.Writer) int {
	if err := replace(t.Root, tree.ManifestName, baseline.Write(res.Elements)); err != nil {
		fmt.Fprintf(stderr, "proviso manifest: cannot write %s: %v\n", path.Join(tree.RootName, tree.ManifestName), err)
		return exitCantRun
	}
	// baseline.manifest.remove: whatever stands under either name goes, a
	// folder with everything in it, a link without what it points to.
	status := exitClean
	for _, name := range []string{tree.CoherenceName, tree.NoteName} {
		if err := os.RemoveAll(filepath.Join(t.Root, name)); err != nil {
			fmt.Fprintf(stderr, "proviso manifest: %s is written, but %s cannot be removed: %v\n", path.Join(tree.RootName, tree.ManifestName), path.Join(tree.RootName, name), err)
			status = exitCantRun
		}
	}
	return status
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
	if !emit("diff", stdout, stderr, func(w io.Writer) error {
		if asJSON {
			return report.DiffJSON(w, reportRoot(t.Root), revision, entries)
		}
		return report.DiffText(w, revision, entries)
	}) {
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
	if !emit("impact", stdout, stderr, func(w io.Writer) error {
		if asJSON {
			return report.ImpactJSON(w, reportRoot(t.Root), revision, candidates, frames)
		}
		return report.ImpactText(w, revision, candidates, frames)
	}) {
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
	if !emit("pack", stdout, stderr, func(w io.Writer) error {
		if asJSON {
			return report.PackJSON(w, reportRoot(t.Root), revision, entries, note, packed)
		}
		return report.PackText(w, revision, entries, note, packed)
	}) {
		return exitCantRun
	}
	return exitClean
}

// emit writes what write writes to stdout, but only once it has written it
// completely, so that a report that fails is never written in part. It
// writes to stderr why it could not.
func emit(cmd string, stdout, stderr io.Writer, write func(io.Writer) error) bool {
	var buf bytes.Buffer
	err := write(&buf)
	if err == nil {
		_, err = buf.WriteTo(stdout)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso %s: %v\n", cmd, err)
		return false
	}
	return true
}

// readRegular reads the file name in the spec root, provided it is a
// regular file. It reports whether the spec root holds an entry of that
// name at all, and why it could not read one it holds.
func readRegular(t *tree.Tree, name string) ([]byte, bool, error) {
	p := filepath.Join(t.Root, name)
	info, err := os.Lstat(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err == nil && !info.Mode().IsRegular() {
		err = errors.New("it is not a regular file")
	}
	var data []byte
	if err == nil {
		data, err = os.ReadFile(p)
	}
	return data, true, err
}

// readNote reads amendment_note, following baseline.pack.note and
// baseline.pack.note-unreadable: its text as the file holds it, or ""
// where the spec root holds none or one that holds only whitespace. It
// writes to stderr why it could not read it, as a regular file in UTF-8.
func readNote(t *tree.Tree, stderr io.Writer) (string, bool) {
	data, exists, err := readRegular(t, tree.NoteName)
	if !exists {
		return "", true
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
	data, exists, err := readRegular(t, tree.CoherenceName)
	if !exists {
		return nil, true
	}
	var c *baseline.Coherence
	if err == nil {
		c, err = baseline.ParseCoherence(data)
	}
	if err != nil {
		fmt.Fprintf(stderr, "proviso %s: %s cannot be read: %v\n", cmd, path.Join(tree.RootName, tree.CoherenceName), err)
		return nil, false
	}
	return c, true
}

// judgeRun records the check_result on stdin in coherence_file. It reads
// coherence_file first, before its input and the element diff.
func judgeRun(t *tree.Tree, res *lint.Result, stderr io.Writer) int {
	before, ok := readCoherence("judge", t, stderr)
	if !ok {
		return exitCantRun
	}
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
// spec tree. It reads coherence_file first, before it derives anything.
func verifyRun(t *tree.Tree, res *lint.Result, asJSON bool, stdout, stderr io.Writer) int {
	c, ok := readCoherence("verify", t, stderr)
	if !ok {
		return exitCantRun
	}
	entries, base, candidates, ok := coherence("verify", t, res, stderr)
	if !ok {
		return exitCantRun
	}
	problems := baseline.Verify(c, hasNote(t), entries, candidates, base)
	if !emit("verify", stdout, stderr, func(w io.Writer) error {
		if asJSON {
			return report.VerifyJSON(w, reportRoot(t.Root), problems)
		}
		return report.VerifyText(w, problems)
	}) {
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
