# Glossary

## <a id="baseline.part"></a> part

- Definition: A piece of an element with a fingerprint of its own: its
  heading, a field, a section, a step, an Examples row, or everything
  else it holds. `part_kinds` lists them.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: parts

## <a id="baseline.fingerprint"></a> fingerprint

- Definition: The hash of a part's text, with only its whitespace
  normalised, by which the baseline recognises a part as unchanged.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: fingerprints

## <a id="baseline.touched-element"></a> touched element

- Definition: An element the element diff lists: added, changed, removed
  or changed through a dependency.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: touched elements

## <a id="baseline.candidate"></a> candidate

- Definition: A rule or invariant the element diff does not list, but
  which an amendment may affect, found through at least one of the
  candidate sources `change_effects` names. Its judgement decides
  whether it is changed after all or stays as it is.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: candidates

## <a id="baseline.event-head"></a> event head

- Definition: The text of a rule's `When` step up to its first
  placeholder, or the whole step if it has none, read as words: every
  character that is not a letter, a mark or a digit separates words,
  case is ignored, and the articles "the", "a" and "an" are left out,
  such as "lint reads row". Two rules with the same event head react to
  the same event.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: event heads

## <a id="baseline.former-text"></a> former text

- Definition: The text an element or a part had in the baseline, read
  from git at the revision that wrote the manifest.
- Contact: Bruno Baketarić
- Status: defined
