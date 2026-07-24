# Ninefold Sudoku — Domain Specification

**Document:** `docs/DOMAIN.md`  
**Status:** Canonical domain specification  
**Product:** Ninefold Sudoku  
**Public URL:** `https://ninefold.recica.dev`  
**Repository:** `ninefold-sudoku`  
**Default branch:** `master`

---

## 1. Purpose and authority

This document defines the canonical gameplay and business behavior of Ninefold Sudoku. It is the source of truth for:

- domain terminology;
- aggregate boundaries;
- entities and value objects;
- room and match lifecycles;
- commands and domain events;
- gameplay invariants;
- Co-op, Race, Duel, Solo, and Daily rules;
- disconnect, reconnect, recovery, and rematch behavior;
- puzzle eligibility and lifecycle;
- replay semantics;
- domain-level privacy and retention;
- domain error codes;
- acceptance scenarios.

This document does **not** define framework-specific package structure, HTTP routing, SQL schema, CSS implementation, container deployment, or UI styling. Those belong in `ARCHITECTURE.md` and `DESIGN.md`.

When another document conflicts with this one about gameplay behavior or invariants, this document wins unless it is intentionally amended.

The rollout scope of a feature is controlled by `PRODUCT.md`. A feature may be fully specified here while still disabled or deferred in the current release.

---

## 2. Product domain summary

Ninefold Sudoku is a privacy-first, multiplayer-first Sudoku game designed for private play among friends and family while also serving as a technically ambitious portfolio project.

The supported play experiences are, in priority order:

1. **Co-op** — several players solve one shared Sudoku board.
2. **Race** — several players independently solve the same puzzle.
3. **Duel** — two players alternate turns on one shared board and compete by score.
4. **Solo** — one player solves a personal puzzle.
5. **Daily Ninefold** — one shared daily puzzle for all visitors.

The product does not require player accounts. Multiplayer uses room-scoped temporary identities. Public matchmaking, permanent social graphs, direct messaging, public profiles, and account-based ranking are outside the defined V1 domain.

The principal product promise is:

> Create a private room, invite people with a code or link, and play Sudoku together with minimal friction.

The signature engineering and product feature is deterministic match replay, including cryptographic integrity verification.

---

## 3. Domain principles

The following rules apply across the domain.

### 3.1 Server authority

For multiplayer:

- the server is authoritative for room state, match state, clocks, accepted commands, completion, ranking, scoring, reconnect eligibility, and replay history;
- client time, client score, client correctness claims, and client completion claims are never authoritative;
- no client may directly mutate shared domain state;
- a state-changing command becomes real only after server validation and durable commit.

### 3.2 Determinism

Given the same:

- assigned puzzle revision;
- transformation seed;
- immutable match rules;
- ordered durable event stream;

the final match state and replay must reconstruct deterministically.

### 3.3 Immutability

- Match rules are immutable after match creation.
- Accepted durable events are append-only.
- Finalized match results are immutable except for explicit administrative invalidation.
- A completed match never becomes active again.
- A rematch creates a new Match; it never resets an existing Match.

### 3.4 Privacy by minimization

- No real name, email address, location, or account is required to play.
- Player display names are temporary and room-scoped.
- Personal solo history and preferences are device-local.
- Multiplayer replays expire automatically.
- Domain events contain stable identifiers and structured data, not unnecessary personal information.

### 3.5 Accessibility as domain behavior

Where gameplay state must be communicated, the domain must expose enough structured information to support:

- text equivalents for heatmaps and progress;
- explicit turn ownership;
- explicit error states;
- explicit completion and score changes;
- participant markers independent of color;
- localized client rendering.

Domain events must not depend on color, animation, sound, or layout as the sole carrier of meaning.

---

## 4. Canonical terminology

| Term | Meaning |
|---|---|
| **Room** | A temporary private gathering space that owns membership, host authority, configuration, readiness, and the current match reference. |
| **Match** | One immutable-rules play session using one assigned puzzle. |
| **Participant** | A room-scoped temporary identity. |
| **Player** | A participant occupying a playable seat in the current room or match. |
| **Spectator** | A participant who may observe but not perform player commands. |
| **Host** | The participant with room-management authority. Host status is separate from player/spectator participation. |
| **Puzzle** | A verified Sudoku definition with clues, solution, difficulty, quality metadata, lifecycle, and revision. |
| **Board** | The current 81-cell state being solved. |
| **Shared Board** | One board jointly mutated by Co-op or Duel participants. |
| **Player Board** | A board visible and mutable only by one Race participant. |
| **Clue** | A fixed original puzzle value that can never be changed. |
| **Value** | A placed digit from 1 through 9. |
| **Note** | A pencil-mark candidate associated with one cell. |
| **Move** | A domain-level value or note mutation attempted by a player. |
| **Command** | A request to perform a domain action. Commands may be accepted or rejected. |
| **Domain Event** | An immutable fact produced by an accepted domain operation. |
| **Durable Event** | An event persisted and used for recovery or replay. |
| **Ephemeral Event** | A short-lived event used for presence or visual coordination but not authoritative recovery. |
| **Replay** | A deterministic reconstruction of a completed or retained match from authoritative events. |
| **Error Preset** | The configured behavior for incorrect value entry. |
| **Soft Lock** | A temporary, overridable Co-op claim indicating active attention on a cell. |
| **Finishing Window** | The post-winner interval during which remaining Race players may finish. |
| **Daily Ninefold** | The globally shared puzzle assigned for one UTC calendar date. |
| **Assisted Result** | A result marked as having used a hint. |
| **Recovery Pending** | A temporary state indicating that a match must be reconstructed or resumed after server interruption. |
| **Invalidated Result** | A completed result retained historically but marked invalid by an administrator. |

---

## 5. Identifiers and value objects

The domain uses strongly typed identifiers. The exact serialization belongs to architecture, but durable identifiers are expected to use non-guessable UUIDv7 values.

Canonical identifier types:

- `RoomID`
- `MatchID`
- `ParticipantID`
- `PuzzleID`
- `ReplayID`
- `DailyChallengeID`
- `RequestID`
- `ConnectionID`
- `ReportID`

### 5.1 Room code

A `RoomCode`:

- contains exactly six characters;
- is case-insensitive;
- uses only the alphabet:

```text
ABCDEFGHJKLMNPQRSTUVWXYZ23456789
```

The characters `I`, `O`, `0`, and `1` are excluded to reduce ambiguity.

Example:

```text
7KMP4R
```

Room codes are invitation secrets, not durable identities. Internal references use `RoomID`.

### 5.2 Cell index

A Sudoku cell is identified by an integer from `0` through `80`.

```text
row    = index / 9
column = index % 9
box    = (row / 3) * 3 + column / 3
```

Rows, columns, and boxes are internally zero-based. User-facing clients may localize them as 1–9.

### 5.3 Digit

A Sudoku `Digit` is an integer from `1` through `9`.

Zero is not a digit. An empty cell is represented by absence of a value, not digit zero.

### 5.4 Candidate set

A `CandidateSet` contains zero or more unique digits from 1 through 9.

### 5.5 Aggregate version

Every mutable Room and Match has a monotonically increasing aggregate version.

A command that depends on current state supplies an expected version. Stale commands are rejected or recovered through the transport protocol; the domain never silently applies a stale mutation.

### 5.6 Event number

Every durable Match event has a strictly increasing event number scoped to one Match.

---

## 6. Aggregate boundaries

Ninefold uses four primary transactional aggregates:

1. `Room`
2. `Match`
3. `Puzzle`
4. `DailyChallenge`

### 6.1 Room aggregate

Room owns:

- room code;
- lifecycle;
- expiration;
- host identity;
- participant membership;
- spectator membership;
- join approval;
- room locking;
- participant blocks;
- readiness;
- configuration;
- current match reference;
- rematch sequence;
- lightweight lobby history.

Room does **not** own:

- Sudoku board state;
- accepted gameplay moves;
- match scoring;
- winner calculation;
- replay event history.

### 6.2 Match aggregate

Match owns:

- immutable match rules;
- assigned puzzle snapshot;
- match participants and seats;
- authoritative match lifecycle;
- match clocks and deadlines;
- one or more boards;
- notes;
- turn state;
- mistakes, hints, timeouts, and score;
- disconnect eligibility;
- finishing state;
- completion and result;
- durable event sequence;
- aggregate version.

### 6.3 Puzzle aggregate

Puzzle owns:

- canonical clues;
- canonical solution;
- difficulty;
- hardest required technique;
- quality score;
- multiplayer approval;
- generator and solver version;
- lifecycle;
- revision;
- retirement reason;
- canonical fingerprint.

### 6.4 DailyChallenge aggregate

DailyChallenge owns:

- UTC date;
- assigned puzzle and revision;
- transformation seed;
- announced difficulty;
- publication status;
- replacement status if the original puzzle is retired before publication.

Daily completion and streaks are device-local in V1 and are not owned by the server-side DailyChallenge aggregate.

---

## 7. Participant identity

### 7.1 Temporary identity

A participant:

- exists only within one Room;
- is not a global account;
- has a temporary display name;
- may be a player or spectator;
- may also hold host authority;
- retains attribution for historical match events even after leaving.

### 7.2 Display-name rules

A display name:

- contains 2–20 visible characters;
- is trimmed at both ends;
- may include Unicode letters, numbers, spaces, punctuation, and emoji;
- is stored in its visible original form;
- is normalized for duplicate detection;
- must be unique within the room using case-insensitive comparison where casing exists;
- must reject control characters;
- must reject bidirectional control-character abuse;
- must reject invisible-formatting abuse;
- must reject URLs;
- must reject excessive repeated-character or emoji-sequence abuse;
- may be rejected by a small configurable prohibited-slur list.

When a duplicate is submitted, the client may suggest a suffix such as `Alex 2`.

Display-name errors must not reveal filtering details. A safe user message is:

> Please choose a different display name.

### 7.3 Participation role and host authority

Participation role is one of:

- `Player`
- `Spectator`

Host authority is a separate boolean or authority reference. Therefore:

- a host may be a player;
- a host may be a spectator;
- transferring host status does not change participation role;
- converting between player and spectator does not inherently transfer host status.

### 7.4 Session continuity

Each participant has a secret reconnect credential. Domain behavior requires:

- a disconnected player retains the same participant identity during the applicable reconnect window;
- reconnecting restores their seat and attribution;
- historical events always retain the original participant identity;
- a removed or blocked participant cannot reuse the same room session.

---

## 8. Room configuration

A room configuration contains at least:

- mode;
- difficulty;
- error preset;
- hint policy;
- optional match time limit;
- Co-op note behavior;
- Race finishing-window duration;
- Duel turn duration;
- spectator policy;
- host-approval policy;
- reconnect policy;
- reaction/ping policy;
- room lock state.

### 8.1 Supported modes

- `Coop`
- `Race`
- `Duel`

Solo and Daily do not use multiplayer rooms.

### 8.2 Supported difficulties

General puzzle difficulties:

- `Easy`
- `Medium`
- `Hard`
- `Expert`

Mode constraints:

- Co-op: Easy, Medium, Hard, Expert
- Race: Easy, Medium, Hard, Expert
- Duel: Easy, Medium, Hard
- Solo: Easy, Medium, Hard, Expert, or Random
- Daily: assigned by schedule

### 8.3 Player capacities

- Co-op: maximum 6 players
- Race: maximum 8 players
- Duel: exactly 2 players

Minimum start counts:

- Co-op: 1 player; UI should encourage 2 or more
- Race: 2 players
- Duel: exactly 2 players

### 8.4 Spectator capacities

- Co-op: maximum 10 spectators
- Race: maximum 10 spectators
- Duel: maximum 20 spectators

Spectators never consume player seats.

### 8.5 Match time limits

Default multiplayer behavior has no overall time limit.

Optional host-selected limits:

- 5 minutes
- 10 minutes
- 15 minutes
- 20 minutes

A mode may impose additional timers, such as Duel turn duration or Race finishing window.

### 8.6 Error presets

The canonical incorrect-entry presets are:

#### Casual

- incorrect value is accepted into the board;
- incorrect value is visibly marked;
- no time penalty;
- correctness is immediately visible.

#### Challenge

- incorrect value is rejected;
- a 5-second penalty is added;
- incorrect value does not remain on the board.

#### Blind

- incorrect value is accepted into the board;
- correctness is not revealed during play;
- completion fails until all values are correct.

#### Clean

- incorrect value is rejected;
- no time penalty;
- correctness is revealed only through rejection.

Default casual behavior for friendly play is `Casual`.

Duel has its own mandatory wrong-entry behavior and does not use these presets for turn resolution.

### 8.7 Hint policy

Canonical hint availability:

- Solo: enabled
- Co-op: enabled
- Casual Race: host-configurable
- Competitive-style Race: disabled
- Duel: disabled

### 8.8 Notes policy

- Notes are supported in all modes.
- Co-op notes are shared.
- Race notes are private per player board.
- Duel notes are private per player.
- Solo notes are private.
- Placing a value clears notes in that same cell.
- Automatic candidate removal from peer cells is configurable and enabled by default.
- Notes may never be added to fixed clue cells.

---

## 9. Room lifecycle

### 9.1 States

```text
Open
ReadyCheck
Countdown
InMatch
Results
Expired
Cancelled
RecoveryPending
TerminatedByAdmin
```

### 9.2 Normal transitions

```text
Open → ReadyCheck
ReadyCheck → Countdown
Countdown → InMatch
InMatch → Results
Results → ReadyCheck        # rematch
Open → Expired
ReadyCheck → Expired
Results → Expired
```

### 9.3 Exceptional transitions

```text
Open → Cancelled
ReadyCheck → Cancelled
Countdown → Cancelled
InMatch → RecoveryPending
RecoveryPending → InMatch
RecoveryPending → Cancelled
Any nonterminal state → TerminatedByAdmin
```

### 9.4 Room invariants

At all times:

1. A room has at most one active host.
2. A room has at most one active Match reference.
3. Display names are unique within the room.
4. Players never exceed mode capacity.
5. Spectators never consume player seats.
6. Only the host may change gameplay settings.
7. A gameplay setting change resets every participant’s ready state.
8. Settings cannot change during Countdown or InMatch.
9. Only ready players are included when a Match is created.
10. Race cannot start with fewer than two players.
11. Duel cannot start with any count other than two players.
12. A room code is not publicly searchable.
13. A locked room rejects all new participants and spectators but permits valid reconnects.
14. Historical participant attribution is never rewritten after host transfer, leave, or removal.

### 9.5 Room expiration

- Empty room: expires after 30 minutes.
- Lobby with participants: expires after 2 hours without activity.
- Active match: does not expire from room inactivity.
- Results room: remains available for rematches for 2 hours.
- Replay retention is independent and lasts 7 days.

### 9.6 Room activity

Activity includes meaningful room operations such as:

- participant joins;
- participant leaves;
- readiness changes;
- gameplay setting changes;
- host transfer;
- rematch initiation;
- active match commands.

Pure presence heartbeats do not necessarily extend room lifetime unless architecture explicitly defines them as meaningful activity.

---

## 10. Room joining and access

### 10.1 Access model

A room has no public directory. A person may request access only when possessing:

- the room code; or
- an invitation URL containing the room code.

There is no room password in V1.

### 10.2 Room preview

Before joining, a visitor may see only:

- room mode;
- difficulty;
- room state;
- available player seats;
- spectator availability;
- whether host approval is required;
- whether the room is locked.

Participant names are not exposed before a successful join.

### 10.3 Join approval

Host approval is optional.

When disabled:

```text
JoinRequested → ParticipantJoined
```

When enabled:

```text
JoinRequested → ApprovalPending
ApprovalPending → ParticipantJoined
ApprovalPending → JoinRejected
```

### 10.4 Locked room

When locked:

- valid reconnects are allowed;
- new player joins are rejected;
- new spectator joins are rejected;
- host may unlock the room.

Duel rooms lock automatically when the match begins.

### 10.5 Late joining

After a match begins:

- Co-op: new visitors may join only as spectators.
- Race: new visitors may join only as spectators.
- Duel: new visitors may join only as delayed spectators if spectator policy allows.
- Existing disconnected players may reconnect and regain their original role.

### 10.6 Player-to-spectator conversion

- Lobby: may switch freely if seat limits permit.
- Co-op: player may voluntarily become spectator.
- Race: player may become spectator only after finishing or abandoning.
- Duel: leaving the player role is resignation.

---

## 11. Host authority

The host may:

- choose mode;
- choose difficulty;
- choose timer;
- choose error preset;
- choose hint policy;
- choose spectator policy;
- enable or disable host approval;
- enable or disable reactions;
- lock or unlock the room;
- remove a participant;
- block a participant from the room;
- start the match;
- cancel Countdown;
- create a rematch;
- transfer host status.

### 11.1 Host transfer

Manual transfer:

- host selects another active participant;
- authority changes immediately;
- gameplay role remains unchanged.

Automatic transfer after host disconnect:

1. Reserve host authority for 60 seconds.
2. If host reconnects in time, they retain authority.
3. Otherwise transfer authority to the longest-present active participant.
4. Host transfer never changes an active Match’s rules or state.

### 11.2 Removing participants

Before match start:

- host may remove a participant immediately;
- removed participant’s room session becomes invalid;
- participant may not immediately rejoin with the same session;
- participant may request a fresh join after 5 minutes unless blocked.

During a match:

- normal behavior is `RemoveAfterMatch`;
- immediate removal is reserved for abuse or severe technical disruption;
- immediate removal must apply mode-specific abandonment, resignation, or spectator-conversion rules.

A host may block a participant from rejoining that room.

---

## 12. Ready check and countdown

### 12.1 Readiness

- Each player controls their own ready state.
- Spectators do not need to become ready.
- The host cannot mark another player ready.
- A gameplay-setting change resets all player ready states.
- Cosmetic changes do not reset readiness.

### 12.2 Start authorization

The host may begin Countdown only when:

- minimum mode player count is met;
- every included player is ready;
- no required approval is pending;
- room is not locked against existing participants in a way that invalidates current seats;
- no other active Match exists.

### 12.3 Countdown

Default countdown duration is 3 seconds.

On Countdown start:

- room settings become locked;
- included player seats are fixed;
- Match rules are copied immutably;
- puzzle assignment is fixed;
- server determines the countdown deadline.

Host may cancel Countdown before Match activation. Cancellation returns to ReadyCheck and preserves room configuration but resets readiness as appropriate.

---

## 13. Match base model

### 13.1 States

```text
Prepared
Countdown
Active
Finishing
Completed
RecoveryPending
Cancelled
Abandoned
Invalidated
```

### 13.2 Normal transitions

```text
Prepared → Countdown
Countdown → Active
Active → Finishing
Finishing → Completed
Active → Completed          # modes without finishing window
```

### 13.3 Exceptional transitions

```text
Prepared → Cancelled
Countdown → Cancelled
Active → RecoveryPending
Finishing → RecoveryPending
RecoveryPending → Active
RecoveryPending → Finishing
RecoveryPending → Cancelled
Active → Abandoned
Completed → Invalidated
```

### 13.4 Match invariants

1. Match rules never change after Match creation.
2. Assigned puzzle revision and transformation never change.
3. Fixed clues never change.
4. Every durable event has exactly one event number.
5. Event numbers strictly increase.
6. Aggregate versions strictly increase with committed mutations.
7. The same `RequestID` cannot produce two domain effects.
8. Only eligible players may issue gameplay commands.
9. Spectators cannot mutate gameplay state.
10. Completion is decided only by authoritative verification.
11. A completed Match cannot return to Active.
12. A result is finalized once.
13. Administrative invalidation does not erase original result history.
14. A rematch is a new Match with a new identifier.

### 13.5 Immutable MatchRules

A MatchRules value contains:

- mode;
- difficulty;
- error preset;
- hint policy;
- overall time limit;
- Duel turn duration;
- Race finishing-window duration;
- spectator policy;
- reconnect policy;
- note auto-removal setting;
- any rule version needed for historical reconstruction.

---

## 14. Puzzle assignment inside a Match

A Match stores an immutable assigned-puzzle snapshot:

- canonical puzzle ID;
- puzzle revision;
- transformation seed;
- transformed clues;
- difficulty at assignment;
- multiplayer approval status at assignment;
- generator version;
- solver version.

The authoritative server retains access to the transformed solution. Multiplayer clients do not receive the complete solution.

Historical replay remains reproducible even if the Puzzle aggregate is later retired or revised.

---

## 15. Board rules

### 15.1 Board representation

A board has exactly 81 cells.

Each cell may contain:

- fixed clue;
- mutable value;
- no value;
- zero or more notes;
- player attribution where applicable;
- correctness visibility state where permitted.

### 15.2 Fixed clues

A fixed clue:

- cannot be erased;
- cannot be overwritten;
- cannot receive notes;
- is identical to the assigned solution;
- remains immutable during recovery and replay.

Any command targeting a fixed clue is rejected with `CELL_FIXED`.

### 15.3 Direct Sudoku conflicts

A board has a direct Sudoku conflict when the same non-empty digit appears more than once in:

- one row;
- one column;
- one 3×3 box.

Direct conflicts may be shown independently from final-solution correctness.

### 15.4 Value placement

A value placement command must validate:

- Match state permits action;
- actor is eligible;
- cell exists;
- cell is not fixed;
- value is 1–9;
- actor has permission for the mode;
- turn ownership where applicable;
- expected aggregate version;
- duplicate request protection;
- error preset behavior.

### 15.5 Erasing

Erasing:

- is prohibited on clues;
- is allowed on mutable values subject to mode permissions;
- retains historical attribution in replay;
- may clear correctness and progress calculations;
- does not remove the original event history.

### 15.6 Correctness visibility

The server always knows correctness, but clients learn only what the MatchRules permit.

- `Casual`: wrong value remains and is marked.
- `Challenge`: wrong value is rejected and penalty applied.
- `Blind`: wrong value remains but is not marked.
- `Clean`: wrong value is rejected without penalty.
- Duel: wrong value is rejected and the turn is lost.

Replay output must preserve the original visibility semantics. A replay must not reveal hidden information earlier than the point at which match rules allow it.

---

## 16. Notes

### 16.1 General note rules

- Notes contain digits 1–9.
- Duplicate candidate entries are ignored idempotently.
- Notes may not exist in fixed cells.
- Placing a value clears notes in that cell.
- Erasing a value does not automatically restore prior notes.
- Auto-removal of matching candidates from peer cells is enabled by default and controlled by immutable MatchRules.
- Note operations never claim puzzle correctness.

### 16.2 Ownership

- Co-op: one shared notes layer.
- Race: one notes layer per player board.
- Duel: one private notes layer per player.
- Solo: one private notes layer.

### 16.3 Duel notes

Duel note changes:

- do not consume a turn;
- are not visible to the opponent;
- are not included in opponent-facing live events;
- may appear in the owning player’s private replay view if that view is supported.

---

## 17. Hints

### 17.1 Hint levels

The final hint model has three levels:

1. `Nudge`
   - highlights a useful row, column, box, or cell;
   - does not reveal a value.

2. `Explain`
   - identifies a valid logical technique;
   - identifies affected cells;
   - provides a localization message key and structured values;
   - must be generated from deterministic solver reasoning, not invented after inspecting the final solution.

3. `Reveal`
   - places or reveals the correct value in one eligible cell.

The same situation may escalate through these levels.

### 17.2 Availability

- Solo: enabled.
- Co-op: enabled.
- Race: host-configurable for casual rooms; disabled in competitive-style rules.
- Duel: disabled.

### 17.3 Penalties and result flags

- Solo hint: adds a visible 20-second result penalty.
- Race hint: adds a 20-second score/time penalty and marks the player as `Assisted`.
- Co-op hint: recorded in shared match statistics; no individual victory penalty because Co-op is collaborative.
- Hint use is always replay-visible.

---

## 18. Co-op mode

### 18.1 Core model

Co-op uses:

- one shared board;
- one shared notes layer;
- 1–6 players;
- optional spectators;
- automatic completion;
- shared success.

There is no individual winner. Contribution statistics are informational only.

### 18.2 Simultaneous editing

Several players may interact concurrently, but authoritative mutations are serialized.

A player selecting an editable cell creates a soft lock:

- duration: 8 seconds;
- refreshed while active interaction continues;
- released after value entry, erase, deselection, disconnect, or expiration;
- another player may override after a warning;
- soft lock is ephemeral;
- soft lock does not make the cell permanently unavailable.

### 18.3 Attribution

Each accepted value retains attribution to the participant who placed it.

Attribution must be representable using:

- participant identifier;
- display-name snapshot;
- player marker;
- accessible non-color indicator.

Attribution persists in replay and result summaries even if the participant disconnects or leaves.

### 18.4 Shared notes

Every Co-op note mutation is immediately shared with connected players after authoritative acceptance.

A player may erase any non-clue value subject to current rules.

Undo behavior:

- a player may undo only their own latest reversible action;
- any player may directly erase a non-fixed value;
- undo does not erase history; it emits inverse events.

### 18.5 Wrong entries

Co-op uses the room’s selected error preset.

Default: `Casual`.

### 18.6 Hints

Hints are allowed. Hint events are shared and replay-visible.

### 18.7 Pings and reactions

Co-op supports lightweight communication without text chat.

Canonical ping/reaction intents:

- `look_here`
- `agree`
- `unsure`
- `try_this_area`
- `nice_move`

Properties:

- structured intent, never pretranslated sentence;
- may target a cell or region;
- expires visually after several seconds;
- rate-limited;
- may be disabled by host;
- may be muted per participant;
- included in replay as lightweight timeline events when replay policy enables them.

### 18.8 Joining after start

A new visitor joining after Match activation becomes a spectator.

A player who participated before disconnecting may reconnect and regain editing rights during the reconnect window.

### 18.9 Disconnect

- Match continues.
- Participant’s seat is retained for 5 minutes.
- Their accepted contributions remain.
- Reopening the room with valid session restores identity.
- After 5 minutes, host may release the seat.
- Released player may not regain the old seat unless re-admitted as a new participant.

### 18.10 Completion

Co-op completes automatically when the authoritative shared board:

- contains 81 values;
- matches the assigned solution;
- has no unresolved validation failure.

Completion produces:

- shared success result;
- total elapsed time;
- contribution counts;
- mistake counts;
- hint counts;
- disconnect summary;
- replay availability.

---

## 19. Race mode

### 19.1 Core model

Race uses:

- the same assigned puzzle for every player;
- one independent board per player;
- private notes per player;
- 2–8 players;
- optional spectators;
- one authoritative first finisher;
- a finishing window for remaining players.

### 19.2 Start

Race start sequence:

```text
Lobby → ReadyCheck → 3-second Countdown → Active
```

Players not ready when Countdown starts are not included.

### 19.3 Opponent visibility

During Race, players may see:

- approximate board heatmap;
- progress based only on correct cells;
- no opponent values;
- no opponent private notes;
- no direct view of opponent wrong values.

Heatmap behavior:

- updates are delayed by 2 seconds;
- updates may be batched;
- progress must have a text equivalent;
- heatmap must not reveal exact digits.

### 19.4 Progress calculation

Race progress equals the number of correctly filled non-clue cells.

Incorrect values do not increase progress.

Erasing a correct value decreases progress.

### 19.5 Late join

After Race activation, a new visitor may join only as spectator.

### 19.6 Winner

The first player whose complete board is authoritatively verified becomes the Race winner.

The determining time is the server timestamp of the accepted final move.

Client completion time is irrelevant.

### 19.7 Finishing window

After the first verified finish:

- winner is declared immediately;
- Match enters `Finishing`;
- remaining players have 60 seconds to finish;
- no new players enter;
- result ranking is finalized when all remaining players finish or the window expires.

### 19.8 Final ranking

Ranking order:

1. Verified finishers by server completion timestamp.
2. Unfinished players by number of correctly completed cells.
3. Fewer incorrect attempts.
4. Earlier timestamp of the player’s last accepted correct move.

If a perfect tie remains after these criteria, tied placement is allowed.

### 19.9 Error presets

Race supports:

- Casual
- Challenge
- Blind
- Clean

Competitive-style Race disables hints. Casual Race may enable them.

### 19.10 Hints

When enabled:

- one useful cell may be revealed;
- 20-second penalty is added;
- player is marked `Assisted`;
- result and replay disclose assistance.

### 19.11 Erasing

A player may erase any non-clue value on their own board.

The server must not expose correctness merely by blocking erase.

### 19.12 Disconnect

- Match and timer continue.
- Player seat is retained for 5 minutes.
- Reconnect restores current board.
- A disconnected player cannot become newly verified as complete without reconnecting and submitting completion through an accepted command.
- If Match ends before return, ranking uses last synchronized state.

### 19.13 Spectators

Spectators may observe delayed progress subject to room policy. They cannot mutate boards, send player commands, or access private notes.

### 19.14 Replay

Race replay:

- may show all player boards side by side;
- uses one synchronized timeline;
- preserves original visibility timing;
- includes finish order, mistakes, hints, and disconnects;
- is private to people with replay capability;
- cannot be hidden by an individual participant in V1;
- expires after 7 days unless deleted sooner.

---

## 20. Duel mode

### 20.1 Core model

Duel uses:

- exactly 2 players;
- one shared value board;
- private notes per player;
- alternating turns;
- score-based victory;
- optional delayed spectators.

### 20.2 Difficulty

Duel supports:

- Easy
- Medium
- Hard

Expert is excluded.

### 20.3 Starting player

- First Duel: starting player is random.
- Rematch: starting player alternates from the prior Duel between the same room participants.

### 20.4 Turn duration

Host-selectable values:

- 15 seconds
- 30 seconds
- 60 seconds
- Unlimited

Default: 30 seconds.

### 20.5 Turn ownership

Only the active player may:

- place a value;
- erase a mutable value;
- pass where supported;
- resign.

Private note operations are permitted outside turn consumption but only for the owning player.

### 20.6 Correct placement

A correct placement:

- is committed to the shared board;
- awards base and completion bonuses;
- ends the player’s turn;
- passes turn to the opponent.

A correct placement never grants an extra turn.

### 20.7 Incorrect placement

An incorrect placement:

- is rejected;
- does not remain on the board;
- loses the player’s turn;
- increments mistake count;
- does not reveal the attempted number to the opponent;
- is replay-visible as a rejected attempt without revealing the secret attempted digit to unauthorized viewers.

### 20.8 Erasing

Erasing a mutable shared value:

- is allowed only for the active player;
- cannot target clues;
- is authoritative;
- ends the current turn;
- does not award points;
- remains replay-visible.

### 20.9 Timeout

When turn time expires:

- active player loses the turn;
- timeout count increments;
- no additional point penalty;
- turn passes to opponent.

Three consecutive timeouts by the same player constitute abandonment. Opponent wins.

A correct move, incorrect attempt, or other turn-ending action resets the consecutive-timeout count.

### 20.10 Scoring

- Correct placement: 1 point
- Completing a row: +1
- Completing a column: +1
- Completing a box: +1
- Completing the puzzle: +3

One move may earn multiple bonuses.

Completion bonuses are based on the board transition caused by that accepted placement. A structure already complete before the move cannot be scored again.

### 20.11 Winner

When the puzzle completes, higher score wins.

Tie-breakers:

1. Fewer incorrect attempts
2. Fewer turn timeouts
3. Lower total thinking time
4. Shared victory if still tied

Resignation, disconnect forfeit, or abandonment awards victory to the opponent regardless of current score.

### 20.12 Cell consideration visibility

Opponent does not see the active player’s selected cell before move acceptance.

Clients may show only a generic `thinking` indicator.

### 20.13 Notes

- Notes are private.
- Notes never consume a turn.
- Notes are never visible to opponent.
- Notes cannot intentionally mislead the opponent because they are not shared.

### 20.14 Resignation

A player may resign at any time.

Resignation:

- ends Match immediately;
- awards opponent victory;
- records resignation event and result reason;
- permits replay.

### 20.15 Disconnect

Each player receives one protected disconnect pause per Duel:

- reconnect window: 60 seconds;
- active turn timer pauses during this protected window;
- if player returns, play resumes;
- if player does not return, player forfeits;
- after protected pause has been used, later disconnections do not pause the turn timer.

### 20.16 Spectators

- maximum 20;
- view delayed by 10 seconds;
- cannot ping or react until Duel ends;
- cannot see private notes;
- cannot see hidden rejected digits.

---

## 21. Solo mode

### 21.1 Entry points

Solo supports:

- Easy
- Medium
- Hard
- Expert
- Random
- Continue Last Puzzle when local unfinished progress exists

### 21.2 Play styles

#### Guided

- incorrect entries are highlighted immediately;
- full Check Board is available;
- direct conflicts are visible.

#### Classic

- incorrect entries are not marked solely because they differ from solution;
- direct Sudoku conflicts may still be shown;
- full-solution correctness remains hidden until completion.

Default for first-time players: Guided.

The device remembers the last choice.

### 21.3 Timer

- starts on first interaction with an editable cell;
- may be paused;
- closing the tab saves state;
- time while closed does not count;
- hint penalties are shown separately and included in result time.

### 21.4 Hints

Solo supports Nudge, Explain, and Reveal.

Each hint:

- is recorded in local attempt history;
- adds 20 seconds to result penalty;
- must be generated from deterministic solver reasoning.

### 21.5 Check Board

- available in Guided mode;
- highlights direct conflicts and incorrect values;
- does not fix values;
- in Classic mode, only direct Sudoku conflicts may be shown.

### 21.6 Completion

Solo completes when:

- all 81 cells contain values;
- values match authoritative or locally permitted validation data;
- completion is explicitly or automatically verified.

### 21.7 Local persistence

Device-local data includes:

- unfinished attempt;
- best times by difficulty;
- completed puzzle count;
- hint count;
- mistake count;
- preferences;
- recent puzzle IDs;
- solo replay;
- Daily progress;
- multiplayer result summaries.

The user must be informed that clearing browser storage may delete this data.

### 21.8 Puzzle repetition

The client retains recently encountered puzzle identities and requests exclusions.

A puzzle should not repeat until the available catalog has rotated sufficiently.

Transformations of the same canonical puzzle count as the same puzzle for repetition prevention.

### 21.9 Offline classification

Two conceptual Solo classifications exist:

- `StandardOnlineSolo`: server may validate hints and completion.
- `OfflineSolo`: local validation package may be used.

Offline results are never eligible for a trusted future leaderboard because browser-side solution protection cannot be authoritative.

The current release scope determines whether OfflineSolo is enabled.

---

## 22. Daily Ninefold

### 22.1 Assignment

- one shared puzzle per UTC date;
- changes at `00:00 UTC`;
- same canonical puzzle and transformation for all visitors that date;
- countdown to next puzzle may be displayed.

### 22.2 Difficulty schedule

| Day | Difficulty |
|---|---|
| Monday | Easy |
| Tuesday | Medium |
| Wednesday | Medium |
| Thursday | Hard |
| Friday | Hard |
| Saturday | Expert |
| Sunday | Mixed surprise |

### 22.3 Progress and streak

In V1:

- Daily completion is device-local;
- streak is device-local;
- no account or global leaderboard;
- replay is device-local;
- replaying a completed Daily does not replace the first official local result unless product rules explicitly allow it.

### 22.4 Replacement

If a scheduled Puzzle is retired before its publication date, DailyChallenge may replace it.

If a Puzzle is retired after publication, historical attempts remain associated with the original assigned revision.

---

## 23. Puzzle domain

### 23.1 Variant scope

Initial canonical variant: classic 9×9 Sudoku.

The domain may later support additional variants, but no non-classic variant may be introduced without explicit rules for:

- clue model;
- constraints;
- solver;
- uniqueness;
- difficulty grading;
- hints;
- multiplayer fairness;
- replay.

### 23.2 Uniqueness

Every published Puzzle must have exactly one valid solution.

### 23.3 Generation pipeline

Canonical generation process:

1. Generate a complete valid grid.
2. Remove clues incrementally.
3. Verify exactly one solution remains.
4. Grade required logical techniques.
5. Evaluate quality and multiplayer fairness.
6. Reject puzzles that fail criteria.
7. Store accepted Puzzle with generator and solver versions.

Generation occurs ahead of gameplay, not in the player request path.

### 23.4 Difficulty grading

Difficulty is determined by deterministic logical-solver requirements, not clue count alone.

#### Easy

- naked singles
- hidden singles

#### Medium

- locked candidates
- naked pairs
- hidden pairs

#### Hard

- triples
- X-Wing
- more complex candidate interactions

#### Expert

- advanced logical techniques
- no guessing required

Brute-force search may prove uniqueness but must not be used to label a puzzle logically solvable if the logical solver cannot complete it.

### 23.5 Multiplayer quality

A Puzzle may be marked `multiplayer_approved`.

Race and Duel use only multiplayer-approved puzzles.

A Puzzle should be rejected for multiplayer when it has:

- excessively slow opening;
- excessive forced-single chain;
- highly uneven logical paths;
- unusually unstable expected solve time;
- ambiguous grading;
- other fairness concerns identified by quality analysis.

Co-op and Solo may use the broader active catalog.

### 23.6 Canonical transformations

Allowed transformations:

- digit renaming;
- row swaps within a band;
- column swaps within a stack;
- band swaps;
- stack swaps;
- rotation;
- reflection.

Transformation must preserve:

- validity;
- uniqueness;
- difficulty class;
- canonical identity;
- replay reproducibility.

A Match stores the transformation seed.

### 23.7 Reuse rules

A rematch must not use:

- the same Puzzle;
- any transformation of the same canonical Puzzle;
- any of the previous 20 Puzzles used in that Room.

### 23.8 Puzzle lifecycle

```text
Draft → Verified → Active → Retired
```

#### Draft

Not eligible for gameplay.

#### Verified

Passed uniqueness and solver checks but not yet active.

#### Active

Eligible for assignment subject to mode approval.

#### Retired

Not eligible for new assignment.

Retirement:

- records reason and timestamp;
- does not delete historical Match or Replay references;
- does not rewrite past results.

---

## 24. Match replay

### 24.1 Purpose

Replay serves:

- post-match understanding;
- entertainment;
- debugging;
- deterministic recovery validation;
- portfolio demonstration;
- integrity verification.

### 24.2 Source of truth

Replay is reconstructed from:

- immutable assigned-puzzle snapshot;
- immutable MatchRules;
- ordered durable event log;
- optional periodic snapshots for fast seeking.

Snapshots are optimizations. Events remain authoritative.

### 24.3 Replay capabilities

Replay supports:

- play;
- pause;
- timeline scrubbing;
- speeds 0.5×, 1×, 2×, 4×;
- event markers;
- player progress;
- mode-specific presentation;
- completion moment;
- disconnects and reconnects;
- mistakes and hints where visibility permits.

Mode-specific views:

- Co-op: attributed shared board.
- Race: synchronized player boards.
- Duel: turns, score changes, time usage, and shared board.
- Solo: local timeline.

### 24.4 Replay event scope

Replay-relevant durable events include:

- match lifecycle transitions;
- value placement;
- value erase;
- note changes;
- hint use;
- Race completion;
- Duel turn changes;
- Duel timeout;
- scoring;
- resignation;
- abandonment;
- join/disconnect/reconnect where relevant;
- pings/reactions when configured for replay.

Raw pointer motion and keyboard events are never recorded.

Cell selection may be included in an extended replay mode, but:

- it is not authoritative state;
- it is not required for recovery;
- it is not required in the MVP replay;
- transient remote focus is ephemeral by default.

### 24.5 Replay privacy

- Multiplayer replay requires an unguessable replay capability.
- Replay is not publicly indexed.
- Only people with the capability may view it.
- Replay expires after 7 days.
- Any authenticated room participant may request immediate deletion.
- One deletion removes the shared replay for everyone.
- Solo replay remains device-local.
- Replay never exposes session credentials.
- Replay never exposes hidden Duel notes.
- Replay respects original correctness visibility.

### 24.6 Result card

A shareable result summary may include:

- Ninefold Sudoku branding;
- mode;
- difficulty;
- elapsed time;
- player count;
- mistake count;
- hint or assistance status;
- temporary display names where appropriate.

It must not expose:

- puzzle solution;
- private notes;
- session tokens;
- hidden opponent data;
- unrelated participant data.

---

## 25. Cryptographically verifiable replay

### 25.1 Hash chain

Every durable event participates in a SHA-256 hash chain.

Conceptually:

```text
event_hash = SHA-256(
    match_id
    + event_number
    + event_type
    + canonical_payload
    + previous_event_hash
)
```

The first event uses a defined genesis value.

Canonicalization rules must be stable and versioned.

### 25.2 Match seal

At Match completion, the final event hash is sealed with Ed25519.

A Match proof contains at least:

- Match ID;
- final event number;
- final event hash;
- completion timestamp;
- signing key ID;
- Ed25519 signature;
- proof format version.

### 25.3 Browser verification

A replay verifier must be able to confirm:

1. Event numbers are sequential.
2. Event payload canonicalization is valid.
3. Each event hash links to the prior hash.
4. No event was inserted, removed, or changed after sealing.
5. Final hash matches the signed digest.
6. Signature matches a trusted Ninefold public key.

### 25.4 Meaning of verification

A valid proof establishes that the replay has not been altered after the authoritative server sealed it.

It does not independently prove that the server’s rule implementation was correct at runtime. That assurance is strengthened through:

- open source implementation;
- deterministic reconstruction;
- domain tests;
- property-based tests;
- TLA+ model checking.

---

## 26. Results

### 26.1 MatchResult

A finalized result contains:

- Match ID;
- mode;
- result reason;
- participant snapshots;
- rankings or shared success;
- completion timestamps;
- elapsed time;
- mistakes;
- hints;
- assistance flags;
- Duel scores;
- timeouts;
- disconnect outcomes;
- replay availability;
- invalidation metadata where applicable.

### 26.2 Result reasons

Canonical reasons include:

- `PuzzleCompleted`
- `FinishingWindowExpired`
- `Resignation`
- `DisconnectForfeit`
- `ConsecutiveTimeoutAbandonment`
- `HostTermination`
- `AdministrativeTermination`
- `RecoveryFailure`
- `CancelledBeforeStart`

### 26.3 Invalidation

An administrator may invalidate a completed result.

Invalidation:

- does not delete original result;
- adds invalidation timestamp;
- adds administrator identity supplied by trusted infrastructure;
- adds reason;
- keeps replay and audit history until retention deletion;
- prevents the result from being treated as valid in future statistics or rankings.

---

## 27. Rematch

A rematch:

- occurs within the same Room;
- creates a new Match ID;
- selects a different canonical Puzzle;
- copies current Room settings into new immutable MatchRules;
- retains current participants;
- returns Room to ReadyCheck;
- resets ready states;
- resets boards, scores, penalties, and timers;
- increments Room rematch number;
- alternates Duel starting player;
- does not mutate or erase prior Match history.

---

## 28. Disconnect, reconnect, and recovery

### 28.1 General disconnect

Disconnect is not the same as voluntary leave.

A network interruption:

- marks participant disconnected;
- preserves identity;
- applies mode-specific grace period;
- does not erase accepted actions;
- does not imply host transfer until host grace expires;
- does not permit queued offline gameplay commands.

### 28.2 General reconnect

A valid reconnect:

- restores same ParticipantID;
- restores role where still eligible;
- restores authoritative room and match state;
- resumes from missing events or snapshot;
- rejects commands produced while disconnected if they are stale.

### 28.3 Mode windows

- Co-op: 5 minutes
- Race: 5 minutes
- Duel: one protected 60-second pause per player

### 28.4 Server restart recovery

After server interruption:

1. Reconstruct Room and Match from latest valid snapshot plus later events.
2. Mark active state `RecoveryPending`.
3. Give participants up to 5 minutes to reconnect.
4. Resume Co-op and Race.
5. Resume Duel at the beginning of the interrupted turn.
6. Cancel if nobody reconnects before recovery deadline.
7. Preserve original event order and aggregate versions.

### 28.5 Recovery invariants

- No accepted event may be applied twice.
- No committed event may be omitted.
- Stale timers may not affect reconstructed state.
- A completed Match may not resume.
- A recovered Match must pass all aggregate invariants before accepting commands.

---

## 29. Timers

Authoritative timers include:

- room expiration;
- ready countdown;
- match time limit;
- Duel turn deadline;
- Race finishing window;
- reconnect grace;
- recovery deadline.

Timer events must include generation identity so stale timers cannot affect newer state.

A timer firing produces a domain command. It never mutates state directly.

Server timestamps determine all outcomes.

---

## 30. Commands

The following catalog defines domain intent. Transport names may differ but must map one-to-one.

### 30.1 Room commands

- `CreateRoom`
- `RequestJoin`
- `ApproveJoin`
- `RejectJoin`
- `ResumeRoomSession`
- `LeaveRoom`
- `ChangeParticipationRole`
- `SetReady`
- `ChangeRoomSettings`
- `LockRoom`
- `UnlockRoom`
- `MuteParticipant`
- `UnmuteParticipant`
- `RemoveParticipant`
- `BlockParticipant`
- `TransferHost`
- `StartCountdown`
- `CancelCountdown`
- `StartMatch`
- `CreateRematch`
- `ExpireRoom`
- `TerminateRoom`

### 30.2 Common Match commands

- `ActivateMatch`
- `PlaceValue`
- `EraseValue`
- `AddNote`
- `RemoveNote`
- `UseHint`
- `ParticipantDisconnected`
- `ParticipantReconnected`
- `MatchDeadlineReached`
- `EnterRecovery`
- `ResumeRecoveredMatch`
- `CancelRecoveredMatch`
- `InvalidateResult`

### 30.3 Co-op commands

- `FocusCell`
- `ReleaseCellFocus`
- `OverrideSoftLock`
- `SendPing`
- `SendReaction`
- `UndoOwnLatestAction`
- `ConvertCoopPlayerToSpectator`

### 30.4 Race commands

- `SubmitRaceCompletion`
- `RaceFinishingWindowElapsed`
- `AbandonRace`

### 30.5 Duel commands

- `PlaceDuelValue`
- `EraseDuelValue`
- `DuelTurnDeadlineReached`
- `PassDuelTurn`
- `ResignDuel`
- `DuelReconnectDeadlineReached`

### 30.6 Puzzle commands

- `GeneratePuzzle`
- `VerifyPuzzle`
- `ActivatePuzzle`
- `RetirePuzzle`
- `AssignPuzzle`
- `ReplaceScheduledDailyPuzzle`

### 30.7 Replay commands

- `CreateReplayCapability`
- `DeleteReplay`
- `SealReplay`
- `VerifyReplay`

---

## 31. Domain events

### 31.1 Room events

- `RoomCreated`
- `JoinRequested`
- `ParticipantJoined`
- `ParticipantJoinApproved`
- `ParticipantJoinRejected`
- `ParticipantLeft`
- `ParticipantRemoved`
- `ParticipantBlocked`
- `ParticipantRoleChanged`
- `ParticipantReadyStateChanged`
- `RoomSettingsChanged`
- `RoomReadyStatesReset`
- `RoomLocked`
- `RoomUnlocked`
- `HostTransferred`
- `CountdownStarted`
- `CountdownCancelled`
- `RoomEnteredMatch`
- `RoomEnteredResults`
- `RematchPrepared`
- `RoomExpired`
- `RoomCancelled`
- `RoomTerminatedByAdmin`

### 31.2 Common Match events

- `MatchPrepared`
- `MatchCountdownStarted`
- `MatchStarted`
- `ValuePlaced`
- `ValueRejected`
- `ValueErased`
- `NoteAdded`
- `NoteRemoved`
- `NotesAutoRemoved`
- `HintUsed`
- `ParticipantDisconnected`
- `ParticipantReconnected`
- `MatchEnteredRecovery`
- `MatchRecovered`
- `MatchCancelled`
- `MatchAbandoned`
- `MatchCompleted`
- `MatchResultInvalidated`

### 31.3 Co-op events

- `CellSoftLocked`
- `CellSoftLockRefreshed`
- `CellSoftLockReleased`
- `CellSoftLockOverridden`
- `CoopPingSent`
- `CoopReactionSent`
- `CoopContributionRecorded`
- `CoopPuzzleCompleted`

### 31.4 Race events

- `RaceProgressChanged`
- `RacePlayerFinished`
- `RaceWinnerDeclared`
- `RaceFinishingWindowStarted`
- `RaceFinishingWindowExpired`
- `RacePlayerAbandoned`
- `RaceRankingFinalized`

### 31.5 Duel events

- `DuelStartingPlayerSelected`
- `DuelTurnStarted`
- `DuelCorrectMoveAccepted`
- `DuelIncorrectMoveRejected`
- `DuelValueErased`
- `DuelScoreChanged`
- `DuelStructureCompleted`
- `DuelTurnTimedOut`
- `DuelTurnPassed`
- `DuelPlayerResigned`
- `DuelPlayerForfeited`
- `DuelAbandonedByTimeouts`
- `DuelCompleted`

### 31.6 Puzzle events

- `PuzzleGenerated`
- `PuzzleVerified`
- `PuzzleActivated`
- `PuzzleAssigned`
- `PuzzleRetired`
- `DailyPuzzleScheduled`
- `DailyPuzzleReplaced`

### 31.7 Replay events

- `ReplayCapabilityCreated`
- `ReplaySealed`
- `ReplayVerified`
- `ReplayDeleted`
- `ReplayExpired`

---

## 32. Durable versus ephemeral events

### 32.1 Durable

Persist when required for authoritative reconstruction or replay:

- Match lifecycle changes
- accepted value mutations
- rejected attempts that affect score, penalty, turn, or replay
- note changes
- hints
- timeouts
- scoring
- finish events
- resignation
- abandonment
- disconnect and reconnect when gameplay-relevant
- replay seal
- result invalidation

### 32.2 Ephemeral

Normally not persisted:

- connection heartbeat
- ordinary remote focus
- temporary hover
- typing indicators
- animation triggers
- transient presence refresh
- soft-lock refreshes that do not affect authoritative state history

A ping or reaction may be persisted when configured as replay content.

---

## 33. Domain error codes

Errors are machine-readable and transport-independent.

### 33.1 Room and identity

- `ROOM_NOT_FOUND`
- `ROOM_EXPIRED`
- `ROOM_CANCELLED`
- `ROOM_TERMINATED`
- `ROOM_LOCKED`
- `ROOM_FULL`
- `SPECTATOR_CAPACITY_REACHED`
- `SESSION_INVALID`
- `SESSION_EXPIRED`
- `NAME_INVALID`
- `NAME_ALREADY_USED`
- `JOIN_APPROVAL_REQUIRED`
- `JOIN_REQUEST_NOT_FOUND`
- `PARTICIPANT_BLOCKED`
- `PARTICIPANT_NOT_FOUND`
- `NOT_ROOM_HOST`
- `HOST_TRANSFER_INVALID`
- `ROLE_CHANGE_INVALID`

### 33.2 Readiness and lifecycle

- `PLAYER_NOT_READY`
- `PLAYERS_NOT_READY`
- `INSUFFICIENT_PLAYERS`
- `INVALID_PLAYER_COUNT`
- `COUNTDOWN_ALREADY_STARTED`
- `COUNTDOWN_NOT_ACTIVE`
- `MATCH_ALREADY_STARTED`
- `MATCH_ALREADY_EXISTS`
- `MATCH_NOT_ACTIVE`
- `MATCH_ALREADY_COMPLETED`
- `ROOM_STATE_INVALID`
- `MATCH_STATE_INVALID`
- `SETTINGS_LOCKED`

### 33.3 Board and gameplay

- `CELL_INDEX_INVALID`
- `DIGIT_INVALID`
- `CELL_FIXED`
- `CELL_SOFT_LOCKED`
- `CELL_NOT_EDITABLE`
- `INVALID_VALUE`
- `VALUE_REJECTED_BY_RULES`
- `NOTE_INVALID`
- `HINTS_DISABLED`
- `HINT_LEVEL_UNAVAILABLE`
- `NOT_YOUR_TURN`
- `ACTION_NOT_ALLOWED_FOR_ROLE`
- `ACTION_NOT_ALLOWED_IN_MODE`
- `UNDO_NOT_AVAILABLE`
- `REPLAY_NOT_AVAILABLE`

### 33.4 Concurrency and recovery

- `STALE_VERSION`
- `DUPLICATE_REQUEST`
- `CLIENT_SEQUENCE_STALE`
- `RECOVERY_REQUIRED`
- `RECOVERY_FAILED`
- `RECONNECT_WINDOW_EXPIRED`
- `TIMER_TOKEN_STALE`
- `COMMAND_NOT_RETRYABLE`
- `SERVER_BUSY`
- `RATE_LIMITED`
- `PERSISTENCE_FAILED`

### 33.5 Replay integrity

- `REPLAY_EXPIRED`
- `REPLAY_DELETED`
- `REPLAY_CAPABILITY_INVALID`
- `REPLAY_EVENT_GAP`
- `REPLAY_HASH_INVALID`
- `REPLAY_SIGNATURE_INVALID`
- `REPLAY_FORMAT_UNSUPPORTED`

Every domain error may include:

- stable code;
- retryable flag;
- structured safe details.

User-facing localized text is not part of the domain error.

---

## 34. Formal properties for TLA+ verification

The TLA+ model should verify at least the following safety properties.

### 34.1 Room safety

- At most one active Match exists per Room.
- At most one participant has host authority.
- A Match cannot start unless mode-specific player count is valid.
- A Match cannot start unless included players are ready.
- Settings cannot mutate after Countdown starts.
- A stale host-transfer event cannot override a newer host.

### 34.2 Command safety

- One `RequestID` produces at most one domain effect.
- A stale aggregate version cannot mutate state.
- A stale client sequence cannot mutate state.
- A stale timer generation cannot mutate state.
- A spectator cannot perform player commands.

### 34.3 Match safety

- A completed Match never becomes Active.
- Completion is finalized at most once.
- Fixed clues never change.
- Duel move acceptance is possible only for active player.
- Race first finisher is declared at most once.
- Finishing window cannot reopen after completion.
- No durable event is broadcast before commit.

### 34.4 Recovery safety

- Recovery never applies a committed event twice.
- Recovery never skips a committed event.
- Recovery cannot resurrect a completed Match.
- Recovery cannot accept commands before invariant validation.

### 34.5 Liveness

Subject to fair scheduling and available persistence:

- a valid ready room eventually starts or is cancelled;
- an active Race entering Finishing eventually completes;
- a disconnected Duel player eventually reconnects or forfeits;
- a Match in RecoveryPending eventually resumes or terminates;
- a Room in Results eventually rematches or expires.

---

## 35. Privacy and retention rules

### 35.1 Server-retained domain data

- Active room records: until completion or expiration
- Abandoned room records: 24 hours
- Replay events: 7 days
- Match result summaries: 90 days
- Temporary participant sessions: 7 days after last room
- Puzzle catalog: indefinitely
- Problem reports: 30 days
- Administrative audit history: 1 year

Retention jobs may remove expired data but must not leave dangling references that break still-retained results.

### 35.2 Local-only data

- Solo progress
- Daily streak
- personal statistics
- preferences
- local replay
- recent puzzle history
- accessibility settings

### 35.3 Prohibited domain data

Ninefold does not require or intentionally store:

- player email;
- real name;
- date of birth;
- precise location;
- permanent friend graph;
- direct messages;
- advertising profile;
- behavioral analytics profile.

---

## 36. Localization requirements

Domain values and events use stable machine identifiers.

Supported locales:

- English: `en`
- German: `de`
- Albanian: `sq`
- Turkish: `tr`

Domain rules:

- events contain structured values, not translated prose;
- errors contain stable codes, not localized logic;
- pluralization and grammar belong to clients;
- participant names preserve original script;
- room codes remain ASCII;
- locale never changes game semantics.

---

## 37. Acceptance scenarios

These scenarios are canonical examples. Implementations must preserve equivalent behavior.

### 37.1 Create and start Co-op

**Given**

- no Room exists for code `7KMP4R`;
- Mila requests a Co-op room with Medium difficulty.

**When**

- Room is created;
- Noah joins;
- both players become ready;
- Mila starts Countdown.

**Then**

- Mila is host;
- both players are included;
- Room settings become locked;
- one Match is created;
- after 3 seconds Match becomes Active;
- no other active Match exists for the Room.

### 37.2 Setting change resets readiness

**Given**

- Mila and Noah are ready in a Race room.

**When**

- host changes difficulty from Medium to Hard.

**Then**

- both ready states become false;
- Match cannot start until both ready again.

### 37.3 Co-op soft-lock override

**Given**

- Mila holds an 8-second soft lock on cell 20.

**When**

- Noah attempts to select cell 20.

**Then**

- Noah receives soft-lock warning;
- Noah may confirm override;
- override releases Mila’s lock and assigns focus to Noah;
- no Sudoku value changes until a value command commits.

### 37.4 Co-op simultaneous commands

**Given**

- Mila and Noah submit different values for the same editable cell nearly simultaneously.

**When**

- Room actor serializes commands.

**Then**

- first valid committed command changes board;
- second command validates against newer version;
- second command is rejected as stale or invalid;
- only one durable value mutation exists.

### 37.5 Race finish

**Given**

- four players are active;
- Mila submits final correct value.

**When**

- server verifies complete solution.

**Then**

- Mila becomes winner at server acceptance timestamp;
- Match enters Finishing;
- 60-second window starts;
- remaining players may continue;
- no client clock can change winner.

### 37.6 Race hint

**Given**

- hints are enabled.

**When**

- Noah uses a Reveal hint.

**Then**

- one eligible value is revealed;
- Noah receives 20-second penalty;
- Noah is marked Assisted;
- hint appears in result and replay.

### 37.7 Duel incorrect move

**Given**

- it is Mila’s turn.

**When**

- Mila submits an incorrect value.

**Then**

- value is rejected;
- board remains unchanged;
- mistake count increments;
- attempted digit is hidden from Noah;
- turn passes to Noah.

### 37.8 Duel scoring

**Given**

- Noah places a correct value that completes one row and one box.

**When**

- command commits.

**Then**

- Noah receives 1 base point;
- Noah receives 1 row bonus;
- Noah receives 1 box bonus;
- total score increase is 3;
- turn passes to Mila.

### 37.9 Duel timeout abandonment

**Given**

- Mila has already timed out twice consecutively.

**When**

- Mila’s next turn deadline expires.

**Then**

- third consecutive timeout is recorded;
- Match ends as abandonment;
- Noah wins.

### 37.10 Refresh and reconnect

**Given**

- Noah refreshes during active Co-op.

**When**

- valid room session resumes within 5 minutes.

**Then**

- same ParticipantID is restored;
- authoritative state is synchronized;
- accepted prior actions remain;
- gameplay input is enabled only after synchronization.

### 37.11 Server restart

**Given**

- active Race has committed events and a valid snapshot.

**When**

- server restarts.

**Then**

- Match enters RecoveryPending;
- state reconstructs from snapshot and later events;
- participants have 5 minutes to reconnect;
- Race resumes without duplicating events.

### 37.12 Replay verification

**Given**

- a completed Match replay and proof.

**When**

- browser recomputes event hashes and verifies Ed25519 signature.

**Then**

- replay is marked verified only if all event links and signature are valid;
- changing one payload causes verification failure.

### 37.13 Rematch

**Given**

- a Duel is complete.

**When**

- host starts rematch.

**Then**

- new Match ID is created;
- new Puzzle is assigned;
- old Match remains immutable;
- prior starting player is not selected again;
- participants return to ReadyCheck.

---

## 38. Explicit non-goals and rejected domain features

Unless future product scope explicitly changes them, the domain excludes:

- public matchmaking;
- permanent accounts;
- ranked ladders;
- public player profiles;
- direct messages;
- persistent teams or clubs;
- paid competitive advantages;
- room passwords;
- user-created public puzzles;
- external solver integration;
- blockchain;
- peer-to-peer authoritative gameplay;
- CRDT-based board authority;
- multiple authoritative game servers in V1;
- permanent cloud storage of personal solo history;
- public replay indexing.

---

## 39. Domain change process

A change requires explicit review when it affects:

- aggregate boundary;
- state transition;
- invariant;
- mode scoring;
- completion;
- disconnect behavior;
- replay visibility;
- privacy retention;
- event meaning;
- cryptographic proof format;
- TLA+ property.

For such changes:

1. Update this document.
2. Update or add domain tests.
3. Update event and command contracts.
4. Update TLA+ model when concurrency-sensitive.
5. Preserve backward replay compatibility or add versioned upcasting.
6. Never rewrite historical event meaning in place.

---

## 40. Final implementation directive

An implementation is domain-correct only when:

- every state mutation occurs through an explicit command;
- every accepted durable mutation emits immutable events;
- every invariant is enforced in domain code;
- mode-specific rules are not duplicated in transports or UI;
- server authority is preserved;
- recovery reconstructs the same state;
- replay reconstructs the same state;
- results cannot be finalized twice;
- privacy retention is enforceable;
- accessible clients can derive complete semantic state from structured events.

When implementation convenience conflicts with these rules, the implementation must change—not the invariant.
