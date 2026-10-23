// Package git holds every call Proviso makes to git: finding the revision
// that wrote the manifest or that a name stands for, and reading the files of the spec tree at a
// revision. It runs the git command line and never writes to the
// repository.
package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// run runs git in dir and returns its standard output. GIT_OPTIONAL_LOCKS=0
// keeps git from refreshing the index, so that a read leaves the
// repository as it was.
func run(dir string, stdin io.Reader, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	cmd.Stdin = stdin
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errOut.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", args[0], msg)
	}
	return out.Bytes(), nil
}

// ManifestRevision returns the full name of the newest revision that
// changed the file name in the folder root, not counting a merge that took
// it unchanged from one of its parents, and checks that the file there
// holds the bytes current.
func ManifestRevision(root, name string, current []byte) (string, error) {
	// git log's default history simplification leaves out a merge whose
	// file is the same as in one of its parents.
	out, err := run(root, nil, "log", "-1", "--format=%H", "--", name)
	if err != nil {
		return "", err
	}
	rev := strings.TrimSpace(string(out))
	if rev == "" {
		return "", fmt.Errorf("no revision in the git history changed %s: it was never committed", name)
	}
	committed, err := run(root, nil, "cat-file", "blob", rev+":./"+name)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(committed, current) {
		return "", fmt.Errorf("%s differs from the one revision %s wrote, so the text there is not the baseline's", name, rev)
	}
	return rev, nil
}

// Commit returns the full name of the commit that name stands for, in the
// repository holding the folder root, as git reads any name of a revision.
func Commit(root, name string) (string, error) {
	// A name git could take for an option names no commit.
	if name == "" || strings.HasPrefix(name, "-") {
		return "", fmt.Errorf("git knows no commit named %q", name)
	}
	out, err := run(root, nil, "rev-parse", "--verify", "--quiet", name+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("git knows no commit named %q", name)
	}
	return strings.TrimSpace(string(out)), nil
}

// Files returns the regular files below the folder root at the revision
// rev, by their slash-separated paths relative to root.
func Files(root, rev string) (map[string][]byte, error) {
	out, err := run(root, nil, "ls-tree", "-r", "-z", rev, "--", ".")
	if err != nil {
		return nil, err
	}
	var paths, oids []string
	for _, rec := range strings.Split(string(out), "\x00") {
		meta, p, ok := strings.Cut(rec, "\t")
		f := strings.Fields(meta)
		if !ok || len(f) != 3 || f[1] != "blob" || f[0] == "120000" {
			continue // a symbolic link or a submodule holds no text of the tree
		}
		paths, oids = append(paths, p), append(oids, f[2])
	}
	if len(oids) == 0 {
		return map[string][]byte{}, nil
	}
	out, err = run(root, strings.NewReader(strings.Join(oids, "\n")+"\n"), "cat-file", "--batch")
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	r := bufio.NewReader(bytes.NewReader(out))
	for _, p := range paths {
		header, err := r.ReadString('\n')
		if err != nil {
			return nil, errors.New("git cat-file: the output ended early")
		}
		f := strings.Fields(header)
		if len(f) != 3 {
			return nil, fmt.Errorf("git cat-file: %s", strings.TrimSpace(header))
		}
		size, err := strconv.Atoi(f[2])
		if err != nil {
			return nil, fmt.Errorf("git cat-file: %s", strings.TrimSpace(header))
		}
		data := make([]byte, size+1) // the content and its line feed
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, errors.New("git cat-file: the output ended early")
		}
		files[p] = data[:size]
	}
	return files, nil
}
