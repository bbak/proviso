package acceptance

// Feature lint.run: proviso/capabilities/lint/features/run.md

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skipIfSpecRootAbove skips a test that needs no spec root up to the file
// system's root when the machine has one above dir.
func skipIfSpecRootAbove(t *testing.T, dir string) {
	t.Helper()
	for d := filepath.Dir(dir); ; d = filepath.Dir(d) {
		if fi, err := os.Stat(filepath.Join(d, "proviso", "system.md")); err == nil && fi.Mode().IsRegular() {
			t.Skipf("%s holds a spec root above the test directory", d)
		}
		if filepath.Dir(d) == d {
			return
		}
	}
}

func TestLintRunRoot(t *testing.T) {
	marked := baseline().with("proviso/notes.md", "# Notes\n")

	t.Run("row1 repo holding proviso/system.md", func(t *testing.T) {
		dir := marked.write(t)
		o := lintIn(t, dir, dir, nil)
		o.has(t, "file-unknown", "proviso/notes.md")
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	})
	t.Run("row2 repo/cmd/proviso found upwards", func(t *testing.T) {
		dir := marked.with("cmd/proviso/main.go", "package main\n").write(t)
		o := lintIn(t, dir, filepath.Join(dir, "cmd", "proviso"), nil)
		o.has(t, "file-unknown", "proviso/notes.md")
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	})
	t.Run("row3 found from inside the tree", func(t *testing.T) {
		dir := marked.write(t)
		o := lintIn(t, dir, filepath.Join(dir, "proviso", "capabilities", "cart"), nil)
		// Paths are relative to the directory holding the spec root,
		// wherever lint was run.
		o.has(t, "file-unknown", "proviso/notes.md")
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	})
	t.Run("row4 the nearest wins", func(t *testing.T) {
		tr := baseline().with("proviso/outer.md", "# Outer\n")
		for k, v := range baseline().with("proviso/inner.md", "# Inner\n").prefixed("a") {
			tr[k] = v
		}
		dir := tr.write(t)
		o := lintIn(t, dir, filepath.Join(dir, "a"), nil)
		o.has(t, "file-unknown", "proviso/inner.md")
		o.lacks(t, "file-unknown", "proviso/outer.md")
		checkRoot(t, o, filepath.Join(dir, "a", "proviso"))
		if len(o.report.Findings) != 1 {
			t.Errorf("want only the nearest tree checked, got:\n%s", dump(o.report.Findings))
		}
	})

	noRoot := []struct {
		name string
		tree tree
		cwd  string
	}{
		{"row5 proviso/ without system.md", tree{"proviso/glossary.md": glossaryMD}, ""},
		{"row6 system.md but no proviso/", tree{"system.md": systemMD}, ""},
		{"row7 no spec root up to the file system's root", tree{"other/readme.txt": "nothing here\n"}, "other"},
	}
	for _, c := range noRoot {
		t.Run(c.name, func(t *testing.T) {
			dir := c.tree.write(t)
			skipIfSpecRootAbove(t, dir)
			o := lintIn(t, dir, filepath.Join(dir, c.cwd), nil)
			for _, r := range []run{o.json, o.text} {
				if r.exit != 2 {
					t.Errorf("want exit status 2, got %d\nstdout: %s", r.exit, r.stdout)
				}
				if strings.TrimSpace(r.stderr) == "" {
					t.Errorf("want stderr to say that no spec root was found, got nothing")
				}
			}
		})
	}
}

func TestLintRunUnreadable(t *testing.T) {
	// The tree also holds an error, so that any finding lint reported
	// despite the unreadable file would show.
	marked := baseline().with("proviso/notes.md", "# Notes\n")
	invalid := "# <a id=\"cart.bad\"></a> Bad \xff\xfe\n\n- Contact: Dana Reyes\n"
	aPath := "proviso/capabilities/cart/features/a.md"
	bPath := "proviso/capabilities/cart/features/b.md"

	// stopped checks exit status 2 and the names on stderr. That nothing
	// goes to stdout is checked for every run with exit status 2.
	stopped := func(t *testing.T, o outcome, names ...string) {
		t.Helper()
		for _, r := range []run{o.json, o.text} {
			if r.exit != 2 {
				t.Fatalf("want exit status 2, got %d\nstdout: %s\nstderr: %s", r.exit, r.stdout, r.stderr)
			}
			for _, n := range names {
				if !strings.Contains(r.stderr, n) {
					t.Errorf("want stderr to name %s, got: %s", n, r.stderr)
				}
			}
		}
	}

	t.Run("row1 none", func(t *testing.T) {
		check(t, baseline()).clean(t)
	})
	t.Run("row2 glossary.md without read permission", func(t *testing.T) {
		dir := marked.write(t)
		path := filepath.Join(dir, filepath.FromSlash(glossaryPath))
		undo, err := makeUnreadable(path)
		if err != nil {
			t.Skipf("cannot make a file unreadable here: %v", err)
		}
		undo()
		o := lintIn(t, dir, dir, func() func() {
			undo, err := makeUnreadable(path)
			if err != nil {
				t.Fatal(err)
			}
			return undo
		})
		stopped(t, o, "glossary.md")
	})
	t.Run("row3 a.md not valid UTF-8", func(t *testing.T) {
		dir := marked.with(aPath, invalid).write(t)
		stopped(t, lintIn(t, dir, dir, nil), "a.md")
	})
	t.Run("row4 shared/ a link to a folder outside the tree", func(t *testing.T) {
		dir := marked.with("outside/glossary.md", glossaryMD).write(t)
		link := filepath.Join(dir, "proviso", "shared")
		if err := os.Symlink(filepath.Join(dir, "outside"), link); err != nil {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		stopped(t, lintIn(t, dir, dir, nil), "shared")
	})
	t.Run("row5 x.md a link to another feature in the tree", func(t *testing.T) {
		dir := marked.write(t)
		link := filepath.Join(dir, filepath.FromSlash(xPath))
		if err := os.Symlink(filepath.Join(dir, filepath.FromSlash(featurePath)), link); err != nil {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		stopped(t, lintIn(t, dir, dir, nil), "x.md")
	})
	t.Run("row6 a.md and b.md both unreadable", func(t *testing.T) {
		dir := marked.with(aPath, invalid, bPath, strings.Replace(invalid, "cart.bad", "cart.worse", 1)).write(t)
		stopped(t, lintIn(t, dir, dir, nil), "a.md", "b.md")
	})
	t.Run("row7 notes.md, not in the layout, without read permission", func(t *testing.T) {
		dir := marked.write(t)
		path := filepath.Join(dir, "proviso", "notes.md")
		undo, err := makeUnreadable(path)
		if err != nil {
			t.Skipf("cannot make a file unreadable here: %v", err)
		}
		undo()
		o := lintIn(t, dir, dir, func() func() {
			undo, err := makeUnreadable(path)
			if err != nil {
				t.Fatal(err)
			}
			return undo
		})
		if o.json.exit == 2 {
			t.Fatalf("want the tree checked, but lint could not run:\n%s", o.json.stderr)
		}
		o.has(t, "file-unknown", "proviso/notes.md", 1)
		if len(o.report.Findings) != 1 {
			t.Errorf("want only notes.md reported, got:\n%s", dump(o.report.Findings))
		}
	})
	// A link counts as unreadable wherever it stands, whatever its name.
	links := []struct{ name, link string }{
		{"derived: a link the layout does not list", "proviso/notes.md"},
		{"derived: a link with a hidden name", "proviso/.shared.md"},
	}
	for _, c := range links {
		t.Run(c.name, func(t *testing.T) {
			dir := marked.write(t)
			link := filepath.Join(dir, filepath.FromSlash(c.link))
			os.Remove(link)
			if err := os.Symlink(filepath.Join(dir, filepath.FromSlash(featurePath)), link); err != nil {
				t.Skipf("cannot create a symbolic link here: %v", err)
			}
			stopped(t, lintIn(t, dir, dir, nil), filepath.Base(link))
		})
	}
}

func TestLintRunOutput(t *testing.T) {
	oneError := baseline().with("proviso/notes.md", "# Notes\n")

	t.Run("row1 one error with --json", func(t *testing.T) {
		o := check(t, oneError)
		if len(o.report.Findings) != 1 {
			t.Errorf("want a report with one finding, got:\n%s", dump(o.report.Findings))
		}
		o.has(t, "file-unknown", "proviso/notes.md")
	})
	t.Run("row2 none with --json", func(t *testing.T) {
		o := check(t, baseline())
		// The report's shape, an array for findings, is checked for
		// every run; here it must be empty.
		if strings.TrimSpace(o.json.stdout) == "" || len(o.report.Findings) != 0 {
			t.Errorf("want a report with no findings, got: %q", o.json.stdout)
		}
	})
	t.Run("row3 one error as text", func(t *testing.T) {
		o := check(t, oneError)
		text := strings.TrimRight(o.text.stdout, "\r\n")
		if text == "" || strings.Contains(text, "\n") {
			t.Errorf("want one line of text, got: %q", o.text.stdout)
		}
	})
	t.Run("row4 none as text", func(t *testing.T) {
		o := check(t, baseline())
		if o.text.stdout != "" {
			t.Errorf("want nothing written, got: %q", o.text.stdout)
		}
	})
}

// featureWith writes a feature file of the cart with fields and extra body.
func featureWith(id, fields, extra string) string {
	return "# <a id=\"" + id + "\"></a> Some feature\n\n" + fields + `## Goal

Something.

## Rationale

Some reason.
` + extra
}

func TestLintRunOrder(t *testing.T) {
	features := "proviso/capabilities/cart/features/"

	t.Run("row1 by file path, then line", func(t *testing.T) {
		// Findings in b.md and a.md, two of them in a.md.
		tr := baseline().with(
			features+"b.md", featureWith("cart.b", "- Contact: Dana Reyes\n- Owner: Team B\n\n", ""),
			features+"a.md", featureWith("cart.a", "- Contact: Dana Reyes\n- Owner: Team A\n\n", "\n## Scope\n\nAll of it.\n"),
		)
		o := check(t, tr)
		owner := o.has(t, "field-unknown", features+"a.md", tr.line(features+"a.md", "- Owner:"))
		scope := o.has(t, "heading-unknown", features+"a.md", tr.line(features+"a.md", "## Scope"))
		other := o.has(t, "field-unknown", features+"b.md", tr.line(features+"b.md", "- Owner:"))
		index := func(f finding) int {
			for i, g := range o.report.Findings {
				if g.Kind == f.Kind && g.File == f.File && g.Line == f.Line {
					return i
				}
			}
			return -1
		}
		if !(index(owner) < index(scope) && index(scope) < index(other)) {
			t.Errorf("want a.md:%d, a.md:%d, b.md:%d in that order, got:\n%s", owner.Line, scope.Line, other.Line, dump(o.report.Findings))
		}
	})

	t.Run("row2 then by kind on one line", func(t *testing.T) {
		// One heading whose identifier is used twice and that holds a
		// second, misplaced anchor.
		tr := withX(ruleMD("cart.add.limit", `B <a id="cart.x.z"></a>`, idBlock, idTable))
		o := check(t, tr)
		line := tr.line(xPath, `id="cart.add.limit"`)
		o.has(t, "anchor-misplaced", xPath, line)
		o.has(t, "identifier-duplicate", xPath, line)
		var kinds []string
		for _, f := range o.report.Findings {
			if norm(f.File) == xPath && f.Line == line {
				kinds = append(kinds, f.Kind)
			}
		}
		for i := 1; i < len(kinds); i++ {
			if kinds[i-1] > kinds[i] {
				t.Errorf("want kinds on one line in order, got %v", kinds)
			}
		}
	})

	t.Run("row3 byte by byte: uppercase first", func(t *testing.T) {
		// Without a Contact, each feature has a finding on its first line.
		tr := baseline().with(
			features+"alpha.md", featureWith("cart.alpha", "", ""),
			features+"Zeta.md", featureWith("cart.zeta", "", ""),
		)
		names := []string{"Zeta.md", "alpha.md"}
		if caseSensitive(t) {
			tr = tr.with(
				features+"b.md", featureWith("cart.b", "", ""),
				features+"B.md", featureWith("cart.upper-b", "", ""),
			)
			names = []string{"B.md", "Zeta.md", "alpha.md", "b.md"}
		}
		o := check(t, tr)
		var got []string
		for _, f := range o.report.Findings {
			if f.Kind == "field-missing" {
				got = append(got, filepath.Base(norm(f.File)))
			}
		}
		if strings.Join(got, ",") != strings.Join(names, ",") {
			t.Errorf("want field-missing in %v in that order, got %v", names, got)
		}
	})

	t.Run("not depending on the order files were read", func(t *testing.T) {
		tr := baseline().with(
			"proviso/notes.md", "# Notes\n",
			features+"a.md", featureWith("cart.a", "- Owner: Team A\n\n", "\n## Scope\n\nAll of it.\n"),
			features+"b.md", featureWith("cart.b", "- Contact: ?\n\n", ""),
			"proviso/capabilities/billing/glossary.md", "# Glossary\n",
		)
		first := t.TempDir()
		tr.writeTo(t, first, false)
		second := t.TempDir()
		tr.writeTo(t, second, true)
		a := lintIn(t, first, first, nil)
		again := lintIn(t, first, first, nil)
		b := lintIn(t, second, second, nil)
		if a.json.stdout != again.json.stdout || a.text.stdout != again.text.stdout {
			t.Errorf("two runs on one tree differ:\n%s\n%s", a.json.stdout, again.json.stdout)
		}
		if dump(a.report.Findings) != dump(b.report.Findings) {
			t.Errorf("the same tree written in another order gives other findings:\n%s\n%s", dump(a.report.Findings), dump(b.report.Findings))
		}
		if len(a.report.Findings) < 4 {
			t.Errorf("want several findings to order, got:\n%s", dump(a.report.Findings))
		}
	})
}

// caseSensitive tells whether the file system of the test directories tells
// b.md from B.md.
func caseSensitive(t *testing.T) bool {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "b"), []byte("b"), 0o644)
	os.WriteFile(filepath.Join(dir, "B"), []byte("B"), 0o644)
	entries, _ := os.ReadDir(dir)
	return len(entries) == 2
}

func TestLintRunExit(t *testing.T) {
	ask := baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?")
	cases := []struct {
		name string
		tree tree
		exit int
	}{
		{"row1 none", baseline(), 0},
		{"row2 one ask", ask, 0},
		{"row3 one error", baseline().with("proviso/notes.md", "# Notes\n"), 1},
		{"row4 one error and one ask", ask.with("proviso/notes.md", "# Notes\n"), 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := check(t, c.tree)
			if o.json.exit != c.exit {
				t.Errorf("want exit status %d, got %d:\n%s", c.exit, o.json.exit, dump(o.report.Findings))
			}
			switch c.name {
			case "row2 one ask", "row4 one error and one ask":
				o.has(t, "undecided", featurePath, 3)
			}
		})
	}
}

// TestLintRunReadOnly checks the invariant lint.run.read-only over runs of
// every kind: clean, with errors, with asks, unable to run. Every other test
// checks it too, for each run it makes (see lintIn).
func TestLintRunReadOnly(t *testing.T) {
	trees := map[string]tree{
		"clean":       baseline(),
		"errors":      baseline().with("proviso/notes.md", "# Notes\n", "proviso/capabilities/billing/glossary.md", "# Glossary\n"),
		"asks":        baseline().edit(featurePath, "- Contact: Dana Reyes", "- Contact: ?"),
		"hidden":      baseline().with("proviso/.DS_Store", "\x00\x01", "proviso/.markdownlint.jsonc", "{}\n"),
		"unreadable":  baseline().with("proviso/capabilities/cart/features/a.md", "\xff\xfe\n"),
		"no root":     tree{"proviso/glossary.md": glossaryMD},
		"many errors": withRule("But the shop\nWhen <provided>", "| x |\n| --- |\n"),
	}
	for name, tr := range trees {
		t.Run(name, func(t *testing.T) {
			dir := tr.write(t)
			before := snapshot(t, dir)
			lintIn(t, dir, dir, nil)
			if name != "no root" && len(before) == 0 {
				t.Errorf("snapshot of the spec root is empty")
			}
		})
	}
}
