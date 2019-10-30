#!/bin/sh
# https://docs.docker.com/compose/startup-order/

set -xe

echo "Waiting for Statechain Daemon..."

until curl -s "$1"; do
  echo "Statechain daemon is unavailable - sleeping ($1)"
  sleep 3
done

sleep 5 # wait for first block to become available

echo "Statechain daemon is up!"
