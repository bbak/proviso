package acceptance

// The glossary entry system.model-line (proviso/glossary.md), checked as a
// property through the two commands whose rules name it: lint
// (lint.run.model, lint.run.model-other) and version
// (install.version.manifest-model, install.version.manifest-other) judge the
// first line of manifest.sum by the same term, read here by an oracle of its
// own.

import (
	"path/filepath"
	"strings"
	"testing"
)

// modelLineOf reads the first line of content as system.model-line defines
// it, and gives the model version it names, and whether it is a model line:
// without its line ending, a line feed or a carriage return directly followed
// by a line feed, the line is "model", a space and a whole number in decimal
// digits without a leading zero. The last line of a file may have no line
// ending, and a carriage return anywhere else belongs to the line.
func modelLineOf(content string) (string, bool) {
	line := content
	if i := strings.IndexByte(content, '\n'); i >= 0 {
		line = strings.TrimSuffix(content[:i], "\r")
	}
	n, ok := strings.CutPrefix(line, "model ")
	if !ok || n == "" || (n[0] == '0' && n != "0") {
		return "", false
	}
	for _, c := range n {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	return n, true
}

// digitsOf gives the digits of line, in their order.
func digitsOf(line string) string {
	var b strings.Builder
	for _, c := range line {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// judgedAsModelLine checks the line l a command wrote about a first line
// whose content is content, judged as the oracle judges it: a model line
// outside the model versions read is named as needed, with those versions;
// any other first line is never named as a model version needed.
func judgedAsModelLine(t *testing.T, cmd, l, content string, nums []string) {
	t.Helper()
	if v, ok := modelLineOf(content); ok {
		if count(nums, v) == 0 || !namesReadVersions(nums) {
			t.Errorf("system.model-line: %s: the first line of %q is a model line naming %s; want it named as the model version the tree needs, with the model versions read, from %s to %s: %q", cmd, content, v, oldestModelVersion, modelVersion, l)
		}
		return
	}
	if count(nums, modelVersion) > 1 {
		t.Errorf("system.model-line: %s: the first line of %q is no model line, but model %s is named as needed: %q", cmd, content, modelVersion, l)
	}
	first := content
	if i := strings.IndexAny(content, "\r\n"); i >= 0 {
		first = content[:i]
	}
	if d := digitsOf(first); d != "" && count(readModelVersions(), d) == 0 && count(nums, d) > 0 && namesReadVersions(nums) {
		t.Errorf("system.model-line: %s: the first line of %q is no model line, but model %s is named as needed: %q", cmd, content, d, l)
	}
}

func TestSystemModelLine(t *testing.T) {
	t.Parallel()
	lines := []string{
		"model 0",
		"model 1",
		"model 2",
		"model 10",
		"model 12345678901234567890",
		"model",
		"model ",
		"model 01",
		"model  1",
		" model 1",
		"model 1 ",
		"model\t1",
		"Model 1",
		"model -1",
		"model +1",
		"model 1.0",
		"model 1a",
		"model 0\rfoo",
		"model 1\r",
		"<<<<<<< HEAD",
	}
	// Each line ends in a line feed, in a carriage return and a line feed, or
	// in nothing, as the last line of the file. A line that follows is a
	// model line of model_version, so that only the first line can decide.
	endings := []struct{ name, end, rest string }{
		{"a line feed", "\n", "model " + modelVersion + "\n"},
		{"a carriage return and a line feed", "\r\n", "model " + modelVersion + "\n"},
		{"nothing, as the last line", "", ""},
	}
	for _, line := range lines {
		for _, e := range endings {
			first := line + e.end
			content := first + e.rest
			v, isModel := modelLineOf(first)
			read := isModel && count(readModelVersions(), v) > 0
			t.Run("derived: "+strings.NewReplacer("\r", `\r`, "\t", `\t`).Replace(line)+", ending in "+e.name, func(t *testing.T) {
				t.Parallel()
				t.Run("lint", func(t *testing.T) {
					t.Parallel()
					dir := baseline().with("proviso/manifest.sum", content).write(t)
					o := lintIn(t, dir, dir, nil)
					if read {
						// lint.run.model: the tree is checked.
						if o.json.exit == 2 {
							t.Errorf("system.model-line: lint: the first line of %q is a model line naming %s; want the tree checked, got exit status 2: %s", content, v, o.json.stderr)
						}
						return
					}
					// lint.run.model-other: exit status 2, and nothing on
					// stdout, checked for every run with exit status 2.
					if o.json.exit != 2 {
						t.Errorf("system.model-line: lint: the first line of %q is no model line naming a model version from %s to %s; want exit status 2, got %d", content, oldestModelVersion, modelVersion, o.json.exit)
						return
					}
					nums := numbers(o.json.stderr, dir, filepath.ToSlash(dir), "manifest.sum")
					judgedAsModelLine(t, "lint", o.json.stderr, first, nums)
				})
				t.Run("version", func(t *testing.T) {
					t.Parallel()
					r := versionRepo(t, installedAs("dev").with("proviso/manifest.sum", content))
					o := r.version("dev")
					// install.version.manifest-model: no line about the
					// tree; otherwise install.version.manifest-other's line.
					o.reports(t, expect{manifest: !read})
					if read || len(o.lines) == 0 {
						return
					}
					l := o.manifestLine()
					nums := numbers(l, r.dir, filepath.ToSlash(r.dir), "manifest.sum")
					judgedAsModelLine(t, "version", l, first, nums)
				})
			})
		}
	}
}
