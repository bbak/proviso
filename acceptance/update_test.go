package acceptance

// The feature install.update: the command `proviso update` of
// install.cli.update. What holds for every run of update (its output, exit
// status, install.update.keeps, and that a file named as written or updated
// is as init of the binary's version writes it) is checked by updateIn.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const coderAgent = ".claude/agents/proviso-coder.md"

// mkdir makes the folder at the slash-separated path p below dir.
func mkdir(t *testing.T, dir, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(p)), 0o755); err != nil {
		t.Fatal(err)
	}
}

// remove removes the file at the slash-separated path p below dir.
func remove(t *testing.T, dir, p string) {
	t.Helper()
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(p))); err != nil {
		t.Fatal(err)
	}
}

// initialised is a fresh repository after init of the binary given no
// version, whose skills and agents are stamped dev.
func initialised(t *testing.T) *repo {
	t.Helper()
	r := freshRepo(t, "repo", false)
	r.init().exits(t, 0)
	return r
}

// TestInstallCliUpdate checks the contract install.cli.update and its
// example: one line per skill or agent definition, in the byte order of
// their paths.
func TestInstallCliUpdate(t *testing.T) {
	t.Parallel()
	t.Run("example: updated .claude/agents/proviso-coder.md, skipped .claude/skills/proviso-accept/SKILL.md", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.2.0").with(coderAgent, stampedAs(coderAgent, "0.1.0")))
		o := r.update("0.2.0")
		o.exits(t, 0)
		want := "skipped .claude/agents/proviso-checker.md\n" +
			"updated .claude/agents/proviso-coder.md\n" +
			"skipped .claude/agents/proviso-tester.md\n" +
			"skipped .claude/skills/proviso-accept/SKILL.md\n" +
			"skipped .claude/skills/proviso-build/SKILL.md\n" +
			"skipped .claude/skills/proviso-check/SKILL.md\n" +
			"skipped .claude/skills/proviso-integrate/SKILL.md\n"
		if got := strings.ReplaceAll(o.stdout, "\r\n", "\n"); got != want {
			t.Errorf("install.cli.update: want stdout\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("derived: written, updated and skipped in one run, in the byte order of the paths", func(t *testing.T) {
		t.Parallel()
		tr := installedAs("0.2.0").without(".claude/agents/proviso-tester.md").with(buildSkill, stampedAs(buildSkill, "0.1.0"))
		o := versionRepo(t, tr).update("0.2.0")
		o.exits(t, 0)
		o.says(t, "written", ".claude/agents/proviso-tester.md")
		o.says(t, "updated", buildSkill)
		o.says(t, "skipped", ".claude/skills/proviso-check/SKILL.md")
		if len(o.lines) != len(stampedFiles) {
			t.Errorf("install.cli.update: want one line per skill and agent definition, got:\n%s", o.stdout)
		}
	})
	t.Run("derived: it takes no argument", func(t *testing.T) {
		t.Parallel()
		r := initialised(t)
		o := updateIn(t, "0.2.0", r.dir, r.dir, "--json")
		o.refused(t, "cli.help.argument", true)
		misused(t, o.run, "update", "--json")
	})
}

func TestInstallCliUpdateExitStatus(t *testing.T) {
	t.Parallel()
	t.Run("0: every skill and agent definition in place, written, updated or skipped", func(t *testing.T) {
		t.Parallel()
		tr := installedAs("0.2.0").without(coderAgent).with(buildSkill, stampedAs(buildSkill, "0.1.0"))
		versionRepo(t, tr).update("0.2.0").exits(t, 0)
	})
	t.Run("2: the binary is no release", func(t *testing.T) {
		t.Parallel()
		initialised(t).update("dev").exits(t, 2)
	})
	t.Run("2: the directory is not the root of a git repository", func(t *testing.T) {
		t.Parallel()
		r := initialised(t)
		r.set(tree{"src/main.go": "package main\n"})
		updateIn(t, "0.2.0", r.dir, filepath.Join(r.dir, "src")).exits(t, 2)
	})
	t.Run("2: no skill or agent definition is installed", func(t *testing.T) {
		t.Parallel()
		freshRepo(t, "repo", false).update("0.2.0").exits(t, 2)
	})
	t.Run("2: an argument was given", func(t *testing.T) {
		t.Parallel()
		r := initialised(t)
		updateIn(t, "0.2.0", r.dir, r.dir, "0.2.0").exits(t, 2)
	})
	t.Run("2: a file could not be written", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.1.0").without(buildSkill))
		mkdir(t, r.dir, buildSkill)
		r.update("0.2.0").exits(t, 2)
	})
}

func TestInstallUpdateExit(t *testing.T) {
	t.Parallel()
	t.Run("row1 none, every file updated", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.1.0"))
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.all(t, "updated")
		// install.update: once update has run, version names no file as
		// installed by another version.
		r.version("0.2.0").reports(t, expect{})
	})
	t.Run("row2 none, every file skipped, some as newer", func(t *testing.T) {
		t.Parallel()
		tester := ".claude/agents/proviso-tester.md"
		tr := installedAs("0.2.0").with(tester, stampedAs(tester, "0.3.0"), buildSkill, stampedAs(buildSkill, "0.3.0"))
		o := versionRepo(t, tr).update("0.2.0")
		o.exits(t, 0)
		o.all(t, "skipped")
		o.newer(t, tester, "0.3.0")
		o.newer(t, buildSkill, "0.3.0")
	})
	t.Run("row3 .claude/skills/proviso-build/SKILL.md, a folder of that name", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, tree{})
		mkdir(t, r.dir, buildSkill)
		o := r.update("0.2.0")
		o.failed(t, buildSkill)
		// The other files are written, and stay.
		if len(o.lines) != len(stampedFiles)-1 {
			t.Errorf("install.update.exit: want the other files named as written, got:\n%s", o.stdout)
		}
		for _, f := range stampedFiles {
			if f != buildSkill {
				o.says(t, "written", f)
			}
		}
	})
	t.Run("derived: two files fail, the others are updated", func(t *testing.T) {
		t.Parallel()
		tester := ".claude/agents/proviso-tester.md"
		r := versionRepo(t, installedAs("0.1.0").without(buildSkill, tester))
		mkdir(t, r.dir, buildSkill)
		mkdir(t, r.dir, tester)
		o := r.update("0.2.0")
		o.failed(t, buildSkill, tester)
		if len(o.lines) != len(stampedFiles)-2 {
			t.Errorf("install.update.exit: want the other files named as updated, got:\n%s", o.stdout)
		}
		for _, f := range stampedFiles {
			if f != buildSkill && f != tester {
				o.says(t, "updated", f)
			}
		}
	})
}

func TestInstallUpdateInstalled(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, after init", func(t *testing.T) {
		t.Parallel()
		// Installed by dev, so each one taken is updated.
		o := initialised(t).update("0.2.0")
		o.exits(t, 0)
		o.all(t, "updated")
	})
	t.Run("row2 repo/, holding only .claude/skills/proviso-build/SKILL.md", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, tree{buildSkill: stampedAs(buildSkill, "0.1.0")})
		o := r.update("0.2.0")
		o.exits(t, 0)
		for _, f := range stampedFiles {
			if f == buildSkill {
				o.says(t, "updated", f)
			} else {
				o.says(t, "written", f)
			}
		}
	})
	refusals := []struct {
		name string
		tr   tree
	}{
		{"row3 repo/, before init", tree{}},
		{"row4 repo/, with proviso/system.md and .proviso/build.md only", tree{"proviso/system.md": "# Shop\n", ".proviso/build.md": "# Notes\n"}},
		{"derived: repo/, with other agents and skills of its own below .claude/", tree{
			".claude/agents/reviewer.md":     stampedAs(".claude/agents/reviewer.md", "0.1.0"),
			".claude/skills/deploy/SKILL.md": stampedAs(".claude/skills/deploy/SKILL.md", "0.1.0"),
		}},
	}
	for _, c := range refusals {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			o := versionRepo(t, c.tr).update("0.2.0")
			o.refused(t, "install.update.installed", true)
			if len(o.stderrLines) == 1 && !strings.Contains(o.stderrLines[0], "init") {
				t.Errorf("install.update.installed: want the line to say that init installs them: %q", o.stderrLines[0])
			}
		})
	}
	t.Run("derived: a folder at one of the paths counts as an entry", func(t *testing.T) {
		t.Parallel()
		// Taken, so the others are written; the folder cannot be.
		r := versionRepo(t, tree{})
		mkdir(t, r.dir, coderAgent)
		o := r.update("0.2.0")
		o.failed(t, coderAgent)
		if len(o.lines) != len(stampedFiles)-1 {
			t.Errorf("install.update.installed: want every other file taken, got:\n%s", o.stdout)
		}
	})
}

// TestInstallUpdateKeeps checks install.update.keeps, a property checked by
// updateIn for every run of update; here on runs that end in every way, in
// repositories with much to keep.
func TestInstallUpdateKeeps(t *testing.T) {
	t.Parallel()
	own := tree{
		"README.md":                             "# Shop\n",
		".gitattributes":                        "* text=auto\n",
		".claude/settings.json":                 "{}\n",
		".claude/agents/reviewer.md":            "---\nname: reviewer\nmetadata:\n  proviso-version: 0.1.0\n---\n",
		".claude/skills/deploy/SKILL.md":        "---\nname: deploy\nmetadata:\n  proviso-version: 0.1.0\n---\n",
		".claude/skills/proviso-build/notes.md": "beside the skill\n",
		"src/main.go":                           "package main\r\n",
		"proviso/capabilities/cart/cart.md":     "# Cart\n",
	}
	t.Run("derived: after init, with the engineers' edits to every other installed file", func(t *testing.T) {
		t.Parallel()
		r := initialised(t)
		tr := own.clone()
		for _, q := range installedFiles {
			if !isStamped(q) {
				tr[q] = "edited by the engineers: " + q + "\n"
			}
		}
		tr.writeTo(t, r.dir, false)
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.all(t, "updated")
	})
	t.Run("derived: some missing, some newer", func(t *testing.T) {
		t.Parallel()
		tr := installedAs("0.1.0").without(coderAgent).with(buildSkill, stampedAs(buildSkill, "0.3.0"))
		for p, c := range own {
			tr[p] = c
		}
		versionRepo(t, tr).update("0.2.0").exits(t, 0)
	})
	t.Run("derived: a run that fails", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, own.clone())
		mkdir(t, r.dir, coderAgent)
		r.update("0.2.0").exits(t, 2)
	})
	t.Run("derived: a run that is refused", func(t *testing.T) {
		t.Parallel()
		tr := installedAs("0.1.0")
		for p, c := range own {
			tr[p] = c
		}
		versionRepo(t, tr).update("dev").exits(t, 2)
	})
}

func TestInstallUpdateMissing(t *testing.T) {
	t.Parallel()
	t.Run("row1 .claude/agents/proviso-coder.md, deleted after init", func(t *testing.T) {
		t.Parallel()
		r := initialised(t)
		remove(t, r.dir, coderAgent)
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.says(t, "written", coderAgent)
	})
	t.Run("row2 .claude/agents/proviso-coder.md, .claude being a file", func(t *testing.T) {
		t.Parallel()
		// With .claude a file, no entry exists at any skill or agent
		// path, so install.update.installed refuses before the file is
		// taken; either way it is not written, with exit status 2.
		r := versionRepo(t, tree{".claude": "not a folder\n"})
		o := r.update("0.2.0")
		o.exits(t, 2)
		for _, l := range o.lines {
			if l.path == coderAgent {
				t.Errorf("install.update.missing: %s could not be written, but stdout names it as %s", coderAgent, l.status)
			}
		}
	})
	t.Run("derived: .claude/agents being a file, the skills in place", func(t *testing.T) {
		t.Parallel()
		tr := installedAs("0.1.0")
		var agents []string
		for _, f := range stampedFiles {
			if strings.HasPrefix(f, ".claude/agents/") {
				agents = append(agents, f)
				delete(tr, f)
			}
		}
		tr[".claude/agents"] = "not a folder\n"
		o := versionRepo(t, tr).update("0.2.0")
		o.failed(t, agents...)
		for _, f := range stampedFiles[len(agents):] {
			o.says(t, "updated", f)
		}
	})
	t.Run("row3 .claude/agents/proviso-coder.md, in place", func(t *testing.T) {
		t.Parallel()
		// Handed over: install.update.stamp updates it.
		r := versionRepo(t, installedAs("0.2.0").with(coderAgent, stampedAs(coderAgent, "0.1.0")))
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.says(t, "updated", coderAgent)
	})
	t.Run("row4 .claude/agents/proviso-coder.md, a folder of that name", func(t *testing.T) {
		t.Parallel()
		// Handed over: install.update.stamp cannot read it as a file, and
		// removes no entry.
		r := versionRepo(t, installedAs("0.2.0").without(coderAgent))
		mkdir(t, r.dir, coderAgent)
		r.update("0.2.0").failed(t, coderAgent)
	})
	t.Run("derived: every file missing but one, with the folders above them", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, tree{".claude/skills/proviso-accept/SKILL.md": stampedAs(".claude/skills/proviso-accept/SKILL.md", "0.2.0")})
		o := r.update("0.2.0")
		o.exits(t, 0)
		for _, f := range stampedFiles {
			if f == ".claude/skills/proviso-accept/SKILL.md" {
				o.says(t, "skipped", f)
			} else {
				o.says(t, "written", f)
			}
		}
	})
}

func TestInstallUpdateRelease(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, version string }{
		{"row1 0.1.0", "0.1.0"},
		{"row2 0.2.0", "0.2.0"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// Installed by dev: every file is written anew, stamped with
			// the version, which updateIn checks.
			o := initialised(t).update(c.version)
			o.exits(t, 0)
			o.all(t, "updated")
		})
	}
	for _, c := range []struct{ name, version string }{
		{"row3 0.1.1-dev", "0.1.1-dev"},
		{"row4 dev", "dev"},
		{"row5 0.2.0-rc.1", "0.2.0-rc.1"},
		{"row6 0.2.0+build.5", "0.2.0+build.5"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// After init, with a file missing and one stale, so that a run
			// that went on would write.
			r := initialised(t)
			remove(t, r.dir, coderAgent)
			o := r.update(c.version)
			o.refused(t, "install.update.release", true)
			if len(o.stderrLines) == 1 && !strings.Contains(o.stderrLines[0], c.version) {
				t.Errorf("install.update.release: want the line to name %s: %q", c.version, o.stderrLines[0])
			}
		})
	}
	t.Run("derived: a refused binary looks for no repository", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		// The version is judged before the directory.
		o := updateIn(t, "dev", dir, dir)
		o.refused(t, "install.update.release", true)
		if len(o.stderrLines) == 1 && !strings.Contains(o.stderrLines[0], "dev") {
			t.Errorf("install.update.release: want the line to name dev: %q", o.stderrLines[0])
		}
	})
}

func TestInstallUpdateRepository(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, a clone with a .git folder", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		installedAs("0.1.0").writeTo(t, r.dir, false)
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.all(t, "updated")
	})
	t.Run("row2 wt/, a worktree with a .git file", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		wt := filepath.Join(t.TempDir(), "wt")
		r.git("worktree", "add", "-q", "--detach", wt)
		if fi, err := os.Lstat(filepath.Join(wt, ".git")); err != nil || !fi.Mode().IsRegular() {
			t.Fatalf("the worktree has no .git file: %v", err)
		}
		installedAs("0.1.0").writeTo(t, wt, false)
		o := updateIn(t, "0.2.0", wt, wt)
		o.exits(t, 0)
		o.all(t, "updated")
	})
	t.Run("row3 repo/src/, inside a clone", func(t *testing.T) {
		t.Parallel()
		// The root above holds stale files; looking upward would update
		// them.
		r := freshRepo(t, "repo", true)
		installedAs("0.1.0").with("src/main.go", "package main\n").writeTo(t, r.dir, false)
		updateIn(t, "0.2.0", r.dir, filepath.Join(r.dir, "src")).refused(t, "install.update.repository", false)
	})
	t.Run("row4 notes/, in no repository", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		updateIn(t, "0.2.0", dir, dir).refused(t, "install.update.repository", false)
	})
	t.Run("derived: no repository below is looked for", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "work")
		tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
		installedAs("0.1.0").with(".git/HEAD", "ref: refs/heads/main\n").writeTo(t, filepath.Join(dir, "repo"), false)
		updateIn(t, "0.2.0", dir, dir).refused(t, "install.update.repository", false)
	})
	t.Run("derived: a .git file is enough, as an entry", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "repo")
		installedAs("0.1.0").with(".git", "gitdir: elsewhere\n").writeTo(t, dir, false)
		o := updateIn(t, "0.2.0", dir, dir)
		o.exits(t, 0)
		o.all(t, "updated")
	})
}

func TestInstallUpdateStamp(t *testing.T) {
	t.Parallel()
	// Updated by 0.2.0 among files stamped 0.2.0, which are skipped.
	rows := []struct {
		name, content string
		status        string // "skipped" or "updated"
		newer         string // the stamp named as newer on stderr, "" for none
	}{
		{"row1 stamped 0.2.0, by 0.2.0", stampedAs(buildSkill, "0.2.0"), "skipped", ""},
		{"row2 stamped 0.3.0, by 0.2.0", stampedAs(buildSkill, "0.3.0"), "skipped", "0.3.0"},
		{"row3 stamped 0.2.1-dev, by 0.2.0", stampedAs(buildSkill, "0.2.1-dev"), "skipped", "0.2.1-dev"},
		{"row4 stamped 0.1.0, by 0.2.0", stampedAs(buildSkill, "0.1.0"), "updated", ""},
		{"row5 stamped 0.2.0-dev, by 0.2.0", stampedAs(buildSkill, "0.2.0-dev"), "updated", ""},
		{"row6 stamped 0.2.0+build.5, by 0.2.0", stampedAs(buildSkill, "0.2.0+build.5"), "updated", ""},
		{"row7 stamped dev, by 0.2.0", stampedAs(buildSkill, "dev"), "updated", ""},
		{"row8 stamped v0.2.0, by 0.2.0", stampedAs(buildSkill, "v0.2.0"), "updated", ""},
		{"row9 stamped twice 0.2.0, by 0.2.0", skillAs("metadata:", "  proviso-version: 0.2.0", "  proviso-version: 0.2.0"), "updated", ""},
		{"row10 with no frontmatter, by 0.2.0", "# Build\n\nInstructions edited by hand.\n", "updated", ""},
		{"derived: stamped 0.2.0, with carriage returns and line feeds", strings.ReplaceAll(stampedAs(buildSkill, "0.2.0"), "\n", "\r\n"), "skipped", ""},
		{"derived: stamped 0.2.0 with a tab as indentation and spaces at the end", skillAs("metadata:", "\tproviso-version:0.2.0   "), "skipped", ""},
		{"derived: stamped 0.10.0, newer by precedence, not as text", stampedAs(buildSkill, "0.10.0"), "skipped", "0.10.0"},
		{"derived: stamped 1.0.0-alpha, a pre-release of a later version", stampedAs(buildSkill, "1.0.0-alpha"), "skipped", "1.0.0-alpha"},
		{"derived: stamped 0.3.0 below metadata: on one line, not read as a stamp", skillAs("metadata: {proviso-version: 0.3.0}"), "updated", ""},
		{"derived: stamped 0.3.0 and 0.3.0, two stamps of a newer version", skillAs("metadata:", "  proviso-version: 0.3.0", "  proviso-version: 0.3.0"), "updated", ""},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := versionRepo(t, installedAs("0.2.0").with(buildSkill, row.content))
			o := r.update("0.2.0")
			o.exits(t, 0)
			o.says(t, row.status, buildSkill)
			if row.newer != "" {
				o.newer(t, buildSkill, row.newer)
			} else {
				o.quietAbout(t, buildSkill)
			}
			for _, f := range stampedFiles {
				if f != buildSkill {
					o.says(t, "skipped", f)
				}
			}
		})
	}
	t.Run("row11 a folder of that name", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.2.0").without(buildSkill))
		mkdir(t, r.dir, buildSkill)
		tree{"keep.md": "inside the folder\n"}.writeTo(t, filepath.Join(r.dir, filepath.FromSlash(buildSkill)), false)
		// That the folder and what it holds stay, updateIn checks.
		r.update("0.2.0").failed(t, buildSkill)
	})
	t.Run("derived: a file updated holds one stamp, which version reads as the binary's", func(t *testing.T) {
		t.Parallel()
		r := versionRepo(t, installedAs("0.2.0").with(buildSkill, stampedAs(buildSkill, "0.1.0")))
		r.update("0.2.0").exits(t, 0)
		r.version("0.2.0").reports(t, expect{})
		// A second run skips every file.
		o := r.update("0.2.0")
		o.exits(t, 0)
		o.all(t, "skipped")
	})
}
