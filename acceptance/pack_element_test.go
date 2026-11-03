package acceptance

// The rule baseline.pack.element (proviso/capabilities/baseline/features/pack.md)
// and the option system.cli.pack.element: the pack of one element named by
// `--element`. What holds for every such run is checked by packElementIn
// (see pack_driver_test.go). The Examples rows name elements of Proviso's
// own tree; each case names the element of the pack tree that stands in for
// it.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaselinePackElement(t *testing.T) {
	t.Parallel()
	t.Run("row1 a rule", func(t *testing.T) {
		t.Parallel()
		// lint.run.exit: cart.pay.cap.
		r, _ := accepted(t, packTree())
		o := r.packElement("cart.pay.cap")
		o.exactly(t, "System", "system")
		o.exactly(t, "Element", "cart.pay.cap")
		o.standsAt(t, packTree())
	})
	t.Run("row2 with an amendment note and an element diff listing another element", func(t *testing.T) {
		t.Parallel()
		// lint.run.root: cart.pay.total, its Then step changed.
		r, _ := accepted(t, packTree())
		tr := amended(packTree(), pTotalThen)
		r.set(tr)
		r.writeNote("- refunds\n")
		o := r.packElement("cart.pay.cap")
		o.exactly(t, "System", "system")
		o.exactly(t, "Element", "cart.pay.cap")
		o.exactly(t, "Amendment")
		nothingTaken(t, o)
		o.standsAt(t, tr)
	})
	t.Run("row3 with no manifest.sum", func(t *testing.T) {
		t.Parallel()
		// Written into a folder of its own, with no manifest and no git
		// repository of its own: neither is read.
		dir := packTree().write(t)
		o := packElementIn(t, dir, dir, filepath.Join(dir, "proviso"), "cart.pay.cap")
		o.exactly(t, "System", "system")
		o.exactly(t, "Element", "cart.pay.cap")
	})
	t.Run("row4 with a folder named .amendment.md", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		r.writeNoteFolder()
		o := r.packElement("cart.pay.cap")
		// ran asserts exit status 0, with --json and without.
		o.exactly(t, "System", "system")
		o.exactly(t, "Element", "cart.pay.cap")
	})
	t.Run("row5 an identifier the tree does not hold", func(t *testing.T) {
		t.Parallel()
		// lint.run.gone: cart.pay.gone.
		r, _ := accepted(t, packTree())
		r.packElement("cart.pay.gone").stoppedNaming(t, "cart.pay.gone")
	})
	t.Run("row6 an element removed since the baseline", func(t *testing.T) {
		t.Parallel()
		// lint.run.root: cart.pay.total.
		r, _ := accepted(t, packTree())
		r.set(amended(packTree(), removedFrom(packTree(), payPath, "cart.pay.total")))
		r.packElement("cart.pay.total").stoppedNaming(t, "cart.pay.total")
	})
	t.Run("derived: a changed element is read as it stands, with no former text", func(t *testing.T) {
		t.Parallel()
		r, _ := accepted(t, packTree())
		tr := amended(packTree(), pTotalThen)
		r.set(tr)
		o := r.packElement("cart.pay.total")
		o.exactly(t, "Element", "cart.pay.total")
		if it := o.item(t, "cart.pay.total"); len(it.formers) > 0 {
			t.Errorf("want no former text, got the parts %v", it.formers)
		}
		o.standsAt(t, tr)
	})
	t.Run("derived: a manifest that cannot be read is not read", func(t *testing.T) {
		t.Parallel()
		// A merge conflict left in manifest.sum, after its first line: it
		// would stop the pack of an amendment (baseline.diff.malformed).
		r, _ := accepted(t, packTree())
		path := filepath.Join(r.root(), "manifest.sum")
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString("<<<<<<< HEAD\n"); err != nil {
			t.Fatal(err)
		}
		f.Close()
		o := r.packElement("cart.pay.cap")
		o.exactly(t, "Element", "cart.pay.cap")
	})
}
