# <a id="system"></a> Proviso

## Goal

Proviso keeps one specification of a system true across many
amendments. It checks a spec tree written in the Proviso Model, derives
from each amendment what changed and what it may affect, and assembles
what the business approves and what the agents build from.

## Rationale

Specifications fail in two ways. Some decisions nobody made: a library
default filled the gap. Others were made elsewhere and never arrived: they
stayed in an archived design, a test or the code. A format can make
someone decide; it cannot make the decision arrive. Making it arrive needs
mechanics nobody can perform reliably by hand: resolving terms, finding
the rules an amendment may affect, telling a changed meaning from a reflowed
line. Proviso is those mechanics.

## Scope

- Checking a spec tree against the model's notation.
- A baseline of element fingerprints, and the element diff against it.
- Impact sets for an amendment.
- Context packs for the testing and coding agents.
- Acceptance checks.
- Installing and updating the files a repository needs beside its code:
  the start of a spec tree, its settings, the agent definitions and the
  skills.
- The authoring skill, with which business requesters write checkable
  requests; distributed on its own, since they never see the tree.

## Out of scope

- Reading, building or running the specified system's code.
- Deciding what a rule means: the business decides, Proviso only checks.
- Driving the agents: the agents call Proviso, not the other way round.
- Adapters for native contract formats, such as OpenAPI (deferred).
- Traceability of requests, decisions and approvals (deferred).
- Systems whose tree spans several repositories (deferred).

## Audience

- Engineers who bring requests into a tree and accept the amendments.
- Agents, through skills, that call the command line for every
  mechanical step.
- Business requesters and owners read what Proviso derives, never the
  tree or the command line.

## Design boundary

- An agent host that runs the skills, Claude Code first.
- Git, which isolates amendments on branches and keeps the former text
  of the tree.
- The specified system's repository, which holds the tree beside its
  code.
