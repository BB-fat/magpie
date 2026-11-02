-------------------------- MODULE ReplayIdentity --------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Requests, SameCaller, ScopeKeys, RecheckAfterWait, HasHeader
VARIABLES pc, cached, cachedBy, gate, starts, receivedFrom
vars == <<pc, cached, cachedBy, gate, starts, receivedFrom>>
Caller(r) == IF SameCaller THEN "A" ELSE r
Key(r) == IF ScopeKeys THEN Caller(r) ELSE "owner+outer+body"
Keys == {Key(r): r \in Requests}
Init == /\ pc = [r \in Requests |-> "lookup"] /\ cached = {}
        /\ cachedBy = [k \in Keys |-> "none"] /\ gate = "none"
        /\ starts = [r \in Requests |-> 0]
        /\ receivedFrom = [r \in Requests |-> "none"]
\* Completed results are checked before enterLiteTurn waits.
Lookup(r) == /\ pc[r] = "lookup"
             /\ pc' = [pc EXCEPT ![r] = IF Key(r) \in cached THEN "done" ELSE "enter"]
             /\ receivedFrom' = [receivedFrom EXCEPT ![r] = IF Key(r) \in cached THEN cachedBy[Key(r)] ELSE "none"]
             /\ UNCHANGED <<cached, cachedBy, gate, starts>>
Enter(r) == /\ pc[r] = "enter" /\ (~HasHeader \/ gate = "none")
            /\ LET hit == RecheckAfterWait /\ Key(r) \in cached IN
               /\ pc' = [pc EXCEPT ![r] = IF hit THEN "done" ELSE "running"]
               /\ starts' = [starts EXCEPT ![r] = IF hit THEN @ ELSE @ + 1]
               /\ receivedFrom' = [receivedFrom EXCEPT ![r] = IF hit THEN cachedBy[Key(r)] ELSE "none"]
               /\ gate' = IF HasHeader /\ ~hit THEN r ELSE gate
            /\ UNCHANGED <<cached, cachedBy>>
Complete(r) == /\ pc[r] = "running"
               /\ pc' = [pc EXCEPT ![r] = "done"]
               /\ cached' = cached \cup {Key(r)}
               /\ cachedBy' = [cachedBy EXCEPT ![Key(r)] = Caller(r)]
               /\ gate' = IF gate = r THEN "none" ELSE gate
               /\ receivedFrom' = [receivedFrom EXCEPT ![r] = Caller(r)]
               /\ UNCHANGED starts
Next == \E r \in Requests: Lookup(r) \/ Enter(r) \/ Complete(r)
Spec == Init /\ [][Next]_vars /\ (\A r \in Requests: WF_vars(Lookup(r)) /\ WF_vars(Enter(r)) /\ WF_vars(Complete(r)))
TypeOK == /\ pc \in [Requests -> {"lookup", "enter", "running", "done"}]
          /\ cached \subseteq Keys /\ starts \in [Requests -> 0..1]
CallerIsolation == \A r \in Requests: receivedFrom[r] \in {"none", Caller(r)}
OneExecutionPerRetry == \A r, s \in Requests: (r # s /\ Caller(r) = Caller(s)) => starts[r] + starts[s] <= 1
OneConcurrentRetry == Cardinality({r \in Requests: pc[r] = "running"}) <= 1
AllRequestsFinish == <> (\A r \in Requests: pc[r] = "done")
=============================================================================
