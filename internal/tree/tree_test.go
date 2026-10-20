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
		"system.md":                             "# x",
		".Hidden/X.md":                          "",
		"capabilities/Lint/glossary.md":         "# G",
		"capabilities/c/Capability.md":          "# c",
		"capabilities/d/features/übersicht.md":  "# u",
		"capabilities/c/features/Übersicht.md":  "# U", // in another folder than übersicht.md, which Windows would take for the same name
		"capabilities/d/features/検索.md":         "# k",
		"capabilities/d/features/my run.md":     "# m",
		"capabilities/d/features/run_old.md":    "# r",
		"capabilities/e/features/u\u0308ber.md": "# u", // stored decomposed
	})
	tr := Load(root)
	want := []string{"capabilities/Lint", "capabilities/c/Capability.md", "capabilities/c/features/Übersicht.md", "capabilities/d/features/my run.md", "capabilities/d/features/run_old.md"}
	if !slices.Equal(tr.Malformed, want) {
		t.Errorf("malformed %v, want %v", tr.Malformed, want)
	}
	if !slices.ContainsFunc(tr.Files, func(f *File) bool { return f.Rel == "capabilities/e/features/\u00fcber.md" && string(f.Data) == "# u" }) {
		t.Error("a decomposed name is not read in NFC")
	}
	for _, c := range tr.CapFolders {
		if c.HasCapability {
			t.Errorf("%s: Capability.md is not capability.md", c.Rel)
		}
	}
}

func TestWellNamed(t *testing.T) {
	cases := map[string]bool{
		"run.md":              true,
		"übersicht.md":        true,
		"u\u0308bersicht.md":  true, // a letter followed by a mark
		"検索.md":               true,
		".markdownlint.jsonc": true,
		"a-b2.md":             true,
		"run٢.md":             true, // a digit from another script
		"Run.md":              false,
		"Übersicht.md":        false,
		"\u01c5.md":           false, // a titlecase letter has a lowercase form
		"my run.md":           false,
		"run_old.md":          false,
		"run.md.":             false,
		"a:b.md":              false,
	}
	for name, want := range cases {
		if got := wellNamed(name); got != want {
			t.Errorf("%q: %v, want %v", name, got, want)
		}
	}
}

// Two names in one folder that are the same in NFC are one file on macOS
// and two on Linux: they are reported once, at the name they share.
func TestLoadNamesSameInNFC(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"system.md":                            "# x",
		"capabilities/c/features/übersicht.md": "# a",
		"capabilities/d/features/übersicht.md": "# a",
	})
	if err := os.WriteFile(filepath.Join(root, "capabilities", "c", "features", "übersicht.md"), []byte("# b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, "capabilities", "c", "features")); len(entries) != 2 {
		t.Skip("the file system keeps one name per NFC form")
	}
	tr := Load(root)
	want := []string{"capabilities/c/features/übersicht.md"}
	if !slices.Equal(tr.Malformed, want) {
		t.Errorf("malformed %q, want %q", tr.Malformed, want)
	}
}
