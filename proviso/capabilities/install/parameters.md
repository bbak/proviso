# Parameters

## <a id="install.installed-files"></a> installed_files

- Configurable: no
- Rationale: a repository needs the start of a spec tree and its
  settings, the definitions of the implementing agents and the checking
  agent and the skills for the first agent host, Claude Code, and a
  place for the engineers' notes to the implementing agents and the
  skills; the checking agent gets no notes, so that nothing steers its
  judgement. Nothing more is installed until it is needed. The agent
  that renders the business view of an amendment gets no definition:
  `proviso-integrate` starts it fresh with only what it renders from,
  which keeps it independent of whoever wrote the amendment, and what
  it writes reaches the engineer before anyone else. A definition of
  its own, `proviso-amendment-renderer`, follows when its instructions
  need to be shared or its tools restricted.

The paths are relative to the directory `init` runs in, with `/` as the
separator.

| path | written as |
| --- | --- |
| `.claude/agents/proviso-checker.md` | the kit's agent definition of the checking agent |
| `.claude/agents/proviso-coder.md` | the kit's agent definition of the coding agent |
| `.claude/agents/proviso-tester.md` | the kit's agent definition of the testing agent |
| `.claude/skills/proviso-accept/SKILL.md` | the kit's skill that accepts an amendment |
| `.claude/skills/proviso-build/SKILL.md` | the kit's skill that builds an amendment |
| `.claude/skills/proviso-check/SKILL.md` | the kit's skill that checks an amendment |
| `.claude/skills/proviso-integrate/SKILL.md` | the kit's skill that brings a request into the spec tree |
| `.proviso/build.md` | the kit's stub for the engineers' notes to the skills |
| `.proviso/coder.md` | the kit's stub for the engineers' notes to the coding agent |
| `.proviso/tester.md` | the kit's stub for the engineers' notes to the testing agent |
| `proviso/.gitattributes` | `gitattributes_file` |
| `proviso/.markdownlint.jsonc` | `markdownlint_file` |
| `proviso/system.md` | `system_skeleton` |
