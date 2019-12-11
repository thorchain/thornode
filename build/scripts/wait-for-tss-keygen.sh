#!/bin/sh
# https://docs.docker.com/compose/startup-order/

set -xe

echo "Waiting for TSS Keygen..."

until curl -s "$1"; do
  echo "TSS Keygen is unavailable - sleeping ($1)"
  sleep 3
done

echo "TSS Keygen is available!"
