package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bbak/proviso/internal/rehome"
)

type jsonRename struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type jsonCollision struct {
	Identifier string   `json:"identifier"`
	From       []string `json:"from"`
}

type jsonMention struct {
	Identifier string  `json:"identifier"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Count      int     `json:"count"`
	Element    *string `json:"element"`
}

type jsonRehomePlan struct {
	Root       string          `json:"root"`
	Renames    []jsonRename    `json:"renames"`
	Collisions []jsonCollision `json:"collisions"`
	Mentions   []jsonMention   `json:"mentions"`
}

// RehomeJSON writes the rehome_plan: the spec root that was planned for,
// every rename, every collision and every mention of a renamed
// identifier.
func RehomeJSON(w io.Writer, root string, renames []rehome.Rename, collisions []rehome.Collision, mentions []rehome.Mention) error {
	p := jsonRehomePlan{Root: root, Renames: []jsonRename{}, Collisions: []jsonCollision{}, Mentions: []jsonMention{}}
	for _, r := range renames {
		p.Renames = append(p.Renames, jsonRename{r.From, r.To})
	}
	for _, c := range collisions {
		p.Collisions = append(p.Collisions, jsonCollision{c.Identifier, c.From})
	}
	for _, m := range mentions {
		var element *string
		if m.Element != "" {
			element = &m.Element
		}
		p.Mentions = append(p.Mentions, jsonMention{m.Identifier, m.File, m.Line, m.Count, element})
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

// RehomeText writes one line per rename, the old identifier, an arrow and
// the new one, then one line per collision, then one line per mention of
// a renamed identifier: where it stands, how often, and in which element.
// Nothing is written when there is no rename.
func RehomeText(w io.Writer, renames []rehome.Rename, collisions []rehome.Collision, mentions []rehome.Mention) error {
	var sb strings.Builder
	for _, r := range renames {
		fmt.Fprintf(&sb, "%s → %s\n", r.From, r.To)
	}
	for _, c := range collisions {
		fmt.Fprintf(&sb, "collision: %s would be taken by %s\n", c.Identifier, strings.Join(c.From, ", "))
	}
	for _, m := range mentions {
		in := "in no element"
		if m.Element != "" {
			in = "in " + m.Element
		}
		fmt.Fprintf(&sb, "mention: %s:%d: %s, %s, %s\n", m.File, m.Line, m.Identifier, times(m.Count), in)
	}
	_, err := io.WriteString(w, sb.String())
	return err
}

func times(n int) string {
	if n == 1 {
		return "once"
	}
	return fmt.Sprintf("%d times", n)
}
