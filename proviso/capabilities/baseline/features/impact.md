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
which ones to look at can be found mechanically, through what stands
beside a changed rule or above it: the rules on the same event, the
invariants above it, the constraints and goals that frame it. What
depends on a changed meaning, such as a rule marking a redefined term,
is no candidate: the element diff lists it as changed through a
dependency. Sharing a term or a reference with a changed rule is no
link at all, as long as the term or the reference keeps its meaning.
Finding too many costs a judgement each, often in bulk; finding too few
hides an effect, the failure Proviso exists to prevent. So the sources
are coarse on purpose, and the coherence check reads the capability for
what they miss.

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
| `lint.examples.cells`, unchanged, on the event of a changed rule | yes | a candidate | |
| `lint.run.root`, changed itself | no | no candidate | it is in the element diff |
| `lint.run.exit`, changed through a dependency | no | no candidate | it is in the element diff |

## <a id="baseline.impact.events"></a> Rule: A touched rule reaches the rules with its event head

### Rationale

Rules reacting to one event share their outcomes' ground: two guards on
one event can overlap or leave a gap. The full text of a `When` almost
never matches between such rules, since each phrases its qualifiers in
its own words; the event head does.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `change_effects` opens "events" for a [part] of <element>
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
Then      it is the words of the "When" [step] of <rule> up to its first [placeholder], every character but a letter, a mark or a digit separating them, with case ignored and the articles left out
  And     two [event heads] are the same when that text is
```

| rule | outcome |
| --- | --- |
| `lint.examples.row-repeated`, `When` `lint reads the row <row> of an [Examples table], below the rows <earlier>` | "lint reads row" |
| `lint.examples.cells`, `When` `lint reads the row <row> of an [Examples table] with the columns <columns>` | "lint reads row", the same |
| `baseline.diff.added`, `When` `` `diff` compares the [element] <element> of the [spec tree] with the [baseline] `` | "diff compares element" |
| a rule whose `When` holds no placeholder | the words of the whole step |
| a `When` reading `lint reads the [hand-over] to <target>` | "lint reads hand over to", the hyphen separating words |

## <a id="baseline.impact.ranges"></a> Rule: A touched rule reaches the invariants above it

### Rationale

An invariant's Ranges over is prose, which no mechanism can match
against a `When`. Its level stands in for it: every invariant of the
rule's feature, its capability and the system may range over the rule's
event. It is coarse, and the mirror of scope.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `change_effects` opens "ranges" for a [part] of <element>
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
everything at and below the level they stand at, whether or not a rule
shares a word with them. A new constraint may share no term and no event
with the rules it touches, and a new Goal changes how the agents fill
the gaps between rules. Scope is the only source that does not depend
on wording.

```guarded
When      `impact` takes the [touched element] <element>
Provided  `change_effects` opens "scope" for a [part] of <element>
Then      every [rule] and invariant at and below the [level] of <element> is found through "scope", naming <element>
  And     the [candidates] found through "scope" are also listed together under <element>, so they can be judged in bulk
Otherwise nothing is found through "scope" from <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `system.git`, its Property changed | yes | every rule and invariant of the tree found | a system constraint |
| `baseline.run.writes`, an invariant of the capability, its Property changed | yes | every rule and invariant of the capability `baseline` found | |
| `lint.run.read-only`, an invariant of the feature `lint.run`, its Property changed | yes | every rule of the feature `lint.run` found | an invariant binds the level it stands at |
| `lint.run`, its Goal changed | yes | every rule and invariant of the feature `lint.run` found | |
| `system`, its Scope changed | yes | every rule and invariant of the tree found | |
| `lint.run`, its Rationale reworded | no | nothing found | |

## <a id="baseline.impact.order"></a> Rule: The candidates are ordered

```guarded
When      `impact` writes the [candidates] <candidates>
Then      they appear ordered by [identifier], compared byte by byte
  And     the reasons of each appear ordered by source, "events", "ranges", then "scope", then by what they name
```

| candidates | outcome |
| --- | --- |
| `lint.run.root`, `lint.run.exit`, `baseline.run.writes` | `baseline.run.writes`, `lint.run.exit`, `lint.run.root` |
| `lint.run.exit`, found through "scope" and through "events" | its "events" reason first, then its "scope" reason |

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
