----------------------------- MODULE ServePool -----------------------------
EXTENDS Integers
CONSTANTS RecheckCancellation, HTTPCompletionFence
VARIABLES turn, state, ready, drained, replied, gen, timerGen, cancelAt, badCancel, completionOwner
vars == <<turn, state, ready, drained, replied, gen, timerGen, cancelAt, badCancel, completionOwner>>
Init == /\ turn = 1 /\ state = "active"
        /\ ready = FALSE /\ drained = FALSE /\ replied = FALSE
        /\ gen = 0 /\ timerGen = -1 /\ cancelAt = 0 /\ badCancel = FALSE /\ completionOwner = "open"
Ready == /\ state = "active" /\ ~ready /\ ready' = TRUE
         /\ UNCHANGED <<turn, state, drained, replied, gen, timerGen, cancelAt, badCancel, completionOwner>>
Drain == /\ state = "active" /\ ~drained /\ drained' = TRUE
         /\ UNCHANGED <<turn, state, ready, replied, gen, timerGen, cancelAt, badCancel, completionOwner>>
Reply == /\ state = "active" /\ ~replied /\ replied' = TRUE
         /\ (~HTTPCompletionFence \/ completionOwner # "cancel")
         /\ completionOwner' = "complete"
         /\ UNCHANGED <<turn, state, ready, drained, gen, timerGen, cancelAt, badCancel>>
Idle == /\ state = "active" /\ ready /\ drained /\ replied
        /\ state' = "idle" /\ gen' = gen + 1
        /\ UNCHANGED <<turn, ready, drained, replied, timerGen, cancelAt, badCancel, completionOwner>>
Arm == /\ state = "idle" /\ timerGen # gen /\ timerGen' = gen
       /\ UNCHANGED <<turn, state, ready, drained, replied, gen, cancelAt, badCancel, completionOwner>>
Claim == /\ state = "idle" /\ turn < 2
         /\ state' = "active" /\ turn' = turn + 1 /\ gen' = gen + 1
         /\ ready' = FALSE /\ drained' = FALSE /\ replied' = FALSE /\ completionOwner' = "open"
         /\ UNCHANGED <<timerGen, cancelAt, badCancel>>
Expire == /\ state = "idle" /\ timerGen = gen
          /\ state' = "ended" /\ gen' = gen + 1
          /\ UNCHANGED <<turn, ready, drained, replied, timerGen, cancelAt, badCancel, completionOwner>>
\* mine/waiting are read under mu; the following abort is outside it.
CancelInspect == /\ turn = 1 /\ state = "active" /\ cancelAt = 0
                 /\ (~HTTPCompletionFence \/ completionOwner = "open")
                 /\ completionOwner' = "cancel"
                 /\ cancelAt' = turn
                 /\ UNCHANGED <<turn, state, ready, drained, replied, gen, timerGen, badCancel>>
CancelAct == /\ cancelAt = 1 /\ state # "ended"
             /\ LET kill == ~RecheckCancellation \/ (turn = cancelAt /\ state = "active") IN
                /\ state' = IF kill THEN "ended" ELSE state
                /\ badCancel' = (badCancel \/ (kill /\ turn # cancelAt))
             /\ cancelAt' = 2
             /\ UNCHANGED <<turn, ready, drained, replied, gen, timerGen, completionOwner>>
Next == Ready \/ Drain \/ Reply \/ Idle \/ Arm \/ Claim \/ Expire \/ CancelInspect \/ CancelAct
Spec == Init /\ [][Next]_vars
        /\ WF_vars(Ready) /\ WF_vars(Drain) /\ WF_vars(Reply) /\ WF_vars(Idle)
        /\ WF_vars(Arm) /\ WF_vars(Expire) /\ WF_vars(CancelAct)
TypeOK == /\ turn \in 1..2 /\ state \in {"active", "idle", "ended"}
          /\ completionOwner \in {"open", "complete", "cancel"}
          /\ gen \in 0..5 /\ timerGen \in -1..4 /\ cancelAt \in 0..2
IdleOnlyAfterWholeReply == state = "idle" => ready /\ drained /\ replied
OldCancellationCannotKillNewTurn == ~badCancel
EventuallyEnds == <> (state = "ended")
=============================================================================
