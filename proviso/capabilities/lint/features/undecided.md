# <a id="lint.undecided"></a> Undecided

- Contact: Bruno Baketarić

## Goal

Every place the spec tree leaves undecided is reported, so that an
author sees what is still open, and nothing else is mistaken for it.

## Rationale

An undecided place is a stop for whoever builds from it, and the tree
may hold one only while an amendment is in progress. Lint reports each
as an ask, never as an error: during an amendment, undecided places are
expected, and an author must be able to lint the tree while asking. A
`?` counts only where the structure puts it, so a real question in prose, a `?` in a code span or
a rule about `?` raises no false alarm.

## <a id="lint.undecided.place"></a> Rule: An undecided place is reported

```guarded
When      lint reads a place <place> where an [undecided] mark can stand: the value of a [field], a cell of an [Examples table], the body of a [section], or a word of a [rule block]
Provided  <place> is not [undecided]
Then      lint reports nothing about <place> being [undecided]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.undecided` at <place>, or at the [section] for the body of one
  And     lint goes on with the rest of the [spec tree]
```

| place | provided | outcome | why |
| --- | --- | --- | --- |
| the field `- Contact: Bruno Baketarić` | yes | nothing reported | |
| the field `- Question: What does the daily limit apply to?` | yes | nothing reported | a question mark ending a sentence |
| the field `- Definition: ?` | no | ask at the field | an open entry's definition |
| the field `- Contact: ?` | no | ask at the field | |
| the field `- Contact: ? (Cash Operations or Treasury)` | no | ask at the field | a note after the mark leaves the value undecided |
| the field `- Contact:   ?   ` | no | ask at the field | surrounding whitespace does not matter |
| the cell `error` | yes | nothing reported | |
| the cell `` `?` `` | yes | nothing reported | a value shown in a code span |
| the cell `?` | no | ask at the cell | an outcome nobody has decided |
| the cell `? depends on the limit` | no | ask at the cell | a note after the mark |
| the cell `??` | yes | nothing reported | not the mark |
| the body "Proviso keeps one specification of a system true …" | yes | nothing reported | |
| the body "Why lowercase? Because links on some hosts …" | yes | nothing reported | a real question in prose |
| the body `?` | no | ask at the section | a skeleton's section, as `proviso init` writes it |
| the body `? to be agreed with the business` | no | ask at the section | a note after the mark |
| the words of `Then      lint reads <file>` | yes | nothing reported | |
| the words of `Provided  <value> is "yes", "no", "no (<part>)" or "?"` | yes | nothing reported | a `?` inside a literal is a value |
| the `?` of `Then      ?` | no | ask at the `?` | an outcome nobody has decided |
| the `?` of `  And     ?` | no | ask at the `?` | what else happens is undecided |
| the `?` of `Provided  <count> ? the [team limit]` | no | ask at the `?` | the comparison itself is open |

## <a id="lint.undecided.term"></a> Rule: A rule using an open term is reported

### Rationale

A rule that marks an open term depends on a meaning nobody has decided
yet, so the rule is undecided too, even where its own text holds no
`?`.

```guarded
When      lint reads a [term] mark in a [rule block], resolving to the [glossary entry] <entry>
Provided  the status of <entry> is "defined"
Then      lint reports nothing about the mark
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.undecided` at the mark, naming <entry>
  And     lint goes on with the rest of the [spec tree]
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| `system.spec-tree`, defined | yes | nothing reported | |
| `atm.daily-limit`, open | no | ask, naming the entry | the rule waits for the term's meaning |
