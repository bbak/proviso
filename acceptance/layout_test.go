package acceptance

// Feature lint.layout: proviso/capabilities/lint/features/layout.md

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLintLayoutHidden(t *testing.T) {
	t.Parallel()
	// A skipped entry holds content that lint would report or could not
	// read, so that reading it would show.
	cases := []struct {
		name    string
		tree    tree
		skipped bool
	}{
		{"row1 .DS_Store", baseline().with("proviso/.DS_Store", "\x00\x01\xff junk"), true},
		{"row2 a hidden folder, with its content", baseline().with(
			"proviso/capabilities/cart/.vscode/settings.json", "{}\n",
			"proviso/capabilities/cart/.vscode/notes.md", "# Notes\n",
			"proviso/capabilities/cart/.vscode/bad.md", "\xff\xfe\n"), true},
		{"row3 .markdownlint.jsonc in the spec root", baseline().with("proviso/.markdownlint.jsonc", "{\n  \"MD013\": false\n}\n"), false},
		{"row4 capabilities/cart/.markdownlint.jsonc", baseline().with("proviso/capabilities/cart/.markdownlint.jsonc", "\xff\xfe not settings"), true},
		{"row5 glossary.md", baseline(), false},
		{"derived: a hidden file in the spec root", baseline().with("proviso/.notes.md", "# Notes\n"), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// Skipped, or handed over to lint.layout.file and read: either
			// way nothing is reported.
			check(t, c.tree).clean(t)
		})
	}
	t.Run("row6 notes.md, marked hidden by Windows", func(t *testing.T) {
		t.Parallel()
		// Hidden means a name starting with a dot: the mark does not
		// travel with the tree, so notes.md is handed over to
		// lint.layout.file, which reports it.
		dir := baseline().with("proviso/notes.md", "# Notes\n").write(t)
		if err := markHidden(filepath.Join(dir, "proviso", "notes.md")); err != nil {
			t.Skipf("cannot mark a file hidden here: %v", err)
		}
		o := lintIn(t, dir, dir, nil)
		checkRoot(t, o, filepath.Join(dir, "proviso"))
		o.has(t, "file-unknown", "proviso/notes.md", 1)
	})
	t.Run("derived: the spec root's settings are read, not skipped", func(t *testing.T) {
		t.Parallel()
		// Handed over to lint.layout.file, which reads the file: a file
		// that cannot be decoded stops lint (lint.run.unreadable).
		dir := baseline().with("proviso/.markdownlint.jsonc", "\xff\xfe not settings").write(t)
		o := lintIn(t, dir, dir, nil)
		if o.json.exit != 2 {
			t.Errorf("want exit status 2, got %d:\n%s", o.json.exit, dump(o.report.Findings))
		}
	})
}

// moved moves the file from, or every file below the folder from, to to.
func (tr tree) moved(from, to string) tree {
	c := tree{}
	for k, v := range tr {
		if k == from || strings.HasPrefix(k, from+"/") {
			k = to + k[len(from):]
		}
		c[k] = v
	}
	return c
}

func TestLintLayoutName(t *testing.T) {
	t.Parallel()
	features := "proviso/capabilities/cart/features/"
	upperCart := baseline().moved("proviso/capabilities/cart", "proviso/capabilities/Cart")
	uebersicht := strings.ReplaceAll(featureMD, "cart.add", "cart.uebersicht")
	// ü as one character (NFC), as u followed by a mark (NFD), and Ü.
	composed, decomposed, upper := "übersicht.md", "übersicht.md", "Übersicht.md"
	rows(t, "name-malformed", []row{
		{name: "row1 capabilities/cart/features/add.md", tree: baseline(), holds: true},
		{name: "row2 übersicht.md, a letter outside ASCII", tree: baseline().with(features+composed, uebersicht), holds: true},
	})
	t.Run("row3 übersicht.md, stored as u followed by a mark", func(t *testing.T) {
		t.Parallel()
		// Without a Contact the feature has a finding on its first line,
		// so that its path is reported.
		o := checkGoesOnAfter(t, baseline().with(features+decomposed, featureWith("cart.uebersicht", "", "")), func(dir string) {
			if !hasEntry(filepath.Join(dir, filepath.FromSlash(features)), decomposed) {
				t.Skip("the file system does not keep the name as u followed by a mark")
			}
		})
		o.lacks(t, "name-malformed", "")
		o.has(t, "field-missing", features+composed, 1)
		for _, f := range o.report.Findings {
			if strings.Contains(f.File, decomposed) {
				t.Errorf("lint.cli.report: path %q is not in the composed form", f.File)
			}
		}
	})
	rows(t, "name-malformed", []row{
		{name: "row4 検索.md, a script without case", tree: baseline().with(features+"検索.md", strings.ReplaceAll(featureMD, "cart.add", "cart.search")), holds: true},
		{name: "row5 .markdownlint.jsonc in the spec root", tree: baseline().with("proviso/.markdownlint.jsonc", "{}\n"), holds: true},
		{name: "row6 capabilities/cart/features/Add.md", tree: baseline().moved(featurePath, features+"Add.md"),
			file: features + "Add.md", lines: lines(1)},
		{name: "row7 capabilities/Cart, a folder", tree: upperCart,
			file: "proviso/capabilities/Cart", lines: lines(1)},
		{name: "row8 Übersicht.md, an uppercase letter outside ASCII", tree: baseline().with(features+upper, uebersicht),
			file: features + upper, lines: lines(1)},
		{name: "row9 my add.md, a space", tree: baseline().moved(featurePath, features+"my add.md"),
			file: features + "my add.md", lines: lines(1)},
		{name: "row10 add_old.md, an underscore", tree: baseline().moved(featurePath, features+"add_old.md"),
			file: features + "add_old.md", lines: lines(1)},
	})
	t.Run("row11 add.md., ending with a dot", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOnAfter(t, baseline(), func(dir string) {
			folder := filepath.Join(dir, filepath.FromSlash(features))
			name := filepath.Join(folder, "add.md") + "."
			if runtime.GOOS == "windows" {
				// Windows drops a trailing dot unless the path is passed
				// on as it is.
				name = `\\?\` + name
			}
			if err := os.WriteFile(name, []byte("# Notes\n"), 0o644); err != nil {
				t.Skipf("cannot create a name ending with a dot here: %v", err)
			}
			// Runs before the temporary directory is removed.
			t.Cleanup(func() { os.Remove(name) })
			if !hasEntry(folder, "add.md.") {
				t.Skip("the file system does not keep a name ending with a dot")
			}
		})
		o.has(t, "name-malformed", features+"add.md.", 1)
	})
	t.Run("row12 übersicht.md, once composed and once as u followed by a mark", func(t *testing.T) {
		t.Parallel()
		tr := baseline().with(features+composed, uebersicht,
			features+decomposed, strings.ReplaceAll(featureMD, "cart.add", "cart.overview"))
		o := checkGoesOnAfter(t, tr, func(dir string) {
			folder := filepath.Join(dir, filepath.FromSlash(features))
			if !hasEntry(folder, composed) || !hasEntry(folder, decomposed) {
				t.Skip("the file system does not keep both names as two files")
			}
		})
		// Both paths read the same in the report, in NFC: the error is
		// reported once, at the name both share.
		if n := len(o.find("name-malformed", features+composed, lines(1))); n != 1 {
			t.Errorf("want name-malformed once, at %s:1, got:\n%s", features+composed, dump(o.report.Findings))
		}
	})
	rows(t, "name-malformed", []row{
		{name: "row13 add٢.md, a digit from another script", tree: baseline().with(features+"add٢.md", strings.ReplaceAll(featureMD, "cart.add", "cart.add2")), holds: true},
		// The names must be the same in the folder of <entry>: the same
		// name in NFC in another folder is not the same entry.
		{name: "derived: übersicht.md composed in one folder and decomposed in another", tree: baseline().with(features+composed, uebersicht, "proviso/"+decomposed, "# Notes\n"), holds: true},
	})
	rows(t, "name-malformed", []row{
		{name: "derived: .DS_Store, skipped as hidden", tree: baseline().with("proviso/.DS_Store", "\x00\x01"), holds: true},
		{name: "derived: malformed names in a hidden folder", tree: baseline().with("proviso/.Trash/My Notes.md", "# Notes\n"), holds: true},
		{name: "derived: README.md, a file the layout does not list", tree: baseline().with("proviso/capabilities/cart/README.md", "# Cart\n"),
			file: "proviso/capabilities/cart/README.md", lines: lines(1)},
	})
	t.Run("derived: the folder's name only, not the names in it", func(t *testing.T) {
		t.Parallel()
		o := checkGoesOn(t, upperCart)
		o.has(t, "name-malformed", "proviso/capabilities/Cart", 1)
		if n := len(o.find("name-malformed", "", nil)); n != 1 {
			t.Errorf("want name-malformed once, at the folder, got:\n%s", dump(o.report.Findings))
		}
	})
}

// hasEntry tells whether the folder dir holds an entry named exactly name.
func hasEntry(dir, name string) bool {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() == name {
			return true
		}
	}
	return false
}

// crlf ends the given (1-based) lines of s in a carriage return and a line
// feed, or every line if none are given.
func crlf(s string, at ...int) string {
	ls := strings.SplitAfter(s, "\n")
	for i, l := range ls {
		hit := len(at) == 0
		for _, n := range at {
			hit = hit || n == i+1
		}
		if hit && strings.HasSuffix(l, "\n") {
			ls[i] = strings.TrimSuffix(l, "\n") + "\r\n"
		}
	}
	return strings.Join(ls, "")
}

func TestLintLayoutLineEnding(t *testing.T) {
	t.Parallel()
	settings := "proviso/.markdownlint.jsonc"
	rows(t, "line-ending", []row{
		{name: "row1 glossary.md, every line ending in a line feed", tree: baseline(), holds: true},
	})
	t.Run("row2 glossary.md, lines 3 and 7 ending in CR LF", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		tr[glossaryPath] = crlf(glossaryMD, 3, 7)
		o := checkGoesOn(t, tr)
		o.has(t, "line-ending", glossaryPath, 3)
		// Once per file, at the first; and the line endings change
		// nothing else that lint reads in the file.
		if n := len(o.find("line-ending", glossaryPath, nil)); n != 1 {
			t.Errorf("want line-ending once in %s, got:\n%s", glossaryPath, dump(o.report.Findings))
		}
		for _, f := range o.report.Findings {
			if f.File == glossaryPath && f.Kind != "line-ending" {
				t.Errorf("want only line-ending in %s, got:\n%s", glossaryPath, dump(o.report.Findings))
				break
			}
		}
	})
	t.Run("row3 glossary.md, line 5 holding a lone carriage return in the middle", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		tr[glossaryPath] = strings.Replace(glossaryMD, "A product placed", "A product\rplaced", 1)
		if tr.line(glossaryPath, "placed in the cart") != 5 {
			t.Fatalf("test tree: the carriage return is not on line 5 of %s", glossaryPath)
		}
		o := checkGoesOn(t, tr)
		o.has(t, "line-ending", glossaryPath, 5)
		if n := len(o.find("line-ending", glossaryPath, nil)); n != 1 {
			t.Errorf("want line-ending once in %s, got:\n%s", glossaryPath, dump(o.report.Findings))
		}
	})
	// Once per file, at the first line holding a carriage return, whether
	// in a line ending or within a line.
	t.Run("derived: a lone carriage return on line 5 before a CR LF on line 7", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		tr[glossaryPath] = strings.Replace(crlf(glossaryMD, 7), "A product placed", "A product\rplaced", 1)
		o := checkGoesOn(t, tr)
		o.has(t, "line-ending", glossaryPath, 5)
		if n := len(o.find("line-ending", glossaryPath, nil)); n != 1 {
			t.Errorf("want line-ending once in %s, got:\n%s", glossaryPath, dump(o.report.Findings))
		}
	})
	t.Run("derived: a lone carriage return after the last line feed", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		tr[glossaryPath] = glossaryMD + "\r"
		o := checkGoesOn(t, tr)
		o.has(t, "line-ending", glossaryPath, strings.Count(glossaryMD, "\n")+1)
	})
	rows(t, "line-ending", []row{
		{name: "row4 .markdownlint.jsonc ending its lines in CR LF", tree: baseline().with(settings, crlf("{\n  \"MD013\": false\n}\n")),
			file: settings, lines: lines(1)},
		{name: "row5 manifest.sum ending its lines in CR LF", tree: baseline().with("proviso/manifest.sum", crlf("model 0\nsystem\theading\th1:"+strings.Repeat("0", 64)+"\n")),
			file: "proviso/manifest.sum", lines: lines(1)},
		{name: "derived: a file the layout does not list is never read", tree: baseline().with("proviso/notes.md", crlf("# Notes\n\nSome notes.\n")), holds: true},
		{name: "derived: a hidden file is skipped", tree: baseline().with("proviso/.notes.md", crlf("# Notes\n")), holds: true},
	})
	t.Run("derived: a lone carriage return does not end a line", func(t *testing.T) {
		t.Parallel()
		// lint.cli.report counts lines by line feeds only, so a finding
		// after the carriage return stands on the line a line feed count
		// gives.
		b := "proviso/capabilities/cart/features/b.md"
		tr := baseline().with(b, featureWith("cart.b", "- Contact: Dana Reyes\n\n", "\n## Scope\n\nAll of it.\n"))
		tr[b] = strings.Replace(tr[b], "Something.", "Some\rthing.", 1)
		o := check(t, tr)
		o.has(t, "line-ending", b, tr.line(b, "thing."))
		o.has(t, "heading-unknown", b, tr.line(b, "## Scope"))
	})
	t.Run("derived: a feature in CR LF throughout", func(t *testing.T) {
		t.Parallel()
		tr := baseline()
		tr[featurePath] = crlf(tr[featurePath])
		o := check(t, tr)
		o.has(t, "line-ending", featurePath, 1)
		if len(o.report.Findings) != 1 {
			t.Errorf("want only line-ending in %s, got:\n%s", featurePath, dump(o.report.Findings))
		}
	})
}

func TestLintLayoutFile(t *testing.T) {
	t.Parallel()
	rows(t, "file-unknown", []row{
		{name: "row1 capabilities/cart/features/add.md", tree: baseline(), holds: true},
		{name: "row2 .markdownlint.jsonc", tree: baseline().with("proviso/.markdownlint.jsonc", "{}\n"), holds: true},
		{name: "row3 notes.md", tree: baseline().with("proviso/notes.md", "# Notes\n"),
			file: "proviso/notes.md", lines: lines(1)},
		{name: "row4 capabilities/cart/README.md", tree: baseline().with("proviso/capabilities/cart/README.md", "# Cart\n"),
			file: "proviso/capabilities/cart/README.md", lines: lines(1)},
		{name: "row5 features are not nested", tree: baseline().with("proviso/capabilities/cart/features/old/add.md", featureMD),
			file: "proviso/capabilities/cart/features/old/add.md", lines: lines(1)},
		{name: "row6 invariants are one file per level", tree: baseline().with("proviso/invariants/read-only.md", "# Read only\n"),
			file: "proviso/invariants/read-only.md", lines: lines(1)},
		{name: "derived: a file that is not Markdown", tree: baseline().with("proviso/notes.txt", "notes\n"),
			file: "proviso/notes.txt", lines: lines(1)},
		{name: "derived: every sibling the layout lists", tree: baseline().with(
			"proviso/glossary.md", systemGlossaryMD,
			"proviso/references.md", systemReferencesMD,
			"proviso/constraints.md", constraintsMD,
			"proviso/invariants.md", invariantsMD,
			"proviso/questions.md", "# Questions\n\nShould refunds be part of the cart?\n",
			"proviso/capabilities/cart/contracts/findings.md", findingsContractMD,
			"proviso/capabilities/cart/questions.md", "# Questions\n\nWho decides the shelf size?\n"), holds: true},
	})
}

func TestLintLayoutCapability(t *testing.T) {
	t.Parallel()
	rows(t, "capability-file-missing", []row{
		{name: "row1 holding capability.md", tree: baseline(), holds: true},
		{name: "row2 holding only glossary.md", tree: baseline().with("proviso/capabilities/billing/glossary.md", `# Glossary

## <a id="billing.invoice"></a> invoice

- Definition: What a buyer pays.
- Contact: Dana Reyes
- Status: defined
`), file: "proviso/capabilities/billing", lines: lines(1)},
		{name: "row3 holding Capability.md", tree: baseline().moved(capPath, "proviso/capabilities/cart/Capability.md"),
			file: "proviso/capabilities/cart", lines: lines(1)},
	})
}

func TestLintLayoutTitle(t *testing.T) {
	t.Parallel()
	rows(t, "title-missing", []row{
		{name: "row1 # Glossary", tree: baseline(), holds: true},
		{name: "row2 ## Glossary", tree: baseline().edit(glossaryPath, "# Glossary", "## Glossary"),
			file: glossaryPath, lines: lines(1)},
		{name: "row3 a title with an anchor", tree: baseline().edit(glossaryPath, "# Glossary", `# <a id="cart.glossary"></a> Glossary`),
			file: glossaryPath, lines: lines(1)},
		{name: "row4 an entry as the first heading", tree: baseline().edit(glossaryPath, "# Glossary\n\n", ""),
			file: glossaryPath, lines: lines(1)},
		{name: "derived: parameters.md", tree: baseline().edit(paramsPath, "# Parameters", "## Parameters"),
			file: paramsPath, lines: lines(1)},
		{name: "derived: a contract", tree: baseline().with("proviso/capabilities/cart/contracts/findings.md",
			"## Findings\n\n"+findingsContractMD[len("# Findings\n\n"):]),
			file: "proviso/capabilities/cart/contracts/findings.md", lines: lines(1)},
		{name: "derived: the system's constraints.md", tree: baseline().with("proviso/constraints.md",
			"Some words first.\n\n## <a id=\"system.web\"></a> The shop runs in a browser\n\n### Rationale\n\nx\n\n### Property\n\ny\n"),
			file: "proviso/constraints.md", lines: lines(3)},
	})
}

func TestLintLayoutEmpty(t *testing.T) {
	t.Parallel()
	rows(t, "sibling-empty", []row{
		{name: "row1 glossary.md with one entry", tree: baseline(), holds: true},
		{name: "row2 glossary.md with only # Glossary", tree: baseline().with(glossaryPath, "# Glossary\n"),
			file: glossaryPath, lines: lines(1)},
		{name: "row3 parameters.md with no content at all", tree: baseline().with(paramsPath, ""),
			file: paramsPath, lines: lines(1)},
		{name: "derived: a contract with only its title", tree: baseline().with("proviso/capabilities/cart/contracts/api.md", "# API\n"),
			file: "proviso/capabilities/cart/contracts/api.md", lines: lines(1)},
	})
}

const constraintsMD = `# Constraints

## <a id="system.web"></a> The shop runs in a browser

### Rationale

Buyers have browsers.

### Property

Every page works in a browser.
`

const invariantsMD = `# Invariants

## <a id="system.read-only"></a> The shop never loses a cart

### Rationale

Buyers come back.

### Ranges over

Every visit.

### Property

A cart is there on the next visit.
`

// TestLintTreeLayoutManifest checks the row of tree_layout for manifest.sum:
// in the spec root it belongs to the layout, anywhere else it does not.
func TestLintTreeLayoutManifest(t *testing.T) {
	t.Parallel()
	rows(t, "file-unknown", []row{
		{name: "manifest.sum in the spec root", tree: baseline().with("proviso/manifest.sum", "model 0\n"), holds: true, file: "proviso/manifest.sum"},
		{name: "manifest.sum in a capability's folder", tree: baseline().with("proviso/capabilities/cart/manifest.sum", "model 0\n"), file: "proviso/capabilities/cart/manifest.sum"},
		{name: "Manifest.sum in the spec root, names matched exactly", tree: baseline().with("proviso/Manifest.sum", "model 0\n"), file: "proviso/Manifest.sum"},
	})
}
