# Ninefold Sudoku — Product Specification

**Document:** `docs/PRODUCT.md`  
**Status:** Canonical product requirements and scope  
**Current implementation scope:** Full MVP (`0.3.0`)
**Product:** Ninefold Sudoku  
**Compact brand:** Ninefold  
**Tagline:** *Solve together. Race the grid.*  
**Public URL:** `https://ninefold.recica.dev`  
**Repository:** `ninefold-sudoku`  
**Default branch:** `master`

---

## 1. Purpose and authority

This document defines what Ninefold Sudoku is, who it serves, which product outcomes matter, what must be built, what belongs in each release phase, and what is intentionally excluded.

This document is the canonical source of truth for:

- product vision;
- target users;
- positioning;
- feature scope;
- release sequencing;
- mode priorities;
- product principles;
- privacy commitments;
- accessibility expectations;
- SEO expectations;
- supported languages;
- MVP boundaries;
- launch criteria;
- non-goals;
- product acceptance criteria.

Gameplay rules and invariants are defined in `docs/DOMAIN.md`.

Implementation architecture, protocols, persistence, deployment, testing, and operational constraints are defined in `docs/ARCHITECTURE.md`.

Visual design, interaction patterns, responsive behavior, and component-level experience are defined in `docs/DESIGN.md`.

When documents conflict:

1. `DOMAIN.md` controls gameplay behavior and invariants.
2. `ARCHITECTURE.md` controls implementation constraints.
3. `PRODUCT.md` controls scope, positioning, and product outcomes.
4. `DESIGN.md` controls visual and interaction expression.

Current-scope rule:

- only features listed in section 29.1 are required for the `0.3.0` MVP;
- sections for Race, Duel, Daily Ninefold, offline Solo, Explain hints, full spectator UX, additional locales, PWA installation, and dedicated mode pages describe future direction and are provisional;
- a provisional feature must receive a focused product and domain review before implementation.

---

## 2. Product summary

Ninefold Sudoku is a privacy-first, multiplayer-first Sudoku web application built for private play among family and friends and designed to function as a high-quality public portfolio showcase.

The complete product direction allows people to:

- create a private Sudoku room;
- invite others with a short code or link;
- solve one board together in Co-op mode;
- race on identical puzzles in Race mode;
- compete turn by turn in Duel mode;
- play alone in Solo mode;
- return for one shared Daily Ninefold puzzle;
- view deterministic, cryptographically verifiable match replays.

The current MVP exposes Co-op and Solo. Race, Duel, and Daily Ninefold are not advertised as available gameplay.

Ninefold does not require player accounts.

The core promise is:

> Start playing with other people quickly, privately, and without an account.

The primary play flow is:

```text
Open Ninefold
→ Create or join a private room
→ Enter a temporary display name
→ Ready up
→ Play
→ View results and replay
→ Rematch
```

---

## 3. Product identity

### 3.1 Name

Full product name:

```text
Ninefold Sudoku
```

Compact brand:

```text
Ninefold
```

### 3.2 Tagline

```text
Solve together. Race the grid.
```

### 3.3 Browser title

```text
Ninefold Sudoku — Multiplayer Sudoku
```

### 3.4 Public domain

```text
https://ninefold.recica.dev
```

### 3.5 Repository

```text
ninefold-sudoku
```

### 3.6 Brand meaning

“Ninefold” references:

- the 9×9 Sudoku grid;
- the repeated 3×3 structure;
- multiple people solving one logical system;
- a compact, memorable identity that is broader than a generic “online Sudoku” label.

The word “Sudoku” must remain visible in:

- product title;
- search metadata;
- social previews;
- portfolio description;
- first-screen product context.

---

## 4. Product vision

Ninefold should become the best small-scale private multiplayer Sudoku experience for people who want to play together without creating accounts, surrendering personal data, or navigating a noisy gaming platform.

The product must feel:

- immediate;
- elegant;
- technically dependable;
- respectful of user privacy;
- accessible;
- fast;
- calm;
- socially engaging;
- polished enough to represent the creator publicly.

The project is both:

1. a real game for personal and public use;
2. a portfolio case study in real-time systems, event-driven replay, formal verification, privacy-conscious design, accessibility, and efficient single-VPS architecture.

Neither purpose may compromise the other.

The portfolio dimension must emerge from excellent engineering and product quality, not from visible technical gimmicks that distract from play.

---

## 5. Product mission

Ninefold exists to make Sudoku more social without making it more complicated.

It should help people:

- solve together;
- compare approaches;
- enjoy friendly competition;
- learn from replay;
- return to a familiar private room experience;
- play on mobile or desktop with equal confidence;
- trust that the game is private and fair.

---

## 6. Product positioning

### 6.1 Positioning statement

For friends, families, and small groups who enjoy Sudoku, Ninefold is a private multiplayer Sudoku game that enables cooperative solving, racing, and turn-based duels directly in the browser without accounts or public matchmaking.

Unlike generic Sudoku apps or broad multiplayer gaming platforms, Ninefold emphasizes:

- private rooms;
- fast entry;
- shared solving;
- deterministic replays;
- strong accessibility;
- minimal data collection;
- refined design;
- high technical integrity.

### 6.2 Primary differentiation

The strongest differentiators are:

1. **Co-op-first multiplayer Sudoku**
2. **No-account private rooms**
3. **Deterministic match replay**
4. **Cryptographically verifiable replay integrity**
5. **Formally checked concurrency-sensitive behavior**
6. **Privacy-first architecture**
7. **Accessible mobile and desktop interaction**
8. **Minimal infrastructure and resource use**

### 6.3 What Ninefold is not

Ninefold is not:

- a social network;
- a ranked esports platform;
- a public matchmaking service;
- a heavily monetized mobile game;
- a puzzle marketplace;
- an ad-supported casual game;
- a chat platform;
- a blockchain project;
- a generic game-engine showcase.

---

## 7. Target users

### 7.1 Primary users

#### Friends and family

People who know each other and want to:

- solve a puzzle together;
- race casually;
- challenge each other;
- play without registration;
- share a room link through messaging apps.

#### The creator and close circle

The product must work extremely well for:

- recurring family play;
- testing with friends;
- informal game nights;
- small private groups.

### 7.2 Secondary users

#### Portfolio visitors

Recruiters, clients, developers, designers, and technical peers who may:

- try the game;
- inspect the public source code;
- read the architecture case study;
- evaluate engineering quality;
- evaluate product thinking;
- evaluate accessibility and performance.

#### Public Sudoku players

People discovering Ninefold through search or shared links who want:

- private multiplayer Sudoku;
- a refined Solo experience;
- a trustworthy account-free product.

The future product direction also serves players who want a shared Daily Ninefold puzzle.

### 7.3 User characteristics

The current MVP must support users who:

- are new to Sudoku;
- are casual players;
- are advanced solvers;
- primarily use a mobile browser;
- primarily use a desktop keyboard;
- use assistive technology;
- have slower networks;
- do not want to create an account;
- use the English UI.

German, Albanian, and Turkish support is required before `1.0.0`, not for the current MVP.

---

## 8. User needs

### 8.1 Multiplayer needs

Users need to:

- create a room quickly;
- understand what mode they are entering;
- invite others easily;
- join using a short code or link;
- see who is present;
- know who is ready;
- understand error, hint, and timer rules before play;
- recover from refresh or connection loss;
- see clear multiplayer state;
- rematch without recreating the room.

### 8.2 Gameplay needs

Users need:

- a readable Sudoku board;
- fast number entry;
- notes;
- clear conflicts;
- configurable error behavior;
- trustworthy timers;
- mode-specific feedback;
- accessible controls;
- a satisfying completion state.

### 8.3 Privacy needs

Users need confidence that:

- no account is required;
- no advertising profile is built;
- no behavioral tracking is present;
- no personal email is required;
- temporary names are not permanent identities;
- private rooms are not publicly listed;
- replays expire automatically;
- local data can be cleared.

### 8.4 Reliability needs

Users need:

- accepted moves not to disappear;
- duplicate commands not to create duplicate effects;
- reconnects not to corrupt Match state;
- winner calculation to be server-authoritative;
- server restarts to recover active games where possible;
- replays to match actual gameplay.

### 8.5 Accessibility needs

Users need:

- full keyboard play;
- screen-reader-compatible board semantics;
- non-color-only state indicators;
- reduced motion;
- readable zoomed layouts;
- large touch targets;
- textual equivalents for progress and status.

---

## 9. Product principles

### 9.1 Play immediately

The product must not put unnecessary friction before play.

Avoid:

- mandatory registration;
- mandatory tutorial;
- marketing interstitial;
- forced install prompt;
- cookie banner when no non-essential tracking exists;
- account upsell;
- public profile setup.

### 9.2 Multiplayer first

The homepage and navigation must emphasize:

1. Create Room
2. Join Room
3. Play Solo

Co-op is the flagship mode.

### 9.3 Privacy by omission

Prefer not collecting data over collecting and protecting more data.

### 9.4 Server-authoritative fairness

Multiplayer outcomes must be decided by the server, not by client claims.

### 9.5 Accessibility is a feature

Accessibility is part of product quality and release readiness, not a later compliance task.

### 9.6 Performance is visible quality

Fast loading, immediate interaction, and low resource consumption are part of the user experience.

### 9.7 Calm design

The game should feel focused and premium, not noisy or overstimulating.

### 9.8 Complexity must earn its place

Advanced engineering is justified only when it improves:

- correctness;
- privacy;
- accessibility;
- performance;
- reliability;
- user trust;
- portfolio credibility.

---

## 10. Product priorities

When product trade-offs arise, use this order:

1. Privacy and security
2. Accessibility and usability
3. Performance and resource efficiency
4. Gameplay integrity
5. Visual quality
6. SEO and public presentation
7. Feature breadth

Examples:

- Do not add analytics merely to improve metrics visibility.
- Do not expose puzzle solutions to simplify client behavior.
- Do not add a heavy UI library for convenience.
- Do not add public matchmaking before private rooms are excellent.
- Do not ship animations that impair reduced-motion users.
- Do not add accounts merely to persist statistics.

---

## 11. Core product modes

Status:

- Co-op and online Solo are current MVP modes.
- Race, Duel, and Daily Ninefold are provisional post-MVP modes.

### 11.1 Co-op

Co-op is the primary mode.

Players:

- share one board;
- share notes;
- see player attribution;
- use soft cell focus;
- use pings and reactions;
- solve together;
- complete automatically;
- share one result.

Co-op must feel collaborative rather than like multiple cursors fighting over one board.

### 11.2 Race

Race is the first post-MVP mode.

Players:

- receive the same puzzle;
- solve independently;
- see delayed opponent progress;
- do not see opponent values;
- compete for first valid completion;
- continue during a finishing window;
- receive final ranking.

Race must feel tense and fair without enabling copying.

### 11.3 Duel

Duel follows Race after Race is stable.

Two players:

- share one value board;
- alternate turns;
- use private notes;
- score correct placements and structure completion;
- lose turns on wrong moves;
- may win by score, resignation, forfeit, or abandonment.

Duel must be meaningfully different from a two-player Race.

### 11.4 Solo

Solo supports:

- Easy;
- Medium;
- Hard;
- Expert;
- Random;
- Continue Last Puzzle;
- Guided mode;
- Classic mode;
- notes;
- hints;
- pause;
- local progress;
- local statistics;
- local replay.

Solo exists to provide practice, onboarding, and a complete account-free experience.

### 11.5 Daily Ninefold

Daily Ninefold follows Duel and is not part of the current MVP.

Daily Ninefold provides:

- one shared puzzle per UTC day;
- predictable difficulty rotation;
- local completion;
- local streak;
- local replay;
- no public leaderboard in V1.

It gives users a reason to return without creating an account.

---

## 12. Product feature hierarchy

### 12.1 Current MVP

- responsive homepage;
- Create Room;
- Join Room;
- temporary display names;
- Co-op;
- ready check;
- room code and share link;
- Sudoku board;
- notes;
- error rules;
- reconnect;
- completion;
- results;
- rematch;
- replay;
- replay integrity verification;
- online Solo with device-local progress;
- Nudge and Reveal hints;
- pings and reactions;
- basic host controls;
- light and dark themes;
- TLA+ models for MVP concurrency-sensitive behavior;
- How to Play/help;
- privacy page;
- accessibility page;
- WCAG 2.2 AA core-flow accessibility;
- English localization.

### 12.2 Required before public-ready 1.0

- Race;
- Duel;
- Daily Ninefold;
- German, Albanian, and Turkish;
- polished SEO pages;

### 12.3 Deferred expansion

- active spectators;
- richer replay;
- Explain hints;
- offline Solo;
- PWA installation;
- advanced admin tools;
- broader public promotion.

---

## 13. Homepage product requirements

The homepage should prioritize play, not marketing.

Recommended order:

1. Ninefold Sudoku brand
2. Headline
3. Create Room
4. Join-code input
5. Play Solo
6. Brief Co-op demonstration
7. Future-mode overview clearly labelled as unavailable
8. Privacy/accessibility highlights
9. Footer

### 13.1 Recommended headline

```text
Sudoku is better together.
```

### 13.2 Supporting copy

```text
Create a private room, invite friends, and solve the same Sudoku grid together.
```

### 13.3 Primary actions

- Create a room
- Join with code

### 13.4 Secondary actions

- Play solo

### 13.5 Homepage constraints

- Must load quickly.
- Must contain meaningful SSR HTML.
- Must not require JavaScript to understand the product.
- Must not show participant or Room data.
- Must not force installation.
- Must not force language selection if browser preference is usable.

---

## 14. Private Room experience

### 14.1 Creation

A player creates a Room by selecting:

- temporary display name;
- difficulty.

Co-op is the only enabled multiplayer mode in the MVP. The contract retains an explicit mode value, but the MVP UI does not present unavailable mode choices.

Additional Co-op settings can be changed in the lobby.

### 14.2 Invitation

Room invitation supports:

- six-character code;
- copyable invitation link.

Canonical invitation URL:

```text
https://ninefold.recica.dev/join/7KMP4R
```

### 14.3 Lobby

Lobby must show:

- Room code;
- copy link;
- selected mode;
- selected difficulty;
- player seats;
- spectators where enabled;
- ready state;
- relevant settings;
- host controls;
- short rules summary;
- Start button.

Only settings relevant to the selected mode should be visible.

### 14.4 Host controls

Host can:

- configure difficulty;
- configure errors;
- configure hints;
- lock Room;
- remove or block participants;
- transfer host;
- start;
- cancel countdown;
- rematch.

### 14.5 Room privacy

- No public Room directory.
- No Room search.
- No public participant list.
- No indexing of invitation pages.
- No password in V1.
- Code/link possession grants join request capability.
- MVP Rooms accept valid joins immediately when unlocked and capacity is available.
- Host-approval workflow is deferred.

---

## 15. Temporary identity

### 15.1 Display name

Players enter a temporary name when creating or joining.

The product should communicate that:

- the name is Room-scoped;
- no permanent profile is created;
- the same browser may remember a preferred name locally;
- another Room may use the same name.

### 15.2 No account wall

There is no account registration requirement for:

- Solo;
- private multiplayer;
- replay access with valid capability.

### 15.3 Local statistics

Local-only statistics may include:

- multiplayer games played;
- Co-op completions;
- best Solo times;
- average times;
- hint usage;
- mistakes.

Race, Duel, and Daily statistics may be added locally when those modes enter scope.

The product must clearly state that these are stored on the device and may disappear if browser data is cleared.

---

## 16. Puzzle content requirements

### 16.1 Initial variant

Classic 9×9 Sudoku only.

### 16.2 Difficulty

- Easy
- Medium
- Hard
- Expert

Duel excludes Expert.

### 16.3 Quality

Every published puzzle must:

- have exactly one solution;
- pass deterministic validation;
- have a logical difficulty grade;
- avoid guessing requirements;
- have stable metadata;
- be reproducible;
- be retired if defective.

### 16.4 Catalog

Development catalog target:

- Easy: 100
- Medium: 100
- Hard: 100
- Expert: 50

Total development catalog:

```text
350 puzzles
```

Public promotion target:

- Easy: 500
- Medium: 500
- Hard: 500
- Expert: 250

Total public catalog:

```text
1,750 puzzles
```

The `0.3.3` closed-beta deployment temporarily seeds 10 reviewed test puzzles: 3 Easy, 3 Medium,
2 Hard, and 2 Expert. This supports an initial match and one rematch at every difficulty but does
not satisfy the development or public catalog targets. It must not be represented as the production
catalog.

### 16.5 Repetition control

Users should not encounter the same canonical puzzle repeatedly.

Rematches exclude the previous 20 Room puzzles and transformations of those puzzles.

---

## 17. Hints product requirements

### 17.1 Full model

- Nudge
- Explain
- Reveal

### 17.2 MVP model

Initially:

- Nudge
- Reveal

Explain may ship later when the logical explanation engine and translated instructional copy are ready.

### 17.3 Product behavior

Hints must:

- be clearly identified;
- affect result where defined;
- appear in replay;
- never pretend to explain a technique when no actual logical derivation exists.

---

## 18. Replay product

Replay is a primary product feature, not an admin-only diagnostic.

### 18.1 User value

Replay helps players:

- understand how the group solved;
- compare Race strategies;
- review Duel turns;
- revisit decisive moments;
- learn;
- share a result;
- trust that the Match history is intact.

### 18.2 Basic replay

Must support:

- play;
- pause;
- scrub;
- speed control;
- value placement;
- erase;
- notes;
- hints;
- disconnects;
- completion.

### 18.3 Mode-specific replay

#### Co-op

- one shared board;
- player attribution;
- pings and reactions where recorded.

#### Race

Provisional:

- side-by-side boards;
- synchronized progress;
- finish order.

#### Duel

Provisional:

- turns;
- score changes;
- time spent;
- accepted and rejected attempts without exposing hidden digits.

### 18.4 Integrity indicator

Completed multiplayer replay should display:

```text
Replay integrity verified
```

only after successful client-side hash-chain and signature verification.

The interface must explain that this verifies the replay was not changed after sealing. It must not overstate the cryptographic guarantee.

### 18.5 Retention

- Multiplayer replay: 7 days
- Solo replay: device-local
- A participant with a still-valid originating Room session may delete the shared replay early
- Replays are not publicly indexed
- Participant-linked Match events, snapshots, and names are scrubbed when replay retention ends
- A non-identifying operational Match tombstone may remain for 30 days

---

## 19. Signature engineering showcase

Ninefold has two advanced engineering features that are also meaningful to product trust.

### 19.1 Formally checked state machines

TLA+ models cover:

- Room lifecycle;
- readiness;
- Match start;
- duplicate commands;
- stale timers;
- reconnect;
- recovery;
- completion.

Race finish and Duel turn ownership are added to the formal model when those modes enter scope.

Portfolio framing:

> Ninefold’s multiplayer lifecycle is formally modeled and checked against duplicate messages, stale timers, reconnections, and competing completion events.

### 19.2 Cryptographically verifiable replay

Durable events form a hash chain and the final digest is signed with Ed25519.

Portfolio framing:

> Ninefold replays are deterministic and tamper-evident, with browser-side integrity verification.

### 19.3 Product constraint

These features must remain subtle during normal play.

They should increase trust and portfolio quality without adding jargon to the main game flow.

---

## 20. Privacy product requirements

### 20.1 Core commitments

Ninefold must have:

- no player account requirement;
- no advertising;
- no product analytics;
- no behavioral tracking;
- no session replay service;
- no public player directory;
- no permanent friend graph;
- no direct messages;
- no real-name requirement;
- no email requirement for play;
- no cross-site tracking;
- no third-party fonts or scripts by default.

### 20.2 Essential storage

The privacy page must explain:

- secure session cookies;
- local browser storage;
- temporary multiplayer data;
- temporary replay data;
- operational security logs;
- local statistics;
- retention.

Full participant-linked multiplayer data is retained for at most seven days. After that, names, participant links, event payloads, snapshots, and replay capabilities are removed. A minimal non-identifying Match tombstone may remain for operational diagnosis for up to 30 days.

IP addresses may be processed transiently for abuse prevention. Application rate limiting should use short-lived keyed representations rather than persist raw addresses where practical.

### 20.3 Cookie banner

Do not show a consent banner when only essential cookies and local storage are used.

### 20.4 Clear local data

Settings must provide a clear action to remove local:

- preferences;
- statistics;
- Solo attempts;
- Daily progress;
- local replay;
- recent puzzle history.

### 20.5 Contact

Public contact:

```text
ninefold@recica.dev
```

It may forward to a private mailbox.

---

## 21. Accessibility product requirements

Target:

```text
WCAG 2.2 AA
```

### 21.1 Required capabilities

- complete keyboard play;
- visible focus;
- screen-reader labels;
- row and column announcements;
- clue/value/note semantics;
- non-color-only status;
- reduced motion;
- scalable text;
- 200% zoom;
- Sudoku cells at least 24×24 CSS pixels;
- number-pad buttons and primary controls at least 44×44 CSS pixels;
- accessible connection status.

Textual Race progress and accessible Duel turn status become requirements when those provisional modes enter scope.

### 21.2 Release quality

A feature is not complete when it works only visually.

Every gameplay state must be understandable without relying only on:

- color;
- sound;
- motion;
- hover;
- spatial position.

### 21.3 Timed play

Before readiness, players must understand:

- reconnect rules;

The MVP has no overall Match deadline. Duel turn duration, Race finishing windows, and timeout consequences become requirements with those modes.

---

## 22. Localization product requirements

Current MVP locale:

- English (`en`)

Planned pre-1.0 locales:

- German (`de`)
- Albanian (`sq`)
- Turkish (`tr`)

Arabic is explicitly excluded.

### 22.1 Language behavior

- Each player may use a different language.
- Room and Match semantics do not depend on locale.
- Browser preference determines initial locale.
- User may override in Settings.
- Saved preference wins on return.
- English is fallback.
- MVP layouts are tested with pseudo-localized expansion before the additional catalogs exist.

### 22.2 Translation quality

Before broad public promotion:

- gameplay terms;
- error messages;
- settings;
- privacy;
- accessibility;
- result copy;

should receive fluent human review.

Machine translation may be used only as a draft.

### 22.3 Text expansion

Layouts must support German expansion and Turkish/Albanian grammar without truncation.

---

## 23. SEO product requirements

SEO matters for portfolio discovery and public consumer discovery.

### 23.1 Indexable pages

- Home
- How to Play
- Privacy
- Accessibility

About and dedicated mode pages are deferred until post-MVP public presentation work.

### 23.2 Search positioning

Content may naturally target:

- multiplayer Sudoku;
- cooperative Sudoku;
- play Sudoku with friends;
- private Sudoku room;
- account-free Sudoku.

Race- and Duel-specific search positioning is deferred until those modes are available.

### 23.3 Metadata

Primary title:

```text
Ninefold Sudoku — Multiplayer Sudoku
```

Primary description:

```text
Play Sudoku together in private co-op rooms or solve a personal puzzle without creating an account.
```

### 23.4 SEO constraints

- Private Room pages are not indexed.
- Replay URLs are not indexed.
- Participant names never appear in search metadata.
- SEO copy must not block immediate access to play.
- Public pages must remain fast and accessible.

---

## 24. Design product direction

The desired personality is:

- modern;
- calm;
- geometric;
- friendly;
- refined;
- slightly playful;
- distinctly multiplayer.

Avoid:

- newspaper imitation;
- generic dashboard appearance;
- loud arcade effects;
- excessive gradients;
- heavy shadows;
- clutter;
- over-animation;
- childish styling.

### 24.1 Board priority

The board is the dominant element during play.

Secondary controls must not visually overwhelm it.

### 24.2 Themes

- Light mode
- Dark mode
- System default initially
- Manual override
- Local persistence

### 24.3 Player identity

Each player receives:

- color;
- marker shape or icon;
- initials;
- display name.

Color is never the sole identity cue.

### 24.4 Sound

Sound is off by default.

Optional effects may include:

- countdown;
- accepted move;
- mistake;
- turn;
- ping;
- completion.

### 24.5 Motion

Motion should be restrained and respect reduced-motion settings.

---

## 25. Mobile and desktop product requirements

### 25.1 Browser priority

Browser is the only planned platform.

No native application is planned.

### 25.2 Mobile

Mobile must support:

- one-handed play where practical;
- fixed number pad below board;
- full-width readable board;
- expandable participant/activity panel;
- large touch targets;
- no horizontal page scrolling.

### 25.3 Desktop

Desktop must support:

- full keyboard entry;
- board-centered layout;
- participant/status panel;
- controls/activity panel;
- efficient number entry;
- visible shortcuts.

### 25.4 Tablet

Tablet layout adapts based on available width.

---

## 26. Performance product requirements

Performance is a visible feature.

### 26.1 Budgets

| Metric | Product target |
|---|---:|
| Public page Lighthouse performance | 95+ |
| Homepage compressed JavaScript | under 100 KiB |
| Gameplay compressed JavaScript | under 200 KiB |
| Local cell feedback | under 50 ms |
| Same-region server acknowledgement | normally under 150 ms |
| Homepage interactive on normal mobile network | under 3 seconds |
| Ordinary replay load | under 2 seconds |

### 26.2 User-facing performance principles

- Gameplay should feel immediate.
- Disconnection status should appear quickly.
- Slow network must not corrupt state.
- Replay must not block result display.
- Mode-specific code should load only when required.
- Marketing content must not make play slow.

---

## 27. Product reliability requirements

Ninefold should recover gracefully from:

- page refresh;
- temporary network loss;
- browser sleep;
- duplicate command;
- stale command;
- server deployment;
- server restart;
- disconnected host;
- disconnected player;
- replay verification failure.

### 27.1 User communication

The UI must distinguish:

- connected;
- reconnecting;
- synchronized;
- read-only;
- server maintenance;
- unrecoverable Match cancellation.

### 27.2 No silent data corruption

When the server cannot persist authoritative commands:

- do not pretend moves succeeded;
- pause Match if necessary;
- communicate temporary failure;
- preserve last committed state.

---

## 28. Product success

### 28.1 Primary outcome

The primary product outcome is:

> People who create a multiplayer Room successfully start and complete a Match.

A conceptual success ratio is:

```text
completed multiplayer sessions / created multiplayer rooms
```

### 28.2 Privacy constraint

Ninefold does not implement product analytics in V1.

Therefore product success is evaluated through:

- private testing;
- direct feedback;
- observed family/friend use;
- bug reports;
- necessary Room and Match records already required to operate the product;
- optional manual analysis that does not create a tracking system.

Do not add analytics solely to measure this ratio.

### 28.3 Qualitative success signals

- Users understand how to create and join without explanation.
- Players complete Co-op games.
- Players rematch.
- Refresh does not cause panic or data loss.
- Mobile board feels comfortable.
- Keyboard users can play quickly.
- Screen-reader users can understand core board state.
- Portfolio visitors recognize technical depth.
- VPS resource use remains low.

---

## 29. MVP definition

The `0.3.0` MVP is the current implementation target. It is not the full product.

### 29.1 MVP scope

MVP includes:

- public home page;
- Create Room;
- Join by code/link;
- temporary display names;
- Co-op mode;
- 1–6 players;
- one active multiplayer Room per browser profile;
- Lobby readiness;
- shared board;
- shared notes;
- soft locks;
- player attribution;
- durable targeted pings;
- ephemeral reactions;
- error presets;
- hints: Nudge and Reveal;
- basic host settings, lock, remove/block, transfer, start/cancel, and rematch controls;
- reconnect;
- server-restart recovery;
- completion;
- result;
- basic replay;
- replay integrity verification;
- TLA+ models for Room lifecycle, idempotency, stale timers, recovery, and completion;
- rematch;
- Solo;
- light/dark mode;
- responsive mobile and desktop;
- English;
- pseudo-localized expansion testing;
- How to Play/help page;
- privacy page;
- accessibility page;
- WCAG 2.2 AA core-flow accessibility;
- private MVP administration;
- Coolify deployment.

### 29.2 MVP replay

MVP replay includes:

- play;
- pause;
- scrub;
- speed;
- value placement;
- erase;
- notes;
- pings;
- disconnects;
- completion;
- integrity status.

MVP replay does not require:

- Key Moments;
- advanced strategic analysis;
- comparison charts;
- public sharing gallery.

### 29.3 MVP admin

- health;
- Room lookup;
- terminate broken Room;
- delete replay;
- retire puzzle.

### 29.4 MVP spectators

Protocol and domain support spectator roles.

Active spectator UX may remain limited until later phases.

### 29.5 MVP exclusions

- Race
- Duel
- Daily Ninefold
- full Explain hints
- full offline Solo
- host-approval join workflow
- multiplayer undo
- optional overall Match deadlines
- full spectator UX
- German, Albanian, and Turkish catalogs
- PWA installation
- About and dedicated mode/SEO pages
- public rankings
- account system
- public matchmaking
- rich reporting workflow
- permanent staging environment
- external error tracking
- off-VPS backup during private testing

---

## 30. First vertical slice

The first complete vertical slice is:

```text
Create Room
→ Join with code
→ Ready check
→ Start Co-op puzzle
→ Place and erase values
→ Complete puzzle
→ View replay
→ Start rematch
```

The slice must:

- work in two isolated browser contexts;
- survive page refresh;
- persist accepted moves;
- reject duplicate commands;
- recover authoritative state;
- support keyboard interaction;
- work on mobile viewport;
- include tests.

No later feature should be built before this slice is coherent.

---

## 31. Release milestones

### 31.1 Milestone A — Local prototype

- one Go server;
- one SvelteKit client;
- hardcoded puzzle;
- in-memory Room;
- two browser players;
- basic Co-op placements;
- no production deployment.

### 31.2 Milestone B — Persistent Co-op alpha

- SQLite;
- Room codes;
- generated puzzle catalog;
- ready check;
- reconnect;
- completion;
- rematch;
- snapshots;
- event persistence.

### 31.3 Milestone C — Portfolio beta

- full `0.3.0` MVP;
- polished responsive design;
- basic replay;
- cryptographic replay verification;
- MVP TLA+ models and CI checks;
- Solo;
- Home, How to Play, Privacy, and Accessibility pages;
- English and pseudo-localized layout testing;
- Coolify;
- public repository;
- private administration and production hardening.

### 31.4 Post-MVP sequence

1. Race
2. Duel
3. Daily Ninefold
4. full spectators, richer replay, Explain hints, and PWA installation
5. additional locales, SEO mode pages, backups, and public-launch hardening

### 31.5 Version map

```text
0.1.0  Co-op local prototype
0.2.0  Persistent multiplayer alpha
0.3.0  Full MVP / portfolio beta
0.4.0  Race mode
0.5.0  Duel mode
0.6.0  Daily Ninefold
1.0.0  Public-ready release
```

---

## 32. Public-ready release criteria

Before `1.0.0`:

- Co-op is stable.
- Race is stable.
- Duel is stable.
- Solo is stable.
- Daily is stable.
- replay verification is reliable;
- recovery tests pass;
- accessibility review passes;
- all four languages receive human review;
- SEO pages are complete;
- off-VPS backups exist;
- restore test passes;
- security tests pass;
- Room-code enumeration protection passes;
- WebSocket payload tests pass;
- admin access is protected;
- public privacy content is accurate;
- performance budgets are met or documented.

---

## 33. Broader public promotion gate

Private testing and portfolio beta may proceed without backups if data loss is accepted.

Broader consumer promotion is blocked until:

- automated off-VPS backup exists;
- restore procedure is tested;
- production monitoring is reliable;
- retention jobs are verified;
- replay deletion works;
- security gate is complete.

---

## 34. Non-functional requirements

### 34.1 Privacy

No non-essential tracking.

### 34.2 Accessibility

WCAG 2.2 AA target.

### 34.3 Performance

Meet published budgets.

### 34.4 Availability

V1 accepts brief deployment interruption and single-instance limits.

### 34.5 Recovery

Active Matches recover where domain rules permit.

### 34.6 Maintainability

One developer must be able to understand and operate the system.

### 34.7 Portability

Avoid unnecessary vendor lock-in.

### 34.8 Security

Server authority, strict validation, secure cookies, replay capabilities, and private admin access.

### 34.9 SEO

Public pages are SSR, indexable, localized, semantic, and fast.

---

## 35. Explicit non-goals

The following are outside current product scope:

- player accounts;
- password login;
- OAuth;
- permanent cloud profiles;
- public matchmaking;
- ranked ladder;
- Elo/Glicko/TrueSkill;
- seasons;
- public leaderboard;
- friend list;
- following;
- direct messages;
- public text chat;
- voice chat;
- persistent teams;
- clubs;
- tournament system;
- paid entry;
- advertisements;
- subscription;
- premium gameplay advantage;
- puzzle marketplace;
- user-generated public puzzles;
- native iOS app;
- native Android app;
- desktop native app;
- public API for third-party developers;
- multiple Sudoku variants;
- Arabic localization;
- client-side SQLite;
- automatic cloud sync;
- blockchain;
- peer-to-peer gameplay;
- CRDT board authority;
- multi-region deployment;
- Kubernetes.

---

## 36. Deferred but possible features

These are not rejected permanently, but must not be implemented without explicit scope change:

- offline Solo;
- advanced educational Explain hints;
- user-entered newspaper puzzle;
- custom puzzle creation;
- advanced replay Key Moments;
- public result card image generation;
- richer spectator UX;
- optional private text chat;
- tournament brackets;
- accounts for cross-device sync;
- permanent statistics;
- additional languages;
- alternative Sudoku variants.

---

## 37. Product risks

### 37.1 Scope expansion

Risk:

- implementing all modes simultaneously delays a playable product.

Mitigation:

- vertical slice first;
- Co-op before Race/Duel;
- feature flags;
- scope authority in this document.

### 37.2 Multiplayer complexity

Risk:

- duplicate events, stale timers, reconnect bugs.

Mitigation:

- Room actors;
- versioning;
- idempotency;
- TLA+;
- deterministic replay;
- recovery tests.

### 37.3 Replay complexity

Risk:

- replay becomes a second product that delays core play.

Mitigation:

- basic replay first;
- advanced analysis deferred;
- event log already required for recovery.

### 37.4 Translation quality

Risk:

- inaccurate game or error terminology.

Mitigation:

- English canonical;
- structured keys;
- human review before broad promotion.

### 37.5 No analytics

Risk:

- limited product funnel visibility.

Mitigation:

- direct testing;
- feedback;
- manual review of operationally necessary data;
- preserve privacy commitment.

### 37.6 Single VPS

Risk:

- downtime and data loss.

Mitigation:

- graceful recovery;
- backups before broad launch;
- clear architecture scaling boundary.

### 37.7 Accessibility complexity

Risk:

- board semantics and multiplayer announcements become difficult.

Mitigation:

- semantic HTML;
- structured events;
- accessibility testing from first slice;
- no canvas board.

---

## 38. Product acceptance criteria

### 38.1 Room creation

- User can create a Room without account.
- Room code is visible and copyable.
- Invitation link works.
- Host identity is clear.
- Safe errors are localized.

### 38.2 Room joining

- User can join with code or link.
- Temporary name is sufficient.
- Participant names are not exposed before join.
- Full Room explains that no player seat is available.
- Locked Room is clearly communicated.

An active spectator option is added when spectator UX enters scope.

### 38.3 Co-op

- Two players can edit one board.
- Simultaneous conflicts resolve consistently.
- Shared notes update.
- Attribution is visible and accessible.
- Soft lock warns but does not permanently block.
- Reconnect restores the same player.
- Completion is automatic.

### 38.4 Race

These criteria are provisional and apply when Race enters scope.

- Same puzzle assigned to all.
- Boards are isolated.
- Opponent values remain hidden.
- Progress reflects correct cells only.
- Winner uses server time.
- Finishing window works.
- Ranking is deterministic.

### 38.5 Duel

These criteria are provisional and apply when Duel enters scope.

- Exactly two players.
- Only active player can move.
- Wrong move loses turn.
- Score is correct.
- Timeout rules work.
- Disconnect pause works once.
- Resignation ends Match.

### 38.6 Solo

- Puzzle can be started and resumed.
- Guided and Classic differ correctly.
- Notes work.
- Pause works.
- Hints apply penalties.
- Local completion history persists.

### 38.7 Replay

- Reconstructs exact accepted gameplay.
- Does not expose hidden information.
- Scrubbing works.
- Integrity verification works.
- Replay can be deleted.
- Replay expires.

### 38.8 Privacy

- No analytics request is sent.
- No third-party tracker loads.
- Local data can be cleared.
- Room pages are not indexed.
- Secure cookies are used in production.

### 38.9 Accessibility

- Core flow works keyboard-only.
- Board state is screen-reader understandable.
- No color-only state.
- 200% zoom is usable.
- Reduced motion works.
- Touch targets meet minimum size.

### 38.10 SEO

- Public pages render useful HTML.
- Canonicals are correct.
- English metadata exists.
- Private pages are noindex.
- Sitemap excludes Rooms and replays.

Additional localized metadata is required when those locales enter scope.

---

## 39. Product content and messaging

### 39.1 Home headline

```text
Sudoku is better together.
```

### 39.2 Home supporting text

```text
Create a private room, invite friends, and solve the same Sudoku grid together.
```

### 39.3 Privacy summary

Suggested concise statement:

```text
No accounts. No ads. No tracking. Your personal progress stays on your device.
```

### 39.4 Replay summary

Suggested concise statement:

```text
Every multiplayer match can be replayed and verified for integrity.
```

### 39.5 Portfolio summary

```text
A privacy-first multiplayer Sudoku game built with SvelteKit, Go, WebSockets, SQLite, formally checked state machines, and cryptographically verifiable replays.
```

---

## 40. About page requirements

This page is deferred until after the MVP.

The About page should explain:

- why Ninefold exists;
- why Co-op is primary;
- why no account is required;
- how private Rooms work;
- how replay works;
- privacy stance;
- accessibility stance;
- technical architecture;
- formal verification;
- public repository.

It should remain understandable to non-engineers, with deeper technical sections clearly separated.

---

## 41. How-to-play content

The product should include clear instructions for:

- Sudoku basics;
- notes;
- Guided versus Classic;
- Co-op;
- hints;
- error presets;
- reconnect;
- replay.

Race, Duel, and mode-specific timer instructions are added when those features enter scope.

Instructions should use:

- plain language;
- localized terminology;
- diagrams or examples where useful;
- no assumption that the player knows Sudoku jargon.

---

## 42. Error and empty-state product requirements

Every major failure needs a purposeful state.

Examples:

- Room not found;
- Room expired;
- Room locked;
- name rejected;
- player seat full;
- reconnecting;
- synchronization failed;
- Match recovered;
- Match cancelled after recovery;
- replay expired;
- replay verification failed;
- unsupported browser capability;
- local storage unavailable.

Error states must:

- explain what happened;
- state whether user can retry;
- avoid technical stack traces;
- include request ID on unexpected failures;
- preserve privacy.

---

## 43. Onboarding

No mandatory tutorial.

Use contextual, dismissible guidance for:

- first Room creation;
- first join;
- first notes use;
- first Co-op ping;
- first replay;

Race and Duel onboarding is added with those modes.

Guidance should not repeatedly interrupt experienced users.

---

## 44. Installation

PWA installation is deferred until after the MVP.

When implemented, do not show an immediate install prompt.

When PWA installation enters scope, offer a subtle install action after:

- completing a Solo puzzle; or
- completing a multiplayer Match.

Installation is optional and must not change functionality expectations.

---

## 45. Source-code visibility

The source repository is public.

The product may link to the repository from:

- footer;
- About page;
- technical case study;
- result page through a subtle “How Ninefold works” link.

Do not place repository links prominently inside active gameplay.

The repository must exclude:

- production secrets;
- live database;
- session keys;
- private signing keys;
- admin access details;
- personal data.

---

## 46. Product decision process

A product-scope change must explicitly update this document when it affects:

- supported modes;
- target users;
- privacy commitments;
- account model;
- public/private access;
- localization;
- monetization;
- MVP scope;
- release gates;
- non-goals;
- platform support.

A technical implementation must not silently introduce a product capability.

Examples:

- Adding OAuth is a product change.
- Adding public matchmaking is a product change.
- Adding analytics is a product and privacy change.
- Adding a new Sudoku variant is a product and domain change.
- Adding public replay sharing is a product and privacy change.

---

## 47. Definition of product done

A feature is product-complete only when:

- it solves the intended user need;
- it is in current scope;
- gameplay rules are documented;
- implementation meets architecture constraints;
- visual and interaction design is complete;
- mobile and desktop behavior is complete;
- accessibility is complete;
- current-scope localization is complete;
- privacy impact is acceptable;
- performance budget is reviewed;
- errors and empty states exist;
- reconnect behavior is defined;
- tests exist;
- documentation is updated.

A feature is not complete merely because the happy path works.

---

## 48. Final product directive

Ninefold should feel like a carefully designed game first and an impressive engineering project second.

The engineering should be discoverable, credible, and exceptional, but normal players should mainly experience:

- fast entry;
- clear rules;
- satisfying Sudoku;
- reliable multiplayer;
- privacy;
- accessibility;
- beautiful design;
- trustworthy replay;
- easy rematch.

Whenever a feature makes Ninefold noisier, heavier, less private, less accessible, or harder to understand without materially improving play, it should be simplified, deferred, or removed.
