// Package tree finds a spec root and reads the files of its spec tree,
// placing each file in the tree's layout.
package tree

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// RootName is the name of the folder that holds a spec tree.
const RootName = "proviso"

// ErrNoRoot is returned when no spec root lies at or above a directory.
var ErrNoRoot = errors.New("no spec root found")

// FindRoot returns the nearest spec root at or above dir: a folder named
// proviso/ holding system.md.
func FindRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		root := filepath.Join(dir, RootName)
		if fi, err := os.Lstat(root); err == nil && fi.IsDir() {
			if fi, err := os.Lstat(filepath.Join(root, "system.md")); err == nil && !fi.IsDir() {
				return root, nil
			}
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

// Load reads the spec tree at root. It never follows a symbolic link: a link,
// like any other entry that is neither a regular file nor a folder, is
// unreadable.
func Load(root string) *Tree {
	t := &Tree{Root: root}
	t.walk("")
	return t
}

func (t *Tree) walk(dir string) {
	entries, err := os.ReadDir(filepath.Join(t.Root, filepath.FromSlash(dir)))
	if err != nil {
		t.Unreadable = append(t.Unreadable, dirName(dir))
		return
	}
	for _, e := range entries {
		rel := path.Join(dir, e.Name())
		if hidden(rel) {
			continue
		}
		mode := e.Type()
		switch {
		case mode&fs.ModeSymlink != 0:
			t.Unreadable = append(t.Unreadable, rel)
		case mode.IsDir():
			if dir == "capabilities" {
				t.CapFolders = append(t.CapFolders, CapFolder{Rel: rel, HasCapability: t.holdsFile(rel, "capability.md")})
			}
			t.walk(rel)
		default:
			t.read(rel, mode.IsRegular())
		}
	}
}

func dirName(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

func (t *Tree) holdsFile(dir, name string) bool {
	fi, err := os.Lstat(filepath.Join(t.Root, filepath.FromSlash(dir), name))
	return err == nil && !fi.IsDir()
}

// read reads a file the layout lists; a file it does not list is never
// opened. A listed entry that is not a regular file, such as a pipe, is
// unreadable.
func (t *Tree) read(rel string, regular bool) {
	kind, cap := Classify(rel)
	if kind == Unknown {
		t.Unknown = append(t.Unknown, rel)
		return
	}
	if !regular {
		t.Unreadable = append(t.Unreadable, rel)
		return
	}
	data, err := os.ReadFile(filepath.Join(t.Root, filepath.FromSlash(rel)))
	if err != nil || !utf8.Valid(data) {
		t.Unreadable = append(t.Unreadable, rel)
		return
	}
	t.Files = append(t.Files, &File{Rel: rel, Kind: kind, Cap: cap, Data: data})
}
