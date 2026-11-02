------------------------------ MODULE ModeRouting ------------------------------
EXTENDS Naturals
CONSTANTS Accounts
VARIABLES desired, live, enabled, deps, pending, requestAccount, ownHome, compatible,
          marks, classifier, oversized, outcome, requestMode
vars == <<desired, live, enabled, deps, pending, requestAccount, ownHome, compatible,
          marks, classifier, oversized, outcome, requestMode>>
Init == /\ desired = [a \in Accounts |-> "standard"] /\ live = desired
        /\ enabled = TRUE /\ deps \in BOOLEAN /\ pending = FALSE /\ requestAccount \in Accounts
        /\ ownHome \in BOOLEAN /\ compatible \in BOOLEAN /\ marks \in BOOLEAN
        /\ classifier \in BOOLEAN /\ oversized \in BOOLEAN
        /\ outcome = "none" /\ requestMode = "none"
Configure(a) == /\ ~pending /\ ~ownHome /\ compatible
                /\ desired' = [desired EXCEPT ![a] = IF @ = "standard" THEN "tui" ELSE "standard"]
                /\ UNCHANGED <<live, enabled, deps, pending, requestAccount, ownHome, compatible,
                               marks, classifier, oversized, outcome, requestMode>>
Restart == /\ ~pending /\ live' = desired
           /\ UNCHANGED <<desired, enabled, deps, pending, requestAccount, ownHome, compatible,
                          marks, classifier, oversized, outcome, requestMode>>
Ask(a) == /\ ~pending /\ outcome = "none" /\ pending' = TRUE
          /\ requestAccount' = a /\ requestMode' = live[a]
          /\ UNCHANGED <<desired, live, enabled, deps, ownHome, compatible,
                         marks, classifier, oversized, outcome>>
Answer == /\ pending /\ pending' = FALSE
          /\ outcome' = IF requestMode = "standard" THEN "standard"
                         ELSE IF ~deps \/ ownHome \/ ~compatible \/ marks \/ classifier \/ oversized
                              THEN "refused" ELSE "tui"
          /\ UNCHANGED <<desired, live, enabled, deps, requestAccount, ownHome, compatible,
                         marks, classifier, oversized, requestMode>>
Next == Restart \/ Answer \/ (\E a \in Accounts: Configure(a) \/ Ask(a))
Spec == Init /\ [][Next]_vars /\ WF_vars(Answer)
TypeOK == /\ desired \in [Accounts -> {"standard", "tui"}] /\ live \in [Accounts -> {"standard", "tui"}]
          /\ outcome \in {"none", "standard", "tui", "refused"}
NoSilentStandardFallback == requestMode = "tui" => outcome # "standard"
NoUnsafeLaunch == outcome = "tui" => deps /\ ~ownHome /\ compatible /\ ~marks /\ ~classifier /\ ~oversized
RequestUsesSnapshot == pending => requestMode = live[requestAccount]
PendingEventuallyAnswered == pending ~> ~pending
=============================================================================
