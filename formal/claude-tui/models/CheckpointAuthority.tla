------------------------ MODULE CheckpointAuthority ------------------------
EXTENDS Naturals, FiniteSets
CONSTANTS Scopes, Prefixes
VARIABLES store, health, request, chosen, mode, dirty, effects, cancelled, done,
          results, summary, busy, badWrite, oldReader
vars == <<store, health, request, chosen, mode, dirty, effects, cancelled, done,
          results, summary, busy, badWrite, oldReader>>
Entries == [scope: Scopes, prefix: Prefixes, kind: {"text", "tool"}, model: {"M", "N"}, instructions: {"I", "J"}]
NoEntry == [scope |-> CHOOSE s \in Scopes: TRUE, prefix |-> CHOOSE p \in Prefixes: TRUE, kind |-> "none", model |-> "none", instructions |-> "none"]
Init == /\ store = {} /\ health = "valid" /\ request = NoEntry
        /\ chosen = NoEntry /\ mode = "none" /\ dirty = FALSE
        /\ effects = FALSE /\ cancelled = FALSE /\ done = FALSE
        /\ results = {} /\ summary = FALSE /\ busy = FALSE /\ badWrite = FALSE /\ oldReader = FALSE
Save(e) == /\ health = "valid" /\ Cardinality(store) < 1 /\ request = NoEntry
           /\ store' = {e}
           /\ UNCHANGED <<health, request, chosen, mode, dirty, effects, cancelled, done,
                          results, summary, busy, badWrite, oldReader>>
Corrupt == /\ health = "valid" /\ health' = "invalid"
           /\ UNCHANGED <<store, request, chosen, mode, dirty, effects, cancelled, done,
                          results, summary, busy, badWrite, oldReader>>
Ask(e, rs, side, live, legacy) ==
    /\ request = NoEntry /\ request' = e /\ results' = rs
    /\ summary' = side /\ busy' = live /\ oldReader' = legacy
    /\ UNCHANGED <<store, health, chosen, mode, dirty, effects, cancelled, done, badWrite>>
Exact(e) == /\ e.scope = request.scope /\ e.prefix = request.prefix
            /\ e.model = request.model /\ e.instructions = request.instructions
            /\ e.kind = request.kind
CompleteResults(e) == IF e.kind = "tool" THEN results = {1, 2} ELSE results = {}
Resume(e) == /\ request # NoEntry /\ mode = "none" /\ health = "valid"
             /\ e \in store /\ Exact(e) /\ CompleteResults(e)
             /\ ~(oldReader /\ e.kind = "tool")
             /\ chosen' = e
             /\ mode' = IF summary \/ busy \/ e.kind = "tool" THEN "fork" ELSE "inplace"
             /\ dirty' = ~summary
             /\ UNCHANGED <<store, health, request, effects, cancelled, done,
                            results, summary, busy, badWrite, oldReader>>
Rebuild == /\ request # NoEntry /\ mode = "none" /\ chosen = NoEntry
           /\ mode' = "rebuild" /\ dirty' = ~summary
           /\ UNCHANGED <<store, health, request, chosen, effects, cancelled, done,
                          results, summary, busy, badWrite, oldReader>>
Act == /\ mode # "none" /\ ~done /\ ~cancelled /\ ~effects /\ ~summary
       /\ effects' = TRUE
       /\ UNCHANGED <<store, health, request, chosen, mode, dirty, cancelled, done,
                      results, summary, busy, badWrite, oldReader>>
Cancel == /\ mode # "none" /\ ~done /\ ~cancelled /\ cancelled' = TRUE
          /\ UNCHANGED <<store, health, request, chosen, mode, dirty, effects, done,
                         results, summary, busy, badWrite, oldReader>>
Finish == /\ mode # "none" /\ ~done /\ done' = TRUE
          /\ dirty' = IF cancelled THEN effects /\ ~summary ELSE FALSE
          /\ UNCHANGED <<store, health, request, chosen, mode, effects, cancelled,
                         results, summary, busy, badWrite, oldReader>>
Next == Corrupt \/ Rebuild \/ Act \/ Cancel \/ Finish
        \/ (\E e \in Entries: Save(e) \/ Resume(e))
        \/ (\E e \in Entries, rs \in SUBSET {1, 2}, side, live, legacy \in BOOLEAN: Ask(e, rs, side, live, legacy))
Spec == Init /\ [][Next]_vars /\ WF_vars(Finish)
TypeOK == /\ store \subseteq Entries /\ health \in {"valid", "invalid"}
          /\ request \in Entries \cup {NoEntry} /\ chosen \in Entries \cup {NoEntry}
          /\ mode \in {"none", "fork", "inplace", "rebuild"} /\ results \subseteq {1, 2}
ResumeAuthorized == chosen # NoEntry => Exact(chosen) /\ CompleteResults(chosen)
SummaryIsolated == summary => ~dirty /\ ~effects /\ mode # "inplace"
ToolCheckpointDowngradeSafe == (oldReader /\ request # NoEntry /\ request.kind = "tool") => chosen = NoEntry
NoDirtyRewindAfterEffects == (done /\ cancelled /\ effects) => dirty
NeverWriteUnreadStore == ~badWrite
FiniteTurnSettles == (mode # "none") ~> done
=============================================================================
