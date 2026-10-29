#!/usr/bin/env python3
"""Check a request written with the proviso-request skill.

Usage:
    python3 lint_request.py <request.md>

Each finding is one line: "line <n>: <level> <code>: <message>".

    error  the format is broken. Fix it, but never by deciding something
           about the domain: write ? instead, and ask.
    open   something nobody has decided yet: a question for the requester.
           Allowed in a draft; a ready request has none.

The items under "Open questions" are counted, not listed again.

Exit status: 0 without errors, 1 with errors, 2 if the file cannot be read.
Standard library only.
"""
import re
import sys

SECTIONS = ["Intent", "Today", "Scope", "Rules", "Glossary", "Open questions",
            "Notes for whoever builds this"]
FRONT_REQUIRED = ["title", "requester", "owner", "date", "status"]
FRONT_KNOWN = FRONT_REQUIRED + ["system", "ref"]
TERM_FIELDS = ["Definition", "Owner", "Status", "Source", "Aliases", "Question"]
TERM_REQUIRED = ["Definition", "Owner", "Status"]
RESERVED = ["provided", "outcome", "why"]

GIVEN, WHEN, PROVIDED, THEN, OTHERWISE, AND = (
    "Given", "When", "Provided", "Then", "Otherwise", "And")
RANK = {GIVEN: 1, WHEN: 2, PROVIDED: 3, THEN: 4, OTHERWISE: 5}

NOTES = """\
- Intent, Today and Scope explain the request. The rules decide. Where the rules are silent, do not fill the gap from the intent: ask.
- Every `?`, every open glossary term and every entry under Open questions is undecided. Treat each as a stop, not as something to guess.
- Rules with no open items can be built now. A rule with open items cannot, and neither can anything that depends on it.
- `<name>` is a value that varies; `[term]` is defined in the Glossary. Each rule's examples are cases the requester confirmed.
- "Provided" is the condition that decides between "Then" and "Otherwise". Both outcomes are required behaviour."""

# Words whose meaning depends on something the line does not say. In a
# Given, When or Provided they hide a comparison the requester has to state.
RELATIVE = re.compile(r"\b(" + "|".join(re.escape(w) for w in [
    "enough", "sufficient", "sufficiently", "insufficient", "valid", "invalid",
    "appropriate", "inappropriate", "adequate", "inadequate", "suitable",
    "unsuitable", "eligible", "ineligible", "acceptable", "unacceptable",
    "reasonable", "unreasonable", "allowed", "permitted", "applicable",
    "excessive", "too much", "too many", "too few", "too little", "too high",
    "too low", "too large", "too small", "too long", "too short", "too old",
    "too late", "too early", "compliant", "non-compliant", "plausible",
    "implausible", "normal", "usual", "unusual", "typical", "proper",
    "improper"]) + r")\b", re.I)
AGENT = re.compile(r"\b(claude|joule|chatgpt|copilot|gemini|llm)\b|^\s*(the\s+)?(ai|agent|assistant)\s*(,|$)", re.I)
HEADING = re.compile(r"^(#{1,6})[ \t]+(.*?)(?:[ \t]+#+)?[ \t]*$")
FENCE = re.compile(r"^ {0,3}(`{3,}|~{3,})(.*)$")
RULE_HEADING = re.compile(r"^R(\d+)\s+[—–-]\s+\S")
FIELD = re.compile(r"^- ([A-Za-z][A-Za-z0-9 -]*?)\s*:(.*)$")
DELIMITER = re.compile(r"^\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?$")
TAG = re.compile(r"^\[(R\d+(?:\s*,\s*R\d+)*|Glossary:\s*[^\]]+|Intent|Today|Scope|General)\]\s*\S")
DATE = re.compile(r"^\d{4}-\d{2}-\d{2}$")
NO_PART = re.compile(r"^no \(.+\)$")


class Report:
    def __init__(self):
        self.findings = []  # (line, level, code, message, rule)

    def error(self, line, code, message, rule=None):
        self.findings.append((line, "error", code, message, rule))

    def open(self, line, code, message, rule=None):
        self.findings.append((line, "open", code, message, rule))

    def count(self, level):
        return sum(1 for f in self.findings if f[1] == level)

    def rules_open(self):
        return {f[4] for f in self.findings if f[1] == "open" and f[4] is not None}


def first_word(s):
    w = s.split()
    return w[0] if w else ""


def undecided(s):
    return first_word(s) == "?"


def unnamed(s):
    return s.strip() in ("", "?", "TBD")


# ------------------------------------------------------------ rule blocks
#
# A line of a rule block is split into spans as proviso lint splits it: a
# literal runs from a double quote to the next one, a code span from a run
# of backticks to the next run of the same length, a term mark from "[" to
# the next "]" (not a Markdown link), and a placeholder is <name>.

TEXT, LITERAL, CODE, MARK, PLACEHOLDER = range(5)
PLACEHOLDER_NAME = re.compile(r"<([A-Za-z0-9_-]+)>")
MARK_TEXT = re.compile(r"\[([^\[\]]+)\]")


def closing_ticks(s, start, n):
    i = start
    while i < len(s):
        if s[i] != "`":
            i += 1
            continue
        j = i
        while j < len(s) and s[j] == "`":
            j += 1
        if j - i == n:
            return i
        i = j
    return None


def lex(s):
    spans, text_start, i = [], 0, 0
    while i < len(s):
        span = None
        c = s[i]
        if c == '"':
            j = s.find('"', i + 1)
            if j >= 0:
                span = (LITERAL, i, j + 1, s[i + 1:j])
        elif c == "`":
            n = 1
            while i + n < len(s) and s[i + n] == "`":
                n += 1
            end = closing_ticks(s, i + n, n)
            if end is None:
                i += n
                continue
            span = (CODE, i, end + n, s[i + n:end].strip())
        elif c == "[":
            m = MARK_TEXT.match(s, i)
            if m and not s.startswith("(", m.end()):
                span = (MARK, i, m.end(), m.group(1))
        elif c == "<":
            m = PLACEHOLDER_NAME.match(s, i)
            if m:
                span = (PLACEHOLDER, i, m.end(), m.group(1))
        if span is None:
            i += 1
            continue
        if span[1] > text_start:
            spans.append((TEXT, text_start, span[1], s[text_start:span[1]]))
        spans.append(span)
        i = text_start = span[2]
    if len(s) > text_start:
        spans.append((TEXT, text_start, len(s), s[text_start:]))
    return spans


class Line:
    def __init__(self, n, text):
        self.n, self.text = n, text
        word = first_word(text)
        self.keyword = word if word in RANK or word == AND else ""
        self.spans = lex(text)

    def values(self, kind):
        return [v for k, _, _, v in self.spans if k == kind]

    def without(self, *kinds):
        """The line with each span of the kinds, and its keyword, replaced
        by a NUL, which is neither a word character nor white space."""
        out = "".join("\0" if k in kinds else self.text[a:b] for k, a, b, _ in self.spans)
        if self.keyword:
            i = out.find(self.keyword)
            out = out[:i] + "\0" + out[i + len(self.keyword):]
        return out


class Step:
    def __init__(self, head):
        self.head, self.ands = head, []

    def lines(self):
        return [self.head] + self.ands

    def placeholders(self):
        return {p for ln in self.lines() for p in ln.values(PLACEHOLDER)}


def parse_block(source):
    lines, steps = [], []
    for n, text in source:
        if not text.strip():
            continue
        ln = Line(n, text)
        lines.append(ln)
        if ln.keyword == AND:
            if steps:
                steps[-1].ands.append(ln)
        elif ln.keyword:
            steps.append(Step(ln))
    return lines, steps


def words(s):
    """The words of s, as (start, end, lowercase): runs of letters, digits,
    hyphens and underscores."""
    return [(m.start(), m.end(), m.group().lower()) for m in re.finditer(r"[\w-]+", s)]


# ------------------------------------------------------------ the file

class Request:
    def __init__(self, text):
        self.lines = text.replace("\r\n", "\n").replace("\r", "\n").split("\n")
        self.front, self.front_end = {}, 0
        self.sections = []  # (name, line, body as [(n, text)])

    def numbered(self, start):
        return [(i + 1, t) for i, t in enumerate(self.lines) if i >= start]


def split_headings(body, level):
    """Split (n, text) lines at the headings of a level, outside fences:
    [(heading text, line, body)], and the lines before the first heading."""
    parts, before, fence = [], [], None
    for n, text in body:
        m = FENCE.match(text)
        if fence is None and m:
            fence = m.group(1)
        elif fence is not None and m and m.group(1)[0] == fence[0] \
                and len(m.group(1)) >= len(fence) and not m.group(2).strip():
            fence = None
        elif fence is None:
            h = HEADING.match(text)
            if h and len(h.group(1)) == level:
                parts.append((h.group(2).strip(), n, []))
                continue
        (parts[-1][2] if parts else before).append((n, text))
    return parts, before


def fences(body):
    """The fenced blocks of the lines: [(label, line, content, end index)]."""
    out, i = [], 0
    while i < len(body):
        n, text = body[i]
        m = FENCE.match(text)
        if not m:
            i += 1
            continue
        marker, label = m.group(1), first_word(m.group(2))
        content, j = [], i + 1
        while j < len(body):
            c = FENCE.match(body[j][1])
            if c and c.group(1)[0] == marker[0] and len(c.group(1)) >= len(marker) \
                    and not c.group(2).strip():
                break
            content.append(body[j])
            j += 1
        out.append((label, n, content, j))
        i = j + 1
    return out


def split_row(text):
    s = text.strip()
    if s.startswith("|"):
        s = s[1:]
    if s.endswith("|") and not s.endswith("\\|"):
        s = s[:-1]
    cells, start = [], 0
    for i, c in enumerate(s):
        if c == "|" and (i == 0 or s[i - 1] != "\\"):
            cells.append(s[start:i].strip())
            start = i + 1
    cells.append(s[start:].strip())
    return cells


def table_after(body, start):
    """The first table in body from index start: (header line, columns,
    [(n, cells)]), or None."""
    for i in range(start, len(body) - 1):
        n, text = body[i]
        if text.strip().startswith("|") and DELIMITER.match(body[i + 1][1].strip()):
            rows, j = [], i + 2
            while j < len(body) and body[j][1].strip().startswith("|"):
                rows.append((body[j][0], split_row(body[j][1])))
                j += 1
            return n, split_row(text), rows
    return None


# ------------------------------------------------------------ glossary

class Term:
    def __init__(self, name, line):
        self.name, self.line, self.fields = name, line, {}

    def names(self):
        aliases = self.fields.get("Aliases", (0, ""))[1]
        return [self.name] + [a.strip() for a in aliases.split(",") if a.strip()]

    def value(self, key):
        return self.fields.get(key, (0, ""))[1]


def read_glossary(sec, rep):
    terms, lookup = [], {}
    if sec is None:
        return terms, lookup
    entries, before = split_headings(sec[2], 3)
    if not entries and first_word(" ".join(t for _, t in before)) not in ("(none)", ""):
        rep.error(sec[1], "glossary-entry", "a glossary entry is a ### heading with its fields, or the section says (none)")
    for name, line, body in entries:
        t = Term(name, line)
        last = None
        for n, text in body:
            m = FIELD.match(text)
            if m:
                last = m.group(1)
                if last not in TERM_FIELDS:
                    rep.error(n, "field-unknown", "%r is no field of a glossary entry" % last)
                if last in t.fields:
                    rep.error(n, "field-repeated", "the field %r appears twice" % last)
                t.fields[last] = (n, m.group(2).strip())
            elif last and text.startswith(" ") and text.strip():
                n0, v = t.fields[last]
                t.fields[last] = (n0, (v + " " + text.strip()).strip())
        terms.append(t)
    for t in terms:
        check_term(t, rep)
        for name in t.names():
            other = lookup.get(name.lower())
            if other is not None and other is not t:
                rep.error(t.line, "term-ambiguous", "%r names both %r and %r" % (name, other.name, t.name))
            else:
                lookup[name.lower()] = t
    return terms, lookup


def check_term(t, rep):
    for key in TERM_REQUIRED:
        if key not in t.fields:
            rep.error(t.line, "field-missing", "the term %r has no %r field" % (t.name, key))
    status = t.value("Status")
    if "Status" in t.fields and status not in ("defined", "open"):
        rep.error(t.fields["Status"][0], "status-unknown", "the status %r is neither defined nor open" % status)
    definition = t.value("Definition")
    if status == "open":
        if not t.value("Question") or not undecided(definition):
            rep.error(t.line, "open-entry-incomplete", "an open term has the Definition ? and a Question")
        else:
            rep.open(t.line, "undecided", "the term %r is open: %s" % (t.name, t.value("Question")))
    if status == "defined":
        if "Question" in t.fields or undecided(definition):
            rep.error(t.line, "defined-entry-open", "a defined term has a Definition and no Question")
        if unnamed(t.value("Owner")):
            rep.error(t.line, "owner-missing", "a defined term names its owner")
        if unnamed(t.value("Source")):
            rep.error(t.line, "source-missing", "a defined term names who gave the definition, and when, as its Source")
    for key in ("Owner", "Source"):
        if AGENT.search(t.value(key)):
            rep.error(t.fields[key][0], "agent-named", "the %s names an AI agent: a definition comes from a person" % key)


# ------------------------------------------------------------ rules

def check_rule(num, line, body, lookup, rep):
    blocks = fences(body)
    guarded = [b for b in blocks if b[0] == "guarded"]
    if len(guarded) != 1:
        hint = " (pasted text may have lost its code fences)" if not guarded else ""
        rep.error(line, "rule-block-count", "R%d holds %d rule blocks fenced as guarded, not 1%s" % (num, len(guarded), hint), num)
        return
    _, fence_line, source, end = guarded[0]
    lines, steps = parse_block(source)
    check_block(num, fence_line, lines, steps, lookup, rep)
    check_examples(num, line, lines, steps, table_after(body, end + 1), rep)
    for n, text in body[end + 1:]:
        m = re.match(r"^Owner:(.*)$", text.strip())
        if m and unnamed(m.group(1)):
            rep.open(n, "undecided", "R%d names no owner yet" % num, num)


def check_block(num, fence_line, lines, steps, lookup, rep):
    for ln in lines:
        if not ln.keyword:
            hint = ": an exception is a condition of its own" if first_word(ln.text) == "But" else ""
            rep.error(ln.n, "keyword-unknown", "the line does not start with Given, When, Provided, Then, Otherwise or And%s" % hint, num)
    top = 0
    for ln in lines:
        if ln.keyword == AND and top == 0:
            rep.error(ln.n, "keyword-order", "an And comes before any other keyword", num)
            break
        if ln.keyword in RANK:
            if RANK[ln.keyword] < top:
                rep.error(ln.n, "keyword-order", "%s is out of the order Given, When, Provided, Then, Otherwise" % ln.keyword, num)
                break
            top = RANK[ln.keyword]
    counts = {k: sum(1 for s in steps if s.head.keyword == k) for k in RANK}
    wrong = ["%s (%d)" % (k, counts[k]) for k in RANK
             if (k in (WHEN, THEN) and counts[k] != 1) or (k not in (WHEN, THEN) and counts[k] > 1)]
    if wrong:
        rep.error(fence_line, "keyword-count", "one When and one Then, at most one Given, Provided and Otherwise; found %s" % ", ".join(wrong), num)
    if (counts[PROVIDED] > 0) != (counts[OTHERWISE] > 0):
        rep.error(fence_line, "guard-incomplete", "a Provided needs its Otherwise, and an Otherwise a Provided: write 'Otherwise ?' with 'And ?' and ask", num)
    for s in steps:
        if s.head.keyword in (THEN, OTHERWISE) and not s.ands:
            rep.error(s.head.n, "and-missing", "the %s has no And line: what else happens or changes, as the requester said it; 'And ?' until asked" % s.head.keyword, num)

    of = lambda k: [s for s in steps if s.head.keyword == k]
    one_when = counts[WHEN] == 1
    event = {p for s in of(WHEN) for p in s.placeholders()}
    introduced = set(event)
    for s in of(GIVEN):
        used = sorted(s.placeholders() & event)
        if one_when and used:
            rep.error(s.head.n, "given-refers-to-event", "the Given uses <%s>, a value of the When: a condition on the event belongs in the Provided" % used[0], num)
        introduced |= s.placeholders()
    for s in of(PROVIDED):
        if one_when and not s.placeholders() & event:
            rep.open(s.head.n, "provided-without-event", "the condition uses no value of the When: ask what about the event decides it; if nothing, it is state for the Given", num)
    for s in steps:
        if s.head.keyword in (GIVEN, WHEN, PROVIDED):
            for ln in s.lines():
                for m in RELATIVE.finditer(ln.without(LITERAL, CODE, MARK, PLACEHOLDER)):
                    rep.open(ln.n, "relative-word", "%r depends on something the line does not say: ask what it means, then state it or make it a [term]" % m.group(1), num)

    phrases = []
    for name in lookup:
        p = [w for _, _, w in words(name)]
        if p and re.fullmatch(r"[\w\s-]+", name):
            phrases.append(p)
    for ln in lines:
        for p in ln.values(PLACEHOLDER):
            if p in RESERVED:
                rep.error(ln.n, "placeholder-reserved", "the value <%s> is named like the column %s" % (p, p), num)
            if one_when and p not in introduced:
                rep.error(ln.n, "placeholder-unbound", "the value <%s> is introduced by neither the When nor a Given" % p, num)
        for name in ln.values(MARK):
            t = lookup.get(name.strip().lower())
            if t is None:
                rep.error(ln.n, "term-unresolved", "[%s] has no Glossary entry: add one, open until its owner defines it" % name, num)
            elif t.value("Status") != "defined":
                # Asked about under the term, not again under the rule.
                rep.open(ln.n, "undecided", "[%s] is an open term" % name)
        check_unmarked(num, ln, phrases, rep)
        if "?" in ln.without(LITERAL, CODE).split():
            rep.open(ln.n, "undecided", "the line holds a ?", num)


def check_unmarked(num, ln, phrases, rep):
    text = ln.without(LITERAL, CODE, MARK, PLACEHOLDER)
    ws = words(text)
    i = 0
    while i < len(ws):
        best = 0
        for p in phrases:
            if len(p) > best and i + len(p) <= len(ws) and all(
                    ws[i + k][2] == p[k] and (k == 0 or not text[ws[i + k - 1][1]:ws[i + k][0]].strip())
                    for k in range(len(p))):
                best = len(p)
        if not best:
            i += 1
            continue
        used = text[ws[i][0]:ws[i + best - 1][1]]
        rep.error(ln.n, "term-unmarked", "%r is a glossary term used without its mark: write [%s]" % (used, used), num)
        i += best


def check_examples(num, line, lines, steps, table, rep):
    if table is None:
        rep.error(line, "examples-missing", "R%d has no examples table after its rule block" % num, num)
        return
    tline, columns, rows = table
    placeholders = []
    for ln in lines:
        for p in ln.values(PLACEHOLDER):
            if p not in placeholders:
                placeholders.append(p)
    missing = [c for c in placeholders + ["outcome"] if c not in columns]
    if missing:
        rep.error(tline, "examples-column-missing", "the examples table lacks the columns: %s" % ", ".join(missing), num)
    guard = [s for s in steps if s.head.keyword == PROVIDED]
    if ("provided" in columns) != bool(guard):
        rep.error(tline, "provided-column", "a rule with a condition, and only such a rule, has a provided column", num)
    for c in columns:
        if c not in placeholders and c not in RESERVED:
            rep.error(tline, "examples-column-unknown", "the column %r is neither a value of the rule nor provided, outcome or why" % c, num)
    if not rows:
        rep.error(tline, "examples-empty", "the examples table has no rows", num)
        return
    prov = columns.index("provided") if "provided" in columns else -1
    keys = [i for i, c in enumerate(columns) if c in placeholders]
    seen, values = {}, []
    for n, cells in rows:
        if len(cells) != len(columns) or any(
                c == "" and i < len(columns) and columns[i] != "why" for i, c in enumerate(cells)):
            rep.error(n, "examples-row-incomplete", "the row needs one cell per column, each with a value except why; ? if not decided", num)
        for i, c in enumerate(cells):
            if undecided(c):
                rep.open(n, "undecided", "the cell %s is not decided yet" % (columns[i] if i < len(columns) else "beyond the last column"), num)
        if all(i < len(cells) and not undecided(cells[i]) for i in keys):
            key = tuple(cells[i] for i in keys)
            if key in seen:
                rep.error(n, "examples-row-repeated", "the row has the same values as the row on line %d" % seen[key], num)
            else:
                seen[key] = n
        if 0 <= prov < len(cells):
            v = cells[prov]
            if v not in ("yes", "?") and v != "no" and not NO_PART.match(v):
                rep.error(n, "provided-value-unknown", "the provided cell %r is not yes, no, no (<part>) or ?" % v, num)
            values.append(v)
    if not guard or prov < 0:
        return
    yes = values.count("yes")
    no = sum(1 for v in values if v == "no" or NO_PART.match(v))
    if not yes or not no:
        side = " or ".join(s for s, k in (("holding (a yes)", yes), ("failing (a no)", no)) if not k)
        rep.open(tline, "examples-side-missing", "no case shows the condition %s: propose one and ask what happens" % side, num)
    parts = 1 + len(guard[0].ands)
    if parts >= 2 and no < parts:
        rep.open(tline, "examples-part-missing", "the condition has %d parts but only %d cases show it failing: propose one per part and ask" % (parts, no), num)


# ------------------------------------------------------------ the whole

def lint(text, rep):
    req = Request(text)
    lines = req.lines
    start = 0
    if lines and lines[0].strip() == "---":
        for i in range(1, len(lines)):
            if lines[i].strip() == "---":
                for n in range(1, i):
                    if ":" in lines[n]:
                        k, v = lines[n].split(":", 1)
                        k = k.strip()
                        if k not in FRONT_KNOWN:
                            rep.error(n + 1, "field-unknown", "%r is no field of the front matter" % k)
                        req.front[k] = (n + 1, v.strip())
                start = i + 1
                break
    if start == 0:
        rep.error(1, "front-matter", "the file starts with front matter between --- lines")
    front = {k: v for k, (_, v) in req.front.items()}
    for k in FRONT_REQUIRED:
        if not front.get(k):
            rep.error(1, "field-missing", "the front matter has no %r" % k)
    status = front.get("status", "")
    if status and status not in ("draft", "ready"):
        rep.error(req.front["status"][0], "status-unknown", "the status %r is neither draft nor ready" % status)
    if front.get("date") and not DATE.match(front["date"]):
        rep.error(req.front["date"][0], "date-invalid", "the date is written YYYY-MM-DD")
    if front.get("owner") in ("TBD", "?"):
        rep.open(req.front["owner"][0], "undecided", "the request names no owner yet")

    sections, before = split_headings(req.numbered(start), 2)
    if not any(HEADING.match(t) and HEADING.match(t).group(1) == "#" for _, t in before):
        rep.error(start + 1, "title-missing", "a # heading with the title follows the front matter")
    sec = {}
    for name, line, body in sections:
        if name not in SECTIONS:
            rep.error(line, "heading-unknown", "%r is no section of a request" % name)
        elif name in sec:
            rep.error(line, "section-repeated", "the section %r appears twice" % name)
        else:
            sec[name] = (name, line, body)
    for name in SECTIONS:
        if name not in sec:
            rep.error(1, "section-missing", "the section '## %s' is missing" % name)
    known = [n for n, _, _ in sections if n in SECTIONS]
    if known != sorted(known, key=SECTIONS.index):
        rep.error(1, "section-order", "the sections follow the order %s" % ", ".join(SECTIONS))

    opened = set()  # the sections, rules and terms with an open item
    for name in ("Intent", "Today"):
        if name in sec:
            body = " ".join(t for _, t in sec[name][2]).strip()
            if not body:
                rep.error(sec[name][1], "section-empty", "the section %s is empty: ? if nobody has said it yet" % name)
            elif undecided(body):
                rep.open(sec[name][1], "undecided", "%s is not stated yet" % name)
                opened.add(name)
    if "Scope" in sec:
        ins = 0
        for n, text in sec["Scope"][2]:
            if not text.strip():
                continue
            m = re.match(r"^- (In|Out):(.*)$", text.strip())
            if not m or not m.group(2).strip():
                rep.error(n, "scope-line", "a scope line is '- In: …' or '- Out: …'")
                continue
            ins += m.group(1) == "In"
            if undecided(m.group(2)):
                rep.open(n, "undecided", "the scope line is not decided yet")
                opened.add("Scope")
        if not ins:
            rep.error(sec["Scope"][1], "scope-in-missing", "the scope has at least one '- In:' line")

    terms, lookup = read_glossary(sec.get("Glossary"), rep)

    numbers = set()
    if "Rules" in sec:
        rules, _ = split_headings(sec["Rules"][2], 3)
        if not rules:
            rep.error(sec["Rules"][1], "rules-missing", "the request has no rule: '### R1 — <the rule in one sentence>'")
        for head, line, body in rules:
            m = RULE_HEADING.match(head)
            if not m:
                rep.error(line, "rule-heading", "a rule heading is 'R<n> — <the rule in one sentence>'")
                continue
            num = int(m.group(1))
            if num in numbers:
                rep.error(line, "rule-number-repeated", "R%d appears twice" % num)
            numbers.add(num)
            check_rule(num, line, body, lookup, rep)
    opened |= {"R%d" % r for r in rep.rules_open()}
    opened |= {"Glossary: " + t.name.lower() for t in terms if t.value("Status") == "open"}

    questions = 0
    asked = set()
    if "Open questions" in sec:
        for n, text in sec["Open questions"][2]:
            m = re.match(r"^\s*(?:-|\d+[.)])\s+(.*)$", text)
            if not m or m.group(1).strip() == "(none)":
                continue
            item = m.group(1).strip()
            questions += 1
            t = TAG.match(item)
            if not t:
                rep.error(n, "question-untagged", "an open question starts with [R<n>], [Glossary: <term>], [Intent], [Today], [Scope] or [General]")
                continue
            tag = t.group(1)
            if tag.startswith("R"):
                for r in re.findall(r"\d+", tag):
                    if int(r) not in numbers:
                        rep.error(n, "question-target-unknown", "[R%s] names no rule of the request" % r)
                    asked.add("R" + r)
            elif tag.startswith("Glossary:"):
                name = tag.split(":", 1)[1].strip().lower()
                if name not in lookup:
                    rep.error(n, "question-target-unknown", "[%s] names no glossary entry" % tag)
                else:
                    asked.add("Glossary: " + lookup[name].name.lower())
            else:
                asked.add(tag)
    for place in sorted(opened - asked):
        rep.error(0, "question-missing", "%s has open items but no [%s] question under Open questions" % (place, place))

    if "Notes for whoever builds this" in sec:
        got = " ".join(" ".join(t.split()) for _, t in sec["Notes for whoever builds this"][2] if t.strip())
        if got != " ".join(" ".join(t.split()) for t in NOTES.split("\n")):
            rep.error(sec["Notes for whoever builds this"][1], "notes-changed", "the notes for whoever builds this are the fixed text of the format")

    if status == "ready" and (rep.count("open") or questions):
        rep.error(req.front["status"][0], "ready-with-open", "the request is ready, but %d items are open and %d questions listed" % (rep.count("open"), questions))
    return len(numbers), len(terms), questions


def main(argv):
    if len(argv) != 2:
        sys.stderr.write(__doc__)
        return 2
    try:
        with open(argv[1], encoding="utf-8-sig") as f:
            text = f.read()
    except (OSError, UnicodeDecodeError) as e:
        sys.stderr.write("lint_request: %s\n" % e)
        return 2
    rep = Report()
    rules, terms, questions = lint(text, rep)
    for line, level, code, message, _ in sorted(rep.findings, key=lambda f: (f[0], f[1] != "error")):
        print("%s: %s %s: %s" % ("line %d" % line if line else "request", level, code, message))
    print("%d rules, %d terms: %d errors, %d open items, %d open questions"
          % (rules, terms, rep.count("error"), rep.count("open"), questions))
    return 1 if rep.count("error") else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
