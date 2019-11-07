#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
DEX_HOST="${DEX_HOST:=testnet-dex.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_HOST="${CHAIN_HOST:=127.0.0.1:1317}"
RPC_HOST="${RPC_HOST:=127.0.0.1:26657}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=1}"
NODE_ID="${NODE_ID:=none}"
SEED="${SEED:=none}" # the hostname of a seed node
GENESIS_URL="${GENESIS_URL:=none}"

$(dirname "$0")/wait-for-statechain-api.sh $CHAIN_HOST

if [ -f ~/.signer/private_key.txt ]; then
  BINANCE_PRIVATE_KEY=$(cat ~/.signer/private_key.txt)
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

SIGNER_PATH=$DB_PATH/signer/
mkdir -p $SIGNER_PATH
mkdir -p /etc/observe/signd

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
    \"chain_id\": \"$CHAIN_ID\",
    \"chain_host\": \"$CHAIN_HOST\",
    \"signer_name\": \"$SIGNER_NAME\",
    \"signer_passwd\": \"$SIGNER_PASSWD\"
  },
  \"metric\": {
    \"enabled\": true
  },
  \"signer_db_path\": \"$SIGNER_PATH\",
  \"binance\": {
    \"private_key\": \"$BINANCE_PRIVATE_KEY\",
    \"dex_host\": \"$DEX_HOST\"
  }
}" > /etc/observe/signd/config.json

exec "$@"
