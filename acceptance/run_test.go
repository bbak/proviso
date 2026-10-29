package acceptance

// Feature lint.run: proviso/capabilities/lint/features/run.md

import (
	"os"
	"path"
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
	t.Parallel()
	marked := baseline().with("proviso/notes.md", "# Notes\n")

	t.Run("row1 repo holding proviso/system.md", func(t *testing.T) {
		t.Parallel()
		dir := marked.write(t)
		o := lintIn(t, dir, dir, nil)
		o.has(t, "file-unknown", "proviso/notes.md")
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	})
	t.Run("row2 repo/cmd/proviso found upwards", func(t *testing.T) {
		t.Parallel()
		dir := marked.with("cmd/proviso/main.go", "package main\n").write(t)
		o := lintIn(t, dir, filepath.Join(dir, "cmd", "proviso"), nil)
		o.has(t, "file-unknown", "proviso/notes.md")
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	})
	t.Run("row3 found from inside the tree", func(t *testing.T) {
		t.Parallel()
		dir := marked.write(t)
		o := lintIn(t, dir, filepath.Join(dir, "proviso", "capabilities", "cart"), nil)
		// Paths are relative to the directory holding the spec root,
		// wherever lint was run.
		o.has(t, "file-unknown", "proviso/notes.md")
		checkRoot(t, o, filepath.Join(dir, "proviso"))
	})
	t.Run("row4 the nearest wins", func(t *testing.T) {
		t.Parallel()
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

	// The rest of the run works on the nearest spec root: its files are
	// read, and its manifest gives the model version, not the outer one's.
	t.Run("derived: the nearest tree's file is unreadable, the outer tree is fine", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		for k, v := range baseline().with("proviso/capabilities/cart/features/a.md", "# <a id=\"cart.bad\"></a> Bad \xff\xfe\n").prefixed("a") {
			tr[k] = v
		}
		dir := tr.write(t)
		o := lintIn(t, dir, filepath.Join(dir, "a"), nil)
		if o.json.exit != 2 || !strings.Contains(o.json.stderr, "a.md") {
			t.Errorf("want exit status 2 naming a.md of the nearest tree, got %d\nstderr: %s", o.json.exit, o.json.stderr)
		}
	})
	t.Run("derived: the nearest tree's manifest names model 1, the outer one model 0", func(t *testing.T) {
		t.Parallel()
		tr := baseline().with("proviso/manifest.sum", "model 0\n")
		for k, v := range baseline().with("proviso/manifest.sum", "model 1\n").prefixed("a") {
			tr[k] = v
		}
		dir := tr.write(t)
		o := lintIn(t, dir, filepath.Join(dir, "a"), nil)
		if o.json.exit != 2 {
			t.Errorf("want exit status 2 for the nearest tree's model version, got %d\nstdout: %s", o.json.exit, o.json.stdout)
		}
	})

	// lint.cli.report: every name in root is in NFC, also when the folder
	// above the spec root is stored as u followed by a mark.
	t.Run("derived: root in NFC", func(t *testing.T) {
		t.Parallel()
		composed, decomposed := "repo-ü", "repo-ü"
		dir := marked.prefixed(decomposed).write(t)
		if !hasEntry(dir, decomposed) {
			t.Skip("the file system does not keep the name as u followed by a mark")
		}
		o := lintIn(t, dir, filepath.Join(dir, decomposed), nil)
		o.has(t, "file-unknown", "proviso/notes.md")
		if strings.Contains(o.report.Root, decomposed) || !strings.HasSuffix(o.report.Root, "/"+composed+"/proviso") {
			t.Errorf("lint.cli.report: root %q, want it to end in /%s/proviso, in NFC", o.report.Root, composed)
		}
	})

	noRoot := []struct {
		name string
		tree tree
		cwd  string
	}{
		{"row5 proviso/ without system.md", tree{"proviso/glossary.md": glossaryMD}, ""},
		{"row6 Proviso/System.md, names matched exactly", tree{"Proviso/System.md": systemMD}, ""},
		{"row7 system.md but no proviso/", tree{"system.md": systemMD}, ""},
		{"row8 no spec root up to the file system's root", tree{"other/readme.txt": "nothing here\n"}, "other"},
	}
	for _, c := range noRoot {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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
	// stoppedIn is stopped, with the test's directory dir taken out of
	// stderr first, since its name holds the test's name.
	stoppedIn := func(t *testing.T, o outcome, dir string, names ...string) {
		t.Helper()
		for _, r := range []*run{&o.json, &o.text} {
			for _, d := range []string{dir, filepath.ToSlash(dir)} {
				r.stderr = strings.ReplaceAll(r.stderr, d, "")
			}
		}
		stopped(t, o, names...)
	}

	t.Run("row1 none", func(t *testing.T) {
		t.Parallel()
		check(t, baseline()).clean(t)
	})
	t.Run("row2 glossary.md without read permission", func(t *testing.T) {
		t.Parallel()
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
		stoppedIn(t, o, dir, "glossary.md")
	})
	t.Run("row3 a.md not valid UTF-8", func(t *testing.T) {
		t.Parallel()
		dir := marked.with(aPath, invalid).write(t)
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "a.md")
	})
	t.Run("row4 shared, a link to a folder outside the tree", func(t *testing.T) {
		t.Parallel()
		dir := marked.with("outside/glossary.md", glossaryMD).write(t)
		link := filepath.Join(dir, "proviso", "shared")
		if err := os.Symlink(filepath.Join(dir, "outside"), link); err != nil {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "shared")
	})
	t.Run("row5 shared, a Windows junction to a folder outside the tree", func(t *testing.T) {
		t.Parallel()
		dir := marked.with("outside/glossary.md", glossaryMD).write(t)
		if err := makeJunction(filepath.Join(dir, "proviso", "shared"), filepath.Join(dir, "outside")); err != nil {
			t.Skipf("cannot create a junction here: %v", err)
		}
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "shared")
	})
	t.Run("row6 x.md a link to another feature in the tree", func(t *testing.T) {
		t.Parallel()
		dir := marked.write(t)
		link := filepath.Join(dir, filepath.FromSlash(xPath))
		if err := os.Symlink(filepath.Join(dir, filepath.FromSlash(featurePath)), link); err != nil {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "x.md")
	})
	t.Run("row7 a.md and b.md both unreadable", func(t *testing.T) {
		t.Parallel()
		dir := marked.with(aPath, invalid, bPath, strings.Replace(invalid, "cart.bad", "cart.worse", 1)).write(t)
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "a.md", "b.md")
	})
	t.Run("row8 manifest.sum without read permission", func(t *testing.T) {
		t.Parallel()
		dir := marked.with("proviso/manifest.sum", "model 0\n").write(t)
		path := filepath.Join(dir, "proviso", "manifest.sum")
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
		stoppedIn(t, o, dir, "manifest.sum")
	})
	// A folder at the path of a file tree_layout lists counts as
	// unreadable: stderr names the folder, not what it holds.
	folders := []struct {
		name, path, holds string // holds: a file in the folder, "" for none
	}{
		{"row9 manifest.sum, a folder", "proviso/manifest.sum", ""},
		{"row10 capabilities/cart/glossary.md, a folder holding a.md", glossaryPath, "a.md"},
		{"row11 capabilities/cart/features/run.md, an empty folder", "proviso/capabilities/cart/features/run.md", ""},
		{"derived: glossary.md of the spec root, a folder holding a.md", "proviso/glossary.md", "a.md"},
	}
	for _, c := range folders {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tr := marked.without(c.path)
			if c.holds != "" {
				tr = tr.with(c.path+"/"+c.holds, glossaryMD)
			}
			dir := tr.write(t)
			if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(c.path)), 0o755); err != nil {
				t.Fatal(err)
			}
			o := lintIn(t, dir, dir, nil)
			stoppedIn(t, o, dir, path.Base(c.path))
			if c.holds != "" {
				for _, r := range []run{o.json, o.text} {
					if strings.Contains(r.stderr, "/"+c.holds) || strings.Contains(r.stderr, `\`+c.holds) {
						t.Errorf("want stderr to name the folder %s, not the file in it: %s", c.path, r.stderr)
					}
				}
			}
		})
	}
	// lint reads the whole tree before it checks the model version: an
	// unreadable file stops the run, and is named, even where the manifest
	// names another model version.
	t.Run("derived: a.md not valid UTF-8, and manifest.sum starting with model 1", func(t *testing.T) {
		t.Parallel()
		dir := marked.with(aPath, invalid, "proviso/manifest.sum", "model 1\n").write(t)
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "a.md")
	})
	t.Run("derived: manifest.sum a folder, and a.md not valid UTF-8", func(t *testing.T) {
		t.Parallel()
		dir := marked.with(aPath, invalid).write(t)
		if err := os.MkdirAll(filepath.Join(dir, "proviso", "manifest.sum"), 0o755); err != nil {
			t.Fatal(err)
		}
		stoppedIn(t, lintIn(t, dir, dir, nil), dir, "a.md", "manifest.sum")
	})
	t.Run("row13 capabilities/cart/features/old/, a folder holding run.md", func(t *testing.T) {
		t.Parallel()
		old := "proviso/capabilities/cart/features/old/run.md"
		dir := marked.with(old, featureMD).write(t)
		o := lintIn(t, dir, dir, nil)
		if o.json.exit == 2 {
			t.Fatalf("want the tree checked, but lint could not run:\n%s", o.json.stderr)
		}
		o.has(t, "file-unknown", old, 1)
		o.has(t, "file-unknown", "proviso/notes.md", 1)
	})
	t.Run("row12 notes.md, not in the layout, without read permission", func(t *testing.T) {
		t.Parallel()
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
	// A link counts as unreadable wherever it stands, whatever its name,
	// unless a hidden name has it skipped.
	t.Run("derived: a link the layout does not list", func(t *testing.T) {
		t.Parallel()
		dir := marked.write(t)
		link := filepath.Join(dir, "proviso", "notes.md")
		os.Remove(link)
		if err := os.Symlink(filepath.Join(dir, filepath.FromSlash(featurePath)), link); err != nil {
			t.Skipf("cannot create a symbolic link here: %v", err)
		}
		stopped(t, lintIn(t, dir, dir, nil), "notes.md")
	})
	// Any other entry that is neither a regular file nor a folder counts as
	// unreadable, as a link or a junction does.
	t.Run("derived: neither a regular file nor a folder (a named pipe)", func(t *testing.T) {
		t.Parallel()
		dir := marked.with("outside/glossary.md", glossaryMD).write(t)
		name, err := makeSpecial(dir)
		if err != nil {
			t.Skipf("cannot create such an entry here: %v", err)
		}
		stopped(t, lintIn(t, dir, dir, nil), name)
	})
	t.Run("derived: .shared, a junction with a hidden name", func(t *testing.T) {
		t.Parallel()
		dir := marked.with("outside/glossary.md", glossaryMD).write(t)
		if err := makeJunction(filepath.Join(dir, "proviso", ".shared"), filepath.Join(dir, "outside")); err != nil {
			t.Skipf("cannot create a junction here: %v", err)
		}
		o := lintIn(t, dir, dir, nil)
		if o.json.exit == 2 {
			t.Fatalf("want the tree checked and the junction skipped, but lint could not run:\n%s", o.json.stderr)
		}
		o.has(t, "file-unknown", "proviso/notes.md", 1)
		if len(o.report.Findings) != 1 {
			t.Errorf("want only notes.md reported, got:\n%s", dump(o.report.Findings))
		}
	})
	hidden := []struct{ name, link, target string }{
		{"row14 .shared.md a link with a hidden name", "proviso/.shared.md", featurePath},
		{"derived: .shared, a hidden link to a folder outside the tree", "proviso/.shared", "outside"},
	}
	for _, c := range hidden {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := marked.with("outside/glossary.md", glossaryMD).write(t)
			link := filepath.Join(dir, filepath.FromSlash(c.link))
			if err := os.Symlink(filepath.Join(dir, filepath.FromSlash(c.target)), link); err != nil {
				t.Skipf("cannot create a symbolic link here: %v", err)
			}
			o := lintIn(t, dir, dir, nil)
			if o.json.exit == 2 {
				t.Fatalf("want the tree checked and the link skipped, but lint could not run:\n%s", o.json.stderr)
			}
			// The tree is checked: notes.md is reported, and nothing else.
			o.has(t, "file-unknown", "proviso/notes.md", 1)
			if len(o.report.Findings) != 1 {
				t.Errorf("want only notes.md reported, got:\n%s", dump(o.report.Findings))
			}
		})
	}
}

func TestLintRunOutput(t *testing.T) {
	t.Parallel()
	oneError := baseline().with("proviso/notes.md", "# Notes\n")

	t.Run("row1 one error with --json", func(t *testing.T) {
		t.Parallel()
		o := check(t, oneError)
		if len(o.report.Findings) != 1 {
			t.Errorf("want a report with one finding, got:\n%s", dump(o.report.Findings))
		}
		o.has(t, "file-unknown", "proviso/notes.md")
	})
	t.Run("row2 none with --json", func(t *testing.T) {
		t.Parallel()
		o := check(t, baseline())
		// The report's shape, an array for findings, is checked for
		// every run; here it must be empty.
		if strings.TrimSpace(o.json.stdout) == "" || len(o.report.Findings) != 0 {
			t.Errorf("want a report with no findings, got: %q", o.json.stdout)
		}
	})
	t.Run("row3 one error as text", func(t *testing.T) {
		t.Parallel()
		o := check(t, oneError)
		text := strings.TrimRight(o.text.stdout, "\r\n")
		if text == "" || strings.Contains(text, "\n") {
			t.Errorf("want one line of text, got: %q", o.text.stdout)
		}
	})
	t.Run("row4 none as text", func(t *testing.T) {
		t.Parallel()
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
	t.Parallel()
	features := "proviso/capabilities/cart/features/"

	t.Run("row1 by file path, then line", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		// One heading whose identifier is used twice and that holds a
		// second, misplaced anchor.
		tr := withX(ruleMD("cart.add.limit", `B <a id="cart.x.z"></a>`, idBlock, idTable))
		o := check(t, tr)
		line := tr.line(xPath, `id="cart.add.limit"`)
		o.has(t, "anchor-misplaced", xPath, line)
		o.has(t, "identifier-duplicate", xPath, line)
		var kinds []string
		for _, f := range o.report.Findings {
			if f.File == xPath && f.Line == line {
				kinds = append(kinds, f.Kind)
			}
		}
		for i := 1; i < len(kinds); i++ {
			if kinds[i-1] > kinds[i] {
				t.Errorf("want kinds on one line in order, got %v", kinds)
			}
		}
	})

	// Without a Contact, each feature has a finding on its first line.
	byteOrder := []struct {
		name  string
		files []string // in the order the findings must appear
	}{
		{"row3 a.md:1, a-b.md:1", []string{"a-b.md", "a.md"}},
		// The composed ä is 0xC3 0xA4, after every ASCII byte.
		{"derived: a letter outside ASCII after z", []string{"z.md", "ä.md"}},
	}
	for _, c := range byteOrder {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tr := baseline()
			for i, f := range c.files {
				tr = tr.with(features+f, featureWith([]string{"cart.first", "cart.second"}[i], "", ""))
			}
			o := check(t, tr)
			var got []string
			for _, f := range o.report.Findings {
				if f.Kind == "field-missing" {
					got = append(got, path.Base(f.File))
				}
			}
			if strings.Join(got, ",") != strings.Join(c.files, ",") {
				t.Errorf("want field-missing in %v in that order, got %v", c.files, got)
			}
		})
	}

	t.Run("not depending on the order files were read", func(t *testing.T) {
		t.Parallel()
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

func TestLintRunExit(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
			dir := tr.write(t)
			before := snapshot(t, dir)
			lintIn(t, dir, dir, nil)
			if name != "no root" && len(before) == 0 {
				t.Errorf("snapshot of the spec root is empty")
			}
		})
	}
}

func TestLintRunModel(t *testing.T) {
	t.Parallel()
	// stopped checks exit status 2 and that stderr says something; that
	// nothing goes to stdout is checked for every run with exit status 2.
	stopped := func(t *testing.T, o outcome) {
		t.Helper()
		for _, r := range []run{o.json, o.text} {
			if r.exit != 2 {
				t.Errorf("want exit status 2, got %d\nstdout: %s", r.exit, r.stdout)
			}
		}
	}
	t.Run("row1 without manifest.sum", func(t *testing.T) {
		t.Parallel()
		check(t, baseline()).clean(t)
	})
	t.Run("row2 manifest.sum starting with model 0", func(t *testing.T) {
		t.Parallel()
		check(t, baseline().with("proviso/manifest.sum", "model 0\n")).clean(t)
	})
	t.Run("row3 manifest.sum starting with model 0 and a carriage return", func(t *testing.T) {
		t.Parallel()
		// The model check passes; the line ending is line-ending's to
		// report, and nothing else.
		o := check(t, baseline().with("proviso/manifest.sum", "model 0\r\n"))
		o.has(t, "line-ending", "proviso/manifest.sum", 1)
		if len(o.report.Findings) != 1 {
			t.Errorf("want only line-ending reported, got:\n%s", dump(o.report.Findings))
		}
	})
	// Handed over to lint.run.model-other, which stops with exit status 2.
	t.Run("row4 manifest.sum starting with model 1", func(t *testing.T) {
		t.Parallel()
		stopped(t, check(t, baseline().with("proviso/manifest.sum", "model 1\n")))
	})
	t.Run("row5 manifest.sum empty", func(t *testing.T) {
		t.Parallel()
		stopped(t, check(t, baseline().with("proviso/manifest.sum", "")))
	})
	t.Run("row6 manifest.sum starting with model 00", func(t *testing.T) {
		t.Parallel()
		stopped(t, check(t, baseline().with("proviso/manifest.sum", "model 00\n")))
	})
	t.Run("row7 manifest.sum starting with model 0, a carriage return and foo", func(t *testing.T) {
		t.Parallel()
		// A lone carriage return belongs to the line, so the line is no
		// model 0; handed over, lint stops before it reports line-ending.
		stopped(t, check(t, baseline().with("proviso/manifest.sum", "model 0\rfoo\n")))
	})
	t.Run("derived: manifest.sum starting with model 0, a carriage return and a line feed, in a tree with an error", func(t *testing.T) {
		t.Parallel()
		// The tree is checked and lint reports what it finds: the error
		// and the line ending, exit status 1.
		o := check(t, baseline().with("proviso/manifest.sum", "model 0\r\n", "proviso/notes.md", "# Notes\n"))
		o.has(t, "line-ending", "proviso/manifest.sum", 1)
		o.has(t, "file-unknown", "proviso/notes.md", 1)
	})
	// Lint checks nothing of the manifest beyond its first line but what it
	// checks of every file it reads: the lines that follow are not checked
	// as manifest lines, but the file must still be readable.
	t.Run("derived: model 0, then lines that are no manifest's", func(t *testing.T) {
		t.Parallel()
		check(t, baseline().with("proviso/manifest.sum", "model 0\nnot a line of a manifest\n<<<<<<< HEAD\n")).clean(t)
	})
	t.Run("derived: model 0, then text that is not UTF-8", func(t *testing.T) {
		t.Parallel()
		o := check(t, baseline().with("proviso/manifest.sum", "model 0\n\xff\xfe\n"))
		stopped(t, o)
		if !strings.Contains(o.json.stderr, "manifest.sum") {
			t.Errorf("want stderr to name manifest.sum, got: %s", o.json.stderr)
		}
	})
	// The model version a tree needs is checked before the tree: a tree
	// with an error gives exit status 2, not 1, and no report.
	t.Run("derived: model 1 and a tree with an error", func(t *testing.T) {
		t.Parallel()
		stopped(t, check(t, baseline().with("proviso/manifest.sum", "model 1\n", "proviso/notes.md", "# Notes\n")))
	})
}

func TestLintRunModelOther(t *testing.T) {
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
			dir := baseline().with("proviso/manifest.sum", row.line+"\n").write(t)
			o := lintIn(t, dir, dir, nil)
			// That nothing goes to stdout is checked for every run with
			// exit status 2.
			for _, r := range []run{o.json, o.text} {
				if r.exit != 2 {
					t.Errorf("want exit status 2, got %d\nstdout: %s", r.exit, r.stdout)
				}
				nums := numbers(r.stderr, dir, filepath.ToSlash(dir), "manifest.sum")
				if row.needs != "" {
					if count(nums, row.needs) == 0 || count(nums, modelVersion) == 0 {
						t.Errorf("lint.run.model-other: want stderr to name model %s as the one the tree needs and model %s as the one lint reads: %q", row.needs, modelVersion, r.stderr)
					}
				} else if count(nums, modelVersion) > 1 {
					t.Errorf("lint.run.model-other: %q names no model version, but stderr names model %s as needed: %q", row.line, modelVersion, r.stderr)
				} else if row.not != "" && count(nums, row.not) > 0 && count(nums, modelVersion) > 0 {
					t.Errorf("lint.run.model-other: %q names no model version, but stderr names model %s as needed and %s as read: %q", row.line, row.not, modelVersion, r.stderr)
				}
			}
		})
	}
}
