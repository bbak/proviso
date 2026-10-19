# Findings

## <a id="lint.findings"></a> findings

- Represents: [finding]

The kinds of finding lint reports. Each kind is a member, and its name is
the value of `kind` in the `report`.

### <a id="lint.findings.anchor-misplaced"></a> anchor-misplaced

An anchor that does not start the text of a heading.

### <a id="lint.findings.identifier-malformed"></a> identifier-malformed

An identifier that is not well-formed.

### <a id="lint.findings.identifier-duplicate"></a> identifier-duplicate

An identifier that another element has too.

### <a id="lint.findings.system-identifier"></a> system-identifier

A system whose identifier is not `system`.

### <a id="lint.findings.file-unknown"></a> file-unknown

A file in the spec tree that the layout does not list.

### <a id="lint.findings.capability-file-missing"></a> capability-file-missing

A capability folder without `capability.md`.

### <a id="lint.findings.title-missing"></a> title-missing

A file of a sibling whose first heading is not a title.

### <a id="lint.findings.element-unplaced"></a> element-unplaced

An element whose heading and file give no element type.

### <a id="lint.findings.heading-unknown"></a> heading-unknown

A heading without an anchor that is no section of its element.

### <a id="lint.findings.section-missing"></a> section-missing

An element without a section its type requires.

### <a id="lint.findings.section-repeated"></a> section-repeated

A section that appears twice in one element.

### <a id="lint.findings.contact-missing"></a> contact-missing

A feature without a contact.

### <a id="lint.findings.sibling-empty"></a> sibling-empty

A file of a sibling with nothing but its title.
