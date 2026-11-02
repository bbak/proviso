package acceptance

// The glossary entry system.regular-file (proviso/glossary.md), checked
// through the commands whose rules name it: an entry that holds no data,
// such as a socket, is neither a regular file nor a folder, on every
// platform. On Windows a socket is a reparse point that is no name
// surrogate. A OneDrive placeholder or a deduplicated file, which count as
// regular files, cannot be made by a test without those services.

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

// makeSocket makes path a socket and leaves it behind, no longer listened
// on. The socket is made in a folder of a short name and then moved to
// path, since the name a socket is bound to is limited to about 100 bytes.
func makeSocket(path string) error {
	dir, err := os.MkdirTemp("", "s")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	short := filepath.Join(dir, "s")
	l, err := net.Listen("unix", short)
	if err != nil {
		return err
	}
	l.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := l.Close(); err != nil {
		return err
	}
	return os.Rename(short, path)
}

func TestSystemRegularFile(t *testing.T) {
	t.Parallel()
	t.Run("derived: pack stops at an amendment note that is a socket", func(t *testing.T) {
		t.Parallel()
		// baseline.pack.note-unreadable
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		if err := makeSocket(filepath.Join(r.root(), noteName)); err != nil {
			t.Skipf("cannot create a socket here: %v", err)
		}
		r.pack().stoppedNaming(t, noteName)
	})
	for _, cmd := range []string{"judge", "verify"} {
		t.Run("derived: "+cmd+" stops at a coherence file that is a socket", func(t *testing.T) {
			t.Parallel()
			// baseline.judge.unreadable
			r := amendedRepo(t, iLimitThen)
			r.judgeAll(nil)
			p := filepath.Join(r.root(), coherenceName)
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			if err := makeSocket(p); err != nil {
				t.Skipf("cannot create a socket here: %v", err)
			}
			if cmd == "judge" {
				refused(t, r.judge(checkResult{}), 2, coherenceName)
				return
			}
			r.verify().stoppedNaming(t, coherenceName)
		})
	}
	t.Run("derived: manifest removes an amendment note that is a socket", func(t *testing.T) {
		t.Parallel()
		// baseline.manifest.remove: whatever stands under the name goes.
		r := amendedRepo(t, iLimitThen)
		if err := makeSocket(filepath.Join(r.root(), noteName)); err != nil {
			t.Skipf("cannot create a socket here: %v", err)
		}
		r.manifest()
		if r.hasNote() {
			t.Errorf("want the socket %s removed", noteName)
		}
	})
	// For lint, see TestLintRunUnreadable: "neither a regular file nor a
	// folder", where makeSpecial makes a socket on Windows.
}
