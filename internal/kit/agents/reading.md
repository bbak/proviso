## How to read the pack

The pack is Markdown with up to four headings:

- **System:** the system's Goal, Rationale, Scope, Out of scope and
  Design boundary.
- **Amendment:** every element the amendment touched, each after a line
  saying what happened to it: `Added`, `Changed` (with the parts that
  changed and their former text), `Removed` (with its former text) or
  `Changed through` (with the chain of changed meanings that reaches it),
  and `Not to be built` where it is undecided. This is your work.
- **Frames:** the features and capabilities holding those elements, and
  the invariants and constraints of their levels. They say what it is
  all for and how to build it.
- **Context:** the terms, contract elements, parameters, references and
  rules the touched elements refer to. Read them to understand the
  touched elements; work on them only as far as a touched element needs.

Read the elements this way:

- **Goals give the frame.** The Goals of the system, the capability and
  each feature say what it is all for. Where the rules leave a choice
  open, choose in the Goal's sense, and list the choice. Where a rule
  seems to work against its Goal, follow the rule and list the conflict;
  do not resolve it yourself.
- **Scope sets the limit.** Build or test nothing the system lists as
  Out of scope, and treat what its Design boundary names as given.
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
- Every Examples row is a case: placeholder columns are inputs,
  `outcome` is the expectation, `provided` says whether the guard holds.
- `Otherwise → <id>` hands the decision to the rule with that identifier.
- **An invariant** holds over every event in its Ranges over. **A
  constraint** binds how the system is built, and is checked by build or
  review, not at run time. A changed constraint binds what was built
  before as well.
- **Changed through a dependency:** the element reads as before but
  means something else. **Removed:** the element is gone, and so is what
  was built and tested for it.
- **Every `?` is a stop, and so is an element listed as not to be
  built:** neither build nor test it.
- Rationale explains a choice; it is not behaviour, but do not "fix" or
  contradict what it justifies.
