#!/bin/bash

set -ex

# This script checks to see if this node should teardown itself. This is
# intended for testnet testing scenarios.

# Required: jq, docker

# check we can get address
docker exec -it thor-daemon thorcli keys show thorchain -a

# get our node address
address=$(docker exec -it thor-daemon thorcli keys show thorchain -a | tr -d "\r")
# address="hello"

node_status=$(curl -s localhost:1317/thorchain/nodeaccount/$address | jq -r '.status')
node_status_since=$(curl -s localhost:1317/thorchain/nodeaccount/$address | jq -r '.status_since')

if [ "$node_status" = "active" ]; then
    echo "node is still active... exiting"
    exit 0
fi

if [ "$node_status_since" = "0" ]; then
    echo "node is hasn't been churned in yet... exiting"
    exit 0
fi

# we have been churned out, we should shutdown
echo "node has been churned out, ready to be shutdown"
shutdown
