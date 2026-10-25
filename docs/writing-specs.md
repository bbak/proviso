# Writing specs

How a spec tree is laid out and written, element by element, with a
cash machine as the running example. `proviso lint` checks every point
of the notation below and tells you what is wrong; this page explains
what each part is for. The essay
[Provided, Otherwise](https://medium.com/@__bbak/provided-otherwise-the-missing-guard-in-gherkin-af562aeaec7a)
introduces Guarded Gherkin and the reasoning behind it, and
[Proviso's own tree](https://github.com/bbak/proviso/tree/main/proviso)
is a complete example.

## The layout

The folders mirror the levels: system, capability, feature, rule.

```text
proviso/                      the system
  system.md                   Goal, Rationale, Scope, Out of scope,
                              Audience, Design boundary
  glossary.md                 terms shared across capabilities
  parameters.md               shared values
  references.md               external standards, with versions
  invariants.md               what holds across every event
  constraints.md              design decisions checked by review or build
  questions.md                doubts about elements, as open questions
  contracts/                  shared interfaces and data shapes
  capabilities/
    cash/
      capability.md           Goal, Rationale
      glossary.md, parameters.md, ...   the same siblings, local to it
      contracts/
      features/
        withdrawal.md         Goal, Rationale, Contact, rules
```

Only `system.md` is required. Every sibling is optional at every level
and exists once it has content. Keep `system.md` to about one screen:
every context pack carries it.

Each element lives at the lowest level that covers all its uses. A term
or parameter used by one capability stays in its folder; once a second
capability uses it, it moves up to the system level.

## Elements and identifiers

Every element has an identifier, written as an anchor at the start of
its heading:

```markdown
## <a id="cash.withdrawal.pay"></a> Rule: The machine pays only what it can make up
```

- **Readable:** people read diffs and impact sets by identifier, so it
  says something, such as `cash.withdrawal.pay`.
- **Permanent:** set when the element is created, never changed, never
  reused. The prefix names where the element was born, not where it
  lives, so moving it breaks nothing.
- **`system` is reserved** for the system and the elements born at its
  level, such as `system.read-only`.

A heading with an anchor is an element. A heading without one is a
section of the element above it, one level below it, such as `## Goal`;
only the sections each element type knows are allowed. Fields are list
items directly below the heading, `- Key: value`. A sibling file opens
with a title, such as `# Glossary`, and its elements follow at level 2.
In a feature file, which holds rules and invariants, headings start with
`Rule:` or `Invariant:`; elsewhere the file gives the type.

## System, capability, feature

**The system** states its Goal, Rationale, Scope, Out of scope, Audience
and Design boundary. Out of scope says what the system does not do; the
design boundary what it relies on without building it, such as a
database or a version control system.

**A capability** groups features, with a Goal and a Rationale.

**A feature** is one file: a Goal, a Rationale, a Contact, and its
rules. The Goal states its purpose in one to three sentences; if that
takes an "and" between unrelated purposes, the feature wants splitting.
The Contact names whom to ask, a person or a team, never an owner.

```markdown
# <a id="cash.withdrawal"></a> Withdrawing cash

- Contact: Cash Operations

## Goal

A customer withdraws an amount, paid in the notes the machine holds.

## Rationale

The machine never pays more or less than asked: an amount it cannot
make up is refused, not rounded.
```

## Rules in Guarded Gherkin

A rule is one block, fenced as `guarded`, and its Examples table.

````markdown
## <a id="cash.withdrawal.pay"></a> Rule: The machine pays only what it can make up

```guarded
Given     the machine holds the notes <notes>
When      the customer asks for <amount>
Provided  <amount> is [composable] from <notes>
Then      the machine pays out notes adding up to exactly <amount>
  And     it holds <notes> less the notes paid out
Otherwise the machine refuses with `error.not-composable`
  And     everything else stays as it was
```

| notes | amount | provided | outcome | why |
| --- | --- | --- | --- | --- |
| one $50, three $20 | $90 | yes | pays $50, $20, $20 | |
| one $50, three $20 | $80 | no | refuses, nothing changes | no selection adds up to $80 |
| one $50, three $20 | $110 | yes | pays every note | the most it can pay |
| one $50, three $20 | $130 | no | refuses, nothing changes | more than it holds |
````

- **`Given`** states what holds before the event, and must not refer to
  the event. **`When`** is the event. **`Provided`** is the guard: the
  condition on the event that decides the outcome, and it must refer to
  the event. **`Then`** is what happens when it holds, **`Otherwise`**
  what happens when it does not.
- **One guard per rule.** Parts that must all hold are `Provided … And
  …`. Several conditions with different outcomes are several rules.
- **Every outcome has at least one `And`:** what else happens, or that
  everything else stays as it was. Write the `And` that says what the
  outcome is for; a filler line misleads the agents.
- **An event that needs no guard** gets a rule without `Provided` and
  `Otherwise`. It always applies, and says so, so that no condition
  hides in a `Given`.
- **Marks:** `[term]` is a glossary term, `<name>` a placeholder that
  the Examples vary, `` `name` `` a contract element or a parameter, and
  `"text"` a literal value.
- **What, not how:** no screens, tables, APIs or algorithms, unless the
  mechanism is itself what is specified.
- **Outcomes may be relations,** such as "notes adding up to exactly
  `<amount>`", rather than one result only one implementation would
  give.

**A decision with more than two outcomes** is a chain of rules. Instead
of an outcome, an `Otherwise` hands over to the next rule by its
identifier:

```guarded
When      the link target <target> is classified
Provided  <target> parses as a URL with a scheme
Then      the link is [external]
  And     nothing else is checked
Otherwise → links.classification.reserved
```

## Examples tables

Every rule has one, after its block. The testing agent derives its
cases from the rows.

- **One column per placeholder,** then `provided`, `outcome` and an
  optional `why`.
- **`provided`** says `yes` or `no` per row, or `no (<which part>)` for a
  guard with several parts. A rule without a guard has no such column.
- **Both sides of the guard:** at least one `yes` row and one `no` row,
  and for a guard with several parts, a `no` row failing each part on
  its own.
- **Boundaries:** a row where the guard only just holds, and one where
  it only just fails, as $110 and $130 above.

Check every expected outcome against the rule as you write it. Rows with
wrong expectations, and rows left stale by a changed rule, are the most
common errors in a tree, and the agents build exactly what the rows say.

## Glossary

A term is a word that two careful readers could understand differently.
Its entry defines it once:

```markdown
## <a id="cash.composable"></a> composable

- Definition: An amount is composable from a set of notes if some
  selection of those notes adds up to exactly that amount.
- Contact: Cash Operations
- Status: defined
- Aliases: can make up
```

- **Once a word is a term, every use of it in a rule block is marked**
  `[composable]`. An unmarked use is an error, since the impact search
  follows the marks. Where you mean something else, use another word.
- **Aliases** cover other names and inflected forms, such as `[terms]`
  for *term*.
- **Status** is `defined` or `open`. An open term carries a `Question`
  and is not built until it is decided.
- **No shadowing:** a capability may not redefine a system term.
- **A term an external standard defines** uses `Defined by:` instead of
  `Definition:`, pointing to a reference, with `Deviations:` where the
  system departs from it.

## Parameters

A value rules refer to, with its unit, whether it is configurable, and
why it has that value:

```markdown
## <a id="cash.max-withdrawal"></a> max_withdrawal

- Value: 500
- Unit: USD
- Configurable: no
- Rationale: the limit the bank's insurance covers per withdrawal.
```

A rule refers to it in backticks: ``Provided  <amount> is at most `max_withdrawal` ``.
The rationale is required: a bare value is what a later agent most
easily "fixes". Examples keep concrete values, with rows at the value
and just past it.

## Invariants and constraints

**An invariant** holds across many events, not one, so it is not a
rule. It ranges over the events it quantifies, and states its property
in glossary terms:

```markdown
## <a id="cash.notes-held"></a> The machine pays only notes it holds

### Rationale

A note paid out that the machine does not hold is money that does not
exist in its count.

### Ranges over

Every event that changes the notes: withdrawal, refill, cancellation.

### Property

The notes paid out never exceed the notes held, for every value of note.
```

The testing agent turns an invariant into a property-based test, and an
amendment that adds an event of a covered kind pulls the invariant into
its impact set.

**A constraint** is a design decision checked by review, a schema or a
build check, never at run time: "no configurable ignore list", "no
query language of our own". It has a Rationale and a Property, and no
Ranges over. Constraints are the negative requirements nobody can read
from the code.

## Contracts

Interfaces and data shapes live in `contracts/`, as prose with embedded
blocks, such as YAML holding a JSON Schema. Every operation, message,
entity and member a rule refers to has its own heading and anchor, and
rules refer to it qualified from the top: `` `customer.id` ``. A
contract states shape only; behaviour, such as when a member is present,
stays in rules. `- Represents: [term]` links a contract element to the
glossary term that carries its meaning.

**An error has three parts:** that it occurs, and when, is the rule's
`Otherwise`; which kind it is, a contract element such as
`` `error.not-composable` ``; its exact text, a template on that element,
only where a consumer depends on the wording.

## References

External standards, listed once per level in `references.md` with their
version, a link and a contact. Rules reach them only through glossary
terms defined by them, so upgrading a standard lists every rule that
uses its terms. Standards only, never libraries: "behaves as the
library does" hands the decision back to a library default.

## Rationale

Rationale is prose in a `Rationale` section, never a step in a rule
block. Systems, capabilities and features always have one. A rule gets
one only where a careful reader could reasonably expect something else:
an alternative was rejected, a value fixed, a standard or a default
departed from, or something deliberately left out. It keeps a later
agent from "fixing" a deliberate choice.

A rationale states its reason itself. It never points to a design
document or a discussion elsewhere: a decision that holds only with a
document beside the tree has not arrived in it.

## Undecided and doubted

**Undecided** is `?`, where structure puts it: the first word of a
field value, an Examples cell or a section body, or a word of its own in
a rule block. A note after it, as in `- Contact: ? (Treasury or Cash
Operations)`, keeps it undecided. Nothing is specified there yet, so the
element is not built, and an amendment is accepted only once every
element it touches is decided.

**Doubted** is behaviour that is specified, and done, but questioned.
Doubts are open questions in `questions.md`, tagged with the element
they concern, such as `[cash.withdrawal.pay] Should a partial amount be
offered?`. They block nothing.
