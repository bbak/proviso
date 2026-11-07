package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/lint"
	"github.com/bbak/proviso/internal/rehome"
	"github.com/bbak/proviso/internal/report"
	"github.com/bbak/proviso/internal/tree"
)

// rehomeRun plans the renames that give every element an identifier
// naming its identifier home, and lists the identifier mentions of the
// identifiers it renames; with write it writes the renames, with mentions
// also the identifier mentions the mention_list on stdin names, and it
// writes the plan. Lint has found no error in t.
func rehomeRun(t *tree.Tree, res *lint.Result, asJSON, write, mentions bool, stdout, stderr io.Writer) int {
	if !recorded(t, res, stderr) {
		return exitCantRun
	}
	renames, collisions := rehome.Plan(res.Elements)
	listed := rehome.Mentions(t, res.Elements, renames)
	status := exitClean
	switch {
	case len(collisions) > 0:
		// upkeep.rehome.collision: nothing is written under the spec root.
		status = exitErrors
	case write:
		var picked []rehome.Picked
		if mentions {
			var ok bool
			if picked, ok = readPicked(listed, stderr); !ok {
				return exitCantRun
			}
		}
		if !rehomeWrite(t, res, renames, picked, stderr) {
			return exitCantRun
		}
	}
	if !emit("rehome", stdout, stderr, func(w io.Writer) error {
		if asJSON {
			return report.RehomeJSON(w, reportRoot(t.Root), renames, collisions, listed)
		}
		return report.RehomeText(w, renames, collisions, listed)
	}) {
		return exitCantRun
	}
	return status
}

// recorded reports whether the spec tree is as manifest_file records it,
// with neither coherence_file nor amendment_note beside it, following
// upkeep.rehome.baseline. It writes to stderr each of these that does not
// hold.
func recorded(t *tree.Tree, res *lint.Result, stderr io.Writer) bool {
	ok := true
	manifest := path.Join(tree.RootName, tree.ManifestName)
	if data, has := t.Manifest(); !has {
		fmt.Fprintf(stderr, "proviso rehome: there is no %s: rehome works only on a spec tree the manifest records\n", manifest)
		ok = false
	} else if base, bad := baseline.Parse(data); len(bad) > 0 {
		// baseline.diff.malformed
		fmt.Fprintf(stderr, "proviso rehome: %s cannot be read:\n", manifest)
		for _, b := range bad {
			fmt.Fprintf(stderr, "  %s\n", b)
		}
		ok = false
	} else if entries := baseline.Compare(res.Elements, base); len(entries) > 0 {
		fmt.Fprintf(stderr, "proviso rehome: the spec tree is not as %s records it; rehome works only outside an amendment:\n", manifest)
		for _, e := range entries {
			fmt.Fprintf(stderr, "  %s %s\n", e.Change, e.ID)
		}
		ok = false
	}
	for _, name := range []string{tree.CoherenceName, tree.NoteName} {
		if _, err := os.Lstat(filepath.Join(t.Root, name)); !errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(stderr, "proviso rehome: the spec root holds %s; rehome works only outside an amendment\n", path.Join(tree.RootName, name))
			ok = false
		}
	}
	return ok
}

// readPicked reads the mention_list on stdin, following
// upkeep.rehome.picked: the identifier mentions it names, each of them one
// of listed. It writes to stderr why the input cannot be taken.
func readPicked(listed []rehome.Mention, stderr io.Writer) ([]rehome.Picked, bool) {
	data, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "proviso rehome: cannot read standard input: %v; nothing is written\n", err)
		return nil, false
	}
	picked, err := rehome.ParseMentionList(data)
	if err != nil {
		fmt.Fprintf(stderr, "proviso rehome: the input of --mentions is no mention_list: %v; nothing is written\n", err)
		return nil, false
	}
	if unlisted := rehome.Unlisted(picked, listed); len(unlisted) > 0 {
		fmt.Fprintln(stderr, "proviso rehome: the input of --mentions names mentions the plan does not list; nothing is written:")
		for _, p := range unlisted {
			fmt.Fprintf(stderr, "  %s:%d: %s\n", p.File, p.Line, p.Identifier)
		}
		return nil, false
	}
	return picked, true
}

// rehomeWrite writes the renames into the files of the spec tree that hold
// a place rehome_references lists for them, the picked identifier mentions,
// and the manifest of the rewritten tree, following upkeep.rehome.write:
// either every file, once each of them and manifest_file can be opened for
// writing, or none. It writes to stderr each file it cannot open for
// writing.
func rehomeWrite(t *tree.Tree, res *lint.Result, renames []rehome.Rename, picked []rehome.Picked, stderr io.Writer) bool {
	files := rehome.Rewrite(t, res.Elements, renames, picked)
	manifest := rehome.Manifest(t, files)
	rels := append(slices.Sorted(maps.Keys(files)), tree.ManifestName)
	ok := true
	for _, rel := range rels {
		f, err := os.OpenFile(filepath.Join(t.Root, filepath.FromSlash(rel)), os.O_WRONLY, 0)
		if err != nil {
			fmt.Fprintf(stderr, "proviso rehome: cannot open %s for writing: %v; nothing is written\n", path.Join(tree.RootName, rel), err)
			ok = false
			continue
		}
		f.Close()
	}
	if !ok {
		return false
	}
	if before, _ := t.Manifest(); !bytes.Equal(before, manifest) {
		files[tree.ManifestName] = manifest
	}
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		// Written in place, as it was opened for writing; a failure here is
		// one of the operating system, which no rule decides.
		if err := os.WriteFile(filepath.Join(t.Root, filepath.FromSlash(rel)), files[rel], 0); err != nil {
			fmt.Fprintf(stderr, "proviso rehome: cannot write %s: %v\n", path.Join(tree.RootName, rel), err)
			return false
		}
	}
	return true
}
