#!/bin/bash

# set the LP1 as incentivised pool
dymd tx gov submit-legacy-proposal update-pool-incentives  1 100  --from local-user -b block --gas auto -y --title sdasd --description dasdas --deposit 1dym

# Vote on the proposal
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block

# verify with:
#dymd q poolincentives distr-info


