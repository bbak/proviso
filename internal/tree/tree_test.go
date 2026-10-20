package tree

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFindRoot(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, map[string]string{
		"proviso/system.md":   "",
		"a/proviso/system.md": "",
		"b/proviso/x.md":      "",
		"c/system.md":         "",
		"cmd/proviso/main.go": "",
	})
	cases := map[string]string{
		"":            "proviso",
		"cmd/proviso": "proviso",
		"proviso":     "proviso",
		"a":           "a/proviso",
		"b":           "proviso",
		"c":           "proviso",
	}
	for dir, want := range cases {
		got, err := FindRoot(filepath.Join(repo, dir))
		if err != nil || got != filepath.Join(repo, filepath.FromSlash(want)) {
			t.Errorf("%s: got %q %v, want %s", dir, got, err, want)
		}
	}
	if _, err := FindRoot(t.TempDir()); err != ErrNoRoot {
		t.Errorf("no root: %v", err)
	}
}

func TestLoad(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"system.md":                             "# x",
		".markdownlint.jsonc":                   "{}",
		".DS_Store":                             "\xff",
		"notes.md":                              "",
		"capabilities/lint/.vscode/x.md":        "",
		"capabilities/lint/.markdownlint.jsonc": "",
		"capabilities/lint/glossary.md":         "# G",
		"capabilities/lint/features/a.md":       "\xff",
		"capabilities/lint/features/old/run.md": "",
		"invariants/read-only.md":               "",
	})
	tr := Load(root)
	var files []string
	for _, f := range tr.Files {
		files = append(files, f.Rel)
	}
	if want := []string{".markdownlint.jsonc", "capabilities/lint/glossary.md", "system.md"}; !slices.Equal(files, want) {
		t.Errorf("files %v, want %v", files, want)
	}
	if want := []string{"capabilities/lint/features/old/run.md", "invariants/read-only.md", "notes.md"}; !slices.Equal(tr.Unknown, want) {
		t.Errorf("unknown %v, want %v", tr.Unknown, want)
	}
	if want := []string{"capabilities/lint/features/a.md"}; !slices.Equal(tr.Unreadable, want) {
		t.Errorf("unreadable %v, want %v", tr.Unreadable, want)
	}
	if len(tr.CapFolders) != 1 || tr.CapFolders[0].HasCapability {
		t.Errorf("capability folders %+v", tr.CapFolders)
	}
}

func TestClassify(t *testing.T) {
	cases := map[string]Kind{
		"system.md":                          System,
		"capabilities/lint/capability.md":    Capability,
		"capabilities/lint/features/run.md":  Feature,
		"capabilities/lint/contracts/cli.md": Contract,
		"contracts/api.md":                   Contract,
		"questions.md":                       Questions,
		"capabilities/lint/README.md":        Unknown,
		"capabilities/lint/system.md":        Unknown,
		"features/run.md":                    Unknown,
	}
	for rel, want := range cases {
		if got, _ := Classify(rel); got != want {
			t.Errorf("%s: %v, want %v", rel, got, want)
		}
	}
}

func TestLoadLinks(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"system.md":                    "# x",
		"capabilities/c/capability.md": "# c",
		"capabilities/c/features/f.md": "# f",
	})
	for link, target := range map[string]string{
		".shared.md":                   "system.md",
		"glossary.md":                  "system.md",
		"capabilities/c/features/x.md": "f.md",
	} {
		if err := os.Symlink(target, filepath.Join(root, filepath.FromSlash(link))); err != nil {
			t.Skipf("cannot create symbolic links here: %v", err)
		}
	}
	tr := Load(root)
	if want := []string{"capabilities/c/features/x.md", "glossary.md"}; !slices.Equal(tr.Unreadable, want) {
		t.Errorf("unreadable %v, want %v", tr.Unreadable, want)
	}
	if len(tr.Unknown) != 0 {
		t.Errorf("unknown %v", tr.Unknown)
	}
}

func TestFindRootMatchesCase(t *testing.T) {
	for _, rel := range []string{"Proviso/system.md", "proviso/System.md"} {
		repo := t.TempDir()
		write(t, repo, map[string]string{rel: ""})
		if _, err := FindRoot(repo); err != ErrNoRoot {
			t.Errorf("%s: %v, want no root", rel, err)
		}
	}
}

func TestLoadNames(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"system.md":                            "# x",
		".Hidden/X.md":                         "",
		"capabilities/Lint/glossary.md":        "# G",
		"capabilities/c/Capability.md":         "# c",
		"capabilities/d/features/übersicht.md": "# u",
		"capabilities/c/features/Übersicht.md": "# U", // in another folder than übersicht.md, which Windows would take for the same name
	})
	tr := Load(root)
	want := []string{"capabilities/Lint", "capabilities/c/Capability.md", "capabilities/c/features/Übersicht.md"}
	if !slices.Equal(tr.Uppercase, want) {
		t.Errorf("uppercase %v, want %v", tr.Uppercase, want)
	}
	for _, c := range tr.CapFolders {
		if c.HasCapability {
			t.Errorf("%s: Capability.md is not capability.md", c.Rel)
		}
	}
}
