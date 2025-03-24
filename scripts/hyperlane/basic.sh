trash ~/.dymension
export SETTLEMENT_EXECUTABLE="dymd"
bash scripts/setup_local.sh
alias hub="dymd"

# should put swagger = true
hub start --log_level=debug

# export HUB_RPC_ENDPOINT="localhost"
# export HUB_RPC_PORT="36657" # default: 36657
# export HUB_RPC_URL="http://${HUB_RPC_ENDPOINT}:${HUB_RPC_PORT}"
# export HUB_CHAIN_ID="dymension_100-1"
# dymd config set chain-id ${HUB_CHAIN_ID}
# dymd config set node ${HUB_RPC_URL}

# Hyperlane https://github.com/bcp-innovations/hyperlane-cosmos/blob/mbreithecker/readme/README.md
export HYPD_FLAGS=--home test --chain-id hyperlane-local --keyring-backend test --from alice --fees 40000uhyp
export DAN_FLAGS=(--from hub-user --fees 20000000000000adym)

/Users/danwt/.dymension/config/app.toml

# create noop ism
hub tx hyperlane ism create-noop "${DAN_FLAGS[@]}"
ISM=$(curl -s http://localhost:1318/hyperlane/v1/isms | jq '.isms.[0].id' -r)
LOCAL_DOMAIN=0

# create mailbox
# ism, local domain
hub tx hyperlane mailbox create  $ISM $LOCAL_DOMAIN "${DAN_FLAGS[@]}"
MAILBOX=$(curl -s http://localhost:1318/hyperlane/v1/mailboxes   | jq '.mailboxes.[0].id' -r)

# create noop hook
hub tx hyperlane hooks noop create "${DAN_FLAGS[@]}"
NOOP_HOOK=$(curl -s http://localhost:1318/hyperlane/v1/noop_hooks | jq '.noop_hooks.[0].id' -r)

# TODO: IGP needed? Gas config?!!

# update mailbox
# mailbox, default hook (e.g. IGP), required hook (e.g. merkle tree)
hub tx hyperlane mailbox set $MAILBOX --default-hook $NOOP_HOOK --required-hook $NOOP_HOOK "${DAN_FLAGS[@]}"

# create warp route
DENOM="adym" # TODO: check
hub tx hyperlane-transfer create-collateral-token $MAILBOX $DENOM "${DAN_FLAGS[@]}"
TOKEN_ID=$(curl -s http://localhost:1318/hyperlane/v1/tokens | jq '.tokens.[0].id' -r)

# TODO: can optionally set override ISM for this token

# create remote router 
ARB_CONTRACT="0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0"
DST_DOMAIN=1
# token, dst domain, recipient contract, gas required on dst chain
hub tx hyperlane-transfer enroll-remote-router $TOKEN_ID $DST_DOMAIN $ARB_CONTRACT 0 "${DAN_FLAGS[@]}"

# transfer
# token, dst domain, recipient, amount
ARB_RECIPIENT="0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0" # TODO: check
# optionally
# 	custom-hook-id: ""
# 	custom-hook-metadata: ""
# 	gas-limit: 0
# 	max-hyperlane-fee 0
TOKEN_AMT=1000
hub tx hyperlane-transfer transfer $TOKEN_ID $DST_DOMAIN $ARB_RECIPIENT $TOKEN_AMT --max-hyperlane-fee 0adym "${DAN_FLAGS[@]}"
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/bridged_supply

