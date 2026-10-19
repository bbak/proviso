# Glossary

## <a id="lint.finding"></a> finding

- Definition: Something lint reports about the spec tree, with where it
  is: an error or an ask.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: findings

## <a id="lint.backtick-reference"></a> backtick reference

- Definition: A name in backticks inside a rule block, referring to a
  contract element or a parameter, such as `` `findings.anchor-misplaced` ``.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: backtick references

## <a id="lint.error"></a> error

- Definition: A finding that the spec tree must not keep: a violation of
  the notation. It cannot be overridden, and lint exits non-zero when it
  reports one.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: errors

## <a id="lint.ask"></a> ask

- Definition: A finding that asks an author to look at a place, such as
  a `?`. It does not change lint's exit status.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: asks

## <a id="lint.heading-level"></a> heading level

- Definition: The number of `#` signs that open a Markdown heading: 1
  for `#`, 2 for `##`, and so on.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: heading levels

## <a id="lint.keyword"></a> keyword

- Definition: One of the words that start the lines of a rule block:
  `Given`, `When`, `Provided`, `Then`, `Otherwise` and `And`, matched
  exactly.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: keywords

## <a id="lint.step"></a> step

- Definition: A line of a rule block that starts with a keyword other
  than `And`, together with the `And` lines that follow it. An outcome
  step is a `Then` or `Otherwise` step.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: steps, outcome step

## <a id="lint.top-level-element"></a> top-level element

- Definition: A contract element that is not a member of another, such
  as `customer` for `customer.id`. A reference is qualified from it.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: top-level elements

## <a id="lint.well-formed"></a> well-formed

- Definition: An identifier is well-formed if it consists of one or more
  segments separated by single dots, and each segment starts with a
  lowercase letter, followed by lowercase letters, digits and hyphens.
- Contact: Bruno Baketarić
- Status: defined
