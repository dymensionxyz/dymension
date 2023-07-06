#! /bin/bash

tmp=$(mktemp)

set_minting_params() {
    echo "setting minting params"

    # jq '.app_state.mint.minter.inflation = "0.300000000000000000"' genesis.json > "$tmp" && mv "$tmp" genesis.json
    # jq '.app_state.mint.params.blocks_per_year = "6311520"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    # jq '.app_state.mint.params.goal_bonded = "0.670000000000000000"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_distribution_params() {
    echo "setting distribution params"
}

set_gov_params() {
    echo "setting gov params"

    jq '.app_state.gov.deposit_params.min_deposit[0].denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    jq '.app_state.gov.deposit_params.min_deposit[0].amount = "10000000000"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    #Two weeks voting_period
    jq '.app_state.gov.voting_params.voting_period = "30s"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_ibc_params() {
    echo "setting IBC params"
    # jq '.app_state.packetforwardmiddleware.params.fee_percentage = "0.00"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_hub_params() {
    echo "setting hub params"
    jq '.app_state.rollapp.params.dispute_period_in_blocks = "2"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_staking_slashing_params() {
    echo "setting slashing params"
    # jq '.app_state.slashing.params.slash_fraction_downtime = "0.01"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"

    echo "setting staking params"
    jq '.app_state.staking.params.max_validators = 110' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"

}

set_EVM_params() {
  jq '.consensus_params["block"]["max_gas"] = "40000000"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state["feemarket"]["params"]["no_base_fee"] = true' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state.evm.params.evm_denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state.evm.params.enable_create = false' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}


set_misc_params() {
    jq '.app_state.crisis.constant_fee.denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

enable_monitoring() {
    sed -i'' -e "/\[telemetry\]/,+8 s/enabled = .*/enabled = true/" "$APP_CONFIG_FILE"
    sed  -i'' -e "s/^prometheus-retention-time =.*/prometheus-retention-time = 31104000/" "$APP_CONFIG_FILE"
    sed  -i'' -e "s/^prometheus =.*/prometheus = true/" "$TENDERMINT_CONFIG_FILE"
    sed -ie 's/enabled-unsafe-cors.*$/enabled-unsafe-cors = true/' "$APP_CONFIG_FILE"
    sed -ie 's/enable-unsafe-cors.*$/enabled-unsafe-cors = true/' "$APP_CONFIG_FILE"
    sed -ie 's/cors_allowed_origins.*$/cors_allowed_origins = ["*"]/' "$TENDERMINT_CONFIG_FILE"
}

set_rollapp_deployer_whitelist() {
  echo "NOT SUPPORTED"
  #TODO
}

add_genesis_accounts() {
    echo "NOT SUPPORTED"
  #TODO
}