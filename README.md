# Ninefold Sudoku

Ninefold Sudoku is a privacy-first multiplayer Sudoku web application for private, account-free play.

## Project status

The repository contains the canonical specifications and the first implementation foundation. The
MVP gameplay itself is not implemented yet.

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

## Local development

Prerequisites:

- Go 1.26.5
- Node.js 24.18.0
- pnpm 11.12.0

Copy `.env.example` to `.env`, review the development-only credentials, then export the variables
before starting the applications:

```sh
set -a
. ./.env
set +a
pnpm install --frozen-lockfile
make dev
```

The web development server listens on `http://localhost:5173`; the Go API listens on
`http://127.0.0.1:8080` and exposes `/health/live`.

Available root checks:

```sh
make test
make lint
make build
```

`make generate`, `make migrate`, `make puzzles`, `make e2e`, and `make tla` fail with the phase
that will implement them. This is intentional: incomplete subsystems are not reported as passing.

## Planning and delivery

- [TASKS.md](TASKS.md) — sequential, one-commit-per-phase implementation ledger for the `0.3.0` MVP
- [ROADMAP.md](ROADMAP.md) — release contracts from the internal prototype checkpoints through `1.0.0`
