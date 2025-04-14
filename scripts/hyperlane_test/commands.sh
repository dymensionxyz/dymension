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
sleep 0.5
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

# It will show permission denied at first
sh scripts/ibc/setup_ibc.sh

# Quickly update whitelist!
RLY_ADDR=$(rly keys show $ROLLAPP_CHAIN_ID)
dymd tx sequencer update-whitelisted-relayers $RLY_ADDR --from sequencer --keyring-dir ~/.rollapp_evm/sequencer_keys --keyring-backend test -y --fees 1dym
# can check with
hub q sequencer list-sequencer -o json | jq '.sequencers[0].whitelisted_relayers'

# relayer should eventually make the bridge

# Can check with 
hub q sequencer list-sequencer
ra q sequencers sequencers # populate sequencer addr
ra q sequencers whitelisted-relayers $SEQUENCER_ADDR # It can take a while for this to show up

# after channel etc finished:
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
# STEP: DEPLOY HYPERLANE ENTITIES TO HUB

# create noop ism
hub tx hyperlane ism create-noop "${HUB_FLAGS[@]}"
ISM=$(curl -s http://localhost:1318/hyperlane/v1/isms | jq '.isms.[0].id' -r); echo $ISM;

# create mailbox
# ism, local domain
hub tx hyperlane mailbox create  $ISM $LOCAL_DOMAIN "${HUB_FLAGS[@]}"
MAILBOX=$(curl -s http://localhost:1318/hyperlane/v1/mailboxes   | jq '.mailboxes.[0].id' -r); echo $MAILBOX;

# create noop hook
hub tx hyperlane hooks noop create "${HUB_FLAGS[@]}"
NOOP_HOOK=$(curl -s http://localhost:1318/hyperlane/v1/noop_hooks | jq '.noop_hooks.[0].id' -r); echo $NOOP_HOOK;

# TODO: IGP needed? Gas config?!! (don't think so, for this test)

# update mailbox
# mailbox, default hook (e.g. IGP), required hook (e.g. merkle tree)
hub tx hyperlane mailbox set $MAILBOX --default-hook $NOOP_HOOK --required-hook $NOOP_HOOK "${HUB_FLAGS[@]}"

############################
# STEP: TEST END-TO-END: ROLLAPP -> HUB -> HYPERLANE
# (First we test putting collateral tokens into the HL warp route, then we test the other direction using those escrowed tokens)

# create the Hyperlane token 
# NOTE: MUST wait for the ibc token to arrive from the genesis bridge transfer to be able to make the collateral token with the right denom
DENOM=$(hub q bank balances $HUB_USER_ADDR -o json | jq '.balances.[1].denom' -r); echo $DENOM; # The IBC denom
hub tx hyperlane-transfer dym-create-collateral-token $MAILBOX $DENOM "${HUB_FLAGS[@]}"
TOKEN_ID=$(curl -s http://localhost:1318/hyperlane/v1/tokens | jq '.tokens.[0].id' -r); echo $TOKEN_ID

# setup the router
hub tx hyperlane-transfer enroll-remote-router $TOKEN_ID $DST_DOMAIN $ETH_TOKEN_CONTRACT 0 "${HUB_FLAGS[@]}"
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/remote_routers # check

# prepare the memo which will be included in the rollapp outbound transfer
# args are [eibc-fee] [token-id] [destination-domain] [recipient] [amount] [max-fee] [recovery-address] [flags]
EIBC_FEE=100
MEMO=$(hub q forward memo-eibc-to-hl $EIBC_FEE $TOKEN_ID $DST_DOMAIN $ETH_TOKEN_CONTRACT 10000 20"$DENOM")

# note, make sure to relay here! If relayer is crashed, restart before initiating the transfer on the next step!

# initiate the transfer from the rollapp
HUB_USER_ADDR=$(dymd keys show hub-user -a)
ra tx ibc-transfer transfer transfer channel-0 $HUB_USER_ADDR 20000arax\
 --memo $MEMO --fees 20000000000000arax --from rol-user -b block --gas auto --gas-adjustment 1.5 -y

# wait for relaying, then fulfill the order with EIBC
ORDER_ID=$(hub q eibc list-demand-orders pending -o json | jq '.demand_orders.[0].id' -r); echo $ORDER_ID; # TODO: check index
hub tx eibc fulfill-order $ORDER_ID $EIBC_FEE --from hub-user --fees 1dym --gas auto --gas-adjustment 1.5 -y

# confirm the result, the token should be bridged: 
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/bridged_supply

############################
# STEP: TEST END-TO-END: HYPERLANE -> HUB -> ROLLAPP
# We have now bridged tokens into the collateral warp route, so we can test the reverse direction immediately with the same token

ROL_USER_ADDR=$($EXECUTABLE keys show $KEY_NAME_ROLLAPP -a)

# get the message to be sent directly the HL server on the Hub. This is just a test utility.
HL_MESSAGE=$(dymd q forward hyperlane-message\
 $HL_NONCE\
 $DST_DOMAIN\
 $ETH_TOKEN_CONTRACT\
 $LOCAL_DOMAIN\
 $TOKEN_ID\
 $HUB_USER_ADDR\
 50\
 "channel-0"\
 $ROL_USER_ADDR\
 50$DENOM\
 5m\
 $RECOVERY_ADDR\
 ); echo $HL_MESSAGE;

# again, make sure relayer is up and running before initiating the transfer
dymd tx hyperlane mailbox process $MAILBOX 0x $HL_MESSAGE --from hub-user --fees 60000000000000adym --gas auto --gas-adjustment 1.5 -y

# while waiting for the tokens to arrive on the rollapp, can check the escrow address on the hub
ESCROW_ADDR=$(dymd q auth module-account forward -o json | jq '.account.value.address' -r); echo $ESCROW_ADDR;

############################
# APPENDIX: EXTRA DEBUG TOOLS

# Queries
# http://localhost:1318/hyperlane/v1/tokens
# http://localhost:1318/hyperlane/v1/tokens/{id}
# http://localhost:1318/hyperlane/v1/tokens/{id}/bridged_supply
# http://localhost:1318/hyperlane/v1/tokens/{id}/remote_routers
# http://localhost:1318/hyperlane/v1/isms
# http://localhost:1318/hyperlane/v1/isms/{id}
# http://localhost:1318/hyperlane/v1/mailboxes
# http://localhost:1318/hyperlane/v1/mailboxes/{id}
# http://localhost:1318/hyperlane/v1/recipient_ism/{recipient}
# http://localhost:1318/hyperlane/v1/verify_dry_run
# http://localhost:1318/hyperlane/v1/igps
# http://localhost:1318/hyperlane/v1/igps/{id}
# http://localhost:1318/hyperlane/v1/merkle_tree_hooks
# http://localhost:1318/hyperlane/v1/merkle_tree_hooks/{id}
# http://localhost:1318/hyperlane/v1/noop_hooks
# http://localhost:1318/hyperlane/v1/noop_hooks/{id}