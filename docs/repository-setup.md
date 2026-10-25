# Repository setup

What a repository needs beyond `proviso init`, so that every amendment
goes through its check, whoever made it and however.

## Run `proviso verify` in CI

`proviso verify` tells whether the spec tree on a branch may be the
baseline as it stands: either it is what the manifest records and no
check is left over, or a coherence check covers the amendment's final
state, with every candidate judged and no finding open. It judges
nothing itself and needs no agent.

| Exit status | Meaning |
| --- | --- |
| 0 | the tree may be the baseline |
| 1 | something is still to check: the check is missing or left over, or a candidate is unjudged, an element stale or a finding open |
| 2 | it could not run, such as for an error in the tree |

Run it wherever the project's checks run:

- **In CI, on every branch and every pull request**, not only on the
  default branch: every target branch holds a manifest of its own. On an
  amendment's branch it fails until the check covers the amendment, so
  make it required where you merge, not for every push.
- **In the local checks,** through whatever the project already uses,
  such as make, npm scripts, Gradle or a pre-push hook.

Proviso needs no build tool and calls none; its exit status is the whole
interface. Put the binary from the
[releases](https://github.com/bbak/proviso/releases) on the CI runner's
path, and run:

```sh
proviso verify
```

**What it catches.** A hand edit committed straight to the target, or a
merge that skipped acceptance, turns the target red: its tree no longer
matches the manifest, and no check covers it. The remedy is the
skipped step: check the amendment on the target and accept it there.

**What makes it a gate.** CI alone only detects. With pull requests,
branch protection that requires the CI check stops an unchecked
amendment before the merge. Teams that merge directly get detection
only. A git hook is no gate either: a clone does not install it, and it
can be skipped.

## Write the engineers' notes

The files in `.proviso/` belong to the engineers. `proviso init` writes
stubs and never touches them again.

- **`build.md`**, for the skills: the commands that build the system, run
  its unit tests and run its acceptance tests, in that order; skips that
  are expected; Proviso's command, if it is not `proviso` from the path;
  and the line `Acceptance: pull request` if acceptance goes through one.
- **`tester.md`**, for the testing agent: how the acceptance tests are
  written, run and kept fast, and commit conventions.
- **`coder.md`**, for the coding agent: code conventions, commands, and
  commit conventions.

Notes hold working practice only. A decision about the system, such as a
library to use or a limit to keep, is a constraint in the spec tree,
where the element diff and the check see it. Where notes contradict the
context pack, the pack wins. The checking agent gets no notes, so that
nothing steers its judgement.

## Commit what Proviso installs

Commit the installed skills and agent definitions in `.claude/`, so
that the whole team works with the same ones. Do not edit them: Proviso
owns them, and they match the version of Proviso that installed them.
Keep that version on everyone's path and in CI.

## Keep the tree readable everywhere

- **Run `proviso init` in the repository's root.** It refuses anywhere
  else, and puts the spec tree in `proviso/`.
- **Keep the settings `init` writes into `proviso/`.** `.gitattributes`
  checks the tree out with line feeds on every platform, as `lint`
  requires; `.markdownlint.jsonc` makes markdownlint accept the
  notation.
- **No symbolic links in the tree.** `lint` never follows one and stops
  instead.
- **Names starting with a dot are skipped:** a hidden file never holds
  spec content.
- **Elements cannot be shared between trees,** and a system spanning
  several repositories is not supported yet.

## Conflicts in the manifest

A conflict in `proviso/manifest.sum` means two amendments met. Never
resolve it by running `proviso manifest`: that would record a tree
nobody checked. Take the amendment's side, and let the check show what
the two did to each other; `proviso-accept` does exactly that.
