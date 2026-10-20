// Package tree finds a spec root and reads the files of its spec tree,
// placing each file in the tree's layout.
package tree

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// RootName is the name of the folder that holds a spec tree.
const RootName = "proviso"

// ErrNoRoot is returned when no spec root lies at or above a directory.
var ErrNoRoot = errors.New("no spec root found")

// FindRoot returns the nearest spec root at or above dir: a folder named
// proviso/ holding system.md. Names match exactly, case included, on every
// platform.
func FindRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		root := filepath.Join(dir, RootName)
		if holds(dir, RootName, true) && holds(root, "system.md", false) {
			return root, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoRoot
		}
		dir = parent
	}
}

// Kind is where a file stands in the layout.
type Kind int

// The kinds of file the layout lists, and Unknown for any other.
const (
	Unknown Kind = iota
	System
	Capability
	Feature
	Glossary
	Parameters
	References
	Invariants
	Constraints
	Questions
	Contract
	Settings
)

var siblings = map[string]Kind{
	"glossary.md":    Glossary,
	"parameters.md":  Parameters,
	"references.md":  References,
	"invariants.md":  Invariants,
	"constraints.md": Constraints,
	"questions.md":   Questions,
}

// IsSibling reports whether a file of kind k belongs to a sibling, and so
// opens with a title.
func (k Kind) IsSibling() bool {
	switch k {
	case Glossary, Parameters, References, Invariants, Constraints, Questions, Contract:
		return true
	}
	return false
}

// File is a file of the spec tree that the layout lists.
type File struct {
	Rel  string // slash-separated, relative to the spec root
	Kind Kind
	Cap  string // the capability whose folder holds the file; "" at system level
	Data []byte
}

// holds reports whether the folder dir holds an entry named exactly name,
// a folder if folder is true, and anything else if not. Lstat alone would
// ignore case where the file system does, as on Windows and macOS; the
// folder's listing gives the name as it is stored.
func holds(dir, name string, folder bool) bool {
	if _, err := os.Lstat(filepath.Join(dir, name)); err != nil {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return hasEntry(entries, name, folder)
}

// hasEntry reports whether entries hold one named exactly name, read in
// NFC, a plain folder if folder is true, and anything but a folder if not.
func hasEntry(entries []os.DirEntry, name string, folder bool) bool {
	return slices.ContainsFunc(entries, func(e os.DirEntry) bool {
		return norm.NFC.String(e.Name()) == name && (e.Type() == fs.ModeDir) == folder
	})
}

// CapFolder is a folder in capabilities/.
type CapFolder struct {
	Rel           string // slash-separated, relative to the spec root
	HasCapability bool
}

// Tree is what lint read from a spec root.
type Tree struct {
	Root       string // absolute path of the spec root
	Files      []*File
	Unknown    []string // files the layout does not list, relative to the spec root
	CapFolders []CapFolder
	Unreadable []string // relative to the spec root
	Malformed  []string // files and folders whose name does not read the same on every platform, or that share their name in NFC with another entry of their folder (once for all of them), relative to the spec root
}

// Classify places a slash-separated path, relative to the spec root, in the
// layout.
func Classify(rel string) (Kind, string) {
	parts := strings.Split(rel, "/")
	if k, cap := classifyLevel(parts, ""); k != Unknown {
		return k, cap
	}
	if len(parts) == 1 {
		switch parts[0] {
		case "system.md":
			return System, ""
		case ".markdownlint.jsonc":
			return Settings, ""
		}
	}
	if len(parts) >= 3 && parts[0] == "capabilities" {
		cap := parts[1]
		rest := parts[2:]
		if len(rest) == 1 && rest[0] == "capability.md" {
			return Capability, cap
		}
		if len(rest) == 2 && rest[0] == "features" && isMarkdown(rest[1]) {
			return Feature, cap
		}
		return classifyLevel(rest, cap)
	}
	return Unknown, ""
}

// classifyLevel places the files that may stand at any level.
func classifyLevel(parts []string, cap string) (Kind, string) {
	if len(parts) == 1 {
		if k, ok := siblings[parts[0]]; ok {
			return k, cap
		}
	}
	if len(parts) == 2 && parts[0] == "contracts" && isMarkdown(parts[1]) {
		return Contract, cap
	}
	return Unknown, ""
}

func isMarkdown(name string) bool {
	return strings.HasSuffix(name, ".md") && len(name) > len(".md")
}

// hidden reports whether lint skips the entry rel.
func hidden(rel string) bool {
	return strings.HasPrefix(path.Base(rel), ".") && rel != ".markdownlint.jsonc"
}

// wellNamed reports whether name, read in NFC, holds only lowercase
// letters, marks, digits, hyphens and dots, and does not end with a dot. A
// letter is lowercase when it has no other lowercase form, so letters of a
// script without case count; a digit is a decimal digit of any script
// (Unicode's Nd).
func wellNamed(name string) bool {
	if strings.HasSuffix(name, ".") {
		return false
	}
	for _, r := range name {
		switch {
		case r == '-', r == '.', unicode.IsMark(r), unicode.IsDigit(r):
		case unicode.IsLetter(r) && unicode.ToLower(r) == r:
		default:
			return false
		}
	}
	return true
}

// Load reads the spec tree at root, every name read in NFC. It never
// follows a link: a symbolic link, a Windows junction or any other entry
// that is neither a regular file nor a folder is unreadable wherever it
// stands.
func Load(root string) *Tree {
	t := &Tree{Root: root}
	t.walk("", t.Root)
	return t
}

// walk reads the folder disk, whose path relative to the spec root, read
// in NFC, is dir.
func (t *Tree) walk(dir, disk string) {
	entries, err := os.ReadDir(disk)
	if path.Dir(dir) == "capabilities" {
		t.CapFolders = append(t.CapFolders, CapFolder{Rel: dir, HasCapability: hasEntry(entries, "capability.md", false)})
	}
	if err != nil {
		t.Unreadable = append(t.Unreadable, dirName(dir))
		return
	}
	seen := map[string]int{}
	for _, e := range entries {
		seen[norm.NFC.String(e.Name())]++
	}
	for _, e := range entries {
		name := norm.NFC.String(e.Name())
		rel := path.Join(dir, name)
		if hidden(rel) {
			continue
		}
		if !wellNamed(name) || seen[name] > 1 {
			if seen[name] > 0 { // entries that share a name in NFC are reported once
				t.Malformed = append(t.Malformed, rel)
			}
			seen[name] = 0
		}
		mode := e.Type()
		switch {
		case mode == fs.ModeDir:
			t.walk(rel, filepath.Join(disk, e.Name()))
		case mode.IsRegular():
			t.read(rel, filepath.Join(disk, e.Name()))
		default:
			t.Unreadable = append(t.Unreadable, rel)
		}
	}
}

func dirName(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

// read reads the regular file disk, at rel, if the layout lists it; a file
// it does not list is never opened.
func (t *Tree) read(rel, disk string) {
	kind, cap := Classify(rel)
	if kind == Unknown {
		t.Unknown = append(t.Unknown, rel)
		return
	}
	data, err := os.ReadFile(disk)
	if err != nil || !utf8.Valid(data) {
		t.Unreadable = append(t.Unreadable, rel)
		return
	}
	t.Files = append(t.Files, &File{Rel: rel, Kind: kind, Cap: cap, Data: data})
}
