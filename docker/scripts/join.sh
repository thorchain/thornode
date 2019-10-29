#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
DEX_HOST="${DEX_HOST:=testnet-dex.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_HOST="${CHAIN_HOST:=127.0.0.1:1317}"
RPC_HOST="${RPC_HOST:=data-seed-pre-0-s3.binance.org}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=0}"
NODE_ID="${NODE_ID:=none}"
SEED="${SEED:=none}" # the hostname of a seed node
GENESIS_URL="${GENESIS_URL:=none}"

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
echo "YOUR NODE ADDRESS: $NODE_ADDRESS. Send your bond with this as your address."

POOL_ADDRESS=$(curl url $SEED:1317/swapservice/pooladdresses | jq -r .current)

OBSERVER_PATH=$DB_PATH/observer/
SIGNER_PATH=$DB_PATH/signer/
mkdir -p $OBSERVER_PATH
mkdir -p $SIGNER_PATH
mkdir -p /etc/observe/observed
mkdir -p /etc/observe/signd

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
sscli config chain-id statechain
sscli config output json
sscli config indent true
sscli config trust-node true

curl $GENESIS_URL --output ~/.ssd/config/genesis.json
ssd validate-genesis

PEER="$NODE_ID@$SEED:26656"
ADDR='addr_book_strict = true'
ADDR_STRICT_FALSE='addr_book_strict = false'
PEERSISTENT_PEER_TARGET='persistent_peers = ""'

sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.ssd/config/config.toml
sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.ssd/config/config.toml

exec "$@"
