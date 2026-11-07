package acceptance

// Feature upkeep.rehome: proviso/capabilities/upkeep/features/rehome.md, the
// parameter rehome_references, the terms identifier home and identifier
// mention, and the contracts system.cli.rehome, system.cli.rehome-plan and
// system.cli.mention-list (proviso/contracts/cli.md).
// What holds for every run (the plan's shape and order, the exit status and
// the invariant upkeep.rehome.writes) is checked by rehomeIn. Every tree is
// accepted in a git repository first, so that it is as its manifest records
// it.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readOnly takes the write permission from the file at the slash-separated
// path p below dir, and gives it back when the test ends, so that the
// temporary folder can be removed.
func readOnly(t *testing.T, dir, p string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(p))
	if err := os.Chmod(full, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(full, 0o644) })
	if f, err := os.OpenFile(full, os.O_WRONLY, 0); err == nil {
		f.Close()
		t.Skipf("%s stays writable without permission (running as root?)", p)
	}
}

// rewritten asserts that every file of want is as the run left it, and
// manifest.sum what `manifest` writes for want.
func (o rehomeOutcome) rewritten(t *testing.T, rule string, want tree) {
	t.Helper()
	for p, content := range want {
		o.holds(t, rule, p, content)
	}
	o.holds(t, rule, manifestPath, manifestWritten(t, want))
}

// mentionOn gives the mention of the mentions tree on the line of the file p
// that holds l, as the plan lists it.
func mentionOn(p, l string) planMention {
	line := mentionsTree().line(p, l)
	for _, m := range mentionsOfTree() {
		if m.File == p && m.Line == line {
			return m
		}
	}
	panic("mentionOn: no mention on " + l)
}

// listedOn gives the mentions the plan lists on the line of the file p that
// holds l in tr.
func (o rehomeOutcome) listedOn(tr tree, p, l string) []planMention {
	line := tr.line(p, l)
	var out []planMention
	for _, m := range o.plan.Mentions {
		if m.File == p && m.Line == line {
			out = append(out, m)
		}
	}
	return out
}

func TestUpkeepRehomeLint(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/, its tree without findings", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.planned(t, mainRehomeTree())
	})
	t.Run("row2 repo/, its tree with an ask only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, undecidedTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.planned(t, undecidedTree())
	})
	t.Run("row3 repo/, its tree with an error, the error on stderr", func(t *testing.T) {
		t.Parallel()
		// notes.md is no file of the layout and no element: lint reports
		// an error, while the tree is still as the manifest records it.
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with("proviso/notes.md", "# Notes\n"))
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		text := execLint(t, r.dir)
		if text.exit != 1 || strings.TrimSpace(text.stdout) == "" {
			t.Fatalf("want lint to report the error, got exit status %d\n%s%s", text.exit, text.stdout, text.stderr)
		}
		norm := func(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n")) }
		if !strings.Contains(norm(o.stderr), norm(text.stdout)) {
			t.Errorf("upkeep.rehome.lint: want stderr to hold what lint writes as text\n%s\ngot\n%s", text.stdout, o.stderr)
		}
	})
	t.Run("row4 repo/, its tree with an unreadable file", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with("proviso/capabilities/cart/features/bad.md", "# <a id=\"cart.bad\"></a> Bad \xff\xfe\n\n- Contact: Dana Reyes\n"))
		r.rehome("--write", "--json").exits(t, 2)
	})
	t.Run("row5 other/, with no spec root up to the file system's root", func(t *testing.T) {
		t.Parallel()
		dir := anywhere(t)
		rehomeIn(t, dir, dir, "", nil, "--write", "--json").exits(t, 2)
	})
}

func TestUpkeepRehomeBaseline(t *testing.T) {
	t.Parallel()
	t.Run("row1 repo/proviso/, its tree as manifest.sum records it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
	t.Run("row2 repo/proviso/, a paragraph reflowed since manifest.sum was written", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		tr := mainRehomeTree().edit(capPath, "Paying once for many goods is easier.", "Paying once for many\ngoods   is easier.")
		r.set(tr)
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
	// The refused rows run with --write, so that a run that went on would
	// have written.
	for _, c := range []struct {
		name string
		tr   func() tree
	}{
		{"row4 repo/proviso/, a rule's Then step edited since manifest.sum was written", func() tree {
			return mainRehomeTree().edit(featurePath, "Then      the cart holds <quantity> more of <product>", "Then      the cart holds <quantity> more pieces of <product>")
		}},
		{"row5 repo/proviso/, a rule added since manifest.sum was written", func() tree {
			return mainRehomeTree().appendTo(featurePath, extraRule)
		}},
		{"row6 repo/proviso/, a rule removed since manifest.sum was written", func() tree {
			empty, _ := elementText(mainRehomeTree()[featurePath], "cart.add.empty")
			return mainRehomeTree().edit(featurePath, empty, "")
		}},
		{"row7 repo/proviso/, holding .coherence.json", func() tree {
			return mainRehomeTree().with(coherencePath, "{}\n")
		}},
		{"row8 repo/proviso/, holding .amendment.md", func() tree {
			return mainRehomeTree().with(notePath, "- refunds\n")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, mainRehomeTree())
			r.set(c.tr())
			r.rehome("--write", "--json").exits(t, 2)
		})
	}
	t.Run("row3 repo/proviso/, without manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		remove(t, r.dir, manifestPath)
		r.rehome("--write", "--json").exits(t, 2)
	})
	t.Run("derived: .amendment.md a folder", func(t *testing.T) {
		t.Parallel()
		// Anything of that name is the note.
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with(notePath+"/note.md", "- refunds\n"))
		r.rehome("--write", "--json").exits(t, 2)
	})
}

func TestUpkeepRehomeHome(t *testing.T) {
	t.Parallel()
	r, _ := accepted(t, mainRehomeTree())
	o := r.rehome("--json")
	o.exits(t, 0)
	o.planned(t, mainRehomeTree())
	// renamed checks that the plan renames from to to; stays that it plans
	// no rename of id.
	renamed := func(t *testing.T, from, to string) {
		t.Helper()
		for _, rn := range o.plan.Renames {
			if rn.From == from {
				if rn.To != to {
					t.Errorf("upkeep.rehome.home: want %s renamed to %s, got %s", from, to, rn.To)
				}
				return
			}
		}
		t.Errorf("upkeep.rehome.home: want %s renamed to %s, got the renames %v", from, to, o.plan.Renames)
	}
	stays := func(t *testing.T, id string) {
		t.Helper()
		for _, rn := range o.plan.Renames {
			if rn.From == id {
				t.Errorf("upkeep.rehome.home: want %s to stay, got it renamed to %s", id, rn.To)
			}
		}
	}
	t.Run("row1 coherence.impact.events, a rule of the capability coherence", func(t *testing.T) {
		t.Parallel()
		stays(t, "cart.add.limit")
		stays(t, "cart.add.show")
	})
	t.Run("row2 baseline.impact.events, a rule of the capability coherence", func(t *testing.T) {
		t.Parallel()
		renamed(t, "billing.x.r", "cart.x.r")
	})
	t.Run("row3 baseline.part, a glossary entry of the system", func(t *testing.T) {
		t.Parallel()
		renamed(t, "cart.item", "system.item")
	})
	t.Run("row4 lint.cli.lint.exit-status, a member in contracts/cli.md of the system", func(t *testing.T) {
		t.Parallel()
		renamed(t, "cart.cli.lint.exit-status", "system.cli.lint.exit-status")
		renamed(t, "cart.cli.lint", "system.cli.lint")
	})
	t.Run("row5 coherence, the capability coherence", func(t *testing.T) {
		t.Parallel()
		stays(t, "cart")
		stays(t, "pay")
	})
	t.Run("row6 system, the system", func(t *testing.T) {
		t.Parallel()
		stays(t, "system")
		stays(t, "system.break")
	})
	t.Run("row7 checks.run, in the folder capabilities/coherence/, whose capability.md is identified as checks", func(t *testing.T) {
		t.Parallel()
		stays(t, "pay.card")
		// Named after the folder, not the capability: renamed.
		renamed(t, "billing.refund", "pay.refund")
	})
	t.Run("derived: a feature moved into another capability's folder", func(t *testing.T) {
		t.Parallel()
		renamed(t, "billing.x", "cart.x")
	})
	t.Run("derived: a reference of the system", func(t *testing.T) {
		t.Parallel()
		renamed(t, "cart.ref.commonmark", "system.ref.commonmark")
	})
}

func TestUpkeepRehomeCollision(t *testing.T) {
	t.Parallel()
	t.Run("row1 baseline.impact.events to coherence.impact.events, the tree holding no coherence.impact.events", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.collisions(t, nil)
		o.renames(t, mainRenames)
	})
	t.Run("row2 none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, nil)
		o.collisions(t, nil)
	})
	// The colliding rows run with --write, so that nothing written under
	// the spec root shows (rehomeIn).
	t.Run("row3 baseline.x to coherence.x, the tree holding a rule coherence.x", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		o.renames(t, []planRename{{"billing.x.r", "cart.x.r"}})
		o.collisions(t, []planCollision{{"cart.x.r", []string{"billing.x.r"}}})
	})
	t.Run("row4 baseline.x to coherence.x, while coherence.x of another capability is renamed to install.x", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, chainRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		o.renames(t, []planRename{{"billing.widget", "cart.widget"}, {"cart.widget", "stock.widget"}})
		o.collisions(t, []planCollision{{"cart.widget", []string{"billing.widget"}}})
	})
	t.Run("row5 baseline.x and lint.x, both to coherence.x", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, order2RehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		o.planned(t, order2RehomeTree())
		for _, c := range o.plan.Collisions {
			if c.Identifier == "cart.gizmo" {
				if strings.Join(c.From, " ") != "billing.gizmo pay.gizmo" {
					t.Errorf("upkeep.rehome.collision: want the collision on cart.gizmo to name billing.gizmo and pay.gizmo, got %v", c.From)
				}
				return
			}
		}
		t.Errorf("upkeep.rehome.collision: want a collision on cart.gizmo, got %v", o.plan.Collisions)
	})
}

func TestUpkeepRehomeWrite(t *testing.T) {
	t.Parallel()
	const rule = "upkeep.rehome.write"
	t.Run("rows 1, 2 and 6", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
		want := mainRehomed()
		t.Run("row1 baseline.impact.events to coherence.impact.events, its anchor in impact.md and a hand-over to it in verify.md", func(t *testing.T) {
			t.Parallel()
			o.holds(t, rule, xPath, want[xPath])
			o.holds(t, rule, featurePath, want[featurePath])
			o.holds(t, rule, manifestPath, manifestWritten(t, want))
		})
		t.Run("row2 the same, with baseline.impact.events also in an Examples cell of verify.md", func(t *testing.T) {
			t.Parallel()
			got, _ := o.file(featurePath)
			if !strings.Contains(got, "| billing.x.r | yes | shown |") {
				t.Errorf("%s: want the Examples cell to hold billing.x.r as it was, got\n%s", rule, got)
			}
			if !strings.Contains(got, "Otherwise → cart.x.r") {
				t.Errorf("%s: want the hand-over renamed to cart.x.r, got\n%s", rule, got)
			}
		})
		t.Run("row6 baseline.dependencies to coherence.dependencies, tagging a question in questions.md", func(t *testing.T) {
			t.Parallel()
			o.holds(t, rule, cartQuestions, want[cartQuestions])
		})
		t.Run("derived: every file of the spec root as the rewritten tree holds it", func(t *testing.T) {
			t.Parallel()
			for p, content := range want {
				o.holds(t, rule, p, content)
			}
		})
	})
	// Rows 3 to 5 run on the mentions tree: billing.x.r to cart.x.r, its
	// anchor in x.md and a hand-over to it in add.md, which also holds it as
	// a code span in Examples cells.
	t.Run("row3 the same, run with --mentions naming that cell's line", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		o := r.rehomePicking(mentionList(mentionOn(featurePath, mentionLineA)), "--write", "--mentions", "--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"billing.x.r", "cart.x.r"}})
		o.rewritten(t, rule, mentionsRehomed(mentionLineA))
	})
	t.Run("row4 the same, the cell's line holding baseline.impact.events twice, run with --mentions naming that line", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		o := r.rehomePicking(mentionList(mentionOn(featurePath, mentionLineB)), "--write", "--mentions", "--json")
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed(mentionLineB))
		got, _ := o.file(featurePath)
		if !strings.Contains(got, md("| ´cart.x.r´ again | yes | shown | as ´cart.x.r´ twice |")) {
			t.Errorf("%s: want both code spans of the line renamed, got\n%s", rule, got)
		}
	})
	t.Run("row5 the same, run with --mentions naming no mention", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		o := r.rehomePicking(mentionList(), "--write", "--mentions", "--json")
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed())
	})
	t.Run("derived: a file that holds only a mention picked, read-only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		readOnly(t, r.dir, cartQuestions)
		o := r.rehomePicking(mentionList(mentionOn(cartQuestions, mentionLineG)), "--write", "--mentions", "--json")
		o.exits(t, 2)
		o.names(t, "questions.md")
	})
	t.Run("derived: a file that holds only a mention not picked, read-only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		readOnly(t, r.dir, cartQuestions)
		o := r.rehomePicking(mentionList(mentionOn(featurePath, mentionLineA)), "--write", "--mentions", "--json")
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed(mentionLineA))
	})
	t.Run("row7 none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, nil)
		o.unchanged(t, rule, manifestPath)
		checkWrites(t, rule, o.before, o.after, "")
	})
	t.Run("row8 baseline.impact.events to coherence.impact.events, impact.md read-only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		readOnly(t, r.dir, xPath)
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		o.names(t, "x.md")
	})
	t.Run("row9 the same, manifest.sum read-only", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		readOnly(t, r.dir, manifestPath)
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		o.names(t, "manifest.sum")
	})
	t.Run("derived: one rename, the manifest of the rewritten tree", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, oneRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		want := oneRehomed()
		o.holds(t, rule, "proviso/glossary.md", want["proviso/glossary.md"])
		o.holds(t, rule, manifestPath, manifestWritten(t, want))
	})
	t.Run("derived: the element diff is empty after the run", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		r.rehome("--write", "--json").exits(t, 0)
		r.diff().only(t)
	})
}

func TestUpkeepRehomeOutput(t *testing.T) {
	t.Parallel()
	t.Run("row1 one, none, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, oneRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		o.collisions(t, nil)
	})
	t.Run("row2 one, its identifier in two Examples cells, none, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, twoCellsTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		o.mentions(t, twoCellsMentions())
		o.collisions(t, nil)
	})
	t.Run("row3 none, none, --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome("--json")
		o.exits(t, 0)
		// An empty list, not an empty output: parsePlan requires both
		// lists as arrays.
		if len(o.plan.Renames) != 0 || len(o.plan.Collisions) != 0 {
			t.Errorf("upkeep.rehome.output: want no renames and no collisions, got %v", o.plan)
		}
	})
	t.Run("row4 one, none, --write --json", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, oneRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		o.collisions(t, nil)
	})
	t.Run("derived: one, its identifier in two Examples cells, none, --write --json, the mentions listed though none is picked", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, twoCellsTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		o.mentions(t, twoCellsMentions())
	})
	t.Run("row5 one, its identifier in two Examples cells, none, --write --mentions --json, naming one of them", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, twoCellsTree())
		o := r.rehomePicking(mentionList(twoCellsMentions()[0]), "--write", "--mentions", "--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"cart.item", "system.item"}})
		// The plan, not what was picked: both mentions, as the tree held
		// them before the run.
		o.mentions(t, twoCellsMentions())
		o.collisions(t, nil)
		o.rewritten(t, "upkeep.rehome.output", twoCellsRehomed())
	})
	t.Run("row6 two, one colliding, one, none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, chainRehomeTree())
		o := r.rehome()
		o.exits(t, 1)
		if strings.TrimSpace(o.stdout) == "" {
			t.Fatalf("upkeep.rehome.output: want the renames and the collision as text, got nothing")
		}
		if strings.HasPrefix(strings.TrimSpace(o.stdout), "{") {
			t.Errorf("upkeep.rehome.output: want text without --json, got:\n%s", o.stdout)
		}
		for _, id := range []string{"billing.widget", "cart.widget", "stock.widget"} {
			if !strings.Contains(o.stdout, id) {
				t.Errorf("upkeep.rehome.output: want the text to name %s, got:\n%s", id, o.stdout)
			}
		}
	})
	t.Run("row7 none, none, none", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, baseline())
		o := r.rehome()
		o.exits(t, 0)
		if o.stdout != "" {
			t.Errorf("upkeep.rehome.output: nothing to say, but stdout holds %q", o.stdout)
		}
	})
}

func TestUpkeepRehomeOrder(t *testing.T) {
	t.Parallel()
	t.Run("row1 lint.keyword to system.keyword, baseline.part to system.part", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, order1RehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		o.renames(t, []planRename{{"billing.part", "system.part"}, {"stock.keyword", "system.keyword"}})
	})
	t.Run("row2 lint.x and baseline.x, both to coherence.x, and lint.y to coherence.y, taken", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, order2RehomeTree())
		o := r.rehome("--json")
		o.exits(t, 1)
		o.collisions(t, []planCollision{
			{"cart.gizmo", []string{"billing.gizmo", "pay.gizmo"}},
			{"cart.widget", []string{"pay.widget"}},
		})
	})
	t.Run("row3 lint.keyword to system.keyword, mentioned in rules.md lines 9 and 3 and in diff.md line 40", func(t *testing.T) {
		t.Parallel()
		// stock.keyword: lint.keyword, in cart/capability.md line 9 and in
		// stock/features/shelf.md lines 7 and 12; billing.part on line 12
		// as well, after it.
		tr := order3RehomeTree()
		r, _ := accepted(t, tr)
		o := r.rehome("--json")
		o.exits(t, 0)
		o.mentions(t, []planMention{
			mention("stock.keyword", capPath, 9, 1, "cart"),
			mention("stock.keyword", shelfPath, 7, 1, "stock.shelf"),
			mention("billing.part", shelfPath, 12, 1, "stock.shelf"),
			mention("stock.keyword", shelfPath, 12, 1, "stock.shelf"),
		})
	})
}

func TestUpkeepRehomeExit(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		tr   func() tree
		want int
	}{
		{"row1 none", mainRehomeTree, 0},
		{"row2 none, and no renames", baseline, 0},
		{"row3 one", takenRehomeTree, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, c.tr())
			r.rehome("--json").exits(t, c.want)
		})
	}
	t.Run("derived: the text form exits as the JSON form", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		r.rehome().exits(t, 1)
	})
}

// TestUpkeepRehomeWrites checks the invariant upkeep.rehome.writes on runs
// that end every way; rehomeIn checks it on every run.
func TestUpkeepRehomeWrites(t *testing.T) {
	t.Parallel()
	t.Run("derived: without --write, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write and exit status 0, only the renamed identifiers and manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 0)
		checkRenamedOnly(t, o.before, o.after, "proviso", mainRenames)
		// The git repository is as it was.
		for p, v := range o.before {
			if strings.HasPrefix(p, ".git/") && o.after[p] != v {
				t.Errorf("upkeep.rehome.writes: %s of the git repository changed", p)
			}
		}
	})
	t.Run("derived: with --write and exit status 1, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		o := r.rehome("--write", "--json")
		o.exits(t, 1)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write and exit status 2, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		r.set(mainRehomeTree().with(coherencePath, "{}\n"))
		o := r.rehome("--write", "--json")
		o.exits(t, 2)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write --mentions and exit status 0, only the renamed identifiers, the mentions picked and manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		o := r.rehomePicking(mentionList(mentionOn(featurePath, mentionLineB), mentionOn(cartQuestions, mentionLineG)), "--write", "--mentions", "--json")
		o.exits(t, 0)
		want := mentionsRehomed(mentionLineB, mentionLineG)
		for p, v := range o.after {
			if p == manifestPath || !strings.HasPrefix(p, "proviso/") || !strings.HasPrefix(v, "file: ") {
				continue
			}
			if content, ok := want[p]; !ok || "file: "+content != v {
				t.Errorf("upkeep.rehome.writes: %s is not as the renames and the mentions picked leave it", p)
			}
		}
		for p, v := range o.before {
			if strings.HasPrefix(p, ".git/") && o.after[p] != v {
				t.Errorf("upkeep.rehome.writes: %s of the git repository changed", p)
			}
		}
	})
	t.Run("derived: with --write --mentions and exit status 1, nothing changes", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, takenRehomeTree())
		o := r.rehomePicking(mentionList(), "--write", "--mentions", "--json")
		o.exits(t, 1)
		checkWrites(t, "upkeep.rehome.writes", o.before, o.after, "")
	})
	t.Run("derived: with --write, in the text form, only the renamed identifiers and manifest.sum", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		o := r.rehome("--write")
		o.exits(t, 0)
		checkRenamedOnly(t, o.before, o.after, "proviso", mainRenames)
	})
}

// TestUpkeepRehomeReferences checks the parameter rehome_references: each
// place it lists is renamed, and an identifier anywhere else stays.
func TestUpkeepRehomeReferences(t *testing.T) {
	t.Parallel()
	r, _ := accepted(t, mainRehomeTree())
	o := r.rehome("--write", "--json")
	o.exits(t, 0)
	get := func(p string) string { s, _ := o.file(p); return s }
	holds := func(t *testing.T, p, sub string) {
		t.Helper()
		if !strings.Contains(get(p), md(sub)) {
			t.Errorf("upkeep.rehome-references: want %s to hold %q, got\n%s", p, md(sub), get(p))
		}
	}
	lacks := func(t *testing.T, p, sub string) {
		t.Helper()
		if strings.Contains(get(p), md(sub)) {
			t.Errorf("upkeep.rehome-references: want %s not to hold %q, got\n%s", p, md(sub), get(p))
		}
	}
	t.Run("row1 an anchor", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, `## <a id="cart.x.r"></a> Rule: Under test`)
		holds(t, payRefundPath, `# <a id="pay.refund"></a> Refunds`)
		lacks(t, payRefundPath, "billing.refund")
	})
	t.Run("row2 a hand-over", func(t *testing.T) {
		t.Parallel()
		holds(t, featurePath, "Otherwise → cart.x.r\n")
		lacks(t, featurePath, "Otherwise → billing.x.r")
	})
	t.Run("row3 an inline link or an autolink, outside code spans and fenced blocks", func(t *testing.T) {
		t.Parallel()
		holds(t, "proviso/glossary.md", "(references.md#system.ref.commonmark)")
		holds(t, xPath, "see [the rule](x.md#cart.x.r), not")
		holds(t, xPath, "<https://example.com/x.md#cart.x.r>")
	})
	t.Run("row4 a question in a questions.md", func(t *testing.T) {
		t.Parallel()
		holds(t, cartQuestions, "- [cart.x.r] Should the shop show pears?")
	})
	t.Run("derived: an identifier in plain text stays", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "and billing.x.r in\nplain text.")
		holds(t, cartQuestions, "- Who decides on [billing.x.r] later?")
	})
	t.Run("row5 a link reference definition, outside code spans and fenced blocks", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "\n[rule]: x.md#cart.x.r\n\n")
		// The link that uses it holds no destination of its own.
		holds(t, xPath, "as the [defined rule][rule],")
	})
	t.Run("derived: an identifier in an Examples cell stays", func(t *testing.T) {
		t.Parallel()
		holds(t, featurePath, "| billing.x.r | yes | shown |")
	})
	t.Run("derived: an inline link or an autolink in an Examples cell stays", func(t *testing.T) {
		t.Parallel()
		holds(t, featurePath, "| a code may look like [an identifier](x.md#billing.x.r) |")
		holds(t, featurePath, "| handed over | <https://example.com/x.md#billing.x.r> |")
	})
	t.Run("derived: an identifier in a code span stays", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "´billing.x.r´ nor ´[the rule](x.md#billing.x.r)´")
	})
	t.Run("derived: an identifier in an image stays", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "![a picture](x.md#billing.x.r)")
	})
	t.Run("derived: an identifier in a fenced block stays, a link reference definition too", func(t *testing.T) {
		t.Parallel()
		holds(t, xPath, "´´´text\n<a id=\"billing.x.r\"></a> [the rule](x.md#billing.x.r)\n[rule]: x.md#billing.x.r\n´´´")
	})
}

func TestUpkeepRehomeMentions(t *testing.T) {
	t.Parallel()
	const rule = "upkeep.rehome.mentions"
	tr := mentionsTree()
	r, _ := accepted(t, tr)
	o := r.rehome("--json")
	o.exits(t, 0)
	o.renames(t, []planRename{{"billing.x.r", "cart.x.r"}})
	// listed asserts that the line of p holding l is listed once, as want.
	listed := func(t *testing.T, p, l string, want planMention) {
		t.Helper()
		got := o.listedOn(tr, p, l)
		if len(got) != 1 || got[0].String() != want.String() {
			t.Errorf("%s: want the line %d of %s listed once, as %v, got %v", rule, tr.line(p, l), p, want, got)
		}
	}
	notListed := func(t *testing.T, p, l string) {
		t.Helper()
		if got := o.listedOn(tr, p, l); len(got) != 0 {
			t.Errorf("%s: want nothing listed for the line %d of %s, got %v", rule, tr.line(p, l), p, got)
		}
	}
	t.Run("row1 baseline.impact.events, standing as a code span in an Examples cell of verify.md", func(t *testing.T) {
		t.Parallel()
		listed(t, featurePath, mentionLineA, mention("billing.x.r", featurePath, tr.line(featurePath, mentionLineA), 1, "cart.add.show"))
	})
	t.Run("row2 baseline.impact.events, standing twice as a code span in one Examples row", func(t *testing.T) {
		t.Parallel()
		listed(t, featurePath, mentionLineB, mention("billing.x.r", featurePath, tr.line(featurePath, mentionLineB), 2, "cart.add.show"))
	})
	t.Run("row3 baseline.impact.events, standing as a code span in the Rationale of a feature", func(t *testing.T) {
		t.Parallel()
		listed(t, xPath, mentionLineF, mention("billing.x.r", xPath, tr.line(xPath, mentionLineF), 1, "cart.x"))
	})
	t.Run("row4 baseline.impact.events, standing only within a longer code span in an Examples cell", func(t *testing.T) {
		t.Parallel()
		notListed(t, featurePath, mentionLineC)
	})
	t.Run("row5 baseline.impact.events, standing only in a fenced block of a contract", func(t *testing.T) {
		t.Parallel()
		for _, m := range o.plan.Mentions {
			if m.File == sysContractPath {
				t.Errorf("%s: want nothing listed in %s, got %v", rule, sysContractPath, m)
			}
		}
	})
	t.Run("row6 baseline.impact.events, standing only in plain text, without backticks", func(t *testing.T) {
		t.Parallel()
		notListed(t, featurePath, mentionLinePlain)
		notListed(t, xPath, mentionLineXText)
	})
	t.Run("row7 lint.run, not planned for a rename, standing as a code span in an Examples cell", func(t *testing.T) {
		t.Parallel()
		notListed(t, featurePath, mentionLineD)
		for _, m := range o.plan.Mentions {
			if m.Identifier == "lint.run" {
				t.Errorf("%s: want no mention of lint.run, got %v", rule, m)
			}
		}
	})
	t.Run("row8 baseline.impact.events, standing as a code span in the text of a question in a questions.md", func(t *testing.T) {
		t.Parallel()
		listed(t, cartQuestions, mentionLineG, mention("billing.x.r", cartQuestions, tr.line(cartQuestions, mentionLineG), 1, ""))
	})
	t.Run("derived: exactly the mentions of the tree", func(t *testing.T) {
		t.Parallel()
		o.mentions(t, mentionsOfTree())
	})
	t.Run("derived: a mention in the renamed element names its identifier before the run", func(t *testing.T) {
		t.Parallel()
		listed(t, xPath, mentionLineH, mention("billing.x.r", xPath, tr.line(xPath, mentionLineH), 1, "billing.x.r"))
	})
	t.Run("derived: with --write, no mention is renamed that --mentions does not name", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		w := r.rehome("--write", "--json")
		w.exits(t, 0)
		w.mentions(t, mentionsOfTree())
		w.rewritten(t, rule, mentionsRehomed())
	})
	t.Run("derived: with --write --mentions, only the mentions it names are renamed", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		w := r.rehomePicking(mentionList(mentionOn(xPath, mentionLineF), mentionOn(xPath, mentionLineH)), "--write", "--mentions", "--json")
		w.exits(t, 0)
		w.rewritten(t, rule, mentionsRehomed(mentionLineF, mentionLineH))
	})
	for name, tr := range map[string]func() tree{
		"main": mainRehomeTree, "two cells": twoCellsTree, "order3": order3RehomeTree, "chain": chainRehomeTree,
	} {
		t.Run("derived: the mentions of the "+name+" tree", func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, tr())
			o := r.rehome("--json")
			if o.exit == 2 {
				t.Fatalf("want rehome to plan, got exit status 2\n%s", o.stderr)
			}
			o.mentioned(t, tr())
		})
	}
}

func TestUpkeepRehomePicked(t *testing.T) {
	t.Parallel()
	const rule = "upkeep.rehome.picked"
	pick := func(t *testing.T, input string) rehomeOutcome {
		t.Helper()
		r, _ := accepted(t, mentionsTree())
		return r.rehomePicking(input, "--write", "--mentions", "--json")
	}
	a := mentionOn(featurePath, mentionLineA)
	t.Run("row1 a mention_list naming one mention the plan lists", func(t *testing.T) {
		t.Parallel()
		o := pick(t, mentionList(a))
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed(mentionLineA))
	})
	t.Run("row2 a mention_list with no mentions", func(t *testing.T) {
		t.Parallel()
		o := pick(t, mentionList())
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed())
	})
	t.Run("row3 a mention_list naming one mention twice", func(t *testing.T) {
		t.Parallel()
		o := pick(t, mentionList(a, a))
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed(mentionLineA))
	})
	// The refused rows: rehomeWith checks that nothing goes to stdout,
	// something to stderr, and nothing is written.
	t.Run("row4 {", func(t *testing.T) {
		t.Parallel()
		pick(t, "{").exits(t, 2)
	})
	t.Run("row5 a mention_list naming a line where the identifier stands in plain text only", func(t *testing.T) {
		t.Parallel()
		tr := mentionsTree()
		pick(t, mentionList(mention("billing.x.r", xPath, tr.line(xPath, mentionLineXText), 1, ""))).exits(t, 2)
	})
	t.Run("row6 a mention_list naming a mention of lint.run, which the plan does not rename", func(t *testing.T) {
		t.Parallel()
		tr := mentionsTree()
		pick(t, mentionList(mention("lint.run", featurePath, tr.line(featurePath, mentionLineD), 1, ""))).exits(t, 2)
	})
	t.Run("derived: a mention_list naming one mention the plan lists and one it does not", func(t *testing.T) {
		t.Parallel()
		tr := mentionsTree()
		pick(t, mentionList(a, mention("billing.x.r", featurePath, tr.line(featurePath, mentionLinePlain), 1, ""))).exits(t, 2)
	})
	t.Run("derived: a mention_list naming every mention the plan lists", func(t *testing.T) {
		t.Parallel()
		o := pick(t, mentionList(mentionsOfTree()...))
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed(mentionLineA, mentionLineB, mentionLineF, mentionLineG, mentionLineH))
	})
}

// TestUpkeepIdentifierMention checks the term identifier mention: a code
// span whose whole content is an identifier, outside rule blocks and fenced
// blocks, through the plan of the mentions tree, against the oracle
// mentionsOf.
func TestUpkeepIdentifierMention(t *testing.T) {
	t.Parallel()
	r, _ := accepted(t, mentionsTree())
	o := r.rehome("--json")
	o.exits(t, 0)
	t.Run("derived: as the oracle gives them", func(t *testing.T) {
		t.Parallel()
		o.mentioned(t, mentionsTree())
	})
}

// TestUpkeepIdentifierHome checks the term identifier home through the
// plans of trees whose elements stand at every kind of place: the oracle
// identifierHome gives each element's home, and every element whose first
// segment names another place is planned to be renamed.
func TestUpkeepIdentifierHome(t *testing.T) {
	t.Parallel()
	for name, tr := range map[string]func() tree{
		"main": mainRehomeTree, "one": oneRehomeTree, "order1": order1RehomeTree, "chain": chainRehomeTree,
	} {
		t.Run("derived: "+name, func(t *testing.T) {
			t.Parallel()
			r, _ := accepted(t, tr())
			o := r.rehome("--json")
			if o.exit == 2 {
				t.Fatalf("want rehome to plan, got exit status 2\n%s", o.stderr)
			}
			o.planned(t, tr())
		})
	}
}

// TestSystemCliRehome checks the contract system.cli.rehome: it runs
// anywhere inside the system's repository.
func TestSystemCliRehome(t *testing.T) {
	t.Parallel()
	t.Run("derived: run in a folder of the spec tree, it plans for the spec root", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree())
		cwd := filepath.Join(r.dir, "proviso", "capabilities", "cart")
		o := rehomeIn(t, r.dir, cwd, r.root(), nil, "--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
	t.Run("derived: run in a folder beside the spec root, it plans for the spec root", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mainRehomeTree().with("src/main.go", "package main\n"))
		o := rehomeIn(t, r.dir, filepath.Join(r.dir, "src"), r.root(), nil, "--json")
		o.exits(t, 0)
		o.renames(t, mainRenames)
	})
}

// TestSystemCliRehomeMentions checks the contract system.cli.rehome.mentions:
// with --write, without --mentions no mention is renamed.
func TestSystemCliRehomeMentions(t *testing.T) {
	t.Parallel()
	t.Run("derived: without --mentions, a mention_list on stdin is not read", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		o := r.rehomePicking(mentionList(mentionOn(featurePath, mentionLineA)), "--write", "--json")
		o.exits(t, 0)
		o.rewritten(t, "system.cli.rehome.mentions", mentionsRehomed())
	})
}

// TestSystemCliMentionList checks the contract system.cli.mention-list: a
// mention is named by its identifier, file and line, other members are not
// read, and an input that is no mention_list is refused.
func TestSystemCliMentionList(t *testing.T) {
	t.Parallel()
	const rule = "system.cli.mention-list"
	t.Run("derived: the plan's own list, with count and element", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		plan := r.rehome("--json")
		plan.exits(t, 0)
		// The plan's mentions without A, as the plan wrote them, with a
		// count that does not hold, which is not read.
		var keep []planMention
		for _, m := range plan.plan.Mentions {
			if m.String() != mentionOn(featurePath, mentionLineA).String() {
				m.Count = 7
				keep = append(keep, m)
			}
		}
		b, _ := json.Marshal(struct {
			Mentions []planMention `json:"mentions"`
			Extra    string        `json:"note"`
		}{keep, "not read"})
		o := r.rehomePicking(string(b)+"\n", "--write", "--mentions", "--json")
		o.exits(t, 0)
		o.rewritten(t, rule, mentionsRehomed(mentionLineB, mentionLineF, mentionLineG, mentionLineH))
	})
	t.Run("derived: an object without mentions", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		r.rehomePicking("{}\n", "--write", "--mentions", "--json").exits(t, 2)
	})
	t.Run("derived: a mention without its line", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		input := `{"mentions": [{"identifier": "billing.x.r", "file": "` + featurePath + `"}]}` + "\n"
		r.rehomePicking(input, "--write", "--mentions", "--json").exits(t, 2)
	})
}

// TestSystemCliRehomePlan checks the contract system.cli.rehome-plan, whose
// shape rehomeIn checks on every run: element is null for a mention no
// element holds.
func TestSystemCliRehomePlan(t *testing.T) {
	t.Parallel()
	t.Run("derived: element null in a questions.md", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, mentionsTree())
		o := r.rehome("--json")
		o.exits(t, 0)
		if !strings.Contains(strings.ReplaceAll(o.stdout, " ", ""), `"element":null`) {
			t.Errorf("system.cli.rehome-plan: want element null for the mention in questions.md, got\n%s", o.stdout)
		}
	})
}
