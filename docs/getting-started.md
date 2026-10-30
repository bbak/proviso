---
title: Getting started
nav_order: 2
---

# Getting started

This page takes a repository from nothing to its first accepted
amendment. It assumes a git repository holding the system's code, and
Claude Code, which runs the skills.

## Install Proviso

Proviso is one binary, `proviso`, with no dependencies. Download the
archive for your platform from the repository's
[releases](https://github.com/bbak/proviso/releases), unpack it, and put
`proviso` in a folder on your path.

The skills run `proviso` from the path. If it is somewhere else, name
the command in `.proviso/build.md` (see below), such as `bin/proviso`.
Keep it the version that installed the skills.

## Start the first amendment

The spec tree itself is the first amendment: until it is accepted, the
repository has no baseline. So everything below happens on a branch:

```sh
git switch -c first-spec
```

## Initialise the repository

In the repository's root, run:

```sh
proviso init
```

It writes what the repository is missing and says, file by file, what
it wrote and what it left alone. It never replaces a file that exists.

| Path | What it is |
| --- | --- |
| `proviso/system.md` | the start of the spec tree: the system, with six sections to decide |
| `proviso/.gitattributes`, `proviso/.markdownlint.jsonc` | settings that keep the tree on line feeds and markdownlint content with the notation |
| `.claude/skills/proviso-integrate/`, `proviso-build/`, `proviso-check/`, `proviso-accept/` | the skills that bring a request into the tree, and build, check and accept an amendment |
| `.claude/agents/proviso-tester.md`, `proviso-coder.md`, `proviso-checker.md` | the testing, coding and checking agents |
| `.proviso/build.md`, `tester.md`, `coder.md` | your notes to the skills and the implementing agents |

Commit all of it. The installed skills and agents are shared by the team
through the repository; Proviso owns them, so do not edit them.

## Write your notes

The skills build, test and check your system with your project's own
commands. Proviso assumes no language or build tool, so tell it in
`.proviso/build.md`:

```markdown
# Notes for the Proviso skills

Commands, in the order integration runs them:

1. `npm ci`
2. `npm run build`
3. `npm test`
4. `npm run test:acceptance`

Acceptance: pull request
```

- The commands build the system, run its unit tests and run its
  acceptance tests, the acceptance tests last. Name skips that are
  expected; any other skip counts as a failure.
- Name Proviso's command here if it is not `proviso` from the path.
- Without the line `Acceptance: pull request`, acceptance merges
  directly into the target branch.

`.proviso/tester.md` and `.proviso/coder.md` hold working practice for
the testing and the coding agent: how the tests are run and kept fast,
code conventions, commit messages. Decisions about the system never go
into notes; they belong in the spec tree, where Proviso sees them.

## Write the tree

Open `proviso/system.md` and decide its six sections: Goal, Rationale,
Scope, Out of scope, Audience and Design boundary. Each starts as `?`,
an undecided place. Then add a capability, a feature and its first
rules. [Writing specs](writing-specs.md) explains the notation, and
[Proviso's own tree](https://github.com/bbak/proviso/tree/main/proviso)
is a complete example.

Check the tree as you write:

```sh
proviso lint      # errors in the notation; exit status 0 means none
proviso diff      # what the amendment adds and changes
proviso impact    # what else it may affect
```

`diff` lists an element with an undecided place as not to be built:
the amendment is neither built nor accepted until it is decided. Commit
the tree on the branch:

```sh
git add proviso
git commit -m "docs(spec): specify the first features"
```

## Integrate, build, check and accept

Instead of writing the tree yourself, you can ask Claude Code to
**integrate a request**: `proviso-integrate` works it into the tree with
you, asking the questions the tree needs answered, and hands over to the
check. [Workflow](workflow.md) describes it.

In Claude Code, on the amendment's branch, ask in turn:

1. **"Build the amendment"** runs `proviso-build`. The testing agent and
   the coding agent build from the tree, each in its own worktree. The
   skill integrates their work, runs your commands, and brings you every
   decision the tree did not make. You decide it in the tree, and the
   agents build again.
2. **"Check the amendment"** runs `proviso-check`. An independent agent
   judges whether the tree still holds together. You resolve each
   finding by an edit, or dismiss it with a reason.
3. **"Accept the amendment"** runs `proviso-accept`. It merges the
   amendment into the target branch, usually `main`, and records the
   tree as the new baseline in `proviso/manifest.sum`. It never pushes.

After that, `proviso verify` exits with 0 on the target branch: the tree
is the baseline, and no check is left over. Every later change is an
amendment of its own, on its own branch.

## Next

- [Repository setup](repository-setup.md): run `proviso verify` in CI,
  so that no amendment reaches the target branch unchecked.
- [Workflow](workflow.md): what each skill does, and what you decide on
  the way.
