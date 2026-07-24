# Ninefold Sudoku roadmap

**Planning horizon:** `0.1.0` through `1.0.0`

**Current implementation target:** `0.3.0` full MVP / portfolio beta

**Scheduling model:** capability and exit gates, not calendar dates

## 1. Authority and roadmap rules

This roadmap describes release intent. It does not move a provisional feature into current scope and does not override:

1. [Domain specification](docs/DOMAIN.md)
2. [Architecture specification](docs/ARCHITECTURE.md)
3. [Product specification](docs/PRODUCT.md)
4. [Design specification](docs/DESIGN.md)

[TASKS.md](TASKS.md) is the executable implementation ledger for `0.3.0` only.

Roadmap rules:

- Versions are completed in order.
- A later version does not begin while the preceding version has an unresolved release-blocking defect.
- Race, Duel, Daily Ninefold, and other deferred features remain provisional until the required focused review updates the canonical documents.
- A focused review must resolve every recorded ambiguity before implementation contracts, database migrations, or UI scaffolding are added.
- Every release preserves privacy, server authority, persistence-before-broadcast, deterministic replay, accessibility, localization readiness, and performance budgets.
- Deferred features must not be hidden behind unused production code or speculative database tables.
- Dates are intentionally absent until team capacity and delivery velocity are known.
- Player accounts, advertising, behavioral analytics, public matchmaking, rankings, and other Product §35 non-goals are not implied by any roadmap version.

## 2. Version sequence

| Version | Status | Outcome |
|---|---|---|
| `0.1.0` | Internal capability gate | Local two-browser Co-op play |
| `0.2.0` | Internal capability gate | Persistent and recoverable Co-op vertical slice |
| `0.3.0` | Current target | Full MVP / portfolio beta |
| `0.4.0` | Planned; provisional rules | Race mode |
| `0.5.0` | Planned; provisional rules | Duel mode |
| `0.6.0` | Planned; provisional rules | Daily Ninefold |
| `1.0.0` | Planned public-ready release | Stable complete V1 and broader-promotion gates |

`0.1.0` and `0.2.0` are progress checkpoints inside `TASKS.md`; they are not tagged or deployed. The first planned tag is `v0.3.0`.

---

## 3. `0.1.0` — Local Co-op capability gate

### Purpose

Prove the core interaction model and authoritative two-browser gameplay locally without treating prototype-only shortcuts as production architecture.

### Included capabilities

- reproducible Go/SvelteKit monorepo and quality commands;
- transport-independent domain primitives and generated contracts;
- verified puzzle validation/solver/catalog path;
- SQLite durability required by higher-authority architecture;
- private Room creation and joining;
- temporary display names;
- Lobby readiness and countdown;
- two isolated browser participants;
- semantic Sudoku board;
- authoritative value placement, erase, and notes;
- basic attribution, soft locks, pings, and reactions;
- keyboard and mobile operation;
- automated domain, persistence, transport, reducer, and basic multi-context tests.

### Exit gates

- Two isolated browser contexts create, join, ready, start, and submit Co-op moves.
- Simultaneous commands resolve deterministically and both clients converge.
- Durable commands commit before acknowledgement or broadcast.
- The board works with keyboard-only input and a compact mobile viewport.
- No standalone multiplayer solution reaches a client.
- All checks through TASKS Phase 9 pass.

### Explicit exclusions

- results;
- replay;
- rematch;
- restart recovery qualification;
- Solo;
- production administration or deployment.

### Release handling

Internal checkpoint only. Do not create a tag, image release, or deployment.

---

## 4. `0.2.0` — Persistent multiplayer alpha gate

### Purpose

Complete the durable Co-op vertical slice and prove that refresh, reconnect, restart, completion, replay, and rematch preserve authoritative state.

### Included capabilities

- all `0.1.0` capabilities;
- durable Room and Match projections, receipts, events, and snapshots;
- refresh and temporary-network reconnect;
- multiple-tab controller lease;
- event-gap and snapshot recovery;
- Go process restart recovery with `RecoveryPending`;
- server-caused pause excluded from Co-op elapsed time;
- automatic completion and Co-op result;
- capability-protected deterministic basic replay;
- rematch returning the Room to Lobby with readiness reset;
- complete two-context create-to-rematch end-to-end coverage.

### Exit gates

- Duplicate, stale, reordered, and uncertain commands cannot create duplicate effects.
- Refresh/reconnect restores the same participant and converges.
- A new Go process reconstructs an active Match from file-backed SQLite.
- Recovery cannot resurrect a completed Match or accept commands before validation.
- Completion finalizes once.
- Replay reconstructs the accepted event history.
- Rematch receives a new MatchID and rejects stale prior-Match commands.
- All checks through TASKS Phase 11 pass.

### Explicit exclusions

- cryptographic replay verification UI;
- online Solo;
- complete public content/audit;
- production administration or packaging;
- every post-MVP mode.

### Release handling

Internal checkpoint only. Do not create a tag, image release, or deployment.

---

## 5. `0.3.0` — Full MVP / portfolio beta

### Purpose

Deliver the complete current-scope product: production-grade Co-op, account-free online Solo, verifiable replay, accessibility, privacy, public pages, administration, and Coolify-ready packaging.

### Included capabilities

- all `0.2.0` capabilities;
- RFC 8785/SHA-256 replay chain and Ed25519 terminal seal;
- salted digest commitments for hidden server payloads;
- browser replay verification and honest integrity wording;
- replay deletion, seven-day expiry/scrub, and 30-day non-identifying tombstone;
- online Solo with signed assignment proof;
- Guided and Classic Solo behavior;
- device-local Solo progress, history, statistics, and replay;
- Home, How to Play, Privacy, Accessibility, and Settings;
- English production UI and pseudo-localized expansion testing;
- WCAG 2.2 AA core-flow target;
- private MVP administration;
- safe health, metrics, logging, maintenance, and graceful shutdown;
- hardened web/server containers and Coolify-ready manifests;
- complete automated and manual release qualification described by TASKS Phase 16.

### Exit gates

- Every checkbox and gate in [TASKS.md](TASKS.md) passes.
- All current Domain invariants and TLA+ properties hold.
- Contract generation and migrations reproduce from a clean checkout.
- Unit, property, integration, race-detector, WebSocket, frontend, Playwright, accessibility, load, resilience, security, privacy, and container suites pass.
- Public Lighthouse and documented browser/server performance budgets pass or have an approved canonical resolution before release.
- Privacy and accessibility pages match actual behavior.
- No current UI advertises a deferred feature as available.
- The release commit is clean and receives the annotated `v0.3.0` tag.

### Explicit exclusions

- Race;
- Duel;
- Daily Ninefold;
- offline Solo/PWA;
- Explain hints;
- host approval;
- multiplayer undo;
- full spectator UX;
- German, Albanian, and Turkish production catalogs;
- About and dedicated mode pages;
- off-VPS backup;
- broader public promotion.

### Operational boundary

The repository provides deployment artifacts and local automated verification. DNS, Coolify project mutation, production secrets, image publication, tag pushing, and remote deployment require a separate owner-authorized operation.

---

## 6. `0.4.0` — Race

**Status:** Planned, but Domain §19 and Design §26 remain provisional until the focused Race review completes.

### Purpose

Add private, fair multiplayer races on the same puzzle without exposing opponent values or private notes.

### Required scope review

Before code or migrations:

- confirm player/spectator capacities and whether MVP-limited spectator behavior changes;
- ratify that every seated Race player must ready or leave/switch role;
- validate the two-second progress/heatmap delay against copy resistance and accessibility;
- ratify correct-cell progress calculation and all ranking tie-breakers;
- ratify the 60-second finishing window and behavior when all remaining players finish early;
- confirm Casual hint policy, 20-second assistance penalty, and competitive hint prohibition;
- define disconnect/reconnect behavior at finishing boundaries;
- define public/private event projections and replay visibility timing;
- update Domain commands/events/errors/acceptance scenarios;
- update Architecture contracts, persistence projections, recovery, load model, and TLA+ scope;
- update Product scope/version and Design Race/live/replay/results behavior.

No Race implementation begins until the canonical documents remove its provisional status.

### Included capabilities

- Room creation/settings for Race;
- 2–8 independent player boards using one immutable puzzle assignment;
- private values and notes;
- server-authoritative progress based only on correct non-clue cells;
- delayed/batched, digit-free opponent heatmap and text equivalent;
- first verified finisher based on accepted server timestamp;
- server-authoritative finishing window;
- deterministic ranking and allowed tied placement;
- supported Race error presets and ratified hint behavior;
- late join as spectator only where the ratified spectator scope permits;
- Race disconnect/reconnect/recovery;
- Race result, replay, and assistance disclosure;
- Race-specific contracts, persistence, reducers, metrics, tests, and TLA+ properties.

### Quality and exit gates

- Opponent values, wrong digits, and private notes never cross visibility boundaries.
- Progress/heatmap has accessible text and non-color equivalents.
- First finisher is declared at most once and the finishing window cannot reopen.
- Server timestamps, not clients, determine finish order.
- Ranking property tests cover every tie-break and exact tie.
- Multi-context E2E covers start, finish, disconnect, hint policy, expiry, replay, and rematch.
- Load tests cover simultaneous finishes and delayed progress fan-out within current resource budgets.
- Recovery, replay integrity, keyboard, mobile, screen-reader, and performance gates pass.

### Explicit exclusions

- Duel;
- Daily Ninefold;
- public matchmaking/rankings/leaderboards;
- exact live opponent digits;
- permanent player identity or statistics.

---

## 7. `0.5.0` — Duel

**Status:** Planned, but Domain §20 and Design §27 are deliberately incomplete implementation contracts.

### Purpose

Add a two-player turn-based Sudoku mode with private thinking, authoritative timers, deterministic scoring, and meaningful strategy distinct from Race.

### Required scope review

Resolve and document all of these before implementation:

- whether accepted correct values are immutable;
- how erase/re-placement avoids score farming;
- whether voluntary pass exists and what it costs;
- exact protected-disconnect and subsequent-disconnect fairness;
- whether note activity appears in replay and, if so, which viewers receive which information;
- interaction between erase, structure bonuses, and already-completed row/column/box state;
- timer behavior for Unlimited turns, server restart, browser sleep, and maintenance;
- spectator delay/capacity and post-Match reaction behavior;
- result amendment/invalidation boundaries without rewriting sealed events.

Then update Domain commands/events/errors/acceptance scenarios, MatchRules, Architecture projections/contracts/recovery/TLA+, Product scope, and Design live/replay/results flows.

No Duel implementation begins while any listed decision remains unresolved.

### Included capabilities

- exactly two seated players;
- one shared authoritative value board;
- private per-player notes;
- server-authoritative alternating turn ownership;
- ratified 15/30/60-second and Unlimited timer behavior;
- correct, incorrect, erase/pass-if-approved, timeout, resignation, and forfeit actions;
- deterministic base/row/column/box/completion scoring without repeat farming;
- private rejected digit and thinking-state protection;
- one ratified protected disconnect pause per player;
- deterministic tie-breaks and shared victory only at the final tie boundary;
- Duel result, replay, rematch starting-player alternation, and spectator projection;
- Duel-specific persistence, reducers, contracts, tests, metrics, recovery, and TLA+ properties.

### Quality and exit gates

- Only the active player can perform turn-consuming actions.
- A correct placement never grants an unintended extra turn.
- Structure bonuses cannot be scored more than once for the same completed structure state.
- Wrong attempted digits and private notes remain hidden from unauthorized clients/replay projections.
- Stale timer generations cannot end a newer turn.
- Disconnect pause can be consumed at most once per player and recovery follows ratified fairness rules.
- Score/ranking property tests cover every compound bonus and tie-break.
- E2E covers every terminal reason, rematch alternation, restart, replay, accessibility, and mobile layout.
- Duel TLA+ safety/liveness, load, privacy, and performance gates pass.

### Explicit exclusions

- Daily Ninefold;
- Elo/ranked ladders, seasons, tournaments, or public matchmaking;
- public opponent notes or rejected digits;
- permanent profiles/statistics.

---

## 8. `0.6.0` — Daily Ninefold

**Status:** Planned, but Domain §22 and Design §29 remain provisional until focused review.

### Purpose

Add one shared daily puzzle that encourages return play without accounts, public leaderboards, or server-side personal tracking.

### Required scope review

Before implementation:

- ratify the canonical UTC day boundary and assignment identifier;
- confirm difficulty rotation and catalog eligibility;
- define behavior when an assigned puzzle is retired, invalidated, or unavailable;
- define local streak semantics across missed days, timezone presentation, cleared data, and multiple devices;
- define assignment-proof reuse and offline/non-offline boundary;
- confirm shareable result content and replay privacy;
- update Product scope/version, Domain aggregate/commands/events/acceptance scenarios, Architecture storage/API/cache/maintenance, and Design homepage/Daily/results states.

No Daily implementation begins until the canonical documents remove its provisional status.

### Included capabilities

- one deterministic puzzle assignment per UTC day;
- ratified difficulty rotation;
- homepage Daily entry with current status and next-day timing;
- account-free play using signed server assignment;
- device-local progress, completion, streak, recent history, and replay;
- replacement behavior for invalid/retired assignments;
- completion result and privacy-safe share summary;
- Daily-specific API/cache/maintenance behavior, IndexedDB migration, tests, and accessibility.

### Quality and exit gates

- Every user receives the same valid assignment for the canonical UTC day.
- Day rollover and countdown use server truth while presenting local-friendly dates.
- No public leaderboard, account, permanent identifier, or server personal streak is created.
- Streak behavior is deterministic for missed days, clock/timezone changes, cleared storage, and assignment replacement.
- Invalid/retired puzzle recovery does not corrupt completion or streak state.
- E2E covers rollover, resume, completion, replacement, replay, clear data, mobile, keyboard, and screen reader.
- Privacy, performance, migration, and catalog-quality gates pass.

### Explicit exclusions

- public leaderboards or rankings;
- cross-device streak sync;
- accounts or notifications;
- offline/PWA guarantees unless separately moved into scope;
- additional Sudoku variants.

---

## 9. `1.0.0` — Public-ready V1

**Status:** Planned. Scope is assembled from Product §§12.2, 31.4, 32–33 and remains subject to focused reviews for each deferred capability.

### Purpose

Stabilize the complete V1 mode set and satisfy every privacy, accessibility, localization, recovery, backup, security, performance, SEO, and operational gate required for broader public promotion.

### Included capabilities

#### Mode and gameplay maturity

- stable Co-op, Race, Duel, Solo, and Daily Ninefold;
- focused defect, balance, recovery, and replay review for every mode;
- full spectator UX under ratified mode-specific visibility/delay rules;
- richer replay after a focused privacy/performance review;
- Explain hints after a focused educational-content/domain review;
- no unresolved provisional gameplay rule.

#### Offline and installation

- PWA installation only after service-worker cache/update behavior is specified;
- offline Solo only with explicit online/offline classification, puzzle-proof validity, conflict-free local migration, and honest availability messaging;
- multiplayer remains online-only and never queues offline moves.

#### Localization and content

- German (`de`), Albanian (`sq`), and Turkish (`tr`) production catalogs;
- human linguistic review of gameplay, errors, privacy, accessibility, metadata, and pluralization;
- locale routing/selection and `hreflang` only for shipped locales;
- About page and dedicated current-mode SEO/help pages;
- no machine-translation-only release.

#### Reliability and operations

- automated encrypted off-VPS SQLite backup/replication;
- documented deletion behavior across retained backup generations;
- successful restore drill into an isolated environment;
- reliable private monitoring and alerting;
- verified retention, replay deletion, maintenance, and recovery jobs;
- stable deployment/rollback procedures and immutable artifact retention.

#### Public readiness

- final privacy/security threat review;
- final WCAG 2.2 AA audit and published known limitations;
- performance budgets met or formally revised with evidence;
- Room-code enumeration and abuse-control qualification;
- WebSocket payload/fan-out qualification for all modes;
- protected administration;
- complete SEO metadata, sitemap, canonicals, structured data, and localized public content;
- public documentation and source repository free of secrets/personal data.

### Required pre-implementation reviews

Separate canonical reviews are required for:

- full spectator visibility and capacity;
- richer replay content and retention;
- Explain hint pedagogy and scoring/assistance effects;
- service-worker/PWA/offline Solo architecture;
- each production locale and locale-routing strategy;
- backup provider, encryption, retention, restore, and deletion semantics;
- public monitoring/alert thresholds and incident response;
- broader-promotion threat model and abuse limits.

### Exit gates

All Product §32 criteria must pass:

- every current mode is stable;
- replay verification and recovery are reliable;
- accessibility audit passes;
- all four languages receive human review;
- SEO pages are complete;
- automated off-VPS backup exists and restore is tested;
- security, enumeration, WebSocket payload, and admin-protection tests pass;
- privacy content is accurate;
- performance budgets are met or have an approved evidence-based revision.

Product §33 promotion gates must also pass:

- production monitoring is reliable;
- retention jobs and replay deletion are verified;
- backup and restore procedures are operational;
- the complete security gate is closed.

### Explicit continuing non-goals

Unless a later explicit product decision changes scope, `1.0.0` still excludes:

- player accounts and OAuth;
- public matchmaking, rankings, ladders, seasons, and tournaments;
- advertising, subscriptions, or paid gameplay advantage;
- public chat, voice, permanent teams, and social graphs;
- native mobile/desktop applications;
- public third-party API;
- multiple Sudoku variants;
- client-side SQLite;
- multi-region, Kubernetes, microservices, or horizontal Go authority;
- behavioral analytics and third-party tracking.

---

## 10. Change control

A roadmap change must update the canonical document that owns the decision:

- scope, version contents, user outcome, or non-goal → Product
- gameplay rule, state, command, event, or invariant → Domain
- protocol, storage, security, deployment, or testing constraint → Architecture
- interaction, responsive, accessibility, or visual behavior → Design

After canonical approval:

1. update this roadmap;
2. add or revise the executable task plan only when the version becomes current;
3. update contract/schema/TLA+ plans with the feature;
4. keep prior sealed events and released replay formats compatible;
5. record an ADR only for an expensive-to-reverse architecture change.
