package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const system = "# <a id=\"system\"></a> S\n\n## Goal\n\ng\n\n## Rationale\n\nr\n\n## Scope\n\ns\n\n## Out of scope\n\no\n\n## Audience\n\na\n\n## Design boundary\n\nd\n"

func inRepo(t *testing.T, files map[string]string) {
	t.Helper()
	repo := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(repo)
}

func TestRun(t *testing.T) {
	cases := []struct {
		name   string
		files  map[string]string
		args   []string
		status int
		stdout string // a substring, or "" for nothing written
	}{
		{"clean text", map[string]string{"proviso/system.md": system}, []string{"lint"}, 0, ""},
		{"clean json", map[string]string{"proviso/system.md": system}, []string{"lint", "--json"}, 0, `"findings": []`},
		{"error text", map[string]string{"proviso/system.md": system, "proviso/notes.md": "x"}, []string{"lint"}, 1, "proviso/notes.md:1: error file-unknown: "},
		{"ask only", map[string]string{"proviso/system.md": strings.Replace(system, "\ng\n", "\n?\n", 1)}, []string{"lint"}, 0, "ask undecided"},
		{"no root", map[string]string{"system.md": system}, []string{"lint"}, 2, ""},
		{"unreadable", map[string]string{"proviso/system.md": system, "proviso/glossary.md": "\xff"}, []string{"lint", "--json"}, 2, ""},
		{"bad option", map[string]string{"proviso/system.md": system}, []string{"lint", "--yaml"}, 2, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inRepo(t, c.files)
			var out, errOut bytes.Buffer
			status := run(c.args, &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d (stderr %q)", status, c.status, errOut.String())
			}
			if c.stdout == "" && out.Len() > 0 || !strings.Contains(out.String(), c.stdout) {
				t.Errorf("stdout %q, want %q", out.String(), c.stdout)
			}
			if c.status == 2 && errOut.Len() == 0 {
				t.Error("nothing on stderr")
			}
		})
	}
}

func TestJSONReport(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/notes.md": "x"})
	var out, errOut bytes.Buffer
	run([]string{"lint", "--json"}, &out, &errOut)
	var r struct {
		Root     string
		Findings []map[string]any
	}
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(r.Root, "/proviso") || len(r.Findings) != 1 {
		t.Fatalf("%+v", r)
	}
	f := r.Findings[0]
	if f["kind"] != "file-unknown" || f["severity"] != "error" || f["file"] != "proviso/notes.md" || f["line"] != 1.0 {
		t.Errorf("%+v", f)
	}
	if _, ok := f["related"]; ok {
		t.Error("an empty related is left out")
	}
}
