package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// showTree is a small tree with a term in two capabilities, a parameter,
// a contract with a member and a capability in the making.
var showTree = map[string]string{
	"proviso/system.md":                            "# <a id=\"system\"></a> Shop\n\n## Goal\n\nx\n",
	"proviso/glossary.md":                          "# Glossary\n\n## <a id=\"system.spec-tree\"></a> spec tree\n\n- Definition: x\n- Aliases: tree\n\n",
	"proviso/parameters.md":                        "# Parameters\n\n## <a id=\"system.tree-layout\"></a> tree_layout\n\n- Configurable: no\n",
	"proviso/contracts/cli.md":                     "# CLI\n\n## <a id=\"system.cli.diff\"></a> diff\n\nx\n\n### <a id=\"system.cli.diff.since\"></a> since\n\ny\n",
	"proviso/capabilities/lint/capability.md":      "# <a id=\"lint\"></a> Lint\n",
	"proviso/capabilities/lint/glossary.md":        "# Glossary\n\n## <a id=\"lint.finding\"></a> finding\n\n- Definition: x\n",
	"proviso/capabilities/lint/features/run.md":    "# <a id=\"lint.run\"></a> Run\n\n## <a id=\"lint.run.exit\"></a> Rule: Exit\n\nx\n\n## <a id=\"lint.run.odd\"></a> Odd\n",
	"proviso/capabilities/coherence/capability.md": "# <a id=\"coherence\"></a> Coherence\n",
	"proviso/capabilities/coherence/glossary.md":   "# Glossary\n\n## <a id=\"coherence.finding\"></a> finding\n\n- Definition: x\n",
	"proviso/capabilities/new/features/a.md":       "# <a id=\"new.a\"></a> A\n",
}

func TestShow(t *testing.T) {
	cases := []struct {
		args   []string
		status int
		ids    []string // with "type/parent" checked through JSON
	}{
		{[]string{"lint.run.exit"}, 0, []string{"lint.run.exit rule lint.run"}},
		{[]string{"spec tree"}, 0, []string{"system.spec-tree glossary entry system"}},
		{[]string{"tree"}, 0, []string{"system.spec-tree glossary entry system"}},
		{[]string{"Spec  Tree"}, 0, []string{"system.spec-tree glossary entry system"}},
		{[]string{"finding"}, 0, []string{"coherence.finding glossary entry coherence", "lint.finding glossary entry lint"}},
		{[]string{"finding", "--in", "lint"}, 0, []string{"lint.finding glossary entry lint"}},
		{[]string{"diff.since"}, 0, []string{"system.cli.diff.since contract element system.cli.diff"}},
		{[]string{"diff"}, 0, []string{"system.cli.diff contract element system"}},
		{[]string{"tree_layout"}, 0, []string{"system.tree-layout parameter system"}},
		{[]string{"lint.run.odd"}, 0, []string{"lint.run.odd <nil> lint.run"}},
		{[]string{"system"}, 0, []string{"system system <nil>"}},
		{[]string{"lint"}, 0, []string{"lint capability system"}},
		{[]string{"lint.run"}, 0, []string{"lint.run feature lint"}},
		{[]string{"new.a", "--in", "new"}, 0, []string{"new.a feature <nil>"}},
		{[]string{"since"}, 1, nil},
		{[]string{"Lint.Run.Exit"}, 1, nil},
		{[]string{"Tree_Layout"}, 1, nil},
		{[]string{"spec"}, 1, nil},
		{[]string{"Rule: Exit"}, 1, nil},
		{[]string{"system.cli.diff.since", "--in", "lint"}, 0, []string{"system.cli.diff.since contract element system.cli.diff"}},
		{[]string{"coherence.finding", "--in", "lint"}, 1, nil},
		{[]string{"x", "--in", "nosuch"}, 2, nil},
		{[]string{"x", "--in", "Lint"}, 2, nil},
	}
	for _, c := range cases {
		inRepo(t, showTree)
		var out, errOut bytes.Buffer
		args := append([]string{"show", "--json"}, c.args...)
		status := run(args, &out, &errOut)
		if status != c.status {
			t.Errorf("%q: status %d, want %d (stderr %q)", c.args, status, c.status, errOut.String())
			continue
		}
		if status == 2 {
			continue
		}
		if errOut.Len() > 0 {
			t.Errorf("%q: stderr %q", c.args, errOut.String())
		}
		var v struct {
			Root     string
			Elements []struct {
				ID, Heading, File, Text string
				Type, Parent            *string
				Line                    int
			}
		}
		if err := json.Unmarshal(out.Bytes(), &v); err != nil || v.Root == "" || v.Elements == nil {
			t.Errorf("%q: %v %q", c.args, err, out.String())
			continue
		}
		var got []string
		for _, e := range v.Elements {
			str := func(p *string) string {
				if p == nil {
					return "<nil>"
				}
				return *p
			}
			got = append(got, e.ID+" "+str(e.Type)+" "+str(e.Parent))
		}
		if strings.Join(got, "|") != strings.Join(c.ids, "|") {
			t.Errorf("%q: %q, want %q", c.args, got, c.ids)
		}
	}
}

func TestShowText(t *testing.T) {
	inRepo(t, showTree)
	var out, errOut bytes.Buffer
	if status := run([]string{"show", "finding"}, &out, &errOut); status != 0 || errOut.Len() > 0 {
		t.Fatalf("status %d, stderr %q", status, errOut.String())
	}
	want := `---
id: "coherence.finding"
type: "glossary entry"
heading: "finding"
file: "proviso/capabilities/coherence/glossary.md"
line: 3
parent: "coherence"
---

## <a id="coherence.finding"></a> finding

- Definition: x

---
id: "lint.finding"
type: "glossary entry"
heading: "finding"
file: "proviso/capabilities/lint/glossary.md"
line: 3
parent: "lint"
---

## <a id="lint.finding"></a> finding

- Definition: x
`
	if out.String() != want {
		t.Errorf("stdout %q, want %q", out.String(), want)
	}
	out.Reset()
	errOut.Reset()
	if status := run([]string{"show", "nosuch", "--in", "lint"}, &out, &errOut); status != 1 || out.Len() > 0 || strings.Count(errOut.String(), "\n") != 1 || !strings.Contains(errOut.String(), `"lint"`) {
		t.Errorf("no match: status %d, stdout %q, stderr %q", status, out.String(), errOut.String())
	}
}

func TestShowNeedsQuery(t *testing.T) {
	for _, args := range [][]string{{"show"}, {"show", "--json"}, {"show", "--in", "lint"}} {
		t.Chdir(t.TempDir()) // refused before any tree is looked for
		var out, errOut bytes.Buffer
		if status := run(args, &out, &errOut); status != 2 || out.Len() > 0 || strings.Count(errOut.String(), "\n") != 1 || !strings.Contains(errOut.String(), helpHint) {
			t.Errorf("%q: status %d, stdout %q, stderr %q", args, status, out.String(), errOut.String())
		}
	}
}

func TestShowLintErrors(t *testing.T) {
	files := map[string]string{}
	for k, v := range showTree {
		files[k] = v
	}
	files["proviso/capabilities/lint/glossary.md"] += "## <a id=\"lint.finding\"></a> finding\n\n- Definition: y\n"
	inRepo(t, files)
	var out, errOut bytes.Buffer
	if status := run([]string{"show", "lint.finding"}, &out, &errOut); status != 0 || errOut.Len() > 0 || strings.Count(out.String(), "id: \"lint.finding\"") != 2 {
		t.Errorf("status %d, stdout %q, stderr %q", status, out.String(), errOut.String())
	}
}
