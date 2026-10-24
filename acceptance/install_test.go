package acceptance

// The install capability: proviso/capabilities/install/, and the command
// `proviso init` of contracts/cli.md. What holds for every run of init (its
// output, exit status and install.init.keeps) is checked by initIn.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestSystemCliInit checks the example of system.cli.init: the lines in the
// byte order of the paths, one per installed file.
func TestSystemCliInit(t *testing.T) {
	t.Parallel()
	r := freshRepo(t, "repo", false)
	r.set(tree{"proviso/system.md": "# Our own system\n"})
	o := r.init()
	o.exits(t, 0)
	want := "written .claude/agents/proviso-checker.md\n" +
		"written .claude/agents/proviso-coder.md\n" +
		"written .claude/agents/proviso-tester.md\n" +
		"written .claude/skills/proviso-accept/SKILL.md\n" +
		"written .claude/skills/proviso-build/SKILL.md\n" +
		"written .claude/skills/proviso-check/SKILL.md\n" +
		"written .proviso/build.md\n" +
		"written .proviso/coder.md\n" +
		"written .proviso/tester.md\n" +
		"written proviso/.gitattributes\n" +
		"written proviso/.markdownlint.jsonc\n" +
		"skipped proviso/system.md\n"
	if got := strings.ReplaceAll(o.stdout, "\r\n", "\n"); got != want {
		t.Errorf("system.cli.init: want stdout\n%s\ngot\n%s", want, got)
	}
}

// TestInstallInstalledFiles checks the rows of installed_files that the
// other tests do not: the skills that accept, build and check an amendment
// are written, and are no empty files. Their wording, and that of the stubs,
// is the kit's.
func TestInstallInstalledFiles(t *testing.T) {
	t.Parallel()
	r := freshRepo(t, "repo", false)
	o := r.init()
	o.exits(t, 0)
	skills := []string{
		".claude/skills/proviso-accept/SKILL.md",
		".claude/skills/proviso-build/SKILL.md",
		".claude/skills/proviso-check/SKILL.md",
	}
	for _, p := range append(skills, ".proviso/build.md") {
		o.says(t, "written", p)
	}
	for _, p := range skills {
		if strings.TrimSpace(r.read(p)) == "" {
			t.Errorf("install.installed-files: the skill %s is empty", p)
		}
	}
}

// guideLines is the fewest lines, blank ones not counted, that a block every
// agent definition shares must have to count as the reading guide.
const guideLines = 10

// sharedBlock gives the longest run of consecutive lines that a and b both
// hold, and the number of its lines that are not blank.
func sharedBlock(a, b string) (string, int) {
	la := strings.Split(strings.ReplaceAll(a, "\r\n", "\n"), "\n")
	lb := strings.Split(strings.ReplaceAll(b, "\r\n", "\n"), "\n")
	prev := make([]int, len(lb)+1)
	best, end := 0, 0
	for i := 1; i <= len(la); i++ {
		cur := make([]int, len(lb)+1)
		for j := 1; j <= len(lb); j++ {
			if la[i-1] == lb[j-1] {
				cur[j] = prev[j-1] + 1
				if cur[j] > best {
					best, end = cur[j], i
				}
			}
		}
		prev = cur
	}
	block := la[end-best : end]
	n := 0
	for _, l := range block {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return strings.Join(block, "\n"), n
}

var (
	linkTarget = regexp.MustCompile(`\]\(([^)\s]+)`)
	importPath = regexp.MustCompile(`(?m)(?:^|\s)@([\w./-]+)`)
)

func TestInstallInitAgents(t *testing.T) {
	t.Parallel()
	rows := []struct{ name, definition, agent string }{
		{"row1 the testing agent's definition", ".claude/agents/proviso-tester.md", "proviso-tester"},
		{"row2 the coding agent's definition", ".claude/agents/proviso-coder.md", "proviso-coder"},
		{"row3 the checking agent's definition", ".claude/agents/proviso-checker.md", "proviso-checker"},
	}
	definitions := make([]string, len(rows))
	for i, row := range rows {
		definitions[i] = row.definition
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := freshRepo(t, "repo", false)
			r.init().exits(t, 0)
			def := strings.ReplaceAll(r.read(row.definition), "\r\n", "\n")

			// Named as the agent host reads it: YAML front matter with
			// the agent's name.
			lines := strings.Split(def, "\n")
			named := false
			if lines[0] == "---" {
				for _, l := range lines[1:] {
					if l == "---" {
						break
					}
					if regexp.MustCompile(`^name:\s*["']?` + regexp.QuoteMeta(row.agent) + `["']?\s*$`).MatchString(l) {
						named = true
					}
				}
			}
			if !named {
				t.Errorf("install.init.agents: %s does not start with front matter naming it %s", row.definition, row.agent)
			}

			// The same reading guide in every definition: one block that
			// this definition shares with both others.
			var others []string
			for _, d := range definitions {
				if d != row.definition {
					others = append(others, d)
				}
			}
			between, _ := sharedBlock(r.read(others[0]), r.read(others[1]))
			if block, n := sharedBlock(def, between); n < guideLines {
				t.Errorf("install.init.agents: %s shares no reading guide with %s and %s; the longest block all three share has %d lines that are not blank:\n%s", row.definition, others[0], others[1], n, block)
			}

			// No reference to a file that is not there.
			var refs []string
			for _, m := range linkTarget.FindAllStringSubmatch(def, -1) {
				target := strings.SplitN(m[1], "#", 2)[0]
				if target == "" || strings.Contains(target, ":") {
					continue
				}
				refs = append(refs, target)
			}
			for _, m := range importPath.FindAllStringSubmatch(def, -1) {
				if strings.Contains(m[1], "/") || strings.HasSuffix(m[1], ".md") {
					refs = append(refs, m[1])
				}
			}
			for _, ref := range refs {
				candidates := []string{
					filepath.Join(r.dir, filepath.FromSlash(filepath.Dir(row.definition)), filepath.FromSlash(ref)),
					filepath.Join(r.dir, filepath.FromSlash(ref)),
				}
				found := false
				for _, c := range candidates {
					if _, err := os.Stat(c); err == nil {
						found = true
					}
				}
				if !found {
					t.Errorf("install.init.agents: %s refers to %s, which is not in the repository", row.definition, ref)
				}
			}
		})
	}
}

func TestInstallInitExisting(t *testing.T) {
	t.Parallel()
	t.Run("row1 proviso/system.md, in a repository without proviso/", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		o := r.init()
		o.exits(t, 0)
		o.says(t, "written", "proviso/system.md")
		if o.after["proviso"] != "folder" {
			t.Errorf("install.init.existing: proviso/ was not made")
		}
	})
	t.Run("row2 proviso/system.md, which already exists", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.set(tree{"proviso/system.md": "# Our own system\n\nNot even a valid tree.\n"})
		o := r.init()
		o.exits(t, 0)
		o.says(t, "skipped", "proviso/system.md")
	})
	t.Run("row3 .proviso/tester.md, an empty file", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.set(tree{".proviso/tester.md": ""})
		o := r.init()
		o.exits(t, 0)
		o.says(t, "skipped", ".proviso/tester.md")
	})
	t.Run("row4 proviso/.gitattributes, a folder of that name", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		if err := os.MkdirAll(filepath.Join(r.dir, "proviso", ".gitattributes"), 0o755); err != nil {
			t.Fatal(err)
		}
		o := r.init()
		o.exits(t, 0)
		o.says(t, "skipped", "proviso/.gitattributes")
	})
	t.Run("row5 .claude/agents/proviso-tester.md, beside other agents", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.set(tree{
			".claude/agents/reviewer.md": "---\nname: reviewer\n---\n\nReview.\n",
			".claude/agents/writer.md":   "---\nname: writer\n---\n\nWrite.\n",
		})
		o := r.init()
		o.exits(t, 0)
		o.says(t, "written", ".claude/agents/proviso-tester.md")
	})
	t.Run("derived: a file left alone is not read", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.set(tree{"proviso/system.md": "# Our own system\n"})
		o := initIn(t, r.dir, r.dir, func() func() {
			undo, err := makeUnreadable(filepath.Join(r.dir, "proviso", "system.md"))
			if err != nil {
				t.Skipf("cannot make the file unreadable: %v", err)
			}
			return undo
		})
		o.exits(t, 0)
		o.says(t, "skipped", "proviso/system.md")
	})
}

func TestInstallInitExit(t *testing.T) {
	t.Parallel()
	t.Run("row1 none, every file written", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		o := r.init()
		o.exits(t, 0)
		o.all(t, "written")
	})
	t.Run("row2 none, every file skipped", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.init().exits(t, 0)
		o := r.init()
		o.exits(t, 0)
		o.all(t, "skipped")
	})
	t.Run("row3 both agent definitions, .claude being a file", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", false)
		r.set(tree{".claude": "not a folder\n"})
		o := r.init()
		o.exits(t, 2)
		// That no line names a failed file, and that a file not named was
		// not written, initIn checks; here, that the others were. The
		// skill lies below .claude too, and cannot be written either.
		below := 0
		for _, q := range installedFiles {
			if strings.HasPrefix(q, ".claude/") {
				below++
			}
		}
		if len(o.lines) != len(installedFiles)-below {
			t.Errorf("install.init.exit: want the %d files outside .claude named as written, got:\n%s", len(installedFiles)-below, o.stdout)
		}
		for _, q := range installedFiles[below:] {
			o.says(t, "written", q)
		}
	})
}

// TestInstallInitKeeps checks install.init.keeps, a property checked by
// initIn for every run of init; here on repositories with much to keep.
func TestInstallInitKeeps(t *testing.T) {
	t.Parallel()
	own := tree{
		"README.md":                         "# Shop\n",
		".gitattributes":                    "* text=auto\n",
		".claude/settings.json":             "{}\n",
		".claude/agents/reviewer.md":        "---\nname: reviewer\n---\n",
		"src/main.go":                       "package main\r\n",
		"proviso/capabilities/cart/cart.md": "# Cart\n",
	}
	t.Run("derived: every installed file there, with the engineers' edits", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		tr := own.clone()
		for _, q := range installedFiles {
			tr[q] = "edited by the engineers: " + q + "\n"
		}
		r.set(tr)
		o := r.init()
		o.exits(t, 0)
		o.all(t, "skipped")
	})
	t.Run("derived: a repository with code, a tree and settings", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		r.set(own)
		r.init().exits(t, 0)
	})
	t.Run("derived: a run that fails", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		r.set(own.without(".claude/settings.json", ".claude/agents/reviewer.md").with(".claude", "not a folder\n"))
		r.init().exits(t, 2)
	})
}

// gitAttributes gives every attribute git gives the slash-separated path p
// of the repository, one "name: value" each.
func gitAttributes(r *repo, p string) []string {
	r.t.Helper()
	var attrs []string
	for _, l := range strings.Split(r.git("check-attr", "-a", "--", p), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			attrs = append(attrs, strings.TrimPrefix(l, p+": "))
		}
	}
	return attrs
}

func TestInstallInitLineEndings(t *testing.T) {
	t.Parallel()
	rows := []struct {
		name, file string
		later      bool
	}{
		{"row1 proviso/system.md", "proviso/system.md", false},
		{"row2 proviso/capabilities/cart/glossary.md, added later", "proviso/capabilities/cart/glossary.md", true},
		{"row3 proviso/manifest.sum", "proviso/manifest.sum", true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			r := freshRepo(t, "repo", false)
			r.init().exits(t, 0)
			if row.later {
				full := filepath.Join(r.dir, filepath.FromSlash(row.file))
				if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte("added later\nwith a second line\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got := gitAttributes(r, row.file)
			found := false
			for _, a := range got {
				found = found || a == "eol: lf"
			}
			if !found {
				t.Errorf("install.init.line-endings: want git to give %s the attribute eol: lf, got %q", row.file, got)
			}

			// Settings that would give carriage returns: the file is
			// committed, removed and checked out again.
			r.git("config", "core.autocrlf", "true")
			r.git("config", "core.eol", "crlf")
			r.commit("tree")
			full := filepath.Join(r.dir, filepath.FromSlash(row.file))
			if err := os.Remove(full); err != nil {
				t.Fatal(err)
			}
			r.git("checkout", "--", row.file)
			if content := r.read(row.file); content == "" || strings.Contains(content, "\r") {
				t.Errorf("install.init.line-endings: want git to check %s out with line feeds, got %q", row.file, content)
			}
		})
	}
}

func TestInstallInitRepository(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, a clone with a .git folder", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		o := r.init()
		o.exits(t, 0)
		o.all(t, "written")
	})
	t.Run("row2 wt/, a worktree with a .git file", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		wt := filepath.Join(t.TempDir(), "wt")
		r.git("worktree", "add", "-q", "--detach", wt)
		if fi, err := os.Lstat(filepath.Join(wt, ".git")); err != nil || !fi.Mode().IsRegular() {
			t.Fatalf("the worktree has no .git file: %v", err)
		}
		o := initIn(t, wt, wt, nil)
		o.exits(t, 0)
		o.all(t, "written")
	})
	t.Run("row3 repo/, its .git folder marked hidden", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		if err := markHidden(filepath.Join(r.dir, ".git")); err != nil {
			t.Skipf("cannot mark .git hidden: %v", err)
		}
		o := r.init()
		o.exits(t, 0)
		o.all(t, "written")
	})
	t.Run("row4 repo/src/, inside a clone", func(t *testing.T) {
		t.Parallel()
		r := freshRepo(t, "repo", true)
		r.set(tree{"src/main.go": "package main\n"})
		o := initIn(t, r.dir, filepath.Join(r.dir, "src"), nil)
		o.exits(t, 2)
		o.nothingWritten(t)
	})
	t.Run("row5 notes/, in no repository", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "notes")
		tree{"todo.md": "# To do\n"}.writeTo(t, dir, false)
		o := initIn(t, dir, dir, nil)
		o.exits(t, 2)
		o.nothingWritten(t)
	})
}

func TestInstallInitSkeleton(t *testing.T) {
	t.Parallel()
	skeleton := func(t *testing.T, r *repo, name string) {
		t.Helper()
		r.init().exits(t, 0)

		heading := ""
		for _, l := range strings.Split(strings.ReplaceAll(r.read("proviso/system.md"), "\r\n", "\n"), "\n") {
			if strings.HasPrefix(l, "#") {
				heading = l
				break
			}
		}
		if want := `# <a id="system"></a> ` + name; heading != want {
			t.Errorf("install.init.skeleton: want the heading %q, got %q", want, heading)
		}

		o := lintIn(t, r.dir, r.dir, nil)
		if o.json.exit != 0 {
			t.Errorf("install.init.skeleton: want lint to exit with 0, got %d:\n%s%s", o.json.exit, dump(o.report.Findings), o.json.stderr)
		}
		seen := map[int]bool{}
		for _, f := range o.report.Findings {
			if f.Kind != "undecided" || f.File != "proviso/system.md" || seen[f.Line] {
				t.Errorf("install.init.skeleton: want lint to ask for each of the six sections in proviso/system.md, and nothing else, got:\n%s", dump(o.report.Findings))
				break
			}
			seen[f.Line] = true
		}
		if len(o.report.Findings) != 6 {
			t.Errorf("install.init.skeleton: want six asks, got:\n%s", dump(o.report.Findings))
		}

		d := diffIn(t, r.dir, r.dir, r.root())
		d.only(t, "system:added")
		if e := d.elem(t, "system"); len(e.Undecided) == 0 {
			t.Errorf("install.init.skeleton: want the system listed as not to be built, but diff names no undecided place")
		}
	}
	t.Run("row1 C:/work/shop", func(t *testing.T) {
		t.Parallel()
		skeleton(t, freshRepo(t, "shop", true), "shop")
	})
	t.Run("row2 /home/me/cash-ops", func(t *testing.T) {
		t.Parallel()
		skeleton(t, freshRepo(t, "cash-ops", true), "cash-ops")
	})
	t.Run("derived: a repository without a commit", func(t *testing.T) {
		t.Parallel()
		skeleton(t, freshRepo(t, "shop", false), "shop")
	})
}
