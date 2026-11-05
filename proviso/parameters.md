# Parameters

## <a id="system.model-version"></a> model_version

- Value: 0
- Configurable: no
- Rationale: 0 says the notation is not declared stable yet, so trees
  and binaries of version 0 promise each other nothing; 1 follows when
  it is declared stable.

The version of the Proviso Model a binary reads and writes. A whole
number, not a semantic version: every change to the notation breaks an
older binary, since a section it does not know is an error to it.

## <a id="system.model-releases"></a> model_releases

- Configurable: no
- Rationale: a binary knows its own model version only; the first
  release of each model version lets it tell which model an older
  release reads, and the table grows with the model, not with every
  release.

Each row holds from its first release up to the next row's first
release; the last row holds for every later release. Versions are
release versions, without a leading `v`.

| model_version | first release |
| --- | --- |
| 0 | 0.1.0 |

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
