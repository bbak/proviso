# Notes for the Proviso skills

Proviso's command: `go run ./cmd/proviso`, from the repository's root,
so that the skills always run the checked-out code.

Commands, in the order integration runs them, all from the integration
worktree:

1. `go vet ./...`
2. `task build`
3. `task test`
4. `task acceptance`, about three minutes.

Expected skips:

- `TestLintRunUnreadable/derived:_neither_a_regular_file_nor_a_folder_(a_named_pipe)`
  on Windows, which has no named pipes in the file system.
