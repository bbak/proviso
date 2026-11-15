package acceptance

// The invariant coherence.run.writes, over the runs of the coherence
// commands `impact`, `judge` and `verify`. It shares its cases with
// baseline.run.writes (see baseline_run_test.go); what lint's gate decides
// for them is checked by TestSystemLintGate.

import "testing"

// TestCoherenceRunWrites checks the invariant coherence.run.writes over runs
// of `impact`, `judge` and `verify`, however they end: every file and folder
// below the test directory, the git repository's .git included, is as it
// was, except .coherence.json after a run of `judge` that exits with 0.
// Every other test checks it too, for each run it makes (see impactIn,
// judgeIn and verifyIn).
func TestCoherenceRunWrites(t *testing.T) {
	t.Parallel()
	// Every case of diff is a case of impact, judge and verify.
	writesOver(t, "impact", "judge", "verify")
}
