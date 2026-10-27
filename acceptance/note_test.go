package acceptance

// The amendment note: the rule baseline.pack.note
// (proviso/capabilities/baseline/features/pack.md) and the contract
// system.amendment-note. The pack cases accept the pack tree (see
// pack_trees_test.go), amend it, run `proviso pack` without a note and then
// again with one, so that what the note adds can be told from everything
// else.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// noteName is the name of amendment_note in the spec root.
const noteName = ".amendment.md"

// notePath is amendment_note of the test trees, relative to the repository
// directory.
const notePath = "proviso/" + noteName

// writeNote writes content as the amendment note of the repository.
func (r *repo) writeNote(content string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.root(), noteName), []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *repo) hasNote() bool {
	_, err := os.Lstat(filepath.Join(r.root(), noteName))
	return err == nil
}

// withoutNote is the pack_text of o with the heading Not in this amendment
// and what stands under it taken out, line feeds as line endings.
func withoutNote(o packOutcome) string {
	lines := strings.Split(strings.ReplaceAll(o.text.stdout, "\r\n", "\n"), "\n")
	var out []string
	in := false
	for _, l := range lines {
		switch {
		case l == "# "+notInAmendment:
			in = true
			continue
		case in && (l == "# Frames" || l == "# Context"):
			in = false
		case in:
			continue
		}
		out = append(out, l)
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

// sameButNote asserts that the pack with differs from the pack without only
// in the note: the context_pack but for note, and pack_text but for the
// heading Not in this amendment and what stands under it.
func sameButNote(t *testing.T, without, with packOutcome) {
	t.Helper()
	var a, b map[string]any
	json.Unmarshal([]byte(without.json.stdout), &a)
	json.Unmarshal([]byte(with.json.stdout), &b)
	delete(a, "note")
	delete(b, "note")
	if !reflect.DeepEqual(a, b) {
		t.Errorf("want the context_pack the same but for note, got\n%s\n---\n%s", without.json.stdout, with.json.stdout)
	}
	if got, want := withoutNote(with), strings.TrimRight(strings.ReplaceAll(without.text.stdout, "\r\n", "\n"), "\n"); got != want {
		t.Errorf("want pack_text the same but for the note, got\n%s\n---\n%s", want, got)
	}
}

// noteTaken asserts that the pack holds want as the note, in both forms.
// The note is compared up to line feeds at its start and end (see the
// report).
func noteTaken(t *testing.T, o packOutcome, want string) {
	t.Helper()
	o.ran(t)
	want = strings.Trim(want, "\n")
	if o.note == nil || *o.note != want {
		t.Errorf("want pack_text to hold under %q\n%s\ngot\n%s", notInAmendment, want, deref(o.note))
	}
	if o.cp.Note == nil || strings.Trim(*o.cp.Note, "\n") != want {
		t.Errorf("want the note %q in the context_pack, got %s", want, deref(o.cp.Note))
	}
}

// nothingTaken asserts that the pack holds no note, in neither form.
func nothingTaken(t *testing.T, o packOutcome) {
	t.Helper()
	o.ran(t)
	if o.note != nil {
		t.Errorf("want no heading %q, got one holding %q", notInAmendment, *o.note)
	}
	if o.cp.Note != nil {
		t.Errorf("want note null in the context_pack, got %q", *o.cp.Note)
	}
}

// packNote accepts the pack tree, amends it with cs and runs pack, then
// writes note as the amendment note and runs pack again. It returns the
// pack without the note and the pack with it.
func packNote(t *testing.T, note string, cs ...change) (packOutcome, packOutcome) {
	t.Helper()
	r, _ := accepted(t, packTree())
	r.set(amended(packTree(), cs...))
	without := r.pack()
	without.ran(t)
	r.writeNote(note)
	with := r.pack()
	with.ran(t)
	return without, with
}

func TestBaselinePackNote(t *testing.T) {
	t.Parallel()
	t.Run(`row1 with .amendment.md holding "- refunds"`, func(t *testing.T) {
		t.Parallel()
		without, with := packNote(t, "- refunds\n", pCapThen)
		noteTaken(t, with, "- refunds")
		with.headingsAre(t, "System", "Amendment", notInAmendment, "Frames", "Context")
		sameButNote(t, without, with)
	})
	t.Run("row2 without .amendment.md", func(t *testing.T) {
		t.Parallel()
		o, _, _ := packOf(t, pCapThen)
		nothingTaken(t, o)
		o.headingsAre(t, "System", "Amendment", "Frames", "Context")
	})
	t.Run("row3 with an empty .amendment.md", func(t *testing.T) {
		t.Parallel()
		without, with := packNote(t, "", pCapThen)
		nothingTaken(t, with)
		sameButNote(t, without, with)
	})
	t.Run("row4 with .amendment.md holding only blank lines", func(t *testing.T) {
		t.Parallel()
		without, with := packNote(t, "\n\n\n", pCapThen)
		nothingTaken(t, with)
		sameButNote(t, without, with)
	})
	t.Run("derived: holding only spaces, tabs and line feeds", func(t *testing.T) {
		t.Parallel()
		without, with := packNote(t, "  \n\t\n \t \n", pCapThen)
		nothingTaken(t, with)
		sameButNote(t, without, with)
	})
	t.Run("derived: the text as the file holds it", func(t *testing.T) {
		t.Parallel()
		// Nested items, a blank line between items, backticks, a fence
		// and letters beyond ASCII: none of it is read as Markdown.
		note := "- refunds, as `billing.refund` would name them\n" +
			"  - and their receipts\n" +
			"\n" +
			"- gift cards für Kinder, Ä–Ö\n" +
			"- the format:\n" +
			"\n" +
			"  ```text\n" +
			"  # not a heading\n" +
			"  ```\n"
		without, with := packNote(t, note, pCapThen)
		noteTaken(t, with, note)
		sameButNote(t, without, with)
	})
	t.Run("derived: an empty element diff with a note", func(t *testing.T) {
		t.Parallel()
		without, with := packNote(t, "- refunds\n")
		noteTaken(t, with, "- refunds")
		with.headingsAre(t, "System", notInAmendment)
		sameButNote(t, without, with)
	})
	t.Run("derived: the note committed on the amendment's branch", func(t *testing.T) {
		t.Parallel()
		r, base := accepted(t, packTree())
		r.git("checkout", "-q", "-b", "amendment")
		r.set(amended(packTree(), pCapThen))
		r.writeNote("- refunds\n")
		r.commit("amend, with a note")
		o := r.pack()
		noteTaken(t, o, "- refunds")
		if deref(o.cp.Revision) != base || o.revision != base {
			t.Errorf("want the revision %s, got %s in context_pack and %q in pack_text", base, deref(o.cp.Revision), o.revision)
		}
	})
}

func TestSystemAmendmentNote(t *testing.T) {
	t.Parallel()
	// What would be an ask and a heading without an identifier if lint
	// read it.
	note := "- refunds?\n- ?\n\n# Refunds\n"
	t.Run("derived: lint skips it", func(t *testing.T) {
		t.Parallel()
		check(t, plainTree().with(notePath, note)).clean(t)
	})
	t.Run("derived: the manifest does not record it", func(t *testing.T) {
		t.Parallel()
		// That manifest removes it is checked for every run that exits
		// with 0 (see manifestIn).
		if got, want := manifestOf(t, plainTree().with(notePath, note)), manifestFor(plainParts()); got != want {
			t.Errorf("want the manifest of the tree without the note\n%s\ngot\n%s", want, got)
		}
	})
	t.Run("derived: the element diff does not list it", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.writeNote(note)
		r.diff().only(t)
	})
}

func TestBaselinePackNoteUnreadable(t *testing.T) {
	t.Parallel()
	// That a stopped pack writes to stderr and nothing to stdout is checked
	// for every run (see packIn); what stderr says is free.
	t.Run(`row1 a file holding "- refunds"`, func(t *testing.T) {
		t.Parallel()
		without, with := packNote(t, "- refunds\n", pCapThen)
		noteTaken(t, with, "- refunds")
		sameButNote(t, without, with)
	})
	t.Run("row2 a folder named .amendment.md", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		if err := os.MkdirAll(filepath.Join(r.root(), noteName, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r.root(), noteName, "sub", "item.md"), []byte("- refunds\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		r.pack().stopped(t)
	})
	t.Run("row3 a file holding bytes that are not UTF-8", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		r.writeNote("- refunds \xff\xfe\n")
		r.pack().stopped(t)
	})
	t.Run("derived: an empty folder named .amendment.md", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		if err := os.Mkdir(filepath.Join(r.root(), noteName), 0o755); err != nil {
			t.Fatal(err)
		}
		r.pack().stopped(t)
	})
	t.Run("derived: not UTF-8, beside an empty element diff", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.writeNote("\xc3\x28\n")
		r.pack().stopped(t)
	})
	t.Run("derived: a file pack may not read", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), pCapThen))
		r.writeNote("- refunds\n")
		undo, err := makeUnreadable(filepath.Join(r.root(), noteName))
		if err != nil {
			t.Skipf("cannot take the read permission here: %v", err)
		}
		defer undo()
		r.pack().stopped(t)
	})
}

// writeNoteFolder makes the amendment note of the repository a folder
// holding a file.
func (r *repo) writeNoteFolder() {
	r.t.Helper()
	dir := filepath.Join(r.root(), noteName)
	if err := os.Mkdir(dir, 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "item.md"), []byte("- refunds\n"), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// TestSystemAmendmentNote_Entry checks system.amendment-note's "anything of
// that name counts as the note" through verify; for pack, see
// TestBaselinePackNoteUnreadable, for manifest TestBaselineManifestRemove.
func TestSystemAmendmentNote_Entry(t *testing.T) {
	t.Parallel()
	t.Run("derived: a folder named .amendment.md beside an unchanged tree is left over", func(t *testing.T) {
		t.Parallel()
		// baseline.verify.leftover
		r, _ := accepted(t, impactTree())
		r.writeNoteFolder()
		r.verify().only(t, "leftover")
	})
	t.Run("derived: a folder named .amendment.md beside a changed tree, without a check", func(t *testing.T) {
		t.Parallel()
		// baseline.verify.leftover, then baseline.verify.missing
		r := amendedRepo(t, iLimitThen)
		r.writeNoteFolder()
		r.verify().only(t, "missing")
	})
	t.Run("derived: a folder named .amendment.md beside a covered tree", func(t *testing.T) {
		t.Parallel()
		r := covered(t, iLimitThen)
		r.writeNoteFolder()
		r.verify().only(t)
	})
}
