# Table Walkthrough: Multi-Epoch, Multi-User Scenario

LSA = last seen accumulator

| Epoch | Event                          | Accumulator             | Total Shares  | U1 Shares | U1 LSA | U1 Claimable                                 | U2 Shares | U2 LSA | U2 Claimable              |
|-------|--------------------------------|-------------------------|---------------|-----------|--------|----------------------------------------------|-----------|--------|---------------------------|
| **1** | User1 endorses with 40 shares  | 0.0                     | 40            | 40        | 0.0    | (0 - 0) * 40 = 0 DYM                         | -         | -      | -                         |
| **2** | +100 DYM unlocked              | 0 + (100 / 40) = 2.5    | 40            | 40        | 0.0    | (2.5 - 0) * 40 = 100 DYM                     | -         | -      | -                         |
| 2     | User2 endorses with 60 shares  | 2.5                     | 40 + 60 = 100 | 40        | 0.0    | 100 DYM                                      | 60        | 2.5    | (2.5 - 2.5) * 60 = 0 DYM  |
| **3** | +100 DYM unlocked              | 2.5 + (100 / 100) = 3.5 | 100           | 40        | 0.0    | (3.5 - 0) * 40 = 140 DYM                     | 60        | 2.5    | (3.5 - 2.5) * 60 = 60 DYM |
| 3     | User1 claims                   | 3.5                     | 100           | 40        | 3.5    | ✅ Claimed: 140 DYM, (3.5 - 3.5) * 40 = 0 DYM | 60        | 2.5    | (3.5 - 2.5) * 60 = 60 DYM |
| 3     | User2 un-endorses (auto-claim) | 3.5                     | 100 - 60 = 40 | 40        | 3.5    | (3.5 - 3.5) * 40 = 0 DYM                     | -         | -      | ✅ Claimed: 60 DYM         |
| **4** | +100 DYM                       | 3.5 + (100 / 40) = 6.0  | 40            | 40        | 3.5    | (6 - 3.5) * 40 = 100 DYM                     | -         | -      | -                         |
| 4     | User1 un-endorses (auto-claim) | 6.0                     | 40 - 40 = 0   | -         | -      | ✅ Claimed: 100 DYM                           | -         | -      | -                         |
| 4     | User2 re-endorses w 100 shares | 6.0                     | 0 + 60 = 60   | -         | -      | -                                            | 60        | 6.0    | (6 - 6) * 60 = 0 DYM      |
| **5** | +100 DYM unlocked              | 6 + (100 / 60) = 7.7    | 60            | -         | -      | -                                            | 60        | 6.0    | (7.7 - 6) * 60 = 100 DYM  |