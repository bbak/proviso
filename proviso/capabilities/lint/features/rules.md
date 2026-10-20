# <a id="lint.rules"></a> Rule blocks

- Contact: Bruno Baketarić

## Goal

Every rule holds one rule block in Guarded Gherkin: one event, at most
one guard with its `Otherwise`, and every outcome followed by what else
happens.

## Rationale

The rule block is what the testing agent turns into tests: the guard
into guard tests, each outcome into an outcome test, and each `And` line
into a frame test. A block of the wrong shape leaves the agent to guess
which is which. The checks are strict: a `Then` without an `And` line
is an error, not a question, because its frame test needs that line.
The checks on a block's Given, guard and placeholders take its one
`When` as their reference; a block without exactly one `When` is
reported only for how often its keywords appear.

## <a id="lint.rules.block"></a> Rule: A rule holds exactly one rule block

```guarded
When      lint reads the [rule] <rule>, holding <count> [rule blocks]
Provided  <count> is 1
Then      lint checks that [rule block]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.rule-block-count` at <rule>
  And     lint goes on with the rest of the [spec tree]
```

| rule | count | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint.run.root` | 1 | yes | checked | |
| a rule with only a `gherkin` block | 0 | no | error | a block counts only when fenced as `guarded` |
| a rule with two `guarded` blocks | 2 | no | error | variation belongs in Examples rows, not in blocks |

## <a id="lint.rules.keyword"></a> Rule: Every line of a rule block starts with a keyword

### Rationale

Gherkin's `But` is not a keyword here: an exception to an outcome is a
guard, and belongs in `Provided` and `Otherwise`.

```guarded
When      lint reads the non-blank line <line> of a [rule block]
Provided  <line> starts with a [keyword], after any spaces or tabs
Then      lint reads <line> as part of a [step]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.keyword-unknown` at <line>
  And     lint goes on with the rest of the [spec tree]
```

| line | provided | outcome | why |
| --- | --- | --- | --- |
| `Given     the spec tree has …` | yes | read | |
| `  And     nothing else changes` | yes | read | `And` may be indented |
| `Otherwise → lint.layout.file` | yes | read | a hand-over |
| `But       nothing is reported` | no | error | not a keyword here |
| `given     the spec tree has …` | no | error | keywords are matched exactly |
| `          lint goes on` | no | error | a step continued on a new line |

## <a id="lint.rules.order"></a> Rule: Keywords follow their order

```guarded
When      lint reads the [keywords] <keywords> of a [rule block], in the order they appear
Provided  <keywords> follow the order Given, When, Provided, Then, Otherwise
  And     no And comes first
Then      lint reports nothing about their order
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.keyword-order` at the first [keyword] out of order
  And     lint goes on with the rest of the [spec tree]
```

| keywords | provided | outcome | why |
| --- | --- | --- | --- |
| When, Then, And | yes | nothing reported | the shortest rule |
| Given, Given, When, Provided, And, Then, And, Otherwise, And | yes | nothing reported | several Givens, a guard in two parts |
| When, Given, Then, And | no (order) | error at Given | |
| Given, When, Then, And, Provided | no (order) | error at Provided | the guard after the outcome |
| And, When, Then, And | no (And first) | error at the first And | |

## <a id="lint.rules.count"></a> Rule: A rule block has one event and one outcome

```guarded
When      lint reads the [keywords] <keywords> of a [rule block]
Provided  When and Then each appear exactly once in <keywords>
  And     Provided and Otherwise each appear at most once
Then      lint reports nothing about how often [keywords] appear
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.keyword-count` at the block, naming each [keyword]
  And     lint goes on with the rest of the [spec tree]
```

| keywords | provided | outcome | why |
| --- | --- | --- | --- |
| When, Then, And | yes | nothing reported | |
| Given, When, Provided, Then, And, Otherwise, And | yes | nothing reported | |
| Given, Then, And | no (When and Then once) | error, naming When | no event |
| When, When, Then, And | no (When and Then once) | error, naming When | two events make two rules |
| When, Provided, Provided, Then, And, Otherwise, And | no (Provided and Otherwise at most once) | error, naming Provided | parts of one condition are `And` lines |

## <a id="lint.rules.guard"></a> Rule: A guard comes with its Otherwise

```guarded
When      lint reads the [keywords] <keywords> of a [rule block]
Provided  <keywords> hold both Provided and Otherwise, or neither
Then      lint reports nothing about the [guard]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.guard-incomplete` at the block
  And     lint goes on with the rest of the [spec tree]
```

| keywords | provided | outcome | why |
| --- | --- | --- | --- |
| When, Provided, Then, And, Otherwise, And | yes | nothing reported | |
| When, Then, And | yes | nothing reported | a rule that always applies |
| When, Provided, Then, And | no | error | what happens otherwise is undecided: write `Otherwise ?` with `And ?` |
| When, Then, And, Otherwise, And | no | error | an Otherwise of nothing |

## <a id="lint.rules.hand-over"></a> Rule: A hand-over has no And lines

### Rationale

The rule receiving a hand-over states what happens. An `And` line after
it would put an outcome where no outcome is.

```guarded
When      lint reads an outcome [step] <step> of a [rule block], with the And lines <ands>
Provided  <step> is a [hand-over]
Then      lint reports the [error] `findings.hand-over-and` at each line of <ands>
  And     lint goes on with the rest of the [spec tree]
Otherwise → lint.rules.and
```

| step | ands | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `Otherwise → lint.layout.file` | none | yes | nothing reported | |
| `Otherwise → lint.layout.file` | `And nothing else changes` | yes | error at that line | |
| `Otherwise lint reports …` | `And lint goes on …` | no | handed over | not a hand-over |

## <a id="lint.rules.and"></a> Rule: An outcome says what else happens

```guarded
When      lint reads an outcome [step] <step> of a [rule block], with the And lines <ands>
Provided  <ands> hold at least one line
Then      lint reports nothing about what else happens
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.and-missing` at <step>
  And     lint goes on with the rest of the [spec tree]
```

| step | ands | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `Then lint reads <file>` | `And lint goes on …` | yes | nothing reported | |
| `Then lint exits with …` | `And nothing else changes` | yes | nothing reported | "nothing else" is a statement too |
| `Otherwise lint reports …` | `And ?` | yes | nothing reported | undecided, reported as an ask elsewhere |
| `Then lint reads <file>` | none | no | error | |
| `Otherwise lint reports …` | none | no | error | |

## <a id="lint.rules.given"></a> Rule: A Given does not refer to the event

```guarded
When      lint reads the Given [step] <given> of a [rule block] whose When [step] uses the [placeholders] <introduced>
Provided  <given> uses none of <introduced>
Then      lint reports nothing about <given>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.given-refers-to-event` at <given>
  And     lint goes on with the rest of the [spec tree]
```

| given | introduced | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `Given the spec tree has … <others>` | `<id>` | yes | nothing reported | |
| `Given lint has found the spec root` | `<unreadable>` | yes | nothing reported | |
| `Given lint has read <id>` | `<id>` | no | error | a condition on the event belongs in Provided |

## <a id="lint.rules.provided"></a> Rule: A guard refers to the event

### Rationale

A condition that says nothing about the event holds or fails regardless
of it, so it may be state that belongs in a Given. Whether it is takes
judgement, so lint asks rather than refuses.

```guarded
When      lint reads the Provided [step] <guard>, with its And lines, of a [rule block] whose When [step] uses the [placeholders] <introduced>
Provided  <guard> uses at least one of <introduced>
Then      lint reports nothing about <guard>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [ask] `findings.provided-without-event` at <guard>
  And     lint goes on with the rest of the [spec tree]
```

| guard | introduced | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `Provided <count> is 1` | `<rule>`, `<count>` | yes | nothing reported | |
| `Provided the name of <entry> …`, `And <entry> is not …` | `<entry>` | yes | nothing reported | |
| `Provided the tree is clean` | `<file>` | no | ask | perhaps state for a Given |

## <a id="lint.rules.bound"></a> Rule: Every placeholder is introduced

```guarded
When      lint reads the [placeholder] <placeholder> in a [rule block] whose When and Given [steps] use the [placeholders] <introduced>
Provided  <placeholder> is among <introduced>
Then      lint reports nothing about <placeholder>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.placeholder-unbound` at <placeholder>
  And     lint goes on with the rest of the [spec tree]
```

| placeholder | introduced | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `<id>` | `<id>` | yes | nothing reported | |
| `<others>` | `<others>`, `<id>` | yes | nothing reported | introduced by a Given |
| `<file>` | `<heading>` | no | error | used in an outcome, introduced nowhere |
