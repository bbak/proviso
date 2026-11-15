package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// oneJSON reports whether b holds exactly one JSON value and nothing
// after it but white space.
func oneJSON(b []byte) bool {
	dec := json.NewDecoder(bytes.NewReader(b))
	var v any
	if err := dec.Decode(&v); err != nil {
		return false
	}
	return dec.Decode(&v) == io.EOF
}

// baseline.pack.output, coherence.impact.output, query.show.match and
// query.show.output: nothing but the report goes to stdout.
func TestStdoutOnly(t *testing.T) {
	inRepo(t, map[string]string{"proviso/system.md": system})
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		return status, out.String(), errOut.String()
	}
	if status, _, errOut := call("manifest"); status != 0 {
		t.Fatalf("manifest: %d %q", status, errOut)
	}
	for _, args := range [][]string{
		{"pack", "--json"},
		{"pack", "--json", "--element", "system"},
		{"impact", "--json"},
		{"show", "--json", "system"},
	} {
		if status, out, errOut := call(args...); status != 0 || !oneJSON([]byte(out)) {
			t.Errorf("%v: %d %q %q", args, status, out, errOut)
		}
	}
	if status, out, errOut := call("show", "--json", "nosuch"); status != 1 || !oneJSON([]byte(out)) || errOut != "" {
		t.Errorf("show, no element, json: %d %q %q", status, out, errOut)
	}
	if status, out, _ := call("show", "nosuch"); status != 1 || out != "" {
		t.Errorf("show, no element, text: %d %q", status, out)
	}
	// show runs whatever lint finds, and writes none of it.
	if err := os.WriteFile(filepath.Join("proviso", "notes.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := "---\nid: \"system\"\n"
	if status, out, _ := call("show", "system"); status != 0 || !strings.HasPrefix(out, want) || strings.Contains(out, "file-unknown") {
		t.Errorf("show, a tree with an error: %d %q", status, out)
	}
	if status, out, _ := call("show", "--json", "system"); status != 0 || !oneJSON([]byte(out)) || strings.Contains(out, "file-unknown") {
		t.Errorf("show, a tree with an error, json: %d %q", status, out)
	}
}
