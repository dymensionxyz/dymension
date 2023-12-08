#! /bin/bash

tmp=$(mktemp)

set_gov_params() {
    echo "setting gov params"

    jq '.app_state.gov.deposit_params.min_deposit[0].denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    jq '.app_state.gov.deposit_params.min_deposit[0].amount = "10000000000"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    #Two weeks voting_period
    jq '.app_state.gov.voting_params.voting_period = "1m"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_hub_params() {
    echo "setting hub params"
    jq '.app_state.rollapp.params.dispute_period_in_blocks = "2"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_EVM_params() {
  jq '.consensus_params["block"]["max_gas"] = "40000000"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state["feemarket"]["params"]["no_base_fee"] = true' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state.evm.params.evm_denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state.evm.params.enable_create = false' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

#Adding a "minute" epoch
set_epochs_params() {
    jq '.app_state.epochs.epochs += [{
    "identifier": "minute",
    "start_time": "0001-01-01T00:00:00Z",
    "duration": "60s",
    "current_epoch": "0",
    "current_epoch_start_time": "0001-01-01T00:00:00Z",
    "epoch_counting_started": false,
    "current_epoch_start_height": "0"
    }]' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

#should be set to days on live net and lockable duration to 2 weeks
set_incentives_params() {
  jq '.app_state.incentives.params.distr_epoch_identifier = "minute"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state.incentives.params.distr_epoch_identifier = "minute"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
  jq '.app_state.incentives.lockable_durations = ["60s"]' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}


set_misc_params() {
    jq '.app_state.crisis.constant_fee.denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    jq -r '.app_state.gamm.params.pool_creation_fee[0].denom = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    jq '.app_state["txfees"]["basedenom"] = "udym"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
    jq '.app_state["txfees"]["params"]["epoch_identifier"] = "minute"' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

set_bank_denom_metadata() {
    jq '.app_state.bank.denom_metadata = [
        {
            "base": "udym",
            "denom_units": [
                {
                    "aliases": [],
                    "denom": "udym",
                    "exponent": 0
                },
                {
                    "aliases": [],
                    "denom": "DYM",
                    "exponent": 18
                }
            ],
            "description": "Denom metadata for DYM (udym)",
            "display": "DYM",
            "name": "DYM",
            "symbol": "DYM"
        }
    ]' "$GENESIS_FILE" > "$tmp" && mv "$tmp" "$GENESIS_FILE"
}

enable_monitoring() {
    sed -i'' -e "/\[telemetry\]/,+8 s/enabled = .*/enabled = true/" "$APP_CONFIG_FILE"
    sed  -i'' -e "s/^prometheus-retention-time =.*/prometheus-retention-time = 31104000/" "$APP_CONFIG_FILE"
    sed  -i'' -e "s/^prometheus =.*/prometheus = true/" "$TENDERMINT_CONFIG_FILE"
    sed -ie 's/enabled-unsafe-cors.*$/enabled-unsafe-cors = true/' "$APP_CONFIG_FILE"
    sed -ie 's/enable-unsafe-cors.*$/enabled-unsafe-cors = true/' "$APP_CONFIG_FILE"
    sed -ie 's/cors_allowed_origins.*$/cors_allowed_origins = ["*"]/' "$TENDERMINT_CONFIG_FILE"
}