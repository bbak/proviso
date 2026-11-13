# <a id="lint.identifiers"></a> Identifiers

- Contact: Bruno Baketarić

## Goal

Every element is found by exactly one identifier, in one form, standing
where tooling and links expect it.

## Rationale

Identifiers are what everything else refers to: links, hand-overs, the
manifest, impact sets and the element diff. An anchor in the wrong place
silently drops its element; two elements with one identifier make every
reference to it ambiguous; an identifier that names no place hides
where its element stands. Lint checks the identifiers as they stand. An
identifier renamed by hand shows in the element diff, as one element
removed and one added. That a removed identifier is never reused is left to the
authors: the manifest forgets a removed element, and within the tree,
unique identifiers are all that references need.

## <a id="lint.identifiers.placement"></a> Rule: An anchor starts a heading

### Rationale

An anchor at the start of its heading is invisible when rendered, and a
link to it from any file lands on the heading itself, on GitHub and in
VS Code, dots included. The heading attribute `{#…}` is not used, since
GitHub does not render it. An anchor on the line before its heading
makes markdownlint report the heading, and a blank line between them
would part the anchor from its element.

```guarded
When      lint reads an [anchor] <anchor>
Provided  <anchor> starts the text of a [heading]
Then      that [heading] opens an [element], with the [anchor]'s id as its [identifier]
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.anchor-misplaced` at <anchor>
  And     lint goes on with the rest of the [spec tree]
```

| anchor | provided | outcome | why |
| --- | --- | --- | --- |
| `## <a id="a.b"></a> Rule: B` | yes | element `a.b` | the prescribed form |
| `## <a id="a.b"></a>Rule: B` | yes | element `a.b` | the space after the anchor is optional |
| `## <a ID="a.b"></a> Rule: B` | yes | element `a.b` | HTML attribute names ignore case |
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
Otherwise lint reports the [error] `findings.identifier-malformed` at <id>
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
Given     the [spec tree] has other [elements] with the [identifiers] <others>
When      lint reads the [identifier] <id> of an [element]
Provided  <id> is not among <others>
Then      lint reports nothing about the uniqueness of <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.identifier-duplicate` at <id>, naming where else it is used
  And     lint goes on with the rest of the [spec tree]
```

| others | id | provided | outcome | why |
| --- | --- | --- | --- | --- |
| `lint`, `lint.run` | `lint.identifiers` | yes | nothing reported | |
| `lint`, `lint.run` | `lint.run` | no | error, naming the other `lint.run` | reported at each of the two |
| `system`, `lint` | `system` | no | error, naming the system | a capability identified as `system`: uniqueness keeps the identifier reserved |

## <a id="lint.identifiers.system"></a> Rule: The system is identified as `system`

### Rationale

`system` identifies the system and prefixes the elements at system
level, as a capability's identifier prefixes the elements in it, so that
a prefix says where an element stands. Once the
system itself always has that identifier, uniqueness keeps every other
element from having it, so reserving it needs no rule of its own.

```guarded
When      lint reads the [identifier] <id> of the system
Provided  <id> is "system"
Then      lint reports nothing about <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.system-identifier` at <id>
  And     lint goes on with the rest of the [spec tree]
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `system` | yes | nothing reported | |
| `proviso` | no | error | the system's name instead of the reserved identifier |
| `system.root` | no | error | a prefix of `system` is not `system` |

## <a id="lint.identifiers.capability"></a> Rule: A capability is identified by one segment

### Rationale

A capability's identifier prefixes the elements in it, and a prefix is
the first segment of an identifier, so a capability identified by
several segments could never prefix its own elements.

```guarded
When      lint reads the [identifier] <id> of a capability
Provided  <id> is a single segment, holding no "."
Then      lint reports nothing about the segments of <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.capability-identifier` at <id>
  And     lint goes on with the rest of the [spec tree]
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `coherence` | yes | nothing reported | |
| `cash.withdrawal` | no | error | two segments |
| `system` | yes | nothing reported | uniqueness reports it, as the system's identifier |

## <a id="lint.identifiers.home"></a> Rule: An identifier starts with a place

### Rationale

The first segment of an identifier says where its element stands, so
an identifier written without one, such as a contract's name alone,
says nowhere, and `rehome` cannot tell its place from its name. Any
place will do: an amendment that moves an element keeps its identifier,
and the identifier names the element's former place until the engineer
rehomes the tree. A capability's identifier of several segments is
already an error of its own, so its first segment counts as a place,
and the one mistake is reported once.

```guarded
When      lint reads the [identifier] <id> of an [element]
Provided  the first segment of <id>, up to its first ".", is "system" or the first segment of the [identifier] of a capability of the [spec tree]
Then      lint reports nothing about the first segment of <id>
  And     lint goes on with the rest of the [spec tree]
Otherwise lint reports the [error] `findings.identifier-unhomed` at <id>
  And     lint goes on with the rest of the [spec tree]
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.root`, a rule of the capability `lint` | yes | nothing reported | |
| `system.read-only`, an invariant of the system | yes | nothing reported | |
| `coherence.impact.events`, a rule of the capability `baseline`, while a capability is identified as `coherence` | yes | nothing reported | a moved element keeps its identifier until a rehome |
| `system` | yes | nothing reported | the system |
| `lint`, the capability so identified | yes | nothing reported | a capability names its own place |
| `cash.withdrawal.pay`, while a capability is identified as `cash.withdrawal` | yes | nothing reported | that capability's identifier is the error |
| `okf-get`, a contract element of the capability `access` | no | error | a contract's name, no place |
| `severity.error`, a member of a contract of the system | no | error | |
| `checks.run`, in the folder `capabilities/checks/`, whose `capability.md` is identified as `coherence`, no capability identified as `checks` | no | error | a place is a capability's identifier, not its folder's name |
| `systems.x`, no capability identified as `systems` | no | error | a segment is compared whole |
