this example can be executed using `podma` or `docker`

1. run `podman compose up`, this will spin up all the necessary infrastructure that includes a local hub, mock da layer, sequencer with `fraud proof` feature enabled and a full node

2. once you see `hub is ready with latest_block_height: 1` in the compose logs, you can attach to the nodes 
and proceed with the following steps

### hub

`podman exec -it fraud_proof_poc-hub-1 /bin/bash`, the following commands are executed inside the hub container

```sh
# fund the wallets that were generated during rollapp initialization
# extract the addresses to fund: you can find the generated wallets inside /go/rollapp_init file on the rollapp-evm and rollapp-evm-fullnode containers
for container in fraud_proof_poc-rollapp-evm-1 fraud_proof_poc-rollapp-evm-fullnode-1; do
  podman exec -it "$container" /bin/bash -c 'grep -Po "(dym[a-zA-Z0-9]+)" /go/rollapp_init' | grep -v 'dymension'
done

wallets=("dym1yk67wjkcu80fqegjskks3km9vz5xshknseqk4j" "dym1672tc2t0f7uq8kqlg2h8da6vm7mu5uhy08luu3" "dym1anfjre42pa7mtnqa0vce8cjpxk66d366v3gg7j" "dym137a5e6k5g2x9w5st2k2u9l80p565lx3qwl7uhp")

for wallet in "${wallets[@]}"; do
  echo "funding ${wallet}"
  dymd tx bank send local-user $wallet \
  10dym --gas-prices 100000000adym --yes -b block --keyring-backend test
done
```

### Sequencer

`podman exec -it fraud_proof_poc-rollapp-evm-1 /bin/bash`, the following commands are executed inside the sequencer container

```sh
# after funding the wallets from the hub node
# register the rollapp
roller tx register
# ! note down the rollapp-id 
# ! note down the node-id
rollapp-evm dymint show-node-id --home ~/.roller/rollapp

# start the sequencer with fraud_proof enabled
# '&' starts the rollapp in the background
rollapp-evm --home ~/.roller/rollapp start --dymint.simulate_fraud --low_level warn &

# ! start the full node before generating transactions
# generate transactions to create fraud with 0.5% probability
rollapp-evm --home ~/.roller/rollapp/ tx bank send \
  rollapp_sequencer ethm1wss9w8e89ntkn73n25lm6c7ul36u282c4sq5qm 100000adum \
  --keyring-backend test --broadcast-mode block -y --keyring-backend test
```

### Full Node
copy the `genesis.json` file from the sequencer to the full node

```sh
podman cp fraud_proof_poc_rollapp-evm_1:/root/.roller/rollapp/config/genesis.json \
    fraud_proof_poc_rollapp-evm-fullnode_1:/root/.roller/rollapp/config/genesis.json 
```

`podman exec -it fraud_proof_poc_rollapp-evm-fullnode_1 /bin/sh`, the following commands are executed inside the fullnode container

```sh
# noted in the sequencer steps
export ROLLAPP_CHAIN_ID="dummy_9361346-1"
# noted in the sequencer steps
export SEQUENCER_NODE_ID="12D3KooWQrNRe8ejp13aQauGze9UZw1kmgYR73K5iyzsKxVirLjz"

sed -i "s/^rollapp_id.*/rollapp_id = \"${ROLLAPP_CHAIN_ID}\"/" ~/.roller/rollapp/config/dymint.toml
sed -i "s/^seeds =.*/seeds = \"tcp:\/\/${SEQUENCER_NODE_ID}@$(dig +short rollapp-evm):26656\/\"/" ~/.roller/rollapp/config/config.toml

rollapp-evm --home ~/.roller/rollapp start
# once a fraud is created - the node will panic and generate a fraud proof file
# named `fraudProof_rollapp_with_tx.json`
```

outside the container, run

```sh
# copy and submit the fraud proof
podman cp fraud_proof_poc_rollapp-evm-fullnode_1:/go/fraudProof_rollapp_with_tx.json  \
  fraud_proof_poc_hub_1:/go/fraudProof_rollapp_with_tx.json 
```

inside the hub container, run the following command to submit the fraud

```sh
# from hub node
dymd tx rollapp submit-fraud ${ROLLAPP_CHAIN_ID} \
  /go/fraudProof_rollapp_with_tx.json --from local-user \
  --gas 50000000 -b block
```

if everything went through fine, hub log will show `fraud proof verified`