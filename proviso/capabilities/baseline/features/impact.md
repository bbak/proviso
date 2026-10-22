# <a id="baseline.impact"></a> Impact candidates

- Contact: Bruno Baketarić

## Goal

Finding the impact candidates lists every rule and invariant an
amendment may affect without touching it, each with every reason it may
be affected, so that the coherence check judges each one and nothing
affected is left unexamined.

## Rationale

A decision that changes one rule can break another that nobody edited.
Which ones is a question of meaning, and only an agent can judge it; but
which ones to look at can be found mechanically, through what the rules
share: a term, a reference, an event, the invariants above them, the
constraints and goals that frame them. Finding too many costs a
judgement each, often in bulk; finding too few hides an effect, the
failure Proviso exists to prevent. So the sources are coarse on
purpose, and the coherence check reads the capability for what they
miss.

`impact` derives the element diff as `diff` does, and starts from its
touched elements. "Formerly" always means the baseline: a touched
element is compared with what it was at the last acceptance, never with
a state in between.

## <a id="baseline.impact.listed"></a> Rule: Only an element the diff does not list is a candidate

### Rationale

Every element the element diff lists is changed by the amendment and
goes to the implementing agents anyway; judging it would add nothing.

```guarded
When      `impact` finds the [element] <element> through a source
Provided  the [element diff] does not list <element>
Then      <element> is a [candidate], naming that source among its reasons
  And     everything else stays as it was
Otherwise <element> is no [candidate]
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, unchanged, marking a term a changed rule marks | yes | a candidate | |
| `lint.run.root`, changed itself | no | no candidate | it is in the element diff |
| `lint.run.exit`, changed by a term | no | no candidate | it is in the element diff |

## <a id="baseline.impact.terms"></a> Rule: A touched rule or definition reaches the rules marking its terms

```guarded
When      `impact` takes the [touched element] <element>
Provided  `impact_triggers` opens "terms" for a [part] of <element>
Then      every [rule] that marks a [term] which <element> marks or defines, now or in the [baseline], is found through "terms", naming that [term]
  And     everything else stays as it was
Otherwise nothing is found through "terms" from <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, its `Then` step changed, marking `findings` | yes | every rule marking `findings` found | |
| `lint.run.exit`, its `Then` step changed, a mark of `findings` removed from it | yes | every rule marking `findings` found | the term it marked in the baseline |
| `lint.step`, its Definition changed | yes | every rule marking `step` found | they are changed by a term, so the diff lists them already |
| `lint.run`, its Contact changed | no | nothing found | |

## <a id="baseline.impact.references"></a> Rule: A touched rule, contract element or parameter reaches the rules that refer to the same

### Rationale

A backtick reference ties a rule to a contract element or a parameter
as a term ties it to a definition. A contract element and the term it
represents are two views of one thing, so the link is followed both
ways.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `impact_triggers` opens "references" for a [part] of <element>
Then      every [rule] that refers to <element>, or to a [contract element] or [parameter] that <element> refers to, now or in the [baseline], is found through "references", naming what it refers to
  And     every [rule] that marks a [term] a [contract element] among them represents, or refers to a [contract element] that represents a [term] <element> defines, is found through "references" as well
Otherwise nothing is found through "references" from <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.cli.lint.exit-status`, a new value | yes | every rule referring to `lint.exit-status` found | |
| `lint.run.exit`, its `Then` step changed, referring to `lint.exit-status` | yes | every other rule referring to `lint.exit-status` found | |
| `lint.findings`, a new member | yes | every rule marking `finding` found | it represents that term |
| `lint.tree-layout`, its Rationale reworded | no | nothing found | |

## <a id="baseline.impact.events"></a> Rule: A touched rule reaches the rules with its event head

### Rationale

Rules reacting to one event share their outcomes' ground: two guards on
one event can overlap or leave a gap. The full text of a `When` almost
never matches between such rules, since each phrases its qualifiers in
its own words; the event head does.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `impact_triggers` opens "events" for a [part] of <element>
Then      every [rule] with the [event head] of <element>, now or in the [baseline], is found through "events", naming that [event head]
  And     everything else stays as it was
Otherwise nothing is found through "events" from <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.examples.row-repeated`, added, with the event head "lint reads row" | yes | `lint.examples.cells` found | its `When` reads `lint reads the row <row> of an [Examples table] with the columns <columns>` |
| `lint.run.exit`, its Rationale reworded | no | nothing found | |

## <a id="baseline.impact.event-head"></a> Rule: The event head is the When up to its first placeholder

```guarded
When      `impact` takes the [event head] of the [rule] <rule>
Then      it is the text of the "When" [step] of <rule> up to its first [placeholder], with case, articles, punctuation, [term] marks and backticks ignored
  And     two [event heads] are the same when that text is
```

| rule | outcome |
| --- | --- |
| `lint.examples.row-repeated`, `When` `lint reads the row <row> of an [Examples table], below the rows <earlier>` | "lint reads row" |
| `lint.examples.cells`, `When` `lint reads the row <row> of an [Examples table] with the columns <columns>` | "lint reads row", the same |
| `baseline.diff.added`, `When` `` `diff` compares the [element] <element> of the [spec tree] with the [baseline] `` | "diff compares element" |
| a rule whose `When` holds no placeholder | the whole text of the step |

## <a id="baseline.impact.ranges"></a> Rule: A touched rule reaches the invariants above it

### Rationale

An invariant's Ranges over is prose, which no mechanism can match
against a `When`. Its level stands in for it: every invariant of the
rule's feature, its capability and the system may range over the rule's
event. It is coarse, and the mirror of scope.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `impact_triggers` opens "ranges" for a [part] of <element>
Then      every invariant of the feature and the capability of <element>, and every invariant of the system, is found through "ranges", naming the [level] it stands at
  And     everything else stays as it was
Otherwise nothing is found through "ranges" from <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, its `Then` step changed | yes | `lint.run.read-only` found | the invariant of its feature |
| `baseline.diff.exit`, added | yes | `baseline.run.writes` found | the invariant of its capability |
| `lint.run.exit`, its Rationale reworded | no | nothing found | |

## <a id="baseline.impact.scope"></a> Rule: A touched frame reaches everything below it

### Rationale

A constraint, an invariant, the system's sections and a Goal frame
everything at and below their level, whether or not a rule shares a
word with them: a new constraint may share no term and no event with the
rules it touches, and a new Goal changes how the agents fill the gaps
between rules. Scope is the only source that does not depend on wording.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `impact_triggers` opens "scope" for a [part] of <element>
Then      every [rule] and invariant at and below the [level] of <element> is found through "scope", naming <element>
  And     the [candidates] found through "scope" are also listed together under <element>, so they can be judged in bulk
Otherwise nothing is found through "scope" from <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `system.git`, its Property changed | yes | every rule and invariant of the tree found | a system constraint |
| `baseline.run.writes`, its Property changed | yes | every rule and invariant of the capability `baseline` found | |
| `lint.run`, its Goal changed | yes | every rule and invariant of the feature `lint.run` found | |
| `system`, its Scope changed | yes | every rule and invariant of the tree found | |
| `lint.run`, its Rationale reworded | no | nothing found | |

## <a id="baseline.impact.order"></a> Rule: The candidates are ordered

```guarded
When      `impact` writes the [candidates] <candidates>
Then      they appear ordered by [identifier], compared byte by byte
  And     the reasons of each appear ordered by source, as `impact_triggers` names them first, then by what they name
```

| candidates | outcome |
| --- | --- |
| `lint.run.root`, `lint.run.exit`, `baseline.run.writes` | `baseline.run.writes`, `lint.run.exit`, `lint.run.root` |
| `lint.run.exit`, found through "events" and through "terms" | its "terms" reason first, then its "events" reason |

## <a id="baseline.impact.output"></a> Rule: The candidates are written as text or as JSON

```guarded
Given     `impact` has found the [candidates] <candidates>
When      `impact` writes them, run with the options <options>
Provided  <options> include `impact.json`
Then      `impact` writes the `impact_set` of <candidates> to "stdout"
  And     `impact` exits with `impact.exit-status` 0
Otherwise `impact` writes <candidates> to "stdout" as text
  And     `impact` exits with `impact.exit-status` 0
```

| candidates | options | provided | outcome | why |
| --- | --- | --- | --- | --- |
| two | `--json` | yes | an `impact_set` with two candidates, exit status 0 | |
| none | `--json` | yes | an `impact_set` with no candidates, exit status 0 | finding none is no failure |
| two | none | no | the two as text, exit status 0 | |
| none | none | no | nothing written, exit status 0 | |
