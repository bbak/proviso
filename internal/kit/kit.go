// Package kit holds the files `proviso init` installs into a repository,
// embedded from this folder, and the paths install.installed-files puts
// them at.
package kit

import (
	"embed"
	"sort"
	"strings"
)

//go:embed agents notes skills tree
var files embed.FS

// readingMark is the line in an agent definition that the shared reading
// guide replaces, so that the installed definition refers to no other
// file of the kit.
const readingMark = "<!-- proviso: reading.md -->\n"

// nameMark is the placeholder in the system skeleton that the system's
// name replaces.
const nameMark = "<name>"

// File is an installed file: its path relative to the directory init runs
// in, with "/" as the separator, and its content.
type File struct {
	Path string
	Data []byte
}

// Files returns the installed files in the byte order of their paths, the
// system skeleton named system.
func Files(system string) []File {
	reading := read("agents/reading.md")
	agent := func(name string) []byte {
		return []byte(strings.Replace(read("agents/"+name), readingMark, reading, 1))
	}
	out := []File{
		{".claude/agents/proviso-coder.md", agent("proviso-coder.md")},
		{".claude/agents/proviso-tester.md", agent("proviso-tester.md")},
		{".claude/skills/proviso-build/SKILL.md", []byte(read("skills/proviso-build/SKILL.md"))},
		{".proviso/build.md", []byte(read("notes/build.md"))},
		{".proviso/coder.md", []byte(read("notes/coder.md"))},
		{".proviso/tester.md", []byte(read("notes/tester.md"))},
		{"proviso/.gitattributes", []byte(read("tree/gitattributes"))},
		{"proviso/.markdownlint.jsonc", []byte(read("tree/markdownlint.jsonc"))},
		{"proviso/system.md", []byte(strings.Replace(read("tree/system.md"), nameMark, system, 1))},
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
