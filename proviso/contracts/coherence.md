# Coherence

## <a id="system.coherence-file"></a> coherence_file

The entry `.coherence.json` in the spec root, whatever it is, written
only by `judge`, read by `judge` and `verify`, removed by `manifest`,
and never edited by hand. As `judge` writes it, it is a file of JSON in
UTF-8.

- **It is committed on the amendment's branch,** so that the check
  survives sessions and reaches other engineers and CI.
- **Its name starts with a dot,** so lint skips it: it is no element,
  and the manifest does not record it.
- **What it holds is not part of the contract.** Nothing but `judge`
  and `verify` reads it; what it records shows through them.
