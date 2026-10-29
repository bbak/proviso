# Example request: the cash machine

The file produced by the session in `worked-example.md`, exactly as
delivered (`atm-cash-payout.md`). Use it as a template: every request
file has this structure, in this order.

It shows both states a session can leave behind. R1 is complete. R2 is
an honest draft: it holds only what the requester said; everything else
is `?` and listed under Open questions, one question per item.

`````markdown
---
title: The ATM pays out only what it can make up from its notes, and respects a daily limit
requester: Maria Keller (Cash Operations)
owner: Maria Keller (Cash Operations)
date: 2026-10-03
status: draft
---

# The ATM pays out only what it can make up from its notes, and respects a daily limit

## Intent

Today the ATM sometimes accepts a request and then cannot pay it,
because it has enough cash in total but not the right notes. Customers
are left confused at the machine, and each case becomes a support
ticket. The ATM should refuse such requests up front. A daily
withdrawal limit of $500 should also be enforced.

## Today

- The ATM checks only whether its total cash covers the requested amount.
- There is no daily limit at the ATM.

## Scope

- In: deciding whether the ATM pays out a requested amount
- In: a daily withdrawal limit of $500
- Out: the account balance check (Retail Banking)
- Out: the card check and returning the card (handled for every transaction)

## Rules

### R1 — The ATM pays out only amounts it can make up from the notes it holds

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
| $110 | yes | pays | every note in the machine |
| $120 | no | refuses | more than the notes held |

Decisions:
- 2026-10-03 — Any combination of notes that sums to the amount is fine; which notes the ATM picks is not part of the rule. (Maria Keller)
- 2026-10-03 — "Enough money in the machine" is not the condition: the machine can hold more than the amount and still be unable to pay it. (Maria Keller)

### R2 — The ATM enforces a daily limit of $500

```guarded
Given     <withdrawn-today> has been withdrawn so far today
  And     the [daily limit] is $500
When      the customer asks for <amount>
Provided  <withdrawn-today> plus <amount> ? the [daily limit]
Then      ?
  And     ?
Otherwise the ATM refuses the request
  And     nothing else changes
```

| withdrawn-today | amount | provided | outcome | why |
| --- | --- | --- | --- | --- |
| $0 | $200 | yes | ? | well inside the limit |
| $300 | $200 | ? | ? | reaches the limit exactly |
| $300 | $220 | no | refuses | $20 over the limit |
| $500 | $20 | no | refuses | the limit is used up |

Owner: TBD

## Glossary

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

## Open questions

- [Glossary: daily limit] What does the daily limit apply to? (ask: Maria Keller)
- [Glossary: daily limit] When does the daily limit start over? (ask: Maria Keller)
- [R2] Is a withdrawal that reaches the limit exactly still allowed? (ask: rule owner)
- [R2] Who decides about the daily limit rule? (ask: Maria Keller)
- [R2] When a request is within the limit, what happens next? (ask: rule owner)
- [R2] When a request is within the limit, does anything else happen or change? (ask: rule owner)

## Notes for whoever builds this

- Intent, Today and Scope explain the request. The rules decide. Where the rules are silent, do not fill the gap from the intent: ask.
- Every `?`, every open glossary term and every entry under Open questions is undecided. Treat each as a stop, not as something to guess.
- Rules with no open items can be built now. A rule with open items cannot, and neither can anything that depends on it.
- `<name>` is a value that varies; `[term]` is defined in the Glossary. Each rule's examples are cases the requester confirmed.
- "Provided" is the condition that decides between "Then" and "Otherwise". Both outcomes are required behaviour.
`````
