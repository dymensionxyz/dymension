#!/bin/bash
set -e

export PROPOSAL_PATH=./scripts/tee/example_proposal.json
export POLICY_VALUES_PATH=./x/rollapp/keeper/testdata/tee/insecure_policy_values.json
export POLICY_STRUCTURE_PATH=./x/rollapp/keeper/testdata/tee/insecure_policy.rego
export POLICY_QUERY_PATH=./x/rollapp/keeper/testdata/tee/insecure_query.rego
export GCP_ROOT_CERT_PATH=./x/rollapp/keeper/testdata/tee/confidential_space_root.pem

./scripts/tee/prepare_proposal.sh