# <a id="upkeep"></a> Upkeep

## Goal

Upkeep keeps a spec tree in order where amendments leave it out of order
without changing what it says: it gives the elements that moved, or
whose identifiers name no place, the identifiers of their place,
together with every reference to them, the mentions of them the engineer
picks, and the manifest, so that an identifier says where its element
stands.

## Rationale

Amendments move elements, such as a term lifted to the system or a
feature regrouped into another capability, and keep their identifiers,
so that a move is cheap and its check small. Over time identifiers stop
saying where their elements stand, and readers, test names and links
follow the wrong place. Upkeep sets this right in a step of its own, on
a tree the manifest records, so that nothing it does mixes with what an
amendment means. It renames what the notation marks as an identifier,
and, of the other mentions, the identifier mentions the engineer
picks, since only a person can tell a reference from data; the code beside the
tree, and every other mention of an identifier, are left to the
engineer.
