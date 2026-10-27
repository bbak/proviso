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
