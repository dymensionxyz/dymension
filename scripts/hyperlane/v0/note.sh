# ~~~~~~~~~~~~~~~~~~~~~~~~
# PREAMBLE 

trash ~/.rollapp_evm
trash ~/.dymension
trash ~/.relayer

source /Users/danwt/Documents/dym/aaa-dym-notes/all_tasks/tasks/202503_feat_bridges/local_tests/v0/env.sh

# ~~~~~~~~~~~~~~~~~~~~~~~~
# HUB

make install

bash scripts/setup_local.sh
dymd start --log_level=debug

# ~~~~~~~~~~~~~~~~~~~~~~~~
# ROLLAPP

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

# ~~~~~~~~~~~~~~~~~~~~~~~~
# RELAYER

# It will show permission denied at first but should eventually work
sh scripts/ibc/setup_ibc.sh
# Can check with 
hub q sequencer list-sequencer
ra q sequencers sequencers
ra q sequencers whitelisted-relayers $SEQUENCER_ID

rly start hub-rollapp

# ~~~~~~~~~~
# Flow with ARAX, need to move some funsd to eibc fulfiller

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



# ~~~~~~~~~~
# Hyperlane 
# Follow steps in hyperlane scripts section to create mailbox, hooks etc

# Must wait for ibc token to arrive before making the collateral token
DENOM=$(hub q bank balances $HUB_USER_ADDR -o json | jq '.balances.[1].denom' -r); echo $DENOM; # The IBC denom
hub tx hyperlane-transfer dym-create-collateral-token $MAILBOX $DENOM "${DAN_FLAGS[@]}"
TOKEN_ID=$(curl -s http://localhost:1318/hyperlane/v1/tokens | jq '.tokens.[0].id' -r); echo $TOKEN_ID

ETH_TOKEN_CONTRACT="0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0" # arb
DST_DOMAIN=1 # arb
hub tx hyperlane-transfer enroll-remote-router $TOKEN_ID $DST_DOMAIN $ETH_TOKEN_CONTRACT 0 "${DAN_FLAGS[@]}"
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/remote_routers

# now the token should be ready...

# Get a memo:
RECOVERY_ADDR=$(hub keys show user -a) # a different addr to be able to check when a refund happens
#                            fee token       dst domain     recipient       amt  maxfee       recovery 
MEMO=$(hub q forward forward 100 $TOKEN_ID $DST_DOMAIN $ETH_TOKEN_CONTRACT 10000 20"$DENOM" $RECOVERY_ADDR)

HUB_USER_ADDR=$(dymd keys show hub-user -a)
ra tx ibc-transfer transfer transfer channel-0 $HUB_USER_ADDR 20000arax\
 --memo $MEMO --fees 20000000000000arax --from rol-user -b block --gas auto --gas-adjustment 1.5 -y

# after it's relayed, we should fulfill
hub q delayedack packets-by-rollapp $ROLLAPP_CHAIN_ID

#TODO: fix this part!

hub q eibc list-demand-orders pending
# look at the ID, fee
FULFILL_ID=$(hub q eibc list-demand-orders pending -o json | jq '.demand_orders.[0].id' -r); echo $FULFILL_ID;
hub tx eibc fulfill-order $FULFILL_ID 100 --from hub-user --fees 1dym --gas auto --gas-adjustment 1.5 -y
# check the tx, for fail/success

# check amt is suitable
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/bridged_supply
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/remote_routers

# Assuming OK, now check the reverse direction

ROL_USER_ADDR=$($EXECUTABLE keys show $KEY_NAME_ROLLAPP -a)

# TODO: fix hyperlane recipient to be arbitrary
HL_NONCE=1

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

# dymd tx hyperlane mailbox process [mailbox-id] [metadata] [message] [flags]
dymd tx hyperlane mailbox process $MAILBOX 0x $HL_MESSAGE --from hub-user --fees 60000000000000adym --gas auto --gas-adjustment 1.5 -y

ESCROW_ADDR=$(dymd q auth module-account forward -o json | jq '.account.value.address' -r); echo $ESCROW_ADDR;