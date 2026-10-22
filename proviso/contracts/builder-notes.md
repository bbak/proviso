# Builder notes

## <a id="system.builder-notes"></a> builder_notes

What every context pack says first to the implementing agents, before
anything of the spec tree. These are the notes, in full:

- **Goals give the frame.** The Goals of the system, the capability and
  each feature say what it is all for. Where the rules leave a choice
  open, choose in the Goal's sense, and list the choice. Where a rule
  seems to work against its Goal, follow the rule and list the conflict;
  do not resolve it yourself.
- **Scope sets the limit.** Build or test nothing that the system lists
  as Out of scope, and treat what its Design boundary names as given.
- **The pack is the only source.** If it does not say something, it is
  not decided: do not decide it silently, list it.
- A rule is one `guarded` block. `Given` is state, `When` is the event,
  `Provided` is the guard. **Both** `Then` and `Otherwise` are required
  behaviour, and each `And` line is part of the outcome, including
  "everything else stays as it was".
- `[term]` means exactly its glossary entry: the capability's glossary
  first, then the system's. `` `name` `` is a contract element or
  parameter. `"…"` is a literal value. `<name>` is a placeholder, varied
  by the rule's Examples table.
- Every Examples row is a case: placeholder columns are inputs, `outcome`
  is the expectation, `provided` says whether the guard holds.
- `Otherwise → <id>` hands the decision to the rule with that identifier.
- An invariant holds over every event in its "Ranges over" (tester: test
  it as a property). A constraint is checked by build or review, not at
  run time (coder: it binds how you build). A changed constraint binds
  what was built before as well: review it.
- An element changed through a dependency reads as before but means
  something else: check what was built for it. What was built for a
  removed element is retired.
- **Every `?` is a stop, and so is an element listed as not to be built:
  do not build or test it.**
- Rationale explains a choice; it is not behaviour, but do not "fix" or
  contradict what it justifies.
