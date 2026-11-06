# <a id="lint.references"></a> References

- Contact: Bruno Baketarić

## Goal

Every term mark, hand-over and backtick reference in the spec tree
resolves to exactly one element, every use of a term in a rule block is
marked, and a word that is a term of another capability is asked about.

## Rationale

Impact sets follow references: the rules that use a term, the rules a
chain hands over to, the rules that refer to a contract element. A
reference that resolves to nothing, or to either of two elements, hides
a rule from the impact set, and so does a term used without its mark.
Resolution follows the levels: a capability's own glossary, contracts
and parameters first, then the system's.

Inside a rule block, a backtick is always a reference and `"…"` always
a literal, so that every backtick resolves and a literal holds anything
without a false alarm. A reference is qualified from its top-level
element, never shortened to what is unique today, since a new element
would then break references that resolved before.

## <a id="lint.references.term"></a> Rule: A term mark resolves

```guarded
When      lint reads a [term] mark naming <name>, in a [rule block], a "Represents" [field], or a "Definition" or "Deviations" [field] of a [glossary entry], of <capability>, a capability or the system
Provided  <name> is the name or an alias of a [glossary entry] in the glossary of <capability> or of the system, compared as `name_comparison` says
Then      the mark refers to that entry
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.term-unresolved` at the mark
  And     lint goes on with the rest of the [spec tree]
```

| name | capability | provided | outcome | why |
| --- | --- | --- | --- | --- |
| spec tree | lint | yes | the system's entry | |
| findings | lint | yes | lint's entry `finding` | by an alias |
| Spec Tree | lint | yes | the system's entry | names are matched ignoring case |
| spec  tree | lint | yes | the system's entry | any run of spaces and tabs joins the words of a name |
| U.S.  law, with two spaces | lint, while lint defines U.S. law | yes | that entry | every run of spaces and tabs within a name counts as one, next to punctuation too |
| spec tree, broken after "spec" onto the next line of a "Represents" field | lint | yes | the system's entry | a line break within a field joins words too |
| spec tree, in the "Definition" field of a glossary entry | lint | yes | the system's entry | a definition may mark the terms it relies on |
| spec tree, in the "Deviations" field of a glossary entry | lint | yes | the system's entry | so may deviations |
| spec-tree, in the "Definition" field of a glossary entry | lint | no | error | a mark in a definition resolves as any other |
| spec-tree, in the "Deviations" field of a glossary entry | lint | no | error | so does a mark in deviations |
| finding, in the "Definition" field of a glossary entry of the system | the system, while lint defines finding | no | error | a capability's terms are not visible at the system level |
| spec tree, in a "Definition" field that holds `` `[spec-tree]` `` before it | lint | yes | the system's entry, and nothing reported for the code span | brackets within a code span mark nothing |
| finding, in a "Represents" field that holds `` `[spec-tree]` `` before it | lint | yes | lint's entry, and nothing reported for the code span | in every field alike |
| spec tree, in a "Definition" field that holds `\[spec-tree]` before it | lint | yes | the system's entry, and nothing reported for the escaped brackets | a bracket escaped by a backslash marks nothing |
| spec-tree, in a "Definition" field, written `\\[spec-tree]` | lint | no | error | an escaped backslash escapes no bracket |
| U.S. law, written `[U\.S\. law]` in a "Definition" field | lint, while lint defines U.S. law | yes | that entry | within a mark in a field, an escape stands for the character it escapes |
| spec] tree, written `[spec\] tree]` in a "Definition" field | lint | no | error | an escaped `]` closes no mark |
| spec tree, in a rule block, written `\[spec tree]` | lint | yes | the system's entry | a rule block is a fenced block, where a backslash escapes nothing |
| " spec tree " | lint | yes | the system's entry | spaces at either end do not count |
| spec tree, followed by a non-breaking space | lint | yes | the system's entry | any white space at either end does not count |
| spec tree, followed by a zero-width space | lint | no | error | a zero-width space is no white space |
| übersicht, its "ü" written as "u" and a combining diaeresis | lint, while lint defines Übersicht, its "Ü" composed | yes | that entry | names are compared in NFC |
| spec tree, with a non-breaking space between the words | lint | no | error | only spaces and tabs join the words of a name |
| STRASSE, while lint defines Straße | lint | no | error | simple lower case, not full case folding |
| J̌, written as "J" and a combining caron | lint, while lint defines ǰ, a single character | yes | that entry | lower case comes before NFC |
| finding | install, while lint defines finding | no | error | lint's own terms are not visible in another capability |
| spec-tree | lint | no | error | neither a name nor an alias |

## <a id="lint.references.marked"></a> Rule: Every use of a term is marked

### Rationale

A term used without its mark is invisible to the impact search through
terms: redefining the term would leave that rule out. Possessives count
as uses, since an apostrophe ends a word, and so do plurals, which is
what aliases are for.

```guarded
When      lint reads the unmarked text <text> of a [rule block] in the capability <capability>
Provided  <text> holds no match that is a name or alias of a [glossary entry] in the glossary of <capability> or of the system, compared as `name_comparison` says
Then      lint reports nothing about <text> holding an unmarked [term]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.term-unmarked` at each such match
  And     lint goes on with the rest of the [spec tree]
```

Unmarked text is what remains of a line once its keyword, term marks,
placeholders, references, literals and the target of a hand-over are
taken out. Brackets that make no term mark, as the term says, are no
piece taken out: the text between them is unmarked text, and what
follows is read as the rest of the line is. Matches are found from its
start onwards: at each place, the longest name or alias of a glossary
entry that starts there as a whole word is a match, and matching goes
on after it. For matching, a word is a run of letters, combining marks,
digits, hyphens and underscores, so that a letter written with a
combining mark stays in its word; any other character ends it, and so
does each piece taken out, which no match spans. Every glossary counts
at once, its own capability's, the system's and the other
capabilities', so the match that starts first wins, and a word within a
match is no use of its own.

| text | capability | provided | outcome | why |
| --- | --- | --- | --- | --- |
| lint goes on with the rest of the | lint | yes | nothing reported | "spec tree" was marked |
| lint goes on with the rest of the spec tree | lint | no | error at "spec tree" | |
| lint reads it as the guard's truth value | lint | no | error at "guard" | a possessive is a use |
| lint reads it as the guard-line | lint | yes | nothing reported | a hyphen joins "guard-line" into one word |
| lint reads the line <x> ending | lint | yes | nothing reported | no match spans the placeholder taken out |
| lint reads the line  ending | lint | no | error at "line  ending" | any run of spaces and tabs joins the words of a name |
| lint reads the übersicht, its "ü" written as "u" and a combining diaeresis | lint, while lint defines Übersicht, its "Ü" composed | no | error at "übersicht" | the combining mark stays in its word, and names are compared in NFC |
| lint reports nothing about | lint | yes | nothing reported | |
| lint reads the [spec tree](notes.md) of | lint | no | error at "spec tree" | brackets followed by `(` make no term mark, so their text is unmarked text |
| lint reads the [rest of it](notes.md) | lint | yes | nothing reported | brackets followed by `(` make no term mark, so their text resolves to nothing |
| lint reads the [spec tree][tree] | lint | yes | nothing reported | two term marks |
| lint reads the date heading | check, which defines heading, while bundle defines date heading | yes | nothing reported | "date heading" is the match that starts first |
| lint reads the line ending date | lint, while diff defines ending date | no | error at "line ending" | the match that starts first wins |

## <a id="lint.references.foreign"></a> Rule: Another capability's term is asked about

### Rationale

Capabilities do not see each other's terms, so a word that is a term of
another capability is plain text here, and the impact search through
terms misses the rule when that term changes. Only a person can weigh
whether the word means that term, so lint asks rather than refuses. If
it does, the term moves to the system's glossary and the word is marked;
if not, the word is reworded, since a word that can be read as either
will be. A word that is a term here as well is left to
`lint.references.marked`.

```guarded
When      lint reads the unmarked text <text> of a [rule block] in the capability <capability>
Provided  <text> holds no match that is a name or alias of a [glossary entry] in the glossary of another capability, and of none in the glossary of <capability> or of the system, compared as `name_comparison` says
Then      lint reports nothing about <text> holding a [term] of another capability
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.term-foreign` at each such match, naming each [glossary entry] that has it
  And     lint goes on with the rest of the [spec tree]
```

Matches are found as `lint.references.marked` says, over every glossary
at once.

| text | capability | provided | outcome | why |
| --- | --- | --- | --- | --- |
| lint reads it as a line of a | lint | yes | nothing reported | |
| lint reports nothing about the parts of the | lint, while baseline defines part, with the alias parts | no | ask at "parts", naming baseline's entry | by an alias |
| lint reads it as Part of a | lint, while baseline defines part | no | ask at "Part", naming baseline's entry | names are matched ignoring case |
| lint reads it as partial | lint, while baseline defines part | yes | nothing reported | not a whole word |
| the finding is read | install, while lint defines finding | no | ask at "finding", naming lint's entry | |
| the part is read | diff, which defines part, as baseline does | yes | nothing reported | a term of its own; `lint.references.marked` reports it unmarked |
| the part is read | cli, while baseline and diff each define part | no | ask at "part", naming both entries | |
| lint reads the date heading | check, which defines heading, while bundle defines date heading | no | ask at "date heading", naming bundle's entry | the match that starts first wins over the capability's own term |
| the line ending is read | lint, while diff defines line | yes | nothing reported | the system's "line ending" is the longer match |
| lint reads the line ending date | lint, while diff defines ending date | yes | nothing reported | "line ending" starts first, so "ending date" is no match |

## <a id="lint.references.term-ambiguous"></a> Rule: A term name names one entry

```guarded
When      lint reads the name or alias <name> of a [glossary entry] <entry>
Provided  no other [glossary entry] in the same glossary has <name> as its name or an alias, compared as `name_comparison` says
  And     no entry of a glossary at the [level] above or below has <name> as its name or an alias, compared so too
Then      lint reports nothing about <name>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.term-ambiguous` at <entry>, naming the other entry
  And     lint goes on with the rest of the [spec tree]
```

| name | entry | provided | outcome | why |
| --- | --- | --- | --- | --- |
| finding | `lint.finding` | yes | nothing reported | |
| Spec  Tree | an entry of lint named Spec  Tree, with two spaces, while the system defines spec tree | no (level above or below) | error | runs of spaces count as one |
| finding | `lint.finding`, and `baseline.finding` in another capability | yes | nothing reported | capabilities do not see each other's terms |
| findings | `lint.finding`, and `lint.findings-list` with the alias findings | no (same glossary) | error | reported at each of the two |
| element | `lint.element`, while the system defines element | no (level above or below) | error | a capability may not redefine a system term; reported at each of the two |

## <a id="lint.references.hand-over"></a> Rule: A hand-over reaches a rule

```guarded
When      lint reads the [hand-over] to <target>
Provided  <target> is the [identifier] of a [rule]
Then      the [hand-over] leads to that [rule]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.hand-over-unresolved` at the [hand-over]
  And     lint goes on with the rest of the [spec tree]
```

| target | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.layout.file` | yes | leads to that rule | |
| `lint.layout` | no | error | a feature, not a rule |
| `lint.layout.files` | no | error | no such identifier |
| `lint.layout.file.` | no | error | the target is taken as written, a trailing dot included |

## <a id="lint.references.chain"></a> Rule: A chain of hand-overs ends

### Rationale

A chain decides by handing over until a rule's own `Otherwise` states an
outcome. A chain that returns to a rule it has passed never decides.

```guarded
When      lint follows the [hand-overs] from a [rule] through the [rules] <chain>
Provided  no [rule] appears twice in <chain>
Then      lint reports nothing about the chain
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.hand-over-cycle` at the [heading] of the first [rule] that appears twice, naming the other [rules] of the cycle
  And     lint goes on with the rest of the [spec tree]
```

| chain | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.layout.hidden`, `lint.layout.file` | yes | nothing reported | ends in an outcome |
| `a`, `a` | no | error at `a` | a rule handing over to itself |
| `a`, `b`, `a` | no | error at `a` | reported once per cycle, with `b` in `related` |

## <a id="lint.references.reference"></a> Rule: A backtick reference resolves

```guarded
When      lint reads the [backtick reference] <reference> in a [rule block] of the capability <capability>
Provided  the first segment of <reference> names a [top-level element] or a [parameter] of <capability> or of the system
  And     each further segment names a [member] of the one before
Then      the [backtick reference] refers to that [element]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.reference-unresolved` at <reference>
  And     lint goes on with the rest of the [spec tree]
```

| reference | capability | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `findings.anchor-misplaced` | lint | yes | that kind | |
| `lint.exit-status` | lint | yes | the exit status | |
| `tree_layout` | lint | yes | the parameter | |
| `finding.anchor-misplaced` | lint | no (first segment) | error | the term's name, not the contract's |
| `findings.anchor-missing` | lint | no (further segment) | error | no such kind |
| `exit-status` | lint | no (first segment) | error | not qualified from its top-level element |
| `lint.cli.lint` | lint | no (first segment) | error | an identifier, not a name |

## <a id="lint.references.reference-ambiguous"></a> Rule: A contract or parameter name names one element

```guarded
When      lint reads the name <name> of a [top-level element] or [parameter] <element>
Provided  no other [top-level element] or [parameter] at the same [level] has <name>
  And     none at the [level] above or below has <name>
Then      lint reports nothing about <name>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.reference-ambiguous` at <element>, naming the other [element]
  And     lint goes on with the rest of the [spec tree]
```

| name | element | provided | outcome | why |
| --- | --- | --- | --- | --- |
| findings | `lint.findings` | yes | nothing reported | |
| report | `lint.cli.report`, and a parameter report in lint | no (same level) | error | contracts and parameters share one namespace; reported at each of the two |
| report | `lint.cli.report`, and a contract report at system level | no (level above or below) | error | no shadowing; reported at each of the two |
