# <a id="install.update"></a> Updating a repository

- Contact: Bruno Baketarić

## Goal

`update` brings what Proviso owns in a repository to the version of the
binary, and says file by file what it wrote, replaced and left alone.

## Rationale

After an upgrade, the skills and agent definitions in a repository still
hold the instructions of the version that installed them, and `init`
never replaces a file. What Proviso owns is today the skills and agent
definitions; once the model changes, the spec tree will follow as a
second part of the same command, with options to run one part alone.
Each part decides on its own and reports its own lines, so that the next
one adds to the command without changing the first.

Proviso goes on owning the skills and agent definitions, so `update`
replaces them without reading what they hold beyond their stamp: an
engineer who changed one by hand copies it first. Everything else
`init` writes is the engineers' own once written, and `update` never
touches it.

Only a release updates: a local build carries a kit nobody released,
and stamping it into a repository would spread it. A file stamped newer
than the binary is left alone, since replacing it would undo a later
release's instructions. A file is left alone otherwise only where its
stamp is the binary's version, so that once `update` has run, `version`
names no file as installed by another version but those `update` warned
about.

## <a id="install.update.release"></a> Rule: Only a release updates

```guarded
When      someone runs `update`, with arguments it takes, as a binary of the [Proviso version] <version>
Provided  <version> is a [release version]
Then      `update` turns to the directory it runs in
  And     every file it writes is stamped with <version>
Otherwise `update` writes to "stderr" one line saying that only a release updates, naming <version>
  And     nothing is written, and nothing goes to "stdout"
  And     `update` exits with `update.exit-status` 2
```

| version | provided | outcome | why |
| --- | --- | --- | --- |
| `0.1.0` | yes | goes on | |
| `0.2.0` | yes | goes on | |
| `0.1.1-dev` | no | refused, exit status 2 | a local build's kit is not released |
| `dev` | no | refused, exit status 2 | |
| `0.2.0-rc.1` | no | refused, exit status 2 | a pre-release is no release |
| `0.2.0+build.5` | no | refused, exit status 2 | a release version has no build part |

## <a id="install.update.repository"></a> Rule: Update works only in the root of a git repository

```guarded
When      `update` turns to the directory <directory> it runs in
Provided  <directory> holds an entry named ".git"
Then      `update` looks for the [agent definitions] and [skills] that `installed_files` names, at their paths below <directory>
  And     `update` looks for no repository above or below <directory>
Otherwise `update` writes to "stderr" that <directory> is not the root of a git repository
  And     nothing is written, and nothing goes to "stdout"
  And     `update` exits with `update.exit-status` 2
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, a clone with a `.git` folder | yes | looks for them | |
| `wt/`, a worktree with a `.git` file | yes | looks for them | a worktree is a repository's root too |
| `repo/src/`, inside a clone | no | refused, exit status 2 | `update` does not look upward, as `init` does not |
| `notes/`, in no repository | no | refused, exit status 2 | |

## <a id="install.update.installed"></a> Rule: Update needs something installed

### Rationale

A repository without any skill or agent was never set up: writing them
alone would leave skills without the spec tree they work on. `init`
sets a repository up.

```guarded
When      `update` looks for the [agent definitions] and [skills] that `installed_files` names, below the root of a git repository <directory>
Provided  an entry exists below <directory> at the path of at least one of them
Then      `update` takes each of them on its own, in the byte order of their paths
  And     `update` touches no other [installed file]
Otherwise `update` writes to "stderr" one line saying that no [skill] or [agent definition] is installed, and that `init` installs them
  And     nothing is written, and nothing goes to "stdout"
  And     `update` exits with `update.exit-status` 2
```

| directory | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/`, after `init` | yes | each taken | |
| `repo/`, holding only `.claude/skills/proviso-build/SKILL.md` | yes | each taken | |
| `repo/`, before `init` | no | refused, exit status 2 | |
| `repo/`, with `proviso/system.md` and `.proviso/build.md` only | no | refused, exit status 2 | only the skills and agents count |

## <a id="install.update.missing"></a> Rule: A missing skill or agent is written

```guarded
When      `update` takes the [agent definition] or [skill] <file>
Provided  no entry exists at the path of <file>
Then      `update` writes <file> as `init` of its own [Proviso version] writes it, with the folders above it that are missing
  And     `update` writes to "stdout" the line naming <file> as written
Otherwise → install.update.stamp
```

| file | provided | outcome | why |
| --- | --- | --- | --- |
| `.claude/agents/proviso-coder.md`, deleted after `init` | yes | written | |
| `.claude/agents/proviso-coder.md`, in place | no | handed over | |
| `.claude/agents/proviso-coder.md`, a folder of that name | no | handed over | any entry counts |

## <a id="install.update.stamp"></a> Rule: A skill or agent is replaced unless its stamp is the binary's version or newer

```guarded
When      `update`, of the [Proviso version] <version>, takes the [agent definition] or [skill] <file>, an entry existing at its path, its [version stamp] naming <stamp>
Provided  <file> can be read as a file whose frontmatter holds exactly one [version stamp]
  And     <stamp> is <version>, compared as text, or a version of semantic versioning 2.0.0 newer than <version> by its precedence
Then      `update` leaves <file> as it is
  And     `update` writes to "stdout" the line naming <file> as skipped
  And     `update` writes to "stderr" one line naming <file> and <stamp> as newer than <version>, where <stamp> is not <version>
Otherwise `update` writes <file> anew, as `init` of its own [Proviso version] writes it
  And     `update` writes to "stdout" the line naming <file> as updated, where it could write it
```

| file | stamp | version | provided | outcome | why |
| --- | --- | --- | --- | --- | --- |
| `.claude/skills/proviso-build/SKILL.md` | `0.2.0` | `0.2.0` | yes | skipped | |
| `.claude/skills/proviso-build/SKILL.md` | `0.3.0` | `0.2.0` | yes | skipped, named as newer on stderr | no downgrade |
| `.claude/skills/proviso-build/SKILL.md` | `0.2.1-dev` | `0.2.0` | yes | skipped, named as newer on stderr | a pre-release of a later version is newer |
| `.claude/skills/proviso-build/SKILL.md` | `0.1.0` | `0.2.0` | no (the version or newer) | updated | |
| `.claude/skills/proviso-build/SKILL.md` | `0.2.0-dev` | `0.2.0` | no (the version or newer) | updated | a pre-release is older than its release |
| `.claude/skills/proviso-build/SKILL.md` | `0.2.0+build.5` | `0.2.0` | no (the version or newer) | updated | another build, as `version` reads it |
| `.claude/skills/proviso-build/SKILL.md` | `dev` | `0.2.0` | no (the version or newer) | updated | `dev` names no released kit |
| `.claude/skills/proviso-build/SKILL.md` | `v0.2.0` | `0.2.0` | no (the version or newer) | updated | not a version |
| `.claude/skills/proviso-build/SKILL.md`, stamped twice | `0.2.0` | `0.2.0` | no (exactly one stamp) | updated | two stamps are damage |
| `.claude/skills/proviso-build/SKILL.md`, with no frontmatter | none | `0.2.0` | no (exactly one stamp) | updated | installed before stamps existed, or edited by hand |
| `.claude/skills/proviso-build/SKILL.md`, a folder of that name | none | `0.2.0` | no (exactly one stamp) | not written, exit status 2 | `update` removes no entry |

## <a id="install.update.exit"></a> Rule: The exit status tells whether every skill and agent is in place

```guarded
When      `update` has taken every [agent definition] and [skill] that `installed_files` names, failing to write <failed>
Provided  <failed> is empty
Then      `update` exits with `update.exit-status` 0
  And     everything else stays as it was
Otherwise `update` writes to "stderr" why each of <failed> could not be written
  And     no line on "stdout" names any of <failed>
  And     the files it wrote stay
  And     `update` exits with `update.exit-status` 2
```

| failed | provided | outcome | why |
| --- | --- | --- | --- |
| none, every file updated | yes | exit status 0 | |
| none, every file skipped, some as newer | yes | exit status 0 | a newer file is no failure |
| `.claude/skills/proviso-build/SKILL.md`, a folder of that name | no | exit status 2, the other files written | |
