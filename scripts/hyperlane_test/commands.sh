###########################
# Prelims
# 1. dymension/ branch with forward module, hyperlane cosmos integration etc
# 2. rollapp-evm branch (https://github.com/dymensionxyz/rollapp-evm/pull/476/files) TODO: latest might work now
# 3. relayer

# For a fresh start
trash ~/.rollapp_evm
trash ~/.dymension
trash ~/.relayer

# get all environment variables set (required in multiple terminals if using multiple)
BASE_PATH="/Users/danwt/Documents/dym/d-dymension/scripts/hyperlane_test"
source $BASE_PATH/env.sh

#########################################################################################
#########################################################################################
# Q: WHAT IS THIS?
# A: It's not a script, but rather some commands, which should be copy pasted as appropriate per the instructions, while in the right directories.
#########################################################################################

############################
# STEP: LAUNCH HUB + ROLLAPP AND OPEN IBC BRIDGE
# for this, it might be out of date, so check the rollapp-evm repo scripts, and the dymension//scripts/setup_local.sh if something doesn't work

###########
# Hub
cd dymension/
bash scripts/setup_local.sh
dymd start --log_level=debug

###########
# Rollapp 
cd rollapp-evm/

make install BECH32_PREFIX=$BECH32_PREFIX

$EXECUTABLE config keyring-backend test
sh scripts/init.sh

dymd keys add sequencer --keyring-dir ~/.rollapp_evm/sequencer_keys --keyring-backend test

SEQUENCER_ADDR=`dymd keys show sequencer --address --keyring-backend test --keyring-dir ~/.rollapp_evm/sequencer_keys`
BOND_AMOUNT="$(dymd q rollapp params -o json | jq -r '.params.min_sequencer_bond_global.amount')$(dymd q rollapp params -o json | jq -r '.params.min_sequencer_bond_global.denom')"
NUMERIC_PART=$(echo $BOND_AMOUNT | sed 's/adym//')
NEW_NUMERIC_PART=$(echo "$NUMERIC_PART + 100000000000000000000" | bc)
TRANSFER_AMOUNT="${NEW_NUMERIC_PART}adym"

dymd tx bank send $HUB_KEY_WITH_FUNDS $SEQUENCER_ADDR ${TRANSFER_AMOUNT} --keyring-backend test --broadcast-mode sync --fees 1dym -y --node ${HUB_RPC_URL}

sh scripts/settlement/register_rollapp_to_hub.sh
sh scripts/settlement/register_sequencer_to_hub.sh

dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "settlement_layer" -v "dymension"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "node_address" -v "$HUB_RPC_URL"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "rollapp_id" -v "$ROLLAPP_CHAIN_ID"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "max_idle_time" -v "2s" # may want to change to something longer after setup (see below)
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "max_proof_time" -v "1s"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/app.toml "minimum-gas-prices" -v "1arax"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "batch_submit_time" -v "30s"

$EXECUTABLE validate-genesis
# will not work unless https://github.com/dymensionxyz/rollapp-evm/pull/471 included
$EXECUTABLE start --log_level=debug

make install BECH32_PREFIX=$BECH32_PREFIX

$EXECUTABLE config keyring-backend test
sh scripts/init.sh

dymd keys add sequencer --keyring-dir ~/.rollapp_evm/sequencer_keys --keyring-backend test

SEQUENCER_ADDR=`dymd keys show sequencer --address --keyring-backend test --keyring-dir ~/.rollapp_evm/sequencer_keys`
BOND_AMOUNT="$(dymd q rollapp params -o json | jq -r '.params.min_sequencer_bond_global.amount')$(dymd q rollapp params -o json | jq -r '.params.min_sequencer_bond_global.denom')"
NUMERIC_PART=$(echo $BOND_AMOUNT | sed 's/adym//')
NEW_NUMERIC_PART=$(echo "$NUMERIC_PART + 100000000000000000000" | bc)
TRANSFER_AMOUNT="${NEW_NUMERIC_PART}adym"

dymd tx bank send $HUB_KEY_WITH_FUNDS $SEQUENCER_ADDR ${TRANSFER_AMOUNT} --keyring-backend test --broadcast-mode sync --fees 1dym -y --node ${HUB_RPC_URL}

sh scripts/settlement/register_rollapp_to_hub.sh
sh scripts/settlement/register_sequencer_to_hub.sh

dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "settlement_layer" -v "dymension"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "node_address" -v "$HUB_RPC_URL"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "rollapp_id" -v "$ROLLAPP_CHAIN_ID"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "max_idle_time" -v "2s" # may want to change to something longer after setup (see below)
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "max_proof_time" -v "1s"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/app.toml "minimum-gas-prices" -v "1arax"
dasel put -f "${ROLLAPP_HOME_DIR}"/config/dymint.toml "batch_submit_time" -v "30s"

$EXECUTABLE validate-genesis
# will not work unless https://github.com/dymensionxyz/rollapp-evm/pull/471 included
$EXECUTABLE start --log_level=debug

###########
# Relayer 

# It will show permission denied at first but should eventually work
sh scripts/ibc/setup_ibc.sh
# Can check with 
hub q sequencer list-sequencer
ra q sequencers sequencers
ra q sequencers whitelisted-relayers $SEQUENCER_ID

rly start hub-rollapp

###########
# Send some ARAX funds to the Hub to open the bridge and make sure the ibc denom is registered

HUB_USER_ADDR=$(dymd keys show hub-user -a)
ra tx ibc-transfer transfer transfer channel-0 $HUB_USER_ADDR 1000000arax\
 --memo '{"eibc":{"fee":"10"}}' --fees 20000000000000arax --from rol-user -b block --gas auto --gas-adjustment 1.5 -y
# Check progress with  
# When allowed do
hub q delayedack packets-by-rollapp $ROLLAPP_CHAIN_ID
PROOF_HEIGHT=$(hub q delayedack packets-by-rollapp $ROLLAPP_CHAIN_ID -o json | jq '.rollappPackets.[0].ProofHeight' -r)
PACKET_SEQ=$(hub q delayedack packets-by-rollapp $ROLLAPP_CHAIN_ID -o json | jq '.rollappPackets.[0].packet.sequence' -r)
# try until success, in the meantime, can set up Hyperlane (below)
hub tx delayedack finalize-packet $ROLLAPP_CHAIN_ID $PROOF_HEIGHT ON_RECV channel-0 $PACKET_SEQ --from hub-user --gas auto --gas-adjustment 1.5 --fees 1dym -y

############################
# STEP: DEPLOY HYPERLANE CONFIG TO HUB
# for this, it might be out of date, so check the rollapp-evm repo scripts, and the dymension//scripts/setup_local.sh if something doesn't work