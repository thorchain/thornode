#!/bin/sh
set -ex

SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
NODES="${NODES:=1}"
SEED="${SEED:=node1}" # the hostname of the master node
ROTATE_BLOCK_HEIGHT="${ROTATE_BLOCK_HEIGHT:=0}" # how often the pools in statechain should rotate

if [ -f ~/.signer/private_key.txt ]; then
  PUBKEY=$(cat ~/.signer/pubkey.txt)
else
  echo "GENERATING BNB ADDRESSES"
  # because the generate command can get API rate limited, we may need to retry
  n=0
  until [ $n -ge 60 ]; do
    generate > /tmp/bnb && break
    n=$[$n+1]
    sleep 1
  done
  BINANCE_PRIVATE_KEY=$(cat /tmp/bnb | grep MASTER_KEY= | awk -F= '{print $NF}')
  echo $BINANCE_PRIVATE_KEY > ~/.signer/private_key.txt
  PUBKEY=$(cat /tmp/bnb | grep MASTER_PUBKEY= | awk -F= '{print $NF}')
  echo $PUBKEY > ~/.signer/pubkey.txt
fi

# create statechain user
echo $SIGNER_PASSWD | sscli keys add $SIGNER_NAME

VALIDATOR=$(ssd tendermint show-validator)
OBSERVER_ADDRESS=$(sscli keys show statechain -a)
NODE_ADDRESS=$(sscli keys show statechain -a)

if [[ "$SEED" == "$(hostname)" ]]; then
  echo "I AM THE SEED NODE"
  ssd tendermint show-node-id > /tmp/shared/node.txt
  echo $PUBKEY > /tmp/shared/pubkey.txt
fi

# write node account data to json file in shared directory
echo "{\"node_address\": \"$NODE_ADDRESS\" ,\"status\":\"active\",\"bond_address\":\"$PUBKEY\",\"accounts\":{\"bnb_signer_acc\":\"$PUBKEY\", \"bepv_validator_acc\": \"$VALIDATOR\", \"bep_observer_acc\": \"$NODE_ADDRESS\"}}" > /tmp/shared/node_$NODE_ADDRESS.json
# write rotate block height as config file
if [[ "$ROTATE_BLOCK_HEIGHT" != "0" ]]; then
  echo "{\"address\": \"$NODE_ADDRESS\" ,\"key\":\"RotatePerBlockHeight\",\"value\":\"$ROTATE_BLOCK_HEIGHT\"}" > /tmp/shared/config_$NODE_ADDRESS.json
fi

# wait until we have the correct number of nodes in our directory before continuing
while [[ "$(ls -1 /tmp/shared/node_*.json | wc -l)" != "$NODES" ]]; do
    # echo "Waiting... '$(ls -1 /tmp/shared | wc -l)' '$NODES'"
    sleep 1
done

POOL_ADDRESS=$(cat /tmp/shared/pubkey.txt)
echo "{\"current\":\"$POOL_ADDRESS\"}" > /tmp/shared/pool_addresses.json


if [ ! -f ~/.ssd/config/genesis.json ]; then
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

    jq --argjson addr "$(cat /tmp/shared/pool_addresses.json)" '.app_state.swapservice.pool_addresses = $addr' ~/.ssd/config/genesis.json > /tmp/genesis.json
    mv /tmp/genesis.json ~/.ssd/config/genesis.json

    if [[ -f /tmp/shared/config_*.json ]]; then
        for f in /tmp/shared/config_*.json; do 
            jq --argjson config "$(cat $f)" '.app_state.swapservice.admin_configs += [$config]' ~/.ssd/config/genesis.json > /tmp/genesis.json
            mv /tmp/genesis.json ~/.ssd/config/genesis.json
        done
    fi
fi

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
