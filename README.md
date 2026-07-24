# Ninefold Sudoku

Ninefold Sudoku is a privacy-first multiplayer Sudoku web application for private, account-free play.

## Project status

This repository currently contains the canonical product, domain, architecture, and design specifications plus the implementation task plan and version roadmap. Application implementation has not started.

The current delivery target is the full `0.3.0` MVP:

- production-grade Co-op multiplayer;
- online Solo with device-local progress;
- private Room creation and joining;
- reconnect, recovery, results, replay, and rematch;
- cryptographically verifiable multiplayer replay;
- English UI built on localization-ready contracts;
- responsive WCAG 2.2 AA gameplay;
- deployment to a single Coolify-managed VPS.

Race, Duel, Daily Ninefold, offline Solo, Explain hints, full spectator UX, additional locales, and PWA installation are deferred. Their detailed specifications are provisional until the corresponding feature enters scope.

## Documentation

Read the canonical documents in authority order:

1. [Domain specification](docs/DOMAIN.md)
2. [Architecture specification](docs/ARCHITECTURE.md)
3. [Product specification](docs/PRODUCT.md)
4. [Design specification](docs/DESIGN.md)

See [AGENTS.md](AGENTS.md) for contributor rules and documentation authority.

## Planning and delivery

- [TASKS.md](TASKS.md) — sequential, one-commit-per-phase implementation ledger for the `0.3.0` MVP
- [ROADMAP.md](ROADMAP.md) — release contracts from the internal prototype checkpoints through `1.0.0`
