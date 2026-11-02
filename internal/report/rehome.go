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

type jsonRehomePlan struct {
	Root       string          `json:"root"`
	Renames    []jsonRename    `json:"renames"`
	Collisions []jsonCollision `json:"collisions"`
}

// RehomeJSON writes the rehome_plan: the spec root that was planned for,
// every rename and every collision.
func RehomeJSON(w io.Writer, root string, renames []rehome.Rename, collisions []rehome.Collision) error {
	p := jsonRehomePlan{Root: root, Renames: []jsonRename{}, Collisions: []jsonCollision{}}
	for _, r := range renames {
		p.Renames = append(p.Renames, jsonRename{r.From, r.To})
	}
	for _, c := range collisions {
		p.Collisions = append(p.Collisions, jsonCollision{c.Identifier, c.From})
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

// RehomeText writes one line per rename, the old identifier, an arrow and
// the new one, then one line per collision. Nothing is written when there
// is no rename.
func RehomeText(w io.Writer, renames []rehome.Rename, collisions []rehome.Collision) error {
	var sb strings.Builder
	for _, r := range renames {
		fmt.Fprintf(&sb, "%s → %s\n", r.From, r.To)
	}
	for _, c := range collisions {
		fmt.Fprintf(&sb, "collision: %s would be taken by %s\n", c.Identifier, strings.Join(c.From, ", "))
	}
	_, err := io.WriteString(w, sb.String())
	return err
}
