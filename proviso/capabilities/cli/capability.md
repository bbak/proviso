# <a id="cli"></a> Command line

## Goal

Every command of Proviso can be found from the command line itself, and
the command line refuses what it does not know, saying where to look.

## Rationale

Engineers call Proviso by hand now and then, and skills call it on
every step; neither should need the documentation to find a command or
its options, nor guess at a mistyped one. One page names everything,
since every command and its options fit on one page, and a mistake
costs one line saying what was wrong and where the page is, rather than
the whole page. The page's wording is for people; what it names, where
it goes and the exit status are specified. The command line is read by
hand, without a library for it, as `system.dependencies` asks: such a
library brings its own help, its own messages and its own exit
statuses, decisions nobody here made.
