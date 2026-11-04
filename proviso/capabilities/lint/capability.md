# <a id="lint"></a> Lint

## Goal

Lint checks that a spec tree follows the Proviso Model's notation, so
that every later step can read the tree reliably. It reports each
violation as an error and each spot an author should look at as an ask.

## Rationale

Every mechanism built on the tree reads it: the manifest, the element
diff, impact sets and context packs. A misread tree gives them wrong
results without any sign of it, so lint is strict where those mechanisms
depend on the notation and only asks where a person has to weigh it.
Lint reads the Markdown as CommonMark with GitHub's tables, as GitHub
and VS Code render it, so that it reads a tree as its authors see it;
where it departs, a term's Deviations say so. It is also the first
check of Proviso's own tree, which until now has been checked by hand.
