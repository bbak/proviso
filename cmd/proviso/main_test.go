package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const system = "# <a id=\"system\"></a> S\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n## Scope\n\ns\n\n## Out of scope\n\no\n\n## Audience\n\na\n\n## Design boundary\n\nd\n"

func inRepo(t *testing.T, files map[string]string) {
	t.Helper()
	repo := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(repo)
}

func TestRun(t *testing.T) {
	cases := []struct {
		name   string
		files  map[string]string
		args   []string
		status int
		stdout string // a substring, or "" for nothing written
	}{
		{"clean text", map[string]string{"proviso/system.md": system}, []string{"lint"}, 0, ""},
		{"clean json", map[string]string{"proviso/system.md": system}, []string{"lint", "--json"}, 0, `"findings": []`},
		{"error text", map[string]string{"proviso/system.md": system, "proviso/notes.md": "x"}, []string{"lint"}, 1, "proviso/notes.md:1: error file-unknown: "},
		{"ask only", map[string]string{"proviso/system.md": strings.Replace(system, "\ng\n", "\n?\n", 1)}, []string{"lint"}, 0, "ask undecided"},
		{"uppercase name", map[string]string{"proviso/system.md": system, "proviso/Glossary.md": "x"}, []string{"lint"}, 1, "proviso/Glossary.md:1: error name-malformed: "},
		{"decomposed name", map[string]string{"proviso/system.md": system, "proviso/ü.md": "x"}, []string{"lint"}, 1, "proviso/ü.md:1: error file-unknown: "},
		{"carriage return", map[string]string{"proviso/system.md": strings.ReplaceAll(system, "\n", "\r\n")}, []string{"lint"}, 1, "proviso/system.md:1: error line-ending: "},
		{"no root", map[string]string{"system.md": system}, []string{"lint"}, 2, ""},
		{"unreadable", map[string]string{"proviso/system.md": system, "proviso/glossary.md": "\xff"}, []string{"lint", "--json"}, 2, ""},
		{"bad option", map[string]string{"proviso/system.md": system}, []string{"lint", "--yaml"}, 2, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inRepo(t, c.files)
			var out, errOut bytes.Buffer
			status := run(c.args, &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d (stderr %q)", status, c.status, errOut.String())
			}
			if c.stdout == "" && out.Len() > 0 || !strings.Contains(out.String(), c.stdout) {
				t.Errorf("stdout %q, want %q", out.String(), c.stdout)
			}
			if c.status == 2 && errOut.Len() == 0 {
				t.Error("nothing on stderr")
			}
		})
	}
}

func TestJSONReport(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/notes.md": "x"})
	var out, errOut bytes.Buffer
	run([]string{"lint", "--json"}, &out, &errOut)
	var r struct {
		Root     string
		Findings []map[string]any
	}
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(r.Root, "/proviso") || strings.Contains(r.Root, "\\") || len(r.Findings) != 1 {
		t.Fatalf("%+v", r)
	}
	f := r.Findings[0]
	if f["kind"] != "file-unknown" || f["severity"] != "error" || f["file"] != "proviso/notes.md" || f["line"] != 1.0 {
		t.Errorf("%+v", f)
	}
	if _, ok := f["related"]; ok {
		t.Error("an empty related is left out")
	}
}

func TestReportRoot(t *testing.T) {
	if runtime.GOOS != "windows" {
		if got := reportRoot("/repo/proviso"); got != "/repo/proviso" {
			t.Errorf("got %q", got)
		}
		if got := reportRoot("/ü/proviso"); got != "/ü/proviso" {
			t.Errorf("a decomposed name: got %q", got)
		}
		return
	}
	cases := map[string]string{
		`c:\repo\proviso`:        "C:/repo/proviso",
		`C:\repo\proviso`:        "C:/repo/proviso",
		`\\server\share\proviso`: "//server/share/proviso",
		"c:\\u\u0308\\proviso":   "C:/\u00fc/proviso",
	}
	for in, want := range cases {
		if got := reportRoot(in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestModelVersion(t *testing.T) {
	for manifest, status := range map[string]int{"model 0\n": 0, "model 1\n": 2, "": 2, "model 0\r\n": 1, "model 0\n\xff\n": 2} {
		inRepo(t, map[string]string{"proviso/system.md": system, "proviso/manifest.sum": manifest})
		var out, errOut bytes.Buffer
		if got := run([]string{"lint", "--json"}, &out, &errOut); got != status || status == 2 && (out.Len() > 0 || errOut.Len() == 0) {
			t.Errorf("%q: status %d, stdout %q, stderr %q", manifest, got, out.String(), errOut.String())
		}
	}
}

func TestBaselineCommands(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		return status, out.String(), errOut.String()
	}
	if status, out, _ := call("diff", "--json"); status != 1 || !strings.Contains(out, `"revision": null`) || !strings.Contains(out, `"change": "added"`) {
		t.Errorf("without a manifest: %d %s", status, out)
	}
	if status, out, errOut := call("manifest"); status != 0 || out != "" || errOut != "" {
		t.Fatalf("manifest: %d %q %q", status, out, errOut)
	}
	data, err := os.ReadFile(filepath.Join("proviso", "manifest.sum"))
	if err != nil || !strings.HasPrefix(string(data), "model 0\nsystem\theading\th1:") {
		t.Fatalf("%q %v", data, err)
	}
	if status, out, _ := call("diff"); status != 0 || out != "" {
		t.Errorf("unchanged, text: %d %q", status, out)
	}
	if status, out, _ := call("diff", "--json"); status != 0 || !strings.Contains(out, `"elements": []`) {
		t.Errorf("unchanged, json: %d %q", status, out)
	}
	if status, _, _ := call("manifest", "--json"); status != 2 {
		t.Error("manifest takes no --json")
	}

	// A changed part needs former text, which a tree in no git repository
	// cannot give.
	os.WriteFile(filepath.Join("proviso", "system.md"), []byte(strings.Replace(system, "\ng\n", "\nh\n", 1)), 0o644)
	if status, out, errOut := call("diff", "--json"); status != 2 || out != "" || errOut == "" {
		t.Errorf("no former text: %d %q %q", status, out, errOut)
	}

	os.WriteFile(filepath.Join("proviso", "manifest.sum"), append(data, "<<<<<<< HEAD\n"...), 0o644)
	if status, out, errOut := call("diff"); status != 2 || out != "" || !strings.Contains(errOut, "<<<<<<< HEAD") {
		t.Errorf("malformed: %d %q %q", status, out, errOut)
	}

	os.WriteFile(filepath.Join("proviso", "notes.md"), []byte("x"), 0o644)
	for _, cmd := range []string{"manifest", "diff"} {
		status, out, errOut := call(cmd)
		if status != 2 || out != "" || !strings.Contains(errOut, "error file-unknown") {
			t.Errorf("%s on a tree with an error: %d %q %q", cmd, status, out, errOut)
		}
	}
	after, _ := os.ReadFile(filepath.Join("proviso", "manifest.sum"))
	if !bytes.Equal(after, append(data, "<<<<<<< HEAD\n"...)) {
		t.Error("a refused manifest changed the former one")
	}
}

func TestImpactCommand(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	var out, errOut bytes.Buffer
	if status := run([]string{"impact", "--json"}, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	var r struct {
		Revision   *string
		Candidates []any
		Scopes     []struct {
			ID         string
			Candidates []string
		}
	}
	if err := json.Unmarshal(out.Bytes(), &r); err != nil || r.Revision != nil || r.Candidates == nil || len(r.Candidates) != 0 || len(r.Scopes) != 1 || r.Scopes[0].ID != "system" {
		t.Errorf("%v %s", err, out.String())
	}
	out.Reset()
	if status := run([]string{"impact"}, &out, &errOut); status != 0 || out.Len() != 0 {
		t.Errorf("text without candidates: %d %q", status, out.String())
	}
	os.WriteFile(filepath.Join("proviso", "notes.md"), []byte("x"), 0o644)
	out.Reset()
	errOut.Reset()
	if status := run([]string{"impact", "--json"}, &out, &errOut); status != 2 || out.Len() != 0 || !strings.Contains(errOut.String(), "file-unknown") {
		t.Errorf("a tree with an error: %d %q %q", status, out.String(), errOut.String())
	}
}

func TestImpactDamagedManifest(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/manifest.sum": "model 0\n<<<<<<< HEAD\n"})
	var out, errOut bytes.Buffer
	if status := run([]string{"impact", "--json"}, &out, &errOut); status != 2 || out.Len() != 0 || !strings.Contains(errOut.String(), "<<<<<<< HEAD") {
		t.Errorf("%d %q %q", status, out.String(), errOut.String())
	}
}

func TestPackCommand(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		return status, out.String(), errOut.String()
	}
	if status, _, errOut := call("manifest"); status != 0 {
		t.Fatalf("manifest: %d %q", status, errOut)
	}
	if status, out, _ := call("pack"); status != 0 || out != "# System\n\n"+system {
		t.Errorf("an empty element diff: %d %q", status, out)
	}
	if status, out, _ := call("pack", "--json"); status != 0 || !strings.Contains(out, `"diff": []`) || !strings.Contains(out, `"note": null`) || !strings.Contains(out, `"heading": "system"`) {
		t.Errorf("json: %d %s", status, out)
	}
	note := filepath.Join("proviso", ".amendment.md")
	os.WriteFile(note, []byte("\n  \n"), 0o644)
	if status, out, _ := call("pack"); status != 0 || out != "# System\n\n"+system {
		t.Errorf("a blank note: %d %q", status, out)
	}
	os.WriteFile(note, []byte("- refunds"), 0o644)
	if status, out, _ := call("pack"); status != 0 || out != "# System\n\n"+system+"\n# Not in this amendment\n\n- refunds\n" {
		t.Errorf("a note: %d %q", status, out)
	}
	if status, out, _ := call("pack", "--json"); status != 0 || !strings.Contains(out, `"note": "- refunds"`) {
		t.Errorf("a note, json: %d %s", status, out)
	}
	os.Remove(note)
	os.Mkdir(note, 0o755)
	if status, out, errOut := call("pack"); status != 2 || out != "" || !strings.Contains(errOut, ".amendment.md cannot be read") {
		t.Errorf("a folder for a note: %d %q %q", status, out, errOut)
	}
	os.Remove(note)
	os.WriteFile(note, []byte("- caf\xe9\n"), 0o644)
	for _, args := range [][]string{{"pack"}, {"pack", "--json"}} {
		if status, out, errOut := call(args...); status != 2 || out != "" || !strings.Contains(errOut, ".amendment.md cannot be read") {
			t.Errorf("a note not in UTF-8, %v: %d %q %q", args, status, out, errOut)
		}
	}
	os.Remove(note)
	os.WriteFile(filepath.Join("proviso", "manifest.sum"), []byte("model 0\n<<<<<<< HEAD\n"), 0o644)
	if status, out, errOut := call("pack"); status != 2 || out != "" || !strings.Contains(errOut, "<<<<<<< HEAD") {
		t.Errorf("malformed: %d %q %q", status, out, errOut)
	}
	os.WriteFile(filepath.Join("proviso", "notes.md"), []byte("x"), 0o644)
	if status, out, errOut := call("pack", "--json"); status != 2 || out != "" || !strings.Contains(errOut, "error file-unknown") {
		t.Errorf("a tree with an error: %d %q %q", status, out, errOut)
	}
}

func TestManifestRemoves(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/.coherence.json/inner/x": "x", "outside.md": "kept"})
	note := filepath.Join("proviso", ".amendment.md")
	if err := os.Symlink(filepath.Join("..", "outside.md"), note); err != nil {
		t.Logf("no link, a folder instead: %v", err)
		os.MkdirAll(filepath.Join(note, "inner"), 0o755)
	}
	var out, errOut bytes.Buffer
	if status := run([]string{"manifest"}, &out, &errOut); status != 0 || errOut.Len() > 0 {
		t.Fatalf("%d %q", status, errOut.String())
	}
	for _, p := range []string{note, filepath.Join("proviso", ".coherence.json")} {
		if _, err := os.Lstat(p); err == nil {
			t.Errorf("%s is left", p)
		}
	}
	if data, err := os.ReadFile("outside.md"); err != nil || string(data) != "kept" {
		t.Errorf("the link's target: %q %v", data, err)
	}
}

func TestInit(t *testing.T) {
	all := []string{".claude/agents/proviso-checker.md", ".claude/agents/proviso-coder.md", ".claude/agents/proviso-tester.md", ".claude/skills/proviso-accept/SKILL.md", ".claude/skills/proviso-build/SKILL.md", ".claude/skills/proviso-check/SKILL.md", ".proviso/build.md", ".proviso/coder.md", ".proviso/tester.md", "proviso/.gitattributes", "proviso/.markdownlint.jsonc", "proviso/system.md"}
	lines := func(verb string, paths ...string) string {
		var b strings.Builder
		for _, p := range paths {
			b.WriteString(verb + " " + p + "\n")
		}
		return b.String()
	}
	cases := []struct {
		name   string
		files  map[string]string
		status int
		stdout string
	}{
		{"not a repository", map[string]string{"a.txt": "x"}, 2, ""},
		{"fresh clone", map[string]string{".git/HEAD": "x"}, 0, lines("written", all...)},
		{"worktree", map[string]string{".git": "gitdir: x"}, 0, lines("written", all...)},
		{"existing system", map[string]string{".git": "x", "proviso/system.md": "mine"}, 0, lines("written", all[:11]...) + lines("skipped", all[11])},
		{".claude a file", map[string]string{".git": "x", ".claude": "mine"}, 2, lines("written", all[6:]...)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inRepo(t, c.files)
			var out, errOut bytes.Buffer
			if status := run([]string{"init"}, &out, &errOut); status != c.status {
				t.Errorf("status %d, want %d (stderr %q)", status, c.status, errOut.String())
			}
			if out.String() != c.stdout {
				t.Errorf("stdout %q, want %q", out.String(), c.stdout)
			}
			for rel, content := range c.files {
				if data, err := os.ReadFile(rel); err != nil || string(data) != content {
					t.Errorf("%s changed: %q, %v", rel, data, err)
				}
			}
			if c.status == 0 {
				out.Reset()
				if status := run([]string{"init"}, &out, &errOut); status != 0 || out.String() != lines("skipped", all...) {
					t.Errorf("second run: %d, %q", status, out.String())
				}
			}
		})
	}
}

func TestInitSkeleton(t *testing.T) {
	inRepo(t, map[string]string{".git/HEAD": "x"})
	wd, _ := os.Getwd()
	var out, errOut bytes.Buffer
	if status := run([]string{"init"}, &out, &errOut); status != 0 {
		t.Fatalf("init: %d %s", status, errOut.String())
	}
	data, _ := os.ReadFile("proviso/system.md")
	if want := "# <a id=\"system\"></a> " + filepath.Base(wd) + "\n"; !strings.HasPrefix(string(data), want) {
		t.Errorf("heading %q, want %q", data, want)
	}
	out.Reset()
	if status := run([]string{"lint"}, &out, &errOut); status != 0 || strings.Count(out.String(), "ask undecided") != 6 {
		t.Errorf("lint: %d %q", status, out.String())
	}
}

func TestSince(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	inRepo(t, map[string]string{"README": "x\n"})
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	commit := func() string {
		git("add", "-A")
		git("commit", "-q", "-m", "x")
		return git("rev-parse", "HEAD")
	}
	write := func(rel, content string) {
		if err := os.MkdirAll(filepath.Dir(rel), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(rel, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q", "-b", "main")
	before := commit()
	write("proviso/system.md", system)
	a1 := commit()
	write("proviso/system.md", strings.Replace(system, "\ng\n", "\ngoal\n", 1))
	write("proviso/manifest.sum", "<<<<<<< HEAD\n")
	write("proviso/notes.md", "x")
	c3 := commit()
	os.Remove("proviso/notes.md")

	cases := []struct {
		name   string
		args   []string
		status int
		stdout string // a substring, or "" for nothing written
	}{
		{"changed since a1", []string{"diff", "--json", "--since", a1[:7]}, 1, `"revision": "` + a1 + `"`},
		{"former text from a1", []string{"diff", "--json", "--since", a1}, 1, `"former": "g\n"`},
		{"before the tree", []string{"diff", "--json", "--since", before}, 1, `"change": "added"`},
		{"no former text needed", []string{"diff", "--json", "--since", before}, 1, `"revision": null`},
		{"pack since a1", []string{"pack", "--since", "HEAD~1"}, 0, "Former text from `" + a1 + "`."},
		{"no such commit", []string{"diff", "--since", "nosuch"}, 2, ""},
		{"lint error there", []string{"pack", "--since", c3}, 2, ""},
		{"no revision", []string{"diff", "--since"}, 2, ""},
		{"twice", []string{"diff", "--since", a1, "--since", a1}, 2, ""},
		{"impact", []string{"impact", "--since", a1}, 2, ""},
		{"damaged manifest without since", []string{"diff"}, 2, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			status := run(c.args, &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d (stderr %q)", status, c.status, errOut.String())
			}
			if c.stdout == "" && out.Len() > 0 || !strings.Contains(out.String(), c.stdout) {
				t.Errorf("stdout %q, want %q", out.String(), c.stdout)
			}
			if c.status == 2 && errOut.Len() == 0 {
				t.Error("nothing on stderr")
			}
		})
	}
}

func TestJudgeAndVerify(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	coherence := filepath.Join("proviso", ".coherence.json")
	call := func(input string, args ...string) (int, string, string) {
		stdin = strings.NewReader(input)
		defer func() { stdin = os.Stdin }()
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		return status, out.String(), errOut.String()
	}
	if status, out, _ := call("", "verify"); status != 1 || !strings.HasPrefix(out, "missing: ") {
		t.Errorf("without a check: %d %q", status, out)
	}
	if status, out, errOut := call("not json", "judge"); status != 1 || out != "" || !strings.Contains(errOut, "no check_result") {
		t.Errorf("not JSON: %d %q %q", status, out, errOut)
	}
	if _, err := os.Stat(coherence); err == nil {
		t.Error("a refused check result is recorded")
	}
	if status, _, _ := call("{}", "judge", "--json"); status != 2 {
		t.Error("judge takes no --json")
	}
	if status, out, errOut := call("{}", "judge"); status != 0 || out != "" || errOut != "" {
		t.Fatalf("judge: %d %q %q", status, out, errOut)
	}
	data, err := os.ReadFile(coherence)
	if err != nil {
		t.Fatal(err)
	}
	if status, out, _ := call("", "verify", "--json"); status != 0 || !strings.Contains(out, `"problems": []`) {
		t.Errorf("covered: %d %q", status, out)
	}
	if status, out, _ := call("", "verify"); status != 0 || out != "" {
		t.Errorf("covered, text: %d %q", status, out)
	}
	os.WriteFile(coherence, append([]byte("<<<<<<< HEAD\n"), data...), 0o644)
	for _, cmd := range []string{"judge", "verify"} {
		if status, out, errOut := call("{}", cmd); status != 2 || out != "" || !strings.Contains(errOut, "cannot be read") {
			t.Errorf("%s on a damaged file: %d %q %q", cmd, status, out, errOut)
		}
	}
	if status, out, errOut := call("not json", "judge"); status != 2 || out != "" || !strings.Contains(errOut, "cannot be read") {
		t.Errorf("judge on a damaged file, with input that is no JSON: %d %q %q", status, out, errOut)
	}
	os.WriteFile(coherence, data, 0o644)
	note := filepath.Join("proviso", ".amendment.md")
	os.WriteFile(note, []byte("- refunds\n"), 0o644)
	if status, out, _ := call("", "verify"); status != 0 || out != "" {
		t.Errorf("covered, with a note: %d %q", status, out)
	}
	if status, _, errOut := call("", "manifest"); status != 0 {
		t.Fatalf("manifest: %d %q", status, errOut)
	}
	if _, err := os.Stat(coherence); err == nil {
		t.Error("manifest leaves the coherence file")
	}
	if _, err := os.Stat(note); err == nil {
		t.Error("manifest leaves the amendment note")
	}
	os.WriteFile(note, nil, 0o644)
	if status, out, _ := call("", "verify"); status != 1 || !strings.HasPrefix(out, "leftover: ") {
		t.Errorf("a note left over: %d %q", status, out)
	}
	os.WriteFile(coherence, data, 0o644)
	if status, out, _ := call("", "verify"); status != 1 || strings.Count(out, "leftover") != 1 {
		t.Errorf("both left over: %d %q", status, out)
	}
	os.Remove(coherence)
	os.Remove(note)
	os.MkdirAll(filepath.Join(note, "inner"), 0o755)
	if status, out, _ := call("", "verify"); status != 1 || !strings.HasPrefix(out, "leftover: ") {
		t.Errorf("a folder for a note left over: %d %q", status, out)
	}
	if status, _, errOut := call("", "manifest"); status != 0 {
		t.Errorf("manifest with a folder for a note: %d %q", status, errOut)
	}
	if _, err := os.Lstat(note); err == nil {
		t.Error("manifest leaves a folder for a note")
	}
	if status, out, _ := call("", "verify"); status != 0 || out != "" {
		t.Errorf("accepted: %d %q", status, out)
	}
	if status, _, errOut := call("{}", "judge"); status != 1 || !strings.Contains(errOut, "no amendment") {
		t.Errorf("nothing to judge: %d %q", status, errOut)
	}
	os.WriteFile(coherence, data, 0o644)
	if status, out, _ := call("", "verify", "--json"); status != 1 || !strings.Contains(out, `"kind": "leftover"`) {
		t.Errorf("leftover: %d %q", status, out)
	}
}
