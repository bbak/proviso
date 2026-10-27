# <a id="install.version"></a> Telling the version

- Contact: Bruno Baketarić

## Goal

`version` names the Proviso version of the binary and the model version
it reads, and, in the root of a git repository, says which installed
skill or agent and which part of the spec tree does not fit them.

## Rationale

Proviso has two versions: its own, which follows semantic versioning,
and the model version, a whole number that changes with every change to
the notation. An engineer who upgrades the binary needs both, and needs
to know what was installed by another version.

Standard output carries the two lines and nothing else, wherever
`version` runs, so that a script or a skill reads them without
filtering. Everything else is a warning or a hint, goes to standard
error, and never changes the exit status: nothing has failed, and
silence in the root of a repository means everything fits. The wording
of those lines is for people; what each one names is specified.

`version` finds the repository's root as `init` does, by the `.git`
entry of the directory it runs in, and looks nowhere else, since it
checks the paths `init` writes there. It only reports: refreshing what
it reports is the job of an update, deferred for now.

## <a id="install.version.output"></a> Rule: Version names the Proviso version and the model version

```guarded
When      someone runs `version` with the arguments <arguments>
Provided  <arguments> is empty
Then      `version` writes to "stdout" the line "proviso", a space and its [Proviso version], and then the line "model", a space and `model_version`
  And     nothing else goes to "stdout"
  And     `version` exits with `version.exit-status` 0, whatever it writes to "stderr"
Otherwise `version` writes to "stderr" that it takes no argument, and its usage
  And     nothing goes to "stdout"
  And     `version` exits with `version.exit-status` 2
```

| arguments | provided | outcome | why |
| --- | --- | --- | --- |
| none, built as release `0.1.0` | yes | "proviso 0.1.0" and "model 0", exit status 0 | |
| none, built locally after release `0.1.0` | yes | "proviso 0.1.1-dev" and "model 0", exit status 0 | |
| none, built with no version given | yes | "proviso dev" and "model 0", exit status 0 | |
| none, in a repository root whose skills are stale | yes | "proviso 0.1.0" and "model 0", exit status 0 | warnings change neither the output nor the exit status |
| `--json` | no | exit status 2, nothing on "stdout" | |
| `0.1.0` | no | exit status 2, nothing on "stdout" | |

## <a id="install.version.repository"></a> Rule: Version checks the installation only in the root of a git repository

```guarded
Given     `version` has written its two lines
When      `version`, run in the directory <directory>, turns to what is installed there
Provided  <directory> holds an entry named ".git"
Then      `version` checks the [agent definitions] and [skills] that `installed_files` names, at their paths below <directory>, and `manifest_file` at "proviso/manifest.sum" below <directory>
  And     `version` looks for no repository above or below <directory>
Otherwise `version` writes to "stderr" one line saying that run in the root of a git repository it checks the installed [skills] and [agent definitions] and the [spec tree] against its version
  And     `version` checks nothing
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, a clone with a `.git` folder | yes | checked | |
| `wt/`, a worktree with a `.git` file | yes | checked | a worktree is a repository's root too |
| `repo/src/`, inside a clone | no | the hint, nothing checked | `version` does not look upward, as `init` does not |
| `notes/`, in no repository | no | the hint, nothing checked | |

## <a id="install.version.kit"></a> Rule: Version says when no skill or agent is installed

```guarded
When      `version`, checking the root of a git repository <directory>, looks for the [agent definitions] and [skills] that `installed_files` names
Provided  an entry exists below <directory> at the path of at least one of them
Then      `version` checks each of them on its own
  And     a line names each of them only where it does not fit
Otherwise `version` writes to "stderr" one line saying that no [skill] or [agent definition] is installed
  And     no line names any of them on its own
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, after `init` | yes | each checked | |
| `repo/`, holding only `.claude/skills/proviso-build/SKILL.md` | yes | each checked | |
| `repo/`, before `init` | no | one line: no skill or agent installed | |
| `repo/`, with `.claude/` but none of the paths `installed_files` names | no | one line: no skill or agent installed | only the named paths count |

## <a id="install.version.missing"></a> Rule: A skill or agent that is missing is named

```guarded
Given     `version` checks the root of a git repository, where an entry exists at the path of at least one [agent definition] or [skill] that `installed_files` names
When      `version` checks the [agent definition] or [skill] <file>
Provided  no entry exists at the path of <file>
Then      `version` writes to "stderr" one line naming <file> as not installed
  And     `version` checks nothing else of <file>
Otherwise → install.version.stamps
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `.claude/agents/proviso-coder.md`, deleted after `init` | yes | named as not installed | |
| `.claude/agents/proviso-coder.md`, in place | no | handed over | |
| `.claude/agents/proviso-coder.md`, a folder of that name | no | handed over | any entry counts, as for `init` |

## <a id="install.version.stamps"></a> Rule: A skill or agent with more than one stamp is named

### Rationale

Proviso writes one stamp per file, so a second one is damage, from a
hand edit or a merge, and the line says so rather than pick one. Every
line `proviso-version:` below `metadata:` counts, whatever its version:
the stamps are counted first, and only then is a version judged.

```guarded
When      `version` checks the [agent definition] or [skill] <file>, an entry existing at its path
Provided  <file> can be read as a file, and its frontmatter holds more than one [version stamp], whatever their versions
Then      `version` writes to "stderr" one line naming <file> as holding more than one [version stamp]
  And     `version` checks nothing else of <file>
Otherwise → install.version.unstamped
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `.claude/skills/proviso-build/SKILL.md`, two lines `proviso-version: 0.1.0` and `proviso-version: 0.2.0` below `metadata:` | yes | named as holding more than one stamp | |
| `.claude/skills/proviso-build/SKILL.md`, two `metadata:` blocks, each stamped `0.1.0` | yes | named as holding more than one stamp | two stamps, even if equal |
| `.claude/skills/proviso-build/SKILL.md`, the lines `proviso-version: 0.1.0` and `proviso-version:` below `metadata:` | yes | named as holding more than one stamp | an empty stamp counts too |
| `.claude/skills/proviso-build/SKILL.md`, stamped `0.1.0` once | no | handed over | |
| `.claude/skills/proviso-build/SKILL.md`, without read permission | no | handed over | an unreadable file names no version |

## <a id="install.version.unstamped"></a> Rule: A skill or agent without a stamp is named

### Rationale

A file that cannot be read and a file without a stamp tell the same:
nobody knows which version installed it. Nor does a stamp that is no
version at all: only `dev` and a version of semantic versioning count, a
wider set than the Proviso versions built today, so that a stamp from a
build given any other semantic version still reads as the version it
is.

```guarded
When      `version` checks the [agent definition] or [skill] <file>, an entry existing at its path
Provided  <file> cannot be read as a file, or it has no frontmatter, or its frontmatter holds no [version stamp] whose version is "dev" or a version of semantic versioning 2.0.0 without a leading "v"
Then      `version` writes to "stderr" one line naming <file> as naming no [Proviso version]
  And     `version` checks nothing else of <file>
Otherwise → install.version.stale
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `.claude/skills/proviso-build/SKILL.md`, installed before stamps existed | yes | named as naming no Proviso version | |
| `.claude/skills/proviso-build/SKILL.md`, a folder of that name | yes | named as naming no Proviso version | |
| `.claude/skills/proviso-build/SKILL.md`, its frontmatter holding `metadata: {proviso-version: 0.1.0}` on one line | yes | named as naming no Proviso version | Proviso reads only the stamp's own lines |
| `.claude/skills/proviso-build/SKILL.md`, its frontmatter holding `proviso-version: 0.1.0` at the top, not below `metadata:` | yes | named as naming no Proviso version | |
| `.claude/skills/proviso-build/SKILL.md`, its frontmatter holding `proviso-version:` below `metadata:`, with no version | yes | named as naming no Proviso version | an empty version names none |
| `.claude/skills/proviso-build/SKILL.md`, its stamp below an indented `metadata:`, nested under another key | yes | named as naming no Proviso version | `metadata:` starts its line |
| `.claude/skills/proviso-build/SKILL.md`, a blank line between `metadata:` and its stamp | yes | named as naming no Proviso version | a blank line ends the lines below `metadata:` |
| `.claude/skills/proviso-build/SKILL.md`, a first line `---` and no later line `---` | yes | named as naming no Proviso version | without its closing line there is no frontmatter |
| `.claude/skills/proviso-build/SKILL.md`, stamped `[1]` | yes | named as naming no Proviso version | not a semantic version |
| `.claude/skills/proviso-build/SKILL.md`, stamped `v0.1.0` | yes | named as naming no Proviso version | the `v` belongs to the tag only |
| `.claude/skills/proviso-build/SKILL.md`, stamped `0.1` | yes | named as naming no Proviso version | a semantic version has three numbers |
| `.claude/skills/proviso-build/SKILL.md`, with no frontmatter | yes | named as naming no Proviso version | |
| `.claude/skills/proviso-build/SKILL.md`, its frontmatter holding `metadata` without `proviso-version` | yes | named as naming no Proviso version | |
| `.claude/skills/proviso-build/SKILL.md`, stamped `0.1.0` | no | handed over | |
| `.claude/skills/proviso-build/SKILL.md`, stamped `0.1.0`, with carriage returns and line feeds | no | handed over, with the stamp `0.1.0` | the line ending is not part of the line |
| `.claude/skills/proviso-build/SKILL.md`, stamped `0.1.1-dev` | no | handed over | a pre-release is a semantic version |
| `.claude/skills/proviso-build/SKILL.md`, stamped `dev` | no | handed over | |
| `.claude/skills/proviso-build/SKILL.md`, an indented comment line between `metadata:` and its stamp `0.1.0` | no | handed over, with the stamp `0.1.0` | an indented line belongs to the lines below `metadata:` |
| `.claude/skills/proviso-build/SKILL.md`, stamped `0.1.0` with a tab as indentation, no space after the colon and spaces at the end | no | handed over, with the stamp `0.1.0` | whitespace changes nothing |

## <a id="install.version.stale"></a> Rule: A skill or agent installed by another version is named

```guarded
When      `version` compares the [version stamp] <stamp> of the [agent definition] or [skill] <file> with its own [Proviso version] <version>
Provided  <stamp> differs from <version>, compared as text
Then      `version` writes to "stderr" one line naming <file>, <stamp> as the version that installed it, and the model version `version` finds when it looks up the model version of <stamp>, where it finds one
  And     the line names no model version where it finds none
Otherwise `version` writes no line about <file>
  And     everything else stays as it was
```

| file | stamp | version | provided | outcome | why |
| --- | --- | --- | --- | --- | --- |
| `.claude/agents/proviso-tester.md` | `0.1.0` | `0.2.0` | yes | named, installed by 0.1.0, which reads model 0 | |
| `.claude/agents/proviso-tester.md` | `dev` | `0.2.0` | yes | named, installed by dev, no model | |
| `.claude/agents/proviso-tester.md` | `0.1.0` | `0.1.1-dev` | yes | named, installed by 0.1.0, which reads model 0 | a local build is another version |
| `.claude/agents/proviso-tester.md` | `0.1.0+build.5` | `0.1.0` | yes | named, installed by 0.1.0+build.5, no model | another build, though semantic versioning ranks them alike |
| `.claude/agents/proviso-tester.md` | `0.3.0` | `0.2.0` | yes | named, installed by 0.3.0, no model | a later release may read a model this binary does not know |
| `.claude/agents/proviso-tester.md` | `0.2.0` | `0.2.0` | no | no line | |
| `.claude/agents/proviso-tester.md` | `dev` | `dev` | no | no line | as in Proviso's own repository |

## <a id="install.version.model"></a> Rule: A stamp's model version is given only where the binary can know it

### Rationale

The last row of `model_releases` holds for every later release only as
far as the binary knows: a release after its own may read a model
version added after it. A pre-release is no release, and `dev` names
none, so neither is looked up.

```guarded
Given     `model_releases` holds the rows <rows>
When      `version`, of the [Proviso version] <version>, looks up the model version of the [version stamp] <stamp>
Provided  <stamp> is a [release version]
  And     <stamp> is no newer than <version> by the precedence of semantic versioning, a <version> of "dev" being newer than every [release version]
  And     <stamp> is no older than the first release of the first of <rows>
Then      the model version is that of the last of <rows> whose first release is no newer than <stamp>
  And     `version` names it in the line about the file
Otherwise `version` gives no model version for <stamp>
  And     the line about the file still names <stamp>
```

| rows | version | stamp | provided | outcome | why |
| --- | --- | --- | --- | --- | --- |
| 0 from `0.1.0` | `0.2.0` | `0.1.0` | yes | model 0 | |
| 0 from `0.1.0`, 1 from `0.3.0` | `0.4.0` | `0.2.5` | yes | model 0 | the row holds until the next one |
| 0 from `0.1.0`, 1 from `0.3.0` | `0.4.0` | `0.3.0` | yes | model 1 | a row's first release is its own |
| 0 from `0.1.0`, 1 from `0.3.0` | `0.3.1-dev` | `0.3.0` | yes | model 1 | a local build after 0.3.0 is newer than 0.3.0 |
| 0 from `0.1.0`, 1 from `0.3.0` | `dev` | `0.4.2` | yes | model 1 | the last row holds for every later release |
| 0 from `0.1.0` | `0.2.0` | `dev` | no (release version) | no model | |
| 0 from `0.1.0` | `0.2.0` | `0.1.1-dev` | no (release version) | no model | a pre-release is no release |
| 0 from `0.1.0` | `0.2.0` | `0.1.0+build.5` | no (release version) | no model | a release version has no build part |
| 0 from `0.1.0` | `0.2.0` | `0.3.0` | no (no newer than the version) | no model | a later release may read a model this binary does not know |
| 0 from `0.1.0` | `0.1.1-dev` | `0.1.1` | no (no newer than the version) | no model | 0.1.1 is newer than its own pre-release |
| 0 from `0.1.0` | `0.2.0` | `0.0.9` | no (no older than the first release) | no model | before the first release |

## <a id="install.version.manifest-missing"></a> Rule: A spec tree without a manifest is named

### Rationale

A tree is accepted under a model version only once its manifest is
written; before `init`, or before the first acceptance, there is none,
and `version` says so rather than stay silent.

```guarded
When      `version` checks "proviso/manifest.sum" below the root of a git repository <directory>
Provided  no entry exists at that path below <directory>
Then      `version` writes to "stderr" one line saying that "proviso/manifest.sum" does not exist, so no model version is recorded for the [spec tree]
  And     `version` checks nothing else of the [spec tree]
Otherwise → install.version.manifest-unreadable
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, before `init` | yes | named as not existing | |
| `repo/`, after `init`, before the first acceptance | yes | named as not existing | |
| `repo/`, with an accepted tree | no | handed over | |

## <a id="install.version.manifest-unreadable"></a> Rule: A manifest that cannot be read is named

```guarded
When      `version` checks the entry <entry> at the path "proviso/manifest.sum"
Provided  <entry> cannot be read as a file
Then      `version` writes to "stderr" one line saying that "proviso/manifest.sum" cannot be read
  And     `version` checks nothing else of the [spec tree]
Otherwise → install.version.manifest-model
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| `manifest.sum`, without read permission | yes | named as unreadable | |
| `manifest.sum`, a folder | yes | named as unreadable | |
| `manifest.sum`, a readable file | no | handed over | |

## <a id="install.version.manifest-model"></a> Rule: A spec tree of another model version is named

### Rationale

`version` judges the manifest's first line by the same rule as `lint`,
so that the two agree on whether the model version a tree needs is the
one the binary reads.

```guarded
When      `version` reads the first line <line> of "proviso/manifest.sum"
Provided  <line>, without its line ending, names `model_version`
Then      `version` writes no line about the [spec tree]
  And     everything else stays as it was
Otherwise → install.version.manifest-other
```

| line | provided | outcome | why |
| --- | --- | --- | --- |
| `model 0` | yes | no line | |
| `model 0`, ending in a carriage return and a line feed | yes | no line | the line ending is not part of the line |
| `model 1` | no | handed over | |
| empty | no | handed over | |
| `model 00` | no | handed over | `model_version` is written without a leading zero |

## <a id="install.version.manifest-other"></a> Rule: The model version the spec tree needs is named

```guarded
When      `version` finds that the first line <line> of "proviso/manifest.sum" does not name `model_version`
Provided  <line>, without its line ending, is "model", a space and a whole number in decimal digits without a leading zero
Then      `version` writes to "stderr" one line naming that number as the model version the [spec tree] needs, and `model_version` as the one it reads
  And     everything else stays as it was
Otherwise `version` writes to "stderr" one line saying that "proviso/manifest.sum" names no model version
  And     everything else stays as it was
```

| line | provided | outcome | why |
| --- | --- | --- | --- |
| `model 1` | yes | the tree needs model 1, this proviso reads model 0 | |
| `model 12` | yes | the tree needs model 12, this proviso reads model 0 | |
| empty | no | names no model version | |
| `model` | no | names no model version | |
| `model one` | no | names no model version | |
| `model 00` | no | names no model version | a leading zero; never "needs 0, reads 0" |
| `<<<<<<< HEAD` | no | names no model version | a merge conflict left in the file |
