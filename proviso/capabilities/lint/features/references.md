# <a id="lint.references"></a> References

- Contact: Bruno Baketarić

## Goal

Every term mark, hand-over and backtick reference in the spec tree
resolves to exactly one element, and every use of a term in a rule block
is marked.

## Rationale

Impact sets follow references: the rules that use a term, the rules a
chain hands over to, the rules that refer to a contract element. A
reference that resolves to nothing, or to either of two elements, hides
a rule from the impact set, and so does a term used without its mark.
Resolution follows the levels: a capability's own glossary, contracts
and parameters first, then the system's.

## <a id="lint.references.term"></a> Rule: A term mark resolves

```guarded
When      lint reads a [term] mark naming <name>, in a [rule block] or a "Represents" [field] of the capability <capability>
Provided  <name> is the name or an alias of a [glossary entry] in the glossary of <capability> or of the system
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
| keyword | diff | no | error | lint's own terms are not visible in another capability |
| spec-tree | lint | no | error | neither a name nor an alias |

## <a id="lint.references.marked"></a> Rule: Every use of a term is marked

### Rationale

A term used without its mark is invisible to the impact search through
terms: redefining the term would leave that rule out. Possessives and
plurals count as uses, which is what aliases are for.

```guarded
When      lint reads the unmarked text <text> of a [rule block] in the capability <capability>
Provided  <text> holds no name or alias of a [glossary entry] in the glossary of <capability> or of the system, as a whole word, ignoring case
Then      lint reports nothing about <text>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.term-unmarked` at each such word
  And     lint goes on with the rest of the [spec tree]
```

Unmarked text is what remains of a line once term marks, placeholders,
references, literals and the target of a hand-over are taken out.

| text | capability | provided | outcome | why |
| --- | --- | --- | --- | --- |
| lint goes on with the rest of the | lint | yes | nothing reported | "spec tree" was marked |
| lint goes on with the rest of the spec tree | lint | no | error at "spec tree" | |
| lint reads it as the guard's truth value | lint | no | error at "guard" | a possessive is a use |
| lint reports nothing about | lint | yes | nothing reported | |

## <a id="lint.references.term-ambiguous"></a> Rule: A term name names one entry

```guarded
When      lint reads the name or alias <name> of a [glossary entry] <entry>
Provided  no other [glossary entry] in the same glossary has <name> as its name or an alias
  And     no entry of a glossary at the [level] above or below has <name> as its name or an alias
Then      lint reports nothing about <name>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.term-ambiguous` at <entry>, naming the other entry
  And     lint goes on with the rest of the [spec tree]
```

| name | entry | provided | outcome | why |
| --- | --- | --- | --- | --- |
| finding | `lint.finding` | yes | nothing reported | |
| keyword | `lint.keyword`, and `diff.keyword` in another capability | yes | nothing reported | capabilities do not see each other's terms |
| findings | `lint.finding`, and `lint.findings-list` with the alias findings | no (same glossary) | error | |
| element | `lint.element`, while the system defines element | no (level above or below) | error | a capability may not redefine a system term |

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

## <a id="lint.references.chain"></a> Rule: A chain of hand-overs ends

### Rationale

A chain decides by handing over until a rule's own `Otherwise` states an
outcome. A chain that returns to a rule it has passed never decides.

```guarded
When      lint follows the [hand-overs] from a [rule] through the [rules] <chain>
Provided  no [rule] appears twice in <chain>
Then      lint reports nothing about the chain
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.hand-over-cycle` at the first [rule] that appears twice
  And     lint goes on with the rest of the [spec tree]
```

| chain | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.layout.hidden`, `lint.layout.file` | yes | nothing reported | ends in an outcome |
| `a`, `a` | no | error at `a` | a rule handing over to itself |
| `a`, `b`, `a` | no | error at `a` | |

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
| report | `lint.cli.report`, and a parameter report in lint | no (same level) | error | contracts and parameters share one namespace |
| report | `lint.cli.report`, and a contract report at system level | no (level above or below) | error | no shadowing |
