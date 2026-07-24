---- MODULE RoomLifecycle ----

EXTENDS Integers, Sequences, FiniteSets, TLC

CONSTANTS
  Participants,
  Host,
  RequestIDs,
  MaxRequests

ASSUME Host \in Participants

VARIABLES
  roomState,
  host,
  players,
  ready,
  locked,
  countdown,
  matchId,
  version,
  usedRequests,
  timerGen

vars == << roomState, host, players, ready, locked, countdown, matchId, version, usedRequests, timerGen >>

Init ==
  /\ roomState = "Empty"
  /\ host = "none"
  /\ players = {}
  /\ ready = {}
  /\ locked = FALSE
  /\ countdown = "none"
  /\ matchId = "none"
  /\ version = 0
  /\ usedRequests = {}
  /\ timerGen = 0

CreateRoom(req) ==
  /\ roomState = "Empty"
  /\ req \notin usedRequests
  /\ roomState' = "Lobby"
  /\ host' = Host
  /\ players' = {Host}
  /\ ready' = {}
  /\ locked' = FALSE
  /\ countdown' = "none"
  /\ matchId' = "none"
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

Join(req, p) ==
  /\ roomState = "Lobby"
  /\ req \notin usedRequests
  /\ p \notin players
  /\ locked = FALSE
  /\ players' = players \union {p}
  /\ ready' = ready
  /\ host' = host
  /\ roomState' = roomState
  /\ locked' = locked
  /\ countdown' = countdown
  /\ matchId' = matchId
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

SetReady(req, p) ==
  /\ roomState = "Lobby"
  /\ req \notin usedRequests
  /\ p \in players
  /\ ready' = ready \union {p}
  /\ roomState' = roomState
  /\ host' = host
  /\ players' = players
  /\ locked' = locked
  /\ countdown' = countdown
  /\ matchId' = matchId
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

ChangeSettings(req, p) ==
  /\ roomState = "Lobby"
  /\ req \notin usedRequests
  /\ p = host
  /\ ready' = {}
  /\ roomState' = roomState
  /\ host' = host
  /\ players' = players
  /\ locked' = locked
  /\ countdown' = countdown
  /\ matchId' = matchId
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

StartCountdown(req, p) ==
  /\ roomState = "Lobby"
  /\ req \notin usedRequests
  /\ p = host
  /\ Cardinality(players) >= 1
  /\ ready = players
  /\ matchId = "none"
  /\ countdown' = "active"
  /\ timerGen' = timerGen + 1
  /\ roomState' = "Countdown"
  /\ locked' = TRUE
  /\ ready' = ready
  /\ host' = host
  /\ players' = players
  /\ matchId' = "prepared"
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}

CancelCountdown(req, p) ==
  /\ roomState = "Countdown"
  /\ req \notin usedRequests
  /\ p = host
  /\ roomState' = "Lobby"
  /\ countdown' = "none"
  /\ matchId' = "none"
  /\ locked' = FALSE
  /\ ready' = {}
  /\ host' = host
  /\ players' = players
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

Activate(gen) ==
  /\ roomState = "Countdown"
  /\ countdown = "active"
  /\ gen = timerGen
  /\ roomState' = "InMatch"
  /\ countdown' = "none"
  /\ matchId' = "active"
  /\ version' = version + 1
  /\ host' = host
  /\ players' = players
  /\ ready' = ready
  /\ locked' = locked
  /\ usedRequests' = usedRequests
  /\ timerGen' = timerGen

StaleActivate(gen) ==
  /\ roomState # "Countdown"
  /\ gen # timerGen
  /\ UNCHANGED vars

Leave(req, p) ==
  /\ roomState \in {"Lobby", "Results"}
  /\ req \notin usedRequests
  /\ p \in players
  /\ players' = players \ {p}
  /\ ready' = ready \ {p}
  /\ IF p = host THEN
        IF players' = {} THEN host' = "none"
        ELSE host' = CHOOSE q \in players' : TRUE
     ELSE host' = host
  /\ roomState' = roomState
  /\ locked' = locked
  /\ countdown' = countdown
  /\ matchId' = matchId
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

TransferHost(req, p, q) ==
  /\ roomState = "Lobby"
  /\ req \notin usedRequests
  /\ p = host
  /\ q \in players
  /\ q # p
  /\ host' = q
  /\ roomState' = roomState
  /\ players' = players
  /\ ready' = ready
  /\ locked' = locked
  /\ countdown' = countdown
  /\ matchId' = matchId
  /\ version' = version + 1
  /\ usedRequests' = usedRequests \union {req}
  /\ timerGen' = timerGen

Next ==
  \/ (\E req \in RequestIDs : CreateRoom(req))
  \/ (\E req \in RequestIDs : \E p \in Participants \ players : Join(req, p))
  \/ (\E req \in RequestIDs : \E p \in players : SetReady(req, p))
  \/ (\E req \in RequestIDs : ChangeSettings(req, Host))
  \/ (\E req \in RequestIDs : StartCountdown(req, Host))
  \/ (\E req \in RequestIDs : CancelCountdown(req, Host))
  \/ (\E req \in RequestIDs : \E p \in players : Leave(req, p))
  \/ (\E req \in RequestIDs : \E p \in players : TransferHost(req, Host, p))
  \/ (\E gen \in 1..timerGen : Activate(gen))
  \/ (\E gen \in 1..timerGen : StaleActivate(gen))

TypeInvariant ==
  /\ roomState \in {"Empty", "Lobby", "Countdown", "InMatch", "Results"}
  /\ host \in Participants \union {"none"}
  /\ players \subseteq Participants
  /\ ready \subseteq players
  /\ locked \in BOOLEAN
  /\ countdown \in {"none", "active"}
  /\ matchId \in {"none", "prepared", "active"}
  /\ version \in Nat
  /\ usedRequests \subseteq RequestIDs
  /\ timerGen \in Nat

AtMostOneHost ==
  (host # "none") => Cardinality({host}) = 1

AtMostOneActiveMatch ==
  roomState \in {"InMatch"} => matchId = "active"

SettingsLockedDuringCountdown ==
  (roomState \in {"Countdown", "InMatch"}) => locked = TRUE

AllReadyBeforeCountdown ==
  (roomState = "Countdown") => ready = players

NoDuplicateRequestEffects ==
  Cardinality(usedRequests) <= MaxRequests

Inv ==
  /\ TypeInvariant
  /\ AtMostOneHost
  /\ AtMostOneActiveMatch
  /\ SettingsLockedDuringCountdown
  /\ AllReadyBeforeCountdown
  /\ NoDuplicateRequestEffects

Spec == Init /\ [][Next]_vars /\ WF_vars(Activate(timerGen))

====
