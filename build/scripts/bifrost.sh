#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
BINANCE_HOST="${BINANCE_HOST:=https://data-seed-pre-0-s3.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_API="${CHAIN_API:=127.0.0.1:1317}"
CHAIN_RPC="${CHAIN_RPC:=127.0.0.1:26657}"
SIGNER_NAME="${SIGNER_NAME:=thorchain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=1}"
TSS_SCHEME="${TSS_SCHEME:=http}"
TSS_HOST="${TSS_HOST:=127.0.0.1}"
TSS_PORT="${TSS_PORT:=4040}"

$(dirname "$0")/wait-for-thorchain-api.sh $CHAIN_API

OBSERVER_PATH=$DB_PATH/bifrost/observer/
SIGNER_PATH=$DB_PATH/bifrost/signer/

mkdir -p $SIGNER_PATH $OBSERVER_PATH /etc/bifrost

# Generate bifrost config file
echo "{
    \"thorchain\": {
        \"chain_id\": \"$CHAIN_ID\",
        \"chain_host\": \"$CHAIN_API\",
        \"signer_name\": \"$SIGNER_NAME\"
    },
    \"metrics\": {
        \"enabled\": true
    },
    \"chains\": [
      {
        \"chain_id\": \"BNB\",
        \"rpc_host\": \"$BINANCE_HOST\",
        \"block_scanner\": {
          \"rpc_host\": \"$BINANCE_HOST\",
          \"enforce_block_height\": false,
          \"block_scan_processors\": 1,
          \"block_height_discover_back_off\": \"1s\",
          \"block_retry_interval\": \"10s\",
          \"chain_id\": \"BNB\",
          \"http_request_timeout\": \"30s\",
          \"http_request_read_timeout\": \"30s\",
          \"http_request_write_timeout\": \"30s\",
          \"max_http_request_retry\": 10,
          \"start_block_height\": 0,
          \"db_path\": \"$OBSERVER_PATH\"
        }
      }
    ],
    \"tss\": {
        \"scheme\": \"$TSS_SCHEME\",
        \"host\": \"$TSS_HOST\",
        \"port\": $TSS_PORT
    },
    \"signer\": {
      \"signer_db_path\": \"$SIGNER_PATH\",
      \"block_scanner\": {
        \"rpc_host\": \"$CHAIN_RPC\",
        \"start_block_height\": $START_BLOCK_HEIGHT,
        \"enforce_block_height\": false,
        \"block_scan_processors\": 1,
        \"block_height_discover_back_off\": \"1s\",
        \"block_retry_interval\": \"10s\",
        \"scheme\": \"http\"
      }
    }
}" > /etc/bifrost/config.json

export SIGNER_PASSWD
exec "$@"
