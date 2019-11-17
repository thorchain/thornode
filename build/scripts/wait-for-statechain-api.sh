#!/bin/sh
# https://docs.docker.com/compose/startup-order/

set -xe

echo "Waiting for Statechain API..."

until curl -s "$1/thorchain/ping"; do
  # echo "Rest server is unavailable - sleeping"
  sleep 1
done

echo "Rest server is up!"
