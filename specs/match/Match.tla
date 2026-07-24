---- MODULE Match ----
EXTENDS Integers, Sequences, TLC, FiniteSets

CONSTANTS Cells,          \* set of cell indices
          Participants,   \* set of participant identifiers
          Clues,          \* subset of Cells that are fixed clues
          RequestIDs,     \* pool of unique request identifiers
          RecoveryGenerations

\* Simple deterministic solution mapping for the abstract model.
Solution == [c \in Cells |-> c]

VARIABLES matchState, placed, mistakes, contributions, usedRequestIDs, events,
          recoveryGeneration, recoveryIntervals

vars == <<matchState, placed, mistakes, contributions, usedRequestIDs, events,
          recoveryGeneration, recoveryIntervals>>

ValidStates == {"Prepared", "Active", "RecoveryPending", "Completed", "Cancelled"}

EventType ==
  UNION {
    [type: {"MatchStarted"}],
    [type: {"ValuePlaced"}, cell: Cells, value: 1..9, req: RequestIDs],
    [type: {"MatchCompleted"}],
    [type: {"MatchEnteredRecovery"}, generation: RecoveryGenerations],
    [type: {"MatchRecovered"}, generation: RecoveryGenerations],
    [type: {"MatchCancelled"}, generation: RecoveryGenerations]
  }

TypeInvariant ==
  /\ matchState \in ValidStates
  /\ placed \in [Cells -> 0..9]
  /\ mistakes \in [Participants -> Nat]
  /\ contributions \in [Participants -> Nat]
  /\ usedRequestIDs \subseteq RequestIDs
  /\ events \in Seq(EventType)
  /\ recoveryGeneration \in RecoveryGenerations
  /\ recoveryIntervals \in Nat

Init ==
  /\ matchState = "Prepared"
  /\ placed = [c \in Cells |-> IF c \in Clues THEN Solution[c] ELSE 0]
  /\ mistakes = [p \in Participants |-> 0]
  /\ contributions = [p \in Participants |-> 0]
  /\ usedRequestIDs = {}
  /\ events = <<>>
  /\ recoveryGeneration = 0
  /\ recoveryIntervals = 0

Activate ==
  /\ matchState = "Prepared"
  /\ matchState' = "Active"
  /\ events' = Append(events, [type |-> "MatchStarted"])
  /\ UNCHANGED <<placed, mistakes, contributions, usedRequestIDs,
                 recoveryGeneration, recoveryIntervals>>

PlaceValue(p, c, v, req) ==
  /\ matchState = "Active"
  /\ c \in Cells \ Clues
  /\ v \in 1..9
  /\ req \in RequestIDs \ usedRequestIDs
  /\ usedRequestIDs' = usedRequestIDs \union {req}
  /\ placed' = [placed EXCEPT ![c] = v]
  /\ IF v = Solution[c]
     THEN /\ contributions' = [contributions EXCEPT ![p] = @ + 1]
          /\ mistakes' = mistakes
     ELSE /\ mistakes' = [mistakes EXCEPT ![p] = @ + 1]
          /\ contributions' = contributions
  /\ events' = Append(events, [type |-> "ValuePlaced", cell |-> c, value |-> v, req |-> req])
  /\ UNCHANGED <<matchState, recoveryGeneration, recoveryIntervals>>

AutoComplete ==
  /\ matchState = "Active"
  /\ \A c \in Cells \ Clues : placed[c] # 0 /\ placed[c] = Solution[c]
  /\ matchState' = "Completed"
  /\ events' = Append(events, [type |-> "MatchCompleted"])
  /\ UNCHANGED <<placed, mistakes, contributions, usedRequestIDs,
                 recoveryGeneration, recoveryIntervals>>

EnterRecovery ==
  /\ matchState = "Active"
  /\ recoveryGeneration + 1 \in RecoveryGenerations
  /\ matchState' = "RecoveryPending"
  /\ recoveryGeneration' = recoveryGeneration + 1
  /\ events' = Append(events,
       [type |-> "MatchEnteredRecovery", generation |-> recoveryGeneration + 1])
  /\ UNCHANGED <<placed, mistakes, contributions, usedRequestIDs, recoveryIntervals>>

Recover(generation) ==
  /\ matchState = "RecoveryPending"
  /\ generation = recoveryGeneration
  /\ matchState' = "Active"
  /\ recoveryIntervals' = recoveryIntervals + 1
  /\ events' = Append(events, [type |-> "MatchRecovered", generation |-> generation])
  /\ UNCHANGED <<placed, mistakes, contributions, usedRequestIDs, recoveryGeneration>>

RecoveryTimeout(generation) ==
  /\ matchState = "RecoveryPending"
  /\ generation = recoveryGeneration
  /\ matchState' = "Cancelled"
  /\ events' = Append(events, [type |-> "MatchCancelled", generation |-> generation])
  /\ UNCHANGED <<placed, mistakes, contributions, usedRequestIDs,
                 recoveryGeneration, recoveryIntervals>>

Next ==
  \/ Activate
  \/ AutoComplete
  \/ EnterRecovery
  \/ \E generation \in RecoveryGenerations :
       Recover(generation) \/ RecoveryTimeout(generation)
  \/ \E p \in Participants, c \in Cells, v \in 1..9, req \in RequestIDs :
       PlaceValue(p, c, v, req)

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

HasReq(e) == "req" \in DOMAIN e

Range(seq) == {seq[i] : i \in 1..Len(seq)}

RequestIDsInEvents ==
  {e.req : e \in {evt \in Range(events) : HasReq(evt)}}

PlaceEventCount ==
  Cardinality({evt \in Range(events) : HasReq(evt)})

UniqueRequests ==
  Cardinality(RequestIDsInEvents) = PlaceEventCount

(* Safety: every request ID produces at most one gameplay event. *)
UniqueRequestIDs == [](UniqueRequests)

(* Safety: once Completed the match never leaves Completed. *)
CompletedPersistent == [](matchState = "Completed" => [](matchState = "Completed"))

(* Safety: recovery requires the current generation and never revives completion. *)
CompletedNeverRecovers ==
  []((matchState = "Completed") => ~ENABLED Recover(recoveryGeneration))

(* Liveness: a recovery wait eventually resumes or cancels. *)
RecoveryEventuallyResolves ==
  [](matchState = "RecoveryPending" => <>(matchState # "RecoveryPending"))

====
