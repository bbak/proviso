package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/stamp"
)

func TestUpdate(t *testing.T) {
	defer func(v string) { version = v }(version)
	version = "0.2.0"
	const build = ".claude/skills/proviso-build/SKILL.md"
	stamped := func(v string) string { return "---\nname: x\nmetadata:\n  proviso-version: " + v + "\n---\n" }
	all := func(v string, edit map[string]string) map[string]string {
		files := map[string]string{".git/HEAD": "", "proviso/system.md": "mine\n", ".proviso/build.md": "mine\n"}
		for _, f := range []string{".claude/agents/proviso-checker.md", ".claude/agents/proviso-coder.md", ".claude/agents/proviso-tester.md", ".claude/skills/proviso-accept/SKILL.md", build, ".claude/skills/proviso-check/SKILL.md", ".claude/skills/proviso-integrate/SKILL.md"} {
			files[f] = stamped(v)
		}
		for k, v := range edit {
			if v == "" {
				delete(files, k)
			} else {
				files[k] = v
			}
		}
		return files
	}
	cases := []struct {
		name   string
		files  map[string]string
		status int
		build  string // the line naming build on stdout, "" for none
		stderr string // a substring of stderr, "" for nothing
	}{
		{"current", all("0.2.0", nil), 0, "skipped", ""},
		{"older", all("0.1.0", nil), 0, "updated", ""},
		{"newer", all("0.2.0", map[string]string{build: stamped("0.3.0")}), 0, "skipped", build + " is stamped 0.3.0, newer"},
		{"newer pre-release", all("0.2.0", map[string]string{build: stamped("0.2.1-dev")}), 0, "skipped", "newer"},
		{"own pre-release", all("0.2.0", map[string]string{build: stamped("0.2.0-dev")}), 0, "updated", ""},
		{"build part", all("0.2.0", map[string]string{build: stamped("0.2.0+build.5")}), 0, "updated", ""},
		{"dev", all("0.2.0", map[string]string{build: stamped("dev")}), 0, "updated", ""},
		{"v", all("0.2.0", map[string]string{build: stamped("v0.2.0")}), 0, "updated", ""},
		{"two stamps", all("0.2.0", map[string]string{build: "---\nmetadata:\n  proviso-version: 0.2.0\n  proviso-version: 0.2.0\n---\n"}), 0, "updated", ""},
		{"no frontmatter", all("0.2.0", map[string]string{build: "# X\n"}), 0, "updated", ""},
		{"missing", all("0.2.0", map[string]string{build: ""}), 0, "written", ""},
		{"folder", all("0.2.0", map[string]string{build: "", build + "/x": "x"}), 2, "", "cannot write " + build},
		{"only one", map[string]string{".git/HEAD": "", build: stamped("0.2.0")}, 0, "skipped", ""},
		{"before init", map[string]string{".git/HEAD": ""}, 2, "", "proviso init installs them"},
		{"notes only", map[string]string{".git/HEAD": "", "proviso/system.md": "x", ".proviso/build.md": "x"}, 2, "", "no Proviso skill"},
		{"no repository", map[string]string{build: stamped("0.1.0")}, 2, "", "not the root of a git repository"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inRepo(t, c.files)
			var out, errOut bytes.Buffer
			status := run([]string{"update"}, &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d (stderr %q)", status, c.status, errOut.String())
			}
			if c.stderr == "" && errOut.Len() > 0 || !strings.Contains(errOut.String(), c.stderr) {
				t.Errorf("stderr %q, want %q", errOut.String(), c.stderr)
			}
			if c.build != "" && !strings.Contains(out.String(), c.build+" "+build+"\n") || c.build == "" && strings.Contains(out.String(), build) {
				t.Errorf("stdout %q, want %s %s", out.String(), c.build, build)
			}
			if c.status == 2 && c.build == "" && c.name != "folder" && out.Len() > 0 {
				t.Errorf("stdout %q, want nothing", out.String())
			}
			for rel, want := range c.files {
				if strings.HasPrefix(rel, ".claude/") {
					continue
				}
				if got, _ := os.ReadFile(filepath.FromSlash(rel)); string(got) != want {
					t.Errorf("%s changed to %q", rel, got)
				}
			}
			if c.build == "updated" || c.build == "written" {
				data, _ := os.ReadFile(filepath.FromSlash(build))
				if s := stamp.Read(data); len(s) != 1 || s[0] != "0.2.0" {
					t.Errorf("stamps %q", s)
				}
			}
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	defer func(v string) { version = v }(version)
	version = "0.2.0"
	inRepo(t, map[string]string{".git/HEAD": "", ".claude/agents/proviso-coder.md": "x"})
	var out, errOut bytes.Buffer
	if status := run([]string{"update"}, &out, &errOut); status != 0 || errOut.Len() > 0 {
		t.Fatalf("status %d, stderr %q", status, errOut.String())
	}
	want := "written .claude/agents/proviso-checker.md\nupdated .claude/agents/proviso-coder.md\nwritten .claude/agents/proviso-tester.md\n" +
		"written .claude/skills/proviso-accept/SKILL.md\nwritten .claude/skills/proviso-build/SKILL.md\nwritten .claude/skills/proviso-check/SKILL.md\nwritten .claude/skills/proviso-integrate/SKILL.md\n"
	if out.String() != want {
		t.Errorf("stdout %q", out.String())
	}
	if _, err := os.Lstat(".proviso"); err == nil {
		t.Error("update wrote a note")
	}
}

func TestUpdateRefused(t *testing.T) {
	defer func(v string) { version = v }(version)
	for _, v := range []string{"0.1.1-dev", "dev", "0.2.0-rc.1", "0.2.0+build.5"} {
		version = v
		inRepo(t, map[string]string{".git/HEAD": "", ".claude/agents/proviso-coder.md": "x"})
		var out, errOut bytes.Buffer
		if status := run([]string{"update"}, &out, &errOut); status != 2 || out.Len() > 0 || !strings.Contains(errOut.String(), "only a release updates") || !strings.Contains(errOut.String(), v) {
			t.Errorf("%s: status %d, stdout %q, stderr %q", v, status, out.String(), errOut.String())
		}
		if got, _ := os.ReadFile(".claude/agents/proviso-coder.md"); string(got) != "x" {
			t.Errorf("%s: wrote %q", v, got)
		}
	}
	version = "0.2.0"
	var out, errOut bytes.Buffer
	if status := run([]string{"update", "--json"}, &out, &errOut); status != 2 || out.Len() > 0 || errOut.Len() == 0 {
		t.Errorf("argument: status %d, stdout %q", status, out.String())
	}
}
