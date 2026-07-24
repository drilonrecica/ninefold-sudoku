# Ninefold Sudoku — Architecture Specification

**Document:** `docs/ARCHITECTURE.md`  
**Status:** Canonical implementation architecture  
**Current implementation scope:** Full MVP (`0.3.0`); deferred-feature architecture is provisional
**Product:** Ninefold Sudoku  
**Public URL:** `https://ninefold.recica.dev`  
**Repository:** `ninefold-sudoku`  
**Default branch:** `master`

---

## 1. Purpose and authority

This document defines how Ninefold Sudoku must be implemented and operated.

It is the canonical source of truth for:

- system topology;
- monorepo organization;
- frontend and backend boundaries;
- module layering;
- real-time communication;
- HTTP contracts;
- WebSocket contracts;
- SQLite persistence;
- event persistence and deterministic replay;
- cryptographic replay verification;
- formal verification;
- recovery and graceful shutdown;
- browser persistence;
- security and privacy controls;
- accessibility-enabling architecture;
- localization architecture;
- SEO architecture;
- performance budgets;
- testing strategy;
- CI/CD;
- Coolify deployment;
- observability;
- scaling boundaries.

Gameplay rules and invariants are defined in `docs/DOMAIN.md`. Product scope and release sequencing belong in `docs/PRODUCT.md`. Visual and interaction rules belong in `docs/DESIGN.md`.

When implementation convenience conflicts with this document or `DOMAIN.md`, the implementation must change.

Current-scope rule:

- architecture described for the `0.3.0` MVP is binding;
- Race, Duel, Daily Ninefold, offline Solo, PWA installation, host approval, full spectator UX, reporting, and additional locales are provisional;
- do not scaffold empty modules, routes, tables, flags, or generated contracts for deferred features.

---

## 2. Architectural priorities

Every architectural trade-off must be evaluated in this order:

1. Privacy and security
2. Accessibility and usability
3. Performance and resource efficiency
4. Gameplay integrity
5. Visual quality
6. SEO
7. Feature breadth

This ordering is intentional. Ninefold must not gain a feature by weakening privacy, server authority, accessibility, reliability, or resource discipline.

The architecture should remain:

- small enough for one developer to understand;
- efficient enough for a modest VPS;
- explicit enough for AI coding agents to implement safely;
- modular enough to evolve;
- simple enough to operate without a platform team.

---

## 3. Selected technology stack

### 3.1 Frontend

- SvelteKit
- TypeScript
- Svelte 5 runes and focused Svelte stores
- Tailwind CSS
- shadcn-svelte
- semantic HTML
- CSS Grid for the Sudoku board
- Dexie over IndexedDB for local data
- optional SvelteKit service worker after PWA installation enters scope
- Vitest
- Testing Library
- Playwright
- `adapter-node`

### 3.2 Backend

- Go
- Chi router
- native RFC 6455 WebSockets
- `github.com/coder/websocket`
- standard `database/sql`
- `modernc.org/sqlite`
- SQLC
- Goose migrations
- standard-library structured logging using `log/slog`
- Prometheus-compatible operational metrics
- TLA+ specifications for concurrency-sensitive behavior

### 3.3 Persistence

- SQLite on the server
- WAL mode
- one authoritative writer connection
- a small read-only connection pool
- append-only Match event stream
- periodic Match snapshots
- state-oriented Room persistence
- IndexedDB in the browser

### 3.4 Deployment

- one VPS
- Coolify
- Docker containers
- one SvelteKit web container
- one Go server container
- one local persistent SQLite volume owned only by the Go server
- GitHub Actions
- GitHub Container Registry

### 3.5 Deliberately excluded in V1

Do not add these without a new architectural decision:

- Redis
- PostgreSQL
- Kafka, NATS, RabbitMQ, or another message broker
- microservices
- Kubernetes
- external analytics
- external session replay
- external error tracking
- player accounts
- public matchmaking
- Socket.IO
- client-side SQLite/WASM/OPFS
- CRDTs
- peer-to-peer authoritative play
- blockchain
- automatic cloud synchronization of personal player data
- Web Locks or BroadcastChannel as required architecture
- a dedicated background-worker service
- a custom administrator identity system

---

## 4. System context

Ninefold has four principal runtime actors:

1. **Player browser**
2. **SvelteKit web server**
3. **Go application server**
4. **SQLite database file**

External infrastructure is deliberately minimal:

- DNS and TLS;
- Coolify reverse proxy;
- GitHub Container Registry;
- optional future off-VPS backup storage.

### 4.1 Public topology

```text
Browser
  │
  │ HTTPS / WSS
  ▼
Coolify reverse proxy
  ├── /, public pages, game shell ───────► SvelteKit web
  ├── /api/v1/* ─────────────────────────► Go server
  ├── /ws ───────────────────────────────► Go server
  ├── /health/live ──────────────────────► Go server
  └── protected internal/admin routes ──► Go server / SvelteKit admin UI
                                                │
                                                ▼
                                      local SQLite volume
```

### 4.2 Canonical origin

All browser-visible traffic uses one origin:

```text
https://ninefold.recica.dev
```

Canonical routes:

```text
https://ninefold.recica.dev/
https://ninefold.recica.dev/api/v1/*
wss://ninefold.recica.dev/ws
```

The browser must never need to know an internal Docker hostname.

### 4.3 Single-instance authority

V1 runs exactly one authoritative Go server instance.

This is required because:

- SQLite is local to one machine;
- active Room actors live in one process;
- no distributed lock or shared runtime-state system exists;
- deterministic ordering is simplest with one authority.

Coolify must not scale the Go server horizontally.

---

## 5. Architectural style

Ninefold is a **modular monolith**.

It is not a monolithic package and it is not a set of microservices.

The application is one deployable Go server with internal business modules:

- Room
- Match
- Puzzle
- Solo
- Replay
- Administration
- Realtime
- Persistence
- Platform

Daily and Reporting modules are added only when their product features enter scope.

The frontend is a separate deployable SvelteKit application within the same monorepo.

The directory examples in this document describe boundaries, not a requirement to create empty packages.

### 5.1 Why a modular monolith

This structure provides:

- clear business boundaries;
- direct in-process calls;
- atomic SQLite transactions across Room and Match;
- low runtime overhead;
- simple deployment;
- straightforward tracing;
- a credible migration path if a module later becomes a service.

### 5.2 Dependency rule

Business dependencies point inward.

```text
transport → application → domain
infrastructure → application/domain interfaces
```

The domain layer must not import:

- Chi;
- WebSocket libraries;
- SQL drivers;
- SQLC;
- JSON transport types;
- Svelte or frontend code;
- logging packages;
- environment configuration.

---

## 6. Monorepo structure

The repository should use this structure:

```text
ninefold-sudoku/
├── AGENTS.md
├── README.md
├── Makefile
├── go.work
├── package.json
├── pnpm-workspace.yaml
├── .env.example
│
├── apps/
│   ├── web/
│   │   ├── src/
│   │   │   ├── routes/
│   │   │   ├── lib/
│   │   │   │   ├── api/
│   │   │   │   ├── components/
│   │   │   │   ├── game/
│   │   │   │   ├── i18n/
│   │   │   │   ├── realtime/
│   │   │   │   ├── replay/
│   │   │   │   ├── state/
│   │   │   │   ├── storage/
│   │   │   │   └── utils/
│   │   │   ├── app.html
│   │   │   ├── hooks.server.ts
│   │   │   └── app.d.ts
│   │   ├── static/
│   │   ├── tests/
│   │   ├── package.json
│   │   ├── svelte.config.js
│   │   ├── vite.config.ts
│   │   └── tsconfig.json
│   │
│   └── server/
│       ├── cmd/
│       │   ├── api/
│       │   ├── puzzle-generator/
│       │   └── maintenance/
│       ├── internal/
│       │   ├── room/
│       │   ├── match/
│       │   ├── puzzle/
│       │   ├── solo/
│       │   ├── replay/
│       │   ├── admin/
│       │   ├── realtime/
│       │   ├── persistence/
│       │   └── platform/
│       ├── migrations/
│       ├── queries/
│       ├── tests/
│       ├── go.mod
│       └── sqlc.yaml
│
├── contracts/
│   ├── openapi/
│   │   └── ninefold.openapi.yaml
│   ├── websocket/
│   │   ├── client-message.schema.json
│   │   ├── server-message.schema.json
│   │   └── event-payloads/
│   ├── replay/
│   │   ├── replay.schema.json
│   │   └── proof.schema.json
│   ├── fixtures/
│   └── generated/
│       ├── go/
│       └── typescript/
│
├── specification/
│   ├── Room.tla
│   ├── Room.cfg
│   ├── Match.tla
│   ├── Match.cfg
│   └── README.md
│
├── docs/
│   ├── README.md
│   ├── PRODUCT.md
│   ├── DOMAIN.md
│   ├── ARCHITECTURE.md
│   ├── DESIGN.md
│   └── decisions/
│
├── deployments/
│   ├── Dockerfile.web
│   ├── Dockerfile.server
│   ├── compose.local.yml
│   ├── coolify/
│   └── scripts/
│
└── .github/
    ├── workflows/
    ├── dependabot.yml
    └── copilot-instructions.md
```

### 6.1 Generated code

Generated Go and TypeScript contract files are committed.

CI must fail when generated output differs from source schemas.

Generated files must include a clear header:

```text
Code generated from contracts. DO NOT EDIT.
```

---

## 7. Backend module structure

Each business module should follow:

```text
internal/<module>/
├── domain/
├── application/
├── infrastructure/
└── transport/
```

### 7.1 Domain

Contains:

- entities;
- aggregates;
- value objects;
- domain services;
- commands understood by the aggregate;
- domain events;
- invariant enforcement;
- typed domain errors.

### 7.2 Application

Contains:

- use-case commands;
- query handlers;
- authorization orchestration;
- transaction boundaries;
- repository interfaces;
- clock interfaces;
- ID generation interfaces;
- event publication coordination.

### 7.3 Infrastructure

Contains:

- SQLite repositories;
- SQLC adapters;
- cryptographic adapters;
- clock implementations;
- signing-key storage;
- metrics integration;
- filesystem adapters.

### 7.4 Transport

Contains:

- Chi HTTP handlers;
- WebSocket message mapping;
- request decoding;
- response encoding;
- cookie interaction;
- protocol validation.

Transport must not contain gameplay decisions.

---

## 8. Strong domain types

Raw strings and integers must not move freely through domain code.

Examples:

```go
type RoomID string
type MatchID string
type ParticipantID string
type PuzzleID string
type RequestID string
type RoomCode string
type CellIndex uint8
type Digit uint8
type RoomVersion uint64
type MatchVersion uint64
type EventNumber uint64
```

Construction must use validated constructors:

```go
func NewCellIndex(value int) (CellIndex, error)
func NewDigit(value int) (Digit, error)
func ParseRoomCode(value string) (RoomCode, error)
```

Transport conversion occurs at the boundary.

---

## 9. Command handling model

All state changes use explicit application commands.

Example:

```go
type PlaceValueCommand struct {
    MatchID         MatchID
    ParticipantID   ParticipantID
    RequestID       RequestID
    ExpectedVersion MatchVersion
    ClientSequence  uint64
    Cell            CellIndex
    Value           Digit
}
```

### 9.1 Processing sequence

For an authoritative command:

1. Authenticate room session.
2. Resolve ParticipantID from server-owned session state.
3. Validate transport shape.
4. Validate rate limit.
5. Locate or activate the Room actor.
6. Enqueue command.
7. Actor validates authorization and current version.
8. Domain produces proposed events without mutating committed state.
9. Persistence transaction commits events and projections.
10. Actor applies committed events to in-memory aggregate.
11. Server broadcasts resulting events.
12. Server sends command acknowledgement.

### 9.2 No mutation before commit

The authoritative in-memory aggregate must not be permanently mutated before SQLite commit.

Approved patterns:

- domain returns events from a pure decision method;
- actor decides against an immutable state snapshot;
- actor works on a copy and replaces committed state only after commit.

Rejected pattern:

```text
mutate aggregate
→ attempt commit
→ try to roll back memory if commit fails
```

---

## 10. Room actor model

Each active Room is owned by one goroutine.

A Room actor owns:

- Room aggregate;
- current Match aggregate;
- connected clients;
- participant-to-controller mapping;
- timers;
- short event recovery buffer;
- ephemeral cell focus and soft locks;
- mode-specific live state;
- command channel.

### 10.1 Registry

```go
type RoomRegistry interface {
    GetOrActivate(ctx context.Context, roomID RoomID) (*RoomActor, error)
    FindActive(roomID RoomID) (*RoomActor, bool)
    Deactivate(roomID RoomID)
}
```

The registry guarantees at most one active actor per Room in the process.

### 10.2 Actor queue

Recommended command-channel capacity:

```text
256 messages
```

If full:

- lifecycle and shutdown commands receive priority;
- ordinary commands are rejected with `SERVER_BUSY`;
- queue must never grow without bound.

### 10.3 Serialization guarantee

The actor processes one authoritative mutation at a time.

This prevents:

- simultaneous conflicting placements;
- duplicate Match completion;
- stale turn timeouts;
- host changes racing Match start;
- reconnect commands racing removal;
- two Race winners;
- old commands mutating new rematches.

### 10.4 Timers

All authoritative timers belong to the actor.

Timer callbacks enqueue commands; they never mutate state directly.

Every timer carries:

```go
type TimerToken struct {
    MatchID    MatchID
    Kind       TimerKind
    Generation uint64
}
```

Stale generations are rejected.

### 10.5 Actor deactivation

An actor may deactivate when:

- Room is expired or terminal;
- no connections remain;
- no active Match exists;
- all required state is durably persisted;
- no relevant timer remains.

Reactivation reconstructs state from persistence.

Long Room-expiration deadlines do not keep an idle actor resident. Persist those deadlines and let the maintenance scheduler enqueue expiration after reactivation. Only near-term Match, Countdown, reconnect, or recovery timers keep an actor active.

---

## 11. WebSocket connection architecture

### 11.1 Endpoint

```text
wss://ninefold.recica.dev/ws
```

One WebSocket connection is created per browser tab.

### 11.2 Connection goroutines

Each connection uses:

- one reader goroutine;
- one writer goroutine;
- one cancellation context;
- one bounded outbound queue.

No goroutine is created per message.

### 11.3 Outbound queue

Recommended capacity:

```text
128 messages
```

Slow-client policy:

1. Drop obsolete ephemeral focus/presence updates first.
2. Never silently drop durable Match events.
3. If durable delivery cannot keep up, disconnect the client.
4. Client reconnects and recovers from event checkpoint.

### 11.4 Message limit

Maximum incoming WebSocket message:

```text
64 KiB
```

Most messages should remain under 2 KiB.

Unknown or oversized messages are rejected before reaching the actor.

### 11.5 Heartbeats

- server ping every 20 seconds;
- connection unavailable after 45 seconds without response;
- browser visibility changes do not count as disconnect;
- heartbeat events are not persisted.

### 11.6 Origin validation

Production connections are accepted only when:

- TLS is active;
- `Origin` matches configured public origins;
- session cookie is valid where required;
- protocol version is supported.

Do not accept arbitrary third-party origins.

---

## 12. WebSocket protocol

### 12.1 Initialization

Client begins with:

```json
{
  "type": "connection.initialize",
  "protocolVersion": 1,
  "roomCode": "7KMP4R",
  "lastRoomVersion": 12,
  "lastMatchId": "019...",
  "lastMatchEventNumber": 42
}
```

Server responds with either:

- accepted connection and current role;
- missing events;
- authoritative snapshot;
- refresh-required response;
- session rejection.

### 12.2 Client command envelope

```json
{
  "type": "match.place_value",
  "requestId": "019...",
  "matchId": "019...",
  "expectedMatchVersion": 42,
  "clientSequence": 18,
  "payload": {
    "cell": 37,
    "value": 8
  }
}
```

Required fields depend on command type, but state-changing commands include:

- `type`
- `requestId`
- `clientSequence`
- `expectedRoomVersion` for Room commands, or
- `matchId` and `expectedMatchVersion` for Match commands;
- structured payload

The server derives Room ID and participant identity from the Room session. It never trusts a client-supplied Room ID. Match ID remains explicit so delayed commands from an older rematch can be rejected.

### 12.3 Server event envelope

```json
{
  "type": "match.value_placed",
  "roomId": "019...",
  "matchId": "019...",
  "eventNumber": 814,
  "matchVersion": 43,
  "occurredAt": "2026-07-24T18:42:10.422Z",
  "publicActorId": "019...",
  "publicPayload": {
    "schemaVersion": 1,
    "cell": 37,
    "value": 8
  }
}
```

Durable Match events carry Match event number and Match version. Room updates carry Room version and are recovered through an authoritative Room snapshot. Ephemeral focus, soft-lock, presence, and reaction messages carry neither durable event number nor aggregate version.

A durable ping receives a Match event number but leaves Match version unchanged because it does not mutate authoritative gameplay state.

### 12.4 Acknowledgement

```json
{
  "type": "command.acknowledged",
  "requestId": "019...",
  "accepted": true,
  "aggregate": "match",
  "resultingVersion": 43
}
```

### 12.5 Rejection

```json
{
  "type": "command.rejected",
  "requestId": "019...",
  "accepted": false,
  "code": "STALE_VERSION",
  "currentVersion": 43,
  "details": {}
}
```

Clients may send a non-mutating `command.status` query with a prior RequestID. The server returns the stored terminal receipt or `COMMAND_OUTCOME_UNKNOWN`. This is the reconciliation path after acknowledgement timeout.

### 12.6 Idempotency

Each state-changing command has a globally unique UUIDv7 RequestID.

The server durably records terminal outcomes for HTTP and WebSocket mutations in a command-receipt store. Match events additionally enforce unique RequestID values.

A retry using the same RequestID returns the original outcome.

A retry with a new RequestID is a new command and must not be used as the default timeout behavior.

Command receipts remain for at least 24 hours and never expire while the associated active Room or Match still needs retry safety.

### 12.7 Client sequence

Each controller connection increments `clientSequence`.

The server rejects:

- sequence numbers older than the last processed sequence;
- delayed commands from a prior controller connection;
- commands after explicit control transfer.

A recognized retry with the same RequestID is still allowed.

### 12.8 Aggregate versions

Room and Match versions are independent. The expected version for the command’s target aggregate prevents stale-state mutations.

On mismatch:

- reject command;
- provide current version;
- provide missing events when available;
- otherwise require snapshot resynchronization.

### 12.9 Protocol evolution

Every connection advertises a protocol version.

Server policy:

- support current version;
- optionally support immediately previous version during rolling web/server deployment;
- reject incompatible clients with `REFRESH_REQUIRED`;
- never guess payload interpretation.

---

## 13. Multiple browser tabs

The domain requires one active controller tab per participant.

Implementation is server-authoritative and does not require Web Locks or BroadcastChannel.

### 13.1 Controller lease

For each Participant session, the server tracks:

- active controller ConnectionID;
- zero or more read-only connections;
- controller generation.

First valid connection becomes controller unless an existing controller is active.

Additional tabs become read-only.

### 13.2 Control transfer

A read-only tab may request control transfer.

Server:

1. increments controller generation;
2. revokes old controller;
3. rejects pending commands from old generation;
4. assigns new controller;
5. informs all tabs.

### 13.3 Read-only tab behavior

Read-only tabs may:

- receive state;
- observe Match;
- navigate replay;
- request control.

They may not submit gameplay mutations.

---

## 14. Recovery protocol

### 14.1 Event buffer

Each active actor keeps the latest 500 server events in memory.

On reconnect:

- if requested gap is present, send missing events;
- if gap is too large, send snapshot;
- if needed, send snapshot plus subsequent events.

### 14.2 Client checkpoint

The browser stores:

- Room ID;
- Match ID;
- last event number;
- last aggregate version;
- session-independent local UI state.

It does not store full authoritative multiplayer state as a durable source of truth.

### 14.3 Reconnect backoff

Recommended client schedule:

```text
0.5s → 1s → 2s → 4s → 8s → 10s repeatedly
```

Apply jitter.

### 14.4 Disconnected input

While disconnected:

- disable authoritative gameplay input;
- do not queue value placements;
- do not replay offline moves later;
- preserve confirmed board;
- preserve local unsent note drafts only as non-authoritative UI state;
- resynchronize before enabling input.

---

## 15. Optimistic UI

Optimistic display is allowed only when reversibility and user feedback are clear.

### 15.1 Allowed

- value placement;
- value erase;
- note change;
- ready toggle;
- lightweight reaction.

### 15.2 Not allowed

- Match start;
- winner declaration;
- completion;
- Duel turn transition;
- penalty;
- score update;
- host transfer;
- participant removal;
- result finalization.

### 15.3 Pending representation

The UI derives displayed state from:

```text
confirmed server state
+ reversible pending commands
```

Pending values must be visibly distinguishable.

### 15.4 Command timeout

After 5 seconds without acknowledgement:

1. mark command uncertain;
2. request current status;
3. query whether RequestID was processed;
4. reconcile with authoritative result;
5. show failure only after recovery proves rejection or loss.

---

## 16. HTTP API architecture

All application APIs use:

```text
/api/v1
```

### 16.1 Core endpoints

| Method | Route | Purpose |
|---|---|---|
| `POST` | `/api/v1/rooms` | Create Room and host participant |
| `GET` | `/api/v1/rooms/{code}` | Safe Room preview |
| `POST` | `/api/v1/rooms/{code}/join` | Join an unlocked Room |
| `POST` | `/api/v1/rooms/{code}/resume` | Resume room session |
| `POST` | `/api/v1/rooms/{code}/leave` | Explicit leave intent |
| `POST` | `/api/v1/solo/puzzles` | Request Solo puzzle |
| `POST` | `/api/v1/solo/attempts/{id}/hint` | Request Solo hint |
| `POST` | `/api/v1/solo/attempts/{id}/complete` | Validate Solo completion |
| `GET` | `/api/v1/replays/{replayId}` | Retrieve replay using capability header |
| `DELETE` | `/api/v1/replays/{replayId}` | Delete replay using originating Room session |
| `GET` | `/health/live` | Liveness |
| `GET` | `/health/ready` | Readiness |
| `GET` | `/health/status` | Protected operational status |
| `GET` | `/internal/metrics` | Private metrics |

Gameplay mutations are not duplicated over HTTP.

`GET /api/v1/daily` is added only when Daily Ninefold enters scope.

Every state-changing HTTP request includes:

```text
Idempotency-Key: <UUIDv7 RequestID>
```

The same key and authenticated scope return the original terminal outcome after retry, including across server restart.

### 16.2 Room creation request

```json
{
  "displayName": "Mila",
  "mode": "coop",
  "difficulty": "medium"
}
```

Create and join responses rotate the single `ninefold_room_session` cookie. If the browser already has a different active Room session, the server rejects with `ACTIVE_ROOM_SESSION_EXISTS` until the client submits an explicit leave/replacement intent. The client must explain the consequence before replacement.

### 16.3 Safe Room preview

Preview may include:

- mode;
- difficulty;
- state;
- available seats;
- spectator availability;
- lock state.

It must not expose participant names.

Approval requirement is added when the deferred host-approval workflow enters scope.

### 16.4 Leave intent

Explicit leave request includes an intent so browser closure is not treated as voluntary abandonment.

Example values:

- `leave_lobby`
- `become_spectator`

Race abandonment and Duel resignation intents are added with those modes.

### 16.5 Error envelope

```json
{
  "error": {
    "code": "ROOM_FULL",
    "messageKey": "error.room_full",
    "requestId": "019...",
    "details": {
      "spectatorAvailable": true
    }
  }
}
```

The frontend acts on `code`, localizes `messageKey`, and interpolates only named safe details. The server does not duplicate user-facing English prose in transport contracts.

### 16.6 Replay capability transport

The shareable browser URL is:

```text
https://ninefold.recica.dev/replay/{replayId}#cap={capability}
```

URL fragments are not sent in HTTP request lines. The client reads the fragment in memory and fetches the replay with:

```text
Authorization: Bearer {capability}
```

Requirements:

- capability tokens are never placed in API paths, query strings, logs, analytics, or persistent browser storage;
- on initial load, the client copies the fragment capability into memory and immediately removes the fragment from browser history with `history.replaceState`;
- the Share action reconstructs the fragment URL from the in-memory capability for copying without navigating to it;
- replay pages use `noindex` and `Referrer-Policy: no-referrer`;
- possession of a read capability permits reading only;
- deletion requires the still-valid originating `ninefold_room_session` cookie and confirmed destructive action;
- replacing the one active Room session removes the browser’s early-delete authority for the older Room.

### 16.7 Solo assignment proof

`POST /api/v1/solo/puzzles` returns clues, puzzle metadata, an attempt ID, and a signed opaque assignment proof. The proof binds puzzle revision, transformation, issued time, and format version without containing the solution.

The client stores the proof with device-local attempt state. Hint and completion requests submit the proof and current board. The server validates them against its Puzzle catalog without storing personal Solo progress. Request bodies are never logged.

The assignment proof is a puzzle-validation capability, not a player identity or Room credential.

### 16.8 Status mapping

| Situation | HTTP status |
|---|---:|
| Successful read | 200 |
| Resource created | 201 |
| Successful action without body | 204 |
| Invalid JSON or shape | 400 |
| Invalid session | 401 |
| Insufficient permission | 403 |
| Unavailable resource | 404 |
| State/version conflict | 409 |
| Domain validation failure | 422 |
| Rate limit | 429 |
| Unexpected failure | 500 |
| Maintenance | 503 |

---

## 17. Contract source of truth

### 17.1 HTTP

OpenAPI file:

```text
contracts/openapi/ninefold.openapi.yaml
```

Generate:

- TypeScript request/response types;
- Go transport types where useful;
- mock fixtures;
- API client helpers.

Domain entities must not be generated from OpenAPI.

### 17.2 WebSocket

JSON Schema files define:

- client envelope;
- server envelope;
- each command payload;
- each event payload;
- acknowledgement;
- rejection;
- snapshot;
- replay event.

### 17.3 Shared fixtures

Both Go and TypeScript tests read fixtures from:

```text
contracts/fixtures/
```

Required fixtures include:

- Room creation;
- Co-op placement;
- Co-op Challenge rejection;
- durable ping;
- stale command;
- snapshot recovery;
- replay proof.

Race and Duel fixtures are added with those modes.

---

## 18. SQLite architecture

### 18.1 Server-only database

SQLite exists only on the Go server.

The browser uses IndexedDB, not SQLite.

The SQLite file:

- resides on a local VPS disk;
- is mounted only into the Go container;
- is not exposed over a network;
- is not mounted by the SvelteKit container;
- is not stored on NFS or another network filesystem.

### 18.2 Driver

Use:

```text
modernc.org/sqlite
```

Reasons:

- pure Go;
- avoids CGO Docker complexity;
- suitable for a small VPS;
- simple cross-compilation.

Requirements:

- pin exact module version;
- verify embedded SQLite version at startup;
- maintain an approved-version policy;
- fail readiness when the embedded version violates known WAL safety requirements.

### 18.3 PRAGMAs

At database initialization:

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = FULL;
```

Additional PRAGMAs may be benchmarked but must not weaken integrity without an ADR.

### 18.4 Connection model

Use two `sql.DB` handles:

#### Writer

- maximum open connections: 1;
- maximum idle connections: 1;
- used for all mutations and migrations;
- transaction duration kept short.

#### Reader pool

- small pool, initially 4;
- read-only/query-only configuration where supported;
- used for previews, replays, public pages, and admin queries.

Do not create one connection per Room.

### 18.5 Why one writer

SQLite permits one writer at a time.

Ninefold already serializes mutations per Room. A single writer:

- makes contention explicit;
- prevents writer storms;
- simplifies busy handling;
- keeps resource use low;
- supports predictable testing.

---

## 19. Persistence model

Ninefold is not pure event sourcing for every aggregate.

### 19.1 Match persistence

Match uses:

- current state/projection row;
- append-only event stream;
- periodic snapshots;
- finalized result projection.

The event stream is authoritative for replay and recovery.

### 19.2 Room persistence

Room uses:

- persisted current state;
- participant rows;
- current Match reference;
- lightweight lobby history;
- aggregate version.

Room is not reconstructed exclusively from an unbounded Room event stream.

Every accepted Room mutation and its command receipt commit atomically. This preserves idempotency for creation, join, readiness, settings, host controls, Countdown, and rematch across process restart.

### 19.3 Session persistence

The MVP supports one active Room session per browser profile.

- The browser holds one opaque `ninefold_room_session` cookie.
- The server stores only its hash.
- Creating or joining another Room requires explicit replacement of the prior Room session.
- Multiple tabs may use the same Room session, subject to controller-lease rules.
- Historical participation remains server-side only long enough to authorize replay deletion and retention cleanup.

### 19.4 Puzzle persistence

Puzzle uses explicit fields and immutable clue/solution blobs.

### 19.5 Cross-aggregate transaction

Critical lifecycle operations update Room and Match atomically.

Examples:

- Match creation;
- Room entering InMatch;
- Match completion;
- Room entering Results;
- recovery cancellation.

---

## 20. SQL schema responsibilities

Canonical tables:

```text
puzzles

rooms
room_participants
room_blocks
room_sessions

matches
match_participants
match_events
match_snapshots
match_results
match_result_players
match_tombstones

command_receipts
replay_capabilities
replay_seals
admin_audit_log
```

Add `daily_challenges`, reporting tables, or feature-flag tables only when the corresponding feature enters scope.

### 20.1 Identifiers

Durable IDs are text-encoded UUIDv7.

Timestamps are UTC Unix milliseconds:

```text
created_at_ms INTEGER NOT NULL
```

### 20.2 Puzzle grids

Clues and solutions are fixed 81-byte blobs.

Do not store Sudoku grids as JSON arrays.

### 20.3 Match events

Minimum fields:

```text
match_id
event_number
aggregate_version
public_event_type
public_actor_id
request_id
occurred_at_ms
public_payload_json
private_payload_blob
private_payload_salt
private_payload_digest
previous_hash
event_hash
```

Constraints:

- primary key `(match_id, event_number)`;
- unique `(match_id, request_id)` where RequestID exists;
- event number strictly increasing;
- public payload contains schema version;
- private fields are nullable and never returned by ordinary replay queries;
- aggregate version may remain unchanged for a durable non-state event such as a Co-op ping.

### 20.4 Snapshots

Snapshots contain:

- Match ID;
- event number;
- aggregate version;
- state format;
- compressed state blob;
- timestamp;
- integrity metadata.

Recommended encoding:

- compact JSON or MessagePack;
- gzip or zstd only if benchmarked and dependency cost is justified.

Prefer standard gzip initially for simplicity.

### 20.5 Sessions and command receipts

`room_sessions` stores:

```text
token_hash
room_id
participant_id
created_at_ms
expires_at_ms
revoked_at_ms
```

Plaintext session tokens are never persisted.

`command_receipts` stores:

```text
request_id
authenticated_scope_hash
command_type
request_fingerprint
terminal_status
safe_response_json
created_at_ms
expires_at_ms
```

A repeated RequestID with a different authenticated scope, command type, or request fingerprint is rejected rather than replaying another outcome.

### 20.6 Replay access and proofs

`replay_capabilities` stores only capability-token hashes, Replay ID, Match ID, expiration, and revocation/deletion state.

`replay_seals` stores:

```text
match_id
final_event_number
final_event_hash
terminal_at_ms
signing_key_id
signature
proof_version
created_at_ms
```

Replay proof rows and access records are not Match events.

---

## 21. Match commit transaction

Every accepted durable Match command commits atomically:

1. insert produced events;
2. update Match aggregate version when authoritative state changed;
3. update next event number;
4. update state projection;
5. update participant projections;
6. update result projection if finalizing;
7. update Room state if lifecycle changed;
8. insert or finalize the command receipt;
9. commit.

Only after commit:

- apply events to actor’s committed state;
- broadcast;
- acknowledge.

### 21.1 Optimistic persistence guard

Persistence update includes expected aggregate version.

If no row is updated:

- treat as version conflict;
- reload aggregate;
- do not broadcast proposed events.

### 21.2 Write failure

On one write failure:

- no success acknowledgement;
- no broadcast;
- in-memory state stays at previous committed version.

After three consecutive write failures for one active Match:

- pause authoritative commands;
- broadcast temporary server-error state;
- perform health check;
- resume if healthy;
- otherwise enter RecoveryPending.

---

## 22. SQLC and migrations

### 22.1 SQLC

SQL lives in:

```text
apps/server/queries/
```

SQLC-generated code remains in infrastructure packages.

Domain and application layers must not import generated query packages directly.

### 22.2 Goose

Migrations live in:

```text
apps/server/migrations/
```

Naming:

```text
00001_initial.sql
00002_puzzles.sql
00003_rooms.sql
00004_matches.sql
00005_match_events.sql
...
```

Requirements:

- forward migrations are authoritative;
- destructive down migrations are not run automatically in production;
- migrations are embedded into Go binary or shipped in server image;
- startup checks migration state;
- production migration occurs before readiness.

### 22.3 Expand-and-contract

Schema changes follow:

```text
add
→ dual read/write if needed
→ migrate data
→ stop old usage
→ remove in later release
```

The immediately preceding server release should remain compatible where practical.

---

## 23. Snapshot policy

Persist a Match snapshot:

- when Match starts;
- every 50 durable events;
- every 30 seconds while durable changes are occurring;
- before planned shutdown;
- when Match reaches a retained terminal result.

### 23.1 Recovery

1. Load latest snapshot.
2. Verify Match ID and aggregate version.
3. Replay subsequent events.
4. Validate all domain invariants.
5. Fall back to older snapshot if invalid.
6. Fall back to full event replay if necessary.
7. Mark RecoveryPending before accepting commands.

A snapshot is never more authoritative than the event stream.

---

## 24. Replay architecture

### 24.1 Replay generation

Replay response includes:

- assigned puzzle snapshot;
- immutable MatchRules;
- participant snapshots;
- ordered public-safe event envelopes;
- optional seek snapshots;
- result;
- replay proof;
- expiration timestamp.

### 24.2 Visibility sanitization

Replay projection must enforce:

- session credentials are omitted;
- server-only correctness metadata is omitted unless rules permit disclosure;
- participant data is limited to temporary display names and markers.

For future visibility-sensitive modes, hidden payloads are represented only by salted commitments. Exact Race and Duel visibility projections must be ratified with those modes.

### 24.3 Seek performance

For ordinary Matches:

- initial replay payload should load under 2 seconds;
- timeline should seek from nearest replay snapshot;
- replay code is lazy-loaded;
- replay data may be compressed at HTTP layer.

---

## 25. Cryptographic replay integrity

### 25.1 Canonical event serialization

Hashing must use one stable canonical representation.

Use RFC 8785 JSON Canonicalization Scheme over a versioned public-safe event envelope. Cross-language Go and TypeScript fixtures are mandatory.

### 25.2 Event hash

Each durable event computes:

```text
event_hash = SHA-256(JCS({
  proofVersion,
  matchId,
  eventNumber,
  aggregateVersion,
  publicEventType,
  publicActorId,
  occurredAtMs,
  publicPayload,
  privatePayloadDigest,
  previousEventHash
}))
```

The first event uses a proof-version-defined genesis hash.

For a hidden payload:

```text
private_payload_digest =
  SHA-256(random_128_bit_or_larger_salt || JCS(private_payload))
```

The browser receives the digest, not the payload or salt. The server verifies private payload integrity during recovery.

### 25.3 Signature

At a replay-retained terminal result:

- sign final digest with Ed25519;
- store signing key ID;
- store proof version;
- store signature;
- expose public key to replay verifier.

The final hashed Match event is the terminal gameplay/lifecycle fact. Replay sealing, capability creation, verification results, deletion, expiration, and result amendments are stored separately and never appended after the signed event.

### 25.4 Key management

Private signing key:

- stored only as Coolify secret;
- never committed;
- never sent to web client;
- loaded into Go process;
- rotated with explicit key ID.

Public keys:

- embedded in the versioned web build;
- retained long enough to verify all unexpired replays.

A same-origin registry may publish key metadata, but it is not a substitute for the client’s embedded trust set. Key rotation deploys the new public key before the server begins signing with it.

### 25.5 Browser verifier

The web client verifies:

- sequential event numbers;
- canonical public envelopes;
- hash chain;
- final digest;
- Ed25519 signature;
- trusted key ID.

For hidden events, the browser verifies the signed digest commitment and does not claim to verify undisclosed content.

Use Web Crypto where supported by the browser matrix. If a fallback is required, it must be small, audited, lazy-loaded, and covered by known test vectors.

### 25.6 Verification wording

UI may say:

```text
Replay integrity verified
```

It must not claim that cryptography proves the server applied correct rules. Correct wording is documented in `DESIGN.md`.

---

## 26. Puzzle engine architecture

### 26.1 Components

```text
Sudoku grid model
Constraint validator
Brute-force uniqueness solver
Deterministic logical solver
Technique detector
Hint generator
Difficulty grader
Quality evaluator
Transformation engine
Puzzle generator
```

### 26.2 Production request path

Production game requests must not generate puzzles.

They select from a pre-generated, verified catalog.

### 26.3 Generator command

```text
apps/server/cmd/puzzle-generator
```

Responsibilities:

- generate;
- verify uniqueness;
- grade;
- evaluate quality;
- mark multiplayer approval;
- output importable records;
- include generator and solver versions.

### 26.4 Determinism

Given:

- canonical puzzle;
- transformation seed;
- solver version;

transformed clues and solution must reproduce exactly.

### 26.5 Testing

Property-based tests must verify:

- generated solution validity;
- uniqueness;
- clue consistency;
- transformation preservation;
- deterministic solver output;
- valid hints;
- successful logical completion.

---

## 27. SvelteKit route architecture

Canonical routes:

```text
/
├── /create
├── /join/[code]
├── /room/[code]
├── /play/[matchId]
├── /solo
├── /replay/[replayId]
├── /settings
├── /how-to-play
├── /privacy
├── /accessibility
└── /admin
```

Daily, About, and dedicated mode routes are added only when their product scope is activated.

### 27.1 Rendering strategy

SSR:

- home;
- privacy;
- accessibility;
- help;
- safe Room preview;
- English SEO content.

Client-heavy:

- lobby;
- active Match;
- Solo board;
- replay;
- settings.

Dynamic Room, Match, replay, and admin routes are not prerendered.

### 27.2 Private-route indexing

Add `noindex` to:

- join pages;
- Room pages;
- active Matches;
- replay URLs;
- admin;
- error diagnostic pages.

---

## 28. Frontend state architecture

Separate four categories:

1. Server state
2. Local gameplay UI state
3. Persistent device state
4. Ephemeral visual state

Do not combine them into one global object.

### 28.1 Suggested modules

```text
room-state.svelte.ts
match-state.svelte.ts
socket-state.svelte.ts
game-ui-state.svelte.ts
preferences-state.svelte.ts
replay-state.svelte.ts
solo-state.svelte.ts
```

### 28.2 Authoritative Match state

```ts
type MatchClientState = {
  confirmed: MatchSnapshot;
  matchVersion: number;
  lastMatchEventNumber: number;
  pendingCommands: Map<string, PendingCommand>;
  connection: ConnectionState;
};
```

### 28.3 Event reducer

All server events pass through one deterministic reducer:

```ts
nextState = applyServerEvent(currentState, event);
```

Reducer responsibilities:

- require sequential events;
- ignore duplicates;
- detect gaps;
- update confirmed state only;
- trigger recovery;
- never perform side effects.

### 28.4 Effects

Side effects such as:

- sounds;
- haptics;
- animations;
- navigation;
- notifications;

listen to reducer outputs but do not mutate authoritative state.

---

## 29. Browser persistence

Use IndexedDB through Dexie.

Canonical stores:

```text
preferences
solo_attempts
solo_replays
local_statistics
recent_puzzles
connection_checkpoints
```

Add `daily_progress` when Daily Ninefold enters scope.

### 29.1 LocalStorage

Use only for tiny boot preferences such as:

- initial theme;
- last locale.

Do not store:

- session tokens;
- multiplayer authoritative state;
- replay capabilities;
- puzzle solutions;
- sensitive diagnostics.

### 29.2 No client SQLite

Do not add SQLite WASM, OPFS VFS, or a browser SQL layer.

IndexedDB is sufficient for the approved local data model.

### 29.3 Session cookies

Session credentials are HTTP-only and are not copied into IndexedDB.

---

## 30. PWA and offline behavior

PWA installation and service-worker caching are deferred. The MVP remains an ordinary responsive web application.

### 30.1 Future service worker cache

When implemented, it may cache:

- application shell;
- fonts;
- icons;
- static assets;
- help and privacy pages;
- previously loaded Solo puzzle assets when allowed.

Must not cache:

- Room preview;
- join responses;
- session responses;
- replay capability responses;
- admin responses;
- WebSocket traffic.

### 30.2 Multiplayer

Multiplayer requires an active connection.

No gameplay mutations are accepted offline.

### 30.3 Solo

MVP Solo requires connectivity to begin an attempt, request a hint, and validate completion. Values, notes, timer, and preferences persist locally.

Offline Solo requires a future scope decision; do not create a dormant feature flag or offline solution package in the MVP.

### 30.4 Future update behavior

When a service worker exists, a new version:

- must not force reload during active Match;
- shows update available;
- applies after Match or explicit user action;
- updates automatically on next fresh visit.

---

## 31. Component architecture

Core components:

```text
SudokuBoard
SudokuCell
CandidateGrid
NumberPad
PlayerRoster
MatchHeader
ConnectionBanner
RoomSettings
LobbyReadyControl
CoopActivityPanel
ReplayTimeline
ResultSummary
LanguageSelector
ThemeSelector
```

### 31.1 Board component

`SudokuBoard`:

- receives derived board state;
- emits user intent;
- does not send network commands;
- does not know WebSocket protocol;
- does not calculate authoritative correctness;
- supports keyboard navigation;
- uses semantic HTML and CSS Grid.

### 31.2 No canvas

The main board must not use canvas.

Reasons:

- accessibility;
- focus management;
- zoom;
- text rendering;
- responsive layout;
- semantic inspection;
- simpler testing.

Canvas may be used only for optional decorative or replay effects.

---

## 32. Accessibility architecture

Target WCAG 2.2 AA.

### 32.1 Semantic state

The client must receive enough structured state to render:

- clue versus entered value;
- selected cell;
- row/column/box context;
- direct conflicts;
- solution error where rules allow;
- participant attribution;
- pending state;
- lock state;
- turn ownership;
- progress text;
- score;
- timer;
- connection state.

### 32.2 Keyboard

Required controls:

- arrow keys: move selection;
- `1–9`: place digit;
- Backspace/Delete: erase;
- `N` or Space: notes mode;
- `Ctrl/Cmd + Z`: undo only in modes that explicitly support it; Co-op MVP does not;
- `H`: hint where allowed;
- Escape: close overlay or clear selection.

### 32.3 Live announcements

Use restrained `aria-live` regions for:

- countdown;
- connection lost/restored;
- turn changed;
- accepted/rejected value;
- row/column/box completion;
- winner;
- Match completion.

Do not announce every remote focus change.

### 32.4 Non-color encoding

Participant and error states combine:

- color;
- icon or shape;
- initials or label;
- text where needed.

### 32.5 Automated testing

Use axe-core in Playwright on major routes.

Manual checks include:

- screen reader;
- keyboard-only;
- 200% zoom;
- reduced motion;
- touch targets;
- color-independent interpretation.

Sizing requirements:

- each Sudoku cell target is at least 24×24 CSS pixels;
- number-pad buttons and primary controls are at least 44×44 CSS pixels;
- the board is treated as an essential two-dimensional layout while surrounding controls reflow at 200% zoom.

---

## 33. Localization architecture

Current MVP locale:

- English `en`

Planned pre-1.0 locales:

- German `de`
- Albanian `sq`
- Turkish `tr`

### 33.1 Catalogs

```text
apps/web/src/lib/i18n/
└── en.json
```

Development also uses a generated pseudo-locale for expansion testing. Add real locale catalogs only when translation work enters scope.

### 33.2 Rules

- English is canonical source.
- Every shipped locale must contain all required keys.
- Development reports missing keys.
- Production may fall back to English.
- Use named placeholders.
- Use locale-aware pluralization.
- Never concatenate sentence fragments.
- Room codes remain ASCII.
- Domain and protocol use stable identifiers, never translated strings.

### 33.3 Locale selection

Order:

1. saved preference;
2. browser language;
3. English fallback.

Locale may vary independently per Room participant.

---

## 34. SEO architecture

SEO applies to public pages, not private gameplay state.

### 34.1 Indexable pages

- `/`
- `/how-to-play`
- `/privacy`
- `/accessibility`

About and mode pages are post-MVP routes.

### 34.2 Requirements

- SSR meaningful content;
- canonical URL;
- English metadata;
- `hreflang` only for locales actually shipped;
- `sitemap.xml`;
- `robots.txt`;
- Open Graph;
- social image;
- semantic heading order;
- structured data for software application where appropriate;
- fast Core Web Vitals;
- no intrusive consent banner because no non-essential tracking exists.

### 34.3 Private data protection

Private pages:

- noindex;
- no participant names in metadata;
- no Room code in public sitemap;
- no replay capability in logs or analytics;
- no server-side preview card exposing Match state.

---

## 35. Security architecture

### 35.1 Cookies

Canonical cookies:

```text
ninefold_room_session
```

Production attributes:

- `Secure`
- `HttpOnly`
- `SameSite=Lax`
- narrow path where practical
- explicit expiration

Server stores session-token hashes, not plaintext tokens.

The MVP supports one active Room session per browser profile. Creating or joining a different Room requires an explicit leave/replacement flow. The cookie rotates on create, join, resume, and privilege-sensitive transitions.

Replay read capabilities are intentionally shareable and therefore cannot be HTTP-only cookies. They are limited read credentials kept in a URL fragment and in memory only, as defined in section 16.6.

### 35.2 CSRF and origin

For cookie-authenticated state-changing HTTP:

- validate `Origin`;
- require JSON content type;
- reject cross-origin browser requests;
- add explicit CSRF token if future browser behavior or integration requires it.

### 35.3 Input validation

Three levels:

```text
transport validation
→ application authorization
→ domain validation
```

Never trust:

- ParticipantID from payload;
- client timestamps;
- score;
- correctness;
- role;
- host status;
- completion claim.

### 35.4 Rate limits

Initial per-session limits:

- value commands: 10/second
- note commands: 30/second
- focus updates: 10/second
- ping/reaction: 5 per 10 seconds
- settings changes: 5 per 10 seconds
- Room creation: 10/hour/IP

Failed Room-code lookups receive progressive delay and temporary blocking.

Room codes are generated with a cryptographically secure random source, inserted under a uniqueness constraint, and retried on collision. Do not reuse a code while any Room, session, block, replay-deletion authority, or retained access-control record still refers to it.

When IP-based abuse controls are needed:

- derive the client address only from a configured trusted proxy chain;
- use a rotating keyed digest for application rate-limit keys;
- do not persist raw IP addresses in product data;
- document any unavoidable reverse-proxy security-log retention.

### 35.5 CAPTCHA

No CAPTCHA in normal flow.

Architecture may expose a provider-neutral challenge interface for suspicious traffic later.

### 35.6 Admin

`/admin` is restricted through VPN/private networking for the initial release.

No custom admin account system.

Trusted proxy identity may be passed in a configured header only when request originates from trusted proxy/network.

### 35.7 Secrets

Coolify environment secrets:

- cookie-signing secret;
- replay-signing private key;
- trusted admin proxy configuration;
- future backup credentials;
- registry credentials where needed.

Secrets are never committed.

### 35.8 Browser security headers

Production responses must define and test:

- Content Security Policy restricted to same-origin application needs;
- `Strict-Transport-Security`;
- `Referrer-Policy: no-referrer` on replay and private gameplay pages;
- `X-Content-Type-Options: nosniff`;
- `Permissions-Policy` disabling unused sensitive browser capabilities;
- `frame-ancestors 'none'` or an equivalent anti-framing policy;
- a restrictive `base-uri` and `form-action`.

Room-code and replay routes must be logged by route template or redacted path, including at the reverse proxy. Raw replay capabilities must never reach an access log.

---

## 36. Privacy architecture

Ninefold collects no product analytics.

### 36.1 No external tracking

Do not include:

- Google Analytics;
- Meta Pixel;
- advertising scripts;
- session replay;
- behavior heatmaps;
- third-party fonts;
- third-party embedded widgets on gameplay pages.

### 36.2 Logs

Never log:

- cookies;
- plaintext session tokens;
- standalone puzzle solutions;
- full WebSocket payloads by default;
- private Duel notes;
- raw replay capability tokens;
- raw Room-session or Solo-assignment proofs;
- raw URL fragments or unredacted capability-bearing paths;
- personal Solo history.

### 36.3 Data minimization

Server stores only data required for:

- Room operation;
- Match operation;
- temporary replay;
- abuse prevention;
- administration;
- operational reliability.

Retention enforcement:

- keep participant-linked Match events, snapshots, names, result projections, private payloads, and replay capabilities for at most seven days;
- early replay deletion immediately revokes capabilities and removes replay-accessible payloads without rewriting the sealed Match stream;
- scrub remaining participant-linked Match data at the seven-day boundary;
- retain only the field-limited non-identifying Match tombstone defined by `DOMAIN.md` for 30 days;
- delete expired Room sessions and command receipts when no active retry, reconnect, or replay-deletion authority depends on them.

### 36.4 Local clear action

Settings must expose a client-side action to clear:

- IndexedDB;
- local preferences;
- cached Solo data;
- local statistics;
- local replays.

If a service worker is introduced later, the action also clears applicable Cache Storage.

Server session cookie clearing is handled separately.

---

## 37. Performance budgets

### 37.1 Browser budgets

| Metric | Target |
|---|---:|
| Homepage compressed JS | under 100 KiB |
| Initial gameplay compressed JS | under 200 KiB |
| Replay tooling | lazy-loaded |
| Local cell feedback | under 50 ms |
| Home page interactive on normal mobile connection | under 3 seconds |
| Lighthouse performance on public pages | 95+ |
| Touch response | immediate visual feedback |
| Layout shift | minimal and budgeted |

### 37.2 Server budgets

| Metric | Target |
|---|---:|
| Typical command processing excluding network | under 20 ms |
| Same-region command acknowledgement | normally under 150 ms |
| Typical WebSocket message | under 2 KiB |
| Simultaneous connections | 100 |
| Active Rooms | 25 |
| Player capacity per Room | mode-specific plus spectators |
| Graceful brief peak | above target without data loss |

### 37.3 Techniques

- pre-generate puzzles;
- keep hot Match state in memory;
- persist only durable events;
- batch noncritical heatmap updates;
- lazy-load mode-specific UI;
- avoid heavyweight state libraries;
- self-host optimized fonts or use system fonts;
- use SVG icons;
- use HTTP compression;
- use immutable cache headers for hashed assets;
- avoid unnecessary polling;
- use prepared queries;
- keep transactions short.

---

## 38. Operational metrics

No product analytics, but operational metrics are required.

Expose a private Prometheus-compatible endpoint:

```text
/internal/metrics
```

Metrics:

- active WebSocket connections;
- active Room actors;
- active Matches;
- actor queue depth;
- connection outbound queue depth;
- command latency;
- command rejection by code;
- SQLite transaction duration;
- SQLite busy events;
- reconnect count;
- recovery success/failure;
- HTTP status totals;
- process memory;
- goroutine count;
- replay verification failures;
- puzzle assignment failures.

Metrics must not contain:

- display names;
- participant IDs as labels;
- Room codes;
- puzzle values;
- replay capabilities.

High-cardinality labels are prohibited.

---

## 39. Logging

Use structured JSON via `log/slog`.

Recommended fields:

- timestamp;
- level;
- service;
- request ID;
- connection ID;
- Room ID;
- Match ID;
- event type;
- latency;
- rejection code;
- recovery stage.

Do not include sensitive payloads.

Retention:

- ordinary logs: 14 days;
- security logs: 30 days;
- admin audit: 1 year;
- debugging payload capture: disabled by default.

---

## 40. Health and readiness

### 40.1 Liveness

```text
GET /health/live
```

Returns success when process event loop is alive.

Must not perform expensive dependency checks.

### 40.2 Readiness

```text
GET /health/ready
```

Returns success only when:

- configuration is valid;
- migrations are current;
- SQLite writer is healthy;
- server accepts new traffic;
- shutdown is not in progress.

### 40.3 Protected status

```text
GET /health/status
```

May include:

- SQLite version;
- migration version;
- actor count;
- queue saturation;
- recovery state;
- build version.

It must be private and must not expose secrets or active Room details.

---

## 41. Graceful shutdown

On termination:

1. mark readiness false;
2. reject new Room creation and joins;
3. notify connected clients of maintenance;
4. drain in-flight HTTP requests;
5. stop accepting new authoritative Match commands;
6. persist active snapshots;
7. close WebSockets with reconnectable maintenance reason;
8. checkpoint WAL;
9. close database;
10. exit before deadline.

Recommended shutdown deadline:

```text
60 seconds
```

If snapshot persistence fails, log prominently and preserve database integrity over speed.

---

## 42. Restart recovery

On startup:

1. load configuration;
2. open SQLite;
3. verify SQLite version;
4. enable PRAGMAs;
5. verify migrations;
6. scan nonterminal Matches;
7. reconstruct from snapshot and events;
8. validate invariants;
9. register recovered Room actors as RecoveryPending;
10. expose readiness;
11. allow reconnect;
12. resume or cancel according to domain rules.

For MVP Co-op:

- resume when at least one eligible player reconnects;
- cancel when nobody reconnects within five minutes;
- exclude the entire server-caused RecoveryPending interval from active elapsed time.

Race and Duel recovery is provisional.

---

## 43. Maintenance scheduler

One internal scheduler in the Go process performs:

- Room expiration;
- session cleanup;
- replay expiration;
- result retention cleanup;
- participant-data scrubbing;
- 30-day Match tombstone expiration;
- old snapshot cleanup;
- WAL checkpoint;
- `PRAGMA optimize`;
- integrity checks after suspicious recovery failures.

Do not deploy a separate worker service in V1.

Jobs must be:

- idempotent;
- bounded;
- cancellable;
- observable;
- safe to rerun after restart.

---

## 44. Backups

Current private-development decision:

- no off-VPS backup is required immediately;
- data loss is accepted during early private testing.

Release gate:

- broader public consumer promotion is blocked until off-VPS automated backup and restore testing exist.

Planned approach:

- Litestream or equivalent continuous SQLite replication;
- S3-compatible storage;
- daily snapshot;
- retention policy;
- encryption in transit and at rest;
- deletion behavior documented for retained backup generations;
- documented restore test.

Do not blindly copy only the SQLite main file while WAL is active.

---

## 45. Coolify deployment

### 45.1 Applications

Two Coolify applications:

```text
ninefold-web
ninefold-server
```

Both build from the same repository.

### 45.2 Persistent volume

```text
ninefold-data:/app/data
```

Mounted only into `ninefold-server`.

### 45.3 Routing

```text
/             → web
/api/*        → server
/ws           → server
/health/*     → server
/internal/*   → private
```

### 45.4 Initial resource limits

| Service | CPU | Memory |
|---|---:|---:|
| Web | 0.5–1 core | 256–512 MiB |
| Server | 1–2 cores | 512 MiB–1 GiB |

Reserve at least 5 GiB disk for the SQLite volume.

Tune after actual VPS details and load tests.

### 45.5 Go server settings

Initial values:

- read-header timeout: 5 seconds;
- read timeout: 15 seconds;
- write timeout: 30 seconds;
- idle timeout: 60 seconds;
- max WebSocket message: 64 KiB;
- `GOMAXPROCS`: container CPU allowance.

Replay download endpoints may use separate response limits.

---

## 46. Docker images

### 46.1 Web image

Multi-stage build:

1. Node build stage;
2. install exact lockfile dependencies;
3. build SvelteKit adapter-node output;
4. minimal runtime image;
5. non-root user;
6. read-only filesystem where practical.

### 46.2 Server image

Multi-stage build:

1. Go build stage;
2. tests and generation happen in CI, not image startup;
3. build static or minimally dynamic binary;
4. copy migrations if not embedded;
5. minimal runtime image;
6. non-root user;
7. writable `/app/data` only.

### 46.3 Image tags

```text
sha-<commit>
master
v0.x.y
```

Production deploys immutable SHA or semantic-version tags, not only `latest`.

---

## 47. CI pipeline

Required GitHub Actions checks:

```text
Go formatting
Go vet
Go unit tests
Go integration tests
Go race detector
Frontend lint
Frontend type checking
Frontend unit tests
Contract generation check
Contract fixture tests
Database migration test
TLA+ model check
Playwright smoke tests
Accessibility smoke tests
Container builds
Dependency vulnerability scan
```

### 47.1 Generated-file check

CI runs generation and fails when Git diff is non-empty.

### 47.2 Database test

CI must:

- create empty SQLite database;
- apply all migrations;
- verify foreign keys;
- verify WAL configuration;
- run repository integration tests;
- test upgrade from representative prior schema where available.

---

## 48. Deployment pipeline

```text
Pull request
→ CI
→ merge to master
→ build immutable images
→ push to GitHub Container Registry
→ Coolify deploy
→ readiness false
→ migrations
→ server startup and recovery
→ readiness true
→ smoke check
```

Production images are not manually built on the VPS.

### 48.1 Rollback

Retain previous web/server image pair.

Rollback:

1. stop readiness;
2. preserve snapshots;
3. deploy previous image pair;
4. verify schema compatibility;
5. recover Matches;
6. restore readiness.

---

## 49. Testing strategy

### 49.1 Unit tests

Domain tests are table-driven and avoid network/database dependencies.

Required coverage:

- Room transitions;
- readiness reset;
- host authorization;
- Co-op commands;
- Countdown cancellation;
- error presets and shared penalties;
- Nudge and Reveal hints;
- durable pings and ephemeral reactions;
- timers;
- reconnects;
- idempotency;
- Match completion;
- replay event application.

Race ranking, Duel scoring, and result invalidation tests are added with those features.

### 49.2 Property tests

Puzzle engine properties:

- valid solution;
- unique solution;
- transformation validity;
- deterministic grading;
- valid hints;
- replayable solver path.

### 49.3 SQLite integration

Use real temporary SQLite database.

Test:

- migrations;
- foreign keys;
- WAL;
- duplicate RequestID;
- version conflict;
- atomic commit;
- snapshot recovery;
- seven-day participant-data scrubbing;
- replay deletion and 30-day tombstone cleanup;
- command receipts;
- graceful restart.

### 49.4 WebSocket tests

Custom Go test client simulates:

- initialization;
- cookies;
- duplicate command;
- stale version;
- reconnect;
- event gap;
- slow reader;
- controller transfer;
- server restart.

### 49.5 Frontend tests

Vitest:

- reducers;
- optimistic reconciliation;
- timer correction;
- keyboard navigation;
- local persistence;
- replay reconstruction;
- English localization and pseudo-localized expansion;
- accessibility helpers.

Replay integrity tests use the same committed fixtures in Go and TypeScript and cover:

- RFC 8785 canonicalization;
- genesis and chained hashes;
- public payload tampering;
- hidden-payload digest commitments;
- event gaps and reordering;
- Ed25519 signatures;
- trusted and unknown key IDs;
- signing-key rotation.

### 49.6 End-to-end tests

Playwright multi-context scenarios:

1. Create and join Co-op Room.
2. Ready and start.
3. Simultaneous Co-op commands.
4. Refresh and reconnect.
5. Complete and replay.
6. Rematch.
7. Error presets and hints.
8. Ping/reaction persistence distinction.
9. Replay verification, tampering, and deletion.
10. Solo resume, hint, and completion validation.
11. Second-tab read-only behavior.
12. Go server restart and recovery with paused elapsed time.

### 49.7 Accessibility

Automated axe-core plus manual release checklist.

### 49.8 Load tests

k6 or custom Go client:

- 100 WebSockets;
- 25 Rooms;
- Room creation burst;
- countdown burst;
- simultaneous finishes;
- restart reconnect storm;
- slow spectators;
- SQLite write pressure.

---

## 50. TLA+ formal verification

Files:

```text
specification/Room.tla
specification/Room.cfg
specification/Match.tla
specification/Match.cfg
```

### 50.1 Model scope

Model:

- Room lifecycle;
- readiness;
- Match creation;
- duplicate RequestIDs;
- stale versions;
- stale timers;
- reconnects;
- recovery;
- completion.

Race finish and Duel turn ownership are added when those modes enter scope.

Do not model all 81 Sudoku cells. Use a reduced finite model.

### 50.2 CI

TLC runs on every change affecting:

- Room state machine;
- Match state machine;
- idempotency;
- timers;
- reconnect;
- completion;
- recovery.

### 50.3 Required invariants

At minimum:

- one active Match per Room;
- one host;
- one effect per RequestID;
- no completed-to-active transition;
- stale timer cannot mutate current state;
- no broadcast before commit as an abstract ordering property.

Mode-specific invariants are added with their modes.

---

## 51. Feature flags

The MVP does not create a feature-flag subsystem or table for features that do not exist.

When staged access is first required, use a small validated server configuration before considering persisted flags.

- flags may enable access;
- flags may not rewrite MatchRules of an existing Match;
- historical replay remains valid after flag changes;
- no third-party feature-flag platform.

---

## 52. Configuration

Parse environment once into one validated struct.

Canonical variables:

```text
NINEFOLD_ENVIRONMENT
NINEFOLD_PUBLIC_URL
NINEFOLD_HTTP_ADDRESS
NINEFOLD_DATABASE_PATH
NINEFOLD_ALLOWED_ORIGINS
NINEFOLD_COOKIE_SECRET
NINEFOLD_REPLAY_SIGNING_KEY
NINEFOLD_REPLAY_SIGNING_KEY_ID
NINEFOLD_ADMIN_PROXY_HEADER
NINEFOLD_LOG_LEVEL
NINEFOLD_REPLAY_RETENTION
NINEFOLD_MATCH_TOMBSTONE_RETENTION
NINEFOLD_COMMAND_RECEIPT_RETENTION
NINEFOLD_SHUTDOWN_TIMEOUT
```

Requirements:

- `NINEFOLD_ENVIRONMENT` is exactly `development`, `test`, or `production`;
- production public and allowed-origin URLs use HTTPS;
- the cookie secret is base64-encoded and decodes to at least 32 bytes;
- the replay signing key is a base64-encoded PKCS#8 Ed25519 private key;
- the replay signing key ID and administrator proxy header use bounded safe syntax;
- fail startup on missing mandatory values;
- reject placeholder secrets in production;
- reject retention values that exceed domain policy;
- do not read environment throughout business code;
- expose sanitized effective config in protected status.

---

## 53. SEO and performance delivery details

### 53.1 Asset strategy

- hashed immutable static assets;
- Brotli/gzip through proxy;
- explicit image dimensions;
- modern image formats;
- self-hosted fonts or system font stack;
- font subsets for supported languages;
- no icon font.

### 53.2 Code splitting

Lazy-load:

- replay;
- admin;

Race, Duel, advanced hints, and non-default locales are added to this list only after they exist.

Do not lazy-load critical board interaction after entering Match.

### 53.3 Caching

Public static pages may use short CDN/proxy caching with revalidation.

Never cache shared responses containing:

- Room session;
- participant state;
- replay capability;
- admin data.

---

## 54. Scaling boundary

SQLite and one Go instance are appropriate while Ninefold has:

- one VPS;
- one authoritative process;
- moderate writes;
- approximately 100 connections;
- approximately 25 active Rooms;
- no multi-region requirement;
- acceptable brief deployment interruption.

### 54.1 Migration triggers

Reconsider PostgreSQL and distributed runtime state when any is required:

- multiple Go replicas;
- rolling zero-downtime backend deployment;
- multi-machine game hosting;
- high sustained write concurrency;
- independent services needing database access;
- automatic failover;
- regional placement.

### 54.2 Migration preparation

To preserve migration ability:

- domain never imports SQLite;
- repositories expose aggregate-oriented interfaces;
- SQL is isolated;
- UUIDv7 identifiers are portable;
- timestamps are UTC integers;
- events are database-agnostic JSON payloads;
- no SQLite-specific behavior leaks into contracts.

Do not prematurely implement distributed infrastructure.

---

## 55. Architecture decision process

Create an ADR for decisions expensive to reverse.

This architecture specification is the initial decision record. Do not duplicate its baseline decisions into boilerplate ADRs.

An ADR is required before changing:

- database;
- authoritative topology;
- actor model;
- event persistence ordering;
- cryptographic proof;
- identity model;
- public API versioning;
- client persistence technology.

---

## 56. Local development

Root commands:

```text
make dev
make test
make lint
make generate
make migrate
make puzzles
make e2e
make build
make tla
```

### 56.1 `make dev`

Starts:

- SvelteKit dev server;
- Go server;
- local SQLite initialization;
- migrations;
- contract watch where practical.

### 56.2 Local HTTP

Ordinary local development may use HTTP on localhost.

Production cookie and Origin behavior must be configurable without scattered code exceptions.

A local reverse-proxy HTTPS test is required before release.

---

## 57. Definition of architecture-complete

A feature is architecture-complete only when:

- domain rule exists;
- transport contract exists;
- authoritative command path exists;
- persistence ordering is correct;
- event reducer exists;
- reconnect behavior is defined;
- error codes are stable;
- accessibility semantics are available;
- localization keys exist;
- performance impact is reviewed;
- tests exist at appropriate layers;
- metrics/logging are safe;
- documentation is updated;
- TLA+ is updated when concurrency-sensitive.

---

## 58. Explicit implementation prohibitions

AI agents and contributors must not:

- send standalone puzzle-solution artifacts to multiplayer clients;
- broadcast before persistence commit;
- mutate Room or Match outside actor serialization;
- place domain logic in Chi handlers;
- place server authority in Svelte stores;
- use client clocks for results;
- queue offline multiplayer moves;
- expose participant names in Room preview;
- use SQLite from the web container;
- mount the database volume into multiple writers;
- horizontally scale the Go service;
- add external analytics;
- add client-side SQLite;
- log secrets or full gameplay payloads;
- make replay public by Match ID;
- skip version checks for convenience;
- remove historical events to “fix” a result;
- weaken TLA+ invariants to satisfy an implementation;
- add microservices or distributed systems only for portfolio novelty.

---

## 59. Implementation sequence

Recommended architecture implementation order:

1. Monorepo and tooling
2. Domain primitives
3. Puzzle validator and solver
4. SQLite migrations and repositories
5. Room aggregate
6. Co-op Match aggregate
7. HTTP Room creation/join
8. WebSocket protocol
9. Room actor registry
10. Svelte lobby
11. Semantic Sudoku board
12. Co-op event reducer
13. persistence-before-broadcast path
14. reconnect and snapshot recovery
15. Match completion
16. replay reconstruction
17. replay hash chain
18. Ed25519 sealing
19. TLA+ model
20. Solo
21. MVP administration and public hardening
22. Race after a focused domain review
23. Duel after a focused domain review
24. Daily Ninefold after a focused domain review
25. remaining public-ready features

---

## 60. Final architectural directive

Ninefold is successful architecturally when it can provide an excellent multiplayer game on a small VPS while remaining:

- server-authoritative;
- deterministic;
- recoverable;
- privacy-preserving;
- accessible;
- fast;
- simple to operate;
- formally reasoned about;
- cryptographically verifiable.

The architecture must remain purposeful. Sophistication is justified only when it improves correctness, privacy, accessibility, performance, recovery, or portfolio credibility without imposing disproportionate operational cost.
