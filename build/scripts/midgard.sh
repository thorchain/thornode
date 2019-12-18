#!/bin/sh
set -ex

PORT="${PORT:=8080}"

CHAIN_SCHEME="${CHAIN_SCHEME:=http}"
CHAIN_API="${CHAIN_API:=localhost:1317}"
CHAIN_RPC="${CHAIN_RPC:=localhost:26657}"

BINANCE_DEX="${BINANCE_HOST:=testnet-dex.binance.org}"
BINANCE_SCHEME="${BINANCE_SCHEME:=https}"
ISTESTNET="${ISTESTNET:=true}"

PG_HOST="${PG_HOST:=localhost}"
PG_PORT="${PG_PORT:=5432}"
PG_USERNAME="${PG_USERNAME:=midgard}"
PG_PASSWORD="${PG_PASSWORD:=password}"
PG_DB="${PG_DB:=midgard}"

mkdir -p /etc/midgard

echo "{
  \"listen_port\": $PORT,
  \"is_testnet\" : $ISTESTNET,
  \"log_level\": \"info\",
  \"thorchain\": {
    \"scheme\": \"$CHAIN_SCHEME\",
    \"host\": \"$CHAIN_API\",
    \"rpc_host\": \"$CHAIN_RPC\",
    \"enable_scan\": true,
    \"scan_start_pos\": 1,
    \"proxied_whitelisted_endpoints\": [
      \"pool_addresses\"
    ]
  },
  \"timescale\": {
    \"host\": \"$PG_HOST\",
    \"port\": $PG_PORT,
    \"user_name\": \"$PG_USERNAME\",
    \"password\": \"$PG_PASSWORD\",
    \"database\": \"$PG_DB\",
    \"sslmode\": \"disable\",
    \"migrationsDir\": \"./db/migrations/\"
  },
  \"binance\": {
    \"scheme\": \"$BINANCE_SCHEME\",
    \"dex_host\": \"$BINANCE_DEX\",
    \"is_testnet\": $ISTESTNET
  }
}" > /etc/midgard/config.json

exec "$@"
