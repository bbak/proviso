# <a id="install.init"></a> Initialising a repository

- Contact: Bruno Baketarić

## Goal

`init` writes every installed file a repository is missing and says,
file by file, what it wrote and what it left alone.

## Rationale

An engineer runs `init` once per repository, by hand, so it works in
the directory it is run in and refuses anywhere else rather than look
for the repository's root. It never reads a file it leaves alone: an
existing file is the engineers' own, whatever it holds. A file it fails
to write is reported, and a second run writes it, since nothing is ever
written twice.

## <a id="install.init.repository"></a> Rule: Init works only in the root of a git repository

```guarded
When      someone runs `init` in the directory <directory>
Provided  <directory> holds an entry named ".git"
Then      `init` takes every [installed file] that `installed_files` names, at its path below <directory>
  And     everything else stays as it was
Otherwise `init` writes to "stderr" that <directory> is not the root of a git repository
  And     `init` exits with `init.exit-status` 2
  And     nothing is written, and nothing goes to "stdout"
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, a clone with a `.git` folder | yes | the installed files taken | |
| `wt/`, a worktree with a `.git` file | yes | the installed files taken | a worktree is a repository's root too |
| `repo/`, its `.git` folder marked hidden, as Git for Windows marks it | yes | the installed files taken | an entry counts whatever its attributes |
| `repo/src/`, inside a clone | no | exit status 2, nothing written | `init` does not look upward |
| `notes/`, in no repository | no | exit status 2, nothing written | |

## <a id="install.init.existing"></a> Rule: Init never replaces an entry

```guarded
When      `init` takes the [installed file] <file>
Provided  no entry exists at the path of <file>
Then      `init` writes <file>, with the folders above it that are missing
  And     `init` writes to "stdout" the line naming <file> as written
Otherwise `init` leaves that entry as it is, without reading it
  And     `init` writes to "stdout" the line naming <file> as skipped
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `proviso/system.md`, in a repository without `proviso/` | yes | written, with `proviso/` | |
| `proviso/system.md`, which already exists | no | skipped, unchanged | |
| `.proviso/tester.md`, an empty file | no | skipped, unchanged | an empty file is the engineers' too |
| `proviso/.gitattributes`, a folder of that name | no | skipped, unchanged | any entry counts |
| `.claude/agents/proviso-tester.md`, beside other agents in `.claude/agents/` | yes | written, the other agents unchanged | only the named path counts |

## <a id="install.init.skeleton"></a> Rule: The system skeleton starts a tree that lint accepts

### Rationale

A tree begins with its system, and an engineer should see at once what
to decide first. Every section of the skeleton is undecided, so lint
asks for each instead of failing, and the element diff lists the system
as not to be built until each is decided. The name is taken in NFC,
as Proviso reports every name, so that the same folder gives the same
heading whichever form its name is stored in.

```guarded
When      `init` writes `system_skeleton` in the directory <directory>, which holds no other [spec tree] file
Then      the heading of the system names it with the last name of the path of <directory>, in NFC
  And     `lint`, run in <directory>, exits with `lint.exit-status` 0, reporting each of the six [sections] of the system as [undecided]
  And     `diff`, run in <directory>, lists the system as added and not to be built
```

| directory | outcome |
| --- | --- |
| `C:/work/shop` | `# <a id="system"></a> shop`; lint exits with 0 and six asks; diff lists `system` as not to be built |
| `/home/me/cash-ops` | `# <a id="system"></a> cash-ops`; lint exits with 0 and six asks; diff lists `system` as not to be built |
| `/home/me/übersicht`, stored as `u` followed by a mark | `# <a id="system"></a> übersicht`, composed; lint exits with 0 and six asks; diff lists `system` as not to be built |

## <a id="install.init.line-endings"></a> Rule: The tree is checked out with line feeds

### Rationale

Git for Windows checks files out with carriage returns by default, and
lint reports a carriage return as an error. The attribute keeps every
file of the tree, those added later included, on line feeds, whatever
the repository's own settings say.

```guarded
When      git reads the attributes of the file <file> below the [spec root], after `init` wrote `gitattributes_file` there
Then      git gives <file> the attribute "eol" with the value "lf"
  And     git checks <file> out with line feeds, even where its settings would give carriage returns
```

| file | outcome | why |
| --- | --- | --- |
| `proviso/system.md` | "eol" is "lf" | |
| `proviso/capabilities/cart/glossary.md`, added later | "eol" is "lf" | the attribute covers what is added later |
| `proviso/manifest.sum` | "eol" is "lf" | the manifest too |

## <a id="install.init.agents"></a> Rule: An agent definition is installed complete

### Rationale

The agent host reads only the installed definition, and the agent reads
only what it is given; a definition that pointed to another file of the
kit would point to nothing. The kit keeps the reading guide once for
every agent that reads a context pack, and `init` puts it in place.

```guarded
When      `init` writes the [agent definition] <definition>
Then      <definition> is the [kit]'s definition of that agent, with the [kit]'s reading guide where the definition marks it
  And     its frontmatter holds the [version stamp] naming the [Proviso version] of `init`, so that a later run of `version` can tell which version installed it
  And     <definition> refers to no file of the [kit] that `installed_files` does not name
  And     <definition> refers to no file below ".proviso/" but the notes to that agent
```

| definition | outcome |
| --- | --- |
| `.claude/agents/proviso-tester.md` | the testing agent's definition, named `proviso-tester`, with the reading guide in it, stamped with `init`'s Proviso version |
| `.claude/agents/proviso-coder.md` | the coding agent's definition, named `proviso-coder`, with the same reading guide in it, stamped with `init`'s Proviso version |
| `.claude/agents/proviso-checker.md` | the checking agent's definition, named `proviso-checker`, with the same reading guide in it, stamped with `init`'s Proviso version, and no engineers' notes |

## <a id="install.init.skills"></a> Rule: A skill is installed complete

### Rationale

The agent host reads only the installed skill; a skill that pointed to
another file of the kit would point to nothing. A skill reads no context
pack, so it carries no reading guide. A skill that edits the spec tree
carries the writing guide instead: the kit keeps it once for every such
skill, so that a decision written into the tree while integrating,
building or checking reads the same, and `init` puts it in place.

```guarded
When      `init` writes the [skill] <skill>
Then      <skill> is the [kit]'s [skill] of that name, with the [kit]'s writing guide where the [skill] marks it
  And     its frontmatter holds the [version stamp] naming the [Proviso version] of `init`, so that a later run of `version` can tell which version installed it
  And     <skill> refers to no file of the [kit] that `installed_files` does not name
```

| skill | outcome | why |
| --- | --- | --- |
| `.claude/skills/proviso-build/SKILL.md` | the skill that builds an amendment, named `proviso-build`, with the writing guide in it, stamped with `init`'s Proviso version | |
| `.claude/skills/proviso-check/SKILL.md` | the skill that checks an amendment, named `proviso-check`, with the same writing guide in it, stamped with `init`'s Proviso version | |
| `.claude/skills/proviso-integrate/SKILL.md` | the skill that brings a request into the spec tree, named `proviso-integrate`, with the same writing guide in it, stamped with `init`'s Proviso version | |
| `.claude/skills/proviso-accept/SKILL.md` | the skill that accepts an amendment, named `proviso-accept`, and no writing guide, stamped with `init`'s Proviso version | it never edits the spec tree |

## <a id="install.init.exit"></a> Rule: The exit status tells whether every installed file is in place

```guarded
When      `init` has taken every [installed file], failing to write the [installed files] <failed>
Provided  <failed> is empty
Then      `init` exits with `init.exit-status` 0
  And     everything else stays as it was
Otherwise `init` writes to "stderr" why each of <failed> could not be written
  And     no line on "stdout" names any of <failed>
  And     the [installed files] it wrote stay
  And     `init` exits with `init.exit-status` 2
```

| failed | provided | outcome | why |
| --- | --- | --- | --- |
| none, every file written | yes | exit status 0 | |
| none, every file skipped | yes | exit status 0 | a second run is no failure |
| every installed file below `.claude/`, `.claude` being a file | no | exit status 2, the other files written | a folder cannot be made below a file |
