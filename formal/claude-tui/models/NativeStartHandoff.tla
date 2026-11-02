-------------------------- MODULE NativeStartHandoff --------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Runs
VARIABLES pc, tracked, closing, closed, child, room, released, stopped, badStart
vars == <<pc, tracked, closing, closed, child, room, released, stopped, badStart>>
Init == /\ pc = [r \in Runs |-> "new"] /\ tracked = {} /\ closing = {}
        /\ closed = {} /\ child = {} /\ room = {} /\ released = {}
        /\ stopped = FALSE /\ badStart = FALSE
Publish(r) == /\ ~stopped /\ pc[r] = "new"
              /\ pc' = [pc EXCEPT ![r] = "room"] /\ tracked' = tracked \cup {r}
              /\ UNCHANGED <<closing, closed, child, room, released, stopped, badStart>>
Admit(r) == /\ pc[r] = "room" /\ r \notin closing /\ r \notin closed
            /\ room' = room \cup {r} /\ pc' = [pc EXCEPT ![r] = "starting"]
            /\ UNCHANGED <<tracked, closing, closed, child, released, stopped, badStart>>
Resolve(r) == /\ pc[r] = "starting"
              /\ pc' = [pc EXCEPT ![r] = "published"] /\ child' = child \cup {r}
              /\ badStart' = (badStart \/ r \in closed)
              /\ UNCHANGED <<tracked, closing, closed, room, released, stopped>>
FailStart(r) == /\ pc[r] = "starting" /\ pc' = [pc EXCEPT ![r] = "done"]
                /\ room' = room \ {r} /\ released' = released \cup {r}
                /\ tracked' = tracked \ {r} /\ closed' = closed \cup {r}
                /\ UNCHANGED <<closing, child, stopped, badStart>>
Shutdown == /\ ~stopped /\ stopped' = TRUE /\ closing' = closing \cup tracked
            /\ closed' = closed \cup {r \in tracked: pc[r] = "room"}
            /\ tracked' = tracked \ {r \in tracked: pc[r] = "room"}
            /\ UNCHANGED <<pc, child, room, released, badStart>>
Refuse(r) == /\ pc[r] = "room" /\ r \in closed
             /\ pc' = [pc EXCEPT ![r] = "done"]
             /\ UNCHANGED <<tracked, closing, closed, child, room, released, stopped, badStart>>
Reap(r) == /\ r \in child /\ pc[r] = "published"
           /\ child' = child \ {r} /\ room' = room \ {r}
           /\ released' = released \cup {r} /\ tracked' = tracked \ {r}
           /\ closed' = closed \cup {r} /\ pc' = [pc EXCEPT ![r] = "done"]
           /\ UNCHANGED <<closing, stopped, badStart>>
Next == Shutdown \/ (\E r \in Runs: Publish(r) \/ Admit(r) \/ Resolve(r) \/ FailStart(r) \/ Refuse(r) \/ Reap(r))
Spec == Init /\ [][Next]_vars /\ (\A r \in Runs: WF_vars(Resolve(r)) /\ WF_vars(Reap(r)))
TypeOK == /\ pc \in [Runs -> {"new", "room", "starting", "published", "done"}]
          /\ tracked \subseteq Runs /\ closed \subseteq Runs /\ child \subseteq Runs
          /\ room \subseteq Runs /\ released \subseteq Runs
NoStartAfterFinish == ~badStart
EveryChildTracked == child \subseteq tracked
NoReleaseBeforeReap == child \cap released = {}
EveryChildHasRoom == child \subseteq room
ShutdownEventuallyReaps == stopped ~> (child = {})
=============================================================================
