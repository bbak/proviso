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

## <a id="lint.undecided.field"></a> Rule: An undecided field is reported

```guarded
When      lint reads the [field] <field>
Provided  the first word of the value of <field> is not "?"
Then      lint reports nothing about <field> being [undecided]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.undecided` at <field>
  And     lint goes on with the rest of the [spec tree]
```

| field | provided | outcome | why |
| --- | --- | --- | --- |
| `- Contact: Bruno Baketarić` | yes | nothing reported | |
| `- Question: What does the daily limit apply to?` | yes | nothing reported | a question mark ending a sentence |
| `- Definition: ?` | no | ask | an open entry's definition |
| `- Contact: ?` | no | ask | |
| `- Contact: ? (Cash Operations or Treasury)` | no | ask | a note after the mark leaves the value undecided |
| `- Contact:   ?   ` | no | ask | surrounding whitespace does not matter |

## <a id="lint.undecided.cell"></a> Rule: An undecided Examples cell is reported

```guarded
When      lint reads the cell <cell> of an [Examples table]
Provided  the first word of <cell> is not "?"
Then      lint reports nothing about <cell> being [undecided]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.undecided` at <cell>
  And     lint goes on with the rest of the [spec tree]
```

| cell | provided | outcome | why |
| --- | --- | --- | --- |
| `error` | yes | nothing reported | |
| `` `?` `` | yes | nothing reported | a value shown in a code span |
| `?` | no | ask | an outcome nobody has decided |
| `? depends on the limit` | no | ask | a note after the mark |
| `??` | yes | nothing reported | not the mark |

## <a id="lint.undecided.section"></a> Rule: An undecided section is reported

```guarded
When      lint reads the body <body> of a [section]
Provided  the first word of <body> is not "?"
Then      lint reports nothing about the [section] being [undecided]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.undecided` at the [section]
  And     lint goes on with the rest of the [spec tree]
```

| body | provided | outcome | why |
| --- | --- | --- | --- |
| Proviso keeps one specification of a system true … | yes | nothing reported | |
| Why lowercase? Because links on some hosts … | yes | nothing reported | a real question in prose |
| `?` | no | ask | a skeleton's section, as `proviso init` writes it |
| `? to be agreed with the business` | no | ask | a note after the mark |

## <a id="lint.undecided.token"></a> Rule: An undecided part of a rule block is reported

```guarded
When      lint reads the line <line> of a [rule block]
Provided  <line> holds no "?" standing as a word of its own, outside literals and [backtick references]
Then      lint reports nothing about <line> being [undecided]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.undecided` at each such "?"
  And     lint goes on with the rest of the [spec tree]
```

| line | provided | outcome | why |
| --- | --- | --- | --- |
| `Then      lint reads <file>` | yes | nothing reported | |
| `Provided  <value> is "yes", "no", "no (<part>)" or "?"` | yes | nothing reported | a `?` inside a literal is a value |
| `Then      ?` | no | ask | an outcome nobody has decided |
| `  And     ?` | no | ask | what else happens is undecided |
| `Provided  <count> ? the [team limit]` | no | ask | the comparison itself is open |

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
