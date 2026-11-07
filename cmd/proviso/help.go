package main

import (
	"fmt"
	"io"
	"slices"
)

// helpPage is system.cli.help.page: what Proviso is, every command with its
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
  show <query> [--json] [--in <capability>]
                      Write every element the identifier or name
                      <query> stands for, with where it stands; quote a
                      name of several words.
  help, --help, -h    Write this page.

Commands the skills run:
  pack [--json] [--since <revision> | --element <identifier>]
                      Write the context pack the agents build from; with
                      --element, the pack of that one element.
  judge               Record the check result on standard input in the
                      coherence file.
  manifest            Record the spec tree as the baseline.
  rehome [--write [--mentions]] [--json]
                      Plan the renames that give moved elements the
                      identifiers of their place, and list the other
                      mentions of the identifiers it renames; with
                      --write, write the renames into the spec tree and
                      the manifest.

Options:
  --json              Write JSON instead of text.
  --since <revision>  Compare with the spec tree at that git revision
                      instead of the baseline.
  --element <identifier>
                      Pack the element with that identifier instead of
                      the amendment (pack only).
  --write             Write the planned renames (rehome only).
  --mentions          Read a mention list on standard input and rename
                      the mentions it names too (rehome, with --write
                      only).
  --in <capability>   Look only among the elements of that capability
                      and of the system level (show only).

Exit statuses:
  0                   The command did what it was asked.
  1                   lint, diff or verify found what it reports,
                      judge refused its input, a rename of rehome
                      would take an identifier, or the query of show
                      names no element.
  2                   The command could not run, or was misused.

Documentation: https://github.com/bbak/proviso/blob/main/docs/commands.md
`

// helpHint ends every line that refuses a command or an argument.
const helpHint = `"proviso help" shows the help page`

// commandOptions is, for every command of cli.help.command, the options
// it takes; "--since" takes a revision after it, "--element" an
// identifier, "--in" a capability.
var commandOptions = map[string][]string{
	"init":     nil,
	"update":   nil,
	"version":  nil,
	"lint":     {"--json"},
	"manifest": nil,
	"diff":     {"--json", "--since"},
	"impact":   {"--json"},
	"pack":     {"--json", "--since", "--element"},
	"judge":    nil,
	"verify":   {"--json"},
	"rehome":   {"--write", "--json", "--mentions"},
	"show":     {"--json", "--in"},
	"help":     nil,
}

// takesQuery names the commands that take one argument that is no option,
// their query: show.query.
var takesQuery = map[string]bool{"show": true}

// readArgs reads the arguments of the command name, which is cmd as it
// was given, following cli.help.argument. It writes to stderr one line
// naming the command and the first argument it does not take.
func readArgs(cmd, name string, args []string, stderr io.Writer) (options, bool) {
	takes := commandOptions[cmd]
	refuse := func(a, why string) (options, bool) {
		fmt.Fprintf(stderr, "proviso %s: %s %q; %s\n", name, why, a, helpHint)
		return options{}, false
	}
	var o options
	for i := 0; i < len(args); i++ {
		a := args[i]
		// A query is not empty and does not start with "-", as no option.
		query := takesQuery[cmd] && a != "" && a[0] != '-'
		switch {
		case query && o.query != "":
			// One lookup per run; a name of several words is quoted.
			return refuse(a, "takes one query, so it refuses")
		case query:
			o.query = a
		case !slices.Contains(takes, a):
			return refuse(a, "does not take the argument")
		case a == "--json":
			o.json = true
		case a == "--write":
			o.write = true
		case a == "--mentions" && !slices.Contains(args, "--write"):
			// Without --write nothing is written, so nothing is picked.
			return refuse(a, "takes --mentions only with --write, so it refuses")
		case a == "--mentions":
			o.mentions = true
		case a == "--in" && o.in != "":
			// One capability per run.
			return refuse(a, "takes one --in, so it refuses")
		case a != "--in" && (o.since != "" || o.element != ""):
			// One revision or one element per run, never both.
			return refuse(a, "takes one --since or one --element, so it refuses")
		case i+1 == len(args) || args[i+1] == "" || args[i+1][0] == '-':
			// No revision git accepts, no identifier and no capability
			// starts with "-".
			return refuse(a, "needs a value after")
		case a == "--since":
			i++
			o.since = args[i]
		case a == "--in":
			i++
			o.in = args[i]
		default:
			i++
			o.element = args[i]
		}
	}
	return o, true
}

// options are the arguments a command read: whether it was given --json,
// --write and --mentions, the values of --since, --element and --in, and
// the query, "" for none.
type options struct {
	json, write, mentions     bool
	since, element, in, query string
}

// helpRun writes the help page to stdout.
func helpRun(stdout io.Writer) int {
	io.WriteString(stdout, helpPage)
	return exitClean
}
