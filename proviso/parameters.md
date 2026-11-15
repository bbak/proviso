# Parameters

## <a id="system.model-version"></a> model_version

- Value: 1
- Configurable: no
- Rationale: a whole number counts the changes of the notation; which
  trees a binary still reads is for `oldest_model_version` to say, not
  the number itself.

The version of the Proviso Model a binary reads and writes. A whole
number, not a semantic version: every change to the notation breaks an
older binary, since a section it does not know is an error to it.

## <a id="system.oldest-model-version"></a> oldest_model_version

- Value: 0
- Configurable: no
- Rationale: a change that only adds to the notation leaves every tree
  as it was, so a binary goes on reading it without a migration; only a
  change that breaks `system.model-compatibility` raises it.

The oldest version of the Proviso Model a binary reads. A binary reads
every model version from it to `model_version`, and writes only
`model_version`.

## <a id="system.model-releases"></a> model_releases

- Configurable: no
- Rationale: a binary knows only the model versions it reads; the
  first release of each model version lets it tell which model an
  older release writes, and the table grows with the model, not with every
  release.

Each row names the model version its releases write, and holds from
its first release up to the next row's first release; the last row holds for every later release. Versions are
release versions, without a leading `v`.

| model_version | first release |
| --- | --- |
| 0 | 0.1.0 |
| 1 | 0.7.0 |

## <a id="system.change-effects"></a> change_effects

- Configurable: no
- Rationale: only a change that can alter behaviour, or the frame in
  which agents fill the gaps between rules, may affect other elements; a
  reworded rationale or a new contact cannot.

What a changed part does beyond its own element. "Dependents" passes the
changed meaning on to every element that depends on it, which the
element diff lists as changed through a dependency; the other entries
are sources of impact candidates. A part of an added or removed element
counts as changed, and so does every part of an element changed through
a dependency, whose meaning changed while its text did not.

| changed part | dependents | candidate sources |
| --- | --- | --- |
| a step or Examples row of a rule | yes | events, ranges |
| the field `Definition`, `Defined by` or `Deviations` of a glossary entry | yes | none |
| the field `Version` or `Link` of a reference | yes | none |
| any part of a contract element, and any part of a parameter but its field `Rationale` | yes | none |
| the section `Property` or `Ranges over` of an invariant or a constraint | no | scope |
| any section of the system | no | scope |
| the section `Goal` of a capability or a feature | no | scope |
| any other part, such as a field `Contact` or a section `Rationale` | no | none |

## <a id="system.element-dependencies"></a> dependencies

- Configurable: no
- Rationale: these are the links the notation makes explicit; a link
  only prose makes, such as a definition naming another term without a
  mark, is left to the coherence check. A contract element depends on
  each term it represents, never the reverse: its shape leaves the
  term's meaning as it is, and a rule that relies on the shape refers to
  the element itself. A rule depends on every member of a contract
  element it refers to, a member added later included, and one removed
  by an amendment, since a changed member can affect any such rule:
  coarse on purpose, at the cost of rules listed that only the added or
  removed member touches. A contract element that builds on another,
  such as a member made of another contract's members and one more,
  depends on it as a rule does on what it refers to, so that a change of
  the other reaches every rule referring to the first; only the field
  `Builds on` says so, as prose names elements in passing too. A
  glossary entry depends on the terms its
  definition or deviations mark, and two entries may each mark the
  other, as a glossary's definitions often lean on each other; an entry
  that marks its own term depends on nothing through that mark.

Which element depends on which: when the meaning of the second changes,
so does the meaning of the first.

| element | depends on |
| --- | --- |
| a rule | every glossary entry whose term its rule block marks |
| a rule | every contract element and parameter its rule block refers to, and every member of such a contract element, one removed from it included |
| a rule | the rule its hand-over names |
| a contract element | every glossary entry whose term its field `Represents` marks |
| a contract element | every contract element and parameter its field `Builds on` refers to, and every member of such a contract element, one removed from it included |
| a glossary entry | the reference its field `Defined by` links to |
| a glossary entry | every other glossary entry whose term its field `Definition` or `Deviations` marks |

## <a id="system.tree-layout"></a> tree_layout

- Configurable: no
- Rationale: the Proviso Model fixes where each type of element stands,
  so that tooling and readers find it without a config file.

"At any level" means in the spec root or in a capability's folder. The
names are matched exactly, case included, on every platform. A
heading whose text starts with `Rule:` or `Invariant:` gives an element
type only in a feature file; anywhere else it gives none, because a
prefix outside a file that mixes types marks an element in the wrong
place.

| path | heading | element type |
| --- | --- | --- |
| `system.md` | level 1, with an anchor | system |
| `capabilities/<capability>/capability.md` | level 1, with an anchor | capability |
| `capabilities/<capability>/features/<feature>.md` | level 1, with an anchor | feature |
| `capabilities/<capability>/features/<feature>.md` | level 2, with an anchor, text starting `Rule:` | rule |
| `capabilities/<capability>/features/<feature>.md` | level 2, with an anchor, text starting `Invariant:` | invariant |
| `glossary.md`, at any level | level 2, with an anchor | glossary entry |
| `parameters.md`, at any level | level 2, with an anchor | parameter |
| `references.md`, at any level | level 2, with an anchor | reference |
| `invariants.md`, at any level | level 2, with an anchor | invariant |
| `constraints.md`, at any level | level 2, with an anchor | constraint |
| `contracts/<name>.md`, at any level | level 2 or deeper, with an anchor | contract element |
| `questions.md`, at any level | none | none |
| `.markdownlint.jsonc`, in the spec root | none | none |
| `manifest.sum`, in the spec root | none | none |

## <a id="system.name-comparison"></a> name_comparison

- Configurable: no
- Rationale: one comparison for every name and alias of a glossary
  entry, every term mark and every match in text, so that a mark
  resolves wherever plain text would match, and the reverse, and
  `show` finds a term by every name a mark of it may use.

Two names are the same when they are equal character by character
after each step below, in this order. A name of a glossary entry, an
alias, a term mark and a run of words in unmarked text are compared
alike; a place reported in text is where the match stands as written.

White space at either end does not count, since it cannot be seen and
the reader of headings and fields already drops it. Of the white space
within a name, only spaces and tabs join words, so that no other white
space makes two words one name. Unicode's simple lower-case mapping,
character by character, and not full case folding, which makes "ß" the
same as "ss" and so changes a word's length while matching. The
composed form, as for paths, since the same word is stored composed or
decomposed depending on the editor; it comes last, since lower case can
leave a name that is no longer composed.

| step | effect |
| --- | --- |
| 1 | every character with Unicode's White_Space property at either end, a non-breaking space included, is dropped; a zero-width space has no such property and stays |
| 2 | each run of spaces and tabs within the name, a line break within a field included, becomes one space |
| 3 | each character becomes its simple lower-case form |
| 4 | the name is brought into Unicode's composed form, NFC |

## <a id="system.command-writes"></a> command_writes

- Configurable: no
- Rationale: a command that changed more than it is for would mix a
  change into a step that must not make it; one row per command says
  what each may change, so that a new command has one row to add.

What each command may change under the spec root; every other file and
folder stays as it was.

| command | may change |
| --- | --- |
| `help` | nothing |
| `version` | nothing |
| `lint` | nothing |
| `show` | nothing |
| `diff` | nothing |
| `pack` | nothing |
| `impact` | nothing |
| `verify` | nothing |
| `manifest` | `manifest.sum`, and `.coherence.json` and `.amendment.md`, each with everything in it, after a run that wrote `manifest.sum` |
| `judge` | `.coherence.json`, after a run that exits with 0 |
| `rehome` | after a run with `rehome.write` that exits with 0: `manifest.sum`, and the places of the spec tree its rules rename, nothing else of their files |
