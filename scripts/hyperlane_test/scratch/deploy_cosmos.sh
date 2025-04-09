trash ~/.dymension
export SETTLEMENT_EXECUTABLE="dymd"
bash scripts/setup_local.sh
alias hub="dymd"

hub start --log_level=debug

export HUB_FLAGS=(--from hub-user --fees 20000000000000adym -y)

# create noop ism
hub tx hyperlane ism create-noop "${HUB_FLAGS[@]}"
ISM=$(curl -s http://localhost:1318/hyperlane/v1/isms | jq '.isms.[0].id' -r); echo $ISM;
LOCAL_DOMAIN=0

# create mailbox
# ism, local domain
hub tx hyperlane mailbox create  $ISM $LOCAL_DOMAIN "${HUB_FLAGS[@]}"
MAILBOX=$(curl -s http://localhost:1318/hyperlane/v1/mailboxes   | jq '.mailboxes.[0].id' -r); echo $MAILBOX;

# create noop hook
hub tx hyperlane hooks noop create "${HUB_FLAGS[@]}"
NOOP_HOOK=$(curl -s http://localhost:1318/hyperlane/v1/noop_hooks | jq '.noop_hooks.[0].id' -r); echo $NOOP_HOOK;

# TODO: IGP needed? Gas config?!!

# update mailbox
# mailbox, default hook (e.g. IGP), required hook (e.g. merkle tree)
hub tx hyperlane mailbox set $MAILBOX --default-hook $NOOP_HOOK --required-hook $NOOP_HOOK "${HUB_FLAGS[@]}"

# create warp route
DENOM="adym" # TODO: check
hub tx hyperlane-transfer create-collateral-token $MAILBOX $DENOM "${HUB_FLAGS[@]}"
TOKEN_ID=$(curl -s http://localhost:1318/hyperlane/v1/tokens | jq '.tokens.[1].id' -r); echo $TOKEN_ID;

# TODO: can optionally set override ISM for this token

# ~~~~~~~~~~~~~
# Now setup Ethereum
# ~~~~~~~~~~~~~

# create remote router 
ETH_TOKEN_CONTRACT="0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0"
ETH_RECIPIENT="0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0" # TODO: check
DST_DOMAIN=1
# token, dst domain, recipient contract, gas required on dst chain
hub tx hyperlane-transfer enroll-remote-router $TOKEN_ID $DST_DOMAIN $ETH_TOKEN_CONTRACT 0 "${HUB_FLAGS[@]}"
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/remote_routers

# transfer
# token, dst domain, recipient, amount
# optionally
# 	custom-hook-id: ""
# 	custom-hook-metadata: ""
# 	gas-limit: 0
# 	max-hyperlane-fee 0
TOKEN_AMT=1000
hub tx hyperlane-transfer transfer $TOKEN_ID $DST_DOMAIN $ETH_RECIPIENT $TOKEN_AMT --max-hyperlane-fee 200adym "${HUB_FLAGS[@]}"
curl -s http://localhost:1318/hyperlane/v1/tokens/$TOKEN_ID/bridged_supply

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