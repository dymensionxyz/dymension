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
    <td>User1 claims</td>
    <td>3.5</td>
    <td>100</td>
    <td>60</td>
    <td>40</td>
    <td>3.5</td>
    <td>✅ Claimed: 140 DYM (3.5 - 3.5) * 40 = 0 DYM</td>
    <td>0</td>
    <td>60</td>
    <td>2.5</td>
    <td>(3.5 - 2.5) * 60 = 60 DYM</td>
    <td>0</td>
  </tr>
  <tr>
    <td>User2 un-endorses</td>
    <td>3.5</td>
    <td>100 - 60 = 40</td>
    <td>0</td>
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
    <td>100</td>
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
    <td>User1 un-endorses</td>
    <td>6.0</td>
    <td>40 - 40 = 0</td>
    <td>0</td>
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
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>✅ Claimed: 100 DYM 0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>60 DYM</td>
  </tr>
  <tr>
    <td>User2 re-endorses w 100 shares</td>
    <td>6.0</td>
    <td>0 + 100 = 100</td>
    <td>0</td>
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
    <td><b>New epoch</b>: +100 DYM</td>
    <td>6 + (100 / 100) = 7.0</td>
    <td>100</td>
    <td>100</td>
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
    <td>User2 stakes 100 DYM</td>
    <td>7.0</td>
    <td>100 + 100 = 200</td>
    <td>100</td>
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
    <td>User2 claims</td>
    <td>7.0</td>
    <td>200</td>
    <td>100</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>0</td>
    <td>200</td>
    <td>7.0</td>
    <td>(7 - 7) * 200 = 0 DYM</td>
    <td>✅ Claimed: 160 DYM 0</td>
  </tr>
</table>

Example scenario – user stakes 100 more and then unstakes 50; a new epoch starts; finally, user claims:

<table>
  <tr>
    <th>Event</th>
    <th>Global Accumulator (GA)</th>
    <th>Total Shares</th>
    <th>Unlocked Coins</th>
    <th>User Shares</th>
    <th>User LSA</th>
    <th>User Claimable</th>
    <th>User AUR</th>
  </tr>
  <tr>
    <td>+100 DYM unlocked</td>
    <td>7.0</td>
    <td>100</td>
    <td>100</td>
    <td>100</td>
    <td>6.0</td>
    <td>(7 - 6) * 100 = 100 DYM</td>
    <td>0</td>
  </tr>
  <tr>
    <td>User stakes 100</td>
    <td>7.0</td>
    <td>100 + 100 = 200</td>
    <td>100</td>
    <td>200</td>
    <td>7.0</td>
    <td>(7 - 7) * 100 = 0 DYM</td>
    <td>(7 - 6) * 100 = 100 DYM</td>
  </tr>
  <tr>
    <td>User unstakes 50</td>
    <td>7.0</td>
    <td>200 - 50 = 150</td>
    <td>100</td>
    <td>150</td>
    <td>7.0</td>
    <td>(7 - 7) * 100 = 0 DYM</td>
    <td>100 DYM</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>7.0 + (100 / 150) = 7,67</td>
    <td>150</td>
    <td>200</td>
    <td>150</td>
    <td>7.0</td>
    <td>(7.67 - 7) * 150 = 100 DYM</td>
    <td>100 DYM</td>
  </tr>
  <tr>
    <td><b>New epoch</b>: +100 DYM</td>
    <td>7,67</td>
    <td>150</td>
    <td>200</td>
    <td>150</td>
    <td>7,67</td>
    <td>✅ Claimed: 100 DYM (7.67 - 7,67) * 150 = 0 DYM</td>
    <td>✅ Claimed: 100 DYM 0 DYM</td>
  </tr>
</table>

Example scenario – user unstakes all coins:

<table>
  <tr>
    <th>Event</th>
    <th>Global Accumulator (GA)</th>
    <th>Total Shares</th>
    <th>Unlocked Coins</th>
    <th>User Shares (S)</th>
    <th>User LSA</th>
    <th>User Claimable (C)</th>
    <th>User AUR</th>
  </tr>
  <tr>
    <td>+100 DYM unlocked</td>
    <td>7.0</td>
    <td>100</td>
    <td>100</td>
    <td>100</td>
    <td>6.0</td>
    <td>(7 - 6) * 100 = 100 DYM</td>
    <td>0</td>
  </tr>
  <tr>
    <td>User unstakes 100</td>
    <td>7.0</td>
    <td>100 - 100 = 0</td>
    <td>100</td>
    <td>0</td>
    <td>7.0</td>
    <td>(7 - 7) * 0 = 0 DYM</td>
    <td>(7 - 6) * 100 = 100 DYM</td>
  </tr>
</table>

Example scenario – user endorses when they did not have the vote before:

<table>
  <tr>
    <th>Event</th>
    <th>Global Accumulator</th>
    <th>Total Shares</th>
    <th>Unlocked Coins</th>
    <th>User Shares (S)</th>
    <th>User LSA</th>
    <th>User Claimable (C)</th>
    <th>User AUR</th>
  </tr>
  <tr>
    <td>+100 DYM unlocked</td>
    <td>7.0</td>
    <td>100</td>
    <td>100</td>
    <td>-</td>
    <td>-</td>
    <td>-</td>
    <td>0</td>
  </tr>
  <tr>
    <td>User endorses 100</td>
    <td>7.0</td>
    <td>100 + 100 = 200</td>
    <td>100</td>
    <td>100</td>
    <td>7.0</td>
    <td>(7 - 7) * 100 = 0 DYM</td>
    <td>(7 - 0) * 0 = 0 DYM</td>
  </tr>
</table>

Example scenario – user casts for the *empty* endorsement (i.e., no one endorsed to it yet)

<table>
  <tr>
    <th>Event</th>
    <th>Global Accumulator</th>
    <th>Total Shares</th>
    <th>Unlocked Coins</th>
    <th>User Shares</th>
    <th>User LSA</th>
    <th>User Claimable</th>
    <th>User AUR</th>
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
  </tr>
  <tr>
    <td>User endorses 100</td>
    <td>0</td>
    <td>100</td>
    <td>0</td>
    <td>100</td>
    <td>0</td>
    <td>(0 - 0) * 100 = 0 DYM</td>
    <td>(0 - 0) * 0 = 0 DYM</td>
  </tr>
</table>
