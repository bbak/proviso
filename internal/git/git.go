// Package git holds every call Proviso makes to git: finding the revision
// that wrote the manifest or that a name stands for, and reading the files of
// the spec tree at a revision. It runs the git command line and never writes
// to the repository.
//
// Each process start costs more than the work it does, so a command makes as
// few as it can: one git log to find the revision that wrote the manifest,
// and one git cat-file session for everything read at a revision.
package git

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// command returns git, run in dir. GIT_OPTIONAL_LOCKS=0 keeps git from
// refreshing the index, so that a read leaves the repository as it was.
func command(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	return cmd
}

// run runs git in dir and returns its standard output.
func run(dir string, args ...string) ([]byte, error) {
	cmd := command(dir, args...)
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

// Former returns the full name of the newest revision that changed the file
// name in the folder root, not counting a merge that took it unchanged from
// one of its parents, and the files below root there, as files says. It
// checks that the file there holds the bytes current.
func Former(root, name string, current []byte) (string, map[string][]byte, []string, error) {
	// git log's default history simplification leaves out a merge whose
	// file is the same as in one of its parents.
	out, err := run(root, "log", "-1", "--format=%H", "--", name)
	if err != nil {
		return "", nil, nil, err
	}
	rev := strings.TrimSpace(string(out))
	if rev == "" {
		return "", nil, nil, fmt.Errorf("no revision in the git history changed %s: it was never committed", name)
	}
	s, err := open(root)
	if err != nil {
		return "", nil, nil, err
	}
	files, links, err := s.files(rev)
	if err = s.close(err); err != nil {
		return "", nil, nil, err
	}
	if committed, ok := files[name]; !ok || !bytes.Equal(committed, current) {
		return "", nil, nil, fmt.Errorf("%s differs from the one revision %s wrote, so the text there is not the baseline's", name, rev)
	}
	return rev, files, links, nil
}

// Since returns the full name of the commit that name stands for, in the
// repository holding the folder root, as git reads any name of a revision,
// and the files below root there, as files says.
func Since(root, name string) (string, map[string][]byte, []string, error) {
	// A name git could take for an option names no commit, and a line
	// break would end the request early.
	if name == "" || strings.HasPrefix(name, "-") || strings.ContainsAny(name, "\r\n") {
		return "", nil, nil, fmt.Errorf("git knows no commit named %q", name)
	}
	s, err := open(root)
	if err != nil {
		return "", nil, nil, err
	}
	rev, typ, _, ok, err := s.get(name + "^{commit}")
	if err == nil && (!ok || typ != "commit") {
		err = fmt.Errorf("git knows no commit named %q", name)
	}
	var files map[string][]byte
	var links []string
	if err == nil {
		files, links, err = s.files(rev)
	}
	if err = s.close(err); err != nil {
		return "", nil, nil, err
	}
	return rev, files, links, nil
}

// session is one git cat-file --batch process, asked for one object at a
// time, so that a tree is walked without a process per folder.
type session struct {
	cmd    *exec.Cmd
	in     io.WriteCloser
	out    *bufio.Reader
	errOut bytes.Buffer
}

func open(dir string) (*session, error) {
	s := &session{cmd: command(dir, "cat-file", "--batch")}
	s.cmd.Stderr = &s.errOut
	in, err := s.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("git cat-file: %v", err)
	}
	out, err := s.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("git cat-file: %v", err)
	}
	if err := s.cmd.Start(); err != nil {
		return nil, fmt.Errorf("git cat-file: %v", err)
	}
	s.in, s.out = in, bufio.NewReader(out)
	return s, nil
}

// close ends the session. It returns err if there is one, and otherwise
// whatever git reported on ending; a session that failed reports what git
// wrote to standard error, which says more than a broken pipe.
func (s *session) close(err error) error {
	s.in.Close()
	_, _ = io.Copy(io.Discard, s.out)
	werr := s.cmd.Wait()
	if msg := strings.TrimSpace(s.errOut.String()); werr != nil && msg != "" {
		werr = fmt.Errorf("git cat-file: %s", msg)
	} else if werr != nil {
		werr = fmt.Errorf("git cat-file: %v", werr)
	}
	if err != nil {
		if errors.Is(err, errSession) && werr != nil {
			return werr
		}
		return err
	}
	return werr
}

// errSession marks a request the session could not answer, such as one
// git ended before answering.
var errSession = errors.New("git cat-file: the session ended early")

// get returns the id, the type and the content of the object name stands
// for; ok is false where there is none.
func (s *session) get(name string) (id, typ string, data []byte, ok bool, err error) {
	if _, err := io.WriteString(s.in, name+"\n"); err != nil {
		return "", "", nil, false, errSession
	}
	header, err := s.out.ReadString('\n')
	if err != nil {
		return "", "", nil, false, errSession
	}
	if strings.HasSuffix(header, " missing\n") || strings.HasSuffix(header, " ambiguous\n") {
		return "", "", nil, false, nil
	}
	f := strings.Fields(header)
	if len(f) != 3 {
		return "", "", nil, false, fmt.Errorf("git cat-file: %s", strings.TrimSpace(header))
	}
	size, err := strconv.Atoi(f[2])
	if err != nil {
		return "", "", nil, false, fmt.Errorf("git cat-file: %s", strings.TrimSpace(header))
	}
	data = make([]byte, size+1) // the content and its line feed
	if _, err := io.ReadFull(s.out, data); err != nil {
		return "", "", nil, false, errSession
	}
	return f[0], f[1], data[:size], true, nil
}

// files returns the regular files below the folder the session runs in,
// at the revision rev, by their slash-separated paths relative to it, and
// the paths of the symbolic links there, whose targets are never read. A
// submodule holds no text of the tree and is left out. Where the folder
// does not exist at rev, it holds no files.
func (s *session) files(rev string) (map[string][]byte, []string, error) {
	files := map[string][]byte{}
	var links []string
	id, typ, data, ok, err := s.get(rev + ":./")
	if err != nil || !ok || typ != "tree" {
		return files, links, err
	}
	// A tree names its entries by raw ids, as long as the hex id of the
	// tree itself is in bytes: SHA-1 or SHA-256, whichever the repository
	// uses.
	err = s.walk("", data, len(id)/2, files, &links)
	return files, links, err
}

// walk reads the entries of the tree object data, in the order git keeps
// them, with dir before each path.
func (s *session) walk(dir string, data []byte, idLen int, files map[string][]byte, links *[]string) error {
	for len(data) > 0 {
		sp, nul := bytes.IndexByte(data, ' '), bytes.IndexByte(data, 0)
		if sp < 0 || nul < sp || nul+1+idLen > len(data) {
			return fmt.Errorf("git cat-file: a tree below %q cannot be read", dir)
		}
		mode, name := string(data[:sp]), string(data[sp+1:nul])
		id := hex.EncodeToString(data[nul+1 : nul+1+idLen])
		data = data[nul+1+idLen:]
		path := dir + name
		switch mode {
		case "40000":
			_, typ, sub, ok, err := s.get(id)
			if err != nil {
				return err
			}
			if !ok || typ != "tree" {
				return fmt.Errorf("git cat-file: the tree %s of %q is missing", id, path)
			}
			if err := s.walk(path+"/", sub, idLen, files, links); err != nil {
				return err
			}
		case "120000":
			*links = append(*links, path)
		case "160000":
			// A submodule holds no text of the tree.
		default:
			_, typ, blob, ok, err := s.get(id)
			if err != nil {
				return err
			}
			if !ok || typ != "blob" {
				return fmt.Errorf("git cat-file: the file %s of %q is missing", id, path)
			}
			files[path] = blob
		}
	}
	return nil
}
