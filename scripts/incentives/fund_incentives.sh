#!/bin/bash

# fund streamer
dymd tx bank send local-user \
        dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv \
        300000dym --keyring-backend test -b block -y --gas-prices 100000000adym

# create new gauges for lockdrop
echo "Creating gauges for lockdrop of uatom"
dymd tx incentives create-gauge uatom 30dym --duration="60s" --perpetual --from local-user -b block --gas auto -y --gas-prices 100000000adym

# create first stream for the LP holders
echo "Gov proposal for creating new stream with LP1 and LP2 as incentives targets"
dymd tx gov submit-legacy-proposal create-stream-proposal 1,2 40,60 20000dym --epoch-identifier minute --from local-user -b block --title sfasfas --description ddasda --deposit 11dym -y --gas auto --gas-prices 100000000adym
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block -y  --gas-prices 100000000adym

# create second stream for the Lockdrop
echo "Gov proposal for creating new stream for lockdrop"
dymd tx gov submit-legacy-proposal create-stream-proposal 3 100 10000dym --epoch-identifier minute --from local-user -b block --title sfasfas --description ddasda --deposit 1dym -y --gas auto --gas-prices 100000000adym
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block -y  --gas-prices 100000000adym
~
