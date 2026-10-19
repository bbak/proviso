# <a id="lint.examples"></a> Examples tables

- Contact: Bruno Baketarić

## Goal

Every rule has an Examples table that gives each placeholder a column,
states each case's outcome, and, for a guarded rule, shows the guard
holding and failing.

## Rationale

The testing agent derives its cases from the rows: the placeholder
values are its inputs, the `outcome` column its expectation, and the
`provided` column the guard's truth value, so it can write guard tests
as well as outcome tests. A missing column or side leaves a test to the
agent's interpretation. Whether the rows also cover the boundaries takes
judgement, and is left to review.

## <a id="lint.examples.present"></a> Rule: A rule has an Examples table

```guarded
When      lint has read the [rule] <rule>
Provided  <rule> holds a table after its [rule block]
Then      lint reads that table as the [rule]'s [Examples table]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-missing` at <rule>
  And     lint goes on with the rest of the [spec tree]
```

| rule | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.root` | yes | read | |
| a rule with a block and no table | no | error | also for a rule without a guard |
| a rule with a table only before its block | no | error | the table follows the block |

## <a id="lint.examples.reserved"></a> Rule: A placeholder does not take a reserved column's name

```guarded
When      lint reads the [placeholder] <placeholder> of a [rule block]
Provided  <placeholder> is none of "<provided>", "<outcome>" and "<why>"
Then      lint reports nothing about the name of <placeholder>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.placeholder-reserved` at <placeholder>
  And     lint goes on with the rest of the [spec tree]
```

| placeholder | provided | outcome | why |
| --- | --- | --- | --- |
| `<guard>` | yes | nothing reported | |
| `<provided>` | no | error | its column would collide with the guard's truth value |
| `<outcome>` | no | error | |

## <a id="lint.examples.columns"></a> Rule: Every placeholder and the outcome have a column

```guarded
When      lint reads the columns <columns> of the [Examples table] of a [rule] with the [placeholders] <placeholders>
Provided  <columns> hold a column named after each of <placeholders>
  And     <columns> hold "outcome"
Then      lint reports nothing about missing columns
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-column-missing` at the table, naming each missing column
  And     lint goes on with the rest of the [spec tree]
```

| columns | placeholders | provided | outcome | why |
| --- | --- | --- | --- | --- |
| id, provided, outcome, why | `<id>` | yes | nothing reported | |
| others, id, provided, outcome, why | `<others>`, `<id>` | yes | nothing reported | a Given's placeholder too |
| provided, outcome, why | `<id>` | no (a column per placeholder) | error, naming id | |
| findings, as found, outcome | `<findings>` | no (a column per placeholder) | error, naming findings | a column is named exactly as its placeholder |
| id, provided, why | `<id>` | no (outcome) | error, naming outcome | |

## <a id="lint.examples.provided-column"></a> Rule: A guarded rule, and only a guarded rule, has a provided column

```guarded
When      lint reads the columns <columns> of the [Examples table] of a [rule]; <guarded> says whether it has a [guard]
Provided  <columns> hold "provided" exactly when <guarded> is "yes"
Then      lint reports nothing about the "provided" column
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.provided-column` at the table
  And     lint goes on with the rest of the [spec tree]
```

| columns | guarded | provided | outcome | why |
| --- | --- | --- | --- | --- |
| id, provided, outcome | yes | yes | nothing reported | |
| findings, outcome | no | yes | nothing reported | a rule that always applies |
| id, outcome | yes | no | error | the guard's truth value is missing |
| findings, provided, outcome | no | no | error | there is no guard to be true or false |

## <a id="lint.examples.unknown-column"></a> Rule: Every column has a purpose

### Rationale

A column that is neither a placeholder nor one of the three reserved
ones holds information the testing agent cannot place: an input it is
not told about, or an outcome under another name.

```guarded
When      lint reads the column <column> of the [Examples table] of a [rule] with the [placeholders] <placeholders>
Provided  <column> is named after one of <placeholders>, or is "provided", "outcome" or "why"
Then      lint reports nothing about <column>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-column-unknown` at <column>
  And     lint goes on with the rest of the [spec tree]
```

| column | placeholders | provided | outcome | why |
| --- | --- | --- | --- | --- |
| id | `<id>` | yes | nothing reported | |
| why | `<id>` | yes | nothing reported | |
| expected | `<id>` | no | error | an outcome under another name |
| notes | `<id>` | no | error | |

## <a id="lint.examples.rows"></a> Rule: An Examples table has rows

```guarded
When      lint reads the [Examples table] of a [rule], with <count> rows
Provided  <count> is at least 1
Then      lint checks each row
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-empty` at the table
  And     lint goes on with the rest of the [spec tree]
```

| count | provided | outcome | why |
| --- | --- | --- | --- |
| 1 | yes | rows checked | the least a rule without a guard needs |
| 0 | no | error | a header alone |

## <a id="lint.examples.cells"></a> Rule: A row fills every column

```guarded
When      lint reads the row <row> of an [Examples table] with the columns <columns>
Provided  <row> has one cell per column of <columns>
  And     every cell of <row> other than "why" holds a value
Then      lint reports nothing about the cells of <row>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-row-incomplete` at <row>
  And     lint goes on with the rest of the [spec tree]
```

| row | columns | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `system`, yes, nothing reported, (empty) | id, provided, outcome, why | yes | nothing reported | `why` may stay empty |
| `system`, yes, `?`, (empty) | id, provided, outcome, why | yes | nothing reported | undecided, reported as an ask elsewhere |
| `system`, yes, nothing reported | id, provided, outcome, why | no (one cell per column) | error | a cell short |
| `system`, yes, (empty), (empty) | id, provided, outcome, why | no (a value in every cell) | error | an outcome left empty: write `?` |

## <a id="lint.examples.provided-value"></a> Rule: A provided cell says whether the guard holds

```guarded
When      lint reads the "provided" cell <value> of a row
Provided  <value> is "yes", "no", "no (<part>)" or "?"
Then      lint reads it as the [guard]'s truth value for that row
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.provided-value-unknown` at <value>
  And     lint goes on with the rest of the [spec tree]
```

| value | provided | outcome | why |
| --- | --- | --- | --- |
| yes | yes | holds | |
| no (level 1) | yes | fails, in the named part | |
| `?` | yes | undecided | reported as an ask elsewhere |
| true | no | error | |
| Yes | no | error | values are matched exactly |

## <a id="lint.examples.sides"></a> Rule: The guard is shown holding and failing

```guarded
When      lint reads the "provided" column <values> of an [Examples table]
Provided  <values> hold at least one "yes"
  And     <values> hold at least one "no" or "no (<part>)"
Then      lint reports nothing about the sides of the [guard]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-side-missing` at the table, naming the missing side
  And     lint goes on with the rest of the [spec tree]
```

| values | provided | outcome | why |
| --- | --- | --- | --- |
| yes, no | yes | nothing reported | |
| yes, yes, no (level 1) | yes | nothing reported | |
| yes, yes | no (a no) | error, naming the failing side | |
| no, no | no (a yes) | error, naming the holding side | |
| yes, ? | no (a no) | error, naming the failing side | an undecided row shows neither side |

## <a id="lint.examples.parts"></a> Rule: Each part of a guard is shown failing

### Rationale

A guard in several parts can fail in each of them. A row failing two
parts at once does not show which one decides, so each part needs a row
of its own.

```guarded
When      lint reads the "provided" column <values> of the [Examples table] of a [rule] whose [guard] has <parts> parts
Provided  <values> hold at least <parts> rows that are "no" or "no (<part>)"
Then      lint reports nothing about the parts of the [guard]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.examples-part-missing` at the table
  And     lint goes on with the rest of the [spec tree]
```

| values | parts | provided | outcome | why |
| --- | --- | --- | --- | --- |
| yes, no | 1 | yes | nothing reported | |
| yes, no (level 1), no (no anchor) | 2 | yes | nothing reported | |
| yes, no (level 1) | 2 | no | error | the second part is never shown failing |
