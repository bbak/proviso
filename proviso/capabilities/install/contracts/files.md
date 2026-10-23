# Installed files

## <a id="install.system-skeleton"></a> system_skeleton

The `system.md` that `init` writes: the system element with its six
sections, each undecided, with a note in brackets saying what belongs
there. Its wording is the kit's.

## <a id="install.gitattributes-file"></a> gitattributes_file

The `.gitattributes` that `init` writes into the spec root, so that git
checks the tree out with line feeds on every platform, as lint
requires, without touching the repository's own settings. Its wording
is the kit's.

## <a id="install.markdownlint-file"></a> markdownlint_file

The `.markdownlint.jsonc` that `init` writes into the spec root, so
that markdownlint accepts the notation: anchors in headings, the same
section under different elements, long Examples rows, rule blocks and
headings, spaces inside code spans, and tabs in code blocks. Its
wording is the kit's.
