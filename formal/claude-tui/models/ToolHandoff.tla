----------------------------- MODULE ToolHandoff -----------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Calls, DropDuplicateGuard, AllowReadOnlyTool
VARIABLES callback, rows, shown, results, consumed, wal, recovery, repaired,
          repairClaimed, closed, readonly, held, deliveries, badDispatch
vars == <<callback, rows, shown, results, consumed, wal, recovery, repaired,
          repairClaimed, closed, readonly, held, deliveries, badDispatch>>
Init == /\ callback = {} /\ rows = {} /\ shown = {} /\ results = {} /\ consumed = {}
        /\ wal = "valid" /\ recovery = 0 /\ repaired = FALSE /\ repairClaimed = FALSE
        /\ closed = FALSE /\ readonly \in BOOLEAN /\ held \in BOOLEAN
        /\ deliveries = [c \in Calls |-> 0] /\ badDispatch = FALSE
Callback(c) == /\ ~closed /\ c \notin callback /\ wal = "valid"
               /\ (~readonly \/ AllowReadOnlyTool) /\ ~held
               /\ callback' = callback \cup {c}
               /\ UNCHANGED <<rows, shown, results, consumed, wal, recovery, repaired,
                              repairClaimed, closed, readonly, held, deliveries, badDispatch>>
Transcript(c) == /\ ~closed /\ c \notin rows /\ rows' = rows \cup {c}
                 /\ UNCHANGED <<callback, shown, results, consumed, wal, recovery, repaired,
                                repairClaimed, closed, readonly, held, deliveries, badDispatch>>
Dispatch(c) == /\ ~closed /\ ~held /\ (~readonly \/ AllowReadOnlyTool)
               /\ c \in callback \cup rows
               /\ (DropDuplicateGuard \/ c \notin shown) /\ deliveries[c] < 2
               /\ shown' = shown \cup {c}
               /\ deliveries' = [deliveries EXCEPT ![c] = @ + 1]
               /\ badDispatch' = (badDispatch \/ readonly)
               /\ UNCHANGED <<callback, rows, results, consumed, wal, recovery, repaired,
                              repairClaimed, closed, readonly, held>>
Result(c) == /\ ~closed /\ c \in shown /\ c \notin results /\ wal = "valid"
             /\ results' = results \cup {c}
             /\ UNCHANGED <<callback, rows, shown, consumed, wal, recovery, repaired,
                            repairClaimed, closed, readonly, held, deliveries, badDispatch>>
Consume(c) == /\ ~closed /\ c \in results \cap callback /\ c \notin consumed
              /\ consumed' = consumed \cup {c}
              /\ UNCHANGED <<callback, rows, shown, results, wal, recovery, repaired,
                             repairClaimed, closed, readonly, held, deliveries, badDispatch>>
Corrupt == /\ wal = "valid" /\ wal' = "invalid"
           /\ UNCHANGED <<callback, rows, shown, results, consumed, recovery, repaired,
                          repairClaimed, closed, readonly, held, deliveries, badDispatch>>
Repair == /\ ~closed /\ ~repaired /\ rows = Calls /\ callback = {}
          /\ repaired' = TRUE
          /\ UNCHANGED <<callback, rows, shown, results, consumed, wal, recovery,
                         repairClaimed, closed, readonly, held, deliveries, badDispatch>>
ClaimRepair == /\ repaired /\ ~repairClaimed /\ results = Calls /\ wal = "valid"
               /\ repairClaimed' = TRUE
               /\ UNCHANGED <<callback, rows, shown, results, consumed, wal, recovery, repaired,
                              closed, readonly, held, deliveries, badDispatch>>
Recover == /\ closed /\ wal = "valid" /\ shown = Calls /\ recovery < 2
           /\ recovery' = recovery + 1 /\ readonly' = TRUE
           /\ UNCHANGED <<callback, rows, shown, results, consumed, wal, repaired,
                          repairClaimed, closed, held, deliveries, badDispatch>>
Close == /\ ~closed /\ closed' = TRUE
         /\ UNCHANGED <<callback, rows, shown, results, consumed, wal, recovery, repaired,
                        repairClaimed, readonly, held, deliveries, badDispatch>>
Next == Corrupt \/ Repair \/ ClaimRepair \/ Recover \/ Close
        \/ (\E c \in Calls: Callback(c) \/ Transcript(c) \/ Dispatch(c) \/ Result(c) \/ Consume(c))
Spec == Init /\ [][Next]_vars /\ WF_vars(Close)
TypeOK == /\ callback \subseteq Calls /\ rows \subseteq Calls /\ shown \subseteq Calls
          /\ results \subseteq Calls /\ consumed \subseteq Calls
          /\ deliveries \in [Calls -> 0..2] /\ recovery \in 0..2 /\ wal \in {"valid", "invalid"}
OnlyShownResults == results \subseteq shown
OnlyKnownConsumption == consumed \subseteq results \cap callback
DispatchAtMostOnce == \A c \in Calls: deliveries[c] <= 1
ReadOnlyNeverDispatches == ~badDispatch
RepairRequiresCompleteResults == repairClaimed => repaired /\ results = Calls
RecoveryBounded == recovery <= 2
EventuallyCloses == <> closed
=============================================================================
