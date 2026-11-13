# Parameters

## <a id="install.installed-files"></a> installed_files

- Configurable: no
- Rationale: a repository needs the start of a spec tree and its
  settings, the definitions of the implementing agents, the checking
  agent and the mutating agent, and the skills for the first agent host, Claude Code, and a
  place for the engineers' notes to the implementing agents and the
  skills; the checking agent gets no notes, so that nothing steers its
  judgement. Nothing more is installed until it is needed. The agent
  that renders the business view of an amendment gets no definition:
  `proviso-integrate` starts it fresh with only what it renders from,
  which keeps it independent of whoever wrote the amendment, and what
  it writes reaches the engineer before anyone else. A definition of
  its own, `proviso-amendment-renderer`, follows when its instructions
  need to be shared or its tools restricted. Recovering a capability of
  an existing system is the first step of an amendment, as bringing in a
  request is: the engineer's, on the amendment's branch, so the skill
  that recovers the spec tree is installed with the other skills; only
  the authoring skill, for requesters who never see the tree, is
  distributed on its own. Measuring the tests with mutants is a step
  outside any amendment, but the engineer's in the same repository, so
  its skill is installed with the others; the mutating agent gets no
  notes, as nothing yet needs them. Answering an engineer's question
  about the system from the tree is the engineer's too, at any time, so
  its skill is installed with the others; it starts no agent of its
  own.

The paths are relative to the directory `init` runs in, with `/` as the
separator.

| path | written as |
| --- | --- |
| `.claude/agents/proviso-checker.md` | the kit's agent definition of the checking agent |
| `.claude/agents/proviso-coder.md` | the kit's agent definition of the coding agent |
| `.claude/agents/proviso-mutator.md` | the kit's agent definition of the mutating agent |
| `.claude/agents/proviso-tester.md` | the kit's agent definition of the testing agent |
| `.claude/skills/proviso-accept/SKILL.md` | the kit's skill that accepts an amendment |
| `.claude/skills/proviso-build/SKILL.md` | the kit's skill that builds an amendment |
| `.claude/skills/proviso-check/SKILL.md` | the kit's skill that checks an amendment |
| `.claude/skills/proviso-explain/SKILL.md` | the kit's skill that answers questions about the system from the spec tree |
| `.claude/skills/proviso-integrate/SKILL.md` | the kit's skill that brings a request into the spec tree |
| `.claude/skills/proviso-mutate/SKILL.md` | the kit's skill that measures with mutants how many slips the tests catch |
| `.claude/skills/proviso-recover/SKILL.md` | the kit's skill that recovers the spec tree of an existing system |
| `.claude/skills/proviso-rehome/SKILL.md` | the kit's skill that gives elements that moved, or whose identifiers name no place, the identifiers of their place |
| `.proviso/build.md` | the kit's stub for the engineers' notes to the skills |
| `.proviso/coder.md` | the kit's stub for the engineers' notes to the coding agent |
| `.proviso/tester.md` | the kit's stub for the engineers' notes to the testing agent |
| `proviso/.gitattributes` | `gitattributes_file` |
| `proviso/.markdownlint.jsonc` | `markdownlint_file` |
| `proviso/system.md` | `system_skeleton` |
