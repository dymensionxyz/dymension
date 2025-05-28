# Endorsement Test Scenarios - Execution Plan

This document outlines the test cases to be implemented for the lazy claiming feature in the sponsorship module, based on the scenarios described in `endorsements_test_scenario.md`.

The existing `TestEndorsements` function in `endorsements_test.go` covers the main multi-user, multi-epoch scenario. The following additional scenarios will be implemented as new test functions:

1.  **TestScenario_StakeUnstakeAndClaim**:
    *   Corresponds to: "Example scenario – user stakes 100 more and then unstakes 50; a new epoch starts; finally, user claims"
    *   Steps:
        *   Initial state: User has 100 shares, LSA = 6.0, AUR = 0. Global Accumulator (GA) = 7.0 (after +100 DYM unlocked).
        *   User stakes an additional 100 (total shares become 200).
            *   Verify AUR becomes 100 DYM ((7-6)\*100).
            *   Verify LSA becomes 7.0.
            *   Verify claimable is 0.
        *   User unstakes 50 (total shares become 150).
            *   Verify AUR remains 100 DYM.
            *   Verify LSA remains 7.0.
            *   Verify claimable is 0.
        *   New epoch: +100 DYM.
            *   Calculate new GA: 7.0 + (100 / 150) = 7.666...
            *   Verify claimable becomes 100 DYM ((7.666... - 7.0) \* 150).
            *   Verify AUR remains 100 DYM.
        *   User claims.
            *   Verify 200 DYM are claimed (100 from AUR + 100 from current epoch).
            *   Verify AUR becomes 0.
            *   Verify LSA updates to current GA (7.666...).
            *   Verify claimable becomes 0.

2.  **TestScenario_UnstakeAll**:
    *   Corresponds to: "Example scenario – user unstakes all coins"
    *   Steps:
        *   Initial state: User has 100 shares, LSA = 6.0, AUR = 0. GA = 7.0 (after +100 DYM unlocked).
        *   User unstakes all 100 shares.
            *   Verify AUR becomes 100 DYM ((7-6)\*100).
            *   Verify user shares become 0.
            *   Verify LSA becomes 7.0.
            *   Verify claimable is 0.
        *   User claims.
            *   Verify 100 DYM are claimed.
            *   Verify AUR becomes 0.

3.  **TestScenario_EndorseWithNoPriorVote**:
    *   Corresponds to: "Example scenario – user endorses when they did not have the vote before"
    *   Steps:
        *   Initial state: GA = 7.0 (after +100 DYM unlocked). User has no prior vote or shares.
        *   User endorses with 100 shares.
            *   Verify user shares become 100.
            *   Verify LSA becomes 7.0.
            *   Verify AUR is 0.
            *   Verify claimable is 0.

4.  **TestScenario_EndorseEmptyEndorsement**:
    *   Corresponds to: "Example scenario – user casts for the *empty* endorsement (i.e., no one endorsed to it yet)"
    *   Steps:
        *   Initial state: First epoch. GA = 0. Total shares for the endorsement = 0.
        *   User endorses with 100 shares.
            *   Verify user shares become 100.
            *   Verify LSA becomes 0.
            *   Verify AUR is 0.
            *   Verify claimable is 0.
            *   Verify total shares for the endorsement become 100.
            *   Verify GA remains 0.
        *   New epoch: +100 DYM.
            *   Calculate new GA: 0 + (100 / 100) = 1.0.
            *   Verify claimable becomes 100 DYM ((1.0 - 0) \* 100).
            *   Verify AUR remains 0.
