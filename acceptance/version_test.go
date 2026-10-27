package acceptance

// Telling the version: proviso/capabilities/install/features/version.md, and
// the command `proviso version` of contracts/cli.md. What holds for every run
// of version (its standard output, exit status and
// install.version.reads-only) is checked by versionIn; the order of the lines
// on standard error by reports.
//
// model_releases is not configurable, and the binary holds the table as it
// is: 0 from 0.1.0. Rows that assume another table are skipped and checked
// by an analogue on the real table.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const buildSkill = ".claude/skills/proviso-build/SKILL.md"

// TestSystemCliVersion checks the contract system.cli.version: its example,
// the order of the lines on standard error, and silence where everything
// fits.
func TestSystemCliVersion(t *testing.T) {
	t.Parallel()
	t.Run("example: proviso 0.1.0 and model 0", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "notes")
		tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
		o := versionIn(t, "0.1.0", dir, dir, nil)
		if got := strings.ReplaceAll(o.stdout, "\r\n", "\n"); got != "proviso 0.1.0\nmodel 0\n" {
			t.Errorf("system.cli.version: want the example's two lines, got %q", got)
		}
	})
	t.Run("derived: in the root, nothing on stderr when everything fits", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.2.0"))
		r.version("0.2.0").reports(t, expect{})
	})
	t.Run("derived: one line per file in the byte order of their paths, then the manifest", func(t *testing.T) {
		t.Parallel()
		// One file of each kind: missing, more than one stamp, no Proviso
		// version, installed by another version; and a manifest of
		// another model version.
		tr := installedAs("0.2.0").without(
			".claude/agents/proviso-coder.md",
		).with(
			".claude/agents/proviso-checker.md", "---\nname: proviso-checker\nmetadata:\n  proviso-version: 0.2.0\n  proviso-version: 0.2.0\n---\n",
			".claude/skills/proviso-accept/SKILL.md", "---\nname: proviso-accept\n---\n",
			".claude/skills/proviso-integrate/SKILL.md", stampedAs(".claude/skills/proviso-integrate/SKILL.md", "0.1.0"),
			"proviso/manifest.sum", "model 7\n",
		)
		r := versionRepo(t, tr)
		o := r.version("0.2.0")
		got := o.reports(t, expect{
			files: []string{
				".claude/agents/proviso-checker.md",
				".claude/agents/proviso-coder.md",
				".claude/skills/proviso-accept/SKILL.md",
				".claude/skills/proviso-integrate/SKILL.md",
			},
			manifest: true,
		})
		if l, ok := got[".claude/skills/proviso-integrate/SKILL.md"]; ok {
			staleLine(t, l, ".claude/skills/proviso-integrate/SKILL.md", "0.1.0", "0.2.0", "0")
		}
		if nums := numbers(o.manifestLine()); count(nums, "7") == 0 {
			t.Errorf("install.version.manifest-other: want the last line to name model 7 as the one the tree needs: %q", o.manifestLine())
		}
	})
}

func TestInstallVersionOutput(t *testing.T) {
	t.Parallel()
	outside := func(t *testing.T) string {
		dir := filepath.Join(t.TempDir(), "notes")
		tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
		return dir
	}
	t.Run("row1 none, built as release 0.1.0", func(t *testing.T) {
		t.Parallel()
		dir := outside(t)
		versionIn(t, "0.1.0", dir, dir, nil)
	})
	t.Run("row2 none, built locally after release 0.1.0", func(t *testing.T) {
		t.Parallel()
		dir := outside(t)
		versionIn(t, "0.1.1-dev", dir, dir, nil)
	})
	t.Run("row3 none, built with no version given", func(t *testing.T) {
		t.Parallel()
		dir := outside(t)
		versionIn(t, "dev", dir, dir, nil)
	})
	t.Run("row4 none, in a repository root whose skills are stale", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		o := r.version("0.1.0")
		if len(o.lines) == 0 {
			t.Errorf("install.version.output: want warnings on stderr about the stale skills, got none")
		}
	})
	for _, c := range []struct{ name, arg string }{
		{"row5 --json", "--json"},
		{"row6 0.1.0", "0.1.0"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// In a root whose skills are stale, so that anything checked
			// would show.
			r := freshRepo(t, "repo", false)
			r.init().exits(t, 0)
			o := versionIn(t, "0.1.0", r.dir, r.dir, nil, c.arg)
			for _, l := range o.lines {
				if n := names(l); len(n) != 0 {
					t.Errorf("system.cli.version: an argument was given, so nothing is checked, but a line names %v: %q", n, l)
				}
			}
		})
	}
}

func TestInstallVersionRepository(t *testing.T) {
	t.Parallel()
	// Before init, a root gives two lines: nothing installed, no manifest;
	// the hint is one line.
	t.Run("row1 repo/, a clone with a .git folder", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		r.version("dev").reports(t, expect{none: true, manifest: true})
	})
	t.Run("row2 wt/, a worktree with a .git file", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		wt := filepath.Join(t.TempDir(), "wt")
		r.git("worktree", "add", "-q", "--detach", wt)
		if fi, err := os.Lstat(filepath.Join(wt, ".git")); err != nil || !fi.Mode().IsRegular() {
			t.Fatalf("the worktree has no .git file: %v", err)
		}
		versionIn(t, "dev", wt, wt, nil).reports(t, expect{none: true, manifest: true})
	})
	t.Run("row3 repo/src/, inside a clone", func(t *testing.T) {
		t.Parallel()
		// The root above holds stale files; looking upward would name them.
		r := freshRepo(t, "repo", true)
		r.set(tree{"src/main.go": "package main\n"})
		r.init().exits(t, 0)
		versionIn(t, "0.1.0", r.dir, filepath.Join(r.dir, "src"), nil).hint(t)
	})
	t.Run("row4 notes/, in no repository", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "notes")
		tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
		versionIn(t, "dev", dir, dir, nil).hint(t)
	})
	t.Run("derived: no repository below is looked for", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "work")
		tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
		installedAs("0.1.0").with(".git/HEAD", "ref: refs/heads/main\n").writeTo(t, filepath.Join(dir, "repo"), false)
		versionIn(t, "dev", dir, dir, nil).hint(t)
	})
	t.Run("derived: a .git file is enough, as an entry", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "repo")
		tree{".git": "gitdir: elsewhere\n"}.writeTo(t, dir, false)
		versionIn(t, "dev", dir, dir, nil).reports(t, expect{none: true, manifest: true})
	})
}

func TestInstallVersionKit(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, after init", func(t *testing.T) {
		t.Parallel()
		// Installed by dev and checked by 0.2.0: each is named, so each was
		// checked.
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		got := r.version("0.2.0").reports(t, expect{files: stampedFiles, manifest: true})
		for f, l := range got {
			staleLine(t, l, f, "dev", "0.2.0", "")
		}
	})
	t.Run("row2 repo/, holding only .claude/skills/proviso-build/SKILL.md", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, tree{buildSkill: stampedAs(buildSkill, "0.2.0"), "proviso/manifest.sum": "model 0\n"})
		var missing []string
		for _, f := range stampedFiles {
			if f != buildSkill {
				missing = append(missing, f)
			}
		}
		r.version("0.2.0").reports(t, expect{files: missing})
	})
	t.Run("row3 repo/, before init", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.version("dev").reports(t, expect{none: true, manifest: true})
	})
	t.Run("row4 repo/, with .claude/ but none of the paths installed_files names", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, tree{
			".claude/settings.json":          "{}\n",
			".claude/agents/reviewer.md":     stampedAs(".claude/agents/reviewer.md", "0.1.0"),
			".claude/skills/deploy/SKILL.md": stampedAs(".claude/skills/deploy/SKILL.md", "0.1.0"),
			"proviso/manifest.sum":           "model 0\n",
		})
		r.version("dev").reports(t, expect{none: true})
	})
	t.Run("derived: a folder at one of the paths counts as an entry", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, tree{"proviso/manifest.sum": "model 0\n"})
		if err := os.MkdirAll(filepath.Join(r.dir, filepath.FromSlash(buildSkill)), 0o755); err != nil {
			t.Fatal(err)
		}
		r.version("dev").reports(t, expect{files: stampedFiles})
	})
}

func TestInstallVersionManifestMissing(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, before init", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.version("dev").reports(t, expect{none: true, manifest: true})
	})
	t.Run("row2 repo/, after init, before the first acceptance", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		o := r.version("dev")
		o.reports(t, expect{manifest: true})
		if l := o.manifestLine(); !strings.Contains(l, "manifest.sum") {
			t.Errorf("install.version.manifest-missing: want the line to name proviso/manifest.sum: %q", l)
		}
	})
	t.Run("row3 repo/, with an accepted tree", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		r.version("dev").reports(t, expect{none: true})
	})
}

func TestInstallVersionManifestUnreadable(t *testing.T) {
	t.Parallel()
	t.Run("row1 manifest.sum, without read permission", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev"))
		path := filepath.Join(r.dir, "proviso", "manifest.sum")
		undo, err := makeUnreadable(path)
		if err != nil {
			t.Skipf("cannot make a file unreadable here: %v", err)
		}
		undo()
		o := versionIn(t, "dev", r.dir, r.dir, func() func() {
			undo, err := makeUnreadable(path)
			if err != nil {
				t.Fatal(err)
			}
			return undo
		})
		o.reports(t, expect{manifest: true})
		if l := o.manifestLine(); !strings.Contains(l, "manifest.sum") {
			t.Errorf("install.version.manifest-unreadable: want the line to name proviso/manifest.sum: %q", l)
		}
	})
	t.Run("row2 manifest.sum, a folder", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev").without("proviso/manifest.sum"))
		if err := os.MkdirAll(filepath.Join(r.dir, "proviso", "manifest.sum"), 0o755); err != nil {
			t.Fatal(err)
		}
		o := r.version("dev")
		o.reports(t, expect{manifest: true})
		if l := o.manifestLine(); !strings.Contains(l, "manifest.sum") {
			t.Errorf("install.version.manifest-unreadable: want the line to name proviso/manifest.sum: %q", l)
		}
	})
	t.Run("row3 manifest.sum, a readable file", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev"))
		r.version("dev").reports(t, expect{})
	})
}

func TestInstallVersionManifestModel(t *testing.T) {
	t.Parallel()
	rows := []struct {
		name, content string
		holds         bool
	}{
		{"row1 model 0", "model 0\n", true},
		{"row2 model 0, ending in a carriage return and a line feed", "model 0\r\n", true},
		{"row3 model 1", "model 1\n", false},
		{"row4 empty", "", false},
		{"row5 model 00", "model 00\n", false},
		{"row6 model 0, a carriage return and foo", "model 0\rfoo\n", false},
		{"derived: model 0 as the last line, without a line ending", "model 0", true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := versionRepo(t, installedAs("dev").with("proviso/manifest.sum", row.content))
			// Handed over, the line is install.version.manifest-other's.
			r.version("dev").reports(t, expect{manifest: !row.holds})
		})
	}
	t.Run("derived: only the first line counts", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev").with("proviso/manifest.sum", "model 0\nnot a manifest line\n<<<<<<< HEAD\n"))
		r.version("dev").reports(t, expect{})
	})
}

func TestInstallVersionManifestOther(t *testing.T) {
	t.Parallel()
	rows := []struct {
		name, line string
		needs      string // the model version the tree needs, "" where none is named
		not        string // a number the line holds that must not be named as needed
	}{
		{"row1 model 1", "model 1", "1", ""},
		{"row2 model 12", "model 12", "12", ""},
		{"row3 empty", "", "", ""},
		{"row4 model", "model", "", ""},
		{"row5 model one", "model one", "", ""},
		{"row6 model 00", "model 00", "", ""},
		{"row7 model 1, followed by two carriage returns and a line feed", "model 1\r\r", "", "1"},
		{"row8 model 0, a carriage return and foo", "model 0\rfoo", "", ""},
		{"row9 <<<<<<< HEAD", "<<<<<<< HEAD", "", ""},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := versionRepo(t, installedAs("dev").with("proviso/manifest.sum", row.line+"\nlint.run.exit\theading\th1:00\n"))
			o := r.version("dev")
			o.reports(t, expect{manifest: true})
			nums := numbers(o.manifestLine(), "manifest.sum")
			if row.needs != "" {
				if count(nums, row.needs) == 0 || count(nums, modelVersion) == 0 {
					t.Errorf("install.version.manifest-other: want the line to name model %s as the one the tree needs and model %s as the one this proviso reads: %q", row.needs, modelVersion, o.manifestLine())
				}
				return
			}
			// Names no model version: never "needs 0, reads 0".
			if count(nums, modelVersion) > 1 {
				t.Errorf("install.version.manifest-other: %q names no model version, but the line names model %s as needed: %q", row.line, modelVersion, o.manifestLine())
			}
			if row.not != "" && count(nums, row.not) > 0 && count(nums, modelVersion) > 0 {
				t.Errorf("install.version.manifest-other: %q names no model version, but the line names model %s as needed and %s as read: %q", row.line, row.not, modelVersion, o.manifestLine())
			}
			if !strings.Contains(o.manifestLine(), "manifest.sum") {
				t.Errorf("install.version.manifest-other: want the line to say that proviso/manifest.sum names no model version: %q", o.manifestLine())
			}
		})
	}
}

func TestInstallVersionMissing(t *testing.T) {
	t.Parallel()
	coder := ".claude/agents/proviso-coder.md"
	t.Run("row1 .claude/agents/proviso-coder.md, deleted after init", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		if err := os.Remove(filepath.Join(r.dir, filepath.FromSlash(coder))); err != nil {
			t.Fatal(err)
		}
		r.version("dev").reports(t, expect{files: []string{coder}, manifest: true})
	})
	t.Run("row2 .claude/agents/proviso-coder.md, in place", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		r.version("dev").reports(t, expect{manifest: true})
	})
	t.Run("row3 .claude/agents/proviso-coder.md, a folder of that name", func(t *testing.T) {
		t.Parallel()
		// Handed over: install.version.unstamped names it.
		r := versionRepo(t, installedAs("dev").without(coder))
		if err := os.MkdirAll(filepath.Join(r.dir, filepath.FromSlash(coder)), 0o755); err != nil {
			t.Fatal(err)
		}
		r.version("dev").reports(t, expect{files: []string{coder}})
	})
	t.Run("derived: every file but one deleted", func(t *testing.T) {
		t.Parallel()
		tr := tree{"proviso/manifest.sum": "model 0\n", coder: stampedAs(coder, "dev")}
		var missing []string
		for _, f := range stampedFiles {
			if f != coder {
				missing = append(missing, f)
			}
		}
		versionRepo(t, tr).version("dev").reports(t, expect{files: missing})
	})
}

// skillAs is the build skill holding the frontmatter lines fm.
func skillAs(fm ...string) string {
	return "---\nname: proviso-build\n" + strings.Join(fm, "\n") + "\n---\n\nInstructions.\n"
}

func TestInstallVersionStamps(t *testing.T) {
	t.Parallel()
	// Checked by 0.1.0 among files stamped 0.1.0: a file read by its first
	// stamp, or by any one stamp 0.1.0, would not be named.
	rows := []struct {
		name, content string
		holds         bool
	}{
		{"row1 two lines proviso-version: 0.1.0 and proviso-version: 0.2.0 below metadata:",
			skillAs("metadata:", "  proviso-version: 0.1.0", "  proviso-version: 0.2.0"), true},
		{"row2 two metadata: blocks, each stamped 0.1.0",
			skillAs("metadata:", "  proviso-version: 0.1.0", "description: Builds.", "metadata:", "  proviso-version: 0.1.0"), true},
		{"row3 the lines proviso-version: 0.1.0 and proviso-version: below metadata:",
			skillAs("metadata:", "  proviso-version: 0.1.0", "  proviso-version:"), true},
		{"row4 stamped 0.1.0 once", stampedAs(buildSkill, "0.1.0"), false},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := versionRepo(t, installedAs("0.1.0").with(buildSkill, row.content))
			var files []string
			if row.holds {
				files = []string{buildSkill}
			}
			r.version("0.1.0").reports(t, expect{files: files})
		})
	}
	t.Run("row5 without read permission", func(t *testing.T) {
		t.Parallel()
		// Handed over: install.version.unstamped names it.
		r := versionRepo(t, installedAs("0.1.0"))
		path := filepath.Join(r.dir, filepath.FromSlash(buildSkill))
		undo, err := makeUnreadable(path)
		if err != nil {
			t.Skipf("cannot make a file unreadable here: %v", err)
		}
		undo()
		o := versionIn(t, "0.1.0", r.dir, r.dir, func() func() {
			undo, err := makeUnreadable(path)
			if err != nil {
				t.Fatal(err)
			}
			return undo
		})
		o.reports(t, expect{files: []string{buildSkill}})
	})
	t.Run("derived: two stamps in an agent definition", func(t *testing.T) {
		t.Parallel()
		tester := ".claude/agents/proviso-tester.md"
		r := versionRepo(t, installedAs("0.1.0").with(tester, "---\nname: proviso-tester\nmetadata:\n  proviso-version: 0.1.0\n  proviso-version: 0.1.0\n---\n"))
		r.version("0.1.0").reports(t, expect{files: []string{tester}})
	})
}

func TestInstallVersionUnstamped(t *testing.T) {
	t.Parallel()
	// Checked by 0.1.0 among files stamped 0.1.0. A row handed over with
	// the stamp 0.1.0 gives no line; with another stamp, the line of
	// install.version.stale.
	rows := []struct {
		name, content string
		holds         bool
		stamp         string // the stamp handed over, where it is not 0.1.0
	}{
		{"row1 installed before stamps existed", "---\nname: proviso-build\ndescription: Builds an amendment.\n---\n\nInstructions.\n", true, ""},
		{"row3 metadata: {proviso-version: 0.1.0} on one line", skillAs("metadata: {proviso-version: 0.1.0}"), true, ""},
		{"row4 proviso-version: 0.1.0 at the top, not below metadata:", skillAs("proviso-version: 0.1.0"), true, ""},
		{"row5 proviso-version: below metadata:, with no version", skillAs("metadata:", "  proviso-version:"), true, ""},
		{"row6 its stamp below an indented metadata:, nested under another key", skillAs("config:", "  metadata:", "    proviso-version: 0.1.0"), true, ""},
		{"row7 a blank line between metadata: and its stamp", skillAs("metadata:", "", "  proviso-version: 0.1.0"), true, ""},
		{"row8 a first line --- and no later line ---", "---\nname: proviso-build\nmetadata:\n  proviso-version: 0.1.0\n\nInstructions.\n", true, ""},
		{"row9 stamped [1]", stampedAs(buildSkill, "[1]"), true, ""},
		{"row10 stamped v0.1.0", stampedAs(buildSkill, "v0.1.0"), true, ""},
		{"row11 stamped 0.1", stampedAs(buildSkill, "0.1"), true, ""},
		{"row12 with no frontmatter", "metadata:\n  proviso-version: 0.1.0\n\n# Build\n\nInstructions.\n", true, ""},
		{"row13 metadata without proviso-version", skillAs("metadata:", "  author: the engineers"), true, ""},
		{"row14 stamped 0.1.0", stampedAs(buildSkill, "0.1.0"), false, ""},
		{"row15 stamped 0.1.0, with carriage returns and line feeds", strings.ReplaceAll(stampedAs(buildSkill, "0.1.0"), "\n", "\r\n"), false, ""},
		{"row16 stamped 0.1.1-dev", stampedAs(buildSkill, "0.1.1-dev"), false, "0.1.1-dev"},
		{"row17 stamped dev", stampedAs(buildSkill, "dev"), false, "dev"},
		{"row18 an indented comment line between metadata: and its stamp 0.1.0", skillAs("metadata:", "  # installed by proviso", "  proviso-version: 0.1.0"), false, ""},
		{"row19 stamped 0.1.0 with a tab as indentation, no space after the colon and spaces at the end", skillAs("metadata:", "\tproviso-version:0.1.0   "), false, ""},
		// install.version-stamp: a line is read without its line ending
		// only; a carriage return before it belongs to the line, and is no
		// space or tab, so the version is 0.1.0 and a carriage return.
		{"derived: stamped 0.1.0, its line followed by two carriage returns and a line feed", strings.Replace(stampedAs(buildSkill, "0.1.0"), "0.1.0\n", "0.1.0\r\r\n", 1), true, ""},
		{"derived: spaces and tabs around both colons", skillAs("metadata \t: ", " \t proviso-version \t:\t 0.1.0"), false, ""},
		{"derived: a stamp among other lines below metadata:", skillAs("metadata:", "  author: the engineers", "  proviso-version: 0.1.0", "  team: payments"), false, ""},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := versionRepo(t, installedAs("0.1.0").with(buildSkill, row.content))
			o := r.version("0.1.0")
			switch {
			case row.holds:
				o.reports(t, expect{files: []string{buildSkill}})
			case row.stamp == "":
				o.reports(t, expect{})
			default:
				if l, ok := o.reports(t, expect{files: []string{buildSkill}})[buildSkill]; ok {
					staleLine(t, l, buildSkill, row.stamp, "0.1.0", "")
				}
			}
		})
	}
	t.Run("row2 a folder of that name", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.1.0").without(buildSkill))
		if err := os.MkdirAll(filepath.Join(r.dir, filepath.FromSlash(buildSkill)), 0o755); err != nil {
			t.Fatal(err)
		}
		r.version("0.1.0").reports(t, expect{files: []string{buildSkill}})
	})
}

func TestInstallVersionStale(t *testing.T) {
	t.Parallel()
	tester := ".claude/agents/proviso-tester.md"
	rows := []struct {
		name, stamp, version string
		holds                bool
		model                string // the model version named, "" for none
	}{
		{"row1 0.1.0 by 0.2.0", "0.1.0", "0.2.0", true, "0"},
		{"row2 dev by 0.2.0", "dev", "0.2.0", true, ""},
		{"row3 0.1.0 by 0.1.1-dev", "0.1.0", "0.1.1-dev", true, "0"},
		{"row4 0.1.0+build.5 by 0.1.0", "0.1.0+build.5", "0.1.0", true, ""},
		{"row5 0.3.0 by 0.2.0", "0.3.0", "0.2.0", true, ""},
		{"row6 0.2.0 by 0.2.0", "0.2.0", "0.2.0", false, ""},
		{"row7 dev by dev", "dev", "dev", false, ""},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := versionRepo(t, installedAs(row.version).with(tester, stampedAs(tester, row.stamp)))
			o := r.version(row.version)
			if !row.holds {
				o.reports(t, expect{})
				return
			}
			if l, ok := o.reports(t, expect{files: []string{tester}})[tester]; ok {
				staleLine(t, l, tester, row.stamp, row.version, row.model)
			}
		})
	}
}

func TestInstallVersionModel(t *testing.T) {
	t.Parallel()
	tester := ".claude/agents/proviso-tester.md"
	rows := []struct {
		name, version, stamp string
		model                string // "" for no model
		table                bool   // whether the row's model_releases is the binary's
	}{
		{"row1 0 from 0.1.0, 0.2.0, 0.1.0", "0.2.0", "0.1.0", "0", true},
		{"row2 0 from 0.1.0, 1 from 0.3.0, 0.4.0, 0.2.5", "0.4.0", "0.2.5", "0", false},
		{"row3 0 from 0.1.0, 1 from 0.3.0, 0.4.0, 0.3.0", "0.4.0", "0.3.0", "1", false},
		{"row4 0 from 0.1.0, 1 from 0.3.0, 0.3.1-dev, 0.3.0", "0.3.1-dev", "0.3.0", "1", false},
		{"row5 0 from 0.1.0, 1 from 0.3.0, dev, 0.4.2", "dev", "0.4.2", "1", false},
		{"row6 0 from 0.1.0, 0.2.0, dev", "0.2.0", "dev", "", true},
		{"row7 0 from 0.1.0, 0.2.0, 0.1.1-dev", "0.2.0", "0.1.1-dev", "", true},
		{"row8 0 from 0.1.0, 0.2.0, 0.1.0+build.5", "0.2.0", "0.1.0+build.5", "", true},
		{"row9 0 from 0.1.0, 0.2.0, 0.3.0", "0.2.0", "0.3.0", "", true},
		{"row10 0 from 0.1.0, 0.1.1-dev, 0.1.1", "0.1.1-dev", "0.1.1", "", true},
		{"row11 0 from 0.1.0, 0.2.0, 0.0.9", "0.2.0", "0.0.9", "", true},
		// Analogues on the binary's table, 0 from 0.1.0.
		{"derived: 0 from 0.1.0, 0.4.0, 0.2.5: the row holds until the next one", "0.4.0", "0.2.5", "0", true},
		{"derived: 0 from 0.1.0, 0.1.1-dev, 0.1.0: a local build after 0.1.0 is newer than 0.1.0", "0.1.1-dev", "0.1.0", "0", true},
		{"derived: 0 from 0.1.0, dev, 0.4.2: the last row holds for every later release", "dev", "0.4.2", "0", true},
		{"derived: 0 from 0.1.0, dev, 0.1.0: the first release is its row's own", "dev", "0.1.0", "0", true},
		{"derived: 0 from 0.1.0, 0.2.0, 0.2.0-rc.1: a pre-release is no release", "0.2.0", "0.2.0-rc.1", "", true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if !row.table {
				t.Skip("model_releases is not configurable: the binary holds 0 from 0.1.0 only; see the derived analogue")
			}
			r := versionRepo(t, installedAs(row.version).with(tester, stampedAs(tester, row.stamp)))
			if l, ok := r.version(row.version).reports(t, expect{files: []string{tester}})[tester]; ok {
				staleLine(t, l, tester, row.stamp, row.version, row.model)
			}
		})
	}
}

// TestInstallVersionReadsOnly checks install.version.reads-only, a property
// checked by versionIn for every run of version; here on runs that end in
// every way.
func TestInstallVersionReadsOnly(t *testing.T) {
	t.Parallel()
	t.Run("derived: after init, every file stale", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		r.version("0.2.0")
	})
	t.Run("derived: files of every kind, with carriage returns", func(t *testing.T) {
		t.Parallel()
		tr := installedAs("0.2.0").without(".claude/agents/proviso-coder.md").with(
			".claude/agents/proviso-checker.md", "---\r\nname: proviso-checker\r\nmetadata:\r\n  proviso-version: 0.1.0\r\n  proviso-version: 0.2.0\r\n---\r\n",
			buildSkill, "no frontmatter\n",
			"proviso/manifest.sum", "model 3\r\n",
			"src/main.go", "package main\r\n",
		)
		versionRepo(t, tr).version("0.2.0")
	})
	t.Run("derived: an unreadable skill", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev"))
		path := filepath.Join(r.dir, filepath.FromSlash(buildSkill))
		undo, err := makeUnreadable(path)
		if err != nil {
			t.Skipf("cannot make a file unreadable here: %v", err)
		}
		undo()
		versionIn(t, "dev", r.dir, r.dir, func() func() {
			undo, err := makeUnreadable(path)
			if err != nil {
				t.Fatal(err)
			}
			return undo
		})
	})
	t.Run("derived: an argument", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev"))
		versionIn(t, "dev", r.dir, r.dir, nil, "--help")
	})
	t.Run("derived: inside a repository", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("dev").with("src/main.go", "package main\n"))
		versionIn(t, "dev", r.dir, filepath.Join(r.dir, "src"), nil)
	})
}
