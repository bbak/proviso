# Concepts

The terms the other pages use, in the order an engineer meets them. The
exact definitions are in the glossaries of Proviso's own spec tree:
[system-wide](https://github.com/bbak/proviso/blob/main/proviso/glossary.md)
and for the [baseline](https://github.com/bbak/proviso/blob/main/proviso/capabilities/baseline/glossary.md).

## The spec tree

**Spec tree.** The Markdown files that specify one system, kept in the
repository beside its code. Its folder, the **spec root**, is
`proviso/`, marked by the `system.md` in it.

**Levels.** A tree has four: system, capability, feature and rule. Each
groups the one below it. A capability is a folder; a feature is one
file, and the readable unit of the tree.

**Element.** A part of the tree with an **identifier**, its permanent
name, written as an anchor at the start of its heading, such as
`baseline.judge.touched`. Besides the four levels, invariants,
constraints, glossary entries, parameters, references and contract
elements are elements. Identifiers never change, so moving an element
breaks no reference.

**Rule.** The lowest level: one block in **Guarded Gherkin**, Gherkin
with a **guard**. `Provided` states the condition, `Then` what happens
when it holds and `Otherwise` what happens when it does not. Below the
block, an **Examples table** gives concrete cases on both sides of the
guard.

**Glossary entry.** Defines a **term**, which rules mark as `[term]`.
Glossaries exist per level; a term means the same wherever it is
marked.

**Undecided.** A place marked `?`: nothing is specified there yet. The
tree may hold undecided places, but an element with one is not built
until it is decided.

[Writing specs](writing-specs.md) explains each element and the
notation.

## Amendments

**Request.** What the business wants changed, written so that it can be
checked. It is the input to an amendment and never becomes part of the
tree.

**Amendment.** An edit of the spec tree on one branch, usually bringing
in one request. It ends when it is accepted. Everything decided on the
way, while building or checking, lands in the tree as part of it.

**Baseline.** The tree as it was last accepted. Every amendment is
compared with it.

**Manifest.** The file `proviso/manifest.sum`, which records the
baseline as a **fingerprint** of every part of every element: a hash of
its text with whitespace normalised. Proviso writes it; nobody edits it
by hand.

**Element diff.** What an amendment did to the tree, element by element:
added, changed (with the parts that changed), removed, or changed
through a dependency, such as a rule whose term's definition changed.
Only whitespace is normalised: a reflowed paragraph shows nothing, a
changed word does.

**Impact set.** The existing elements an amendment may affect without
having edited them: its **candidates**, rules and invariants found
mechanically, such as rules that react to the same event as a changed
rule. Whether a candidate is really affected takes judgement.

**Context pack.** What an agent receives for an amendment: the element
diff and everything needed to read it, such as the glossary entries,
invariants, constraints and parameters it uses, and `system.md`. Nothing
else of the tree.

**Business view.** What an amendment does, in plain sentences, for the
request's owner, who never reads the tree: rendered after the check by
a fresh agent from the pack, the candidates and the request, with every
open question as a concrete case. It says what the tree says, and is
never stored.

**Amendment note.** `proviso/.amendment.md`, on the amendment's branch:
a list of what this amendment deliberately leaves out, such as a part
of the request kept for later. The pack carries it under "Not in this
amendment", so the agents do not build more than asked. It holds
nothing else: whatever decides behaviour belongs in the tree.

## Agents and checks

**Implementing agents.** The **testing agent** and the **coding agent**.
They build from the same context pack, each in its own worktree, and
never see each other's work, so that the tests are a second, independent
reading of the tree.

**Checking agent.** Runs the **coherence check**: does the tree still
make sense with the amendment? It reads the amendment against its
capabilities and the system level, and sees no code, tests or reports.

**Judgement.** The checking agent's decision that a candidate stays as
it is, with the reason why.

**Finding.** What the check reports besides its judgements: a
contradiction, a gap, an ignored term, invariant or constraint, or a
candidate the mechanics missed. A finding is open until an edit
resolves it or the engineer dismisses it with a reason.

**Coherence file.** `proviso/.coherence.json`, on the amendment's
branch: the check's judgements and findings, with fingerprints of what
it covered. An edit after the check makes the elements it touched
**stale**, and they are checked again.

**Acceptance.** The merge of an amendment into its **target**, the
branch it is accepted into, usually `main`. It requires a coherence
check that covers the amendment's final state, and rewrites the manifest
in the same step, which removes the coherence file and the amendment
note.

**Verify.** `proviso verify` tells whether a branch's tree may be the
baseline as it stands: either it is what the manifest records, with no
coherence file or amendment note left over, or a coherence check covers
it. It judges nothing, so CI can run it on every
branch.

## What Proviso installs

**Kit.** The files Proviso carries in its binary and `proviso init`
installs. Proviso owns them.

**Skill.** Instructions the agent host follows for one step of an
amendment: `proviso-integrate`, `proviso-build`, `proviso-check` and
`proviso-accept`. A skill runs Proviso's commands for every mechanical
step and starts the agents. The skills that edit the tree share a
**writing guide**: how to write it for the agents that build from it.

**Agent definition.** The instructions of one agent, in the form the
agent host reads: its role, what it may read and write, how it reads a
context pack and what it reports.

**Engineers' notes.** The files in `.proviso/` that the engineers own:
`build.md` for the skills, `tester.md` and `coder.md` for the
implementing agents. They hold working practice, such as commands. The
checking agent gets no notes, so that nothing steers its judgement.
