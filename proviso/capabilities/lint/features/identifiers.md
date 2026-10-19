# <a id="lint.identifiers"></a> Identifiers

- Contact: Bruno Baketarić

## Goal

Every element is found by exactly one identifier, in one form, standing
where tooling and links expect it.

## Rationale

Identifiers are what everything else refers to: links, hand-overs, the
manifest, impact sets and the logical diff. An anchor in the wrong place
silently drops its element; two elements with one identifier make every
reference to it ambiguous. Whether an identifier is permanent and never
reused can only be checked against a baseline, so that check comes with
the manifest, not here.

## <a id="lint.identifiers.placement"></a> Rule: An anchor starts a heading

```guarded
When      lint reads an [anchor] <anchor>
Provided  <anchor> starts the text of a heading
Then      that heading opens an [element], identified by the anchor's id
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `finding.anchor-misplaced` at <anchor>
  And     lint goes on with the rest of the [spec tree]
```

| anchor | provided | outcome | why |
| --- | --- | --- | --- |
| `## <a id="a.b"></a> Rule: B` | yes | element `a.b` | the prescribed form |
| `## <a id="a.b"></a>Rule: B` | yes | element `a.b` | the space after the anchor is optional |
| `<a id="a.b"></a>` on the line before `## Rule: B` | no | error | the rejected older form |
| `## Rule: B <a id="a.b"></a>` | no | error | inside the heading, but not at its start |
| `<a id="a.b"></a>` within a paragraph | no | error | not in a heading at all |

## <a id="lint.identifiers.form"></a> Rule: An identifier is well-formed

### Rationale

One form avoids identifiers that differ only in case or punctuation,
which links on some hosts would not tell apart. Lowercase and hyphens
match the anchors GitHub generates for headings.

```guarded
When      lint reads the [identifier] <id> of an [element]
Provided  <id> is [well-formed]
Then      lint reports nothing about the form of <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `finding.identifier-malformed` at <id>
  And     lint goes on with the rest of the [spec tree]
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `search.attempts.simplified` | yes | nothing reported | |
| `system` | yes | nothing reported | a single segment |
| `lint.run.root-2` | yes | nothing reported | digits and hyphens after the first letter |
| `Search.attempts` | no | error | uppercase |
| `search..attempts` | no | error | an empty segment |
| `search.attempts.` | no | error | an empty last segment |
| `search.2nd` | no | error | a segment starting with a digit |
| `search.first_attempt` | no | error | an underscore |

## <a id="lint.identifiers.unique"></a> Rule: An identifier is unique

```guarded
Given     the [spec tree] has other [elements] with the identifiers <others>
When      lint reads the [identifier] <id> of an [element]
Provided  <id> is not among <others>
Then      lint reports nothing about the uniqueness of <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `finding.identifier-duplicate` at <id>, naming where else it is used
  And     lint goes on with the rest of the [spec tree]
```

| others | id | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint`, `lint.run` | `lint.identifiers` | yes | nothing reported | |
| `lint`, `lint.run` | `lint.run` | no | error, naming the other `lint.run` | reported at each of the two |
| `system`, `lint` | `system` | no | error, naming the system | a capability identified as `system`: uniqueness keeps the identifier reserved |

## <a id="lint.identifiers.system"></a> Rule: The system is identified as `system`

### Rationale

`system` identifies the system and prefixes the elements born at system
level, so that a prefix always says where an element was born. Once the
system itself always has that identifier, uniqueness keeps every other
element from having it, so reserving it needs no rule of its own.

```guarded
When      lint reads the [identifier] <id> of the system
Provided  <id> is `system`
Then      lint reports nothing about <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `finding.system-identifier` at <id>
  And     lint goes on with the rest of the [spec tree]
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `system` | yes | nothing reported | |
| `proviso` | no | error | the system's name instead of the reserved identifier |
| `system.root` | no | error | a prefix of `system` is not `system` |
