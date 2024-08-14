#!/bin/bash

# fund streamer
dymd tx bank send user \
        dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv \
        300000dym --keyring-backend test -y --gas-prices 100000000adym
sleep 7

# create new gauges for lockdrop
echo "Creating gauges for lockdrop of uatom"
dymd tx incentives create-gauge uatom 30dym --duration="60s" --perpetual --from user --gas auto -y --gas-prices 100000000adym
sleep 7

# create first stream for the LP holders
echo "Gov proposal for creating new stream with LP1 and LP2 as incentives targets"
dymd tx gov submit-legacy-proposal create-stream-proposal 1,2 40,60 20000dym --epoch-identifier minute --from user --title sfasfas --description ddasda --deposit 11dym -y --gas auto --gas-prices 100000000adym
sleep 7

last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')

dymd tx gov vote "$last_proposal_id" yes --from hub-user -y  --gas-prices 100000000adym
dymd tx gov vote "$last_proposal_id" yes --from user -y  --gas-prices 100000000adym
dymd tx gov vote "$last_proposal_id" yes --from pools -y  --gas-prices 100000000adym
sleep 7

# create second stream for the Lockdrop
echo "Gov proposal for creating new stream for lockdrop"
dymd tx gov submit-legacy-proposal create-stream-proposal 3 100 10000dym --epoch-identifier minute --from user --title sfasfas --description ddasda --deposit 11dym -y --gas auto --gas-prices 100000000adym
sleep 7

last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')

dymd tx gov vote "$last_proposal_id" yes --from hub-user -y  --gas-prices 100000000adym
dymd tx gov vote "$last_proposal_id" yes --from user -y  --gas-prices 100000000adym
dymd tx gov vote "$last_proposal_id" yes --from pools -y  --gas-prices 100000000adym
sleep 7

# create sponsored stream
echo "Gov proposal for creating a new sponsored stream"
dymd tx gov submit-legacy-proposal create-stream-proposal - - 10000dym --sponsored --epoch-identifier minute --from user --title sfasfas --description ddasda --deposit 11dym -y --gas auto --gas-prices 100000000adym
sleep 7

last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')

dymd tx gov vote "$last_proposal_id" yes --from hub-user -y  --gas-prices 100000000adym
dymd tx gov vote "$last_proposal_id" yes --from user -y  --gas-prices 100000000adym
dymd tx gov vote "$last_proposal_id" yes --from pools -y  --gas-prices 100000000adym
