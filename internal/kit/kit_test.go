package kit

import (
	"bytes"
	"strings"
	"testing"
)

func TestFiles(t *testing.T) {
	got := Files("shop")
	if len(got) != 9 {
		t.Fatalf("%d files, want 9", len(got))
	}
	reading := read("agents/reading.md")
	for i, f := range got {
		if i > 0 && got[i-1].Path >= f.Path {
			t.Errorf("%s not after %s", f.Path, got[i-1].Path)
		}
		if bytes.Contains(f.Data, []byte("\r")) {
			t.Errorf("%s holds a carriage return", f.Path)
		}
		if strings.HasPrefix(f.Path, ".claude/agents/") {
			if bytes.Contains(f.Data, []byte("proviso: reading.md")) || !bytes.Contains(f.Data, []byte(reading)) {
				t.Errorf("%s: reading guide not in place", f.Path)
			}
			name := strings.TrimSuffix(strings.TrimPrefix(f.Path, ".claude/agents/"), ".md")
			if !bytes.Contains(f.Data, []byte("\nname: "+name+"\n")) {
				t.Errorf("%s not named %s", f.Path, name)
			}
		}
	}
	if sys := got[len(got)-1]; sys.Path != "proviso/system.md" || !bytes.HasPrefix(sys.Data, []byte("# <a id=\"system\"></a> shop\n")) {
		t.Errorf("skeleton %s: %q", sys.Path, sys.Data)
	}
}
