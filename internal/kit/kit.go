// Package kit holds the files `proviso init` installs into a repository,
// embedded from this folder, and the paths install.installed-files puts
// them at.
package kit

import (
	"embed"
	"sort"
	"strings"

	"github.com/bbak/proviso/internal/stamp"
)

//go:embed agents notes skills tree
var files embed.FS

// readingMark is the line in an agent definition that the shared reading
// guide replaces, so that the installed definition refers to no other
// file of the kit.
const readingMark = "<!-- proviso: reading.md -->\n"

// writingMark is the line in a skill that edits the spec tree that the
// shared writing guide replaces, so that the installed skill refers to no
// other file of the kit. A skill without it carries no writing guide.
const writingMark = "<!-- proviso: writing.md -->\n"

// nameMark is the placeholder in the system skeleton that the system's
// name replaces.
const nameMark = "<name>"

// File is an installed file: its path relative to the directory init runs
// in, with "/" as the separator, and its content. Stamped tells a skill or
// agent definition, which carries a version stamp.
type File struct {
	Path    string
	Data    []byte
	Stamped bool
}

// Files returns the installed files in the byte order of their paths, the
// system skeleton named system, the skills and agent definitions stamped
// with the Proviso version version.
func Files(system, version string) []File {
	reading := read("agents/reading.md")
	agent := func(name string) []byte {
		return []byte(stamp.Add(strings.Replace(read("agents/"+name), readingMark, reading, 1), version))
	}
	writing := read("skills/writing.md")
	skill := func(name string) []byte {
		return []byte(stamp.Add(strings.Replace(read("skills/"+name+"/SKILL.md"), writingMark, writing, 1), version))
	}
	out := []File{
		{".claude/agents/proviso-checker.md", agent("proviso-checker.md"), true},
		{".claude/agents/proviso-coder.md", agent("proviso-coder.md"), true},
		{".claude/agents/proviso-tester.md", agent("proviso-tester.md"), true},
		{".claude/skills/proviso-accept/SKILL.md", skill("proviso-accept"), true},
		{".claude/skills/proviso-build/SKILL.md", skill("proviso-build"), true},
		{".claude/skills/proviso-check/SKILL.md", skill("proviso-check"), true},
		{".claude/skills/proviso-integrate/SKILL.md", skill("proviso-integrate"), true},
		{".proviso/build.md", []byte(read("notes/build.md")), false},
		{".proviso/coder.md", []byte(read("notes/coder.md")), false},
		{".proviso/tester.md", []byte(read("notes/tester.md")), false},
		{"proviso/.gitattributes", []byte(read("tree/gitattributes")), false},
		{"proviso/.markdownlint.jsonc", []byte(read("tree/markdownlint.jsonc")), false},
		{"proviso/system.md", []byte(strings.Replace(read("tree/system.md"), nameMark, system, 1)), false},
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func read(name string) string {
	data, err := files.ReadFile(name)
	if err != nil {
		panic(err) // embedded at build time; a missing file is a build error
	}
	return string(data)
}
