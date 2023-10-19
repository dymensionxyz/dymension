#!/bin/bash

# create new gauges for lockdrop
echo "Creating gauges for lockdrop of uatom"
dymd tx incentives create-gauge uatom 30dym --duration="3600s" --epochs 30 --from local-user -b block --gas auto -y

# set the LP1 as incentivised pool
echo "Setting the LP1 and the lockdrop as incentives target"
dymd tx gov submit-legacy-proposal update-pool-incentives  1,3 50,50  --from local-user -b block --gas auto -y --title sdasd --description dasdas --deposit 1dym

# Vote on the proposal
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block

# verify with:
#dymd q poolincentives distr-info


