#!/bin/bash

echo "Delegate to the active validator"
# dymvaloper139mq752delxv78jvtmwxhasyrycufsvrd7p5rl is a validator address of hub-user
dymd tx staking delegate dymvaloper139mq752delxv78jvtmwxhasyrycufsvrd7p5rl 10000dym --from user -y --gas-prices 100000000adym
sleep 7

echo "Voting on gauge as 'user'"
dymd tx sponsorship vote 1=30,2=30,3=40 --from user -y  --gas-prices 100000000adym
sleep 7

echo "Voting on gauge as 'hub-user'"
dymd tx sponsorship vote 1=20,2=20 --from hub-user -y  --gas-prices 100000000adym
