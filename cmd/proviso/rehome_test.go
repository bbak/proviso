package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// glossaryMoved holds a term lifted to the system that kept the
// identifier of its capability, and a link to it.
const glossaryMoved = "# Glossary\n\n## <a id=\"c.thing\"></a> thing\n\n- Definition: A thing.\n- Contact: Ann\n- Status: defined\n\n" +
	"## <a id=\"system.other\"></a> other\n\n- Definition: Not a [thing](#c.thing).\n- Contact: Ann\n- Status: defined\n"

func TestRehome(t *testing.T) {
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		return status, out.String(), errOut.String()
	}
	glossary := filepath.Join("proviso", "glossary.md")
	manifest := filepath.Join("proviso", "manifest.sum")
	read := func(p string) string {
		data, _ := os.ReadFile(p)
		return string(data)
	}

	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/glossary.md": glossaryMoved})
	if status, out, errOut := call("rehome"); status != 2 || out != "" || !strings.Contains(errOut, "no proviso/manifest.sum") {
		t.Errorf("without a manifest: %d %q %q", status, out, errOut)
	}
	if status, _, errOut := call("manifest"); status != 0 {
		t.Fatalf("manifest: %d %q", status, errOut)
	}
	recorded := read(manifest)

	if err := os.WriteFile(filepath.Join("proviso", ".amendment.md"), []byte("- x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if status, out, errOut := call("rehome"); status != 2 || out != "" || !strings.Contains(errOut, ".amendment.md") {
		t.Errorf("with a note: %d %q %q", status, out, errOut)
	}
	os.Remove(filepath.Join("proviso", ".amendment.md"))

	if status, out, errOut := call("rehome"); status != 0 || out != "c.thing → system.thing\n" || errOut != "" {
		t.Errorf("plan: %d %q %q", status, out, errOut)
	}
	if status, out, _ := call("rehome", "--json"); status != 0 || !strings.Contains(out, `"from": "c.thing",`) || !strings.Contains(out, `"collisions": []`) {
		t.Errorf("plan, json: %d %q", status, out)
	}
	if read(glossary) != glossaryMoved || read(manifest) != recorded {
		t.Error("a plan wrote")
	}

	// A file that cannot be opened for writing stops every write.
	if err := os.Chmod(manifest, 0o444); err != nil {
		t.Fatal(err)
	}
	if status, out, errOut := call("rehome", "--write"); status != 2 || out != "" || !strings.Contains(errOut, "proviso/manifest.sum") {
		t.Errorf("read-only manifest: %d %q %q", status, out, errOut)
	}
	if err := os.Chmod(manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	if read(glossary) != glossaryMoved || read(manifest) != recorded {
		t.Error("a refused write wrote")
	}

	if status, out, errOut := call("rehome", "--write", "--json"); status != 0 || !strings.Contains(out, `"to": "system.thing"`) {
		t.Fatalf("write: %d %q %q", status, out, errOut)
	}
	if want := strings.ReplaceAll(glossaryMoved, "c.thing", "system.thing"); read(glossary) != want {
		t.Errorf("glossary %q", read(glossary))
	}
	if status, out, _ := call("diff"); status != 0 || out != "" {
		t.Errorf("after the write, the diff: %d %q", status, out)
	}
	if status, out, _ := call("rehome", "--write"); status != 0 || out != "" {
		t.Errorf("nothing left to rename: %d %q", status, out)
	}

	// The mentions of a renamed identifier are listed, and renamed only
	// where the mention list on stdin names them.
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/glossary.md": glossaryMoved + "\n## <a id=\"system.third\"></a> third\n\n- Definition: Not `c.thing`, nor `c.thing`.\n- Contact: Ann\n- Status: defined\n\n## <a id=\"system.fourth\"></a> fourth\n\n- Definition: Not `c.thing`.\n- Contact: Ann\n- Status: defined\n"})
	call("manifest")
	before := read(glossary)
	if status, out, _ := call("rehome"); status != 0 || !strings.Contains(out, "mention: proviso/glossary.md:17: c.thing, 2 times, in system.third\n") || !strings.Contains(out, "mention: proviso/glossary.md:23: c.thing, once, in system.fourth\n") {
		t.Errorf("mentions: %d %q", status, out)
	}
	if status, out, _ := call("rehome", "--json"); status != 0 || !strings.Contains(out, `"line": 17,`) || !strings.Contains(out, `"element": "system.third"`) {
		t.Errorf("mentions, json: %d %q", status, out)
	}
	for _, in := range []string{"{", `{"mentions": [{"identifier": "c.thing", "file": "proviso/glossary.md", "line": 19}]}`} {
		stdin = strings.NewReader(in)
		if status, out, errOut := call("rehome", "--write", "--mentions"); status != 2 || out != "" || !strings.Contains(errOut, "--mentions") {
			t.Errorf("%q: %d %q %q", in, status, out, errOut)
		}
		if read(glossary) != before {
			t.Errorf("%q wrote", in)
		}
	}
	stdin = strings.NewReader(`{"mentions": [{"identifier": "c.thing", "file": "proviso/glossary.md", "line": 17}]}`)
	if status, _, errOut := call("rehome", "--write", "--mentions"); status != 0 {
		t.Errorf("picked: %d %q", status, errOut)
	}
	stdin = os.Stdin
	if s := read(glossary); !strings.Contains(s, "Not `system.thing`, nor `system.thing`.") || !strings.Contains(s, "Not `c.thing`.\n") || strings.Contains(s, `id="c.thing"`) {
		t.Errorf("picked: %q", s)
	}
	if status, out, _ := call("diff"); status != 0 || out != "" {
		t.Errorf("after picking, the diff: %d %q", status, out)
	}

	// A rename that takes an identifier refuses the run.
	inRepo(t, map[string]string{"proviso/system.md": system, "proviso/glossary.md": glossaryMoved + "\n## <a id=\"system.thing\"></a> thing2\n\n- Definition: A thing.\n- Contact: Ann\n- Status: defined\n"})
	call("manifest")
	before = read(glossary)
	if status, out, _ := call("rehome", "--write"); status != 1 || !strings.Contains(out, "collision: system.thing would be taken by c.thing") {
		t.Errorf("collision: %d %q", status, out)
	}
	if read(glossary) != before {
		t.Error("a collision wrote")
	}
}
