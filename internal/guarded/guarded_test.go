package guarded

import (
	"testing"
)

func TestLex(t *testing.T) {
	l := parseLine(SourceLine{1, `Provided  <value> is "yes", "no (<part>)" or "?" in [spec tree] ` + "`lint.exit-status`" + ` [x](y)`})
	var got []string
	for _, sp := range l.Spans {
		if sp.Kind != Text {
			got = append(got, string("TLCMP"[sp.Kind])+":"+sp.Value)
		}
	}
	want := []string{"P:value", `L:yes`, `L:no (<part>)`, `L:?`, "M:spec tree", "C:lint.exit-status"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if l.Keyword != Provided {
		t.Errorf("keyword %q", l.Keyword)
	}
}

func TestHandOver(t *testing.T) {
	l := parseLine(SourceLine{1, "Otherwise → lint.layout.file"})
	if !l.IsHandOver() || l.Target != "lint.layout.file" {
		t.Fatalf("%+v", l)
	}
	if got := l.Without(true); got != "Otherwise → \x00" {
		t.Errorf("without target: %q", got)
	}
}

func TestSteps(t *testing.T) {
	b := Parse([]SourceLine{{1, "  And x"}, {2, "When a"}, {3, ""}, {4, "  And b"}, {5, "given c"}, {6, "Then d"}})
	if len(b.Lines) != 5 || len(b.LeadingAnds) != 1 || len(b.Steps) != 2 || len(b.Steps[0].Ands) != 1 || b.Lines[3].Keyword != "" {
		t.Errorf("%+v", b)
	}
}
