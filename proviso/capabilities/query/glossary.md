# Glossary

## <a id="query.element-parent"></a> element parent

- Definition: The element another element belongs to. Every element
  of a feature file but the feature, such as a rule or an invariant,
  belongs to the feature; a feature belongs to its capability; a
  capability to the system; a [member] to the contract element it is
  a member of; any other element to the capability whose folder holds
  its file, or to the system for an element in a file of the system
  level. The system belongs to none, and so does an element where the
  tree holds no element it would belong to, such as a feature in a
  capability folder without its `capability.md`.
- Contact: Bruno Baketarić
- Status: defined
- Aliases: element parents
