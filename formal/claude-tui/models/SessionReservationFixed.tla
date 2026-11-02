-------------------------- MODULE SessionReservationFixed --------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Requests
VARIABLES pc, reservation, writers, advanced
vars == <<pc, reservation, writers, advanced>>
Init == /\ pc = [r \in Requests |-> "reserve"] /\ reservation = "none"
        /\ writers = {} /\ advanced = FALSE
Reserve(r) == /\ pc[r] = "reserve" /\ reservation = "none"
              /\ reservation' = r /\ pc' = [pc EXCEPT ![r] = "inspect"]
              /\ UNCHANGED <<writers, advanced>>
Inspect(r) == /\ pc[r] = "inspect" /\ reservation = r
              /\ pc' = [pc EXCEPT ![r] = IF writers = {} THEN "checkrow" ELSE "fork"]
              /\ reservation' = IF writers = {} THEN reservation ELSE "none"
              /\ UNCHANGED <<writers, advanced>>
CheckRow(r) == /\ pc[r] = "checkrow" /\ ~advanced
               /\ pc' = [pc EXCEPT ![r] = "start"]
               /\ UNCHANGED <<reservation, writers, advanced>>
Publish(r) == /\ pc[r] = "start" /\ reservation = r
              /\ pc' = [pc EXCEPT ![r] = "published"] /\ writers' = writers \cup {r}
              /\ UNCHANGED <<reservation, advanced>>
Release(r) == /\ pc[r] = "published" /\ reservation = r
              /\ reservation' = "none" /\ pc' = [pc EXCEPT ![r] = "running"]
              /\ UNCHANGED <<writers, advanced>>
Advance == /\ writers # {} /\ ~advanced /\ advanced' = TRUE
           /\ UNCHANGED <<pc, reservation, writers>>
Reap(r) == /\ pc[r] = "running" /\ r \in writers
           /\ pc' = [pc EXCEPT ![r] = "done"] /\ writers' = writers \ {r}
           /\ UNCHANGED <<reservation, advanced>>
Next == Advance \/ (\E r \in Requests: Reserve(r) \/ Inspect(r) \/ CheckRow(r) \/ Publish(r) \/ Release(r) \/ Reap(r))
Spec == Init /\ [][Next]_vars
TypeOK == /\ pc \in [Requests -> {"reserve", "inspect", "checkrow", "start", "published", "running", "done", "fork"}]
          /\ reservation \in Requests \cup {"none"} /\ writers \subseteq Requests
SingleNativeWriter == Cardinality(writers) <= 1
=============================================================================
