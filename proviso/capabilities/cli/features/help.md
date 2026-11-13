# <a id="cli.help"></a> Help and misuse

- Contact: Bruno Baketarić

## Goal

`help` shows the help page on request, and every command refuses an
argument it does not take, pointing to `help`.

## Rationale

The help page goes to standard output with exit status 0 when someone
asks for it, so that it can be read or paged like any output. Without a
command, Proviso has nothing to do, and a script that calls it so has
failed: the page goes to standard error and the exit status is 2. The
arguments are read before anything else, so that a misused command
checks and writes nothing.

## <a id="cli.help.start"></a> Rule: Proviso needs a command

```guarded
When      someone runs "proviso" with the arguments <arguments>
Provided  <arguments> hold at least one argument
Then      "proviso" reads the first of <arguments> as its command
  And     the rest of <arguments> as that command's arguments
Otherwise "proviso" writes `help.page` to "stderr"
  And     "proviso" exits with `help.exit-status` 2
  And     nothing goes to "stdout"
```

| arguments | provided | outcome | why |
| --- | --- | --- | --- |
| `lint --json` | yes | `lint` is the command | |
| `help` | yes | `help` is the command | |
| `""` | yes | `""` is the command | an empty argument is an argument |
| none | no | the page on stderr, exit status 2 | |

## <a id="cli.help.command"></a> Rule: Proviso knows its commands

```guarded
When      "proviso" reads its command <command>
Provided  <command> is `init`, `update`, `version`, `lint`, `manifest`, `diff`, `impact`, `pack`, `judge`, `verify`, `rehome`, `show` or `help`, or "--help" or "-h"
Then      "proviso" runs <command>, and runs `help` for "--help" and "-h"
  And     the command reads its arguments as its own [rules] say
Otherwise "proviso" writes to "stderr" one line naming <command> as no command of Proviso, and that "proviso help" shows the help page
  And     "proviso" exits with `help.exit-status` 2
  And     nothing goes to "stdout"
```

| command | provided | outcome | why |
| --- | --- | --- | --- |
| lint | yes | `lint` runs | |
| verify | yes | `verify` runs | |
| update | yes | `update` runs | |
| rehome | yes | `rehome` runs | |
| show | yes | `show` runs | |
| help | yes | `help` runs | |
| --help | yes | `help` runs | |
| -h | yes | `help` runs | |
| Lint | no | refused, exit status 2 | commands are matched exactly |
| --version | no | refused, exit status 2 | the version is a command, `version` |
| build | no | refused, exit status 2 | |
| `""` | no | refused, exit status 2 | |

## <a id="cli.help.page"></a> Rule: Help shows the help page

```guarded
When      someone runs `help` with the arguments <arguments>
Provided  <arguments> is empty
Then      `help` writes `help.page` to "stdout"
  And     nothing goes to "stderr"
  And     `help` exits with `help.exit-status` 0
Otherwise → cli.help.argument
```

| arguments | provided | outcome | why |
| --- | --- | --- | --- |
| none | yes | the page on stdout, exit status 0 | |
| `diff` | no | handed over | there is one page, not one per command |
| `--json` | no | handed over | |

## <a id="cli.help.argument"></a> Rule: A command reads only the arguments it takes

### Rationale

`--json` given twice asks for the same thing twice, so it is read
twice; a second `--since` would ask for two revisions, and the command
would have to choose one. A second `--element` would ask for a pack of
two elements, which `pack` does not assemble, and a pack of a named
element comes from no revision, so `--since` and `--element` exclude
each other. The value after either belongs to that option and is not
read as an argument of its own. No revision git accepts and no
identifier starts with "-", so an argument that does is a forgotten
value, not one to look up. A second `--in` would name two capabilities. `--rename` is given once
for each element whose new identifier the engineer gives; what its
value holds is `rehome`'s to check.
`show` takes one argument that is no option, its query, since it has
nothing to do without one. A second would ask for two lookups, which
`show` does not make; a name of several words is quoted for the shell
rather than joined, so that a mistyped option is refused rather than
read as part of a name.

```guarded
When      <command> reads the argument <argument>, after the arguments <before> and before the arguments <rest>
Provided  <command> takes <argument>, as `lint.json`, `diff.json`, `diff.since`, `impact.json`, `pack.json`, `pack.since`, `pack.element`, `verify.json`, `rehome.write`, `rehome.json`, `rehome.mentions`, `rehome.rename`, `show.query`, `show.in` and `show.json` define
  And     <rest> start with a value that is not empty and does not start with "-", where <argument> is "--since", "--element", "--in" or "--rename"
  And     <before> hold neither "--since" nor "--element", where <argument> is "--since" or "--element", no "--in", where <argument> is "--in", and no `show.query`, where <argument> is one
  And     <before> or <rest> hold "--write", where <argument> is "--mentions"
Then      <command> reads <argument>, with the value after it where <argument> is "--since", "--element", "--in" or "--rename"
  And     <command> goes on with the arguments after them
Otherwise <command> writes to "stderr" one line naming <command> and <argument>, and that "proviso help" shows the help page
  And     <command> exits with 2, as its exit status defines
  And     nothing goes to "stdout", and <command> checks and writes nothing
```

| command | before | argument | rest | provided | outcome | why |
| --- | --- | --- | --- | --- | --- | --- |
| lint | none | --json | none | yes | read | |
| lint | --json | --json | none | yes | read | the same option twice asks for the same |
| diff | none | --since | `HEAD~1` | yes | read, with `HEAD~1` | |
| pack | --json | --since | `a1b2c3d` | yes | read, with `a1b2c3d` | |
| pack | --json | --element | `lint.run.exit` | yes | read, with `lint.run.exit` | |
| rehome | --json | --write | none | yes | read | |
| rehome | --write | --mentions | none | yes | read | |
| rehome | none | --mentions | --write | yes | read | the order of options is free |
| rehome | none | --rename | `serving.snapshot=system.whole-snapshot` | yes | read, with `serving.snapshot=system.whole-snapshot` | |
| rehome | --rename `a.x=b.x` | --rename | `a.y=b.y` | yes | read, with `a.y=b.y` | one value per element to rename |
| show | none | `spec tree` | none | yes | read as the query | |
| show | `lint.run.exit` | --in | `lint` | yes | read, with `lint` | |
| show | `lint.run.exit` | --json | none | yes | read | |
| version | none | 0.1.0 | none | no (takes) | refused, exit status 2 | `version` takes no argument |
| init | none | --json | none | no (takes) | refused, exit status 2 | |
| judge | none | --json | none | no (takes) | refused, exit status 2 | `judge` reads its input on stdin |
| help | none | diff | none | no (takes) | refused, exit status 2 | |
| lint | none | --since | `HEAD~1` | no (takes) | refused, exit status 2 | only `diff` and `pack` compare with a revision |
| diff | none | --help | none | no (takes) | refused, exit status 2 | the page is `help`'s alone |
| diff | none | --element | `lint.run.exit` | no (takes) | refused, exit status 2 | only `pack` names an element |
| lint | none | `spec tree` | none | no (takes) | refused, exit status 2 | only `show` takes a query |
| show | none | `""` | none | no (takes) | refused, exit status 2 | an empty query names nothing |
| show | none | -x | none | no (takes) | refused, exit status 2 | an argument starting with "-" is an option |
| diff | none | --since | none | no (a value after it) | refused, exit status 2 | |
| pack | none | --since | `""` | no (a value after it) | refused, exit status 2 | an empty revision names nothing |
| diff | none | --since | --json | no (a value after it) | refused, exit status 2 | an option is no revision |
| pack | none | --element | none | no (a value after it) | refused, exit status 2 | |
| rehome | none | --rename | --write | no (a value after it) | refused, exit status 2 | an option is no value |
| show | `lint.run.exit` | --in | none | no (a value after it) | refused, exit status 2 | |
| diff | --since `HEAD~1` | --since | `HEAD~2` | no (neither before it) | refused, exit status 2 | two revisions contradict |
| pack | --element `lint.run.exit` | --element | `lint.run.root` | no (neither before it) | refused, exit status 2 | one element per pack |
| pack | --since `HEAD~1` | --element | `lint.run.exit` | no (neither before it) | refused, exit status 2 | a named element comes from no revision |
| pack | --element `lint.run.exit` | --since | `HEAD~1` | no (neither before it) | refused, exit status 2 | |
| show | `lint.run.exit` --in `lint` | --in | `query` | no (neither before it) | refused, exit status 2 | two capabilities contradict |
| show | `spec` | `tree` | none | no (neither before it) | refused, exit status 2 | one query per run; a name of several words is quoted |
| rehome | --json | --mentions | none | no (with --write) | refused, exit status 2 | a plan alone writes nothing, so there is nothing to pick for |
