# Ninefold Sudoku — Product Specification

**Document:** `docs/PRODUCT.md`  
**Status:** Canonical product requirements and scope  
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

---

## 2. Product summary

Ninefold Sudoku is a privacy-first, multiplayer-first Sudoku web application built for private play among family and friends and designed to function as a high-quality public portfolio showcase.

The product allows people to:

- create a private Sudoku room;
- invite others with a short code or link;
- solve one board together in Co-op mode;
- race on identical puzzles in Race mode;
- compete turn by turn in Duel mode;
- play alone in Solo mode;
- return for one shared Daily Ninefold puzzle;
- view deterministic, cryptographically verifiable match replays.

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
- Daily Ninefold;
- a trustworthy account-free product.

### 7.3 User characteristics

The product must support users who:

- are new to Sudoku;
- are casual players;
- are advanced solvers;
- primarily use a mobile browser;
- primarily use a desktop keyboard;
- use assistive technology;
- have slower networks;
- do not want to create an account;
- speak English, German, Albanian, or Turkish.

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
4. Daily Ninefold

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

Race is the second-priority multiplayer mode.

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

Duel is the third-priority multiplayer mode.

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

### 12.1 Essential

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
- privacy page;
- accessibility basics;
- English localization.

### 12.2 Important

- Solo;
- Daily Ninefold;
- light and dark themes;
- German, Albanian, and Turkish;
- pings and reactions;
- host controls;
- replay integrity verification;
- polished SEO pages;
- mobile installability;
- formal verification artifacts.

### 12.3 Expansion

- Race;
- Duel;
- active spectators;
- richer replay;
- Explain hints;
- offline Solo;
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
6. Daily Ninefold
7. Brief Co-op demonstration
8. Mode overview
9. Privacy/accessibility highlights
10. Portfolio and architecture links
11. Footer

### 13.1 Recommended headline

```text
Sudoku is better together.
```

### 13.2 Supporting copy

```text
Create a private room, invite friends, and solve the same grid together—or compete in Race and Duel modes.
```

### 13.3 Primary actions

- Create a room
- Join with code

### 13.4 Secondary actions

- Play solo
- Daily Ninefold

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
- mode;
- difficulty.

Additional settings can be changed in the lobby.

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

- configure mode;
- configure difficulty;
- configure timer;
- configure errors;
- configure hints;
- lock Room;
- enable approval;
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
- Daily;
- private multiplayer;
- replay access with valid capability.

### 15.3 Local statistics

Local-only statistics may include:

- multiplayer games played;
- Race wins;
- Duel record;
- Co-op completions;
- best Solo times;
- average times;
- Daily streak;
- hint usage;
- mistakes.

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

- side-by-side boards;
- synchronized progress;
- finish order.

#### Duel

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
- Participant may delete shared replay early
- Replays are not publicly indexed

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
- Race finish;
- Duel turn ownership;
- recovery;
- completion.

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
- touch targets at least 44×44 CSS pixels;
- textual Race progress;
- accessible Duel turn status;
- accessible connection status.

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

- overall time limit;
- Duel turn duration;
- reconnect rules;
- finishing window;
- timeout consequences.

---

## 22. Localization product requirements

Launch locales:

- English (`en`)
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
- About
- How to Play
- Co-op mode page
- Race mode page
- Duel mode page
- Privacy
- Accessibility

### 23.2 Search positioning

Content may naturally target:

- multiplayer Sudoku;
- cooperative Sudoku;
- play Sudoku with friends;
- private Sudoku room;
- online Sudoku race;
- turn-based Sudoku;
- account-free Sudoku.

### 23.3 Metadata

Primary title:

```text
Ninefold Sudoku — Multiplayer Sudoku
```

Primary description:

```text
Play Sudoku together in private co-op rooms, race friends on identical puzzles, or challenge someone to a turn-based duel.
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

The MVP is not the full product.

### 29.1 MVP scope

MVP includes:

- public home page;
- Create Room;
- Join by code/link;
- temporary display names;
- Co-op mode;
- 1–6 players;
- ready check;
- shared board;
- shared notes;
- soft locks;
- player attribution;
- pings/reactions;
- error presets;
- hints: Nudge and Reveal;
- reconnect;
- completion;
- result;
- basic replay;
- replay integrity verification;
- rematch;
- Solo;
- light/dark mode;
- responsive mobile and desktop;
- English;
- privacy page;
- accessibility baseline;
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
- full Explain hints
- full offline Solo
- advanced Daily experience
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

- polished responsive design;
- basic replay;
- cryptographic replay verification;
- Solo;
- Daily;
- public pages;
- SEO;
- English plus additional translations;
- Coolify;
- public repository;
- architecture case study.

### 31.4 Milestone D — Multiplayer expansion

- Race;
- Duel;
- full spectators;
- richer replay;
- Explain hints;
- stronger public-launch hardening.

### 31.5 Version map

```text
0.1.0  Co-op local prototype
0.2.0  Persistent multiplayer alpha
0.3.0  Portfolio beta
0.4.0  Race mode
0.5.0  Duel mode
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
- Full Room offers spectator option where permitted.
- Locked Room is clearly communicated.

### 38.3 Co-op

- Two players can edit one board.
- Simultaneous conflicts resolve consistently.
- Shared notes update.
- Attribution is visible and accessible.
- Soft lock warns but does not permanently block.
- Reconnect restores the same player.
- Completion is automatic.

### 38.4 Race

- Same puzzle assigned to all.
- Boards are isolated.
- Opponent values remain hidden.
- Progress reflects correct cells only.
- Winner uses server time.
- Finishing window works.
- Ranking is deterministic.

### 38.5 Duel

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
- Localized metadata exists.
- Private pages are noindex.
- Sitemap excludes Rooms and replays.

---

## 39. Product content and messaging

### 39.1 Home headline

```text
Sudoku is better together.
```

### 39.2 Home supporting text

```text
Create a private room, invite friends, and solve the same grid together—or compete in Race and Duel modes.
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
- Race;
- Duel;
- hints;
- error presets;
- timers;
- reconnect;
- replay.

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
- first Duel turn;
- first Race progress view.

Guidance should not repeatedly interrupt experienced users.

---

## 44. Installation

Ninefold may be installable as a PWA.

Do not show an immediate install prompt.

Offer a subtle install action after:

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
- localization is complete;
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
