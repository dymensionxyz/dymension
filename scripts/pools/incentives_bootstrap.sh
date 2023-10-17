#/bin/bash
DENOM=gamm/pool/1




dymd tx incentives create-gauge gamm/pool/1 100dym \
--duration 1h --start-time 1640081402 --epochs 20 --from user -b block -y

