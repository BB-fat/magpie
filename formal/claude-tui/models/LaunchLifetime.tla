-------------------------- MODULE LaunchLifetime --------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Runs, FenceClosedStart
VARIABLES pc, registered, closed, child, freed, stopped, badStart
vars == <<pc, registered, closed, child, freed, stopped, badStart>>
Init == /\ pc = [r \in Runs |-> "new"]
        /\ registered = {} /\ closed = {} /\ child = {} /\ freed = {}
        /\ stopped = FALSE /\ badStart = FALSE
\* startInteractiveWithPromptImages publishes before Room and StartTree.
Publish(r) == /\ ~stopped /\ pc[r] = "new"
              /\ pc' = [pc EXCEPT ![r] = "room"]
              /\ registered' = registered \cup {r}
              /\ UNCHANGED <<closed, child, freed, stopped, badStart>>
RoomGranted(r) == /\ pc[r] = "room"
                  /\ pc' = [pc EXCEPT ![r] = "start"]
                  /\ UNCHANGED <<registered, closed, child, freed, stopped, badStart>>
Start(r) == /\ pc[r] = "start"
            /\ (~FenceClosedStart \/ r \notin closed)
            /\ pc' = [pc EXCEPT ![r] = "running"]
            /\ child' = child \cup {r}
            /\ badStart' = (badStart \/ r \in closed)
            /\ UNCHANGED <<registered, closed, freed, stopped>>
\* With no tree, finish/removeRun removes the run before launch resumes.
Shutdown == /\ ~stopped /\ stopped' = TRUE
            /\ closed' = closed \cup registered
            /\ registered' = registered \cap child
            /\ UNCHANGED <<pc, child, freed, badStart>>
Reap(r) == /\ r \in child
           /\ child' = child \ {r} /\ registered' = registered \ {r}
           /\ freed' = freed \cup {r} /\ closed' = closed \cup {r}
           /\ pc' = [pc EXCEPT ![r] = "done"]
           /\ UNCHANGED <<stopped, badStart>>
RefuseClosed(r) == /\ FenceClosedStart /\ r \in closed
                   /\ pc[r] \in {"room", "start"}
                   /\ pc' = [pc EXCEPT ![r] = "done"]
                   /\ UNCHANGED <<registered, closed, child, freed, stopped, badStart>>
Next == Shutdown \/ (\E r \in Runs: Publish(r) \/ RoomGranted(r) \/ Start(r) \/ Reap(r) \/ RefuseClosed(r))
Spec == Init /\ [][Next]_vars /\ (\A r \in Runs: WF_vars(Reap(r)))
TypeOK == /\ pc \in [Runs -> {"new", "room", "start", "running", "done"}]
          /\ registered \subseteq Runs /\ closed \subseteq Runs /\ child \subseteq Runs /\ freed \subseteq Runs
NoStartAfterFinish == ~badStart
EveryChildTracked == child \subseteq registered
NoFreeBeforeReap == child \cap freed = {}
ShutdownEventuallyReaps == stopped ~> (child = {})
=============================================================================
