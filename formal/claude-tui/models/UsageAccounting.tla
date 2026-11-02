---------------------------- MODULE UsageAccounting ----------------------------
EXTENDS Naturals, FiniteSets
CONSTANT AddSnapshotsInsteadOfReplacing
VARIABLES seen, readByMessage, writeByMessage, hourByMessage, outputByMessage, billed
vars == <<seen, readByMessage, writeByMessage, hourByMessage, outputByMessage, billed>>
Messages == {"message1", "message2"}
Rows == { [msg |-> "message1", read |-> 2, write |-> 1, hour |-> 1, output |-> 1],
          [msg |-> "message1", read |-> 2, write |-> 1, hour |-> 1, output |-> 2],
          [msg |-> "message2", read |-> 1, write |-> 2, hour |-> 0, output |-> 1] }
Max(a,b) == IF a > b THEN a ELSE b
MaxSet(s) == CHOOSE n \in s: \A v \in s: n >= v
Total(f) == f["message1"] + f["message2"]
Init == /\ seen = {} /\ readByMessage = [m \in Messages |-> 0]
        /\ writeByMessage = [m \in Messages |-> 0] /\ hourByMessage = [m \in Messages |-> 0]
        /\ outputByMessage = [m \in Messages |-> 0] /\ billed = 0
Observe(row) == /\ billed <= 12
                /\ seen' = seen \cup {row}
                /\ readByMessage' = [readByMessage EXCEPT ![row.msg] = Max(@,row.read)]
                /\ writeByMessage' = [writeByMessage EXCEPT ![row.msg] = Max(@,row.write)]
                /\ hourByMessage' = [hourByMessage EXCEPT ![row.msg] = Max(@,row.hour)]
                /\ outputByMessage' = [outputByMessage EXCEPT ![row.msg] = Max(@,row.output)]
                /\ LET snapshot == Total(readByMessage') + Total(writeByMessage') + Total(outputByMessage') IN
                   billed' = IF AddSnapshotsInsteadOfReplacing THEN billed + snapshot ELSE snapshot
Next == \E row \in Rows: Observe(row)
Spec == Init /\ [][Next]_vars
TypeOK == /\ seen \subseteq Rows /\ readByMessage \in [Messages -> 0..2]
          /\ writeByMessage \in [Messages -> 0..2] /\ hourByMessage \in [Messages -> 0..2]
          /\ outputByMessage \in [Messages -> 0..2]
CountsMatchMessageMaxima == \A m \in Messages:
    /\ readByMessage[m] = MaxSet({row.read: row \in {r \in seen: r.msg = m}} \cup {0})
    /\ writeByMessage[m] = MaxSet({row.write: row \in {r \in seen: r.msg = m}} \cup {0})
    /\ outputByMessage[m] = MaxSet({row.output: row \in {r \in seen: r.msg = m}} \cup {0})
OneHourIsSubsetOfWrites == \A m \in Messages: hourByMessage[m] <= writeByMessage[m]
NoRepeatedSnapshotBilling == billed = Total(readByMessage) + Total(writeByMessage) + Total(outputByMessage)
=============================================================================
