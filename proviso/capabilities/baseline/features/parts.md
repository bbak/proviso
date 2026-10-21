# <a id="baseline.parts"></a> Parts and fingerprints

- Contact: Bruno Baketarić

## Goal

Every element splits into parts that keep their identity wherever the
element moves, and each part gets a fingerprint that changes whenever
its meaning can have changed, and only then.

## Rationale

A part is the smallest unit the element diff reports, so it follows what
readers and agents act on: a guard, an outcome, an Examples row, a
definition. Only whitespace is normalised. Anything more, such as
reading `*a*` and `_a_` as the same, would need the fingerprint to
understand Markdown; a false "changed" costs a look, a false "unchanged"
hides a change.

## <a id="baseline.parts.split"></a> Rule: An element splits into the parts its kinds give

```guarded
When      `manifest` or `diff` reads the [element] <element>
Then      <element> has one [part] for each entry of `part_kinds` it holds, with that entry's key
  And     an [element] below <element> is not a [part] of <element>
```

| element | outcome | why |
| --- | --- | --- |
| `lint.run.exit`, a rule without a Rationale | `heading`, `step When`, `step Provided`, `step Then`, `step Otherwise`, one `row` per Examples row | |
| `lint.run`, a feature with rules | `heading`, `field Contact`, `section Goal`, `section Rationale` | its rules have parts of their own |
| `lint.step`, a glossary entry | `heading`, `field Definition`, `field Contact`, `field Status`, `field Aliases` | |
| `lint.tree-layout`, a parameter with a table | `heading`, `field Configurable`, `field Rationale`, `body` | the prose and the table |
| `lint.cli.lint`, a contract element with members | `heading`, `body` | `json` and `exit-status` are elements of their own |

## <a id="baseline.parts.fingerprint"></a> Rule: A fingerprint ignores whitespace and nothing else

```guarded
When      `manifest` or `diff` takes the [fingerprint] of the [part] <part>
Then      the [fingerprint] is the hash of the text of <part>, each run of spaces, tabs and line feeds in it made one space, none left at its start or end
  And     nothing else in the text counts as the same as anything else
```

| part | outcome | why |
| --- | --- | --- |
| a section's body, and the same body wrapped at other places | the same fingerprint | a reflowed line is no change |
| `Then      lint exits`, and `Then lint exits` | the same fingerprint | the padding after a keyword |
| `*a*`, and `_a_` | different fingerprints | Markdown is not interpreted |
| `- Contact: Bruno`, and `- Contact: bruno` | different fingerprints | case counts |

## <a id="baseline.parts.row-key"></a> Rule: A row is known by its placeholder values

### Rationale

The placeholder values are what the testing agent reads as a case, so
they are what makes a row the same row after an edit. A row whose values
change is another case: one row removed, one added. Lint reports rows
that repeat each other's values, except where a cell is undecided, so
only an undecided row can share its values with another.

```guarded
When      `manifest` or `diff` keys the row <row> of an [Examples table] as a [part]
Provided  no row above <row> in the same table has the same [placeholder] values
Then      the key of <row> is the [fingerprint] of its [placeholder] values, each with its column's name
  And     the order of the columns does not change the key
Otherwise the key of <row> is that [fingerprint], numbered by how many rows above <row> share it
  And     the order of the columns does not change the key
```

| row | provided | outcome | why |
| --- | --- | --- | --- |
| `id` "a", `outcome` "error" | yes | keyed by `id` "a" | |
| `id` "a", `outcome` "nothing reported" | yes | the same key as before | a new outcome is a changed row, not another case |
| `id` "a", after the columns `id` and `outcome` were swapped | yes | the same key as before | |
| `id` "?", below another row with `id` "?" | no | keyed as the second row with that fingerprint | |
