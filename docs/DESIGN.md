# Ninefold Sudoku — Design Specification

**Document:** `docs/DESIGN.md`  
**Status:** Canonical product design, interaction, and accessibility specification  
**Current implementation scope:** Full MVP (`0.3.0`); deferred-feature designs are provisional
**Product:** Ninefold Sudoku  
**Compact brand:** Ninefold  
**Tagline:** *Solve together. Race the grid.*  
**Public URL:** `https://ninefold.recica.dev`  
**Repository:** `ninefold-sudoku`  
**Default branch:** `master`

---

## 1. Purpose and authority

This document defines how Ninefold Sudoku should look, feel, communicate, and behave across mobile, tablet, desktop, keyboard, touch, screen reader, light mode, dark mode, and all supported languages.

It is the canonical source of truth for:

- brand expression;
- visual identity;
- layout;
- component behavior;
- interaction design;
- Sudoku board presentation;
- player attribution;
- mobile and desktop controls;
- lobby and Match flows;
- Co-op, Race, Duel, Solo, Daily, and Replay interfaces;
- loading, empty, error, disconnected, and recovery states;
- accessibility behavior;
- localization-sensitive layout;
- motion;
- sound;
- haptics;
- responsive design;
- SEO-facing page presentation;
- content style;
- design quality gates.

This document does not define authoritative gameplay rules. Those are in `docs/DOMAIN.md`.

This document does not define technical implementation, persistence, protocols, or deployment. Those are in `docs/ARCHITECTURE.md`.

This document does not define release scope. That is in `docs/PRODUCT.md`.

When documents conflict:

1. `DOMAIN.md` controls rules and invariants.
2. `ARCHITECTURE.md` controls implementation constraints.
3. `PRODUCT.md` controls scope and priorities.
4. `DESIGN.md` controls visual and interaction expression.

Current-scope rule:

- Co-op, online Solo, replay, and the MVP public pages are binding designs;
- Race, Duel, Daily Ninefold, offline Solo, PWA installation, host approval, full spectator UX, additional locales, About, and dedicated mode pages are provisional;
- do not implement or scaffold provisional screens merely because they are described here.

---

## 2. Design vision

Ninefold should feel like a calm, refined, modern game that respects the player’s attention.

The experience should be:

- clear before clever;
- elegant without becoming sterile;
- social without becoming noisy;
- polished without becoming heavy;
- playful without becoming childish;
- competitive without becoming aggressive;
- technical without looking like a developer tool;
- accessible without looking compromised;
- fast without feeling abrupt.

The design goal is not to imitate newspaper Sudoku, mobile casino games, generic SaaS dashboards, or arcade interfaces.

The intended emotional tone is:

> Quiet focus, shared presence, and satisfying precision.

---

## 3. Core design principles

### 3.1 The board is the product

During gameplay, the Sudoku board must remain the dominant visual element.

The interface should not compete with it through:

- large navigation;
- oversized sidebars;
- decorative illustration;
- excessive statistics;
- persistent marketing;
- unnecessary panels;
- large animated effects.

### 3.2 Multiplayer state must be understandable at a glance

A player should always know:

- who is present;
- who is host;
- whether they are ready;
- whose turn it is;
- whether the connection is healthy;
- whether their move is pending or confirmed;
- whether a cell is fixed, entered, conflicted, incorrect, or focused;
- how opponents are progressing where the mode allows it;
- whether the Match is active, finishing, completed, reconnecting, or recovering.

### 3.3 Color is never the only carrier

Color may reinforce meaning, but every important state also requires one or more of:

- icon;
- shape;
- border style;
- text;
- initials;
- pattern;
- accessible label;
- semantic announcement.

### 3.4 Interaction should reward fluency

The interface must support:

- quick first-time use;
- one-handed mobile play where practical;
- fast keyboard play;
- number-first and cell-first input;
- minimal pointer travel;
- predictable focus;
- shortcuts;
- persistent preferences.

### 3.5 Privacy should be visible but not theatrical

Ninefold should communicate privacy through:

- no account wall;
- no tracking consent banner;
- concise privacy language;
- local-data controls;
- limited personal information;
- clear replay retention;
- no invasive prompts.

Do not use exaggerated security claims or fear-based messaging.

### 3.6 Accessibility is part of visual quality

Accessible states should look intentional and polished.

Visible focus, non-color indicators, large targets, reduced motion, readable contrast, and screen-reader support must be built into the primary design—not added as secondary patches.

### 3.7 Performance is a design constraint

Avoid design choices that require:

- large image downloads;
- heavy animation libraries;
- complex canvas rendering;
- oversized font files;
- continuous polling;
- decorative video;
- excessive blur;
- expensive visual effects.

---

## 4. Brand system

### 4.1 Product naming hierarchy

Use:

- Full name: **Ninefold Sudoku**
- Compact brand: **Ninefold**
- Descriptor: **Multiplayer Sudoku**
- Tagline: **Solve together. Race the grid.**

Use the full name in:

- page title;
- first homepage heading context;
- social image;
- search metadata;
- About page;
- footer;
- repository description.

Use “Ninefold” alone in:

- compact navigation;
- favicon context;
- in-game header;
- mobile app title where space is limited;
- result cards.

### 4.2 Logo concept

The logo should derive from a 3×3 grid.

Preferred conceptual properties:

- simple enough for a favicon;
- recognizable in monochrome;
- geometric;
- visually balanced;
- not a literal full Sudoku puzzle;
- suggests connection, collaboration, or movement through one altered cell or connecting stroke.

Possible logo construction:

- 3×3 grid of rounded squares;
- one square offset, connected, or highlighted;
- subtle pathway through three cells;
- negative-space “N” implied inside the grid.

Avoid:

- tiny Sudoku digits inside the logo;
- gradients that fail in monochrome;
- excessive detail;
- speech bubbles;
- generic game-controller iconography;
- puzzle-piece clichés.

### 4.3 Logo lockups

Required lockups:

1. Mark only
2. Mark + Ninefold
3. Mark + Ninefold Sudoku
4. Monochrome light
5. Monochrome dark
6. Social preview lockup

### 4.4 Favicon and app icon

The icon should remain legible at:

- 16×16
- 32×32
- 48×48
- 180×180
- 192×192
- 512×512

The smallest versions may simplify borders or remove internal detail.

### 4.5 Social preview

Static Open Graph image should include:

- Ninefold mark;
- “Ninefold Sudoku”;
- “Multiplayer Sudoku”;
- simplified 3×3 or 9×9 board motif;
- `ninefold.recica.dev`.

It should not show:

- private Room data;
- participant names;
- a real puzzle solution;
- live Match screenshots requiring context.

---

## 5. Visual personality

Ninefold should feel:

- modern;
- geometric;
- calm;
- friendly;
- premium;
- exact;
- slightly playful;
- distinctly multiplayer.

Avoid:

- neon cyberpunk;
- sports-betting aesthetics;
- corporate dashboard styling;
- hand-drawn children’s game visuals;
- skeuomorphic wood/paper textures;
- overly clinical grayscale;
- extreme glassmorphism;
- exaggerated depth.

---

## 6. Design tokens

All visual decisions should be expressed through semantic design tokens.

Do not hard-code arbitrary color, radius, spacing, or shadow values inside individual components.

### 6.1 Token categories

```text
color
typography
spacing
radius
border
shadow
motion
z-index
layout
touch-target
player-marker
```

### 6.2 Semantic color roles

Required semantic roles:

```text
surface.canvas
surface.primary
surface.secondary
surface.elevated
surface.overlay
surface.interactive
surface.selected
surface.related
surface.pending

text.primary
text.secondary
text.muted
text.inverse
text.link
text.disabled

border.default
border.strong
border.focus
border.selected
border.conflict
border.pending

brand.primary
brand.primaryHover
brand.primaryActive
brand.subtle

state.success
state.successSubtle
state.warning
state.warningSubtle
state.error
state.errorSubtle
state.info
state.infoSubtle

board.clue
board.entry
board.note
board.grid
board.boxGrid
board.matchHighlight
board.conflict
board.incorrect
board.pending
board.remoteFocus
```

### 6.3 Player-color tokens

Each Match participant receives a semantic slot:

```text
player.1
player.2
player.3
player.4
player.5
player.6
player.7
player.8
```

Each slot must define:

- foreground;
- subtle background;
- strong border;
- high-contrast marker;
- dark-mode equivalents.

Player colors must be tested for:

- deuteranopia;
- protanopia;
- tritanopia;
- grayscale;
- light mode;
- dark mode.

### 6.4 Player markers

Each player color pairs with a marker:

- circle;
- square;
- triangle;
- diamond;
- hexagon;
- star;
- cross;
- ring.

The exact marker set may change, but a player always has both:

- color;
- non-color marker.

---

## 7. Color direction

### 7.1 Brand accent

Recommended brand family:

- indigo;
- violet;
- blue-violet.

The accent should feel thoughtful and modern rather than energetic or playful.

Use brand accent for:

- primary actions;
- active navigation;
- selected control states;
- Ninefold branding;
- focus reinforcement where it does not conflict with accessibility.

Do not use brand accent for all player attribution.

### 7.2 Light mode

Preferred light-mode qualities:

- warm off-white canvas;
- white or lightly tinted surfaces;
- graphite text;
- subtle cool-gray borders;
- restrained brand accent;
- high board contrast.

Avoid pure white everywhere because it can feel harsh.

### 7.3 Dark mode

Preferred dark-mode qualities:

- deep graphite canvas;
- slightly lighter board surface;
- off-white text;
- strong but not glowing player colors;
- clear clue/entry distinction;
- reduced blur and glow.

Avoid pure black as the only background.

### 7.4 Semantic color behavior

Error color must not be the only indication of:

- conflict;
- incorrect value;
- rejected action;
- failed connection;
- invalid replay proof.

Combine with:

- icon;
- border style;
- label;
- accessible text.

---

## 8. Typography

### 8.1 Font strategy

Prefer:

1. high-quality system font stack; or
2. one carefully optimized, self-hosted sans-serif family.

Do not use third-party font CDNs.

The selected typeface must support:

- English;
- German;
- Albanian;
- Turkish;
- all required diacritics;
- tabular numerals;
- clear differentiation between `1`, `I`, `l`, `0`, `O`.

### 8.2 Typography roles

```text
display
heading-1
heading-2
heading-3
body
body-small
label
caption
button
numeric-board
numeric-timer
code
```

### 8.3 Board numerals

Board numerals should:

- use tabular lining numbers;
- have strong legibility;
- avoid excessive geometric stylization;
- make clues visibly heavier than entries;
- remain readable at mobile sizes;
- scale fluidly with cell size.

### 8.4 Typographic hierarchy

Use weight and size before relying on color.

Avoid:

- excessive all-caps;
- tiny muted text;
- very light font weights;
- condensed body text;
- center-aligned paragraphs longer than two lines.

### 8.5 Line length

Public content pages:

- ideal text line length: 60–75 characters;
- maximum readable width: approximately 70ch.

---

## 9. Spacing and layout rhythm

### 9.1 Base spacing

Use a consistent spacing scale, preferably based on 4px increments.

Suggested semantic steps:

```text
2, 4, 8, 12, 16, 20, 24, 32, 40, 48, 64, 80
```

### 9.2 Grid alignment

The Ninefold 3×3 motif should influence:

- card group spacing;
- section rhythm;
- loading indicators;
- decorative dividers;
- icon geometry;
- result summary layout.

Do not force every layout into a literal 3-column grid.

### 9.3 Density

Gameplay surfaces should be moderately dense.

Marketing and About pages may use more whitespace.

Settings and lobby should avoid large empty gaps that push controls below the fold unnecessarily.

---

## 10. Shape, border, and depth

### 10.1 Radius

Use modest radii.

Recommended hierarchy:

- small controls: 6–8px;
- buttons/inputs: 8–10px;
- cards/panels: 12–16px;
- large modal: 16–20px.

The Sudoku grid itself should remain relatively precise and geometric.

### 10.2 Borders

Use borders to communicate structure more than shadows.

Preferred:

- subtle panel borders;
- strong 3×3 box dividers;
- clear focus outline;
- semantic conflict outline.

### 10.3 Shadows

Use restrained shadows only for:

- dialogs;
- floating menus;
- elevated mobile panels;
- active drag/focus layers where necessary.

Avoid heavy card shadows across the entire product.

---

## 11. Iconography

Use a consistent SVG icon set.

Icons should be:

- simple;
- stroked or consistently filled;
- recognizable at 16–24px;
- paired with labels for ambiguous actions.

Do not use icons alone for:

- ready;
- host;
- hints;
- privacy;
- connection;
- replay verification;
- destructive actions.

Directional icons may reflect locale direction only if future RTL locales are added. Current supported languages are left-to-right.

---

## 12. Motion system

### 12.1 Motion principles

Motion should:

- clarify state change;
- confirm interaction;
- maintain spatial continuity;
- celebrate completion;
- never delay input.

### 12.2 Standard durations

Suggested ranges:

- immediate feedback: 80–120ms;
- control transition: 120–180ms;
- panel transition: 180–240ms;
- completion celebration: 400–800ms;
- replay timeline changes: tied to playback speed.

### 12.3 Easing

Use:

- ease-out for entering;
- ease-in for leaving;
- standard cubic-bezier for movement;
- no spring animation unless subtle and justified.

### 12.4 Reduced motion

When reduced motion is enabled:

- remove pulsing;
- remove large board celebrations;
- replace animated movement with fades or instant state;
- stop decorative looping animation;
- keep necessary progress and timer updates.

### 12.5 Prohibited motion

Do not:

- shake the entire screen on error;
- zoom the board dramatically;
- animate every cell selection;
- use confetti that obscures the board;
- create motion that affects reading;
- autoplay animated hero scenes.

---

## 13. Sound and haptics

### 13.1 Defaults

- Sound: off by default
- Haptics: off by default unless platform conventions and explicit preference support justify otherwise

### 13.2 Optional sounds

- countdown;
- accepted move;
- incorrect move;
- turn change;
- ping;
- Match completion.

### 13.3 Sound requirements

- volume must be restrained;
- no background music;
- no repeated ticking by default;
- each sound has visual and textual equivalent;
- sound preference persists locally.

### 13.4 Haptics

Possible supported haptics:

- subtle tap on accepted move;
- stronger short pulse on mistake;
- double pulse on turn change;
- completion pulse.

Haptics must never be required to understand state.

---

## 14. Responsive breakpoints

Use content-driven breakpoints rather than device labels.

Suggested ranges:

```text
compact: < 640px
medium: 640–1023px
wide: 1024–1439px
large: ≥ 1440px
```

Do not assume one specific phone width.

### 14.1 Compact layout goals

- board fits width;
- number pad remains reachable;
- no horizontal scroll;
- Match status remains visible;
- secondary panels collapse;
- Sudoku cells remain at least 24×24 CSS pixels;
- number-pad buttons and primary controls remain at least 44×44 CSS pixels.

### 14.2 Medium layout goals

- board remains central;
- compact side information may appear;
- number pad can move beside or below board;
- panels may use two columns.

### 14.3 Wide layout goals

Recommended gameplay layout:

```text
Player/status panel | Sudoku board | Controls/activity panel
```

The board should still appear visually dominant.

---

## 15. Global navigation

### 15.1 Public pages

Desktop navigation may include:

- Play
- How to Play
- Privacy
- Accessibility
- Settings or theme controls

Modes and About may be added after those pages enter scope.

Mobile navigation should use a compact menu.

### 15.2 Active gameplay

During a Match, navigation should be reduced to:

- Ninefold mark;
- Room/Match status;
- connection indicator;
- settings;
- leave action.

Do not expose full marketing navigation during active play.

### 15.3 Back behavior

Back actions must clearly communicate consequences when they may:

- leave Room;
- abandon Race;
- resign Duel;
- interrupt Solo;
- delete local progress.

---

## 16. Homepage design

### 16.1 Above the fold

Must include:

- Ninefold Sudoku identity;
- headline;
- concise supporting copy;
- Create Room;
- Join code input;
- Play Solo;

Recommended headline:

```text
Sudoku is better together.
```

Recommended supporting copy:

```text
Create a private room, invite friends, and solve the same Sudoku grid together.
```

### 16.2 Primary interaction block

Create and Join should appear as the primary visual cluster.

Possible desktop structure:

```text
[ Create a room ]    [ Room code ______ ] [ Join ]
```

Mobile:

```text
[ Create a room ]
Room code
[ ______ ]
[ Join ]
```

### 16.3 Secondary content

Below the play block:

- short Co-op visual;
- future-mode cards only when clearly labelled as unavailable;
- privacy statement;
- replay integrity statement;
- accessibility statement;
- portfolio/technical link.

### 16.4 Homepage constraints

- no autoplay video;
- no large hero illustration that pushes play below fold;
- no testimonial carousel;
- no pop-up;
- no install prompt on first visit;
- no newsletter form;
- no cookie banner.

---

## 17. Create Room flow

### 17.1 Entry

Create Room should be one concise flow.

Required fields:

- display name;
- difficulty.

Optional advanced settings should remain collapsed until lobby or advanced toggle.

Co-op is the only enabled multiplayer mode in the MVP.

### 17.2 Mode selection

Mode selection is deferred until Race enters scope.

Mode cards should include:

- icon;
- name;
- one-sentence summary;
- player count;
- status if deferred or disabled.

Order:

1. Co-op
2. Race
3. Duel

### 17.3 Difficulty selection

Use clear labeled controls:

- Easy
- Medium
- Hard
- Expert

Duel hides Expert.

### 17.4 Validation

Validation should be:

- inline;
- specific;
- non-destructive;
- localized;
- announced accessibly.

### 17.5 Completion

After creation:

- transition to lobby;
- show code prominently;
- provide copy link;
- confirm host identity.

---

## 18. Join flow

### 18.1 Code entry

Room-code input:

- uppercase display;
- accepts lowercase;
- ignores surrounding spaces;
- excludes ambiguous character guidance;
- groups six characters visually if helpful;
- supports paste.

### 18.2 Invitation link

Opening `/join/[code]` shows:

- safe Room summary;
- mode;
- difficulty;
- available role;
- lock state;
- display-name field.

Do not show participant names before successful join.

Approval state is added only with the deferred host-approval workflow.

### 18.3 Full player seats

If players are full and spectator access exists:

- explain clearly;
- offer Join as Spectator;
- do not silently change role.

### 18.4 Awaiting approval

This state is deferred with the host-approval workflow.

Approval pending state should show:

- Room summary;
- waiting indicator;
- cancel request;
- no false promise of immediate approval.

---

## 19. Lobby design

### 19.1 Information hierarchy

Lobby should show:

1. Room code and share action
2. Mode and difficulty
3. participant list
4. ready states
5. settings
6. Start action
7. activity feed

### 19.2 Participant row

Each row includes:

- marker;
- display name;
- Host badge where applicable;
- role;
- ready state;
- connection state;
- host actions in overflow menu.

### 19.3 Ready state

Ready control should be obvious and persistent.

States:

- Not ready
- Ready
- Waiting for host
- Countdown

Do not rely only on a green dot.

### 19.4 Host settings

Show only mode-relevant options.

Changing a gameplay setting must visibly communicate:

> Ready states were reset because the game settings changed.

### 19.5 Room code

Code should be:

- large enough to read across a room;
- copyable;
- accessible label included;
- visually distinct from body text;
- never translated.

### 19.6 Activity feed

Show lightweight events such as:

- Mila joined
- Difficulty changed to Hard
- Noah is ready
- Host transferred to Mila

Retain only latest relevant entries visually.

### 19.7 Start button

Disabled state should explain why:

- waiting for players;
- not everyone ready;
- invalid Duel count;
- pending approval;
- Match already active.

---

## 20. Countdown design

### 20.1 Visual

Countdown should appear:

- centered over board/lobby transition;
- large;
- high contrast;
- calm, not explosive.

### 20.2 Accessibility

Announce:

- “Match starts in 3”
- “Match starts in 2”
- “Match starts in 1”
- “Match started”

Reduced-motion mode may use instant number changes without scale animation.

### 20.3 Cancellation

If host cancels:

- transition back to Lobby;
- explain cancellation;
- show that readiness was reset;
- do not leave users on a frozen overlay.

---

## 21. Gameplay shell

### 21.1 Desktop layout

Recommended:

```text
┌─────────────────────────────────────────────────────────────┐
│ Header: Room / mode / timer / connection / settings         │
├───────────────┬────────────────────────┬────────────────────┤
│ Players       │ Sudoku board           │ Controls/activity  │
│ Progress      │                        │ Notes / hints       │
│ Status        │                        │ Mode-specific panel │
└───────────────┴────────────────────────┴────────────────────┘
```

### 21.2 Mobile layout

```text
Header
Match status
Sudoku board
Number pad
Primary controls
Expandable player/activity panel
```

### 21.3 Board sizing

- board should use available width;
- preserve square aspect ratio;
- never shrink below usable touch size merely to keep side panels visible;
- side panels should collapse first.

### 21.4 Header

May include:

- mode;
- difficulty;
- timer;
- Room code;
- current turn or Match status;
- connection state;
- settings;
- leave.

Avoid crowding.

---

## 22. Sudoku board design

### 22.1 Structural grid

- 9×9 CSS Grid
- clear 3×3 box dividers
- lighter individual cell dividers
- square cells
- consistent focus and selection outlines

### 22.2 Cell state hierarchy

Each cell may visually represent:

- clue;
- player entry;
- selected;
- related row/column/box;
- same digit;
- note mode;
- direct conflict;
- incorrect value;
- pending optimistic value;
- remote focus;
- soft lock;
- hint target;
- Duel active consideration hidden from opponent;
- completed structure animation.

### 22.3 Clues

Clues:

- heavier weight;
- primary text color;
- no player attribution;
- visually stable;
- cannot look interactive when disabled.

### 22.4 Player entries

Player entries:

- slightly lighter weight than clues;
- use player attribution treatment where mode requires;
- remain readable without color;
- pending entries receive distinct pending indicator.

### 22.5 Notes

Candidate grid:

- aligned 3×3 miniature positions;
- readable at compact widths;
- no overlap;
- consistent numeral placement;
- private/shared ownership communicated outside the tiny digits when necessary.

### 22.6 Selection

Selected cell should use:

- clear border or outline;
- background shift;
- focus ring when keyboard focus is active.

Selection must remain visible in both themes.

### 22.7 Related cells

Same row, column, and box receive a subtle background.

Do not over-highlight to the point that the board becomes noisy.

### 22.8 Matching numbers

Cells containing same digit as selected value may receive stronger but restrained highlight.

### 22.9 Direct conflict

Conflict should combine:

- error border;
- icon or patterned corner;
- accessible label;
- optional subtle background.

### 22.10 Incorrect value

Where rules permit revealing solution error:

- distinct from direct conflict;
- do not use only red text;
- include semantic icon or underline treatment.

### 22.11 Pending move

Optimistic pending value:

- appears immediately;
- uses reduced opacity or dotted underline;
- includes hidden text for assistive technology such as “pending confirmation”;
- finalizes on acknowledgement;
- visibly reverts on rejection.

### 22.12 Remote focus and soft lock

Co-op remote focus:

- player-colored outline;
- player marker badge;
- no full-cell fill that obscures value.

Soft lock:

- player marker;
- subtle timer/fade;
- warning before override.

### 22.13 Fixed clue interaction

Hovering or selecting a clue may still highlight relationships but must not imply editability.

### 22.14 Completed row/column/box

Use a brief subtle sweep or border emphasis.

Reduced motion:

- static success highlight.

---

## 23. Number pad

### 23.1 Mobile

Fixed below board.

Required buttons:

- digits 1–9;
- erase;
- notes toggle;

Optional controls:

- hint;
- check;
- input mode switch.

### 23.2 Desktop

May appear beside board.

Keyboard remains primary for experienced desktop users.

### 23.3 Digit state

Digit buttons may show:

- completed digit count;
- disabled when all nine occurrences are fixed/placed if rule-safe;
- active in number-first mode;
- note-mode state.

Do not disable digits in ways that reveal hidden correctness.

### 23.4 Input modes

Support:

#### Cell-first

1. Select cell.
2. Select digit.

Default.

#### Number-first

1. Select digit.
2. Select one or more cells.

Preference persists locally.

### 23.5 Notes toggle

Notes mode must have:

- labeled toggle;
- icon;
- visible active state;
- keyboard shortcut;
- accessible announcement.

---

## 24. Keyboard behavior

### 24.1 Required shortcuts

- Arrow keys: move selection
- `1–9`: place digit
- Backspace/Delete: erase
- `N` or Space: toggle notes
- `Ctrl/Cmd + Z`: undo only in a mode that explicitly supports it; Co-op MVP does not
- `H`: hint where enabled
- Escape: close overlay or clear selection

### 24.2 Focus model

The board should use a predictable roving-tabindex or grid navigation pattern.

Tab should move through major interface regions, not every one of 81 cells unless the chosen accessible grid pattern requires it.

Arrow keys navigate cells.

### 24.3 Shortcut help

Provide a discoverable shortcut reference in:

- settings/help;
- first desktop use;
- board overflow menu.

Do not show a large shortcut overlay automatically.

---

## 25. Co-op design

### 25.1 Primary feeling

Co-op should feel like people working on one physical puzzle together.

### 25.2 Player attribution

Accepted entries show a subtle contributor marker.

Possible treatments:

- small marker in corner;
- colored baseline;
- shaped badge;
- accessible label.

Do not make every cell visually loud.

### 25.3 Shared notes

Shared notes should update smoothly.

When another player changes notes:

- briefly indicate remote change;
- avoid distracting animation;
- preserve local focus.

### 25.4 Soft lock warning

When selecting a cell focused by another player:

```text
Mila is working on this cell.
[Choose another] [Use anyway]
```

On mobile, keep warning compact and non-modal where possible.

### 25.5 Pings

Ping interaction:

- long press or contextual button;
- select intent;
- temporary pulse;
- label visible in activity panel;
- rate-limited feedback.

MVP pings use the targeted intents `look_here`, `unsure`, and `try_this_area`. A target cell or region is required. Pings are durable and appear in replay.

### 25.6 Reactions

Use concise icons with localized labels:

- I agree
- Nice move

Do not add free-form text chat in V1.

Reactions are untargeted, ephemeral, and omitted from replay.

### 25.7 Undo

Co-op MVP has no undo command or control. Players reverse values with explicit erase and reverse notes with the normal note toggle.

### 25.8 Contribution summary

Results may show:

- values placed;
- notes contributed;
- hints used;
- reconnects;
- completion time.

Do not rank Co-op contributors as winners/losers.

---

## 26. Race design

**Status:** Provisional; do not implement until Race enters scope.

### 26.1 Primary feeling

Race should create tension without allowing direct copying.

### 26.2 Opponent panel

Each opponent row may show:

- display name;
- marker;
- correct progress count or percentage;
- delayed heatmap;
- status: solving, finished, disconnected, assisted.

### 26.3 Heatmap

Heatmap:

- shows filled/correct regions without digits;
- updates in delayed batches;
- includes text equivalent;
- uses pattern/opacity plus player marker, not only color.

### 26.4 Own board

Own board remains full size and primary.

Do not show multiple full opponent boards during live play.

### 26.5 Winner declaration

When first player finishes:

- show winner banner;
- start visible 60-second finishing window;
- do not block remaining players’ boards;
- allow winner to observe results or replay preview.

### 26.6 Finishing window

Remaining time should be clear, but avoid frantic animation.

### 26.7 Final ranking

Show:

- position;
- finish status;
- time;
- correct progress if unfinished;
- mistakes;
- assistance;
- disconnect status.

---

## 27. Duel design

**Status:** Provisional; do not implement until Duel enters scope and its erase, pass, recovery, scoring, and privacy rules are ratified.

### 27.1 Primary feeling

Duel should feel strategic, turn-based, and focused.

### 27.2 Turn indicator

Must clearly show:

- whose turn;
- remaining turn time;
- score;
- timeout count;
- connection state.

Use both text and player marker.

### 27.3 Opponent thinking

Show:

```text
Noah is thinking…
```

Do not show selected cell before accepted move.

### 27.4 Scoreboard

Display:

- player markers;
- names;
- total score;
- mistakes;
- timeouts;
- current turn.

Avoid complex sports-style graphics.

### 27.5 Move scoring feedback

When a move scores:

```text
+1 correct
+1 row
+1 box
```

Use concise temporary feedback near scoreboard, not obstructing board.

### 27.6 Wrong move

For acting player:

- clear rejection;
- no board mutation;
- “Turn lost” explanation;
- mistake count update.

For opponent:

- show that an attempt was rejected;
- do not show attempted digit.

### 27.7 Disconnect pause

Show:

- reconnect countdown;
- whether protected pause has been used;
- what happens when time expires.

### 27.8 Resignation

Resign action:

- clearly destructive;
- requires confirmation;
- explains immediate loss.

---

## 28. Solo design

Online Solo is current MVP scope. Offline behavior is provisional.

### 28.1 Solo entry

Show:

- Continue Last Puzzle, when available;
- difficulty choices;
- Random;
- Guided/Classic explanation;
- recent personal bests where locally available.

### 28.2 Guided mode

Explain:

```text
Mistakes are highlighted immediately, and Check Board is available.
```

### 28.3 Classic mode

Explain:

```text
Only direct Sudoku conflicts are shown until you finish.
```

### 28.4 Timer

Timer:

- starts on first editable interaction;
- pause control;
- hint penalty displayed separately;
- closed-tab time excluded.

### 28.5 Hint ladder

Hint panel:

1. Nudge
2. Explain
3. Reveal

If Explain is not yet released, do not show disabled noise. Show only currently supported levels.

### 28.6 Check Board

Available only where rules permit.

Use clear confirmation that it checks but does not fix.

### 28.7 Continue state

Resume should restore:

- values;
- notes;
- timer;
- hint state;
- input preference;
- theme/locale.

---

## 29. Daily Ninefold design

**Status:** Provisional; do not implement until Daily Ninefold enters scope.

### 29.1 Entry card

Homepage Daily card shows:

- today’s difficulty;
- completion status;
- local streak;
- time until next Daily.

### 29.2 Daily identity

Daily should feel special but not like a separate product.

Use subtle date or sequence badge.

### 29.3 Completion

Show:

- completion time;
- hints;
- mistakes;
- local streak;
- replay;
- shareable result summary.

No public leaderboard in V1.

---

## 30. Replay design

### 30.1 Replay shell

Required controls:

- play/pause;
- timeline;
- current time;
- total time;
- speed: 0.5×, 1×, 2×, 4×;
- previous/next event;
- integrity status;
- event filters if needed.

### 30.2 Timeline

Timeline markers for:

- mistakes;
- hints;
- disconnects;
- Race finish;
- Duel score events;
- completion.

Markers must be distinguishable without color alone.

### 30.3 Integrity status

Verified:

```text
✓ Replay integrity verified
```

Failed:

```text
Replay integrity could not be verified
```

Explanation:

> Verification confirms that the replay has not been altered since the Ninefold server sealed it.

Do not claim:

- cheating was impossible;
- server implementation was mathematically proven correct;
- replay proves identity.

### 30.4 Loading

Replay data should load after result summary.

Do not block result screen while verifier code downloads.

### 30.5 Co-op replay

Show:

- shared board;
- contributor markers;
- player list;
- targeted pings;
- activity timeline.

### 30.6 Race replay

Desktop:

- side-by-side boards where space allows.

Mobile:

- one board at a time;
- player switcher;
- synchronized progress list.

### 30.7 Duel replay

Show:

- shared board;
- current turn;
- score timeline;
- move result;
- time spent;
- hidden-digit protection.

### 30.8 Expired replay

Explain:

```text
This replay has expired and is no longer stored.
```

Offer:

- return home;
- create Room;
- play Solo.

---

## 31. Results design

### 31.1 Co-op result

Emphasize shared success:

```text
Puzzle solved together
```

Show:

- time;
- difficulty;
- players;
- mistakes;
- hints;
- contribution summary;
- Replay;
- Rematch.

### 31.2 Race result

Emphasize ranking:

- winner;
- final order;
- finish times;
- unfinished progress;
- assisted flag;
- replay;
- rematch.

### 31.3 Duel result

Emphasize:

- winner or draw;
- score;
- result reason;
- mistakes;
- timeouts;
- thinking time;
- replay;
- rematch.

### 31.4 Primary actions

Order:

1. Rematch
2. View replay
3. Return home / leave Room

### 31.5 Result card

A compact shareable summary may include:

```text
Ninefold Sudoku
Co-op · Hard · 8:42
4 players · 2 mistakes · No reveals
```

Do not include solution or sensitive data.

---

## 32. Connection and recovery states

### 32.1 Connection indicator

Use explicit states:

- Connected
- Reconnecting
- Synchronizing
- Read-only in this tab
- Maintenance
- Recovery failed

### 32.2 Reconnecting banner

Nonblocking banner:

```text
Connection lost. Reconnecting…
```

Gameplay input disabled.

Board remains visible.

### 32.3 Synchronizing

```text
Connected. Syncing the latest moves…
```

Do not enable input until complete.

### 32.4 Read-only tab

```text
This Room is active in another tab.
[Control from this tab]
```

### 32.5 Server restart recovery

```text
Ninefold restarted and is restoring this Match.
```

If resumed:

```text
Match restored.
```

If cancelled:

```text
This Match could not be restored.
```

Provide next action.

### 32.6 Persistence failure

Do not say “Your move was saved” before confirmation.

If authoritative saving pauses:

```text
The game is temporarily paused while Ninefold protects the Match state.
```

---

## 33. Loading states

### 33.1 Principles

- Show structure quickly.
- Avoid spinner-only screens.
- Use skeletons for public pages and lobby.
- Do not fake progress percentages.
- Keep loading copy concise.

### 33.2 Gameplay loading

Show:

- board frame;
- Match status;
- “Loading puzzle…”;
- no editable controls.

### 33.3 Replay loading

Show result first, then lazy replay loading.

### 33.4 Slow network

After a reasonable delay, add:

```text
This is taking longer than usual.
```

Do not immediately show failure.

---

## 34. Empty states

Required empty states:

- no local Solo history;
- no Daily streak;
- no replay events yet;
- no spectators;
- no activity feed entries;
- no unfinished puzzle;
- no admin reports.

Empty states should include one useful action, not decorative copy only.

---

## 35. Error states

### 35.1 Room not found

```text
This Room does not exist or is no longer available.
```

Actions:

- enter another code;
- create Room;
- return home.

### 35.2 Room expired

```text
This Room has expired.
```

### 35.3 Room locked

```text
This Room is locked to new participants.
```

### 35.4 Name rejected

```text
Please choose a different display name.
```

### 35.5 Replay verification failed

```text
Replay integrity could not be verified.
```

Do not automatically imply malicious tampering; could be unsupported format or corrupted data.

### 35.6 Unexpected error

Show:

- plain-language message;
- retry action;
- request ID;
- contact address.

Never show stack trace.

---

## 36. Settings design

Sections:

1. Appearance
2. Gameplay
3. Accessibility
4. Sound and haptics
5. Privacy and local data
6. Language, after more than one locale ships

### 36.1 Appearance

- System
- Light
- Dark

### 36.2 Gameplay

- Cell-first / Number-first
- Auto-remove notes
- Show matching digits
- Keyboard shortcut help
- Sound preference

### 36.3 Accessibility

- Reduced motion override
- High-contrast enhancement if offered
- Larger board labels
- Screen-reader verbosity if needed
- Haptics

Do not create unnecessary duplicate settings for OS preferences unless user control adds value.

### 36.4 Privacy

- Explain device-local data
- Clear local data
- Replay retention explanation
- Privacy page link

### 36.5 Language

The MVP ships English only and does not show a one-option language selector.

When additional locales ship, display language names in their own language.

---

## 37. Accessibility details

### 37.1 Board semantics

Use an accessible grid pattern.

Each cell label should communicate as applicable:

- row;
- column;
- clue or editable;
- current value;
- notes;
- selected;
- conflict;
- incorrect;
- pending;
- contributor;
- soft lock.

Example:

```text
Row 3, column 7, editable, value 8, entered by Mila, pending confirmation.
```

### 37.2 Focus visibility

Focus ring must:

- meet contrast requirements;
- not be clipped;
- remain visible in selected and conflict states;
- be distinguishable from remote focus.

### 37.3 Live regions

Separate regions may be used for:

- urgent Match state;
- routine move feedback;
- connection feedback.

Avoid repeatedly interrupting screen-reader users.

### 37.4 Progress text

Race heatmap equivalent:

```text
Noah: 31 of 47 editable cells correct.
```

### 37.5 Duel turn text

```text
Your turn. 18 seconds remaining.
```

### 37.6 Timers

Do not announce every second by default.

Announce meaningful thresholds:

- turn begins;
- 10 seconds;
- 5 seconds;
- expired.

### 37.7 Zoom and reflow

At 200% zoom:

- board remains usable;
- side panels stack;
- no clipped controls;
- horizontal page scrolling avoided where possible;
- dialogs remain within viewport.

### 37.8 Touch targets

Sudoku board cells:

- minimum 24×24 CSS pixels;
- maximize available size without forcing horizontal page scrolling;
- remain one coherent two-dimensional grid.

Number-pad buttons and primary interactive controls:

- minimum 44×44 CSS pixels;
- adequate spacing;
- no tiny icon-only destructive controls.

---

## 38. Localization-sensitive design

Current MVP locale:

- English

Planned pre-1.0 locales:

- German
- Albanian
- Turkish

Use a generated pseudo-locale during MVP development to test expansion and placeholder reordering.

### 38.1 Text expansion

Test with at least 35% expansion.

Avoid:

- fixed-width buttons;
- single-line assumptions;
- icon-only replacements for longer labels;
- truncation of critical actions.

### 38.2 Turkish casing

Do not uppercase human-language labels mechanically.

Room codes may use ASCII uppercase.

### 38.3 Diacritics

Fonts and line heights must render:

- ä, ö, ü, ß;
- ë, ç;
- ğ, ş, ı, İ, ö, ü.

### 38.4 Locale-independent Room state

Each participant sees localized UI independently.

Participant-generated names remain unchanged.

### 38.5 Translation placeholders

Layouts must support reordered placeholders.

Do not visually compose a sentence from separately positioned fragments that assume English order.

---

## 39. Content design and tone

### 39.1 Voice

Ninefold copy should be:

- clear;
- friendly;
- calm;
- direct;
- concise;
- respectful.

Avoid:

- slang;
- forced humor;
- exaggerated hype;
- technical jargon in primary play flows;
- blame-oriented errors.

### 39.2 Button labels

Prefer explicit verbs:

- Create Room
- Join Room
- Ready
- Start Match
- View Replay
- Rematch
- Resume Puzzle
- Clear Local Data

Avoid vague labels such as:

- Continue
- Submit
- Okay

when a more precise verb exists.

### 39.3 Error tone

Use:

```text
That cell is fixed.
```

Avoid:

```text
Invalid action!
```

### 39.4 Privacy tone

Use factual language:

```text
No accounts. No ads. No tracking.
```

Avoid absolute claims that exceed implementation:

```text
Completely anonymous and impossible to track.
```

### 39.5 Competitive tone

Use friendly competition language.

Avoid humiliation, taunting, or aggressive loss messaging.

---

## 40. Public content pages

### 40.1 About

Deferred until after the MVP.

Sections:

- product concept;
- multiplayer modes;
- privacy;
- replay;
- architecture;
- formal verification;
- accessibility;
- source code.

### 40.2 How to Play

Sections:

- Sudoku basics;
- selecting cells;
- entering values;
- notes;
- conflicts;
- hints;
- mode rules;
- reconnect;
- replay.

### 40.3 Mode pages

Deferred until the relevant modes are available.

Each mode page should include:

- plain-language overview;
- player count;
- key rules;
- screenshots or diagrams;
- accessibility notes;
- primary action.

### 40.4 Privacy page

Explain:

- essential cookies;
- local storage;
- Room data;
- Match events;
- replay retention;
- logs;
- data deletion;
- no analytics;
- contact.

### 40.5 Accessibility page

Explain:

- keyboard;
- screen reader;
- non-color indicators;
- reduced motion;
- zoom;
- touch targets;
- known limitations;
- contact for feedback.

---

## 41. SEO-facing visual design

Public pages should:

- prioritize readable content;
- use meaningful headings;
- include visible text, not image-only copy;
- avoid hidden SEO text;
- keep primary play action above fold;
- render well without client JavaScript.

Localized pages should maintain visual parity.

---

## 42. Admin design

Admin is functional, private, and minimal.

Required views:

- server health;
- Room lookup;
- Match details;
- replay deletion;
- puzzle retirement;
- Room termination.

Admin visual style may reuse product components but should prioritize:

- density;
- clarity;
- auditability;
- explicit destructive confirmation.

Do not expose admin links publicly.

---

## 43. Destructive actions

Destructive actions include:

- leave active Match;
- abandon Race;
- resign Duel;
- remove participant;
- block participant;
- delete replay;
- clear local data;
- retire puzzle;
- terminate Room.

Requirements:

- explicit label;
- consequence explanation;
- confirmation;
- keyboard-accessible;
- safe default focus;
- no color-only warning.

For low-risk reversible actions, avoid unnecessary confirmation.

---

## 44. Notifications and toasts

Use toasts sparingly for:

- copied Room code;
- preference saved;
- transient noncritical confirmation;
- replay link copied.

Do not use toasts for:

- important errors;
- turn changes;
- connection state;
- Match completion;
- destructive confirmations.

Important state belongs in persistent UI.

---

## 45. Dialogs and sheets

### 45.1 Desktop

Use modal dialogs for:

- confirmation;
- advanced settings;
- rule explanation;
- replay deletion.

### 45.2 Mobile

Use bottom sheets where they improve reachability.

### 45.3 Requirements

- focus trap;
- initial focus;
- Escape close where safe;
- visible close action;
- prevent background scroll;
- restore prior focus;
- descriptive title;
- no nested dialogs unless unavoidable.

---

## 46. Component states

Every interactive component must define:

- default;
- hover;
- focus;
- active;
- selected;
- disabled;
- loading;
- error;
- success;
- pending;
- read-only.

Do not leave browser default disabled styling if contrast is insufficient.

---

## 47. Shadcn-svelte usage

Use shadcn-svelte for primitives such as:

- dialog;
- sheet;
- dropdown menu;
- tooltip;
- tabs;
- switch;
- select;
- alert dialog;
- toast;
- command menu where useful.

Do not force the Sudoku board or mode-specific gameplay into generic table/card components.

Customize primitives through Ninefold semantic tokens.

Avoid importing large unused component sets.

---

## 48. Design QA matrix

Every major screen must be reviewed in:

### Viewports

- 320px
- 375px
- 430px
- 768px
- 1024px
- 1280px
- 1440px+

### Themes

- light
- dark
- system transition

### Locales

- English
- generated pseudo-locale

Add German, Albanian, and Turkish to the matrix when their catalogs enter scope.

### Input

- touch
- mouse
- keyboard
- screen reader

### States

- loading
- empty
- error
- disconnected
- recovering
- read-only
- reduced motion
- 200% zoom

---

## 49. Screen-level acceptance criteria

### 49.1 Homepage

- Create and Join are immediately visible.
- Product identity includes Sudoku.
- No account prompt.
- No tracking banner.
- Public content is readable without JS.
- Mobile layout has no horizontal scroll.
- Lighthouse target is feasible.

### 49.2 Lobby

- Room code is obvious.
- Participants and ready states are clear.
- Host is clear.
- Start-disabled reason is clear.
- Settings do not overwhelm.
- Keyboard navigation works.
- pseudo-localized expanded labels do not truncate.

### 49.3 Active board

- Board is dominant.
- Cells remain readable.
- Number pad is reachable.
- Connection state is visible.
- Pending state is distinct.
- Color is not sole cue.
- 200% zoom remains usable.
- Screen-reader labels are meaningful.

### 49.4 Results

- Outcome is clear.
- Rematch is primary.
- Replay is discoverable.
- Temporary names are handled respectfully.
- Integrity status is accurate.
- Share card does not expose solution.

### 49.5 Replay

- Controls are keyboard accessible.
- Timeline has text equivalents.
- Mobile Race replay remains usable.
- Integrity verification loads without blocking result.
- Hidden information remains hidden.

---

## 50. Design quality gates

A feature is design-complete only when:

- all states are designed;
- mobile and desktop layouts exist;
- keyboard behavior is defined;
- screen-reader semantics are defined;
- error and loading states exist;
- dark mode exists;
- localization expansion is tested;
- reduced motion is supported;
- touch targets pass;
- color-independent cues exist;
- content is finalized;
- interaction feedback is timely;
- performance impact is acceptable.

---

## 51. Explicit design prohibitions

Do not:

- use canvas for the main Sudoku board;
- rely on hover for required information;
- use red/green alone for state;
- shrink Sudoku cells below 24×24;
- shrink number-pad buttons or primary controls below 44×44;
- force account creation;
- show cookie consent without non-essential cookies;
- auto-play sound;
- auto-play video;
- force PWA installation;
- show Room participant names before join;
- display standalone puzzle-solution artifacts in replay metadata;
- block result display while replay loads;
- put marketing navigation inside active Match;
- use excessive confetti;
- animate every move heavily;
- use tiny low-contrast notes;
- truncate critical pseudo-localized or future German text;
- hide host or turn status in icons only;
- use generic “Submit” where a precise verb exists;
- put technical stack jargon in primary gameplay;
- expose private admin navigation;
- overstate cryptographic guarantees.

---

## 52. Implementation sequence for design

Recommended order:

1. Brand mark and token foundation
2. Light/dark color system
3. Typography and spacing
4. Homepage
5. Create/Join flow
6. Lobby
7. Semantic Sudoku board
8. Number pad and keyboard interaction
9. Co-op attribution and soft locks
10. Connection/recovery states
11. Results
12. Basic replay
13. Solo
14. Home, How to Play, Privacy, and Accessibility pages
15. MVP administration
16. Race after scope review
17. Duel after scope review
18. Daily Ninefold after scope review
19. additional locales and deferred public pages

Accessibility, keyboard behavior, responsive layout, and pseudo-localization are implemented continuously from the first interactive component; they are not a late refinement phase.

---

## 53. Final design directive

Ninefold should feel effortless to players even though its engineering is sophisticated.

The design succeeds when:

- people understand how to start without explanation;
- the board is clear and satisfying;
- multiplayer presence is visible but not distracting;
- privacy is obvious without being alarming;
- accessibility feels native;
- mobile play is comfortable;
- desktop play is fast;
- replay feels trustworthy;
- design supports, rather than advertises, the technical architecture.

When a design choice makes the product heavier, noisier, less accessible, less private, or harder to understand without materially improving play, remove or simplify it.
