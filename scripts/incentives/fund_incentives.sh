#!/bin/bash

# fund streamer
dymd tx bank send local-user dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv 300000dym --keyring-backend test -b block


# create stream governance proposal
dymd tx gov submit-legacy-proposal create-stream-proposal dym1upfuxznarpja3sywq0tzd2kktg9wv8mczfufge 100000dym --from local-user -b block --title sfasfas --description ddasda --deposit 1dym

# Vote on the proposal
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block




# --------------------- Fund streamer through governance --------------------- #
# Retrieve the streamer module address
recipient_address=$(dymd q auth module-account streamer -o json | jq -r '.account.base_account.address' | tr -d '"')

# Create the proposal JSON file
cat <<EOL > proposal.json
{
  "title": "Community Pool Spend",
  "description": "Fund streamer from the community pool",
  "recipient": "$recipient_address",
  "amount": "30dym",
  "deposit": "1dym"
}
EOL

# Fund the community pool
dymd tx distribution fund-community-pool 100dym --from local-user -b block -y

# Submit the proposal
dymd tx gov submit-legacy-proposal community-pool-spend proposal.json --from local-user -b block --gas auto -y

# Vote on the proposal
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block



