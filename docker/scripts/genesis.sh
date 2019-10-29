#!/bin/sh
set -ex

NODES="${NODES:=1}"
SEED="${SEED:=node1}" # the hostname of the master node
ROTATE_BLOCK_HEIGHT="${ROTATE_BLOCK_HEIGHT:=0}" # how often the pools in statechain should rotate

# create statechain user
echo $SIGNER_PASSWD | sscli keys add statechain

VALIDATOR=$(ssd tendermint show-validator)
OBSERVER_ADDRESS=$(sscli keys show statechain -a)
NODE_ADDRESS=$(sscli keys show statechain -a)

if [[ "$SEED" == "$(hostname)" ]]; then
  echo "I AM THE SEED NODE"
  ssd tendermint show-node-id > /tmp/shared/node.txt
  echo $ADDRESS > /tmp/shared/pool_address.txt
fi

node() {
    echo "{\"node_address\": \"$1\" ,\"status\":\"active\",\"bond_address\":\"$2\",\"accounts\":{\"bnb_signer_acc\":\"$2\", \"bepv_validator_acc\": \"$3\", \"bep_observer_acc\": \"$1\"}}" > /tmp/shared/node_$1.json
    if [[ "$ROTATE_BLOCK_HEIGHT" != "0" ]]; then
        echo "{\"address\": \"$1\" ,\"key\":\"RotatePerBlockHeight\",\"value\":\"$ROTATE_BLOCK_HEIGHT\"}" > /tmp/shared/config_$1.json
    fi
}

node $NODE_ADDRESS $ADDRESS $VALIDATOR

# wait until we have the correct number of nodes in our directory before continuing
while [[ "$(ls -1 /tmp/shared/node_*.json | wc -l)" != "$NODES" ]]; do
    # echo "Waiting... '$(ls -1 /tmp/shared | wc -l)' '$NODES'"
    sleep 1
done

# Setup SSD
ssd init local --chain-id statechain
for f in /tmp/shared/node_*.json; do 
    ssd add-genesis-account $(cat $f | jq -r .node_address) 100000000000thor
done
sscli config chain-id statechain
sscli config output json
sscli config indent true
sscli config trust-node true

# add node accounts to genesis file
for f in /tmp/shared/node_*.json; do 
    jq --argjson nodeInfo "$(cat $f)" '.app_state.swapservice.node_accounts += [$nodeInfo]' ~/.ssd/config/genesis.json > /tmp/genesis.json
    mv /tmp/genesis.json ~/.ssd/config/genesis.json
done

for f in /tmp/shared/config_*.json; do 
    jq --argjson config "$(cat $f)" '.app_state.swapservice.admin_configs += [$config]' ~/.ssd/config/genesis.json > /tmp/genesis.json
    mv /tmp/genesis.json ~/.ssd/config/genesis.json
done

cat ~/.ssd/config/genesis.json
ssd validate-genesis

# setup peer connection
if [[ "$SEED" != "$(hostname)" ]]; then
  echo "I AM NOT THE SEED"
  NODE_ID=$(cat /tmp/shared/node.txt)
  PEER="$NODE_ID@$SEED:26656"
  ADDR='addr_book_strict = true'
  ADDR_STRICT_FALSE='addr_book_strict = false'
  PEERSISTENT_PEER_TARGET='persistent_peers = ""'

  sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.ssd/config/config.toml
  sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.ssd/config/config.toml
fi

if [[ "$SEED" == "$(hostname)" ]]; then
    # copy the genesis file to shared directory, so we can access it later
    cp ~/.ssd/config/genesis.json /tmp/shared
fi

exec "$@"
