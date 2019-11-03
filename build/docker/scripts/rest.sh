#!/bin/sh
set -ex

CHAIN_DAEMON="${CHAIN_DAEMON:=127.0.0.1:26657}"
echo $CHAIN_DAEMON

$(dirname "$0")/wait-for-statechain-daemon.sh $CHAIN_DAEMON

exec "$@"
