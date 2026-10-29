# Request file format

A request is one Markdown file. It travels as it is: pasted into a
ticket, sent in a message or an e-mail, committed to a repository. The
engineer who brings it into the specification can split it into its
parts because the structure is fixed. `scripts/lint_request.py` checks
it. Do not invent variations.

The rules inside use a notation called Guarded Gherkin (Given / When /
Provided / Then / Otherwise). The requester never needs to know the name.

## Contents

1. The file as a whole
2. Front matter
3. Intent, Today, Scope
4. Rules
5. The rule block
6. Values, terms and literals
7. Examples
8. Glossary
9. Open questions
10. Notes for whoever builds this (fixed text)
11. Status and what "ready" means
12. The check

---

## 1. The file as a whole

One file per request, named after the request in kebab-case:
`expense-auto-approval.md`. Sections in this order, all as `##`
headings, each exactly once:

| section | content |
| --- | --- |
| `## Intent` | why this is wanted |
| `## Today` | how it works now |
| `## Scope` | what changes and what explicitly does not |
| `## Rules` | one `###` subsection per rule |
| `## Glossary` | one `###` subsection per term; `(none)` if empty |
| `## Open questions` | everything not yet known; `(none)` if empty |
| `## Notes for whoever builds this` | fixed text, see section 10 |

The file starts with front matter, then a `#` heading with the title.
No other `##` sections. See `example-request.md` for a complete file.

## 2. Front matter

```
---
title: Pay out cash only when the ATM can make up the amount
requester: Maria Keller (Cash Operations)
owner: Maria Keller (Cash Operations)
date: 2026-10-03
status: draft
---
```

| field | content |
| --- | --- |
| `title` | the request in one plain sentence |
| `requester` | who brought the request |
| `owner` | who decides what the rules mean. Ask; it is often not the requester. `TBD` until someone is named |
| `date` | date of the last change, `YYYY-MM-DD` |
| `status` | `draft` or `ready` (section 11) |
| `system` | optional: the application or process this concerns, as the requester names it |
| `ref` | optional: a ticket key or link |

No other fields.

## 3. Intent, Today, Scope

These sections are prose, written from what the requester said, in
their terms. Read them back like the rules: prose is where invented
content hides most easily. "To reduce manager workload while
maintaining compliance" sounds right, and nobody said it.

**Intent**: two to five sentences. What problem this solves, for whom,
and why now. If the requester has not said why, write `?` and ask.

**Today**: how this works now, in a few sentences or bullets. "Like
today" in a request carries meaning nobody writes down; this section is
where it goes. For something entirely new, write `Nothing yet.` If
unknown, `?`.

**Scope**: one bullet per line, each starting with `In:` or `Out:`.

```
- In: automatic approval of travel expense reports
- Out: changes to the expense policy itself
- Out: other expense types (e.g. hospitality)
```

At least one `In:` line. `Out:` lines are what the requester explicitly
excluded, never what you assume they would exclude. They stop the
builders from building more than was asked.

A section that is not known yet starts with `?`, followed by a note if
useful: `? (Maria asks Card Products)`.

## 4. Rules

Each rule is a `###` subsection under `## Rules`:

````markdown
### R1 — The ATM pays out only amounts it can make up from its notes

```guarded
Given     the ATM holds one $50 note and three $20 notes
When      the customer asks for <amount>
Provided  <amount> is [composable] from the notes the ATM holds
Then      the ATM pays out notes that sum to exactly <amount>
  And     its stock of notes is reduced by exactly the notes paid out
Otherwise the ATM refuses the request
  And     nothing else changes
```

| amount | provided | outcome | why |
| --- | --- | --- | --- |
| $90 | yes | pays | $50 + 2 × $20 |
| $80 | no | refuses | no combination makes $80 |

Decisions:
- 2026-10-03 — Any combination of notes is acceptable. (Maria Keller)
````

- Heading: `### R<n> — <the rule in one plain sentence>`. Numbers are
  local to the file: R1, R2, … Never renumber; a removed rule leaves a
  gap.
- Then the rule block (section 5), then the examples table (section 7),
  then optional lines: `Owner: <name>` (only if this rule has another
  owner than the request; `Owner: TBD` until named) and `Decisions:`
  with a bullet list: what was decided, close to the words of whoever
  decided it, who, and the date if you know it. Never guess a date.
- Open items of a rule are not listed here; they go under
  `## Open questions`, tagged with the rule number.

## 5. The rule block

A fenced code block with the info string `guarded`, so the alignment
survives rendering. Each line starts with a keyword, matched exactly,
padded so the text lines up.

| keyword | count | meaning |
| --- | --- | --- |
| `Given` | 0 or 1 | The world before the event. State only. More state goes into `And` lines under it. |
| `When` | exactly 1 | The event, with the value it is about. |
| `Provided` | 0 or 1 | The condition that decides the outcome; it relates the event to the state. Further parts that must all hold go into `And` lines under it. |
| `Then` | exactly 1 | What happens if the condition holds, or always if there is none. |
| `Otherwise` | 1 with a `Provided`, else 0 | What happens if the condition does not hold. |
| `And` | any | Continues the step above it. Indented by two spaces. |

Order: `Given`, `When`, `Provided`, `Then`, `Otherwise`, each with its
`And` lines. There is no `But`: an exception to an outcome is a
condition, and gets its own rule. A step does not continue on a line
without a keyword; write one long line instead.

**A Given does not refer to the event.** It uses no value of the When,
does not mention the event in words (*the requested amount*), and holds
no word whose meaning depends on the event (*enough, sufficient,
valid…*). A Given describes the world as it is, whether or not anyone
acts.

**A Provided refers to the event.** It uses at least one value of the
When. A condition that mentions nothing of the event is state, and
belongs in the Given.

**Then and Otherwise each have at least one `And` line:** what else
happens or changes, as the requester stated it (`And a reminder is
created…`, or `And nothing else changes`). On the success side this is
often what matters most: payment, notifications, audit records. Until
asked, `And ?`. Once answered, their answer is the line; do not add
`And ?` after it. Never fill it in yourself: "nothing else changes" is
a business statement like any other.

Further conventions:

- **One condition per rule.** Several conditions with different
  outcomes for one event are several rules.
- **A rule without Provided always applies.** Confirm that with the
  requester; a condition hidden in a Given loses its alternative.
- **Undecided parts are written as `?`,** standing as a word of its
  own: `Then      ?`, `And     ?`, or
  `<withdrawn-today> plus <amount> ? the [daily limit]` when the
  comparison itself is open.
- **Outcomes may be relations.** "pays out notes that sum to exactly
  `<amount>`" admits every valid result. Prefer this over one result only
  a particular implementation would produce.
- **No implementation detail.** No screens, buttons, tables, APIs or
  algorithms unless the business actually requires them.

## 6. Values, terms and literals

- **`<name>` is a value that varies.** Its name holds letters, digits,
  hyphens and underscores, no spaces: `<amount>`, `<withdrawn-today>`.
  Write it in lowercase, and the same way everywhere: `<Amount>` is
  another value. `<provided>`, `<outcome>` and `<why>` are taken by the
  examples table. A value introduced in the When is what the event is
  about; one that appears in a Given is state. A value used in the
  Provided, Then or Otherwise is introduced by the When or the Given.
- **`[term]` is a glossary term:** a word whose meaning the owner must
  pin down. Every `[term]` has an entry under `## Glossary`, by name or
  alias, ignoring case. **Every use of a term in a rule block is
  marked,** including its aliases and plurals: once `daily limit` is a
  term, an unmarked "daily limit" in a rule line is an error.
- **`"text"` is a literal:** a value written exactly as the system
  shows or stores it, such as `"Card blocked"`. Nothing inside a
  literal is a value, a term or a `?`.

Make a word a `[term]` only if it decides an outcome *and* two careful
readers could understand it differently (*reasonable, within policy,
team, blackout period*). Everyday nouns (*employee, report, licence*)
are not terms. Spelled-out comparisons (`<amount> is at most the
account balance`) need no term.

## 7. Examples

A Markdown table after the rule block. **Every rule has one.**

| column | when | content |
| --- | --- | --- |
| one per value | always | named exactly like the value, without `<>`: `withdrawn-today` |
| `provided` | exactly when the rule has a Provided | `yes`, `no`, `no (<which part>)` or `?` |
| `outcome` | always | short name of what happens, matching Then or Otherwise |
| `why` | optional | one line a business reader can check |

No other columns. Every row has one cell per column, and every cell but
`why` holds something: `?` for a value not decided yet.

- **Each row is a case of its own:** no two rows with the same values.
  A rule without values has exactly one row.
- **Both sides of the condition:** at least one `yes` and one `no`.
- **A condition of *n* parts** (`Provided` with *n* − 1 `And` lines)
  needs at least *n* `no` rows, each failing one part on its own:
  `no (<which part>)`.
- **Boundaries** (the condition just holds, just fails) belong in the
  table whenever the wording leaves them open.

## 8. Glossary

One `###` subsection per term under `## Glossary`, its fields as a
bullet list:

```markdown
### composable

- Definition: An amount is composable from a set of notes if some
  selection of those notes adds up to exactly that amount.
- Owner: Maria Keller (Cash Operations)
- Status: defined
- Source: Maria Keller, 2026-10-03
- Aliases: can make up

### daily limit

- Definition: ?
- Owner: TBD
- Status: open
- Question: What does the daily limit apply to?
```

| field | when | content |
| --- | --- | --- |
| `Definition` | always | the owner's definition, in their words; `?` while open |
| `Owner` | always | who decides what the term means; `TBD` while open |
| `Status` | always | exactly `defined` or `open` |
| `Source` | defined | who gave the definition, and when; always a person, never an agent |
| `Aliases` | optional | other names, comma-separated, including plurals and the requester's own word in their language (*Tageslimit*) |
| `Question` | open | what has to be answered first; a defined term has none |

No other fields. A field may continue on lines indented by two spaces.
The definition comes from the owner or someone they name, never from
the authoring agent.

## 9. Open questions

One bullet per open item, tagged with where it belongs, and with who
can answer it if known:

```
- [R2] When a request is within the limit, what happens next? (ask: Maria Keller)
- [Glossary: daily limit] What does the daily limit apply to? (ask: Card Products)
- [R1, R2] Which of the two rules applies first? (ask: Maria Keller)
- [Intent] Why is this needed now? (ask: requester)
```

One open question per item, phrased openly: no options to choose from,
no candidate answers. Tags: `[R<n>]` (or several: `[R1, R2]`),
`[Glossary: <term>]`, `[Intent]`, `[Today]`, `[Scope]`, `[General]`.
Every rule, term and section with an open item has at least one
question here, so this section is the complete list of what is not yet
known. An open term's question is listed here as well as in the
glossary.

## 10. Notes for whoever builds this

Always this exact text, unchanged:

```markdown
## Notes for whoever builds this

- Intent, Today and Scope explain the request. The rules decide. Where the rules are silent, do not fill the gap from the intent: ask.
- Every `?`, every open glossary term and every entry under Open questions is undecided. Treat each as a stop, not as something to guess.
- Rules with no open items can be built now. A rule with open items cannot, and neither can anything that depends on it.
- `<name>` is a value that varies; `[term]` is defined in the Glossary. Each rule's examples are cases the requester confirmed.
- "Provided" is the condition that decides between "Then" and "Otherwise". Both outcomes are required behaviour.
```

## 11. Status and what "ready" means

- `draft`: open items are allowed. They are the honest list of what is
  not yet known.
- `ready`: no open items anywhere: no `?`, no open term, nothing under
  Open questions, a named owner.

Only the owner can set `ready`; the authoring agent does it only when
they explicitly ask. A request may be handed over as a draft; its open
items then travel with it, visibly.

## 12. The check

```
python3 <skill-dir>/scripts/lint_request.py <request-file.md>
```

(`python` where there is no `python3`.) Each finding is one line,
`line <n>: <level> <code>: <message>`, then a summary:

- **error**: the format is broken. Fix it, but never by deciding
  something about the domain: write `?` instead, and ask. The exit
  status is 1 while an error remains.
- **open**: something nobody has decided yet. It becomes a question for
  the requester, or stays as an open question. Allowed in a draft.

Two findings are hints rather than plain gaps: `relative-word` (a word
like *enough* in a Given, When or Provided) and `provided-without-event`
(a condition that uses no value of the When). Both are settled by
asking, then rewriting the line; never by explaining them away.
