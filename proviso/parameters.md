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
