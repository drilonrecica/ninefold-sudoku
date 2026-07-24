# Ninefold Sudoku documentation

Canonical documents are read in this authority order:

1. [DOMAIN.md](DOMAIN.md) — gameplay rules, invariants, state machines, commands, and events
2. [ARCHITECTURE.md](ARCHITECTURE.md) — implementation constraints, protocols, persistence, deployment, and testing
3. [PRODUCT.md](PRODUCT.md) — current scope, priorities, releases, and non-goals
4. [DESIGN.md](DESIGN.md) — visual, interaction, responsive, accessibility, and localization rules

The current implementation scope is the full `0.3.0` MVP. Sections for Race, Duel, Daily Ninefold, and other deferred features describe product direction but are provisional until `PRODUCT.md` moves the feature into scope and the domain rules receive a focused review.

The architecture specification is the baseline decision record. Add an ADR under `docs/decisions/` only when changing an expensive-to-reverse baseline decision or introducing a new one.
