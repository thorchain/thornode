#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_HOST="${CHAIN_HOST:=127.0.0.1:1317}"
RPC_SCHEME="${RPC_RPC_SCHEME:=https}"
BINANCE_HOST="${BINANCE_HOST:=https://data-seed-pre-0-s3.binance.org}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"

$(dirname "$0")/wait-for-statechain-api.sh $CHAIN_HOST

OBSERVER_PATH=$DB_PATH/observer/
mkdir -p $OBSERVER_PATH
mkdir -p /etc/observe/observed

# Generate observer config file
echo "{
  \"chain_id\": \"$CHAIN_ID\",
  \"binance_host\": \"$BINANCE_HOST\",
  \"observer_db_path\": \"$OBSERVER_PATH\",
  \"block_scanner\": {
    \"rpc_host\": \"$BINANCE_HOST\",
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
  \"signer_db_path\": \"$SIGNER_PATH\",
  \"binance\": {
    \"private_key\": \"$BINANCE_PRIVATE_KEY\",
    \"rpc_host\": \"$RPC_HOST\"
  }
}" > /etc/observe/observed/config.json

exec "$@"
