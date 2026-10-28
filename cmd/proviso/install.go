package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/bbak/proviso/internal/baseline"
	"github.com/bbak/proviso/internal/kit"
	"github.com/bbak/proviso/internal/model"
	"github.com/bbak/proviso/internal/stamp"
	"github.com/bbak/proviso/internal/tree"
	"golang.org/x/text/unicode/norm"
)

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
	for _, f := range kit.Files(norm.NFC.String(filepath.Base(wd)), version) {
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

// versionRun writes the Proviso version and the model version and then, in
// the root of a git repository, one line to stderr for each installed
// skill or agent definition and for a spec tree that does not fit them. It
// writes nothing to disk, and its warnings never change its exit status.
func versionRun(stdout, stderr io.Writer) int {
	fmt.Fprintf(stdout, "proviso %s\nmodel %d\n", version, model.Version)
	wd, err := os.Getwd()
	if err == nil {
		// Any entry named .git counts, as for init; no other directory is
		// looked at.
		_, err = os.Lstat(filepath.Join(wd, ".git"))
	}
	if err != nil {
		fmt.Fprintln(stderr, "proviso version: run in the root of a git repository, version checks the installed skills and agent definitions and the spec tree against its version")
		return exitClean
	}
	var paths []string
	installed := false
	for _, f := range kit.Files("", version) {
		if f.Stamped {
			paths = append(paths, f.Path)
			if _, err := os.Lstat(filepath.Join(wd, filepath.FromSlash(f.Path))); err == nil {
				installed = true
			}
		}
	}
	if !installed {
		fmt.Fprintln(stderr, "proviso version: no Proviso skill or agent definition is installed here")
	} else {
		for _, p := range paths {
			if msg := checkStamp(filepath.Join(wd, filepath.FromSlash(p))); msg != "" {
				fmt.Fprintf(stderr, "proviso version: %s %s\n", p, msg)
			}
		}
	}
	if msg := checkManifest(filepath.Join(wd, tree.RootName, tree.ManifestName)); msg != "" {
		fmt.Fprintf(stderr, "proviso version: %s %s\n", path.Join(tree.RootName, tree.ManifestName), msg)
	}
	return exitClean
}

// checkStamp returns how the installed skill or agent definition at p does
// not fit this binary, or "" where it fits.
func checkStamp(p string) string {
	if _, err := os.Lstat(p); err != nil {
		return "is not installed"
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "names no Proviso version: it cannot be read"
	}
	stamps := stamp.Read(data)
	switch {
	case len(stamps) > 1:
		return "holds more than one version stamp"
	case len(stamps) == 0 || !stamp.Names(stamps[0]):
		return "names no Proviso version"
	case stamps[0] == version:
		return ""
	}
	if m, ok := stamp.Model(stamps[0], version); ok {
		return fmt.Sprintf("was installed by Proviso %s, which reads model %d; this is Proviso %s", stamps[0], m, version)
	}
	return fmt.Sprintf("was installed by Proviso %s; this is Proviso %s", stamps[0], version)
}

// checkManifest returns how the manifest at p does not fit this binary's
// model version, or "" where it fits.
func checkManifest(p string) string {
	if _, err := os.Lstat(p); err != nil {
		return "does not exist, so no model version is recorded for the spec tree"
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "cannot be read"
	}
	if baseline.CheckModel(data) == nil {
		return ""
	}
	if v, ok := baseline.NamedModel(baseline.FirstLine(data)); ok {
		return fmt.Sprintf("says the spec tree needs model version %s; this proviso reads model version %d", v, model.Version)
	}
	return "names no model version"
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
