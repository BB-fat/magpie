---------------------------- MODULE ProcessCleanup ----------------------------
EXTENDS Naturals
CONSTANTS Pids
VARIABLES actual, seen, alive, killedUnrelated, room, released, wrapper, tmux,
          dirSafe, rssKnown, overRSS, stopRequested
vars == <<actual, seen, alive, killedUnrelated, room, released, wrapper, tmux,
          dirSafe, rssKnown, overRSS, stopRequested>>
Init == /\ actual = [p \in Pids |-> 1] /\ seen = [p \in Pids |-> 0]
        /\ alive = Pids /\ killedUnrelated = FALSE /\ room = FALSE /\ released = FALSE
        /\ wrapper = "new" /\ tmux = FALSE /\ dirSafe \in BOOLEAN
        /\ rssKnown \in BOOLEAN /\ overRSS \in BOOLEAN /\ stopRequested = FALSE
Start == /\ wrapper = "new" /\ dirSafe
         /\ wrapper' = "live" /\ tmux' = TRUE /\ room' = TRUE
         /\ UNCHANGED <<actual, seen, alive, killedUnrelated, released,
                        dirSafe, rssKnown, overRSS, stopRequested>>
Sample(p) == /\ wrapper = "live" /\ p \in alive /\ seen[p] = 0 /\ actual[p] = 1
             /\ seen' = [seen EXCEPT ![p] = actual[p]]
             /\ UNCHANGED <<actual, alive, killedUnrelated, room, released, wrapper, tmux,
                            dirSafe, rssKnown, overRSS, stopRequested>>
Reuse(p) == /\ p \in alive /\ actual[p] = 1
            /\ actual' = [actual EXCEPT ![p] = 2]
            /\ UNCHANGED <<seen, alive, killedUnrelated, room, released, wrapper, tmux,
                           dirSafe, rssKnown, overRSS, stopRequested>>
RequestStop == /\ wrapper = "live" /\ ~stopRequested /\ stopRequested' = TRUE
               /\ UNCHANGED <<actual, seen, alive, killedUnrelated, room, released, wrapper, tmux,
                              dirSafe, rssKnown, overRSS>>
RSSStop == /\ wrapper = "live" /\ rssKnown /\ overRSS /\ ~stopRequested
           /\ stopRequested' = TRUE
           /\ UNCHANGED <<actual, seen, alive, killedUnrelated, room, released, wrapper, tmux,
                          dirSafe, rssKnown, overRSS>>
ReapWrapper == /\ wrapper = "live" /\ stopRequested
               /\ wrapper' = "reaped" /\ tmux' = FALSE
               /\ UNCHANGED <<actual, seen, alive, killedUnrelated, room, released,
                              dirSafe, rssKnown, overRSS, stopRequested>>
Sweep(p) == /\ wrapper = "reaped" /\ p \in alive /\ seen[p] # 0
            /\ seen[p] = actual[p]
            /\ alive' = alive \ {p}
            /\ killedUnrelated' = (killedUnrelated \/ actual[p] = 2)
            /\ UNCHANGED <<actual, seen, room, released, wrapper, tmux,
                           dirSafe, rssKnown, overRSS, stopRequested>>
Release == /\ wrapper = "reaped" /\ ~released
           /\ room' = FALSE /\ released' = TRUE
           /\ UNCHANGED <<actual, seen, alive, killedUnrelated, wrapper, tmux,
                          dirSafe, rssKnown, overRSS, stopRequested>>
Next == Start \/ RequestStop \/ RSSStop \/ ReapWrapper \/ Release
        \/ (\E p \in Pids: Sample(p) \/ Reuse(p) \/ Sweep(p))
Spec == Init /\ [][Next]_vars /\ WF_vars(ReapWrapper) /\ WF_vars(Release)
TypeOK == /\ actual \in [Pids -> 1..2] /\ seen \in [Pids -> 0..2]
          /\ alive \subseteq Pids /\ wrapper \in {"new", "live", "reaped"}
NoPrematureRoomRelease == wrapper = "live" => room /\ ~released
NoUnsafeDirectoryStart == ~dirSafe => wrapper = "new"
NoReusedPidKilled == ~killedUnrelated
RoomEventuallyReleased == (stopRequested /\ wrapper # "new") ~> released
=============================================================================
