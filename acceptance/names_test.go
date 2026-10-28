package acceptance

// Constraint system.test-names. It binds the
// acceptance tests themselves, so this test reads their source and the
// identifiers of the spec tree beside them, not the binary.

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// testName gives the name system.test-names derives from an identifier: Test,
// then every part between dots and hyphens, those starting with a letter
// with a capital first letter, those starting with a digit as they are.
func testName(id string) string {
	var b strings.Builder
	b.WriteString("Test")
	for _, part := range strings.FieldsFunc(id, func(r rune) bool { return r == '.' || r == '-' }) {
		r := []rune(part)
		if unicode.IsLetter(r[0]) {
			r[0] = unicode.ToUpper(r[0])
		}
		b.WriteString(string(r))
	}
	return b.String()
}

var (
	suffixPattern  = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	fixturePattern = regexp.MustCompile(`^TestFixture[A-Z][A-Za-z0-9]*$`)
	anchorPattern  = regexp.MustCompile(`^#{1,6}\s+<a id="([^"]+)"></a>`)
)

// testNameKind tells what a top-level test function's name is under
// system.test-names, given the names the identifiers give: "element" with the
// identifier, "fixture", "main", or "" where the name is none of them.
func testNameKind(name string, names map[string]string) (kind, id string) {
	if name == "TestMain" {
		return "main", ""
	}
	if fixturePattern.MatchString(name) {
		return "fixture", ""
	}
	base, suffix, hasSuffix := strings.Cut(name, "_")
	if hasSuffix && !suffixPattern.MatchString(suffix) {
		return "", ""
	}
	if id, ok := names[base]; ok {
		return "element", id
	}
	return "", ""
}

// specIdentifiers lists the identifiers of the spec tree in proviso/ beside
// the acceptance tests: the anchors that open a heading, outside fenced code
// blocks, with the file each stands in.
func specIdentifiers(t *testing.T) map[string][]string {
	t.Helper()
	root := filepath.Join("..", "proviso")
	ids := map[string][]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		fence := ""
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			trimmed := strings.TrimLeft(line, " ")
			if fence == "" && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
				fence = trimmed[:3]
				continue
			}
			if fence != "" {
				if strings.HasPrefix(trimmed, fence) {
					fence = ""
				}
				continue
			}
			if m := anchorPattern.FindStringSubmatch(line); m != nil {
				ids[m[1]] = append(ids[m[1]], filepath.ToSlash(p))
			}
		}
		return sc.Err()
	})
	if err != nil {
		t.Fatalf("reading the spec tree: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("the spec tree in proviso/ holds no identifier")
	}
	return ids
}

// topLevelTests lists the top-level functions named Test… in the acceptance
// tests, with the file each stands in.
func topLevelTests(t *testing.T) map[string]string {
	t.Helper()
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{}
	fset := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", file, err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "Test") {
				tests[fn.Name.Name] = file
			}
		}
	}
	return tests
}

func TestSystemTestNames(t *testing.T) {
	t.Parallel()
	ids := specIdentifiers(t)
	byName := map[string][]string{}
	for id := range ids {
		byName[testName(id)] = append(byName[testName(id)], id)
	}
	names := map[string]string{}
	for name, from := range byName {
		names[name] = from[0]
	}

	t.Run("derived: the name of install.version.stale", func(t *testing.T) {
		t.Parallel()
		if got := testName("install.version.stale"); got != "TestInstallVersionStale" {
			t.Errorf("system.test-names: want TestInstallVersionStale, got %s", got)
		}
	})
	t.Run("derived: the name of lint.run.root-2", func(t *testing.T) {
		t.Parallel()
		if got := testName("lint.run.root-2"); got != "TestLintRunRoot2" {
			t.Errorf("system.test-names: want TestLintRunRoot2, got %s", got)
		}
	})
	t.Run("derived: the names the property gives as examples", func(t *testing.T) {
		t.Parallel()
		known := map[string]string{"TestBaselineDiffFormer": "baseline.diff.former"}
		for _, c := range []struct{ name, kind, id string }{
			{"TestBaselineDiffFormer", "element", "baseline.diff.former"},
			{"TestBaselineDiffFormer_Pack", "element", "baseline.diff.former"},
			{"TestBaselineDiffFormer_2", "element", "baseline.diff.former"},
			{"TestBaselineDiffFormer_", "", ""},
			{"TestBaselineDiffFormer_Pack_Two", "", ""},
			{"TestBaselineDiffFormerPack", "", ""},
			{"TestFixtureLint", "fixture", ""},
			{"TestFixture2", "", ""},
			{"TestFixturelint", "", ""},
			{"TestFixture", "", ""},
			{"TestMain", "main", ""},
		} {
			if kind, id := testNameKind(c.name, known); kind != c.kind || id != c.id {
				t.Errorf("system.test-names: %s: want %q %q, got %q %q", c.name, c.kind, c.id, kind, id)
			}
		}
	})
	t.Run("derived: no two identifiers give the same name", func(t *testing.T) {
		t.Parallel()
		for name, from := range byName {
			if len(from) > 1 {
				sort.Strings(from)
				t.Errorf("system.test-names: the identifiers %s all give the name %s", strings.Join(from, ", "), name)
			}
		}
	})
	t.Run("derived: no identifier gives TestMain or a name starting with TestFixture", func(t *testing.T) {
		t.Parallel()
		for name, from := range byName {
			if name == "TestMain" || strings.HasPrefix(name, "TestFixture") {
				t.Errorf("system.test-names: the identifier %s gives the name %s", from[0], name)
			}
		}
	})
	t.Run("derived: every top-level test is named after an element, TestFixture… or TestMain", func(t *testing.T) {
		t.Parallel()
		for name, file := range topLevelTests(t) {
			if kind, _ := testNameKind(name, names); kind == "" {
				t.Errorf("system.test-names: %s in %s is named after no identifier of the spec tree, nor TestFixture… nor TestMain", name, file)
			}
		}
	})
}
