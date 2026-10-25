# <a id="install"></a> Install

## Goal

Installing gives a repository what Proviso needs beside the system's
code: the start of a spec tree that lint accepts, the settings that keep
the tree readable on every platform, the definitions of the agents and
the skills that bring a request into the spec tree and build, check and
accept an amendment, so that an engineer begins with one command.

## Rationale

Every file `init` writes is one the engineers go on to change, so `init`
writes only what is missing and never replaces a file; refreshing what
Proviso itself owns is the job of an update, deferred for now. The
wording of every installed file comes from the kit: the agents'
instructions are Proviso's own behaviour, not the specified system's,
and the files that become part of the spec tree are stated here by
what they do, where a test can check it from outside. Keeping their
wording right is the duty of Proviso's developers, since checking it
would mean keeping it twice.

`init` works only in the root of a git repository, recognised by its
`.git` entry rather than by asking git, since Proviso asks git only for
history.
