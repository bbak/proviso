# Parameters

## <a id="upkeep.rehome-references"></a> rehome_references

- Configurable: no
- Rationale: these are the places where the notation itself marks an
  identifier as one; anywhere else an identifier may be data, such as
  the Examples rows of a rule about identifiers, and only a person can
  tell.

The places where `rehome` renames an identifier. Nothing else is
rewritten: an identifier in plain text, in an Examples cell, in a code
span or in a fenced block stays as it is.

| reference | the identifier it holds | such as |
| --- | --- | --- |
| an anchor | its `id` | `<a id="baseline.impact.events"></a>` |
| a hand-over | the identifier after `Otherwise →` | `Otherwise → lint.run.exit` |
| a link, outside code spans and fenced blocks | the fragment of its destination, after `#` | `(../../references.md#system.ref.commonmark)` |
| a question in a `questions.md` | the identifier in brackets that starts its list item | `- [baseline.dependencies] Should …` |
