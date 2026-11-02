---------------------------- MODULE DurableRecovery ----------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Claims
VARIABLES pc, file, lock, resultHash, claimedHash, fence, childAlive, recoveryStarts,
          readonlyStarts, emittedTools, accepted
vars == <<pc, file, lock, resultHash, claimedHash, fence, childAlive, recoveryStarts,
          readonlyStarts, emittedTools, accepted>>
Init == /\ pc = [c \in Claims |-> "new"] /\ file = "registered" /\ lock = "none"
        /\ resultHash \in {"result-A", "result-B"} /\ claimedHash = "none"
        /\ fence \in BOOLEAN /\ childAlive \in BOOLEAN
        /\ recoveryStarts = 0 /\ readonlyStarts = 0 /\ emittedTools = 0 /\ accepted = {}
Acquire(c) == /\ pc[c] = "new" /\ lock = "none" /\ fence /\ ~childAlive
              /\ pc' = [pc EXCEPT ![c] = "check"] /\ lock' = c
              /\ UNCHANGED <<file, resultHash, claimedHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools, accepted>>
\* Both append and claimOrphanRecovery hold the OS file lock and validate
\* every checksum/transition. A partial write is unknown, not an empty WAL.
Claim(c) == /\ pc[c] = "check" /\ lock = c
            /\ file \in {"registered", "verified", "claimed"}
            /\ (file = "registered" \/ claimedHash = resultHash)
            /\ file' = IF file = "claimed" THEN "retry" ELSE "claimed"
            /\ claimedHash' = resultHash /\ accepted' = accepted \cup {c}
            /\ pc' = [pc EXCEPT ![c] = "launch"] /\ lock' = "none"
            /\ UNCHANGED <<resultHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools>>
TornWrite(c) == /\ pc[c] = "check" /\ lock = c
                /\ file' = "corrupt" /\ pc' = [pc EXCEPT ![c] = "refused"] /\ lock' = "none"
                /\ UNCHANGED <<resultHash, claimedHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools, accepted>>
Refuse(c) == /\ pc[c] \in {"new", "check"}
             /\ pc' = [pc EXCEPT ![c] = "refused"]
             /\ lock' = IF lock = c THEN "none" ELSE lock
             /\ UNCHANGED <<file, resultHash, claimedHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools, accepted>>
Launch(c) == /\ pc[c] = "launch" /\ fence /\ ~childAlive
             /\ pc' = [pc EXCEPT ![c] = "done"]
             /\ recoveryStarts' = recoveryStarts + 1 /\ readonlyStarts' = readonlyStarts + 1
             /\ UNCHANGED <<file, lock, resultHash, claimedHash, fence, childAlive, emittedTools, accepted>>
Crash(c) == /\ pc[c] = "launch" /\ pc' = [pc EXCEPT ![c] = "done"]
            /\ UNCHANGED <<file, lock, resultHash, claimedHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools, accepted>>
Verify == /\ file = "registered" /\ lock = "none"
          /\ file' = "verified" /\ claimedHash' = resultHash
          /\ UNCHANGED <<pc, lock, resultHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools, accepted>>
ChangeResult == /\ resultHash' = IF resultHash = "result-A" THEN "result-B" ELSE "result-A"
                /\ UNCHANGED <<pc, file, lock, claimedHash, fence, childAlive, recoveryStarts, readonlyStarts, emittedTools, accepted>>
Next == Verify \/ ChangeResult \/ (\E c \in Claims: Acquire(c) \/ Claim(c) \/ TornWrite(c) \/ Refuse(c) \/ Launch(c) \/ Crash(c))
Spec == Init /\ [][Next]_vars
TypeOK == /\ file \in {"registered", "verified", "claimed", "retry", "corrupt"}
          /\ pc \in [Claims -> {"new", "check", "launch", "refused", "done"}]
          /\ accepted \subseteq Claims
AtMostTwoClaims == Cardinality(accepted) <= 2
OnlyReadOnlyRecovery == recoveryStarts = readonlyStarts /\ emittedTools = 0
NoRecoveryWhileNativeAlive == childAlive => recoveryStarts = 0
NoRecoveryWithoutFence == ~fence => recoveryStarts = 0
=============================================================================
