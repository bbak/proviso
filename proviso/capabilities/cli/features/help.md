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
Provided  <command> is `init`, `version`, `lint`, `manifest`, `diff`, `impact`, `pack`, `judge`, `verify` or `help`, or "--help" or "-h"
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
would have to choose one. The revision after `--since` belongs to that
option and is not read as an argument of its own.

```guarded
When      <command> reads the argument <argument>, after the arguments <before> and before the arguments <rest>
Provided  <command> takes <argument>, as `lint.json`, `diff.json`, `diff.since`, `impact.json`, `pack.json`, `pack.since` and `verify.json` define
  And     <rest> start with a revision that is not empty, where <argument> is "--since"
  And     <before> hold no "--since", where <argument> is "--since"
Then      <command> reads <argument>, with the revision after it where <argument> is "--since"
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
| version | none | 0.1.0 | none | no (takes) | refused, exit status 2 | `version` takes no argument |
| init | none | --json | none | no (takes) | refused, exit status 2 | |
| judge | none | --json | none | no (takes) | refused, exit status 2 | `judge` reads its input on stdin |
| help | none | diff | none | no (takes) | refused, exit status 2 | |
| lint | none | --since | `HEAD~1` | no (takes) | refused, exit status 2 | only `diff` and `pack` compare with a revision |
| diff | none | --help | none | no (takes) | refused, exit status 2 | the page is `help`'s alone |
| diff | none | --since | none | no (a revision after it) | refused, exit status 2 | |
| pack | none | --since | `""` | no (a revision after it) | refused, exit status 2 | an empty revision names nothing |
| diff | --since `HEAD~1` | --since | `HEAD~2` | no (no "--since" before it) | refused, exit status 2 | two revisions contradict |
