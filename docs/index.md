# Proviso

Proviso keeps one specification of a system true across many changes.
The specification is a **spec tree**: Markdown files beside the system's
code, written in the **Proviso Model**. Every change to the system starts
as an **amendment**, an edit of that tree on a branch. Proviso checks the
tree, works out what the amendment changed and what it may affect, and
hands the agents that build and check it exactly what they need.

Proviso serves systems in any language. It reads the spec tree, never the
system's code.

## Why

Agents write code faster than anyone can review it. The work that
remains sits at the two ends: saying what should be built, and checking
that it was. At that volume, checking has to be automated as well:
tests written by an agent, independent of the agent that writes the
code. Such tests check something only if the specification tells the
testing agent enough about the system's behaviour: what happens when a
condition does not hold, where its boundaries lie, what each term means.
Where it does not, the agent guesses, and its tests confirm the code
instead of checking it. The Proviso Model is written for that reader:
rules in Guarded Gherkin, with examples on both sides of every
condition, and a glossary that defines their terms. The essay
[Provided, Otherwise](https://medium.com/devops-dev/provided-otherwise-the-missing-guard-in-gherkin-af562aeaec7a)
sets out this argument and the format.

A format alone is not enough, though. Every specification leaves
something unsaid, and a running system does something in every case
anyway. Where the specification is silent, something else decides,
often a tool's default that nobody chose. And
where someone did decide, the decision often never reaches the
specification: it stays with the change that made it, where the next
reader does not look. Over many changes, specification and system drift
apart.

A format can make someone decide. It cannot make the decision arrive.
Making it arrive takes mechanics nobody performs reliably by hand:
finding every rule a change may affect, resolving terms, telling a
changed meaning from a reflowed line. Proviso is those mechanics.

The essay [Decided Elsewhere](https://medium.com/@__bbak/decided-elsewhere-30bcfc233b6e)
sets out the findings behind this.

## How it works

![The path of an amendment: request, specification, build, check, accept](images/pipeline.png)

1. **Request.** A business requester says what should change. The
   authoring skill, which guides a requester to a checkable request, is in
   the works.
2. **Specification.** An engineer brings the request into the spec tree
   with the Integrator, a skill that asks the tough questions, on the
   amendment's own branch: rules in Guarded Gherkin, a glossary,
   contracts, invariants, constraints and parameters. Proviso lints the
   tree and lists what the amendment changed and what it may affect.
   What only the business can answer goes back to it, with a business
   view of what was integrated.
3. **Build.** A testing agent and a coding agent build from the same
   context pack, each in its own worktree, without seeing each other's
   work. Whatever the tree does not decide comes back to the engineer and
   is decided in the tree.
4. **Check.** An independent checking agent, which sees no code, tests or
   reports, judges whether the tree still holds together with the
   amendment. Each finding is resolved by an edit, or dismissed with a
   reason.
5. **Accept.** Once the check covers the amendment, it is merged, and the
   tree becomes the new baseline. CI tells whether the tree on a branch
   is either the recorded baseline or an amendment its check covers.

The agents call Proviso; Proviso never drives them. Four skills,
installed into the repository, take an amendment through steps 2 to 5.
Claude Code is the first agent host they run on.

## Limits

Proviso has been validated only on command-line applications written in
Go. How it works for systems in other languages, or for systems heavy in
user interface or business logic, is untested so far.

The skills and agent definitions were developed with Claude and used
only in Claude Code, primarily with Claude Opus 5.5 as the model. How
they work with other agent hosts or models is untested as well.

## Where to go next

- [Getting started](getting-started.md): install Proviso, set up a
  repository and take a first amendment through.
- [Concepts](concepts.md): the terms the other pages use.
- [Workflow](workflow.md): an amendment from request to acceptance, and
  what the engineer decides on the way.
- [Writing specs](writing-specs.md): the spec tree and its notation.
- [Repository setup](repository-setup.md): CI, the engineers' notes and
  what to commit.
- [Commands](commands.md): every command, its exit statuses and output.

Proviso specifies itself: its own spec tree is
[`proviso/`](https://github.com/bbak/proviso/tree/main/proviso) in its
repository, and it is the exact statement of what every command does.
