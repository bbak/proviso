package kit

import (
	"bytes"
	"strings"
	"testing"
)

func TestFiles(t *testing.T) {
	got := Files("shop")
	if len(got) != 12 {
		t.Fatalf("%d files, want 12", len(got))
	}
	reading := read("agents/reading.md")
	for i, f := range got {
		if i > 0 && got[i-1].Path >= f.Path {
			t.Errorf("%s not after %s", f.Path, got[i-1].Path)
		}
		if bytes.Contains(f.Data, []byte("\r")) {
			t.Errorf("%s holds a carriage return", f.Path)
		}
		if strings.HasPrefix(f.Path, ".claude/") {
			// A kit file that is not installed, or a kit path, would point to
			// nothing on the agent host.
			for _, ref := range []string{"reading.md", "proviso: ", "agents/", "notes/", "tree/", "skills/"} {
				if bytes.Contains(f.Data, []byte(ref)) {
					t.Errorf("%s refers to %q of the kit", f.Path, ref)
				}
			}
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
	skills := 0
	for _, f := range got {
		if !strings.HasPrefix(f.Path, ".claude/skills/") {
			continue
		}
		skills++
		name := strings.TrimSuffix(strings.TrimPrefix(f.Path, ".claude/skills/"), "/SKILL.md")
		if want := read("skills/" + name + "/SKILL.md"); string(f.Data) != want {
			t.Errorf("%s is not the kit's skill as the kit holds it", f.Path)
		}
		if !bytes.Contains(f.Data, []byte("\nname: "+name+"\n")) {
			t.Errorf("%s not named %s", f.Path, name)
		}
	}
	if skills != 3 {
		t.Errorf("%d skills, want 3", skills)
	}
	if sys := got[len(got)-1]; sys.Path != "proviso/system.md" || !bytes.HasPrefix(sys.Data, []byte("# <a id=\"system\"></a> shop\n")) {
		t.Errorf("skeleton %s: %q", sys.Path, sys.Data)
	}
}
