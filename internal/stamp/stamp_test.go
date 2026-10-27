package stamp

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/baseline"
)

func TestAdd(t *testing.T) {
	got := Add("---\nname: x\n---\n\n# X\n", "0.1.0")
	if got != "---\nname: x\nmetadata:\n  proviso-version: 0.1.0\n---\n\n# X\n" {
		t.Errorf("%q", got)
	}
	if v := Read([]byte(got)); !slices.Equal(v, []string{"0.1.0"}) {
		t.Errorf("read back %q", v)
	}
	if got := Add("# X\n", "0.1.0"); got != "# X\n" {
		t.Errorf("%q", got)
	}
}

func TestRead(t *testing.T) {
	const head = "---\nname: proviso-build\n"
	cases := []struct {
		name string
		data string
		want []string
	}{
		{"one", head + "metadata:\n  proviso-version: 0.1.0\n---\n", []string{"0.1.0"}},
		{"two", head + "metadata:\n  proviso-version: 0.1.0\n  proviso-version: 0.2.0\n---\n", []string{"0.1.0", "0.2.0"}},
		{"two blocks", head + "metadata:\n  proviso-version: 0.1.0\nmetadata:\n  proviso-version: 0.1.0\n---\n", []string{"0.1.0", "0.1.0"}},
		{"empty counts", head + "metadata:\n  proviso-version: 0.1.0\n  proviso-version:\n---\n", []string{"0.1.0", ""}},
		{"flow", head + "metadata: {proviso-version: 0.1.0}\n---\n", nil},
		{"top", head + "proviso-version: 0.1.0\n---\n", nil},
		{"lone carriage return at the end", head + "metadata:\n  proviso-version: 0.1.0\n---\r", nil},
		{"carriage return within the stamp", head + "metadata:\n  proviso-version: 0.1.0\r \n---\n", []string{"0.1.0\r"}},
		{"nested metadata", head + "other:\n  metadata:\n    proviso-version: 0.1.0\n---\n", nil},
		{"blank line", head + "metadata:\n\n  proviso-version: 0.1.0\n---\n", nil},
		{"unclosed", head + "metadata:\n  proviso-version: 0.1.0\n", nil},
		{"no frontmatter", "# X\nmetadata:\n  proviso-version: 0.1.0\n", nil},
		{"crlf", strings.ReplaceAll(head+"metadata:\n  proviso-version: 0.1.0\n---\n", "\n", "\r\n"), []string{"0.1.0"}},
		{"comment", head + "metadata:\n  # installed by init\n  proviso-version: 0.1.0\n---\n", []string{"0.1.0"}},
		{"whitespace", head + "metadata :  \n\tproviso-version :0.1.0  \n---\n", []string{"0.1.0"}},
		{"after the frontmatter", head + "---\nmetadata:\n  proviso-version: 0.1.0\n", nil},
	}
	for _, c := range cases {
		if got := Read([]byte(c.data)); !slices.Equal(got, c.want) {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestNames(t *testing.T) {
	for v, want := range map[string]bool{"0.1.0": true, "0.1.1-dev": true, "dev": true, "0.1.0+build.5": true, "[1]": false, "v0.1.0": false, "0.1": false, "": false, "01.0.0": false} {
		if Names(v) != want {
			t.Errorf("%q: %v", v, !want)
		}
	}
}

func TestModel(t *testing.T) {
	one := []baseline.ModelRelease{{Model: 0, First: "0.1.0"}}
	two := []baseline.ModelRelease{{Model: 0, First: "0.1.0"}, {Model: 1, First: "0.3.0"}}
	cases := []struct {
		rows       []baseline.ModelRelease
		own, stamp string
		model      int
		ok         bool
	}{
		{one, "0.2.0", "0.1.0", 0, true},
		{two, "0.4.0", "0.2.5", 0, true},
		{two, "0.4.0", "0.3.0", 1, true},
		{two, "0.3.1-dev", "0.3.0", 1, true},
		{two, "dev", "0.4.2", 1, true},
		{one, "0.2.0", "dev", 0, false},
		{one, "0.2.0", "0.1.1-dev", 0, false},
		{one, "0.2.0", "0.1.0+build.5", 0, false},
		{one, "0.2.0", "0.3.0", 0, false},
		{one, "0.1.1-dev", "0.1.1", 0, false},
		{one, "0.2.0", "0.0.9", 0, false},
		{one, "0.10.0", "0.9.0", 0, true},
	}
	for _, c := range cases {
		m, ok := model(c.rows, c.stamp, c.own)
		if m != c.model || ok != c.ok {
			t.Errorf("%v, %s, stamp %s: %d %v", c.rows, c.own, c.stamp, m, ok)
		}
	}
}

func TestCompare(t *testing.T) {
	// The example order of semantic versioning 2.0.0.
	order := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1-dev", "1.0.1", "1.10.0"}
	for i := 1; i < len(order); i++ {
		a, _ := parse(order[i-1])
		b, _ := parse(order[i])
		if compare(a, b) >= 0 || compare(b, a) <= 0 {
			t.Errorf("%s not before %s", order[i-1], order[i])
		}
	}
}

// TestModelReleases checks system.model-release: every first release is a
// release version with patch 0, both columns rise, and the last row names
// model_version.
func TestModelReleases(t *testing.T) {
	rows := baseline.ModelReleases
	patch0 := regexp.MustCompile(`\.0$`)
	for i, r := range rows {
		if !isRelease(r.First) || !patch0.MatchString(r.First) {
			t.Errorf("%s is no release version with patch 0", r.First)
		}
		if i > 0 {
			a, _ := parse(rows[i-1].First)
			b, _ := parse(r.First)
			if r.Model <= rows[i-1].Model || compare(a, b) >= 0 {
				t.Errorf("row %d does not rise", i)
			}
		}
	}
	if rows[len(rows)-1].Model != baseline.ModelVersion {
		t.Error("the last row does not name model_version")
	}
}
