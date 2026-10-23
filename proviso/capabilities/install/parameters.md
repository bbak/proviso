# Parameters

## <a id="install.installed-files"></a> installed_files

- Configurable: no
- Rationale: a repository needs the start of a spec tree and its
  settings, the definitions of the implementing agents for the first
  agent host, Claude Code, and a place for the engineers' notes to them;
  nothing more is installed until it is needed.

The paths are relative to the directory `init` runs in, with `/` as the
separator.

| path | written as |
| --- | --- |
| `.claude/agents/proviso-coder.md` | the kit's agent definition of the coding agent |
| `.claude/agents/proviso-tester.md` | the kit's agent definition of the testing agent |
| `.proviso/coder.md` | the kit's stub for the engineers' notes to the coding agent |
| `.proviso/tester.md` | the kit's stub for the engineers' notes to the testing agent |
| `proviso/.gitattributes` | `gitattributes_file` |
| `proviso/.markdownlint.jsonc` | `markdownlint_file` |
| `proviso/system.md` | `system_skeleton` |
