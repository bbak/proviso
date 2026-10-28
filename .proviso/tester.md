# Notes for the testing agent

Working practice for Proviso's own acceptance tests. They add to the
definition of the testing agent; the context pack decides what is
tested.

- **Build and run:** `go vet ./acceptance/...` must be clean; run the
  suite with `task acceptance`, or one test with
  `go test ./acceptance/... -run '^TestBaselinePackFrames$'`.
- **Runtime matters.** Every process start costs about 0.2 s on Windows,
  and engineers skip a slow suite. Keep the whole suite within a few
  minutes:
  - **Run in parallel.** Every top-level test and every subtest calls
    `t.Parallel()`, unless it shares state that cannot be separated;
    say so in a comment where it does not.
  - **Start few processes.** Build a test tree's git history once per
    test, for example behind a `sync.Once`, and give each subtest a copy
    of the folder rather than running `git init`, `add` and `commit`
    again. Run the binary only as often as a check needs.
  - **Never share a folder** between parallel subtests: each writes only
    to its own `t.TempDir()`.
- **Names:** a test is named after the element it checks, as
  `system.test-names` says. While you work, run only the tests of the
  pack's elements and those you changed, such as
  `go test ./acceptance/... -run '^(TestFixture.*|TestLintRunModel(_.*)?)$'`.
- **Commit messages** use the prefix `test:`.
