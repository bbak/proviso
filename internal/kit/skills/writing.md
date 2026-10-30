## How to write the spec tree

Whatever you write into the tree, a testing agent and a coding agent
build from, through a context pack, and they read it literally: both
outcomes of a rule are required behaviour, every `And` line is part of
the outcome, every `?` is a stop, and the pack is their only source.
Write for that reader.

### Never decide what nobody decided

- Every line you write traces back to what the business said, to what
  the engineer decided in this session, or to what the tree already
  says. If you cannot point to one of them, the line is a question.
- Where something is not decided, write `?` and add an open question
  to the `questions.md` of the element's capability, or of the system,
  tagged with the element's identifier and naming whom to ask:
  `- [cash.withdrawal.pay] Should a partial amount be offered? (ask: Cash Operations)`.
  Never fill the place with a plausible value, a library's default, or
  an "everything else stays as it was" that nobody said.
- Never define a glossary term yourself. A definition comes from the
  term's contact or the request's owner, in their words.
- Behaviour that is specified and built, but questioned, is doubted,
  not undecided: an open question in `questions.md`, and no `?`.

### Where an element goes

- At the lowest level that covers all its uses: rules and the
  invariants of one feature in its feature file; a term, parameter,
  invariant, constraint or contract used by one capability in that
  capability's sibling file; anything used by several at the system
  level. When a second capability starts using one, move it up; its
  identifier stays.
- An identifier is readable, `<prefix>.<name>` in lowercase with
  hyphens, such as `cash.withdrawal.pay`. The prefix names where the
  element was born, the capability or `system`; it is permanent, and
  never reused once its element is removed. Search the tree before you
  choose one.
- Before you add a term, read the glossaries of the capability and the
  system. Reuse a term that means the same thing; never add a second
  term for it, and never redefine a system term in a capability.

### Rules

A rule is one block fenced as `guarded`, its keywords padded so the
text lines up, then its Examples table.

- **`Given`** is the state before the event. It never refers to the
  event, and holds no word that needs the event to be understood, such
  as *enough* or *valid*.
- **`When`** is the event: subject, verb and object first, then the
  placeholder, then any qualifier, such as "the customer asks for
  <amount>". Rules about the same event then share the words up to the
  first placeholder, which is how the impact search finds them. A
  qualifier that narrows the event, such as "a customer with enough
  funds", hides a condition: make it the guard.
- **`Provided`** is the one condition that decides the outcome, and it
  refers to the event through a placeholder of the `When`. Parts that
  must all hold are `Provided … And …`. Several conditions with
  different outcomes are several rules, or a chain whose `Otherwise →
  <identifier>` hands over to the next rule.
- **`Then` and `Otherwise`** each have at least one `And`: what the
  outcome is for, what else happens or changes, or that everything else
  stays as it was, where someone decided that. No filler.
- **A rule without a guard** has no `Provided` and no `Otherwise`. Write
  one only where the event needs no condition by its nature, and say so
  in the session, so that no condition hides in a `Given`.
- **Outcomes may be relations,** such as "notes adding up to exactly
  <amount>", where several results are acceptable. No screens, tables,
  APIs or algorithms, unless the mechanism is what is specified.
- **Marks:** `[term]` for a glossary term, and every use of a term in a
  rule block is marked; `` `name` `` for a contract element or a
  parameter, a contract element qualified from its top-level element,
  such as `` `customer.id` ``, and never a rule's identifier; `"text"`
  for a literal value; `<name>` for a placeholder.
- **An error** is named in the `Otherwise` by its kind, a contract
  element such as `` `error.not-composable` ``.

### Examples

- One column per placeholder, then `provided`, `outcome` and `why`.
- Both sides of the guard: at least one `yes` row and one `no` row; for
  a guard of *n* parts, *n* `no` rows, each failing one part, written
  `no (<part>)`; and the boundaries, where the guard only just holds and
  only just fails.
- Check every outcome against the rule as you write it. When you change
  a rule, go through every row of its table: rows with wrong outcomes,
  and rows left stale by a changed rule, are the most common errors in
  a tree, and the agents build exactly what the rows say.

### Terms, parameters, invariants, constraints, contracts

- **A term** is a word that decides an outcome and that two careful
  readers could understand differently. Its entry has `Definition`,
  `Contact`, `Status` (`defined`, or `open` with a `Question`) and
  `Aliases` for other names and forms.
- **A parameter** has `Value`, `Unit`, `Configurable` and a one-line
  `Rationale`, which a bare value always needs.
- **An invariant** holds across many events: `Rationale`, `Ranges over`
  and `Property`. Look for one in *always*, *never*, *every*, in the
  same consequence repeated across rules, and in a lasting "the system
  never …"; ask about it with a case on another event.
- **A constraint** is checked by review or a build, never at run time:
  `Rationale` and `Property`.
- **A contract** states shape only. Every operation, message, entity
  and member a rule refers to has its own heading and identifier;
  behaviour stays in rules.
- **When you widen what something counts as,** such as "anything of
  that name", follow it into every rule that handles it: each must
  still decide every case it now covers.

### Rationale

- Systems, capabilities and features always have one; a rule only where
  a careful reader would expect something else: an alternative
  rejected, a value fixed, a standard or a default departed from,
  something deliberately left out.
- A rationale states its reason itself. It never points to the request,
  a discussion, a design document or any other text beside the tree.
- The reason for a decision goes into the Rationale of the element it
  concerns; why the amendment is made now goes into the commit message.

### Holes to look for

Every one you find ends decided in the tree or marked `?`:

- a relative word in a `Given`, or a condition hidden in a `When`;
- a `Provided` that says nothing about the event;
- an outcome without its `And`, or a guard without its alternative;
- state the outcome depends on but no step mentions, including "as
  today" or "like before";
- two rules on the same event whose guards can both hold, with
  different outcomes, or a case no rule covers;
- a condition the actor can make true by themselves;
- an outcome only one implementation would produce.

### After every edit

- `proviso lint` exits with 0.
- `proviso diff` lists what you meant to change, and `proviso impact`
  what it may affect: read every candidate, and edit the ones the
  amendment does affect.
- Commit only the tree files you edited, as `docs(spec): <what>`, with
  why this amendment is made now as the message's body, unless
  `.proviso/build.md` names another convention for commit messages;
  then follow that.
