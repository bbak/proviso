package lint

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/tree"
)

const baseSystem = `# <a id="system"></a> Sys

## Goal

g

## Rationale

r

## Scope

s

## Out of scope

o

## Audience

a

## Design boundary

d
`

const baseGlossary = `# Glossary

## <a id="system.spec-tree"></a> spec tree

- Definition: The files.
- Contact: Ann
- Status: defined
- Aliases: tree

## <a id="system.guard"></a> guard

- Definition: The Provided line.
- Contact: Ann
- Status: defined
- Aliases: guards
`

const baseCapability = `# <a id="c"></a> C

## Goal

g

## Rationale

r
`

const baseParameters = `# Parameters

## <a id="c.limit"></a> limit

- Value: 3
- Configurable: no
- Rationale: because
`

const baseContract = `# Command line

## <a id="c.cli.lint"></a> lint

### <a id="c.cli.lint.exit-status"></a> exit-status

The status.
`

const baseFeature = `# <a id="c.f"></a> F

- Contact: Ann

## Goal

g

## Rationale

r

## <a id="c.f.one"></a> Rule: One

### Rationale

Why.

` + "```guarded" + `
Given     the [spec tree] has <others>
When      lint reads <id>
Provided  <id> is not among <others>
Then      lint reports nothing
  And     lint goes on with the rest of the [spec tree]
Otherwise lint exits with ` + "`lint.exit-status`" + ` 1
  And     nothing else changes
` + "```" + `

| others | id | provided | outcome | why |
| --- | --- | --- | --- | --- |
| a | b | yes | nothing | |
| a | a | no | error | |
`

func base() map[string]string {
	return map[string]string{
		"system.md":                       baseSystem,
		"glossary.md":                     baseGlossary,
		"capabilities/c/capability.md":    baseCapability,
		"capabilities/c/parameters.md":    baseParameters,
		"capabilities/c/contracts/cli.md": baseContract,
		"capabilities/c/features/f.md":    baseFeature,
		"capabilities/c/glossary.md":      "# Glossary\n\n## <a id=\"c.finding\"></a> finding\n\n- Definition: A thing.\n- Contact: Ann\n- Status: defined\n- Aliases: findings\n",
		".markdownlint.jsonc":             "{}",
	}
}

func lintFiles(files map[string]string) []Finding {
	t := &tree.Tree{Root: "proviso"}
	var rels []string
	for rel := range files {
		rels = append(rels, rel)
	}
	slices.Sort(rels)
	caps := map[string]bool{}
	for _, rel := range rels {
		if parts := strings.Split(rel, "/"); len(parts) >= 3 && parts[0] == "capabilities" {
			if _, seen := caps[parts[1]]; !seen {
				caps[parts[1]] = false
			}
			if parts[2] == "capability.md" {
				caps[parts[1]] = true
			}
		}
		k, c := tree.Classify(rel)
		if k == tree.Unknown {
			t.Unknown = append(t.Unknown, rel)
			continue
		}
		t.Files = append(t.Files, &tree.File{Rel: rel, Kind: k, Cap: c, Data: []byte(files[rel])})
	}
	for c, has := range caps {
		t.CapFolders = append(t.CapFolders, tree.CapFolder{Rel: "capabilities/" + c, HasCapability: has})
	}
	return Run(t)
}

// kinds returns "kind@file:line" for each finding, file relative to the
// spec root.
func kinds(fs []Finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Kind+"@"+strings.TrimPrefix(f.File, "proviso/")+":"+strconv.Itoa(f.Line))
	}
	return out
}

func TestBaseIsClean(t *testing.T) {
	if got := kinds(lintFiles(base())); len(got) != 0 {
		t.Fatalf("base tree has findings: %v", got)
	}
}

// withRule replaces the base feature's rule block and table.
func withRule(block, table string) map[string]string {
	f := base()
	head := baseFeature[:strings.Index(baseFeature, "```guarded")]
	f["capabilities/c/features/f.md"] = head + "```guarded\n" + block + "```\n\n" + table
	return f
}

func TestFindings(t *testing.T) {
	const okTable = "| id | provided | outcome |\n| --- | --- | --- |\n| a | yes | x |\n| b | no | y |\n"
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"keyword unknown and But", withRule("When      lint reads <id>\nProvided  <id> is 1\nThen      x\n  And     y\nOtherwise z\n  But     w\n", okTable),
			[]string{"and-missing@capabilities/c/features/f.md:24", "keyword-unknown@capabilities/c/features/f.md:25"}},
		{"keyword order", withRule("When      lint reads <id>\nGiven     a\nThen      x\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"keyword-order@capabilities/c/features/f.md:21"}},
		{"and first", withRule("  And     a\nWhen      lint reads <id>\nThen      x\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"keyword-order@capabilities/c/features/f.md:20"}},
		{"two whens", withRule("When      lint reads <id>\nWhen      again\nThen      x\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"keyword-count@capabilities/c/features/f.md:19"}},
		{"guard without otherwise", withRule("When      lint reads <id>\nProvided  <id> is 1\nThen      x\n  And     y\n", okTable),
			[]string{"guard-incomplete@capabilities/c/features/f.md:19"}},
		{"and missing", withRule("When      lint reads <id>\nThen      x\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"and-missing@capabilities/c/features/f.md:21"}},
		{"hand-over self cycle", withRule("When      lint reads <id>\nProvided  <id> is 1\nThen      x\n  And     y\nOtherwise → c.f.one\n", okTable),
			[]string{"hand-over-cycle@capabilities/c/features/f.md:13"}},
		{"hand-over to feature with and", withRule("When      lint reads <id>\nProvided  <id> is 1\nThen      x\n  And     y\nOtherwise → c.f\n  And     z\n", okTable),
			[]string{"hand-over-and@capabilities/c/features/f.md:25", "hand-over-unresolved@capabilities/c/features/f.md:24"}},
		{"unbound and reserved", withRule("When      lint reads <id>\nThen      <provided>\n  And     y\n", "| id | provided | outcome |\n| --- | --- | --- |\n| a | yes | x |\n"),
			[]string{"placeholder-reserved@capabilities/c/features/f.md:21", "placeholder-unbound@capabilities/c/features/f.md:21", "provided-column@capabilities/c/features/f.md:25"}},
		{"literal placeholders are values", withRule("When      lint reads <id>\nThen      it is \"<why>\" or \"?\"\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			nil},
		{"provided without event", withRule("When      lint reads <id>\nProvided  the sky is blue\nThen      x\n  And     y\nOtherwise z\n  And     w\n", okTable),
			[]string{"provided-without-event@capabilities/c/features/f.md:21"}},
		{"terms", withRule("When      lint reads <id> of the [Spec Tree] and [findings]\nThen      the guard's tree\n  And     [nope]\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"term-unmarked@capabilities/c/features/f.md:21", "term-unmarked@capabilities/c/features/f.md:21", "term-unresolved@capabilities/c/features/f.md:22"}},
		{"references", withRule("When      lint reads <id> `limit` `lint.exit-status`\nThen      `lint.nope` `exit-status` `c.cli.lint`\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"reference-unresolved@capabilities/c/features/f.md:21", "reference-unresolved@capabilities/c/features/f.md:21", "reference-unresolved@capabilities/c/features/f.md:21"}},
		{"undecided tokens", withRule("When      lint reads <id> `?`\nThen      ?\n  And     ?\n", "| id | outcome |\n| --- | --- |\n| ? x | ?? |\n"),
			[]string{"reference-unresolved@capabilities/c/features/f.md:20", "undecided@capabilities/c/features/f.md:21", "undecided@capabilities/c/features/f.md:22", "undecided@capabilities/c/features/f.md:27"}},
		{"examples missing", withRule("When      lint reads <id>\nThen      x\n  And     y\n", ""),
			[]string{"examples-missing@capabilities/c/features/f.md:13"}},
		{"examples empty and missing columns", withRule("When      lint reads <id>\nThen      x\n  And     y\n", "| why |\n| --- |\n"),
			[]string{"examples-column-missing@capabilities/c/features/f.md:25", "examples-empty@capabilities/c/features/f.md:25"}},
		{"provided values and sides", withRule("When      lint reads <id>\nProvided  <id> is 1\n  And     <id> is 2\nThen      x\n  And     y\nOtherwise z\n  And     w\n",
			"| id | provided | outcome |\n| --- | --- | --- |\n| a | yes | x |\n| b | Yes | x |\n| c | no (one) | x |\n"),
			[]string{"examples-part-missing@capabilities/c/features/f.md:29", "provided-value-unknown@capabilities/c/features/f.md:32"}},
		{"no failing side", withRule("When      lint reads <id>\nProvided  <id> is 1\nThen      x\n  And     y\nOtherwise z\n  And     w\n",
			"| id | provided | outcome |\n| --- | --- | --- |\n| a | yes | x |\n"),
			[]string{"examples-side-missing@capabilities/c/features/f.md:28"}}, // one part: no part check
		{"two whens skip event checks", withRule("When      lint reads <id>\nWhen      <x>\nProvided  the sky\nThen      <y>\n  And     z\nOtherwise w\n  And     v\n", "| id | x | y | provided | outcome |\n| --- | --- | --- | --- | --- |\n| a | b | c | yes | d |\n| a | b | c | no | d |\n"),
			[]string{"keyword-count@capabilities/c/features/f.md:19"}},
		{"two whens keep the other checks", withRule("When      lint reads <id>\nWhen      <x> [nope]\nThen      y\n", "| id | x | outcome |\n| --- | --- | --- |\n| a | b | d |\n"),
			[]string{"and-missing@capabilities/c/features/f.md:22", "keyword-count@capabilities/c/features/f.md:19", "term-unresolved@capabilities/c/features/f.md:21"}},
		{"keyword is not unmarked text", func() map[string]string {
			f := withRule("When      lint reads <id>\nThen      x\n  And     y\n", "| id | outcome |\n| --- | --- |\n| a | x |\n")
			f["capabilities/c/glossary.md"] += "\n## <a id=\"c.then\"></a> then\n\n- Definition: x\n- Contact: x\n- Status: defined\n"
			return f
		}(), nil},
		{"level files and section levels", map[string]string{
			"system.md":                    "Nothing here.\n",
			"capabilities/c/capability.md": "## <a id=\"c\"></a> C\n",
			"capabilities/c/features/f.md": "# F\n\n# <a id=\"c.f\"></a> F\n\n- Contact: x\n\n# Goal\n\ng\n\n## <a id=\"c.f.r\"></a> Rule: R\n\n## Rationale\n\nr\n",
			"glossary.md":                  "# Glossary\n\n## <a id=\"system.r\"></a> Rule: r\n",
		}, []string{
			"element-missing@system.md:1", "element-missing@capabilities/c/capability.md:1", "element-unplaced@capabilities/c/capability.md:1",
			"element-missing@capabilities/c/features/f.md:1", "section-level@capabilities/c/features/f.md:7", "section-missing@capabilities/c/features/f.md:3",
			"section-level@capabilities/c/features/f.md:13", "rule-block-count@capabilities/c/features/f.md:11", "element-unplaced@glossary.md:3",
		}},
		{"two blocks", withRule("When      lint reads <id>\nThen      x\n  And     y\n```\n\n```guarded\nWhen x\n", "| id | outcome |\n| --- | --- |\n| a | x |\n"),
			[]string{"rule-block-count@capabilities/c/features/f.md:13"}},
		{"structure", map[string]string{
			"system.md":                    strings.Replace(baseSystem, `id="system"`, `id="proviso"`, 1),
			"capabilities/c/capability.md": baseCapability + "\n## Scope\n\nx\n\n## Goal\n\n?\n",
			"capabilities/c/features/f.md": "# <a id=\"c.f\"></a> F\n\n- contact: Ann\n- Contact: Ann\n- Contact:\n\n## Goal\n\nx\n\n## Rationale\n\nr\n\n### <a id=\"c.f.x\"></a> Rule: X\n\n## Rule: forgotten <a id=\"Bad\"></a>\n",
		}, []string{
			"anchor-misplaced@capabilities/c/features/f.md:17", "element-unplaced@capabilities/c/features/f.md:15", "field-repeated@capabilities/c/features/f.md:5",
			"field-unknown@capabilities/c/features/f.md:3", // the forgotten rule lies within an unplaced element
			"heading-unknown@capabilities/c/capability.md:11", "section-repeated@capabilities/c/capability.md:15", "undecided@capabilities/c/capability.md:15",
			"system-identifier@system.md:1",
		}},
		{"layout", map[string]string{
			"system.md":                      baseSystem,
			"notes.md":                       "x",
			"capabilities/c/features/x/a.md": "x",
			"glossary.md":                    "## <a id=\"system.x\"></a> x\n\n- Definition: d\n- Contact: c\n- Status: defined\n",
			"parameters.md":                  "",
			"capabilities/d/glossary.md":     "# Glossary\n",
		}, []string{
			"capabilities/c/features/x/a.md:1", "capabilities/c:1", "capabilities/d/glossary.md:1", "capabilities/d:1", "capability-file-missing", "capability-file-missing", "file-unknown", "file-unknown", "glossary.md:1", "notes.md:1", "parameters.md:1", "sibling-empty", "sibling-empty", "title-missing",
		}},
		{"glossary entries", map[string]string{
			"system.md":     baseSystem,
			"references.md": "# References\n\n## <a id=\"system.ref.cm\"></a> CommonMark\n\n- Version: 0.31\n- Link: https://spec.commonmark.org\n- Contact: Ann\n",
			"glossary.md": "# Glossary\n\n" +
				"## <a id=\"system.a\"></a> a\n\n- Definition: ?\n- Contact: x\n- Status: open\n\n" + // no question
				"## <a id=\"system.b\"></a> b\n\n- Definition: d\n- Contact: x\n- Status: defined\n- Question: q?\n\n" +
				"## <a id=\"system.c\"></a> c\n\n- Defined by: [CM](references.md#system.ref.cm)\n- Deviations: none\n- Contact: x\n- Status: defined\n\n" +
				"## <a id=\"system.d\"></a> d\n\n- Defined by: [CM](references.md#system.ref.nope)\n- Contact: x\n- Status: draft\n\n" +
				"## <a id=\"system.e\"></a> e\n\n- Deviations: x\n- Contact: x\n- Status: defined\n- Source: y\n- Aliases: A\n",
		}, []string{
			"definition-count@glossary.md:29", "defined-by-unresolved@glossary.md:25", "defined-entry-open@glossary.md:9", "deviations-without-reference@glossary.md:29",
			"field-unknown@glossary.md:34", "open-entry-incomplete@glossary.md:3", "status-unknown@glossary.md:27",
			"term-ambiguous@glossary.md:3", "term-ambiguous@glossary.md:29", "undecided@glossary.md:5",
		}},
		{"duplicates and ambiguity", func() map[string]string {
			f := base()
			f["capabilities/c/glossary.md"] += "\n## <a id=\"c.f\"></a> tree\n\n- Definition: x\n- Contact: x\n- Status: defined\n"
			f["capabilities/c/parameters.md"] += "\n## <a id=\"c.lint\"></a> lint\n\n- Configurable: no\n- Rationale: r\n"
			return f
		}(), []string{
			"identifier-duplicate@capabilities/c/features/f.md:1", "identifier-duplicate@capabilities/c/glossary.md:10",
			"reference-ambiguous@capabilities/c/contracts/cli.md:3", "reference-ambiguous@capabilities/c/parameters.md:9",
			"term-ambiguous@capabilities/c/glossary.md:10", "term-ambiguous@glossary.md:3",
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := kinds(lintFiles(c.files))
			if c.name == "layout" {
				got = flatten(got)
			}
			slices.Sort(got)
			want := slices.Clone(c.want)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("got  %v\nwant %v", got, want)
			}
		})
	}
}

// flatten splits "kind@place" into its two parts, so that a case can list
// kinds and places without pairing them.
func flatten(ks []string) []string {
	var out []string
	for _, k := range ks {
		a, b, _ := strings.Cut(k, "@")
		out = append(out, a, b)
	}
	return out
}

func TestOrder(t *testing.T) {
	fs := []Finding{
		{Kind: "x", File: "b.md", Line: 3}, {Kind: "x", File: "a.md", Line: 9}, {Kind: "x", File: "a.md", Line: 2},
		{Kind: "identifier-duplicate", File: "a.md", Line: 4}, {Kind: "anchor-misplaced", File: "a.md", Line: 4},
		{Kind: "x", File: "B.md", Line: 1},
	}
	Sort(fs)
	var got []string
	for _, f := range fs {
		got = append(got, f.File+":"+strconv.Itoa(f.Line)+" "+f.Kind)
	}
	want := []string{"B.md:1 x", "a.md:2 x", "a.md:4 anchor-misplaced", "a.md:4 identifier-duplicate", "a.md:9 x", "b.md:3 x"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRelatedAtHeading(t *testing.T) {
	f := base()
	f["capabilities/c/glossary.md"] += "\n## <a id=\"c.f\"></a> other\n\n- Definition: x\n- Contact: x\n- Status: defined\n"
	for _, fd := range lintFiles(f) {
		if fd.Kind != "identifier-duplicate" || fd.File != "proviso/capabilities/c/features/f.md" {
			continue
		}
		if want := []Pos{{"proviso/capabilities/c/glossary.md", 10}}; !slices.Equal(fd.Related, want) {
			t.Errorf("related %v, want %v", fd.Related, want)
		}
		return
	}
	t.Error("no identifier-duplicate in f.md")
}
