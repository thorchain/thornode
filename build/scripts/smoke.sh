#!/bin/sh

#
# Smoke Tests.
#
# Run our smoke tests against a Thorchain instance.
#

cd "$(dirname "$0")"/../docker
NET=$1

if [ -z "$NET" ]; then
  NET="testnet"
fi

make NET="$NET" validate-smoke-test
