# Worked example: the cash machine

A complete session, condensed. The requester is Maria Keller from Cash
Operations. The resulting file is in `example-request.md`.

## What she brings

> We need the ATM to only pay out when it has enough money. Our old
> acceptance criteria:
>
> ```
> Scenario: Account has sufficient funds
>   Given the account balance is $100
>     And the card is valid
>     And the machine contains enough money
>    When the Account Holder requests $20
>    Then the ATM should dispense $20
>     And the account balance should be $80
>     And the card should be returned
> ```

## Read and split

> I see one event here, the customer asking for an amount, and three
> conditions on it: the account balance, the card, and the money in the
> machine. You said the money in the machine is what you need now, so
> I'll start there. Should the balance and the card be part of this
> request too?

She says no: balance checks belong to Retail Banking, and the card is
handled for every transaction. Both go under *Out* in Scope, in her
words.

## Round 1

The agent drafts the whole file with `?` wherever she has not said
something: Intent, Today, the condition, what happens otherwise, and
what else happens after each outcome. It runs the check once, fixes the
format, and replies:

> I see one rule here: the ATM pays out an amount only if it has enough
> money. Is that right?
>
> **The rule:** when a customer asks for an amount, the ATM pays it out
> if it has enough money; what happens otherwise isn't decided yet.
>
> ```guarded
> Given     the ATM holds (not decided yet)
> When      the customer asks for <amount>
> Provided  (not decided yet)
> Then      the ATM pays out <amount>
>   And     (not decided yet)
> Otherwise (not decided yet)
>   And     (not decided yet)
> ```
>
> Say the ATM holds one $50 note and three $20 notes:
>
> | amount | what happens? |
> | --- | --- |
> | $90 | (not decided yet) |
> | $80 | (not decided yet) |
> | $110 | (not decided yet) |
> | $120 | (not decided yet) |
>
> Before I continue, a few questions:
>
> 1. What should this improve compared with today?
> 2. What exactly do you mean by *enough* money?
> 3. What should happen if the ATM doesn't have enough?
> 4. What should happen in each of the four cases in the table?

Note what is *not* in it: no options inside the questions, no remarks
about the draft or the check, no announcement of what comes next round.
The question about what else happens after each outcome waits for
round 2, because four questions is the limit.

Maria: "Today it sometimes takes the request and then can't pay, and
the customer stands there confused. Every one of those is a support
ticket. Enough for the amount the customer wants. If not, refuse,
nothing else happens. $90 pays out. $80… it has $110 in there, so yes.
Wait, no. It can't make $80 out of a fifty and twenties. That's a
refusal. $110 pays, $120 refuses."

The table surfaced the domain hole. *Enough money* was never the
condition; the condition is whether the notes can make up the amount.

## Round 2

> Here's what I've changed: the ATM refuses an amount it can't make up
> from its notes, and nothing else happens then. The goal: refuse up
> front what it can't pay, because each failure confuses a customer and
> creates a ticket. Today it only checks whether its total cash covers
> the amount.
>
> Is that right? And a few more questions:
>
> 1. I've called the condition *composable* for now: how would you
>    define it, in your words?
> 2. With two $20 notes and four $10 notes, a request for $40 can be
>    paid three different ways: does it matter which notes the ATM
>    picks?
> 3. When the ATM pays out, does anything else happen or change?
> 4. Who decides about these rules?

Maria defines *composable* ("some selection of the notes adds up to
exactly the amount"), says the choice of notes doesn't matter, and that
the stock of notes goes down by exactly what was paid out. She decides
about these rules, for cash handling.

The agent writes the outcome as a relation ("pays out notes that sum to
exactly `<amount>`"), records two decisions and the definition with her
as Source, and reads R1 back in one plain sentence. She confirms.

## Round 3: a second rule appears

Maria: "Oh, and there's a daily limit, $500. Anything over it is
refused, nothing else happens."

The agent adds R2 with only what she said. The comparison in the
condition stays open (`<withdrawn-today> plus <amount> ? the [daily
limit]`): "at most" or "less than" would decide for her whether exactly
$500 is allowed. *Daily limit* goes into the glossary as `open`. What
happens within the limit stays `Then ?` with `And ?`. The agent asks:

> 1. What does the daily limit apply to?
> 2. Someone has already taken $300 today and asks for $200, which is
>    exactly $500. What should happen?
> 3. When a request is within the limit, what happens next?
> 4. Who decides about this limit?

Maria knows none of these for sure; Card Products might own the limit.
All four stay open, with who to ask. When the limit starts over is a
further open question, recorded rather than asked as a fifth.

This is the third round, so the agent writes the whole file and offers
it: "Here is the request as it stands. If we continue another time,
bring this file along, and we'll pick up where we left off."

## Closing

Maria wants to stop. The agent runs the check once more, delivers
`atm-cash-payout.md`, and closes:

> **What the request covers:** the ATM refuses amounts it can't make up
> from its notes, and enforces a $500 daily limit.
>
> **Rules**
>
> - R1: pays out only amounts it can make up from its notes — complete
> - R2: daily limit of $500 — still open in places
>
> **Still open**
>
> - What the daily limit applies to, and when it starts over (Maria
>   asks Card Products)
> - Whether a withdrawal that reaches the limit exactly is allowed
> - Who decides about the daily limit
> - What happens when a request is within the limit, and what else
>   happens or changes then
>
> The request is a draft. It can be handed over as it is: whoever
> builds it sees exactly what is still open.

## A week later: the questions come back

Maria returns with the file and a message from the engineer, who has
brought the request into the specification: a summary of what the
system will do, and two questions. One asks what the daily limit
applies to, which is still open in the file; Card Products has
answered meanwhile: per card, from midnight to midnight. The other is
new, found in the specification: customers can also withdraw cash at
the counter with their card. Does that count towards the daily limit?

The agent reads the file and the message, and does not start over. It
folds in Card Products' answer as the definition, with Card Products as
Source, and asks the new question as a case:

> A customer takes $300 at the counter in the morning and asks the ATM
> for $250 in the afternoon. What should happen?

Then it reads back the engineer's summary, rule by rule, and asks
whether that is what she means. Her answers go into the file as
decisions, and the file goes back to the engineer once, complete.
