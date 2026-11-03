# <a id="baseline.pack"></a> The context pack

- Contact: Bruno Baketarić

## Goal

The context pack gives the implementing agents and the checking agent
the element diff, or one element a skill names, and everything they
need to read it, and nothing else of the spec tree, so that each builds or checks from what the tree says
and lists what it does not say.

## Rationale

An agent that browses the tree fills a gap with whatever it happens to
read, and one that misses a term guesses its meaning. The pack follows
what an agent needs to read a listed element, not what changed: a
rule's terms and the standard a term is defined by, the contract
elements and parameters it refers to, the element holding a member and
the term an element represents, and the rules it hands over to. Unlike
the element diff, it does not stop at an unchanged element: an
unchanged term defined by a standard still needs the standard's
version. Other rules sharing a term stay out, as they do for impact. A
link only prose makes, such as an invariant naming a term, is left out,
as it is for the dependencies.

The frames, the Goals and Rationale of the levels with their invariants
and constraints, say what the listed elements are for, and how to fill
the gaps between rules. One pack serves every agent; what differs
between testing, coding and checking is the skill's job. The pack holds what the
tree says about the system; how to read it is Proviso's own, and comes
with the agents, not with the pack.

`pack` derives the element diff as `diff` does, unless it is run with
`pack.element`.

## <a id="baseline.pack.contents"></a> Rule: The system and the element diff are always in the pack

```guarded
When      `pack`, run without `pack.element`, assembles the [context pack] of the [element diff] <diff>
Then      the system is taken for "System"
  And     every [element] <diff> lists is taken for "Amendment", with everything <diff> lists for it
```

| diff | outcome | why |
| --- | --- | --- |
| `lint.run.exit`, its `Then` step changed | `system` and `lint.run.exit` taken | |
| `lint.run.root`, removed | `system` and `lint.run.root` with its former text taken | its tests can be retired |
| none: the tree is as the manifest records it | `system` taken | nothing to build is no failure |

## <a id="baseline.pack.note"></a> Rule: The amendment note is in the pack

### Rationale

The agents read the pack and nothing else, so what the amendment
deliberately leaves out reaches them there, as the engineers wrote it.
`pack` copies the note; what it says is the engineers' concern. It holds
only what the amendment leaves out: whatever decides behaviour belongs
in the tree, where the check sees it. It is a file of its own rather
than the commit message, which is written for people, so that the pack
does not depend on how a message is worded.

```guarded
When      `pack`, run without `pack.element`, assembles the [context pack] of the [spec root] <root>
Provided  <root> holds an `amendment_note` that holds more than whitespace
Then      the text of that `amendment_note` is taken for "Not in this amendment", as the file holds it
  And     everything else stays as it was
Otherwise nothing is taken for "Not in this amendment"
  And     everything else stays as it was
```

| root | provided | outcome | why |
| --- | --- | --- | --- |
| `repo/proviso/`, with `.amendment.md` holding "- refunds" | yes | "- refunds" taken | |
| `repo/proviso/`, without `.amendment.md` | no | nothing taken | most amendments leave nothing out |
| `repo/proviso/`, with an empty `.amendment.md` | no | nothing taken | |
| `repo/proviso/`, with `.amendment.md` holding only blank lines | no | nothing taken | |

## <a id="baseline.pack.note-unreadable"></a> Rule: An amendment note that cannot be read stops the pack

### Rationale

A note `pack` cannot read would leave out of the pack what the
amendment deliberately leaves out, and the agents would build it
without anyone seeing why. A symbolic link is not followed: what it
points to may lie outside the branch.

```guarded
When      `pack`, run without `pack.element`, reads the `amendment_note` <note>
Provided  <note> is a [regular file] `pack` can read, in UTF-8
Then      `pack` goes on with the text of <note>
  And     everything else stays as it was
Otherwise `pack` writes to "stderr" that `amendment_note` cannot be read
  And     `pack` exits with `pack.exit-status` 2
  And     nothing goes to "stdout"
```

| note | provided | outcome | why |
| --- | --- | --- | --- |
| a file holding "- refunds" | yes | read | |
| a folder named `.amendment.md` | no | exit status 2 | |
| a file holding bytes that are not UTF-8 | no | exit status 2 | the JSON could not carry it unchanged |
| a symbolic link to a file holding "- refunds" | no | exit status 2 | a link is not followed |

## <a id="baseline.pack.element"></a> Rule: A named element takes the place of the element diff

### Rationale

A test of an element can need writing when no amendment touches it,
such as when a mutant of the code that carries out a rule survives
every test. Its pack gives the testing agent that element and what it
needs to read it, as the pack of an amendment would. It derives no
element diff, so that nothing of the manifest is read beyond the first
line `lint`, run first, checks, and no git history: the element is
read as it stands. The amendment note belongs to an
amendment, and is not read either, so that an amendment in progress
never stops the pack of an element. One element per pack keeps the option as
simple as `--since`; a skill asks again for the next.

```guarded
When      `pack` assembles the [context pack], run with `pack.element` naming the [identifier] <id>
Provided  the [spec tree] has an [element] with the [identifier] <id>
Then      the system is taken for "System", and that [element] for "Element"
  And     nothing is taken for "Amendment" or "Not in this amendment", and no [element diff] is derived, so that nothing of `manifest_file` is read beyond what `lint` checks, and neither `amendment_note` nor git history
Otherwise `pack` writes to "stderr" that the [spec tree] holds no [element] <id>
  And     `pack` exits with `pack.exit-status` 2
  And     nothing goes to "stdout"
```

| id | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit` | yes | `system` and `lint.run.exit` taken | |
| `lint.run.exit`, with `.amendment.md` holding "- refunds" and an element diff listing `lint.run.root` | yes | `system` and `lint.run.exit` taken, nothing else for "Amendment" or "Not in this amendment" | the amendment in progress stays out |
| `lint.run.exit`, with no `manifest.sum` | yes | `system` and `lint.run.exit` taken | no manifest is read |
| `lint.run.exit`, with a folder named `.amendment.md` | yes | `system` and `lint.run.exit` taken, exit status 0 | the note is not read |
| `lint.run.gone`, an identifier the tree does not hold | no | exit status 2 | |
| `lint.run.root`, removed since the baseline | no | exit status 2 | a removed element is no longer in the tree |

## <a id="baseline.pack.frames"></a> Rule: A listed element brings the frames of its levels

```guarded
When      `pack` takes the [element] <element>, which the [element diff] lists or `pack.element` names
Provided  <element> is named by `pack.element`, or listed as added, changed or changed through a dependency
Then      the feature and the capability that hold <element> or are <element>, where there are such, are taken for "Frames"
  And     so is every invariant of that feature, every invariant and constraint of that capability, and every invariant and constraint of the system
Otherwise nothing is taken for "Frames" because of <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `lint.run.exit`, its `Then` step changed | yes | `lint`, `lint.run`, `lint.run.read-only` and the system's constraints taken | |
| `baseline.former-text`, its Definition changed | yes | `baseline`, `baseline.run.writes` and the system's constraints taken | no feature holds a glossary entry of the capability |
| `system.git`, its Property changed | yes | the system's constraints taken | no feature or capability holds it |
| `lint.run.root`, removed | no | nothing taken because of it | what was built for it is retired, not built |
| `lint.run.exit`, named by `--element` | yes | `lint`, `lint.run`, `lint.run.read-only` and the system's constraints taken | a named element is read as an added one |

## <a id="baseline.pack.chain"></a> Rule: The context follows the reading chain to its end

```guarded
When      `pack` takes the [element] <element> for "Amendment", "Element" or "Context"
Provided  <element> is not listed as removed
Then      every [element] <element> depends on by `dependencies`, and the [contract element] holding <element> if it is a [member], is taken for "Context"
  And     every [element] taken so is taken in turn, and every [element] once
Otherwise nothing is taken for "Context" because of <element>
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `baseline.diff.added`, its `Then` step changed | yes | `baseline.diff.changed` taken, and every term both rules mark | the hand-over, followed to its end |
| a changed rule referring to `findings.anchor-misplaced` | yes | `lint.findings.anchor-misplaced`, `lint.findings` and `lint.finding` taken | the member's holder, and the term the holder represents |
| a changed rule marking a term defined by `system.ref.commonmark` | yes | the term and `system.ref.commonmark` taken | an unchanged standard still needs its version read |
| a changed rule marking [spec root] | yes | `system.spec-root` taken, and none of the other rules marking it | sharing a term is no link |
| `lint.run.root`, removed | no | nothing taken because of it | |
| `baseline.diff.added`, named by `--element` | yes | `baseline.diff.changed` taken, and every term both rules mark | followed as for a listed element |

## <a id="baseline.pack.once"></a> Rule: Every element appears once

### Rationale

An element written twice costs the agents its length again, and two
copies read as two elements. A listed element shows what happened to
it, so it stands under "Amendment" wherever else it was taken; a named
element is the one the agent works on, so it stands under "Element".

```guarded
When      `pack` has taken the [element] <element> for "System", "Frames" or "Context"
Provided  the [element diff] lists <element>, or `pack.element` names it
Then      <element> appears under "Amendment" or "Element" only, whichever it was taken for
  And     everything else stays as it was
Otherwise <element> appears once under the heading it was taken for, however often it was taken
  And     everything else stays as it was
```

| element | provided | outcome | why |
| --- | --- | --- | --- |
| `system`, its Scope changed | yes | under "Amendment" only | the change shows |
| `lint.run`, its Goal changed, holding a changed rule | yes | under "Amendment" only | not repeated under "Frames" |
| `lint.step`, its Definition changed, marked by a changed rule | yes | under "Amendment" only | not repeated under "Context" |
| `system`, named by `--element` | yes | under "Element" only | not repeated under "System" |
| `lint.run`, unchanged, holding two changed rules | no | under "Frames" once | |
| `system.spec-root`, unchanged, marked by two changed rules | no | under "Context" once | |

## <a id="baseline.pack.order"></a> Rule: The elements under each heading are ordered as the tree holds them

### Rationale

The agents read the pack in place of the tree, so it keeps the tree's
order. Where elements stand can carry meaning, such as the members of a
contract element standing in the order of a message's fields, and an
order by identifier would lose it. Paths are compared byte by byte, so
that the order is the same on every platform.

```guarded
When      `pack` writes the [elements] <elements> under the heading <heading> of `pack_text`
Then      they appear in the order the [spec tree] holds them: by the path of their file below the [spec root], written as `report` writes paths, compared byte by byte, then by where they stand in that file
  And     removed [elements] follow them, ordered by [identifier], compared byte by byte, since they no longer stand in the [spec tree]
  And     the [element parts] named for each appear in the order `diff` gives them
  And     the chains of [identifiers] and the lines not decided named for each appear in the order `diff` gives them
```

| elements | heading | outcome | why |
| --- | --- | --- | --- |
| `system.git`, `lint.run`, `lint` | "Frames" | `lint`, `lint.run`, `system.git` | `capabilities/lint/capability.md`, then `capabilities/lint/features/run.md`, then `constraints.md` |
| `lint.run.exit`, `lint.run.root` | "Amendment" | `lint.run.root`, then `lint.run.exit` | as they stand in `run.md`, not by identifier |
| `lint.run.root`, removed, and `lint.run.exit`, changed | "Amendment" | `lint.run.exit`, then `lint.run.root` | a removed element follows |
| `system.cli.diff.exit-status`, `system.cli.diff`, `system.cli.diff.since`, `system.cli.diff.json` | "Context" | `system.cli.diff`, `system.cli.diff.json`, `system.cli.diff.since`, `system.cli.diff.exit-status` | the members as the contract gives them |

## <a id="baseline.pack.output"></a> Rule: The pack is written as text or as JSON

```guarded
Given     `pack` has assembled the [context pack] <pack>
When      `pack` writes it, run with the options <options>
Provided  <options> include `pack.json`
Then      `pack` writes the `context_pack` of <pack> to "stdout"
  And     `pack` exits with `pack.exit-status` 0
Otherwise `pack` writes <pack> to "stdout" as `pack_text`
  And     `pack` exits with `pack.exit-status` 0
```

| pack | options | provided | outcome | why |
| --- | --- | --- | --- | --- |
| for one changed rule | `--json` | yes | a `context_pack`, exit status 0 | |
| for one changed rule | none | no | `pack_text`, exit status 0 | |
| for an empty element diff, without `.amendment.md` | none | no | `pack_text` with "System" only, exit status 0 | |
| for `lint.run.exit`, named by `--element` | `--json` | yes | a `context_pack` with an empty `diff` and a null `note`, exit status 0 | |
