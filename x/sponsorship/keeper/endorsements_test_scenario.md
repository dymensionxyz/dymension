# Test Scenarios

LSA = last seen accumulator

AUR = Accumulated Unclaimed Rewards

The core idea is that whenever a user's shares change, the rewards accrued up to that point (with the old share
quantity) are calculated and "banked." The user's (LSA) is then updated to the current global accumulator value. Future
rewards will accrue based on the new share quantity and the updated LSA. This allows to seamlessly update user’s shares
in case of voting power change, un-endorsement, and update of the existing vote.

<table>
  <tr>
    <th>Event</th>
    <th>Accumulator</th>
    <th>Total Shares</th>
    <th>Unlocked Coins</th>
    <th>U1 Shares</th>
    <th>U1 LSA</th>
    <th>U1 Claimable</th>
    <th>U1 AUR</th>
    <th>U2 Shares</th>
    <th>U2 LSA</th>
    <th>U2 Claimable</th>
    <th>U2 AUR</th>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: distributing rewards to the empty endorsement results in noop</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user casts for the empty endorsement (i.e., no one endorsed to it yet)</td>
  </tr>
  <tr>
    <td>First epoch</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
  </tr>
  <tr>
    <td>User1 endorses with 40 shares</td>
    <td>0.0</td>
    <td>40</td>
    <td>0</td>
    <td>40</td>
    <td>0.0</td>
    <td>(0 - 0) * 40 = 0 DYM</td>
    <td>0</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: new epoch begins, user's claimable balance increases</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>0 + (100 / 40) = 2.5</td>
    <td>40</td>
    <td>100</td>
    <td>40</td>
    <td>0.0</td>
    <td>(2.5 - 0) * 40 = 100 DYM</td>
    <td>0</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user endorses when they did not have the vote before</td>
  </tr>
  <tr>
    <td>User2 endorses with 60 shares</td>
    <td>2.5</td>
    <td>40 + 60 = 100</td>
    <td>100</td>
    <td>40</td>
    <td>0.0</td>
    <td>100 DYM</td>
    <td>0</td>
    <td>60</td>
    <td>2.5</td>
    <td>(2.5 - 2.5) * 60 = 0 DYM</td>
    <td>0</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: new epoch begins, user's claimable balances increase</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>2.5 + (100 / 100) = 3.5</td>
    <td>100</td>
    <td>200</td>
    <td>40</td>
    <td>0.0</td>
    <td>(3.5 - 0) * 40 = 140 DYM</td>
    <td>0</td>
    <td>60</td>
    <td>2.5</td>
    <td>(3.5 - 2.5) * 60 = 60 DYM</td>
    <td>0</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user claims and keeps endorsement</td>
  </tr>
  <tr>
    <td>User1 claims</td>
    <td>3.5</td>
    <td>100</td>
    <td>60</td>
    <td>40</td>
    <td>3.5</td>
    <td>✅ Claimed: 140 DYM; (3.5 - 3.5) * 40 = 0 DYM</td>
    <td>0</td>
    <td>60</td>
    <td>2.5</td>
    <td>(3.5 - 2.5) * 60 = 60 DYM</td>
    <td>0</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user un-endorses without claiming; their rewards are persisted</td>
  </tr>
  <tr>
    <td>User2 un-endorses</td>
    <td>3.5</td>
    <td>100 - 60 = 40</td>
    <td>60</td>
    <td>40</td>
    <td>3.5</td>
    <td>(3.5 - 3.5) * 40 = 0 DYM</td>
    <td>0</td>
    <td>0</td>
    <td>3.5</td>
    <td>(3.5 - 3.5) * 0 = 0 DYM</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>3.5 + (100 / 40) = 6.0</td>
    <td>40</td>
    <td>160</td>
    <td>40</td>
    <td>3.5</td>
    <td>(6 - 3.5) * 40 = 100 DYM</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user un-endorses and claims after</td>
  </tr>
  <tr>
    <td>User1 un-endorses</td>
    <td>6.0</td>
    <td>40 - 40 = 0</td>
    <td>160</td>
    <td>0</td>
    <td>0</td>
    <td>(6 - 6) * 0 = 0 DYM</td>
    <td>100 DYM</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td>User1 claims</td>
    <td>6.0</td>
    <td>0</td>
    <td>60</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>✅ Claimed: 100 DYM; 0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user returns; their accumulated rewards are still persisted</td>
  </tr>
  <tr>
    <td>User2 re-endorses w 100 shares</td>
    <td>6.0</td>
    <td>0 + 100 = 100</td>
    <td>60</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>100</td>
    <td>6.0</td>
    <td>(6 - 6) * 100 = 0 DYM</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user has both accumulated rewards and claimable rewards</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>6 + (100 / 100) = 7.0</td>
    <td>100</td>
    <td>160</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>100</td>
    <td>6.0</td>
    <td>(7 - 6) * 100 = 100 DYM</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user stakes, unstakes, and claims</td>
  </tr>
  <tr>
    <td>User2 stakes 200 DYM</td>
    <td>7.0</td>
    <td>100 + 200 = 300</td>
    <td>160</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>300</td>
    <td>7.0</td>
    <td>(7 - 7) * 300 = 0 DYM</td>
    <td>160 DYM</td>
  </tr>
  <tr>
    <td>User2 unstakes 100 DYM</td>
    <td>7.0</td>
    <td>300 - 100 = 200</td>
    <td>160</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>200</td>
    <td>7.0</td>
    <td>(7 - 7) * 200 = 0 DYM</td>
    <td>160 DYM</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user updates the existing vote</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>7 + (100 / 200) = 7.5</td>
    <td>200</td>
    <td>260</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>200</td>
    <td>7.0</td>
    <td>(7.5 - 7) * 200 = 100 DYM</td>
    <td>160 DYM</td>
  </tr>
  <tr>
    <td>User2 removes 100 shares</td>
    <td>7.5</td>
    <td>200 - 100 = 100</td>
    <td>260</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>100</td>
    <td>7.5</td>
    <td>(7.5 - 7.5) * 100 = 0 DYM</td>
    <td>260 DYM</td>
  </tr>
  <tr>
    <td colspan="12" style="background-color: #fffae6;">Scenario: user unstakes all (their vote is removed)</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>7 + (100 / 100) = 8.5</td>
    <td>100</td>
    <td>360</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>100</td>
    <td>7.5</td>
    <td>(8.5 - 7.5) * 100 = 100 DYM</td>
    <td>260 DYM</td>
  </tr>
  <tr>
    <td>User2 unstakes all</td>
    <td>8.5</td>
    <td>0</td>
    <td>360</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>8.5</td>
    <td>(8.5 - 8.5) * 0 = 0 DYM</td>
    <td>360 DYM</td>
  </tr>
  <tr>
    <td>User2 claims</td>
    <td>8.5</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>8.5</td>
    <td>0 DYM</td>
    <td>✅ Claimed: 360 DYM</td>
  </tr>
</table>
