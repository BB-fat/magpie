-------------------------- MODULE SessionReservation --------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Requests, RecheckLiveAtReserve
VARIABLES pc, reservation, writers, rowAdvanced
vars == <<pc, reservation, writers, rowAdvanced>>
Init == /\ pc = [r \in Requests |-> "scan"] /\ reservation = "none"
        /\ writers = {} /\ rowAdvanced = FALSE
\* sessionRun snapshots live runs separately from claiming b.claimedSessions.
Scan(r) == /\ pc[r] = "scan" /\ writers = {}
           /\ pc' = [pc EXCEPT ![r] = "reserve"]
           /\ UNCHANGED <<reservation, writers, rowAdvanced>>
Reserve(r) == /\ pc[r] = "reserve" /\ reservation = "none"
              /\ (~RecheckLiveAtReserve \/ writers = {})
              /\ reservation' = r /\ pc' = [pc EXCEPT ![r] = "checkrow"]
              /\ UNCHANGED <<writers, rowAdvanced>>
CheckRow(r) == /\ pc[r] = "checkrow" /\ ~rowAdvanced
               /\ pc' = [pc EXCEPT ![r] = "start"]
               /\ UNCHANGED <<reservation, writers, rowAdvanced>>
Publish(r) == /\ pc[r] = "start" /\ pc' = [pc EXCEPT ![r] = "published"]
              /\ writers' = writers \cup {r}
              /\ UNCHANGED <<reservation, rowAdvanced>>
Release(r) == /\ pc[r] = "published" /\ reservation = r
              /\ reservation' = "none" /\ pc' = [pc EXCEPT ![r] = "running"]
              /\ UNCHANGED <<writers, rowAdvanced>>
Advance == /\ writers # {} /\ ~rowAdvanced /\ rowAdvanced' = TRUE
           /\ UNCHANGED <<pc, reservation, writers>>
Reap(r) == /\ pc[r] = "running" /\ r \in writers
           /\ pc' = [pc EXCEPT ![r] = "done"] /\ writers' = writers \ {r}
           /\ UNCHANGED <<reservation, rowAdvanced>>
Next == Advance \/ (\E r \in Requests: Scan(r) \/ Reserve(r) \/ CheckRow(r) \/ Publish(r) \/ Release(r) \/ Reap(r))
Spec == Init /\ [][Next]_vars
TypeOK == /\ pc \in [Requests -> {"scan", "reserve", "checkrow", "start", "published", "running", "done"}]
          /\ reservation \in Requests \cup {"none"} /\ writers \subseteq Requests
SingleNativeWriter == Cardinality(writers) <= 1
=============================================================================
