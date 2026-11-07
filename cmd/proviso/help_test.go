package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"--help"}, {"-h"}} {
		var out, errOut bytes.Buffer
		if status := run(args, &out, &errOut); status != 0 || out.String() != helpPage || errOut.Len() != 0 {
			t.Errorf("%v: status %d, stdout %q, stderr %q", args, status, out.String(), errOut.String())
		}
	}
	var out, errOut bytes.Buffer
	if status := run(nil, &out, &errOut); status != 2 || out.Len() != 0 || errOut.String() != helpPage {
		t.Errorf("no command: status %d, stdout %q, stderr %q", status, out.String(), errOut.String())
	}
}

func TestHelpPageNamesEverything(t *testing.T) {
	for cmd, opts := range commandOptions {
		if !strings.Contains(helpPage, "  "+cmd) {
			t.Errorf("the page does not name %s", cmd)
		}
		for _, o := range opts {
			if !strings.Contains(helpPage, o) {
				t.Errorf("the page does not name %s", o)
			}
		}
	}
}

func TestRefused(t *testing.T) {
	cases := [][]string{
		{"Lint"}, {"--version"}, {"build"}, {""},
		{"version", "0.1.0"}, {"init", "--json"}, {"judge", "--json"}, {"manifest", "--json"},
		{"help", "diff"}, {"--help", "--json"}, {"lint", "--since", "HEAD~1"}, {"diff", "--help"},
		{"diff", "--since"}, {"pack", "--since", ""}, {"diff", "--since", "--json"},
		{"diff", "--since", "-x"}, {"diff", "--since", "HEAD~1", "--since", "HEAD~2"},
		{"diff", "HEAD~1"}, {"rehome", "--since", "HEAD~1"}, {"lint", "--write"},
		{"diff", "--element", "lint.run.exit"}, {"pack", "--element"}, {"pack", "--element", "-x"},
		{"pack", "--element", "a", "--element", "b"}, {"pack", "--since", "HEAD~1", "--element", "a"},
		{"pack", "--element", "a", "--since", "HEAD~1"},
		{"lint", "spec tree"}, {"show", ""}, {"show", "-x"}, {"show", "lint.run.exit", "--in"},
		{"show", "a", "--in", "lint", "--in", "query"}, {"show", "spec", "tree"}, {"show", "a", "--since", "HEAD~1"},
		{"show", "a", "--in", "-x"},
	}
	for _, args := range cases {
		t.Chdir(t.TempDir()) // nothing to check: a refusal reads no tree
		var out, errOut bytes.Buffer
		status := run(args, &out, &errOut)
		msg := errOut.String()
		if status != 2 || out.Len() != 0 || strings.Count(msg, "\n") != 1 || !strings.Contains(msg, `"proviso help" shows the help page`) {
			t.Errorf("%q: status %d, stdout %q, stderr %q", args, status, out.String(), msg)
		}
		if !strings.Contains(msg, "proviso "+args[0]+":") && !strings.Contains(msg, `"`+args[0]+`" is no command`) {
			t.Errorf("%q: stderr %q does not name the command", args, msg)
		}
	}
}

func TestReadArgs(t *testing.T) {
	cases := []struct {
		cmd  string
		args []string
		want options
	}{
		{"lint", []string{"--json"}, options{json: true}},
		{"lint", []string{"--json", "--json"}, options{json: true}},
		{"diff", []string{"--since", "HEAD~1"}, options{since: "HEAD~1"}},
		{"pack", []string{"--json", "--since", "a1b2c3d"}, options{json: true, since: "a1b2c3d"}},
		{"pack", []string{"--json", "--element", "lint.run.exit"}, options{json: true, element: "lint.run.exit"}},
		{"diff", []string{"--since", "HEAD~1", "--json"}, options{json: true, since: "HEAD~1"}},
		{"rehome", []string{"--json", "--write"}, options{json: true, write: true}},
		{"rehome", nil, options{}},
		{"show", []string{"spec tree"}, options{query: "spec tree"}},
		{"show", []string{"lint.run.exit", "--in", "lint"}, options{query: "lint.run.exit", in: "lint"}},
		{"show", []string{"--in", "lint", "--json", "lint.run.exit", "--json"}, options{json: true, query: "lint.run.exit", in: "lint"}},
		{"show", []string{"--in", "lint"}, options{in: "lint"}},
	}
	for _, c := range cases {
		var errOut bytes.Buffer
		got, ok := readArgs(c.cmd, c.cmd, c.args, &errOut)
		if !ok || got != c.want {
			t.Errorf("%s %q: %+v %v (stderr %q)", c.cmd, c.args, got, ok, errOut.String())
		}
	}
}
