package acceptance

// Feature baseline.parts: proviso/capabilities/baseline/features/parts.md,
// observed through the manifest that `proviso manifest` writes.

import (
	"sort"
	"strings"
	"testing"
)

func TestBaselinePartsSplit(t *testing.T) {
	t.Parallel()
	m := manifestOf(t, richTree())
	limitRows := []string{"row " + rowKey("quantity", "1", "product", "apple"), "row " + rowKey("quantity", "11", "product", "apple")}
	cases := []struct {
		name, id string
		want     []string
	}{
		{"row1 a rule without a Rationale", "cart.add.limit",
			append([]string{"heading", "step When", "step Provided", "step Then", "step Otherwise"}, limitRows...)},
		{"row2 a feature with rules", "cart.add",
			[]string{"heading", "field Contact", "section Goal", "section Rationale"}},
		{"row3 a glossary entry", "cart.item",
			[]string{"heading", "field Definition", "field Contact", "field Status", "field Aliases"}},
		{"row4 a parameter with a table", "cart.sizes",
			[]string{"heading", "field Configurable", "field Rationale", "body"}},
		{"row5 a contract element with members", "cart.cli.lint",
			[]string{"heading", "body"}},
		{"derived: the member of a contract element", "cart.cli.lint.exit-status",
			[]string{"heading", "body"}},
		{"derived: a rule with a Given and a Rationale", "cart.x.open",
			[]string{"heading", "section Rationale", "step Given", "step When", "step Then", "row " + rowKey("product", "apple")}},
		{"derived: a rule without a guard", "cart.add.empty",
			[]string{"heading", "step When", "step Then", "row " + rowKey("cart", "the first cart")}},
		{"derived: an invariant", "cart.x.never",
			[]string{"heading", "section Rationale", "section Ranges over", "section Property"}},
		{"derived: the system", "system",
			[]string{"heading", "section Goal", "section Rationale", "section Scope", "section Out of scope", "section Audience", "section Design boundary"}},
		{"derived: a parameter with a value", "cart.max-quantity",
			[]string{"heading", "field Value", "field Unit", "field Configurable", "field Rationale"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := partsOf(m, c.id)
			want := append([]string(nil), c.want...)
			sort.Strings(want)
			if strings.Join(got, "|") != strings.Join(want, "|") {
				t.Errorf("want the parts of %s to be %q, got %q", c.id, want, got)
			}
		})
	}
	t.Run("derived: one line for every part of every element", func(t *testing.T) {
		t.Parallel()
		ids := map[string]bool{}
		for k := range manifestParts(m) {
			ids[k[0]] = true
		}
		for _, id := range []string{"system", "cart", "cart.item", "cart.max-quantity", "cart.sizes", "cart.add",
			"cart.add.limit", "cart.add.empty", "cart.cli.lint", "cart.cli.lint.exit-status", "cart.x", "cart.x.open", "cart.x.never"} {
			if !ids[id] {
				t.Errorf("want parts of %s in the manifest, got none", id)
			}
			delete(ids, id)
		}
		for id := range ids {
			t.Errorf("the manifest records %s, which is no element of the tree", id)
		}
	})
}

// fpOf writes the manifest of tr and returns the fingerprint of id's part.
func fpOf(t *testing.T, tr tree, id, part string) string {
	t.Helper()
	got, ok := manifestParts(manifestOf(t, tr))[[2]string{id, part}]
	if !ok {
		t.Fatalf("the manifest records no part %q of %s", part, id)
	}
	return got
}

func TestBaselinePartsFingerprint(t *testing.T) {
	t.Parallel()
	goal := "Buyers put goods into the cart."

	t.Run("row1 a section's body, and the same body wrapped", func(t *testing.T) {
		t.Parallel()
		a := fpOf(t, plainTree(), "cart.add", "section Goal")
		b := fpOf(t, plainTree().edit(featurePath, goal, "Buyers put goods\ninto the cart."), "cart.add", "section Goal")
		if a != b {
			t.Errorf("want the same fingerprint, got %s and %s", a, b)
		}
		if a != fp(goal) {
			t.Errorf("want the fingerprint %s of the body, got %s", fp(goal), a)
		}
	})
	t.Run("row2 the padding after a keyword", func(t *testing.T) {
		t.Parallel()
		a := fpOf(t, baseline(), "cart.add.limit", "step Then")
		b := fpOf(t, baseline().edit(featurePath, "Then      the cart holds", "Then the cart holds"), "cart.add.limit", "step Then")
		if a != b {
			t.Errorf("want the same fingerprint, got %s and %s", a, b)
		}
	})
	t.Run("row3 *a* and _a_", func(t *testing.T) {
		t.Parallel()
		a := fpOf(t, plainTree().edit(featurePath, goal, "Buyers put *some* goods into the cart."), "cart.add", "section Goal")
		b := fpOf(t, plainTree().edit(featurePath, goal, "Buyers put _some_ goods into the cart."), "cart.add", "section Goal")
		if a == b {
			t.Errorf("want different fingerprints, got %s for both", a)
		}
	})
	t.Run("row4 case counts", func(t *testing.T) {
		t.Parallel()
		a := fpOf(t, plainTree(), "cart.item", "field Contact")
		b := fpOf(t, plainTree().edit(glossaryPath, "- Contact: Bruno", "- Contact: bruno"), "cart.item", "field Contact")
		if a == b {
			t.Errorf("want different fingerprints, got %s for both", a)
		}
	})
	// system.section: in a rule, a section ends at its rule block.
	t.Run("derived: a rule's Rationale ends at its rule block", func(t *testing.T) {
		t.Parallel()
		if got, want := fpOf(t, richTree(), "cart.x.open", "section Rationale"), fp("Closed shops sell nothing."); got != want {
			t.Errorf("want the fingerprint %s of the Rationale's text alone, got %s", want, got)
		}
	})
	t.Run("derived: tabs and trailing spaces are whitespace", func(t *testing.T) {
		t.Parallel()
		a := fpOf(t, plainTree(), "cart.item", "field Contact")
		b := fpOf(t, plainTree().edit(glossaryPath, "- Contact: Bruno Baketarić", "- Contact: Bruno\tBaketarić   "), "cart.item", "field Contact")
		if a != b || a != fp("Bruno Baketarić") {
			t.Errorf("want the fingerprint %s for both, got %s and %s", fp("Bruno Baketarić"), a, b)
		}
	})
	t.Run("derived: a no-break space is not whitespace", func(t *testing.T) {
		t.Parallel()
		a := fpOf(t, plainTree(), "cart.item", "field Contact")
		b := fpOf(t, plainTree().edit(glossaryPath, "- Contact: Bruno Baketarić", "- Contact: Bruno Baketarić"), "cart.item", "field Contact")
		if a == b {
			t.Errorf("want different fingerprints, got %s for both", a)
		}
	})
	// The text of each part of a rule, as part_kinds gives it: the heading's
	// text after its anchor, a row's line, a step's lines with its And lines.
	t.Run("derived: the fingerprints of a rule's parts", func(t *testing.T) {
		t.Parallel()
		m := manifestParts(manifestOf(t, baseline()))
		get := func(part string) string { return m[[2]string{"cart.add.limit", part}] }
		if got, want := get("heading"), fp("Rule: The cart takes a limited quantity"); got != want {
			t.Errorf("heading: want %s, got %s", want, got)
		}
		if got, want := get("row "+rowKey("quantity", "11", "product", "apple")), fp("| 11 | apple | no | refused | more than a shelf holds |"); got != want {
			t.Errorf("row: want the fingerprint %s of its line, got %s", want, got)
		}
		steps := map[string]string{
			"step When":      "When      a buyer adds the quantity <quantity> of the [item] <product>",
			"step Provided":  "Provided  <quantity> is at most ´max_quantity´",
			"step Then":      "Then      the cart holds <quantity> more of <product>\n  And     nothing else changes",
			"step Otherwise": "Otherwise the cart refuses <product>\n  And     the cart stays as it was",
		}
		for part, text := range steps {
			got, want := get(part), stepFPs(text)
			if got != want[0] && got != want[1] {
				t.Errorf("%s: want %s (with its keyword) or %s (without), got %s", part, want[0], want[1], got)
			}
		}
	})
}

func TestBaselinePartsRowKey(t *testing.T) {
	t.Parallel()
	head := "| id | provided | outcome | why |\n| --- | --- | --- | --- |\n"
	tail := "| b | no | nothing shown | |\n"
	keyed := func(t *testing.T, table string) map[[2]string]string {
		t.Helper()
		return manifestParts(manifestOf(t, withRule(idBlock, table)))
	}
	rowA := "| a | yes | error | |"
	key := "row " + rowKey("id", "a")

	t.Run("row1 keyed by its placeholder values", func(t *testing.T) {
		t.Parallel()
		m := keyed(t, head+rowA+"\n"+tail)
		if got, ok := m[[2]string{"cart.x.r", key}]; !ok {
			t.Errorf("want the part %q, got %v", key, partsOf(manifestOf(t, withRule(idBlock, head+rowA+"\n"+tail)), "cart.x.r"))
		} else if got != fp(rowA) {
			t.Errorf("want the fingerprint %s of the row's line, got %s", fp(rowA), got)
		}
	})
	t.Run("row2 a new outcome keeps the key", func(t *testing.T) {
		t.Parallel()
		changed := "| a | yes | nothing reported | |"
		m := keyed(t, head+changed+"\n"+tail)
		if got, ok := m[[2]string{"cart.x.r", key}]; !ok {
			t.Errorf("want the part %q", key)
		} else if got != fp(changed) {
			t.Errorf("want the fingerprint %s of the changed line, got %s", fp(changed), got)
		}
	})
	t.Run("row3 the order of the columns does not change the key", func(t *testing.T) {
		t.Parallel()
		m := keyed(t, "| outcome | id | provided | why |\n| --- | --- | --- | --- |\n| error | a | yes | |\n| nothing shown | b | no | |\n")
		if _, ok := m[[2]string{"cart.x.r", key}]; !ok {
			t.Errorf("want the part %q after the columns were swapped", key)
		}
	})
	t.Run("row4 a row below another with the same values", func(t *testing.T) {
		t.Parallel()
		first, second, third := "| ? | yes | shown | |", "| ? | no | nothing shown | |", "| ? | no | ? | |"
		m := keyed(t, head+first+"\n"+second+"\n"+third+"\n")
		q := "row " + rowKey("id", "?")
		want := map[string]string{q: first, q + " #2": second, q + " #3": third}
		for part, line := range want {
			if got, ok := m[[2]string{"cart.x.r", part}]; !ok {
				t.Errorf("want the part %q", part)
			} else if got != fp(line) {
				t.Errorf("want %q to be the row %q, got the fingerprint %s", part, line, got)
			}
		}
	})
	t.Run("derived: several placeholder columns in the byte order of their names", func(t *testing.T) {
		t.Parallel()
		m := manifestParts(manifestOf(t, baseline()))
		// The table's columns are quantity, then product.
		part := "row " + fp("product=apple\tquantity=1")
		if _, ok := m[[2]string{"cart.add.limit", part}]; !ok {
			t.Errorf("want the part %q, got %v", part, partsOf(manifestOf(t, baseline()), "cart.add.limit"))
		}
	})
}
