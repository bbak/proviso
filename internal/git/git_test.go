package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type repo struct {
	t    *testing.T
	dir  string
	root string // the spec root
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	r := &repo{t, dir, filepath.Join(dir, "proviso")}
	if err := os.MkdirAll(r.root, 0o755); err != nil {
		t.Fatal(err)
	}
	r.git("init", "-q", "-b", "main")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.dir, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r *repo) write(rel, content string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.root, rel), []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *repo) commit(msg string) string {
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
	return r.git("rev-parse", "HEAD")
}

func TestManifestRevision(t *testing.T) {
	r := newRepo(t)
	r.write("system.md", "a\n")
	r.commit("tree")
	r.write("manifest.sum", "model 0\n")
	if _, err := ManifestRevision(r.root, "manifest.sum", []byte("model 0\n")); err == nil {
		t.Error("a manifest never committed has a revision")
	}
	a1 := r.commit("manifest")
	r.write("system.md", "b\n")
	r.commit("edit")
	if rev, err := ManifestRevision(r.root, "manifest.sum", []byte("model 0\n")); err != nil || rev != a1 {
		t.Errorf("%s %v, want %s", rev, err, a1)
	}
	if _, err := ManifestRevision(r.root, "manifest.sum", []byte("model 0\nx\n")); err == nil {
		t.Error("a manifest written again since was taken for the committed one")
	}

	// A merge that takes the manifest unchanged from one parent.
	r.git("checkout", "-q", "-b", "side")
	r.write("manifest.sum", "model 0\nside\n")
	e5f6 := r.commit("side manifest")
	r.git("checkout", "-q", "main")
	r.write("other.md", "x\n")
	r.commit("main edit")
	r.git("merge", "-q", "--no-ff", "-m", "merge", "side")
	if rev, err := ManifestRevision(r.root, "manifest.sum", []byte("model 0\nside\n")); err != nil || rev != e5f6 {
		t.Errorf("after a merge: %s %v, want %s", rev, err, e5f6)
	}

	files, err := Files(r.root, e5f6)
	if err != nil || string(files["system.md"]) != "b\n" || string(files["manifest.sum"]) != "model 0\nside\n" || len(files) != 2 {
		t.Errorf("%q %v", files, err)
	}
}

func TestCommit(t *testing.T) {
	r := newRepo(t)
	r.write("system.md", "a\n")
	a1 := r.commit("first")
	r.write("system.md", "b\n")
	b2 := r.commit("second")
	for name, want := range map[string]string{"HEAD": b2, "HEAD~1": a1, a1[:7]: a1} {
		if rev, err := Commit(r.root, name); err != nil || rev != want {
			t.Errorf("%s: %s %v, want %s", name, rev, err, want)
		}
	}
	for _, name := range []string{"nosuch", "HEAD:proviso/system.md", "--all", ""} {
		if rev, err := Commit(r.root, name); err == nil {
			t.Errorf("%q names the commit %s", name, rev)
		}
	}
}

func TestNoRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	if _, err := ManifestRevision(dir, "manifest.sum", nil); err == nil {
		t.Error("a folder in no repository has a revision")
	}
}
