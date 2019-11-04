#!/bin/bash
set -ex

SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
NODES="${NODES:=1}"
SEED="${SEED:=thor1}" # the hostname of the master node
ROTATE_BLOCK_HEIGHT="${ROTATE_BLOCK_HEIGHT:=5}" # how often the pools in statechain should rotate

if [ -f ~/.signer/private_key.txt ]; then
    PUBKEY=$(cat ~/.signer/pubkey.txt)
    ADDRESS=$(cat ~/.signer/address.txt)
else
    echo "GENERATING BNB ADDRESSES"
    # because the generate command can get API rate limited, we may need to retry
    n=0
    until [ $n -ge 60 ]; do
        generate > /tmp/bnb && break
        n=$[$n+1]
        sleep 1
    done
    ADDRESS=$(cat /tmp/bnb | grep MASTER= | awk -F= '{print $NF}')
    mkdir -p ~/.signer
    echo $ADDRESS > ~/.signer/address.txt
    BINANCE_PRIVATE_KEY=$(cat /tmp/bnb | grep MASTER_KEY= | awk -F= '{print $NF}')
    echo $BINANCE_PRIVATE_KEY > ~/.signer/private_key.txt
    PUBKEY=$(cat /tmp/bnb | grep MASTER_PUBKEY= | awk -F= '{print $NF}')
    echo $PUBKEY > ~/.signer/pubkey.txt
fi

# create statechain user
echo $SIGNER_PASSWD | thorcli keys add $SIGNER_NAME 2>&1

VALIDATOR=$(thord tendermint show-validator)
OBSERVER_ADDRESS=$(thorcli keys show statechain -a)
NODE_ADDRESS=$(thorcli keys show statechain -a)

if [ "$SEED" == "$(hostname)" ]; then
    echo "I AM THE SEED NODE"
    thord tendermint show-node-id > /tmp/shared/node.txt
    echo $PUBKEY > /tmp/shared/pubkey.txt
    echo $ADDRESS > /tmp/shared/pool_address.txt
fi

# write node account data to json file in shared directory
echo "{\"node_address\": \"$NODE_ADDRESS\" ,\"status\":\"active\",\"bond_address\":\"$ADDRESS\",\"accounts\":{\"bnb_signer_acc\":\"$PUBKEY\", \"bepv_validator_acc\": \"$VALIDATOR\", \"bep_observer_acc\": \"$NODE_ADDRESS\"}}" > /tmp/shared/node_$NODE_ADDRESS.json
# write rotate block height as config file
if [ "$ROTATE_BLOCK_HEIGHT" != "0" ]; then
    echo "{\"address\": \"$NODE_ADDRESS\" ,\"key\":\"RotatePerBlockHeight\",\"value\":\"$ROTATE_BLOCK_HEIGHT\"}" > /tmp/shared/config_$NODE_ADDRESS.json
fi

# wait until we have the correct number of nodes in our directory before continuing
while [ "$(ls -1 /tmp/shared/node_*.json | wc -l)" != "$NODES" ]; do
    # echo "Waiting... '$(ls -1 /tmp/shared | wc -l)' '$NODES'"
    sleep 1
done

POOL_ADDRESS=$(cat /tmp/shared/pool_address.txt)
echo "{\"current\":\"$POOL_ADDRESS\",\"rotate_at\":\"28800\",\"rotate_window_open_at\":\"27800\"}" > /tmp/shared/pool_addresses.json

if [ "$SEED" == "$(hostname)" ]; then
    if [ ! -f ~/.thord/config/genesis.json ]; then
        # Setup SSD
        thord init local --chain-id statechain
        for f in /tmp/shared/node_*.json; do 
            thord add-genesis-account $(cat $f | jq -r .node_address) 100000000000thor
        done
        thorcli config chain-id statechain
        thorcli config output json
        thorcli config indent true
        thorcli config trust-node true

        # add node accounts to genesis file
        for f in /tmp/shared/node_*.json; do 
            jq --argjson nodeInfo "$(cat $f)" '.app_state.swapservice.node_accounts += [$nodeInfo]' ~/.thord/config/genesis.json > /tmp/genesis.json
            mv /tmp/genesis.json ~/.thord/config/genesis.json
        done

        jq --argjson addr "$(cat /tmp/shared/pool_addresses.json)" '.app_state.swapservice.pool_addresses = $addr' ~/.thord/config/genesis.json > /tmp/genesis.json
        mv /tmp/genesis.json ~/.thord/config/genesis.json

        if [[ -f /tmp/shared/config_*.json ]]; then
            for f in /tmp/shared/config_*.json; do 
                jq --argjson config "$(cat $f)" '.app_state.swapservice.admin_configs += [$config]' ~/.thord/config/genesis.json > /tmp/genesis.json
                mv /tmp/genesis.json ~/.thord/config/genesis.json
            done
        fi

        cat ~/.thord/config/genesis.json
        thord validate-genesis

        cp ~/.thord/config/genesis.json /tmp/shared
    fi
fi

# setup peer connection
if [ "$SEED" != "$(hostname)" ]; then
    if [ ! -f ~/.thord/config/genesis.json ]; then
        echo "I AM NOT THE SEED"
        
        thord init local --chain-id statechain
        thorcli config chain-id statechain
        thorcli config output json
        thorcli config indent true
        thorcli config trust-node true

        NODE_ID=$(cat /tmp/shared/node.txt)
        PEER="$NODE_ID@$SEED:26656"
        echo "PEER $PEER"
        ADDR='addr_book_strict = true'
        ADDR_STRICT_FALSE='addr_book_strict = false'
        PEERSISTENT_PEER_TARGET='persistent_peers = ""'

        sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.thord/config/config.toml
        sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.thord/config/config.toml

        while [ ! -f /tmp/shared/genesis.json ]; do
            sleep 1
        done

        cp /tmp/shared/genesis.json ~/.thord/config/genesis.json

        cat ~/.thord/config/genesis.json
        thord validate-genesis
    fi
fi

thord start --rpc.laddr tcp://0.0.0.0:26657 &> ~/daemon.log &
