# Amendment

## <a id="system.amendment-note"></a> amendment_note

The file `.amendment.md` in the spec root: what the amendment on its
branch deliberately leaves out, for the agents that build it. Written
by the engineers or a skill, read by `pack`, looked for by `verify`,
removed by `manifest`.

- **It is committed on the amendment's branch,** like `coherence_file`,
  so that it reaches every round of the build, and CI.
- **Its name starts with a dot,** so lint skips it: it is no element,
  and the manifest does not record it.
- **It is Markdown in UTF-8:** a list, one item for each thing the
  amendment leaves out, and no headings, since `pack_text` carries it
  under a heading of its own. That binds whoever writes it: Proviso
  does not read what it says, and does not check it, beyond reading it
  as UTF-8.
