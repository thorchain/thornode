#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
DEX_HOST="${DEX_HOST:=testnet-dex.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_HOST="${CHAIN_HOST:=127.0.0.1:1317}"
RPC_HOST="${RPC_HOST:=data-seed-pre-0-s3.binance.org}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
BINANCE_TESTNET="${BINANCE_TESTNET:=Binance-Chain-Nile}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=0}"
NODES="${NODES:=1}"
MASTER="${MASTER:=node1}" # the hostname of the master node
ROTATE_BLOCK_HEIGHT="${ROTATE_BLOCK_HEIGHT:=180}" # how often the pools in statechain should rotate

if [ -z ${ADDRESS+x} ]; then
  echo "GENERATING BNB ADDRESSES"
  # because the generate command can get API rate limited, we may need to retry
  n=0
  until [ $n -ge 60 ]; do
    generate > /tmp/bnb && break
    n=$[$n+1]
    sleep 1
  done
  ADDRESS=$(cat /tmp/bnb | grep MASTER= | awk -F= '{print $NF}')
  BINANCE_PRIVATE_KEY=$(cat /tmp/bnb | grep MASTER_KEY= | awk -F= '{print $NF}')
fi

# create statechain user
echo $SIGNER_PASSWD | sscli keys add statechain

VALIDATOR=$(ssd tendermint show-validator)
OBSERVER_ADDRESS=$(sscli keys show statechain -a)
NODE_ADDRESS=$(sscli keys show statechain -a)

if [[ "$MASTER" == "$(hostname)" ]]; then
  echo "I AM THE MASTER NODE"
  ssd tendermint show-node-id > /tmp/shared/node.txt
  echo $ADDRESS > /tmp/shared/pool_address.txt
fi

OBSERVER_PATH=$DB_PATH/observer/
SIGNER_PATH=$DB_PATH/signer/
mkdir -p $OBSERVER_PATH
mkdir -p $SIGNER_PATH
mkdir -p /etc/observe/observed
mkdir -p /etc/observe/signd

node() {
    echo "{\"node_address\": \"$1\" ,\"status\":\"active\",\"bond_address\":\"$2\",\"accounts\":{\"bnb_signer_acc\":\"$2\", \"bepv_validator_acc\": \"$3\", \"bep_observer_acc\": \"$1\"}}" > /tmp/shared/node_$1.json
    echo "{\"address\": \"$1\" ,\"key\":\"RotatePerBlockHeight\",\"value\":\"$ROTATE_BLOCK_HEIGHT\"}" > /tmp/shared/config_$1.json
}

node $NODE_ADDRESS $ADDRESS $VALIDATOR

# wait until we have the correct number of nodes in our directory before continuing
while [[ "$(ls -1 /tmp/shared/node_*.json | wc -l)" != "$NODES" ]]; do
    # echo "Waiting... '$(ls -1 /tmp/shared | wc -l)' '$NODES'"
    sleep 1
done

POOL_ADDRESS=$(cat /tmp/shared/pool_address.txt)

# Generate observer config file
echo "{
  \"chain_id\": \"$CHAIN_ID\",
  \"dex_host\": \"$DEX_HOST\",
  \"observer_db_path\": \"$OBSERVER_PATH\",
  \"block_scanner\": {
    \"rpc_host\": \"$RPC_HOST\",
    \"start_block_height\": \"$START_BLOCK_HEIGHT\",
    \"enforce_block_height\": false,
    \"block_scan_processors\": 1,
    \"block_height_discover_back_off\": \"1s\",
    \"block_retry_interval\": \"10s\"
  },
  \"state_chain\": {
    \"chain_id\": \"$CHAIN_ID\",
    \"chain_host\": \"$CHAIN_HOST\",
    \"signer_name\": \"$SIGNER_NAME\",
    \"signer_passwd\": \"$SIGNER_PASSWD\"
  },
  \"metric\": {
    \"enabled\": true
  },
  \"pool_address\": \"$POOL_ADDRESS\",
  \"signer_db_path\": \"$SIGNER_PATH\",
  \"binance\": {
    \"private_key\": \"$BINANCE_PRIVATE_KEY\",
    \"dex_host\": \"$DEX_HOST\"
  }
}" > /etc/observe/observed/config.json

# Generate Signer config file
echo "{
  \"chain_id\": \"$CHAIN_ID\",
  \"dex_host\": \"$DEX_HOST\",
  \"observer_db_path\": \"$OBSERVER_PATH\",
  \"block_scanner\": {
    \"rpc_host\": \"$RPC_HOST\",
    \"start_block_height\": $START_BLOCK_HEIGHT,
    \"enforce_block_height\": false,
    \"block_scan_processors\": 1,
    \"block_height_discover_back_off\": \"1s\",
    \"block_retry_interval\": \"10s\",
    \"scheme\": \"http\"
  },
  \"state_chain\": {
    \"chain_id\": \"statechain\",
    \"chain_host\": \"localhost:1317\",
    \"signer_name\": \"statechain\",
    \"signer_passwd\": \"PASSWORD1234\"
  },
  \"metric\": {
    \"enabled\": true
  },
  \"pool_address\": \"$POOL_ADDRESS\",
  \"signer_db_path\": \"$SIGNER_PATH\",
  \"binance\": {
    \"private_key\": \"$BINANCE_PRIVATE_KEY\",
    \"dex_host\": \"$DEX_HOST\"
  }
}" > /etc/observe/signd/config.json

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
if [[ "$MASTER" != "$(hostname)" ]]; then
  echo "I AM NOT THE MASTER"
  NODE_ID=$(cat /tmp/shared/node.txt)
  PEER="$NODE_ID@$MASTER:26656"
  ADDR='addr_book_strict = true'
  ADDR_STRICT_FALSE='addr_book_strict = false'
  PEERSISTENT_PEER_TARGET='persistent_peers = ""'

  sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.ssd/config/config.toml
  sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.ssd/config/config.toml
fi

if [[ "$MASTER" == "$(hostname)" ]]; then
    # copy the genesis file to shared directory, so we can access it later
    cp ~/.ssd/config/genesis.json /tmp/shared
fi

exec "$@"
