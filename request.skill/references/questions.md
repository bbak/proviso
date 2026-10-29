# Finding holes and asking about them

Two kinds of holes exist in a request:

- **Structural holes** follow from the form: relative words, missing
  alternatives, one-sided examples, ambiguous outcomes. They can be
  found in any domain. The check finds many of them; this file covers
  the rest.
- **Domain holes** follow from what the request is about, such as notes
  in a cash machine that cannot make up every amount. Only someone who
  knows the domain can find them. You can prompt with analogies, but
  the answer comes from the requester or the person they name.

Every hole you leave is a question the engineer has to send back to
the business later, which costs a round trip of days. Every hole you
close now costs one question of a session. So look for all of them;
ration only how many you ask per round.

Ask in priority order. At most four questions per round, each one
sentence with one question mark, with no candidate answers in it.

## Standard questions

Use these wordings, translated if needed, **word for word: add nothing
after them**, no "such as…", no "is X notified, or is Y logged?".
Anything you add becomes the answer.

| when | question |
| --- | --- |
| the request does not say why | "What should this improve compared with today?" |
| a condition has no alternative | "What should happen if [condition, in their words] is not the case?" |
| an outcome has `And ?` | "When [outcome, in their words] happens, does anything else happen or change?" |
| a vague word | "What exactly do you mean by *[their word]*?", or better, a case: "Would [concrete case] count as *[their word]*?" |
| the actor can set the condition themselves | "Can [actor] make [condition] true by themselves?" (next round, if yes: "What is it meant to protect?") |
| the condition says nothing about the event | "What about [the request / the joiner / the order] decides whether [condition]?" |
| "like today" or "as before" | "Apart from [condition], are there other reasons today why [the current outcome] happens?" |
| a boundary the wording leaves open | "If [case exactly at the boundary], what should happen?" |
| a case no rule covers | "If [case], what should happen?" |

---

## Priority 1: ask first

### 1.0 Missing intent (check: open, Intent is `?`)

*Spot:* the requester said what they want but not why.

*Ask:* "What should this improve compared with today?" No candidate
reasons ("bottleneck? slow reimbursement?"): those get nodded through.

*Then:* write the Intent from their answer, in their words, and read it
back. Do not polish it into something they did not say. The builders
use the intent for the many small decisions the rules do not cover, so
an invented intent misleads in ways nobody will trace back.

### 1.1 Relative word in a Given (check: open, `relative-word`)

*Spot:* *enough, sufficient, valid, appropriate, allowed, too many…*
in a Given. Also adjectives that sound absolute but hide a comparison
("the card is valid": valid on what date, for what kind of
transaction?).

*Ask:* "What exactly do you mean by *enough*?" Or, better, a case: "If
the machine holds one $50 note and three $20 notes, is that enough for
$80?" Let the requester supply the comparison; do not suggest it.

*Then:* if the answer relates state to the event, it becomes the
Provided. If it does not (for example "valid" means "not expired and
not blocked"), state it concretely in the Given, or make it a glossary
term with an owner.

### 1.2 Condition hidden in the event (check: open, `relative-word` in a When)

*Spot:* the When describes who acts *and* a condition: "When a user
with insufficient funds requests a withdrawal".

*Ask:* "This mixes what happens (a withdrawal request) with a condition
(insufficient funds). Let me separate them, so we can say what happens
in both cases. What should happen when the funds are sufficient?"

### 1.3 Missing alternative (check: error `guard-incomplete`, but only the requester can fill it)

*Spot:* a Provided without an Otherwise, or a requirement that only
describes the success case.

*Ask:* "What should happen if [the condition] is not the case?" and
then "Besides that, does anything else happen or change?" Do not list
candidates (logging, notifications, counters), and do not offer "or
nothing?": both invite a quick nod.

Ask the same about the success side: "When [the outcome] happens, does
anything else happen or change?" For approvals, payments and anything
audited, the success side is often where the important side effects
are.

Until answered, the draft says `Otherwise ?` and `And ?` under the Then
and under the Otherwise.

### 1.4 Word that needs pinning down

*Spot:* a condition or concept two careful readers could understand
differently, and the rule does not spell it out: *active customer,
premium account, business day, daily limit, composable, duplicate*.

*Ask:* "What exactly counts as an *active* customer here?" If they
hesitate, in a later question: "Who decides what counts as an active
customer?"

If the requester is stuck, do not offer candidate definitions or values
("say, under €1,000?"). Offer a case to judge instead: "Would a report
of €80 for a taxi count as reasonable? And €950 for a hotel?" Their
answers to cases often add up to a definition they can then confirm.

*Then:* record it in the glossary. With a definition and owner from the
requester or the person they name, it is `defined`, with them as
`Source`. Otherwise `open`, with the question. **Do not propose a
definition**, even a "draft" one. Mark every use of the word in the
rule blocks.

### 1.5 Unstated state

*Spot:* the outcome depends on something the rule never mentions.
"Then the account balance is $80" with no starting balance; "Then the
order ships" with no word on stock.

*Ask:* "What has to be true before [the event] happens?"

**"Like today" is unstated state too.** "Everything else goes to the
manager like today" hides whatever today's rules are. *Ask:* "Apart
from [the new condition], are there other reasons today why [the
current outcome] happens?" Any reason named either stays in force
(record it in Today, and check it does not overlap the new rule) or is
changed by this request (then it is a rule).

### 1.6 Condition that says nothing about the event (check: open, `provided-without-event`)

*Spot:* the Provided uses no value from the When: "When a new joiner
starts, provided the pool has [enough] licences". The condition looks
only at the state, so it cannot depend on *who* is joining. That is
often exactly where the domain hole is.

*Ask:* "What about [the joiner / the request] decides whether
[condition]?" If the answer is "nothing, it's just about the pool", the
condition is state and belongs in the Given. If the answer names
something (edition, region, role), that becomes a `<value>` in the When
and part of the condition.

### 1.7 Rules that overlap, contradict each other, or leave a gap

*Spot:* within the request, two rules with the same event whose
conditions can hold at the same time with different outcomes; a
combination of conditions no rule covers; or two outcomes that say
different things for the same case. Check whenever you add or change a
rule, and **always after splitting one sentence into several rules**:
"approve automatically if X, otherwise the manager, except in the
freeze period, then reject" becomes two rules, and unless the second
says "outside the freeze period", both apply during the freeze period
with opposite outcomes.

If the requester stated the precedence ("except", "but never during"),
write it into the rule. If not, ask with a case.

*Then:* never resolve it yourself. If it stays open, record it under
Open questions, tagged with both rules: `[R1, R2]`.

### 1.8 Condition the actor can satisfy alone

*Spot:* the condition is a status, a checkbox or a confirmation
("resolved", "approved", "confirmed", "declared compliant") that the
person or system doing the action can set themselves.

*Ask, once and briefly:* "Can [actor] make [condition] true by
themselves?" If yes, next round: "What is it meant to protect?"

*Then:* if the requester names a different real condition, use that.
If they agree the self-declaration is acceptable, record that as a
decision. Either way it is their call.

---

## Priority 2: ask once the rule's shape is settled

### 2.1 Examples on both sides (check: open, `examples-side-missing`)

*Do:* propose the input values yourself: at least one case where the
condition clearly holds, one where it clearly fails, and the boundary
on each side. Leave every outcome as `?` and ask the requester to fill
them in:

> Here are four cases. What should happen in each?
>
> | amount | what happens? |
> | --- | --- |
> | $90 | ? |
> | $80 | ? |
> | $110 | ? |
> | $120 | ? |

Choosing inputs is structural work; deciding outcomes is domain work.
Surprising answers here are the most valuable result of the whole
session.

### 2.2 Each part of a condition failing on its own (check: open, `examples-part-missing`)

*Do:* propose one case per part in which only that part fails. Ask
whether the outcome is the same in each. If it is not, the rule needs
splitting.

### 2.3 Boundaries

*Spot:* the condition compares with a limit, a date or a count, and the
wording does not say on which side the limit itself falls: "up to
$500", "within 30 days", "after the deadline". The check cannot see
this.

*Ask:* a case exactly at the limit: "If someone has taken $300 today
and asks for $200, which makes exactly $500, what should happen?"

### 2.4 Outcome only one implementation would produce

*Spot:* a Then that names one specific result where several would
satisfy the requester: "dispenses two $20 notes" when any notes summing
to $40 would do; "sorted by date" when the order does not matter.

*Ask:* "Does it have to be exactly this, or is any result fine as long
as [the relation] holds?"

*Then:* if any valid result is fine, write the outcome as a relation,
and record the decision.

---

## Priority 3: advice; mention once, record if deferred

### 3.1 Several conditions with different outcomes (check: error `keyword-count`)

*Do:* split into one rule per condition. Ask in which order the
conditions are checked, if it matters to the requester; record their
answer as a decision, or as an open question.

### 3.2 Nothing explicitly out of scope

*Spot:* `## Scope` has no `Out:` line.

*Ask:* "Is there anything this should explicitly not change or not
cover?" If the answer is no, leave it; never add `Out:` lines yourself.

### 3.3 Implementation detail

*Spot:* screens, buttons, field names, tables, APIs, algorithms.

*Ask:* "Is the [button / screen / table] itself a requirement, or is it
how you picture it working? If it's just a picture, I'll describe what
has to happen instead."

---

## Before you deliver

These are the holes the engineer will look for when bringing the
request into the specification. Each one you leave is a question that
comes back. Go through them once per rule:

- every condition has its alternative, and both outcomes say what else
  happens or changes;
- no word in a condition is left vague, and every term has its owner's
  definition or is open with its question;
- no outcome depends on state no line mentions, including "as today";
- no two rules can apply to the same case with different outcomes, and
  no case of the same event falls under no rule;
- no condition is one the actor can make true by themselves, unless the
  requester decided that is fine;
- the examples show both sides, each part of the condition failing,
  and the boundaries the wording leaves open.

Whatever is still open goes under Open questions, with who can answer
it. An open item listed there costs nobody a round trip; the engineer
sees it at once. Only a hole nobody saw comes back.

## Analogies

When a domain hole seems likely but the requester has not seen it, you
may offer an analogy as a question:

> In quota and licence-pool systems, "available" often turns out to
> mean "available in the right units", not just "enough in total".
> Could that apply to the notes in the machine?

An analogy is a prompt, never an answer. If the requester says no, drop
it.

## Phrasing

- One question per bullet. No compound questions except the "what else
  happens?" follow-up in 1.3.
- Quote their own words back ("You wrote *enough*…").
- No jargon: no *guard, precondition, placeholder, Given/When* in
  questions.
- Never talk about the format, the file or the check. Talk about their
  domain.
- No candidate answers inside a question ("say, €1,000?"). Ask openly,
  or offer a case.
- Say why a question matters only if they ask, or if it is not obvious;
  one short clause at most.
