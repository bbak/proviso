# Notes for the Proviso skills

Proviso's command: `go run ./cmd/proviso`, from the repository's root,
so that the skills always run the checked-out code.

Commands, in the order integration runs them, all from the integration
worktree:

1. `go vet ./...`
2. `task build`
3. `task test`
4. `task acceptance`, about five minutes.

The acceptance tests of a list of identifiers, with the fixture tests
every such run needs:

```sh
go test -timeout 30m ./acceptance/... -run '^(TestFixture.*|<names>)$'
```

`<names>` joins with `|` one `<name>(_.*)?` per identifier, `<name>`
being `Test` and the identifier with its dots and hyphens dropped and
each part that starts with a letter capitalised, as `system.test-names`
says: `install.version.stale` gives `TestInstallVersionStale(_.*)?`. An
identifier without a test of that name, such as a glossary entry,
selects nothing.

Expected skips:

- `TestLintRunUnreadable/derived:_neither_a_regular_file_nor_a_folder_(a_named_pipe)`
  on Windows, which has no named pipes in the file system.
- `TestInstallVersionModel/row2…` to `row5…`: the rows assume a second
  row of `model_releases`, "1 from 0.3.0", which no binary holds; the
  unit tests of `internal/stamp` cover them.
