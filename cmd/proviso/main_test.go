package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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
		{"uppercase name", map[string]string{"proviso/system.md": system, "proviso/Glossary.md": "x"}, []string{"lint"}, 1, "proviso/Glossary.md:1: error name-malformed: "},
		{"decomposed name", map[string]string{"proviso/system.md": system, "proviso/ü.md": "x"}, []string{"lint"}, 1, "proviso/ü.md:1: error file-unknown: "},
		{"carriage return", map[string]string{"proviso/system.md": strings.ReplaceAll(system, "\n", "\r\n")}, []string{"lint"}, 1, "proviso/system.md:1: error line-ending: "},
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
	if !strings.HasSuffix(r.Root, "/proviso") || strings.Contains(r.Root, "\\") || len(r.Findings) != 1 {
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

func TestReportRoot(t *testing.T) {
	if runtime.GOOS != "windows" {
		if got := reportRoot("/repo/proviso"); got != "/repo/proviso" {
			t.Errorf("got %q", got)
		}
		if got := reportRoot("/ü/proviso"); got != "/ü/proviso" {
			t.Errorf("a decomposed name: got %q", got)
		}
		return
	}
	cases := map[string]string{
		`c:\repo\proviso`:        "C:/repo/proviso",
		`C:\repo\proviso`:        "C:/repo/proviso",
		`\\server\share\proviso`: "//server/share/proviso",
		"c:\\u\u0308\\proviso":   "C:/\u00fc/proviso",
	}
	for in, want := range cases {
		if got := reportRoot(in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestModelVersion(t *testing.T) {
	for manifest, status := range map[string]int{"model 0\n": 0, "model 1\n": 2, "": 2, "model 0\r\n": 1, "model 0\n\xff\n": 2} {
		inRepo(t, map[string]string{"proviso/system.md": system, "proviso/manifest.sum": manifest})
		var out, errOut bytes.Buffer
		if got := run([]string{"lint", "--json"}, &out, &errOut); got != status || status == 2 && (out.Len() > 0 || errOut.Len() == 0) {
			t.Errorf("%q: status %d, stdout %q, stderr %q", manifest, got, out.String(), errOut.String())
		}
	}
}

func TestBaselineCommands(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		return status, out.String(), errOut.String()
	}
	if status, out, _ := call("diff", "--json"); status != 1 || !strings.Contains(out, `"revision": null`) || !strings.Contains(out, `"change": "added"`) {
		t.Errorf("without a manifest: %d %s", status, out)
	}
	if status, out, errOut := call("manifest"); status != 0 || out != "" || errOut != "" {
		t.Fatalf("manifest: %d %q %q", status, out, errOut)
	}
	data, err := os.ReadFile(filepath.Join("proviso", "manifest.sum"))
	if err != nil || !strings.HasPrefix(string(data), "model 0\nsystem\theading\th1:") {
		t.Fatalf("%q %v", data, err)
	}
	if status, out, _ := call("diff"); status != 0 || out != "" {
		t.Errorf("unchanged, text: %d %q", status, out)
	}
	if status, out, _ := call("diff", "--json"); status != 0 || !strings.Contains(out, `"elements": []`) {
		t.Errorf("unchanged, json: %d %q", status, out)
	}
	if status, _, _ := call("manifest", "--json"); status != 2 {
		t.Error("manifest takes no --json")
	}

	// A changed part needs former text, which a tree in no git repository
	// cannot give.
	os.WriteFile(filepath.Join("proviso", "system.md"), []byte(strings.Replace(system, "\ng\n", "\nh\n", 1)), 0o644)
	if status, out, errOut := call("diff", "--json"); status != 2 || out != "" || errOut == "" {
		t.Errorf("no former text: %d %q %q", status, out, errOut)
	}

	os.WriteFile(filepath.Join("proviso", "manifest.sum"), append(data, "<<<<<<< HEAD\n"...), 0o644)
	if status, out, errOut := call("diff"); status != 2 || out != "" || !strings.Contains(errOut, "<<<<<<< HEAD") {
		t.Errorf("malformed: %d %q %q", status, out, errOut)
	}

	os.WriteFile(filepath.Join("proviso", "notes.md"), []byte("x"), 0o644)
	for _, cmd := range []string{"manifest", "diff"} {
		status, out, errOut := call(cmd)
		if status != 2 || out != "" || !strings.Contains(errOut, "error file-unknown") {
			t.Errorf("%s on a tree with an error: %d %q %q", cmd, status, out, errOut)
		}
	}
	after, _ := os.ReadFile(filepath.Join("proviso", "manifest.sum"))
	if !bytes.Equal(after, append(data, "<<<<<<< HEAD\n"...)) {
		t.Error("a refused manifest changed the former one")
	}
}

func TestImpactCommand(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	var out, errOut bytes.Buffer
	if status := run([]string{"impact", "--json"}, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	var r struct {
		Revision   *string
		Candidates []any
		Scopes     []struct {
			ID         string
			Candidates []string
		}
	}
	if err := json.Unmarshal(out.Bytes(), &r); err != nil || r.Revision != nil || r.Candidates == nil || len(r.Candidates) != 0 || len(r.Scopes) != 1 || r.Scopes[0].ID != "system" {
		t.Errorf("%v %s", err, out.String())
	}
	out.Reset()
	if status := run([]string{"impact"}, &out, &errOut); status != 0 || out.Len() != 0 {
		t.Errorf("text without candidates: %d %q", status, out.String())
	}
	os.WriteFile(filepath.Join("proviso", "notes.md"), []byte("x"), 0o644)
	out.Reset()
	errOut.Reset()
	if status := run([]string{"impact", "--json"}, &out, &errOut); status != 2 || out.Len() != 0 || !strings.Contains(errOut.String(), "file-unknown") {
		t.Errorf("a tree with an error: %d %q %q", status, out.String(), errOut.String())
	}
}

func TestImpactDamagedManifest(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/manifest.sum": "model 0\n<<<<<<< HEAD\n"})
	var out, errOut bytes.Buffer
	if status := run([]string{"impact", "--json"}, &out, &errOut); status != 2 || out.Len() != 0 || !strings.Contains(errOut.String(), "<<<<<<< HEAD") {
		t.Errorf("%d %q %q", status, out.String(), errOut.String())
	}
}
