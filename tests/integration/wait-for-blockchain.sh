#!/bin/sh
# wait-for-blockchain.sh
# https://docs.docker.com/compose/startup-order/

set -e

echo "Waiting for Blockchain..."

cmd="$@"

until curl -s "$DAEMONHOST:$DAEMONPORT"; do
  >&2 echo "Blockchain is unavailable - sleeping"
  sleep 1
done

until curl -s "$APIHOST:$APIPORT/swapservice/ping"; do
  >&2 echo "Rest server is unavailable - sleeping"
  sleep 1
done

# sleep a little more to give time to add its first block
sleep 10

>&2 echo "Blockchain is up - executing command"
exec $cmd
