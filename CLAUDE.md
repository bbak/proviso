# Proviso

Tooling for a specification model built on Guarded Gherkin and a glossary:
one living spec tree per system, with derived deltas, impact sets and
independent test and coding agents.

- The model, and with it the requirements for this tooling:
  [model/IMPROVEMENTS.md](model/IMPROVEMENTS.md)
- Where things stand and what is next: [HANDOVER.md](HANDOVER.md). Read it
  first in a new session.
- Docs, specs and code comments: English

## User Preferences

### Conceptual Integrity

- Conceptual integrity is important. I care for these three things:
  - **Cohesion** measures how closely related the elements within a single module, class, or component are, ideally focusing on one well-defined purpose. High cohesion means functions and data inside a module work together tightly, improving readability, maintainability, and reusability.
  - **Coherence** in programming refers to the logical flow and meaningful organization of code or system components, ensuring they form a unified, understandable whole. Coherent code avoids disjointed logic that confuses developers.
  - **Consistency** ensures uniformity in coding style, naming conventions, architecture, and behavior across a codebase or system. It covers standards like indentation, variable naming (e.g., camelCase throughout), and design patterns, reducing errors and easing collaboration in teams.

### Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify

### Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## Project Preferences

- Ensure Unix-style line endings.
- Use Conventional Commits for git commit messages with the following prefixes:
  - `feat`: A new feature
  - `fix`: A bug fix
  - `refactor`: Code refactoring
  - `docs`: Documentation changes
  - `chore`: Maintenance, scaffolding, or project setup
  - `build`: Build system changes
