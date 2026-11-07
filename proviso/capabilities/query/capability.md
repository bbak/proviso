# <a id="query"></a> Query

## Goal

Query shows engineers and agents what an element of a spec tree says,
found by the identifier or the name a reader meets in the tree, with
where it stands.

## Rationale

The tree refers to an element by its identifier or by its name, and a
reader who meets one has to find the element it stands for: a term may
be defined in its capability or at the system level, a contract element
in a file of its own. Query finds it as lint resolves it, so that what
query shows is what lint reads, and says where it stands, so that a
person or an agent can go on reading there. It reads the tree as it
stands, also in the middle of an amendment, when lint may still find
errors in it, since that is when the tree is read most. It shows; it
neither searches the text, which any search tool does, nor follows what
an element refers to, which the context pack does.
