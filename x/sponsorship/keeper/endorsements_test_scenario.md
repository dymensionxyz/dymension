# Test Scenarios

LSA = last seen accumulator

AUR = Accumulated Unclaimed Rewards

The core idea is that whenever a user's shares change, the rewards accrued up to that point (with the old share
quantity) are calculated and "banked." The user's (LSA) is then updated to the current global accumulator value. Future
rewards will accrue based on the new share quantity and the updated LSA. This allows to seamlessly update user’s shares
in case of voting power change, un-endorsement, and update of the existing vote.

| Event                          | Accumulator             | Total Shares    | Unlocked Coins | U1 Shares | U1 LSA | U1 Claimable                                      | U1 AUR                | U2 Shares | U2 LSA | U2 Claimable              | U2 AUR                |
|--------------------------------|-------------------------|-----------------|----------------|-----------|--------|---------------------------------------------------|-----------------------|-----------|--------|---------------------------|-----------------------|
| User1 endorses with 40 shares  | 0.0                     | 40              | 0              | 40        | 0.0    | (0 - 0) * 40 = 0 DYM                              | 0                     | -         | -      | -                         | -                     |
| **New epoch**: +100 DYM        | 0 + (100 / 40) = 2.5    | 40              | 100            | 40        | 0.0    | (2.5 - 0) * 40 = 100 DYM                          | 0                     | -         | -      | -                         | -                     |
| User2 endorses with 60 shares  | 2.5                     | 40 + 60 = 100   | 100            | 40        | 0.0    | 100 DYM                                           | 0                     | 60        | 2.5    | (2.5 - 2.5) * 60 = 0 DYM  | 0                     |
| **New epoch**: +100 DYM        | 2.5 + (100 / 100) = 3.5 | 100             | 200            | 40        | 0.0    | (3.5 - 0) * 40 = 140 DYM                          | 0                     | 60        | 2.5    | (3.5 - 2.5) * 60 = 60 DYM | 0                     |
| User1 claims                   | 3.5                     | 100             | 60             | 40        | 3.5    | ✅ Claimed: 140 DYM       (3.5 - 3.5) * 40 = 0 DYM | 0                     | 60        | 2.5    | (3.5 - 2.5) * 60 = 60 DYM | 0                     |
| User2 un-endorses              | 3.5                     | 100 - 60 = 40   | 0              | 40        | 3.5    | (3.5 - 3.5) * 40 = 0 DYM                          | 0                     | 0         | 3.5    | (3.5 - 3.5) * 0 = 0 DYM   | 60 DYM                |
| **New epoch**: +100 DYM        | 3.5 + (100 / 40) = 6.0  | 40              | 100            | 40        | 3.5    | (6 - 3.5) * 40 = 100 DYM                          | 0                     | 0         | 0      | 0                         | 60 DYM                |
| User1 un-endorses              | 6.0                     | 40 - 40 = 0     | 0              | 0         | 0      | (6 - 6) * 0 = 0 DYM                               | 100 DYM               | 0         | 0      | 0                         | 60 DYM                |
| User1 claims                   | 6.0                     | 0               | 0              | 0         | 0      | 0                                                 | ✅ Claimed: 100 DYM  0 | 0         | 0      | 0                         | 60 DYM                |
| User2 re-endorses w 100 shares | 6.0                     | 0 + 100 = 100   | 0              | 0         | 0      | 0                                                 | 0                     | 100       | 6.0    | (6 - 6) * 100 = 0 DYM     | 60 DYM                |
| **New epoch**: +100 DYM        | 6 + (100 / 100) = 7.0   | 100             | 100            | 0         | 0      | 0                                                 | 0                     | 100       | 6.0    | (7 - 6) * 100 = 100 DYM   | 60  DYM               |
| User2 stakes 100 DYM           | 7.0                     | 100 + 100 = 200 | 100            | 0         | 0      | 0                                                 | 0                     | 200       | 7.0    | (7 - 7) * 200 = 0 DYM     | 160 DYM               |
| User2 claims                   | 7.0                     | 200             | 100            | 0         | 0      | 0                                                 | 0                     | 200       | 7.0    | (7 - 7) * 200 = 0 DYM     | ✅ Claimed: 160 DYM  0 |

Example scenario – user stakes 100 more and then unstakes 50; a new epoch starts; finally, user claims:

| Event                   | Global Accumulator (GA)  | Total Shares    | Unlocked Coins | User Shares | User LSA | User Claimable                                         | User AUR                 |
|-------------------------|--------------------------|-----------------|----------------|-------------|----------|--------------------------------------------------------|--------------------------|
| +100 DYM unlocked       | 7.0                      | 100             | 100            | 100         | 6.0      | (7 - 6) * 100 = 100 DYM                                | 0                        |
| User stakes 100         | 7.0                      | 100 + 100 = 200 | 100            | 200         | 7.0      | (7 - 7) * 100 = 0 DYM                                  | (7 - 6) * 100 = 100 DYM  |
| User unstakes 50        | 7.0                      | 200 - 50 = 150  | 100            | 150         | 7.0      | (7 - 7) * 100 = 0 DYM                                  | 100 DYM                  |
| **New epoch**: +100 DYM | 7.0 + (100 / 150) = 7,67 | 150             | 200            | 150         | 7.0      | (7.67 - 7) * 150 = 100 DYM                             | 100 DYM                  |
| **New epoch**: +100 DYM | 7,67                     | 150             | 200            | 150         | 7,67     | ✅ Claimed: 100 DYM         (7.67 - 7,67) * 150 = 0 DYM | ✅ Claimed: 100 DYM 0 DYM |

Example scenario – user unstakes all coins:

| Event             | Global Accumulator (GA) | Total Shares  | Unlocked Coins | User Shares (S) | User LSA | User Claimable (C)      | User AUR                |
|-------------------|-------------------------|---------------|----------------|-----------------|----------|-------------------------|-------------------------|
| +100 DYM unlocked | 7.0                     | 100           | 100            | 100             | 6.0      | (7 - 6) * 100 = 100 DYM | 0                       |
| User unstakes 100 | 7.0                     | 100 - 100 = 0 | 100            | 0               | 7.0      | (7 - 7) * 0 = 0 DYM     | (7 - 6) * 100 = 100 DYM |

Example scenario – user endorses when they did not have the vote before:

| Event             | Global Accumulator | Total Shares    | Unlocked Coins | User Shares (S) | User LSA | User Claimable (C)    | User AUR            |
|-------------------|--------------------|-----------------|----------------|-----------------|----------|-----------------------|---------------------|
| +100 DYM unlocked | 7.0                | 100             | 100            | -               | -        | -                     | 0                   |
| User endorses 100 | 7.0                | 100 + 100 = 200 | 100            | 100             | 7.0      | (7 - 7) * 100 = 0 DYM | (7 - 0) * 0 = 0 DYM |

Example scenario – user casts for the *empty* endorsement (i.e., no one endorsed to it yet)

| Event             | Global Accumulator | Total Shares | Unlocked Coins | User Shares | User LSA | User Claimable        | User AUR            |
|-------------------|--------------------|--------------|----------------|-------------|----------|-----------------------|---------------------|
| First epoch       | 0                  | 0            | 0              | -           | -        | -                     | -                   |
| User endorses 100 | 0                  | 100          | 0              | 100         | 0        | (0 - 0) * 100 = 0 DYM | (0 - 0) * 0 = 0 DYM |