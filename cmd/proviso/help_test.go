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
		{"diff", "HEAD~1"},
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
		cmd    string
		args   []string
		asJSON bool
		since  string
	}{
		{"lint", []string{"--json"}, true, ""},
		{"lint", []string{"--json", "--json"}, true, ""},
		{"diff", []string{"--since", "HEAD~1"}, false, "HEAD~1"},
		{"pack", []string{"--json", "--since", "a1b2c3d"}, true, "a1b2c3d"},
		{"diff", []string{"--since", "HEAD~1", "--json"}, true, "HEAD~1"},
	}
	for _, c := range cases {
		var errOut bytes.Buffer
		asJSON, since, ok := readArgs(c.cmd, c.cmd, c.args, &errOut)
		if !ok || asJSON != c.asJSON || since != c.since {
			t.Errorf("%s %q: %v %q %v (stderr %q)", c.cmd, c.args, asJSON, since, ok, errOut.String())
		}
	}
}
