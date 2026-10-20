# <a id="lint.glossary"></a> Glossary entries

- Contact: Bruno Baketarić

## Goal

Every glossary entry is either defined, with one source for its
meaning, or openly undecided, with the question that will decide it.

## Rationale

A term carries meaning into every rule that marks it, so an entry must
say plainly whether that meaning is settled. An open entry without its
question gives its contact nothing to answer; a defined entry with a
question left in it reads as settled and unsettled at once. A meaning
taken from an external standard must reach that standard, because
upgrading the standard changes every term defined by it.

## <a id="lint.glossary.status"></a> Rule: A glossary entry is defined or open

```guarded
When      lint reads the "Status" [field] <status> of a [glossary entry]
Provided  <status> is "defined" or "open"
Then      lint reads the [glossary entry] as defined or as open
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.status-unknown` at <status>
  And     lint goes on with the rest of the [spec tree]
```

| status | provided | outcome | why |
| --- | --- | --- | --- |
| defined | yes | defined | |
| open | yes | open | |
| `?` | no | error | "open" already says that the meaning is undecided |
| draft | no | error | a request's status, not a term's |
| Defined | no | error | values are matched exactly |

## <a id="lint.glossary.open"></a> Rule: An open entry asks its question

```guarded
When      lint reads an open [glossary entry] <entry>
Provided  <entry> has a "Question" [field] with a value
  And     the "Definition" of <entry> is [undecided]
Then      lint reads <entry> as [undecided]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.open-entry-incomplete` at <entry>
  And     lint goes on with the rest of the [spec tree]
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| Definition "?", Question "What does the daily limit apply to?" | yes | undecided | |
| Definition "?", no Question | no (a Question) | error | nothing for the contact to answer |
| Definition "The amount withdrawn per day", Question "…" | no (Definition "?") | error | a definition arrives with the decision, and the status changes with it |
| Defined by a reference, Question "…" | no (Definition "?") | error | |

## <a id="lint.glossary.defined"></a> Rule: A defined entry has no question left

```guarded
When      lint reads a defined [glossary entry] <entry>
Provided  <entry> has no "Question" [field]
  And     the "Definition" of <entry> is not [undecided]
Then      lint reports nothing about the status of <entry>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.defined-entry-open` at <entry>
  And     lint goes on with the rest of the [spec tree]
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| Definition "An amount is composable …" | yes | nothing reported | |
| Defined by a reference | yes | nothing reported | |
| Definition "An amount is composable …", Question "…" | no (no Question) | error | the question was answered, or the status is wrong |
| Definition "?" | no (Definition not "?") | error | undecided, so the status is open |

## <a id="lint.glossary.definition"></a> Rule: An entry has one source for its meaning

```guarded
When      lint reads the [glossary entry] <entry> with the [fields] <fields>
Provided  <fields> hold exactly one of "Definition" and "Defined by"
Then      lint reads the meaning of <entry> from that [field]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.definition-count` at <entry>
  And     lint goes on with the rest of the [spec tree]
```

| entry | fields | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `search.composable` | Definition, Contact, Status | yes | from Definition | |
| `links.link-destination` | Defined by, Deviations, Contact, Status | yes | from Defined by | |
| `links.link-destination` | Definition, Defined by, Contact, Status | no | error | two meanings that may drift apart |
| `search.composable` | Contact, Status | no | error | no meaning at all |

## <a id="lint.glossary.defined-by"></a> Rule: A definition by reference reaches its reference

### Rationale

The link is written as an ordinary Markdown link, so a reader can follow
it on any host. Lint checks it the way a reader would follow it: the
path from the glossary file, then the identifier in the file it leads
to. The path's separator is `/`, as in every Markdown link: lint never
takes a `\` for one, so a link that only works on Windows works nowhere.

```guarded
When      lint reads the "Defined by" [field] <field> of a [glossary entry] at the [level] <level>
Provided  <field> holds a link whose path leads from the glossary to a "references.md" at <level> or at the system's [level]
  And     the link's fragment is the [identifier] of a [reference] in that file
Then      the [glossary entry] is defined by that [reference]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.defined-by-unresolved` at <field>
  And     lint goes on with the rest of the [spec tree]
```

| field | level | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `[CommonMark §6.3](../../references.md#system.ref.commonmark)` | capability `links` | yes | defined by CommonMark | the system's references, seen from a capability |
| `[CommonMark](references.md#system.ref.commonmark)` | system | yes | defined by CommonMark | |
| `CommonMark §6.3` | capability `links` | no (a link to a references.md) | error | no link to follow |
| `[CommonMark](../../glossary.md#system.ref.commonmark)` | capability `links` | no (a link to a references.md) | error | |
| `[CommonMark](../search/references.md#search.ref.x)` | capability `links` | no (a link to a references.md) | error | another capability's references are not visible |
| `[CommonMark](..\..\references.md#system.ref.commonmark)` | capability `links` | no (a link to a references.md) | error | `\` is not a separator |
| `[CommonMark](../../references.md#system.ref.commonmar)` | capability `links` | no (an identifier of a reference) | error | no such identifier |

## <a id="lint.glossary.deviations"></a> Rule: Deviations come only with a reference

```guarded
When      lint reads the "Deviations" [field] of a [glossary entry] <entry>
Provided  <entry> has a "Defined by" [field]
Then      lint reads the deviations as departures from that [reference]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.deviations-without-reference` at <entry>
  And     lint goes on with the rest of the [spec tree]
```

| entry | provided | outcome | why |
| --- | --- | --- | --- |
| `links.link-destination`, Defined by CommonMark | yes | departures from CommonMark | |
| `search.composable`, Definition only | no | error | a deviation needs something to deviate from |
