package main

import (
	"fmt"
	"io"
	"slices"
)

// helpPage is cli.cli.help.page: what Proviso is, every command with its
// options, first those engineers run and then those the skills run, in
// the order the page names them, the exit statuses the commands share,
// and where the documentation is. Its wording is for people and not part
// of the contract.
const helpPage = `proviso keeps one specification of a system true across many amendments.

Usage: proviso <command> [options]

Commands engineers run:
  init                Write the files a repository is missing; run in
                      the root of a git repository.
  update              Bring the installed skills and agent definitions
                      to this release; run in the root of a git
                      repository.
  version             Write the Proviso version and the model version,
                      and check what is installed against them.
  lint [--json]       Check the spec tree against the model's notation.
  diff [--json] [--since <revision>]
                      Write the element diff against the baseline.
  impact [--json]     Write the rules and invariants the amendment may
                      affect.
  verify [--json]     Tell whether the spec tree may become the baseline.
  help, --help, -h    Write this page.

Commands the skills run:
  pack [--json] [--since <revision>]
                      Write the context pack the agents build from.
  judge               Record the check result on standard input in the
                      coherence file.
  manifest            Record the spec tree as the baseline.

Options:
  --json              Write JSON instead of text.
  --since <revision>  Compare with the spec tree at that git revision
                      instead of the baseline.

Exit statuses:
  0                   The command did what it was asked.
  1                   lint, diff or verify found what it reports, or
                      judge refused its input.
  2                   The command could not run, or was misused.

Documentation: https://github.com/bbak/proviso/blob/main/docs/commands.md
`

// helpHint ends every line that refuses a command or an argument.
const helpHint = `"proviso help" shows the help page`

// commandOptions is, for every command of cli.help.command, the options
// it takes; "--since" takes a revision after it.
var commandOptions = map[string][]string{
	"init":     nil,
	"update":   nil,
	"version":  nil,
	"lint":     {"--json"},
	"manifest": nil,
	"diff":     {"--json", "--since"},
	"impact":   {"--json"},
	"pack":     {"--json", "--since"},
	"judge":    nil,
	"verify":   {"--json"},
	"help":     nil,
}

// readArgs reads the arguments of the command name, which is cmd as it
// was given, following cli.help.argument. It writes to stderr one line
// naming the command and the first argument it does not take.
func readArgs(cmd, name string, args []string, stderr io.Writer) (asJSON bool, since string, ok bool) {
	takes := commandOptions[cmd]
	refuse := func(a, why string) (bool, string, bool) {
		fmt.Fprintf(stderr, "proviso %s: %s %q; %s\n", name, why, a, helpHint)
		return false, "", false
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case !slices.Contains(takes, a):
			return refuse(a, "does not take the argument")
		case a == "--json":
			asJSON = true
		case since != "":
			return refuse(a, "takes one revision only, so it refuses a second")
		case i+1 == len(args) || args[i+1] == "" || args[i+1][0] == '-':
			// No revision git accepts starts with "-".
			return refuse(a, "needs a revision after")
		default:
			i++
			since = args[i]
		}
	}
	return asJSON, since, true
}

// helpRun writes the help page to stdout.
func helpRun(stdout io.Writer) int {
	io.WriteString(stdout, helpPage)
	return exitClean
}
