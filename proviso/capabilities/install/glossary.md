# Glossary

## <a id="install.kit"></a> kit

- Definition: The files Proviso carries in its binary for `init` to
  install: the start of a spec tree and its settings, the agent
  definitions, the reading guide they share, the skills, the writing
  guide the skills that edit the spec tree share, and the stubs
  of the engineers' notes to the implementing agents and the skills.
  Their wording is Proviso's own.
- Contact: Bruno Baketarić
- Status: defined

## <a id="install.installed-file"></a> installed file

- Definition: A file `init` writes into a repository, at a path that
  `installed_files` names.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: installed files

## <a id="install.agent-definition"></a> agent definition

- Definition: The instructions of an implementing agent or of the
  checking agent in the form an agent host reads, such as a file in `.claude/agents/`: its role, what
  it may read and write, how it reads a context pack and what it
  reports.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: agent definitions

## <a id="install.skill"></a> skill

- Definition: The instructions an agent host follows for one step of an
  amendment, such as a `SKILL.md` in a folder of `.claude/skills/`: when
  it applies, which commands of Proviso it runs and which agents it
  starts, and what it leaves to the engineer.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: skills

## <a id="install.release-version"></a> release version

- Definition: A version of Proviso in the form
  `<major>.<minor>.<patch>` of semantic versioning, in plain numbers,
  without a leading `v` and without a pre-release or build part, such
  as `0.1.0`: the version a release carries, its git tag without the
  `v`.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: release versions

## <a id="install.proviso-version"></a> Proviso version

- Definition: The version a proviso binary was built with: the release
  version of a release; for a local build after a release, the next
  patch version with the pre-release `dev`, such as `0.1.1-dev`; or
  `dev` for a build that was given none.
- Contact: Bruno Baketarić
- Status: defined

## <a id="install.version-stamp"></a> version stamp

- Definition: The line `proviso-version: <version>` in the frontmatter
  of an installed skill or agent definition, among the lines directly
  below the line `metadata:` that start with a space or a tab: the
  Proviso version of the `init` that wrote the file. The frontmatter is
  the lines between a first line `---` and the next line `---`; without
  that next line there is none. `metadata:` starts its line; the first
  line below it that starts with neither a space nor a tab, a blank one
  included, ends the lines below it. Each of these lines is read without
  its line ending, a line feed or a carriage return and a line feed.
  Spaces and tabs at the end of any of these lines, around the colons,
  and before `proviso-version`, change nothing. The agent host reads it
  as YAML; Proviso reads only these lines.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: stamp, version stamps, stamps
