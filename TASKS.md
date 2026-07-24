# Ninefold Sudoku implementation tasks

**Current target:** `0.3.0` full MVP / portfolio beta

**Execution model:** sequential phases, exactly one commit per completed phase

**Current status:** documentation complete; application implementation has not started

## 1. Authority and scope

This file is an execution ledger. It does not override the canonical specifications.

Before implementing a task, read the cited sections in this authority order:

1. [Domain specification](docs/DOMAIN.md) — gameplay rules and invariants
2. [Architecture specification](docs/ARCHITECTURE.md) — implementation constraints
3. [Product specification](docs/PRODUCT.md) — current scope and outcomes
4. [Design specification](docs/DESIGN.md) — interaction and presentation
5. [AGENTS.md](AGENTS.md) — contributor workflow and Definition of Done

If this file conflicts with a canonical document, the canonical document wins. Do not infer a new rule or silently choose between contradictory requirements. Stop, update the appropriate canonical document, and create an ADR under `docs/decisions/` only when the decision meets the ADR threshold in Architecture §55.

`TASKS.md` contains executable work only for `0.3.0`. Race, Duel, Daily Ninefold, offline Solo, Explain hints, host approval, multiplayer undo, full spectator UX, additional locales, PWA installation, About, dedicated mode pages, public matchmaking, accounts, and off-VPS backup are not implementation tasks here.

## 2. Phase execution protocol

Every coding agent must follow this protocol:

1. Start from the completed commit of the preceding phase and a clean working tree.
2. Read every canonical section listed by the phase.
3. Complete phase tasks in task-ID order unless a task explicitly permits parallel work.
4. Add tests with the behavior they protect; do not postpone tests to a later phase.
5. Keep all phase work uncommitted until every task and phase gate passes.
6. Run the phase-specific checks and every cumulative check that exists at that point.
7. Preserve required evidence in CI artifacts or the explicitly required release report, then mark completed checkboxes. Do not commit volatile raw logs.
8. Review the complete staged diff for secrets, generated-file drift, unrelated changes, and deferred features.
9. Create exactly one commit using the phase’s exact commit message.
10. Confirm the working tree is clean before starting the next phase.

Do not create WIP, fixup, formatting-only, or test-only commits inside a phase. If a gate fails, fix it before committing. If a later phase discovers an integration defect in earlier behavior, fix and test it in the current phase and describe the integration correction in that phase’s commit body; do not rewrite published history.

Checkbox meanings:

- `[ ]` not completed
- `[x]` completed and verified

A phase is incomplete while any task or gate remains unchecked.

## 3. Global engineering gates

These gates apply continuously, not only in the final phase:

- Domain packages import no HTTP, WebSocket, SQLite, Svelte, or UI packages.
- HTTP and WebSocket handlers contain transport parsing, authentication, and mapping only.
- Every durable command is idempotent through a UUIDv7 RequestID and durable command receipt.
- Room and Match mutations run through the Room actor.
- Durable events commit before acknowledgement or broadcast.
- Multiplayer clients never receive a standalone solution artifact.
- Client time is visual; server time decides deadlines, elapsed results, and ordering.
- Completed Matches never return to an active state.
- Historical Match events are append-only and schema-versioned.
- Replay reconstruction is deterministic.
- Privacy-sensitive material never appears in logs, metrics, URLs sent to servers, fixtures, snapshots, or test output.
- Accessibility, keyboard operation, mobile layout, dark mode, reduced motion, and pseudo-localized expansion are implemented with each interactive feature.
- Performance budgets in Product §26 and Architecture §37 are measured when the affected path first exists.
- Generated contracts are committed, carry the generated-file header, and reproduce without a diff.
- No phase adds infrastructure or dependencies solely for future use.

Root commands must progressively become real checks as their subsystems are introduced:

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

Targets must fail honestly when their required tool or check fails. Do not use placeholder targets that always succeed.

---

## Phase 1 — Repository and quality foundation

### Goal

Create a reproducible monorepo, pinned toolchains, validated configuration boundary, root developer commands, and a CI foundation without creating empty feature packages.

### Canonical references

- Architecture §§3, 6–7, 17, 46–47, 52, 56
- Product §§26, 34
- `AGENTS.md` Stack, Workflow, and Definition of Done

### Tasks

- [x] **P01-T01 — Establish workspace metadata.** Add the Go workspace, pnpm workspace, root package metadata, exact lockfiles, editor-neutral formatting rules, `.gitignore`, `.env.example`, and license-preserving package metadata. Pin supported Go and Node/pnpm versions in committed toolchain configuration.
- [x] **P01-T02 — Scaffold only active applications.** Create `apps/server` and `apps/web` with the minimum files needed to compile and run. Do not create empty Race, Duel, Daily, reporting, PWA, or future-locale modules.
- [x] **P01-T03 — Bootstrap the Go service.** Add `apps/server/go.mod`, a minimal `cmd/api` entry point, internal platform configuration package, graceful context wiring, and a deterministic build-version value. The service may expose only a minimal development liveness response in this phase.
- [x] **P01-T04 — Bootstrap the SvelteKit application.** Configure SvelteKit, TypeScript strict mode, Vite, Tailwind CSS, shadcn-svelte prerequisites, adapter-node, ESLint, Prettier, Vitest, and Playwright. Render a minimal semantic page without third-party network assets.
- [x] **P01-T05 — Implement one configuration boundary.** Parse environment variables once into a validated Go struct. Reject missing mandatory values, malformed URLs/origins, invalid timeouts, and production placeholder secrets. Business packages must not read environment variables directly.
- [x] **P01-T06 — Add root commands.** Implement the root `Makefile` commands defined in Architecture §56. Commands for not-yet-existing subsystems must explain that the subsystem is not present and return a non-success status when invoked as a required gate; convert them to real checks in the owning phase.
- [x] **P01-T07 — Add baseline CI.** Create GitHub Actions jobs for Go formatting/vet/test, frontend formatting/lint/typecheck/unit test, and both builds. Pin action revisions, use lockfile-frozen installs, enable dependency caching without caching secrets, and cancel superseded branch runs.
- [x] **P01-T08 — Add dependency maintenance.** Configure Dependabot or the repository-approved equivalent for Go, pnpm, and GitHub Actions with bounded update frequency. Do not auto-merge dependency changes.
- [x] **P01-T09 — Add test foundations.** Establish Go table-test helpers, frontend test setup, deterministic timezone/locale defaults, and temporary-directory conventions. Tests must not rely on network access or developer-global state.
- [x] **P01-T10 — Document local startup.** Update README development instructions with prerequisites, environment setup, root commands, ports, and the fact that implementation is incomplete. Do not claim features that do not exist.

### Phase gates

- [x] `go test ./...`, `go vet ./...`, and Go formatting checks pass from the server module.
- [x] Frozen pnpm install, frontend lint, formatting check, typecheck, unit tests, and production build pass.
- [x] Root `make test`, `make lint`, and `make build` execute the real available checks.
- [x] Server startup fails safely for invalid configuration and succeeds with development configuration.
- [x] Web and server artifacts contain no production secret, absolute developer path, or unexpected external request.
- [x] CI reproduces local checks from a clean checkout.
- [x] `git diff --check` passes and the staged diff contains no unrelated feature implementation.

### Commit

```text
chore(repo): scaffold production monorepo

- pin Go and pnpm toolchains and workspace dependencies
- add root development, generation, build, and test commands
- establish baseline CI and validated configuration
```

---

## Phase 2 — Domain primitives and generated contracts

### Goal

Create transport-independent domain foundations and one versioned contract-generation path shared by Go and TypeScript.

### Canonical references

- Domain §§3–7, 13, 15, 29–33
- Architecture §§5, 7–9, 12.8–12.9, 17
- Product §27

### Tasks

- [ ] **P02-T01 — Implement strong identifiers and values.** Add validated types for RoomID, RoomCode, ParticipantID, MatchID, PuzzleID, ReplayID, RequestID, aggregate version, event number, cell index, digit, candidate set, timestamps, mode, difficulty, role, and current-scope lifecycle states. Keep parsing at boundaries and prohibit unvalidated primitive aliases in domain APIs.
- [ ] **P02-T02 — Implement identity validation.** Validate display names using the Domain §7.2 grapheme, Unicode normalization, comparison, and structural rules. Preserve the NFC display form, compare with the NFKC case-folded form, and keep moderation policy outside automatic validation.
- [ ] **P02-T03 — Define aggregate boundaries.** Add Room, Match, Puzzle, participant, rules, result, command, and event interfaces without persistence or transport imports. Model current Co-op/Solo scope only; deferred mode constants may exist only when required for stable stored-enum compatibility and must not enable behavior.
- [ ] **P02-T04 — Define stable errors.** Implement every current-scope Domain §33 error code with safe structured details and retryability metadata. Domain errors contain no localized prose, HTTP status, or UI behavior.
- [ ] **P02-T05 — Define command and event metadata.** Add UUIDv7 RequestID, authenticated scope, expected aggregate version, client sequence, server occurrence time, schema version, and event number rules. Separate RoomVersion from MatchVersion.
- [ ] **P02-T06 — Establish contract sources.** Add the OpenAPI root, WebSocket schema roots, replay schema roots, fixture directories, and generation configuration. Schemas must reject unknown required-shape violations while preserving explicit version evolution.
- [ ] **P02-T07 — Generate typed contracts.** Generate committed Go and TypeScript types from source schemas with `Code generated from contracts. DO NOT EDIT.` headers. Generated types must not become domain entities.
- [ ] **P02-T08 — Add cross-language fixtures.** Commit fixtures for successful envelopes, each error family, maximum safe integers, Unicode display names, stale versions, and unknown schema versions. Validate the same fixtures in Go and TypeScript.
- [ ] **P02-T09 — Enforce dependency rules.** Add a static or test-based import check proving domain packages do not import transport, persistence, configuration, generated transport, or UI packages.
- [ ] **P02-T10 — Document contract evolution.** Record additive/minor and breaking/major rules, unknown-event handling, fixture update requirements, and the rule that schema changes land with producer and consumer handling in the same phase.

### Phase gates

- [ ] Domain unit tests cover every value boundary, Unicode name rule, lifecycle enum, and current error code.
- [ ] UUIDv7 generation and parsing tests verify uniqueness assumptions and invalid input rejection.
- [ ] Go and TypeScript decode the shared fixtures to equivalent typed values.
- [ ] `make generate` completes and a second run produces no diff.
- [ ] Contract validation rejects missing identifiers, unsafe numeric ranges, malformed versions, and unexpected credential fields.
- [ ] Dependency-rule checks prove the domain remains transport- and storage-independent.
- [ ] All Phase 1 checks continue to pass.

### Commit

```text
feat(domain): add core domain contracts

- define identifiers, aggregates, commands, events, and errors
- generate versioned Go and TypeScript contracts from schemas
- enforce transport-free domain dependencies with unit tests
```

---

## Phase 3 — Puzzle engine and verified catalog

### Goal

Provide deterministic, tested Sudoku validation, solving, grading, transformation, hints, generation, and catalog ingestion outside production request paths.

### Canonical references

- Product §§16–17
- Domain §§15, 17, 23
- Architecture §26 and §§37.3, 49.2

### Tasks

- [ ] **P03-T01 — Implement compact board representation.** Represent clues and solutions as validated 81-cell structures with explicit empty-cell handling. Reject invalid length, digits, clue/solution disagreement, and duplicate row/column/box values.
- [ ] **P03-T02 — Implement deterministic validation and solving.** Provide candidate calculation, constraint propagation, deterministic search order, solution validation, and solution counting capped once non-uniqueness is proven.
- [ ] **P03-T03 — Verify uniqueness.** Accept catalog puzzles only when exactly one valid solution exists. Return structured tooling errors that identify the violated invariant without printing a full solution in ordinary logs.
- [ ] **P03-T04 — Implement logical techniques and grading.** Implement the techniques and deterministic grade boundaries defined by Domain §23.4. Store the grading algorithm version with every puzzle revision.
- [ ] **P03-T05 — Implement Nudge and Reveal support.** Produce deterministic useful-cell selection and reveal data suitable for server-side authorization. Do not implement deferred Explain hints.
- [ ] **P03-T06 — Implement canonical transformations.** Add digit permutation, row/column band/stack transformations, transpose/rotation/reflection rules, and proof that clues and solution transform together.
- [ ] **P03-T07 — Implement catalog tooling.** Add `cmd/puzzle-generator`, seedable deterministic generation, validation, grading, duplicate/canonical-equivalence detection, revision metadata, and import/export suitable for code review.
- [ ] **P03-T08 — Add verified fixtures.** Commit representative Easy, Medium, Hard, and Expert puzzles with immutable IDs/revisions, clues, server-only solutions, grades, and solver paths. Public/frontend fixtures must contain clues only unless a test is explicitly server-confined.
- [ ] **P03-T09 — Keep generation off request paths.** Expose catalog assignment interfaces that read pre-generated verified puzzles. Production HTTP handlers must never invoke puzzle generation or uniqueness search.
- [ ] **P03-T10 — Add property and fuzz tests.** Cover solver determinism, unique solutions, transformations, invalid encodings, hint validity, grade stability, and generated-puzzle replayability.
- [ ] **P03-T11 — Add benchmarks.** Benchmark validation, solving, grading, and hint selection using representative worst-case catalog candidates. Record regression thresholds that keep puzzle work outside gameplay latency paths.

### Phase gates

- [ ] Every committed puzzle passes validity, uniqueness, clue consistency, deterministic grading, and logical-completion checks.
- [ ] Property tests prove every transformation preserves the solution and uniqueness.
- [ ] Fuzz tests do not panic, hang, allocate without bound, or accept malformed boards.
- [ ] Frontend/generated public artifacts contain no full solution fixtures.
- [ ] `make puzzles` validates the catalog deterministically and produces no diff on a second run.
- [ ] Puzzle benchmarks are recorded and do not run in the production request path.
- [ ] All cumulative checks pass.

### Commit

```text
feat(puzzle): add verified puzzle engine

- validate, solve, grade, transform, and hint standard puzzles
- add deterministic catalog tooling and verified fixtures
- keep solutions server-only and generation off request paths
```

---

## Phase 4 — SQLite persistence foundation

### Goal

Create WAL-safe, migration-driven SQLite persistence with atomic command receipts, append-only Match events, snapshots, and repository boundaries suitable for Room actors.

### Canonical references

- Domain §§6, 24–25, 32, 35
- Architecture §§18–23, 25, 35.1, 36.3, 43
- Architecture §§47.2, 49.3

### Tasks

- [ ] **P04-T01 — Configure the approved driver.** Pin `modernc.org/sqlite`, verify the embedded SQLite version at startup, and fail readiness when it violates the approved WAL safety policy.
- [ ] **P04-T02 — Open separate database handles.** Configure one writer connection and a small reader pool. Apply and verify `WAL`, foreign keys, `busy_timeout=5000`, and `synchronous=FULL` on every relevant connection.
- [ ] **P04-T03 — Add ordered migrations.** Create migrations for every current MVP table listed in Architecture §20 with UUIDv7 text IDs, UTC millisecond timestamps, foreign keys, uniqueness constraints, status checks, and deletion behavior consistent with retention.
- [ ] **P04-T04 — Encode grids compactly.** Store puzzle clues and solutions as validated 81-byte blobs. Prevent JSON-array grid storage and prevent ordinary public queries from selecting solution blobs.
- [ ] **P04-T05 — Implement SQLC repositories.** Add aggregate-oriented repositories for puzzles, Rooms, participants, blocks, sessions, Matches, events, snapshots, results, command receipts, replay capabilities/seals, tombstones, and admin audit records. Keep SQL and SQLite types inside persistence packages.
- [ ] **P04-T06 — Implement atomic Room writes.** Persist the Room projection, participant changes, aggregate version, and command receipt in one transaction guarded by the expected version.
- [ ] **P04-T07 — Implement atomic Match writes.** Persist ordered events, projection changes, Match/Room lifecycle changes, results, next event number, aggregate version, and command receipt in one transaction. A failed guard updates nothing.
- [ ] **P04-T08 — Implement idempotent receipts.** Store scope hash, command type, request fingerprint, terminal outcome, safe response, and expiration. Reject reuse with a different scope/type/fingerprint and replay an identical terminal outcome after restart.
- [ ] **P04-T09 — Implement snapshots.** Store compressed, versioned Match state with event number, aggregate version, timestamp, and integrity metadata. Start with standard gzip; do not add another codec without benchmark evidence.
- [ ] **P04-T10 — Prepare replay integrity fields.** Persist public envelopes, nullable private payload/salt/digest, previous hash, event hash, and terminal seal metadata exactly as Architecture §§20.3 and 20.6 require. Ordinary replay queries may return public data and digests only.
- [ ] **P04-T11 — Add retention-safe queries.** Add bounded queries for session/receipt expiry, seven-day participant-data scrubbing, replay revocation/deletion, 30-day tombstones, and referential cleanup without scheduling them yet.
- [ ] **P04-T12 — Add migration and repository integration tests.** Test empty-database migration, rollback where supported, foreign keys, WAL, duplicate RequestID, version conflict, transaction failure, snapshot round-trip, and retention behavior with a real temporary database.

### Phase gates

- [ ] A clean database migrates to the latest schema and reports the expected version.
- [ ] The migration test upgrades from every committed representative prior schema.
- [ ] Foreign-key and constraint violations fail atomically.
- [ ] No test uses an in-memory SQLite mode that behaves differently from file-backed WAL.
- [ ] Concurrent repository tests show one writer, bounded busy handling, and no lost update.
- [ ] Public repository methods cannot return solution blobs, session tokens, replay capabilities, salts, or private payloads accidentally.
- [ ] `make migrate`, SQLC generation, integration tests, and every cumulative check pass.

### Commit

```text
feat(storage): add SQLite persistence

- add migrations, SQLC repositories, and WAL-safe setup
- persist rooms, matches, events, receipts, snapshots, and replays
- verify atomic writes, recovery, and retention constraints
```

---

## Phase 5 — Authoritative Room lifecycle

### Goal

Implement secure account-free Room creation and joining, host authority, Lobby readiness, Countdown, and actor-serialized Room mutations.

### Canonical references

- Domain §§7–12, 27, 30.1, 31.1, 33–34, 37.1–37.2, 37.14
- Architecture §§9–10, 13, 16, 19.2–19.3, 35
- Product §§14–15, 29.1

### Tasks

- [ ] **P05-T01 — Generate secure Room codes.** Use a CSPRNG, the canonical alphabet and length, a database uniqueness constraint, bounded collision retry, and the no-reuse policy. Normalize lowercase input without accepting ambiguous characters.
- [ ] **P05-T02 — Create temporary participants and hosts.** Create the host atomically with the Room, validate the display name, assign a non-color-only marker, and persist only required temporary identity.
- [ ] **P05-T03 — Implement hashed Room sessions.** Issue the `ninefold_room_session` cookie with production attributes, store only its hash, rotate it on sensitive transitions, and never expose it to JavaScript or logs.
- [ ] **P05-T04 — Enforce one active Room.** Reject create/join with `ACTIVE_ROOM_SESSION_EXISTS` while another Room session is active. Require an explicit leave/replacement request and return safe consequence data for UI confirmation.
- [ ] **P05-T05 — Implement safe preview and joining.** Return only mode, difficulty, Room state, seat availability, spectator availability, and lock state. Never expose names before successful join. Unlocked valid joins are immediate; host approval is absent.
- [ ] **P05-T06 — Implement Lobby state.** Model participant seats, host, current Co-op settings, lock state, ready flags, aggregate version, and activity timestamps. Do not introduce a separate ReadyCheck state.
- [ ] **P05-T07 — Implement host controls.** Authorize settings, lock/unlock, remove/block, host transfer, start/cancel, and later rematch hooks. Define deterministic host transfer on explicit leave/disconnect according to Domain §11.
- [ ] **P05-T08 — Implement readiness rules.** Require every seated player to be ready. Reset readiness on relevant setting/roster changes. Reject stale and unauthorized readiness mutations.
- [ ] **P05-T09 — Implement Countdown.** Atomically create the prepared Match with immutable MatchRules and a verified puzzle assignment when Countdown starts, create a generation-tagged three-second server timer, freeze seats/settings, cancel on canonical invalidating changes, return to Lobby with readiness reset, and never expose the assigned board before activation.
- [ ] **P05-T10 — Implement the Room actor registry.** Guarantee one active actor per Room, use bounded priority-aware queues, serialize every Room mutation, reject overload safely, and deactivate only when no timers/connections/in-flight work require the actor.
- [ ] **P05-T11 — Implement Room HTTP endpoints.** Add create, safe preview, join, resume, and explicit leave/replacement endpoints with typed OpenAPI contracts, JSON-only state changes, Origin validation, Idempotency-Key, structured safe errors, and thin Chi handlers.
- [ ] **P05-T12 — Implement Room TLA+ specification.** Model one host, one active Match reference, readiness, settings freeze, duplicate RequestIDs, stale versions, stale countdown generations, cancellation, and rematch return to Lobby.
- [ ] **P05-T13 — Add Room acceptance tests.** Cover code collision, Unicode names, duplicate normalized names, blocked participants, full seats, locked Rooms, one-active-Room replacement, host transfer, readiness reset, simultaneous start/settings, stale timers, and restart-safe receipts.
- [ ] **P05-T14 — Preserve minimal spectator protocol support.** Model the spectator role, read-only authorization, seat accounting, permitted Lobby conversion, and safe availability fields in domain/contracts. Do not add the deferred active spectator screens or full live-spectating experience.

### Phase gates

- [ ] HTTP contract fixtures and generated clients reproduce without drift.
- [ ] Domain tests cover every Room transition and current Room error code.
- [ ] Real-SQLite integration tests prove Room projection and receipt atomicity.
- [ ] Actor concurrency tests prove no double host, double start, stale timer mutation, or unbounded queue.
- [ ] TLC completes the bounded Room model with all Domain §34 current invariants.
- [ ] Security tests verify cookie attributes, Origin enforcement, preview minimization, rate limits, and redacted routes.
- [ ] All cumulative checks pass.

### Commit

```text
feat(room): implement room lifecycle

- add secure room identity, sessions, and one-room enforcement
- implement lobby, host, readiness, countdown, and join flows
- serialize mutations and model room invariants in TLA+
```

---

## Phase 6 — Authoritative Co-op Match domain

### Goal

Implement every current Co-op rule as deterministic domain commands and events, including completion and recovery-safe timers, without transport or UI logic.

### Canonical references

- Domain §§13–18, 26, 28–34, 37.3–37.5, 37.15–37.16
- Architecture §§9, 19.1, 21, 49.1, 50
- Product §§11.1, 17, 29.1

### Tasks

- [ ] **P06-T01 — Enforce immutable MatchRules.** Load the Co-op mode, difficulty, puzzle revision/transformation, participants, error preset, hint policy, notes policy, and schema versions snapshotted when Phase 5 created the prepared Match. Reject any later mutation or activation mismatch.
- [ ] **P06-T02 — Activate the prepared Match atomically.** Load the immutable puzzle assignment and MatchRules fixed when Phase 5 entered Countdown, reject any changed/stale preparation, and transition Room to InMatch plus Match to Active in one transaction. Never reassign at activation or include the solution in public state.
- [ ] **P06-T03 — Implement board commands.** Validate clues, cell/digit ranges, role, Match state, expected version, and direct Sudoku conflicts. Accept placement and erase only through authoritative commands.
- [ ] **P06-T04 — Implement correctness presets.** Implement Casual, Challenge, Blind, and Clean exactly as Domain §8.6 defines, including Challenge rejection, mistake count, and five-second shared penalty separated from active elapsed time.
- [ ] **P06-T05 — Implement shared notes.** Add deterministic note toggles/sets, ownership attribution, clue/value constraints, and removal behavior. Do not add multiplayer undo.
- [ ] **P06-T06 — Implement hints.** Add Nudge and Reveal authorization, deterministic selection, assistance flags, counts, and penalties. Do not expose Explain or standalone solution data.
- [ ] **P06-T07 — Implement attribution and simultaneous-edit resolution.** Attribute accepted values/notes to participants, allow soft-lock override, process commands in actor order, reject stale expected versions, and return enough safe state for client reconciliation.
- [ ] **P06-T08 — Split durable and ephemeral social behavior.** Persist targeted cell pings without incrementing MatchVersion when no authoritative state changes. Keep reactions, focus, soft locks, pointer motion, and animation out of SQLite and replay.
- [ ] **P06-T09 — Implement server-authoritative elapsed time.** Store server start/pause/penalty intervals, generation-tag timers, and result values. Client timestamps may not affect acceptance, completion, or ordering.
- [ ] **P06-T10 — Implement automatic completion.** Detect the accepted final correct move, finalize once, write the result and terminal event atomically, transition Room to Results, and prohibit completed-to-active transitions.
- [ ] **P06-T11 — Implement deterministic event application.** Reconstruct Match state from assignment, rules, snapshot, and ordered events. Reject gaps, duplicates with different payloads, invalid schema versions, and invariant-breaking histories.
- [ ] **P06-T12 — Implement Match TLA+ specification.** Model idempotency, stale versions/timers, actor ordering, persistence-before-broadcast as an ordering property, completion once, and no completed-to-active transition.
- [ ] **P06-T13 — Add exhaustive domain tests.** Cover all commands, error presets, hint limits, notes, pings/reactions distinction, simultaneous placements, clues, unauthorized roles, completion, stale commands, duplicate RequestIDs, and deterministic reconstruction.

### Phase gates

- [ ] Domain tests use no network or database packages and cover every current Co-op command/event/error.
- [ ] Persistence integration tests prove Match events/projections/results/receipts commit atomically.
- [ ] Replay-application tests reconstruct byte-equivalent authoritative state from fixtures.
- [ ] Tests prove pings are durable but version-neutral and reactions/focus never persist.
- [ ] Tests prove a standalone solution cannot appear in public Match state, events, snapshots, errors, or logs.
- [ ] TLC completes the Match model and required invariants.
- [ ] Typical domain command processing benchmarks remain comfortably below the 20 ms server budget before I/O.
- [ ] All cumulative checks pass.

### Commit

```text
feat(coop): implement authoritative match rules

- add board, notes, errors, hints, attribution, and completion
- enforce idempotent versioned commands and durable events
- model completion, timer, and command invariants
```

---

## Phase 7 — Committed realtime transport

### Goal

Expose Room and Match commands through a bounded native WebSocket protocol that persists durable outcomes before acknowledgement or broadcast.

### Canonical references

- Domain §§28, 30–34
- Architecture §§10–15, 17, 20.3, 21, 25.1–25.2, 35.2–35.4
- Architecture §§37, 49.4

### Tasks

- [ ] **P07-T01 — Implement the WebSocket endpoint.** Authenticate from the Room session cookie, validate Origin, enforce message size, set heartbeat/read/write deadlines, and isolate one read pump and one write pump per connection.
- [ ] **P07-T02 — Implement typed initialization.** Accept last RoomVersion, MatchID, and Match event number; return current identity, controller state, Room snapshot, Match snapshot/events, server time, and protocol versions without private data.
- [ ] **P07-T03 — Implement command envelopes.** Require RequestID, client sequence, command type, expected RoomVersion or MatchVersion, and typed payload. Derive participant, role, host status, RoomID, and MatchID from authenticated server state.
- [ ] **P07-T04 — Implement acknowledgements and rejections.** Return the durable terminal outcome, authoritative versions/event number, safe error code/details, and request correlation. Never acknowledge a proposed but uncommitted mutation.
- [ ] **P07-T05 — Enforce commit ordering.** Route authoritative commands to the Room actor, execute the Phase 4 transaction, update committed actor state, then broadcast and acknowledge. On failure, preserve prior state and send no success event.
- [ ] **P07-T06 — Implement command-status recovery.** Allow a timed-out client to query a RequestID outcome. Return committed success/rejection, pending, or unknown without resubmitting a possibly durable command blindly.
- [ ] **P07-T07 — Implement aggregate-version handling.** Keep RoomVersion, MatchVersion, and Match event number distinct. Broadcast durable version-neutral pings with event numbers and broadcast ephemeral messages without event numbers.
- [ ] **P07-T08 — Implement controller leases.** Permit one controlling tab per Room session, deterministic transfer, read-only secondary tabs, lease expiry, and safe rejection of gameplay commands from non-controller connections.
- [ ] **P07-T09 — Bound all queues.** Add actor and per-connection queue capacities, priority for lifecycle/shutdown messages, safe `SERVER_BUSY` behavior, slow-reader disconnect policy, and metrics hooks without high-cardinality labels.
- [ ] **P07-T10 — Add realtime rate limits.** Enforce the canonical per-session limits for values, notes, focus, social messages, settings, and Room creation. Keep ephemeral overload from delaying durable gameplay.
- [ ] **P07-T11 — Start the replay hash chain.** Canonicalize each durable public event envelope with RFC 8785, calculate hidden-payload commitments using a CSPRNG salt of at least 128 bits where applicable, chain SHA-256 hashes, and persist the hash fields in the same transaction. Do not sign until Phase 12.
- [ ] **P07-T12 — Generate protocol clients and fixtures.** Update schemas/generated Go/TypeScript types and fixtures for initialization, every current command/event, gaps, stale versions, controller transfer, busy handling, and unknown compatible events.
- [ ] **P07-T13 — Build a Go WebSocket test client.** Support cookies, typed messages, intentional duplicate/stale commands, disconnects, slow reads, event capture, and deterministic assertions for integration/load tests.

### Phase gates

- [ ] Tests prove no acknowledgement or broadcast occurs before the corresponding SQLite commit.
- [ ] Duplicate RequestIDs across reconnect return exactly the original terminal outcome.
- [ ] Stale versions, stale client sequences, wrong controllers, spectators, and malformed payloads cannot mutate state.
- [ ] Slow readers and full queues are bounded and cannot block Room actors.
- [ ] Origin, cookie, payload-limit, heartbeat, and rate-limit tests pass.
- [ ] Hash-chain fixtures reproduce identically in Go and TypeScript.
- [ ] Typical messages remain below 2 KiB or carry a documented reason for an allowed larger snapshot.
- [ ] A 100-connection transport smoke test shows no unbounded goroutine, queue, or memory growth.
- [ ] All cumulative checks pass.

### Commit

```text
feat(realtime): add committed event transport

- add bounded WebSocket transport and controller leases
- persist and hash events before acknowledgement or broadcast
- handle versions, gaps, receipts, and slow clients
```

---

## Phase 8 — Web shell, entry flows, and Lobby

### Goal

Build the public shell and accessible Room entry/Lobby experience on typed HTTP and WebSocket clients.

### Canonical references

- Product §§13–15, 20–26, 29.1
- Design §§4–20, 37–41, 43–48
- Architecture §§27–34, 53

### Tasks

- [ ] **P08-T01 — Implement design foundations.** Add semantic color, typography, spacing, radius, focus, motion, player-marker, light, and dark tokens. Use system/self-hosted assets only and no icon font.
- [ ] **P08-T02 — Implement the application shell.** Add semantic landmarks, skip link, responsive navigation, theme control, connection-status region, error boundary, and route-level loading/empty/error states.
- [ ] **P08-T03 — Establish localization.** Use English translation keys with named placeholders and locale-independent machine values. Add a pseudo-locale used only in development/tests, support expansion without truncation, and never concatenate translated sentence fragments.
- [ ] **P08-T04 — Build typed HTTP and WebSocket clients.** Use generated contracts, centralized safe-error mapping, RequestID creation, command status recovery, cookie credentials, reconnect hooks, and no duplicated gameplay rules.
- [ ] **P08-T05 — Build the SSR homepage.** Present Create Room, Join by code, current Co-op messaging, clearly unavailable future modes, and privacy/accessibility highlights in meaningful HTML without requiring JavaScript to understand the product. Do not render the Play Solo action until Phase 13 activates the working Solo route.
- [ ] **P08-T06 — Build Create Room.** Collect display name and difficulty only, identify Co-op as the sole current mode, validate locally for feedback and server-side for authority, handle active-Room replacement explicitly, and show precise errors.
- [ ] **P08-T07 — Build Join flow.** Normalize/paste Room codes, render safe preview without names, collect display name, handle locked/full/expired/blocked states, and never show deferred approval UI.
- [ ] **P08-T08 — Build Lobby.** Render Room code/share link, participants, host, ready state, current settings, activity, connection state, and accessible start-disabled reasons. Provide only current host controls.
- [ ] **P08-T09 — Build Countdown.** Render server-aligned three-second countdown with a text/live-region equivalent, reduced-motion treatment, host cancellation, and deterministic return to Lobby with readiness reset.
- [ ] **P08-T10 — Implement frontend state boundaries.** Keep authoritative Room state in deterministic reducers, transport effects outside reducers, transient UI state separate, and no server authority in Svelte stores.
- [ ] **P08-T11 — Add responsive and accessibility behavior.** Support keyboard-only creation/join/Lobby, visible focus, 200% zoom, 44×44 primary controls, screen-reader status, no color-only readiness/host markers, and compact layouts without horizontal scroll.
- [ ] **P08-T12 — Add frontend tests.** Cover reducers, safe errors, active-Room replacement, keyboard order, pseudo-localized expansion, dark mode, reduced motion, mobile/desktop layouts, and accessibility helpers.

### Phase gates

- [ ] Public homepage HTML is useful with client JavaScript disabled.
- [ ] Create/Join/Lobby work against real server endpoints in Playwright.
- [ ] No Room preview, metadata, client log, or error exposes participant names or credentials.
- [ ] Keyboard, axe-core, 200% zoom, reduced-motion, and compact viewport checks pass.
- [ ] Pseudo-localized labels do not overlap, truncate critical content, or make controls unreachable.
- [ ] Homepage compressed JavaScript remains below 100 KiB.
- [ ] No third-party analytics, fonts, embeds, or tracking requests are emitted.
- [ ] All cumulative checks pass.

### Commit

```text
feat(web): add room entry and lobby

- add the SSR shell, themes, home, create, join, and lobby
- integrate typed APIs, sessions, English, and pseudo-localization
- cover responsive, keyboard, and accessible lobby states
```

---

## Phase 9 — Accessible Co-op board

### Goal

Deliver responsive, keyboard-complete Co-op gameplay using a semantic board and deterministic authoritative-event reduction.

### Canonical references

- Domain §§15–18, 28–33
- Design §§21–25, 37–39, 44–46, 49–51
- Architecture §§15, 28–33, 37
- Product §§17, 21, 25–27, 38.3, 38.9

### Tasks

- [ ] **P09-T01 — Build the semantic Sudoku grid.** Render 81 focusable semantic cells with CSS Grid, row/column/box context, clue/editable state, accessible names, related-cell relationships, and no canvas.
- [ ] **P09-T02 — Implement the focus model.** Support roving focus, arrow movement, Home/End behavior where documented, direct cell selection, focus restoration after overlays, and no keyboard trap.
- [ ] **P09-T03 — Implement number input.** Support physical digits, number pad, erase, optional input-first/cell-first preference, and pending feedback. Keep Sudoku cells at least 24×24 CSS pixels and primary/number-pad controls at least 44×44.
- [ ] **P09-T04 — Implement notes.** Add explicit note mode, keyboard shortcut, candidate rendering, ownership/attribution, deterministic toggles, and server reconciliation. Do not add Co-op undo.
- [ ] **P09-T05 — Implement the Match reducer.** Reduce snapshots and ordered durable events into authoritative client state, verify MatchVersion/event-number continuity, ignore known-compatible non-state events correctly, and request recovery on gaps or unknown incompatible versions.
- [ ] **P09-T06 — Implement pending reconciliation.** Track RequestID-scoped pending commands, reconcile acknowledgement/event ordering, resolve command-status queries, and never leave a value displayed as accepted before authoritative confirmation.
- [ ] **P09-T07 — Render correctness presets and hints.** Present only information allowed by the active error preset, show shared Challenge penalties separately, expose Nudge/Reveal only when permitted, and label assisted results without leaking the solution.
- [ ] **P09-T08 — Render collaboration.** Show non-color-only player attribution, remote focus/soft locks, override warnings, targeted durable pings, and ephemeral reactions. Keep transient focus/reactions out of durable reducer state.
- [ ] **P09-T09 — Implement gameplay status.** Show server-aligned elapsed time, penalties, connection/controller state, participant list, settings/rules summary, and host leave/cancel actions without displacing the board.
- [ ] **P09-T10 — Implement responsive layouts.** Keep the board dominant on compact, medium, and wide viewports; place number controls within thumb reach; avoid horizontal scrolling; and preserve usable layout at 200% zoom.
- [ ] **P09-T11 — Implement accessible announcements.** Announce accepted/rejected moves, mistakes, hints, remote attribution, connection changes, completion, and destructive consequences without flooding live regions.
- [ ] **P09-T12 — Respect user preferences.** Apply dark/light theme, reduced motion, sound-off default, optional haptics only after user action, and locally stored input/accessibility preferences.
- [ ] **P09-T13 — Add reducer and interaction tests.** Cover every durable event, event gaps, duplicate delivery, pending race order, keyboard navigation, notes, error presets, hints, soft locks, pings/reactions, themes, and screen-reader labels.
- [ ] **P09-T14 — Add multi-context Playwright play.** Create/join/ready/start in two isolated contexts, submit simultaneous commands, verify deterministic winners of conflicts, confirm both clients converge, and exercise mobile plus desktop input.

### Phase gates

- [ ] The complete create/join/ready/start/place/erase/note flow works in two isolated browser contexts.
- [ ] Every gameplay operation is possible with keyboard alone and has a visible focus state.
- [ ] Axe-core, screen-reader-label assertions, non-color cue checks, reduced motion, and 200% zoom checks pass.
- [ ] No canvas, hover-only behavior, hidden authority, disabled-but-deferred control, or multiplayer undo exists.
- [ ] Local cell feedback begins within 50 ms even when the network acknowledgement is delayed.
- [ ] Initial gameplay compressed JavaScript remains below 200 KiB.
- [ ] Clients converge after simultaneous, duplicate, stale, and reordered network delivery scenarios.
- [ ] This closes the internal `0.1.0` capability gate; no tag or deployment is created.
- [ ] All cumulative checks pass.

### Commit

```text
feat(game): add accessible co-op board

- add the semantic board, number pad, keyboard, and notes
- reduce events and render attribution, locks, and social signals
- verify multi-context play, accessibility, and response budgets
```

---

## Phase 10 — Reconnect, snapshots, tabs, and restart recovery

### Goal

Make authoritative Room and Co-op Match state recover correctly across network loss, refresh, browser sleep, duplicate delivery, multiple tabs, and Go process restart.

### Canonical references

- Domain §28, §29, §34.4–34.5, §37.10–37.11
- Architecture §§13–14, 23, 28.3, 41–43, 49.3–49.6, 50
- Design §32
- Product §27

### Tasks

- [ ] **P10-T01 — Persist client checkpoints.** Store only non-secret RoomVersion, MatchID, Match event number, pending RequestIDs, and UI-safe recovery metadata. Never persist the Room cookie, replay capability, Solo proof, or private server payload.
- [ ] **P10-T02 — Implement reconnect state machine.** Use bounded exponential backoff with jitter, stop on explicit leave/terminal authorization errors, distinguish offline/reconnecting/synchronizing/connected/read-only, and avoid polling while WebSocket recovery is active.
- [ ] **P10-T03 — Recover from the event buffer.** Resume from the last contiguous event when the server buffer covers the gap; otherwise request and apply a fresh snapshot followed by ordered events.
- [ ] **P10-T04 — Resolve disconnected input.** Prevent new multiplayer mutations while disconnected. Preserve unsent UI intent only as non-authoritative draft state and never queue it for later automatic submission.
- [ ] **P10-T05 — Resolve uncertain commands.** Query durable receipt status for every timed-out RequestID before allowing a replacement action. Reconcile committed success, rejection, pending, and unknown outcomes without double application.
- [ ] **P10-T06 — Complete controller transfer.** Coordinate tabs with platform APIs, show secondary tabs as read-only, transfer control explicitly or after lease loss, and prevent simultaneous controllers after wake/sleep races.
- [ ] **P10-T07 — Schedule snapshots.** Create snapshots at bounded event/time thresholds and before graceful shutdown without blocking ordinary commands longer than the command budget. Continue to treat events as authoritative.
- [ ] **P10-T08 — Reconstruct on startup.** Verify configuration/migrations, scan nonterminal Matches, load snapshot plus subsequent events, validate invariants/hash continuity, and register recovered actors as `RecoveryPending` before accepting gameplay commands.
- [ ] **P10-T09 — Resume current Co-op safely.** Resume when an eligible player reconnects, cancel when nobody reconnects within five minutes, exclude the full server-caused recovery interval from active elapsed time, and use generation-tagged recovery timers.
- [ ] **P10-T10 — Handle unrecoverable state.** Keep readiness false, preserve the last valid committed state, emit a safe terminal cancellation/audit result where allowed, and never guess missing events or resurrect a completed Match.
- [ ] **P10-T11 — Implement recovery UI.** Add accessible maintenance, reconnecting, synchronizing, recovered, read-only-tab, and unrecoverable-cancellation states with precise retry/consequence messaging.
- [ ] **P10-T12 — Extend TLA+ models.** Cover recovery pending, reconnect eligibility, duplicate replay, stale recovery timers, completed-state protection, and eventual resume/cancel liveness.
- [ ] **P10-T13 — Add recovery integration tests.** Exercise refresh, sleep/wake, network loss, missing buffer, snapshot fallback, uncertain receipt, controller race, host disconnect, server termination/restart, corrupted snapshot, corrupted event sequence, and recovery timeout.

### Phase gates

- [ ] Refresh and reconnect restore the same participant and converge without duplicate effects.
- [ ] A stale tab, stale timer, or stale recovery command cannot mutate current state.
- [ ] Server restart tests use a real file-backed SQLite database and a new Go process.
- [ ] Recovery never skips/reapplies a committed event, accepts a command before validation, or resurrects a completed Match.
- [ ] Co-op elapsed time excludes exactly the server-caused recovery interval.
- [ ] Controller tests cover simultaneous tabs, background throttling, browser sleep, explicit transfer, and lease expiry.
- [ ] TLC and every cumulative test pass.

### Commit

```text
feat(recovery): add reconnect and restart recovery

- add checkpoints, snapshots, gap recovery, and tab transfer
- recover co-op matches after restart with paused elapsed time
- expose reconnecting, synchronized, read-only, and cancel states
```

---

## Phase 11 — Completion, results, rematch, and basic replay

### Goal

Complete the first vertical slice from Room creation through deterministic replay and rematch.

### Canonical references

- Domain §§24, 26–27, §37.12–37.14
- Product §§18, 29.2, 30
- Architecture §§16.6, 24, 27–29
- Design §§30–31, 49.4–49.5

### Tasks

- [ ] **P11-T01 — Finalize Co-op results.** Project terminal reason, server elapsed time, separate penalty time, mistakes, hints/assistance, contribution counts, disconnect summary, and replay availability from committed facts.
- [ ] **P11-T02 — Build Results UI.** Present completion clearly, make Rematch primary, keep Replay discoverable, load replay lazily, show honest pending integrity wording, and never rank Co-op contributors as winners/losers.
- [ ] **P11-T03 — Implement rematch.** Authorize the host, create a new MatchID and assignment/rules, return the Room to Lobby, reset readiness, preserve eligible participants/host/settings as specified, and prevent old Match commands from affecting the rematch.
- [ ] **P11-T04 — Create replay capabilities.** Generate CSPRNG read capabilities, store only hashes, bind them to ReplayID/MatchID/expiry, and return the share URL with capability in the fragment. Deletion still requires the originating Room session.
- [ ] **P11-T05 — Handle fragment capabilities safely.** Copy the capability into memory, immediately remove it from browser history with `history.replaceState`, send it only in the Authorization header, and reconstruct a share URL only for an explicit copy action.
- [ ] **P11-T06 — Implement replay projection.** Read immutable assignment/rules and ordered public events, exclude private payloads/salts, apply deterministic reducers, and create optional seek checkpoints without making them authoritative.
- [ ] **P11-T07 — Build basic replay controls.** Implement play, pause, scrub, 0.5×/1×/2×/4× speed, event markers, shared board, attribution, notes, targeted pings, disconnect/reconnect, and completion.
- [ ] **P11-T08 — Preserve replay privacy.** Keep replay routes noindex, require capability possession, use `Referrer-Policy: no-referrer`, exclude reactions/focus/private data, and show expired/revoked/unauthorized states without confirming unrelated Match existence.
- [ ] **P11-T09 — Add replay reducer tests.** Reconstruct identical state at every event boundary, seek forward/backward, verify speed does not alter order, and reject event gaps/unknown incompatible schemas.
- [ ] **P11-T10 — Complete first-slice E2E.** Cover create, join, ready, start, simultaneous gameplay, completion, results, replay, rematch, refresh, and mobile/keyboard operation across two browser contexts.

### Phase gates

- [ ] Completion finalizes once and the Results Room cannot return the completed Match to Active.
- [ ] Rematch uses a new MatchID and stale prior-Match commands are rejected.
- [ ] Replay reconstructs the accepted event history exactly and never blocks Results rendering.
- [ ] Replay capabilities never appear in server request paths, query strings, logs, metrics, persistent browser storage, or referrers.
- [ ] The canonical first vertical slice passes in two isolated contexts, after refresh, on compact mobile, and keyboard-only.
- [ ] This closes the internal `0.2.0` capability gate; no tag or deployment is created.
- [ ] All cumulative checks pass.

### Commit

```text
feat(results): complete room gameplay loop

- finalize completion, results, rematch, and readiness reset
- add capability-protected deterministic replay controls
- prove the create-to-rematch flow in end-to-end tests
```

---

## Phase 12 — Cryptographically verifiable replay

### Goal

Seal terminal multiplayer event chains and verify them independently in the browser without overstating what hidden-payload commitments prove.

### Canonical references

- Domain §§24–25, §33.5, §37.12, §37.17
- Product §§18–19
- Architecture §§20.3, 20.6, 24–25, 35.7, 49.5
- Design §§30.3, 49.4–49.5, 51

### Tasks

- [ ] **P12-T01 — Freeze proof version 1 vectors.** Commit the exact genesis hash, RFC 8785 input envelopes, byte encodings, SHA-256 outputs, hidden payload commitments, Ed25519 keys/signatures, and expected failures used by Go and TypeScript. Fixture private keys must be unmistakably test-only and rejected by production configuration.
- [ ] **P12-T02 — Validate event-chain creation.** Verify Phase 7 hashing includes proof version, MatchID, event number, aggregate version, public type/actor/payload, occurrence time, private digest, and previous hash exactly as Architecture §25.2 defines.
- [ ] **P12-T03 — Seal terminal Matches.** After the terminal event commits, sign the final event hash with Ed25519, store key ID/proof version/signature separately, and never append replay/audit/deletion facts to the sealed Match stream.
- [ ] **P12-T04 — Protect signing keys.** Load the private key only from validated server configuration, reject malformed/placeholder production keys, never log/export it, and expose only public trusted-key metadata needed by the browser verifier.
- [ ] **P12-T05 — Implement key rotation.** Embed the trusted public-key set in the web build, accept unexpired older replay keys, reject unknown/retired-invalid keys, and document deployment ordering that publishes the new public key before new signatures.
- [ ] **P12-T06 — Implement browser verification.** Verify event sequence, canonical envelopes, each hash, final digest, trusted key ID, and Ed25519 signature. Use Web Crypto when supported and add a fallback only if the supported-browser test matrix proves it necessary.
- [ ] **P12-T07 — Report integrity accurately.** Distinguish verified, verifying, unavailable/legacy, corrupted/tampered, unknown key, and unsupported browser. State that verification proves the sealed replay was not altered, not that the server applied correct gameplay rules.
- [ ] **P12-T08 — Respect hidden commitments.** Show that a signed digest commitment exists without claiming the browser verified undisclosed content. Never return private payloads, salts, or hidden Duel-oriented placeholders.
- [ ] **P12-T09 — Implement early deletion and expiry.** Require the valid originating Room session and confirmation; replacing that one active Room session removes early-delete authority. Revoke all read capabilities and remove replay-accessible projections without rewriting the sealed Match stream. At seven days scrub the remaining participant-linked Match data, retain only the permitted 30-day tombstone, and make repeated deletion/cleanup idempotent.
- [ ] **P12-T10 — Add tamper and rotation tests.** Cover every byte/payload/time/order/hash/signature mutation, missing/duplicate events, hidden-digest changes, unknown keys, valid old keys, changed scope, deleted/expired capability, and cross-language vector equality.
- [ ] **P12-T11 — Measure replay loading.** Lazy-load verifier code, keep Results independent, compress replay HTTP responses, and measure an ordinary replay under the two-second product target.

### Phase gates

- [ ] Go and TypeScript produce identical canonical bytes, hashes, commitments, and verification outcomes for every fixture.
- [ ] A changed public field, event order, previous hash, final digest, signature, or trusted key fails verification.
- [ ] Hidden payload/salt data never reaches browser contracts or public fixtures.
- [ ] Completed seals remain immutable; deletion/expiration/audit state lives outside the signed event stream.
- [ ] Replay deletion and seven-day expiry remove participant-linked replay data while preserving only the allowed 30-day tombstone.
- [ ] Replay verification UI passes keyboard, screen-reader, reduced-motion, mobile, and honest-wording checks.
- [ ] All cumulative checks pass.

### Commit

```text
feat(replay): seal verifiable match replays

- add hash verification, private commitments, and Ed25519 seals
- add key rotation, tamper handling, deletion, and expiration
- share one proof fixture suite across Go and TypeScript
```

---

## Phase 13 — Account-free online Solo

### Goal

Provide server-validated online Solo puzzles while keeping personal progress, history, statistics, and replay on the device.

### Canonical references

- Domain §21, §35.2, §37
- Product §§11.4, 17, 20, 29.1
- Architecture §§16.1, 16.7, 29–30
- Design §28

### Tasks

- [ ] **P13-T01 — Implement Solo entry and assignment API.** Enable the homepage Play Solo action and `/solo` route. Return clues, metadata, AttemptID, and a signed opaque assignment proof for Easy/Medium/Hard/Expert/Random. The proof binds revision, transformation, issue time, and format without containing the solution.
- [ ] **P13-T02 — Validate Solo proof use.** Authenticate hint/completion requests with the proof, enforce request IDs and input limits, reject tampering or mismatched attempt/revision/transformation data, and keep request bodies/proofs out of logs. Do not invent an expiry policy absent from the canonical specification.
- [ ] **P13-T03 — Implement Guided mode.** Show immediate incorrect-value feedback as authorized, support Check Board, notes, pause, Nudge/Reveal, and separate hint penalties.
- [ ] **P13-T04 — Implement Classic mode.** Show direct Sudoku conflicts only until completion/check behavior allowed by the domain, while preserving notes, pause, hints, and authoritative completion validation.
- [ ] **P13-T05 — Implement Solo timing.** Start on first editable interaction, exclude paused and closed-tab intervals, store local elapsed/penalty components, and never use client time to claim server-validated completion ordering.
- [ ] **P13-T06 — Implement IndexedDB storage.** Persist active attempt values, notes, timer, hints, input preference, local history/statistics, recent-puzzle IDs, and local replay. Version and migrate the schema without losing a valid active attempt.
- [ ] **P13-T07 — Implement Continue Last Puzzle.** Restore exactly one current attempt, recover after refresh/browser restart, handle retired/missing puzzle validation safely, and avoid cross-device/cloud claims.
- [ ] **P13-T08 — Implement repetition control.** Prefer puzzles outside bounded device-local recent history, handle a small catalog deterministically, and never create a permanent server player profile.
- [ ] **P13-T09 — Implement local completion/results.** Validate completion through the server, store only the personal result locally, display hints/mistakes/time accurately, and produce a device-local replay.
- [ ] **P13-T10 — Implement local data clearing.** Clear IndexedDB Solo attempts/history/statistics/replays and related preferences after confirmation without implying that shared multiplayer replay is deleted.
- [ ] **P13-T11 — Add Solo tests.** Cover proof tampering, wrong attempt/revision, Guided/Classic visibility, pause/closed-tab time, hints, Check Board, completion, migration, quota/storage failure, resume, repetition, local replay, and clear-data behavior.

### Phase gates

- [ ] No Solo solution or solution-bearing proof appears in frontend state, IndexedDB, logs, errors, or public fixtures.
- [ ] Server stores no personal Solo progress, history, statistics, or replay.
- [ ] Guided and Classic behavior differs exactly as documented and deferred Explain is absent.
- [ ] IndexedDB migration, unavailable-storage, quota, corruption, clear, and restart tests pass.
- [ ] Solo is online-only; there is no service worker or offline claim.
- [ ] Solo keyboard, mobile, 200% zoom, screen-reader, pseudo-localization, and performance checks pass.
- [ ] All cumulative checks pass.

### Commit

```text
feat(solo): add account-free solo play

- add signed assignments, server hints, and completion validation
- keep progress, history, statistics, and replays in IndexedDB
- implement Guided and Classic play without exposing solutions
```

---

## Phase 14 — MVP public experience and accessibility audit

### Goal

Complete current public pages, privacy controls, English content, SEO fundamentals, and a full WCAG 2.2 AA core-flow audit.

### Canonical references

- Product §§20–26, 29.1, 38–43, 47
- Architecture §§27, 32–34, 36, 53
- Design §§15–16, 33–50

### Tasks

- [ ] **P14-T01 — Complete How to Play.** Explain Sudoku basics, notes, Guided/Classic, Co-op, error presets, Nudge/Reveal, reconnect, replay, keyboard controls, and accessibility features. Omit future-mode instructions.
- [ ] **P14-T02 — Complete Privacy.** Accurately describe no accounts/analytics/tracking, essential cookies, local Solo data, temporary multiplayer retention, replay capabilities/deletion, logs, admin access, and clear-data behavior.
- [ ] **P14-T03 — Complete Accessibility.** Publish supported input/assistive behavior, WCAG target, known limitations, contact path, and tested browser/assistive-technology matrix without overstating certification.
- [ ] **P14-T04 — Complete Settings.** Provide theme, motion, sound/haptics where available, input preference, accessibility options, local-data clear, and session-leave behavior. Hide a one-option language selector.
- [ ] **P14-T05 — Finalize English content.** Replace placeholders, use precise verbs, ensure errors have useful next actions, explain destructive consequences, and keep technical claims out of primary gameplay.
- [ ] **P14-T06 — Complete pseudo-localization coverage.** Exercise every route, dialog, toast, live region, replay control, admin-safe error, and metadata template with expansion/diacritics/placeholders. Pseudo-localization is test-only, not user-selectable production content.
- [ ] **P14-T07 — Implement SEO essentials.** SSR meaningful Home/How to Play/Privacy/Accessibility HTML, canonical URLs, English metadata, Open Graph, software-application structured data where valid, `sitemap.xml`, and `robots.txt`.
- [ ] **P14-T08 — Protect private indexing.** Apply noindex and safe metadata to create/join/Room/Match/replay/settings/admin routes. Exclude Room codes, Match IDs, Replay IDs, participant names, and capabilities from sitemap/preview data.
- [ ] **P14-T09 — Audit keyboard and focus.** Complete every current flow without pointer input; verify focus order/restoration/visibility, skip links, dialogs, destructive confirmation, shortcuts, and no traps.
- [ ] **P14-T10 — Audit screen readers and semantics.** Manually verify the supported matrix for board navigation, values/notes/clues, participant attribution, readiness, errors, connection/recovery, results, replay, and Solo.
- [ ] **P14-T11 — Audit visual accessibility.** Verify text/UI contrast, non-color cues, 200% zoom, reflow, 24×24 cells, 44×44 primary controls, reduced motion, dark mode, Windows high-contrast behavior where supported, and no critical truncation.
- [ ] **P14-T12 — Audit public performance.** Measure compressed JavaScript, interactive time, layout shift, Lighthouse performance/accessibility/SEO, caching, image dimensions, font behavior, and absence of third-party requests.

### Phase gates

- [ ] Home, How to Play, Privacy, and Accessibility render useful English HTML without client JavaScript.
- [ ] Private routes are noindex and absent from sitemap; canonical and metadata tests pass.
- [ ] Automated axe-core reports no serious/critical violations in current core flows.
- [ ] Manual keyboard, screen-reader, 200% zoom, dark mode, reduced-motion, and compact/mobile checklists are recorded and pass.
- [ ] Public-page Lighthouse performance reaches 95+ in the documented test environment or a canonical decision records the measured exception and fix owner; the phase is not complete with an unexplained miss.
- [ ] Homepage compressed JavaScript remains under 100 KiB and gameplay under 200 KiB.
- [ ] Privacy-page claims match actual network, cookie, storage, log, retention, and deletion behavior.
- [ ] All cumulative checks pass.

### Commit

```text
feat(web): complete MVP public experience

- add help, privacy, accessibility, and settings pages
- finish local-data controls, English, and pseudo-localization
- audit WCAG, responsive layouts, SEO, and frontend budgets
```

---

## Phase 15 — Administration and production packaging

### Goal

Add private operational controls, safe observability, maintenance, hardened runtime behavior, and deployable Coolify/container artifacts without performing owner-only production operations.

### Canonical references

- Product §§20, 29.3, 32–34
- Architecture §§35–48, 52–54
- Architecture §§57–58

### Tasks

- [ ] **P15-T01 — Implement private administration boundary.** Restrict `/admin` and protected operational endpoints to the configured trusted private network/proxy identity. Do not build an admin account/password system or expose admin navigation publicly.
- [ ] **P15-T02 — Implement MVP admin actions.** Provide Room lookup with minimized fields, terminate broken Room, delete replay, retire puzzle, and display health. Require confirmation, authorization, structured errors, idempotency where applicable, and an append-only audit record.
- [ ] **P15-T03 — Implement health endpoints.** Keep liveness cheap; make readiness require valid config/current migrations/healthy writer/not shutting down; keep detailed status private and free of active Room details.
- [ ] **P15-T04 — Implement safe metrics.** Expose the private Prometheus-compatible metrics in Architecture §38 with bounded labels. Exclude names, participant IDs, Room codes, puzzle values, replay capabilities, and other high-cardinality identifiers.
- [ ] **P15-T05 — Implement structured logging.** Use `log/slog`, correlation IDs, route templates/redaction, safe event/rejection/recovery fields, configured levels, and documented retention. Never log cookies, tokens, proofs, solutions, private notes, or full WebSocket/request payloads.
- [ ] **P15-T06 — Implement maintenance scheduling.** Run bounded, cancellable, observable, idempotent jobs for Room/session/receipt/replay expiry, seven-day participant scrubbing, 30-day tombstone expiry, snapshots, WAL checkpoint, `PRAGMA optimize`, and conditional integrity checks.
- [ ] **P15-T07 — Implement graceful shutdown.** Mark readiness false, reject new joins, notify clients, drain requests, stop commands, snapshot active state, close WebSockets with reconnectable maintenance reason, checkpoint WAL, and exit within the validated 60-second deadline.
- [ ] **P15-T08 — Harden HTTP responses.** Add CSP, HSTS in production, no-referrer private/replay policy, nosniff, Permissions-Policy, anti-framing, restrictive base/form directives, compression, immutable hashed assets, and safe cache headers.
- [ ] **P15-T09 — Harden abuse controls.** Verify CSPRNG Room/replay material, progressive failed-code delay, temporary blocking, trusted proxy parsing, rotating keyed IP digests, request/body limits, WebSocket size limits, and absence of raw IP product storage.
- [ ] **P15-T10 — Build the web image.** Use frozen dependency install, SvelteKit adapter-node build, minimal non-root runtime, read-only filesystem where practical, explicit health behavior, and immutable version metadata.
- [ ] **P15-T11 — Build the server image.** Use a multi-stage Go build, minimal non-root runtime, `/app/data` as the only required writable path, embedded/copied migrations, no compiler/toolchain in runtime, and correct signal handling.
- [ ] **P15-T12 — Add local orchestration.** Provide local compose/reverse-proxy configuration with only the server mounting SQLite, correct `/api`, `/ws`, `/health`, and private `/internal` routing, resource limits, and a local HTTPS production-cookie/Origin test path.
- [ ] **P15-T13 — Add Coolify manifests/runbook.** Define separate web/server applications, persistent volume mount, routing, health checks, environment contract, immutable image tags, migration/startup order, rollback compatibility checks, and owner-supplied secret placeholders.
- [ ] **P15-T14 — Complete CI security/packaging.** Add container builds, dependency vulnerability scanning, generated-file drift, migration testing, contract fixtures, TLA+, Playwright smoke, accessibility smoke, and race-detector jobs with least required permissions.
- [ ] **P15-T15 — Add operational tests.** Cover unauthorized/spoofed admin access, audit records, health transitions, metric-label privacy, log redaction, maintenance reruns/interruption, graceful termination, security headers, rate limits, container read-only behavior, and rollback-compatible startup.

### Phase gates

- [ ] Private admin cannot be reached by an untrusted request or spoofed trusted-proxy header.
- [ ] Every destructive admin action is confirmed, audited, idempotent where required, and privacy-safe.
- [ ] Maintenance tests prove exact seven-day scrub and 30-day tombstone behavior.
- [ ] Logs/metrics/header tests contain no prohibited data or high-cardinality labels.
- [ ] Web and server images run as non-root, pass vulnerability policy, and expose only required writable paths.
- [ ] Local container smoke verifies routing, HTTPS cookies/Origin, WebSocket upgrade, migrations, persistence across restart, readiness, and graceful shutdown.
- [ ] Coolify artifacts are complete, but no DNS, secret, remote deployment, or production mutation is performed.
- [ ] All cumulative checks pass.

### Commit

```text
feat(ops): add secure production operations

- add private admin, health, metrics, maintenance, and safe logs
- add hardened containers, Coolify manifests, and lifecycle handling
- enforce headers, rate limits, retention, and security checks
```

---

## Phase 16 — MVP release qualification

### Goal

Qualify the complete `0.3.0` repository, resolve every discovered defect, finalize release metadata, and create the first release tag without performing owner-only production deployment.

### Canonical references

- Product §§26–34, 38, 47
- Architecture §§37, 47–50, 57–59
- Design §§49–53
- `AGENTS.md` Workflow and Definition of Done

### Tasks

- [ ] **P16-T01 — Freeze the release candidate.** Confirm only current MVP features are enabled, migrations/contracts/proof versions are final for `0.3.0`, dependencies and actions are pinned, and no known critical/high defect remains open.
- [ ] **P16-T02 — Run backend qualification.** Run Go formatting, vet, unit, property/fuzz corpus, real-SQLite integration, WebSocket integration, race detector, solver/catalog validation, command/recovery/replay/admin/maintenance tests, and coverage review.
- [ ] **P16-T03 — Run frontend qualification.** Run formatting, lint, strict typecheck, unit/reducer/storage/replay tests, production build, bundle analysis, localization/pseudo-localization checks, and supported-browser matrix.
- [ ] **P16-T04 — Run contract and migration qualification.** Regenerate everything and require a clean diff; validate all shared fixtures; migrate an empty database and every representative prior schema; verify foreign keys, WAL, downgrade/rollback policy, and startup compatibility.
- [ ] **P16-T05 — Run formal verification.** Execute TLC for Room and Match configurations with committed bounds, record state counts/runtime, and verify every current Domain §34 invariant and liveness property.
- [ ] **P16-T06 — Run full E2E matrix.** Execute every Architecture §49.6 scenario on desktop/mobile projects plus locked/full/blocked Room, host transfer, countdown cancel, Challenge penalty, replay expiry/deletion, storage failure, and current public pages.
- [ ] **P16-T07 — Run accessibility qualification.** Complete automated axe and manual keyboard/screen-reader/zoom/reduced-motion/contrast/touch-target checks for every current route and state; record supported browser/assistive versions.
- [ ] **P16-T08 — Run load and resilience qualification.** Exercise 100 WebSockets, 25 Rooms, creation/countdown bursts, simultaneous commands/completions, reconnect storm, slow readers/spectators, SQLite pressure, maintenance overlap, graceful shutdown, and restart recovery.
- [ ] **P16-T09 — Verify performance budgets.** Measure homepage/gameplay bundle sizes, public Lighthouse, cell feedback, command processing, acknowledgement latency in the documented same-region setup, message sizes, replay load, process resources, and layout shift.
- [ ] **P16-T10 — Run security/privacy qualification.** Scan dependencies/images/secrets, test Room-code enumeration controls, cookie/Origin/CSRF assumptions, WebSocket payload limits, capability handling, headers, admin boundary, log/metric redaction, retention, replay deletion, and absence of third-party tracking.
- [ ] **P16-T11 — Verify containers and deployment artifacts.** Build immutable SHA/version images, run local HTTPS container smoke from empty and existing databases, test rollback-compatible previous image pair where available, and validate Coolify configuration without remote deployment.
- [ ] **P16-T12 — Finalize documentation and evidence.** Update README status, operations/recovery/rollback instructions, supported environment, privacy/accessibility statements, known limitations, release notes, and canonical docs for any behavior changed during qualification. Add `docs/releases/0.3.0.md` with commands, tool versions, CI links, measured budgets, manual accessibility matrix, security/privacy results, and accepted non-blocking limitations; do not commit raw logs or secrets.
- [ ] **P16-T13 — Resolve every failed gate.** Fix defects inside this phase, add regression tests, rerun the affected suite and full cumulative checks, and document any accepted non-critical limitation in the appropriate canonical document. Do not waive an invariant, privacy, security, or accessibility failure.
- [ ] **P16-T14 — Set release metadata.** Set the version to `0.3.0`, ensure build/status endpoints and images report the immutable commit/version, generate release checksums, and verify no dirty generated output.
- [ ] **P16-T15 — Prepare release tagging.** Confirm `v0.3.0` does not already exist, prepare an annotation summarizing the qualified MVP and release-report path, and verify the candidate is ready to tag after the phase commit. Do not push, publish images, configure secrets, or deploy remotely without a separate explicit request.

### Phase gates

- [ ] Every required GitHub Actions check passes on the release commit.
- [ ] `make test`, `make lint`, `make generate`, `make migrate`, `make puzzles`, `make e2e`, `make build`, and `make tla` pass from a clean checkout.
- [ ] The complete functional, concurrency, recovery, replay-integrity, accessibility, localization, security, privacy, load, and performance evidence is recorded.
- [ ] There is no unexplained performance-budget miss or unresolved critical/high vulnerability.
- [ ] Fresh-install, upgrade, restart, rollback-compatibility, and retention tests pass with real file-backed SQLite.
- [ ] No deferred feature is exposed or advertised as current.
- [ ] Release notes and public privacy/accessibility claims match actual behavior.
- [ ] The staged release commit contains the completed phase ledger, release report, final metadata, and no unrelated changes.

### Commit

```text
chore(release): qualify 0.3.0 MVP

- run the full functional, concurrency, accessibility, and load suites
- verify migrations, containers, HTTPS, recovery, and budgets
- finalize release artifacts and v0.3.0 version metadata
```

### Post-commit release action

After the exact Phase 16 commit succeeds and the working tree is clean, create the prepared annotated `v0.3.0` tag pointing to that commit. Tag creation does not modify the working tree and is not a second phase commit. Pushing the commit/tag, publishing images, configuring secrets, and deploying remain separately authorized operations.

---

## 4. Capability checkpoints

These checkpoints measure progress but do not create tags or deployments:

| Checkpoint | Owning phase | Required capability |
|---|---:|---|
| `0.1.0` local Co-op capability | 9 | Two isolated browser players can create/join/ready/start and perform basic authoritative Co-op play locally. |
| `0.2.0` persistent multiplayer alpha | 11 | SQLite persistence, recovery, completion, results, basic replay, and rematch complete the first vertical slice. |
| `0.3.0` full MVP / portfolio beta | 16 | Every current-scope feature and release gate passes; one annotated `v0.3.0` tag is created. |

The earlier checkpoints may exceed their minimum Product §31 milestone internals because higher-authority architecture requires SQLite durability before broadcast. Do not add a disposable in-memory production path merely to reproduce the minimum prototype wording.

## 5. MVP coverage matrix

Every Product §29.1 capability has one primary owning phase. Cross-cutting verification may recur later.

| MVP requirement | Primary phase |
|---|---:|
| Public homepage | 8 |
| Create Room | 5 |
| Join by code/link | 5 |
| Temporary display names | 2 |
| Co-op, 1–6 players | 6 |
| One active Room per browser profile | 5 |
| Lobby readiness | 5 |
| Shared board and notes | 9 |
| Soft locks and player attribution | 9 |
| Durable targeted pings and ephemeral reactions | 6 |
| Error presets | 6 |
| Nudge and Reveal | 6 |
| Host settings, lock, remove/block, transfer, start/cancel, rematch | 5 and 11 |
| Reconnect and server-restart recovery | 10 |
| Completion and result | 11 |
| Basic replay | 11 |
| Replay integrity verification | 12 |
| TLA+ current models | 5, 6, and 10 |
| Rematch | 11 |
| Online Solo | 13 |
| Light/dark mode | 8 |
| Responsive mobile/desktop | 8 and 9 |
| English and pseudo-localized expansion testing | 8 and 14 |
| How to Play, Privacy, and Accessibility pages | 14 |
| WCAG 2.2 AA core flow | continuous; final owner 14 |
| Private MVP administration | 15 |
| Coolify-ready deployment artifacts | 15 |

Architecture §59 current implementation steps map as follows:

| Architecture step | Primary phase |
|---|---:|
| Monorepo and tooling | 1 |
| Domain primitives | 2 |
| Puzzle validator and solver | 3 |
| SQLite migrations and repositories | 4 |
| Room aggregate | 5 |
| Co-op Match aggregate | 6 |
| HTTP Room creation/join | 5 |
| WebSocket protocol | 7 |
| Room actor registry | 5 |
| Svelte Lobby | 8 |
| Semantic Sudoku board | 9 |
| Co-op event reducer | 9 |
| Persistence-before-broadcast | 7 |
| Reconnect and snapshot recovery | 10 |
| Match completion | 6 and 11 |
| Replay reconstruction | 11 |
| Replay hash chain | 7 and 12 |
| Ed25519 sealing | 12 |
| TLA+ model | 5, 6, and 10 |
| Solo | 13 |
| MVP administration and public hardening | 14 and 15 |

## 6. Explicit exclusions

The following terms may appear in this file only to prevent accidental scope expansion:

- Race
- Duel
- Daily Ninefold
- offline Solo
- Explain hints
- host-approval joining
- multiplayer undo
- optional overall Match deadline
- full spectator UX
- German, Albanian, and Turkish production catalogs
- PWA/service worker installation
- About and dedicated mode pages
- public rankings or matchmaking
- accounts or cross-device sync
- rich reporting
- external behavioral analytics/error tracking
- off-VPS backup

Future work is defined only in [ROADMAP.md](ROADMAP.md) and remains provisional until the required canonical review moves it into current scope.
