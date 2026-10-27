package kit

import (
	"bytes"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/bbak/proviso/internal/stamp"
)

// proviso matches a path below .proviso/.
var proviso = regexp.MustCompile("\\.proviso/[^\\s`'\")]*")

func TestFiles(t *testing.T) {
	got := Files("shop", "0.1.1-dev")
	if len(got) != 13 {
		t.Fatalf("%d files, want 13", len(got))
	}
	reading := read("agents/reading.md")
	for i, f := range got {
		if i > 0 && got[i-1].Path >= f.Path {
			t.Errorf("%s not after %s", f.Path, got[i-1].Path)
		}
		if bytes.Contains(f.Data, []byte("\r")) {
			t.Errorf("%s holds a carriage return", f.Path)
		}
		if f.Stamped != strings.HasPrefix(f.Path, ".claude/") {
			t.Errorf("%s: stamped %v", f.Path, f.Stamped)
		}
		if v := stamp.Read(f.Data); f.Stamped && !slices.Equal(v, []string{"0.1.1-dev"}) || !f.Stamped && v != nil {
			t.Errorf("%s: stamps %q", f.Path, v)
		}
		if strings.HasPrefix(f.Path, ".claude/") {
			// A kit file that is not installed, or a kit path, would point to
			// nothing on the agent host.
			for _, ref := range []string{"reading.md", "writing.md", "proviso: ", "agents/", "notes/", "tree/", "skills/"} {
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
			// Below .proviso/ only the notes to that agent; the checker has none.
			notes := map[string]string{"proviso-coder": ".proviso/coder.md", "proviso-tester": ".proviso/tester.md"}[name]
			for _, ref := range proviso.FindAllString(string(f.Data), -1) {
				if ref != notes {
					t.Errorf("%s refers to %s", f.Path, ref)
				}
			}
		}
	}
	writing := read("skills/writing.md")
	skills := 0
	for _, f := range got {
		if !strings.HasPrefix(f.Path, ".claude/skills/") {
			continue
		}
		skills++
		name := strings.TrimSuffix(strings.TrimPrefix(f.Path, ".claude/skills/"), "/SKILL.md")
		// The skills that edit the spec tree carry the writing guide where
		// they mark it; proviso-accept never edits it and carries none.
		kit := read("skills/" + name + "/SKILL.md")
		want := stamp.Add(kit, "0.1.1-dev")
		if name != "proviso-accept" {
			if !strings.Contains(kit, writingMark) {
				t.Errorf("%s: the kit's skill marks no writing guide", f.Path)
			}
			want = stamp.Add(strings.Replace(kit, writingMark, writing, 1), "0.1.1-dev")
		}
		if string(f.Data) != want {
			t.Errorf("%s is not the kit's skill with the writing guide where it marks it", f.Path)
		}
		if !bytes.Contains(f.Data, []byte("\nname: "+name+"\n")) {
			t.Errorf("%s not named %s", f.Path, name)
		}
	}
	if skills != 4 {
		t.Errorf("%d skills, want 4", skills)
	}
	if sys := got[len(got)-1]; sys.Path != "proviso/system.md" || !bytes.HasPrefix(sys.Data, []byte("# <a id=\"system\"></a> shop\n")) {
		t.Errorf("skeleton %s: %q", sys.Path, sys.Data)
	}
}
