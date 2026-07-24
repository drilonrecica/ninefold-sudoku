# AGENTS.md

## Project

**Ninefold Sudoku** is a privacy-first, multiplayer-first Sudoku web app hosted at `ninefold.recica.dev`.

Repository: `ninefold-sudoku`  
Default branch: `master`

## Documentation Authority

Before coding, read the relevant documents in this order:

1. `docs/DOMAIN.md` — gameplay rules, invariants, state machines, commands, and events
2. `docs/ARCHITECTURE.md` — implementation constraints, protocols, persistence, deployment, and testing
3. `docs/PRODUCT.md` — product scope, priorities, MVP boundaries, and non-goals
4. `docs/DESIGN.md` — visual, interaction, responsive, accessibility, and localization rules

Higher-ranked documents override lower-ranked ones. Do not silently resolve contradictions; update the canonical document or record an explicit architectural decision.

## Stack

- Frontend: SvelteKit, TypeScript, Tailwind CSS, shadcn-svelte
- Backend: Go, Chi, native WebSockets
- Database: SQLite in WAL mode
- Deployment: Docker on a Coolify-managed VPS
- Architecture: modular monolith in a monorepo
- Client persistence: IndexedDB
- No Redis, message broker, external analytics, or player-account system in V1

## Non-Negotiable Engineering Rules

- Multiplayer state is server-authoritative.
- Never send the complete multiplayer puzzle solution to clients.
- Commit durable events to SQLite before acknowledgement or broadcast.
- Serialize room and match mutations through the room actor.
- Domain code must not import HTTP, WebSocket, SQLite, or UI packages.
- Transport handlers must remain thin and contain no gameplay rules.
- Commands must be idempotent through request IDs.
- Clients must handle versions, event gaps, snapshots, and reconnection.
- Client clocks are visual only; server time determines deadlines and results.
- Completed matches never return to an active state.
- Historical events are append-only and schema-versioned.
- Replays must reconstruct deterministically from the authoritative event log.
- Replay events use a SHA-256 hash chain; completed digests are sealed with Ed25519.
- Concurrency-sensitive behavior must remain consistent with the TLA+ specification.
- Do not add infrastructure or dependencies solely for novelty.

## Priorities

Evaluate trade-offs in this order:

1. Privacy and security
2. Accessibility and usability
3. Performance and resource efficiency
4. Gameplay integrity
5. Visual quality
6. SEO
7. Feature breadth

Mode priority: Co-op, Race, Duel, Solo, Daily Ninefold.

The first vertical slice is:

`Create room → Join by code → Ready check → Co-op game → Complete → Replay → Rematch`

Do not implement deferred features unless `docs/PRODUCT.md` explicitly includes them in the current scope.

## Privacy

- No accounts, advertising, behavioral analytics, or third-party tracking.
- Use temporary room-scoped display names.
- Store credentials only in secure HTTP-only cookies.
- Store server-side session tokens as hashes.
- Keep personal solo history and preferences on the device.
- Never log cookies, tokens, puzzle solutions, private notes, or full WebSocket payloads.
- Expire temporary rooms, sessions, and replays according to policy.
- Keep admin access private or reverse-proxy protected.

## Accessibility and Localization

Target WCAG 2.2 AA. Every gameplay feature must support:

- complete keyboard operation,
- visible focus and screen-reader labeling,
- non-color-only indicators,
- reduced motion,
- 200% zoom,
- touch targets of at least 44×44 CSS pixels.

Supported locales: English (`en`), German (`de`), Albanian (`sq`), Turkish (`tr`).

Use translation keys and named placeholders. Never concatenate translated sentence fragments.

## Performance

- Prefer platform APIs and small focused dependencies.
- Lazy-load replay and mode-specific code.
- Keep ephemeral presence, focus, and animation events out of SQLite.
- Use bounded actor and WebSocket queues.
- Pre-generate and verify puzzles outside request paths.
- Render the board with semantic HTML and CSS Grid, not canvas.
- Avoid polling, unnecessary background work, and duplicated state.

## Workflow

For each behavior change:

1. Confirm it is in scope and read the relevant domain rules.
2. Add or update domain tests first.
3. Update HTTP, WebSocket, or event contracts.
4. Implement backend behavior, then deterministic frontend reducer handling.
5. Verify stale commands, reconnects, keyboard, mobile, localization, and accessibility.
6. Update canonical documentation and run all relevant checks.

## Definition of Done

A feature requires correct invariants, typed contracts, backend tests, frontend state handling, structured errors, reconnect behavior where applicable, keyboard and mobile usability, accessibility, localization, performance review, and documentation updates.

Never weaken an invariant, bypass persistence ordering, or duplicate business logic merely to make a test pass.
