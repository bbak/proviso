package acceptance

// Feature lint.layout: proviso/capabilities/lint/features/layout.md

import "testing"

func TestLintLayoutHidden(t *testing.T) {
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
			// Skipped, or handed over to lint.layout.file and read: either
			// way nothing is reported.
			check(t, c.tree).clean(t)
		})
	}
	t.Run("derived: the spec root's settings are read, not skipped", func(t *testing.T) {
		// Handed over to lint.layout.file, which reads the file: a file
		// that cannot be decoded stops lint (lint.run.unreadable).
		dir := baseline().with("proviso/.markdownlint.jsonc", "\xff\xfe not settings").write(t)
		o := lintIn(t, dir, dir, nil)
		if o.json.exit != 2 {
			t.Errorf("want exit status 2, got %d:\n%s", o.json.exit, dump(o.report.Findings))
		}
	})
}

func TestLintLayoutFile(t *testing.T) {
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
	rows(t, "capability-file-missing", []row{
		{name: "row1 holding capability.md", tree: baseline(), holds: true},
		{name: "row2 holding only glossary.md", tree: baseline().with("proviso/capabilities/billing/glossary.md", `# Glossary

## <a id="billing.invoice"></a> invoice

- Definition: What a buyer pays.
- Contact: Dana Reyes
- Status: defined
`), file: "proviso/capabilities/billing", lines: lines(1)},
	})
}

func TestLintLayoutTitle(t *testing.T) {
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
