# <a id="coherence"></a> Coherence check

## Goal

The coherence check finds the untouched rules and invariants an
amendment may affect, records how the checking agent judged each and
what it found, and tells whether the check covers the amendment as it
stands, so that no amendment becomes the baseline with an effect nobody
examined.

## Rationale

A decision that changes one rule can break another that nobody edited,
and only an agent can judge which. What can be done mechanically is
done here: finding which rules and invariants to look at, recording what
each judgement covered, and comparing that with the tree as it stands.
Judging stays with the checking agent, and deciding a finding with the
engineer. Every command starts from the element diff of the amendment.
