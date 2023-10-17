#!/bin/bash

# Retrieve the recipient address
recipient_address=$(dymd q auth module-account poolincentives -o json | jq -r '.account.base_account.address')

# Create the proposal JSON file
cat <<EOL > proposal.json
{
  "title": "Community Pool Spend",
  "description": "Fund pool incentives from the community pool",
  "recipient": "$recipient_address",
  "amount": "100dym",
  "deposit": "1dym"
}
EOL

# Fund the community pool
dymd tx distribution fund-community-pool 100dym --from local-user -b block -y

# Submit the proposal
dymd tx gov submit-legacy-proposal community-pool-spend proposal.json --from local-user -b block --gas auto -y

# Vote on the proposal
    dymd tx gov vote 1 yes --from local-user -b block