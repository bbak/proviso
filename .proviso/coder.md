# Notes for the coding agent

Working practice for Proviso's own code. They add to the definition of
the coding agent; the context pack decides what is built.

- **Build and check:** `task build`, `task test` and `task vet` must
  pass. Build a binary for your own runs into a temporary folder, never
  into `bin/` or the repository root: a virus scanner may quarantine
  binaries in `bin/`.
- **Run Proviso on its own tree:** `lint`, and the command the amendment
  is about, on the worktree's `proviso/`, unless the command would write
  into it; then run it in a scratch repository. Report the results.
- **Reuse before adding:** the packages in `internal/` already read the
  tree, the manifest and git; extend them rather than duplicating them.
- **The files `init` installs** live as real files in `internal/kit/`,
  embedded with `//go:embed`, never as string literals in Go code.
- **Line endings** are LF in every file you write.
- **Commit messages** use the prefixes `feat:`, `fix:` or `refactor:`.
