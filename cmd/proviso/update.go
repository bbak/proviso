package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bbak/proviso/internal/kit"
	"github.com/bbak/proviso/internal/stamp"
)

// updateRun brings the skills and agent definitions installed in the
// working directory to this binary's version, provided the binary is a
// release and the directory is the root of a git repository holding at
// least one of them. It touches no other installed file.
func updateRun(stdout, stderr io.Writer) int {
	// install.update.release: a local build's kit is not released.
	if !stamp.IsRelease(version) {
		fmt.Fprintf(stderr, "proviso update: only a release updates, and this is Proviso %s\n", version)
		return exitCantRun
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "proviso update: %v\n", err)
		return exitCantRun
	}
	// install.update.repository: any entry named .git counts, as for init;
	// no other directory is looked at.
	if _, err := os.Lstat(filepath.Join(wd, ".git")); err != nil {
		fmt.Fprintf(stderr, "proviso update: %s is not the root of a git repository\n", wd)
		return exitCantRun
	}
	var files []kit.File
	installed := false
	for _, f := range kit.Files("", version) {
		if f.Stamped {
			files = append(files, f)
			if _, err := os.Lstat(filepath.Join(wd, filepath.FromSlash(f.Path))); err == nil {
				installed = true
			}
		}
	}
	// install.update.installed: only the skills and agents count.
	if !installed {
		fmt.Fprintln(stderr, "proviso update: no Proviso skill or agent definition is installed here; proviso init installs them")
		return exitCantRun
	}
	status := exitClean
	for _, f := range files {
		dst := filepath.Join(wd, filepath.FromSlash(f.Path))
		word, err := updateFile(dst, f.Data, f.Path, stderr)
		if err != nil {
			// install.update.exit: no line on stdout names a failed file.
			fmt.Fprintf(stderr, "proviso update: cannot write %s: %v\n", f.Path, err)
			status = exitCantRun
			continue
		}
		fmt.Fprintf(stdout, "%s %s\n", word, f.Path)
	}
	return status
}

// updateFile writes the skill or agent definition at dst, shown as rel,
// where it is missing or its stamp is neither this binary's version nor a
// newer one, and returns the word naming what it did.
func updateFile(dst string, data []byte, rel string, stderr io.Writer) (string, error) {
	// install.update.missing: any entry counts.
	if _, err := os.Lstat(dst); err != nil {
		return "written", install(dst, data)
	}
	// install.update.stamp: a file that cannot be read names no stamp.
	if old, err := os.ReadFile(dst); err == nil {
		if s := stamp.Read(old); len(s) == 1 && (s[0] == version || stamp.Newer(s[0], version)) {
			if s[0] != version {
				fmt.Fprintf(stderr, "proviso update: %s is stamped %s, newer than this Proviso %s, and is left as it is\n", rel, s[0], version)
			}
			return "skipped", nil
		}
	}
	return "updated", overwrite(dst, data)
}

// overwrite writes data over the regular file dst. It removes no entry, so
// it refuses a folder, and it writes through no link, so that no file
// outside the installed paths changes.
func overwrite(dst string, data []byte) error {
	info, err := os.Lstat(dst)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("it is not a regular file, and update removes no entry")
	}
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
