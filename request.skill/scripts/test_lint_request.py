"""Tests of lint_request.py. The cases of the rule block, examples,
glossary and undecided checks follow the Examples tables of the lint
capability's rules in the spec tree (proviso/capabilities/lint/), which
these checks port; the rest follow references/format.md.

Run from the repository root: python -m unittest discover -s request.skill/scripts
"""
import os
import re
import unittest

import lint_request

HERE = os.path.dirname(os.path.abspath(__file__))

BLOCK = """\
```guarded
When      the customer asks for <amount>
Provided  <amount> is at most $500
Then      the ATM pays out <amount>
  And     nothing else changes
Otherwise the ATM refuses the request
  And     nothing else changes
```"""

TABLE = """\
| amount | provided | outcome | why |
| --- | --- | --- | --- |
| $500 | yes | pays | the limit exactly |
| $501 | no | refuses | |"""

GLOSSARY = """\
### daily limit

- Definition: The most a customer may withdraw in one calendar day.
- Owner: Maria Keller
- Status: defined
- Source: Maria Keller, 2026-10-03"""


def request(block=BLOCK, table=TABLE, glossary="(none)", questions="(none)",
            front="status: draft", intent="The ATM refuses what it cannot pay.",
            scope="- In: paying out cash", after=""):
    return """\
---
title: A request
requester: Maria Keller
owner: Maria Keller
date: 2026-10-03
{front}
---

# A request

## Intent

{intent}

## Today

Nothing yet.

## Scope

{scope}

## Rules

### R1 — The ATM pays out at most $500

{block}

{table}
{after}
## Glossary

{glossary}

## Open questions

{questions}

## Notes for whoever builds this

{notes}
""".format(front=front, intent=intent, scope=scope, block=block, table=table,
           after=after, glossary=glossary, questions=questions, notes=lint_request.NOTES)


def codes(text, level=None):
    rep = lint_request.Report()
    lint_request.lint(text, rep)
    return sorted(c for _, lv, c, _, _ in rep.findings if level in (None, lv))


def block(*lines):
    return "```guarded\n" + "\n".join(lines) + "\n```"


class Whole(unittest.TestCase):
    def test_clean(self):
        self.assertEqual(codes(request()), [])

    def test_example(self):
        with open(os.path.join(HERE, "..", "references", "example-request.md"), encoding="utf-8") as f:
            text = re.search(r"^`````markdown\n(.*?)^`````$", f.read(), re.S | re.M).group(1)
        self.assertEqual(codes(text, "error"), [])

    def test_front_matter(self):
        self.assertIn("front-matter", codes(request().split("---\n", 2)[2]))
        self.assertIn("status-unknown", codes(request(front="status: done")))
        self.assertIn("field-unknown", codes(request(front="status: draft\npriority: high")))
        text = request().replace("date: 2026-10-03", "date: 3 October")
        self.assertIn("date-invalid", codes(text))
        text = request().replace("owner: Maria Keller", "owner: TBD")
        self.assertEqual(codes(text), ["undecided"])

    def test_sections(self):
        self.assertIn("section-missing", codes(request().replace("## Today\n\nNothing yet.\n", "")))
        self.assertIn("heading-unknown", codes(request(after="\n## Background\n\nText.\n")))
        text = request().replace("## Intent", "## Tmp").replace("## Today", "## Intent").replace("## Tmp", "## Today")
        self.assertIn("section-order", codes(text))
        self.assertIn("section-empty", codes(request(intent="")))
        self.assertIn("notes-changed", codes(request().replace("The rules decide.", "The rules guide.")))

    def test_heading_in_a_fence_is_no_section(self):
        self.assertEqual(codes(request(after="\n```\n## Not a section\n```\n")), [])

    def test_scope(self):
        self.assertIn("scope-in-missing", codes(request(scope="- Out: the card check")))
        self.assertIn("scope-line", codes(request(scope="- In: paying out\n- Maybe: the card")))
        self.assertEqual(codes(request(scope="- In: ?", questions="- [Scope] What is in? (ask: Maria)")), ["undecided"])

    def test_rule_heading(self):
        text = request().replace("### R1 — The ATM", "### The ATM")
        self.assertIn("rule-heading", codes(text))

    def test_undecided_intent_needs_its_question(self):
        self.assertEqual(codes(request(intent="?")), ["question-missing", "undecided"])
        self.assertEqual(codes(request(intent="?", questions="- [Intent] Why now? (ask: Maria)")), ["undecided"])

    def test_ready(self):
        self.assertEqual(codes(request(front="status: ready")), [])
        text = request(front="status: ready", intent="?", questions="- [Intent] Why now? (ask: Maria)")
        self.assertIn("ready-with-open", codes(text))


class RuleBlock(unittest.TestCase):
    # lint.rules.block
    def test_block_count(self):
        self.assertIn("rule-block-count", codes(request(block=BLOCK.replace("guarded", "gherkin"))))
        self.assertIn("rule-block-count", codes(request(block=BLOCK + "\n\n" + BLOCK)))

    # lint.rules.keyword
    def test_keyword(self):
        for line, ok in [("  And     nothing else changes", True),
                         ("But       nothing is reported", False),
                         ("given     the ATM holds notes", False),
                         ("          the ATM goes on", False)]:
            b = BLOCK.replace("  And     nothing else changes\nOtherwise", line + "\n  And     x\nOtherwise", 1)
            self.assertEqual("keyword-unknown" in codes(request(block=b)), not ok, line)

    # lint.rules.order
    def test_order(self):
        b = block("When      the customer asks for <amount>", "Given     the ATM holds notes",
                  "Then      the ATM pays out <amount>", "  And     nothing else changes")
        self.assertIn("keyword-order", codes(request(block=b, table="| amount | outcome |\n| --- | --- |\n| $5 | pays |")))
        b = block("  And     the ATM holds notes", "When      the customer asks for <amount>",
                  "Then      the ATM pays out <amount>", "  And     nothing else changes")
        self.assertIn("keyword-order", codes(request(block=b, table="| amount | outcome |\n| --- | --- |\n| $5 | pays |")))

    # lint.rules.count
    def test_count(self):
        b = block("When      the customer asks for <amount>", "When      the customer asks again",
                  "Then      the ATM pays out <amount>", "  And     nothing else changes")
        self.assertIn("keyword-count", codes(request(block=b, table="| amount | outcome |\n| --- | --- |\n| $5 | pays |")))
        b = BLOCK.replace("Then", "Provided  <amount> is a multiple of $10\nThen", 1)
        self.assertIn("keyword-count", codes(request(block=b)))

    # lint.rules.guard
    def test_guard(self):
        b = block("When      the customer asks for <amount>", "Provided  <amount> is at most $500",
                  "Then      the ATM pays out <amount>", "  And     nothing else changes")
        self.assertIn("guard-incomplete", codes(request(block=b)))

    # lint.rules.and
    def test_and(self):
        b = BLOCK.replace("Then      the ATM pays out <amount>\n  And     nothing else changes",
                          "Then      the ATM pays out <amount>")
        self.assertIn("and-missing", codes(request(block=b)))
        b = BLOCK.replace("Otherwise the ATM refuses the request\n  And     nothing else changes",
                          "Otherwise the ATM refuses the request")
        self.assertIn("and-missing", codes(request(block=b)))
        b = BLOCK.replace("  And     nothing else changes\nOtherwise", "  And     ?\nOtherwise")
        self.assertEqual(codes(request(block=b, questions="- [R1] Anything else? (ask: Maria)")), ["undecided"])

    # lint.rules.given
    def test_given(self):
        b = BLOCK.replace("When", "Given     the ATM has noted <amount>\nWhen", 1)
        self.assertIn("given-refers-to-event", codes(request(block=b)))
        b = BLOCK.replace("When", "Given     the ATM holds <notes>\nWhen", 1)
        t = TABLE.replace("| amount |", "| notes | amount |").replace("| --- |", "| --- | --- |", 1)
        t = t.replace("| $500 |", "| 3 | $500 |").replace("| $501 |", "| 3 | $501 |")
        self.assertEqual(codes(request(block=b, table=t)), [])

    # lint.rules.provided
    def test_provided(self):
        b = BLOCK.replace("Provided  <amount> is at most $500", "Provided  the ATM is switched on")
        self.assertEqual(codes(request(block=b, questions="- [R1] Is it state? (ask: Maria)")), ["provided-without-event"])

    # lint.rules.bound
    def test_bound(self):
        b = BLOCK.replace("Then      the ATM pays out <amount>", "Then      the ATM pays out <sum>")
        self.assertIn("placeholder-unbound", codes(request(block=b)))

    # lint.examples.reserved
    def test_reserved(self):
        b = BLOCK.replace("<amount>", "<outcome>")
        self.assertIn("placeholder-reserved", codes(request(block=b)))

    def test_placeholder_names(self):
        # A name with a space is no placeholder; its column has no purpose.
        b = BLOCK.replace("<amount>", "<the amount>")
        self.assertIn("examples-column-unknown", codes(request(block=b)))

    # lint.undecided.token
    def test_undecided_token(self):
        q = "- [R1] What happens? (ask: Maria)"
        for line, n in [('Then      the ATM shows "?"', 0), ("Then      ?", 1),
                        ("Then      <amount> ? the limit", 1), ("Then      `?` is shown", 0)]:
            b = BLOCK.replace("Then      the ATM pays out <amount>", line)
            self.assertEqual(codes(request(block=b, questions=q)).count("undecided"), n, line)

    def test_relative_word(self):
        b = BLOCK.replace("Provided  <amount> is at most $500", "Provided  <amount> is reasonable")
        self.assertIn("relative-word", codes(request(block=b)))
        b = BLOCK.replace("Then      the ATM pays out <amount>", "Then      the ATM pays out a reasonable <amount>")
        self.assertNotIn("relative-word", codes(request(block=b)))
        b = BLOCK.replace("Provided  <amount> is at most $500", "Provided  <amount> is [reasonable]")
        self.assertNotIn("relative-word", codes(request(block=b)))

    def test_open_items_need_a_question(self):
        b = BLOCK.replace("Then      the ATM pays out <amount>", "Then      ?")
        self.assertIn("question-missing", codes(request(block=b)))
        self.assertNotIn("question-missing", codes(request(block=b, questions="- [R1] What happens? (ask: Maria)")))


class Examples(unittest.TestCase):
    # lint.examples.present
    def test_present(self):
        self.assertIn("examples-missing", codes(request(table="")))
        b = block("When      the card is inserted", "Then      the ATM asks for the PIN", "  And     nothing else changes")
        self.assertIn("examples-missing", codes(request(block=b, table="")))

    # lint.examples.columns
    def test_columns(self):
        self.assertIn("examples-column-missing", codes(request(table=TABLE.replace("| amount |", "| sum |"))))
        t = "| amount | provided | why |\n| --- | --- | --- |\n| $500 | yes | |\n| $501 | no | |"
        self.assertIn("examples-column-missing", codes(request(table=t)))

    # lint.examples.provided-column
    def test_provided_column(self):
        t = "| amount | outcome |\n| --- | --- |\n| $500 | pays |\n| $501 | refuses |"
        self.assertIn("provided-column", codes(request(table=t)))
        b = block("When      the customer asks for <amount>", "Then      the ATM pays out <amount>",
                  "  And     nothing else changes")
        self.assertIn("provided-column", codes(request(block=b, table=TABLE)))

    # lint.examples.unknown-column
    def test_unknown_column(self):
        t = TABLE.replace("| why |", "| expected |")
        self.assertIn("examples-column-unknown", codes(request(table=t)))

    # lint.examples.rows
    def test_rows(self):
        self.assertIn("examples-empty", codes(request(table="| amount | provided | outcome | why |\n| --- | --- | --- | --- |")))

    # lint.examples.cells
    def test_cells(self):
        self.assertIn("examples-row-incomplete", codes(request(table=TABLE.replace("| $501 | no | refuses | |", "| $501 | no | refuses |"))))
        self.assertIn("examples-row-incomplete", codes(request(table=TABLE.replace("| $501 | no | refuses | |", "| $501 | no | | |"))))
        q = "- [R1] What happens? (ask: Maria)"
        self.assertEqual(codes(request(table=TABLE.replace("| $501 | no | refuses | |", "| $501 | no | ? | |"), questions=q)), ["undecided"])

    # lint.examples.row-repeated
    def test_row_repeated(self):
        self.assertIn("examples-row-repeated", codes(request(table=TABLE + "\n| $500 | yes | pays | again |")))
        self.assertIn("examples-row-repeated", codes(request(table=TABLE + "\n| $500 | yes | refuses | |")))
        q = "- [R1] Which amount? (ask: Maria)"
        t = TABLE + "\n| ? | yes | pays | |\n| ? | yes | pays | |"
        self.assertNotIn("examples-row-repeated", codes(request(table=t, questions=q)))

    # lint.examples.provided-value
    def test_provided_value(self):
        for v, ok in [("no (over the limit)", True), ("true", False), ("Yes", False)]:
            t = TABLE.replace("| $501 | no |", "| $501 | %s |" % v)
            self.assertEqual("provided-value-unknown" in codes(request(table=t)), not ok, v)

    # lint.examples.sides
    def test_sides(self):
        t = TABLE.replace("| $501 | no | refuses | |", "| $400 | yes | pays | |")
        self.assertIn("examples-side-missing", codes(request(table=t, questions="- [R1] Too much? (ask: Maria)"), "open"))

    # lint.examples.parts
    def test_parts(self):
        b = BLOCK.replace("Then", "  And     <amount> is a multiple of $10\nThen", 1)
        self.assertIn("examples-part-missing", codes(request(block=b, questions="- [R1] Odd amounts? (ask: Maria)"), "open"))
        t = TABLE + "\n| $15 | no (a multiple of $10) | refuses | |"
        self.assertEqual(codes(request(block=b, table=t)), [])


class Glossary(unittest.TestCase):
    def rule_with_term(self, glossary, questions="(none)"):
        b = BLOCK.replace("<amount> is at most $500", "<amount> is within the [daily limit]")
        return codes(request(block=b, glossary=glossary, questions=questions))

    # lint.references.term
    def test_term(self):
        self.assertEqual(self.rule_with_term(GLOSSARY), [])
        self.assertEqual(self.rule_with_term(GLOSSARY.replace("daily limit", "Daily Limit", 1)), [])
        self.assertEqual(self.rule_with_term(GLOSSARY.replace("daily limit", "day limit", 1) + "\n- Aliases: daily limit"), [])
        self.assertIn("term-unresolved", self.rule_with_term("(none)"))

    # lint.references.marked
    def test_marked(self):
        b = BLOCK.replace("<amount> is at most $500", "<amount> is within the daily limit")
        self.assertIn("term-unmarked", codes(request(block=b, glossary=GLOSSARY)))
        b = BLOCK.replace("<amount> is at most $500", 'the "daily limit" holds <amount>')
        self.assertNotIn("term-unmarked", codes(request(block=b, glossary=GLOSSARY)))

    # lint.references.term-ambiguous
    def test_ambiguous(self):
        other = GLOSSARY.replace("### daily limit", "### limit") + "\n- Aliases: Daily Limit"
        self.assertIn("term-ambiguous", self.rule_with_term(GLOSSARY + "\n\n" + other))

    # lint.glossary.status
    def test_status(self):
        for v in ["?", "draft", "Defined"]:
            self.assertIn("status-unknown", self.rule_with_term(GLOSSARY.replace("Status: defined", "Status: " + v)), v)

    # lint.glossary.open and lint.undecided.term
    def test_open(self):
        g = "### daily limit\n\n- Definition: ?\n- Owner: TBD\n- Status: open\n- Question: What does it apply to?"
        q = "- [Glossary: daily limit] What does it apply to? (ask: Maria)"
        self.assertEqual(self.rule_with_term(g, q), ["undecided", "undecided"])
        self.assertIn("question-missing", self.rule_with_term(g))
        self.assertIn("open-entry-incomplete", self.rule_with_term(g.replace("\n- Question: What does it apply to?", ""), q))
        self.assertIn("open-entry-incomplete", self.rule_with_term(g.replace("Definition: ?", "Definition: The amount per day"), q))

    # lint.glossary.defined
    def test_defined(self):
        self.assertIn("defined-entry-open", self.rule_with_term(GLOSSARY + "\n- Question: Per card?"))
        self.assertIn("defined-entry-open", self.rule_with_term(GLOSSARY.replace("Definition: The most a customer may withdraw in one calendar day.", "Definition: ?")))
        self.assertIn("source-missing", self.rule_with_term(GLOSSARY.replace("\n- Source: Maria Keller, 2026-10-03", "")))
        self.assertIn("owner-missing", self.rule_with_term(GLOSSARY.replace("Owner: Maria Keller", "Owner: TBD")))
        self.assertIn("agent-named", self.rule_with_term(GLOSSARY.replace("Source: Maria Keller", "Source: Claude")))

    def test_fields(self):
        self.assertIn("field-missing", self.rule_with_term(GLOSSARY.replace("\n- Owner: Maria Keller", "")))
        self.assertIn("field-unknown", self.rule_with_term(GLOSSARY + "\n- Contact: Maria Keller"))
        self.assertIn("field-repeated", self.rule_with_term(GLOSSARY + "\n- Owner: Hans Huber"))

    def test_definition_continues(self):
        g = GLOSSARY.replace("in one calendar day.", "in one\n  calendar day.")
        self.assertEqual(self.rule_with_term(g), [])


class OpenQuestions(unittest.TestCase):
    def test_tags(self):
        self.assertIn("question-untagged", codes(request(questions="- Why now? (ask: Maria)")))
        self.assertIn("question-target-unknown", codes(request(questions="- [R7] What? (ask: Maria)")))
        self.assertIn("question-target-unknown", codes(request(questions="- [Glossary: limit] What? (ask: Maria)")))
        self.assertEqual(codes(request(questions="- [General] Who else? (ask: Maria)")), [])


class Main(unittest.TestCase):
    def test_exit(self):
        import io
        import tempfile
        from contextlib import redirect_stderr, redirect_stdout
        with tempfile.TemporaryDirectory() as d:
            path = os.path.join(d, "r.md")
            for text, status in [(request(), 0), (request(table=""), 1)]:
                with open(path, "w", encoding="utf-8") as f:
                    f.write(text)
                with redirect_stdout(io.StringIO()):
                    self.assertEqual(lint_request.main(["lint_request.py", path]), status)
            with redirect_stderr(io.StringIO()):
                self.assertEqual(lint_request.main(["lint_request.py", os.path.join(d, "missing.md")]), 2)


if __name__ == "__main__":
    unittest.main()
