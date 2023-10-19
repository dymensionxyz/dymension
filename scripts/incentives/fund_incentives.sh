#!/bin/bash

# Retrieve the recipient address
recipient_address=$(dymd q auth module-account poolincentives -o json | jq -r '.account.base_account.address')

# Create the proposal JSON file
cat <<EOL > proposal.json
{
  "title": "Community Pool Spend",
  "description": "Fund pool incentives from the community pool",
  "recipient": "$recipient_address",
  "amount": "30dym",
  "deposit": "1dym"
}
EOL

# Pass fund manually to the poolincentives module
dymd tx bank send local-user dym1upfuxznarpja3sywq0tzd2kktg9wv8mczfufge 30dym --keyring-backend test -b block

# Fund the community pool
dymd tx distribution fund-community-pool 100dym --from local-user -b block -y

# Submit the proposal
dymd tx gov submit-legacy-proposal community-pool-spend proposal.json --from local-user -b block --gas auto -y

# Vote on the proposal
last_proposal_id=$(dymd q gov proposals -o json | jq '.proposals | map(.id | tonumber) | max')
dymd tx gov vote "$last_proposal_id" yes --from local-user -b block



