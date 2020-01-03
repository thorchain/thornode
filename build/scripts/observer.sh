#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_API="${CHAIN_API:=127.0.0.1:1317}"
BINANCE_HOST="${BINANCE_HOST:=https://data-seed-pre-0-s3.binance.org}"
SIGNER_NAME="${SIGNER_NAME:=thorchain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"

$(dirname "$0")/wait-for-thorchain-api.sh $CHAIN_API

OBSERVER_PATH=$DB_PATH/observer/
mkdir -p $OBSERVER_PATH
mkdir -p /etc/observe/observed

# Generate observer config file
echo "{
  \"observer\": {
      \"observer_db_path\": \"$OBSERVER_PATH\",
      \"block_scanner\": {
        \"rpc_host\": \"$BINANCE_HOST\",
        \"enforce_block_height\": false,
        \"block_scan_processors\": 1,
        \"block_height_discover_back_off\": \"1s\",
        \"block_retry_interval\": \"10s\"
      },
      \"thorchain\": {
        \"chain_id\": \"$CHAIN_ID\",
        \"chain_host\": \"$CHAIN_API\",
        \"signer_name\": \"$SIGNER_NAME\",
        \"signer_passwd\": \"$SIGNER_PASSWD\"
      },
      \"metric\": {
        \"enabled\": true
      },
      \"binance\": {
        \"rpc_host\": \"$BINANCE_HOST\"
      }
  }
}" > /etc/observe/observed/config.json

exec "$@"
