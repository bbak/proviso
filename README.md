# Proviso

Proviso keeps one specification of a system true across many changes.
The specification is a **spec tree**: Markdown files beside the system's
code. Every change starts as an **amendment**, an edit of that tree on a
branch. Proviso checks the tree, works out what the amendment changed and
what it may affect, and hands the agents that build and check it exactly
what they need. It serves systems in any language, and never reads their
code.

Agents write code faster than anyone can review it, so checking it has
to be automated as well: a testing agent, independent of the coding
agent, derives tests from the specification. Those tests check
something only if the specification states the system's behaviour
precisely: what happens when a condition does not hold, where its
boundaries lie, what its terms mean. The spec tree is written for that
reader, and Proviso keeps it true as the system changes: a format can
make someone decide, but it cannot make the decision arrive.

![The path of an amendment: request, specification, build, check, accept](docs/images/pipeline.png)

## Install

Download the archive for your platform from the
[releases](https://github.com/bbak/proviso/releases), unpack it, and put
`proviso` in a folder on your path.

## Set up a repository

The first spec tree is the first amendment, so start on a branch, in
the repository's root:

```sh
git switch -c first-spec
proviso init
```

`init` writes the start of the spec tree in `proviso/`, the skills and
agent definitions for Claude Code in `.claude/`, and stubs for your
notes in `.proviso/`. Name your project's build and test commands in
`.proviso/build.md`, and commit everything.

## Use it

1. **Integrate:** ask Claude Code to integrate a request, in any form.
   The Integrator works it into the tree with you on the amendment's
   branch, asks the questions the tree needs answered, sends what only
   the business can decide back as questions, and checks the result. Or
   edit the tree yourself: `proviso lint` checks the notation,
   `proviso diff` shows what changed, and `proviso impact` what it may
   affect; what the amendment deliberately leaves out goes into
   `proviso/.amendment.md`.
2. **Build:** ask Claude Code to build the amendment. A testing agent
   and a coding agent build from the tree, apart from each other; every
   decision the tree did not make comes back to you.
3. **Check:** ask it to check the amendment. An independent agent judges
   whether the tree still holds together; you resolve or dismiss each
   finding.
4. **Accept:** ask it to accept the amendment. It is merged, and the tree
   becomes the new baseline.

Run `proviso verify` in CI, so that no amendment reaches the target
branch unchecked.

## For business requesters

Requesters never see the spec tree. The skill `proviso-request` guides
them through a short session with their own AI assistant to one request
file, which the Integrator reads best: what is wanted and why, the rules
with examples, the words that need pinning down, and what is still
open. Each release carries it as `proviso-request_<version>.skill.zip`,
for any assistant that reads Agent Skills, such as Claude, ChatGPT,
Microsoft Copilot or Gemini. Its check needs Python 3, which those
assistants bring along; without it, the skill checks by reading.

## Status

Proviso has been validated only on command-line applications written in
Go, with Claude Code and primarily Claude Opus 5.5.

## Documentation

The [documentation](docs/index.md) covers
[getting started](docs/getting-started.md),
[concepts](docs/concepts.md), the [workflow](docs/workflow.md),
[writing specs](docs/writing-specs.md),
[repository setup](docs/repository-setup.md) and the
[commands](docs/commands.md). Proviso specifies itself: its own spec tree
is in [`proviso/`](proviso/).

## Building Proviso

Building Proviso needs [Go](https://go.dev/), [Task](https://taskfile.dev/)
and [Python 3](https://www.python.org/), for the tests of the request
skill's check; `task snapshot` also needs
[GoReleaser](https://goreleaser.com/). `task build` builds the binary
into `bin/`, and `task check` runs every check before a push.

## License

Copyright 2026 Bruno Baketarić. Licensed under the
[Apache License, Version 2.0](LICENSE). The binary contains code from
other projects under their own licenses; see
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
